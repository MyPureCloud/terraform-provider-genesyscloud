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
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func getAllRoutingWrapupCodes(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		wrapupcodes, _, getErr := routingAPI.GetRoutingWrapupcodes(pageSize, pageNum, "", "", "", []string{}, []string{})
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of wrapupcodes: %v", getErr)
		}

		if wrapupcodes.Entities == nil || len(*wrapupcodes.Entities) == 0 {
			break
		}

		for _, wrapupcode := range *wrapupcodes.Entities {
			resources[*wrapupcode.Id] = &resourceExporter.ResourceMeta{Name: *wrapupcode.Name}
		}
	}

	return resources, nil
}

func RoutingWrapupCodeExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllRoutingWrapupCodes),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceRoutingWrapupCode() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Wrapup Code",

		CreateContext: CreateWithPooledClient(createRoutingWrapupCode),
		ReadContext:   ReadWithPooledClient(readRoutingWrapupCode),
		UpdateContext: UpdateWithPooledClient(updateRoutingWrapupCode),
		DeleteContext: DeleteWithPooledClient(deleteRoutingWrapupCode),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Wrapup Code name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func createRoutingWrapupCode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Creating wrapupcode %s", name)
	wrapupcode, _, err := routingAPI.PostRoutingWrapupcodes(platformclientv2.Wrapupcoderequest{
		Name: &name,
	})
	if err != nil {
		return diag.Errorf("Failed to create wrapupcode %s: %s", name, err)
	}

	d.SetId(*wrapupcode.Id)
	log.Printf("Created wrapupcode %s %s", name, *wrapupcode.Id)
	return readRoutingWrapupCode(ctx, d, meta)
}

func readRoutingWrapupCode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Reading wrapupcode %s", d.Id())
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		wrapupcode, resp, getErr := routingAPI.GetRoutingWrapupcode(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read wrapupcode %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read wrapupcode %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingWrapupCode())
		d.Set("name", *wrapupcode.Name)

		log.Printf("Read wrapupcode %s %s", d.Id(), *wrapupcode.Name)
		return cc.CheckState()
	})
}

func updateRoutingWrapupCode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Updating wrapupcode %s", name)
	_, _, err := routingAPI.PutRoutingWrapupcode(d.Id(), platformclientv2.Wrapupcoderequest{
		Name: &name,
	})
	if err != nil {
		return diag.Errorf("Failed to update wrapupcode %s: %s", name, err)
	}

	log.Printf("Updated wrapupcode %s", name)

	return readRoutingWrapupCode(ctx, d, meta)
}

func deleteRoutingWrapupCode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Deleting wrapupcode %s", name)
	_, err := routingAPI.DeleteRoutingWrapupcode(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete wrapupcode %s: %s", name, err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := routingAPI.GetRoutingWrapupcode(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Routing wrapup code deleted
				log.Printf("Deleted Routing wrapup code %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting Routing wrapup code %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("Routing wrapup code %s still exists", d.Id()))
	})
}

func GenerateRoutingWrapupcodeResource(
	resourceID string,
	name string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_wrapupcode" "%s" {
		name = "%s"
	}
	`, resourceID, name)
}
