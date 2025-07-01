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
	if version.Id != nil {
		d.SetId(guideId + "/" + *version.Id)
	}

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

	guideId, versionId, err := parseId(d.Id())
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to Parse Guide id"), err)
	}

	log.Printf("Reading Guide Version for guide: %s", guideId)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		version, resp, err := proxy.getGuideVersionById(ctx, versionId, guideId)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read guide version %s | Error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read guide version %s | Error: %s", d.Id(), err), resp))
		}

		_ = d.Set("guide_id", version.Guide.Id)
		_ = d.Set("instruction", version.Instruction)

		if version.State != "" {
			_ = d.Set("state", version.State)
		}

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

	guideId, versionId, err := parseId(d.Id())
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to Parse Guide id"), err)
	}

	log.Printf("Updating Guide Version %s", d.Id())

	versionReq := buildGuideVersionForUpdate(d)

	version, resp, err := proxy.updateGuideVersion(ctx, versionId, guideId, versionReq)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update guide version | error: %s", err), resp)
	}

	version.Id = &version.Version
	if version.Id != nil {
		d.SetId(guideId + "/" + *version.Id)
	}

	_ = d.Set("guide_id", version.Guide.Id)

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
	state := d.Get("state").(string)

	guideId, versionId, err := parseId(d.Id())
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to Parse Guide id"), err)
	}

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

	jobStatus, resp, err := proxy.getGuideVersionPublishJobStatus(ctx, versionId, jobId, guideId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get guide version publish job status | error: %s", err), resp)
	}

	status := *jobStatus.Status

	switch status {
	case "InProgress":
		log.Printf("Publish job for guide: %s, version: %s still in progress with status: %s", guideId, d.Id(), status)
	case "Succeeded":
		log.Printf("Published successfully")
		return readGuideVersion(ctx, d, meta)
	case "Failed":
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish guide: %s, version: %s, with error: %s", guideId, versionId, jobStatus.Errors[0].Message), resp)
	default:
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Unknown job status: %s", status), nil)
	}

	return readGuideVersion(ctx, d, meta)
}
