package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v129/platformclientv2"
)

func getAllRoutingUtilizationLabels(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		labels, resp, getErr := routingAPI.GetRoutingUtilizationLabels(pageSize, pageNum, "", "")
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_routing_utilization_label", fmt.Sprintf("Failed to get page of labels error: %s", getErr), resp)
		}

		if labels.Entities == nil || len(*labels.Entities) == 0 {
			break
		}

		for _, label := range *labels.Entities {
			resources[*label.Id] = &resourceExporter.ResourceMeta{Name: *label.Name}
		}
	}

	return resources, nil
}

func RoutingUtilizationLabelExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRoutingUtilizationLabels),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceRoutingUtilizationLabel() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Utilization Label. This resource is not yet widely available. Only use it if the feature is enabled.",

		CreateContext: provider.CreateWithPooledClient(createRoutingUtilizationLabel),
		ReadContext:   provider.ReadWithPooledClient(readRoutingUtilizationLabel),
		UpdateContext: provider.UpdateWithPooledClient(updateRoutingUtilizationLabel),
		DeleteContext: provider.DeleteWithPooledClient(deleteRoutingUtilizationLabel),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Label name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func createRoutingUtilizationLabel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Creating label %s", name)
	label, resp, err := routingAPI.PostRoutingUtilizationLabels(platformclientv2.Createutilizationlabelrequest{
		Name: &name,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_routing_utilization_label", fmt.Sprintf("Failed to create label %s error: %s", name, err), resp)
	}

	d.SetId(*label.Id)

	log.Printf("Created label %s %s", name, *label.Id)
	return readRoutingUtilizationLabel(ctx, d, meta)
}

func updateRoutingUtilizationLabel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	id := d.Id()
	name := d.Get("name").(string)

	log.Printf("Updating label %s with name %s", id, name)

	_, resp, err := routingAPI.PutRoutingUtilizationLabel(id, platformclientv2.Updateutilizationlabelrequest{
		Name: &name,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_routing_utilization_label", fmt.Sprintf("Failed to update label %s error: %s", id, err), resp)
	}

	log.Printf("Updated label %s", id)
	return readRoutingUtilizationLabel(ctx, d, meta)
}

func readRoutingUtilizationLabel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Reading label %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		label, resp, getErr := routingApi.GetRoutingUtilizationLabel(d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_utilization_label", fmt.Sprintf("Failed to read label %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_utilization_label", fmt.Sprintf("Failed to read label %s | error: %s", d.Id(), getErr), resp))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingUtilizationLabel())
		d.Set("name", *label.Name)
		log.Printf("Read label %s %s", d.Id(), *label.Name)
		return cc.CheckState()
	})
}

func deleteRoutingUtilizationLabel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Deleting label %s", name)
	resp, err := routingApi.DeleteRoutingUtilizationLabel(d.Id(), true)

	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_routing_utilization_label", fmt.Sprintf("Failed to delete label %s error: %s", name, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := routingApi.GetRoutingUtilizationLabel(d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Routing label deleted
				log.Printf("Deleted Routing label %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_utilization_label", fmt.Sprintf("Error deleting Routing label %s: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_utilization_label", fmt.Sprintf("Routing label %s still exists", d.Id()), resp))
	})
}
