package genesyscloud

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v48/platformclientv2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func getAllFlows(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	architectAPI := platformclientv2.NewArchitectApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		flows, _, err := architectAPI.GetFlows(nil, pageNum, 25, "", "", nil, "", "", "", "", "", "", "", "", false, true, "", "", nil)
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

func architectFlowExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllFlows),
		RefAttrs:         map[string]*RefAttrSettings{},
	}
}

func resourceFlow() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Flow",

		CreateContext: createWithPooledClient(createFlow),
		ReadContext:   readWithPooledClient(readFlow),
		UpdateContext: updateWithPooledClient(updateFlow),
		DeleteContext: deleteWithPooledClient(deleteFlow),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"description": {
				Description: "Description for the flow. This won't affect the flow at all. Configuration of the flow should be in the provided file path. ",
				Type:        schema.TypeString,
				Required:    true,
			},
			"filepath": {
				Description: "YAML file path for flow configuration. ",
				Type:        schema.TypeString,
				Required:    true,
				StateFunc: func(v interface{}) string {
					return hashFileContent(v.(string))
				},
			},
			"debug": {
				Description: "Boolean value for debug mode. ",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"force_unlock": {
				Description: "Whether to force unlocking the flow. ",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"recreate": {
				Description: "Whether to recreate the flow. ",
				Type:        schema.TypeBool,
				Required:    true,
			},
		},
	}
}

func createFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	sdkConfig := meta.(*providerMeta).ClientConfig
	orgsAPI := platformclientv2.NewOrganizationApiWithConfig(sdkConfig)

	org, _, err := orgsAPI.GetOrganizationsMe()
	if err != nil {
		return diag.Errorf("Failed to get organization.", err)
	}

	orgID := org.Id

	//Todo: Call architect/archy/jobs to register
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

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

	var successPayload *IntegrationAction
	response, err := apiClient.CallAPI(path, "POST", body, headerParams, nil, nil, "", nil)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if err == nil && response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		err = json.Unmarshal([]byte(response.RawBody), &successPayload)
	}
	return successPayload, response, err

	//TODO: parse response, get jobId and presigned url
	presignedUrl := "asdf"
	jobId := "asdf"
	correlationId := "asdf"
	headers := make(map[string]interface{})

	filePath := d.Get("filepath").(string)

	_, err = prepareAndUploadFile(filePath, headers, presignedUrl, jobId, *orgID, correlationId)

	if err != nil {
		return diag.Errorf(err.Error())
	}
	//
	//var result map[string]interface{}
	//
	//err = json.Unmarshal(body, &result)
	//if err != nil {
	//	return diag.Errorf("Failed to unmarshal response body when creating flow. %s", err)
	//}
	//
	//if result["statusCode"] == nil || int(result["statusCode"].(float64)) != http.StatusOK {
	//	return diag.Errorf("Create flow Request failed. Result: %s", string(body))
	//}

	//flowID := getFlowID((result["body"]).(map[string]interface{})["stdout"].(string))

	retryErr := withRetries(ctx, 16*time.Minute, func() *resource.RetryError {
		body, resp, err := architectAPI.getStatus()
		if body.status == "ERRORED" {
			return resource.NonRetryableError(fmt.Errorf("Error occurred publishing the flow. JobID: %s, error code: %s", jobId, body.error_code))
		}

		if body.status == "COMPLETED" {
			// TODO: get ID from archy result from getStatus endpoint
			d.SetId(*ID)
			return nil
		}

		time.Sleep(15 * time.Second) // Wait 15 seconds for next retry
		return resource.RetryableError(fmt.Errorf("Job is still in progress %s", jobId))
	})

	if retryErr != nil {
		return retryErr
	}

	log.Printf("Created flow %s. ", d.Id())
	return readFlow(ctx, d, meta)
}

func readFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	flow, resp, err := architectAPI.GetFlow(d.Id(), false)
	if err != nil {
		if isStatus404(resp) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read flow %s: %s", d.Id(), err)
	}

	description := fmt.Sprintf("Flow name: %s, Flow type: %s", *flow.Name, *flow.VarType)

	d.Set("description", description)

	log.Printf("Read flow %s %s", d.Id(), *flow.Name)
	return nil
}

func updateFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clientID := os.Getenv("GENESYSCLOUD_OAUTHCLIENT_ID")
	clientSecret := os.Getenv("GENESYSCLOUD_OAUTHCLIENT_SECRET")
	location := getLocationForArchy(os.Getenv("GENESYSCLOUD_REGION"))
	debug := strconv.FormatBool(d.Get("debug").(bool))
	forceUnlock := strconv.FormatBool(d.Get("force_unlock").(bool))
	recreate := strconv.FormatBool(d.Get("recreate").(bool))
	filepath := d.Get("filepath").(string)

	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return diag.Errorf("Failed to read configuration file from this path: %s. %s", filepath, err)
	}
	filecontent := string(file)

	requestBody, _ := json.Marshal(map[string]interface{}{
		"command": "update",
		"files": map[string]interface{}{
			"main.yaml": filecontent,
		},
		"settings": map[string]interface{}{
			"clientSecret": &clientSecret,
			"clientId":     &clientID,
			"debug":        &debug,
			"forceUnlock":  &forceUnlock,
			"location":     &location,
			"recreate":     &recreate,
		},
	})

	body, err := sendRequest(requestBody)

	if err != nil {
		return diag.Errorf(err.Error())
	}

	var result map[string]interface{}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return diag.Errorf("Failed to unmarshal response body when updating flow. %s", err)
	}
	if result["statusCode"] == nil || int(result["statusCode"].(float64)) != http.StatusOK {
		return diag.Errorf("Update flow Request failed. Result: %s", string(body))
	}

	flowID := getFlowID((result["body"]).(map[string]interface{})["stdout"].(string))

	d.SetId(flowID)

	log.Printf("Updated flow %s", d.Id())
	return readFlow(ctx, d, meta)
}

func deleteFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	_, err := architectAPI.DeleteFlow(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete the flow %s: %s", d.Id(), err)
	}
	log.Printf("Deleted flow %s", d.Id())
	return nil
}

func hashFileContent(path string) string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Fatal(err)
	}

	return hex.EncodeToString(hash.Sum(nil))
}

func prepareAndUploadFile(filename string, headers map[string]interface{}, presignedUrl string, jobId string, orgId string, correlationId string) ([]byte, error) {

	file, err := os.Open(filename)

	if err != nil {
		return nil, fmt.Errorf("Failed to open file %s . Error: %s ", filename, err)
	}

	defer file.Close()

	req, _ := http.NewRequest("PUT", presignedUrl, file)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-amz-meta-organizationid", orgId)
	req.Header.Set("x-amz-meta-correlationid", correlationId)
	req.Header.Set("x-amz-meta-jobid", jobId)

	for key, value := range headers {
		req.Header.Set(key, value.(string))
	}

	//TODO: Set headers using presignedUrl object returned by public api endpoint

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to upload flow configuration file to S3 bucket. Error: %s ", err)
	}

	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body when uploading flow configuration file. %s", err)
	}

	return response, nil
}

func getFlowID(response string) string {
	regex := regexp.MustCompile(`Flow Id: (\b[a-fA-F0-9]{8}\b-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-\b[a-fA-F0-9]{12}\b)`)
	regexResult := regex.FindAllStringSubmatch(response, -1)
	var flowID string

	if len(regexResult) > 0 {
		flowID = regexResult[0][1]
	}

	return flowID
}

func getLocationForArchy(region string) string {
	switch region {
	// TODO: Can add more regions here when Archy support them
	case "":
		return ""
	default:
		return "dev"
	}
}
