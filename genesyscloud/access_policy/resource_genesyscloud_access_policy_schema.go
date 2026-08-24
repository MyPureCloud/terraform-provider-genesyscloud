package access_policy

import (
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesyscloud_access_policy_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the access_policy resource.
3.  The datasource schema definitions for the access_policy datasource.
4.  The resource exporter configuration for the access_policy exporter.
*/

const ResourceType = "genesyscloud_access_policy"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceAccessPolicy())
	regInstance.RegisterDataSource(ResourceType, DataSourceAccessPolicy())
	regInstance.RegisterExporter(ResourceType, AccessPolicyExporter())
}

// ResourceAccessPolicy registers the genesyscloud_access_policy resource with Terraform
func ResourceAccessPolicy() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Access Policy. Manages attribute-based access control (ABAC) policies 
that provide granular control over user permissions and system access.`,

		CreateContext: provider.CreateWithPooledClient(createAccessPolicy),
		ReadContext:   provider.ReadWithPooledClient(readAccessPolicy),
		UpdateContext: provider.UpdateWithPooledClient(updateAccessPolicy),
		DeleteContext: provider.DeleteWithPooledClient(deleteAccessPolicy),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the access policy.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "A description of what this policy does.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"target_resource": {
				Description: "The targeted resource to which the policy applies, in the form of domain:entity:action (e.g. 'authorization:role:add').",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"effect": {
				Description:  "The effect this policy has when conditions are met. Valid values: ALLOW, DENY.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"ALLOW", "DENY"}, false),
			},
			"subject_type": {
				Description:  "The type of subject the policy applies to. Valid values: ALL, USER, CLIENT, GROUP, TEAM.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"ALL", "USER", "CLIENT", "GROUP", "TEAM"}, false),
			},
			"subject_id": {
				Description: "The ID of the subject. Required when subject_type is not 'ALL'. May be computed by the API.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// The API returns subject_id as "all" when subject_type is "ALL" and no explicit ID was provided.
					// Suppress the diff in this case to avoid a perpetual plan.
					subjectType := d.Get("subject_type").(string)
					if new == "" && strings.EqualFold(old, subjectType) {
						return true
					}
					return false
				},
			},
			"condition_json": {
				Description: "The condition tree as a JSON string. Use jsonencode() to construct this value. The condition defines when the policy effect is applied based on attribute comparisons.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enabled": {
				Description: "Whether the policy is actively enforced. Defaults to false.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"apply_to_clients": {
				Description: "Whether the policy also applies to OAuth Client (service account) API calls. Defaults to false.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"preset_attributes_json": {
				Description: "A JSON string containing a map of preset attribute names and their typed values to use in policy condition evaluation.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

// AccessPolicyExporter returns the resourceExporter object used to hold the genesyscloud_access_policy exporter's config
func AccessPolicyExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAccessPolicies),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
		JsonEncodeAttributes: []string{
			"condition_json",
			"preset_attributes_json",
		},
	}
}

// DataSourceAccessPolicy registers the genesyscloud_access_policy data source
func DataSourceAccessPolicy() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Access Policies. Select a policy by name.`,
		ReadContext: provider.ReadWithPooledClient(dataSourceAccessPolicyRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the access policy.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
