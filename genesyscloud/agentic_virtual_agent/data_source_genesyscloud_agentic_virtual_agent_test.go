package agentic_virtual_agent

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
   data_source_genesyscloud_agentic_virtual_agent_test.go contains acceptance tests for the
   agentic_virtual_agent data source.
*/

func TestAccDataSourceAgenticVirtualAgent(t *testing.T) {
	var (
		agentResourceLabel   = "test_agent"
		agentDataSourceLabel = "agent_data"
		agentName            = "Terraform DS Test Agent " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create agent and then look it up via data source
				Config: generateAgenticVirtualAgentResource(
					agentResourceLabel,
					agentName,
					"",
				) + generateAgenticVirtualAgentDataSource(
					agentDataSourceLabel,
					agentName,
					ResourceType+"."+agentResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data."+ResourceType+"."+agentDataSourceLabel, "id",
						ResourceType+"."+agentResourceLabel, "id",
					),
				),
			},
		},
		CheckDestroy: testVerifyAgenticVirtualAgentDestroyed,
	})
}

// generateAgenticVirtualAgentDataSource generates the Terraform config for a data source lookup.
func generateAgenticVirtualAgentDataSource(resourceLabel, name, dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name       = "%s"
		depends_on = [%s]
	}
	`, ResourceType, resourceLabel, name, dependsOnResource)
}
