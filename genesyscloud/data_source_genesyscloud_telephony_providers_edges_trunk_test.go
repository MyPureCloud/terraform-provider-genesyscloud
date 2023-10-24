package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceTrunk(t *testing.T) {
	t.Skip("Skipping because we need to manage edges in order to successfully implement and test this resource")
	var (
		// trunk base settings used to create the trunk
		trunkBaseSettingsRes = "trunkBaseSettingsRes"

		//edge groups
		edgeGroupRes1 = "edgeGroupRes1"

		// trunk base settings of type PHONE needed for creating edge groups
		phoneTrunkBaseSettingsRes = "phoneTrunkBaseSettingsRes"

		// trunk
		trunkRes  = "trunkRes"
		trunkData = "trunkData"
	)

	phoneTrunkBaseSettings := GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		phoneTrunkBaseSettingsRes,
		"phone trunk base settings "+uuid.NewString(),
		"",
		"phone_connections_webrtc.json",
		"PHONE",
		false)

	trunkBaseSettingsConfig := GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		trunkBaseSettingsRes,
		"test trunk base settings "+uuid.NewString(),
		"test description 1",
		"phone_connections_webrtc.json",
		"PHONE",
		false)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateEdgeGroupResourceWithCustomAttrs(
					edgeGroupRes1,
					"test edge group "+uuid.NewString(),
					"edge group description 1",
					false,
					false,
					generatePhoneTrunkBaseIds("genesyscloud_telephony_providers_edges_trunkbasesettings."+phoneTrunkBaseSettingsRes+".id"),
				) + phoneTrunkBaseSettings + trunkBaseSettingsConfig + generateTrunk(
					trunkRes,
					"genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes+".id",
					"genesyscloud_telephony_providers_edges_edge_group."+edgeGroupRes1+".id",
				) + generateTrunkDataSource(
					trunkData,
					"genesyscloud_telephony_providers_edges_trunk."+trunkRes+".name",
					"genesyscloud_telephony_providers_edges_trunk."+trunkRes,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_trunk."+trunkData, "id", "genesyscloud_telephony_providers_edges_trunk."+trunkRes, "id"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func generateTrunkDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_trunk" "%s" {
		name = %s
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
