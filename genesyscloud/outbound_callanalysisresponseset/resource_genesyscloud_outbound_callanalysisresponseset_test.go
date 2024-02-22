package outbound_callanalysisresponseset

import (
	"fmt"
	"strconv"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
)

func TestAccResourceOutboundCallAnalysisResponseSet(t *testing.T) {
	t.Parallel()
	var (
		resourceId          = "cars"
		name                = "Terraform test CAR " + uuid.NewString()
		identifier1         = "callable_person"
		identifier2         = "callable_fax"
		identifier3         = "callable_machine"
		reactionType        = "transfer"
		reactionTypeUpdated = "hangup"

		contactListResourceId = "contact-list"
		wrapupCodeResourceId  = "wrapup"
		flowResourceId        = "flow"

		contactListName      = "Terraform Test Contact List " + uuid.NewString()
		wrapupCodeName       = "Terraform Test WrapUpCode " + uuid.NewString()
		outboundFlowName     = "Terraform Test Flow " + uuid.NewString()
		outboundFlowFilePath = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateOutboundCallAnalysisResponseSetResource(
					resourceId,
					name,
					gcloud.TrueValue,
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
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "beep_detection_enabled", gcloud.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0."+identifier1+".0.reaction_type", "transfer"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0."+identifier2+".0.reaction_type", "transfer"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0."+identifier3+".0.reaction_type", "transfer"),
				),
			},
			// Update
			{
				Config: GenerateOutboundCallAnalysisResponseSetResource(
					resourceId,
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
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "beep_detection_enabled", FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0."+identifier1+".0.reaction_type", reactionTypeUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0."+identifier2+".0.reaction_type", reactionTypeUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0."+identifier3+".0.reaction_type", reactionTypeUpdated),
					// Check computed values are set
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0.callable_busy.0.reaction_type", "hangup"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0.callable_disconnect.0.reaction_type", "hangup"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0.callable_machine.0.reaction_type", "hangup"),
				),
			},
			{
				// Test outbound flow reference when reactionType is 'transfer_flow'
				Config: `data "genesyscloud_auth_division_home" "home" {}` + obContactList.GenerateOutboundContactList(
					contactListResourceId,
					contactListName,
					gcloud.NullValue,
					gcloud.NullValue,
					[]string{},
					[]string{strconv.Quote("Cell")},
					FalseValue,
					gcloud.NullValue,
					gcloud.NullValue,
					obContactList.GeneratePhoneColumnsBlock("Cell",
						"cell",
						gcloud.NullValue,
					),
				) + gcloud.GenerateRoutingWrapupcodeResource(
					wrapupCodeResourceId,
					wrapupCodeName,
				) + gcloud.GenerateFlowResource(
					flowResourceId,
					outboundFlowFilePath,
					"",
					false,
					gcloud.GenerateSubstitutionsMap(map[string]string{
						"flow_name":          outboundFlowName,
						"home_division_name": "${data.genesyscloud_auth_division_home.home.name}",
						"contact_list_name":  "${genesyscloud_outbound_contact_list." + contactListResourceId + ".name}",
						"wrapup_code_name":   "${genesyscloud_routing_wrapupcode." + wrapupCodeResourceId + ".name}",
					}),
				) + GenerateOutboundCallAnalysisResponseSetResource(
					resourceId,
					name,
					FalseValue,
					GenerateCarsResponsesBlock(
						GenerateCarsResponse(
							"callable_person",
							"transfer_flow",
							outboundFlowName,
							"${genesyscloud_flow."+flowResourceId+".id}",
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "beep_detection_enabled", FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0.callable_person.0.reaction_type", "transfer_flow"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0.callable_person.0.name", outboundFlowName),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0.callable_person.0.data",
						"genesyscloud_flow."+flowResourceId, "id"),
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
		} else if gcloud.IsStatus404(resp) {
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
