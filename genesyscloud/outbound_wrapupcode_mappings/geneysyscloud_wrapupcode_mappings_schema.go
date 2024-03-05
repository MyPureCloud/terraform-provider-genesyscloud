package outbound_wrapupcode_mappings

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const resourceName = "genesyscloud_outbound_wrapupcodemappings"

var (
	flagsSchema = &schema.Schema{
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"CONTACT_UNCALLABLE", "NUMBER_UNCALLABLE", "RIGHT_PARTY_CONTACT"}, true),
	}

	mappingResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`wrapup_code_id`: {
				Description: `The wrap-up code identifier.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`flags`: {
				Description: `The set of wrap-up flags.`,
				Required:    true,
				Type:        schema.TypeSet,
				Elem:        flagsSchema,
			},
		},
	}
)

// SetRegistrar registers the resource objects and the exporter.  Note:  There is no datasource implementation
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(resourceName, ResourceOutboundWrapUpCodeMappings())
	l.RegisterExporter(resourceName, OutboundWrapupCodeMappingsExporter())
}

// OutboundWrapupCodeMappingsExporter() returns the exporter used for exporting the outbound wrapping codes
func OutboundWrapupCodeMappingsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getOutboundWrapupCodeMappings),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			`mappings.wrapup_code_id`: {
				RefType: `genesyscloud_routing_wrapupcode`,
			},
		},
		AllowEmptyArrays: []string{"default_set", "mappings.flags"},
	}
}

// ResourceOutboundWrapUpCodeMappings returns the schema definition for outbound wrappings
func ResourceOutboundWrapUpCodeMappings() *schema.Resource {
	return &schema.Resource{
		Description:   `Genesys Cloud Outbound Wrap-up Code Mappings`,
		CreateContext: provider.CreateWithPooledClient(createOutboundWrapUpCodeMappings),
		ReadContext:   provider.ReadWithPooledClient(readOutboundWrapUpCodeMappings),
		UpdateContext: provider.UpdateWithPooledClient(updateOutboundWrapUpCodeMappings),
		DeleteContext: provider.DeleteWithPooledClient(deleteOutboundWrapUpCodeMappings),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`default_set`: {
				Description: `The default set of wrap-up flags. These will be used if there is no entry for a given wrap-up code in the mapping.`,
				Required:    true,
				Type:        schema.TypeSet,
				Elem:        flagsSchema,
			},
			`mappings`: {
				Description: `A map from wrap-up code identifiers to a set of wrap-up flags.`,
				Optional:    true,
				Type:        schema.TypeSet,
				Elem:        mappingResource,
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
