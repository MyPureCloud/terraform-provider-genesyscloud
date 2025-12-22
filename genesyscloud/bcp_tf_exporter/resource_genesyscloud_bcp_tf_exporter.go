package bcp_tf_exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	dependentconsumers "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/dependent_consumers"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
	lists "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
)

type BcpResource struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Dependencies []string `json:"dependencies"`
}

type BcpExportData map[string][]BcpResource

func createBcpTfExporter(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	directory := d.Get("directory").(string)
	filename := d.Get("filename").(string)

	exporters := resourceExporter.GetResourceExporters()
	filteredExporters := filterExporters(ctx, exporters, d)
	exportData := make(BcpExportData)

	tflog.Info(ctx, "Starting BCP export", map[string]interface{}{
		"directory":            directory,
		"filename":             filename,
		"resource_types_count": len(filteredExporters),
	})

	for resourceType, exporter := range filteredExporters {
		tflog.Debug(ctx, "Processing resource type", map[string]interface{}{
			"resource_type": resourceType,
		})

		err := exporter.LoadSanitizedResourceMap(ctx, resourceType, nil)
		if err != nil {
			tflog.Error(ctx, "Error loading resources", map[string]interface{}{
				"resource_type": resourceType,
				"error":         fmt.Sprintf("%v", err),
			})
			continue
		}

		resourceMap := exporter.GetSanitizedResourceMap()
		var resources []BcpResource

		for id, resMeta := range resourceMap {
			name := resMeta.BlockLabel
			if name == "" {
				name = resMeta.OriginalLabel
			}

			// Get dependencies using the dependent consumers proxy for flows, RefAttrs for others
			deps := getResourceDependencies(ctx, resourceType, id, resMeta, exporter, filteredExporters, meta)

			resources = append(resources, BcpResource{
				ID:           id,
				Name:         name,
				Dependencies: deps,
			})
		}

		if len(resources) > 0 {
			exportData[resourceType] = resources
			tflog.Debug(ctx, "Added resources to export", map[string]interface{}{
				"resource_type":  resourceType,
				"resource_count": len(resources),
			})
		}
	}

	if err := os.MkdirAll(directory, 0755); err != nil {
		return diag.FromErr(err)
	}

	filePath := filepath.Join(directory, filename)
	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return diag.FromErr(err)
	}

	diagErr := files.WriteToFile(jsonData, filePath)
	if diagErr.HasError() {
		return diagErr
	}

	d.SetId(filePath)
	tflog.Info(ctx, "BCP export completed", map[string]interface{}{
		"file_path":            filePath,
		"total_resource_types": len(exportData),
	})
	return nil
}

func readBcpTfExporter(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	path := d.Id()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		d.SetId("")
		return nil
	}
	return nil
}

func deleteBcpTfExporter(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	filePath := d.Id()
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return diag.FromErr(err)
	}
	return nil
}

func filterExporters(ctx context.Context, exporters map[string]*resourceExporter.ResourceExporter, d *schema.ResourceData) map[string]*resourceExporter.ResourceExporter {
	filtered := make(map[string]*resourceExporter.ResourceExporter)

	if includeList, ok := d.GetOk("include_filter_resources"); ok {
		includeTypes := lists.InterfaceListToStrings(includeList.([]interface{}))
		tflog.Debug(ctx, "Applying include filter", map[string]interface{}{
			"include_types": includeTypes,
		})
		for _, resourceType := range includeTypes {
			if exporter, exists := exporters[resourceType]; exists {
				filtered[resourceType] = exporter
			}
		}
		return filtered
	}

	// Start with all exporters
	for k, v := range exporters {
		filtered[k] = v
	}

	if excludeList, ok := d.GetOk("exclude_filter_resources"); ok {
		excludeTypes := lists.InterfaceListToStrings(excludeList.([]interface{}))
		tflog.Debug(ctx, "Applying exclude filter", map[string]interface{}{
			"exclude_types": excludeTypes,
		})
		for _, resourceType := range excludeTypes {
			delete(filtered, resourceType)
		}
	}

	return filtered
}

func getFlowDependencies(ctx context.Context, flowID string, resMeta *resourceExporter.ResourceMeta, meta interface{}) []string {
	// Check if we have proper provider meta for flow dependencies
	providerMeta, ok := meta.(*provider.ProviderMeta)
	if !ok || providerMeta.ClientConfig == nil {
		tflog.Warn(ctx, "No valid client config for flow, skipping dependency resolution", map[string]interface{}{
			"flow_id": flowID,
		})
		return []string{}
	}

	proxy := dependentconsumers.GetDependentConsumerProxy(providerMeta.ClientConfig)
	if proxy == nil {
		tflog.Error(ctx, "Failed to get dependent consumer proxy for flow", map[string]interface{}{
			"flow_id": flowID,
		})
		return []string{}
	}

	// Create a ResourceInfo for the flow
	resourceInfo := resourceExporter.ResourceInfo{
		State:      &terraform.InstanceState{ID: flowID},
		Type:       "genesyscloud_flow",
		BlockLabel: resMeta.BlockLabel,
	}

	// Get dependencies using the proxy
	_, dependsStruct, _, err := proxy.GetDependentConsumers(ctx, resourceInfo, []string{})
	if err != nil {
		tflog.Error(ctx, "Error getting dependencies for flow", map[string]interface{}{
			"flow_id": flowID,
			"error":   err.Error(),
		})
		return []string{}
	}

	if dependsStruct == nil || dependsStruct.DependsMap == nil {
		tflog.Debug(ctx, "No dependency structure returned for flow", map[string]interface{}{
			"flow_id": flowID,
		})
		return []string{}
	}

	// Extract dependencies for this specific flow
	var deps []string
	if flowDeps, ok := dependsStruct.DependsMap[flowID]; ok {
		for _, dep := range flowDeps {
			// Convert from "resourceType.resourceID" to "resourceType::resourceID"
			parts := strings.SplitN(dep, ".", 2)
			if len(parts) == 2 {
				deps = append(deps, fmt.Sprintf("%s::%s", parts[0], parts[1]))
			}
		}
	}

	tflog.Debug(ctx, "Flow dependencies resolved", map[string]interface{}{
		"flow_id":            flowID,
		"dependencies":       deps,
		"dependencies_count": len(deps),
	})
	return deps
}

func getResourceDependencies(ctx context.Context, resourceType, resourceID string, resMeta *resourceExporter.ResourceMeta, exporter *resourceExporter.ResourceExporter, allExporters map[string]*resourceExporter.ResourceExporter, meta interface{}) []string {
	// For flows, use the dependent consumers proxy
	if resourceType == "genesyscloud_flow" {
		return getFlowDependencies(ctx, resourceID, resMeta, meta)
	}

	// For other resources, extract specific dependencies from resource data
	return extractSpecificDependencies(ctx, resourceType, resourceID, resMeta, exporter, allExporters, meta)
}

func extractSpecificDependencies(ctx context.Context, resourceType, resourceID string, resMeta *resourceExporter.ResourceMeta, exporter *resourceExporter.ResourceExporter, allExporters map[string]*resourceExporter.ResourceExporter, meta interface{}) []string {
	var deps []string

	if exporter.RefAttrs == nil || len(exporter.RefAttrs) == 0 {
		return deps
	}

	// Get the resource schema and read the actual resource data
	providerResources, _ := registrar.GetResources()
	resource, exists := providerResources[resourceType]
	if !exists {
		return deps
	}

	// Read the resource state
	instanceState, err := getResourceState(ctx, resource, resourceID, resMeta, meta)
	if err != nil || instanceState == nil {
		return deps
	}

	// Convert instance state to JSON map
	ctyType := resource.CoreConfigSchema().ImpliedType()
	stateVal, err := schema.StateValueFromInstanceState(instanceState, ctyType)
	if err != nil {
		return deps
	}

	resourceData, err := schema.StateValueToJSONMap(stateVal, ctyType)
	if err != nil {
		return deps
	}

	// Extract specific dependencies using RefAttrs
	depSet := make(map[string]bool)
	deps = extractDepsFromMap(resourceData, "", exporter, depSet)

	return deps
}

func extractDepsFromMap(data map[string]interface{}, prefix string, exporter *resourceExporter.ResourceExporter, depSet map[string]bool) []string {
	var deps []string

	for key, val := range data {
		fullPath := key
		if prefix != "" {
			fullPath = prefix + "." + key
		}

		// Check if this attribute has reference settings
		refSettings := exporter.GetRefAttrSettings(fullPath)
		if refSettings == nil {
			// Check wildcard
			wildcardPath := prefix + ".*"
			if prefix == "" {
				wildcardPath = "*"
			}
			refSettings = exporter.GetRefAttrSettings(wildcardPath)
		}

		if refSettings != nil && refSettings.RefType != "" {
			// Extract GUIDs from this value
			guids := extractGUIDsFromValue(val)
			for _, guid := range guids {
				if guid != "" && isValidGUID(guid) {
					depRef := fmt.Sprintf("%s::%s", refSettings.RefType, guid)
					if !depSet[depRef] {
						depSet[depRef] = true
						deps = append(deps, depRef)
					}
				}
			}
		}

		// Recurse into nested structures
		switch v := val.(type) {
		case map[string]interface{}:
			deps = append(deps, extractDepsFromMap(v, fullPath, exporter, depSet)...)
		case []interface{}:
			for _, item := range v {
				if mapItem, ok := item.(map[string]interface{}); ok {
					deps = append(deps, extractDepsFromMap(mapItem, fullPath, exporter, depSet)...)
				}
			}
		}
	}

	return deps
}

func extractGUIDsFromValue(val interface{}) []string {
	switch v := val.(type) {
	case string:
		return []string{v}
	case []interface{}:
		var guids []string
		for _, item := range v {
			if strItem, ok := item.(string); ok {
				guids = append(guids, strItem)
			}
		}
		return guids
	default:
		return []string{}
	}
}

func getResourceState(ctx context.Context, resource *schema.Resource, resID string, resMeta *resourceExporter.ResourceMeta, meta interface{}) (*terraform.InstanceState, error) {
	// Create instance state with ID
	instanceState := &terraform.InstanceState{ID: resMeta.IdPrefix + resID}

	// Create resource data
	resourceData := resource.Data(instanceState)

	// If resource has importer, use it
	if resource.Importer != nil && resource.Importer.StateContext != nil {
		resourceDataArr, err := resource.Importer.StateContext(ctx, resourceData, meta)
		if err != nil {
			return nil, fmt.Errorf("importer failed: %v", err)
		}
		if len(resourceDataArr) > 0 {
			instanceState = resourceDataArr[0].State()
		}
	}

	// Refresh the resource to get current state
	state, err := resource.RefreshWithoutUpgrade(ctx, instanceState, meta)
	if err != nil {
		return nil, fmt.Errorf("refresh failed: %v", err)
	}

	if state == nil || state.ID == "" {
		return nil, nil
	}

	return state, nil
}

func isValidGUID(s string) bool {
	// Basic GUID validation - 36 characters with hyphens in correct positions
	if len(s) != 36 {
		return false
	}
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return false
	}
	return true
}
