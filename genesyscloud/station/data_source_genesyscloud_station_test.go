package station

import (
	"fmt"
	"strconv"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	edgePhone "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phone"
	phoneBaseSettings "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phonebasesettings"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
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

	defaultSiteId, err := edgeSite.GetOrganizationDefaultSiteId(sdkConfig)
	if err != nil {
		t.Fatal(err)
	}

	config := gcloud.GenerateUserResource(
		userRes1,
		userEmail1,
		userName1,
		gcloud.NullValue, // Defaults to active
		strconv.Quote(userTitle),
		strconv.Quote(userDepartment),
		gcloud.NullValue, // No manager
		gcloud.NullValue, // Default acdAutoAnswer
		"",               // No profile skills
		"",               // No certs
	) + phoneBaseSettings.GeneratePhoneBaseSettingsResourceWithCustomAttrs(
		phoneBaseSettingsRes,
		phoneBaseSettingsName,
		"phoneBaseSettings description",
		"inin_webrtc_softphone.json",
	) + edgePhone.GeneratePhoneResourceWithCustomAttrs(&edgePhone.PhoneConfig{
		PhoneRes:            phoneRes,
		Name:                name1,
		State:               stateActive,
		SiteId:              fmt.Sprintf("\"%s\"", defaultSiteId),
		PhoneBaseSettingsId: "genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsRes + ".id",
		LineAddresses:       nil, // no line addresses
		WebRtcUserId:        "genesyscloud_user." + userRes1 + ".id",
		Depends_on:          "", // no depends on
	},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
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
		CheckDestroy: edgePhone.TestVerifyWebRtcPhoneDestroyed,
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
