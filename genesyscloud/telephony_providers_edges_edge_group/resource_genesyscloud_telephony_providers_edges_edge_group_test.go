package telephony_providers_edges_edge_group

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	tbs "terraform-provider-genesyscloud/genesyscloud/telephony_provider_edges_trunkbasesettings"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

func TestAccResourceEdgeGroup(t *testing.T) {
	t.Parallel()
	var (
		edgeGroupResourceLabel = "edgeGroup1234"
		edgeGroupName1         = "test edge group " + uuid.NewString()
		edgeGroupName2         = "test edge group " + uuid.NewString()
		edgeGroupDescription1  = "test description 1"
		edgeGroupDescription2  = "test description 2"

		phoneTrunkBaseSettingsResourceLabel1 = "phoneTrunkBaseSettingsRes1"
		phoneTrunkBaseSettingsResourceLabel2 = "phoneTrunkBaseSettingsRes2"
	)

	// Original phone settings
	phoneTrunkBaseSetting1 := tbs.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		phoneTrunkBaseSettingsResourceLabel1,
		"phone trunk base settings "+uuid.NewString(),
		"",
		"phone_connections_webrtc.json",
		"PHONE",
		false)

	// Updated phone settings
	phoneTrunkBaseSetting2 := tbs.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		phoneTrunkBaseSettingsResourceLabel2,
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
				Config: phoneTrunkBaseSetting1 + GenerateEdgeGroupResourceWithCustomAttrs(
					edgeGroupResourceLabel,
					edgeGroupName1,
					edgeGroupDescription1,
					false,
					false,
					GeneratePhoneTrunkBaseIds("genesyscloud_telephony_providers_edges_trunkbasesettings."+phoneTrunkBaseSettingsResourceLabel1+".id")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupResourceLabel, "name", edgeGroupName1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupResourceLabel, "description", edgeGroupDescription1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupResourceLabel, "managed", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupResourceLabel, "hybrid", util.FalseValue),
				),
			},
			// Update with new name, description and phone trunk base
			{
				Config: phoneTrunkBaseSetting1 + phoneTrunkBaseSetting2 + GenerateEdgeGroupResourceWithCustomAttrs(
					edgeGroupResourceLabel,
					edgeGroupName2,
					edgeGroupDescription2,
					false,
					false,
					GeneratePhoneTrunkBaseIds("genesyscloud_telephony_providers_edges_trunkbasesettings."+phoneTrunkBaseSettingsResourceLabel2+".id")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupResourceLabel, "name", edgeGroupName2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupResourceLabel, "description", edgeGroupDescription2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupResourceLabel, "managed", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupResourceLabel, "hybrid", util.FalseValue),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_edge_group."+edgeGroupResourceLabel, "phone_trunk_base_ids.0",
						"genesyscloud_telephony_providers_edges_trunkbasesettings."+phoneTrunkBaseSettingsResourceLabel2, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_edge_group." + edgeGroupResourceLabel,
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
