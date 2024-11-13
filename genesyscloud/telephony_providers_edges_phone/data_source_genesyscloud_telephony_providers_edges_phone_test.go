package telephony_providers_edges_phone

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	phoneBaseSettings "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phonebasesettings"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"terraform-provider-genesyscloud/genesyscloud/user"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourcePhone(t *testing.T) {
	t.Parallel()
	var (
		phoneResourceLabel     = "phone1234"
		phoneDataResourceLabel = "phoneData"
		name1                  = "test-phone" + uuid.NewString()
		stateActive            = "active"

		phoneBaseSettingsResourceLabel = "phoneBaseSettings1234"
		phoneBaseSettingsName          = "phoneBaseSettings " + uuid.NewString()

		userResourceLabel1 = "user1"
		userName1          = "test_webrtc_user_" + uuid.NewString()
		userEmail1         = userName1 + "@test.com"

		userTitle      = "Senior Director"
		userDepartment = "Development"
	)

	defaultSiteId, err := edgeSite.GetOrganizationDefaultSiteId(sdkConfig)
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: user.GenerateUserResource(
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
				) + phoneBaseSettings.GeneratePhoneBaseSettingsResourceWithCustomAttrs(
					phoneBaseSettingsResourceLabel,
					phoneBaseSettingsName,
					"phoneBaseSettings description",
					"inin_webrtc_softphone.json",
				) + GeneratePhoneResourceWithCustomAttrs(&PhoneConfig{
					phoneResourceLabel,
					name1,
					stateActive,
					fmt.Sprintf("\"%s\"", defaultSiteId),
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
					), generatePhoneProperties(uuid.NewString()),
				) + generatePhoneDataSource(
					phoneDataResourceLabel,
					name1,
					"genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_phone."+phoneDataResourceLabel, "id", "genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "id"),
				),
			},
		},
	})
}

func generatePhoneDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_phone" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
