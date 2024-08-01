package group_roles

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const resourceName = "genesyscloud_group_roles"

// SetRegistrar registers all the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(resourceName, ResourceGroupRoles())
	l.RegisterExporter(resourceName, GroupRolesExporter())
}

var (
	RoleAssignmentResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"role_id": {
				Description: "Role ID.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"division_ids": {
				Description: "Division IDs applied to this resource. If not set, the home division will be used. '*' may be set for all divisions.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
)

// ResourceGroupRoles registers the genesyscloud_group_roles resource with Terraform
func ResourceGroupRoles() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Group Roles maintains group role assignments.`,

		CreateContext: provider.CreateWithPooledClient(createGroupRoles),
		ReadContext:   provider.ReadWithPooledClient(readGroupRoles),
		UpdateContext: provider.UpdateWithPooledClient(updateGroupRoles),
		DeleteContext: provider.DeleteWithPooledClient(deleteGroupRoles),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"group_id": {
				Description: "Group ID that will be managed by this resource. Changing the group_id attribute for the groups_role object will cause the existing group_roles object to be dropped and recreated with a new ID",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"roles": {
				Description: "Roles and their divisions assigned to this group.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        RoleAssignmentResource,
			},
		},
	}
}

// GroupRolesExporter returns the resourceExporter object used to hold the genesyscloud_group_roles exporter's config
func GroupRolesExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllGroups),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"group_id":           {RefType: "genesyscloud_group"},
			"roles.role_id":      {RefType: "genesyscloud_auth_role"},
			"roles.division_ids": {RefType: "genesyscloud_auth_division", AltValues: []string{"*"}},
		},
		RemoveIfMissing: map[string][]string{
			"roles": {"role_id"},
		},
	}
}

// Duplicated this from the group package to break a cyclical dependency.  We should be asking ourselves why we are doing this at some point
func getAllGroups(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	groupsAPI := platformclientv2.NewGroupsApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		groups, resp, getErr := groupsAPI.GetGroups(pageSize, pageNum, nil, nil, "")
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get page of groups error: %s", getErr), resp)
		}

		if groups.Entities == nil || len(*groups.Entities) == 0 {
			break
		}

		for _, group := range *groups.Entities {
			resources[*group.Id] = &resourceExporter.ResourceMeta{Name: *group.Name}
		}
	}

	return resources, nil
}
