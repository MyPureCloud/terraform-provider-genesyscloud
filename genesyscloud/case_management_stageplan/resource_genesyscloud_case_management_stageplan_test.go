package case_management_stageplan

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	caseplanpkg "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/case_management_caseplan"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	workitemSchema "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// Serial: creates a workitem schema; avoids parallel acc + org max workitem schemas (often 100).
func TestAccResourceCaseManagementStageplan(t *testing.T) {
	suffix := uuid.NewString()
	caseplanName := "tf_acc_stg_" + suffix
	refPrefix := caseplanpkg.AccReferencePrefix(suffix)
	schemaName := caseplanpkg.AccSubstrSchema("tf_stg_" + suffix)
	emailLocal := "tf_acc_stg_" + strings.ReplaceAll(suffix, "-", "")
	stageName := "TF Acc Stage 1 " + suffix[:8]

	stagePath := "genesyscloud_case_management_stageplan.s1"
	dataPath := "data.genesyscloud_case_management_stageplan.lookup"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: testAccCaseplanForStageplan(caseplanName, refPrefix, schemaName, emailLocal) + fmt.Sprintf(`
resource "genesyscloud_case_management_stageplan" "s1" {
  caseplan_id   = genesyscloud_case_management_caseplan.cp.id
  stage_number  = 1
  name          = "%s"
  description   = "acc stageplan"
}

data "genesyscloud_case_management_stageplan" "lookup" {
  caseplan_id   = genesyscloud_case_management_caseplan.cp.id
  stage_number  = 1
}
`, stageName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(stagePath, "name", stageName),
					resource.TestCheckResourceAttr(stagePath, "stage_number", "1"),
					resource.TestCheckResourceAttrPair(stagePath, "caseplan_id", caseplanpkg.ResourceType+".cp", "id"),
					resource.TestCheckResourceAttrPair(dataPath, "id", stagePath, "id"),
				),
			},
			{
				ResourceName:      stagePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: caseplanpkg.AccVerifyCaseplanDestroyed,
	})
}

func testAccCaseplanForStageplan(caseplanName, refPrefix, schemaName, emailLocal string) string {
	props := `jsonencode({
    acc_note_text = {
      allOf     = [{ "$ref" = "#/definitions/text" }]
      title     = "n"
      minLength = 1
      maxLength = 100
    }
  })`

	return gcloud.GenerateAuthDivisionHomeDataSource("home") +
		caseplanpkg.AccCustomerIntentDepsHCL(caseplanName, "acc stageplan deps") +
		workitemSchema.GenerateWorkitemSchemaResource("schema", schemaName, "acc", props, util.TrueValue) +
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
  description                     = "acc caseplan for stageplan test"
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
    id = genesyscloud_task_management_workitem_schema.schema.id
  }

  lifecycle {
    ignore_changes = [data_schema]
  }
}
`, emailLocal, caseplanName, strings.ToUpper(strings.TrimSpace(refPrefix)), caseplanpkg.AccOwnerRoleAndUserRolesHCL(caseplanName))
}
