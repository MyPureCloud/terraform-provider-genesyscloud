package genesyscloud

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/platformclientv2"
)

var (
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
		},
	}
)

func getAllAuthRoles(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDNameMap, diag.Diagnostics) {
	resources := make(map[string]string)
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		roles, _, getErr := authAPI.GetAuthorizationRoles(100, pageNum, "", nil, "", "", "", nil, nil, false, nil)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of roles: %v", getErr)
		}

		if roles.Entities == nil || len(*roles.Entities) == 0 {
			break
		}

		for _, role := range *roles.Entities {
			resources[*role.Id] = *role.Name
		}
	}

	return resources, nil
}

func authRoleExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllAuthRoles),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceAuthRole() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Authorization Role",

		CreateContext: createWithPooledClient(createAuthRole),
		ReadContext:   readWithPooledClient(readAuthRole),
		UpdateContext: updateWithPooledClient(updateAuthRole),
		DeleteContext: deleteWithPooledClient(deleteAuthRole),
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
				Description: "Internal ID for an existing default role, e.g. 'employee'. This can be set to manage permissions on existing default roles.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
			},
		},
	}
}

func createAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	defaultRoleID := d.Get("default_role_id").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Creating role %s", name)
	if defaultRoleID != "" {
		// Default roles must already exist, or they cannot be modified
		id, diagErr := getRoleID(defaultRoleID, authAPI)
		if diagErr != nil {
			return diagErr
		}
		d.SetId(id)
		return updateAuthRole(ctx, d, meta)
	}

	role, _, err := authAPI.PostAuthorizationRoles(platformclientv2.Domainorganizationrolecreate{
		Name:               &name,
		Description:        &description,
		Permissions:        buildSdkRolePermissions(d),
		PermissionPolicies: buildSdkRolePermPolicies(d),
	})
	if err != nil {
		return diag.Errorf("Failed to create role %s: %s", name, err)
	}

	d.SetId(*role.Id)
	log.Printf("Created role %s %s", name, *role.Id)
	return readAuthRole(ctx, d, meta)
}

func readAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Reading role %s", d.Id())

	role, resp, getErr := authAPI.GetAuthorizationRole(d.Id(), nil)
	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read role %s: %s", d.Id(), getErr)
	}

	d.Set("name", *role.Name)

	if role.Description != nil {
		d.Set("description", *role.Description)
	}

	if role.DefaultRoleId != nil {
		d.Set("default_role_id", *role.DefaultRoleId)
	}

	if role.Permissions != nil {
		d.Set("permissions", stringListToSet(*role.Permissions))
	}

	if role.PermissionPolicies != nil {
		d.Set("permission_policies", flattenRolePermissionPolicies(*role.PermissionPolicies))
	}
	log.Printf("Read role %s %s", d.Id(), *role.Name)
	return nil
}

func updateAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	defaultRoleID := d.Get("default_role_id").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Updating role %s", name)
	_, _, err := authAPI.PutAuthorizationRole(d.Id(), platformclientv2.Domainorganizationroleupdate{
		Name:               &name,
		Description:        &description,
		Permissions:        buildSdkRolePermissions(d),
		PermissionPolicies: buildSdkRolePermPolicies(d),
		DefaultRoleId:      &defaultRoleID,
	})
	if err != nil {
		return diag.Errorf("Failed to update role %s: %s", name, err)
	}

	log.Printf("Updated role %s", name)
	return readAuthRole(ctx, d, meta)
}

func deleteAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	defaultRoleID := d.Get("default_role_id").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	if defaultRoleID != "" {
		// Restore default roles to their default state instead of deleting them
		log.Printf("Restoring default role %s", name)
		id := d.Id()
		_, _, err := authAPI.PutAuthorizationRolesDefault([]platformclientv2.Domainorganizationrole{
			{
				Id: &id,
			},
		})
		if err != nil {
			return diag.Errorf("Failed to restore default role %s: %s", defaultRoleID, err)
		}
		return nil
	}

	log.Printf("Deleting role %s", name)
	_, err := authAPI.DeleteAuthorizationRole(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete role %s: %s", name, err)
	}
	log.Printf("Deleted role %s", name)
	return nil
}

func buildSdkRolePermissions(d *schema.ResourceData) *[]string {
	if permConfig, ok := d.GetOk("permissions"); ok {
		return setToStringList(permConfig.(*schema.Set))
	}
	return nil
}

func buildSdkRolePermPolicies(d *schema.ResourceData) *[]platformclientv2.Domainpermissionpolicy {
	var sdkPolicies []platformclientv2.Domainpermissionpolicy
	if configPolicies, ok := d.GetOk("permission_policies"); ok {
		policyList := configPolicies.(*schema.Set).List()
		for _, configPolicy := range policyList {
			policyMap := configPolicy.(map[string]interface{})
			domain := policyMap["domain"].(string)
			entityName := policyMap["entity_name"].(string)
			sdkPolicies = append(sdkPolicies, platformclientv2.Domainpermissionpolicy{
				Domain:     &domain,
				EntityName: &entityName,
				ActionSet:  buildSdkPermPolicyActions(policyMap),
			})
		}
	}
	return &sdkPolicies
}

func buildSdkPermPolicyActions(policyAttrs map[string]interface{}) *[]string {
	if actions, ok := policyAttrs["action_set"]; ok {
		return setToStringList(actions.(*schema.Set))
	}
	return nil
}

func flattenRolePermissionPolicies(policies []platformclientv2.Domainpermissionpolicy) *schema.Set {
	policySet := schema.NewSet(schema.HashResource(rolePermPolicyResource), []interface{}{})

	for _, sdkPolicy := range policies {
		policyMap := make(map[string]interface{})
		if sdkPolicy.Domain != nil {
			policyMap["domain"] = *sdkPolicy.Domain
		}
		if sdkPolicy.EntityName != nil {
			policyMap["entity_name"] = *sdkPolicy.EntityName
		}
		if sdkPolicy.ActionSet != nil {
			policyMap["action_set"] = stringListToSet(*sdkPolicy.ActionSet)
		}
		policySet.Add(policyMap)
	}

	return policySet
}

func getRoleID(defaultRoleID string, authAPI *platformclientv2.AuthorizationApi) (string, diag.Diagnostics) {
	roles, _, getErr := authAPI.GetAuthorizationRoles(1, 1, "", nil, "", "", "", nil, []string{defaultRoleID}, false, nil)
	if getErr != nil {
		return "", diag.Errorf("Error requesting default role %s: %s", defaultRoleID, getErr)
	}
	if roles.Entities == nil || len(*roles.Entities) == 0 {
		return "", diag.Errorf("Default role not found: %s", defaultRoleID)
	}

	return *(*roles.Entities)[0].Id, nil
}
