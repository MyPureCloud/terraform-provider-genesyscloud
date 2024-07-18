package genesyscloud

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
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func getAllRoutingWrapupCodes(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		wrapupcodes, resp, getErr := routingAPI.GetRoutingWrapupcodes(pageSize, pageNum, "", "", "", []string{}, []string{})
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_routing_wrapupcode", fmt.Sprintf("Failed to get wrapupcodes error: %s", getErr), resp)
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
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRoutingWrapupCodes),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceRoutingWrapupCode() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Wrapup Code",

		CreateContext: provider.CreateWithPooledClient(createRoutingWrapupCode),
		ReadContext:   provider.ReadWithPooledClient(readRoutingWrapupCode),
		UpdateContext: provider.UpdateWithPooledClient(updateRoutingWrapupCode),
		DeleteContext: provider.DeleteWithPooledClient(deleteRoutingWrapupCode),
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

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Creating wrapupcode %s", name)
	wrapupcode, resp, err := routingAPI.PostRoutingWrapupcodes(platformclientv2.Wrapupcoderequest{
		Name: &name,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_routing_wrapupcode", fmt.Sprintf("Failed to create wrapupcode %s error: %s", name, err), resp)
	}

	d.SetId(*wrapupcode.Id)
	log.Printf("Created wrapupcode %s %s", name, *wrapupcode.Id)
	return readRoutingWrapupCode(ctx, d, meta)
}

func readRoutingWrapupCode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingWrapupCode(), constants.DefaultConsistencyChecks, "genesyscloud_routing_wrapupcode")

	log.Printf("Reading wrapupcode %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		wrapupcode, resp, getErr := routingAPI.GetRoutingWrapupcode(d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_wrapupcode", fmt.Sprintf("Failed to read wrapupcode %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_wrapupcode", fmt.Sprintf("Failed to read wrapupcode %s | error: %s", d.Id(), getErr), resp))
		}

		d.Set("name", *wrapupcode.Name)

		log.Printf("Read wrapupcode %s %s", d.Id(), *wrapupcode.Name)
		return cc.CheckState(d)
	})
}

func updateRoutingWrapupCode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Updating wrapupcode %s", name)
	_, resp, err := routingAPI.PutRoutingWrapupcode(d.Id(), platformclientv2.Wrapupcoderequest{
		Name: &name,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_routing_wrapupcode", fmt.Sprintf("Failed to update wrapupcode %s error: %s", name, err), resp)
	}

	log.Printf("Updated wrapupcode %s", name)

	return readRoutingWrapupCode(ctx, d, meta)
}

func deleteRoutingWrapupCode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Deleting wrapupcode %s", name)
	resp, err := routingAPI.DeleteRoutingWrapupcode(d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_routing_wrapupcode", fmt.Sprintf("Failed to delete wrapupcode %s error: %s", name, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := routingAPI.GetRoutingWrapupcode(d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Routing wrapup code deleted
				log.Printf("Deleted Routing wrapup code %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_wrapupcode", fmt.Sprintf("Error deleting Routing wrapup code %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_wrapupcode", fmt.Sprintf("Routing wrapup code %s still exists", d.Id()), resp))
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
