package telephony_providers_edges_phonebasesettings

import (
	"fmt"
	"strconv"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
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
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
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
						gcloud.TrueValue,
						gcloud.TrueValue,
						gcloud.FalseValue,
						[]string{strconv.Quote("station 1")}),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "description", description1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "phone_meta_base_id", phoneMetaBaseId),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_label", "Generic SIP Phone"),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_maxLineKeys", "1"),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_mwi_enabled", gcloud.TrueValue),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_mwi_subscribe", gcloud.TrueValue),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_standalone", gcloud.FalseValue),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_stations", strings.Join([]string{"station 1"}, ",")),
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
						gcloud.FalseValue,
						gcloud.FalseValue,
						gcloud.TrueValue,
						[]string{strconv.Quote("station 2"), strconv.Quote("station 1")}),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "description", description2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "phone_meta_base_id", phoneMetaBaseId),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_label", "Generic SIP Phone 1"),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_maxLineKeys", "2"),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_mwi_enabled", gcloud.FalseValue),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_mwi_subscribe", gcloud.FalseValue),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_standalone", gcloud.TrueValue),
					gcloud.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties", "phone_stations", strings.Join([]string{"station 2", "station 1"}, ",")),
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
		} else if gcloud.IsStatus404(resp) {
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
	return "properties = " + gcloud.GenerateJsonEncodedProperties(
		gcloud.GenerateJsonProperty(
			"phone_label", gcloud.GenerateJsonObject(
				gcloud.GenerateJsonProperty(
					"value", gcloud.GenerateJsonObject(
						gcloud.GenerateJsonProperty("instance", strconv.Quote(phoneLabel)),
					)))),
		gcloud.GenerateJsonProperty(
			"phone_maxLineKeys", gcloud.GenerateJsonObject(
				gcloud.GenerateJsonProperty(
					"value", gcloud.GenerateJsonObject(
						gcloud.GenerateJsonProperty("instance", phoneMaxLineKeys),
					)))),
		gcloud.GenerateJsonProperty(
			"phone_mwi_enabled", gcloud.GenerateJsonObject(
				gcloud.GenerateJsonProperty(
					"value", gcloud.GenerateJsonObject(
						gcloud.GenerateJsonProperty("instance", phoneMwiEnabled),
					)))),
		gcloud.GenerateJsonProperty(
			"phone_mwi_subscribe", gcloud.GenerateJsonObject(
				gcloud.GenerateJsonProperty(
					"value", gcloud.GenerateJsonObject(
						gcloud.GenerateJsonProperty("instance", phoneMwiSubscribe),
					)))),
		gcloud.GenerateJsonProperty(
			"phone_standalone", gcloud.GenerateJsonObject(
				gcloud.GenerateJsonProperty(
					"value", gcloud.GenerateJsonObject(
						gcloud.GenerateJsonProperty("instance", phoneStandalone),
					)))),
		gcloud.GenerateJsonProperty(
			"phone_stations", gcloud.GenerateJsonObject(
				gcloud.GenerateJsonProperty(
					"value", gcloud.GenerateJsonObject(
						gcloud.GenerateJsonArrayProperty("instance", strings.Join(phoneStations, ",")),
					)))),
	)
}
