package telephony_providers_edges_phone

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccResourcePhoneBasic(t *testing.T) {
	var (
		phoneRes    = "phone1234"
		name1       = "test-phone_" + uuid.NewString()
		name2       = "test-phone_" + uuid.NewString()
		stateActive = "active"

		phoneBaseSettingsRes  = "phoneBaseSettings1234"
		phoneBaseSettingsName = "phoneBaseSettings " + uuid.NewString()

		userRes1   = "user1"
		userName1  = "test_webrtc_user_" + uuid.NewString()
		userEmail1 = userName1 + "@test.com"

		userRes2   = "user2"
		userName2  = "test_webrtc_user_" + uuid.NewString()
		userEmail2 = userName2 + "@test.com"

		userTitle      = "Senior Director"
		userDepartment = "Development"
	)

	user1 := gcloud.GenerateUserResource(
		userRes1,
		userEmail1,
		userName1,
		nullValue, // Defaults to active
		strconv.Quote(userTitle),
		strconv.Quote(userDepartment),
		nullValue, // No manager
		nullValue, // Default acdAutoAnswer
		"",        // No profile skills
		"",        // No certs
	)

	user2 := gcloud.GenerateUserResource(
		userRes2,
		userEmail2,
		userName2,
		nullValue, // Defaults to active
		strconv.Quote(userTitle),
		strconv.Quote(userDepartment),
		nullValue, // No manager
		nullValue, // Default acdAutoAnswer
		"",        // No profile skills
		"",        // No certs
	)

	_, err := gcloud.AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	siteId, err := edgeSite.GetOrganizationDefaultSiteId()
	if err != nil {
		t.Fatal(err)
	}

	config1 := gcloud.GenerateOrganizationMe() + user1 + user2 +
		gcloud.GeneratePhoneBaseSettingsResourceWithCustomAttrs(
			phoneBaseSettingsRes,
			phoneBaseSettingsName,
			"phoneBaseSettings description",
			"inin_webrtc_softphone.json",
		) + GeneratePhoneResourceWithCustomAttrs(&PhoneConfig{
		phoneRes,
		name1,
		stateActive,
		fmt.Sprintf("\"%s\"", siteId),
		"genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsRes + ".id",
		nil, // no line addresses
		"genesyscloud_user." + userRes1 + ".id",
		"", // no depends on
	},
		generatePhoneCapabilities(
			false,
			false,
			false,
			false,
			false,
			false,
			true,
			"mac",
			[]string{strconv.Quote("audio/opus")},
		),
	)

	// Update phone with new user and name
	config2 := gcloud.GenerateOrganizationMe() + user1 + user2 +
		gcloud.GeneratePhoneBaseSettingsResourceWithCustomAttrs(
			phoneBaseSettingsRes,
			phoneBaseSettingsName,
			"phoneBaseSettings description",
			"inin_webrtc_softphone.json",
		) + GeneratePhoneResourceWithCustomAttrs(&PhoneConfig{
		phoneRes,
		name2,
		stateActive,
		fmt.Sprintf("\"%s\"", siteId),
		"genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsRes + ".id",
		nil, // no line addresses
		"genesyscloud_user." + userRes2 + ".id",
		"", // no depends_on
	},
		generatePhoneCapabilities(
			false,
			false,
			false,
			false,
			false,
			false,
			true,
			"mac",
			[]string{strconv.Quote("audio/opus")},
		),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: config1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "state", stateActive),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "site_id", siteId),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneRes, "phone_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneRes, "line_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "line_base_settings_id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneRes, "web_rtc_user_id", "genesyscloud_user."+userRes1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.provisions", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.registers", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.dual_registers", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.allow_reboot", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.no_rebalance", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.no_cloud_provisioning", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.cdm", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.hardware_id_type", "mac"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.media_codecs.0", "audio/opus"),
				),
			},
			{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "state", stateActive),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "site_id", siteId),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneRes, "phone_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneRes, "line_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "line_base_settings_id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneRes, "web_rtc_user_id", "genesyscloud_user."+userRes2, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_phone." + phoneRes,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: TestVerifyWebRtcPhoneDestroyed,
	})
}

func TestAccResourcePhoneStandalone(t *testing.T) {
	t.Parallel()
	number := "+14175538114"
	platformConfig, err := gcloud.AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	// TODO: Use did pool resource inside config once cyclic dependency issue is resolved between genesyscloud and did_pools package
	didPoolId, err := createDidPoolForEdgesPhoneTest(platformConfig, number)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := deleteDidPool(platformConfig, didPoolId); err != nil {
			t.Logf("failed to delete did pool '%s': %v", didPoolId, err)
		}
	}()

	lineAddresses := []string{number}
	phoneRes := "phone_standalone1234"
	name1 := "test-phone-standalone_" + uuid.NewString()
	stateActive := "active"
	phoneBaseSettingsRes := "phoneBaseSettings1234"
	phoneBaseSettingsName := "phoneBaseSettings " + uuid.NewString()

	locationRes := "test-location"

	emergencyNumber := "+13173114121"
	if err := edgeSite.DeleteLocationWithNumber(emergencyNumber); err != nil {
		t.Log(err)
	}

	locationConfig := gcloud.GenerateLocationResource(
		locationRes,
		"Terraform location"+uuid.NewString(),
		"HQ1",
		[]string{},
		gcloud.GenerateLocationEmergencyNum(
			emergencyNumber,
			nullValue, // Default number type
		), gcloud.GenerateLocationAddress(
			"0176 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46279",
		))

	siteRes := "test-site"
	siteConfig := edgeSite.GenerateSiteResourceWithCustomAttrs(
		siteRes,
		"tf site "+uuid.NewString(),
		"test site description",
		"genesyscloud_location."+locationRes+".id",
		"Premises",
		false,
		`["us-east-1"]`,
		nullValue,
		nullValue,
		"primary_sites   = []",
		"secondary_sites = []",
	)

	capabilities := generatePhoneCapabilities(
		false,
		true,
		true,
		true,
		true,
		false,
		false,
		"mac",
		[]string{},
	)

	config := gcloud.GeneratePhoneBaseSettingsResourceWithCustomAttrs(
		phoneBaseSettingsRes,
		phoneBaseSettingsName,
		"phoneBaseSettings description",
		"generic_sip.json",
	) + GeneratePhoneResourceWithCustomAttrs(&PhoneConfig{
		phoneRes,
		name1,
		stateActive,
		"genesyscloud_telephony_providers_edges_site." + siteRes + ".id",
		"genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsRes + ".id",
		lineAddresses,
		"", // no web rtc user
		"",
	}, capabilities)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: locationConfig + siteConfig + config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "state", stateActive),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneRes, "site_id", "genesyscloud_telephony_providers_edges_site."+siteRes, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneRes, "line_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "line_base_settings_id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneRes, "phone_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "line_addresses.0", lineAddresses[0]),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.provisions", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.registers", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.dual_registers", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.allow_reboot", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.no_rebalance", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.no_cloud_provisioning", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.cdm", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.hardware_id_type", "mac"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_phone." + phoneRes,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: TestVerifyWebRtcPhoneDestroyed,
	})
}

func generatePhoneCapabilities(
	provisions,
	registers,
	dualRegisters,
	allowReboot,
	noRebalance,
	noCloudProvisioning,
	cdm bool,
	hardwareIdType string,
	mediaCodecs []string) string {
	return fmt.Sprintf(`
		capabilities {
			provisions = %v
			registers = %v
			dual_registers = %v
			allow_reboot = %v
			no_rebalance = %v
			no_cloud_provisioning = %v
			cdm = %v
			hardware_id_type = "%s"
			media_codecs = [%s]
		}
	`, provisions, registers, dualRegisters, allowReboot, noRebalance, noCloudProvisioning, cdm, hardwareIdType, strings.Join(mediaCodecs, ","))
}

// TODO: Generate DID Pool resource inside test config when edges_phone has been moved to its own package
// and the cyclic dependency issue is resolved
func createDidPoolForEdgesPhoneTest(config *platformclientv2.Configuration, number string) (string, error) {
	api := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(config)
	body := &platformclientv2.Didpool{
		StartPhoneNumber: &number,
		EndPhoneNumber:   &number,
	}
	didPool, _, err := api.PostTelephonyProvidersEdgesDidpools(*body)
	if err != nil {
		return "", fmt.Errorf("failed to create did pool: %v", err)
	}
	return *didPool.Id, nil
}

func deleteDidPool(config *platformclientv2.Configuration, id string) error {
	api := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(config)
	if _, err := api.DeleteTelephonyProvidersEdgesDidpool(id); err != nil {
		return fmt.Errorf("error deleting did pool: %v", err)
	}
	return nil
}
