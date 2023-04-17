package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v98/platformclientv2"
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
				Description:  "YAML file path for flow configuration.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validatePath,
			},
			"file_content_hash": {
				Description: "Hash value of the YAML file content. Used to detect changes.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"substitutions": {
				Description: "A substitution is a key value pair where the key is the value you want to replace, and the value is the value to substitute in its place.",
				Type:        schema.TypeMap,
				Optional:    true,
			},
			"force_unlock": {
				Description: `Will perform a force unlock on an architect flow before beginning the publication process.  NOTE: The force unlock publishes the 'draft'
				              architect flow and then publishes the flow named in this resource. This mirrors the behavior found in the archy CLI tool.`,
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func createFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating flow")
	return updateFlow(ctx, d, meta)
}

func readFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

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

func forceUnlockFlow(flowId string, sdkConfig *platformclientv2.Configuration) error {
	log.Printf("Attempting to perform an unlock on flow: %s", flowId)
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)
	_, _, err := architectAPI.PostFlowsActionsUnlock(flowId)

	if err != nil {
		return err
	}
	return nil
}

func isForceUnlockEnabled(d *schema.ResourceData) bool {
	forceUnlock := d.Get("force_unlock").(bool)
	log.Printf("ForceUnlock: %v, id %v", forceUnlock, d.Id())

	if forceUnlock && d.Id() != "" {
		return true
	}
	return false
}

func updateFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	//Check to see if we need to force and unlock on an architect flow
	if isForceUnlockEnabled(d) {
		err := forceUnlockFlow(d.Id(), sdkConfig)
		if err != nil {
			return diag.Errorf("Failed to unlock targeted flow %s with error %s", d.Id(), err)
		}
	}

	flowJob, response, err := architectAPI.PostFlowsJobs()

	if err != nil {
		return diag.Errorf("Failed to update job %s", err)
	}

	if err == nil && response.Error != nil {
		return diag.Errorf("Failed to register job. %s", err)
	}

	presignedUrl := *flowJob.PresignedUrl
	jobId := *flowJob.Id
	headers := *flowJob.Headers

	filePath := d.Get("filepath").(string)
	substitutions := d.Get("substitutions").(map[string]interface{})

	reader, _, err := downloadOrOpenFile(filePath)
	if err != nil {
		return diag.Errorf(err.Error())
	}

	s3Uploader := NewS3Uploader(reader, substitutions, headers, presignedUrl)
	_, err = s3Uploader.Upload()
	if err != nil {
		return diag.Errorf(err.Error())
	}

	// Pre-define here before entering retry function, otherwise it will be overwritten
	flowID := ""

	retryErr := withRetries(ctx, 16*time.Minute, func() *resource.RetryError {
		flowJob, response, err := architectAPI.GetFlowsJob(jobId, []string{"messages"})
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error retrieving job status. JobID: %s, error: %s ", jobId, response.ErrorMessage))
		}

		if *flowJob.Status == "Failure" {
			if flowJob.Messages == nil {
				return resource.NonRetryableError(fmt.Errorf("Flow publish failed. JobID: %s, no tracing messages available.", jobId))
			}
			messages := make([]string, 0)
			for _, m := range *flowJob.Messages {
				messages = append(messages, *m.Text)
			}
			return resource.NonRetryableError(fmt.Errorf("Flow publish failed. JobID: %s, tracing messages: %v ", jobId, strings.Join(messages, "\n\n")))
		}

		if *flowJob.Status == "Success" {
			flowID = *flowJob.Flow.Id
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

	d.SetId(flowID)

	log.Printf("Updated flow %s. ", d.Id())
	return readFlow(ctx, d, meta)
}

func deleteFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	//Check to see if we need to force
	if isForceUnlockEnabled(d) {
		err := forceUnlockFlow(d.Id(), sdkConfig)
		if err != nil {
			return diag.Errorf("Failed to unlock targeted flow %s with error %s", d.Id(), err)
		}
	}

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
