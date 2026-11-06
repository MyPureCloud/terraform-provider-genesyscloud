package integration_credential

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
Acceptance tests for the write-only credential fields (fields_wo / fields_wo_version), DEVTOOLING-1899.

Write-only attributes need Terraform 1.11 or later, so these tests skip on older CLIs. The framework runs a
plan after every applied step and fails on a non-empty plan, so each step below also proves there is no
perpetual diff.
*/

// skipIfTerraformBelow skips the test when the Terraform CLI used by the acceptance framework is older
// than major.minor. It honors TF_ACC_TERRAFORM_PATH like the framework does.
func skipIfTerraformBelow(t *testing.T, major, minor int) {
	t.Helper()
	bin := os.Getenv("TF_ACC_TERRAFORM_PATH")
	if bin == "" {
		bin = "terraform"
	}
	out, err := exec.Command(bin, "version", "-json").Output()
	if err != nil {
		t.Skipf("could not determine the Terraform CLI version: %v", err)
	}
	var v struct {
		Version string `json:"terraform_version"`
	}
	if err := json.Unmarshal(out, &v); err != nil {
		t.Skipf("could not parse the Terraform CLI version: %v", err)
	}
	parts := strings.Split(v.Version, ".")
	if len(parts) < 2 {
		t.Skipf("unexpected Terraform CLI version %q", v.Version)
	}
	gotMajor, err1 := strconv.Atoi(parts[0])
	gotMinor, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		t.Skipf("unexpected Terraform CLI version %q", v.Version)
	}
	if gotMajor < major || (gotMajor == major && gotMinor < minor) {
		t.Skipf("write-only attributes need Terraform %d.%d or later, found %s", major, minor, v.Version)
	}
}

// generateFieldsWoAttrs builds the fields_wo / fields_wo_version attributes for a basicAuth credential.
func generateFieldsWoAttrs(userName, password string, version int) string {
	return fmt.Sprintf("fields_wo = jsonencode({ userName = %s, password = %s })\n\tfields_wo_version = %d",
		strconv.Quote(userName), strconv.Quote(password), version)
}

// checkSecretsNotInState fails if any of the secrets appears in any attribute value of the resource in state.
func checkSecretsNotInState(resourcePath string, secrets ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourcePath)
		}
		for key, val := range rs.Primary.Attributes {
			for _, secret := range secrets {
				if secret != "" && strings.Contains(val, secret) {
					return fmt.Errorf("a secret was found in state under attribute %q", key)
				}
			}
		}
		return nil
	}
}

func TestAccResourceCredentialWriteOnlyFields(t *testing.T) {
	skipIfTerraformBelow(t, 1, 11)

	var (
		label     = "test_credential_wo"
		path      = "genesyscloud_integration_credential." + label
		credName  = "Terraform Credential WO Test-" + uuid.NewString()
		credName2 = "Terraform Credential WO Test Renamed-" + uuid.NewString()
		typeName  = "basicAuth"
		userName  = "wo-user-" + uuid.NewString()

		secretV1 = "wo-secret-v1-" + uuid.NewString()
		secretV2 = "wo-secret-v2-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with the write-only fields. The secret must not be stored in state.
				Config: GenerateCredentialResource(label, strconv.Quote(credName), strconv.Quote(typeName),
					generateFieldsWoAttrs(userName, secretV1, 1)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "name", credName),
					resource.TestCheckResourceAttr(path, "credential_type_name", typeName),
					resource.TestCheckResourceAttr(path, "fields_wo_version", "1"),
					resource.TestCheckNoResourceAttr(path, "fields_wo"),
					resource.TestCheckNoResourceAttr(path, "fields.password"),
					checkSecretsNotInState(path, secretV1),
				),
			},
			{
				// Rotate the secret and bump the version: the update is applied and no secret reaches state.
				Config: GenerateCredentialResource(label, strconv.Quote(credName), strconv.Quote(typeName),
					generateFieldsWoAttrs(userName, secretV2, 2)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "fields_wo_version", "2"),
					resource.TestCheckNoResourceAttr(path, "fields.password"),
					checkSecretsNotInState(path, secretV1, secretV2),
				),
			},
			{
				// Rename only: the update replaces the credential, so the secret must still be sent from
				// the config (it is not in state). Apply succeeding without the API rejecting an empty
				// credential proves it was re-sent.
				Config: GenerateCredentialResource(label, strconv.Quote(credName2), strconv.Quote(typeName),
					generateFieldsWoAttrs(userName, secretV2, 2)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "name", credName2),
					checkSecretsNotInState(path, secretV1, secretV2),
				),
			},
			{
				// Import: write-only values and the version are not in the imported state.
				ResourceName:            path,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"fields_wo", "fields_wo_version"},
			},
		},
		CheckDestroy: testVerifyCredentialDestroyed,
	})
}

func TestAccResourceCredentialWriteOnlyConflicts(t *testing.T) {
	skipIfTerraformBelow(t, 1, 11)

	label := "test_credential_wo_invalid"
	name := strconv.Quote("Terraform Credential WO Invalid-" + uuid.NewString())
	typeName := strconv.Quote("basicAuth")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateCredentialResource(label, name, typeName,
					GenerateCredentialFields(map[string]string{"userName": strconv.Quote("u")})+"\n\t"+
						generateFieldsWoAttrs("u", "p", 1)),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)Conflicting configuration arguments`),
			},
			{
				Config:      GenerateCredentialResource(label, name, typeName, "fields_wo = jsonencode({ userName = \"u\" })"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)Missing required argument`),
			},
			{
				Config:      GenerateCredentialResource(label, name, typeName, "fields_wo = \"{\"\nfields_wo_version = 1"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)fields_wo must be a valid JSON object`),
			},
		},
	})
}

// TestAccResourceCredentialMigrateFieldsToWriteOnly proves that moving an existing credential from the
// stored `fields` map to the write-only `fields_wo` also removes the old plaintext secret from state.
func TestAccResourceCredentialMigrateFieldsToWriteOnly(t *testing.T) {
	skipIfTerraformBelow(t, 1, 11)

	var (
		label    = "test_credential_migrate"
		path     = "genesyscloud_integration_credential." + label
		credName = "Terraform Credential Migrate Test-" + uuid.NewString()
		typeName = "basicAuth"
		userName = "migrate-user-" + uuid.NewString()

		oldSecret = "migrate-old-" + uuid.NewString()
		newSecret = "migrate-new-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Existing behaviour: the secret supplied through `fields` is stored in state.
				Config: GenerateCredentialResource(label, strconv.Quote(credName), strconv.Quote(typeName),
					GenerateCredentialFields(map[string]string{
						"userName": strconv.Quote(userName),
						"password": strconv.Quote(oldSecret),
					})),
				Check: resource.TestCheckResourceAttr(path, "fields.password", oldSecret),
			},
			{
				// Migrate to the write-only fields: neither the old nor the new secret may remain in state.
				Config: GenerateCredentialResource(label, strconv.Quote(credName), strconv.Quote(typeName),
					generateFieldsWoAttrs(userName, newSecret, 1)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "fields_wo_version", "1"),
					resource.TestCheckNoResourceAttr(path, "fields.password"),
					checkSecretsNotInState(path, oldSecret, newSecret),
				),
			},
		},
		CheckDestroy: testVerifyCredentialDestroyed,
	})
}
