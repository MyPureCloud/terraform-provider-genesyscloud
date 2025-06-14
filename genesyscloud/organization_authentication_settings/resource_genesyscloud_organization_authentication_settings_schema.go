package organization_authentication_settings

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_organization_authentication_settings_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the organization_authentication_settings resource.
3.  The datasource schema definitions for the organization_authentication_settings datasource.
4.  The resource exporter configuration for the organization_authentication_settings exporter.
*/
const ResourceType = "genesyscloud_organization_authentication_settings"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceOrganizationAuthenticationSettings())
	l.RegisterExporter(ResourceType, OrganizationAuthenticationSettingsExporter())
}

var passwordRequirements = &schema.Resource{
	Schema: map[string]*schema.Schema{
		`minimum_length`: {
			Description: "The minimum character length for passwords",
			Optional:    true,
			Type:        schema.TypeInt,
		},
		`minimum_digits`: {
			Description: "The minimum number of numerals (0-9) that must be included in passwords",
			Optional:    true,
			Type:        schema.TypeInt,
		},
		`minimum_letters`: {
			Description: "The minimum number of characters required for passwords",
			Optional:    true,
			Type:        schema.TypeInt,
		},
		`minimum_upper`: {
			Description: "The minimum number of upper case letters that must be included in passwords",
			Optional:    true,
			Type:        schema.TypeInt,
		},
		`minimum_lower`: {
			Description: "The minimum number of lower case letters that must be included in passwords",
			Optional:    true,
			Type:        schema.TypeInt,
		},
		`minimum_specials`: {
			Description: "The minimum number of special characters that must be included in passwords",
			Optional:    true,
			Type:        schema.TypeInt,
		},
		`minimum_age_seconds`: {
			Description: "Minimum age of the password (in seconds) before it can be changed",
			Optional:    true,
			Type:        schema.TypeInt,
		},
		`expiration_days`: {
			Description: "Length of time (in days) before a password must be changed",
			Optional:    true,
			Type:        schema.TypeInt,
		},
	},
}

var timeOutSettings = &schema.Resource{
	Schema: map[string]*schema.Schema{
		`enable_idle_token_timeout`: {
			Description: `Indicates whether the Token Timeout should be enabled or disabled.`,
			Required:    true,
			Type:        schema.TypeBool,
		},
		`idle_token_timeout_seconds`: {
			Description: `Token timeout length in seconds. Must be at least 5 minutes and 8 hours or less (if HIPAA is disabled) or 15 minutes or less (if HIPAA is enabled).`,
			Required:    true,
			Type:        schema.TypeInt,
		},
	},
}

// ResourceOrganizationAuthenticationSettings registers the genesyscloud_organization_authentication_settings resource with Terraform
func ResourceOrganizationAuthenticationSettings() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud organization authentication settings`,

		CreateContext: provider.CreateWithPooledClient(createOrganizationAuthenticationSettings),
		ReadContext:   provider.ReadWithPooledClient(readOrganizationAuthenticationSettings),
		UpdateContext: provider.UpdateWithPooledClient(updateOrganizationAuthenticationSettings),
		DeleteContext: provider.DeleteWithPooledClient(deleteOrganizationAuthenticationSettings),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`multifactor_authentication_required`: {
				Description: `Indicates whether multi-factor authentication is required.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`domain_allowlist_enabled`: {
				Description: `Indicates whether the domain allowlist is enabled.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`domain_allowlist`: {
				Description: `The list of domains that will be allowed to embed Genesys Cloud applications.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`ip_address_allowlist`: {
				Description: `The list of IP addresses that will be allowed to authenticate with Genesys Cloud. Warning: Changing these will result in only allowing specified ip Addresses to log in and will invalidate credentials with a different ip address`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`password_requirements`: {
				Description: `The password requirements for the organization.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        passwordRequirements,
			},
			`timeout_settings`: {
				Description: `the time out settings for the tokens`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        timeOutSettings,
			},
		},
	}
}

func OrganizationAuthenticationSettingsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllOrganizationAuthenticationSettings),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
		AllowZeroValues: []string{
			"timeout_settings.idle_token_timeout_seconds",
		},
	}
}
