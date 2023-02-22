package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v92/platformclientv2"
)

func TestAccResourceTrunkBaseSettings(t *testing.T) {
	t.Parallel()
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
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: generateTrunkBaseSettingsResourceWithCustomAttrs(
					trunkBaseSettingsRes,
					name1,
					description1,
					trunkMetaBaseId,
					trunkType,
					managed,
					generateTrunkBaseSettingsProperties(
						name1,
						"1m",
						"25",
						falseValue,
						[]string{strconv.Quote("audio/pcmu")}),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "description", description1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "trunk_meta_base_id", trunkMetaBaseId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "trunk_type", trunkType),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "managed", falseValue),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_label", name1),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_max_dial_timeout", "1m"),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_transport_sip_dscp_value", "25"),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_media_disconnect_on_idle_rtp", falseValue),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_media_codec", strings.Join([]string{"audio/pcmu"}, ",")),
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
					generateTrunkBaseSettingsProperties(name2,
						"2m",
						"50",
						trueValue,
						[]string{strconv.Quote("audio/opus")}),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "description", description2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "trunk_meta_base_id", trunkMetaBaseId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "trunk_type", trunkType),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "managed", falseValue),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_label", name2),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_max_dial_timeout", "2m"),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_transport_sip_dscp_value", "50"),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_media_disconnect_on_idle_rtp", trueValue),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_media_codec", strings.Join([]string{"audio/opus"}, ",")),
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
		} else if isStatus404(resp) {
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

func generateTrunkBaseSettingsProperties(settingsName, trunkMaxDialTimeout, trunkTransportSipDscpValue, trunkMediaDisconnectOnIdleRtp string, trunkMediaCodec []string) string {
	// A random selection of properties
	return "properties = " + generateJsonEncodedProperties(
		generateJsonProperty(
			"trunk_label", generateJsonObject(
				generateJsonProperty(
					"value", generateJsonObject(
						generateJsonProperty("instance", strconv.Quote(settingsName)),
					)))),
		generateJsonProperty(
			"trunk_max_dial_timeout", generateJsonObject(
				generateJsonProperty(
					"value", generateJsonObject(
						generateJsonProperty("instance", strconv.Quote(trunkMaxDialTimeout)),
					)))),
		generateJsonProperty(
			"trunk_transport_sip_dscp_value", generateJsonObject(
				generateJsonProperty(
					"value", generateJsonObject(
						generateJsonProperty("instance", trunkTransportSipDscpValue),
					)))),
		generateJsonProperty(
			"trunk_media_disconnect_on_idle_rtp", generateJsonObject(
				generateJsonProperty(
					"value", generateJsonObject(
						generateJsonProperty("instance", trunkMediaDisconnectOnIdleRtp),
					)))),
		generateJsonProperty(
			"trunk_media_codec", generateJsonObject(
				generateJsonProperty(
					"value", generateJsonObject(
						generateJsonArrayProperty("instance", strings.Join(trunkMediaCodec, ",")),
					)))),
	)
}
