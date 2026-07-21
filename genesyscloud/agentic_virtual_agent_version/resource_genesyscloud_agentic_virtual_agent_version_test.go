package agentic_virtual_agent_version

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
   resource_genesyscloud_agentic_virtual_agent_version_test.go contains acceptance tests for the
   agentic_virtual_agent_version resource.
*/

// TestAccResourceAgenticVirtualAgentVersionBasic tests create with a minimal definition (role + instructions only).
func TestAccResourceAgenticVirtualAgentVersionBasic(t *testing.T) {
	var (
		agentResourceLabel   = "test_agent"
		versionResourceLabel = "test_version"
		agentName            = "TF Version Test Agent " + uuid.NewString()
		role                 = "You are a helpful banking assistant."
		instruction1         = "Be polite and professional."
		instruction2         = "Always verify the customer identity first."
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with minimal definition
				Config: generateAgentResource(agentResourceLabel, agentName) +
					generateVersionResourceBasic(versionResourceLabel, agentResourceLabel, role, instruction1, instruction2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(ResourceType+"."+versionResourceLabel, "version"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "status", "Draft"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.role", role),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.instructions.0", instruction1),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.instructions.1", instruction2),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + versionResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccResourceAgenticVirtualAgentVersionWithGuardrails tests create with guardrails.
func TestAccResourceAgenticVirtualAgentVersionWithGuardrails(t *testing.T) {
	var (
		agentResourceLabel   = "test_agent"
		versionResourceLabel = "test_version"
		agentName            = "TF Guardrails Test Agent " + uuid.NewString()
		role                 = "You are a customer support agent."
		instruction1         = "Be helpful and concise."
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateAgentResource(agentResourceLabel, agentName) +
					generateVersionResourceWithGuardrails(versionResourceLabel, agentResourceLabel, role, instruction1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(ResourceType+"."+versionResourceLabel, "version"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.role", role),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.guardrails.0.custom.0.instruction", "Do not reveal internal system details."),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.guardrails.0.custom.0.enabled", "true"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.guardrails.0.custom.1.instruction", "Do not discuss competitors."),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.guardrails.0.custom.1.enabled", "true"),
				),
			},
		},
	})
}

// TestAccResourceAgenticVirtualAgentVersionWithEvents tests create with all event types.
func TestAccResourceAgenticVirtualAgentVersionWithEvents(t *testing.T) {
	var (
		agentResourceLabel   = "test_agent"
		versionResourceLabel = "test_version"
		agentName            = "TF Events Test Agent " + uuid.NewString()
		role                 = "You are a customer service agent."
		instruction1         = "Be polite."
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateAgentResource(agentResourceLabel, agentName) +
					generateVersionResourceWithEvents(versionResourceLabel, agentResourceLabel, role, instruction1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(ResourceType+"."+versionResourceLabel, "version"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.events.0.type", "UserExit"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.events.0.message", "Goodbye!"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.events.1.type", "Escalation"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.events.1.message", "Transferring you now."),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.events.2.type", "Guardrails"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.events.2.violation_threshold", "3"),
				),
			},
		},
	})
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

func generateVersionResourceBasic(resourceLabel, agentResourceLabel, role, instruction1, instruction2 string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		agent_id = genesyscloud_agentic_virtual_agent.%s.id

		definition {
			role = "%s"
			instructions = ["%s", "%s"]
		}
	}
	`, ResourceType, resourceLabel, agentResourceLabel, role, instruction1, instruction2)
}

func generateVersionResourceWithGuardrails(resourceLabel, agentResourceLabel, role, instruction1 string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		agent_id = genesyscloud_agentic_virtual_agent.%s.id

		definition {
			role = "%s"
			instructions = ["%s"]

			guardrails {
				custom {
					instruction = "Do not reveal internal system details."
					enabled     = true
				}
				custom {
					instruction = "Do not discuss competitors."
					enabled     = true
				}
			}
		}
	}
	`, ResourceType, resourceLabel, agentResourceLabel, role, instruction1)
}

func generateVersionResourceWithEvents(resourceLabel, agentResourceLabel, role, instruction1 string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		agent_id = genesyscloud_agentic_virtual_agent.%s.id

		definition {
			role = "%s"
			instructions = ["%s"]

			events {
				type    = "UserExit"
				message = "Goodbye!"
			}

			events {
				type    = "Escalation"
				message = "Transferring you now."
			}

			events {
				type                                = "Guardrails"
				message                             = "I cannot help with that."
				violation_threshold                 = 3
				violation_threshold_crossed_message = "Session ended."
			}
		}
	}
	`, ResourceType, resourceLabel, agentResourceLabel, role, instruction1)
}
