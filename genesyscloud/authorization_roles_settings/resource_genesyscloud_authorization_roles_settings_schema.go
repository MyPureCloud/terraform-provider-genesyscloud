package authorization_roles_settings

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesyscloud_authorization_roles_settings_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the authorization_roles_settings resource.
3.  The datasource schema definitions for the authorization_roles_settings datasource.
4.  The resource exporter configuration for the authorization_roles_settings exporter.
*/
const ResourceType = "genesyscloud_authorization_roles_settings"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceAuthorizationRolesSettings())
	l.RegisterDataSource(ResourceType, DataSourceAuthorizationRolesSettings())
	l.RegisterExporter(ResourceType, AuthorizationRolesSettingsExporter())
}

// ResourceAuthorizationRolesSettings registers the genesyscloud_authorization_roles_settings resource with Terraform
//
// Every attribute is Optional+Computed on purpose. The underlying API exposes only a PUT
// (full replace), so omitting an attribute from the request body would reset it. Marking the
// attributes Computed lets Terraform carry the organization's current value for anything the
// practitioner has not configured, and the update path merges those values over a fresh read
// before writing (see buildAuthorizationRolesSettingsFromResourceData). That keeps a config
// which manages only one attribute from silently resetting the others.
func ResourceAuthorizationRolesSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud organization authorization role settings. This is a singleton resource; only one instance exists per organization. The underlying API supports only read and replace, so destroying this resource removes it from Terraform state without changing the organization's settings.",

		CreateContext: provider.CreateWithPooledClient(createAuthorizationRolesSettings),
		ReadContext:   provider.ReadWithPooledClient(readAuthorizationRolesSettings),
		UpdateContext: provider.UpdateWithPooledClient(updateAuthorizationRolesSettings),
		DeleteContext: provider.DeleteWithPooledClient(deleteAuthorizationRolesSettings),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"backfill_enabled": {
				Description: "Indicates whether the organization is opted in to role backfills.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"authorization_grant_division_aware": {
				Description: "Indicates whether division aware enforcement is enabled for authorization grant add and delete.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"genesys_org_policy_bypass": {
				Description: "Indicates whether attribute-based access control policy enforcement is skipped when the organization is on the Genesys bypass list. Genesys Cloud enables this by default. Setting it to `false` stops paired Genesys Services organizations and Genesys CIDR ranges from bypassing ABAC policies, which can block Genesys support access to this organization.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// DataSourceAuthorizationRolesSettings registers the genesyscloud_authorization_roles_settings data source with Terraform
func DataSourceAuthorizationRolesSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud organization authorization role settings. This is a singleton, so no arguments are required.",
		ReadContext: provider.ReadWithPooledClient(dataSourceAuthorizationRolesSettingsRead),
		Schema: map[string]*schema.Schema{
			"backfill_enabled": {
				Description: "Indicates whether the organization is opted in to role backfills.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"authorization_grant_division_aware": {
				Description: "Indicates whether division aware enforcement is enabled for authorization grant add and delete.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"genesys_org_policy_bypass": {
				Description: "Indicates whether attribute-based access control policy enforcement is skipped when the organization is on the Genesys bypass list.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func AuthorizationRolesSettingsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthorizationRolesSettings),
		IsSingleton:      true,
		ExportId:         ResourceType,
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
	}
}
