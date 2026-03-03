package bcp_tf_exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	dependentconsumers "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/dependent_consumers"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/errors"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
	lists "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
)

type BcpResource struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Dependencies BcpResourceDependency `json:"dependencies"`
}

type BcpResourceDependency struct {
	AsProviderResourceList []string            `json:"as_provider_resource_list"`
	AsObjectMap            map[string][]string `json:"as_object_map"`
}

type BcpExportData map[string][]BcpResource

func createBcpTfExporter(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	directory := d.Get("directory").(string)
	filename := d.Get("filename").(string)
	logPermissionErrorsFilename := d.Get("log_permissions_filename").(string)

	validatedDir, validatedFilename, err := validatePath(directory, filename)
	if err != nil {
		return diag.FromErr(err)
	}

	exporters := resourceExporter.GetResourceExporters()
	filteredExporters := filterExporters(ctx, exporters, d)
	exportData := make(BcpExportData)

	tflog.Info(ctx, "Starting BCP export", map[string]interface{}{
		"directory":            validatedDir,
		"filename":             validatedFilename,
		"resource_types_count": len(filteredExporters),
	})

	for resourceType, exporter := range filteredExporters {
		tflog.Debug(ctx, "Processing resource type", map[string]interface{}{
			"resource_type": resourceType,
		})

		diagErr := exporter.LoadSanitizedResourceMap(ctx, resourceType, nil)
		if diagErr != nil {
			if errors.ContainsPermissionsErrorOnly(diagErr) {
				// Log permission errors but continue processing other resources
				errMsg := fmt.Sprintf("%v", diagErr)

				tflog.Warn(ctx, "Permission denied loading resource", map[string]interface{}{
					"resource_type": resourceType,
					"error":         errMsg,
				})
				if logPermissionErrorsFilename != "" {
					// Log permission errors to file
					jsonLog, _ := json.Marshal(map[string]string{
						"resource_type": resourceType,
						"error":         errMsg,
					})
					// amazonq-ignore-next-line
					files.WriteToFile(jsonLog, logPermissionErrorsFilename)
				}
				continue
			} else {
				tflog.Error(ctx, "Error loading resources", map[string]interface{}{
					"resource_type": resourceType,
					"error":         fmt.Sprintf("%v", diagErr),
				})
				return diagErr
			}
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

	if err := os.MkdirAll(validatedDir, 0755); err != nil {
		return diag.FromErr(err)
	}

	// Paths have already been validated and sanitized
	// amazonq-ignore-next-line
	filePath := filepath.Join(validatedDir, validatedFilename)
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

// Add path validation
func validatePath(dir, filename string) (string, string, error) {
	// Clean and validate directory path
	if strings.Contains(dir, "..") {
		return "", "", fmt.Errorf("directory path contains invalid traversal sequences")
	}

	// Validate filename
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		return "", "", fmt.Errorf("filename contains invalid characters")
	}

	// Safety check for path traversal
	cleanFilePath := filepath.Clean(filename)
	cleanBaseFilePath := filepath.Base(cleanFilePath)
	if cleanBaseFilePath != cleanFilePath {
		return "", "", fmt.Errorf("filename contains invalid path traversal sequences")
	}

	// Explicit base path validation of dir
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", "", fmt.Errorf("invalid directory path: %v", err)
	}

	return absDir, cleanBaseFilePath, nil
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

func getFlowDependencies(ctx context.Context, flowID string, resMeta *resourceExporter.ResourceMeta, meta interface{}) BcpResourceDependency {
	bcpResourceDependencies := BcpResourceDependency{}

	// Check if we have proper provider meta for flow dependencies
	providerMeta, ok := meta.(*provider.ProviderMeta)
	if !ok || providerMeta.ClientConfig == nil {
		tflog.Warn(ctx, "No valid client config for flow, skipping dependency resolution", map[string]interface{}{
			"flow_id": flowID,
		})
		return bcpResourceDependencies
	}

	proxy := dependentconsumers.GetDependentConsumerProxy(providerMeta.ClientConfig)
	if proxy == nil {
		tflog.Error(ctx, "Failed to get dependent consumer proxy for flow", map[string]interface{}{
			"flow_id": flowID,
		})
		return bcpResourceDependencies
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
		return bcpResourceDependencies
	}

	if dependsStruct == nil || dependsStruct.DependsMap == nil {
		tflog.Debug(ctx, "No dependency structure returned for flow", map[string]interface{}{
			"flow_id": flowID,
		})
		return bcpResourceDependencies
	}

	// Extract dependencies for this specific flow
	var depsAsProviderList []string
	depsAsObjectMap := make(map[string][]string)
	if flowDeps, ok := dependsStruct.DependsMap[flowID]; ok {
		for _, dep := range flowDeps {
			// Convert from "resourceType.resourceID" to "resourceType::resourceID"
			parts := strings.SplitN(dep, ".", 2)
			if len(parts) == 2 {
				depsAsProviderList = append(depsAsProviderList, fmt.Sprintf("%s::%s", parts[0], parts[1]))
				if _, exists := depsAsObjectMap[parts[0]]; !exists {
					depsAsObjectMap[parts[0]] = []string{}
				}
				depsAsObjectMap[parts[0]] = append(depsAsObjectMap[parts[0]], parts[1])
			}
		}
	}

	deps := BcpResourceDependency{
		AsProviderResourceList: depsAsProviderList,
		AsObjectMap:            depsAsObjectMap,
	}

	tflog.Debug(ctx, "Flow dependencies resolved", map[string]interface{}{
		"flow_id":            flowID,
		"dependencies":       deps,
		"dependencies_count": len(deps.AsProviderResourceList),
	})
	return deps
}

func getResourceDependencies(ctx context.Context, resourceType, resourceID string, resMeta *resourceExporter.ResourceMeta, exporter *resourceExporter.ResourceExporter, allExporters map[string]*resourceExporter.ResourceExporter, meta interface{}) BcpResourceDependency {
	// For flows, use the dependent consumers proxy
	if resourceType == "genesyscloud_flow" {
		return getFlowDependencies(ctx, resourceID, resMeta, meta)
	}

	// For other resources, extract specific dependencies from resource data
	return extractSpecificDependencies(ctx, resourceType, resourceID, resMeta, exporter, allExporters, meta)
}

func extractSpecificDependencies(ctx context.Context, resourceType, resourceID string, resMeta *resourceExporter.ResourceMeta, exporter *resourceExporter.ResourceExporter, allExporters map[string]*resourceExporter.ResourceExporter, meta interface{}) BcpResourceDependency {
	bcpResourceDependencies := BcpResourceDependency{
		AsProviderResourceList: []string{},
		AsObjectMap:            make(map[string][]string),
	}

	if exporter.RefAttrs == nil || len(exporter.RefAttrs) == 0 {
		return bcpResourceDependencies
	}

	// Get the resource schema and read the actual resource data
	providerResources, _ := registrar.GetResources()
	resource, exists := providerResources[resourceType]
	if !exists {
		return bcpResourceDependencies
	}

	// Read the resource state
	instanceState, err := getResourceState(ctx, resource, resourceID, resMeta, meta)
	if err != nil || instanceState == nil {
		return bcpResourceDependencies
	}

	// Convert instance state to JSON map
	ctyType := resource.CoreConfigSchema().ImpliedType()
	stateVal, err := schema.StateValueFromInstanceState(instanceState, ctyType)
	if err != nil {
		return bcpResourceDependencies
	}

	resourceData, err := schema.StateValueToJSONMap(stateVal, ctyType)
	if err != nil {
		return bcpResourceDependencies
	}

	// Extract specific dependencies using RefAttrs
	depSet := make(map[string]map[string]bool)
	depSet = extractDepsFromMap(resourceData, "", exporter, depSet, 1)

	for refType, guids := range depSet {
		guidList := make([]string, 0, len(guids))
		resourceDepList := make([]string, 0, len(guids))
		for guid := range guids {
			guidList = append(guidList, guid)
			resourceDepList = append(resourceDepList, fmt.Sprintf("%s::%s", refType, guid))
		}
		bcpResourceDependencies.AsObjectMap[refType] = guidList
		bcpResourceDependencies.AsProviderResourceList = append(bcpResourceDependencies.AsProviderResourceList, resourceDepList...)
	}

	return bcpResourceDependencies
}

func extractDepsFromMap(data map[string]interface{}, prefix string, exporter *resourceExporter.ResourceExporter, depSet map[string]map[string]bool, depth int) map[string]map[string]bool {

	// Because this is a recursive function, we need to ensure safety within the recursion so it doesn't get caught in an infinite loop
	const maxDepth = 100
	if depth > maxDepth {
		return depSet
	}

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
					if _, exists := depSet[refSettings.RefType]; !exists {
						depSet[refSettings.RefType] = make(map[string]bool)
					}
					if _, exists := depSet[refSettings.RefType][guid]; !exists {
						depSet[refSettings.RefType][guid] = true
					}
				}
			}
		}

		// Recurse into nested structures
		switch v := val.(type) {
		case map[string]interface{}:
			depSet = extractDepsFromMap(v, fullPath, exporter, depSet, depth+1)
		case []interface{}:
			for _, item := range v {
				if mapItem, ok := item.(map[string]interface{}); ok {
					depSet = extractDepsFromMap(mapItem, fullPath, exporter, depSet, depth+1)
				}
			}
		}
	}

	return depSet
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
