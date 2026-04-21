package case_management_stageplan

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"
	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	caseplanpkg "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/case_management_caseplan"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	workitemSchema "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccResourceCaseManagementStageplan(t *testing.T) {
	t.Parallel()
	suffix := uuid.NewString()
	caseplanName := "tf_acc_stg_" + suffix
	refPrefix := strings.ReplaceAll(suffix, "-", "")
	if len(refPrefix) > 8 {
		refPrefix = refPrefix[:8]
	}
	schemaName := substrForSchema("tf_stg_" + suffix)
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
		CheckDestroy: testAccVerifyCaseManagementCaseplanDestroyed,
	})
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

func testAccCaseplanForStageplan(caseplanName, refPrefix, schemaName, emailLocal string) string {
	props := `jsonencode({
    acc_note_text = {
      allOf     = [{ "$ref" = "#/definitions/text" }]
      title     = "n"
      maxLength = 100
    }
  })`

	return gcloud.GenerateAuthDivisionHomeDataSource("home") +
		generateAccCustomerIntentDeps(caseplanName) +
		workitemSchema.GenerateWorkitemSchemaResource("schema", schemaName, "acc", props, util.TrueValue) +
		fmt.Sprintf(`
resource "genesyscloud_user" "owner" {
  email       = "%[1]s@exampleuser.com"
  name        = "%[2]s owner"
  password    = "TfAccCaseplan1!"
  division_id = data.genesyscloud_auth_division_home.home.id
}

resource "genesyscloud_case_management_caseplan" "cp" {
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
    id      = genesyscloud_task_management_workitem_schema.schema.id
    version = floor(genesyscloud_task_management_workitem_schema.schema.version)
  }
}
`, emailLocal, caseplanName, refPrefix)
}

func generateAccCustomerIntentDeps(namePrefix string) string {
	return fmt.Sprintf(`
resource "genesyscloud_intent_category" "cat" {
  name        = "%[1]s_cat"
  description = "acc stageplan deps"
}

resource "genesyscloud_customer_intent" "intent" {
  name        = "%[1]s_intent"
  description = "acc"
  expiry_time = 24
  category_id = genesyscloud_intent_category.cat.id
}
`, namePrefix)
}

func substrForSchema(s string) string {
	if len(s) <= 50 {
		return s
	}
	return s[:50]
}
