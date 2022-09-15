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

func TestAccResourceOutboundCampaign(t *testing.T) {
	t.Parallel()
	var (
		resourceId            = "campaign"
		name                  = "Test Campaign " + uuid.NewString()
		dialingMode           = "agentless"
		callerName            = "Test Name"
		callerAddress         = "+353371111111"
		contactListResourceId = "contact_list"
		carResourceId         = "car"
		siteId                = "site"
		outboundFlowFilePath  = "../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
	)

	referencedResources := generateReferencedResourcesForOutboundCampaignTests(
		contactListResourceId,
		"",
		"",
		carResourceId,
		outboundFlowFilePath,
		"",
		siteId,
	)

	// TODO: Test out campaign resource with dialingMode as 'agentless', thereby omitting the scriptId field
	// Note: site can be set in place of edge group
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
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
						nullValue, // division id
						nullValue, // script id
						nullValue, // queue id required for all except agentless
						"genesyscloud_telephony_providers_edges_site."+siteId+".id",
						"1",
						nullValue, // callabletimeset id
						"genesyscloud_outbound_callanalysisresponseset."+carResourceId+".id",
						"1",
						nullValue,
						nullValue,
						nullValue,
						nullValue,
						nullValue,
						[]string{},
						[]string{},
						[]string{},
						generatePhoneColumnsBlock("Cell", "cell", ""),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "dialing_mode", dialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "caller_name", callerName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "caller_address", callerAddress),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "outbound_line_count", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "abandon_rate", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "site_id",
						"genesyscloud_telephony_providers_edges_site."+siteId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceId, "id"),
					testDefaultHomeDivision("genesyscloud_outbound_campaign."+resourceId),
				),
			},
		},
	})
}

func TestAccResourceOutboundCampaignWithScriptId(t *testing.T) {
	t.Parallel()
	var (
		resourceId            = "campaign"
		name                  = "Test Campaign " + uuid.NewString()
		dialingMode           = "preview"
		callerName            = "Test Name 123"
		callerAddress         = "+353371111111"
		contactListResourceId = "contact_list"
		scriptDataSourceId    = "data_script"
		queueResourceId       = "queue"
		dncListResourceId     = "dnc_list"
		carResourceId         = "car"
		clfResourceId         = "clf"

		contactSortFieldName = "zipcode"
		contactSortDirection = "ASC"
		contactSortNumeric   = falseValue

		// TODO: Replace these hard-coded GUIDS with resource references
		ruleSetId         = "8523bd1e-aaed-4fdf-8283-6884a516b298"
		callableTimeSetId = "5654dc1a-874d-439f-83af-f3c1271dcf7c"
	)

	// To test this function in your own org, pass the name of an existing script into the variable below.
	scriptName := "Outbound 1"
	if scriptName == "" {
		t.Skip("Skipping test until script resource is defined")
	}

	referencedResources := generateReferencedResourcesForOutboundCampaignTests(
		contactListResourceId,
		"",
		queueResourceId,
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
		clfResourceId,
		"",
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`data "genesyscloud_script" "%s" { name = "%s" }`, scriptDataSourceId, scriptName) +
					referencedResources +
					generateOutboundCampaign(
						resourceId,
						name,
						dialingMode,
						callerName,
						callerAddress,
						"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
						nullValue,
						"data.genesyscloud_script."+scriptDataSourceId+".id",
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
						generatePhoneColumnsBlock("Cell", "cell", ""),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "dialing_mode", dialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "caller_name", callerName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "caller_address", callerAddress),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "queue_id",
						"genesyscloud_routing_queue."+queueResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "script_id",
						"data.genesyscloud_script."+scriptDataSourceId, "id"),
				),
			},
			{
				Config: fmt.Sprintf(`data "genesyscloud_script" "%s" { name = "%s" }`, scriptDataSourceId, scriptName) +
					referencedResourcesUpdate +
					generateOutboundCampaign(
						resourceId,
						name,
						dialingMode,
						callerName,
						callerAddress,
						"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
						nullValue,
						"data.genesyscloud_script."+scriptDataSourceId+".id",
						"genesyscloud_routing_queue."+queueResourceId+".id",
						nullValue,
						"1",
						strconv.Quote(callableTimeSetId),
						"genesyscloud_outbound_callanalysisresponseset."+carResourceId+".id",
						"1",
						falseValue,
						"1",
						falseValue,
						"3",
						"2",
						[]string{"genesyscloud_outbound_dnclist." + dncListResourceId + ".id"},
						[]string{strconv.Quote(ruleSetId)},
						[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceId + ".id"},
						generatePhoneColumnsBlock(
							"Cell",
							"cell",
							"",
						),
						// TODO: Use generate func for contact sorts created on messaging campaigns branch
						fmt.Sprintf(`
	contact_sorts {
		field_name = "%s"
		direction  = "%s"
		numeric    = %s
	}
`, contactSortFieldName, contactSortDirection, contactSortNumeric),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "dialing_mode", dialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "caller_name", callerName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "caller_address", callerAddress),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "phone_columns.0.type", "cell"),
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
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "script_id",
						"data.genesyscloud_script."+scriptDataSourceId, "id"),
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
					validateStringInArray("genesyscloud_outbound_campaign."+resourceId, "rule_set_ids", ruleSetId),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "callable_time_set_id", callableTimeSetId),
					testDefaultHomeDivision("genesyscloud_outbound_campaign."+resourceId),
				),
			},
		},
		CheckDestroy: testVerifyOutboundCampaignDestroyed,
	})
}

func generateOutboundCampaign(
	resourceId string,
	name string,
	dialingMode string, // required
	callerName string, // required
	callerAddress string, // required
	contactListId string, // required
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
`, resourceId, name, dialingMode, callerName, callerAddress, contactListId, divisionId, scriptId, queueId, siteId, abandonRate, callableTimeSetId,
		callAnalysisResponseSetId, outboundLineCount, skipPreviewDisabled, previewTimeOutSeconds, alwaysRunning, noAnswerTimeout,
		priority, strings.Join(dncListIds, ", "), strings.Join(ruleSetIds, ", "), strings.Join(contactListFilterIds, ", "),
		strings.Join(nestedBlocks, "\n"))
}

func generateReferencedResourcesForOutboundCampaignTests(
	contactListResourceId string,
	dncListResourceId string,
	queueResourceId string,
	carResourceId string,
	outboundFlowFilePath string,
	clfResourceId string,
	siteId string) string {
	var (
		contactList             string
		dncList                 string
		queue                   string
		callAnalysisResponseSet string
		contactListFilter       string
		site                    string
	)
	if contactListResourceId != "" {
		contactList = generateOutboundContactList(
			contactListResourceId,
			"terraform contact list "+uuid.NewString(),
			"",
			"Cell",
			[]string{strconv.Quote("Cell")},
			[]string{strconv.Quote("Cell"), strconv.Quote("Home"), strconv.Quote("zipcode")},
			falseValue,
			"",
			"",
			generatePhoneColumnsBlock(
				"Cell",
				"cell",
				"Cell",
			),
			generatePhoneColumnsBlock(
				"Home",
				"home",
				"Home",
			))
	}
	if dncListResourceId != "" {
		dncList = generateOutboundDncListBasic(dncListResourceId, "tf dnc list "+uuid.NewString())
	}
	if queueResourceId != "" {
		queue = generateRoutingQueueResourceBasic(queueResourceId, "tf test queue "+uuid.NewString())
	}
	if carResourceId != "" {
		flowName := "test flow " + uuid.NewString()
		if outboundFlowFilePath != "" {
			callAnalysisResponseSet = fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) + generateRoutingWrapupcodeResource(
				"wrap-up-code",
				"wrapupcode "+uuid.NewString(),
			) + generateFlowResource(
				"test-flow",
				outboundFlowFilePath,
				"",
				generateFlowSubstitutions(map[string]string{
					"flow_name":          flowName,
					"home_division_name": "${data.genesyscloud_auth_division_home.home.name}",
					"contact_list_name":  "${genesyscloud_outbound_contact_list." + contactListResourceId + ".name}",
					"wrapup_code_name":   "${genesyscloud_routing_wrapupcode.wrap-up-code.name}",
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
						"${genesyscloud_flow.test-flow.id}",
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
resource "genesyscloud_location" "location" {
	name  = "%s"
	notes = "HQ1"
	path  = []
	emergency_number {
		number = "3173124740"
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
	location_id                     = genesyscloud_location.location.id
	media_model                     = "Cloud"
	media_regions_use_latency_based = false	
}
`, locationName, siteId, siteName)
	}
	return fmt.Sprintf(`
%s
%s
%s
%s
%s
%s
`, contactList, dncList, queue, callAnalysisResponseSet, contactListFilter, site)
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
		} else if isStatus404(resp) {
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
