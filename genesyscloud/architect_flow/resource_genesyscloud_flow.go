package architect_flow

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/files"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
)

func getAllFlows(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	p := getArchitectFlowProxy(clientConfig)

	flows, err := p.GetAllFlows(ctx)
	if err != nil {
		return nil, diag.Errorf("failed to get architect flows %v", err)
	}

	for _, flow := range *flows {

		//DEVTOOLING-393:  Putting this in here to deal with the situation where Cesar's BCP app is reliant on the naming structure
		//This should be removed once the CX as Code architect export process is complete and will export files with the type in the name.
		overrideBCPNaming := os.Getenv("OVERRIDE_BCP_NAMING")

		if overrideBCPNaming != "" {
			resources[*flow.Id] = &resourceExporter.ResourceMeta{Name: *flow.Name}
			continue
		}

		//This is our go forward naming standard for flows.
		resources[*flow.Id] = &resourceExporter.ResourceMeta{Name: *flow.VarType + "_" + *flow.Name}
	}

	return resources, nil
}

func readFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig

	proxy := getArchitectFlowProxy(sdkConfig)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		flow, resp, err := proxy.GetFlow(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read flow %s: %s", d.Id(), err))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read flow %s: %s", d.Id(), err))
		}

		log.Printf("Read flow %s %s", d.Id(), *flow.Name)
		return nil
	})
}

func createFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating flow")
	return updateFlow(ctx, d, meta)
}

func updateFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	p := getArchitectFlowProxy(sdkConfig)

	log.Printf("Updating flow")

	//Check to see if we need to force and unlock on an architect flow
	if isForceUnlockEnabled(d) {
		err := p.ForceUnlockFlow(ctx, d.Id())
		if err != nil {
			setFileContentHashToNil(d)
			return diag.Errorf("Failed to unlock targeted flow %s with error %s", d.Id(), err)
		}
	}

	flowJob, response, err := p.CreateFlowsDeployJob(ctx)

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

	_, uploadErr := s3Uploader.UploadWithRetries(ctx, filePath, 20*time.Second)
	if uploadErr != nil {
		setFileContentHashToNil(d)
		return diag.FromErr(uploadErr)
	}

	// Pre-define here before entering retry function, otherwise it will be overwritten
	flowID := ""

	retryErr := util.WithRetries(ctx, 16*time.Minute, func() *retry.RetryError {
		flowJob, response, err := p.GetFlowsDeployJob(ctx, jobId)
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
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	p := getArchitectFlowProxy(sdkConfig)

	//Check to see if we need to force
	if isForceUnlockEnabled(d) {
		err := p.ForceUnlockFlow(ctx, d.Id())
		if err != nil {
			return diag.Errorf("Failed to unlock targeted flow %s with error %s", d.Id(), err)
		}
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		resp, err := p.DeleteFlow(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
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

func updateFile(filepath, content string) {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	file.WriteString(content)
}
