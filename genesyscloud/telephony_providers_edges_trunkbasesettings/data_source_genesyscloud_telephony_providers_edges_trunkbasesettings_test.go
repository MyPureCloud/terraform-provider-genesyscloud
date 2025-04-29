package telephony_providers_edges_trunkbasesettings

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceTrunkBaseSettings(t *testing.T) {
	t.Parallel()
	var (
		trunkBaseSettingsResourceLabel     = "trunkBaseSettings"
		trunkBaseSettingsDataResourceLabel = "trunkBaseSettingsData"
		name                               = "test trunk base settings-" + uuid.NewString()
		description                        = "test description"
		trunkMetaBaseId                    = "phone_connections_webrtc.json"
		trunkType                          = "PHONE"
		managed                            = false

		dataResourcePath = "data." + ResourceType + "." + trunkBaseSettingsDataResourceLabel
		resourcePath     = ResourceType + "." + trunkBaseSettingsResourceLabel
	)

	resource.Test(t, resource.TestCase{

		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateTrunkBaseSettingsResourceWithCustomAttrs(
					trunkBaseSettingsResourceLabel,
					name,
					description,
					trunkMetaBaseId,
					trunkType,
					managed,
				) + generateTrunkBaseSettingsDataSource(
					trunkBaseSettingsDataResourceLabel,
					name,
					resourcePath),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataResourcePath, "id", resourcePath, "id"),
				),
			},
		},
	})
}

func generateTrunkBaseSettingsDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_trunkbasesettings" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
