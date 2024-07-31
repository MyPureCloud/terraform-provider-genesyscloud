package flow_milestone

import (
	"fmt"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceFlowMilestone(t *testing.T) {
	var (
		milestoneResource1 = "flow-milestone1"
		name1              = "Terraform Code-" + uuid.NewString()
		description1       = "Sample flow milestone for CX as Code"
		divResource        = "test-division"
		divName            = "terraform-" + uuid.NewString()

		name2        = "Terraform Code-" + uuid.NewString()
		description2 = "Edited description for flow milestone"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateFlowMilestoneResource(
					milestoneResource1,
					name1,
					util.NullValue,
					description1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_flow_milestone."+milestoneResource1, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_flow_milestone."+milestoneResource1, "description", description1),
					provider.TestDefaultHomeDivision("genesyscloud_flow_milestone."+milestoneResource1),
				),
			},
			{
				// Update with a new name and description
				Config: generateFlowMilestoneResource(
					milestoneResource1,
					name2,
					util.NullValue,
					description2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_flow_milestone."+milestoneResource1, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_flow_milestone."+milestoneResource1, "description", description2),
					provider.TestDefaultHomeDivision("genesyscloud_flow_milestone."+milestoneResource1),
				),
			},
			{
				// Update with a new division
				Config: gcloud.GenerateAuthDivisionBasic(divResource, divName) + generateFlowMilestoneResource(
					milestoneResource1,
					name2,
					"genesyscloud_auth_division."+divResource+".id",
					description2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_flow_milestone."+milestoneResource1, "name", name2),
					resource.TestCheckResourceAttrPair("genesyscloud_flow_milestone."+milestoneResource1, "division_id", "genesyscloud_auth_division."+divResource, "id"),
					resource.TestCheckResourceAttr("genesyscloud_flow_milestone."+milestoneResource1, "description", description2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_flow_milestone." + milestoneResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyFlowMilestoneDestroyed,
	})
}

func generateFlowMilestoneResource(
	milestoneResource string,
	name string,
	divisionId string,
	description string) string {
	return fmt.Sprintf(`resource "genesyscloud_flow_milestone" "%s" {
		name = "%s"
		division_id = %s
		description = "%s"
	}
	`, milestoneResource, name, divisionId, description)
}

func testVerifyFlowMilestoneDestroyed(state *terraform.State) error {
	archAPi := platformclientv2.NewArchitectApi()

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_flow_milestone" {
			continue
		}

		milestone, resp, err := archAPi.GetFlowsMilestone(rs.Primary.ID)
		if milestone != nil {
			return fmt.Errorf("Milestone (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Milestone not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}

	// Success. All milestones destroyed
	return nil
}
