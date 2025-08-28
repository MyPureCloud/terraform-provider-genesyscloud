package guide_version

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

func getAllGuideVersions(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getGuideVersionProxy(clientConfig)

	log.Printf("Retrieving all Guides")

	guides, resp, err := proxy.GetAllGuides(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get all guides | error: %s", err), resp)
	}

	if guides == nil {
		log.Printf("No guides found")
		return resources, nil
	}

	for _, guide := range *guides {
		// For guide versions, we need both guide ID and version ID
		// Use the latest saved version if available, otherwise skip this guide
		if guide.LatestSavedVersion == nil || guide.LatestSavedVersion.Version == nil {
			log.Printf("Skipping guide %s - no latest saved version available", *guide.Id)
			continue
		}

		resourceId := *guide.Id + "/" + *guide.LatestSavedVersion.Version
		resources[resourceId] = &resourceExporter.ResourceMeta{
			BlockLabel: *guide.Name,
		}
	}

	return resources, nil
}

func createGuideVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	skdConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideVersionProxy(skdConfig)
	guideId := d.Get("guide_id").(string)

	log.Printf("Creating Guide Version for Guide: %s", guideId)

	versionReq := buildGuideVersionFromResourceData(d)

	version, resp, err := proxy.createGuideVersion(ctx, versionReq, guideId)
	if err != nil {
		// if the error is about the latest version not being in Production Ready state, we try to update the existing version
		if strings.Contains(err.Error(), "Latest version is not in Production Ready state") {
			log.Printf("Latest version is not in Production Ready state, trying to update the existing version ")
			return updateGuideVersion(ctx, d, meta)
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create guide version: %s", err), resp)
	}

	version.Id = &version.Version
	d.SetId(guideId + "/" + version.Version)

	log.Printf("Created Guide Version: %s for Guide: %s", version.Version, guideId)

	publishErr := publishGuideVersion(ctx, d, meta)
	if publishErr != nil {
		return publishErr
	}

	return readGuideVersion(ctx, d, meta)
}

func readGuideVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	skdConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideVersionProxy(skdConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceGuideVersion(), constants.ConsistencyChecks(), ResourceType)

	guideId, versionId, err := parseId(d.Id())
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "Failed to parse guide id", err)
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

		resourcedata.SetNillableValue(d, "guide_id", version.Guide.Id)
		resourcedata.SetNillableValue(d, "instruction", &version.Instruction)

		if len(version.Resources.DataActions) > 0 {
			resourcesList := flattenGuideVersionResources(version.Resources)
			_ = d.Set("resources", resourcesList)
		}

		if version.Variables != nil {
			variablesList := flattenGuideVersionVariables(version.Variables)
			_ = d.Set("variables", variablesList)
		}

		guide, resp, err := proxy.getGuideById(ctx, guideId)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to get guide %s: %v", guideId, err), resp))
		}

		log.Printf("Read Guide Version %s for guide %s", d.Id(), *guide.Name)
		return cc.CheckState(d)
	})
}

func updateGuideVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Since published versions are immutable, we must create a new version
	skdConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideVersionProxy(skdConfig)
	guideId := d.Get("guide_id").(string)

	var version *VersionResponse
	guide, resp, err := proxy.getGuideById(ctx, guideId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get guide: %s", err), resp)
	}

	log.Printf("Updating Guide Version %s", d.Id())

	// Check if there's a production ready version
	// If there is, we need to create a new version. If not, we can update the existing one.
	if guide.LatestProductionReadyVersion != nil && guide.LatestProductionReadyVersion.Version != nil {
		// There's a published version, create a new version
		versionReq := buildGuideVersionFromResourceData(d)
		version, resp, err = proxy.createGuideVersion(ctx, versionReq, guideId)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create guide version: %s", err), resp)
		}
	} else {
		// No published version, update the existing version
		if guide.LatestSavedVersion == nil || guide.LatestSavedVersion.Version == nil {
			return util.BuildDiagnosticError(ResourceType, "No latest saved version found to update", fmt.Errorf("guide has no latest saved version"))
		}

		versionReq := buildGuideVersionForUpdate(d)
		version, resp, err = proxy.updateGuideVersion(ctx, *guide.LatestSavedVersion.Version, guideId, versionReq)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update guide version: %s", err), resp)
		}
	}

	d.SetId(guideId + "/" + version.Version)

	publishErr := publishGuideVersion(ctx, d, meta)
	if publishErr != nil {
		return publishErr
	}

	log.Printf("Updated Guide Version")
	return readGuideVersion(ctx, d, meta)
}

func publishGuideVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	skdConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideVersionProxy(skdConfig)
	state := "ProductionReady"

	guideId, versionId, err := parseId(d.Id())
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "Failed to parse guide id", err)
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

func deleteGuideVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil // No delete operation for guide versions, return nil
}
