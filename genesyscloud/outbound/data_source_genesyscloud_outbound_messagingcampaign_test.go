package outbound

import (
	"fmt"
	"strconv"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

var trueValue = "true"

/*
This test can only pass in a test org because it requires an active provisioned sms phone number
Endpoint `POST /api/v2/routing/sms/phonenumbers` creates an active/valid phone number in test orgs only.
*/
func TestAccDataSourceOutboundMessagingCampaign(t *testing.T) {

	var (
		resourceId          = "campaign"
		dataSourceId        = "campaign_data"
		digitalCampaignName = "Test Digital Campaign " + uuid.NewString()

		clfResourceId         = "clf"
		clfName               = "Test CLF " + uuid.NewString()
		contactListResourceId = "contact_list"
		contactListName       = "Test Contact List " + uuid.NewString()
		column1               = "phone"
		column2               = "zipcode"

		smsConfigSenderSMSPhoneNumber = "+19198793428"

		callableTimeSetResourceId = "callable_time_set"
		callableTimeSetName       = "Test CTS " + uuid.NewString()
		callableTimeSetResource   = generateOutboundCallabletimeset(
			callableTimeSetResourceId,
			callableTimeSetName,
			generateCallableTimesBlock(
				"Europe/Dublin",
				generateTimeSlotsBlock("07:00:00", "18:00:00", "3"),
				generateTimeSlotsBlock("09:30:00", "22:30:00", "5"),
			),
		)

		contactListResource = obContactList.GenerateOutboundContactList(
			contactListResourceId,
			contactListName,
			nullValue,
			nullValue,
			[]string{},
			[]string{strconv.Quote(column1), strconv.Quote(column2)},
			falseValue,
			nullValue,
			nullValue,
			obContactList.GeneratePhoneColumnsBlock(
				column1,
				"cell",
				strconv.Quote(column1),
			),
		)

		contactListFilterResource = generateOutboundContactListFilter(
			clfResourceId,
			clfName,
			"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
			"",
			generateOutboundContactListFilterClause(
				"",
				generateOutboundContactListFilterPredicates(
					column1,
					"alphabetic",
					"EQUALS",
					"XYZ",
					"",
					"",
				),
			),
		)
	)

	config, err := gcloud.AuthorizeSdk()
	if err != nil {
		t.Errorf("failed to authorize client: %v", err)
	}
	api := platformclientv2.NewRoutingApiWithConfig(config)
	err = createRoutingSmsPhoneNumber(smsConfigSenderSMSPhoneNumber, api)
	if err != nil {
		t.Errorf("error creating sms phone number %s: %v", smsConfigSenderSMSPhoneNumber, err)
	}
	defer func() {
		_, err := api.DeleteRoutingSmsPhonenumber(smsConfigSenderSMSPhoneNumber)
		if err != nil {
			t.Logf("error deleting phone number %s: %v", smsConfigSenderSMSPhoneNumber, err)
		}
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: contactListResource +
					contactListFilterResource +
					callableTimeSetResource +
					generateOutboundMessagingCampaignResource(
						resourceId,
						digitalCampaignName,
						"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
						"",
						"10",
						falseValue,
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceId+".id",
						[]string{},
						[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceId + ".id"},
						generateOutboundMessagingCampaignSmsConfig(
							column1,
							column1,
							smsConfigSenderSMSPhoneNumber,
						),
						generateOutboundMessagingCampaignContactSort(
							column1,
							"",
							"",
						),
						generateOutboundMessagingCampaignContactSort(
							column2,
							"DESC",
							trueValue,
						),
					) + generateOutboundMessagingCampaignDataSource(
					dataSourceId,
					digitalCampaignName,
					"genesyscloud_outbound_messagingcampaign."+resourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_messagingcampaign."+dataSourceId, "id",
						"genesyscloud_outbound_messagingcampaign."+resourceId, "id"),
				),
			},
		},
	})
}

func generateOutboundMessagingCampaignDataSource(id string, name string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_outbound_messagingcampaign" "%s" {
	name = "%s"
	depends_on = [%s]
}
`, id, name, dependsOn)
}
