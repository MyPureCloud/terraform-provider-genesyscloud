package guide_version

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"log"
	"time"
)

func createGuideVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	skdConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideVersionProxy(skdConfig)
	guideId := d.Get("guide_id").(string)

	log.Printf("Creating Guide Version for Guide: %s", guideId)

	versionReq := buildGuideVersionFromResourceData(d)

	version, resp, err := proxy.createGuideVersion(ctx, versionReq, guideId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create guide version | error: %s", err), resp)
	}

	version.Id = &version.Version
	d.SetId(*version.Id)
	log.Printf("Created Guide Version: %s for Guide: %s", *version.Id, guideId)

	if d.Get("state") != nil && d.Get("state").(string) != "Draft" {
		log.Printf("Guide Version is not Draft")
		return publishGuideVersion(ctx, d, meta)
	}

	return readGuideVersion(ctx, d, meta)
}

func readGuideVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	skdConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideVersionProxy(skdConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceGuideVersion(), constants.ConsistencyChecks(), ResourceType)
	guideId := d.Get("guide_id").(string)

	log.Printf("Reading Guide Version")

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		version, resp, err := proxy.getGuideVersionById(ctx, d.Id(), guideId)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read guide version %s | Error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read guide version %s | Error: %s", d.Id(), err), resp))
		}

		_ = d.Set("instruction", version.Instruction)

		if version.Resources.DataActions != nil {
			resourcesList := flattenGuideVersionResources(version.Resources)
			_ = d.Set("resources", resourcesList)
		}

		if version.Variables != nil {
			variablesList := flattenGuideVersionVariables(version.Variables)
			_ = d.Set("variables", variablesList)
		}

		log.Printf("Read Guide Version")
		return cc.CheckState(d)
	})
}

func updateGuideVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	skdConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideVersionProxy(skdConfig)
	guideId := d.Get("guide_id").(string)

	log.Printf("Updating Guide Version %s", d.Id())

	versionReq := buildGuideVersionForUpdate(d)

	_, resp, err := proxy.updateGuideVersion(ctx, d.Id(), guideId, versionReq)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update guide version | error: %s", err), resp)
	}

	if d.Get("state") != nil && d.Get("state").(string) != "Draft" {
		log.Printf("Guide Version is not Draft")
		return publishGuideVersion(ctx, d, meta)
	}

	log.Printf("Updated Guide Version")
	return readGuideVersion(ctx, d, meta)
}

func deleteGuideVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func publishGuideVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	skdConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideVersionProxy(skdConfig)

	guideId := d.Get("guide_id").(string)
	versionId := d.Id()
	state := d.Get("state").(string)

	log.Printf("Attempting to publish Guide Version: %s for Guide: %s in State: %s", versionId, guideId, state)

	version := GuideVersionPublishJobRequest{
		GuideId:   guideId,
		VersionId: versionId,
		GuideVersion: GuideVersionPublish{
			State: state,
		},
	}

	job, resp, err := proxy.publishGuideVersion(ctx, &version)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish guide version | error: %s", err), resp)
	}

	jobId := *job.Id

	// Sleep to allow time for publish job to finish
	time.Sleep(10 * time.Second)

	jobStatus, resp, err := proxy.getGuideVersionPublishJobStatus(ctx, d.Id(), jobId, guideId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get guide version publish job status | error: %s", err), resp)
	}

	status := *jobStatus.Status

	if status == "InProgress" {
		log.Printf("Publish job for guide: %s, version: %s still in progress with status: %s", guideId, d.Id(), status)
	}

	if status == "Succeeded" {
		log.Printf("Published Guide: %s, Version: %s | Status: %s", guideId, versionId, status)
		return nil
	}

	if status == "Failed" {
		if len(jobStatus.Errors) > 0 && jobStatus.Errors[0].Message != "" {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish guide: %s, version: %s, with error: %s", guideId, versionId, jobStatus.Errors[0].Message), resp)
		}
	}
	return readGuideVersion(ctx, d, meta)
}
