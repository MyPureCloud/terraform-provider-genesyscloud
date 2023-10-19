package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccResourcePhoneBaseSettings(t *testing.T) {
	t.Parallel()
	var (
		phoneBaseSettingsRes = "phoneBaseSettings1234"
		name1                = "test phone base settings " + uuid.NewString()
		name2                = "test phone base settings " + uuid.NewString()
		description1         = "test description 1"
		description2         = "test description 2"
		phoneMetaBaseId      = "generic_sip.json"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GeneratePhoneBaseSettingsResourceWithCustomAttrs(
					phoneBaseSettingsRes,
					name1,
					description1,
					phoneMetaBaseId,
					generatePhoneBaseSettingsProperties(
						"Generic SIP Phone",
						"1",
						trueValue,
						trueValue,
						falseValue,
						[]string{strconv.Quote("station 1")}),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "description", description1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "phone_meta_base_id", phoneMetaBaseId),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_label", "Generic SIP Phone"),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_maxLineKeys", "1"),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_mwi_enabled", trueValue),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_mwi_subscribe", trueValue),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_standalone", falseValue),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_stations", strings.Join([]string{"station 1"}, ",")),
				),
			},
			// Update with new name, description and properties
			{
				Config: GeneratePhoneBaseSettingsResourceWithCustomAttrs(
					phoneBaseSettingsRes,
					name2,
					description2,
					phoneMetaBaseId,
					generatePhoneBaseSettingsProperties(
						"Generic SIP Phone 1",
						"2",
						falseValue,
						falseValue,
						trueValue,
						[]string{strconv.Quote("station 2"), strconv.Quote("station 1")}),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "description", description2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "phone_meta_base_id", phoneMetaBaseId),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_label", "Generic SIP Phone 1"),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_maxLineKeys", "2"),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_mwi_enabled", falseValue),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_mwi_subscribe", falseValue),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_standalone", trueValue),
					validateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_stations", strings.Join([]string{"station 2", "station 1"}, ",")),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsRes,
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
		} else if IsStatus404(resp) {
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

func generatePhoneBaseSettingsProperties(phoneLabel, phoneMaxLineKeys, phoneMwiEnabled, phoneMwiSubscribe, phoneStandalone string, phoneStations []string) string {
	// A random selection of properties
	return "properties = " + GenerateJsonEncodedProperties(
		GenerateJsonProperty(
			"phone_label", GenerateJsonObject(
				GenerateJsonProperty(
					"value", GenerateJsonObject(
						GenerateJsonProperty("instance", strconv.Quote(phoneLabel)),
					)))),
		GenerateJsonProperty(
			"phone_maxLineKeys", GenerateJsonObject(
				GenerateJsonProperty(
					"value", GenerateJsonObject(
						GenerateJsonProperty("instance", phoneMaxLineKeys),
					)))),
		GenerateJsonProperty(
			"phone_mwi_enabled", GenerateJsonObject(
				GenerateJsonProperty(
					"value", GenerateJsonObject(
						GenerateJsonProperty("instance", phoneMwiEnabled),
					)))),
		GenerateJsonProperty(
			"phone_mwi_subscribe", GenerateJsonObject(
				GenerateJsonProperty(
					"value", GenerateJsonObject(
						GenerateJsonProperty("instance", phoneMwiSubscribe),
					)))),
		GenerateJsonProperty(
			"phone_standalone", GenerateJsonObject(
				GenerateJsonProperty(
					"value", GenerateJsonObject(
						GenerateJsonProperty("instance", phoneStandalone),
					)))),
		GenerateJsonProperty(
			"phone_stations", GenerateJsonObject(
				GenerateJsonProperty(
					"value", GenerateJsonObject(
						GenerateJsonArrayProperty("instance", strings.Join(phoneStations, ",")),
					)))),
	)
}
