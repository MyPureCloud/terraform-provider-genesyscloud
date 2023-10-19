package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

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

func getAllAuthRoles(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		roles, _, getErr := authAPI.GetAuthorizationRoles(pageSize, pageNum, "", nil, "", "", "", nil, nil, false, nil)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of roles: %v", getErr)
		}

		if roles.Entities == nil || len(*roles.Entities) == 0 {
			break
		}

		for _, role := range *roles.Entities {
			resources[*role.Id] = &resourceExporter.ResourceMeta{Name: *role.Name}
		}
	}

	return resources, nil
}

func AuthRoleExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllAuthRoles),
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

func ResourceAuthRole() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Authorization Role",

		CreateContext: CreateWithPooledClient(createAuthRole),
		ReadContext:   ReadWithPooledClient(readAuthRole),
		UpdateContext: UpdateWithPooledClient(updateAuthRole),
		DeleteContext: DeleteWithPooledClient(deleteAuthRole),
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

func createAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	defaultRoleID := d.Get("default_role_id").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
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
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Reading role %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		role, resp, getErr := authAPI.GetAuthorizationRole(d.Id(), false, nil)
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read role %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read role %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceAuthRole())
		d.Set("name", *role.Name)

		if role.Description != nil {
			d.Set("description", *role.Description)
		} else {
			d.Set("description", nil)
		}

		if role.DefaultRoleId != nil {
			d.Set("default_role_id", *role.DefaultRoleId)
		} else {
			d.Set("default_role_id", nil)
		}

		if role.Permissions != nil {
			d.Set("permissions", lists.StringListToSet(*role.Permissions))
		} else {
			d.Set("permissions", nil)
		}

		if role.PermissionPolicies != nil {
			d.Set("permission_policies", flattenRolePermissionPolicies(*role.PermissionPolicies))
		} else {
			d.Set("permission_policies", nil)
		}

		log.Printf("Read role %s %s", d.Id(), *role.Name)
		return cc.CheckState()
	})
}

func updateAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	defaultRoleID := d.Get("default_role_id").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
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

	sdkConfig := meta.(*ProviderMeta).ClientConfig
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

	return WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		_, resp, err := authAPI.GetAuthorizationRole(d.Id(), false, nil)
		if err != nil {
			if IsStatus404(resp) {
				// role deleted
				log.Printf("Deleted role %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting role %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("Role %s still exists", d.Id()))
	})
}

func buildSdkRolePermissions(d *schema.ResourceData) *[]string {
	if permConfig, ok := d.GetOk("permissions"); ok {
		return lists.SetToStringList(permConfig.(*schema.Set))
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
			policy := platformclientv2.Domainpermissionpolicy{
				Domain:     &domain,
				EntityName: &entityName,
				ActionSet:  buildSdkPermPolicyActions(policyMap),
			}
			if conditions, ok := policyMap["conditions"]; ok {
				conditionsList := conditions.([]interface{})
				policy.ResourceConditionNode = buildSdkPermPolicyConditions(conditionsList)
			}
			sdkPolicies = append(sdkPolicies, policy)
		}
	}
	return &sdkPolicies
}

func buildSdkPermPolicyActions(policyAttrs map[string]interface{}) *[]string {
	if actions, ok := policyAttrs["action_set"]; ok {
		return lists.SetToStringList(actions.(*schema.Set))
	}
	return nil
}

func buildSdkPermPolicyConditions(conditions []interface{}) *platformclientv2.Domainresourceconditionnode {
	if len(conditions) > 0 {
		conditionAttrs := conditions[0].(map[string]interface{})
		conjunction := conditionAttrs["conjunction"].(string)
		terms := conditionAttrs["terms"].(*schema.Set).List()
		return &platformclientv2.Domainresourceconditionnode{
			Conjunction: &conjunction,
			Terms:       buildSdkPermPolicyCondTerms(terms),
		}
	}
	return nil
}

func buildSdkPermPolicyCondTerms(terms []interface{}) *[]platformclientv2.Domainresourceconditionnode {
	sdkTerms := make([]platformclientv2.Domainresourceconditionnode, len(terms))
	for i, term := range terms {
		termMap := term.(map[string]interface{})
		varName := termMap["variable_name"].(string)
		operator := termMap["operator"].(string)
		operands := termMap["operands"].(*schema.Set).List()
		sdkTerms[i] = platformclientv2.Domainresourceconditionnode{
			VariableName: &varName,
			Operator:     &operator,
			Operands:     buildSdkPermPolicyCondOperands(operands),
		}
	}
	return &sdkTerms
}

func buildSdkPermPolicyCondOperands(operands []interface{}) *[]platformclientv2.Domainresourceconditionvalue {
	sdkOperands := make([]platformclientv2.Domainresourceconditionvalue, len(operands))
	for i, operand := range operands {
		operandMap := operand.(map[string]interface{})
		varType := operandMap["type"].(string)

		sdkOperand := platformclientv2.Domainresourceconditionvalue{
			VarType: &varType,
		}
		switch varType {
		case "USER":
			value := operandMap["user_id"].(string)
			sdkOperand.User = &platformclientv2.User{Id: &value}
		case "QUEUE":
			value := operandMap["queue_id"].(string)
			sdkOperand.Queue = &platformclientv2.Queue{Id: &value}
		default:
			value := operandMap["value"].(string)
			sdkOperand.Value = &value
		}
		sdkOperands[i] = sdkOperand
	}
	return &sdkOperands
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
			policyMap["action_set"] = lists.StringListToSet(*sdkPolicy.ActionSet)
		}
		if sdkPolicy.ResourceConditionNode != nil {
			policyMap["conditions"] = flattenRoleConditionNode(*sdkPolicy.ResourceConditionNode)
		}
		policySet.Add(policyMap)
	}

	return policySet
}

func flattenRoleConditionNode(conditions platformclientv2.Domainresourceconditionnode) []interface{} {
	conditionMap := make(map[string]interface{})

	if conditions.Conjunction != nil {
		conditionMap["conjunction"] = *conditions.Conjunction
	}
	if conditions.Terms != nil {
		conditionMap["terms"] = flattenRoleConditionTerms(*conditions.Terms)
	}

	return []interface{}{conditionMap}
}

func flattenRoleConditionTerms(terms []platformclientv2.Domainresourceconditionnode) *schema.Set {
	termSet := schema.NewSet(schema.HashResource(rolePermPolicyCondTerms), []interface{}{})
	for _, term := range terms {
		termMap := make(map[string]interface{})
		if term.VariableName != nil {
			termMap["variable_name"] = *term.VariableName
		}
		if term.Operator != nil {
			termMap["operator"] = *term.Operator
		}
		if term.Operands != nil {
			termMap["operands"] = flattenRoleConditionOperands(*term.Operands)
		}
		termSet.Add(termMap)
	}
	return termSet
}

func flattenRoleConditionOperands(operands []platformclientv2.Domainresourceconditionvalue) *schema.Set {
	operandSet := schema.NewSet(schema.HashResource(rolePermPolicyCondOperands), []interface{}{})
	for _, operand := range operands {
		operandMap := make(map[string]interface{})
		if operand.VarType != nil {
			operandMap["type"] = *operand.VarType
			switch *operand.VarType {
			case "USER":
				if operand.User != nil {
					operandMap["user_id"] = *operand.User.Id
				}
			case "QUEUE":
				if operand.Queue != nil {
					operandMap["queue_id"] = *operand.Queue.Id
				}
			default:
				if operand.Value != nil {
					operandMap["value"] = *operand.Value
				}
			}
		}
		operandSet.Add(operandMap)
	}
	return operandSet
}

func getRoleID(defaultRoleID string, authAPI *platformclientv2.AuthorizationApi) (string, diag.Diagnostics) {
	const pageSize = 1
	const pageNum = 1
	roles, _, getErr := authAPI.GetAuthorizationRoles(pageSize, pageNum, "", nil, "", "", "", nil, []string{defaultRoleID}, false, nil)
	if getErr != nil {
		return "", diag.Errorf("Error requesting default role %s: %s", defaultRoleID, getErr)
	}
	if roles.Entities == nil || len(*roles.Entities) == 0 {
		return "", diag.Errorf("Default role not found: %s", defaultRoleID)
	}

	return *(*roles.Entities)[0].Id, nil
}

func GenerateAuthRoleResource(
	resourceID string,
	name string,
	description string,
	nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_auth_role" "%s" {
		name = "%s"
		description = "%s"
		%s
	}
	`, resourceID, name, description, strings.Join(nestedBlocks, "\n"))
}

func GenerateRolePermissions(permissions ...string) string {
	return fmt.Sprintf(`
		permissions = [%s]
	`, strings.Join(permissions, ","))
}

func GenerateRolePermPolicy(domain string, entityName string, actions ...string) string {
	return fmt.Sprintf(` permission_policies {
		domain = "%s"
		entity_name = "%s"
		action_set = [%s]
	}
	`, domain, entityName, strings.Join(actions, ","))
}
