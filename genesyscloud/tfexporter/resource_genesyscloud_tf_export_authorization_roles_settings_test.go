package tfexporter

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	authorizationRolesSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/authorization_roles_settings"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"
)

func TestUnitAuthorizationRolesSettingsExporterRegistered(t *testing.T) {
	if resourceExporters[authorizationRolesSettings.ResourceType] == nil {
		t.Fatalf("no exporter registered for %s", authorizationRolesSettings.ResourceType)
	}
	if providerResources[authorizationRolesSettings.ResourceType] == nil {
		t.Fatalf("no resource registered for %s", authorizationRolesSettings.ResourceType)
	}
	exporter := authorizationRolesSettings.AuthorizationRolesSettingsExporter()
	if !exporter.IsSingleton {
		t.Fatal("authorization roles settings exporter must be a singleton")
	}
	if exporter.ExportId != authorizationRolesSettings.ResourceType {
		t.Errorf("ExportId = %q, want %q", exporter.ExportId, authorizationRolesSettings.ResourceType)
	}
}

// Read-only export of the org singleton; optionally verifies exported values match a managed resource.
func TestAccResourceTfExportAuthorizationRolesSettings(t *testing.T) {
	testSetup(t)

	exportDir := testrunner.GetTestTempPath(".terraform" + uuid.NewString())
	defer removeExportDir(exportDir)

	resourceType := authorizationRolesSettings.ResourceType

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateAuthorizationRolesSettingsExportConfig(
					exportDir,
					util.TrueValue,
					util.FalseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					assertAuthorizationRolesSettingsExportedBlock(exportDir, resourceType, util.TrueValue, util.FalseValue),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportDir),
	})
}

func generateAuthorizationRolesSettingsExportConfig(exportDir, backfillEnabled, grantDivisionAware string) string {
	return fmt.Sprintf(`
resource "%[1]s" "roles_settings" {
  backfill_enabled                   = %[4]s
  authorization_grant_division_aware = %[5]s
}

resource "genesyscloud_tf_export" "export" {
  directory                = %[2]q
  include_state_file       = false
  export_format            = "json"
  log_permission_errors    = true
  include_filter_resources = [%[3]q]
  depends_on               = [%[1]s.roles_settings]
}
`, authorizationRolesSettings.ResourceType, exportDir, authorizationRolesSettings.ResourceType, backfillEnabled, grantDivisionAware)
}

func assertAuthorizationRolesSettingsExportedBlock(exportDir, resourceType, backfillEnabled, grantDivisionAware string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		blocks, err := getResourceDefinition(filepath.Join(exportDir, defaultTfJSONFile), resourceType)
		if err != nil {
			return err
		}
		if len(blocks) != 1 {
			return fmt.Errorf("expected exactly 1 exported %s block, got %d", resourceType, len(blocks))
		}

		var attrs map[string]interface{}
		for _, raw := range blocks {
			if err := json.Unmarshal(*raw, &attrs); err != nil {
				return fmt.Errorf("failed to unmarshal exported %s block: %w", resourceType, err)
			}
		}

		if got, ok := attrs["backfill_enabled"].(bool); !ok || fmt.Sprintf("%t", got) != backfillEnabled {
			return fmt.Errorf("exported backfill_enabled = %v, want %s", attrs["backfill_enabled"], backfillEnabled)
		}
		if got, ok := attrs["authorization_grant_division_aware"].(bool); !ok || fmt.Sprintf("%t", got) != grantDivisionAware {
			return fmt.Errorf("exported authorization_grant_division_aware = %v, want %s", attrs["authorization_grant_division_aware"], grantDivisionAware)
		}
		if _, ok := attrs["genesys_org_policy_bypass"].(bool); !ok {
			return fmt.Errorf("exported genesys_org_policy_bypass missing or not a bool: %v", attrs["genesys_org_policy_bypass"])
		}
		return nil
	}
}
