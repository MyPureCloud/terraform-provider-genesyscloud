package outbound_callanalysisresponseset

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/architect_flow"
	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingWrapupcode "terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

func TestAccResourceOutboundCallAnalysisResponseSet(t *testing.T) {
	t.Parallel()
	var (
		resourceLabel       = "cars"
		name                = "Terraform test CAR " + uuid.NewString()
		identifier1         = "callable_person"
		identifier2         = "callable_fax"
		identifier3         = "callable_machine"
		reactionType        = "transfer"
		reactionTypeUpdated = "hangup"

		contactListResourceLabel = "contact-list"
		wrapupCodeResourceLabel  = "wrapup"
		flowResourceLabel        = "flow"

		contactListName      = "Terraform Test Contact List " + uuid.NewString()
		wrapupCodeName       = "Terraform Test WrapUpCode " + uuid.NewString()
		outboundFlowName     = "Terraform Test Flow " + uuid.NewString()
		outboundFlowFilePath = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"

		divResourceLabel = "test-division"
		divName          = "terraform-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateOutboundCallAnalysisResponseSetResource(
					resourceLabel,
					name,
					util.TrueValue,
					GenerateCarsResponsesBlock(
						GenerateCarsResponse(
							identifier1,
							reactionType,
							"",
							"",
						),
						GenerateCarsResponse(
							identifier2,
							reactionType,
							"",
							"",
						),
						GenerateCarsResponse(
							identifier3,
							reactionType,
							"",
							"",
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "beep_detection_enabled", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "responses.0."+identifier1+".0.reaction_type", "transfer"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "responses.0."+identifier2+".0.reaction_type", "transfer"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "responses.0."+identifier3+".0.reaction_type", "transfer"),
				),
			},
			// Update
			{
				Config: GenerateOutboundCallAnalysisResponseSetResource(
					resourceLabel,
					name,
					FalseValue,
					GenerateCarsResponsesBlock(
						GenerateCarsResponse(
							identifier1,
							reactionTypeUpdated,
							"",
							"",
						),
						GenerateCarsResponse(
							identifier2,
							reactionTypeUpdated,
							"",
							"",
						),
						GenerateCarsResponse(
							identifier3,
							reactionTypeUpdated,
							"",
							"",
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "beep_detection_enabled", FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "responses.0."+identifier1+".0.reaction_type", reactionTypeUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "responses.0."+identifier2+".0.reaction_type", reactionTypeUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "responses.0."+identifier3+".0.reaction_type", reactionTypeUpdated),
					// Check computed values are set
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "responses.0.callable_busy.0.reaction_type", "hangup"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "responses.0.callable_disconnect.0.reaction_type", "hangup"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "responses.0.callable_machine.0.reaction_type", "hangup"),
				),
			},
			{
				// Test outbound flow reference when reactionType is 'transfer_flow'
				Config: `data "genesyscloud_auth_division_home" "home" {}` + obContactList.GenerateOutboundContactList(
					contactListResourceLabel,
					contactListName,
					util.NullValue,
					util.NullValue,
					[]string{},
					[]string{strconv.Quote("Cell")},
					FalseValue,
					util.NullValue,
					util.NullValue,
					obContactList.GeneratePhoneColumnsBlock("Cell",
						"cell",
						util.NullValue,
					),
				) + authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + routingWrapupcode.GenerateRoutingWrapupcodeResource(
					wrapupCodeResourceLabel,
					wrapupCodeName,
					"genesyscloud_auth_division."+divResourceLabel+".id",
				) + architect_flow.GenerateFlowResource(
					flowResourceLabel,
					outboundFlowFilePath,
					"",
					false,
					util.GenerateSubstitutionsMap(map[string]string{
						"flow_name":          outboundFlowName,
						"home_division_name": "${data.genesyscloud_auth_division_home.home.name}",
						"contact_list_name":  "${genesyscloud_outbound_contact_list." + contactListResourceLabel + ".name}",
						"wrapup_code_name":   "${genesyscloud_routing_wrapupcode." + wrapupCodeResourceLabel + ".name}",
					}),
				) + GenerateOutboundCallAnalysisResponseSetResource(
					resourceLabel,
					name,
					FalseValue,
					GenerateCarsResponsesBlock(
						GenerateCarsResponse(
							"callable_person",
							"transfer_flow",
							outboundFlowName,
							"${genesyscloud_flow."+flowResourceLabel+".id}",
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "beep_detection_enabled", FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "responses.0.callable_person.0.reaction_type", "transfer_flow"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "responses.0.callable_person.0.name", outboundFlowName),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_callanalysisresponseset."+resourceLabel, "responses.0.callable_person.0.data",
						"genesyscloud_flow."+flowResourceLabel, "id"),
				),
			},
		},
		CheckDestroy: testVerifyCallAnalysisResponseSetDestroyed,
	})
}

func testVerifyCallAnalysisResponseSetDestroyed(state *terraform.State) error {
	outboundAPI := platformclientv2.NewOutboundApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_outbound_callanalysisresponseset" {
			continue
		}

		cars, resp, err := outboundAPI.GetOutboundCallanalysisresponseset(rs.Primary.ID)
		if cars != nil {
			return fmt.Errorf("call analysis response set (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// CARS not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All CARS destroyed
	return nil
}
