package agentic_virtual_agent

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
   resource_genesyscloud_agentic_virtual_agent_test.go contains acceptance tests for the
   agentic_virtual_agent resource. These tests create real resources in a Genesys Cloud org.
*/

// TestAccResourceAgenticVirtualAgentBasic tests basic create, update, read, import, and destroy.
func TestAccResourceAgenticVirtualAgentBasic(t *testing.T) {
	var (
		resourceLabel = "test_agent"
		name1         = "Terraform Test Agent " + uuid.NewString()
		name2         = "Terraform Test Agent Updated " + uuid.NewString()
		imageUri1     = "https://example.com/images/agent-avatar-1.png"
		imageUri2     = "https://example.com/images/agent-avatar-2.png"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with name only
				Config: generateAgenticVirtualAgentResource(resourceLabel, name1, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "status", "Draft"),
				),
			},
			{
				// Update name and add image_uri
				Config: generateAgenticVirtualAgentResource(resourceLabel, name2, imageUri1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "image_uri", imageUri1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "status", "Draft"),
				),
			},
			{
				// Update image_uri only
				Config: generateAgenticVirtualAgentResource(resourceLabel, name2, imageUri2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "image_uri", imageUri2),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyAgenticVirtualAgentDestroyed,
	})
}

// testVerifyAgenticVirtualAgentDestroyed verifies that the agent was successfully destroyed.
func testVerifyAgenticVirtualAgentDestroyed(state *terraform.State) error {
	sdkConfig := provider.GetProviderMeta().ClientConfig
	proxy := getAgenticVirtualAgentProxy(sdkConfig)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}
		agent, resp, err := proxy.getAgenticVirtualAgentById(context.Background(), rs.Primary.ID)
		if agent != nil {
			return fmt.Errorf("%s (%s) still exists", ResourceType, rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			continue
		} else {
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	return nil
}

// generateAgenticVirtualAgentResource generates the Terraform config for a test agent.
func generateAgenticVirtualAgentResource(resourceLabel, name, imageUri string) string {
	imageUriAttr := ""
	if imageUri != "" {
		imageUriAttr = fmt.Sprintf(`image_uri = "%s"`, imageUri)
	}

	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		%s
	}
	`, ResourceType, resourceLabel, name, imageUriAttr)
}
