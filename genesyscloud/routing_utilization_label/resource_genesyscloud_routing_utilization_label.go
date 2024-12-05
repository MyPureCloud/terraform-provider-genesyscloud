package routing_utilization_label

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllRoutingUtilizationLabels(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getRoutingUtilizationLabelProxy(clientConfig)

	labels, resp, getErr := proxy.getAllRoutingUtilizationLabels(ctx, "")
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of labels error: %s", getErr), resp)
	}

	for _, label := range *labels {
		resources[*label.Id] = &resourceExporter.ResourceMeta{BlockLabel: *label.Name}
	}
	return resources, nil
}

func createRoutingUtilizationLabel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingUtilizationLabelProxy(sdkConfig)

	log.Printf("Creating label %s", name)

	label, resp, err := proxy.createRoutingUtilizationLabel(ctx, &platformclientv2.Createutilizationlabelrequest{
		Name: &name,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create label %s error: %s", name, err), resp)
	}

	d.SetId(*label.Id)

	log.Printf("Created label %s %s", name, *label.Id)
	return readRoutingUtilizationLabel(ctx, d, meta)
}

func readRoutingUtilizationLabel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingUtilizationLabelProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingUtilizationLabel(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading label %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		label, resp, getErr := proxy.getRoutingUtilizationLabel(ctx, d.Id())

		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read label %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read label %s | error: %s", d.Id(), getErr), resp))
		}

		_ = d.Set("name", *label.Name)
		log.Printf("Read label %s %s", d.Id(), *label.Name)
		return cc.CheckState(d)
	})
}

func updateRoutingUtilizationLabel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingUtilizationLabelProxy(sdkConfig)

	id := d.Id()
	name := d.Get("name").(string)

	log.Printf("Updating label %s with name %s", id, name)
	_, resp, err := proxy.updateRoutingUtilizationLabel(ctx, id, &platformclientv2.Updateutilizationlabelrequest{
		Name: &name,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update label %s error: %s", id, err), resp)
	}

	log.Printf("Updated label %s", id)
	return readRoutingUtilizationLabel(ctx, d, meta)
}

func deleteRoutingUtilizationLabel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingUtilizationLabelProxy(sdkConfig)

	log.Printf("Deleting label %s %s", d.Id(), name)
	resp, err := proxy.deleteRoutingUtilizationLabel(ctx, d.Id(), true)

	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete label %s error: %s", name, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getRoutingUtilizationLabel(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted Routing label %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting Routing label %s: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Routing label %s still exists", d.Id()), resp))
	})
}
