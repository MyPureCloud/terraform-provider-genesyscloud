package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	sdkConfig *platformclientv2.Configuration
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
	t.Parallel()
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

	config := generateOrganizationMe() + generateUserResource(
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
	) + generatePhoneBaseSettingsResourceWithCustomAttrs(
		phoneBaseSettingsRes,
		phoneBaseSettingsName,
		"phoneBaseSettings description",
		"inin_webrtc_softphone.json",
	) + generatePhoneResourceWithCustomAttrs(&phoneConfig{
		phoneRes,
		name1,
		stateActive,
		"data.genesyscloud_organizations_me.me.default_site_id",
		"genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsRes + ".id",
		"genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsRes + ".line_base_settings_id",
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

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "state", stateActive),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneRes, "site_id", "data.genesyscloud_organizations_me.me", "default_site_id"),
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
				// Update phone with new user and name
				Config: generateOrganizationMe() + generateUserResource(
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
				) + generatePhoneBaseSettingsResourceWithCustomAttrs(
					phoneBaseSettingsRes,
					phoneBaseSettingsName,
					"phoneBaseSettings description",
					"inin_webrtc_softphone.json",
				) + generatePhoneResourceWithCustomAttrs(&phoneConfig{
					phoneRes,
					name2,
					stateActive,
					"data.genesyscloud_organizations_me.me.default_site_id",
					"genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsRes + ".id",
					"genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsRes + ".line_base_settings_id",
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
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "state", stateActive),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneRes, "site_id", "data.genesyscloud_organizations_me.me", "default_site_id"),
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
		CheckDestroy: testVerifyWebRtcPhoneDestroyed,
	})
}

func deleteDidPoolWithNumber(number string) {
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

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
				edgesAPI.DeleteTelephonyProvidersEdgesDidpool(*didPool.Id)
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func TestAccResourcePhoneStandalone(t *testing.T) {
	t.Parallel()
	didPoolResource1 := "test-didpool1"
	number := "+14175530013"
	err := authorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	deleteDidPoolWithNumber(number)
	lineAddresses := []string{number}
	phoneRes := "phone_standalone1234"
	name1 := "test-phone-standalone_" + uuid.NewString()
	stateActive := "active"
	phoneBaseSettingsRes := "phoneBaseSettings1234"
	phoneBaseSettingsName := "phoneBaseSettings " + uuid.NewString()

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

	config += generatePhoneBaseSettingsResourceWithCustomAttrs(
		phoneBaseSettingsRes,
		phoneBaseSettingsName,
		"phoneBaseSettings description",
		"generic_sip.json",
	) + generatePhoneResourceWithCustomAttrs(&phoneConfig{
		phoneRes,
		name1,
		stateActive,
		"data.genesyscloud_organizations_me.me.default_site_id",
		"genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsRes + ".id",
		"genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsRes + ".line_base_settings_id",
		lineAddresses,
		"", // no web rtc user
		"genesyscloud_telephony_providers_edges_did_pool." + didPoolResource1,
	}, capabilities)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateOrganizationMe() + config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_phone."+phoneRes, "state", stateActive),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneRes, "site_id", "data.genesyscloud_organizations_me.me", "default_site_id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneRes, "line_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "line_base_settings_id"),
					resource.TestCheckResourceAttrPair("genesyscloud_telephony_providers_edges_phone."+phoneRes, "phone_base_settings_id", "genesyscloud_telephony_providers_edges_phonebasesettings."+phoneBaseSettingsRes, "id"),
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
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_telephony_providers_edges_phone" {
			continue
		}

		phone, resp, err := edgesAPI.GetTelephonyProvidersEdgesPhone(rs.Primary.ID)
		if phone != nil {
			return fmt.Errorf("Phone (%s) still exists", rs.Primary.ID)
		} else if isStatus404(resp) {
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

func authorizeSdk() error {
	// Create new config
	sdkConfig = platformclientv2.GetDefaultConfiguration()

	sdkConfig.BasePath = getRegionBasePath(os.Getenv("GENESYSCLOUD_REGION"))

	err := sdkConfig.AuthorizeClientCredentials(os.Getenv("GENESYSCLOUD_OAUTHCLIENT_ID"), os.Getenv("GENESYSCLOUD_OAUTHCLIENT_SECRET"))
	if err != nil {
		return err
	}

	return nil
}

func generatePhoneResourceWithCustomAttrs(config *phoneConfig, otherAttrs ...string) string {
	lineStrs := make([]string, len(config.lineAddresses))
	for i, val := range config.lineAddresses {
		lineStrs[i] = fmt.Sprintf("\"%s\"", val)
	}

	webRtcUser := ""
	if len(config.webRtcUserId) != 0 {
		webRtcUser = fmt.Sprintf(`web_rtc_user_id = %s`, config.webRtcUserId)
	}

	finalConfig := fmt.Sprintf(`resource "genesyscloud_telephony_providers_edges_phone" "%s" {
		name = "%s"
		state = "%s"
		site_id = %s
		phone_base_settings_id = %s
		line_base_settings_id = %s
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

func generateOrganizationMe() string {
	return `
data "genesyscloud_organizations_me" "me" {}
`
}
