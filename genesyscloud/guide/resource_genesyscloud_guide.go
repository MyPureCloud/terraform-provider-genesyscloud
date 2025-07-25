package guide

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

func getAllGuides(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getGuideProxy(clientConfig)

	log.Printf("Retrieving all Guides")

	guides, resp, err := proxy.getAllGuides(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get all guides | error: %s", err), resp)
	}

	if guides == nil {
		return resources, nil
	}

	for _, guide := range *guides {
		resources[*guide.Id] = &resourceExporter.ResourceMeta{BlockLabel: *guide.Name}
	}

	log.Printf("Successfully retrieved all Guides")
	return resources, nil
}

func createGuide(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideProxy(sdkConfig)

	// Required Attributes
	name := d.Get("name").(string)
	source := d.Get("source").(string)

	log.Printf("Creating Guide: %s", name)

	guideReq := &CreateGuide{
		Name:   &name,
		Source: &source,
	}

	guide, resp, err := proxy.createGuide(ctx, guideReq)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create guide: %s | error: %s", name, err), resp)
	}

	d.SetId(*guide.Id)
	log.Printf("Created guide: %s", *guide.Name)
	return readGuide(ctx, d, meta)
}

func readGuide(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceGuide(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Guide: %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		guide, resp, err := proxy.getGuideById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read guide %s | Error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read guide %s | Error: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "name", guide.Name)

		d.SetId(*guide.Id)

		if guide.Status != nil {
			_ = d.Set("status", *guide.Status)
		}
		if guide.Source != nil {
			_ = d.Set("source", guide.Source)
		}
		if guide.LatestSavedVersion != nil && guide.LatestSavedVersion.Version != nil {
			_ = d.Set("latest_saved_version", guide.LatestSavedVersion.Version)
		}
		if guide.LatestProductionReadyVersion != nil && guide.LatestProductionReadyVersion.Version != nil {
			_ = d.Set("latest_production_ready_version", guide.LatestProductionReadyVersion.Version)
		}

		log.Printf("Read Guide: %s", d.Id())
		return cc.CheckState(d)
	})
}

func deleteGuide(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideProxy(sdkConfig)

	log.Printf("Deleting Guide: %s", d.Id())

	job, resp, err := proxy.deleteGuide(ctx, d.Id())
	if err != nil {
		if util.IsStatus404(resp) {
			log.Printf("Guide %s already deleted", d.Id())
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete guide %s | Error: %s", d.Id(), err), resp)
	}

	if resp.StatusCode == 202 {
		log.Printf("Delete Job for Guide: %s started", d.Id())
	}

	// Large Timeout to allow for delete job to complete before context deadline exceeded
	return util.WithRetries(ctx, 20*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getGuideById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted Guide: %s", d.Id())
				return nil
			}
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error retrieving guide %s | error: %s", d.Id(), err), resp))
		}

		jobStatus, jobResp, jobErr := proxy.getDeleteJobStatusById(ctx, job.Id, d.Id())
		if jobErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error checking delete job status for guide %s | Error: %s", d.Id(), jobErr), jobResp))
		}
		status := jobStatus.Status

		// Check the status of the delete job
		if status == "InProgress" {
			return retry.RetryableError(fmt.Errorf("delete job for guide %s still in progress: %s", d.Id(), status))
		}

		if status == "Succeeded" {
			log.Printf("Deleted Guide: %s | Status: %s", d.Id(), status)
			return nil
		}

		if status == "Failed" {
			if len(jobStatus.Errors) > 0 && jobStatus.Errors[0].Message != "" {
				return retry.NonRetryableError(fmt.Errorf("delete job failed for guide %s: %s", d.Id(), jobStatus.Errors[0].Message))
			}
			return retry.NonRetryableError(fmt.Errorf("delete job failed for guide %s | Status: %s", d.Id(), status))
		}

		return retry.RetryableError(fmt.Errorf("unexpected job status for: %s | Status: %s", d.Id(), status))
	})
}
