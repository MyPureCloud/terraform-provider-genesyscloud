package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v80/platformclientv2"
	"strconv"
	"strings"
	"testing"
)

func TestAccResourceOutboundMessagingCampaign(t *testing.T) {
	t.Parallel()
	var (
		// Contact list
		contactListResourceId = "contact_list"
		contactListName       = "Contact List " + uuid.NewString()
		column1               = "phone"
		column2               = "zipcode"
		contactListResource   = generateOutboundContactList(
			contactListResourceId,
			contactListName,
			"",
			"",
			[]string{},
			[]string{strconv.Quote(column1), strconv.Quote(column2)},
			"",
			"",
			"",
			"",
			generatePhoneColumnsBlock(
				column1,
				"cell",
				column1,
			),
		)

		// Messaging Campaign
		resourceId                    = "messaging_campaign"
		name                          = "Test Messaging Campaign " + uuid.NewString()
		messagesPerMin                = "10"
		callableTimeSetId             = "5654dc1a-874d-439f-83af-f3c1271dcf7c"
		alwaysRunning                 = falseValue
		smsConfigMessageColumn        = column1
		smsConfigPhoneColumn          = column1
		smsConfigSenderSMSPhoneNumber = "+19197050640"

		// Messaging Campaign Updated fields
		nameUpdate           = "Test Messaging Campaign " + uuid.NewString()
		messagesPerMinUpdate = "15"
		alwaysRunningUpdate  = trueValue

		// DNC List
		dncListResourceId = "dnc_list"
		dncListName       = "Test DNC List " + uuid.NewString()
		dncListResource   = generateOutboundDncListBasic(
			dncListResourceId,
			dncListName,
		)

		// Contact List Filter
		clfResourceId             = "contact_list_filter"
		clfName                   = "Contact List Filter " + uuid.NewString()
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

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: dncListResource + contactListResource + contactListFilterResource + generateOutboundMessagingCampaignResource(
					resourceId,
					name,
					"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
					"off",
					messagesPerMin,
					alwaysRunning,
					callableTimeSetId,
					[]string{"genesyscloud_outbound_dnclist." + dncListResourceId + ".id"},
					[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceId + ".id"},
					generateOutboundMessagingCampaignSmsConfig(
						smsConfigMessageColumn,
						smsConfigPhoneColumn,
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
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "messages_per_minute", messagesPerMin),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "always_running", alwaysRunning),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "campaign_status", "off"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "callable_time_set_id", callableTimeSetId),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.message_column", smsConfigMessageColumn),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.phone_column", smsConfigPhoneColumn),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.sender_sms_phone_number", smsConfigSenderSMSPhoneNumber),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.field_name", column1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.direction", "ASC"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.numeric", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.field_name", column2),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.direction", "DESC"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.numeric", trueValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "dnc_list_ids.0",
						"genesyscloud_outbound_dnclist."+dncListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					testDefaultHomeDivision("genesyscloud_outbound_messagingcampaign."+resourceId),
				),
			},
			{
				Config: dncListResource + contactListResource + contactListFilterResource + generateOutboundMessagingCampaignResource(
					resourceId,
					name,
					"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
					"on",
					messagesPerMin,
					alwaysRunning,
					callableTimeSetId,
					[]string{"genesyscloud_outbound_dnclist." + dncListResourceId + ".id"},
					[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceId + ".id"},
					generateOutboundMessagingCampaignSmsConfig(
						smsConfigMessageColumn,
						smsConfigPhoneColumn,
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
				),
				Check: resource.ComposeTestCheckFunc(
					// Check that the DiffSuppressFunc is working
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "campaign_status", "complete"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "messages_per_minute", messagesPerMin),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "always_running", alwaysRunning),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "callable_time_set_id", callableTimeSetId),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.message_column", smsConfigMessageColumn),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.phone_column", smsConfigPhoneColumn),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.sender_sms_phone_number", smsConfigSenderSMSPhoneNumber),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.field_name", column1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.direction", "ASC"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.numeric", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.field_name", column2),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.direction", "DESC"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.numeric", trueValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "dnc_list_ids.0",
						"genesyscloud_outbound_dnclist."+dncListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					testDefaultHomeDivision("genesyscloud_outbound_messagingcampaign."+resourceId),
				),
			},
			{
				// Update
				Config: dncListResource + contactListResource + contactListFilterResource + generateOutboundMessagingCampaignResource(
					resourceId,
					nameUpdate,
					"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
					"on",
					messagesPerMinUpdate,
					alwaysRunningUpdate,
					callableTimeSetId,
					[]string{"genesyscloud_outbound_dnclist." + dncListResourceId + ".id"},
					[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceId + ".id"},
					generateOutboundMessagingCampaignSmsConfig(
						smsConfigMessageColumn,
						smsConfigPhoneColumn,
						smsConfigSenderSMSPhoneNumber,
					),
					generateOutboundMessagingCampaignContactSort(
						column1,
						"DESC",
						trueValue,
					),
					generateOutboundMessagingCampaignContactSort(
						column2,
						"",
						"",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "name", nameUpdate),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "messages_per_minute", messagesPerMinUpdate),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "always_running", alwaysRunningUpdate),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "campaign_status", "complete"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "callable_time_set_id", callableTimeSetId),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.message_column", smsConfigMessageColumn),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.phone_column", smsConfigPhoneColumn),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "sms_config.0.sender_sms_phone_number", smsConfigSenderSMSPhoneNumber),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.field_name", column1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.direction", "DESC"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.0.numeric", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.field_name", column2),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.direction", "ASC"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_sorts.1.numeric", falseValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_messagingcampaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					testDefaultHomeDivision("genesyscloud_outbound_messagingcampaign."+resourceId),
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
		callableTimeSetId = fmt.Sprintf(`callable_time_set_id = "%s"`, callableTimeSetId)
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

func generateOutboundMessagingCampaignContactSort(fieldName string, direction string, numeric string) string {
	if direction != "" {
		direction = fmt.Sprintf(`direction = "%s"`, direction)
	}
	if numeric != "" {
		numeric = fmt.Sprintf(`numeric = %s`, numeric)
	}
	return fmt.Sprintf(`
	contact_sorts {
		field_name = "%s"
		%s
        %s
	}
`, fieldName, direction, numeric)
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
		} else if isStatus404(resp) {
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
