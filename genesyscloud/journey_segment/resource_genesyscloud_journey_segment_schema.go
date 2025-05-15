package journey_segment

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const ResourceType = "genesyscloud_journey_segment"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceJourneySegment())
	regInstance.RegisterDataSource(ResourceType, DataSourceJourneySegment())
	regInstance.RegisterExporter(ResourceType, JourneySegmentExporter())
}

var (
	journeySegmentSchema = map[string]*schema.Schema{
		"is_active": {
			Description: "Whether or not the segment is active.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"display_name": {
			Description: "The display name of the segment.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"description": {
			Description: "A description of the segment.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"color": {
			Description: "The hexadecimal color value of the segment.",
			Type:        schema.TypeString,
			Required:    true,
			ValidateFunc: validation.StringMatch(func() *regexp.Regexp {
				r, _ := regexp.Compile("^#[a-fA-F\\d]{6}$")
				return r
			}(), ""),
		},
		"should_display_to_agent": {
			Description: "Whether or not the segment should be displayed to agent/supervisor users.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"context": {
			Description: "The context of the segment.",
			Type:        schema.TypeSet,
			Optional:    true,
			MaxItems:    1,
			Elem:        contextResource,
		},
		"journey": {
			Description: "The pattern of rules defining the segment.",
			Type:        schema.TypeSet,
			Optional:    true,
			MaxItems:    1,
			Elem:        journeyResource,
		},
		"assignment_expiration_days": {
			Description:  "Time, in days, from when the segment is assigned until it is automatically unassigned.",
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntAtLeast(0),
		},
	}

	contextResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"patterns": {
				Description: "A list of one or more patterns to match.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        contextPatternResource,
			},
		},
	}

	contextPatternResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"criteria": {
				Description: "A list of one or more criteria to satisfy.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        entityTypeCriteriaResource,
			},
		},
	}

	entityTypeCriteriaResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Description:  "The criteria key.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"device.category", "device.type", "device.osFamily", "browser.family", "browser.lang", "browser.version", "mktCampaign.source", "mktCampaign.medium", "mktCampaign.name", "mktCampaign.term", "mktCampaign.content", "mktCampaign.clickId", "mktCampaign.network", "geolocation.countryName", "geolocation.locality", "geolocation.region", "geolocation.postalCode", "geolocation.country", "ipOrganization", "referrer.url", "referrer.medium", "referrer.hostname", "authenticated"}, false),
			},
			"values": {
				Description: "The criteria values.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"should_ignore_case": {
				Description: "Should criteria be case insensitive.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"operator": {
				Description:  "The comparison operator. Valid values: containsAll, containsAny, notContainsAll, notContainsAny, equal, notEqual, greaterThan, greaterThanOrEqual, lessThan, lessThanOrEqual, startsWith, endsWith.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "equal",
				ValidateFunc: validation.StringInSlice([]string{"containsAll", "containsAny", "notContainsAll", "notContainsAny", "equal", "notEqual", "greaterThan", "greaterThanOrEqual", "lessThan", "lessThanOrEqual", "startsWith", "endsWith"}, false),
			},
			"entity_type": {
				Description:  "The entity to match the pattern against.Valid values: visit.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"visit"}, false),
			},
		},
	}

	journeyResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"patterns": {
				Description: "A list of zero or more patterns to match.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        journeyPatternResource,
			},
		},
	}

	journeyPatternResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"criteria": {
				Description: "A list of one or more criteria to satisfy.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        criteriaResource,
			},
			"count": {
				Description: "The number of times the pattern must match.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"stream_type": {
				Description:  "The stream type for which this pattern can be matched on. Valid values: Web, App.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Web", "App" /*, "Custom", "Conversation"*/}, false), // Custom and Conversation seem not to be supported by the API despite the documentation (DEVENGSD-607)
			},
			"session_type": {
				Description:  "The session type for which this pattern can be matched on. Valid values: web, app.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"web", "app"}, false), // custom value seems not to be supported by the API despite the documentation
			},
			"event_name": {
				Description: "The name of the event for which this pattern can be matched on.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
			},
		},
	}

	criteriaResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "The criteria key.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.Any(
					validation.StringInSlice([]string{"eventName", "page.url", "page.title", "page.hostname", "page.domain", "page.fragment", "page.keywords", "page.pathname", "searchQuery", "page.queryString"}, false),
					validation.StringMatch(func() *regexp.Regexp {
						r, _ := regexp.Compile("attributes\\..*\\.value")
						return r
					}(), ""),
				),
			},
			"values": {
				Description: "The criteria values.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"should_ignore_case": {
				Description: "Should criteria be case insensitive.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"operator": {
				Description:  "The comparison operator.Valid values: containsAll, containsAny, notContainsAll, notContainsAny, equal, notEqual, greaterThan, greaterThanOrEqual, lessThan, lessThanOrEqual, startsWith, endsWith.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "equal",
				ValidateFunc: validation.StringInSlice([]string{"containsAll", "containsAny", "notContainsAll", "notContainsAny", "equal", "notEqual", "greaterThan", "greaterThanOrEqual", "lessThan", "lessThanOrEqual", "startsWith", "endsWith"}, false),
			},
		},
	}
)

func DataSourceJourneySegment() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Journey Segment. Select a journey segment by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceJourneySegmentRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Journey Segment name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func JourneySegmentExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllJourneySegments),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceJourneySegment() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Journey Segment",
		CreateContext: provider.CreateWithPooledClient(createJourneySegment),
		ReadContext:   provider.ReadWithPooledClient(readJourneySegment),
		UpdateContext: provider.UpdateWithPooledClient(updateJourneySegment),
		DeleteContext: provider.DeleteWithPooledClient(deleteJourneySegment),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema:        journeySegmentSchema,
	}
}
