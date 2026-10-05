package case_management_caseplan

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	workbin "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	worktype "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

const accCaseplanPath = "genesyscloud_case_management_caseplan.cp"

type accCaseplanNames struct {
	caseplan, refPrefix, schema, workbin, worktype, emailLocal string
}

func newAccCaseplanNames(prefix string) accCaseplanNames {
	suffix := uuid.NewString()
	return accCaseplanNames{
		caseplan:   "tf_acc_" + prefix + "_" + suffix,
		refPrefix:  AccReferencePrefix(suffix),
		schema:     AccSubstrSchema("tf_" + prefix + "_" + suffix),
		workbin:    "tf_acc_" + prefix + "_wb_" + suffix,
		worktype:   "tf_acc_" + prefix + "_wt_" + suffix,
		emailLocal: "tf_acc_" + prefix + "_" + strings.ReplaceAll(suffix, "-", ""),
	}
}

type accCaseplanOptions struct {
	name        string
	description string
	dueSeconds  int
	stageplans  []string
	intake      string
}

func accStageplan(name, stepName string, workitem bool) string {
	activity := ""
	if workitem {
		activity = `
      activity_type = "Workitem"
      workitem_settings {
        worktype_id = genesyscloud_task_management_worktype.wt.id
      }`
	}
	return fmt.Sprintf(`
  stageplan {
    name = %q
    stepplan {
      name = %q%s
    }
  }
`, name, stepName, activity)
}

func accIntake(required bool, order int) string {
	return fmt.Sprintf(`
  intake_settings {
    property      = "acc_note_text"
    required      = %t
    display_order = %d
  }
`, required, order)
}

// Do not use t.Parallel(): each test creates a workitem schema; parallel acc runs can exceed org limits (e.g. 100 schemas).
func TestAccResourceCaseManagementCaseplan(t *testing.T) {
	n := newAccCaseplanNames("cp")
	dataPath := "data.genesyscloud_case_management_caseplan.by_name"

	ids := map[string]string{}
	base := accCaseplanOptions{
		name:        n.caseplan,
		description: "acc caseplan",
		dueSeconds:  86400,
		stageplans: []string{
			accStageplan("Intake", "Triage", false),
			accStageplan("Investigate", "Work it", true),
			accStageplan("Resolve", "Close out", false),
		},
	}
	renamed := base
	renamed.name = n.caseplan + "_renamed"
	renamed.description = "acc caseplan renamed"

	reshaped := renamed
	reshaped.stageplans = []string{
		accStageplan("Resolve", "Close out", false),
		accStageplan("Escalate", "Hand off", false),
		accStageplan("Investigation", "Work it", true),
		accStageplan("Intake", "Triage", false),
	}

	shrunk := reshaped
	shrunk.stageplans = []string{accStageplan("Investigation", "Work it", true)}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create everything except the caseplan first so the owner's roles propagate.
				Config: testAccCaseplanDeps(n),
			},
			{
				PreConfig: func() { time.Sleep(15 * time.Second) },
				Config: testAccCaseplanConfig(n, base) + fmt.Sprintf(`
data "genesyscloud_case_management_caseplan" "by_name" {
  name       = %q
  depends_on = [genesyscloud_case_management_caseplan.cp]
}
`, n.caseplan),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(accCaseplanPath, "name", n.caseplan),
					resource.TestCheckResourceAttr(accCaseplanPath, "reference_prefix", n.refPrefix),
					resource.TestCheckResourceAttr(accCaseplanPath, "published_version", "1"),
					resource.TestCheckResourceAttr(accCaseplanPath, "has_draft", "false"),
					resource.TestCheckResourceAttrPair(accCaseplanPath, "division_id", "data.genesyscloud_auth_division_home.home", "id"),
					resource.TestCheckResourceAttrPair(accCaseplanPath, "customer_intent.0.id", "genesyscloud_intents_customerintents.intent", "id"),
					resource.TestCheckResourceAttrPair(accCaseplanPath, "default_case_owner.0.id", "genesyscloud_user.owner", "id"),
					resource.TestCheckResourceAttrPair(accCaseplanPath, "data_schema.0.id", "genesyscloud_task_management_workitem_schema.schema", "id"),
					resource.TestCheckResourceAttrPair(dataPath, "id", accCaseplanPath, "id"),
					resource.TestCheckResourceAttr(accCaseplanPath, "stageplan.#", "3"),
					resource.TestCheckResourceAttr(accCaseplanPath, "stageplan.0.name", "Intake"),
					resource.TestCheckResourceAttr(accCaseplanPath, "stageplan.0.stepplan.0.activity_type", "None"),
					resource.TestCheckResourceAttr(accCaseplanPath, "stageplan.1.name", "Investigate"),
					resource.TestCheckResourceAttr(accCaseplanPath, "stageplan.1.stepplan.0.activity_type", "Workitem"),
					resource.TestCheckResourceAttrPair(accCaseplanPath, "stageplan.1.stepplan.0.workitem_settings.0.worktype_id", "genesyscloud_task_management_worktype.wt", "id"),
					resource.TestCheckResourceAttr(accCaseplanPath, "stageplan.2.name", "Resolve"),
					captureAttr(accCaseplanPath, "stageplan.0.id", ids, "intake"),
					captureAttr(accCaseplanPath, "stageplan.1.id", ids, "investigate"),
					captureAttr(accCaseplanPath, "stageplan.2.id", ids, "resolve"),
				),
			},
			{
				// name and description are unversioned: no new draft, no publish.
				Config: testAccCaseplanConfig(n, renamed),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(accCaseplanPath, "name", renamed.name),
					resource.TestCheckResourceAttr(accCaseplanPath, "description", renamed.description),
					resource.TestCheckResourceAttr(accCaseplanPath, "published_version", "1"),
				),
			},
			{
				// Reorder, rename in place, and insert in one apply; matched stageplans keep their ids.
				Config: testAccCaseplanConfig(n, reshaped),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(accCaseplanPath, "published_version", "2"),
					resource.TestCheckResourceAttr(accCaseplanPath, "has_draft", "false"),
					resource.TestCheckResourceAttr(accCaseplanPath, "stageplan.#", "4"),
					resource.TestCheckResourceAttr(accCaseplanPath, "stageplan.0.name", "Resolve"),
					resource.TestCheckResourceAttr(accCaseplanPath, "stageplan.1.name", "Escalate"),
					resource.TestCheckResourceAttr(accCaseplanPath, "stageplan.2.name", "Investigation"),
					resource.TestCheckResourceAttr(accCaseplanPath, "stageplan.3.name", "Intake"),
					checkCapturedAttr(accCaseplanPath, "stageplan.0.id", ids, "resolve"),
					checkCapturedAttr(accCaseplanPath, "stageplan.3.id", ids, "intake"),
				),
			},
			{
				Config: testAccCaseplanConfig(n, shrunk),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(accCaseplanPath, "published_version", "3"),
					resource.TestCheckResourceAttr(accCaseplanPath, "stageplan.#", "1"),
					resource.TestCheckResourceAttr(accCaseplanPath, "stageplan.0.name", "Investigation"),
					resource.TestCheckResourceAttr(accCaseplanPath, "stageplan.0.stepplan.0.activity_type", "Workitem"),
				),
			},
			{
				ResourceName:      accCaseplanPath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: AccVerifyCaseplanDestroyed,
	})
}

func TestAccResourceCaseManagementCaseplanIntakeSettings(t *testing.T) {
	n := newAccCaseplanNames("cpin")
	opts := accCaseplanOptions{
		name:        n.caseplan,
		description: "acc caseplan intake",
		dueSeconds:  86400,
		stageplans:  []string{accStageplan("Only stage", "Only step", false)},
		intake:      accIntake(false, 1),
	}
	updated := opts
	updated.intake = accIntake(true, 2)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: testAccCaseplanDeps(n),
			},
			{
				PreConfig: func() { time.Sleep(15 * time.Second) },
				Config:    testAccCaseplanConfig(n, opts),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(accCaseplanPath, "published_version", "1"),
					resource.TestCheckResourceAttr(accCaseplanPath, "intake_settings.#", "1"),
					resource.TestCheckResourceAttr(accCaseplanPath, "intake_settings.0.property", "acc_note_text"),
					resource.TestCheckResourceAttr(accCaseplanPath, "intake_settings.0.required", "false"),
					resource.TestCheckResourceAttr(accCaseplanPath, "intake_settings.0.display_order", "1"),
				),
			},
			{
				// Intake settings are versioned but not frozen: editable after publish via a new draft.
				Config: testAccCaseplanConfig(n, updated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(accCaseplanPath, "published_version", "2"),
					resource.TestCheckResourceAttr(accCaseplanPath, "intake_settings.0.required", "true"),
					resource.TestCheckResourceAttr(accCaseplanPath, "intake_settings.0.display_order", "2"),
				),
			},
			{
				ResourceName:      accCaseplanPath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: AccVerifyCaseplanDestroyed,
	})
}

func captureAttr(path, key string, into map[string]string, as string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(path, key, func(v string) error {
		if v == "" {
			return fmt.Errorf("%s.%s is empty", path, key)
		}
		into[as] = v
		return nil
	})
}

func checkCapturedAttr(path, key string, from map[string]string, as string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(path, key, func(v string) error {
		if v != from[as] {
			return fmt.Errorf("%s.%s = %q, expected the id captured as %q (%q)", path, key, v, as, from[as])
		}
		return nil
	})
}

// testAccCaseplanDeps returns every dependency of the caseplan without the caseplan itself.
func testAccCaseplanDeps(n accCaseplanNames) string {
	props := `jsonencode({
    acc_note_text = {
      allOf     = [{ "$ref" = "#/definitions/text" }]
      title     = "n"
      minLength = 1
      maxLength = 100
    }
  })`
	wtExtra := `
  schema_id          = genesyscloud_task_management_workitem_schema.schema.id
  schema_version     = floor(genesyscloud_task_management_workitem_schema.schema.version)
  assignment_enabled = false
`

	return gcloud.GenerateAuthDivisionHomeDataSource("home") +
		AccCustomerIntentDepsHCL(n.caseplan, "acc caseplan deps") +
		workitemSchema.GenerateWorkitemSchemaResource("schema", n.schema, "acc caseplan schema", props, util.TrueValue) +
		workbin.GenerateWorkbinResource("wb", n.workbin, "acc", "data.genesyscloud_auth_division_home.home.id") +
		worktype.GenerateWorktypeResourceBasic("wt", n.worktype, "acc", "genesyscloud_task_management_workbin.wb.id", wtExtra) +
		fmt.Sprintf(`
resource "genesyscloud_user" "owner" {
  email       = "%[1]s@exampleuser.com"
  name        = "%[2]s owner"
  password    = "TfAccCaseplan1!"
  division_id = data.genesyscloud_auth_division_home.home.id
}

%[3]s
`, n.emailLocal, n.caseplan, AccOwnerRoleAndUserRolesHCL(n.caseplan))
}

func testAccCaseplanConfig(n accCaseplanNames, o accCaseplanOptions) string {
	return testAccCaseplanDeps(n) + fmt.Sprintf(`
resource "genesyscloud_case_management_caseplan" "cp" {
  depends_on = [genesyscloud_user_roles.cp_owner_roles]

  name                            = %[1]q
  division_id                     = data.genesyscloud_auth_division_home.home.id
  description                     = %[2]s
  reference_prefix                = %[3]q
  default_due_duration_in_seconds = %[4]d
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
%[5]s%[6]s}
`, o.name, strconv.Quote(o.description), strings.ToUpper(n.refPrefix), o.dueSeconds, o.intake, strings.Join(o.stageplans, ""))
}
