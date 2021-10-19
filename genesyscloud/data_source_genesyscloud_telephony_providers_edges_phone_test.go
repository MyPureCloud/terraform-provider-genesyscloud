package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strconv"
	"testing"
)

func TestAccDataSourcePhone(t *testing.T) {
	var (
		phoneRes     = "phone1234"
		phoneDataRes = "phoneData"
		name1        = "test-phone" + uuid.NewString()
		stateActive  = "active"

		phoneBaseSettingsRes  = "phoneBaseSettings1234"
		phoneBaseSettingsName = "phoneBaseSettings " + uuid.NewString()

		userRes1   = "user1"
		userName1  = "test_webrtc_user_" + uuid.NewString()
		userEmail1 = userName1 + "@test.com"

		userTitle      = "Senior Director"
		userDepartment = "Development"
	)

	err := authorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	siteId, err := getDefaultSiteId()
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateUserResource(
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
					siteId,
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
				) + generatePhoneDataSource(
					phoneDataRes,
					name1,
					"genesyscloud_telephony_providers_edges_phone."+phoneRes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_phone."+phoneDataRes, "id", "genesyscloud_telephony_providers_edges_phone."+phoneRes, "id"),
				),
			},
		},
	})
}

func generatePhoneDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_phone" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
