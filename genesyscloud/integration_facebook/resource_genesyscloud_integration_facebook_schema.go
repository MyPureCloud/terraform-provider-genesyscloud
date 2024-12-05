package integration_facebook

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_integration_facebook_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the integration_facebook resource.
3.  The datasource schema definitions for the integration_facebook datasource.
4.  The resource exporter configuration for the integration_facebook exporter.
*/
const ResourceType = "genesyscloud_integration_facebook"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceIntegrationFacebook())
	regInstance.RegisterDataSource(ResourceType, DataSourceIntegrationFacebook())
	regInstance.RegisterExporter(ResourceType, IntegrationFacebookExporter())
}

// ResourceIntegrationFacebook registers the genesyscloud_integration_facebook resource with Terraform
func ResourceIntegrationFacebook() *schema.Resource {

	return &schema.Resource{
		Description: `Genesys Cloud integration facebook`,

		CreateContext: provider.CreateWithPooledClient(createIntegrationFacebook),
		ReadContext:   provider.ReadWithPooledClient(readIntegrationFacebook),
		UpdateContext: provider.UpdateWithPooledClient(updateIntegrationFacebook),
		DeleteContext: provider.DeleteWithPooledClient(deleteIntegrationFacebook),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(8 * time.Minute),
			Read:   schema.DefaultTimeout(8 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the Facebook Integration`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`supported_content_id`: {
				Description: `The SupportedContent unique identifier associated with this integration.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`messaging_setting_id`: {
				Description: `The messaging Setting unique identifier associated with this integration.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`page_access_token`: {
				Description: `The long-lived Page Access Token of Facebook page.
			See https://developers.facebook.com/docs/facebook-login/access-tokens.
			Either pageAccessToken or userAccessToken should be provided.`,
				Optional: true,
				Type:     schema.TypeString,
			},
			`user_access_token`: {
				Description: `The short-lived User Access Token of the Facebook user logged into the Facebook app.
			See https://developers.facebook.com/docs/facebook-login/access-tokens.
			Either pageAccessToken or userAccessToken should be provided.`,
				Optional: true,
				Type:     schema.TypeString,
			},
			`page_id`: {
				Description: `The page Id of Facebook page. The pageId is required when userAccessToken is provided.`,
				Optional:    true,
				Type:        schema.TypeString,
				ForceNew:    true,
			},
			`app_id`: {
				Description: `The app Id of Facebook app. The appId is required when a customer wants to use their own approved Facebook app.`,
				Optional:    true,
				Type:        schema.TypeString,
				ForceNew:    true,
			},
			`app_secret`: {
				Description: `The app Secret of Facebook app. The appSecret is required when appId is provided.`,
				Optional:    true,
				Type:        schema.TypeString,
				ForceNew:    true,
			},
		},
	}
}

// IntegrationFacebookExporter returns the resourceExporter object used to hold the genesyscloud_integration_facebook exporter's config
func IntegrationFacebookExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthIntegrationFacebooks),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"messaging_setting_id": {
				RefType: "genesyscloud_conversations_messaging_settings",
			},
			"supported_content_id": {
				RefType: "genesyscloud_conversations_messaging_supportedcontent",
			},
		},
	}
}

// DataSourceIntegrationFacebook registers the genesyscloud_integration_facebook data source
func DataSourceIntegrationFacebook() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud integration facebook data source. Select an integration facebook by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceIntegrationFacebookRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `integration facebook name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
