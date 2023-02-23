package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceTrunkBaseSettings(t *testing.T) {
	t.Parallel()
	var (
		trunkBaseSettingsRes     = "trunkBaseSettings"
		trunkBaseSettingsDataRes = "trunkBaseSettingsData"
		name                     = "test trunk base settings " + uuid.NewString()
		description              = "test description"
		trunkMetaBaseId          = "external_sip_pcv_byoc_carrier.json"
		trunkType                = "EXTERNAL"
		managed                  = false
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: generateTrunkBaseSettingsResourceWithCustomAttrs(
					trunkBaseSettingsRes,
					name,
					description,
					trunkMetaBaseId,
					trunkType,
					managed,
				) + generateTrunkBaseSettingsDataSource(
					trunkBaseSettingsDataRes,
					name,
					"genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsDataRes, "id", "genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "id"),
				),
			},
		},
	})
}

func generateTrunkBaseSettingsDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_trunkbasesettings" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
