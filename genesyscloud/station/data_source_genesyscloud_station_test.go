package station

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	edgePhone "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phone"
	phoneBaseSettings "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phonebasesettings"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"terraform-provider-genesyscloud/genesyscloud/user"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceStation(t *testing.T) {
	var (
		phoneResourceLabel = "phone1234"
		name1              = "test-phone_" + uuid.NewString()
		stateActive        = "active"

		phoneBaseSettingsResourceLabel = "phoneBaseSettings1234"
		phoneBaseSettingsName          = "phoneBaseSettings " + uuid.NewString()

		userResourceLabel1 = "user1"
		userName1          = "test_webrtc_user_" + uuid.NewString()
		userEmail1         = userName1 + "@test.com"

		userTitle      = "Senior Director"
		userDepartment = "Development"

		// station
		stationDataResourceLabel = "station1234"
	)

	defaultSiteId, err := edgeSite.GetOrganizationDefaultSiteId(sdkConfig)
	if err != nil {
		t.Fatal(err)
	}

	config := user.GenerateUserResource(
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
	) + edgePhone.GeneratePhoneResourceWithCustomAttrs(&edgePhone.PhoneConfig{
		PhoneResourceLabel:  phoneResourceLabel,
		Name:                name1,
		State:               stateActive,
		SiteId:              fmt.Sprintf("\"%s\"", defaultSiteId),
		PhoneBaseSettingsId: "genesyscloud_telephony_providers_edges_phonebasesettings." + phoneBaseSettingsResourceLabel + ".id",
		WebRtcUserId:        "genesyscloud_user." + userResourceLabel1 + ".id",
		DependsOn:           "", // no depends on
	},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: config + generateStationDataSource(
					stationDataResourceLabel,
					"genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel+".name",
					"genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_station."+stationDataResourceLabel, "name", "genesyscloud_telephony_providers_edges_phone."+phoneResourceLabel, "name"),
				),
			},
		},
		CheckDestroy: edgePhone.TestVerifyWebRtcPhoneDestroyed,
	})
}

func generateStationDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_station" "%s" {
		name = %s
        depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
