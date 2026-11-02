package authorization_roles_settings

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
NOTE: genesyscloud_authorization_roles_settings is a singleton, organization-wide resource.
These acceptance tests mutate real org-level authorization role settings. There is no
CheckDestroy because the settings object always exists and cannot be deleted.

The OAuth client used for these tests needs directory:organization:admin, or both
authorization:settings:view and authorization:settings:edit. Without the edit permission the
update step fails with a 403 that reads like a provider bug.

These tests never write genesys_org_policy_bypass. Setting it to false on a real organization
stops paired Genesys Services organizations from bypassing ABAC policy enforcement and can cut
off Genesys support access. The tests assert instead that the flag is left alone, which is the
behaviour that actually needs protecting.
*/

func TestAccResourceAuthorizationRolesSettings(t *testing.T) {
	var (
		resourceLabel = "roles_settings"
		resourcePath  = ResourceType + "." + resourceLabel

		// Captured from the first apply so later steps can assert it never moved.
		bypassBefore string
	)

	captureBypass := func(target *string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources[resourcePath]
			if !ok {
				return fmt.Errorf("resource %s not found in state", resourcePath)
			}
			*target = rs.Primary.Attributes["genesys_org_policy_bypass"]
			if *target == "" {
				return fmt.Errorf("genesys_org_policy_bypass was not populated in state")
			}
			return nil
		}
	}

	assertBypassUnchanged := func(expected *string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources[resourcePath]
			if !ok {
				return fmt.Errorf("resource %s not found in state", resourcePath)
			}
			if got := rs.Primary.Attributes["genesys_org_policy_bypass"]; got != *expected {
				return fmt.Errorf("genesys_org_policy_bypass changed from %q to %q across an apply that did not configure it", *expected, got)
			}
			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateAuthorizationRolesSettingsResource(
					resourceLabel,
					util.TrueValue,
					util.FalseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "backfill_enabled", util.TrueValue),
					resource.TestCheckResourceAttr(resourcePath, "authorization_grant_division_aware", util.FalseValue),
					resource.TestCheckResourceAttrSet(resourcePath, "genesys_org_policy_bypass"),
					captureBypass(&bypassBefore),
				),
			},
			{
				// Update
				Config: generateAuthorizationRolesSettingsResource(
					resourceLabel,
					util.FalseValue,
					util.TrueValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "backfill_enabled", util.FalseValue),
					resource.TestCheckResourceAttr(resourcePath, "authorization_grant_division_aware", util.TrueValue),
					assertBypassUnchanged(&bypassBefore),
				),
			},
			{
				// Partial config: backfill_enabled is left out, so it must keep the value the
				// previous step wrote rather than being reset by the full-replace PUT. The
				// bypass flag, which this config never mentions, must likewise be untouched.
				Config: generateAuthorizationRolesSettingsResourcePartial(
					resourceLabel,
					util.TrueValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "backfill_enabled", util.FalseValue),
					resource.TestCheckResourceAttr(resourcePath, "authorization_grant_division_aware", util.TrueValue),
					assertBypassUnchanged(&bypassBefore),
				),
			},
			{
				// Import/Read
				ResourceName:      resourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func generateAuthorizationRolesSettingsResource(resourceLabel string, backfillEnabled string, grantDivisionAware string) string {
	return fmt.Sprintf(`
resource "%s" "%s" {
	backfill_enabled                   = %s
	authorization_grant_division_aware = %s
}
`, ResourceType, resourceLabel, backfillEnabled, grantDivisionAware)
}

func generateAuthorizationRolesSettingsResourcePartial(resourceLabel string, grantDivisionAware string) string {
	return fmt.Sprintf(`
resource "%s" "%s" {
	authorization_grant_division_aware = %s
}
`, ResourceType, resourceLabel, grantDivisionAware)
}
