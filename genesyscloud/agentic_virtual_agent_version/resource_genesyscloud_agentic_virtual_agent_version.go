package agentic_virtual_agent_version

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
)

/*
   resource_genesyscloud_agentic_virtual_agent_version.go contains the CRUD logic for the
   genesyscloud_agentic_virtual_agent_version resource.

   Key behaviors:
   - Create: POST definition → store composite ID (agentId/versionId)
   - Read: GET → flatten definition into Terraform state, ignoring computed fields
   - Update: PATCH with full definition (resets status to Draft)
   - Delete: No-op (state removal only — no DELETE endpoint exists)
   - ProductionReady versions are immutable — PATCH is rejected by API
*/

// Composite ID format: agentId/versionId
func buildVersionId(agentId, versionId string) string {
	return agentId + "/" + versionId
}

func parseVersionId(id string) (agentId string, versionId string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid version resource ID format: %s (expected agentId/versionId)", id)
	}
	return parts[0], parts[1], nil
}

// createAgenticVirtualAgentVersion creates a new version with the full definition.
func createAgenticVirtualAgentVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAgenticVirtualAgentVersionProxy(sdkConfig)

	agentId := d.Get("agent_id").(string)
	definition := buildDefinitionFromResourceData(d)

	versionReq := &AgenticVirtualAgentVersionCreate{
		Definition: definition,
	}

	log.Printf("Creating Agentic Virtual Agent Version for agent: %s", agentId)

	versionResp, resp, err := proxy.createVersion(ctx, agentId, versionReq)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create agentic virtual agent version for agent %s: %s", agentId, err), resp)
	}

	d.SetId(buildVersionId(agentId, *versionResp.Version))
	log.Printf("Created Agentic Virtual Agent Version: %s", d.Id())

	return readAgenticVirtualAgentVersion(ctx, d, meta)
}

// readAgenticVirtualAgentVersion reads a version by agent ID and version ID.
func readAgenticVirtualAgentVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAgenticVirtualAgentVersionProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceAgenticVirtualAgentVersion(), constants.ConsistencyChecks(), ResourceType)

	agentId, versionId, err := parseVersionId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("Reading Agentic Virtual Agent Version: %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		versionResp, resp, err := proxy.getVersionById(ctx, agentId, versionId)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read version %s: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read version %s: %s", d.Id(), err), resp))
		}

		_ = d.Set("agent_id", agentId)
		_ = d.Set("version", versionResp.Version)
		_ = d.Set("status", versionResp.Status)

		if versionResp.Definition != nil {
			flattenedDef := flattenDefinitionToResourceData(versionResp.Definition)
			_ = d.Set("definition", flattenedDef)
		}

		log.Printf("Read Agentic Virtual Agent Version: %s", d.Id())
		return cc.CheckState(d)
	})
}

// updateAgenticVirtualAgentVersion updates a version's definition (PATCH — full replacement).
func updateAgenticVirtualAgentVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAgenticVirtualAgentVersionProxy(sdkConfig)

	agentId, versionId, err := parseVersionId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	definition := buildDefinitionFromResourceData(d)

	versionReq := &AgenticVirtualAgentVersionUpdate{
		Definition: definition,
	}

	log.Printf("Updating Agentic Virtual Agent Version: %s", d.Id())

	_, resp, updateErr := proxy.updateVersion(ctx, agentId, versionId, versionReq)
	if updateErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update version %s: %s", d.Id(), updateErr), resp)
	}

	log.Printf("Updated Agentic Virtual Agent Version: %s", d.Id())
	return readAgenticVirtualAgentVersion(ctx, d, meta)
}

// deleteAgenticVirtualAgentVersion is a no-op — no DELETE endpoint exists for versions.
func deleteAgenticVirtualAgentVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Removing Agentic Virtual Agent Version from state: %s (no DELETE endpoint exists)", d.Id())
	return nil
}

// =============================================================================
// Build definition from Terraform ResourceData → Go struct (for create/update)
// =============================================================================

func buildDefinitionFromResourceData(d *schema.ResourceData) *AgenticVirtualAgentVersionDefinition {
	defList := d.Get("definition").([]interface{})
	if len(defList) == 0 {
		return nil
	}
	defMap := defList[0].(map[string]interface{})

	def := &AgenticVirtualAgentVersionDefinition{
		Role:         defMap["role"].(string),
		Instructions: expandStringList(defMap["instructions"].([]interface{})),
	}

	// Model
	if v, ok := defMap["model"].(string); ok && v != "" {
		def.Model = v
	}

	// Guardrails
	if v, ok := defMap["guardrails"].([]interface{}); ok && len(v) > 0 {
		def.Guardrails = expandGuardrails(v[0].(map[string]interface{}))
	}

	// Tools
	if v, ok := defMap["tools"].([]interface{}); ok && len(v) > 0 {
		def.Tools = expandTools(v)
	}

	// Types
	if v, ok := defMap["types"].([]interface{}); ok && len(v) > 0 {
		def.Types = expandTypes(v)
	}

	// Events
	if v, ok := defMap["events"].([]interface{}); ok && len(v) > 0 {
		def.Events = expandEvents(v)
	}

	// Settings
	if v, ok := defMap["settings"].([]interface{}); ok && len(v) > 0 {
		def.Settings = expandSettings(v[0].(map[string]interface{}))
	}

	return def
}

func expandGuardrails(m map[string]interface{}) *AgenticVirtualAgentGuardrails {
	guardrails := &AgenticVirtualAgentGuardrails{}
	if customList, ok := m["custom"].([]interface{}); ok {
		for _, item := range customList {
			customMap := item.(map[string]interface{})
			instruction := AgenticVirtualAgentGuardrailInstruction{
				Instruction: customMap["instruction"].(string),
			}
			if v, ok := customMap["enabled"].(bool); ok {
				instruction.Enabled = &v
			}
			guardrails.Custom = append(guardrails.Custom, instruction)
		}
	}
	return guardrails
}

func expandTools(toolsList []interface{}) []AgenticVirtualAgentTool {
	tools := make([]AgenticVirtualAgentTool, 0, len(toolsList))
	for _, item := range toolsList {
		toolMap := item.(map[string]interface{})
		tool := AgenticVirtualAgentTool{
			Type:        toolMap["type"].(string),
			Name:        toolMap["name"].(string),
			Description: toolMap["description"].(string),
		}

		// Target
		if targetList, ok := toolMap["target"].([]interface{}); ok && len(targetList) > 0 {
			targetMap := targetList[0].(map[string]interface{})
			tool.Target = DomainEntityRef{
				Id:   targetMap["id"].(string),
				Name: targetMap["name"].(string),
			}
		}

		// Input instructions
		if v, ok := toolMap["input_instructions"].([]interface{}); ok && len(v) > 0 {
			tool.InputInstructions = expandStringList(v)
		}

		// Output instructions
		if v, ok := toolMap["output_instructions"].([]interface{}); ok && len(v) > 0 {
			tool.OutputInstructions = expandOutputInstructions(v)
		}

		// DataAction-specific: inputs
		if v, ok := toolMap["inputs"].([]interface{}); ok && len(v) > 0 {
			tool.Inputs = expandToolInputs(v)
		}

		// DataAction-specific: output
		if v, ok := toolMap["output"].(string); ok && v != "" {
			tool.Output = v
		}

		// DataAction-specific: errors
		if v, ok := toolMap["errors"].([]interface{}); ok && len(v) > 0 {
			tool.Errors = expandToolErrors(v)
		}

		// DataAction-specific: input_validation
		if v, ok := toolMap["input_validation"].([]interface{}); ok && len(v) > 0 {
			tool.InputValidation = expandInputValidation(v)
		}

		tools = append(tools, tool)
	}
	return tools
}

func expandOutputInstructions(list []interface{}) []AgenticVirtualAgentToolOutputInstruction {
	instructions := make([]AgenticVirtualAgentToolOutputInstruction, 0, len(list))
	for _, item := range list {
		m := item.(map[string]interface{})
		instruction := AgenticVirtualAgentToolOutputInstruction{
			Type: m["type"].(string),
			Then: m["then"].(string),
		}
		if whenStr, ok := m["when"].(string); ok && whenStr != "" {
			if m["type"].(string) == "Python" {
				// Python: when is a plain string, marshal as JSON string
				whenBytes, _ := json.Marshal(whenStr)
				instruction.When = whenBytes
			} else {
				// Structured: when is a JSON object, pass through as raw JSON
				instruction.When = json.RawMessage(whenStr)
			}
		}
		instructions = append(instructions, instruction)
	}
	return instructions
}

func expandToolInputs(list []interface{}) []AgenticVirtualAgentToolInput {
	inputs := make([]AgenticVirtualAgentToolInput, 0, len(list))
	for _, item := range list {
		m := item.(map[string]interface{})
		input := AgenticVirtualAgentToolInput{
			TargetName: m["target_name"].(string),
			Type:       m["type"].(string),
			Source:     m["source"].(string),
		}
		if v, ok := m["required"].(bool); ok {
			input.Required = &v
		}
		if v, ok := m["fallback_to_user"].(bool); ok {
			input.FallbackToUser = &v
		}
		if v, ok := m["mapping"].(string); ok && v != "" {
			var mapping []interface{}
			_ = json.Unmarshal([]byte(v), &mapping)
			input.Mapping = mapping
		}
		inputs = append(inputs, input)
	}
	return inputs
}

func expandToolErrors(list []interface{}) []AgenticVirtualAgentToolError {
	errors := make([]AgenticVirtualAgentToolError, 0, len(list))
	for _, item := range list {
		m := item.(map[string]interface{})
		toolError := AgenticVirtualAgentToolError{
			Type: m["type"].(string),
		}
		if v, ok := m["instruction"].(string); ok && v != "" {
			toolError.Instruction = v
		}
		errors = append(errors, toolError)
	}
	return errors
}

func expandInputValidation(list []interface{}) []AgenticVirtualAgentInputValidation {
	validations := make([]AgenticVirtualAgentInputValidation, 0, len(list))
	for _, item := range list {
		m := item.(map[string]interface{})
		validation := AgenticVirtualAgentInputValidation{
			Type: m["type"].(string),
			Else: m["else_instruction"].(string),
		}
		if ifStr, ok := m["if_condition"].(string); ok && ifStr != "" {
			if m["type"].(string) == "Python" {
				ifBytes, _ := json.Marshal(ifStr)
				validation.If = ifBytes
			} else {
				validation.If = json.RawMessage(ifStr)
			}
		}
		validations = append(validations, validation)
	}
	return validations
}

func expandTypes(list []interface{}) []AgenticVirtualAgentTypeDefinition {
	types := make([]AgenticVirtualAgentTypeDefinition, 0, len(list))
	for _, item := range list {
		m := item.(map[string]interface{})
		typeDef := AgenticVirtualAgentTypeDefinition{
			Name: m["name"].(string),
		}
		if v, ok := m["description"].(string); ok && v != "" {
			typeDef.Description = v
		}
		if v, ok := m["direction"].(string); ok && v != "" {
			typeDef.Direction = v
		}
		if v, ok := m["type"].(string); ok && v != "" {
			typeDef.Type = v
		}
		if v, ok := m["user_utterance_substring"].(bool); ok && v {
			typeDef.UserUtteranceSubstring = &v
		}
		if v, ok := m["undisclosed"].(bool); ok && v {
			typeDef.Undisclosed = &v
		}
		if v, ok := m["items"].(string); ok && v != "" {
			typeDef.Items = v
		}
		if v, ok := m["default_instruction"].(string); ok && v != "" {
			typeDef.DefaultInstruction = v
		}
		if v, ok := m["status_codes"].([]interface{}); ok && len(v) > 0 {
			typeDef.StatusCodes = expandIntList(v)
		}
		if v, ok := m["enum_values"].([]interface{}); ok && len(v) > 0 {
			typeDef.Enum = expandStringList(v)
		}
		if v, ok := m["properties"].([]interface{}); ok && len(v) > 0 {
			typeDef.Properties = expandProperties(v)
		}
		types = append(types, typeDef)
	}
	return types
}

func expandProperties(list []interface{}) []AgenticVirtualAgentPropertyDefinition {
	props := make([]AgenticVirtualAgentPropertyDefinition, 0, len(list))
	for _, item := range list {
		m := item.(map[string]interface{})
		prop := AgenticVirtualAgentPropertyDefinition{
			Name: m["name"].(string),
			Type: m["type"].(string),
		}
		if v, ok := m["required"].(bool); ok {
			prop.Required = &v
		}
		if v, ok := m["description"].(string); ok && v != "" {
			prop.Description = v
		}
		if v, ok := m["items"].(string); ok && v != "" {
			prop.Items = v
		}
		if v, ok := m["mapping"].(string); ok && v != "" {
			var mapping []interface{}
			_ = json.Unmarshal([]byte(v), &mapping)
			prop.Mapping = mapping
		}
		props = append(props, prop)
	}
	return props
}

func expandEvents(list []interface{}) []AgenticVirtualAgentEventSettings {
	events := make([]AgenticVirtualAgentEventSettings, 0, len(list))
	for _, item := range list {
		m := item.(map[string]interface{})
		event := AgenticVirtualAgentEventSettings{
			Type: m["type"].(string),
		}
		if v, ok := m["message"].(string); ok && v != "" {
			event.Message = v
		}
		if v, ok := m["violation_threshold"].(int); ok && v > 0 {
			event.ViolationThreshold = &v
		}
		if v, ok := m["violation_threshold_crossed_message"].(string); ok && v != "" {
			event.ViolationThresholdCrossedMessage = v
		}
		events = append(events, event)
	}
	return events
}

func expandSettings(m map[string]interface{}) *AgenticVirtualAgentVersionSettings {
	settings := &AgenticVirtualAgentVersionSettings{}
	if csList, ok := m["comfort_statement"].([]interface{}); ok && len(csList) > 0 {
		csMap := csList[0].(map[string]interface{})
		cs := &AgenticVirtualAgentComfortStatementSettings{}
		if v, ok := csMap["enabled"].(bool); ok {
			cs.Enabled = &v
		}
		settings.ComfortStatement = cs
	}
	return settings
}

// =============================================================================
// Flatten definition from Go struct → Terraform ResourceData (for read)
// =============================================================================

func flattenDefinitionToResourceData(def *AgenticVirtualAgentVersionDefinition) []interface{} {
	defMap := map[string]interface{}{
		"role":         def.Role,
		"instructions": def.Instructions,
	}

	if def.Model != "" {
		defMap["model"] = def.Model
	}

	if def.Guardrails != nil {
		defMap["guardrails"] = flattenGuardrails(def.Guardrails)
	}
	if len(def.Tools) > 0 {
		defMap["tools"] = flattenTools(def.Tools)
	}
	if len(def.Types) > 0 {
		defMap["types"] = flattenTypes(def.Types)
	}
	if len(def.Events) > 0 {
		defMap["events"] = flattenEvents(def.Events)
	}
	if def.Settings != nil {
		defMap["settings"] = flattenSettings(def.Settings)
	}

	return []interface{}{defMap}
}

func flattenGuardrails(g *AgenticVirtualAgentGuardrails) []interface{} {
	guardrailMap := map[string]interface{}{}
	if len(g.Custom) > 0 {
		customList := make([]interface{}, 0, len(g.Custom))
		for _, c := range g.Custom {
			m := map[string]interface{}{
				"instruction": c.Instruction,
			}
			if c.Enabled != nil {
				m["enabled"] = *c.Enabled
			} else {
				m["enabled"] = true
			}
			customList = append(customList, m)
		}
		guardrailMap["custom"] = customList
	}
	return []interface{}{guardrailMap}
}

func flattenTools(tools []AgenticVirtualAgentTool) []interface{} {
	result := make([]interface{}, 0, len(tools))
	for _, tool := range tools {
		m := map[string]interface{}{
			"type":        tool.Type,
			"name":        tool.Name,
			"description": tool.Description,
			"target": []interface{}{
				map[string]interface{}{
					"id":       tool.Target.Id,
					"name":     tool.Target.Name,
					"self_uri": tool.Target.SelfUri,
				},
			},
		}

		if len(tool.InputInstructions) > 0 {
			m["input_instructions"] = tool.InputInstructions
		}
		if len(tool.OutputInstructions) > 0 {
			m["output_instructions"] = flattenOutputInstructions(tool.OutputInstructions)
		}
		if len(tool.Inputs) > 0 {
			m["inputs"] = flattenToolInputs(tool.Inputs)
		}
		if tool.Output != "" {
			m["output"] = tool.Output
		}
		if len(tool.Errors) > 0 {
			m["errors"] = flattenToolErrors(tool.Errors)
		}
		if len(tool.InputValidation) > 0 {
			m["input_validation"] = flattenInputValidation(tool.InputValidation)
		}

		result = append(result, m)
	}
	return result
}

func flattenOutputInstructions(instructions []AgenticVirtualAgentToolOutputInstruction) []interface{} {
	result := make([]interface{}, 0, len(instructions))
	for _, instr := range instructions {
		m := map[string]interface{}{
			"type": instr.Type,
			"then": instr.Then,
		}
		if instr.When != nil {
			if instr.Type == "Python" {
				// Unmarshal the JSON string to get the raw Python expression
				var whenStr string
				if err := json.Unmarshal(instr.When, &whenStr); err == nil {
					m["when"] = whenStr
				} else {
					m["when"] = string(instr.When)
				}
			} else {
				// Structured: store as JSON string
				m["when"] = string(instr.When)
			}
		}
		result = append(result, m)
	}
	return result
}

func flattenToolInputs(inputs []AgenticVirtualAgentToolInput) []interface{} {
	result := make([]interface{}, 0, len(inputs))
	for _, input := range inputs {
		m := map[string]interface{}{
			"target_name": input.TargetName,
			"type":        input.Type,
			"source":      input.Source,
		}
		if input.Required != nil {
			m["required"] = *input.Required
		}
		if input.FallbackToUser != nil {
			m["fallback_to_user"] = *input.FallbackToUser
		}
		if len(input.Mapping) > 0 {
			mappingBytes, _ := json.Marshal(input.Mapping)
			m["mapping"] = string(mappingBytes)
		}
		result = append(result, m)
	}
	return result
}

func flattenToolErrors(errors []AgenticVirtualAgentToolError) []interface{} {
	result := make([]interface{}, 0, len(errors))
	for _, e := range errors {
		m := map[string]interface{}{
			"type": e.Type,
		}
		if e.Instruction != "" {
			m["instruction"] = e.Instruction
		}
		result = append(result, m)
	}
	return result
}

func flattenInputValidation(validations []AgenticVirtualAgentInputValidation) []interface{} {
	result := make([]interface{}, 0, len(validations))
	for _, v := range validations {
		m := map[string]interface{}{
			"type":             v.Type,
			"else_instruction": v.Else,
		}
		if v.If != nil {
			if v.Type == "Python" {
				var ifStr string
				if err := json.Unmarshal(v.If, &ifStr); err == nil {
					m["if_condition"] = ifStr
				} else {
					m["if_condition"] = string(v.If)
				}
			} else {
				m["if_condition"] = string(v.If)
			}
		}
		result = append(result, m)
	}
	return result
}

func flattenTypes(types []AgenticVirtualAgentTypeDefinition) []interface{} {
	result := make([]interface{}, 0, len(types))
	for _, t := range types {
		m := map[string]interface{}{
			"name": t.Name,
		}
		if t.Description != "" {
			m["description"] = t.Description
		}
		if t.Direction != "" {
			m["direction"] = t.Direction
		}
		if t.Type != "" {
			m["type"] = t.Type
		}
		if t.UserUtteranceSubstring != nil && *t.UserUtteranceSubstring {
			m["user_utterance_substring"] = true
		}
		if t.Undisclosed != nil && *t.Undisclosed {
			m["undisclosed"] = true
		}
		if t.Items != "" {
			m["items"] = t.Items
		}
		if t.DefaultInstruction != "" {
			m["default_instruction"] = t.DefaultInstruction
		}
		if len(t.StatusCodes) > 0 {
			m["status_codes"] = t.StatusCodes
		}
		if len(t.Enum) > 0 {
			m["enum_values"] = t.Enum
		}
		if len(t.Properties) > 0 {
			m["properties"] = flattenProperties(t.Properties)
		}
		result = append(result, m)
	}
	return result
}

func flattenProperties(props []AgenticVirtualAgentPropertyDefinition) []interface{} {
	result := make([]interface{}, 0, len(props))
	for _, p := range props {
		m := map[string]interface{}{
			"name": p.Name,
			"type": p.Type,
		}
		if p.Required != nil {
			m["required"] = *p.Required
		}
		if p.Description != "" {
			m["description"] = p.Description
		}
		if p.Items != "" {
			m["items"] = p.Items
		}
		if len(p.Mapping) > 0 {
			mappingBytes, _ := json.Marshal(p.Mapping)
			m["mapping"] = string(mappingBytes)
		}
		result = append(result, m)
	}
	return result
}

func flattenEvents(events []AgenticVirtualAgentEventSettings) []interface{} {
	result := make([]interface{}, 0, len(events))
	for _, e := range events {
		m := map[string]interface{}{
			"type": e.Type,
		}
		if e.Message != "" {
			m["message"] = e.Message
		}
		if e.ViolationThreshold != nil {
			m["violation_threshold"] = *e.ViolationThreshold
		}
		if e.ViolationThresholdCrossedMessage != "" {
			m["violation_threshold_crossed_message"] = e.ViolationThresholdCrossedMessage
		}
		result = append(result, m)
	}
	return result
}

func flattenSettings(s *AgenticVirtualAgentVersionSettings) []interface{} {
	settingsMap := map[string]interface{}{}
	if s.ComfortStatement != nil {
		csMap := map[string]interface{}{}
		if s.ComfortStatement.Enabled != nil {
			csMap["enabled"] = *s.ComfortStatement.Enabled
		}
		settingsMap["comfort_statement"] = []interface{}{csMap}
	}
	return []interface{}{settingsMap}
}

// =============================================================================
// Helper functions
// =============================================================================

func expandStringList(items []interface{}) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func expandIntList(items []interface{}) []int {
	result := make([]int, 0, len(items))
	for _, item := range items {
		if i, ok := item.(int); ok {
			result = append(result, i)
		}
	}
	return result
}
