package telephony_providers_edges_edge_group

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/telephony"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceEdgeGroup(t *testing.T) {
	t.Parallel()
	var (
		edgeGroupRes          = "edgeGroup1234"
		edgeGroupData         = "edgeGroupData"
		edgeGroupName1        = "test edge group " + uuid.NewString()
		edgeGroupDescription1 = "test description 1"

		phoneTrunkBaseSettingsRes1 = "phoneTrunkBaseSettingsRes1"
		phoneTrunkBaseSettingsRes2 = "phoneTrunkBaseSettingsRes2"
	)

	phoneTrunkBaseSetting1 := telephony.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		phoneTrunkBaseSettingsRes1,
		"phone trunk base settings "+uuid.NewString(),
		"",
		"phone_connections_webrtc.json",
		"PHONE",
		false)
	phoneTrunkBaseSetting2 := telephony.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		phoneTrunkBaseSettingsRes2,
		"phone trunk base settings "+uuid.NewString(),
		"",
		"phone_connections_webrtc.json",
		"PHONE",
		false)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: phoneTrunkBaseSetting1 + phoneTrunkBaseSetting2 + GenerateEdgeGroupResourceWithCustomAttrs(
					edgeGroupRes,
					edgeGroupName1,
					edgeGroupDescription1,
					false,
					false,
					GeneratePhoneTrunkBaseIds("genesyscloud_telephony_providers_edges_trunkbasesettings."+phoneTrunkBaseSettingsRes1+".id",
						"genesyscloud_telephony_providers_edges_trunkbasesettings."+phoneTrunkBaseSettingsRes2+".id"),
				) + generateEdgeGroupDataSource(
					edgeGroupData,
					edgeGroupName1,
					"genesyscloud_telephony_providers_edges_edge_group."+edgeGroupRes,
					false,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_edge_group."+edgeGroupData, "id", "genesyscloud_telephony_providers_edges_edge_group."+edgeGroupRes, "id"),
				),
			},
		},
	})
}

/*
This test expects that the org has a product called "voice" enabled on it. If the test org does not have this product on it, the test can be skipped or ignored.
*/
func TestAccDataSourceEdgeGroupManaged(t *testing.T) {
	t.Parallel()
	var (
		edgeGroupData  = "edgeGroupData"
		edgeGroupName1 = "PureCloud Voice - AWS"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateEdgeGroupDataSource(
					edgeGroupData,
					edgeGroupName1,
					"",
					true,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.genesyscloud_telephony_providers_edges_edge_group."+edgeGroupData, "name", edgeGroupName1),
				),
			},
		},
	})
}

func generateEdgeGroupDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string,
	managed bool) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_edge_group" "%s" {
		name = "%s"
		managed = %t
		depends_on=[%s]
	}
	`, resourceID, name, managed, dependsOnResource)
}
