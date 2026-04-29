package case_management_stepplan

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"
	authrole "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_role"
	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	caseplanpkg "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/case_management_caseplan"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	workbin "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	worktype "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// Serial: creates workitem schema/workbin/worktype; parallel acc can hit org quotas.
func TestAccResourceCaseManagementStepplan(t *testing.T) {
	suffix := uuid.NewString()
	caseplanName := "tf_acc_stp_" + suffix
	refPrefix := testAccCaseplanReferencePrefix(suffix)
	schemaName := substrForSchema("tf_stp_" + suffix)
	wbName := substrForSchema("tf_wb_" + suffix)
	wtName := substrForSchema("tf_wt_" + suffix)
	emailLocal := "tf_acc_stp_" + strings.ReplaceAll(suffix, "-", "")
	stepName := "TF Acc Step " + suffix[:8]

	stepPath := "genesyscloud_case_management_stepplan.step1"
	dataPath := "data.genesyscloud_case_management_stepplan.lookup"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: testAccCaseplanStackForStepplan(caseplanName, refPrefix, schemaName, wbName, wtName, emailLocal) + fmt.Sprintf(`
resource "genesyscloud_case_management_stageplan" "st1" {
  caseplan_id  = genesyscloud_case_management_caseplan.cp.id
  stage_number = 1
  name         = "Acc stage 1 %[1]s"
}

resource "genesyscloud_case_management_stepplan" "step1" {
  caseplan_id     = genesyscloud_case_management_caseplan.cp.id
  stage_number    = 1
  name            = "%[2]s"
  description     = "acc stepplan"
  activity_type   = "Workitem"
  workitem_settings {
    worktype_id = genesyscloud_task_management_worktype.wt.id
  }
  depends_on = [
    genesyscloud_case_management_stageplan.st1,
    genesyscloud_task_management_worktype.wt,
  ]
}

data "genesyscloud_case_management_stepplan" "lookup" {
  caseplan_id   = genesyscloud_case_management_caseplan.cp.id
  stage_number  = 1
}
`, suffix[:8], stepName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(stepPath, "name", stepName),
					resource.TestCheckResourceAttr(stepPath, "stage_number", "1"),
					resource.TestCheckResourceAttrPair(stepPath, "caseplan_id", caseplanpkg.ResourceType+".cp", "id"),
					resource.TestCheckResourceAttrPair(stepPath, "workitem_settings.0.worktype_id", "genesyscloud_task_management_worktype.wt", "id"),
					resource.TestCheckResourceAttrPair(dataPath, "id", stepPath, "id"),
				),
			},
			{
				ResourceName:      stepPath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testAccVerifyCaseManagementCaseplanDestroyed,
	})
}

func testAccCaseplanStackForStepplan(caseplanName, refPrefix, schemaName, wbName, wtName, emailLocal string) string {
	props := `jsonencode({
    acc_note_text = {
      allOf     = [{ "$ref" = "#/definitions/text" }]
      title     = "n"
      minLength = 1
      maxLength = 100
    }
  })`

	wtExtra := `
		schema_id = genesyscloud_task_management_workitem_schema.schema.id
		schema_version = floor(genesyscloud_task_management_workitem_schema.schema.version)
		assignment_enabled = false
`

	return gcloud.GenerateAuthDivisionHomeDataSource("home") +
		generateAccCustomerIntentDeps(caseplanName) +
		workitemSchema.GenerateWorkitemSchemaResource("schema", schemaName, "acc", props, util.TrueValue) +
		workbin.GenerateWorkbinResource("wb", wbName, "acc", "data.genesyscloud_auth_division_home.home.id") +
		worktype.GenerateWorktypeResourceBasic("wt", wtName, "acc", "genesyscloud_task_management_workbin.wb.id", wtExtra) +
		fmt.Sprintf(`
resource "genesyscloud_user" "owner" {
  email       = "%[1]s@exampleuser.com"
  name        = "%[2]s owner"
  password    = "TfAccCaseplan1!"
  division_id = data.genesyscloud_auth_division_home.home.id
}

%[4]s

resource "genesyscloud_case_management_caseplan" "cp" {
  depends_on = [genesyscloud_user_roles.cp_owner_roles]

  name                            = "%[2]s"
  division_id                     = data.genesyscloud_auth_division_home.home.id
  description                     = "acc caseplan for stepplan test"
  reference_prefix                = "%[3]s"
  default_due_duration_in_seconds = 86400
  default_ttl_seconds             = 604800

  customer_intent {
    id = genesyscloud_customer_intent.intent.id
  }

  default_case_owner {
    id = genesyscloud_user.owner.id
  }

  data_schema {
    id      = genesyscloud_task_management_workitem_schema.schema.id
    version = floor(genesyscloud_task_management_workitem_schema.schema.version)
  }

  lifecycle {
    ignore_changes = [data_schema]
  }
}
`, emailLocal, caseplanName, strings.ToUpper(strings.TrimSpace(refPrefix)), testAccCaseplanOwnerRoleAndUserRolesHCL(caseplanName))
}

func generateAccCustomerIntentDeps(namePrefix string) string {
	return fmt.Sprintf(`
resource "genesyscloud_intent_category" "cat" {
  name        = "%[1]s_cat"
  description = "acc stepplan deps"
}

resource "genesyscloud_customer_intent" "intent" {
  name        = "%[1]s_intent"
  description = "acc"
  expiry_time = 24
  category_id = genesyscloud_intent_category.cat.id
}
`, namePrefix)
}

func testAccCaseplanReferencePrefix(suffix string) string {
	p := strings.ReplaceAll(suffix, "-", "")
	if len(p) > 8 {
		p = p[:8]
	}
	return strings.ToUpper(p)
}

func substrForSchema(s string) string {
	if len(s) <= 50 {
		return s
	}
	return s[:50]
}

func testAccCaseplanOwnerRoleAndUserRolesHCL(roleDisplayName string) string {
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

func testAccVerifyCaseManagementCaseplanDestroyed(state *terraform.State) error {
	api := platformclientv2.NewCaseManagementApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != caseplanpkg.ResourceType {
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
