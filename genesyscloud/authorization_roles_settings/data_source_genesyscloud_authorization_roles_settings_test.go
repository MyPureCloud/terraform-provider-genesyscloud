package authorization_roles_settings

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
The data source is a full read of the singleton settings object, so the test asserts that each
attribute it exposes matches the resource managing the same settings.
*/

func TestAccDataSourceAuthorizationRolesSettings(t *testing.T) {
	var (
		resourceLabel   = "roles_settings"
		dataSourceLabel = "roles_settings_data"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateAuthorizationRolesSettingsResource(
					resourceLabel,
					util.TrueValue,
					util.FalseValue,
				) + generateAuthorizationRolesSettingsDataSource(dataSourceLabel, ResourceType+"."+resourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data."+ResourceType+"."+dataSourceLabel, "backfill_enabled",
						ResourceType+"."+resourceLabel, "backfill_enabled",
					),
					resource.TestCheckResourceAttrPair(
						"data."+ResourceType+"."+dataSourceLabel, "authorization_grant_division_aware",
						ResourceType+"."+resourceLabel, "authorization_grant_division_aware",
					),
					resource.TestCheckResourceAttrPair(
						"data."+ResourceType+"."+dataSourceLabel, "genesys_org_policy_bypass",
						ResourceType+"."+resourceLabel, "genesys_org_policy_bypass",
					),
				),
			},
		},
	})
}

func generateAuthorizationRolesSettingsDataSource(dataSourceLabel string, dependsOn string) string {
	return fmt.Sprintf(`
data "%s" "%s" {
	depends_on = [%s]
}
`, ResourceType, dataSourceLabel, dependsOn)
}
