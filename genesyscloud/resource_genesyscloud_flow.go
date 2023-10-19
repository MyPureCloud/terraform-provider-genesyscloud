package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/files"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func getAllFlows(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
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
			resources[*flow.Id] = &resourceExporter.ResourceMeta{Name: *flow.Name}
		}
	}

	return resources, nil
}

func FlowExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllFlows),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
		UnResolvableAttributes: map[string]*schema.Schema{
			"filepath": ResourceFlow().Schema["filepath"],
		},
		CustomFlowResolver: map[string]*resourceExporter.CustomFlowResolver{
			"file_content_hash": {ResolverFunc: resourceExporter.FileContentHashResolver},
		},
	}
}

func ResourceFlow() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Flow`,

		CreateContext: CreateWithPooledClient(createFlow),
		UpdateContext: UpdateWithPooledClient(updateFlow),
		ReadContext:   ReadWithPooledClient(readFlow),
		DeleteContext: DeleteWithPooledClient(deleteFlow),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"filepath": {
				Description:  "YAML file path for flow configuration. Note: Changing the flow name will result in the creation of a new flow with a new GUID, while the original flow will persist in your org.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: ValidatePath,
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

func readFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		flow, resp, err := architectAPI.GetFlow(d.Id(), false)
		if err != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read flow %s: %s", d.Id(), err))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read flow %s: %s", d.Id(), err))
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

func createFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating flow")
	return updateFlow(ctx, d, meta)
}

func updateFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Updating flow")

	//Check to see if we need to force and unlock on an architect flow
	if isForceUnlockEnabled(d) {
		err := forceUnlockFlow(d.Id(), sdkConfig)
		if err != nil {
			setFileContentHashToNil(d)
			return diag.Errorf("Failed to unlock targeted flow %s with error %s", d.Id(), err)
		}
	}

	flowJob, response, err := architectAPI.PostFlowsJobs()

	if err != nil {
		setFileContentHashToNil(d)
		return diag.Errorf("Failed to update job %s", err)
	}

	if err == nil && response.Error != nil {
		setFileContentHashToNil(d)
		return diag.Errorf("Failed to register job. %s", err)
	}

	presignedUrl := *flowJob.PresignedUrl
	jobId := *flowJob.Id
	headers := *flowJob.Headers

	filePath := d.Get("filepath").(string)
	substitutions := d.Get("substitutions").(map[string]interface{})

	reader, _, err := files.DownloadOrOpenFile(filePath)
	if err != nil {
		setFileContentHashToNil(d)
		return diag.Errorf(err.Error())
	}

	s3Uploader := files.NewS3Uploader(reader, nil, substitutions, headers, "PUT", presignedUrl)
	_, err = s3Uploader.Upload()
	if err != nil {
		setFileContentHashToNil(d)
		return diag.Errorf(err.Error())
	}

	// Pre-define here before entering retry function, otherwise it will be overwritten
	flowID := ""

	retryErr := WithRetries(ctx, 16*time.Minute, func() *retry.RetryError {
		flowJob, response, err := architectAPI.GetFlowsJob(jobId, []string{"messages"})
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Error retrieving job status. JobID: %s, error: %s ", jobId, response.ErrorMessage))
		}

		if *flowJob.Status == "Failure" {
			if flowJob.Messages == nil {
				return retry.NonRetryableError(fmt.Errorf("Flow publish failed. JobID: %s, no tracing messages available.", jobId))
			}
			messages := make([]string, 0)
			for _, m := range *flowJob.Messages {
				messages = append(messages, *m.Text)
			}
			return retry.NonRetryableError(fmt.Errorf("Flow publish failed. JobID: %s, tracing messages: %v ", jobId, strings.Join(messages, "\n\n")))
		}

		if *flowJob.Status == "Success" {
			flowID = *flowJob.Flow.Id
			return nil
		}

		time.Sleep(15 * time.Second) // Wait 15 seconds for next retry
		return retry.RetryableError(fmt.Errorf("Job (%s) could not finish in 16 minutes and timed out ", jobId))
	})

	if retryErr != nil {
		setFileContentHashToNil(d)
		return retryErr
	}

	if flowID == "" {
		setFileContentHashToNil(d)
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

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		resp, err := architectAPI.DeleteFlow(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Flow deleted
				log.Printf("Deleted Flow %s", d.Id())
				return nil
			}
			if resp.StatusCode == http.StatusConflict {
				return retry.RetryableError(fmt.Errorf("Error deleting flow %s: %s", d.Id(), err))
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting flow %s: %s", d.Id(), err))
		}
		return nil
	})
}

func GenerateFlowResource(resourceID, srcFile, filecontent string, force_unlock bool, substitutions ...string) string {
	fullyQualifiedPath, _ := filepath.Abs(srcFile)

	if filecontent != "" {
		updateFile(srcFile, filecontent)
	}

	flowResourceStr := fmt.Sprintf(`resource "genesyscloud_flow" "%s" {
        filepath = %s
		file_content_hash =  filesha256(%s)
		force_unlock = %v
		%s
	}
	`, resourceID, strconv.Quote(srcFile), strconv.Quote(fullyQualifiedPath), force_unlock, strings.Join(substitutions, "\n"))

	return flowResourceStr
}

func updateFile(filepath, content string) {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	file.WriteString(content)
}

// setFileContentHashToNil This operation is required after a flow update fails because we want Terraform to detect changes
// in the file content hash and re-attempt an update, should the user re-run terraform apply without making changes to the file contents
func setFileContentHashToNil(d *schema.ResourceData) {
	_ = d.Set("file_content_hash", nil)
}
