package flow_outcome

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceFlowOutcome(t *testing.T) {
	t.Skip("Skipping until a DELETE method is publicly available for flow outcomes.")
	var (
		outcomeResource1 = "flow-outcome1"
		name1            = "Terraform Code-" + uuid.NewString()

		name2       = "Terraform Code-" + uuid.NewString()
		description = "Edited description for flow outcome"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create using only required fields i.e. name
				Config: generateFlowOutcomeResource(
					outcomeResource1,
					name1,
					util.NullValue,
					util.NullValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_flow_outcome."+outcomeResource1, "name", name1),
					provider.TestDefaultHomeDivision("genesyscloud_flow_outcome."+outcomeResource1),
				),
			},
			{
				// Update with a new name and description
				Config: generateFlowOutcomeResource(
					outcomeResource1,
					name2,
					util.NullValue,
					description,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_flow_outcome."+outcomeResource1, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_flow_outcome."+outcomeResource1, "description", description),
					provider.TestDefaultHomeDivision("genesyscloud_flow_outcome."+outcomeResource1),
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
