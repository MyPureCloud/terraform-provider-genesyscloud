package outbound_messagingcampaign

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	obDigRuleset "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_digitalruleset"
	obDnclist "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_dnclist"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"

	obCallableTimeset "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_callabletimeset"
	obContactList "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	obContactListFilter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contactlistfilter"
	responseManagementLibrary "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/responsemanagement_library"
	responseManagementResponse "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/responsemanagement_response"
	routingEmailDomain "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	routingEmailRoute "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_route"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
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
							util.TrueValue,  // filter
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
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "dynamic_contact_queueing_settings.0.filter", util.TrueValue),
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
							util.TrueValue,  // filter
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
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "dynamic_contact_queueing_settings.0.filter", util.TrueValue),
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

/*
Since (api/v2/routing/emails/outbound/domains) has no terraform resource currently,
this tests relies on a pre-created outbound domain "terraformemailconfig.com"
*/
func TestAccResourceOutboundMessagingCampaignWithEmailConfig(t *testing.T) {
	var (
		// contact_list variables

		contactListResourceLabel = "email-config-test"
		contactListName          = "Terraform Test Contact List " + uuid.NewString()
		contactListColumnNames   = []string{
			strconv.Quote("PERSONAL"),
			strconv.Quote("WORK"),
		}
		contactListResource = obContactList.GenerateOutboundContactList(
			contactListResourceLabel,
			contactListName,
			util.NullValue,
			util.NullValue,
			[]string{},
			contactListColumnNames,
			util.NullValue,
			util.NullValue,
			util.NullValue,
			obContactList.GenerateEmailColumnsBlock(
				"PERSONAL",
				"PERSONAL",
				util.NullValue,
			),
			obContactList.GenerateEmailColumnsBlock(
				"WORK",
				"WORK",
				util.NullValue,
			),
		)

		// routing variables

		routingQueueResourceLabel = "email-config-test"
		routingQueueName          = "Terraform Test Routing Queue " + uuid.NewString()
		routingQueueResource      = routingQueue.GenerateRoutingQueueResourceBasic(
			routingQueueResourceLabel,
			routingQueueName,
		)

		routingEmailDomainResourceLabel = "email-config-test"
		routingEmailDomainId            = "terraformtest"
		routingEmailDomainSubdomain     = util.TrueValue
		routingEmailDomainResource      = routingEmailDomain.GenerateRoutingEmailDomainResource(
			routingEmailDomainResourceLabel,
			routingEmailDomainId,
			routingEmailDomainSubdomain,
			util.NullValue,
		)

		routingEmailRouteResourceLabel = "email-config-test"
		routingEmailRoutePattern       = "test"
		routingEmailRouteFromName      = "TerraformAccTest"
		routingEmailRouteResource      = routingEmailRoute.GenerateRoutingEmailRouteResource(
			routingEmailRouteResourceLabel,
			routingEmailDomain.ResourceType+"."+routingEmailDomainResourceLabel+".id",
			routingEmailRoutePattern,
			routingEmailRouteFromName,
			"queue_id = "+routingQueue.ResourceType+"."+routingQueueResourceLabel+".id",
		)

		// responsemanagement variables

		responseManagementLibraryResourceLabel = "email-config-test"
		responseManagementLibraryName          = "Terraform Test Response Management Library " + uuid.NewString()
		responseManagementLibraryResource      = responseManagementLibrary.GenerateResponseManagementLibraryResource(
			responseManagementLibraryResourceLabel,
			responseManagementLibraryName,
		)

		responseManagementResponseResourceLabel = "email-config-test"
		responseManagementResponseName          = "Terraform Test Response Management Response " + uuid.NewString()
		responseManagementResponseName2         = "Terraform Test Response Management Response 2" + uuid.NewString()
		responseManagementResponseType          = "CampaignEmailTemplate"
		responseManagementResponseResource      = responseManagementResponse.GenerateResponseManagementResponseResource(
			responseManagementResponseResourceLabel,
			responseManagementResponseName,
			[]string{responseManagementLibrary.ResourceType + "." + responseManagementLibraryResourceLabel + ".id"},
			util.NullValue,
			util.NullValue,
			strconv.Quote(responseManagementResponseType),
			[]string{},
			responseManagementResponse.GenerateTextsBlock(
				"test@email.com",
				"text/plain",
				strconv.Quote("subject"),
			),
			responseManagementResponse.GenerateTextsBlock(
				"Testing Email Content",
				"text/html",
				strconv.Quote("body"),
			),
		)
		responseManagementResponseResource2 = responseManagementResponse.GenerateResponseManagementResponseResource(
			responseManagementResponseResourceLabel,
			responseManagementResponseName2,
			[]string{responseManagementLibrary.ResourceType + "." + responseManagementLibraryResourceLabel + ".id"},
			util.NullValue,
			util.NullValue,
			strconv.Quote(responseManagementResponseType),
			[]string{},
			responseManagementResponse.GenerateTextsBlock(
				"test@email.com",
				"text/plain",
				strconv.Quote("subject"),
			),
			responseManagementResponse.GenerateTextsBlock(
				"Testing Email Content",
				"text/html",
				strconv.Quote("body"),
			),
		)

		// messaging_campaign variables

		resourceLabel  = "test_messaging_campaign_email_config"
		name           = "Test Messaging Campaign " + uuid.NewString()
		messagesPerMin = "10"
		alwaysRunning  = util.FalseValue
		campaignStatus = "off"

		emailColumns  = "PERSONAL"
		emailColumns2 = "WORK"
		/*
			The API for creating this outbound domain (api/v2/routing/email/outbound/domains) does not have a terraform resource yet.
			currently this test relies on the pre-created outbound domain "terraformemailconfig.com".
		*/
		fromAddressDomainId   = "terraformemailconfig.com"
		fromAddressLocalPart  = "TestEmail"
		fromAddressLocalPart2 = "newPart"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Can be removed once `api/v2/routing/email/outbound/*` is implemented in provider
					err := CheckOutboundDomainExists(fromAddressDomainId)
					if err != nil {
						t.Fatal(err)
					}
				},

				// Create
				Config: contactListResource +
					routingQueueResource +
					routingEmailDomainResource +
					routingEmailRouteResource +
					responseManagementLibraryResource +
					responseManagementResponseResource +
					generateOutboundMessagingCampaignResource(
						resourceLabel,
						name,
						obContactList.ResourceType+"."+contactListResourceLabel+".id",
						strconv.Quote(campaignStatus),
						messagesPerMin,
						alwaysRunning,
						util.NullValue,
						[]string{},
						[]string{},
						[]string{},
						generateOutboundMessagingCampaignEmailConfig(
							[]string{strconv.Quote(emailColumns)},
							responseManagementResponse.ResourceType+"."+responseManagementResponseResourceLabel+".id",
							strconv.Quote(fromAddressDomainId),
							strconv.Quote(fromAddressLocalPart),
							routingEmailDomain.ResourceType+"."+routingEmailDomainResourceLabel+".id",
							routingEmailRoute.ResourceType+"."+routingEmailRouteResourceLabel+".id",
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "messages_per_minute", messagesPerMin),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "always_running", alwaysRunning),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "campaign_status", campaignStatus),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_config.0.email_columns.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_config.0.email_columns.0", emailColumns),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_config.0.from_address.0.domain_id", fromAddressDomainId),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_config.0.from_address.0.local_part", fromAddressLocalPart),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_config.0.reply_to_address.0.domain_id", "terraformtest.mypurecloud.com"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "email_config.0.reply_to_address.0.route_id",
						routingEmailRoute.ResourceType+"."+routingEmailRouteResourceLabel, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "email_config.0.content_template_id",
						responseManagementResponse.ResourceType+"."+responseManagementResponseResourceLabel, "id"),
				),
			},
			{
				// Update
				Config: contactListResource +
					routingQueueResource +
					routingEmailDomainResource +
					routingEmailRouteResource +
					responseManagementLibraryResource +
					responseManagementResponseResource2 +
					generateOutboundMessagingCampaignResource(
						resourceLabel,
						name,
						obContactList.ResourceType+"."+contactListResourceLabel+".id",
						strconv.Quote(campaignStatus),
						messagesPerMin,
						alwaysRunning,
						util.NullValue,
						[]string{},
						[]string{},
						[]string{},
						generateOutboundMessagingCampaignEmailConfig(
							[]string{strconv.Quote(emailColumns2)},
							responseManagementResponse.ResourceType+"."+responseManagementResponseResourceLabel+".id",
							strconv.Quote(fromAddressDomainId),
							strconv.Quote(fromAddressLocalPart2),
							routingEmailDomain.ResourceType+"."+routingEmailDomainResourceLabel+".id",
							routingEmailRoute.ResourceType+"."+routingEmailRouteResourceLabel+".id",
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "messages_per_minute", messagesPerMin),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "always_running", alwaysRunning),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "campaign_status", campaignStatus),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_config.0.email_columns.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_config.0.email_columns.0", emailColumns2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_config.0.from_address.0.domain_id", fromAddressDomainId),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_config.0.from_address.0.local_part", fromAddressLocalPart2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_config.0.reply_to_address.0.domain_id", "terraformtest.mypurecloud.com"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "email_config.0.reply_to_address.0.route_id",
						routingEmailRoute.ResourceType+"."+routingEmailRouteResourceLabel, "id"),
					resource.TestCheckResourceAttrPair(ResourceType+"."+resourceLabel, "email_config.0.content_template_id",
						responseManagementResponse.ResourceType+"."+responseManagementResponseResourceLabel, "id"),
				),
			},
			{
				// Import
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

func generateOutboundMessagingCampaignEmailConfig(
	emailColumns []string,
	contentTemplateId string,
	fromAddressDomainId string,
	fromAddressLocalPart string,
	replyToAddressDomainId string,
	replyToAddressRouteId string,
) string {
	return fmt.Sprintf(`
   	email_config {
		email_columns = [%s]
		content_template_id = %s
		from_address {
			domain_id = %s
			local_part = %s
		}
		reply_to_address {
			domain_id = %s
			route_id = %s
		}
	}
`, strings.Join(emailColumns, ", "), contentTemplateId, fromAddressDomainId, fromAddressLocalPart, replyToAddressDomainId, replyToAddressRouteId)
}

func generateDynamicContactQueueingSettings(sort string, filter string) string {
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

func CheckOutboundDomainExists(id string) error {
	config, _ := provider.AuthorizeSdk()
	routingApi := platformclientv2.NewRoutingApiWithConfig(config)

	log.Printf("Checking if outbound domain (%s) exists", id)
	outboundDomain, resp, err := routingApi.GetRoutingEmailOutboundDomain(id)
	if err != nil && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error getting outbound domain (%s): %v", id, err)
	}

	if isDomainVerified(outboundDomain) {
		log.Printf("Outbound domain (%s) exists and is verified", id)
		return nil
	}

	if resp.StatusCode == 404 {
		// outbound domain for test does not exist so create it
		log.Printf("Outbound domain (%s) does not exist. Creating...", id)
		_, _, postErr := routingApi.PostRoutingEmailOutboundDomains(platformclientv2.Outbounddomainrequest{
			Id: &id,
		})
		if postErr != nil {
			return fmt.Errorf("failed to create outbound domain: %v", err)
		}

		log.Printf("Outbound domain (%s) created", id)
		time.Sleep(3 * time.Second)
	}

	// ensure domain is verified
	if !isDomainVerified(outboundDomain) {
		log.Printf("Verifying outbound domain (%s)...", id)

		result, _, putErr := routingApi.PutRoutingEmailOutboundDomainActivation(id)
		if putErr != nil {
			return fmt.Errorf("failed to verify outbound domain (%s): %v", id, err)
		}

		if checkDomainVerification(result) {
			log.Printf("Outbound domain (%s) verified", id)
		}
	}
	return nil
}

func isDomainVerified(domain *platformclientv2.Outbounddomain) bool {
	if domain == nil ||
		domain.DkimVerificationResult == nil ||
		domain.CnameVerificationResult == nil ||
		domain.DkimVerificationResult.Status == nil ||
		domain.CnameVerificationResult.Status == nil {
		return false
	}

	return *domain.DkimVerificationResult.Status == "VERIFIED" &&
		*domain.CnameVerificationResult.Status == "VERIFIED"
}

func checkDomainVerification(domain *platformclientv2.Emailoutbounddomainresult) bool {
	if domain == nil ||
		domain.DnsTxtSendingRecord == nil ||
		domain.DnsCnameBounceRecord == nil ||
		domain.DnsTxtSendingRecord.VerificationStatus == nil ||
		domain.DnsCnameBounceRecord.VerificationStatus == nil {
		return false
	}

	return *domain.DnsTxtSendingRecord.VerificationStatus == "Verified" &&
		*domain.DnsCnameBounceRecord.VerificationStatus == "Verified"
}
