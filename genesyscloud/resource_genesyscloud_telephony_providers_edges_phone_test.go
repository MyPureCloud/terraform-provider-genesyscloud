package genesyscloud

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v55/platformclientv2"
)

var (
	webRtcUserId1         string
	webRtcUserId2         string
	genericSIPPhoneBaseId string
	sdkConfig             *platformclientv2.Configuration
)

type phoneConfig struct {
	phoneRes            string
	name                string
	state               string
	siteId              string
	phoneBaseSettingsId string
	lineBaseSettingsId  string
	lineAddresses       []string
	webRtcUserId        string
	depends_on          string
}

func TestAccResourcePhoneBasic(t *testing.T) {
	var (
		phoneRes    = "phone1234"
		name1       = "test-phone_" + uuid.NewString()
		name2       = "test-phone_" + uuid.NewString()
		stateActive = "active"
	)

	err := authorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	siteId, err := getDefaultSiteId()
	if err != nil {
		t.Fatal(err)
	}

	phoneBaseSettings, err := getWebRTCPhoneBaseSettings()
	if err != nil {
		t.Fatal(err)
	}
	phoneBaseSettingsId := *phoneBaseSettings.Id

	line := *phoneBaseSettings.Lines
	lineBaseSettingsId := *line[0].Id

	user, err := createWebRTCUser()
	if err != nil {
		t.Fatal(err)
	}

	// ID of the initial user
	webRtcUserId1 = *user.Id

	user, err = createWebRTCUser()
	if err != nil {
		t.Fatal(err)
	}

	// ID of the second user
	webRtcUserId2 = *user.Id

	capabilities := generatePhoneCapabilities(
		false,
		false,
		false,
		false,
		false,
		false,
		true,
		"mac",
		[]string{strconv.Quote("audio/opus")},
	)

	config := generatePhoneResourceWithCustomAttrs(&phoneConfig{
		phoneRes,
		name1,
		stateActive,
		siteId,
		phoneBaseSettingsId,
		lineBaseSettingsId,
		nil, // no line addresses
		webRtcUserId1,
		"", // no depends on
	}, capabilities)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "state", stateActive),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "site_id", siteId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "phone_base_settings_id", phoneBaseSettingsId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "web_rtc_user_id", webRtcUserId1),
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
				// Update phone with new user and name
				Config: generatePhoneResourceWithCustomAttrs(&phoneConfig{
					phoneRes,
					name2,
					stateActive,
					siteId,
					phoneBaseSettingsId,
					lineBaseSettingsId,
					nil, // no line addresses
					webRtcUserId2,
					"", // no depends_on
				}, capabilities),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "state", stateActive),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "site_id", siteId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "phone_base_settings_id", phoneBaseSettingsId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "web_rtc_user_id", webRtcUserId2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_phone." + phoneRes,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyWebRtcPhoneDestroyed,
	})
}

func TestAccResourcePhoneStandalone(t *testing.T) {
	didPoolResource1 := "test-didpool1"
	lineAddresses := []string{"+15175550010"}
	phoneRes := "phone_standalone1234"
	name1 := "test-phone-standalone_" + uuid.NewString()
	stateActive := "active"

	err := authorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	siteId, err := getDefaultSiteId()
	if err != nil {
		t.Fatal(err)
	}

	phoneBaseSettings, err := createGenericSIPPhoneBase()
	if err != nil {
		t.Fatal(err)
	}
	genericSIPPhoneBaseId = *phoneBaseSettings.Id

	line := *phoneBaseSettings.Lines
	lineBaseSettingsId := *line[0].Id

	capabilities := generatePhoneCapabilities(
		false,
		true,
		false,
		true,
		true,
		false,
		true,
		"mac",
		[]string{},
	)

	config := generateDidPoolResource(&didPoolStruct{
		didPoolResource1,
		lineAddresses[0],
		lineAddresses[0],
		nullValue, // No description
		nullValue, // No comments
		nullValue, // No provider
	})

	config += generatePhoneResourceWithCustomAttrs(&phoneConfig{
		phoneRes,
		name1,
		stateActive,
		siteId,
		genericSIPPhoneBaseId,
		lineBaseSettingsId,
		lineAddresses,
		"", // no web rtc user
		"genesyscloud_telephony_providers_edges_did_pool." + didPoolResource1,
	}, capabilities)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "state", stateActive),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "site_id", siteId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "phone_base_settings_id", genericSIPPhoneBaseId),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "line_addresses.0", lineAddresses[0]),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.provisions", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.registers", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.dual_registers", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.allow_reboot", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.no_rebalance", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.no_cloud_provisioning", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "capabilities.0.cdm", trueValue),
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
		CheckDestroy: testVerifyWebRtcPhoneDestroyed,
	})
}

func testVerifyWebRtcPhoneDestroyed(state *terraform.State) error {
	deleteWebRTCUser(webRtcUserId1)
	deleteWebRTCUser(webRtcUserId2)
	deleteBaseSetting(genericSIPPhoneBaseId)

	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_telephony_providers_edges_phone" {
			continue
		}

		phone, resp, err := edgesAPI.GetTelephonyProvidersEdgesPhone(rs.Primary.ID)
		if phone != nil {
			return fmt.Errorf("Phone (%s) still exists", rs.Primary.ID)
		} else if resp != nil && resp.StatusCode == 404 {
			// Phone not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	//Success. Phone destroyed
	return nil
}

func getWebRTCPhoneBaseSettings() (*platformclientv2.Phonebase, error) {
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		settings, _, err := edgesAPI.GetTelephonyProvidersEdgesPhonebasesettings(100, 1, "", "", nil, "")
		if err != nil {
			return nil, err
		}

		if settings.Entities == nil || len(*settings.Entities) == 0 {
			break
		}

		for _, setting := range *settings.Entities {
			// Creating a WebRTC phone for the tests
			if *setting.PhoneMetaBase.Id == "inin_webrtc_softphone.json" {
				return &setting, nil
			}
		}
	}

	return nil, errors.New("could not find webrtc phone settings")
}

func createGenericSIPPhoneBase() (*platformclientv2.Phonebase, error) {
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)
	name := "TestGenericSip-cxascode-12345"+uuid.NewString()
	phoneMetaBaseId := "generic_sip.json"

	baseSettingBody := platformclientv2.Phonebase{
		Name: &name,
		PhoneMetaBase: &platformclientv2.Domainentityref{
			Id: &phoneMetaBaseId,
		},
		Lines: &[]platformclientv2.Linebase{
			{
				Name: &name,
				LineMetaBase: &platformclientv2.Domainentityref{
					Id: &phoneMetaBaseId,
				},
			},
		},
	}

	baseSetting, _, err := edgesAPI.PostTelephonyProvidersEdgesPhonebasesettings(baseSettingBody)

	return baseSetting, err
}

func deleteBaseSetting(id string) error {
	if len(id) == 0 {
		return nil
	}

	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Deleting base setting %s", id)
	_, err := edgesAPI.DeleteTelephonyProvidersEdgesPhonebasesetting(id)

	return err
}

func getDefaultSiteId() (string, error) {
	orgsAPI := platformclientv2.NewOrganizationApiWithConfig(sdkConfig)

	org, _, err := orgsAPI.GetOrganizationsMe()
	if err != nil {
		return "", err
	}

	return *org.DefaultSiteId, nil
}

func createWebRTCUser() (*platformclientv2.User, error) {
	email := "webRtcUser_" + uuid.NewString() + "@email.com"
	name := "webRtcUserTest"

	createUser := platformclientv2.Createuser{
		Email: &email,
		Name:  &name,
	}

	// Create API instance using config
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)

	user, _, err := usersAPI.PostUsers(createUser)

	return user, err
}

func authorizeSdk() error {
	// Create new config
	sdkConfig = platformclientv2.NewConfiguration()

	sdkConfig.BasePath = getRegionBasePath(os.Getenv("GENESYSCLOUD_REGION"))

	err := sdkConfig.AuthorizeClientCredentials(os.Getenv("GENESYSCLOUD_OAUTHCLIENT_ID"), os.Getenv("GENESYSCLOUD_OAUTHCLIENT_SECRET"))
	if err != nil {
		return err
	}

	return nil
}

func deleteWebRTCUser(id string) error {
	if len(id) == 0 {
		return nil
	}

	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)

	log.Printf("Deleting user %s", id)
	_, _, err := usersAPI.DeleteUser(id)

	return err
}

func generatePhoneResourceWithCustomAttrs(config *phoneConfig, otherAttrs ...string) string {
	lineStrs := make([]string, len(config.lineAddresses))
	for i, val := range config.lineAddresses {
		lineStrs[i] = fmt.Sprintf("\"%s\"", val)
	}

	webRtcUser := ""
	if len(config.webRtcUserId) != 0 {
		webRtcUser = fmt.Sprintf(`web_rtc_user_id = "%s"`, config.webRtcUserId)
	}

	finalConfig := fmt.Sprintf(`resource "genesyscloud_telephony_providers_edges_phone" "%s" {
		name = "%s"
		state = "%s"
		site_id = "%s"
		phone_base_settings_id = "%s"
		line_base_settings_id = "%s"
		line_addresses = [%s]
		depends_on=[%s]
		%s
		%s
	}
	`, config.phoneRes,
		config.name,
		config.state,
		config.siteId,
		config.phoneBaseSettingsId,
		config.lineBaseSettingsId,
		strings.Join(lineStrs, ","),
		config.depends_on,
		webRtcUser,
		strings.Join(otherAttrs, "\n"),
	)

	return finalConfig
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
