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
		role1                = "You are a helpful banking assistant."
		role2                = "You are a friendly customer support agent."
		instruction1         = "Be polite and professional."
		instruction2         = "Always verify the customer identity first."
		instruction3         = "Escalate billing issues to a live agent."
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with minimal definition
				Config: generateAgentResource(agentResourceLabel, agentName) +
					generateVersionResourceBasic(versionResourceLabel, agentResourceLabel, role1, instruction1, instruction2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(ResourceType+"."+versionResourceLabel, "version"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "status", "Draft"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.role", role1),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.instructions.0", instruction1),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.instructions.1", instruction2),
				),
			},
			{
				// Update — change role and add an instruction
				Config: generateAgentResource(agentResourceLabel, agentName) +
					generateVersionResourceThreeInstructions(versionResourceLabel, agentResourceLabel, role2, instruction1, instruction2, instruction3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.role", role2),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.instructions.0", instruction1),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.instructions.1", instruction2),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.instructions.2", instruction3),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "status", "Draft"),
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
			{
				// Import/Read
				ResourceName:      ResourceType + "." + versionResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
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
			{
				// Import/Read
				ResourceName:      ResourceType + "." + versionResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccResourceAgenticVirtualAgentVersionWithModel tests create with an explicit model selection.
func TestAccResourceAgenticVirtualAgentVersionWithModel(t *testing.T) {
	var (
		agentResourceLabel   = "test_agent"
		versionResourceLabel = "test_version"
		agentName            = "TF Model Test Agent " + uuid.NewString()
		role                 = "You are a helpful assistant."
		instruction1         = "Be concise."
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateAgentResource(agentResourceLabel, agentName) +
					generateVersionResourceWithModelAndSettings(versionResourceLabel, agentResourceLabel, role, instruction1, "Preview"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(ResourceType+"."+versionResourceLabel, "version"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.role", role),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.model", "Preview"),
					// settings.comfort_statement coverage
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.settings.0.comfort_statement.0.enabled", "true"),
				),
			},
			{
				// Plan should be a no-op after apply (round-trip check for model + settings)
				Config: generateAgentResource(agentResourceLabel, agentName) +
					generateVersionResourceWithModelAndSettings(versionResourceLabel, agentResourceLabel, role, instruction1, "Preview"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
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

// TestAccResourceAgenticVirtualAgentVersionWithTypes tests create with type definitions
// (object with properties, an array type, an enum type, and a DataActionHttpError type).
func TestAccResourceAgenticVirtualAgentVersionWithTypes(t *testing.T) {
	var (
		agentResourceLabel   = "test_agent"
		versionResourceLabel = "test_version"
		agentName            = "TF Types Test Agent " + uuid.NewString()
		role                 = "You are a data-driven assistant."
		instruction1         = "Use the defined types."
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateAgentResource(agentResourceLabel, agentName) +
					generateVersionResourceWithTypes(versionResourceLabel, agentResourceLabel, role, instruction1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(ResourceType+"."+versionResourceLabel, "version"),
					// Object type with a property
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.types.0.name", "CustomerInfo"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.types.0.type", "object"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.types.0.properties.0.name", "accountId"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.types.0.properties.0.type", "string"),
					// Array type
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.types.1.name", "TagList"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.types.1.type", "array"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.types.1.items", "string"),
					// Enum type
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.types.2.name", "Priority"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.types.2.enum_values.0", "Low"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.types.2.enum_values.1", "High"),
					// DataActionHttpError type
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.types.3.name", "HttpError"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.types.3.type", "DataActionHttpError"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.types.3.status_codes.0", "404"),
				),
			},
			{
				// Round-trip check
				Config: generateAgentResource(agentResourceLabel, agentName) +
					generateVersionResourceWithTypes(versionResourceLabel, agentResourceLabel, role, instruction1),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
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

// TestAccResourceAgenticVirtualAgentVersionWithKnowledgeBaseTool tests create with a
// KnowledgeBase tool that references a real knowledge knowledgebase resource. This exercises
// the tools expand/flatten round-trip and target reference resolution.
func TestAccResourceAgenticVirtualAgentVersionWithKnowledgeBaseTool(t *testing.T) {
	var (
		agentResourceLabel   = "test_agent"
		versionResourceLabel = "test_version"
		kbResourceLabel      = "test_kb"
		agentName            = "TF KB Tool Test Agent " + uuid.NewString()
		kbName               = "TF KB Tool " + uuid.NewString()
		role                 = "You are a knowledge base assistant."
		instruction1         = "Answer using the knowledge base."
		toolName             = "LookupArticles"
		toolDesc             = "Search the knowledge base for relevant articles."
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateKnowledgebaseResource(kbResourceLabel, kbName) +
					generateAgentResource(agentResourceLabel, agentName) +
					generateVersionResourceWithKnowledgeBaseTool(
						versionResourceLabel, agentResourceLabel, kbResourceLabel, role, instruction1, toolName, toolDesc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(ResourceType+"."+versionResourceLabel, "version"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.tools.0.type", "KnowledgeBase"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.tools.0.name", toolName),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.tools.0.description", toolDesc),
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+versionResourceLabel, "definition.0.tools.0.target.0.id",
						"genesyscloud_knowledge_knowledgebase."+kbResourceLabel, "id",
					),
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

// TestAccResourceAgenticVirtualAgentVersionWithDataActionTool tests create with a DataAction
// tool referencing a real integration_action. This exercises the most complex tool expand/flatten
// paths: inputs (source/mapping), output, errors, input_validation (Python), and output_instructions.
func TestAccResourceAgenticVirtualAgentVersionWithDataActionTool(t *testing.T) {
	var (
		agentResourceLabel   = "test_agent"
		versionResourceLabel = "test_version"
		integResourceLabel   = "test_integ"
		actionResourceLabel  = "test_action"
		agentName            = "TF DataAction Tool Agent " + uuid.NewString()
		actionName           = "TF DA Tool Action " + uuid.NewString()
		role                 = "You are a data action assistant."
		instruction1         = "Use the data action tool to look up data."
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateDataActionDeps(integResourceLabel, actionResourceLabel, actionName) +
					generateAgentResource(agentResourceLabel, agentName) +
					generateVersionResourceWithDataActionTool(versionResourceLabel, agentResourceLabel, actionResourceLabel, role, instruction1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(ResourceType+"."+versionResourceLabel, "version"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.tools.0.type", "DataAction"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.tools.0.name", "LookupCustomer"),
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+versionResourceLabel, "definition.0.tools.0.target.0.id",
						"genesyscloud_integration_action."+actionResourceLabel, "id",
					),
					// Input
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.tools.0.inputs.0.target_name", "customerId"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.tools.0.inputs.0.source", "User"),
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.tools.0.inputs.0.required", "true"),
					// Output
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.tools.0.output", "CustomerResult"),
					// Error handling
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.tools.0.errors.0.type", "HttpError"),
					// Input validation (Python)
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.tools.0.input_validation.0.type", "Python"),
					// Output instructions (Python)
					resource.TestCheckResourceAttr(ResourceType+"."+versionResourceLabel, "definition.0.tools.0.output_instructions.0.type", "Python"),
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

func generateVersionResourceThreeInstructions(resourceLabel, agentResourceLabel, role, instruction1, instruction2, instruction3 string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		agent_id = genesyscloud_agentic_virtual_agent.%s.id

		definition {
			role = "%s"
			instructions = ["%s", "%s", "%s"]
		}
	}
	`, ResourceType, resourceLabel, agentResourceLabel, role, instruction1, instruction2, instruction3)
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

func generateVersionResourceWithModelAndSettings(resourceLabel, agentResourceLabel, role, instruction1, model string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		agent_id = genesyscloud_agentic_virtual_agent.%s.id

		definition {
			role         = "%s"
			instructions = ["%s"]
			model        = "%s"

			settings {
				comfort_statement {
					enabled = true
				}
			}
		}
	}
	`, ResourceType, resourceLabel, agentResourceLabel, role, instruction1, model)
}

func generateVersionResourceWithTypes(resourceLabel, agentResourceLabel, role, instruction1 string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		agent_id = genesyscloud_agentic_virtual_agent.%s.id

		definition {
			role         = "%s"
			instructions = ["%s"]

			types {
				name = "CustomerInfo"
				type = "object"
				properties {
					name     = "accountId"
					type     = "string"
					required = true
				}
			}

			types {
				name  = "TagList"
				type  = "array"
				items = "string"
			}

			types {
				name        = "Priority"
				type        = "string"
				enum_values = ["Low", "High"]
			}

			types {
				name                = "HttpError"
				type                = "DataActionHttpError"
				status_codes        = [404, 500]
				default_instruction = "Inform the user the request failed."
			}
		}
	}
	`, ResourceType, resourceLabel, agentResourceLabel, role, instruction1)
}

func generateKnowledgebaseResource(resourceLabel, name string) string {
	return fmt.Sprintf(`resource "genesyscloud_knowledge_knowledgebase" "%s" {
		name                   = "%s"
		description            = "Knowledge base for AVA tool test"
		core_language          = "en-US"
		content_search_enabled = true
	}
	`, resourceLabel, name)
}

func generateVersionResourceWithKnowledgeBaseTool(resourceLabel, agentResourceLabel, kbResourceLabel, role, instruction1, toolName, toolDesc string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		agent_id = genesyscloud_agentic_virtual_agent.%s.id

		definition {
			role         = "%s"
			instructions = ["%s"]

			tools {
				type        = "KnowledgeBase"
				name        = "%s"
				description = "%s"
				target {
					id   = genesyscloud_knowledge_knowledgebase.%s.id
					name = genesyscloud_knowledge_knowledgebase.%s.name
				}
				input_instructions = ["Use when the user asks a question"]
			}
		}
	}
	`, ResourceType, resourceLabel, agentResourceLabel, role, instruction1, toolName, toolDesc, kbResourceLabel, kbResourceLabel)
}

// generateDataActionDeps creates an integration (purecloud-data-actions) and an integration_action
// with a single-string input ("customerId") and output ("result") contract, to be used as a
// DataAction tool target.
func generateDataActionDeps(integResourceLabel, actionResourceLabel, actionName string) string {
	return fmt.Sprintf(`
resource "genesyscloud_integration" "%[1]s" {
	integration_type = "purecloud-data-actions"
}

resource "genesyscloud_integration_action" "%[2]s" {
	name            = "%[3]s"
	category        = "Genesys Cloud Data Actions"
	integration_id  = genesyscloud_integration.%[1]s.id
	secure          = false
	contract_input  = jsonencode({
		type       = "object"
		properties = { customerId = { type = "string" } }
		required   = ["customerId"]
	})
	contract_output = jsonencode({
		type       = "object"
		properties = { result = { type = "string" } }
		required   = ["result"]
	})
	config_request {
		request_url_template = "/api/v2/users/$${input.customerId}"
		request_type         = "GET"
	}
}
`, integResourceLabel, actionResourceLabel, actionName)
}

// generateVersionResourceWithDataActionTool creates a version with a DataAction tool that
// references the integration_action, exercising inputs, output, errors, input_validation, and
// output_instructions.
func generateVersionResourceWithDataActionTool(resourceLabel, agentResourceLabel, actionResourceLabel, role, instruction1 string) string {
	return fmt.Sprintf(`resource "%[1]s" "%[2]s" {
		agent_id = genesyscloud_agentic_virtual_agent.%[3]s.id

		definition {
			role         = "%[4]s"
			instructions = ["%[5]s"]

			types {
				name      = "CustomerId"
				type      = "string"
				direction = "Input"
			}

			types {
				name      = "CustomerResult"
				type      = "object"
				direction = "Output"
				properties {
					name = "result"
					type = "string"
				}
			}

			types {
				name         = "HttpError"
				type         = "DataActionHttpError"
				status_codes = [404, 500]
			}

			tools {
				type        = "DataAction"
				name        = "LookupCustomer"
				description = "Look up a customer by id via a data action."
				target {
					id   = genesyscloud_integration_action.%[6]s.id
					name = genesyscloud_integration_action.%[6]s.name
				}

				inputs {
					target_name = "customerId"
					type        = "CustomerId"
					source      = "User"
					required    = true
				}

				output = "CustomerResult"

				errors {
					type        = "HttpError"
					instruction = "Tell the user the lookup failed."
				}

				input_validation {
					type             = "Python"
					if_condition     = "len(customerId) > 0"
					else_instruction = "Ask the user for a valid customer id."
				}

				output_instructions {
					type = "Python"
					when = "result is not None"
					then = "Summarize the customer result."
				}
			}
		}
	}
	`, ResourceType, resourceLabel, agentResourceLabel, role, instruction1, actionResourceLabel)
}
