package agentic_virtual_agent_version

import "encoding/json"

/*
   resource_genesyscloud_agentic_virtual_agent_version_utils.go contains Go struct definitions
   for the Agentic Virtual Agent Version API request/response models.

   These structs map to the public API contract defined in AgenticVirtualAgentVersionDefinition
   and all its sub-models (33 total definitions from the swagger).

   Key design decisions:
   - Tool types use a single struct with omitempty for type-specific fields (handles discriminator)
   - Output instructions use a single struct with a flexible `When` field (string or structured)
   - Input validation uses a single struct with a flexible `If` field (string or structured)
   - Computed fields (schemas, skills, target.selfUri) are included for reads but not sent on create/update
*/

// =============================================================================
// Top-Level Version Request/Response
// =============================================================================

// AgenticVirtualAgentVersionResponse represents the full version response from the API.
type AgenticVirtualAgentVersionResponse struct {
	Version               *string                               `json:"version,omitempty"`
	AgenticVirtualAgentId *string                               `json:"agenticVirtualAgentId,omitempty"`
	Status                *string                               `json:"status,omitempty"`
	Definition            *AgenticVirtualAgentVersionDefinition `json:"definition,omitempty"`
	DateCreated           *string                               `json:"dateCreated,omitempty"`
	DateModified          *string                               `json:"dateModified,omitempty"`
	SelfUri               *string                               `json:"selfUri,omitempty"`
}

// AgenticVirtualAgentVersionCreate represents the request body for creating a version.
type AgenticVirtualAgentVersionCreate struct {
	Definition *AgenticVirtualAgentVersionDefinition `json:"definition"`
}

// AgenticVirtualAgentVersionUpdate represents the request body for updating (PATCH) a version.
type AgenticVirtualAgentVersionUpdate struct {
	Definition *AgenticVirtualAgentVersionDefinition `json:"definition"`
}

// =============================================================================
// Version Definition
// =============================================================================

// AgenticVirtualAgentVersionDefinition is the core definition of a virtual agent version.
// Required fields: role, instructions
type AgenticVirtualAgentVersionDefinition struct {
	Role         string                              `json:"role"`
	Instructions []string                            `json:"instructions"`
	Model        string                              `json:"model,omitempty"`
	Guardrails   *AgenticVirtualAgentGuardrails      `json:"guardrails,omitempty"`
	Tools        []AgenticVirtualAgentTool           `json:"tools,omitempty"`
	Types        []AgenticVirtualAgentTypeDefinition `json:"types,omitempty"`
	Events       []AgenticVirtualAgentEventSettings  `json:"events,omitempty"`
	Settings     *AgenticVirtualAgentVersionSettings `json:"settings,omitempty"`
}

// =============================================================================
// Guardrails
// =============================================================================

// AgenticVirtualAgentGuardrails contains custom guardrail rules.
type AgenticVirtualAgentGuardrails struct {
	Custom []AgenticVirtualAgentGuardrailInstruction `json:"custom,omitempty"`
}

// AgenticVirtualAgentGuardrailInstruction is a single custom guardrail rule.
// Required: instruction
type AgenticVirtualAgentGuardrailInstruction struct {
	Instruction string `json:"instruction"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

// =============================================================================
// Tools (Discriminator: type)
// Uses a single struct for all tool types with omitempty for type-specific fields.
// =============================================================================

// AgenticVirtualAgentTool represents any tool type.
// Common fields: type, name, description, target, inputInstructions, outputInstructions
// DataAction-specific: errors, inputs, output, inputValidation, schemas (computed)
// ExternalA2AServer-specific: skills (computed)
type AgenticVirtualAgentTool struct {
	Type               string                                     `json:"type"`
	Name               string                                     `json:"name"`
	Description        string                                     `json:"description"`
	Target             DomainEntityRef                            `json:"target"`
	InputInstructions  []string                                   `json:"inputInstructions,omitempty"`
	OutputInstructions []AgenticVirtualAgentToolOutputInstruction `json:"outputInstructions,omitempty"`
	// DataAction-specific fields
	Errors          []AgenticVirtualAgentToolError        `json:"errors,omitempty"`
	Inputs          []AgenticVirtualAgentToolInput        `json:"inputs,omitempty"`
	Output          string                                `json:"output,omitempty"`
	InputValidation []AgenticVirtualAgentInputValidation  `json:"inputValidation,omitempty"`
	Schemas         *AgenticVirtualAgentDataActionSchemas `json:"schemas,omitempty"`
	// ExternalA2AServer-specific fields (computed at publish time)
	Skills []AgenticVirtualAgentAgentCardSkill `json:"skills,omitempty"`
}

// DomainEntityRef is a reference to a Genesys Cloud resource.
type DomainEntityRef struct {
	Id      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	SelfUri string `json:"selfUri,omitempty"`
}

// =============================================================================
// Tool Output Instructions (Discriminator: type — Python or Structured)
// =============================================================================

// AgenticVirtualAgentToolOutputInstruction represents an output instruction.
// For Python type: When is a string (Python condition).
// For Structured type: When is a structured condition group object.
// We use json.RawMessage to handle both cases during marshal/unmarshal.
type AgenticVirtualAgentToolOutputInstruction struct {
	Type string          `json:"type"`
	Then string          `json:"then"`
	When json.RawMessage `json:"when,omitempty"`
}

// AgenticVirtualAgentStructuredOutputConditionGroup is a group of structured output rules.
type AgenticVirtualAgentStructuredOutputConditionGroup struct {
	Group string                                    `json:"group"`
	Rules []AgenticVirtualAgentStructuredOutputRule `json:"rules"`
}

// AgenticVirtualAgentStructuredOutputRule is a single structured output rule.
type AgenticVirtualAgentStructuredOutputRule struct {
	Mapping  []interface{} `json:"mapping,omitempty"`
	Operator string        `json:"operator,omitempty"`
	Value    interface{}   `json:"value,omitempty"`
	// Nested group support (rules can be groups themselves)
	Group string                                    `json:"group,omitempty"`
	Rules []AgenticVirtualAgentStructuredOutputRule `json:"rules,omitempty"`
}

// =============================================================================
// Tool Input Validation (Discriminator: type — Python or Structured)
// =============================================================================

// AgenticVirtualAgentInputValidation represents input validation for a DataAction tool.
// For Python type: If is a string (Python condition).
// For Structured type: If is a structured condition group object.
type AgenticVirtualAgentInputValidation struct {
	Type string          `json:"type"`
	If   json.RawMessage `json:"if,omitempty"`
	Else string          `json:"else,omitempty"`
}

// AgenticVirtualAgentStructuredConditionGroup is a group of structured input validation rules.
type AgenticVirtualAgentStructuredConditionGroup struct {
	Group string                              `json:"group"`
	Rules []AgenticVirtualAgentStructuredRule `json:"rules"`
}

// AgenticVirtualAgentStructuredRule is a single structured input validation rule.
type AgenticVirtualAgentStructuredRule struct {
	Name     string      `json:"name,omitempty"`
	Operator string      `json:"operator,omitempty"`
	Value    interface{} `json:"value,omitempty"`
	// Nested group support
	Group string                              `json:"group,omitempty"`
	Rules []AgenticVirtualAgentStructuredRule `json:"rules,omitempty"`
}

// =============================================================================
// Tool Inputs (DataAction only)
// =============================================================================

// AgenticVirtualAgentToolInput represents an input for a DataAction tool.
// Required: targetName, type, source
type AgenticVirtualAgentToolInput struct {
	TargetName     string        `json:"targetName"`
	Type           string        `json:"type"`
	Source         string        `json:"source"`
	Required       *bool         `json:"required,omitempty"`
	FallbackToUser *bool         `json:"fallbackToUser,omitempty"`
	Mapping        []interface{} `json:"mapping,omitempty"`
}

// =============================================================================
// Tool Errors (DataAction only)
// =============================================================================

// AgenticVirtualAgentToolError represents error handling for a tool.
// Required: type
type AgenticVirtualAgentToolError struct {
	Type        string `json:"type"`
	Instruction string `json:"instruction,omitempty"`
}

// =============================================================================
// Data Action Schemas (Computed — populated at publish time)
// =============================================================================

// AgenticVirtualAgentDataActionSchemas holds the input/output JSON schemas.
type AgenticVirtualAgentDataActionSchemas struct {
	Inputs  map[string]interface{} `json:"inputs,omitempty"`
	Outputs map[string]interface{} `json:"outputs,omitempty"`
}

// =============================================================================
// A2A Agent Card Skills (Computed — populated at publish time)
// =============================================================================

// AgenticVirtualAgentAgentCardSkill represents an A2A agent card skill.
type AgenticVirtualAgentAgentCardSkill struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Examples    []string `json:"examples,omitempty"`
	InputModes  []string `json:"inputModes,omitempty"`
	OutputModes []string `json:"outputModes,omitempty"`
}

// =============================================================================
// Type Definitions
// =============================================================================

// AgenticVirtualAgentTypeDefinition defines a type used by tools.
// Required: name
type AgenticVirtualAgentTypeDefinition struct {
	Name                   string                                  `json:"name"`
	Description            string                                  `json:"description,omitempty"`
	Direction              string                                  `json:"direction,omitempty"`
	Type                   string                                  `json:"type,omitempty"`
	UserUtteranceSubstring *bool                                   `json:"userUtteranceSubstring,omitempty"`
	Undisclosed            *bool                                   `json:"undisclosed,omitempty"`
	Properties             []AgenticVirtualAgentPropertyDefinition `json:"properties,omitempty"`
	Items                  string                                  `json:"items,omitempty"`
	StatusCodes            []int                                   `json:"statusCodes,omitempty"`
	DefaultInstruction     string                                  `json:"defaultInstruction,omitempty"`
	Enum                   []string                                `json:"enum,omitempty"`
}

// AgenticVirtualAgentPropertyDefinition defines a property of an object type.
// Required: name, type
type AgenticVirtualAgentPropertyDefinition struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Required    *bool         `json:"required,omitempty"`
	Description string        `json:"description,omitempty"`
	Items       string        `json:"items,omitempty"`
	Mapping     []interface{} `json:"mapping,omitempty"`
}

// =============================================================================
// Event Settings (Discriminator: type — Guardrails, UserExit, Escalation)
// Uses a single struct for all event types with omitempty for type-specific fields.
// =============================================================================

// AgenticVirtualAgentEventSettings represents event settings.
// Common: type, message
// Guardrails-specific: violationThreshold (required), violationThresholdCrossedMessage (required)
type AgenticVirtualAgentEventSettings struct {
	Type                             string `json:"type"`
	Message                          string `json:"message,omitempty"`
	ViolationThreshold               *int   `json:"violationThreshold,omitempty"`
	ViolationThresholdCrossedMessage string `json:"violationThresholdCrossedMessage,omitempty"`
}

// =============================================================================
// Version Settings
// =============================================================================

// AgenticVirtualAgentVersionSettings holds runtime behavior settings.
type AgenticVirtualAgentVersionSettings struct {
	ComfortStatement *AgenticVirtualAgentComfortStatementSettings `json:"comfortStatement,omitempty"`
}

// AgenticVirtualAgentComfortStatementSettings controls comfort statements during tool calls.
type AgenticVirtualAgentComfortStatementSettings struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// =============================================================================
// Agent Summary (used by exporter to list agents and discover versions)
// =============================================================================

// AgentSummary is a minimal agent representation used to discover versions for export.
type AgentSummary struct {
	Id                 *string          `json:"id,omitempty"`
	Name               *string          `json:"name,omitempty"`
	LatestSavedVersion *AgentVersionRef `json:"latestSavedVersion,omitempty"`
}

// AgentVersionRef is the version reference on the agent summary.
type AgentVersionRef struct {
	Version *string `json:"version,omitempty"`
	SelfUri *string `json:"selfUri,omitempty"`
}

// AgentSummaryListing is the paginated list response for agents.
type AgentSummaryListing struct {
	Entities   *[]AgentSummary `json:"entities,omitempty"`
	PageSize   *int            `json:"pageSize,omitempty"`
	PageNumber *int            `json:"pageNumber,omitempty"`
	Total      *int            `json:"total,omitempty"`
	PageCount  *int            `json:"pageCount,omitempty"`
}
