package genesyscloud

import (
	"context"
	"log"

	"github.com/MyPureCloud/platform-client-sdk-go/platformclientv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func resourceAuthRole() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Authorization Role",

		CreateContext: createAuthRole,
		UpdateContext: updateAuthRole,
		ReadContext:   readAuthRole,
		DeleteContext: deleteAuthRole,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Role name.",
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
		},
	}
}

func createAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	authAPI := platformclientv2.NewAuthorizationApi()

	log.Printf("Creating role %s", name)
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

	return readAuthRole(ctx, d, meta)
}

func readAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	authAPI := platformclientv2.NewAuthorizationApi()

	role, _, getErr := authAPI.GetAuthorizationRole(d.Id(), nil)
	if getErr != nil {
		return diag.Errorf("Failed to read role %s: %s", d.Id(), getErr)
	}

	d.Set("name", *role.Name)

	if role.Description != nil {
		d.Set("description", *role.Description)
	}

	if role.Permissions != nil {
		d.Set("permissions", stringListToSet(*role.Permissions))
	}

	if role.PermissionPolicies != nil {
		d.Set("permission_policies", flattenRolePermissionPolicies(*role.PermissionPolicies))
	}
	return nil
}

func updateAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	authAPI := platformclientv2.NewAuthorizationApi()

	log.Printf("Updating role %s", name)
	role, _, err := authAPI.PutAuthorizationRole(d.Id(), platformclientv2.Domainorganizationroleupdate{
		Name:               &name,
		Description:        &description,
		Permissions:        buildSdkRolePermissions(d),
		PermissionPolicies: buildSdkRolePermPolicies(d),
	})
	if err != nil {
		return diag.Errorf("Failed to create role %s: %s", name, err)
	}

	d.SetId(*role.Id)

	return readAuthRole(ctx, d, meta)
}

func deleteAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	authAPI := platformclientv2.NewAuthorizationApi()

	log.Printf("Deleting role %s", name)
	_, err := authAPI.DeleteAuthorizationRole(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete role %s: %s", name, err)
	}
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
