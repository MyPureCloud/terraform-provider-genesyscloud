package conversations_messaging_supportedcontent

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_supported_content_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the supported_content resource.
3.  The datasource schema definitions for the supported_content datasource.
4.  The resource exporter configuration for the supported_content exporter.
*/
const resourceName = "genesyscloud_conversations_messaging_supportedcontent"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceSupportedContent())
	regInstance.RegisterDataSource(resourceName, DataSourceSupportedContent())
	regInstance.RegisterExporter(resourceName, SupportedContentExporter())
}

var (
	mediaTypeResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`type`: {
				Description: `The media type string as defined by RFC 2046. You can define specific types such as 'image/jpeg', 'video/mpeg', or specify wild cards for a range of types, 'image/*', or all types '*/*'. See https://www.iana.org/assignments/media-types/media-types.xhtml for a list of registered media types.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	mediaTypeAccessResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`inbound`: {
				Description: `List of media types allowed for inbound messages from customers. If inbound messages from a customer contain media that is not in this list, the media will be dropped from the outbound message.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        mediaTypeResource,
			},
			`outbound`: {
				Description: `List of media types allowed for outbound messages to customers. If an outbound message is sent that contains media that is not in this list, the message will not be sent.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        mediaTypeResource,
			},
		},
	}

	mediaTypesResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`allow`: {
				Description: `Specify allowed media types for inbound and outbound messages. If this field is empty, all inbound and outbound media will be blocked.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        mediaTypeAccessResource,
			},
		},
	}
)

// ResourceSupportedContent registers the genesyscloud_conversations_messaging_supportedcontent resource with Terraform
func ResourceSupportedContent() *schema.Resource {

	return &schema.Resource{
		Description: `Genesys Cloud supported content`,

		CreateContext: provider.CreateWithPooledClient(createSupportedContent),
		ReadContext:   provider.ReadWithPooledClient(readSupportedContent),
		UpdateContext: provider.UpdateWithPooledClient(updateSupportedContent),
		DeleteContext: provider.DeleteWithPooledClient(deleteSupportedContent),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the supported content profile`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`media_types`: {
				Description: `Defines the allowable media that may be accepted for an inbound message or to be sent in an outbound message.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        mediaTypesResource,
			},
		},
	}
}

// SupportedContentExporter returns the resourceExporter object used to hold the genesyscloud_conversations_messaging_supportedcontent exporter's config
func SupportedContentExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthSupportedContents),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
	}
}

// DataSourceSupportedContent registers the genesyscloud_conversations_messaging_supportedcontent data source
func DataSourceSupportedContent() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud supported content data source. Select an supported content by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceSupportedContentRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `supported content name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
