package agentic_virtual_agent_version

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
   resource_genesyscloud_agentic_virtual_agent_version_schema.go holds:

   1. The registration code that registers the Resource for the package.
   2. The resource schema definition for the agentic_virtual_agent_version resource.
*/

// SetRegistrar registers the resource.
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceAgenticVirtualAgentVersion())
}

// ResourceAgenticVirtualAgentVersion registers the Terraform resource.
func ResourceAgenticVirtualAgentVersion() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Agentic Virtual Agent Version. Manages a version containing the full agent definition.",
		CreateContext: provider.CreateWithPooledClient(createAgenticVirtualAgentVersion),
		ReadContext:   provider.ReadWithPooledClient(readAgenticVirtualAgentVersion),
		UpdateContext: provider.UpdateWithPooledClient(updateAgenticVirtualAgentVersion),
		DeleteContext: provider.DeleteWithPooledClient(deleteAgenticVirtualAgentVersion),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"agent_id": {
				Description: "ID of the parent agentic virtual agent.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"version": {
				Description: "Auto-assigned version number (e.g. '1.0'). Set by the API on creation.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"status": {
				Description: "Current status of the version: Draft, TestReady, or ProductionReady.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"definition": {
				Description: "The full definition of the virtual agent version.",
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem:        definitionResource(),
			},
		},
	}
}

func definitionResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"role": {
				Description: "A brief description of the virtual agent's high-level capabilities.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"model": {
				Description:  "The model powering the virtual agent version. Where a new model version is available, Preview can be used to opt in.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"Stable", "Preview"}, false),
			},
			"instructions": {
				Description: "List of instructions, rules, or guidelines the virtual agent must always follow.",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"guardrails": {
				Description: "Custom guardrail rules for the virtual agent.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        guardrailsResource(),
			},
			"tools": {
				Description: "Tools available to the virtual agent.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        toolResource(),
			},
			"types": {
				Description: "Type definitions used by tools for inputs, outputs, and errors.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        typeDefinitionResource(),
			},
			"events": {
				Description: "Event settings for the virtual agent.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        eventSettingsResource(),
			},
			"settings": {
				Description: "Runtime behavior settings for the virtual agent version.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem:        versionSettingsResource(),
			},
		},
	}
}

func guardrailsResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"custom": {
				Description: "Custom guardrail rules.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instruction": {
							Description: "Natural language rule describing user behavior to detect and block.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"enabled": {
							Description: "Whether this custom guardrail rule is active.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
						},
					},
				},
			},
		},
	}
}

func toolResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "Tool type discriminator.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"KnowledgeSetting", "KnowledgeBase", "DataAction", "ExternalA2AServer"}, false),
			},
			"name": {
				Description: "Name of the tool.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Description of how this tool works.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"target": {
				Description: "Resource selected for this tool.",
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "ID of the target resource.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"name": {
							Description: "Name of the target resource.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
						"self_uri": {
							Description: "Self URI of the target resource.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"input_instructions": {
				Description: "Additional instructions specific to using this tool.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"output_instructions": {
				Description: "Instructions that apply after successful tool execution.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        outputInstructionResource(),
			},
			// DataAction-specific fields
			"inputs": {
				Description: "Inputs passed to this data action tool. Only applicable when type is DataAction.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        toolInputResource(),
			},
			"output": {
				Description: "Name of the output type this tool returns. Only applicable when type is DataAction.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"errors": {
				Description: "Error types this tool can raise. Only applicable when type is DataAction.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        toolErrorResource(),
			},
			"input_validation": {
				Description: "Conditions that must be checked before invoking this tool. Only applicable when type is DataAction.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        inputValidationResource(),
			},
		},
	}
}

func outputInstructionResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "Output instruction type discriminator.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Python", "Structured"}, false),
			},
			"then": {
				Description: "Instruction to follow when the condition is met.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"when": {
				Description: "Condition for when this instruction applies. For Python type: a Python expression string. For Structured type: a JSON-encoded structured condition group.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func toolInputResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"target_name": {
				Description: "Name of the input field in the data action input schema.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description: "Input type name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"source": {
				Description:  "Source of the input value.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"User", "ToolInput", "ToolOutput", "External"}, false),
			},
			"required": {
				Description: "Whether this input must be supplied.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"fallback_to_user": {
				Description: "Whether the virtual agent should ask the user for this input when not available from the configured source.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"mapping": {
				Description: "Path used to extract this input from a previous tool output. JSON-encoded array. Only valid when source is ToolOutput.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func toolErrorResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description: "Error type name as defined in the types list.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"instruction": {
				Description: "Instruction for how the virtual agent should handle this error.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func inputValidationResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "Validation type discriminator.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Python", "Structured"}, false),
			},
			"if_condition": {
				Description: "Condition that must evaluate to true before invoking the tool. For Python type: a Python expression. For Structured type: a JSON-encoded structured condition group.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"else_instruction": {
				Description: "Instruction when the validation condition is not met.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func typeDefinitionResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Type name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Additional context about what this type is used for.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"direction": {
				Description:  "Intended direction of use for this type.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Input", "Output", "AgentInput", "AgentOutput"}, false),
			},
			"type": {
				Description:  "Type value.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"string", "integer", "number", "boolean", "null", "object", "array", "DataActionHttpError"}, false),
			},
			"user_utterance_substring": {
				Description: "Whether values of this string type must be copied as a contiguous substring from recent user messages.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"undisclosed": {
				Description: "Whether values of this string type are hidden from the virtual agent.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"properties": {
				Description: "Properties of this object type. Applies when type is object.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        propertyDefinitionResource(),
			},
			"items": {
				Description: "Type of items in this array type. Applies when type is array.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"status_codes": {
				Description: "HTTP status codes this error type can handle. Applies when type is DataActionHttpError.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"default_instruction": {
				Description: "Default instruction for handling this error type. Applies when type is DataActionHttpError.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enum_values": {
				Description: "Allowed enum values.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func propertyDefinitionResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Property name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description: "Property type name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"required": {
				Description: "Whether this property must be supplied.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"description": {
				Description: "Additional context about what this property means.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"items": {
				Description: "Type of items in this array property. Applies when type is array.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"mapping": {
				Description: "Path used to extract this output property from a tool output. JSON-encoded array.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func eventSettingsResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "Event type discriminator.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Guardrails", "UserExit", "Escalation"}, false),
			},
			"message": {
				Description: "Message the virtual agent should return when this event is handled.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"violation_threshold": {
				Description: "Number of guardrail violations allowed before the threshold is crossed. Required when type is Guardrails.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"violation_threshold_crossed_message": {
				Description: "Message when the guardrail violation threshold is crossed. Required when type is Guardrails.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func versionSettingsResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"comfort_statement": {
				Description: "Comfort statement settings for tool calls.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Description: "Whether comfort statements are enabled during eligible tool calls.",
							Type:        schema.TypeBool,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}
