package telephony

import (
	"fmt"
	"strconv"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v116/platformclientv2"
)

func TestAccResourceTrunkBaseSettings(t *testing.T) {
	t.Parallel()
	var (
		trunkBaseSettingsRes = "trunkBaseSettings1234" + uuid.NewString()
		name1                = "test trunk base settings " + uuid.NewString()
		name2                = "test trunk base settings " + uuid.NewString()
		description1         = "test description 1"
		description2         = "test description 2"
		trunkMetaBaseId      = "phone_connections_webrtc.json"
		trunkType            = "PHONE"
		managed              = false
		siteId               = "siteTest"
		locationResourceId   = "location"
	)

	referencedResources := gcloud.GenerateLocationResource(
		locationResourceId,
		"tf location "+uuid.NewString(),
		"HQ1",
		[]string{},
		gcloud.GenerateLocationEmergencyNum(
			"+13178791021",
			gcloud.NullValue,
		),
		gcloud.GenerateLocationAddress(
			"7601 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46278",
		),
	) + edgeSite.GenerateSiteResourceWithCustomAttrs(
		siteId,
		"tf site "+uuid.NewString(),
		"test description",
		"genesyscloud_location."+locationResourceId+".id",
		"Cloud",
		false,
		"[\"us-east-1\"]",
		gcloud.NullValue,
		gcloud.NullValue,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: referencedResources + GenerateTrunkBaseSettingsResourceWithCustomAttrs(
					trunkBaseSettingsRes,
					name1,
					description1,
					trunkMetaBaseId,
					"genesyscloud_telephony_providers_edges_site."+siteId+".id",
					trunkType,
					managed,
					generateTrunkBaseSettingsProperties(
						name1,
						"1m",
						"25",
						gcloud.FalseValue,
						[]string{strconv.Quote("audio/pcmu")}),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "description", description1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "trunk_meta_base_id", trunkMetaBaseId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "trunk_type", trunkType),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "inbound_site_id", "genesyscloud_telephony_providers_edges_site."+siteId, "id"),

					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "managed", gcloud.FalseValue),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_label", name1),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_max_dial_timeout", "1m"),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_transport_sip_dscp_value", "25"),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_media_disconnect_on_idle_rtp", gcloud.FalseValue),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_media_codec", strings.Join([]string{"audio/pcmu"}, ",")),
				),
			},
			// Update with new name, description and properties
			{
				Config: referencedResources + GenerateTrunkBaseSettingsResourceWithCustomAttrs(
					trunkBaseSettingsRes,
					name2,
					description2,
					trunkMetaBaseId,
					"genesyscloud_telephony_providers_edges_site."+siteId+".id",
					trunkType,
					managed,
					generateTrunkBaseSettingsProperties(name2,
						"2m",
						"50",
						gcloud.TrueValue,
						[]string{strconv.Quote("audio/opus")}),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "description", description2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "trunk_meta_base_id", trunkMetaBaseId),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "inbound_site_id", "genesyscloud_telephony_providers_edges_site."+siteId, "id"),

					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "trunk_type", trunkType),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "managed", gcloud.FalseValue),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_label", name2),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_max_dial_timeout", "2m"),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_transport_sip_dscp_value", "50"),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_media_disconnect_on_idle_rtp", gcloud.TrueValue),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "properties", "trunk_media_codec", strings.Join([]string{"audio/opus"}, ",")),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_telephony_providers_edges_trunkbasesettings." + trunkBaseSettingsRes,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"inbound_site_id"},
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
		} else if gcloud.IsStatus404(resp) {
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

func generateTrunkBaseSettingsProperties(settingsName, trunkMaxDialTimeout, trunkTransportSipDscpValue, trunkMediaDisconnectOnIdleRtp string, trunkMediaCodec []string) string {
	// A random selection of properties
	return "properties = " + gcloud.GenerateJsonEncodedProperties(
		gcloud.GenerateJsonProperty(
			"trunk_label", gcloud.GenerateJsonObject(
				gcloud.GenerateJsonProperty(
					"value", gcloud.GenerateJsonObject(
						gcloud.GenerateJsonProperty("instance", strconv.Quote(settingsName)),
					)))),
		gcloud.GenerateJsonProperty(
			"trunk_max_dial_timeout", gcloud.GenerateJsonObject(
				gcloud.GenerateJsonProperty(
					"value", gcloud.GenerateJsonObject(
						gcloud.GenerateJsonProperty("instance", strconv.Quote(trunkMaxDialTimeout)),
					)))),
		gcloud.GenerateJsonProperty(
			"trunk_transport_sip_dscp_value", gcloud.GenerateJsonObject(
				gcloud.GenerateJsonProperty(
					"value", gcloud.GenerateJsonObject(
						gcloud.GenerateJsonProperty("instance", trunkTransportSipDscpValue),
					)))),
		gcloud.GenerateJsonProperty(
			"trunk_media_disconnect_on_idle_rtp", gcloud.GenerateJsonObject(
				gcloud.GenerateJsonProperty(
					"value", gcloud.GenerateJsonObject(
						gcloud.GenerateJsonProperty("instance", trunkMediaDisconnectOnIdleRtp),
					)))),
		gcloud.GenerateJsonProperty(
			"trunk_media_codec", gcloud.GenerateJsonObject(
				gcloud.GenerateJsonProperty(
					"value", gcloud.GenerateJsonObject(
						gcloud.GenerateJsonArrayProperty("instance", strings.Join(trunkMediaCodec, ",")),
					)))),
	)
}
