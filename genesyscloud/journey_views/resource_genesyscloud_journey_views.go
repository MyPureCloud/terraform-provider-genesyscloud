package journey_views

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllJourneyViews(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getJourneyViewProxy(clientConfig)

	journeys, resp, err := proxy.getAllJourneyViews(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get journey views: %s", err), resp)
	}

	if journeys == nil || len(*journeys) == 0 {
		return resources, nil
	}

	for _, journey := range *journeys {
		resources[*journey.Id] = &resourceExporter.ResourceMeta{BlockLabel: *journey.Name}
	}

	return resources, nil
}

func createJourneyView(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	gp := getJourneyViewProxy(sdkConfig)

	journeyView := makeJourneyViewFromSchema(d)
	log.Printf("Creating journeyView %s", name)
	journeyView, resp, err := gp.createJourneyView(ctx, journeyView)

	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create journeyView %s: %s", name, err), resp)
	}
	d.SetId(*journeyView.Id)
	log.Printf("Created journeyView with viewId: %s", d.Id())
	return readJourneyView(ctx, d, meta)
}

func updateJourneyView(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	versionId := d.Get("version").(int)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	gp := getJourneyViewProxy(sdkConfig)

	journeyView := makeJourneyViewFromSchema(d)
	log.Printf("Updating journeyView %s", d.Id())
	journeyView, resp, err := gp.updateJourneyView(ctx, d.Id(), versionId, journeyView)

	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create journeyView %s: %s", name, err), resp)
	}
	log.Printf("Updated journeyView %s", d.Id())
	return readJourneyView(ctx, d, meta)
}

func readJourneyView(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	viewId := d.Id()

	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceJourneyViews(), constants.ConsistencyChecks(), ResourceType)
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	gp := getJourneyViewProxy(sdkConfig)
	log.Printf("Getting journeyView with viewId: %s", viewId)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		journeyView, resp, err := gp.getJourneyViewById(ctx, viewId)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to get journeyView with viewId %s | error: %s", viewId, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to get journeyView with viewId %s | error: %s", viewId, err), resp))
		}

		resourcedata.SetNillableValue(d, "name", journeyView.Name)
		resourcedata.SetNillableValue(d, "description", journeyView.Description)
		resourcedata.SetNillableValue(d, "interval", journeyView.Interval)
		resourcedata.SetNillableValue(d, "duration", journeyView.Duration)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "elements", journeyView.Elements, flattenElements)

		return cc.CheckState(d)
	})
}

func deleteJourneyView(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	viewId := d.Id()

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	gp := getJourneyViewProxy(sdkConfig)

	util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Directory occasionally returns version errors on deletes if an object was updated at the same time.
		log.Printf("Deleting journeyView with viewId %s", viewId)
		resp, err := gp.deleteJourneyView(ctx, viewId)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete journeyView with viewId %s: %s", viewId, err), resp)
		}
		return nil, nil
	})

	return util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		_, resp, err := gp.getJourneyViewById(ctx, viewId)
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("JourneyView %s deleted", viewId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting joruneyView with viewId %s | error: %s", viewId, err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("JourneyView with viewId %s still exists", viewId), resp))
	})
}

func makeJourneyViewFromSchema(d *schema.ResourceData) *platformclientv2.Journeyview {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	interval := d.Get("interval").(string)
	duration := d.Get("duration").(string)
	elements, _ := buildElements(d)

	journeyView := &platformclientv2.Journeyview{
		Name:        &name,
		Description: &description,
		Interval:    &interval,
		Duration:    &duration,
		Elements:    elements,
	}
	return journeyView
}
