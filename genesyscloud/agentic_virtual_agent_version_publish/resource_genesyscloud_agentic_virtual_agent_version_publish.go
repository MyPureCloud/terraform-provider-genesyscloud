package agentic_virtual_agent_version_publish

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
   resource_genesyscloud_agentic_virtual_agent_version_publish.go contains the CRUD logic for the
   publish resource.

   Key behaviors:
   - Create: POST a publish job → poll until Succeeded or Failed
   - Read: GET the version and check its status matches the published state
   - Delete: No-op (state removal only — no unpublish API exists)
   - All fields are ForceNew — any change requires a new publish
*/

// Composite ID format: agentId/versionId/status
func buildPublishId(agentId, versionId, status string) string {
	return agentId + "/" + versionId + "/" + status
}

func parsePublishId(id string) (agentId, versionId, status string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid publish resource ID format: %s (expected agentId/versionId/status)", id)
	}
	return parts[0], parts[1], parts[2], nil
}

// createPublish creates a publish job and polls until it completes.
func createPublish(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getPublishProxy(sdkConfig)

	agentId := d.Get("agent_id").(string)
	versionId := d.Get("version").(string)
	status := d.Get("status").(string)

	publishReq := &PublishJobRequest{
		VirtualAgentVersion: &PublishJobVersionStatus{
			Status: status,
		},
	}

	log.Printf("Publishing Agentic Virtual Agent Version %s/%s to %s", agentId, versionId, status)

	jobResp, resp, err := proxy.createPublishJob(ctx, agentId, versionId, publishReq)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create publish job for version %s/%s: %s", agentId, versionId, err), resp)
	}

	jobId := jobResp.Id
	log.Printf("Publish job created: %s", jobId)

	// Poll until the job completes
	pollErr := util.WithRetries(ctx, 2*time.Minute, func() *retry.RetryError {
		jobStatus, jobResp, jobErr := proxy.getPublishJobStatus(ctx, agentId, versionId, jobId)
		if jobErr != nil {
			if util.IsStatus404(jobResp) {
				return retry.NonRetryableError(fmt.Errorf("publish job %s not found", jobId))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error polling publish job %s: %s", jobId, jobErr), jobResp))
		}

		switch jobStatus.Status {
		case "InProgress":
			return retry.RetryableError(fmt.Errorf("publish job %s still in progress", jobId))
		case "Succeeded":
			log.Printf("Publish job %s succeeded. Version %s/%s is now %s", jobId, agentId, versionId, status)
			return nil
		case "Failed":
			errMsg := "publish job failed"
			if len(jobStatus.Errors) > 0 {
				errMsg = fmt.Sprintf("publish job failed: %s (code: %s)", jobStatus.Errors[0].Message, jobStatus.Errors[0].Code)
			}
			return retry.NonRetryableError(fmt.Errorf("%s for version %s/%s: %s", errMsg, agentId, versionId, errMsg))
		default:
			return retry.RetryableError(fmt.Errorf("unexpected publish job status: %s", jobStatus.Status))
		}
	})

	if pollErr != nil {
		return pollErr
	}

	d.SetId(buildPublishId(agentId, versionId, status))
	return readPublish(ctx, d, meta)
}

// readPublish reads the version's current status to verify the publish is still valid.
func readPublish(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getPublishProxy(sdkConfig)

	agentId, versionId, expectedStatus, err := parsePublishId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("Reading publish state for version %s/%s (expected: %s)", agentId, versionId, expectedStatus)

	versionResp, resp, getErr := proxy.getVersionStatus(ctx, agentId, versionId)
	if getErr != nil {
		if util.IsStatus404(resp) {
			log.Printf("Version %s/%s not found, removing publish from state", agentId, versionId)
			d.SetId("")
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read version %s/%s: %s", agentId, versionId, getErr), resp)
	}

	currentStatus := ""
	if versionResp.Status != nil {
		currentStatus = *versionResp.Status
	}

	// Check if the version is still in the expected published state
	// ProductionReady is terminal — it never changes
	// TestReady can be reset to Draft via PATCH — detect that as drift
	if expectedStatus == "ProductionReady" && currentStatus == "ProductionReady" {
		// Still valid
	} else if expectedStatus == "TestReady" && (currentStatus == "TestReady" || currentStatus == "ProductionReady") {
		// TestReady is valid, ProductionReady is a higher state (not drift)
	} else if currentStatus == "Draft" {
		// Version was patched back to Draft — publish is no longer valid
		log.Printf("Version %s/%s has been reset to Draft. Removing publish from state.", agentId, versionId)
		d.SetId("")
		return nil
	}

	_ = d.Set("agent_id", agentId)
	_ = d.Set("version", versionId)
	_ = d.Set("status", expectedStatus)

	return nil
}

// deletePublish is a no-op — there is no unpublish API.
func deletePublish(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Removing publish resource from state: %s (no unpublish API exists)", d.Id())
	return nil
}
