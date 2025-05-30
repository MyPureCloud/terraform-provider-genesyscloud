package telephony_providers_edges_linebasesettings

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	phoneBaseSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phonebasesettings"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceLineBaseSettings(t *testing.T) {
	t.Parallel()
	phoneBaseSettingsResourceLabel := "phoneBaseSettings1234"
	phoneBaseSettingsName := "phoneBaseSettings " + uuid.NewString()

	lineBaseSettingsDataResourceLabel := "lineBaseSettings1234"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Creating a phone base settings will result in a line base settings of the same name being created
				Config: phoneBaseSettings.GeneratePhoneBaseSettingsResourceWithCustomAttrs(
					phoneBaseSettingsResourceLabel,
					phoneBaseSettingsName,
					"phoneBaseSettings description",
					"generic_sip.json",
				) + generateLineBaseSettingsDataSource(
					lineBaseSettingsDataResourceLabel,
					phoneBaseSettingsName,
					"genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func generateLineBaseSettingsDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_linebasesettings" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
