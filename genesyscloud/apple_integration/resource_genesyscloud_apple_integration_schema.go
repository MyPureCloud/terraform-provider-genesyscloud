package apple_integration

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_apple_integration_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the apple_integration resource.
3.  The datasource schema definitions for the apple_integration datasource.
4.  The resource exporter configuration for the apple_integration exporter.
*/
const resourceName = "genesyscloud_apple_integration"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceAppleIntegration())
	regInstance.RegisterDataSource(resourceName, DataSourceAppleIntegration())
	regInstance.RegisterExporter(resourceName, AppleIntegrationExporter())
}

// ResourceAppleIntegration registers the genesyscloud_apple_integration resource with Terraform
func ResourceAppleIntegration() *schema.Resource {
	mediaTypeResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`type`: {
				Description: `The media type string as defined by RFC 2046. You can define specific types such as 'image/jpeg', 'video/mpeg', or specify wild cards for a range of types, 'image/*', or all types '*/*'. See https://www.iana.org/assignments/media-types/media-types.xhtml for a list of registered media types.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	mediaTypeAccessResource := &schema.Resource{
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

	mediaTypesResource := &schema.Resource{
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

	supportedContentReferenceResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The SupportedContent profile name`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`media_types`: {
				Description: `Media types definition for the supported content`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        mediaTypesResource,
			},
		},
	}

	inboundOnlySettingResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`inbound`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	storySettingResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`mention`: {
				Description: `Setting relating to Story Mentions`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        inboundOnlySettingResource,
			},
			`reply`: {
				Description: `Setting relating to Story Replies`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        inboundOnlySettingResource,
			},
		},
	}

	contentSettingResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`story`: {
				Description: `Settings relating to facebook and instagram stories feature`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        storySettingResource,
			},
		},
	}

	settingDirectionResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`inbound`: {
				Description: `Status for the Inbound Direction`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`outbound`: {
				Description: `Status for the Outbound Direction`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	typingSettingResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`on`: {
				Description: `Should typing indication Events be sent`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        settingDirectionResource,
			},
		},
	}

	eventSettingResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`typing`: {
				Description: `Settings regarding typing events`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        typingSettingResource,
			},
		},
	}

	messagingSettingReferenceResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The messaging Setting profile name`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`updated_by_id`: {
				Description: `User reference that modified this Setting`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`content`: {
				Description: `Settings relating to message contents`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        contentSettingResource,
			},
			`event`: {
				Description: `Settings relating to events which may occur`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        eventSettingResource,
			},
		},
	}

	errorBodyResource := &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}

	appleIMessageAppResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`application_name`: {
				Description: `Application Name.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`application_id`: {
				Description: `Application ID.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`bundle_id`: {
				Description: `Bundle ID.`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}

	appleAuthenticationResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`oauth_client_id`: {
				Description: `The Apple Messages for Business OAuth client id.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`oauth_client_secret`: {
				Description: `The Apple Messages for Business OAuth client secret.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`oauth_token_url`: {
				Description: `The Apple Messages for Business token URL.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`oauth_scope`: {
				Description: `The Apple Messages for Business OAuth scope.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	applePayResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`store_name`: {
				Description: `The name of the store.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`merchant_id`: {
				Description: `The stores merchant identifier.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`domain_name`: {
				Description: `The domain name associated with the merchant account.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`payment_capabilities`: {
				Description: `The payment capabilities supported by the merchant.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`supported_payment_networks`: {
				Description: `The payment networks supported by the merchant.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`payment_certificate_credential_id`: {
				Description: `The Genesys credentialId the payment certificates are stored under.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`payment_gateway_url`: {
				Description: `The url used to process payments.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`fallback_url`: {
				Description: `The url opened in a web browser if the customers device is unable to make payments using Apple Pay.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`shipping_method_update_url`: {
				Description: `The url called when the customer changes the shipping method.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`shipping_contact_update_url`: {
				Description: `The url called when the customer changes their shipping address information.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`payment_method_update_url`: {
				Description: `The url called when the customer changes their payment method.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`order_tracking_url`: {
				Description: `The url called after completing the order to update the order information in your system`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	appleIdentityResolutionConfigResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`division_id`: {
				Description: `The division to use when performing identity resolution.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`resolve_identities`: {
				Description: `Whether the channel should resolve identities`,
				Required:    true,
				Type:        schema.TypeBool,
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud apple integration`,

		CreateContext: provider.CreateWithPooledClient(createAppleIntegration),
		ReadContext:   provider.ReadWithPooledClient(readAppleIntegration),
		UpdateContext: provider.UpdateWithPooledClient(updateAppleIntegration),
		DeleteContext: provider.DeleteWithPooledClient(deleteAppleIntegration),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the Apple messaging integration.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`supported_content`: {
				Description: `Defines the SupportedContent profile configured for an integration`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        supportedContentReferenceResource,
			},
			`messaging_setting`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        messagingSettingReferenceResource,
			},
			`messages_for_business_id`: {
				Description: `The Apple Messages for Business Id for the Apple messaging integration.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`business_name`: {
				Description: `The name of the business.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`logo_url`: {
				Description: `The url of the businesses logo.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`status`: {
				Description: `The status of the Apple Integration`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`recipient_id`: {
				Description: `The recipient associated to the Apple messaging Integration. This recipient is used to associate a flow to an integration`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`create_status`: {
				Description: `Status of asynchronous create operation`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`create_error`: {
				Description: `Error information returned, if createStatus is set to Error`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        errorBodyResource,
			},
			`apple_i_message_app`: {
				Description: `Interactive Application (iMessage App) Settings.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        appleIMessageAppResource,
			},
			`apple_authentication`: {
				Description: `The Apple Messages for Business authentication setting.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        appleAuthenticationResource,
			},
			`apple_pay`: {
				Description: `Apple Pay settings.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        applePayResource,
			},
			`identity_resolution`: {
				Description: `The configuration to control identity resolution.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        appleIdentityResolutionConfigResource,
			},
		},
	}
}

// AppleIntegrationExporter returns the resourceExporter object used to hold the genesyscloud_apple_integration exporter's config
func AppleIntegrationExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthAppleIntegrations),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"supported_content.id":           {RefType: "genesyscloud_conversations_messaging_supportedcontent"},
			"messaging_setting.id":           {RefType: "genesyscloud_conversations_messaging_settings"},
			"recipient_id":                   {RefType: "genesyscloud_routing_message_recipient"},
			"identity_resolution.division_id": {RefType: "genesyscloud_auth_division"},
		},
		ExcludedAttributes: []string{
			"create_status",
			"create_error",
		},
	}
}

// DataSourceAppleIntegration registers the genesyscloud_apple_integration data source
func DataSourceAppleIntegration() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud apple integration data source. Select an apple integration by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceAppleIntegrationRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `apple integration name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
