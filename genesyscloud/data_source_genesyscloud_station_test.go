package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceStation(t *testing.T) {
	var (
		phoneRes    = "phone1234"
		name1       = "test-phone_" + uuid.NewString()
		stateActive = "active"

		phoneBaseSettingsRes  = "phoneBaseSettings1234"
		phoneBaseSettingsName = "phoneBaseSettings " + uuid.NewString()

		userRes1   = "user1"
		userName1  = "test_webrtc_user_" + uuid.NewString()
		userEmail1 = userName1 + "@test.com"

		userTitle      = "Senior Director"
		userDepartment = "Development"

		// station
		stationDataRes = "station1234"
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
		nil, // no line addresses
		"genesyscloud_user." + userRes1 + ".id",
		"", // no depends on
	},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: config + generateStationDataSource(
					stationDataRes,
					"genesyscloud_telephony_providers_edges_phone."+phoneRes+".name",
					"genesyscloud_telephony_providers_edges_phone."+phoneRes,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_station."+stationDataRes, "name", "genesyscloud_telephony_providers_edges_phone."+phoneRes, "name"),
				),
			},
		},
		CheckDestroy: testVerifyWebRtcPhoneDestroyed,
	})
}

func generateStationDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_station" "%s" {
		name = %s
        depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
