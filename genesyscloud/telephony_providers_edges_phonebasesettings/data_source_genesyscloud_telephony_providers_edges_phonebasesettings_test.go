package telephony_providers_edges_phonebasesettings

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourcePhoneBaseSettings(t *testing.T) {
	t.Parallel()
	var (
		phoneBaseSettingsResourceLabel     = "phoneBaseSettings"
		phoneBaseSettingsDataResourceLabel = "phoneBaseSettingsData"
		name                               = "test phone base settings " + uuid.NewString()
		description                        = "test description"
		phoneMetaBaseId                    = "generic_sip.json"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GeneratePhoneBaseSettingsResourceWithCustomAttrs(
					phoneBaseSettingsResourceLabel,
					name,
					description,
					phoneMetaBaseId,
				) + generatePhoneBaseSettingsDataSource(
					phoneBaseSettingsDataResourceLabel,
					name,
					"genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsDataResourceLabel, "id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "id"),
				),
			},
		},
	})
}
