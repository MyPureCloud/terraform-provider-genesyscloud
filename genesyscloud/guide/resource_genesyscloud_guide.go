package guide

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"log"
)

func getAllGuides(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getGuideProxy(clientConfig)

	log.Printf("Retrieving all Guides")
	guides, resp, err := proxy.getAllGuides(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get all guides | error: %s", err), resp)
	}

	if guides.Entities == nil {
		return resources, nil
	}

	for _, guide := range *guides.Entities {
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

	guideReq := &Createguide{
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

		if guide.Status != nil {
			_ = d.Set("status", *guide.Status)
		}

		if guide.Source != nil {
			_ = d.Set("source", guide.Source)
		}

		if guide.LatestSavedVersion != nil && guide.LatestSavedVersion.version != nil {
			log.Println("Latest Saved Version:", *guide.LatestSavedVersion)
			_ = d.Set("latest_saved_version", guide.LatestSavedVersion.version)
		}
		if guide.LatestProductionReadyVersion != nil && guide.LatestProductionReadyVersion.version != nil {
			_ = d.Set("latest_production_ready_version", guide.LatestProductionReadyVersion.version)
		}

		log.Printf("Read Guide: %s", d.Id())
		return cc.CheckState(d)
	})
}

func deleteGuide(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideProxy(sdkConfig)

	log.Printf("Deleting Guide: %s", d.Id())

	resp, err := proxy.deleteGuide(ctx, d.Id())
	if err != nil {
		if util.IsStatus404(resp) {
			log.Printf("Guide %s already deleted", d.Id())
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete guide %s | Error: %s", d.Id(), err), resp)
	}

	if resp.StatusCode == 202 {
		log.Printf("Delete guide job started for: %s", d.Id())
		return nil
	}

	return util.WithRetries(ctx, 180, func() *retry.RetryError {
		_, resp, err := proxy.getGuideById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted Guide: %s", d.Id())
				return nil
			}
			log.Printf("Error checking if guide %s is deleted: %s", d.Id(), err)
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Guide %s still exists", d.Id()), resp))
	})
}
