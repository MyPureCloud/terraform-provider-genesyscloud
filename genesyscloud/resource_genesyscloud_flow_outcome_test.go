package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccResourceFlowOutcome(t *testing.T) {
	var (
		outcomeResource1 = "flow-outcome1"
		name1            = "Terraform Code-" + uuid.NewString()
		description1     = "Sample flow outcome for CX as Code"

		name2        = "Terraform Code-" + uuid.NewString()
		description2 = "Edited description for flow outcome"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateFlowOutcomeResource(
					outcomeResource1,
					name1,
					nullValue,
					description1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_flow_outcome."+outcomeResource1, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_flow_outcome."+outcomeResource1, "description", description1),
					testDefaultHomeDivision("genesyscloud_flow_outcome."+outcomeResource1),
				),
			},
			{
				// Update with a new name and description
				Config: generateFlowOutcomeResource(
					outcomeResource1,
					name2,
					nullValue,
					description2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_flow_outcome."+outcomeResource1, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_flow_outcome."+outcomeResource1, "description", description2),
					testDefaultHomeDivision("genesyscloud_flow_outcome."+outcomeResource1),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_flow_outcome." + outcomeResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func generateFlowOutcomeResource(
	outcomeResource string,
	name string,
	divisionId string,
	description string) string {
	return fmt.Sprintf(`resource "genesyscloud_flow_outcome" "%s" {
		name = "%s"
		division_id = %s
		description = "%s"
	}
	`, outcomeResource, name, divisionId, description)
}
