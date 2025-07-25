package telephony_providers_edges_phone

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	location "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/location"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	didPool "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_did_pool"
	phoneBaseSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phonebasesettings"
	edgeSite "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func TestAccResourcePhoneBasic(t *testing.T) {
	var (
		phoneResourceLabel  = "phone1234"
		phoneResourceLabel2 = "phone5555"
		name1               = "test-phone_" + uuid.NewString()
		name2               = "test-phone_" + uuid.NewString()
		stateActive         = "active"

		phoneBaseSettingsResourceLabel = "phoneBaseSettings1234"
		phoneBaseSettingsName          = "phoneBaseSettings " + uuid.NewString()

		userResourceLabel1 = "user1"
		userName1          = "test_webrtc_user_" + uuid.NewString()
		userEmail1         = userName1 + "@test.com"

		userResourceLabel2 = "user2"
		userName2          = "test_webrtc_user_" + uuid.NewString()
		userEmail2         = userName2 + "@test.com"

		phoneBaseSettingsResourceLabel2 = "phoneBaseSettings123"
		phoneBaseSettingsName2          = "phoneBaseSettings " + uuid.NewString()

		userTitle      = "Senior Director"
		userDepartment = "Development"
	)

	user1 := user.GenerateUserResource(
		userResourceLabel1,
		userEmail1,
		userName1,
		util.NullValue, // Defaults to active
		strconv.Quote(userTitle),
		strconv.Quote(userDepartment),
		util.NullValue, // No manager
		util.NullValue, // Default acdAutoAnswer
		"",             // No profile skills
		"",             // No certs
	)

	user2 := user.GenerateUserResource(
		userResourceLabel2,
		userEmail2,
		userName2,
		util.NullValue, // Defaults to active
		strconv.Quote(userTitle),
		strconv.Quote(userDepartment),
		util.NullValue, // No manager
		util.NullValue, // Default acdAutoAnswer
		"",             // No profile skills
		"",             // No certs
	)

	siteId, err := edgeSite.GetOrganizationDefaultSiteId(sdkConfig)
	if err != nil {
		t.Fatal(err)
	}

	config1 := gcloud.GenerateOrganizationMe() + user1 + user2 +
		phoneBaseSettings.GeneratePhoneBaseSettingsResourceWithCustomAttrs(
			phoneBaseSettingsResourceLabel,
			phoneBaseSettingsName,
			"phoneBaseSettings description",
			"inin_webrtc_softphone.json",
		) + GeneratePhoneResourceWithCustomAttrs(&PhoneConfig{
		phoneResourceLabel,
		name1,
		stateActive,
		fmt.Sprintf("\"%s\"", siteId),
		"genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsResourceLabel + ".id",
		"genesyscloud_user." + userResourceLabel1 + ".id",
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
		generatePhoneProperties(uuid.NewString()),
	)

	// Update phone with new user and name
	config2 := gcloud.GenerateOrganizationMe() + user1 + user2 +
		phoneBaseSettings.GeneratePhoneBaseSettingsResourceWithCustomAttrs(
			phoneBaseSettingsResourceLabel2,
			phoneBaseSettingsName2,
			"phoneBaseSettings description",
			"inin_webrtc_softphone.json",
		) + GeneratePhoneResourceWithCustomAttrs(&PhoneConfig{
		phoneResourceLabel2,
		name2,
		stateActive,
		fmt.Sprintf("\"%s\"", siteId),
		"genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsResourceLabel2 + ".id",
		"genesyscloud_user." + userResourceLabel2 + ".id",
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
		generatePhoneProperties(uuid.NewString()),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
				Config: config1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "state", stateActive),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "site_id", siteId),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "phone_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "line_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "line_base_settings_id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "web_rtc_user_id", "genesyscloud_user."+userResourceLabel1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.provisions", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.registers", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.dual_registers", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.allow_reboot", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.no_rebalance", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.no_cloud_provisioning", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.cdm", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.hardware_id_type", "mac"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.media_codecs.0", "audio/opus"),
					checkifDefaultPhoneAdded("genesyscloud_user."+userResourceLabel1),
				),
			},
			{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel2, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel2, "state", stateActive),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel2, "site_id", siteId),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel2, "phone_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel2, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel2, "line_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel2, "line_base_settings_id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel2, "web_rtc_user_id", "genesyscloud_user."+userResourceLabel2, "id"),
					checkifDefaultPhoneAdded("genesyscloud_user."+userResourceLabel2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_phone." + phoneResourceLabel2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: TestVerifyWebRtcPhoneDestroyed,
	})
}

func TestAccResourceHardPhoneStandalone(t *testing.T) {
	number := "+13172128941"
	phoneMac := "AB12CD34"
	phoneMacUpdated := "BANANAS"
	// TODO: Use did pool resource inside config once cyclic dependency issue is resolved between genesyscloud and did_pools package
	didPoolId, err := createDidPoolForEdgesPhoneTest(sdkConfig, number)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := deleteDidPool(sdkConfig, didPoolId); err != nil {
			t.Logf("failed to delete did pool '%s': %v", didPoolId, err)
		}
	}()

	phoneResourceLabel := "phone_standalone987"
	name := "test-phone-standalone_" + uuid.NewString()

	stateActive := "active"
	phoneBaseSettingsResourceLabel := "phoneBaseSettings987"
	phoneBaseSettingsName := "phoneBaseSettings " + uuid.NewString()

	locationResourceLabel := "test-location-test111"

	emergencyNumber := "+13293100121"
	if err = edgeSite.DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s: %v", emergencyNumber, err)
	}

	locationConfig := location.GenerateLocationResource(
		locationResourceLabel,
		"Terraform-location"+uuid.NewString(),
		"HQ1",
		[]string{},
		location.GenerateLocationEmergencyNum(
			emergencyNumber,
			util.NullValue, // Default number type
		), location.GenerateLocationAddress(
			"0176 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"41119",
		))

	siteResourceLabel := "test-site"
	siteConfig := edgeSite.GenerateSiteResourceWithCustomAttrs(
		siteResourceLabel,
		"tf site "+uuid.NewString(),
		"test site description",
		"genesyscloud_location."+locationResourceLabel+".id",
		"Premises",
		false,
		`["us-east-1"]`,
		util.NullValue,
		util.NullValue,
		"primary_sites   = []",
		"secondary_sites = []",
	)

	capabilities := generatePhoneCapabilities(
		true,
		true,
		true,
		true,
		false,
		true,
		false,
		"mac",
		[]string{strconv.Quote("audio/opus"), strconv.Quote("audio/pcmu"), strconv.Quote("audio/pcma")},
	)
	config := locationConfig + siteConfig + phoneBaseSettings.GeneratePhoneBaseSettingsResourceWithCustomAttrs(
		phoneBaseSettingsResourceLabel,
		phoneBaseSettingsName,
		"phoneBaseSettings description",
		"audiocodes_400hd.json",
	)
	phone1 := GeneratePhoneResourceWithCustomAttrs(&PhoneConfig{
		phoneResourceLabel,
		name,
		stateActive,
		"genesyscloud_telephony_providers_edges_site." + siteResourceLabel + ".id",
		"genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsResourceLabel + ".id",
		"", // no web rtc user
		"", // no Depends On
	}, capabilities, generatePhoneProperties(phoneMac))

	//only mac is updated here, same resource as phone 1
	phone2 := GeneratePhoneResourceWithCustomAttrs(&PhoneConfig{
		phoneResourceLabel,
		name,
		stateActive,
		"genesyscloud_telephony_providers_edges_site." + siteResourceLabel + ".id",
		"genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsResourceLabel + ".id",
		"", // no web rtc user
		"", // no Depends On
	}, capabilities, generatePhoneProperties(phoneMacUpdated))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{

				Config: config + phone1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "state", stateActive),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "site_id", "genesyscloud_telephony_providers_edges_site."+siteResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "line_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "line_base_settings_id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "phone_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.provisions", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.registers", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.dual_registers", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.allow_reboot", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.no_rebalance", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.no_cloud_provisioning", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.cdm", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.hardware_id_type", "mac"),
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "properties", "phone_hardwareId", phoneMac),
				),
			},
			{
				Config: config + phone2,
				Check: resource.ComposeTestCheckFunc(
					util.ValidateValueInJsonPropertiesAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "properties", "phone_hardwareId", phoneMacUpdated),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_phone." + phoneResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: TestVerifyWebRtcPhoneDestroyed,
	})
}

func TestAccResourcePhoneStandalone(t *testing.T) {
	lineAddresses := "+12005537112"
	deleteDidPoolWithNumber(lineAddresses)
	didPoolResourceLabel1 := "test-didpool1"
	phoneResourceLabel := "phone_standalone1234"
	name1 := "test-phone-standalone_" + uuid.NewString()
	stateActive := "active"
	phoneBaseSettingsResourceLabel := "phoneBaseSettings1234"
	phoneBaseSettingsName := "phoneBaseSettings " + uuid.NewString()
	resourcePath := "genesyscloud_telephony_providers_edges_did_pool" + "." + didPoolResourceLabel1
	locationResourceLabel := "test-location"

	emergencyNumber := "+13173114121"
	if err := edgeSite.DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s: %v", emergencyNumber, err)
	}

	locationConfig := location.GenerateLocationResource(
		locationResourceLabel,
		"Terraform location"+uuid.NewString(),
		"HQ1",
		[]string{},
		location.GenerateLocationEmergencyNum(
			emergencyNumber,
			util.NullValue, // Default number type
		), location.GenerateLocationAddress(
			"0176 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46279",
		))

	siteResourceLabel := "test-site"
	siteConfig := edgeSite.GenerateSiteResourceWithCustomAttrs(
		siteResourceLabel,
		"tf site "+uuid.NewString(),
		"test site description",
		"genesyscloud_location."+locationResourceLabel+".id",
		"Premises",
		false,
		util.AssignRegion(),
		util.NullValue,
		util.NullValue,
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
		true,
		"mac",
		[]string{},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: didPool.GenerateDidPoolResource(&didPool.DidPoolStruct{
					ResourceLabel:    didPoolResourceLabel1,
					StartPhoneNumber: lineAddresses,
					EndPhoneNumber:   lineAddresses,
					Description:      util.NullValue, // No description
					Comments:         util.NullValue, // No comments
					PoolProvider:     util.NullValue, // No provider
				}),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper updation
						return nil
					},
					resource.TestCheckResourceAttr(resourcePath, "start_phone_number", lineAddresses)),
			},
			{
				Config: didPool.GenerateDidPoolResource(&didPool.DidPoolStruct{
					ResourceLabel:    didPoolResourceLabel1,
					StartPhoneNumber: lineAddresses,
					EndPhoneNumber:   lineAddresses,
					Description:      util.NullValue, // No description
					Comments:         util.NullValue, // No comments
					PoolProvider:     util.NullValue, // No provider
				}) + locationConfig + siteConfig + phoneBaseSettings.GeneratePhoneBaseSettingsResourceWithCustomAttrs(
					phoneBaseSettingsResourceLabel,
					phoneBaseSettingsName,
					"phoneBaseSettings description",
					"generic_sip.json",
				) + GeneratePhoneResourceWithCustomAttrs(&PhoneConfig{
					phoneResourceLabel,
					name1,
					stateActive,
					"genesyscloud_telephony_providers_edges_site." + siteResourceLabel + ".id",
					"genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsResourceLabel + ".id",
					"", // no web rtc user
					"genesyscloud_telephony_providers_edges_did_pool." + didPoolResourceLabel1,
				}, capabilities, generateLinePropertiesLineAddress(strconv.Quote(lineAddresses)), generatePhoneProperties(uuid.NewString())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "state", stateActive),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "site_id", "genesyscloud_telephony_providers_edges_site."+siteResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "line_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "line_base_settings_id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "phone_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "line_properties.0.line_address.0", lineAddresses),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.provisions", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.registers", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.dual_registers", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.allow_reboot", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.no_rebalance", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.no_cloud_provisioning", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.cdm", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.hardware_id_type", "mac"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_phone." + phoneResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: TestVerifyWebRtcPhoneDestroyed,
	})
}

func TestAccResourcePhoneStandaloneRemoteStation(t *testing.T) {
	remoteStationAddress := "+11005538454"
	phoneResourceLabel := "phone_standalone1234"
	name1 := "test-phone-Kstandalone_" + uuid.NewString()
	stateActive := "active"
	phoneBaseSettingsResourceLabel := "phoneBaseSettings1234"
	phoneBaseSettingsName := "phoneBaseSettings " + uuid.NewString()

	locationResourceLabel := "test-location"

	emergencyNumber := "+13173117632"
	if err := edgeSite.DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s: %v", emergencyNumber, err)
	}

	locationConfig := location.GenerateLocationResource(
		locationResourceLabel,
		"TerraformLocationRemote"+uuid.NewString(),
		"HQ1",
		[]string{},
		location.GenerateLocationEmergencyNum(
			emergencyNumber,
			util.NullValue, // Default number type
		), location.GenerateLocationAddress(
			"0176 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46279",
		))

	siteResourceLabel := "test-site"
	siteConfig := edgeSite.GenerateSiteResourceWithCustomAttrs(
		siteResourceLabel,
		"tf site "+uuid.NewString(),
		"test site description",
		"genesyscloud_location."+locationResourceLabel+".id",
		"Premises",
		false,
		`["us-east-1"]`,
		util.NullValue,
		util.NullValue,
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
		true,
		"mac",
		[]string{},
	)

	config := phoneBaseSettings.GeneratePhoneBaseSettingsResourceWithCustomAttrs(
		phoneBaseSettingsResourceLabel,
		phoneBaseSettingsName,
		"phoneBaseSettings description",
		"generic_sip.json",
	) + GeneratePhoneResourceWithCustomAttrs(&PhoneConfig{
		phoneResourceLabel,
		name1,
		stateActive,
		"genesyscloud_telephony_providers_edges_site." + siteResourceLabel + ".id",
		"genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsResourceLabel + ".id",
		"", // no web rtc user
		"", // no depends on
	}, capabilities, generateLinePropertiesRemoteAddress(strconv.Quote(remoteStationAddress)), generatePhoneProperties(uuid.NewString()))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
				Config: locationConfig + siteConfig + config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "state", stateActive),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "site_id", "genesyscloud_telephony_providers_edges_site."+siteResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "line_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "line_base_settings_id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "phone_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsResourceLabel, "id"),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "line_properties.0.remote_address.0", remoteStationAddress),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.provisions", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.registers", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.dual_registers", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.allow_reboot", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.no_rebalance", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.no_cloud_provisioning", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.cdm", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "capabilities.0.hardware_id_type", "mac"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_phone." + phoneResourceLabel,
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

// Check if flow is published, then check if flow name and type are correct
func checkifDefaultPhoneAdded(userName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		userResource, ok := state.RootModule().Resources[userName]
		if !ok {
			return fmt.Errorf("Failed to find user %s in state", userName)
		}
		userId := userResource.Primary.ID
		usersApi := platformclientv2.NewUsersApi()
		stationsApi := platformclientv2.NewStationsApi()
		const pageSize = 100
		const pageNum = 1
		stations, _, err := stationsApi.GetStations(pageSize, pageNum, "", "", "", userId, "", "")
		if err != nil {
			return fmt.Errorf("Unexpected error: %s", err)
		}
		if stations.Entities == nil || len(*stations.Entities) == 0 {
			return fmt.Errorf("Failed to find user %s in state", userName)
		}

		user, _, err := usersApi.GetUserStation(userId)

		if err != nil {
			return fmt.Errorf("Unexpected error: %s", err)
		}

		if user == nil || user.DefaultStation == nil {
			return fmt.Errorf("User Stations (%s) not found. ", userId)
		}

		station := &(*stations.Entities)[0]
		if *user.DefaultStation.Id != *station.Id {
			return fmt.Errorf("User  (%s) has incorrect default station Id. Expect: %s, Actual: %s", userId, *station.Id, *user.DefaultStation.Id)
		}

		return nil
	}
}

func deleteDidPool(config *platformclientv2.Configuration, id string) error {
	api := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(config)
	if _, err := api.DeleteTelephonyProvidersEdgesDidpool(id); err != nil {
		return fmt.Errorf("error deleting did pool: %v", err)
	}
	return nil
}

func deleteDidPoolWithNumber(number string) {
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)
	var didPoolsToDelete []string

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		didPools, _, getErr := edgesAPI.GetTelephonyProvidersEdgesDidpools(pageSize, pageNum, "", nil)
		if getErr != nil {
			return
		}

		if didPools.Entities == nil || len(*didPools.Entities) == 0 {
			break
		}

		for _, didPool := range *didPools.Entities {
			if (didPool.StartPhoneNumber != nil && *didPool.StartPhoneNumber == number) ||
				(didPool.EndPhoneNumber != nil && *didPool.EndPhoneNumber == number) {
				didPoolsToDelete = append(didPoolsToDelete, *didPool.Id)
			}
		}
	}

	for _, didPoolId := range didPoolsToDelete {
		edgesAPI.DeleteTelephonyProvidersEdgesDidpool(didPoolId)
		time.Sleep(5 * time.Second)
	}
}
