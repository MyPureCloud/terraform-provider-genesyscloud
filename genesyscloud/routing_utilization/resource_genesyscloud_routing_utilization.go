package routing_utilization

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

func getAllRoutingUtilization(_ context.Context, _ *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Routing utilization config always exists
	resources := make(resourceExporter.ResourceIDMetaMap)
	resources["0"] = &resourceExporter.ResourceMeta{Name: "routing_utilization"}
	return resources, nil
}

func createRoutingUtilization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating Routing Utilization")
	d.SetId("routing_utilization")
	return updateRoutingUtilization(ctx, d, meta)
}

func readRoutingUtilization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Calling the Utilization API directly while the label feature is not available.
	// Once it is, this code can go back to using platformclientv2's RoutingApi to make the call.
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingUtilizationProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingUtilization(), constants.DefaultConsistencyChecks, resourceName)
	orgUtilization := &OrgUtilizationWithLabels{}

	log.Printf("Reading Routing Utilization")

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		resp, err := proxy.getRoutingUtilization(ctx)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read Routing Utilization: %s", err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read Routing Utilization: %s", err), resp))
		}

		err = json.Unmarshal(resp.RawBody, &orgUtilization)

		if orgUtilization.Utilization != nil {
			for sdkType, schemaType := range UtilizationMediaTypes {
				if mediaSettings, ok := orgUtilization.Utilization[sdkType]; ok {
					_ = d.Set(schemaType, FlattenUtilizationSetting(mediaSettings))
				} else {
					_ = d.Set(schemaType, nil)
				}
			}
		}

		if orgUtilization.LabelUtilizations != nil {
			originalLabelUtilizations := d.Get("label_utilizations").([]interface{})
			// Only add to the state the configured labels, in the configured order, but not any extras, to help terraform with matching new and old state.
			flattenedLabelUtilizations := FilterAndFlattenLabelUtilizations(orgUtilization.LabelUtilizations, originalLabelUtilizations)
			_ = d.Set("label_utilizations", flattenedLabelUtilizations)
		}

		log.Printf("Read Routing Utilization")
		return cc.CheckState(d)
	})
}

func updateRoutingUtilization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingUtilizationProxy(sdkConfig)

	labelUtilizations := d.Get("label_utilizations").([]interface{})
	var resp *platformclientv2.APIResponse
	var err error

	log.Printf("Updating Routing Utilization")

	// Retrying on 409s because if a label is created immediately before the utilization update, it can lead to a conflict while the utilization is being updated to handle the new label.
	diagErr := util.RetryWhen(util.IsStatus409, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// If the resource has label(s), calls the Utilization API directly.
		// This code can go back to using platformclientv2's RoutingApi to make the call once label utilization is available in platformclientv2's RoutingApi.
		if labelUtilizations != nil && len(labelUtilizations) > 0 {
			resp, err := proxy.updateDirectly(ctx, d, labelUtilizations)
			if err != nil {
				return resp, util.BuildAPIDiagnosticError(resourceName, "Failed to update routing utilization directly", resp)
			}
		} else {
			_, resp, err = proxy.updateRoutingUtilization(ctx, &platformclientv2.Utilizationrequest{
				Utilization: buildSdkMediaUtilizations(d),
			})
		}

		if err != nil {
			return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update Routing Utilization %s error: %s", d.Id(), err), resp)
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
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to reset Routing Utilization | error: %s", err), resp)
	}
	log.Printf("Reset Routing Utilization")
	return nil
}
