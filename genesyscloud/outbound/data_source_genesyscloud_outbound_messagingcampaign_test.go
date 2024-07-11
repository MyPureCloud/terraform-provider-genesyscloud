package outbound

import (
	"fmt"
	"os"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	obCallableTimeset "terraform-provider-genesyscloud/genesyscloud/outbound_callabletimeset"
	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	obContactListFilter "terraform-provider-genesyscloud/genesyscloud/outbound_contactlistfilter"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var TrueValue = "true"

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
		callableTimeSetResource   = obCallableTimeset.GenerateOutboundCallabletimeset(
			callableTimeSetResourceId,
			callableTimeSetName,
			obCallableTimeset.GenerateCallableTimesBlock(
				"Europe/Dublin",
				obCallableTimeset.GenerateTimeSlotsBlock("07:00:00", "18:00:00", "3"),
				obCallableTimeset.GenerateTimeSlotsBlock("09:30:00", "22:30:00", "5"),
			),
		)

		contactListResource = obContactList.GenerateOutboundContactList(
			contactListResourceId,
			contactListName,
			util.NullValue,
			util.NullValue,
			[]string{},
			[]string{strconv.Quote(column1), strconv.Quote(column2)},
			util.FalseValue,
			util.NullValue,
			util.NullValue,
			obContactList.GeneratePhoneColumnsBlock(
				column1,
				"cell",
				strconv.Quote(column1),
			),
		)

		contactListFilterResource = obContactListFilter.GenerateOutboundContactListFilter(
			clfResourceId,
			clfName,
			"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
			"",
			obContactListFilter.GenerateOutboundContactListFilterClause(
				"",
				obContactListFilter.GenerateOutboundContactListFilterPredicates(
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

	if v := os.Getenv("GENESYSCLOUD_REGION"); v == "tca" {
		smsConfigSenderSMSPhoneNumber = "+18159823725"
	}

	config, err := provider.AuthorizeSdk()
	if err != nil {
		t.Errorf("failed to authorize client: %v", err)
	}

	if v := os.Getenv("GENESYSCLOUD_REGION"); v == "us-east-1" {
		api := platformclientv2.NewRoutingApiWithConfig(config)
		err = createRoutingSmsPhoneNumber(smsConfigSenderSMSPhoneNumber, api)
		if err != nil {
			t.Errorf("error creating sms phone number %s: %v", smsConfigSenderSMSPhoneNumber, err)
		}
		//Do not delete the smsPhoneNumber
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
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
						util.FalseValue,
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceId+".id",
						[]string{},
						[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceId + ".id"},
						generateOutboundMessagingCampaignSmsConfig(
							column1,
							column1,
							smsConfigSenderSMSPhoneNumber,
						),
						GenerateOutboundMessagingCampaignContactSort(
							column1,
							"",
							"",
						),
						GenerateOutboundMessagingCampaignContactSort(
							column2,
							"DESC",
							TrueValue,
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
