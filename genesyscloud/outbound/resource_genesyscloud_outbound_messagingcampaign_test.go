package outbound

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	obDigRuleset "terraform-provider-genesyscloud/genesyscloud/outbound_digitalruleset"
	obDnclist "terraform-provider-genesyscloud/genesyscloud/outbound_dnclist"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"

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
		contactListResourceLabel = "contact_list"
		contactListName          = "Contact List " + uuid.NewString()
		column1                  = "phone"
		column2                  = "zipcode"
		contactListResource      = obContactList.GenerateOutboundContactList(
			contactListResourceLabel,
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
		resourceLabel                 = "messaging_campaign"
		name                          = "Test Messaging Campaign " + uuid.NewString()
		messagesPerMin                = "10"
		alwaysRunning                 = util.FalseValue
		smsConfigMessageColumn        = column1
		smsConfigPhoneColumn          = column1
		smsConfigSenderSMSPhoneNumber = "+19198793429"

		// Messaging Campaign Updated fields
		nameUpdate           = "Test Messaging Campaign " + uuid.NewString()
		messagesPerMinUpdate = "15"
		alwaysRunningUpdate  = util.TrueValue

		// DNC List
		dncListResourceLabel = "dnc_list"
		dncListName          = "Test DNC List " + uuid.NewString()
		dncListResource      = obDnclist.GenerateOutboundDncListBasic(
			dncListResourceLabel,
			dncListName,
		)

		// Contact List Filter
		clfResourceLabel          = "contact_list_filter"
		clfName                   = "Contact List Filter " + uuid.NewString()
		contactListFilterResource = obContactListFilter.GenerateOutboundContactListFilter(
			clfResourceLabel,
			clfName,
			"genesyscloud_outbound_contact_list."+contactListResourceLabel+".id",
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

		callableTimeSetResourceLabel = "callable_time_set"
		callableTimeSetName          = "Test CTS " + uuid.NewString()
		callableTimeSetResource      = obCallableTimeset.GenerateOutboundCallabletimeset(
			callableTimeSetResourceLabel,
			callableTimeSetName,
			obCallableTimeset.GenerateCallableTimesBlock(
				"Europe/Dublin",
				obCallableTimeset.GenerateTimeSlotsBlock("07:00:00", "18:00:00", "3"),
				obCallableTimeset.GenerateTimeSlotsBlock("09:30:00", "22:30:00", "5"),
			),
		)

		digRuleSetResourceLabel                     = "ruleset"
		ruleSetName                                 = "a tf test digitalruleset " + uuid.NewString()
		digitalRulesetResource, digRulesetReference = obDigRuleset.GenerateSimpleOutboundDigitalRuleSet(
			digRuleSetResourceLabel,
			ruleSetName,
		)

		digRuleSet2ResourceLabel                      = "ruleset_2"
		ruleSet2Name                                  = "a tf test digitalruleset 2" + uuid.NewString()
		digitalRuleset2Resource, digRuleset2Reference = obDigRuleset.GenerateSimpleOutboundDigitalRuleSet(
			digRuleSet2ResourceLabel,
			ruleSet2Name,
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
					digitalRulesetResource +
					digitalRuleset2Resource +
					generateOutboundMessagingCampaignResource(
						resourceLabel,
						name,
						"genesyscloud_outbound_contact_list."+contactListResourceLabel+".id",
						strconv.Quote("off"),
						messagesPerMin,
						alwaysRunning,
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceLabel+".id",
						[]string{"genesyscloud_outbound_dnclist." + dncListResourceLabel + ".id"},
						[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceLabel + ".id"},
						[]string{}, // rule_set_ids
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
							util.TrueValue,
						),
						generateDynamicContactQueueingSettings(
							util.FalseValue, // sort
							util.TrueValue,  // filter
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "messages_per_minute", messagesPerMin),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "always_running", alwaysRunning),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "campaign_status", "off"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "sms_config.0.message_column", smsConfigMessageColumn),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "sms_config.0.phone_column", smsConfigPhoneColumn),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "sms_config.0.sender_sms_phone_number", smsConfigSenderSMSPhoneNumber),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.0.field_name", column1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.0.direction", "ASC"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.0.numeric", util.FalseValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.1.field_name", column2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.1.direction", "DESC"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.1.numeric", util.TrueValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "dynamic_contact_queueing_settings.0.sort", util.FalseValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "dynamic_contact_queueing_settings.0.filter", util.TrueValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "rule_set_ids.#", "0"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "callable_time_set_id",
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceLabel, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "dnc_list_ids.0",
						"genesyscloud_outbound_dnclist."+dncListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceLabel, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceLabel, "id"),
					provider.TestDefaultHomeDivision(ResourceType+"."+resourceLabel),
				),
			},
			// Update dynamicContactQueueingSettings
			{
				Config: dncListResource +
					contactListResource +
					contactListFilterResource +
					callableTimeSetResource +
					digitalRulesetResource +
					digitalRuleset2Resource +
					generateOutboundMessagingCampaignResource(
						resourceLabel,
						name,
						"genesyscloud_outbound_contact_list."+contactListResourceLabel+".id",
						strconv.Quote("off"),
						messagesPerMin,
						alwaysRunning,
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceLabel+".id",
						[]string{"genesyscloud_outbound_dnclist." + dncListResourceLabel + ".id"},
						[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceLabel + ".id"},
						[]string{digRulesetReference + ".id", digRuleset2Reference + ".id"}, // rule_set_ids
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
							util.TrueValue,
						),
						generateDynamicContactQueueingSettings(
							util.FalseValue, // sort
							util.FalseValue, // filter
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "messages_per_minute", messagesPerMin),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "always_running", alwaysRunning),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "campaign_status", "off"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "sms_config.0.message_column", smsConfigMessageColumn),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "sms_config.0.phone_column", smsConfigPhoneColumn),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "sms_config.0.sender_sms_phone_number", smsConfigSenderSMSPhoneNumber),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.0.field_name", column1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.0.direction", "ASC"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.0.numeric", util.FalseValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.1.field_name", column2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.1.direction", "DESC"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.1.numeric", util.TrueValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "dynamic_contact_queueing_settings.0.sort", util.FalseValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "dynamic_contact_queueing_settings.0.filter", util.FalseValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "rule_set_ids.#", "2"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "callable_time_set_id",
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceLabel, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "dnc_list_ids.0",
						"genesyscloud_outbound_dnclist."+dncListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceLabel, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "rule_set_ids.0",
						digRulesetReference, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "rule_set_ids.1",
						digRuleset2Reference, "id"),
					provider.TestDefaultHomeDivision(ResourceType+"."+resourceLabel),
				),
			},
			{
				Config: dncListResource +
					contactListResource +
					contactListFilterResource +
					callableTimeSetResource +
					digitalRulesetResource +
					digitalRuleset2Resource +
					generateOutboundMessagingCampaignResource(
						resourceLabel,
						name,
						"genesyscloud_outbound_contact_list."+contactListResourceLabel+".id",
						strconv.Quote("on"),
						messagesPerMin,
						alwaysRunning,
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceLabel+".id",
						[]string{"genesyscloud_outbound_dnclist." + dncListResourceLabel + ".id"},
						[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceLabel + ".id"},
						[]string{digRulesetReference + ".id", digRuleset2Reference + ".id"}, // rule_set_ids
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
							util.TrueValue,
						),
						generateDynamicContactQueueingSettings(
							util.FalseValue, // sort
							util.FalseValue, // filter
						),
					),
				Check: resource.ComposeTestCheckFunc(
					// Check that the DiffSuppressFunc is working
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "messages_per_minute", messagesPerMin),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "always_running", alwaysRunning),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "sms_config.0.message_column", smsConfigMessageColumn),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "sms_config.0.phone_column", smsConfigPhoneColumn),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "sms_config.0.sender_sms_phone_number", smsConfigSenderSMSPhoneNumber),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.0.field_name", column1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.0.direction", "ASC"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.0.numeric", util.FalseValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.1.field_name", column2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.1.direction", "DESC"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.1.numeric", util.TrueValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "dynamic_contact_queueing_settings.0.sort", util.FalseValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "dynamic_contact_queueing_settings.0.filter", util.FalseValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "rule_set_ids.#", "2"),
					util.VerifyAttributeInArrayOfPotentialValues(ResourceType+"."+resourceLabel, "campaign_status", []string{"on", "complete"}),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "callable_time_set_id",
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceLabel, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "dnc_list_ids.0",
						"genesyscloud_outbound_dnclist."+dncListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceLabel, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "rule_set_ids.0",
						digRulesetReference, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "rule_set_ids.1",
						digRuleset2Reference, "id"),
					provider.TestDefaultHomeDivision(ResourceType+"."+resourceLabel),
				),
			},
			{
				// Update
				Config: dncListResource +
					contactListResource +
					contactListFilterResource +
					callableTimeSetResource +
					digitalRulesetResource +
					digitalRuleset2Resource +
					generateOutboundMessagingCampaignResource(
						resourceLabel,
						nameUpdate,
						"genesyscloud_outbound_contact_list."+contactListResourceLabel+".id",
						strconv.Quote("on"),
						messagesPerMinUpdate,
						alwaysRunningUpdate,
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceLabel+".id",
						[]string{"genesyscloud_outbound_dnclist." + dncListResourceLabel + ".id"},
						[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceLabel + ".id"},
						[]string{}, // rule_set_ids
						generateOutboundMessagingCampaignSmsConfig(
							smsConfigMessageColumn,
							smsConfigPhoneColumn,
							smsConfigSenderSMSPhoneNumber,
						),
						GenerateOutboundMessagingCampaignContactSort(
							column1,
							"DESC",
							util.TrueValue,
						),
						GenerateOutboundMessagingCampaignContactSort(
							column2,
							"",
							"",
						),
						generateDynamicContactQueueingSettings(
							util.FalseValue, // sort
							util.FalseValue, // filter
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", nameUpdate),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "messages_per_minute", messagesPerMinUpdate),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "always_running", alwaysRunningUpdate),
					util.VerifyAttributeInArrayOfPotentialValues(ResourceType+"."+resourceLabel, "campaign_status", []string{"on", "complete"}),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "sms_config.0.message_column", smsConfigMessageColumn),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "sms_config.0.phone_column", smsConfigPhoneColumn),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "sms_config.0.sender_sms_phone_number", smsConfigSenderSMSPhoneNumber),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.0.field_name", column1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.0.direction", "DESC"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.0.numeric", util.TrueValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.1.field_name", column2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.1.direction", "ASC"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contact_sorts.1.numeric", util.FalseValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "dynamic_contact_queueing_settings.0.sort", util.FalseValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "dynamic_contact_queueing_settings.0.filter", util.FalseValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "rule_set_ids.#", "0"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "callable_time_set_id",
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceLabel, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceLabel, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceLabel, "id"),
					provider.TestDefaultHomeDivision(ResourceType+"."+resourceLabel),
				),
			},
			{
				ResourceName:      ResourceType + "." + resourceLabel,
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
	resourceLabel string,
	name string,
	contactListId string,
	campaignStatus string,
	messagesPerMinute string,
	alwaysRunning string,
	callableTimeSetId string,
	dncListIds,
	contactListFilterIds,
	ruleSetIds []string,
	nestedBlocks ...string,
) string {
	return fmt.Sprintf(`
resource "%s" "%s" {
	name                    = "%s"
	contact_list_id         = %s
    campaign_status         = %s
    messages_per_minute     = %s
    always_running          = %s
    callable_time_set_id    = %s
	dnc_list_ids            = [%s]
	contact_list_filter_ids = [%s]
	rule_set_ids            = [%s]
    %s
}
`, ResourceType, resourceLabel, name, contactListId, campaignStatus, messagesPerMinute, alwaysRunning, callableTimeSetId,
		strings.Join(dncListIds, ", "), strings.Join(contactListFilterIds, ", "),
		strings.Join(ruleSetIds, ", "), strings.Join(nestedBlocks, "\n"))
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

func generateDynamicContactQueueingSettings(sort, filter string) string {
	return fmt.Sprintf(`
	dynamic_contact_queueing_settings {
		sort   = %s
		filter = %s
	}
`, sort, filter)
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
