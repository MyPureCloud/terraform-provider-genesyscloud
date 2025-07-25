package architect_flow

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func getAllFlows(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	p := newArchitectFlowProxy(clientConfig)

	flows, resp, err := p.GetAllFlows(ctx, "", nil)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get architect flows %v", err), resp)
	}

	for _, flow := range *flows {

		blockHash, err := util.QuickHashFields(flow.VarType)
		if err != nil {
			return nil, diag.Errorf("failed to generate quick hash for flow %s: %v", *flow.Id, err)
		}

		//DEVTOOLING-393:  Putting this in here to deal with the situation where Cesar's BCP app is reliant on the naming structure
		//This should be removed once the CX as Code architect export process is complete and will export files with the type in the name.
		overrideBCPNaming := os.Getenv("OVERRIDE_BCP_NAMING")

		if overrideBCPNaming != "" {
			resources[*flow.Id] = &resourceExporter.ResourceMeta{BlockLabel: *flow.Name, BlockHash: blockHash}
			continue
		}

		//This is our go forward naming standard for flows.
		resources[*flow.Id] = &resourceExporter.ResourceMeta{BlockLabel: *flow.VarType + "_" + *flow.Name, BlockHash: blockHash}
	}

	return resources, nil
}

func readFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig

	proxy := newArchitectFlowProxy(sdkConfig)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		flow, resp, err := proxy.GetFlow(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read flow %s: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read flow %s: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "name", flow.Name)
		resourcedata.SetNillableValue(d, "type", flow.VarType)

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
		resp, err := p.ForceUnlockFlow(ctx, d.Id())
		if err != nil {
			setFileContentHashToNil(d)
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to unlock targeted flow %s with error %s", d.Id(), err), resp)
		}
	}

	flowJob, response, err := p.CreateFlowsDeployJob(ctx)

	if err != nil || response.Error != nil {
		var errorString string
		if err != nil {
			errorString = err.Error()
		} else {
			errorString = response.ErrorMessage
		}
		setFileContentHashToNil(d)
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to register job %s", errorString), response)
	}

	presignedUrl := *flowJob.PresignedUrl
	jobId := *flowJob.Id
	headers := *flowJob.Headers

	filePath := d.Get("filepath").(string)
	substitutions := d.Get("substitutions").(map[string]interface{})

	reader, _, err := files.DownloadOrOpenFile(filePath)
	if err != nil {
		setFileContentHashToNil(d)
		return diag.FromErr(err)
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
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error retrieving job status. JobID: %s, error: %s ", jobId, err), response))
		}

		if *flowJob.Status == "Failure" {
			if flowJob.Messages == nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("flow publish failed. JobID: %s, no tracing messages available", jobId), response))
			}
			messages := make([]string, 0)
			for _, m := range *flowJob.Messages {
				messages = append(messages, *m.Text)
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("flow publish failed. JobID: %s, tracing messages: %v ", jobId, strings.Join(messages, "\n\n")), response))
		}

		if *flowJob.Status == "Success" {
			flowID = *flowJob.Flow.Id
			return nil
		}

		time.Sleep(15 * time.Second) // Wait 15 seconds for next retry
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Job (%s) could not finish in 16 minutes and timed out ", jobId), response))
	})

	if retryErr != nil {
		setFileContentHashToNil(d)
		return retryErr
	}

	if flowID == "" {
		setFileContentHashToNil(d)
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to get the flowId from Architect Job (%s).", jobId), fmt.Errorf("FlowID is nil"))
	}

	d.SetId(flowID)

	log.Printf("Updated flow %s. ", d.Id())
	return readFlow(ctx, d, meta)
}

func deleteFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	p := getArchitectFlowProxy(sdkConfig)

	log.Printf("Deleting flow %s", d.Id())

	//Check to see if we need to force
	if isForceUnlockEnabled(d) {
		resp, err := p.ForceUnlockFlow(ctx, d.Id())
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to unlock targeted flow %s with error %v", d.Id(), err), resp)
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
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting flow %s | error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting flow %s | error: %s", d.Id(), err), resp))
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
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Printf("failed to close file %s: %v", filepath, err)
		}
	}(file)

	_, _ = file.WriteString(content)
}
