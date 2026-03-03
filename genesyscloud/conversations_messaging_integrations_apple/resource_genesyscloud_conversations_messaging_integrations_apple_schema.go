package conversations_messaging_integrations_apple

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

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
const resourceName = "genesyscloud_conversations_messaging_integrations_apple"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceConversationsMessagingIntegrationsApple())
	regInstance.RegisterDataSource(resourceName, DataSourceConversationsMessagingIntegrationsApple())
	regInstance.RegisterExporter(resourceName, ConversationsMessagingIntegrationsAppleExporter())
}

// ResourceAppleIntegration registers the genesyscloud_apple_integration resource with Terraform
func ResourceConversationsMessagingIntegrationsApple() *schema.Resource {

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
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"3DS",
						"creditCard",
						"debitCard",
					}, false),
				},
			},
			`supported_payment_networks`: {
				Description: `The payment networks supported by the merchant.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"amex",
						"discover",
						"jcb",
						"masterCard",
						"privateLabel",
						"visa",
					}, false),
				},
			},
			`payment_certificate_credential_id`: {
				Description: `The Genesys credentialId the payment certificates are stored under. Must be a valid and existing credential ID created via /api/v2/integrations/credentials endpoint. See API documentation: https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-integrations-credentials. Example payload: {"type": "applePayCertificate", "credentialFields": {"merchantKey": "base-64-key", "merchantCertificate": "base-64-cert"}}`,
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

	return &schema.Resource{
		Description: `Genesys Cloud apple integration`,

		CreateContext: provider.CreateWithPooledClient(createConversationsMessagingIntegrationsApple),
		ReadContext:   provider.ReadWithPooledClient(readConversationsMessagingIntegrationsApple),
		UpdateContext: provider.UpdateWithPooledClient(updateConversationsMessagingIntegrationsApple),
		DeleteContext: provider.DeleteWithPooledClient(deleteConversationsMessagingIntegrationsApple),
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
			`supported_content_id`: {
				Description: `The ID of the supported content profile configured for this integration. If not set, the default supported content profile will be used.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			`messaging_setting_id`: {
				Description: `The ID of the messaging setting configured for this integration`,
				Optional:    true,
				Type:        schema.TypeString,
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
		},
	}
}

// AppleIntegrationExporter returns the resourceExporter object used to hold the genesyscloud_apple_integration exporter's config
func ConversationsMessagingIntegrationsAppleExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllConversationsMessagingIntegrationsApple),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"supported_content_id":                        {RefType: "genesyscloud_conversations_messaging_supportedcontent"},
			"messaging_setting_id":                        {RefType: "genesyscloud_conversations_messaging_settings"},
			"apple_pay.payment_certificate_credential_id": {RefType: "genesyscloud_integration_credential"},
		},
	}
}

// DataSourceAppleIntegration registers the genesyscloud_apple_integration data source
func DataSourceConversationsMessagingIntegrationsApple() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud apple integration data source. Select an apple integration by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceConversationsMessagingIntegrationsAppleRead),
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
