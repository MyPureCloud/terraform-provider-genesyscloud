package telephony_providers_edges_phonebasesettings

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

func TestAccResourcePhoneBaseSettings(t *testing.T) {
	t.Parallel()
	var (
		phoneBaseSettingsResourceLabel = "phoneBaseSettings1234"
		name1                          = "test phone base settings resource" + uuid.NewString()
		name2                          = "test phone base settings resource" + uuid.NewString()
		description1                   = "test description 1"
		description2                   = "test description 2"
		phoneMetaBaseId                = "generic_sip.json"
		stationPersistTimeout          = "2000"
		stationPersistTimeoutUpdate    = "3000"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GeneratePhoneBaseSettingsResourceWithCustomAttrs(
					phoneBaseSettingsResourceLabel,
					name1,
					description1,
					phoneMetaBaseId,
					generatePhoneBaseSettingsProperties(
						"Generic SIP Phone",
						"1",
						util.TrueValue,
						util.TrueValue,
						util.FalseValue,
						[]string{strconv.Quote("station 1")}),
					generatePhoneBaseSettingsLineBase(
						util.FalseValue, // station_persistent_enabled
						util.FalseValue, // station_persistent_webrtc_enabled
						stationPersistTimeout,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "description", description1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "phone_meta_base_id", phoneMetaBaseId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "line_base.0.station_persistent_timeout", stationPersistTimeout),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "line_base.0.station_persistent_enabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "line_base.0.station_persistent_webrtc_enabled", util.FalseValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "properties", "phone_label", "Generic SIP Phone"),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "properties", "phone_maxLineKeys", "1"),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "properties", "phone_mwi_enabled", util.TrueValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "properties", "phone_mwi_subscribe", util.TrueValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "properties", "phone_standalone", util.FalseValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "properties", "phone_stations", strings.Join([]string{"station 1"}, ",")),
				),
			},
			// Update with new name, description and properties
			{
				Config: GeneratePhoneBaseSettingsResourceWithCustomAttrs(
					phoneBaseSettingsResourceLabel,
					name2,
					description2,
					phoneMetaBaseId,
					generatePhoneBaseSettingsProperties(
						"Generic SIP Phone 1",
						"2",
						util.FalseValue,
						util.FalseValue,
						util.TrueValue,
						[]string{strconv.Quote("station 2"), strconv.Quote("station 1")}),
					generatePhoneBaseSettingsLineBase(
						util.TrueValue, // station_persistent_enabled
						util.TrueValue, // station_persistent_webrtc_enabled
						stationPersistTimeoutUpdate,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "description", description2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "phone_meta_base_id", phoneMetaBaseId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "line_base.0.station_persistent_timeout", stationPersistTimeoutUpdate),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "line_base.0.station_persistent_enabled", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "line_base.0.station_persistent_webrtc_enabled", util.TrueValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "properties", "phone_label", "Generic SIP Phone 1"),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "properties", "phone_maxLineKeys", "2"),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "properties", "phone_mwi_enabled", util.FalseValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "properties", "phone_mwi_subscribe", util.FalseValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "properties", "phone_standalone", util.TrueValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "properties", "phone_stations", strings.Join([]string{"station 2", "station 1"}, ",")),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyPhoneBaseSettingsDestroyed,
	})
}

func testVerifyPhoneBaseSettingsDestroyed(state *terraform.State) error {
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_telephony_providers_edges_phonebasesettings" {
			continue
		}

		phoneBaseSettings, resp, err := edgesAPI.GetTelephonyProvidersEdgesPhonebasesetting(rs.Primary.ID)
		if phoneBaseSettings != nil {
			return fmt.Errorf("PhoneBaseSettings (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// PhoneBaseSettings not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	//Success. PhoneBaseSettings destroyed
	return nil
}

func generatePhoneBaseSettingsLineBase(enabled, webRtcEnabled, timeout string) string {
	return fmt.Sprintf(`
	line_base {
		station_persistent_enabled        = %s
		station_persistent_webrtc_enabled = %s
		station_persistent_timeout        = %s
	}
`, enabled, webRtcEnabled, timeout)
}

func generatePhoneBaseSettingsProperties(phoneLabel, phoneMaxLineKeys, phoneMwiEnabled, phoneMwiSubscribe, phoneStandalone string, phoneStations []string) string {
	// A random selection of properties
	return "properties = " + util.GenerateJsonEncodedProperties(
		util.GenerateJsonProperty(
			"phone_label", util.GenerateJsonObject(
				util.GenerateJsonProperty(
					"value", util.GenerateJsonObject(
						util.GenerateJsonProperty("instance", strconv.Quote(phoneLabel)),
					)))),
		util.GenerateJsonProperty(
			"phone_maxLineKeys", util.GenerateJsonObject(
				util.GenerateJsonProperty(
					"value", util.GenerateJsonObject(
						util.GenerateJsonProperty("instance", phoneMaxLineKeys),
					)))),
		util.GenerateJsonProperty(
			"phone_mwi_enabled", util.GenerateJsonObject(
				util.GenerateJsonProperty(
					"value", util.GenerateJsonObject(
						util.GenerateJsonProperty("instance", phoneMwiEnabled),
					)))),
		util.GenerateJsonProperty(
			"phone_mwi_subscribe", util.GenerateJsonObject(
				util.GenerateJsonProperty(
					"value", util.GenerateJsonObject(
						util.GenerateJsonProperty("instance", phoneMwiSubscribe),
					)))),
		util.GenerateJsonProperty(
			"phone_standalone", util.GenerateJsonObject(
				util.GenerateJsonProperty(
					"value", util.GenerateJsonObject(
						util.GenerateJsonProperty("instance", phoneStandalone),
					)))),
		util.GenerateJsonProperty(
			"phone_stations", util.GenerateJsonObject(
				util.GenerateJsonProperty(
					"value", util.GenerateJsonObject(
						util.GenerateJsonArrayProperty("instance", strings.Join(phoneStations, ",")),
					)))),
	)
}
