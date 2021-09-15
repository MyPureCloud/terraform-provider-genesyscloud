package genesyscloud

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v53/platformclientv2"
)

func getAllRoutingWrapupCodes(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		wrapupcodes, _, getErr := routingAPI.GetRoutingWrapupcodes(100, pageNum, "", "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of wrapupcodes: %v", getErr)
		}

		if wrapupcodes.Entities == nil || len(*wrapupcodes.Entities) == 0 {
			break
		}

		for _, wrapupcode := range *wrapupcodes.Entities {
			resources[*wrapupcode.Id] = &ResourceMeta{Name: *wrapupcode.Name}
		}
	}

	return resources, nil
}

func routingWrapupCodeExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllRoutingWrapupCodes),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceRoutingWrapupCode() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Wrapup Code",

		CreateContext: createWithPooledClient(createRoutingWrapupCode),
		ReadContext:   readWithPooledClient(readRoutingWrapupCode),
		UpdateContext: updateWithPooledClient(updateRoutingWrapupCode),
		DeleteContext: deleteWithPooledClient(deleteRoutingWrapupCode),
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

	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Creating wrapupcode %s", name)
	wrapupcode, _, err := routingAPI.PostRoutingWrapupcodes(platformclientv2.Wrapupcode{
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
	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Reading wrapupcode %s", d.Id())

	wrapupcode, resp, getErr := routingAPI.GetRoutingWrapupcode(d.Id())
	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read wrapupcode %s: %s", d.Id(), getErr)
	}

	d.Set("name", *wrapupcode.Name)

	log.Printf("Read wrapupcode %s %s", d.Id(), *wrapupcode.Name)
	return nil
}

func updateRoutingWrapupCode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Updating wrapupcode %s", name)
	_, _, err := routingAPI.PutRoutingWrapupcode(d.Id(), platformclientv2.Wrapupcode{
		Name: &name,
	})
	if err != nil {
		return diag.Errorf("Failed to update wrapupcode %s: %s", name, err)
	}

	log.Printf("Updated wrapupcode %s", name)

	// Give time for public API caches to update
	time.Sleep(5 * time.Second)
	return readRoutingWrapupCode(ctx, d, meta)
}

func deleteRoutingWrapupCode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Deleting wrapupcode %s", name)
	_, err := routingAPI.DeleteRoutingWrapupcode(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete wrapupcode %s: %s", name, err)
	}
	return nil
}
