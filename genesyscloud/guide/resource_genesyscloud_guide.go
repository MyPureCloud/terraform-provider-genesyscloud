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
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get all guide error: %s", err), resp)
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
	name := d.Get("name").(string)
	source := d.Get("source").(string)

	log.Printf("Creating Guide")
	guideReq := &Guide{
		Name:   &name,
		Source: &source,
	}

	guide, resp, err := proxy.createGuide(ctx, guideReq)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create guide error: %s", err), resp)
	}

	d.SetId(*guide.Id)
	log.Printf("Created guide: %s", *guide.Id)
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
		resourcedata.SetNillableValue(d, "source", guide.Source)
		resourcedata.SetNillableValue(d, "status", guide.Status)
		resourcedata.SetNillableValue(d, "latest_saved_version", guide.LatestSavedVersion)
		resourcedata.SetNillableValue(d, "latest_production_ready_version", guide.LatestProductionReadyVersion)

		log.Printf("Read Guide: %s", d.Id())
		return cc.CheckState(d)
	})
}
