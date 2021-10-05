package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v55/platformclientv2"
	"strings"
	"testing"
)

func TestAccResourceTrunkBaseSettings(t *testing.T) {
	var (
		trunkBaseSettingsRes = "trunkBaseSettings1234"
		name1                = "test trunk base settings " + uuid.NewString()
		name2                = "test trunk base settings " + uuid.NewString()
		description1         = "test description 1"
		description2         = "test description 2"
		trunkMetaBaseId      = "phone_connections_webrtc.json"
		trunkType            = "PHONE"
		managed              = false
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateTrunkBaseSettingsResourceWithCustomAttrs(
					trunkBaseSettingsRes,
					name1,
					description1,
					trunkMetaBaseId,
					trunkType,
					managed,
					generateTrunkBaseSettingsProperties(name1, "1m", "audio/pcmu", false, 25),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "description", description1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "trunk_meta_base_id", trunkMetaBaseId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "trunk_type", trunkType),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "managed", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties.0.trunk_max_dial_timeout", "1m"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties.0.trunk_media_codec.0", "audio/pcmu"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties.0.trunk_media_disconnect_on_idle_rtp", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties.0.trunk_transport_sip_dscp_value", "25"),
				),
			},
			// Update with new name, description and properties
			{
				Config: generateTrunkBaseSettingsResourceWithCustomAttrs(
					trunkBaseSettingsRes,
					name2,
					description2,
					trunkMetaBaseId,
					trunkType,
					managed,
					generateTrunkBaseSettingsProperties(name2, "2m", "audio/opus", true, 50),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "description", description2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "trunk_meta_base_id", trunkMetaBaseId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "trunk_type", trunkType),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "managed", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties.0.trunk_max_dial_timeout", "2m"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties.0.trunk_media_codec.0", "audio/opus"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties.0.trunk_media_disconnect_on_idle_rtp", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties.0.trunk_transport_sip_dscp_value", "50"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_trunkbasesettings." + trunkBaseSettingsRes,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyTrunkBaseSettingsDestroyed,
	})
}

func testVerifyTrunkBaseSettingsDestroyed(state *terraform.State) error {
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_telephony_providers_edges_trunkbasesettings" {
			continue
		}

		trunkBaseSettings, resp, err := edgesAPI.GetTelephonyProvidersEdgesTrunkbasesetting(rs.Primary.ID, true)
		if trunkBaseSettings != nil {
			return fmt.Errorf("TrunkBaseSettings (%s) still exists", rs.Primary.ID)
		} else if resp != nil && resp.StatusCode == 404 {
			// TrunkBaseSettings not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	//Success. TrunkBaseSettings destroyed
	return nil
}

func generateTrunkBaseSettingsResourceWithCustomAttrs(
	trunkBaseSettingsRes,
	name,
	description,
	trunkMetaBaseId,
	trunkType string,
	managed bool,
	otherAttrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_telephony_providers_edges_trunkbasesettings" "%s" {
		name = "%s"
		description = "%s"
		trunk_meta_base_id = "%s"
		trunk_type = "%s"
		managed = %v
		%s
	}
	`, trunkBaseSettingsRes, name, description, trunkMetaBaseId, trunkType, managed, strings.Join(otherAttrs, "\n"))
}

func generateTrunkBaseSettingsProperties(settingsName, trunkMaxDialTimeout, trunkMediaCodec string, trunkMediaDisconnectOnIdleRtp bool, trunkTransportSipDscpValue int) string {
	// A random selection of properties
	return fmt.Sprintf(`properties {
			trunk_label = "%s"
			trunk_max_dial_timeout = "%s"
			trunk_transport_sip_dscp_value = %d
			trunk_media_codec = ["%s"]
			trunk_media_disconnect_on_idle_rtp = %v
		}`, settingsName, trunkMaxDialTimeout, trunkTransportSipDscpValue, trunkMediaCodec, trunkMediaDisconnectOnIdleRtp)
}
