package telephony_providers_edges_trunk

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/telephony"
	edgeGroup "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_edge_group"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceTrunk(t *testing.T) {
	t.Skip("Skipping because we need to manage edges in order to successfully implement and test this resource")
	var (
		// trunk base settings used to create the trunk
		trunkBaseSettingsRes = "trunkBaseSettingsRes"

		//edge groups
		edgeGroupRes1 = "edgeGroupRes1"
		edgeGroupRes2 = "edgeGroupRes2"

		// trunk base settings of type PHONE needed for creating edge groups
		phoneTrunkBaseSettingsRes = "phoneTrunkBaseSettingsRes"

		// trunk
		trunkRes = "trunkRes"
	)

	phoneTrunkBaseSettings := telephony.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		phoneTrunkBaseSettingsRes,
		"phone trunk base settings "+uuid.NewString(),
		"",
		"phone_connections_webrtc.json",
		"PHONE",
		false)

	trunkBaseSettingsConfig := telephony.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		trunkBaseSettingsRes,
		"test trunk base settings "+uuid.NewString(),
		"test description 1",
		"phone_connections_webrtc.json",
		"PHONE",
		false)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			// Create the trunk by creating trunk base settings and an edge group and assigning the trunk base settings to the edge group
			{
				Config: edgeGroup.GenerateEdgeGroupResourceWithCustomAttrs(
					edgeGroupRes1,
					"test edge group "+uuid.NewString(),
					"edge group description 1",
					false,
					false,
					edgeGroup.GeneratePhoneTrunkBaseIds(
						"genesyscloud_telephony_providers_edges_trunkbasesettings."+phoneTrunkBaseSettingsRes+".id",
						"genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes+".id",
					),
				) + phoneTrunkBaseSettings + trunkBaseSettingsConfig + generateTrunk(
					trunkRes,
					"genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes+".id",
					"genesyscloud_telephony_providers_edges_edge_group."+edgeGroupRes1+".id",
				),
			},
			//Create a new edge group and assign the trunk base settings to a new edge group to update the trunk
			{
				Config: edgeGroup.GenerateEdgeGroupResourceWithCustomAttrs(
					edgeGroupRes2,
					"test edge group "+uuid.NewString(),
					"edge group description 2",
					false,
					false,
					edgeGroup.GeneratePhoneTrunkBaseIds(
						"genesyscloud_telephony_providers_edges_trunkbasesettings."+phoneTrunkBaseSettingsRes+".id",
						"genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes+".id",
					),
				) + phoneTrunkBaseSettings + trunkBaseSettingsConfig + generateTrunk(
					trunkRes,
					"genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes+".id",
					"genesyscloud_telephony_providers_edges_edge_group."+edgeGroupRes2+".id",
				),
			},
		},
	})
}
