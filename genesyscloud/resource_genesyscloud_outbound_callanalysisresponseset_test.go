package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v74/platformclientv2"
	"strings"
	"testing"
)

func TestAccResourceCallAnalysisResponseSet(t *testing.T) {
	t.Parallel()
	var (
		resourceId          = "cars"
		name                = "Terraform test CAR " + uuid.NewString()
		identifier1         = "callable_person"
		identifier2         = "callable_fax"
		identifier3         = "callable_machine"
		reactionType        = "transfer"
		reactionTypeUpdated = "hangup"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateOutboundCallAnalysisResponseSetResource(
					resourceId,
					name,
					trueValue,
					generateCarsResponsesBlock(
						generateCarsResponse(
							identifier1,
							reactionType,
							"",
							"",
						),
						generateCarsResponse(
							identifier2,
							reactionType,
							"",
							"",
						),
						generateCarsResponse(
							identifier3,
							reactionType,
							"",
							"",
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "beep_detection_enabled", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0."+identifier1+".0.reaction_type", "transfer"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0."+identifier2+".0.reaction_type", "transfer"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0."+identifier3+".0.reaction_type", "transfer"),
				),
			},
			// Update
			{
				Config: generateOutboundCallAnalysisResponseSetResource(
					resourceId,
					name,
					falseValue,
					generateCarsResponsesBlock(
						generateCarsResponse(
							identifier1,
							reactionTypeUpdated,
							"",
							"",
						),
						generateCarsResponse(
							identifier2,
							reactionTypeUpdated,
							"",
							"",
						),
						generateCarsResponse(
							identifier3,
							reactionTypeUpdated,
							"",
							"",
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "beep_detection_enabled", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0."+identifier1+".0.reaction_type", reactionTypeUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0."+identifier2+".0.reaction_type", reactionTypeUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0."+identifier3+".0.reaction_type", reactionType),
					// Check computed values are set
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0.callable_busy.0.reaction_type", "hangup"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0.callable_disconnect.0.reaction_type", "hangup"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callanalysisresponseset."+resourceId, "responses.0.callable_machine.0.reaction_type", "hangup"),
				),
			},
			{
				// Generate contact list, wrap up code, outbound flow, and then CARS
			},
		},
		CheckDestroy: testVerifyCallAnalysisResponseSetDestroyed,
	})
}

func generateOutboundCallAnalysisResponseSetResource(resourceId string, name string, beepDetectionEnabled string, responsesBlock string) string {
	return fmt.Sprintf(`
resource "genesyscloud_outbound_callanalysisresponseset" "%s" {
	name                   = "%s"
	beep_detection_enabled = %s
	%s
}
`, resourceId, name, beepDetectionEnabled, responsesBlock)
}

func generateCarsResponsesBlock(nestedBlocks ...string) string {
	return fmt.Sprintf(`
	responses {
		%s
	}
`, strings.Join(nestedBlocks, "\n"))
}

func generateCarsResponse(identifier string, reactionType string, name string, data string) string {
	if name != "" {
		name = fmt.Sprintf(`name = "%s"`, name)
	}
	if data != "" {
		data = fmt.Sprintf(`data = "%s"`, data)
	}
	return fmt.Sprintf(`
		%s {
			reaction_type = "%s"
			%s
			%s
		}
`, identifier, reactionType, name, data)
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
		} else if isStatus404(resp) {
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
