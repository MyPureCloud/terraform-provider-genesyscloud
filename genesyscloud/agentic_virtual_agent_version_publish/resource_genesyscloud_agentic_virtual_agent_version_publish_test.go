package agentic_virtual_agent_version_publish

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
   resource_genesyscloud_agentic_virtual_agent_version_publish_test.go contains acceptance tests
   for the publish resource.
*/

// TestAccResourceAgenticVirtualAgentVersionPublishTestReady tests publishing to TestReady.
func TestAccResourceAgenticVirtualAgentVersionPublishTestReady(t *testing.T) {
	var (
		agentResourceLabel   = "test_agent"
		versionResourceLabel = "test_version"
		publishResourceLabel = "test_publish"
		agentName            = "TF Publish Test Agent " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateAgentResource(agentResourceLabel, agentName) +
					generateVersionResource(versionResourceLabel, agentResourceLabel) +
					generatePublishResource(publishResourceLabel, agentResourceLabel, versionResourceLabel, "TestReady"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+publishResourceLabel, "status", "TestReady"),
					resource.TestCheckResourceAttrSet(ResourceType+"."+publishResourceLabel, "agent_id"),
					resource.TestCheckResourceAttrSet(ResourceType+"."+publishResourceLabel, "version"),
				),
			},
		},
	})
}

// TestAccResourceAgenticVirtualAgentVersionPublishProductionReady tests publishing to ProductionReady.
func TestAccResourceAgenticVirtualAgentVersionPublishProductionReady(t *testing.T) {
	var (
		agentResourceLabel   = "test_agent"
		versionResourceLabel = "test_version"
		publishResourceLabel = "test_publish"
		agentName            = "TF Publish Prod Test Agent " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateAgentResource(agentResourceLabel, agentName) +
					generateVersionResource(versionResourceLabel, agentResourceLabel) +
					generatePublishResource(publishResourceLabel, agentResourceLabel, versionResourceLabel, "ProductionReady"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+publishResourceLabel, "status", "ProductionReady"),
				),
			},
		},
	})
}

// TestAccResourceAgenticVirtualAgentVersionPublishLifecycle publishes a version to TestReady
// and then, in a second step, publishes a newer version to ProductionReady. This exercises the
// publish lifecycle across versions (TestReady -> ProductionReady on a subsequent version) and
// verifies cleanup via CheckDestroy on the parent agent.
func TestAccResourceAgenticVirtualAgentVersionPublishLifecycle(t *testing.T) {
	var (
		agentResourceLabel    = "test_agent"
		versionResourceLabel1 = "test_version_1"
		versionResourceLabel2 = "test_version_2"
		publishResourceLabel  = "test_publish"
		agentName             = "TF Publish Lifecycle Agent " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Publish the first version to TestReady
				Config: generateAgentResource(agentResourceLabel, agentName) +
					generateVersionResource(versionResourceLabel1, agentResourceLabel) +
					generatePublishResource(publishResourceLabel, agentResourceLabel, versionResourceLabel1, "TestReady"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+publishResourceLabel, "status", "TestReady"),
				),
			},
			{
				// Create a newer version and publish it to ProductionReady.
				// The publish resource's status is ForceNew, so this replaces the publish.
				Config: generateAgentResource(agentResourceLabel, agentName) +
					generateVersionResource(versionResourceLabel2, agentResourceLabel) +
					generatePublishResource(publishResourceLabel, agentResourceLabel, versionResourceLabel2, "ProductionReady"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+publishResourceLabel, "status", "ProductionReady"),
				),
			},
		},
		CheckDestroy: testVerifyAgenticVirtualAgentPublishParentDestroyed,
	})
}

// testVerifyAgenticVirtualAgentPublishParentDestroyed verifies the parent agent is destroyed.
// The publish resource itself has a no-op delete (no unpublish API), so destruction is verified
// via the parent agent being removed.
func testVerifyAgenticVirtualAgentPublishParentDestroyed(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type == "genesyscloud_agentic_virtual_agent" {
			return fmt.Errorf("agentic virtual agent %s still exists in state after destroy", rs.Primary.ID)
		}
	}
	return nil
}

// =============================================================================
// Config generators
// =============================================================================

func generateAgentResource(resourceLabel, name string) string {
	return fmt.Sprintf(`resource "genesyscloud_agentic_virtual_agent" "%s" {
		name = "%s"
	}
	`, resourceLabel, name)
}

func generateVersionResource(resourceLabel, agentResourceLabel string) string {
	return fmt.Sprintf(`resource "genesyscloud_agentic_virtual_agent_version" "%s" {
		agent_id = genesyscloud_agentic_virtual_agent.%s.id

		definition {
			role         = "You are a helpful test agent."
			instructions = ["Be concise and helpful."]
		}
	}
	`, resourceLabel, agentResourceLabel)
}

func generatePublishResource(resourceLabel, agentResourceLabel, versionResourceLabel, status string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		agent_id = genesyscloud_agentic_virtual_agent.%s.id
		version  = genesyscloud_agentic_virtual_agent_version.%s.version
		status   = "%s"
	}
	`, ResourceType, resourceLabel, agentResourceLabel, versionResourceLabel, status)
}
