package telephony_providers_edges_trunkbasesettings

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	location "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/location"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	edgeSite "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func TestAccResourceTrunkBaseSettings(t *testing.T) {
	t.Parallel()
	var (
		trunkBaseSettingsResourceLabel = "trunkBaseSettings1234"
		name1                          = "test trunk base settings " + uuid.NewString()
		name2                          = "test trunk base settings " + uuid.NewString()
		description1                   = "test description 1"
		description2                   = "test description 2"
		trunkMetaBaseId                = "phone_connections_webrtc.json"
		trunkType                      = "PHONE"
		managed                        = false
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateTrunkBaseSettingsResourceWithCustomAttrs(
					trunkBaseSettingsResourceLabel,
					name1,
					description1,
					trunkMetaBaseId,
					trunkType,
					managed,
					//GenerateTrunkBaseSettingsInboundSite("InboundSiteTest"),
					generateTrunkBaseSettingsProperties(
						name1,
						"1m",
						"25",
						util.FalseValue,
						[]string{strconv.Quote("audio/pcmu")}),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "description", description1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "trunk_meta_base_id", trunkMetaBaseId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "trunk_type", trunkType),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "managed", util.FalseValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_label", name1),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_max_dial_timeout", "1m"),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_transport_sip_dscp_value", "25"),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_media_disconnect_on_idle_rtp", util.FalseValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_media_codec", strings.Join([]string{"audio/pcmu"}, ",")),
				),
			},
			// Update with new name, description and properties
			{
				Config: GenerateTrunkBaseSettingsResourceWithCustomAttrs(
					trunkBaseSettingsResourceLabel,
					name2,
					description2,
					trunkMetaBaseId,
					trunkType,
					managed,
					generateTrunkBaseSettingsProperties(name2,
						"2m",
						"50",
						util.TrueValue,
						[]string{strconv.Quote("audio/opus")}),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "description", description2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "trunk_meta_base_id", trunkMetaBaseId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "trunk_type", trunkType),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "managed", util.FalseValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_label", name2),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_max_dial_timeout", "2m"),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_transport_sip_dscp_value", "50"),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_media_disconnect_on_idle_rtp", util.TrueValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_media_codec", strings.Join([]string{"audio/opus"}, ",")),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_trunkbasesettings." + trunkBaseSettingsResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyTrunkBaseSettingsDestroyed,
	})
}

func TestAccResourceExternalTrunkBaseSettingsInboundSite(t *testing.T) {

	var (
		trunkBaseSettingsResourceLabel = "trunkBaseSettings1234"
		name1                          = "test trunk base settings " + uuid.NewString()
		name2                          = "test trunk base settings " + uuid.NewString()
		description1                   = "test description 1"
		description2                   = "test description 2"
		trunkMetaBaseId                = "external_sip_pcv_byoc_carrier.json"
		trunkType                      = "EXTERNAL"
		managed                        = false
		locationResourceLabel          = "locationtest2"
		siteResourceLabel              = "sitetest2"
	)
	referencedResources :=
		location.GenerateLocationResource(
			locationResourceLabel,
			"tf location "+uuid.NewString(),
			"HQ",
			[]string{},
			location.GenerateLocationEmergencyNum(
				"+13100000003",
				util.NullValue,
			),
			location.GenerateLocationAddress(
				"7601 Interactive Way",
				"Orlando",
				"FL",
				"US",
				"32826",
			),
		) + edgeSite.GenerateSiteResourceWithCustomAttrs(
			siteResourceLabel,
			"tf site "+uuid.NewString(),
			"test description",
			"genesyscloud_location."+locationResourceLabel+".id",
			"Cloud",
			false,
			"[\"us-east-1\"]",
			util.NullValue,
			util.NullValue,
		)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: referencedResources + GenerateTrunkBaseSettingsResourceWithCustomAttrs(
					trunkBaseSettingsResourceLabel,
					name1,
					description1,
					trunkMetaBaseId,
					trunkType,
					managed,
					GenerateTrunkBaseSettingsInboundSite(siteResourceLabel),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "description", description1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "trunk_meta_base_id", trunkMetaBaseId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "trunk_type", trunkType),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "managed", util.FalseValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_label", name1),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_maxDialTimeout", "2m"),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_media_disconnectOnIdleRTP", util.TrueValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_media_codec", "audio/opus,audio/pcmu,audio/pcma")),
			},
			// Update with new name, description and properties
			{
				Config: referencedResources + GenerateTrunkBaseSettingsResourceWithCustomAttrs(
					trunkBaseSettingsResourceLabel,
					name2,
					description2,
					trunkMetaBaseId,
					trunkType,
					managed,
					GenerateTrunkBaseSettingsInboundSite(siteResourceLabel),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "description", description2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "trunk_meta_base_id", trunkMetaBaseId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "trunk_type", trunkType),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "managed", util.FalseValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_label", name2),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_maxDialTimeout", "2m"),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_media_disconnectOnIdleRTP", util.TrueValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsResourceLabel, "properties", "trunk_media_codec", "audio/opus,audio/pcmu,audio/pcma")),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_trunkbasesettings." + trunkBaseSettingsResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: location.GenerateLocationResource(
					locationResourceLabel,
					"tf location "+uuid.NewString(),
					"HQ",
					[]string{},
					location.GenerateLocationEmergencyNum(
						"+13100000003",
						util.NullValue,
					),
					location.GenerateLocationAddress(
						"7601 Interactive Way",
						"Orlando",
						"FL",
						"US",
						"32826",
					),
				) + edgeSite.GenerateSiteResourceWithCustomAttrs(
					siteResourceLabel,
					"tf site "+uuid.NewString(),
					"test description",
					"genesyscloud_location."+locationResourceLabel+".id",
					"Cloud",
					false,
					"[\"us-east-1\"]",
					util.NullValue,
					util.NullValue,
				),
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
		} else if util.IsStatus404(resp) {
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
	return "properties = " + util.GenerateJsonEncodedProperties(
		util.GenerateJsonProperty(
			"trunk_label", util.GenerateJsonObject(
				util.GenerateJsonProperty(
					"value", util.GenerateJsonObject(
						util.GenerateJsonProperty("instance", strconv.Quote(settingsName)),
					)))),
		util.GenerateJsonProperty(
			"trunk_max_dial_timeout", util.GenerateJsonObject(
				util.GenerateJsonProperty(
					"value", util.GenerateJsonObject(
						util.GenerateJsonProperty("instance", strconv.Quote(trunkMaxDialTimeout)),
					)))),
		util.GenerateJsonProperty(
			"trunk_transport_sip_dscp_value", util.GenerateJsonObject(
				util.GenerateJsonProperty(
					"value", util.GenerateJsonObject(
						util.GenerateJsonProperty("instance", trunkTransportSipDscpValue),
					)))),
		util.GenerateJsonProperty(
			"trunk_media_disconnect_on_idle_rtp", util.GenerateJsonObject(
				util.GenerateJsonProperty(
					"value", util.GenerateJsonObject(
						util.GenerateJsonProperty("instance", trunkMediaDisconnectOnIdleRtp),
					)))),
		util.GenerateJsonProperty(
			"trunk_media_codec", util.GenerateJsonObject(
				util.GenerateJsonProperty(
					"value", util.GenerateJsonObject(
						util.GenerateJsonArrayProperty("instance", strings.Join(trunkMediaCodec, ",")),
					)))),
	)
}

func GenerateTrunkBaseSettingsInboundSite(inboundSiteId string) string {
	return fmt.Sprintf(`inbound_site_id = genesyscloud_telephony_providers_edges_site.%s.id`, inboundSiteId)
}
