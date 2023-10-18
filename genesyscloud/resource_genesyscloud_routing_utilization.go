package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

var (
	// Map of SDK media type name to schema media type name
	utilizationMediaTypes = map[string]string{
		"call":     "call",
		"callback": "callback",
		"chat":     "chat",
		"email":    "email",
		"message":  "message",
	}

	utilizationSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"maximum_capacity": {
				Description:  "Maximum capacity of conversations of this media type. Value must be between 0 and 25.",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 25),
			},
			"interruptible_media_types": {
				Description: fmt.Sprintf("Set of other media types that can interrupt this media type (%s).", strings.Join(getSdkUtilizationTypes(), " | ")),
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"include_non_acd": {
				Description: "Block this media type when on a non-ACD conversation.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
)

func getSdkUtilizationTypes() []string {
	types := make([]string, 0, len(utilizationMediaTypes))
	for t := range utilizationMediaTypes {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

func getAllRoutingUtilization(_ context.Context, _ *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Routing utilization config always exists
	resources := make(resourceExporter.ResourceIDMetaMap)
	resources["0"] = &resourceExporter.ResourceMeta{Name: "routing_utilization"}
	return resources, nil
}

func RoutingUtilizationExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllRoutingUtilization),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
		AllowZeroValues:  []string{"maximum_capacity"},
	}
}

func ResourceRoutingUtilization() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Org-wide Routing Utilization Settings.",

		CreateContext: CreateWithPooledClient(createRoutingUtilization),
		ReadContext:   ReadWithPooledClient(readRoutingUtilization),
		UpdateContext: UpdateWithPooledClient(updateRoutingUtilization),
		DeleteContext: DeleteWithPooledClient(deleteRoutingUtilization),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(8 * time.Minute),
			Read:   schema.DefaultTimeout(8 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"call": {
				Description: "Call media settings. If not set, this reverts to the default media type settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        utilizationSettingsResource,
			},
			"callback": {
				Description: "Callback media settings. If not set, this reverts to the default media type settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        utilizationSettingsResource,
			},
			"message": {
				Description: "Message media settings. If not set, this reverts to the default media type settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        utilizationSettingsResource,
			},
			"email": {
				Description: "Email media settings. If not set, this reverts to the default media type settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        utilizationSettingsResource,
			},
			"chat": {
				Description: "Chat media settings. If not set, this reverts to the default media type settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        utilizationSettingsResource,
			},
		},
	}
}

func createRoutingUtilization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating Routing Utilization")
	d.SetId("routing_utilization")
	return updateRoutingUtilization(ctx, d, meta)
}

func readRoutingUtilization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Reading Routing Utilization")
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		settings, resp, getErr := routingAPI.GetRoutingUtilization()
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read Routing Utilization: %s", getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Routing Utilization: %s", getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingSkill())
		if settings.Utilization != nil {
			for sdkType, schemaType := range utilizationMediaTypes {
				if mediaSettings, ok := (*settings.Utilization)[sdkType]; ok {
					d.Set(schemaType, flattenUtilizationSetting(mediaSettings))
				} else {
					d.Set(schemaType, nil)
				}
			}
		}

		log.Printf("Read Routing Utilization")
		return cc.CheckState()
	})
}

func updateRoutingUtilization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Updating Routing Utilization")

	_, _, err := routingAPI.PutRoutingUtilization(platformclientv2.Utilization{
		Utilization: buildSdkRoutingUtilizations(d),
	})
	if err != nil {
		return diag.Errorf("Failed to update Routing Utilization: %s", err)
	}

	log.Printf("Updated Routing Utilization")
	return readRoutingUtilization(ctx, d, meta)
}

func deleteRoutingUtilization(_ context.Context, _ *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	// Resets to default values
	log.Printf("Resetting Routing Utilization")
	_, err := routingAPI.DeleteRoutingUtilization()
	if err != nil {
		return diag.Errorf("Failed to reset Routing Utilization: %s", err)
	}
	log.Printf("Reset Routing Utilization")
	return nil
}

func flattenUtilizationSetting(settings platformclientv2.Mediautilization) []interface{} {
	settingsMap := make(map[string]interface{})
	if settings.MaximumCapacity != nil {
		settingsMap["maximum_capacity"] = *settings.MaximumCapacity
	}
	if settings.InterruptableMediaTypes != nil {
		settingsMap["interruptible_media_types"] = lists.StringListToSet(*settings.InterruptableMediaTypes)
	}
	if settings.IncludeNonAcd != nil {
		settingsMap["include_non_acd"] = *settings.IncludeNonAcd
	}
	return []interface{}{settingsMap}
}

func buildSdkRoutingUtilizations(d *schema.ResourceData) *map[string]platformclientv2.Mediautilization {
	settings := make(map[string]platformclientv2.Mediautilization)

	for sdkType, schemaType := range utilizationMediaTypes {
		mediaSettings := d.Get(schemaType).([]interface{})
		if mediaSettings != nil && len(mediaSettings) > 0 {
			settings[sdkType] = buildSdkMediaUtilization(mediaSettings)
		}
	}

	return &settings
}

func buildSdkMediaUtilization(settings []interface{}) platformclientv2.Mediautilization {
	settingsMap := settings[0].(map[string]interface{})

	maxCapacity := settingsMap["maximum_capacity"].(int)
	includeNonAcd := settingsMap["include_non_acd"].(bool)

	// Optional
	interruptableMediaTypes := &[]string{}
	if types, ok := settingsMap["interruptible_media_types"]; ok {
		interruptableMediaTypes = lists.SetToStringList(types.(*schema.Set))
	}

	return platformclientv2.Mediautilization{
		MaximumCapacity:         &maxCapacity,
		IncludeNonAcd:           &includeNonAcd,
		InterruptableMediaTypes: interruptableMediaTypes,
	}
}
