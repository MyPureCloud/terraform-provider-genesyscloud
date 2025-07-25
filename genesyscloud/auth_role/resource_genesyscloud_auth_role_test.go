package auth_role

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	lists "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func TestAccResourceAuthRoleDefault(t *testing.T) {
	var (
		roleResourceLabel2    = "auth-role2"
		roleDesc1             = "Terraform test role"
		perm1                 = "group_creation"
		directoryDom          = "directory"
		userEntity            = "user"
		addAction             = "add"
		viewAction            = "view"
		defaultRoleName       = "Trusted External User"
		defaultRoleID         = "trustedUser"
		authDom               = "authorization"
		orgTrusteeGroupEntity = "orgTrusteeGroup"
		orgTrusteeUserEntity  = "orgTrusteeUser"
		orgTrustorEntity      = "orgTrustor"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Modify default role
				Config: GenerateAuthRoleResource(
					roleResourceLabel2,
					defaultRoleName,
					roleDesc1,
					"default_role_id = "+strconv.Quote(defaultRoleID),
					GenerateRolePermissions(strconv.Quote(perm1)),
					GenerateRolePermPolicy(directoryDom, userEntity, strconv.Quote(addAction)),
					// Keep existing permissions on default role
					GenerateRolePermPolicy(authDom, orgTrusteeGroupEntity, strconv.Quote(viewAction)),
					GenerateRolePermPolicy(authDom, orgTrusteeUserEntity, strconv.Quote(viewAction)),
					GenerateRolePermPolicy(authDom, orgTrustorEntity, strconv.Quote(viewAction)),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_auth_role."+roleResourceLabel2, "name", defaultRoleName),
					resource.TestCheckResourceAttr("genesyscloud_auth_role."+roleResourceLabel2, "description", roleDesc1),
					resource.TestCheckResourceAttr("genesyscloud_auth_role."+roleResourceLabel2, "default_role_id", defaultRoleID),
					// New permissions
					validateRolePermissions("genesyscloud_auth_role."+roleResourceLabel2, perm1),
					validatePermissionPolicyTest("genesyscloud_auth_role."+roleResourceLabel2, directoryDom, userEntity, addAction),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_auth_role." + roleResourceLabel2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyRolesDestroyed,
	})
}

func TestAccResourceAuthRoleBasic(t *testing.T) {
	var (
		roleResourceLabel1 = "auth-role1"
		roleName1          = "Terraform Role-" + uuid.NewString()
		roleDesc1          = "Terraform test role"
		roleDesc2          = "Terraform test role updated"
		perm1              = "group_creation"
		perm2              = "admin"
		directoryDom       = "directory"
		userEntity         = "user"
		groupEntity        = "group"
		allAction          = "*"
		addAction          = "add"
		editAction         = "edit"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateAuthRoleResource(
					roleResourceLabel1,
					roleName1,
					roleDesc1,
					GenerateRolePermissions(strconv.Quote(perm1)),
					GenerateRolePermPolicy(directoryDom, userEntity, strconv.Quote(addAction)),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_auth_role."+roleResourceLabel1, "name", roleName1),
					resource.TestCheckResourceAttr("genesyscloud_auth_role."+roleResourceLabel1, "description", roleDesc1),
					validateRolePermissions("genesyscloud_auth_role."+roleResourceLabel1, perm1),
					validatePermissionPolicyTest("genesyscloud_auth_role."+roleResourceLabel1, directoryDom, userEntity, addAction),
				),
			},
			{
				// Update
				Config: GenerateAuthRoleResource(
					roleResourceLabel1,
					roleName1,
					roleDesc2,
					GenerateRolePermissions(strconv.Quote(perm1), strconv.Quote(perm2)),
					GenerateRolePermPolicy(directoryDom, userEntity, strconv.Quote(allAction)),
					GenerateRolePermPolicy(directoryDom, groupEntity, strconv.Quote(addAction), strconv.Quote(editAction)),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_auth_role."+roleResourceLabel1, "name", roleName1),
					resource.TestCheckResourceAttr("genesyscloud_auth_role."+roleResourceLabel1, "description", roleDesc2),
					validateRolePermissions("genesyscloud_auth_role."+roleResourceLabel1, perm1, perm2),
					validatePermissionPolicyTest("genesyscloud_auth_role."+roleResourceLabel1, directoryDom, userEntity, allAction),
					validatePermissionPolicyTest("genesyscloud_auth_role."+roleResourceLabel1, directoryDom, groupEntity, addAction, editAction),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_auth_role." + roleResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyRolesDestroyed,
	})
}

func TestAccResourceAuthRoleConditions(t *testing.T) {
	var (
		roleResourceLabel1  = "auth-role1"
		queueResourceLabel1 = "queue-resource1"
		queueName1          = "Terraform Queue-" + uuid.NewString()
		roleName1           = "Terraform Role-" + uuid.NewString()
		roleDesc1           = "Terraform test condition role"
		qualityDom          = "quality"
		calibrationEntity   = "calibration"
		addAction           = "add"
		conjAnd             = "AND"
		varNameMedia        = "Conversation.mediaType"
		varNameQueue        = "Conversation.queues"
		opEq                = "EQ"
		typeScalar          = "SCALAR"
		typeQueue           = "QUEUE"
		valueCall           = "CALL"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with a scalar condition
				Config: GenerateAuthRoleResource(
					roleResourceLabel1,
					roleName1,
					roleDesc1,
					generateRolePermPolicyCondition(
						qualityDom,
						calibrationEntity,
						addAction,
						conjAnd,
						generateRolePermPolicyCondTerm(
							varNameMedia,
							opEq,
							generateRoleCondValue(typeScalar, "value", strconv.Quote(valueCall)),
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					validatePermPolicyCondition(
						"genesyscloud_auth_role."+roleResourceLabel1,
						qualityDom,
						calibrationEntity,
						conjAnd,
						varNameMedia,
						opEq,
						typeScalar,
						valueCall,
					),
				),
			},
			{
				// Create a queue and update with a queue condition
				Config: routingQueue.GenerateRoutingQueueResourceBasic(queueResourceLabel1, queueName1) +
					GenerateAuthRoleResource(
						roleResourceLabel1,
						roleName1,
						roleDesc1,
						generateRolePermPolicyCondition(
							qualityDom,
							calibrationEntity,
							addAction,
							conjAnd,
							generateRolePermPolicyCondTerm(
								varNameQueue,
								opEq,
								generateRoleCondValue(typeQueue, "queue_id", "genesyscloud_routing_queue."+queueResourceLabel1+".id"),
							),
						),
					),
				Check: resource.ComposeTestCheckFunc(
					validatePermPolicyCondition(
						"genesyscloud_auth_role."+roleResourceLabel1,
						qualityDom,
						calibrationEntity,
						conjAnd,
						varNameQueue,
						opEq,
						typeQueue,
						"genesyscloud_routing_queue."+queueResourceLabel1),
				),
			},
			{
				// Queue condition without setting a queue_id
				Config: GenerateAuthRoleResource(
					roleResourceLabel1,
					roleName1,
					roleDesc1,
					generateRolePermPolicyCondition(
						qualityDom,
						calibrationEntity,
						addAction,
						conjAnd,
						generateRolePermPolicyCondTerm(
							varNameQueue,
							opEq,
							fmt.Sprintf(`
								operands {
									type  = "%s"
								}
								`, typeQueue),
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					validatePermPolicyCondition(
						"genesyscloud_auth_role."+roleResourceLabel1,
						qualityDom,
						calibrationEntity,
						conjAnd,
						varNameQueue,
						opEq,
						typeQueue,
						""),
				),
			},
			{
				// User condition without setting a user_id
				Config: GenerateAuthRoleResource(
					roleResourceLabel1,
					roleName1,
					roleDesc1,
					generateRolePermPolicyCondition(
						qualityDom,
						calibrationEntity,
						addAction,
						conjAnd,
						generateRolePermPolicyCondTerm(
							varNameQueue,
							opEq,
							`operands {
								type  = "USER"
							}`,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					validatePermPolicyCondition(
						"genesyscloud_auth_role."+roleResourceLabel1,
						qualityDom,
						calibrationEntity,
						conjAnd,
						varNameQueue,
						opEq,
						"USER",
						""),
				),
			},
			{
				// VARIABLE condition without setting a value
				Config: GenerateAuthRoleResource(
					roleResourceLabel1,
					roleName1,
					roleDesc1,
					generateRolePermPolicyCondition(
						"analytics",
						"userObservation",
						"*",
						conjAnd,
						generateRolePermPolicyCondTerm(
							varNameQueue,
							opEq,
							`operands {
								type  = "VARIABLE"
							}`,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					validatePermPolicyCondition(
						"genesyscloud_auth_role."+roleResourceLabel1,
						"analytics",
						"userObservation",
						conjAnd,
						varNameQueue,
						opEq,
						"VARIABLE",
						""),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_auth_role." + roleResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyRolesDestroyed,
	})
}

func generateRolePermPolicyCondition(domain string, entityName string, action string, conj string, terms ...string) string {
	return fmt.Sprintf(` permission_policies {
		domain = "%s"
		entity_name = "%s"
		action_set = ["%s"]
		conditions {
			conjunction = "%s"
			%s
		}
	}
	`, domain, entityName, action, conj, strings.Join(terms, "\n"))
}

func generateRolePermPolicyCondTerm(varName string, op string, operands ...string) string {
	return fmt.Sprintf(`
	terms {
		variable_name = "%s"
		operator      = "%s"
		%s
	}
	`, varName, op, strings.Join(operands, "\n"))
}

func generateRoleCondValue(varType string, attr string, val string) string {
	return fmt.Sprintf(`
	operands {
		type  = "%s"
		%s = %s
	}
	`, varType, attr, val)
}

func testVerifyRolesDestroyed(state *terraform.State) error {
	authAPI := platformclientv2.NewAuthorizationApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_auth_role" {
			continue
		}

		if rs.Primary.Attributes["default_role_id"] != "" {
			// We do not delete default roles
			continue
		}

		role, resp, err := authAPI.GetAuthorizationRole(rs.Primary.ID, false, nil)
		if role != nil {
			return fmt.Errorf("Role (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Role not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All roles destroyed
	return nil
}

func validateRolePermissions(roleResourcePath string, permissions ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		roleResource, ok := state.RootModule().Resources[roleResourcePath]
		if !ok {
			return fmt.Errorf("failed to find role %s in state", roleResourcePath)
		}

		numPermsAttr := roleResource.Primary.Attributes["permissions.#"]
		numPerms, _ := strconv.Atoi(numPermsAttr)
		configPerms := make([]string, numPerms)
		for i := 0; i < numPerms; i++ {
			configPerms[i] = roleResource.Primary.Attributes["permissions."+strconv.Itoa(i)]
		}

		extraPerms := lists.SliceDifference(configPerms, permissions)
		if len(extraPerms) > 0 {
			return fmt.Errorf("Unexpected permissions found for role %s in state: %v", roleResource.Primary.ID, extraPerms)
		}

		missingPerms := lists.SliceDifference(permissions, configPerms)
		if len(missingPerms) > 0 {
			return fmt.Errorf("Missing expected permissions for role %s in state: %v", roleResource.Primary.ID, missingPerms)
		}

		// All expected permissions found
		return nil
	}
}

func validatePermissionPolicyTest(roleResourcePath string, domain string, entityName string, actionSet ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		roleResource, ok := state.RootModule().Resources[roleResourcePath]
		if !ok {
			return fmt.Errorf("Failed to find role %s in state", roleResourcePath)
		}

		roleAttrs := roleResource.Primary.Attributes
		numPermsAttr := roleAttrs["permission_policies.#"]
		numPerms, _ := strconv.Atoi(numPermsAttr)
		for i := 0; i < numPerms; i++ {
			if roleAttrs["permission_policies."+strconv.Itoa(i)+".domain"] == domain &&
				roleAttrs["permission_policies."+strconv.Itoa(i)+".entity_name"] == entityName {

				numActionsAttr := roleAttrs["permission_policies."+strconv.Itoa(i)+".action_set.#"]
				numActions, _ := strconv.Atoi(numActionsAttr)
				stateActions := make([]string, numActions)
				for j := 0; j < numActions; j++ {
					stateActions[j] = roleAttrs["permission_policies."+strconv.Itoa(i)+".action_set."+strconv.Itoa(j)]
				}

				extraActions := lists.SliceDifference(stateActions, actionSet)
				if len(extraActions) > 0 {
					return fmt.Errorf("Unexpected permission actions found for role %s in state: %v", roleResource.Primary.ID, extraActions)
				}

				missingActions := lists.SliceDifference(actionSet, stateActions)
				if len(missingActions) > 0 {
					return fmt.Errorf("Missing expected permission actions for role %s in state: %v", roleResource.Primary.ID, missingActions)
				}

				// Found expected policy
				return nil
			}
		}

		return fmt.Errorf("Missing expected permission policy for role %s in state: %s %s", roleResource.Primary.ID, domain, entityName)
	}
}

func validatePermPolicyCondition(
	roleResourcePath string,
	domain string,
	entityName string,
	conjunction string,
	variableName string,
	operator string,
	typeVar string,
	value string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		roleResource, ok := state.RootModule().Resources[roleResourcePath]
		if !ok {
			return fmt.Errorf("Failed to find role %s in state", roleResourcePath)
		}

		roleAttrs := roleResource.Primary.Attributes
		numPermsAttr := roleAttrs["permission_policies.#"]
		numPerms, _ := strconv.Atoi(numPermsAttr)
		for i := 0; i < numPerms; i++ {
			strNum := strconv.Itoa(i)
			if roleAttrs["permission_policies."+strNum+".domain"] == domain &&
				roleAttrs["permission_policies."+strNum+".entity_name"] == entityName {

				// Check condition exists and matches
				numCondAttr := roleAttrs["permission_policies."+strNum+".conditions.#"]
				numCond, _ := strconv.Atoi(numCondAttr)

				if numCond == 0 {
					return fmt.Errorf("Missing conditions in role %s", roleResource.Primary.ID)
				}

				stateConjunction := roleAttrs["permission_policies."+strNum+".conditions.0.conjunction"]
				if stateConjunction != conjunction {
					return fmt.Errorf("Invalid condition conjunction role %s: %v", roleResource.Primary.ID, stateConjunction)
				}

				stateVarName := roleAttrs["permission_policies."+strNum+".conditions.0.terms.0.variable_name"]
				if stateVarName != variableName {
					return fmt.Errorf("Invalid condition variable name in role %s: %v", roleResource.Primary.ID, stateVarName)
				}

				stateOp := roleAttrs["permission_policies."+strNum+".conditions.0.terms.0.operator"]
				if stateOp != operator {
					return fmt.Errorf("Invalid condition operator name in role %s: %v", roleResource.Primary.ID, stateOp)
				}

				stateType := roleAttrs["permission_policies."+strNum+".conditions.0.terms.0.operands.0.type"]
				if stateType != typeVar {
					return fmt.Errorf("Invalid condition operand type in role %s: %v", roleResource.Primary.ID, stateType)
				}

				// Don't check value since the roles api allows the type to be set without setting a value
				if value == "" {
					return nil
				}

				if typeVar == "QUEUE" {
					// Get the ID of the queue in the expected value and compare
					stateQueue := roleAttrs["permission_policies."+strNum+".conditions.0.terms.0.operands.0.queue_id"]
					queueResource, ok := state.RootModule().Resources[value]
					if !ok {
						return fmt.Errorf("Failed to find queue %s in state", value)
					}
					if stateQueue != queueResource.Primary.ID {
						return fmt.Errorf("Condition operand value in role %s did not match queue ID: %v", roleResource.Primary.ID, stateQueue)
					}
				} else if typeVar == "USER" {
					// Get the ID of the user in the expected value and compare
					stateUser := roleAttrs["permission_policies."+strNum+".conditions.0.terms.0.operands.0.user_id"]
					userResource, ok := state.RootModule().Resources[value]
					if !ok {
						return fmt.Errorf("Failed to find queue %s in state", value)
					}
					if stateUser != userResource.Primary.ID {
						return fmt.Errorf("Condition operand value in role %s did not match user ID: %v", roleResource.Primary.ID, stateUser)
					}
				} else {
					stateVal := roleAttrs["permission_policies."+strNum+".conditions.0.terms.0.operands.0.value"]
					if stateVal != value {
						return fmt.Errorf("Invalid condition operand value in role %s: %v", roleResource.Primary.ID, stateVal)
					}
				}

				return nil
			}
		}

		return fmt.Errorf("Missing expected permission policy for role %s in state: %s %s", roleResource.Primary.ID, domain, entityName)
	}
}
