package outbound

import (
	"context"
	"fmt"
	"log"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

func getOutboundWrapupCodeMappings(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	outboundApi := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, getErr := outboundApi.GetOutboundWrapupcodemappings()
	if getErr != nil {
		if gcloud.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, diag.Errorf("Failed to get wrap-up code mappings: %v", getErr)
	}

	resources["0"] = &resourceExporter.ResourceMeta{Name: "wrapupcodemappings"}
	return resources, nil
}

func OutboundWrapupCodeMappingsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getOutboundWrapupCodeMappings),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			`mappings.wrapup_code_id`: {
				RefType: `genesyscloud_routing_wrapupcode`,
			},
		},
	}
}

func ResourceOutboundWrapUpCodeMappings() *schema.Resource {
	return &schema.Resource{
		Description:   `Genesys Cloud Outbound Wrap-up Code Mappings`,
		CreateContext: gcloud.CreateWithPooledClient(createOutboundWrapUpCodeMappings),
		ReadContext:   gcloud.ReadWithPooledClient(readOutboundWrapUpCodeMappings),
		UpdateContext: gcloud.UpdateWithPooledClient(updateOutboundWrapUpCodeMappings),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteOutboundWrapUpCodeMappings),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`default_set`: {
				Description: `The default set of wrap-up flags. These will be used if there is no entry for a given wrap-up code in the mapping.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"CONTACT_UNCALLABLE", "NUMBER_UNCALLABLE", "RIGHT_PARTY_CONTACT"}, true),
				},
			},
			`mappings`: {
				Description: `A map from wrap-up code identifiers to a set of wrap-up flags.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						`wrapup_code_id`: {
							Description: `The wrap-up code identifier.`,
							Required:    true,
							Type:        schema.TypeString,
						},
						`flags`: {
							Description: `The set of wrap-up flags.`,
							Required:    true,
							Type:        schema.TypeList,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{"CONTACT_UNCALLABLE", "NUMBER_UNCALLABLE", "RIGHT_PARTY_CONTACT"}, true),
							},
						},
					},
				},
			},
			`placeholder`: {
				Description:  `Placeholder data used internally by the provider.`,
				Optional:     true,
				Type:         schema.TypeString,
				Default:      "***",
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func createOutboundWrapUpCodeMappings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating Outbound Wrap-up Code Mappings")
	d.SetId("wrapupcodemappings")
	return updateOutboundWrapUpCodeMappings(ctx, d, meta)
}

func updateOutboundWrapUpCodeMappings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	log.Printf("Updating Outbound Wrap-up Code Mappings")
	diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		wrapupCodeMappings, resp, err := outboundApi.GetOutboundWrapupcodemappings()
		if err != nil {
			return resp, diag.Errorf("failed to read wrap-up code mappings: %s", err)
		}
		wrapupCodeUpdate := platformclientv2.Wrapupcodemapping{
			DefaultSet: lists.BuildSdkStringListFromInterfaceArray(d, "default_set"),
			Mapping:    buildWrapupCodeMappings(d),
			Version:    wrapupCodeMappings.Version,
		}
		_, _, err = outboundApi.PutOutboundWrapupcodemappings(wrapupCodeUpdate)
		if err != nil {
			return resp, diag.Errorf("failed to update wrap-up code mappings: %s", err)
		}
		return resp, nil
	})

	if diagErr != nil {
		return diagErr
	}

	log.Print("Updated Outbound Wrap-up Code Mappings")
	return readOutboundWrapUpCodeMappings(ctx, d, meta)
}

func readOutboundWrapUpCodeMappings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	log.Printf("Reading Outbound Wrap-up Code Mappings")

	return gcloud.WithRetriesForRead(ctx, d, func() *resource.RetryError {
		sdkWrapupCodeMappings, resp, err := outboundApi.GetOutboundWrapupcodemappings()
		if err != nil {
			if gcloud.IsStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read Outbound Wrap-up Code Mappings: %s", err))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read Outbound Wrap-up Code Mappings: %s", err))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundWrapUpCodeMappings())

		// Match new random ordering of list returned from API
		if sdkWrapupCodeMappings.DefaultSet != nil {
			defaultSet := make([]string, 0)
			schemaDefaultSet := d.Get("default_set").([]interface{})
			for _, v := range schemaDefaultSet {
				defaultSet = append(defaultSet, v.(string))
			}
			if lists.ListsAreEquivalent(defaultSet, *sdkWrapupCodeMappings.DefaultSet) {
				_ = d.Set("default_set", defaultSet)
			} else {
				_ = d.Set("default_set", lists.StringListToInterfaceList(*sdkWrapupCodeMappings.DefaultSet))
			}
		}

		if sdkWrapupCodeMappings.Mapping != nil {
			_ = d.Set("mappings", flattenOutboundWrapupCodeMappings(d, sdkWrapupCodeMappings))
		}

		log.Print("Read Outbound Wrap-up Code Mappings")
		return cc.CheckState()
	})
}

func deleteOutboundWrapUpCodeMappings(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Does not delete the wrap-up code mappings. This resource will just no longer manage them.
	return nil
}

// Mapping objects and flags lists come back ordered differently than what is defined by the user in their config
// To avoid plan not empty errors, this function:
// checks that the maps/lists from the schema & sdk returned data are equivalent before returning the data in it's original order.
func flattenOutboundWrapupCodeMappings(d *schema.ResourceData, sdkWrapupcodemapping *platformclientv2.Wrapupcodemapping) []interface{} {
	mappings := make([]interface{}, 0)
	schemaMappings := d.Get("mappings").([]interface{})

	// If read is called from export function, placeholder field should not exist
	// In this case, dump whatever is returned from the API.
	if _, exists := d.GetOkExists("placeholder"); !exists {
		for sdkId, sdkFlags := range *sdkWrapupcodemapping.Mapping {
			currentMap := make(map[string]interface{}, 0)
			currentMap["wrapup_code_id"] = sdkId
			currentMap["flags"] = lists.StringListToInterfaceList(sdkFlags)
			mappings = append(mappings, currentMap)
		}
		return mappings
	}

	for _, m := range schemaMappings {
		if mMap, ok := m.(map[string]interface{}); ok {
			var schemaFlags []string
			if flags, ok := mMap["flags"].([]interface{}); ok {
				schemaFlags = lists.InterfaceListToStrings(flags)
			}
			for sdkId, sdkFlags := range *sdkWrapupcodemapping.Mapping {
				if mMap["wrapup_code_id"].(string) == sdkId {
					currentMap := make(map[string]interface{}, 0)
					currentMap["wrapup_code_id"] = sdkId
					if lists.ListsAreEquivalent(schemaFlags, sdkFlags) {
						currentMap["flags"] = lists.StringListToInterfaceList(schemaFlags)
					} else {
						currentMap["flags"] = lists.StringListToInterfaceList(sdkFlags)
					}
					mappings = append(mappings, currentMap)
				}
			}
		}
	}
	return mappings
}

func buildWrapupCodeMappings(d *schema.ResourceData) *map[string][]string {
	wrapupCodeMappings := make(map[string][]string, 0)
	if mappings := d.Get("mappings").([]interface{}); mappings != nil && len(mappings) > 0 {
		for _, m := range mappings {
			if mapping, ok := m.(map[string]interface{}); ok {
				id := mapping["wrapup_code_id"].(string)
				flags := lists.InterfaceListToStrings(mapping["flags"].([]interface{}))
				wrapupCodeMappings[id] = flags
			}
		}
	}
	return &wrapupCodeMappings
}
