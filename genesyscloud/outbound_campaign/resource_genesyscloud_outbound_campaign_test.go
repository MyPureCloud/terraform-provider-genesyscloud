package outbound_campaign

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/architect_flow"
	location "terraform-provider-genesyscloud/genesyscloud/location"
	"terraform-provider-genesyscloud/genesyscloud/outbound"
	obDnclist "terraform-provider-genesyscloud/genesyscloud/outbound_dnclist"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
	obCallableTimeset "terraform-provider-genesyscloud/genesyscloud/outbound_callabletimeset"
	obResponseSet "terraform-provider-genesyscloud/genesyscloud/outbound_callanalysisresponseset"
	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	obContactListFilter "terraform-provider-genesyscloud/genesyscloud/outbound_contactlistfilter"
	routingWrapupcode "terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
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
	var (
		resourceLabel            = "campaign1"
		name                     = "Test Campaign " + uuid.NewString()
		dialingMode              = "agentless"
		callerName               = "Test Name"
		callerAddress            = "+353371111111"
		contactListResourceLabel = "contact_list"
		dncListResourceLabel     = "dnc"
		wrapupCodeResourceLabel  = "wrapupcode"
		locationResourceLabel    = "location"
		clfResourceLabel         = "clf"
		carResourceLabel         = "car"
		ruleSetResourceLabel     = "rule_set"
		siteId                   = "site"
		callableTimeSetId        = "time_set"
		outboundFlowFilePath     = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		flowName                 = "test flow " + uuid.NewString()

		contactSortFieldName = "zipcode"
		contactSortDirection = "ASC"
		contactSortNumeric   = util.FalseValue

		nameUpdated          = "Test Campaign " + uuid.NewString()
		callerNameUpdated    = "Test Name 2"
		callerAddressUpdated = "+353371112111"
		divResourceLabel     = "test-division"
		divName              = "terraform-" + uuid.NewString()
	)

	emergencyNumber := "+13178793428"
	if err := edgeSite.DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s: %v", emergencyNumber, err)
	}

	referencedResources := `data "genesyscloud_auth_division_home" "home" {}` + obContactList.GenerateOutboundContactList(
		contactListResourceLabel,
		"contact list "+uuid.NewString(),
		util.NullValue,
		strconv.Quote("Cell"),
		[]string{strconv.Quote("Cell")},
		[]string{strconv.Quote("Cell"), strconv.Quote("Home"), strconv.Quote("zipcode")},
		util.FalseValue,
		util.NullValue,
		util.NullValue,
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
			util.NullValue,           // min
			util.NullValue,           // max
			"10",                     // maxLength
		),
	) + obDnclist.GenerateOutboundDncListBasic(
		dncListResourceLabel,
		"dnc list "+uuid.NewString(),
	) + authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + routingWrapupcode.GenerateRoutingWrapupcodeResource(
		wrapupCodeResourceLabel,
		"tf wrapup code"+uuid.NewString(),
		"genesyscloud_auth_division."+divResourceLabel+".id",
	) + architect_flow.GenerateFlowResource(
		"flow",
		outboundFlowFilePath,
		"",
		false,
		util.GenerateSubstitutionsMap(map[string]string{
			"flow_name":          flowName,
			"home_division_name": "${data.genesyscloud_auth_division_home.home.name}",
			"contact_list_name":  "${genesyscloud_outbound_contact_list." + contactListResourceLabel + ".name}",
			"wrapup_code_name":   "${genesyscloud_routing_wrapupcode." + wrapupCodeResourceLabel + ".name}",
		}),
	) + obResponseSet.GenerateOutboundCallAnalysisResponseSetResource(
		carResourceLabel,
		"tf car "+uuid.NewString(),
		util.FalseValue,
		obResponseSet.GenerateCarsResponsesBlock(
			obResponseSet.GenerateCarsResponse(
				"callable_person",
				"transfer_flow",
				flowName,
				"${genesyscloud_flow.flow.id}",
			),
		),
	) + obContactListFilter.GenerateOutboundContactListFilter(
		clfResourceLabel,
		"tf clf "+uuid.NewString(),
		"genesyscloud_outbound_contact_list."+contactListResourceLabel+".id",
		"",
		obContactListFilter.GenerateOutboundContactListFilterClause(
			"",
			obContactListFilter.GenerateOutboundContactListFilterPredicates(
				"Cell",
				"alphabetic",
				"EQUALS",
				"+12345123456",
				"",
				"",
			),
		),
	) + location.GenerateLocationResource(
		locationResourceLabel,
		"tf location "+uuid.NewString(),
		"HQ1",
		[]string{},
		location.GenerateLocationEmergencyNum(
			"+13178793428",
			util.NullValue,
		),
		location.GenerateLocationAddress(
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
		"genesyscloud_location."+locationResourceLabel+".id",
		"Cloud",
		false,
		"[\"us-east-1\"]",
		util.NullValue,
		util.NullValue,
	) + fmt.Sprintf(`
		resource "genesyscloud_outbound_ruleset" "%s" {
			name            = "%s"
			contact_list_id = genesyscloud_outbound_contact_list.%s.id
		}
		`, ruleSetResourceLabel, "tf ruleset "+uuid.NewString(), contactListResourceLabel,
	) + obCallableTimeset.GenerateOutboundCallabletimeset(
		callableTimeSetId,
		"tf timeset "+uuid.NewString(),
		obCallableTimeset.GenerateCallableTimesBlock(
			"Africa/Abidjan",
			obCallableTimeset.GenerateTimeSlotsBlock("07:00:00", "18:00:00", "3"),
		))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
				Config: referencedResources + generateOutboundCampaign(
					resourceLabel,
					name,
					dialingMode,
					callerName,
					callerAddress,
					"genesyscloud_outbound_contact_list."+contactListResourceLabel+".id",
					util.NullValue, // campaign_status
					util.NullValue, // division id
					util.NullValue, // script id
					util.NullValue, // queue id
					"genesyscloud_telephony_providers_edges_site."+siteId+".id",
					"1",
					util.NullValue,
					"genesyscloud_outbound_callabletimeset."+callableTimeSetId+".id",
					"genesyscloud_outbound_callanalysisresponseset."+carResourceLabel+".id",
					"1",
					util.NullValue,
					"0",
					util.FalseValue,
					"40",
					"4",
					[]string{"genesyscloud_outbound_dnclist." + dncListResourceLabel + ".id"},
					[]string{"genesyscloud_outbound_ruleset." + ruleSetResourceLabel + ".id"},
					[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceLabel + ".id"},
					[]string{""},
					strconv.Quote("false"),
					generatePhoneColumnNoTypeBlock("Cell"),
					outbound.GenerateOutboundMessagingCampaignContactSort(
						contactSortFieldName,
						contactSortDirection,
						contactSortNumeric,
					),
					generateDynamicContactQueueingSettingsBlock(util.TrueValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "dialing_mode", dialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "caller_name", callerName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "caller_address", callerAddress),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "outbound_line_count", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "campaign_status", "off"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "abandon_rate", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "preview_time_out_seconds", "0"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "always_running", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "no_answer_timeout", "40"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "priority", "4"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "contact_sorts.0.field_name", contactSortFieldName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "contact_sorts.0.direction", contactSortDirection),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "contact_sorts.0.numeric", contactSortNumeric),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "dynamic_contact_queueing_settings.0.sort", util.TrueValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "callable_time_set_id",
						"genesyscloud_outbound_callabletimeset."+callableTimeSetId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "dnc_list_ids.0",
						"genesyscloud_outbound_dnclist."+dncListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "rule_set_ids.0",
						"genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "site_id",
						"genesyscloud_telephony_providers_edges_site."+siteId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceLabel, "id"),
					provider.TestDefaultHomeDivision("genesyscloud_outbound_campaign."+resourceLabel),
				),
			},
			{
				// Update DCQSettings (ForceNew)
				Config: referencedResources + generateOutboundCampaign(
					resourceLabel,
					name,
					dialingMode,
					callerName,
					callerAddress,
					"genesyscloud_outbound_contact_list."+contactListResourceLabel+".id",
					util.NullValue, // campaign_status
					util.NullValue, // division id
					util.NullValue, // script id
					util.NullValue, // queue id
					"genesyscloud_telephony_providers_edges_site."+siteId+".id",
					"1",
					util.NullValue,
					"genesyscloud_outbound_callabletimeset."+callableTimeSetId+".id",
					"genesyscloud_outbound_callanalysisresponseset."+carResourceLabel+".id",
					"1",
					util.NullValue,
					"0",
					util.FalseValue,
					"40",
					"4",
					[]string{"genesyscloud_outbound_dnclist." + dncListResourceLabel + ".id"},
					[]string{"genesyscloud_outbound_ruleset." + ruleSetResourceLabel + ".id"},
					[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceLabel + ".id"},
					[]string{"en-us"},
					strconv.Quote("false"),
					generatePhoneColumnNoTypeBlock("Cell"),
					outbound.GenerateOutboundMessagingCampaignContactSort(
						contactSortFieldName,
						contactSortDirection,
						contactSortNumeric,
					),
					generateDynamicContactQueueingSettingsBlock(util.FalseValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "dialing_mode", dialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "caller_name", callerName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "caller_address", callerAddress),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "outbound_line_count", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "campaign_status", "off"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "abandon_rate", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "preview_time_out_seconds", "0"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "always_running", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "no_answer_timeout", "40"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "priority", "4"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "contact_sorts.0.field_name", contactSortFieldName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "contact_sorts.0.direction", contactSortDirection),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "contact_sorts.0.numeric", contactSortNumeric),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "dynamic_contact_queueing_settings.0.sort", util.FalseValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "callable_time_set_id",
						"genesyscloud_outbound_callabletimeset."+callableTimeSetId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "dnc_list_ids.0",
						"genesyscloud_outbound_dnclist."+dncListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "rule_set_ids.0",
						"genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "site_id",
						"genesyscloud_telephony_providers_edges_site."+siteId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceLabel, "id"),
					provider.TestDefaultHomeDivision("genesyscloud_outbound_campaign."+resourceLabel),
				),
			},
			{
				// Update
				Config: referencedResources + generateOutboundCampaign(
					resourceLabel,
					nameUpdated,
					dialingMode,
					callerNameUpdated,
					callerAddressUpdated,
					"genesyscloud_outbound_contact_list."+contactListResourceLabel+".id",
					util.NullValue, // campaign_status
					util.NullValue, // division id
					util.NullValue, // script id
					util.NullValue, // queue id
					"genesyscloud_telephony_providers_edges_site."+siteId+".id",
					"2",
					util.NullValue,
					"genesyscloud_outbound_callabletimeset."+callableTimeSetId+".id",
					"genesyscloud_outbound_callanalysisresponseset."+carResourceLabel+".id",
					"2",
					util.NullValue,
					"1",
					util.TrueValue,
					"30",
					"3",
					[]string{"genesyscloud_outbound_dnclist." + dncListResourceLabel + ".id"},
					[]string{"genesyscloud_outbound_ruleset." + ruleSetResourceLabel + ".id"},
					[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceLabel + ".id"},
					[]string{},
					strconv.Quote("false"),
					generatePhoneColumnNoTypeBlock("Cell"),
					outbound.GenerateOutboundMessagingCampaignContactSort(
						contactSortFieldName,
						contactSortDirection,
						contactSortNumeric,
					),
					generateDynamicContactQueueingSettingsBlock(util.FalseValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "name", nameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "dialing_mode", dialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "caller_name", callerNameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "caller_address", callerAddressUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "outbound_line_count", "2"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "abandon_rate", "2"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "preview_time_out_seconds", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "always_running", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "no_answer_timeout", "30"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "priority", "3"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "contact_sorts.0.field_name", contactSortFieldName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "contact_sorts.0.direction", contactSortDirection),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "contact_sorts.0.numeric", contactSortNumeric),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "dynamic_contact_queueing_settings.0.sort", "false"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "callable_time_set_id",
						"genesyscloud_outbound_callabletimeset."+callableTimeSetId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "dnc_list_ids.0",
						"genesyscloud_outbound_dnclist."+dncListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "rule_set_ids.0",
						"genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "site_id",
						"genesyscloud_telephony_providers_edges_site."+siteId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceLabel, "id"),
					provider.TestDefaultHomeDivision("genesyscloud_outbound_campaign."+resourceLabel),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_outbound_campaign." + resourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"queue_id"},
			},
		},
		CheckDestroy: testVerifyOutboundCampaignDestroyed,
	})
}

func TestAccResourceOutboundCampaignCampaignStatus(t *testing.T) {

	var (
		resourceLabel            = "campaign2"
		name                     = "Test Campaign " + uuid.NewString()
		contactListResourceLabel = "contact_list"
		carResourceLabel         = "car"
		siteId                   = "site"
		wrapupCodeResourceLabel  = "wrapupcode"
		outboundFlowFilePath     = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		flowName                 = "test flow " + uuid.NewString()
		flowResourceLabel        = "flow"
		wrapupcodeResourceLabel  = "wrapupcode"
		locationResourceLabel    = "location"
		divResourceLabel         = "test-division"
		divName                  = "terraform-" + uuid.NewString()
	)

	emergencyNumber := "+13178793429"
	if err := edgeSite.DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s: %v", emergencyNumber, err)
	}

	referencedResources := obContactList.GenerateOutboundContactList(
		contactListResourceLabel,
		"contact list "+uuid.NewString(),
		util.NullValue,
		strconv.Quote("Cell"),
		[]string{strconv.Quote("Cell")},
		[]string{strconv.Quote("Cell"), strconv.Quote("Home"), strconv.Quote("zipcode")},
		util.FalseValue,
		util.NullValue,
		util.NullValue,
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
	) + authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + routingWrapupcode.GenerateRoutingWrapupcodeResource(
		wrapupcodeResourceLabel,
		"tf wrapup code"+uuid.NewString(),
		"genesyscloud_auth_division."+divResourceLabel+".id",
	) + architect_flow.GenerateFlowResource(
		flowResourceLabel,
		outboundFlowFilePath,
		"",
		false,
		util.GenerateSubstitutionsMap(map[string]string{
			"flow_name":          flowName,
			"home_division_name": "${data.genesyscloud_auth_division_home.home.name}",
			"contact_list_name":  "${genesyscloud_outbound_contact_list." + contactListResourceLabel + ".name}",
			"wrapup_code_name":   "${genesyscloud_routing_wrapupcode." + wrapupCodeResourceLabel + ".name}",
		}),
	) + obResponseSet.GenerateOutboundCallAnalysisResponseSetResource(
		carResourceLabel,
		"tf car "+uuid.NewString(),
		util.FalseValue,
		obResponseSet.GenerateCarsResponsesBlock(
			obResponseSet.GenerateCarsResponse(
				"callable_person",
				"transfer_flow",
				flowName,
				"${genesyscloud_flow.flow.id}",
			),
		),
	) + location.GenerateLocationResource(
		locationResourceLabel,
		"tf location "+uuid.NewString(),
		"HQ1",
		[]string{},
		location.GenerateLocationEmergencyNum(
			"+13178793429",
			util.NullValue,
		),
		location.GenerateLocationAddress(
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
		"genesyscloud_location."+locationResourceLabel+".id",
		"Cloud",
		false,
		"[\"us-east-1\"]",
		util.NullValue,
		util.NullValue,
	) + "\ndata \"genesyscloud_auth_division_home\" \"home\" {}\n"

	// Test campaign_status can be turned on in a second run after first run's initial creation in off state, and then back off again
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
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
					`, resourceLabel, name, contactListResourceLabel, siteId, carResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "campaign_status", "off"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "site_id",
						"genesyscloud_telephony_providers_edges_site."+siteId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceLabel, "id"),
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
					`, resourceLabel, name, contactListResourceLabel, siteId, carResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "name", name),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "site_id",
						"genesyscloud_telephony_providers_edges_site."+siteId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceLabel, "id"),
					util.VerifyAttributeInArrayOfPotentialValues("genesyscloud_outbound_campaign."+resourceLabel, "campaign_status", []string{"on", "complete"}),
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
				`, resourceLabel, name, contactListResourceLabel, siteId, carResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "name", name),
					util.VerifyAttributeInArrayOfPotentialValues("genesyscloud_outbound_campaign."+resourceLabel, "campaign_status", []string{"off", "complete"}),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "site_id",
						"genesyscloud_telephony_providers_edges_site."+siteId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceLabel, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_outbound_campaign." + resourceLabel,
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
		resourceLabel            = "campaign3"
		name                     = "Test Campaign - " + uuid.NewString()
		contactListResourceLabel = "contact_list"
		carResourceLabel         = "car"
		siteId                   = "site"
		outboundFlowFilePath     = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		flowName                 = "test flow " + uuid.NewString()
		flowResourceLabel        = "flow"
		wrapupcodeResourceLabel  = "wrapupcode"
		locationResourceLabel    = "location"
		divResourceLabel         = "test-outbound-campaign-division"
		divName                  = "terraform-" + uuid.NewString()
	)

	emergencyNumber := "+13178793430"
	if err := edgeSite.DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s: %v", emergencyNumber, err)
	}

	// Test campaign_status can be turned on at time of creation as well
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			// Create resources for outbound campaign
			{
				Config: `data "genesyscloud_auth_division_home" "home" {}` + "\n" +
					authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
					GenerateReferencedResourcesForOutboundCampaignTests(
						contactListResourceLabel,
						"",
						"",
						carResourceLabel,
						outboundFlowFilePath,
						flowResourceLabel,
						flowName,
						"",
						siteId,
						emergencyNumber,
						"",
						"",
						"${data.genesyscloud_auth_division_home.home.name}",
						locationResourceLabel,
						wrapupcodeResourceLabel,
						divResourceLabel,
					),
				// Add contacts to the contact list (because we have access to the state and can pull out the contactlist ID to pass to the API)
				Check: addContactsToContactList,
			},
			// Now, we create the outbound campaign and it should stay running because it has contacts to call. We leave it running to test
			// the destroy command takes care of turning it off before deleting.
			{
				Config: `data "genesyscloud_auth_division_home" "home" {}` + "\n" +
					authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
					GenerateOutboundCampaignBasic(
						resourceLabel,
						name,
						contactListResourceLabel,
						siteId,
						emergencyNumber,
						carResourceLabel,
						strconv.Quote("on"),
						outboundFlowFilePath,
						flowResourceLabel,
						flowName,
						"${data.genesyscloud_auth_division_home.home.name}",
						locationResourceLabel,
						wrapupcodeResourceLabel,
						divResourceLabel,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "name", name),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "site_id",
						"genesyscloud_telephony_providers_edges_site."+siteId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceLabel, "id"),
					util.VerifyAttributeInArrayOfPotentialValues("genesyscloud_outbound_campaign."+resourceLabel, "campaign_status", []string{"on", "complete"}),
					func(s *terraform.State) error {
						time.Sleep(300 * time.Second) // Takes approx. 300 seconds for campaign to be completed / stopped
						return nil
					},
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_outbound_campaign." + resourceLabel,
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
		resourceLabel                = "campaign4"
		name                         = "Test Campaign " + uuid.NewString()
		dialingMode                  = "preview"
		callerName                   = "Test Name 123"
		callerAddress                = "+353371111111"
		contactListResourceLabel     = "contact_list"
		queueResourceLabel           = "queue"
		dncListResourceLabel         = "dnc_list"
		carResourceLabel             = "car"
		clfResourceLabel             = "clf"
		ruleSetResourceLabel         = "rule_set"
		callableTimeSetResourceLabel = "time_set"

		contactSortFieldName = "zipcode"
		contactSortDirection = "ASC"
		contactSortNumeric   = util.FalseValue
	)

	scriptId, err := getPublishedScriptId()
	if err != nil || scriptId == "" {
		t.Skip("Skipping as a published script ID is needed to run this test")
	}

	referencedResources := GenerateReferencedResourcesForOutboundCampaignTests(
		contactListResourceLabel,
		"",
		queueResourceLabel,
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
		"",
	)

	referencedResourcesUpdate := GenerateReferencedResourcesForOutboundCampaignTests(
		contactListResourceLabel,
		dncListResourceLabel,
		queueResourceLabel,
		carResourceLabel,
		"",
		"",
		"",
		clfResourceLabel,
		"",
		"",
		ruleSetResourceLabel,
		callableTimeSetResourceLabel,
		"",
		"",
		"",
		"",
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: referencedResources +
					generateOutboundCampaign(
						resourceLabel,
						name,
						dialingMode,
						callerName,
						callerAddress,
						"genesyscloud_outbound_contact_list."+contactListResourceLabel+".id",
						util.NullValue,
						util.NullValue,
						strconv.Quote(scriptId),
						"genesyscloud_routing_queue."+queueResourceLabel+".id",
						util.NullValue,
						util.NullValue,
						util.NullValue,
						util.NullValue,
						util.NullValue,
						util.NullValue,
						util.NullValue,
						util.NullValue,
						util.NullValue,
						util.NullValue,
						util.NullValue,
						[]string{},
						[]string{},
						[]string{},
						[]string{},
						strconv.Quote("false"),
						generatePhoneColumnNoTypeBlock("Cell"),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "dialing_mode", dialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "caller_name", callerName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "caller_address", callerAddress),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "script_id", scriptId),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "queue_id",
						"genesyscloud_routing_queue."+queueResourceLabel, "id"),
				),
			},
			{
				// Update
				Config: referencedResourcesUpdate +
					generateOutboundCampaign(
						resourceLabel,
						name,
						dialingMode,
						callerName,
						callerAddress,
						"genesyscloud_outbound_contact_list."+contactListResourceLabel+".id",
						util.NullValue,
						util.NullValue,
						strconv.Quote(scriptId),
						"genesyscloud_routing_queue."+queueResourceLabel+".id",
						util.NullValue,
						"1",
						util.NullValue,
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceLabel+".id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceLabel+".id",
						"1",
						util.FalseValue,
						"1",
						util.FalseValue,
						"3",
						"2",
						[]string{"genesyscloud_outbound_dnclist." + dncListResourceLabel + ".id"},
						[]string{"genesyscloud_outbound_ruleset." + ruleSetResourceLabel + ".id"},
						[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceLabel + ".id"},
						[]string{strconv.Quote("language")},
						strconv.Quote("false"),
						generatePhoneColumnNoTypeBlock("Cell"),
						outbound.GenerateOutboundMessagingCampaignContactSort(
							contactSortFieldName,
							contactSortDirection,
							contactSortNumeric,
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "dialing_mode", dialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "caller_name", callerName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "caller_address", callerAddress),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "abandon_rate", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "outbound_line_count", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "skip_preview_disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "preview_time_out_seconds", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "always_running", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "contact_sorts.0.field_name", contactSortFieldName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "contact_sorts.0.direction", contactSortDirection),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "contact_sorts.0.numeric", contactSortNumeric),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "no_answer_timeout", "3"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "priority", "2"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "script_id", scriptId),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "skill_columns.0", "language"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "auto_answer", "false"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "contact_list_filter_ids.0",
						"genesyscloud_outbound_contactlistfilter."+clfResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "dnc_list_ids.0",
						"genesyscloud_outbound_dnclist."+dncListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "queue_id",
						"genesyscloud_routing_queue."+queueResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "rule_set_ids.0",
						"genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "callable_time_set_id",
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceLabel, "id"),
					provider.TestDefaultHomeDivision("genesyscloud_outbound_campaign."+resourceLabel),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_campaign." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyOutboundCampaignDestroyed,
	})
}

func TestAccResourceOutboundCampaignPower(t *testing.T) {
	t.Parallel()
	var (
		resourceLabel            = "campaign5"
		name                     = "Test Campaign " + uuid.NewString()
		dialingMode              = "power"
		callerName               = "Test Name 123"
		callerAddress            = "+353371111111"
		contactListResourceLabel = "contact_list"
		queueResourceLabel       = "queue"
		locationResourceLabel    = "location"
		siteId                   = "site"
		carResourceLabel         = "car"
	)

	emergencyNumber := "+13178793431"
	if err := edgeSite.DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s: %v", emergencyNumber, err)
	}

	scriptId, err := getPublishedScriptId()
	if err != nil || scriptId == "" {
		t.Skip("Skipping as a published script ID is needed to run this test")
	}

	referencedResources := GenerateReferencedResourcesForOutboundCampaignTests(
		contactListResourceLabel,
		"",
		queueResourceLabel,
		carResourceLabel,
		"",
		"",
		"",
		"",
		siteId,
		emergencyNumber,
		"",
		"",
		"",
		locationResourceLabel,
		"",
		"",
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: referencedResources +
					generateOutboundCampaign(
						resourceLabel,
						name,
						dialingMode,
						callerName,
						callerAddress,
						"genesyscloud_outbound_contact_list."+contactListResourceLabel+".id",
						util.NullValue,
						util.NullValue,
						strconv.Quote(scriptId),
						"genesyscloud_routing_queue."+queueResourceLabel+".id",
						"genesyscloud_telephony_providers_edges_site."+siteId+".id",
						"1",
						"1",
						util.NullValue,
						"genesyscloud_outbound_callanalysisresponseset."+carResourceLabel+".id",
						"1",
						util.NullValue,
						util.NullValue,
						util.NullValue,
						util.NullValue,
						util.NullValue,
						[]string{},
						[]string{},
						[]string{},
						[]string{},
						strconv.Quote("true"),
						generatePhoneColumnNoTypeBlock("Cell"),
						generateDynamicLineBalancingSettingsBlock(util.FalseValue, "0"),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "dialing_mode", dialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "caller_name", callerName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "caller_address", callerAddress),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "max_calls_per_agent", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "outbound_line_count", "1"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "queue_id",
						"genesyscloud_routing_queue."+queueResourceLabel, "id"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel,
						"dynamic_line_balancing_settings.0.enabled", "false"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel,
						"dynamic_line_balancing_settings.0.relative_weight", "0"),
				),
			},
			{
				// update
				Config: referencedResources +
					generateOutboundCampaign(
						resourceLabel,
						name,
						dialingMode,
						callerName,
						callerAddress,
						"genesyscloud_outbound_contact_list."+contactListResourceLabel+".id",
						util.NullValue,
						util.NullValue,
						strconv.Quote(scriptId),
						"genesyscloud_routing_queue."+queueResourceLabel+".id",
						"genesyscloud_telephony_providers_edges_site."+siteId+".id",
						"1",
						"2",
						util.NullValue,
						"genesyscloud_outbound_callanalysisresponseset."+carResourceLabel+".id",
						util.NullValue,
						util.NullValue,
						util.NullValue,
						util.NullValue,
						util.NullValue,
						util.NullValue,
						[]string{},
						[]string{},
						[]string{},
						[]string{},
						strconv.Quote("true"),
						generatePhoneColumnNoTypeBlock("Cell"),
						generateDynamicLineBalancingSettingsBlock(util.TrueValue, "15"),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "dialing_mode", dialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "caller_name", callerName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "caller_address", callerAddress),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "max_calls_per_agent", "2"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "queue_id",
						"genesyscloud_routing_queue."+queueResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceLabel, "queue_id",
						"genesyscloud_routing_queue."+queueResourceLabel, "id"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel,
						"dynamic_line_balancing_settings.0.enabled", "true"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceLabel,
						"dynamic_line_balancing_settings.0.relative_weight", "15"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_campaign." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyOutboundCampaignDestroyed,
	})
}

func addContactsToContactList(state *terraform.State) error {
	outboundAPI := platformclientv2.NewOutboundApi()
	contactListResource := state.RootModule().Resources["genesyscloud_outbound_contact_list.contact_list"]
	if contactListResource == nil {
		return fmt.Errorf("genesyscloud_outbound_contact_list.contact_list contactListResource not found in state")
	}

	contactList, _, err := outboundAPI.GetOutboundContactlist(contactListResource.Primary.ID, false, false)
	if err != nil {
		return fmt.Errorf("genesyscloud_outbound_contact_list (%s) not available", contactListResource.Primary.ID)
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

func generateOutboundCampaign(
	resourceLabel string,
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
	maxCallsPerAgent string,
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
	skillColumns []string,
	autoAnswer string,
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
	max_calls_per_agent			  = %s
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
	skill_columns				  = [%s]
	auto_answer                   = %s
	%s
}
`, resourceLabel, name, dialingMode, callerName, callerAddress, contactListId, campaignStatus, divisionId, scriptId, queueId, siteId, abandonRate, maxCallsPerAgent, callableTimeSetId,
		callAnalysisResponseSetId, outboundLineCount, skipPreviewDisabled, previewTimeOutSeconds, alwaysRunning, noAnswerTimeout,
		priority, strings.Join(dncListIds, ", "), strings.Join(ruleSetIds, ", "), strings.Join(contactListFilterIds, ", "), strings.Join(skillColumns, ", "), autoAnswer,
		strings.Join(nestedBlocks, "\n"))
}

func generateDynamicContactQueueingSettingsBlock(sort string) string {
	return fmt.Sprintf(`
	dynamic_contact_queueing_settings {
		sort = %s
	}
	`, sort)
}

func generateDynamicLineBalancingSettingsBlock(enabled string, weight string) string {
	return fmt.Sprintf(`
	dynamic_line_balancing_settings {
		enabled = %s
		relative_weight = %s
	}
	`, enabled, weight)
}

func getPublishedScriptId() (string, error) {
	api := platformclientv2.NewScriptsApiWithConfig(sdkConfig)
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
		} else if util.IsStatus404(resp) {
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
