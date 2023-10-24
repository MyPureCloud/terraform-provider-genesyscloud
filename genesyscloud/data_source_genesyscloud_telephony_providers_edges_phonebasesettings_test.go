package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourcePhoneBaseSettings(t *testing.T) {
	t.Parallel()
	var (
		phoneBaseSettingsRes     = "phoneBaseSettings"
		phoneBaseSettingsDataRes = "phoneBaseSettingsData"
		name                     = "test phone base settings " + uuid.NewString()
		description              = "test description"
		phoneMetaBaseId          = "generic_sip.json"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GeneratePhoneBaseSettingsResourceWithCustomAttrs(
					phoneBaseSettingsRes,
					name,
					description,
					phoneMetaBaseId,
				) + generatePhoneBaseSettingsDataSource(
					phoneBaseSettingsDataRes,
					name,
					"genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsDataRes, "id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "id"),
				),
			},
		},
	})
}

func generatePhoneBaseSettingsDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_phonebasesettings" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
