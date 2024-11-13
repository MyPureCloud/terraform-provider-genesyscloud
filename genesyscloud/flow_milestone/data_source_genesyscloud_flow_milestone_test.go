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
		milestoneResourceLabel = "flow-milestone"
		milestoneDataLabel     = "milestoneData"
		name                   = "Terraform Code-" + uuid.NewString()
		description            = "Sample Milestone by CX as Code"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateFlowMilestoneResource(
					milestoneResourceLabel,
					name,
					util.NullValue,
					description,
				),
			},
			{
				Config: generateFlowMilestoneResource(
					milestoneResourceLabel,
					name,
					util.NullValue,
					description,
				) + generateFlowMilestoneDataSource(
					milestoneDataLabel,
					name,
					"genesyscloud_flow_milestone."+milestoneResourceLabel,
				),
				PreConfig: func() {
					t.Log("sleeping to allow for eventual consistency")
					time.Sleep(3 * time.Second)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_flow_milestone."+milestoneDataLabel, "id", "genesyscloud_flow_milestone."+milestoneResourceLabel, "id"),
				),
			},
		},
	})
}

func generateFlowMilestoneDataSource(resourceLabel, name, dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_flow_milestone" "%s" {
		name       = "%s"
		depends_on =[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
