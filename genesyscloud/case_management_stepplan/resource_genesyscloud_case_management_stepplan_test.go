package case_management_stepplan

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	caseplanpkg "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/case_management_caseplan"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccResourceCaseManagementStepplan(t *testing.T) {
	t.Parallel()
	suffix := uuid.NewString()
	caseplanName := "tf_acc_stp_" + suffix
	refPrefix := strings.ReplaceAll(suffix, "-", "")
	if len(refPrefix) > 8 {
		refPrefix = refPrefix[:8]
	}
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
	return fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}

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

resource "genesyscloud_task_management_workitem_schema" "schema" {
  name        = "%[2]s"
  description = "acc"
  enabled     = true
  properties = jsonencode({
    acc_note_text = {
      allOf     = [{ "$ref" = "#/definitions/text" }]
      title     = "n"
      maxLength = 100
    }
  })
}

resource "genesyscloud_task_management_workbin" "wb" {
  name        = "%[3]s"
  description = "acc"
  division_id = data.genesyscloud_auth_division_home.home.id
}

resource "genesyscloud_task_management_worktype" "wt" {
  name               = "%[4]s"
  description        = "acc"
  default_workbin_id = genesyscloud_task_management_workbin.wb.id
  schema_id          = genesyscloud_task_management_workitem_schema.schema.id
  schema_version     = floor(genesyscloud_task_management_workitem_schema.schema.version)
  assignment_enabled = false
}

resource "genesyscloud_user" "owner" {
  email       = "%[5]s@exampleuser.com"
  name        = "%[1]s owner"
  password    = "TfAccCaseplan1!"
  division_id = data.genesyscloud_auth_division_home.home.id
}

resource "genesyscloud_case_management_caseplan" "cp" {
  name                            = "%[6]s"
  division_id                     = data.genesyscloud_auth_division_home.home.id
  description                     = "acc caseplan for stepplan test"
  reference_prefix                = "%[7]s"
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
}
`, caseplanName, schemaName, wbName, wtName, emailLocal, caseplanName, refPrefix)
}

func substrForSchema(s string) string {
	if len(s) <= 50 {
		return s
	}
	return s[:50]
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
