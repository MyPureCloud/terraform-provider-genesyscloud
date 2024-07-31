package telephony_providers_edges_phonebasesettings

import (
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourcePhoneBaseSettings(t *testing.T) {
	t.Parallel()
	var (
		phoneBaseSettingsRes = "phoneBaseSettings1234"
		name1                = "test phone base settings resource" + uuid.NewString()
		name2                = "test phone base settings resource" + uuid.NewString()
		description1         = "test description 1"
		description2         = "test description 2"
		phoneMetaBaseId      = "generic_sip.json"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
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
						util.TrueValue,
						util.TrueValue,
						util.FalseValue,
						[]string{strconv.Quote("station 1")}),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "description", description1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "phone_meta_base_id", phoneMetaBaseId),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_label", "Generic SIP Phone"),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_maxLineKeys", "1"),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_mwi_enabled", util.TrueValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_mwi_subscribe", util.TrueValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_standalone", util.FalseValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_stations", strings.Join([]string{"station 1"}, ",")),
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
						util.FalseValue,
						util.FalseValue,
						util.TrueValue,
						[]string{strconv.Quote("station 2"), strconv.Quote("station 1")}),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "description", description2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "phone_meta_base_id", phoneMetaBaseId),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_label", "Generic SIP Phone 1"),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_maxLineKeys", "2"),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_mwi_enabled", util.FalseValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_mwi_subscribe", util.FalseValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_standalone", util.TrueValue),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_stations", strings.Join([]string{"station 2", "station 1"}, ",")),
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
