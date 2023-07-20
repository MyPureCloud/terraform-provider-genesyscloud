package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceFlowMilestone(t *testing.T) {
	var (
		milestoneRes  = "flow-milestone"
		milestoneData = "milestoneData"
		name          = "Terraform Code-" + uuid.NewString()
		description   = "Sample Milestone by CX as Code"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateFlowMilestoneResource(
					milestoneRes,
					name,
					nullValue,
					description,
				) + generateFlowMilestoneDataSource(
					milestoneData,
					name,
					"genesyscloud_flow_milestone."+milestoneRes,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_flow_milestone."+milestoneData, "id", "genesyscloud_flow_milestone."+milestoneRes, "id"),
				),
			},
		},
	})
}

func generateFlowMilestoneDataSource(resourceID string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_flow_milestone" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
