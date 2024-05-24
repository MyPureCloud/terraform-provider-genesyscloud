package auth_role

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_auth_role_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the auth_role resource.
3.  The datasource schema definitions for the auth_role datasource.
4.  The resource exporter configuration for the auth_role exporter.
*/
const resourceName = "genesyscloud_auth_role"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceAuthRole())
	regInstance.RegisterDataSource(resourceName, DataSourceAuthRole())
	regInstance.RegisterExporter(resourceName, AuthRoleExporter())
}

var (
	rolePermPolicyCondOperands = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "Value type (USER | QUEUE | SCALAR | VARIABLE).",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"USER", "QUEUE", "SCALAR", "VARIABLE"}, false),
			},
			"queue_id": {
				Description: "Queue ID for QUEUE types.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"user_id": {
				Description: "User ID for USER types.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"value": {
				Description: "Value for operand. For USER or QUEUE types, use user_id or queue_id instead.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	rolePermPolicyCondTerms = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"variable_name": {
				Description: "Variable name being compared. This varies depending on the permission.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"operator": {
				Description:  "Operator type (EQ | IN | GE | GT | LE | LT).",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"EQ", "IN", "GE", "GT", "LE", "LT"}, false),
			},
			"operands": {
				Description: "Operands for this condition.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        rolePermPolicyCondOperands,
			},
		},
	}

	rolePermPolicyResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"domain": {
				Description: "Permission domain. e.g 'directory'",
				Type:        schema.TypeString,
				Required:    true,
			},
			"entity_name": {
				Description: "Permission entity or '*' for all. e.g. 'user'",
				Type:        schema.TypeString,
				Required:    true,
			},
			"action_set": {
				Description: "Actions allowed on the entity or '*' for all. e.g. 'add'",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    1,
			},
			"conditions": {
				Description: "Conditions specific to this resource. This is only applicable to some permission types.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"conjunction": {
							Description:  "Conjunction for condition terms (AND | OR).",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"AND", "OR"}, false),
						},
						"terms": {
							Description: "Terms of the condition.",
							Type:        schema.TypeSet,
							Required:    true,
							Elem:        rolePermPolicyCondTerms,
						},
					},
				},
			},
		},
	}
)

// ResourceAuthRole registers the genesyscloud_auth_role resource with Terraform
func ResourceAuthRole() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Authorization Role",

		CreateContext: provider.CreateWithPooledClient(createAuthRole),
		ReadContext:   provider.ReadWithPooledClient(readAuthRole),
		UpdateContext: provider.UpdateWithPooledClient(updateAuthRole),
		DeleteContext: provider.DeleteWithPooledClient(deleteAuthRole),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Role name. This cannot be modified for default roles.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Role description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"permissions": {
				Description: "General role permissions. e.g. 'group_creation'",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"permission_policies": {
				Description: "Role permission policies.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        rolePermPolicyResource,
			},
			"default_role_id": {
				Description: "Internal ID for an existing default role, e.g. 'employee'. This can be set to manage permissions on existing default roles.  Note: Changing the default_role_id attribute will cause this auth_role to be dropped and recreated with a new ID.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
			},
		},
	}
}

// AuthRoleExporter returns the resourceExporter object used to hold the genesyscloud_auth_role exporter's config
func AuthRoleExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthAuthRoles),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"permission_policies.conditions.terms.operands.queue_id": {RefType: "genesyscloud_routing_queue"},
			"permission_policies.conditions.terms.operands.user_id":  {RefType: "genesyscloud_user"},
		},
		RemoveIfMissing: map[string][]string{
			"permission_policies.conditions.terms.operands": {"queue_id", "user_id", "value"},
			"permission_policies.conditions.terms":          {"operands"},
			"permission_policies.conditions":                {"terms"},
		},
	}
}

// DataSourceAuthRole registers the genesyscloud_auth_role data source
func DataSourceAuthRole() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Roles. Select a role by name.`,
		ReadContext: provider.ReadWithPooledClient(DataSourceAuthRoleRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Role name.`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
