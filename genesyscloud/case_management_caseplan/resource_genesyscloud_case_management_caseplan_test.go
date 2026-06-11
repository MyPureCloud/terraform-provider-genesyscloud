package case_management_caseplan

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	workitemSchema "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// Do not use t.Parallel(): each test creates a workitem schema; parallel acc runs can exceed org limits (e.g. 100 schemas).
func TestAccResourceCaseManagementCaseplan(t *testing.T) {
	suffix := uuid.NewString()
	caseplanName := "tf_acc_cp_" + suffix
	refPrefix := AccReferencePrefix(suffix)
	schemaName := AccSubstrSchema("tf_cp_" + suffix)
	emailLocal := "tf_acc_cp_" + strings.ReplaceAll(suffix, "-", "")

	resourcePath := "genesyscloud_case_management_caseplan.cp"
	dataPath := "data.genesyscloud_case_management_caseplan.by_name"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: testAccCaseManagementCaseplanConfig(caseplanName, refPrefix, schemaName, emailLocal) + fmt.Sprintf(`
data "genesyscloud_case_management_caseplan" "by_name" {
  name       = "%s"
  depends_on = [genesyscloud_case_management_caseplan.cp]
}
`, caseplanName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", caseplanName),
					resource.TestCheckResourceAttr(resourcePath, "reference_prefix", refPrefix),
					resource.TestCheckResourceAttrPair(resourcePath, "division_id", "data.genesyscloud_auth_division_home.home", "id"),
					resource.TestCheckResourceAttrPair(resourcePath, "customer_intent.0.id", "genesyscloud_intents_customerintents.intent", "id"),
					resource.TestCheckResourceAttrPair(resourcePath, "default_case_owner.0.id", "genesyscloud_user.owner", "id"),
					resource.TestCheckResourceAttrPair(resourcePath, "data_schema.0.id", "genesyscloud_task_management_workitem_schema.schema", "id"),
					resource.TestCheckResourceAttrPair(dataPath, "id", resourcePath, "id"),
				),
			},
			{
				ResourceName:      resourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: AccVerifyCaseplanDestroyed,
	})
}

func TestAccResourceCaseManagementCaseplanIntakeSettings(t *testing.T) {
	suffix := uuid.NewString()
	caseplanName := "tf_acc_cpin_" + suffix
	refPrefix := AccReferencePrefix(suffix)
	schemaName := AccSubstrSchema("tf_cpin_" + suffix)
	emailLocal := "tf_acc_cpin_" + strings.ReplaceAll(suffix, "-", "")
	resourcePath := "genesyscloud_case_management_caseplan.cp"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: testAccCaseManagementCaseplanConfigIntake(caseplanName, refPrefix, schemaName, emailLocal, "acc caseplan intake", 86400, 604800, false, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "intake_settings.#", "1"),
					resource.TestCheckResourceAttr(resourcePath, "intake_settings.0.property", "acc_note_text"),
					resource.TestCheckResourceAttr(resourcePath, "intake_settings.0.required", "false"),
					resource.TestCheckResourceAttr(resourcePath, "intake_settings.0.display_order", "1"),
				),
			},
			{
				Config: testAccCaseManagementCaseplanConfigIntake(caseplanName, refPrefix, schemaName, emailLocal, "acc caseplan intake", 86400, 604800, true, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "intake_settings.#", "1"),
					resource.TestCheckResourceAttr(resourcePath, "intake_settings.0.property", "acc_note_text"),
					resource.TestCheckResourceAttr(resourcePath, "intake_settings.0.required", "true"),
					resource.TestCheckResourceAttr(resourcePath, "intake_settings.0.display_order", "2"),
				),
			},
			{
				ResourceName:      resourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: AccVerifyCaseplanDestroyed,
	})
}

func TestAccResourceCaseManagementCaseplanPublish(t *testing.T) {
	suffix := uuid.NewString()
	caseplanName := "tf_acc_cppub_" + suffix
	refPrefix := AccReferencePrefix(suffix)
	schemaName := AccSubstrSchema("tf_cpp_" + suffix)
	emailLocal := "tf_acc_cpp_" + strings.ReplaceAll(suffix, "-", "")

	resourcePath := "genesyscloud_case_management_caseplan.cp"
	publishPath := "genesyscloud_case_management_caseplan_publish.pub"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: testAccCaseManagementCaseplanConfig(caseplanName, refPrefix, schemaName, emailLocal) + `
resource "genesyscloud_case_management_caseplan_publish" "pub" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(publishPath, "caseplan_id", resourcePath, "id"),
				),
			},
			{
				ResourceName:            publishPath,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revision"},
			},
		},
		CheckDestroy: AccVerifyCaseplanDestroyed,
	})
}

func TestAccResourceCaseManagementCaseplanPublish_revisionBump(t *testing.T) {
	suffix := uuid.NewString()
	caseplanName := "tf_acc_cppubr_" + suffix
	refPrefix := AccReferencePrefix(suffix)
	schemaName := AccSubstrSchema("tf_cppr_" + suffix)
	emailLocal := "tf_acc_cppr_" + strings.ReplaceAll(suffix, "-", "")

	resourcePath := "genesyscloud_case_management_caseplan.cp"
	publishPath := "genesyscloud_case_management_caseplan_publish.pub"
	base := testAccCaseManagementCaseplanConfig(caseplanName, refPrefix, schemaName, emailLocal)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: base + `
resource "genesyscloud_case_management_caseplan_publish" "pub" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
  revision    = 0
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(publishPath, "revision", "0"),
					resource.TestCheckResourceAttrPair(publishPath, "caseplan_id", resourcePath, "id"),
				),
			},
			{
				Config: base + `
resource "genesyscloud_case_management_caseplan_publish" "pub" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
  revision    = 0
}

resource "genesyscloud_case_management_caseplan_create_version" "new_draft" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
  revision    = 0
  depends_on  = [genesyscloud_case_management_caseplan_publish.pub]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(publishPath, "caseplan_id", resourcePath, "id"),
				),
			},
			{
				Config: base + `
resource "genesyscloud_case_management_caseplan_publish" "pub" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
  revision    = 1
}

resource "genesyscloud_case_management_caseplan_create_version" "new_draft" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
  revision    = 0
  depends_on  = [genesyscloud_case_management_caseplan_publish.pub]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(publishPath, "revision", "1"),
					resource.TestCheckResourceAttrPair(publishPath, "caseplan_id", resourcePath, "id"),
				),
			},
		},
		CheckDestroy: AccVerifyCaseplanDestroyed,
	})
}

func TestAccResourceCaseManagementCaseplanCreateVersion(t *testing.T) {
	suffix := uuid.NewString()
	caseplanName := "tf_acc_cpver_" + suffix
	refPrefix := AccReferencePrefix(suffix)
	schemaName := AccSubstrSchema("tf_cpv_" + suffix)
	emailLocal := "tf_acc_cpv_" + strings.ReplaceAll(suffix, "-", "")

	resourcePath := "genesyscloud_case_management_caseplan.cp"
	publishPath := "genesyscloud_case_management_caseplan_publish.pub"
	versionPath := "genesyscloud_case_management_caseplan_create_version.new_draft"
	base := testAccCaseManagementCaseplanConfig(caseplanName, refPrefix, schemaName, emailLocal)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: base + `
resource "genesyscloud_case_management_caseplan_publish" "pub" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
}

resource "genesyscloud_case_management_caseplan_create_version" "new_draft" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
  depends_on  = [genesyscloud_case_management_caseplan_publish.pub]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(versionPath, "caseplan_id", resourcePath, "id"),
					resource.TestCheckResourceAttrPair(publishPath, "caseplan_id", resourcePath, "id"),
				),
			},
			{
				ResourceName:            versionPath,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revision"},
			},
		},
		CheckDestroy: AccVerifyCaseplanDestroyed,
	})
}

func TestAccResourceCaseManagementCaseplanCreateVersion_revisionAfterRepublish(t *testing.T) {
	suffix := uuid.NewString()
	caseplanName := "tf_acc_cpverr_" + suffix
	refPrefix := AccReferencePrefix(suffix)
	schemaName := AccSubstrSchema("tf_cpvr_" + suffix)
	emailLocal := "tf_acc_cpvr_" + strings.ReplaceAll(suffix, "-", "")

	resourcePath := "genesyscloud_case_management_caseplan.cp"
	publishPath := "genesyscloud_case_management_caseplan_publish.pub"
	versionPath := "genesyscloud_case_management_caseplan_create_version.new_draft"
	base := testAccCaseManagementCaseplanConfig(caseplanName, refPrefix, schemaName, emailLocal)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: base + `
resource "genesyscloud_case_management_caseplan_publish" "pub" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
  revision    = 0
}

resource "genesyscloud_case_management_caseplan_create_version" "new_draft" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
  revision    = 0
  depends_on  = [genesyscloud_case_management_caseplan_publish.pub]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(publishPath, "revision", "0"),
					resource.TestCheckResourceAttr(versionPath, "revision", "0"),
					resource.TestCheckResourceAttrPair(versionPath, "caseplan_id", resourcePath, "id"),
				),
			},
			{
				Config: base + `
resource "genesyscloud_case_management_caseplan_publish" "pub" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
  revision    = 1
}

resource "genesyscloud_case_management_caseplan_create_version" "new_draft" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
  revision    = 1
  depends_on  = [genesyscloud_case_management_caseplan_publish.pub]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(publishPath, "revision", "1"),
					resource.TestCheckResourceAttr(versionPath, "revision", "1"),
					resource.TestCheckResourceAttrPair(versionPath, "caseplan_id", resourcePath, "id"),
				),
			},
		},
		CheckDestroy: AccVerifyCaseplanDestroyed,
	})
}

// TestAccResourceCaseManagementCaseplan_publishDraftUpdateRepublish exercises: create → publish → POST new draft
// (create_version) → PATCH caseplan (allowed fields after publish) → publish again (revision bump).
func TestAccResourceCaseManagementCaseplan_publishDraftUpdateRepublish(t *testing.T) {
	suffix := uuid.NewString()
	caseplanName := "tf_acc_cpup_" + suffix
	refPrefix := AccReferencePrefix(suffix)
	schemaName := AccSubstrSchema("tf_cpup_" + suffix)
	emailLocal := "tf_acc_cpup_" + strings.ReplaceAll(suffix, "-", "")

	resourcePath := "genesyscloud_case_management_caseplan.cp"
	publishPath := "genesyscloud_case_management_caseplan_publish.pub"
	versionPath := "genesyscloud_case_management_caseplan_create_version.new_draft"

	descInitial := "acc caseplan draft before publish"
	descUpdated := "acc caseplan updated patch after draft"
	dueInitial, ttlInitial := 86400, 604800
	dueUpdated, ttlUpdated := 86401, 604801

	base := func(desc string, due, ttl int) string {
		return testAccCaseManagementCaseplanConfigFlexible(caseplanName, refPrefix, schemaName, emailLocal, desc, due, ttl)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: base(descInitial, dueInitial, ttlInitial),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "description", descInitial),
					resource.TestCheckResourceAttr(resourcePath, "default_due_duration_in_seconds", fmt.Sprintf("%d", dueInitial)),
					resource.TestCheckResourceAttr(resourcePath, "default_ttl_seconds", fmt.Sprintf("%d", ttlInitial)),
				),
			},
			{
				Config: base(descInitial, dueInitial, ttlInitial) + `
resource "genesyscloud_case_management_caseplan_publish" "pub" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
  revision    = 0
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(publishPath, "revision", "0"),
					resource.TestCheckResourceAttrPair(publishPath, "caseplan_id", resourcePath, "id"),
				),
			},
			{
				Config: base(descInitial, dueInitial, ttlInitial) + `
resource "genesyscloud_case_management_caseplan_publish" "pub" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
  revision    = 0
}

resource "genesyscloud_case_management_caseplan_create_version" "new_draft" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
  revision    = 0
  depends_on  = [genesyscloud_case_management_caseplan_publish.pub]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(versionPath, "caseplan_id", resourcePath, "id"),
					resource.TestCheckResourceAttr(versionPath, "revision", "0"),
				),
			},
			{
				Config: base(descUpdated, dueUpdated, ttlUpdated) + `
resource "genesyscloud_case_management_caseplan_publish" "pub" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
  revision    = 0
}

resource "genesyscloud_case_management_caseplan_create_version" "new_draft" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
  revision    = 0
  depends_on  = [genesyscloud_case_management_caseplan_publish.pub]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "description", descUpdated),
					resource.TestCheckResourceAttr(resourcePath, "default_due_duration_in_seconds", fmt.Sprintf("%d", dueUpdated)),
					resource.TestCheckResourceAttr(resourcePath, "default_ttl_seconds", fmt.Sprintf("%d", ttlUpdated)),
				),
			},
			{
				Config: base(descUpdated, dueUpdated, ttlUpdated) + `
resource "genesyscloud_case_management_caseplan_publish" "pub" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
  revision    = 1
}

resource "genesyscloud_case_management_caseplan_create_version" "new_draft" {
  caseplan_id = genesyscloud_case_management_caseplan.cp.id
  revision    = 0
  depends_on  = [genesyscloud_case_management_caseplan_publish.pub]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(publishPath, "revision", "1"),
					resource.TestCheckResourceAttrPair(publishPath, "caseplan_id", resourcePath, "id"),
				),
			},
		},
		CheckDestroy: AccVerifyCaseplanDestroyed,
	})
}

func testAccCaseManagementCaseplanConfig(caseplanName, refPrefix, schemaName, emailLocal string) string {
	return testAccCaseManagementCaseplanConfigFlexible(caseplanName, refPrefix, schemaName, emailLocal, "acc caseplan", 86400, 604800)
}

func testAccCaseManagementCaseplanConfigFlexible(caseplanName, refPrefix, schemaName, emailLocal, description string, defaultDueSec, defaultTtlSec int) string {
	props := `jsonencode({
    acc_note_text = {
      allOf     = [{ "$ref" = "#/definitions/text" }]
      title     = "n"
      minLength = 1
      maxLength = 100
    }
  })`

	ownerGrants := AccOwnerRoleAndUserRolesHCL(caseplanName)
	// POST accepts mixed case but GET canonicalizes uppercase; mismatch fails SDK post-apply empty-plan checks.
	refNormalized := strings.ToUpper(strings.TrimSpace(refPrefix))

	return gcloud.GenerateAuthDivisionHomeDataSource("home") +
		AccCustomerIntentDepsHCL(caseplanName, "acc caseplan deps") +
		workitemSchema.GenerateWorkitemSchemaResource("schema", schemaName, "acc caseplan schema", props, util.TrueValue) +
		fmt.Sprintf(`
resource "genesyscloud_user" "owner" {
  email       = "%[1]s@exampleuser.com"
  name        = "%[2]s owner"
  password    = "TfAccCaseplan1!"
  division_id = data.genesyscloud_auth_division_home.home.id
}

%[7]s

resource "genesyscloud_case_management_caseplan" "cp" {
  depends_on = [genesyscloud_user_roles.cp_owner_roles]

  name                            = "%[2]s"
  division_id                     = data.genesyscloud_auth_division_home.home.id
  description                     = %[4]s
  reference_prefix                = "%[3]s"
  default_due_duration_in_seconds = %[5]d
  default_ttl_seconds             = %[6]d

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
`, emailLocal, caseplanName, refNormalized, strconv.Quote(description), defaultDueSec, defaultTtlSec, ownerGrants)
}

func testAccCaseManagementCaseplanConfigIntake(caseplanName, refPrefix, schemaName, emailLocal, description string, defaultDueSec, defaultTtlSec int, intakeRequired bool, intakeOrder int) string {
	props := `jsonencode({
    acc_note_text = {
      allOf     = [{ "$ref" = "#/definitions/text" }]
      title     = "n"
      minLength = 1
      maxLength = 100
    }
  })`

	ownerGrants := AccOwnerRoleAndUserRolesHCL(caseplanName)
	refNormalized := strings.ToUpper(strings.TrimSpace(refPrefix))

	return gcloud.GenerateAuthDivisionHomeDataSource("home") +
		AccCustomerIntentDepsHCL(caseplanName, "acc caseplan deps") +
		workitemSchema.GenerateWorkitemSchemaResource("schema", schemaName, "acc caseplan schema", props, util.TrueValue) +
		fmt.Sprintf(`
resource "genesyscloud_user" "owner" {
  email       = "%[1]s@exampleuser.com"
  name        = "%[2]s owner"
  password    = "TfAccCaseplan1!"
  division_id = data.genesyscloud_auth_division_home.home.id
}

%[7]s

resource "genesyscloud_case_management_caseplan" "cp" {
  depends_on = [genesyscloud_user_roles.cp_owner_roles]

  name                            = "%[2]s"
  division_id                     = data.genesyscloud_auth_division_home.home.id
  description                     = %[4]s
  reference_prefix                = "%[3]s"
  default_due_duration_in_seconds = %[5]d
  default_ttl_seconds             = %[6]d

  customer_intent {
    id = genesyscloud_intents_customerintents.intent.id
  }

  default_case_owner {
    id = genesyscloud_user.owner.id
  }

  data_schema {
    id = genesyscloud_task_management_workitem_schema.schema.id
  }

  intake_settings {
    property      = "acc_note_text"
    required      = %[8]t
    display_order = %[9]d
  }

  lifecycle {
    ignore_changes = [data_schema]
  }
}
`, emailLocal, caseplanName, refNormalized, strconv.Quote(description), defaultDueSec, defaultTtlSec, ownerGrants, intakeRequired, intakeOrder)
}
