package outbound

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	obDnclist "terraform-provider-genesyscloud/genesyscloud/outbound_dnclist"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"

	obCallableTimeset "terraform-provider-genesyscloud/genesyscloud/outbound_callabletimeset"
	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	obContactListFilter "terraform-provider-genesyscloud/genesyscloud/outbound_contactlistfilter"
)

/*
This test can only pass in a test org because it requires an active provisioned sms phone number
Endpoint `POST /api/v2/routing/sms/phonenumbers` creates an active/valid phone number in test orgs only.
*/
func TestAccResourceOutboundMessagingCampaign(t *testing.T) {

	t.Parallel()
	var (
		// Contact list
		contactListResourceId = "contact_list"
		contactListName       = "Contact List " + uuid.NewString()
		column1               = "phone"
		column2               = "zipcode"
		contactListResource   = obContactList.GenerateOutboundContactList(
			contactListResourceId,
			contactListName,
			util.NullValue,
			util.NullValue,
			[]string{},
			[]string{strconv.Quote(column1), strconv.Quote(column2)},
			util.NullValue,
			util.NullValue,
			util.NullValue,
			obContactList.GeneratePhoneColumnsBlock(
				column1,
				"cell",
				strconv.Quote(column1),
			),
		)

		// Messaging Campaign
		resourceId                    = "messaging_campaign"
		name                          = "Test Messaging Campaign " + uuid.NewString()
		messagesPerMin                = "10"
		alwaysRunning                 = util.FalseValue
		smsConfigMessageColumn        = column1
		smsConfigPhoneColumn          = column1
		smsConfigSenderSMSPhoneNumber = "+19198793429"

		// Messaging Campaign Updated fields
		nameUpdate           = "Test Messaging Campaign " + uuid.NewString()
		messagesPerMinUpdate = "15"
		alwaysRunningUpdate  = TrueValue

		// DNC List
		dncListResourceId = "dnc_list"
		dncListName       = "Test DNC List " + uuid.NewString()
		dncListResource   = obDnclist.GenerateOutboundDncListBasic(
			dncListResourceId,
			dncListName,
		)

		// Contact List Filter
		clfResourceId             = "contact_list_filter"
		clfName                   = "Contact List Filter " + uuid.NewString()
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
				Config: dncListResource +
					contactListResource +
					contactListFilterResource +
					callableTimeSetResource +
					generateOutboundMessagingCampaignResource(
						resourceId,
						name,
						"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
						"off",
						messagesPerMin,
						alwaysRunning,
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceId+".id",
						[]string{"genesyscloud_outbound_dnclist." + dncListResourceId + ".id"},
						[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceId + ".id"},
						generateOutboundMessagingCampaignSmsConfig(
							smsConfigMessageColumn,
							smsConfigPhoneColumn,
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
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "messages_per_minute", messagesPerMin),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "always_running", alwaysRunning),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "campaign_status", "off"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.message_column", smsConfigMessageColumn),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.phone_column", smsConfigPhoneColumn),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.sender_sms_phone_number", smsConfigSenderSMSPhoneNumber),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.field_name", column1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.direction", "ASC"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.numeric", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.field_name", column2),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.direction", "DESC"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.numeric", TrueValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "callable_time_set_id",
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "dnc_list_ids.0",
						"genesyscloud_outbound_dnclist."+dncListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					provider.TestDefaultHomeDivision("genesyscloud_outbound_messagingcampaign."+resourceId),
				),
			},
			{
				Config: dncListResource +
					contactListResource +
					contactListFilterResource +
					callableTimeSetResource +
					generateOutboundMessagingCampaignResource(
						resourceId,
						name,
						"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
						"on",
						messagesPerMin,
						alwaysRunning,
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceId+".id",
						[]string{"genesyscloud_outbound_dnclist." + dncListResourceId + ".id"},
						[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceId + ".id"},
						generateOutboundMessagingCampaignSmsConfig(
							smsConfigMessageColumn,
							smsConfigPhoneColumn,
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
					),
				Check: resource.ComposeTestCheckFunc(
					// Check that the DiffSuppressFunc is working
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "messages_per_minute", messagesPerMin),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "always_running", alwaysRunning),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.message_column", smsConfigMessageColumn),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.phone_column", smsConfigPhoneColumn),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.sender_sms_phone_number", smsConfigSenderSMSPhoneNumber),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.field_name", column1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.direction", "ASC"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.numeric", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.field_name", column2),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.direction", "DESC"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.numeric", TrueValue),
					util.VerifyAttributeInArrayOfPotentialValues("genesyscloud_outbound_messagingcampaign."+resourceId, "campaign_status", []string{"on", "complete"}),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "callable_time_set_id",
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "dnc_list_ids.0",
						"genesyscloud_outbound_dnclist."+dncListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					provider.TestDefaultHomeDivision("genesyscloud_outbound_messagingcampaign."+resourceId),
				),
			},
			{
				// Update
				Config: dncListResource +
					contactListResource +
					contactListFilterResource +
					callableTimeSetResource +
					generateOutboundMessagingCampaignResource(
						resourceId,
						nameUpdate,
						"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
						"on",
						messagesPerMinUpdate,
						alwaysRunningUpdate,
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceId+".id",
						[]string{"genesyscloud_outbound_dnclist." + dncListResourceId + ".id"},
						[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceId + ".id"},
						generateOutboundMessagingCampaignSmsConfig(
							smsConfigMessageColumn,
							smsConfigPhoneColumn,
							smsConfigSenderSMSPhoneNumber,
						),
						GenerateOutboundMessagingCampaignContactSort(
							column1,
							"DESC",
							TrueValue,
						),
						GenerateOutboundMessagingCampaignContactSort(
							column2,
							"",
							"",
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "name", nameUpdate),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "messages_per_minute", messagesPerMinUpdate),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "always_running", alwaysRunningUpdate),
					util.VerifyAttributeInArrayOfPotentialValues("genesyscloud_outbound_messagingcampaign."+resourceId, "campaign_status", []string{"on", "complete"}),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.message_column", smsConfigMessageColumn),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.phone_column", smsConfigPhoneColumn),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.sender_sms_phone_number", smsConfigSenderSMSPhoneNumber),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.field_name", column1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.direction", "DESC"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.numeric", TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.field_name", column2),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.direction", "ASC"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.numeric", util.FalseValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "callable_time_set_id",
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					provider.TestDefaultHomeDivision("genesyscloud_outbound_messagingcampaign."+resourceId),
				),
			},
			{
				ResourceName:      "genesyscloud_outbound_messagingcampaign." + resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyOutboundMessagingCampaignDestroyed,
	})
}

func createRoutingSmsPhoneNumber(inputSmsPhoneNumber string, api *platformclientv2.RoutingApi) error {
	var (
		phoneNumberType = "local"
		countryCode     = "US"
		status          string
		maxRetries      = 10
	)
	_, resp, err := api.GetRoutingSmsPhonenumber(inputSmsPhoneNumber, "compliance")
	if resp.StatusCode == 200 {
		// Number already exists
		return nil
	} else if resp.StatusCode == http.StatusNotFound {
		body := platformclientv2.Smsphonenumberprovision{
			PhoneNumber:     &inputSmsPhoneNumber,
			PhoneNumberType: &phoneNumberType,
			CountryCode:     &countryCode,
		}
		// POST /api/v2/routing/sms/phonenumbers
		address, _, err := api.PostRoutingSmsPhonenumbers(body)
		if err != nil {
			return err
		}
		// Ensure status transitions to complete before proceeding
		for i := 0; i <= maxRetries; i++ {
			time.Sleep(3 * time.Second)
			// GET /api/v2/routing/sms/phonenumbers/{addressId}
			sdkSmsPhoneNumber, _, err := api.GetRoutingSmsPhonenumber(*address.PhoneNumber, "compliance")
			if err != nil {
				return err
			}
			status = *sdkSmsPhoneNumber.ProvisioningStatus.State
			if status == "Running" {
				if i == maxRetries {
					return fmt.Errorf(`sms phone number status did not transition to "Completed" within max retries %v`, maxRetries)
				}
				continue
			}
			break
		}
		if status == "Failed" {
			return fmt.Errorf(`sms phone number provisioning failed`)
		}
	} else if err != nil {
		return fmt.Errorf("error checking for sms phone number %v: %v", inputSmsPhoneNumber, err)
	}
	return nil
}

func generateOutboundMessagingCampaignResource(
	resourceId string,
	name string,
	contactListId string,
	campaignStatus string,
	messagesPerMinute string,
	alwaysRunning string,
	callableTimeSetId string,
	dncListIds []string,
	contactListFilterIds []string,
	nestedBlocks ...string,
) string {
	if callableTimeSetId != "" {
		callableTimeSetId = fmt.Sprintf(`callable_time_set_id = %s`, callableTimeSetId)
	}
	if alwaysRunning != "" {
		alwaysRunning = fmt.Sprintf(`always_running = %s`, alwaysRunning)
	}
	if campaignStatus != "" {
		campaignStatus = fmt.Sprintf(`campaign_status = "%s"`, campaignStatus)
	}
	return fmt.Sprintf(`
resource "genesyscloud_outbound_messagingcampaign" "%s" {
	name                = "%s"
	contact_list_id     = %s
    %s
    messages_per_minute = %s
    %s
    %s
	dnc_list_ids = [%s]
	contact_list_filter_ids = [%s]
    %s
}
`, resourceId, name, contactListId, campaignStatus, messagesPerMinute, alwaysRunning, callableTimeSetId,
		strings.Join(dncListIds, ", "), strings.Join(contactListFilterIds, ", "), strings.Join(nestedBlocks, "\n"))
}

func generateOutboundMessagingCampaignSmsConfig(
	smsConfigMessageColumn string,
	smsConfigPhoneColumn string,
	smsConfigSenderSMSPhoneNumber string,
) string {
	return fmt.Sprintf(`
    sms_config {
		message_column          = "%s"
		phone_column            = "%s"
 		sender_sms_phone_number = "%s"
	}
`, smsConfigMessageColumn, smsConfigPhoneColumn, smsConfigSenderSMSPhoneNumber)
}

func testVerifyOutboundMessagingCampaignDestroyed(state *terraform.State) error {
	outboundAPI := platformclientv2.NewOutboundApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_outbound_messagingcampaign" {
			continue
		}
		campaign, resp, err := outboundAPI.GetOutboundMessagingcampaign(rs.Primary.ID)
		if campaign != nil {
			return fmt.Errorf("messaging campaign (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Messaging Campaign not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All messaging campaigns destroyed
	return nil
}
