package case_management_caseplan

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v191/platformclientv2"
	authrole "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_role"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// Shared acceptance-test HCL / check helpers for case_management_caseplan and dependent packages (stageplan, stepplan).

// AccReferencePrefix derives a short uppercase reference_prefix from a UUID suffix (same logic as historical test helpers).
func AccReferencePrefix(suffix string) string {
	p := strings.ReplaceAll(suffix, "-", "")
	if len(p) > 8 {
		p = p[:8]
	}
	return strings.ToUpper(p)
}

// AccSubstrSchema truncates s to the max workitem schema name length used in tests.
func AccSubstrSchema(s string) string {
	if len(s) <= 50 {
		return s
	}
	return s[:50]
}

// AccCustomerIntentDepsHCL returns intent_category + customer_intent resources wired for a caseplan test stack.
func AccCustomerIntentDepsHCL(namePrefix, categoryDescription string) string {
	return fmt.Sprintf(`
resource "genesyscloud_intents_categories" "cat" {
  name        = "%[1]s_cat"
  description = "%[2]s"
}

resource "genesyscloud_intents_customerintents" "intent" {
  name        = "%[1]s_intent"
  description = "acc"
  expiry_time = 24
  category_id = genesyscloud_intents_categories.cat.id
}
`, namePrefix, categoryDescription)
}

// AccOwnerRoleAndUserRolesHCL grants default_case_owner caseManagement caseplan/case view in home division (auth role + user_roles).
func AccOwnerRoleAndUserRolesHCL(roleDisplayName string) string {
	roleName := roleDisplayName
	if len(roleName) > 100 {
		roleName = roleName[:100]
	}
	return authrole.GenerateAuthRoleResource(
		"cp_owner_cm",
		roleName,
		"TF acc: caseManagement caseplan and case view for default_case_owner in home division",
		authrole.GenerateRolePermPolicy("caseManagement", "caseplan", `"view"`),
		authrole.GenerateRolePermPolicy("caseManagement", "case", `"view"`),
	) + `
resource "genesyscloud_user_roles" "cp_owner_roles" {
  user_id = genesyscloud_user.owner.id
  roles {
    role_id      = genesyscloud_auth_role.cp_owner_cm.id
    division_ids = [data.genesyscloud_auth_division_home.home.id]
  }
}
`
}

// AccVerifyCaseplanDestroyed fails CheckDestroy if any genesyscloud_case_management_caseplan still exists in the org.
func AccVerifyCaseplanDestroyed(state *terraform.State) error {
	api := platformclientv2.NewCaseManagementApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}
		cp, resp, err := api.GetCasemanagementCaseplan(rs.Primary.ID)
		if cp != nil {
			return fmt.Errorf("case management caseplan (%s) still exists", rs.Primary.ID)
		}
		if util.IsStatus404(resp) {
			continue
		}
		if err != nil {
			return fmt.Errorf("unexpected error verifying caseplan destroy: %s", err)
		}
	}
	return nil
}
