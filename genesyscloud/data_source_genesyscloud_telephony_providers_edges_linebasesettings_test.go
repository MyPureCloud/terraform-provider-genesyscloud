package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceLineBaseSettings(t *testing.T) {
	t.Parallel()
	phoneBaseSettingsRes := "phoneBaseSettings1234"
	phoneBaseSettingsName := "phoneBaseSettings " + uuid.NewString()

	lineBaseSettingsDataRes := "lineBaseSettings1234"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Creating a phone base settings will result in a line base settings of the same name being created
				Config: GeneratePhoneBaseSettingsResourceWithCustomAttrs(
					phoneBaseSettingsRes,
					phoneBaseSettingsName,
					"phoneBaseSettings description",
					"generic_sip.json",
				) + generateLineBaseSettingsDataSource(
					lineBaseSettingsDataRes,
					phoneBaseSettingsName,
					"genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes,
				),
				Check: resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func generateLineBaseSettingsDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_linebasesettings" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
