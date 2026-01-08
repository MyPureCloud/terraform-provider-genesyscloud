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
	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
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

func readFlow(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig

	proxy := newArchitectFlowProxy(sdkConfig)

	log.Printf("Reading flow  %s", d.Id())

	// Set resource context for SDK debug logging before entering retry loop
	ctx = util.SetResourceContext(ctx, d, ResourceType)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		flow, resp, err := proxy.GetFlow(ctx, d.Id())
		if err != nil {
			diagErr := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read flow %s: %s", d.Id(), err), resp)
			if util.IsStatus404(resp) {
				return retry.RetryableError(diagErr)
			}
			return retry.NonRetryableError(diagErr)
		}

		resourcedata.SetNillableValue(d, "name", flow.Name)
		resourcedata.SetNillableValue(d, "type", flow.VarType)

		log.Printf("Read flow %s, %s", d.Id(), *flow.Name)
		return nil
	})
}

func createFlow(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("Creating flow")
	return updateFlow(ctx, d, meta)
}

func updateFlow(ctx context.Context, d *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	p := getArchitectFlowProxy(sdkConfig)

	var flowName string

	if nameInterface := d.Get("name"); nameInterface != nil {
		if name, ok := nameInterface.(string); ok && name != "" {
			flowName = name
		} else if !ok {
			log.Printf("Warning: 'name' attribute is not a string, got type: %T, using empty name", nameInterface)
			flowName = ""
		}
	} else {
		log.Printf("Info: 'name' attribute is nil, using empty name")
		flowName = ""
	}

	// If name is not in ResourceData but we have a flow ID, try to get it from the flow
	if flowName == "" && d.Id() != "" {
		flow, _, err := p.GetFlow(ctx, d.Id())
		if err == nil && flow != nil && flow.Name != nil && *flow.Name != "" {
			flowName = *flow.Name
			log.Printf("Retrieved flow name '%s' from API for flow ID %s", flowName, d.Id())
		}
	}

	// If still no name, set to "unavailable" so the resource context preserves it
	if flowName == "" {
		flowName = "unavailable"
	}

	log.Printf("Updating flow  %s, %s", flowName, d.Id())

	// Set resource context for SDK debug logging with flow name
	// Use the existing flow ID if available, otherwise it will be updated once we get it from the job
	flowIdForContext := d.Id()
	if flowIdForContext == "" {
		flowIdForContext = "unavailable"
	}
	ctx = provider.WithResourceContext(ctx, ResourceType, flowIdForContext, flowName)

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
	substitutions := d.Get("substitutions").(map[string]any)

	reader, _, err := files.DownloadOrOpenFile(ctx, filePath, S3Enabled)
	if err != nil {
		setFileContentHashToNil(d)
		return append(diags, diag.FromErr(err)...)
	}

	log.Printf("Uploading flow  %s, %s, %s", flowName, d.Id(), jobId)

	s3Uploader := files.NewS3Uploader(reader, nil, substitutions, headers, "PUT", presignedUrl)

	_, uploadErr := s3Uploader.UploadWithRetries(ctx, filePath, 20*time.Second)
	if uploadErr != nil {
		setFileContentHashToNil(d)
		return append(diags, diag.FromErr(uploadErr)...)
	}

	log.Printf("Uploaded flow %s, %s, %s", flowName, d.Id(), jobId)

	// Pre-define here before entering retry function, otherwise it will be overwritten
	flowID := ""

	retryErr := util.WithRetries(ctx, 16*time.Minute, func() *retry.RetryError {
		flowJob, response, err := p.GetFlowsDeployJob(ctx, jobId)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error retrieving job status. JobID: %s, flowName: %s, error: %s", jobId, flowName, err), response))
		}

		// Update resource context with flow ID if we get it from the job response
		// This ensures subsequent API calls have the correct flow ID in logs
		if flowJob.Flow != nil && flowJob.Flow.Id != nil {
			flowID = *flowJob.Flow.Id
			// Update context with the actual flow ID
			ctx = provider.WithResourceContext(ctx, ResourceType, flowID, flowName)
		}

		if flowJob.Status != nil {
			log.Printf("Job status for flow %s, jobId %s: %s", flowName, jobId, *flowJob.Status)
		} else {
			log.Printf("Job status for flow %s, jobId %s: <nil status>", flowName, jobId)
		}

		if flowJob.Status != nil && *flowJob.Status == "Failure" {
			log.Printf("Failed to Get flow %s, %s, %s", flowName, d.Id(), jobId)
			if flowJob.Messages == nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("flow publish failed. JobID: %s, flowName: %s,  no tracing messages available", jobId, flowName), response))
			}
			messages := make([]string, 0)
			for _, m := range *flowJob.Messages {
				if m.Text != nil {
					log.Printf("API Message for flow %s, jobId %s: %s", flowName, jobId, *m.Text)
					messages = append(messages, *m.Text)
				} else {
					log.Printf("API Message for flow %s, jobId %s: <nil message>", flowName, jobId)
				}
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("flow publish failed. JobID: %s, flowName: %s, tracing messages: %v ", jobId, flowName, strings.Join(messages, "\n\n")), response))
		}

		if flowJob.Status != nil && *flowJob.Status == "Success" {
			log.Printf("Success for flow %s, %s", flowName, jobId)
			if flowJob.Flow != nil && flowJob.Flow.Id != nil {
				flowID = *flowJob.Flow.Id
				// Context was already updated above when we extracted flowID
			} else {
				log.Printf("Warning: Flow or Flow.Id is nil for successful job %s", jobId)
			}
			return nil
		}

		time.Sleep(15 * time.Second) // Wait 15 seconds for next retry
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Job (%s) could not finish in 16 minutes and timed out ", jobId), response))
	})

	if retryErr != nil {
		setFileContentHashToNil(d)
		return append(diags, retryErr...)
	}

	if flowID == "" {
		setFileContentHashToNil(d)
		return append(diags, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to get the flowId from Architect Job (%s).", jobId), fmt.Errorf("FlowID is nil"))...)
	}

	filePathHash, err := files.HashFileContent(ctx, filePath, S3Enabled)
	if err != nil {
		return append(diags, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to get the file content hash for the flow %s", flowID), err)...)
	}
	_ = d.Set("file_content_hash", filePathHash)

	d.SetId(flowID)

	log.Printf("Updated flow %s, %s", flowName, d.Id())
	return append(diags, readFlow(ctx, d, meta)...)
}

func deleteFlow(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	// Set resource context for SDK debug logging before entering retry loop
	ctx = util.SetResourceContext(ctx, d, ResourceType)

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
