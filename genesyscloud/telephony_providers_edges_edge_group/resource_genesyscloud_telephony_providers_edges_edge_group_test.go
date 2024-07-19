package telephony_providers_edges_edge_group

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	telephony "terraform-provider-genesyscloud/genesyscloud/telephony"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceEdgeGroup(t *testing.T) {
	t.Parallel()
	var (
		edgeGroupRes          = "edgeGroup1234"
		edgeGroupName1        = "test edge group " + uuid.NewString()
		edgeGroupName2        = "test edge group " + uuid.NewString()
		edgeGroupDescription1 = "test description 1"
		edgeGroupDescription2 = "test description 2"

		phoneTrunkBaseSettingsRes1 = "phoneTrunkBaseSettingsRes1"
		phoneTrunkBaseSettingsRes2 = "phoneTrunkBaseSettingsRes2"
		phoneTrunkBaseSettingsRes3 = "phoneTrunkBaseSettingsRes3"
	)

	// Original phone settings
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

	// Updated phone settings
	phoneTrunkBaseSetting3 := telephony.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		phoneTrunkBaseSettingsRes3,
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
						"genesyscloud_telephony_providers_edges_trunkbasesettings."+phoneTrunkBaseSettingsRes2+".id")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupRes, "name", edgeGroupName1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupRes, "description", edgeGroupDescription1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupRes, "managed", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupRes, "hybrid", util.FalseValue),
				),
			},
			// Update with new name, description and phone trunk base
			{
				Config: phoneTrunkBaseSetting1 + phoneTrunkBaseSetting2 + phoneTrunkBaseSetting3 + GenerateEdgeGroupResourceWithCustomAttrs(
					edgeGroupRes,
					edgeGroupName2,
					edgeGroupDescription2,
					false,
					false,
					GeneratePhoneTrunkBaseIds("genesyscloud_telephony_providers_edges_trunkbasesettings."+phoneTrunkBaseSettingsRes3+".id")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupRes, "name", edgeGroupName2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupRes, "description", edgeGroupDescription2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupRes, "managed", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupRes, "hybrid", util.FalseValue),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupRes, "phone_trunk_base_ids.0",
						"genesyscloud_telephony_providers_edges_trunkbasesettings."+phoneTrunkBaseSettingsRes3, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_edge_group." + edgeGroupRes,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyEdgeGroupsDestroyed,
	})
}

func testVerifyEdgeGroupsDestroyed(state *terraform.State) error {
	edgeAPI := platformclientv2.NewTelephonyProvidersEdgeApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_telephony_providers_edges_edge_group" {
			continue
		}

		edgeGroup, resp, err := edgeAPI.GetTelephonyProvidersEdgesEdgegroup(rs.Primary.ID, nil)
		if edgeGroup != nil {
			return fmt.Errorf("edge group (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// edge group not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All edge groups destroyed
	return nil
}
