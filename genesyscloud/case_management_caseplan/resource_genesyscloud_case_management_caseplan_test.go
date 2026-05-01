package case_management_caseplan

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"
	authrole "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_role"
	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	workitemSchema "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// Do not use t.Parallel(): each test creates a workitem schema; parallel acc runs can exceed org limits (e.g. 100 schemas).
func TestAccResourceCaseManagementCaseplan(t *testing.T) {
	suffix := uuid.NewString()
	caseplanName := "tf_acc_cp_" + suffix
	refPrefix := testAccCaseplanReferencePrefix(suffix)
	schemaName := substrForSchema("tf_cp_" + suffix)
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
					resource.TestCheckResourceAttrPair(resourcePath, "customer_intent.0.id", "genesyscloud_customer_intent.intent", "id"),
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
		CheckDestroy: testVerifyCaseManagementCaseplanDestroyed,
	})
}

func TestAccResourceCaseManagementCaseplanPublish(t *testing.T) {
	suffix := uuid.NewString()
	caseplanName := "tf_acc_cppub_" + suffix
	refPrefix := testAccCaseplanReferencePrefix(suffix)
	schemaName := substrForSchema("tf_cpp_" + suffix)
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
		CheckDestroy: testVerifyCaseManagementCaseplanDestroyed,
	})
}

func TestAccResourceCaseManagementCaseplanPublish_revisionBump(t *testing.T) {
	suffix := uuid.NewString()
	caseplanName := "tf_acc_cppubr_" + suffix
	refPrefix := testAccCaseplanReferencePrefix(suffix)
	schemaName := substrForSchema("tf_cppr_" + suffix)
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
		CheckDestroy: testVerifyCaseManagementCaseplanDestroyed,
	})
}

func TestAccResourceCaseManagementCaseplanCreateVersion(t *testing.T) {
	suffix := uuid.NewString()
	caseplanName := "tf_acc_cpver_" + suffix
	refPrefix := testAccCaseplanReferencePrefix(suffix)
	schemaName := substrForSchema("tf_cpv_" + suffix)
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
		CheckDestroy: testVerifyCaseManagementCaseplanDestroyed,
	})
}

func TestAccResourceCaseManagementCaseplanCreateVersion_revisionAfterRepublish(t *testing.T) {
	suffix := uuid.NewString()
	caseplanName := "tf_acc_cpverr_" + suffix
	refPrefix := testAccCaseplanReferencePrefix(suffix)
	schemaName := substrForSchema("tf_cpvr_" + suffix)
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
		CheckDestroy: testVerifyCaseManagementCaseplanDestroyed,
	})
}

// TestAccResourceCaseManagementCaseplan_publishDraftUpdateRepublish exercises: create → publish → POST new draft
// (create_version) → PATCH caseplan (allowed fields after publish) → publish again (revision bump).
func TestAccResourceCaseManagementCaseplan_publishDraftUpdateRepublish(t *testing.T) {
	suffix := uuid.NewString()
	caseplanName := "tf_acc_cpup_" + suffix
	refPrefix := testAccCaseplanReferencePrefix(suffix)
	schemaName := substrForSchema("tf_cpup_" + suffix)
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
		CheckDestroy: testVerifyCaseManagementCaseplanDestroyed,
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

	ownerGrants := testAccCaseplanOwnerRoleAndUserRolesHCL(caseplanName)
	// POST accepts mixed case but GET canonicalizes uppercase; mismatch fails SDK post-apply empty-plan checks.
	refNormalized := strings.ToUpper(strings.TrimSpace(refPrefix))

	return gcloud.GenerateAuthDivisionHomeDataSource("home") +
		generateAccCustomerIntentDeps(caseplanName) +
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
    # Version read from dataschemas can lag floor(workitem_schema.version) briefly after create.
    ignore_changes = [data_schema]
  }
}
`, emailLocal, caseplanName, refNormalized, strconv.Quote(description), defaultDueSec, defaultTtlSec, ownerGrants)
}

func generateAccCustomerIntentDeps(namePrefix string) string {
	return fmt.Sprintf(`
resource "genesyscloud_intent_category" "cat" {
  name        = "%[1]s_cat"
  description = "acc caseplan deps"
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

// Ensures genesyscloud_user.owner can be default_case_owner (caseManagement:{caseplan,case}:view in home division).
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

func testVerifyCaseManagementCaseplanDestroyed(state *terraform.State) error {
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
