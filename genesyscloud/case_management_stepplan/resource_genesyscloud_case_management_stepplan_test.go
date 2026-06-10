package case_management_stepplan

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
	refPrefix := caseplanpkg.AccReferencePrefix(suffix)
	schemaName := caseplanpkg.AccSubstrSchema("tf_stp_" + suffix)
	wbName := caseplanpkg.AccSubstrSchema("tf_wb_" + suffix)
	wtName := caseplanpkg.AccSubstrSchema("tf_wt_" + suffix)
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
		CheckDestroy: caseplanpkg.AccVerifyCaseplanDestroyed,
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
		caseplanpkg.AccCustomerIntentDepsHCL(caseplanName, "acc stepplan deps") +
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
    id = genesyscloud_intents_customerintents.intent.id
  }

  default_case_owner {
    id = genesyscloud_user.owner.id
  }

  data_schema {
    id = genesyscloud_task_management_workitem_schema.schema.id
  }

  lifecycle {
    ignore_changes = [data_schema]
  }
}
`, emailLocal, caseplanName, strings.ToUpper(strings.TrimSpace(refPrefix)), caseplanpkg.AccOwnerRoleAndUserRolesHCL(caseplanName))
}
