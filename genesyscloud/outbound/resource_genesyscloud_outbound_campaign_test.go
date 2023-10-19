package outbound

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

// Add a special generator DEVENGAGE-1646.  Basically, the API makes it look like you need a full phone_columns field here.  However, the API ignores the type because the devs reused the phone_columns object.  However,
// we still need to pass in a phone column block to get the column name.
func generatePhoneColumnNoTypeBlock(columnName string) string {

	return fmt.Sprintf(`
	phone_columns {
		column_name = "%s"
	}
`, columnName)
}

func TestAccResourceOutboundCampaignBasic(t *testing.T) {
	t.Parallel()

	var (
		resourceId            = "campaign1"
		name                  = "Test Campaign " + uuid.NewString()
		dialingMode           = "agentless"
		callerName            = "Test Name"
		callerAddress         = "+353371111111"
		contactListResourceId = "contact_list"
		dncListResourceId     = "dnc"
		queueResourceId       = "queue"
		wrapupCodeResourceId  = "wrapupcode"
		locationResourceId    = "location"
		clfResourceId         = "clf"
		carResourceId         = "car"
		ruleSetResourceId     = "rule_set"
		siteId                = "site"
		callableTimeSetId     = "time_set"
		outboundFlowFilePath  = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		flowName              = "test flow " + uuid.NewString()

		contactSortFieldName = "zipcode"
		contactSortDirection = "ASC"
		contactSortNumeric   = falseValue

		nameUpdated          = "Test Campaign " + uuid.NewString()
		callerNameUpdated    = "Test Name 2"
		callerAddressUpdated = "+353371112111"
	)

	// necessary to avoid errors during site creation
	_, err := gcloud.AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	emergencyNumber := "+13178793428"
	err = edgeSite.DeleteLocationWithNumber(emergencyNumber)
	if err != nil {
		t.Fatal(err)
	}

	referencedResources := fmt.Sprintf(`
		data "genesyscloud_auth_division_home" "home" {}
		`) + obContactList.GenerateOutboundContactList(
		contactListResourceId,
		"contact list "+uuid.NewString(),
		nullValue,
		strconv.Quote("Cell"),
		[]string{strconv.Quote("Cell")},
		[]string{strconv.Quote("Cell"), strconv.Quote("Home"), strconv.Quote("zipcode")},
		falseValue,
		nullValue,
		nullValue,
		obContactList.GeneratePhoneColumnsBlock(
			"Cell",
			"cell",
			strconv.Quote("Cell"),
		),
		obContactList.GeneratePhoneColumnsBlock(
			"Home",
			"home",
			strconv.Quote("Home"),
		),
		obContactList.GeneratePhoneColumnsDataTypeSpecBlock(
			strconv.Quote("zipcode"), // columnName
			strconv.Quote("TEXT"),    // columnDataType
			nullValue,                // min
			nullValue,                // max
			"10",                     // maxLength
		),
	) + generateOutboundDncListBasic(
		dncListResourceId,
		"dnc list "+uuid.NewString(),
	) + gcloud.GenerateRoutingQueueResourceBasic(
		queueResourceId,
		"tf queue"+uuid.NewString(),
		"",
	) + gcloud.GenerateRoutingWrapupcodeResource(
		wrapupCodeResourceId,
		"tf wrapup code"+uuid.NewString(),
	) + gcloud.GenerateFlowResource(
		"flow",
		outboundFlowFilePath,
		"",
		false,
		gcloud.GenerateSubstitutionsMap(map[string]string{
			"flow_name":          flowName,
			"home_division_name": "${data.genesyscloud_auth_division_home.home.name}",
			"contact_list_name":  "${genesyscloud_outbound_contact_list." + contactListResourceId + ".name}",
			"wrapup_code_name":   "${genesyscloud_routing_wrapupcode." + wrapupCodeResourceId + ".name}",
		}),
	) + generateOutboundCallAnalysisResponseSetResource(
		carResourceId,
		"tf car "+uuid.NewString(),
		falseValue,
		generateCarsResponsesBlock(
			generateCarsResponse(
				"callable_person",
				"transfer_flow",
				flowName,
				"${genesyscloud_flow.flow.id}",
			),
		),
	) + generateOutboundContactListFilter(
		clfResourceId,
		"tf clf "+uuid.NewString(),
		"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
		"",
		generateOutboundContactListFilterClause(
			"",
			generateOutboundContactListFilterPredicates(
				"Cell",
				"alphabetic",
				"EQUALS",
				"+12345123456",
				"",
				"",
			),
		),
	) + gcloud.GenerateLocationResource(
		locationResourceId,
		"tf location "+uuid.NewString(),
		"HQ1",
		[]string{},
		gcloud.GenerateLocationEmergencyNum(
			"+13178793428",
			nullValue,
		),
		gcloud.GenerateLocationAddress(
			"7601 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46278",
		),
	) + edgeSite.GenerateSiteResourceWithCustomAttrs(
		siteId,
		"tf site "+uuid.NewString(),
		"test description",
		"genesyscloud_location."+locationResourceId+".id",
		"Cloud",
		false,
		"[\"us-east-1\"]",
		nullValue,
		nullValue,
	) + fmt.Sprintf(`
		resource "genesyscloud_outbound_ruleset" "%s" {
			name            = "%s"
			contact_list_id = genesyscloud_outbound_contact_list.%s.id
		}
		`, ruleSetResourceId, "tf ruleset "+uuid.NewString(), contactListResourceId,
	) + generateOutboundCallabletimeset(
		callableTimeSetId,
		"tf timeset "+uuid.NewString(),
		generateCallableTimesBlock(
			"Africa/Abidjan",
			generateTimeSlotsBlock("07:00:00", "18:00:00", "3"),
		))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: referencedResources + generateOutboundCampaign(
					resourceId,
					name,
					dialingMode,
					callerName,
					callerAddress,
					"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
					nullValue, // campaign_status
					nullValue, // division id
					nullValue, // script id
					"genesyscloud_routing_queue."+queueResourceId+".id",
					"genesyscloud_telephony_providers_edges_site."+siteId+".id",
					"1",
					"genesyscloud_outbound_callabletimeset."+callableTimeSetId+".id",
					"genesyscloud_outbound_callanalysisresponseset."+carResourceId+".id",
					"1",
					nullValue,
					"0",
					falseValue,
					"40",
					"4",
					[]string{"genesyscloud_outbound_dnclist." + dncListResourceId + ".id"},
					[]string{"genesyscloud_outbound_ruleset." + ruleSetResourceId + ".id"},
					[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceId + ".id"},
					generatePhoneColumnNoTypeBlock("Cell"),
					generateOutboundMessagingCampaignContactSort(
						contactSortFieldName,
						contactSortDirection,
						contactSortNumeric,
					),
					generateDynamicContactQueueingSettingsBlock(trueValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "dialing_mode", dialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "caller_name", callerName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "caller_address", callerAddress),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "outbound_line_count", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "campaign_status", "off"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "abandon_rate", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "preview_time_out_seconds", "0"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "always_running", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "no_answer_timeout", "40"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "priority", "4"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.field_name", contactSortFieldName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.direction", contactSortDirection),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.numeric", contactSortNumeric),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "dynamic_contact_queueing_settings.0.sort", trueValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "callable_time_set_id",
						"genesyscloud_outbound_callabletimeset."+callableTimeSetId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "queue_id",
						"genesyscloud_routing_queue."+queueResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "dnc_list_ids.0",
						"genesyscloud_outbound_dnclist."+dncListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "rule_set_ids.0",
						"genesyscloud_outbound_ruleset."+ruleSetResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "site_id",
						"genesyscloud_telephony_providers_edges_site."+siteId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceId, "id"),
					gcloud.TestDefaultHomeDivision("genesyscloud_outbound_campaign."+resourceId),
				),
			},
			{
				// Update DCQSettings (ForceNew)
				Config: referencedResources + generateOutboundCampaign(
					resourceId,
					name,
					dialingMode,
					callerName,
					callerAddress,
					"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
					nullValue, // campaign_status
					nullValue, // division id
					nullValue, // script id
					"genesyscloud_routing_queue."+queueResourceId+".id",
					"genesyscloud_telephony_providers_edges_site."+siteId+".id",
					"1",
					"genesyscloud_outbound_callabletimeset."+callableTimeSetId+".id",
					"genesyscloud_outbound_callanalysisresponseset."+carResourceId+".id",
					"1",
					nullValue,
					"0",
					falseValue,
					"40",
					"4",
					[]string{"genesyscloud_outbound_dnclist." + dncListResourceId + ".id"},
					[]string{"genesyscloud_outbound_ruleset." + ruleSetResourceId + ".id"},
					[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceId + ".id"},
					generatePhoneColumnNoTypeBlock("Cell"),
					generateOutboundMessagingCampaignContactSort(
						contactSortFieldName,
						contactSortDirection,
						contactSortNumeric,
					),
					generateDynamicContactQueueingSettingsBlock(falseValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "dialing_mode", dialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "caller_name", callerName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "caller_address", callerAddress),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "outbound_line_count", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "campaign_status", "off"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "abandon_rate", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "preview_time_out_seconds", "0"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "always_running", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "no_answer_timeout", "40"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "priority", "4"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.field_name", contactSortFieldName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.direction", contactSortDirection),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.numeric", contactSortNumeric),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "dynamic_contact_queueing_settings.0.sort", falseValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "callable_time_set_id",
						"genesyscloud_outbound_callabletimeset."+callableTimeSetId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "queue_id",
						"genesyscloud_routing_queue."+queueResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "dnc_list_ids.0",
						"genesyscloud_outbound_dnclist."+dncListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "rule_set_ids.0",
						"genesyscloud_outbound_ruleset."+ruleSetResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "site_id",
						"genesyscloud_telephony_providers_edges_site."+siteId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceId, "id"),
					gcloud.TestDefaultHomeDivision("genesyscloud_outbound_campaign."+resourceId),
				),
			},
			{
				// Update
				Config: referencedResources + generateOutboundCampaign(
					resourceId,
					nameUpdated,
					dialingMode,
					callerNameUpdated,
					callerAddressUpdated,
					"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
					nullValue, // campaign_status
					nullValue, // division id
					nullValue, // script id
					"genesyscloud_routing_queue."+queueResourceId+".id",
					"genesyscloud_telephony_providers_edges_site."+siteId+".id",
					"2",
					"genesyscloud_outbound_callabletimeset."+callableTimeSetId+".id",
					"genesyscloud_outbound_callanalysisresponseset."+carResourceId+".id",
					"2",
					nullValue,
					"1",
					trueValue,
					"30",
					"3",
					[]string{"genesyscloud_outbound_dnclist." + dncListResourceId + ".id"},
					[]string{"genesyscloud_outbound_ruleset." + ruleSetResourceId + ".id"},
					[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceId + ".id"},
					generatePhoneColumnNoTypeBlock("Cell"),
					generateOutboundMessagingCampaignContactSort(
						contactSortFieldName,
						contactSortDirection,
						contactSortNumeric,
					),
					generateDynamicContactQueueingSettingsBlock(falseValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "name", nameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "dialing_mode", dialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "caller_name", callerNameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "caller_address", callerAddressUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "outbound_line_count", "2"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "abandon_rate", "2"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "preview_time_out_seconds", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "always_running", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "no_answer_timeout", "30"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "priority", "3"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.field_name", contactSortFieldName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.direction", contactSortDirection),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.numeric", contactSortNumeric),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "dynamic_contact_queueing_settings.0.sort", "false"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "callable_time_set_id",
						"genesyscloud_outbound_callabletimeset."+callableTimeSetId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "queue_id",
						"genesyscloud_routing_queue."+queueResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "dnc_list_ids.0",
						"genesyscloud_outbound_dnclist."+dncListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "rule_set_ids.0",
						"genesyscloud_outbound_ruleset."+ruleSetResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "site_id",
						"genesyscloud_telephony_providers_edges_site."+siteId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceId, "id"),
					gcloud.TestDefaultHomeDivision("genesyscloud_outbound_campaign."+resourceId),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_outbound_campaign." + resourceId,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"queue_id"},
			},
		},
		CheckDestroy: testVerifyOutboundCampaignDestroyed,
	})
}

func TestAccResourceOutboundCampaignCampaignStatus(t *testing.T) {
	t.Parallel()
	var (
		resourceId            = "campaign2"
		name                  = "Test Campaign " + uuid.NewString()
		contactListResourceId = "contact_list"
		carResourceId         = "car"
		siteId                = "site"
		wrapupCodeResourceId  = "wrapupcode"
		outboundFlowFilePath  = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		flowName              = "test flow " + uuid.NewString()
		flowResourceId        = "flow"
		wrapupcodeResourceId  = "wrapupcode"
		locationResourceId    = "location"
	)

	// necessary to avoid errors during site creation
	_, err := gcloud.AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	emergencyNumber := "+13178793429"
	err = edgeSite.DeleteLocationWithNumber(emergencyNumber)
	if err != nil {
		t.Fatal(err)
	}

	referencedResources := obContactList.GenerateOutboundContactList(
		contactListResourceId,
		"contact list "+uuid.NewString(),
		nullValue,
		strconv.Quote("Cell"),
		[]string{strconv.Quote("Cell")},
		[]string{strconv.Quote("Cell"), strconv.Quote("Home"), strconv.Quote("zipcode")},
		falseValue,
		nullValue,
		nullValue,
		obContactList.GeneratePhoneColumnsBlock(
			"Cell",
			"cell",
			strconv.Quote("Cell"),
		),
		obContactList.GeneratePhoneColumnsBlock(
			"Home",
			"home",
			strconv.Quote("Home"),
		),
	) + gcloud.GenerateRoutingWrapupcodeResource(
		wrapupcodeResourceId,
		"tf wrapup code"+uuid.NewString(),
	) + gcloud.GenerateFlowResource(
		flowResourceId,
		outboundFlowFilePath,
		"",
		false,
		gcloud.GenerateSubstitutionsMap(map[string]string{
			"flow_name":          flowName,
			"home_division_name": "${data.genesyscloud_auth_division_home.home.name}",
			"contact_list_name":  "${genesyscloud_outbound_contact_list." + contactListResourceId + ".name}",
			"wrapup_code_name":   "${genesyscloud_routing_wrapupcode." + wrapupCodeResourceId + ".name}",
		}),
	) + generateOutboundCallAnalysisResponseSetResource(
		carResourceId,
		"tf car "+uuid.NewString(),
		falseValue,
		generateCarsResponsesBlock(
			generateCarsResponse(
				"callable_person",
				"transfer_flow",
				flowName,
				"${genesyscloud_flow.flow.id}",
			),
		),
	) + gcloud.GenerateLocationResource(
		locationResourceId,
		"tf location "+uuid.NewString(),
		"HQ1",
		[]string{},
		gcloud.GenerateLocationEmergencyNum(
			"+13178793429",
			nullValue,
		),
		gcloud.GenerateLocationAddress(
			"7601 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46278",
		),
	) + edgeSite.GenerateSiteResourceWithCustomAttrs(
		siteId,
		"tf site "+uuid.NewString(),
		"test description",
		"genesyscloud_location."+locationResourceId+".id",
		"Cloud",
		false,
		"[\"us-east-1\"]",
		nullValue,
		nullValue,
	) + fmt.Sprintf("\ndata \"genesyscloud_auth_division_home\" \"home\" {}\n")

	// Test campaign_status can be turned on in a second run after first run's initial creation in off state, and then back off again
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: referencedResources + fmt.Sprintf(`
					resource "genesyscloud_outbound_campaign" "%s" {
						name                          = "%s"
						dialing_mode                  = "agentless"
						caller_name                   = "Test Name"
						caller_address                = "+353371111111"
						outbound_line_count           = 2
						campaign_status               = "off"
						contact_list_id               = genesyscloud_outbound_contact_list.%s.id
						site_id                       = genesyscloud_telephony_providers_edges_site.%s.id
						call_analysis_response_set_id = genesyscloud_outbound_callanalysisresponseset.%s.id
						phone_columns {
							column_name = "Cell"
						}
					}
					`, resourceId, name, contactListResourceId, siteId, carResourceId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "campaign_status", "off"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "site_id",
						"genesyscloud_telephony_providers_edges_site."+siteId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceId, "id"),
					// Add contacts to the contact list (because we have access to the state and can pull out the contactlist ID to pass to the API)
					addContactsToContactList,
				),
			},
			{
				Config: referencedResources + fmt.Sprintf(`
					resource "genesyscloud_outbound_campaign" "%s" {
						name                          = "%s"
						dialing_mode                  = "agentless"
						caller_name                   = "Test Name"
						caller_address                = "+353371111111"
						outbound_line_count           = 2
						campaign_status               = "on"
						contact_list_id               = genesyscloud_outbound_contact_list.%s.id
						site_id                       = genesyscloud_telephony_providers_edges_site.%s.id
						call_analysis_response_set_id = genesyscloud_outbound_callanalysisresponseset.%s.id
						phone_columns {
							column_name = "Cell"
						}
					}
					`, resourceId, name, contactListResourceId, siteId, carResourceId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "name", name),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "site_id",
						"genesyscloud_telephony_providers_edges_site."+siteId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceId, "id"),
					gcloud.VerifyAttributeInArrayOfPotentialValues("genesyscloud_outbound_campaign."+resourceId, "campaign_status", []string{"on", "complete"}),
				),
			},
			{
				Config: referencedResources + fmt.Sprintf(`
				resource "genesyscloud_outbound_campaign" "%s" {
					name                          = "%s"
					dialing_mode                  = "agentless"
					caller_name                   = "Test Name"
					caller_address                = "+353371111111"
					outbound_line_count           = 2
					campaign_status               = "off"
					contact_list_id               = genesyscloud_outbound_contact_list.%s.id
					site_id                       = genesyscloud_telephony_providers_edges_site.%s.id
					call_analysis_response_set_id = genesyscloud_outbound_callanalysisresponseset.%s.id
					phone_columns {
						column_name = "Cell"
					}
				}
				`, resourceId, name, contactListResourceId, siteId, carResourceId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "name", name),
					gcloud.VerifyAttributeInArrayOfPotentialValues("genesyscloud_outbound_campaign."+resourceId, "campaign_status", []string{"off", "complete"}),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "site_id",
						"genesyscloud_telephony_providers_edges_site."+siteId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceId, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_outbound_campaign." + resourceId,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"campaign_status"},
			},
		},
		CheckDestroy: testVerifyOutboundCampaignDestroyed,
	})
}

func TestAccResourceOutboundCampaignStatusOn(t *testing.T) {
	t.Parallel()
	var (
		resourceId            = "campaign3"
		name                  = "Test Campaign " + uuid.NewString()
		contactListResourceId = "contact_list"
		carResourceId         = "car"
		siteId                = "site"
		outboundFlowFilePath  = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		flowName              = "test flow " + uuid.NewString()
		flowResourceId        = "flow"
		wrapupcodeResourceId  = "wrapupcode"
		locationResourceId    = "location"
	)

	// necessary to avoid errors during site creation
	_, err := gcloud.AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	emergencyNumber := "+13178793430"
	err = edgeSite.DeleteLocationWithNumber(emergencyNumber)
	if err != nil {
		t.Fatal(err)
	}

	// Test campaign_status can be turned on at time of creation as well
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			// Create resources for outbound campaign
			{
				Config: fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) + generateReferencedResourcesForOutboundCampaignTests(
					contactListResourceId,
					"",
					"",
					carResourceId,
					outboundFlowFilePath,
					flowResourceId,
					flowName,
					"",
					siteId,
					emergencyNumber,
					"",
					"",
					"${data.genesyscloud_auth_division_home.home.name}",
					locationResourceId,
					wrapupcodeResourceId,
				),
				// Add contacts to the contact list (because we have access to the state and can pull out the contactlist ID to pass to the API)
				Check: addContactsToContactList,
			},
			// Now, we create the outbound campaign and it should stay running because it has contacts to call. We leave it running to test
			// the destroy command takes care of turning it off before deleting.
			{
				Config: fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) + generateOutboundCampaignBasic(
					resourceId,
					name,
					contactListResourceId,
					siteId,
					emergencyNumber,
					carResourceId,
					strconv.Quote("on"),
					outboundFlowFilePath,
					flowResourceId,
					flowName,
					"${data.genesyscloud_auth_division_home.home.name}",
					locationResourceId,
					wrapupcodeResourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "name", name),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "site_id",
						"genesyscloud_telephony_providers_edges_site."+siteId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceId, "id"),
					gcloud.VerifyAttributeInArrayOfPotentialValues("genesyscloud_outbound_campaign."+resourceId, "campaign_status", []string{"on", "complete"}),
				),
			},
			// Don't turn campaign back off to ensure campaign can be destroyed properly by turning it off within the destroy handler
			{
				// Import/Read
				ResourceName:            "genesyscloud_outbound_campaign." + resourceId,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"campaign_status"},
			},
		},
		CheckDestroy: testVerifyOutboundCampaignDestroyed,
	})
}

func TestAccResourceOutboundCampaignWithScriptId(t *testing.T) {
	t.Parallel()
	var (
		resourceId                = "campaign4"
		name                      = "Test Campaign " + uuid.NewString()
		dialingMode               = "preview"
		callerName                = "Test Name 123"
		callerAddress             = "+353371111111"
		contactListResourceId     = "contact_list"
		queueResourceId           = "queue"
		dncListResourceId         = "dnc_list"
		carResourceId             = "car"
		clfResourceId             = "clf"
		ruleSetResourceId         = "rule_set"
		callableTimeSetResourceId = "time_set"

		contactSortFieldName = "zipcode"
		contactSortDirection = "ASC"
		contactSortNumeric   = falseValue
	)

	scriptId, err := getPublishedScriptId()
	if err != nil || scriptId == "" {
		t.Skip("Skipping as a published script ID is needed to run this test")
	}

	referencedResources := generateReferencedResourcesForOutboundCampaignTests(
		contactListResourceId,
		"",
		queueResourceId,
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
	)

	referencedResourcesUpdate := generateReferencedResourcesForOutboundCampaignTests(
		contactListResourceId,
		dncListResourceId,
		queueResourceId,
		carResourceId,
		"",
		"",
		"",
		clfResourceId,
		"",
		"",
		ruleSetResourceId,
		callableTimeSetResourceId,
		"",
		"",
		"",
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: referencedResources +
					generateOutboundCampaign(
						resourceId,
						name,
						dialingMode,
						callerName,
						callerAddress,
						"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
						nullValue,
						nullValue,
						strconv.Quote(scriptId),
						"genesyscloud_routing_queue."+queueResourceId+".id",
						nullValue,
						nullValue,
						nullValue,
						nullValue,
						nullValue,
						nullValue,
						nullValue,
						nullValue,
						nullValue,
						nullValue,
						[]string{},
						[]string{},
						[]string{},
						generatePhoneColumnNoTypeBlock("Cell"),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "dialing_mode", dialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "caller_name", callerName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "caller_address", callerAddress),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "script_id", scriptId),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "queue_id",
						"genesyscloud_routing_queue."+queueResourceId, "id"),
				),
			},
			{
				// Update
				Config: referencedResourcesUpdate +
					generateOutboundCampaign(
						resourceId,
						name,
						dialingMode,
						callerName,
						callerAddress,
						"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
						nullValue,
						nullValue,
						strconv.Quote(scriptId),
						"genesyscloud_routing_queue."+queueResourceId+".id",
						nullValue,
						"1",
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceId+".id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceId+".id",
						"1",
						falseValue,
						"1",
						falseValue,
						"3",
						"2",
						[]string{"genesyscloud_outbound_dnclist." + dncListResourceId + ".id"},
						[]string{"genesyscloud_outbound_ruleset." + ruleSetResourceId + ".id"},
						[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceId + ".id"},
						generatePhoneColumnNoTypeBlock("Cell"),
						generateOutboundMessagingCampaignContactSort(
							contactSortFieldName,
							contactSortDirection,
							contactSortNumeric,
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "dialing_mode", dialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "caller_name", callerName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "caller_address", callerAddress),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "abandon_rate", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "outbound_line_count", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "skip_preview_disabled", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "preview_time_out_seconds", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "always_running", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.field_name", contactSortFieldName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.direction", contactSortDirection),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.numeric", contactSortNumeric),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "no_answer_timeout", "3"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "priority", "2"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "script_id", scriptId),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "dnc_list_ids.0",
						"genesyscloud_outbound_dnclist."+dncListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "queue_id",
						"genesyscloud_routing_queue."+queueResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "rule_set_ids.0",
						"genesyscloud_outbound_ruleset."+ruleSetResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "callable_time_set_id",
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceId, "id"),
					gcloud.TestDefaultHomeDivision("genesyscloud_outbound_campaign."+resourceId),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_campaign." + resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyOutboundCampaignDestroyed,
	})
}

func addContactsToContactList(state *terraform.State) error {
	outboundAPI := platformclientv2.NewOutboundApi()
	resource := state.RootModule().Resources["genesyscloud_outbound_contact_list.contact_list"]
	if resource == nil {
		return fmt.Errorf("genesyscloud_outbound_contact_list.contact_list resource not found in state")
	}

	contactList, _, err := outboundAPI.GetOutboundContactlist(resource.Primary.ID, false, false)
	if err != nil {
		return fmt.Errorf("genesyscloud_outbound_contact_list (%s) not available", resource.Primary.ID)
	}
	contactsJSON := `[{
			"data": {
			  "FirstName": "Asa",
			  "LastName": "Acosta",
			  "Cell": "3335551234",
			  "Home": "3335552345",
			  "zipcode": "23849"
			},
			"callable": true,
			"phoneNumberStatus": {}
		  },
		  {
			"data": {
			  "FirstName": "Leonidas",
			  "LastName": "Acosta",
			  "Cell": "4445551234",
			  "Home": "4445552345",
			  "zipcode": "34567"
			},
			"callable": true,
			"phoneNumberStatus": {}
		  },
		  {
			"data": {
			  "FirstName": "Nolan",
			  "LastName": "Adams",
			  "Cell": "6665551234",
			  "Home": "6665552345",
			  "zipcode": "56789"
			},
			"callable": true,
			"phoneNumberStatus": {}
		  }]`
	var contacts []platformclientv2.Writabledialercontact
	err = json.Unmarshal([]byte(contactsJSON), &contacts)
	if err != nil {
		return fmt.Errorf("could not unmarshall JSON contacts to add to contact list")
	}
	_, _, err = outboundAPI.PostOutboundContactlistContacts(*contactList.Id, contacts, false, false, false)
	if err != nil {
		return fmt.Errorf("could not post contacts to contact list")
	}
	return nil
}

func generateOutboundCampaignBasic(resourceId string,
	name string,
	contactListResourceId string,
	siteResourceId string,
	emergencyNumber string,
	carResourceId string,
	campaignStatus string,
	outboundFlowFilePath string,
	flowResourceId string,
	flowName string,
	divisionName,
	locationResourceId string,
	wrapupcodeResourceId string) string {
	referencedResources := generateReferencedResourcesForOutboundCampaignTests(
		contactListResourceId,
		"",
		"",
		carResourceId,
		outboundFlowFilePath,
		flowResourceId,
		flowName,
		"",
		siteResourceId,
		emergencyNumber,
		"",
		"",
		divisionName,
		locationResourceId,
		wrapupcodeResourceId,
	)
	return fmt.Sprintf(`
resource "genesyscloud_outbound_campaign" "%s" {
	name                          = "%s"
	dialing_mode                  = "agentless"
	caller_name                   = "Test Name"
	caller_address                = "+353371111111"
	outbound_line_count           = 2
	campaign_status               = %s
	contact_list_id 			  = genesyscloud_outbound_contact_list.%s.id
	site_id 				      = genesyscloud_telephony_providers_edges_site.%s.id
	call_analysis_response_set_id = genesyscloud_outbound_callanalysisresponseset.%s.id
	phone_columns {
		column_name = "Cell"
	}
}
%s
`, resourceId, name, campaignStatus, contactListResourceId, siteResourceId, carResourceId, referencedResources)
}

func generateOutboundCampaign(
	resourceId string,
	name string,
	dialingMode string, // required
	callerName string, // required
	callerAddress string, // required
	contactListId string, // required
	campaignStatus string,
	divisionId string,
	scriptId string,
	queueId string,
	siteId string,
	abandonRate string,
	callableTimeSetId string,
	callAnalysisResponseSetId string,
	outboundLineCount string,
	skipPreviewDisabled string,
	previewTimeOutSeconds string,
	alwaysRunning string,
	noAnswerTimeout string,
	priority string,
	dncListIds []string,
	ruleSetIds []string,
	contactListFilterIds []string,
	nestedBlocks ...string,
) string {
	return fmt.Sprintf(`
resource "genesyscloud_outbound_campaign" "%s" {
	name                          = "%s"
	dialing_mode                  = "%s"
	caller_name                   = "%s"
	caller_address                = "%s"
	contact_list_id               = %s
	campaign_status               = %s
	division_id                   = %s
	script_id                     = %s
	queue_id                      = %s
	site_id 					  = %s
	abandon_rate                  = %s
	callable_time_set_id          = %s
	call_analysis_response_set_id = %s
	outbound_line_count           = %s
	skip_preview_disabled         = %s
	preview_time_out_seconds      = %s
	always_running                = %s
	no_answer_timeout             = %s
	priority                      = %s
	dnc_list_ids                  = [%s]
	rule_set_ids 			      = [%s]
	contact_list_filter_ids       = [%s]
	%s
}
`, resourceId, name, dialingMode, callerName, callerAddress, contactListId, campaignStatus, divisionId, scriptId, queueId, siteId, abandonRate, callableTimeSetId,
		callAnalysisResponseSetId, outboundLineCount, skipPreviewDisabled, previewTimeOutSeconds, alwaysRunning, noAnswerTimeout,
		priority, strings.Join(dncListIds, ", "), strings.Join(ruleSetIds, ", "), strings.Join(contactListFilterIds, ", "),
		strings.Join(nestedBlocks, "\n"))
}

func generateDynamicContactQueueingSettingsBlock(sort string) string {
	return fmt.Sprintf(`
	dynamic_contact_queueing_settings {
		sort = %s
	}
	`, sort)
}

func generateReferencedResourcesForOutboundCampaignTests(
	contactListResourceId string,
	dncListResourceId string,
	queueResourceId string,
	carResourceId string,
	outboundFlowFilePath string,
	flowResourceId string,
	flowName string,
	clfResourceId string,
	siteId string,
	emergencyNumber string,
	ruleSetId string,
	callableTimeSetResourceId string,
	divisionName string,
	locationResourceId string,
	wrapUpCodeResourceId string,
) string {
	var (
		contactList             string
		dncList                 string
		queue                   string
		callAnalysisResponseSet string
		contactListFilter       string
		site                    string
		ruleSet                 string
		callableTimeSet         string
	)
	if contactListResourceId != "" {
		contactList = obContactList.GenerateOutboundContactList(
			contactListResourceId,
			"terraform contact list "+uuid.NewString(),
			nullValue,
			strconv.Quote("Cell"),
			[]string{strconv.Quote("Cell")},
			[]string{strconv.Quote("Cell"), strconv.Quote("Home"), strconv.Quote("zipcode")},
			falseValue,
			nullValue,
			nullValue,
			obContactList.GeneratePhoneColumnsBlock("Cell", "cell", strconv.Quote("Cell")),
			obContactList.GeneratePhoneColumnsBlock("Home", "home", strconv.Quote("Home")))
	}
	if dncListResourceId != "" {
		dncList = generateOutboundDncListBasic(dncListResourceId, "tf dnc list "+uuid.NewString())
	}
	if queueResourceId != "" {
		queue = gcloud.GenerateRoutingQueueResourceBasic(queueResourceId, "tf test queue "+uuid.NewString())
	}
	if carResourceId != "" {
		if outboundFlowFilePath != "" {
			callAnalysisResponseSet = gcloud.GenerateRoutingWrapupcodeResource(
				wrapUpCodeResourceId,
				"wrapupcode "+uuid.NewString(),
			) + gcloud.GenerateFlowResource(
				flowResourceId,
				outboundFlowFilePath,
				"",
				false,
				gcloud.GenerateSubstitutionsMap(map[string]string{
					"flow_name":          flowName,
					"home_division_name": divisionName,
					"contact_list_name":  "${genesyscloud_outbound_contact_list." + contactListResourceId + ".name}",
					"wrapup_code_name":   "${genesyscloud_routing_wrapupcode." + wrapUpCodeResourceId + ".name}",
				}),
			) + generateOutboundCallAnalysisResponseSetResource(
				carResourceId,
				"tf test car "+uuid.NewString(),
				falseValue,
				generateCarsResponsesBlock(
					generateCarsResponse(
						"callable_person",
						"transfer_flow",
						flowName,
						"${genesyscloud_flow."+flowResourceId+".id}",
					),
				))
		} else {
			callAnalysisResponseSet = generateOutboundCallAnalysisResponseSetResource(
				carResourceId,
				"tf test car "+uuid.NewString(),
				trueValue,
				generateCarsResponsesBlock(
					generateCarsResponse(
						"callable_machine",
						"transfer",
						"",
						"",
					),
				),
			)
		}
	}
	if clfResourceId != "" {
		contactListFilter = generateOutboundContactListFilter(
			clfResourceId,
			"tf test clf "+uuid.NewString(),
			"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
			"",
			generateOutboundContactListFilterClause(
				"",
				generateOutboundContactListFilterPredicates(
					"Cell",
					"alphabetic",
					"EQUALS",
					"+12345123456",
					"",
					"",
				),
			),
		)
	}
	if siteId != "" {
		siteName := "site " + uuid.NewString()
		locationName := "location " + uuid.NewString()
		site = fmt.Sprintf(`
resource "genesyscloud_location" "%s" {
	name  = "%s"
	notes = "HQ1"
	path  = []
	emergency_number {
		number = "%s"
		type   = null
	}
	address {
		street1  = "7601 Interactive Way"
		city     = "Indianapolis"
		state    = "IN"
		country  = "US"
		zip_code = "46278"
	}
}

resource "genesyscloud_telephony_providers_edges_site" "%s" {
	name                            = "%s"
	description                     = "TestAccResourceSite description 1"
	location_id                     = genesyscloud_location.%s.id
	media_model                     = "Cloud"
	media_regions_use_latency_based = false
}
`, locationResourceId, locationName, emergencyNumber, siteId, siteName, locationResourceId)
	}
	if ruleSetId != "" {
		ruleSetName := "ruleset " + uuid.NewString()
		ruleSet = fmt.Sprintf(`
resource "genesyscloud_outbound_ruleset" "%s" {
  name            = "%s"
  contact_list_id = genesyscloud_outbound_contact_list.%s.id
}
`, ruleSetId, ruleSetName, contactListResourceId)
	}
	if callableTimeSetResourceId != "" {
		callableTimeSetName := "test time set " + uuid.NewString()
		callableTimeSet = fmt.Sprintf(`
resource "genesyscloud_outbound_callabletimeset" "%s"{
	name = "%s"
	callable_times {
		time_zone_id = "Africa/Abidjan"
		time_slots {
			start_time = "07:00:00"
			stop_time = "18:00:00"
			day = 3
		}
	}
}
`, callableTimeSetResourceId, callableTimeSetName)
	}
	return fmt.Sprintf(`
%s
%s
%s
%s
%s
%s
%s
%s
`, contactList, dncList, queue, callAnalysisResponseSet, contactListFilter, site, ruleSet, callableTimeSet)
}

func getPublishedScriptId() (string, error) {
	config, err := gcloud.AuthorizeSdk()
	if err != nil {
		return "", fmt.Errorf("failed to authorize client: %v", err)
	}
	api := platformclientv2.NewScriptsApiWithConfig(config)
	// Get the published scripts.
	data, _, err := api.GetScriptsPublished(50, 1, "", "", "", "", "", "")
	if err != nil {
		return "", err
	}
	script := (*data.Entities)[0]
	return *script.Id, nil
}

func testVerifyOutboundCampaignDestroyed(state *terraform.State) error {
	outboundAPI := platformclientv2.NewOutboundApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_outbound_campaign" {
			continue
		}
		campaign, resp, err := outboundAPI.GetOutboundCampaign(rs.Primary.ID)
		if campaign != nil {
			return fmt.Errorf("campaign (%s) still exists", rs.Primary.ID)
		} else if gcloud.IsStatus404(resp) {
			// campaign not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All campaigns destroyed
	return nil
}
