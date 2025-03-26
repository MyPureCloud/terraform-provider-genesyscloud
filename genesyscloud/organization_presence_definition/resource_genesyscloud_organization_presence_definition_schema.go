package organization_presence_definition

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_organization_presence_definition_schema.go holds four functions within it:

1.  The registration code that registers the Resource and Exporter for the package.
2.  The resource schema definitions for the organization_presence_definition resource.
4.  The resource exporter configuration for the organization_presence_definition exporter.
*/
const ResourceType = "genesyscloud_organization_presence_definition"

// SetRegistrar registers all of the resources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceOrganizationPresenceDefinition())
	regInstance.RegisterExporter(ResourceType, OrganizationPresenceDefinitionExporter())
}

var validLanguageLabels = []string{
	"ar",
	"cs",
	"da",
	"de",
	"en",
	"en_US",
	"es",
	"fi",
	"fr",
	"he",
	"hi",
	"it",
	"ja",
	"ko",
	"nl",
	"no",
	"pl",
	"pt_BR",
	"pt_PT",
	"ru",
	"sv",
	"th",
	"tr",
	"uk",
	"zh_CN",
	"zh_TW",
}

var validSystemPresences = []string{
	"Available",
	"Away",
	"Break",
	"Busy",
	"Meal",
	"Meeting",
	"Training",
}

// ResourceOrganizationPresenceDefinition registers the genesyscloud_organization_presence_definition resource with Terraform
func ResourceOrganizationPresenceDefinition() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud organization presence definition`,

		CreateContext: provider.CreateWithPooledClient(createOrganizationPresenceDefinition),
		ReadContext:   provider.ReadWithPooledClient(readOrganizationPresenceDefinition),
		UpdateContext: provider.UpdateWithPooledClient(updateOrganizationPresenceDefinition),
		DeleteContext: provider.DeleteWithPooledClient(deleteOrganizationPresenceDefinition),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`language_labels`: {
				Description:      `The localized language labels for the presence definition. Valid labels: ` + strings.Join(validLanguageLabels, `, `),
				Type:             schema.TypeMap,
				Required:         true,
				Elem:             &schema.Schema{Type: schema.TypeString},
				ValidateDiagFunc: StringInMap(validLanguageLabels, true),
			},
			`system_presence`: {
				Description:  `System presence to create presence definition for. Once presence definition is created, this cannot be changed. Valid presences: ` + strings.Join(validSystemPresences, `, `),
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(validSystemPresences, true),
			},
			`division_id`: {
				Description: `The division to which the presence definition will belong. If not set, the presence definition will apply to all divisions.`,
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			`deactivated`: {
				Description: `If true, the presence definition is not active. If not set, the presence definition defaults to active.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
		},
	}
}

// Custom validator for language_labels
func StringInMap(valid []string, ignoreCase bool) schema.SchemaValidateDiagFunc {
	// Create the regular expression pattern
	pattern := strings.Join(valid, "|")
	if ignoreCase {
		pattern = fmt.Sprintf(`(?i)^(%s)$`, pattern)
	} else {
		pattern = fmt.Sprintf(`^(%s)$`, pattern)
	}

	return validation.MapKeyMatch(
		regexp.MustCompile(pattern),
		fmt.Sprintf(`expected key to be one of ["%s"], got`, strings.Join(valid, `", "`)),
	)
}

// OrganizationPresenceDefinitionExporter returns the resourceExporter object used to hold the genesyscloud_organization_presence_definition exporter's config
func OrganizationPresenceDefinitionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthOrganizationPresenceDefinitions),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
		},
	}
}
