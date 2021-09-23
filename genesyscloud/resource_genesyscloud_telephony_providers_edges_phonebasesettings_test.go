package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v53/platformclientv2"
	"strconv"
	"strings"
	"testing"
)

func TestAccResourcePhoneBaseSettings(t *testing.T) {
	var (
		phoneBaseSettingsRes = "phoneBaseSettings1234"
		name1                = "test phone base settings " + uuid.NewString()
		name2                = "test phone base settings " + uuid.NewString()
		description1         = "test description 1"
		description2         = "test description 2"
		phoneMetaBaseId      = "generic_sip.json"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generatePhoneBaseSettingsResourceWithCustomAttrs(
					phoneBaseSettingsRes,
					name1,
					description1,
					phoneMetaBaseId,
					generatePhoneBaseSettingsProperties("Generic SIP Phone", 1, true, true, false, []string{strconv.Quote("station 1")}),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "description", description1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "phone_meta_base_id", phoneMetaBaseId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties.0.phone_label", "Generic SIP Phone"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties.0.phone_max_line_keys", "1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties.0.phone_mwi_enabled", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties.0.phone_mwi_subscribe", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties.0.phone_standalone", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties.0.phone_stations.0", "station 1"),
				),
			},
			// Update with new name, description and properties
			{
				Config: generatePhoneBaseSettingsResourceWithCustomAttrs(
					phoneBaseSettingsRes,
					name2,
					description2,
					phoneMetaBaseId,
					generatePhoneBaseSettingsProperties("Generic SIP Phone 1", 2, false, false, true, []string{strconv.Quote("station 2"), strconv.Quote("station 1")}),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "description", description2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "phone_meta_base_id", phoneMetaBaseId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties.0.phone_label", "Generic SIP Phone 1"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties.0.phone_max_line_keys", "2"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties.0.phone_mwi_enabled", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties.0.phone_mwi_subscribe", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties.0.phone_standalone", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties.0.phone_stations.0", "station 2"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "properties.0.phone_stations.1", "station 1"),
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
		} else if resp != nil && resp.StatusCode == 404 {
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

func generatePhoneBaseSettingsResourceWithCustomAttrs(
	phoneBaseSettingsRes,
	name,
	description,
	phoneMetaBaseId string,
	otherAttrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_telephony_providers_edges_phonebasesettings" "%s" {
		name = "%s"
		description = "%s"
		phone_meta_base_id = "%s"
		%s
	}
	`, phoneBaseSettingsRes, name, description, phoneMetaBaseId, strings.Join(otherAttrs, "\n"))
}

func generatePhoneBaseSettingsProperties(phoneLabel string, phoneMaxLineKeys int, phoneMwiEnabled, phoneMwiSubscribe, phoneStandalone bool, phoneStations []string) string {
	// A random selection of properties
	return fmt.Sprintf(`properties {
			phone_label = "%s"
			phone_max_line_keys = %d
			phone_mwi_enabled = %v
			phone_mwi_subscribe = %v
			phone_standalone = %v
			phone_stations = [%s]
		}`, phoneLabel, phoneMaxLineKeys, phoneMwiEnabled, phoneMwiSubscribe, phoneStandalone, strings.Join(phoneStations, ","))
}
