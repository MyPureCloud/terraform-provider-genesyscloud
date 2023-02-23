package genesyscloud

import (
	"fmt"
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

	phoneTrunkBaseSetting1 := generateTrunkBaseSettingsResourceWithCustomAttrs(
		phoneTrunkBaseSettingsRes1,
		"phone trunk base settings "+uuid.NewString(),
		"",
		"phone_connections_webrtc.json",
		"PHONE",
		false)
	phoneTrunkBaseSetting2 := generateTrunkBaseSettingsResourceWithCustomAttrs(
		phoneTrunkBaseSettingsRes2,
		"phone trunk base settings "+uuid.NewString(),
		"",
		"phone_connections_webrtc.json",
		"PHONE",
		false)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: phoneTrunkBaseSetting1 + phoneTrunkBaseSetting2 + generateEdgeGroupResourceWithCustomAttrs(
					edgeGroupRes,
					edgeGroupName1,
					edgeGroupDescription1,
					false,
					false,
					generatePhoneTrunkBaseIds("genesyscloud_telephony_providers_edges_trunkbasesettings."+phoneTrunkBaseSettingsRes1+".id",
						"genesyscloud_telephony_providers_edges_trunkbasesettings."+phoneTrunkBaseSettingsRes2+".id"),
				) + generateEdgeGroupDataSource(
					edgeGroupData,
					edgeGroupName1,
					"genesyscloud_telephony_providers_edges_edge_group."+edgeGroupRes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_edge_group."+edgeGroupData, "id", "genesyscloud_telephony_providers_edges_edge_group."+edgeGroupRes, "id"),
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
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_edge_group" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
