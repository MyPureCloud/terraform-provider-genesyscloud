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
		resourceId            = "campaign1"
		name                  = "Test Campaign " + uuid.NewString()
		dialingMode           = "agentless"
		callerName            = "Test Name"
		callerAddress         = "+353371111111"
		contactListResourceId = "contact_list"
		dncListResourceId     = "dnc"
		queueResourceId       = "queue"
		clfResourceId         = "clf"
		carResourceId         = "car"
		ruleSetResourceId     = "rule_set"
		siteId                = "site"
		callableTimeSetId     = "time_set"
		outboundFlowFilePath  = "../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		flowName              = "test flow " + uuid.NewString()

		contactSortFieldName = "zipcode"
		contactSortDirection = "ASC"
		contactSortNumeric   = falseValue

		nameUpdated          = "Test Campaign " + uuid.NewString()
		callerNameUpdated    = "Test Name 2"
		callerAddressUpdated = "+353371112111"
	)

	// necessary to avoid errors during site creation
	err := authorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	emergencyNumber := "+13178793430"
	err = deleteLocationWithNumber(emergencyNumber)
	if err != nil {
		t.Fatal(err)
	}

	referencedResources := generateReferencedResourcesForOutboundCampaignTests(
		contactListResourceId,
		dncListResourceId,
		queueResourceId,
		carResourceId,
		outboundFlowFilePath,
		flowName,
		clfResourceId,
		siteId,
		emergencyNumber,
		ruleSetResourceId,
		callableTimeSetId,
	)

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
						generatePhoneColumnsBlock("Cell", "cell", ""),
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
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "outbound_line_count", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "campaign_status", "off"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "abandon_rate", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "preview_time_out_seconds", "0"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "always_running", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "no_answer_timeout", "40"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "priority", "4"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.field_name", contactSortFieldName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.direction", contactSortDirection),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.numeric", contactSortNumeric),
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
					testDefaultHomeDivision("genesyscloud_outbound_campaign."+resourceId),
				),
			},
			{
				// Update
				Config: referencedResources +
					generateOutboundCampaign(
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
						generatePhoneColumnsBlock("Cell", "cell", ""),
						generateOutboundMessagingCampaignContactSort(
							contactSortFieldName,
							contactSortDirection,
							contactSortNumeric,
						),
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
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.field_name", contactSortFieldName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.direction", contactSortDirection),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "contact_sorts.0.numeric", contactSortNumeric),
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
					testDefaultHomeDivision("genesyscloud_outbound_campaign."+resourceId),
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

func TestAccResourceOutboundCampaignBasic(t *testing.T) {
	t.Parallel()
	var (
		resourceId            = "campaign2"
		name                  = "Test Campaign " + uuid.NewString()
		contactListResourceId = "contact_list"
		carResourceId         = "car"
		siteId                = "site"
		outboundFlowFilePath  = "../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		flowName              = "test flow " + uuid.NewString()
	)

	// necessary to avoid errors during site creation
	err := authorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	emergencyNumber := "+13178793429"
	err = deleteLocationWithNumber(emergencyNumber)
	if err != nil {
		t.Fatal(err)
	}

	// Test campaign_status update
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateOutboundCampaignBasic(
					resourceId,
					name,
					contactListResourceId,
					siteId,
					emergencyNumber,
					carResourceId,
					strconv.Quote("off"),
					outboundFlowFilePath,
					flowName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "campaign_status", "off"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "site_id",
						"genesyscloud_telephony_providers_edges_site."+siteId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceId, "id"),
				),
			},
			{
				Config: generateOutboundCampaignBasic(
					resourceId,
					name,
					contactListResourceId,
					siteId,
					emergencyNumber,
					carResourceId,
					strconv.Quote("on"),
					outboundFlowFilePath,
					flowName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaign."+resourceId, "name", name),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "site_id",
						"genesyscloud_telephony_providers_edges_site."+siteId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "call_analysis_response_set_id",
						"genesyscloud_outbound_callanalysisresponseset."+carResourceId, "id"),
					verifyAttributeInArrayOfPotentialValues("genesyscloud_outbound_campaign."+resourceId, "campaign_status", []string{"on", "complete"}),
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

func TestAccResourceOutboundCampaignWithScriptId(t *testing.T) {
	t.Parallel()
	var (
		resourceId                = "campaign3"
		name                      = "Test Campaign " + uuid.NewString()
		dialingMode               = "preview"
		callerName                = "Test Name 123"
		callerAddress             = "+353371111111"
		contactListResourceId     = "contact_list"
		scriptDataSourceId        = "data_script"
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

	// To test this function in your own org, pass the name of an existing script into the variable below.
	scriptName := ""
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
		clfResourceId,
		"",
		"",
		ruleSetResourceId,
		callableTimeSetResourceId,
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
				// Update
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
						nullValue,
						"data.genesyscloud_script."+scriptDataSourceId+".id",
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
						generatePhoneColumnsBlock(
							"Cell",
							"cell",
							"",
						),
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
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "rule_set_ids.0",
						"genesyscloud_outbound_ruleset."+ruleSetResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaign."+resourceId, "callable_time_set_id",
						"genesyscloud_outbound_callabletimeset."+callableTimeSetResourceId, "id"),
					testDefaultHomeDivision("genesyscloud_outbound_campaign."+resourceId),
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

func generateOutboundCampaignBasic(resourceId string,
	name string,
	contactListResourceId string,
	siteResourceId string,
	emergencyNumber string,
	carResourceId string,
	campaignStatus string,
	outboundFlowFilePath string,
	flowName string) string {
	referencedResources := generateReferencedResourcesForOutboundCampaignTests(
		contactListResourceId,
		"",
		"",
		carResourceId,
		outboundFlowFilePath,
		flowName,
		"",
		siteResourceId,
		emergencyNumber,
		"",
		"",
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
		type        = "cell"
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

func generateReferencedResourcesForOutboundCampaignTests(
	contactListResourceId string,
	dncListResourceId string,
	queueResourceId string,
	carResourceId string,
	outboundFlowFilePath string,
	flowName string,
	clfResourceId string,
	siteId string,
	emergencyNumber string,
	ruleSetId string,
	callableTimeSetResourceId string) string {
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
		if outboundFlowFilePath != "" {
			callAnalysisResponseSet = fmt.Sprintf(`
data "genesyscloud_auth_division_home" "division" {}
`) + generateRoutingWrapupcodeResource(
				"wrap-up-code",
				"wrapupcode "+uuid.NewString(),
			) + generateFlowResource(
				"test-flow",
				outboundFlowFilePath,
				"",
				false,
				generateFlowSubstitutions(map[string]string{
					"flow_name":          flowName,
					"home_division_name": "${data.genesyscloud_auth_division_home.division.name}",
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
	location_id                     = genesyscloud_location.location.id
	media_model                     = "Cloud"
	media_regions_use_latency_based = false	
}
`, locationName, emergencyNumber, siteId, siteName)
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
