package routing_utilization

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllRoutingUtilization(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Although this resource typically has only a single instance,
	// we are attempting to fetch the data from the API in order to
	// verify the user's permission to access this resource's API endpoint(s).

	proxy := getRoutingUtilizationProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, err := proxy.getRoutingUtilization(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get %s due to error: %s", ResourceType, err), resp)
	}
	resources["0"] = &resourceExporter.ResourceMeta{BlockLabel: "routing_utilization"}
	return resources, nil
}

func createRoutingUtilization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating Routing Utilization")
	d.SetId("routing_utilization")
	return updateRoutingUtilization(ctx, d, meta)
}

func readRoutingUtilization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingUtilizationProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingUtilization(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Routing Utilization")

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		orgUtilization, resp, err := proxy.getRoutingUtilization(ctx)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Routing Utilization: %s", err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Routing Utilization: %s", err), resp))
		}

		if orgUtilization.Utilization != nil {
			for sdkType, schemaType := range UtilizationMediaTypes {
				if mediaSettings, ok := (*orgUtilization.Utilization)[sdkType]; ok {
					_ = d.Set(schemaType, FlattenMediaUtilization(mediaSettings))
				} else {
					_ = d.Set(schemaType, nil)
				}
			}
		}

		if orgUtilization.LabelUtilizations != nil {
			originalLabelUtilizations := d.Get("label_utilizations").([]interface{})
			// Only add the configured labels to the state, in the configured order, but not any extras, to help terraform with matching new and old state.
			flattenedLabelUtilizations := FilterAndFlattenLabelUtilizations(*orgUtilization.LabelUtilizations, originalLabelUtilizations)
			_ = d.Set("label_utilizations", flattenedLabelUtilizations)
		}

		log.Printf("Read Routing Utilization")
		return cc.CheckState(d)
	})
}

func updateRoutingUtilization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingUtilizationProxy(sdkConfig)

	log.Printf("Updating Routing Utilization")

	// Retrying on 409s because if a label is created immediately before the utilization update, it can lead to a conflict while the utilization is being updated to handle the new label.
	diagErr := util.RetryWhen(util.IsStatus409, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		_, resp, err := proxy.updateRoutingUtilization(ctx, &platformclientv2.Utilizationrequest{
			Utilization:       BuildSdkMediaUtilizations(d),
			LabelUtilizations: BuildSdkLabelUtilizations(d.Get("label_utilizations").([]interface{})),
		})

		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update Routing Utilization %s error: %s", d.Id(), err), resp)
		}
		return resp, nil
	})

	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Routing Utilization")
	return readRoutingUtilization(ctx, d, meta)
}

func deleteRoutingUtilization(ctx context.Context, _ *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingUtilizationProxy(sdkConfig)

	// Resets to default values
	log.Printf("Resetting Routing Utilization")

	resp, err := proxy.deleteRoutingUtilization(ctx)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to reset Routing Utilization | error: %s", err), resp)
	}
	log.Printf("Reset Routing Utilization")
	return nil
}
