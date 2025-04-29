package telephony_providers_edges_site

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/validators"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesyscloud_telephony_providers_edges_site_schema_v1.go holds information about the old v1 schema and how to upgrade to the next schema version
*/

func resourceSiteResourceV1() *schema.Resource {

	edgeAutoUpdateConfigSchema := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"time_zone": {
				Description: "The timezone of the window in which any updates to the edges assigned to the site can be applied. The minimum size of the window is 2 hours.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"rrule": {
				Description:      "A reoccurring rule for updating the Edges assigned to the site. The only supported frequencies are daily and weekly. Weekly frequencies require a day list with at least oneday specified. All other configurations are not supported.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validators.ValidateRrule,
			},
			"start": {
				Description: "Date time is represented as an ISO-8601 string without a timezone. For example: yyyy-MM-ddTHH:mm:ss.SSS",
				Type:        schema.TypeString,
				Required:    true,
			},
			"end": {
				Description: "Date time is represented as an ISO-8601 string without a timezone. For example: yyyy-MM-ddTHH:mm:ss.SSS",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}

	numberPlansSchema := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the entity.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"match_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"digitLength", "e164NumberList", "interCountryCode", "intraCountryCode", "numberList", "regex"}, false),
			},
			"normalized_format": {
				Description: "Use regular expression capture groups to build the normalized number",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"match_format": {
				Description: "Use regular expression capture groups to build the normalized number",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"numbers": {
				Description: "Numbers must be 2-9 digits long. Numbers within ranges must be the same length. (e.g. 888, 888-999, 55555-77777, 800).",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"end": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"digit_length": {
				Description: "Allowed values are between 1-20 digits.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"end": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"classification": {
				Description: "Used to classify this number plan",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}

	return &schema.Resource{
		Description:   "Genesys Cloud Site",
		CreateContext: provider.CreateWithPooledClient(createSite),
		ReadContext:   provider.ReadWithPooledClient(readSite),
		UpdateContext: provider.UpdateWithPooledClient(updateSite),
		DeleteContext: provider.DeleteWithPooledClient(deleteSite),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the entity.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The resource's description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"location_id": {
				Description: "Site location ID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"media_model": {
				Description:  "Media model for the site Valid Values: Premises, Cloud. Changing the media_model attribute will cause the site object to be dropped and created with a new ID.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Premises", "Cloud"}, false),
				ForceNew:     true,
			},
			"media_regions_use_latency_based": {
				Description: "Latency based on media region",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"media_regions": {
				Description: "The ordered list of AWS regions through which media can stream. A full list of available media regions can be found at the GET /api/v2/telephony/mediaregions endpoint",
				Type:        schema.TypeList, //This has to be a list because it must be ordered
				Optional:    true,
				Computed:    true, //This needs to be a computed field because the sites API automatically adds the home region to whatever regions you add add.
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"caller_id": {
				Description:      "The caller ID value for the site. The callerID must be a valid E.164 formatted phone number",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validators.ValidatePhoneNumber,
			},
			"caller_name": {
				Description: "The caller name for the site",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"edge_auto_update_config": {
				Description: "Recurrence rule, time zone, and start/end settings for automatic edge updates for this site",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        edgeAutoUpdateConfigSchema,
			},
			"number_plans": {
				Description: "Number plans for the site. The order of the plans in the resource file determines the priority of the plans. Specifying number plans will not result in the default plans being overwritten.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem:        numberPlansSchema,
			},
			"outbound_routes": {
				Description: "Outbound Routes for the site. The default outbound route will be deleted if routes are specified",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				ConfigMode:  schema.SchemaConfigModeAttr,
				Elem:        outboundRouteSchema,
				Deprecated:  fmt.Sprintf("The outbound routes property is deprecated in %s, please use independent outbound routes resource instead, genesyscloud_telephony_providers_edges_site_outbound_route", ResourceType),
			},
			"primary_sites": {
				Description: `Used for primary phone edge assignment on physical edges only.  List of primary sites the phones can be assigned to. If no primary_sites are defined, the site id for this site will be used as the primary site id.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"secondary_sites": {
				Description: `Used for secondary phone edge assignment on physical edges only.  List of secondary sites the phones can be assigned to.  If no primary_sites or secondary_sites are defined then the current site will defined as primary and secondary. `,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"set_as_default_site": {
				Description: `Set this site as the default site for the organization. Only one genesyscloud_telephony_providers_edges_site resource should be set as the default.`,
				Optional:    true,
				Default:     false,
				Type:        schema.TypeBool,
			},
			"managed": {
				Description: "Is this site managed by Genesys Cloud",
				Type:        schema.TypeBool,
				Optional:    false,
				Required:    false,
				Computed:    true,
			},
		},
		CustomizeDiff: customizeSiteDiff,
	}
}

func resourceSiteUpgradeV1ToV2(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	// Get the site ID
	siteID := rawState["id"].(string)

	// Get the outbound routes from the old state
	if outboundRoutes, ok := rawState["outbound_routes"].(*schema.Set); ok && outboundRoutes != nil {
		fmt.Printf("\n[MIGRATION] Found %d outbound routes in site %s that need to be migrated to the new resource type.\n", outboundRoutes.Len(), siteID)
		fmt.Println("[MIGRATION] Add the following blocks to your configuration to manage these routes:")

		for _, route := range outboundRoutes.List() {
			routeMap := route.(map[string]interface{})
			routeID := routeMap["id"].(string)

			// Output the configuration block for each route
			fmt.Printf(`
resource "genesyscloud_telephony_providers_edges_site_outbound_route" "example_%s" {
  site_id               = %q
  name                  = %q
  description           = %q
  classification_types  = %v
  enabled              = %v
  distribution         = %q
  external_trunk_base_ids = %v
}
`,
				strings.ToLower(routeMap["name"].(string)),
				siteID,
				routeMap["name"].(string),
				routeMap["description"].(string),
				routeMap["classification_types"],
				routeMap["enabled"].(bool),
				routeMap["distribution"].(string),
				routeMap["external_trunk_base_ids"])

			// Output the exact import command with the route ID
			fmt.Printf("terraform import genesyscloud_telephony_providers_edges_site_outbound_route.example_%s %s:%s\n",
				strings.ToLower(routeMap["name"].(string)),
				siteID,
				routeID)
		}
	}

	// Remove the outbound_routes from the site resource state
	delete(rawState, "outbound_routes")
	return rawState, nil
}
