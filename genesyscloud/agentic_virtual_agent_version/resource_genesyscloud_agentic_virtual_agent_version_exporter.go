package agentic_virtual_agent_version

import (
	"fmt"
	"strings"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

const (
	knowledgeKnowledgebaseResourceType = "genesyscloud_knowledge_knowledgebase"
	integrationActionResourceType      = "genesyscloud_integration_action"

	customIntegrationActionIDPrefix = "custom_-_"
	staticIntegrationActionIDPrefix = "static"
)

// resolveToolTargetRefType determines the Terraform resource type for a tool target ID
// during export. The configMap is the target block (id, name, self_uri).
func resolveToolTargetRefType(configMap map[string]interface{}) (string, error) {
	if selfURI, ok := configMap["self_uri"].(string); ok && selfURI != "" {
		switch {
		case strings.Contains(selfURI, "/knowledge/knowledgebases/"):
			return knowledgeKnowledgebaseResourceType, nil
		case strings.Contains(selfURI, "/integrations/actions/"):
			return integrationActionResourceType, nil
		}
	}

	id, _ := configMap["id"].(string)
	if id == "" {
		return "", nil
	}

	if strings.HasPrefix(id, customIntegrationActionIDPrefix) || strings.HasPrefix(id, staticIntegrationActionIDPrefix) {
		return integrationActionResourceType, nil
	}

	return "", nil
}

// resolveAgenticToolTargets resolves any tool target IDs that were not rewritten during
// the standard RefAttr pass (for example when self_uri was omitted from export).
func resolveAgenticToolTargets(configMap map[string]interface{}, exporters map[string]*resourceExporter.ResourceExporter, _ string) error {
	tools, ok := configMap["tools"].([]interface{})
	if !ok || len(tools) == 0 {
		return nil
	}

	for _, toolItem := range tools {
		toolMap, ok := toolItem.(map[string]interface{})
		if !ok {
			continue
		}

		refType := toolTypeToRefType(toolMap["type"])
		if refType == "" {
			continue
		}

		targetList, ok := toolMap["target"].([]interface{})
		if !ok || len(targetList) == 0 {
			continue
		}

		targetMap, ok := targetList[0].(map[string]interface{})
		if !ok {
			continue
		}

		targetID, _ := targetMap["id"].(string)
		if targetID == "" || strings.HasPrefix(targetID, "${") {
			continue
		}

		if resolved := resolveTargetReference(refType, targetID, exporters); resolved != targetID {
			targetMap["id"] = resolved
		}
	}

	return nil
}

func toolTypeToRefType(toolType any) string {
	toolTypeStr, _ := toolType.(string)
	switch toolTypeStr {
	case "KnowledgeBase":
		return knowledgeKnowledgebaseResourceType
	case "DataAction":
		return integrationActionResourceType
	default:
		return ""
	}
}

func resolveTargetReference(refType, targetID string, exporters map[string]*resourceExporter.ResourceExporter) string {
	exporter, ok := exporters[refType]
	if !ok || exporter == nil {
		return targetID
	}

	meta := exporter.SanitizedResourceMap[targetID]
	if meta == nil || meta.BlockLabel == "" {
		return targetID
	}

	return fmt.Sprintf("${%s.%s.id}", refType, meta.BlockLabel)
}
