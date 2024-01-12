package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

func getAllRoutingUtilizationLabels(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		labels, _, getErr := routingAPI.GetRoutingUtilizationLabels(pageSize, pageNum, "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of labels: %v", getErr)
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
		GetResourcesFunc: GetAllWithPooledClient(getAllRoutingUtilizationLabels),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceRoutingUtilizationLabel() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Utilization Label. This resource is not yet widely available. Only use it if the feature is enabled.",

		CreateContext: CreateWithPooledClient(createRoutingUtilizationLabel),
		ReadContext:   ReadWithPooledClient(readRoutingUtilizationLabel),
		UpdateContext: UpdateWithPooledClient(updateRoutingUtilizationLabel),
		DeleteContext: DeleteWithPooledClient(deleteRoutingUtilizationLabel),
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

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Creating label %s", name)
	label, _, err := routingAPI.PostRoutingUtilizationLabels(platformclientv2.Createutilizationlabelrequest{
		Name: &name,
	})
	if err != nil {
		return diag.Errorf("Failed to create label %s: %s", name, err)
	}

	d.SetId(*label.Id)

	log.Printf("Created label %s %s", name, *label.Id)
	return readRoutingUtilizationLabel(ctx, d, meta)
}

func updateRoutingUtilizationLabel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	id := d.Id()
	name := d.Get("name").(string)

	log.Printf("Updating label %s with name %s", id, name)

	_, _, err := routingAPI.PutRoutingUtilizationLabel(id, platformclientv2.Updateutilizationlabelrequest{
		Name: &name,
	})
	if err != nil {
		return diag.Errorf("Failed to update label %s: %s", id, err)
	}

	log.Printf("Updated label %s", id)
	return readRoutingUtilizationLabel(ctx, d, meta)
}

func readRoutingUtilizationLabel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Reading label %s", d.Id())
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		label, resp, getErr := routingApi.GetRoutingUtilizationLabel(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read label %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read label %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingUtilizationLabel())
		d.Set("name", *label.Name)
		log.Printf("Read label %s %s", d.Id(), *label.Name)
		return cc.CheckState()
	})
}

func deleteRoutingUtilizationLabel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Deleting label %s", name)
	_, err := routingApi.DeleteRoutingUtilizationLabel(d.Id(), true)

	if err != nil {
		return diag.Errorf("Failed to delete label %s: %s", name, err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := routingApi.GetRoutingUtilizationLabel(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Routing label deleted
				log.Printf("Deleted Routing label %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting Routing label %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Routing label %s still exists", d.Id()))
	})
}
