package genesyscloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v72/platformclientv2"
)

func getAllFlows(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	architectAPI := platformclientv2.NewArchitectApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 50
		flows, _, err := architectAPI.GetFlows(nil, pageNum, pageSize, "", "", nil, "", "", "", "", "", "", "", "", false, true, "", "", nil)
		if err != nil {
			return nil, diag.Errorf("Failed to get page of flows: %v", err)
		}

		if flows.Entities == nil || len(*flows.Entities) == 0 {
			break
		}

		for _, flow := range *flows.Entities {
			resources[*flow.Id] = &ResourceMeta{Name: *flow.Name}
		}
	}

	return resources, nil
}

func flowExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllFlows),
		RefAttrs:         map[string]*RefAttrSettings{},
		UnResolvableAttributes: map[string]*schema.Schema{
			"filepath": resourceFlow().Schema["filepath"],
		},
	}
}

func resourceFlow() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Flow`,

		CreateContext: createWithPooledClient(createFlow),
		ReadContext:   readWithPooledClient(readFlow),
		UpdateContext: updateWithPooledClient(updateFlow),
		DeleteContext: deleteWithPooledClient(deleteFlow),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"filepath": {
				Description:  "YAML file path or URL for flow configuration.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validatePath,
			},
			"file_content_hash": {
				Description: "Hash value of the YAML file content. Used to detect changes.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"substitutions": {
				Description: "A substitution is a key value pair where the key is the value you want to replace, and the value is the value to substitute in its place.",
				Type:        schema.TypeMap,
				Optional:    true,
			},
		},
	}
}

func createFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	// TODO: After public API endpoint is published and exposed to public, change to SDK method instead of direct invocation
	apiClient := &architectAPI.Configuration.APIClient
	path := architectAPI.Configuration.BasePath + "/api/v2/flows/jobs"

	headerParams := make(map[string]string)

	// add default headers if any
	for key := range architectAPI.Configuration.DefaultHeader {
		headerParams[key] = architectAPI.Configuration.DefaultHeader[key]
	}

	headerParams["Authorization"] = "Bearer " + architectAPI.Configuration.AccessToken
	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	successPayload := make(map[string]interface{})
	response, err := apiClient.CallAPI(path, "POST", nil, headerParams, nil, nil, "", nil)
	if err != nil {
		return diag.Errorf("Failed to initiate archy job %s", err)
	} else if err == nil && response.Error != nil {
		return diag.Errorf("Failed to register Archy job. %s", err)
	} else {
		err = json.Unmarshal(response.RawBody, &successPayload)
		if err != nil {
			return diag.Errorf("Failed to unmarshal response after registering the Archy job. %s", err)
		}
	}

	// Once the endpoint is ready for SDK, can extract data from the InitiateArchitectJobResponse type, instead of map of string interface
	presignedUrl := successPayload["presignedUrl"].(string)
	jobId := successPayload["id"].(string)
	headers := successPayload["headers"].(map[string]interface{})

	filePath := d.Get("filepath").(string)
	substitutions := d.Get("substitutions").(map[string]interface{})

	// Upload flow configuration file
	_, err = prepareAndUploadFile(filePath, substitutions, headers, presignedUrl)
	if err != nil {
		return diag.Errorf(err.Error())
	}

	// Pre-define here before entering retry function, otherwise it will be overwritten
	flowID := ""

	// Retry every 15 seconds to fetch job status for 16 minutes until job succeeds or fails
	retryErr := withRetries(ctx, 16*time.Minute, func() *resource.RetryError {
		// TODO: After public API endpoint is published and exposed to public, change to SDK method instead of direct invocation
		path := architectAPI.Configuration.BasePath + "/api/v2/flows/jobs/" + jobId + "?expand=messages"
		res := make(map[string]interface{})
		// If possible, after changing to SDK method invocation, include correlationId we get earlier in this function when making the GET request
		response, err := apiClient.CallAPI(path, "GET", nil, headerParams, nil, nil, "", nil)
		if err != nil {
			// Nothing special to do here, but do avoid processing the response
		} else if err == nil && response.Error != nil {
			resource.NonRetryableError(fmt.Errorf("Error retrieving job status. JobID: %s, error: %s ", jobId, response.ErrorMessage))
		} else {
			err = json.Unmarshal(response.RawBody, &res)
			if err != nil {
				resource.NonRetryableError(fmt.Errorf("Unable to unmarshal response when retrieving job status. JobID: %s, error: %s ", jobId, response.ErrorMessage))
			}
		}
		// Once SDK is ready, get status from ArchitectJobStateResponse type
		if res["status"] == "Failure" {
			return resource.NonRetryableError(fmt.Errorf("Flow publish failed. JobID: %s, tracing messages: %v ", jobId, res["messages"].([]interface{})))
		}

		// Once SDK is ready, get status from ArchitectJobStateResponse type
		if res["status"] == "Success" {
			// Once SDK is ready, get flow id from ArchitectJobStateResponse type
			flowID = res["flow"].(map[string]interface{})["id"].(string)
			return nil
		}

		time.Sleep(15 * time.Second) // Wait 15 seconds for next retry
		return resource.RetryableError(fmt.Errorf("Job (%s) could not finish in 16 minutes and timed out ", jobId))
	})

	if retryErr != nil {
		return retryErr
	}

	if flowID == "" {
		return diag.Errorf("Failed to get the flowId from Architect Job (%s).", jobId)
	}

	d.Set("file_content_hash", hashFileContent(d.Get("filepath").(string)))
	d.SetId(flowID)
	log.Printf("Created flow %s.", d.Id())
	return readFlow(ctx, d, meta)
}

func readFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	filePath := d.Get("filepath").(string)
	if filePath != "" {
		fileContentHash := hashFileContent(filePath)
		if fileContentHash != d.Get("file_content_hash") {
			d.Set("file_content_hash", fileContentHash)
			log.Println("Detected change to config file, updating")
			return updateFlow(ctx, d, meta)
		}
	}

	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		flow, resp, err := architectAPI.GetFlow(d.Id(), false)
		if err != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read flow %s: %s", d.Id(), err))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read flow %s: %s", d.Id(), err))
		}

		log.Printf("Read flow %s %s", d.Id(), *flow.Name)
		return nil
	})
}

func updateFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	// TODO: After public API endpoint is published and exposed to public, change to SDK method instead of direct invocation
	apiClient := &architectAPI.Configuration.APIClient
	path := architectAPI.Configuration.BasePath + "/api/v2/flows/jobs"

	headerParams := make(map[string]string)

	// add default headers if any
	for key := range architectAPI.Configuration.DefaultHeader {
		headerParams[key] = architectAPI.Configuration.DefaultHeader[key]
	}

	headerParams["Authorization"] = "Bearer " + architectAPI.Configuration.AccessToken
	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	successPayload := make(map[string]interface{})
	response, err := apiClient.CallAPI(path, "POST", nil, headerParams, nil, nil, "", nil)
	if err != nil {
		return diag.Errorf("Failed to update archy job %s", err)
	} else if err == nil && response.Error != nil {
		return diag.Errorf("Failed to register Archy job. %s", err)
	} else {
		err = json.Unmarshal([]byte(response.RawBody), &successPayload)
		if err != nil {
			return diag.Errorf("Failed to unmarshal response after registering the Archy job. %s", err)
		}
	}

	// Once the endpoint is ready for SDK, can extract data from the InitiateArchitectJobResponse type, instead of map of string interface
	presignedUrl := successPayload["presignedUrl"].(string)
	jobId := successPayload["id"].(string)
	headers := successPayload["headers"].(map[string]interface{})

	filePath := d.Get("filepath").(string)
	substitutions := d.Get("substitutions").(map[string]interface{})

	_, err = prepareAndUploadFile(filePath, substitutions, headers, presignedUrl)
	if err != nil {
		return diag.Errorf(err.Error())
	}

	// Pre-define here before entering retry function, otherwise it will be overwritten
	flowID := ""

	retryErr := withRetries(ctx, 16*time.Minute, func() *resource.RetryError {
		// TODO: After public API endpoint is published and exposed to public, change to SDK method instead of direct invocation
		path := architectAPI.Configuration.BasePath + "/api/v2/flows/jobs/" + jobId + "?expand=messages"
		res := make(map[string]interface{})
		// If possible, after changing to SDK method invocation, include correlationId we get earlier in this function when making the GET request
		response, err := apiClient.CallAPI(path, "GET", nil, headerParams, nil, nil, "", nil)
		if err != nil {
			// Nothing special to do here, but do avoid processing the response
		} else if err == nil && response.Error != nil {
			resource.NonRetryableError(fmt.Errorf("Error retrieving job status. JobID: %s, error: %s ", jobId, response.ErrorMessage))
		} else {
			err = json.Unmarshal(response.RawBody, &res)
			if err != nil {
				resource.NonRetryableError(fmt.Errorf("Unable to unmarshal response when retrieving job status. JobID: %s, error: %s ", jobId, response.ErrorMessage))
			}
		}
		// Once SDK is ready, get status from ArchitectJobStateResponse type
		if res["status"] == "Failure" {
			return resource.NonRetryableError(fmt.Errorf("Flow publish failed. JobID: %s, tracing messages: %v ", jobId, res["messages"].([]interface{})))
		}

		// Once SDK is ready, get status from ArchitectJobStateResponse type
		if res["status"] == "Success" {
			// Once SDK is ready, get flow id from ArchitectJobStateResponse type
			flowID = res["flow"].(map[string]interface{})["id"].(string)
			return nil
		}

		time.Sleep(15 * time.Second) // Wait 15 seconds for next retry
		return resource.RetryableError(fmt.Errorf("Job (%s) could not finish in 16 minutes and timed out ", jobId))
	})

	if retryErr != nil {
		return retryErr
	}

	if flowID == "" {
		return diag.Errorf("Failed to get the flowId from Architect Job (%s).", jobId)
	}

	d.Set("file_content_hash", hashFileContent(d.Get("filepath").(string)))
	d.SetId(flowID)

	log.Printf("Updated flow %s. ", d.Id())
	return readFlow(ctx, d, meta)
}

func deleteFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		resp, err := architectAPI.DeleteFlow(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// Flow deleted
				log.Printf("Deleted Flow %s", d.Id())
				return nil
			}
			if resp.StatusCode == http.StatusConflict {
				return resource.RetryableError(fmt.Errorf("Error deleting flow %s: %s", d.Id(), err))
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting flow %s: %s", d.Id(), err))
		}
		return nil
	})
}

// Read and upload input file path to S3 pre-signed URL
func prepareAndUploadFile(filename string, substitutions map[string]interface{}, headers map[string]interface{}, presignedUrl string) ([]byte, error) {
	bodyBuf := &bytes.Buffer{}

	reader, file, err := downloadOrOpenFile(filename)
	if err != nil {
		return nil, err
	}
	if file != nil {
		defer file.Close()
	}

	_, err = io.Copy(bodyBuf, reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to copy file content to the handler. Error: %s ", err)
	}

	if len(substitutions) > 0 {
		fileContents := bodyBuf.String()
		for k, v := range substitutions {
			fileContents = strings.Replace(fileContents, fmt.Sprintf("{{%s}}", k), v.(string), -1)
		}
		bodyBuf.Reset()
		bodyBuf.WriteString(fileContents)
	}

	req, _ := http.NewRequest("PUT", presignedUrl, bodyBuf)
	for key, value := range headers {
		req.Header.Set(key, value.(string))
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to upload flow configuration file to S3 bucket. Error: %s ", err)
	}

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body when uploading flow configuration file. %s", err)
	}

	return response, nil
}
