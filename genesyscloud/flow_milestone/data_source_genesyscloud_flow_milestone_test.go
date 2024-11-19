package flow_milestone

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

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
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateFlowMilestoneResource(
					milestoneRes,
					name,
					util.NullValue,
					description,
				),
			},
			{
				Config: generateFlowMilestoneResource(
					milestoneRes,
					name,
					util.NullValue,
					description,
				) + generateFlowMilestoneDataSource(
					milestoneData,
					name,
					"genesyscloud_flow_milestone."+milestoneRes,
				),
				PreConfig: func() {
					t.Log("sleeping to allow for eventual consistency")
					time.Sleep(3 * time.Second)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_flow_milestone."+milestoneData, "id", "genesyscloud_flow_milestone."+milestoneRes, "id"),
				),
			},
		},
	})
}

func generateFlowMilestoneDataSource(resourceID, name, dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_flow_milestone" "%s" {
		name       = "%s"
		depends_on =[%s]
	}
	`, resourceID, name, dependsOnResource)
}
