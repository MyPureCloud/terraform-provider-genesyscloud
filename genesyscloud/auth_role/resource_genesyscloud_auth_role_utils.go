package auth_role

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func validatePermissionPolicy(proxy *authRoleProxy, policy platformclientv2.Domainpermissionpolicy) (*platformclientv2.APIResponse, error) {
	allowedPermissions, resp, err := proxy.getAllowedPermissions(*policy.Domain)
	if err != nil {
		return resp, fmt.Errorf("error requesting org permissions: %s", err)
	}
	if len(*allowedPermissions) == 0 {
		return resp, fmt.Errorf("domain %s not found", *policy.Domain)
	}

	if *policy.EntityName == "*" {
		return resp, nil
	}

	// Check entity type (e.g. callableTimeSet) exists in the map of allowed permissions
	if entityPermissions, ok := (*allowedPermissions)[*policy.EntityName]; ok {
		// Check if the policy actions exist for the given domain permission e.g. callableTimeSet: add
		for _, action := range *policy.ActionSet {
			if action == "*" && len(entityPermissions) >= 1 {
				break
			}

			var found bool
			for _, entityPermission := range entityPermissions {
				if action == *entityPermission.Action {
					// action found, move to next action
					found = true
					break
				}
			}
			if !found {
				return resp, fmt.Errorf("action %s not found for domain %s, entity name %s", action, *policy.Domain, *policy.EntityName)
			}
		}
		// All actions have been found, permission exists
		return resp, nil
	}

	return resp, fmt.Errorf("entity_name %s not found for domain %s", *policy.EntityName, *policy.Domain)
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

func GenerateDefaultAuthRoleDataSource(
	resourceID string,
	name string) string {
	return fmt.Sprintf(`data "genesyscloud_auth_role" "%s" {
		name = %s
	}
	`, resourceID, name)
}
