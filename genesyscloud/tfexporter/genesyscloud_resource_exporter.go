package tfexporter

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"maps"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	architectFlow "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	dependentconsumers "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/dependent_consumers"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	rRegistrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	featureToggles "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/stringmap"

	"github.com/google/uuid"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mohae/deepcopy"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

/*
   This file contains all logic associated with the process of exporting a file.
*/

// Used to store the TF config block as a string so that it can be ignored when testing the exported HCL config file.
var (
	terraformHCLBlock string
	mockError         diag.Diagnostics

	// UID : "jsonencode({decoded representation of json string})"
	// Attrs which can be exported in jsonencode objects are populated with a UID
	// The same UID is stored as the key in attributesDecoded, with the value being the jsonencode representation of the json string.
	// When the bytes are being written to the file, the UID is found and replaced with the unquoted jsonencode object
	attributesDecoded = make(map[string]string)

	providerDataSources map[string]*schema.Resource
	providerResources   map[string]*schema.Resource
	resourceExporters   map[string]*resourceExporter.ResourceExporter
)

type unresolvableAttributeInfo struct {
	ResourceType  string
	ResourceLabel string
	Name          string
	Schema        *schema.Schema
}

const (
	formatHCL     = "hcl"
	formatJSON    = "json"
	formatJSONHCL = "json_hcl"
	formatHCLJSON = "hcl_json"
)

type ResourceErrorInfo struct {
	ErrorMessage  string `json:"error_message"`
	ResourceID    string `json:"resource_id"`
	ResourceLabel string `json:"resource_label"`
	ResourceType  string `json:"resource_type"`
	IsTimeout     bool   `json:"is_timeout"`
}

type GenesysCloudResourceExporter struct {
	// Ordering of objects in a struct is important with regards to 32-bit systems.
	// To align padding, put complex objects at the top, grouping like objects together

	// 8-byte alignment
	// .. Pointers and reference types
	buildSecondDeps     map[string][]string
	configExporter      Exporter
	ctx                 context.Context
	cyclicDependsList   []string
	d                   *schema.ResourceData
	dataSourceTypesMaps map[string]ResourceJSONMaps
	dependsList         map[string][]string
	exporters           *map[string]*resourceExporter.ResourceExporter
	filterList          *[]string
	filterType          ExporterFilterType
	flowResourcesList   []string

	meta                  interface{}
	provider              *schema.Provider
	replaceWithDatasource []string
	resources             []resourceExporter.ResourceInfo
	resourceErrors        map[string][]ResourceErrorInfo
	resourceFilter        ExporterResourceFilter
	resourceTypeFilter    ExporterResourceTypeFilter
	resourceTypesMaps     map[string]ResourceJSONMaps
	unresolvedAttrs       []unresolvableAttributeInfo

	// .. Strings
	exportDirPath    string
	exportFormat     string
	providerRegistry string
	version          string

	// 4-byte alignment
	// .. Mutex
	buildSecondDepsMutex       sync.RWMutex
	dataSourceTypesMapsMutex   sync.RWMutex
	exMutex                    sync.RWMutex
	replaceWithDatasourceMutex sync.Mutex
	resourceErrorsMutex        sync.RWMutex
	resourcesMutex             sync.Mutex
	resourceStateMutex         sync.Mutex
	resourceTypesMapsMutex     sync.RWMutex
	unresolvedAttrsMutex       sync.Mutex

	// .. Int
	maxConcurrentOps int // New field to control concurrency

	// 1-byte alignment
	// .. Booleans
	addDependsOn         bool
	exportComputed       bool
	ignoreCyclicDeps     bool
	includeStateFile     bool
	logPermissionErrors  bool
	splitFilesByResource bool
}

func configureExporterType(ctx context.Context, d *schema.ResourceData, gre *GenesysCloudResourceExporter, filterType ExporterFilterType) {
	switch filterType {
	case LegacyInclude:
		var filter []string
		if resourceTypes, ok := d.GetOk("resource_types"); ok {
			filter = lists.InterfaceListToStrings(resourceTypes.([]interface{}))
			gre.filterList = &filter
		}

		//Setting up the resource type filter
		gre.resourceTypeFilter = IncludeFilterByResourceType //Setting up the resource type filter
		gre.resourceFilter = FilterResourceByLabel           //Setting up the resource filters
	case IncludeResources:
		var filter []string
		if resourceTypes, ok := d.GetOk("include_filter_resources"); ok {
			filter = lists.InterfaceListToStrings(resourceTypes.([]interface{}))
			gre.filterList = &filter
		}

		//Setting up the resource type filter
		gre.resourceTypeFilter = IncludeFilterByResourceType //Setting up the resource type filter
		gre.resourceFilter = IncludeFilterResourceByRegex    //Setting up the resource filters
	case ExcludeResources:
		var filter []string
		if resourceTypes, ok := d.GetOk("exclude_filter_resources"); ok {
			filter = lists.InterfaceListToStrings(resourceTypes.([]interface{}))
			gre.filterList = &filter
		}

		//Setting up the resource type filter
		gre.resourceTypeFilter = ExcludeFilterByResourceType //Setting up the resource type filter
		gre.resourceFilter = ExcludeFilterResourceByRegex    //Setting up the resource filters
	}
}

func NewGenesysCloudResourceExporter(ctx context.Context, d *schema.ResourceData, meta interface{}, filterType ExporterFilterType) (*GenesysCloudResourceExporter, diag.Diagnostics) {
	if providerResources == nil {
		providerResources, providerDataSources = rRegistrar.GetResources()
	}
	gre := &GenesysCloudResourceExporter{
		exportFormat:         identifyExportFormat(d),
		splitFilesByResource: d.Get("split_files_by_resource").(bool),
		logPermissionErrors:  d.Get("log_permission_errors").(bool),
		exportComputed:       d.Get("export_computed").(bool),
		addDependsOn:         computeDependsOn(d),
		filterType:           filterType,
		includeStateFile:     d.Get("include_state_file").(bool),
		ignoreCyclicDeps:     d.Get("ignore_cyclic_deps").(bool),
		version:              meta.(*provider.ProviderMeta).Version,
		providerRegistry:     meta.(*provider.ProviderMeta).Registry,
		provider:             provider.New(meta.(*provider.ProviderMeta).Version, providerResources, providerDataSources)(),
		d:                    d,
		ctx:                  ctx,
		meta:                 meta,
		maxConcurrentOps:     10, // Default to 10 concurrent operations
	}

	// Set max concurrent operations based on provider configuration if available
	if providerMeta, ok := meta.(*provider.ProviderMeta); ok && providerMeta.MaxClients > 0 {
		gre.maxConcurrentOps = providerMeta.MaxClients
	}

	err := gre.setUpExportDirPath()
	if err != nil {
		return nil, err
	}

	gre.setupDataSource()

	//Setting up the filter
	configureExporterType(ctx, d, gre, filterType)
	return gre, nil
}

// NewThreadSafeGenesysCloudResourceExporter creates a new exporter with thread-safe features
func NewThreadSafeGenesysCloudResourceExporter(d *schema.ResourceData, ctx context.Context, meta interface{}, provider *schema.Provider, exporters *map[string]*resourceExporter.ResourceExporter) *GenesysCloudResourceExporter {
	exporter := &GenesysCloudResourceExporter{
		configExporter:        nil,                         // Will be set later based on export format
		filterType:            LegacyInclude,               // Default value
		resourceTypeFilter:    IncludeFilterByResourceType, // Default value
		resourceFilter:        FilterResourceByLabel,       // Default value
		filterList:            &[]string{},
		exportFormat:          d.Get("export_format").(string),
		splitFilesByResource:  d.Get("split_files_by_resource").(bool),
		logPermissionErrors:   d.Get("log_permission_errors").(bool),
		addDependsOn:          d.Get("add_depends_on").(bool),
		replaceWithDatasource: []string{},
		includeStateFile:      d.Get("include_state_file").(bool),
		version:               d.Get("version").(string),
		providerRegistry:      d.Get("provider_registry").(string),
		provider:              provider,
		exportDirPath:         d.Get("export_dir_path").(string),
		exporters:             exporters,
		resources:             []resourceExporter.ResourceInfo{},
		resourceTypesMaps:     make(map[string]ResourceJSONMaps),
		dataSourceTypesMaps:   make(map[string]ResourceJSONMaps),
		unresolvedAttrs:       []unresolvableAttributeInfo{},
		d:                     d,
		ctx:                   ctx,
		meta:                  meta,
		dependsList:           make(map[string][]string),
		buildSecondDeps:       make(map[string][]string),
		exMutex:               sync.RWMutex{},
		cyclicDependsList:     []string{},
		ignoreCyclicDeps:      d.Get("ignore_cyclic_dependencies").(bool),
		flowResourcesList:     []string{},
		exportComputed:        d.Get("export_computed").(bool),
		maxConcurrentOps:      10, // Default to 10 concurrent operations
	}

	// Set max concurrent operations based on configuration if available
	if maxClients, ok := d.GetOk("max_concurrent_operations"); ok {
		exporter.maxConcurrentOps = maxClients.(int)
	}

	return exporter
}

func identifyExportFormat(d *schema.ResourceData) string {
	if d.Get("export_as_hcl").(bool) {
		return formatHCL
	}
	return strings.ToLower(d.Get("export_format").(string))
}
func computeDependsOn(d *schema.ResourceData) bool {
	addDependsOn := d.Get("enable_dependency_resolution").(bool)
	if addDependsOn {
		if exportableResourceTypes, ok := d.GetOk("include_filter_resources"); ok {
			filter := lists.InterfaceListToStrings(exportableResourceTypes.([]interface{}))
			addDependsOn = len(filter) > 0
		} else {
			addDependsOn = false
		}
	}
	return addDependsOn
}

func (g *GenesysCloudResourceExporter) Export() (diagErr diag.Diagnostics) {
	// Step #1 Retrieve the exporters we are have registered and have been requested by the user
	tflog.Info(g.ctx, "Retrieving exporters")
	diagErr = append(diagErr, g.retrieveExporters()...)
	if diagErr.HasError() {
		tflog.Error(g.ctx, fmt.Sprintf("Failed to retrieve exporters: %v", diagErr))
		return diagErr
	}
	// Step #2 Retrieve all the individual resources we are going to export
	diagErr = append(diagErr, g.retrieveSanitizedResourceMaps()...)
	if diagErr.HasError() {
		return diagErr
	}

	// Step #3 Retrieve the individual genesys cloud object instances
	diagErr = append(diagErr, g.retrieveGenesysCloudObjectInstances()...)
	if diagErr.HasError() {
		return diagErr
	}

	// Step #4 export dependent resources for the flows
	diagErr = append(diagErr, g.buildAndExportDependsOnResourcesForFlows()...)
	if diagErr.HasError() {
		return diagErr
	}

	// Step #5 Convert the Genesys Cloud resources to neutral format (e.g. map of maps)
	diagErr = append(diagErr, g.buildResourceConfigMap()...)
	if diagErr.HasError() {
		return diagErr
	}

	// Step #6 export dependents for other resources
	diagErr = append(diagErr, g.buildAndExportDependentResources()...)
	if diagErr.HasError() {
		return diagErr
	}

	// Step #7 Write the terraform state file along with either the HCL or JSON
	diagErr = append(diagErr, g.generateOutputFiles()...)
	if diagErr.HasError() {
		return diagErr
	}

	// step #8 Verify the terraform state file with Exporter Resources
	diagErr = append(diagErr, g.verifyTerraformState()...)

	// step #9 Report any resources that errored
	if len(g.resourceErrors) > 0 {
		var timeoutErrorsTotalLen, otherErrorsTotalLen int
		var errorSummary strings.Builder
		errorSummary.WriteString("WARNING: Some resources encountered errors during export:\n")

		for resType, errors := range g.resourceErrors {
			var timeoutCount, otherCount int
			errorSummary.WriteString(fmt.Sprintf("  %s: %d resources\n", resType, len(errors)))

			// Count timeouts vs other errors
			for _, errorInfo := range errors {
				if errorInfo.IsTimeout {
					timeoutCount++
					timeoutErrorsTotalLen++
				} else {
					otherCount++
					otherErrorsTotalLen++
				}
			}

			errorSummary.WriteString(fmt.Sprintf("  - Timeout errors: %d\n", timeoutCount))
			errorSummary.WriteString(fmt.Sprintf("  - Other errors: %d\n", otherCount))

			// List individual errors
			for _, errorInfo := range errors {
				errorType := "Error"
				if errorInfo.IsTimeout {
					errorType = "Timeout"
				}
				errorSummary.WriteString(fmt.Sprintf("    - %s: %s (%s) - %v\n",
					errorType, errorInfo.ResourceLabel, errorInfo.ResourceID, errorInfo.ErrorMessage))
			}
		}

		tflog.Warn(g.ctx, errorSummary.String())

		// Write JSON error data
		jsonFilepath := filepath.Join(g.exportDirPath, "export_errors.json")
		jsonData, err := json.MarshalIndent(g.resourceErrors, "", "  ")
		if err == nil {
			diagErr = append(diagErr, files.WriteToFile(jsonData, jsonFilepath)...)
			tflog.Info(g.ctx, fmt.Sprintf("Export errors written to %s", jsonFilepath))
		}

		// Add a warning diagnostic
		if timeoutErrorsTotalLen > 0 {
			diagErr = append(diagErr, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("%d resources were not exported due to timeouts. Consider increasing the token_pool_size and/or token_acquire_timeout provider configs.", timeoutErrorsTotalLen),
				Detail:   fmt.Sprintf("See %s for details", jsonFilepath),
			})
		}

		if otherErrorsTotalLen > 0 {
			diagErr = append(diagErr, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("%d resources encountered non-timeout errors during export.", otherErrorsTotalLen),
				Detail:   fmt.Sprintf("See %s for details", jsonFilepath),
			})
		}
	} else {
		tflog.Info(g.ctx, "Export completed successfully with no errors")
	}
	return diagErr
}

func (g *GenesysCloudResourceExporter) setUpExportDirPath() (diagErr diag.Diagnostics) {
	tflog.Info(g.ctx, "Setting up export directory path")

	g.exportDirPath, diagErr = getDirPath(g.d)
	if diagErr != nil {
		return diagErr
	}

	return nil
}

func (g *GenesysCloudResourceExporter) setupDataSource() {
	if replaceWithDatasource, ok := g.d.GetOk("replace_with_datasource"); ok {
		dataSourceList := lists.InterfaceListToStrings(replaceWithDatasource.([]interface{}))
		g.replaceWithDatasource = dataSourceList
	}
	SetDataSourceExports()
	g.replaceWithDatasource = append(g.replaceWithDatasource, DataSourceExports...)
}

// retrieveExporters will return a list of all the registered exporters. If the resource_type on the exporter contains any elements, only the defined
// elements in the resource_type attribute will be returned.
func (g *GenesysCloudResourceExporter) retrieveExporters() (diagErr diag.Diagnostics) {
	tflog.Debug(g.ctx, "Retrieving exporters list")
	exports := resourceExporter.GetResourceExporters()

	tflog.Trace(g.ctx, fmt.Sprintf("Retrieving exporters filtered list %v", g.filterList))

	if g.resourceTypeFilter != nil && g.filterList != nil {
		exports = g.resourceTypeFilter(exports, *g.filterList)
	}

	g.exporters = &exports

	// Assign excluded attributes to the config Map
	if excludedAttrs, ok := g.d.GetOk("exclude_attributes"); ok {
		if diagErr := g.populateConfigExcluded(*g.exporters, lists.InterfaceListToStrings(excludedAttrs.([]interface{}))); diagErr != nil {
			return diagErr
		}
	}
	return nil
}

// Removes the ::resource_label from the resource_types list
func formatFilter(filter []string) []string {
	newFilter := make([]string, 0)
	for _, str := range filter {
		newFilter = append(newFilter, strings.Split(str, "::")[0])
	}
	return newFilter
}

// retrieveSanitizedResourceMaps will retrieve a list of all resources to be exported.  It will also apply a filter (e.g the :: ) and only return the specific Genesys Cloud
// resources that are specified via :: delimiter
func (g *GenesysCloudResourceExporter) retrieveSanitizedResourceMaps() (diagErr diag.Diagnostics) {
	tflog.Info(g.ctx, "Retrieving map of Genesys Cloud resources to export")
	var filter []string
	if exportableResourceTypes, ok := g.d.GetOk("resource_types"); ok {
		filter = lists.InterfaceListToStrings(exportableResourceTypes.([]interface{}))
	}

	if exportableResourceTypes, ok := g.d.GetOk("include_filter_resources"); ok {
		filter = lists.InterfaceListToStrings(exportableResourceTypes.([]interface{}))
	}

	if exportableResourceTypes, ok := g.d.GetOk("exclude_filter_resources"); ok {
		filter = lists.InterfaceListToStrings(exportableResourceTypes.([]interface{}))
	}

	newFilter := make([]string, 0)
	for _, f := range filter {
		if strings.Contains(f, "::") {
			newFilter = append(newFilter, f)
		}
	}

	//Retrieve a map of all objects we are going to build.  Apply the filter that will remove specific classes of an object
	log.Println("Building sanitized resource maps")
	diagErr = g.buildSanitizedResourceMaps(*g.exporters, newFilter, g.logPermissionErrors)
	if diagErr.HasError() {
		return diagErr
	}

	//Check to see if we found any exporters.  If we did find the exporter
	if len(*g.exporters) == 0 {
		diagErr = append(diagErr, diag.Errorf("No valid resource types to export.")...)
		return diagErr
	}

	return diagErr
}

// retrieveGenesysCloudObjectInstances will take a list of exporters and then return the actual terraform Genesys Cloud data
func (g *GenesysCloudResourceExporter) retrieveGenesysCloudObjectInstances() diag.Diagnostics {
	tflog.Info(g.ctx, "Starting to retrieve Genesys Cloud objects from Genesys Cloud")
	tflog.Info(g.ctx, fmt.Sprintf("Number of exporters to process: %d", len(*g.exporters)))

	// Log all exporter types being processed
	for resType, exporter := range *g.exporters {
		tflog.Debug(g.ctx, fmt.Sprintf("Exporter for %s has %d resources in SanitizedResourceMap", resType, len(exporter.SanitizedResourceMap)))
	}

	// Retrieves data on each individual Genesys Cloud object from each registered exporter

	// Buffer error channel to prevent goroutine leaks or deadlocks
	errorChan := make(chan diag.Diagnostics, len(*g.exporters))
	wgDone := make(chan bool)
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(g.ctx)
	defer cancel()
	tflog.Trace(g.ctx, "Created context with cancellation")

	// Track successful and failed resource types
	var successfulTypes []string
	var failedTypes []string
	var statsMutex sync.Mutex

	// We use concurrency here to spin off each exporter type and getting the data
	for resType, exporter := range *g.exporters {
		tflog.Debug(g.ctx, fmt.Sprintf("Starting processing for resource type: %s", resType))
		wg.Add(1)
		go func(resType string, exporter *resourceExporter.ResourceExporter) {
			defer wg.Done()
			tflog.Trace(g.ctx, fmt.Sprintf("Starting goroutine for resource type: %s", resType))

			// Check if context was cancelled before processing
			select {
			case <-ctx.Done():
				tflog.Warn(g.ctx, fmt.Sprintf("Context cancelled before processing resource type: %s", resType))
				return
			default:
			}

			tflog.Debug(g.ctx, fmt.Sprintf("Getting exported resources for [%s]", resType))
			typeResources, err := g.getResourcesForType(resType, g.provider, exporter, g.meta)

			if err != nil {
				tflog.Error(g.ctx, fmt.Sprintf("Error getting resources for type %s: %v", resType, err))

				// Track failed resource type
				statsMutex.Lock()
				failedTypes = append(failedTypes, resType)
				statsMutex.Unlock()

				select {
				case <-ctx.Done():
					tflog.Warn(g.ctx, fmt.Sprintf("Context cancelled while handling error for %s", resType))
				case errorChan <- err:
					tflog.Trace(g.ctx, fmt.Sprintf("Successfully sent error to channel for resource type: %s", resType))
				default:
					tflog.Error(g.ctx, fmt.Sprintf("Could not send error to channel for resource type: %s (channel full)", resType))
				}
				cancel()
				return
			}

			tflog.Info(g.ctx, fmt.Sprintf("Successfully retrieved %d resources for type %s", len(typeResources), resType))

			// Use thread-safe method to add resources
			if len(typeResources) > 0 {
				g.addResources(typeResources)
				tflog.Debug(g.ctx, fmt.Sprintf("Successfully added %d resources for type %s to global resources list", len(typeResources), resType))

				// Track successful resource type
				statsMutex.Lock()
				successfulTypes = append(successfulTypes, resType)
				statsMutex.Unlock()
			} else {
				tflog.Warn(g.ctx, fmt.Sprintf("No resources found for type %s", resType))
			}

			tflog.Debug(g.ctx, fmt.Sprintf("Completed processing for resource type: %s", resType))
		}(resType, exporter)
	}

	tflog.Trace(g.ctx, "Started all goroutines, waiting for completion")

	go func() {
		wg.Wait()
		tflog.Info(g.ctx, "All resource type exports completed")
		close(wgDone)
	}()

	// Wait until either WaitGroup is done or an error is received
	select {
	case <-wgDone:
		tflog.Info(g.ctx, "Successfully retrieved all Genesys Cloud object instances")
		tflog.Info(g.ctx, fmt.Sprintf("Summary - Successful types: %v", successfulTypes))
		tflog.Info(g.ctx, fmt.Sprintf("Summary - Failed types: %v", failedTypes))
		tflog.Info(g.ctx, fmt.Sprintf("Summary - Total successful: %d, Total failed: %d", len(successfulTypes), len(failedTypes)))
		return nil
	case err := <-errorChan:
		tflog.Error(g.ctx, fmt.Sprintf("Error received from channel: %v", err))

		// Give other goroutines a chance to clean up
		go func() {
			<-wgDone // Wait for all goroutines to finish
			tflog.Trace(g.ctx, "All goroutines finished after error")
		}()

		tflog.Error(g.ctx, fmt.Sprintf("Returning error retrieving cloud object instances: %v", err))
		return err
	}
}

// buildResourceConfigMap Builds a map of all the Terraform resources data returned for each resource
func (g *GenesysCloudResourceExporter) buildResourceConfigMap() (diagnostics diag.Diagnostics) {
	tflog.Info(g.ctx, "Build Genesys Cloud Resources Map")

	// Initialize maps using thread-safe methods
	g.setResourceTypesMaps(make(map[string]ResourceJSONMaps))
	g.setDataSourceTypesMaps(make(map[string]ResourceJSONMaps))

	// Get resources using thread-safe method
	resources := g.getResources()

	for _, resource := range resources {
		// 1. Get instance state as JSON Map
		jsonResult, diagErr := g.instanceStateToMap(resource.State, resource.CtyType)
		if diagErr != nil {
			diagnostics = append(diagnostics, diagErr...)
			if diagnostics.HasError() {
				return
			}
		}

		// 2. Determine if instance is a data source
		isDataSource := g.isDataSource(resource.Type, resource.BlockLabel, resource.OriginalLabel)
		if isDataSource {
			dataSourceMaps := g.getDataSourceTypesMaps()
			if dataSourceMaps[resource.Type] == nil {
				dataSourceMaps[resource.Type] = make(ResourceJSONMaps)
			}
			g.setDataSourceTypesMaps(dataSourceMaps)
		} else {
			// 3. Ensure the resource type is instantiated
			resourceMaps := g.getResourceTypesMaps()
			if resourceMaps[resource.Type] == nil {
				resourceMaps[resource.Type] = make(ResourceJSONMaps)
			}
			g.setResourceTypesMaps(resourceMaps)
		}

		// Theoretically this should only ever occur when using the Original Sanitizer as it doesn't have guaranteed
		// uniqueness for generating the block labels. See resource_name_sanitizer_test.go
		resourceMaps := g.getResourceTypesMaps()
		dataSourceMaps := g.getDataSourceTypesMaps()

		if len(resourceMaps[resource.Type][resource.BlockLabel]) > 0 || len(dataSourceMaps[resource.Type][resource.BlockLabel]) > 0 {
			algorithm := fnv.New32()
			algorithm.Write([]byte(uuid.NewString()))
			// The _BRCM prefix is meant to be an identifier so we can tell that the hash was generated here and not in the sanitizer.
			resource.BlockLabel = resource.BlockLabel + "_BRCM" + strconv.FormatUint(uint64(algorithm.Sum32()), 10)
			g.updateSanitizeMap(*g.exporters, resource)
		}

		if resource.Type == architectFlow.ResourceType && !g.d.Get("use_legacy_architect_flow_exporter").(bool) {
			(*g.exporters)[architectFlow.ResourceType] = resourceExporter.GetNewFlowResourceExporter()
		}

		// 4. Convert the instance state to a map
		configMap := maps.Clone(jsonResult)

		// 5. Sanitize the config map
		unresolvableAttrs, _ := g.sanitizeConfigMap(resource, configMap, "", *g.exporters, g.includeStateFile, g.exportFormat, true)
		if len(unresolvableAttrs) > 0 {
			g.addUnresolvedAttrs(unresolvableAttrs)
		}

		// 6. Add the resource to the appropriate map
		if isDataSource {
			dataSourceMaps = g.getDataSourceTypesMaps()
			dataSourceMaps[resource.Type][resource.BlockLabel] = configMap
			g.setDataSourceTypesMaps(dataSourceMaps)
		} else {
			// 6. Handles writing external files as part of the export process
			diagnostics = append(diagnostics, g.customWriteAttributes(configMap, resource)...)
			if diagnostics.HasError() {
				return diagnostics
			}
			resourceMaps = g.getResourceTypesMaps()
			resourceMaps[resource.Type][resource.BlockLabel] = configMap
			g.setResourceTypesMaps(resourceMaps)
		}
	}

	tflog.Info(g.ctx, fmt.Sprintf("Successfully built resource config map with %d resources", len(resources)))
	return diagnostics
}

func (g *GenesysCloudResourceExporter) customWriteAttributes(jsonResult util.JsonMap,
	resource resourceExporter.ResourceInfo) (diagnostics diag.Diagnostics) {
	exporters := *g.exporters

	if resourceFilesWriterFunc := exporters[resource.Type].CustomFileWriter.RetrieveAndWriteFilesFunc; resourceFilesWriterFunc != nil {
		exportDir, getFilePathDiags := getFilePath(g.d, "")
		diagnostics = append(diagnostics, getFilePathDiags...)
		if diagnostics.HasError() {
			return
		}
		if err := resourceFilesWriterFunc(resource.State.ID, exportDir, exporters[resource.Type].CustomFileWriter.SubDirectory, jsonResult, g.meta, resource); err != nil {
			tflog.Error(g.ctx, fmt.Sprintf("An error has occurred while trying invoking the RetrieveAndWriteFilesFunc for resource type %s: %v", resource.Type, err))
			diagnostics = append(diagnostics, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Failed to invoke %s custom resolver method.", resource.Type),
				Detail:   err.Error(),
			})
		}
	}

	return diagnostics
}

func (g *GenesysCloudResourceExporter) updateInstanceStateAttributes(jsonResult util.JsonMap, resource resourceExporter.ResourceInfo) {
	for attr, val := range jsonResult {
		if valStr, ok := val.(string); ok {
			// Directly add string attribute for rest of the flows
			resource.State.Attributes[attr] = valStr
		}
	}
}

func (g *GenesysCloudResourceExporter) updateSanitizeMap(exporters map[string]*resourceExporter.ResourceExporter, //Map of all of the exporters
	resource resourceExporter.ResourceInfo) {
	if exporters[resource.Type] != nil {
		// Get the sanitized label from the ID returned as a reference expression
		if idMetaMap := exporters[resource.Type].SanitizedResourceMap; idMetaMap != nil {
			if meta := idMetaMap[resource.State.ID]; meta != nil && meta.BlockLabel != "" {
				meta.BlockLabel = resource.BlockLabel
				meta.OriginalLabel = resource.OriginalLabel
			}
		}
	}
}

func (g *GenesysCloudResourceExporter) instanceStateToMap(state *terraform.InstanceState, ctyType cty.Type) (util.JsonMap, diag.Diagnostics) {
	stateVal, err := schema.StateValueFromInstanceState(state, ctyType)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	jsonMap, err := schema.StateValueToJSONMap(stateVal, ctyType)

	if err != nil {
		return nil, diag.FromErr(err)
	}
	return jsonMap, nil
}

// generateOutputFiles is used to generate the tfStateFile and either the tf export or the json based export
func (g *GenesysCloudResourceExporter) generateOutputFiles() (diags diag.Diagnostics) {

	if g.resourceTypesMaps == nil || g.dataSourceTypesMaps == nil {
		return diag.Errorf("required fields resourceTypesMaps or dataSourceTypesMaps are nil")
	}

	// Ensure export directory exists and is writable
	if err := os.MkdirAll(g.exportDirPath, 0755); err != nil {
		return diag.FromErr(err)
	}

	if g.includeStateFile {
		t, err := NewTFStateWriter(g.ctx, g.resources, g.d, g.providerRegistry)
		if err != nil {
			return diag.FromErr(err)
		}

		diags = append(diags, t.writeTfState()...)
		if diags.HasError() {
			return diags
		}
	}

	if g.matchesExportFormat(formatHCL, formatJSONHCL) {
		hclExporter := NewHClExporter(g.resourceTypesMaps, g.dataSourceTypesMaps, g.unresolvedAttrs, g.providerRegistry, g.version, g.exportDirPath, g.splitFilesByResource)
		diags = append(diags, hclExporter.exportHCLConfig()...)
	}

	if g.matchesExportFormat(formatJSON, formatJSONHCL) {
		jsonExporter := NewJsonExporter(g.resourceTypesMaps, g.dataSourceTypesMaps, g.unresolvedAttrs, g.providerRegistry, g.version, g.exportDirPath, g.splitFilesByResource)
		diags = append(diags, jsonExporter.exportJSONConfig()...)
	}

	if diags != nil && diags.HasError() {
		return diags
	}

	if g.cyclicDependsList != nil && len(g.cyclicDependsList) > 0 {
		diags = append(diags, files.WriteToFile([]byte(strings.Join(g.cyclicDependsList, "\n")), filepath.Join(g.exportDirPath, "cyclicDepends.txt"))...)
		if diags.HasError() {
			return diags
		}
	}

	diags = append(diags, g.generateZipForExporter()...)
	return diags
}

func (g *GenesysCloudResourceExporter) generateZipForExporter() diag.Diagnostics {
	zipFileName := filepath.Join(g.exportDirPath, "..", "archive_genesyscloud_tf_export"+uuid.NewString()+".zip")
	if compress := g.d.Get("compress").(bool); compress { //if true, compress directory name of where the export is going to occur
		// read all the files
		var files []fileMeta
		ferr := filepath.Walk(g.exportDirPath, func(path string, info os.FileInfo, ferr error) error {
			files = append(files, fileMeta{Path: path, IsDir: info.IsDir()})
			return nil
		})
		if ferr != nil {
			return diag.Errorf("Failed to fetch file path %s", ferr)
		}
		// create a zip
		archive, ferr := os.Create(zipFileName)
		if ferr != nil {
			return diag.Errorf("Failed to create zip %s", ferr)
		}
		defer archive.Close()
		zipWriter := zip.NewWriter(archive)

		for _, f := range files {
			if !f.IsDir {
				fPath := f.Path

				w, ferr := zipWriter.Create(path.Base(fPath))
				if ferr != nil {
					return diag.Errorf("Failed to create base path for zip %s", ferr)
				}

				file, ferr := os.Open(f.Path)
				if ferr != nil {
					return diag.Errorf("Failed to open the original file %s", ferr)
				}
				defer file.Close()

				if _, ferr = io.Copy(w, file); ferr != nil {
					return diag.Errorf("Failed to copy the file to zip %s", ferr)
				}
			}
		}
		zipWriter.Close()
	}

	return nil
}

func (g *GenesysCloudResourceExporter) buildAndExportDependsOnResourcesForFlows() diag.Diagnostics {

	if g.addDependsOn {
		filterList, resources, err := g.processAndBuildDependencies()
		if err != nil {
			return err
		}
		if len(filterList) > 0 {
			diagErr := g.exportDependentResources(filterList, resources)
			if diagErr != nil {
				return diagErr
			}
		}
		return nil
	}
	return nil
}

func (g *GenesysCloudResourceExporter) processAndBuildDependencies() (filters []string, resources resourceExporter.ResourceIDMetaMap, diagErr diag.Diagnostics) {
	filterList := make([]string, 0)
	totalResources := make(resourceExporter.ResourceIDMetaMap)
	proxy := dependentconsumers.GetDependentConsumerProxy(nil)

	retrieveDependentConsumers := func(resourceKeys resourceExporter.ResourceInfo) func(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, diag.Diagnostics) {
		return func(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, diag.Diagnostics) {
			proxy = dependentconsumers.GetDependentConsumerProxy(clientConfig)
			resources := make(resourceExporter.ResourceIDMetaMap)
			resources, dependsMap, err := proxy.GetDependentConsumers(ctx, resourceKeys)

			if err != nil {
				return nil, nil, diag.Errorf("Failed to retrieve Dependent Flows %s: %s", resourceKeys.State.ID, err)
			}
			return resources, dependsMap, nil
		}
	}

	for _, resourceKeys := range g.resources {

		exists := util.StringExists(resourceKeys.State.ID, g.flowResourcesList)
		if exists {
			tflog.Debug(g.ctx, fmt.Sprintf("dependent consumers retrieved %v", resourceKeys.State.ID))
			continue
		}

		resources, dependsStruct, err := proxy.GetAllWithPooledClient(retrieveDependentConsumers(resourceKeys))

		g.flowResourcesList = append(g.flowResourcesList, resourceKeys.State.ID)

		if err != nil {
			return nil, nil, err
		}

		if len(resources) > 0 {
			resourcesTobeExported := retrieveExportResources(g.resources, resources)
			for _, meta := range resourcesTobeExported {

				resource := strings.Split(meta.BlockLabel, "::::")
				filterList = append(filterList, fmt.Sprintf("%s::%s", resource[0], resource[1]))
			}
			g.dependsList = stringmap.MergeMaps(g.dependsList, dependsStruct.DependsMap)
			g.cyclicDependsList = append(g.cyclicDependsList, dependsStruct.CyclicDependsList...)
			totalResources = stringmap.MergeSingularMaps(totalResources, resources)
		}
	}

	if !g.ignoreCyclicDeps && len(g.cyclicDependsList) > 0 {
		return nil, nil, diag.Errorf("Cyclic Dependencies Identified:  %v ", strings.Join(g.cyclicDependsList, "\n"))
	}
	return filterList, totalResources, nil
}

func (g *GenesysCloudResourceExporter) rebuildExports(filterList []string) (diagErr diag.Diagnostics) {
	tflog.Info(g.ctx, "Rebuilding exporters list")
	diagErr = g.retrieveExporters()
	if diagErr != nil {
		return diagErr
	}

	diagErr = g.buildSanitizedResourceMaps(*g.exporters, filterList, g.logPermissionErrors)
	if diagErr != nil {
		return diagErr
	}

	diagErr = g.retrieveGenesysCloudObjectInstances()
	if diagErr != nil {
		return diagErr
	}
	return nil
}

func (g *GenesysCloudResourceExporter) exportDependentResources(filterList []string, resources resourceExporter.ResourceIDMetaMap) (diagErr diag.Diagnostics) {
	g.reAssignFilters()
	g.filterList = &filterList
	existingExporters := g.copyExporters()
	existingResources := g.copyResources()
	tflog.Info(g.ctx, "Rebuilding exports from exportDependentResources")

	err := g.rebuildExports(filterList)
	if err != nil {
		return err
	}

	// retain the exporters and resources
	g.retainExporterList(resources)
	uniqueResources := g.attainUniqueResourceList(resources)

	// deep copy is needed here else exporters being overridden
	depExporters := g.copyExporters()

	// this is done before the merge of exporters and this will make sure only dependency resources are resolved
	diagErr = append(diagErr, g.buildResourceConfigMap()...)
	if diagErr.HasError() {
		return diagErr
	}
	diagErr = append(diagErr, g.exportAndResolveDependencyAttributes()...)
	if diagErr.HasError() {
		return diagErr
	}
	g.appendResources(uniqueResources)
	g.appendResources(existingResources)
	g.exporters = mergeExporters(existingExporters, *mergeExporters(depExporters, *g.exporters))

	return diagErr
}

func (g *GenesysCloudResourceExporter) buildAndExportDependentResources() (diagErr diag.Diagnostics) {
	if g.addDependsOn {
		g.reAssignFilters()
		existingExporters := g.copyExporters()
		existingResources := g.copyResources()

		// this will make sure all the dependency resources are resolved
		g.exportAndResolveDependencyAttributes()

		// merge the resources and exporters after the dependencies are resolved
		g.appendResources(existingResources)
		g.exporters = mergeExporters(existingExporters, *g.exporters)

		// rebuild the config map
		diagErr = g.buildResourceConfigMap()
	}
	return
}

func (g *GenesysCloudResourceExporter) copyExporters() map[string]*resourceExporter.ResourceExporter {
	// deep copy is needed here else exporters are being overridden
	existingExportersInterface := deepcopy.Copy(*g.exporters)
	existingExporters, _ := existingExportersInterface.(map[string]*resourceExporter.ResourceExporter)
	return existingExporters
}

func (g *GenesysCloudResourceExporter) copyResources() []resourceExporter.ResourceInfo {
	existingResources := g.copyResource()
	g.resources = nil
	return existingResources
}

func (g *GenesysCloudResourceExporter) copyResource() []resourceExporter.ResourceInfo {
	existingResources := make([]resourceExporter.ResourceInfo, 0)
	for _, resource := range g.resources {
		existingResource := extractResource(resource)
		existingResources = append(existingResources, existingResource)
	}
	return existingResources
}

func (g *GenesysCloudResourceExporter) copyResourceAddtoG(resourcesToAdd []resourceExporter.ResourceInfo) {
	existingResources := make([]resourceExporter.ResourceInfo, 0)
	for _, resource := range resourcesToAdd {
		existingResource := extractResource(resource)
		existingResources = append(existingResources, existingResource)
	}
	g.resources = existingResources
}

func extractResource(resource resourceExporter.ResourceInfo) resourceExporter.ResourceInfo {
	existingResourceInterface := deepcopy.Copy(resource)
	existingResource, _ := existingResourceInterface.(resourceExporter.ResourceInfo)
	existingResourceCtyTypeInterface := deepcopy.Copy(resource.CtyType)
	existingResource.CtyType, _ = existingResourceCtyTypeInterface.(cty.Type)
	if existingResource.CtyType == cty.NilType {
		existingResource.CtyType = resource.CtyType
	}
	return existingResource
}

func (g *GenesysCloudResourceExporter) retainExporterList(resources resourceExporter.ResourceIDMetaMap) diag.Diagnostics {
	removeChan := make([]string, 0)
	for _, exporter := range *g.exporters {
		for id, _ := range exporter.SanitizedResourceMap {
			_, exists := resources[id]
			if !exists {
				removeChan = append(removeChan, id)
			}
		}
		for _, removeId := range removeChan {
			tflog.Debug(g.ctx, fmt.Sprintf("Deleted removeId %v", removeId))
			delete(exporter.SanitizedResourceMap, removeId)
		}
	}
	return nil
}

func (g *GenesysCloudResourceExporter) reAssignFilters() {
	g.resourceTypeFilter = IncludeFilterByResourceType
	g.resourceFilter = FilterResourceById
}

func (g *GenesysCloudResourceExporter) attainUniqueResourceList(resources resourceExporter.ResourceIDMetaMap) []resourceExporter.ResourceInfo {
	uniqueResources := make([]resourceExporter.ResourceInfo, 0)
	for _, resource := range g.resources {
		_, exists := resources[resource.State.ID]
		if exists {
			uniqueResources = append(uniqueResources, resource)
		}
	}
	return uniqueResources
}

func (g *GenesysCloudResourceExporter) exportAndResolveDependencyAttributes() (diagErr diag.Diagnostics) {
	if g.addDependsOn {
		g.resources = nil
		exp := make(map[string]*resourceExporter.ResourceExporter, 0)
		filterListById := make([]string, 0)

		// build filter list with guid.
		for refType, guidList := range g.buildSecondDeps {
			if refType != "" {
				for _, guid := range guidList {
					if guid != "" {
						filterListById = append(filterListById, fmt.Sprintf("%s::%s", refType, guid))
					}
				}
			}
		}

		if len(filterListById) > 0 {
			g.resourceFilter = FilterResourceById
			diagErr = append(diagErr, g.chainDependencies(make([]resourceExporter.ResourceInfo, 0), exp)...)
			if diagErr.HasError() {
				return
			}
		}
	}
	return
}

// Recursive function to perform operations based on filterListById length
func (g *GenesysCloudResourceExporter) chainDependencies(
	existingResources []resourceExporter.ResourceInfo,
	existingExporters map[string]*resourceExporter.ResourceExporter) (diagErr diag.Diagnostics) {
	filterListById := make([]string, 0)

	for refType, guidList := range g.buildSecondDeps {
		if refType != "" {
			for _, guid := range guidList {
				if guid != "" {
					if !g.resourceIdExists(guid, existingResources) {
						filterListById = append(filterListById, fmt.Sprintf("%s::%s", refType, guid))
					} else {
						tflog.Warn(g.ctx, fmt.Sprintf("Resource already present in the resources. %v", guid))
					}

				}
			}
		}
	}
	g.filterList = &filterListById
	g.buildSecondDeps = nil

	if len(*g.filterList) > 0 {
		g.resources = nil
		g.exporters = nil
		tflog.Debug(g.ctx, "Rebuilding exporters list from chainDependencies")
		diagErr = append(diagErr, g.rebuildExports(*g.filterList)...)
		if diagErr.HasError() {
			return
		}
		// checks and exports if there are any dependent flow resources
		diagErr = append(diagErr, g.buildAndExportDependsOnResourcesForFlows()...)
		if diagErr.HasError() {
			return
		}

		diagErr = append(diagErr, g.buildResourceConfigMap()...)
		if diagErr.HasError() {
			return
		}
		//append the resources and exporters
		g.appendResources(existingResources)
		g.exporters = mergeExporters(existingExporters, *g.exporters)

		// deep copy is needed here else exporters being overridden
		existingExportersInterface := deepcopy.Copy(*g.exporters)
		existingExporters, _ = existingExportersInterface.(map[string]*resourceExporter.ResourceExporter)
		existingResources = g.resources

		// Recursive call until all the dependencies are addressed.
		return append(diagErr, g.chainDependencies(existingResources, existingExporters)...)
	}
	return
}

func (g *GenesysCloudResourceExporter) appendResources(resourcesToAdd []resourceExporter.ResourceInfo) {
	tflog.Debug(g.ctx, fmt.Sprintf("Appending %d resources to existing resources", len(resourcesToAdd)))

	// Get existing resources using thread-safe method
	existingResources := g.getResources()

	// Create a map for efficient duplicate checking
	existingResourceMap := make(map[string]bool)
	for _, resource := range existingResources {
		key := resource.Type + ":" + resource.State.ID
		existingResourceMap[key] = true
	}

	var newResources []resourceExporter.ResourceInfo
	for _, resourceToAdd := range resourcesToAdd {
		key := resourceToAdd.Type + ":" + resourceToAdd.State.ID
		if !existingResourceMap[key] {
			newResources = append(newResources, resourceToAdd)
			existingResourceMap[key] = true // Mark as added
		} else {
			tflog.Warn(g.ctx, fmt.Sprintf("Skipping duplicate resource: %s", key))
		}
	}

	if len(newResources) > 0 {
		// Use thread-safe method to add new resources
		g.addResources(newResources)
		tflog.Debug(g.ctx, fmt.Sprintf("Successfully added %d new resources", len(newResources)))
	} else {
		tflog.Debug(g.ctx, "No new resources to add")
	}
}

func (g *GenesysCloudResourceExporter) buildSanitizedResourceMaps(exporters map[string]*resourceExporter.ResourceExporter, filter []string, logErrors bool) diag.Diagnostics {
	tflog.Info(g.ctx, fmt.Sprintf("Starting to build sanitized resource maps for %d exporters", len(exporters)))

	// Buffer error channel to prevent goroutine leaks or deadlocks
	errorChan := make(chan diag.Diagnostics, len(exporters))
	wgDone := make(chan bool)

	// Cancel remaining goroutines if an error occurs
	ctx, cancel := context.WithCancel(g.ctx)
	defer cancel()

	// Create semaphore to limit concurrent operations to the configured maximum
	sem := make(chan struct{}, g.maxConcurrentOps)
	tflog.Trace(g.ctx, fmt.Sprintf("Using semaphore to limit concurrent operations to %d", g.maxConcurrentOps))

	var wg sync.WaitGroup
	for resourceType, exporter := range exporters {
		wg.Add(1)
		go func(resourceType string, exporter *resourceExporter.ResourceExporter) {
			defer wg.Done()
			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }() // Release semaphore when done
			case <-ctx.Done():
				return
			}

			tflog.Info(g.ctx, fmt.Sprintf("Getting all resources for type %s", resourceType))
			exporter.FilterResource = g.resourceFilter

			err := exporter.LoadSanitizedResourceMap(ctx, resourceType, filter)

			// Used in tests
			if mockError != nil {
				err = mockError
			}
			if containsPermissionsErrorOnly(err) && logErrors {
				tflog.Error(g.ctx, fmt.Sprintf("%v", err[0].Summary))
				tflog.Warn(g.ctx, fmt.Sprintf("Logging permission error for %s. Resuming export...", resourceType))
				return
			}
			if err != nil {
				if !logErrors {
					err = addLogAttrInfoToErrorSummary(err)
				}
				select {
				case <-ctx.Done():
				case errorChan <- err:
					cancel() // Cancel other operations on error
				}
				return
			}
			tflog.Info(g.ctx, fmt.Sprintf("Found %d resources for type %s", len(exporter.SanitizedResourceMap), resourceType))
		}(resourceType, exporter)
	}

	go func() {
		wg.Wait()
		tflog.Info(g.ctx, `Finished building sanitized resource maps`)
		close(wgDone)
	}()

	// Wait until either WaitGroup is done or an error is received
	select {
	case <-wgDone:
		tflog.Info(g.ctx, "Successfully completed building sanitized resource maps")
		return nil
	case err := <-errorChan:
		// Give other goroutines a chance to clean up
		go func() {
			<-wgDone // Wait for all goroutines to finish
		}()
		tflog.Warn(g.ctx, fmt.Sprintf("Error occurred while building sanitized resource maps: %v", err))
		return err
	}
}

func mergeExporters(m1, m2 map[string]*resourceExporter.ResourceExporter) *map[string]*resourceExporter.ResourceExporter {
	result := make(map[string]*resourceExporter.ResourceExporter)

	for k, v := range m1 {
		result[k] = v
	}

	for k, v := range m2 {
		_, exists := result[k]
		if exists {
			for id, value := range v.SanitizedResourceMap {
				result[k].SanitizedResourceMap[id] = value

			}
			if result[k].ExcludedAttributes != nil {
				result[k].ExcludedAttributes = append(result[k].ExcludedAttributes, v.ExcludedAttributes...)
			} else {
				result[k].ExcludedAttributes = v.ExcludedAttributes
			}

		} else {
			result[k] = v
		}
	}
	return &result
}

func retrieveExportResources(existingResources []resourceExporter.ResourceInfo, resources resourceExporter.ResourceIDMetaMap) map[string]*resourceExporter.ResourceMeta {
	foundTypes := make(map[string]bool)
	resourcesTobeExported := make(map[string]*resourceExporter.ResourceMeta)

	for _, data := range existingResources {
		if _, ok := resources[data.State.ID]; ok {
			foundTypes[data.State.ID] = true
		}
	}

	for resourceType, meta := range resources {
		if !foundTypes[resourceType] {
			resourcesTobeExported[resourceType] = meta
		}
	}

	return resourcesTobeExported
}

func containsPermissionsErrorOnly(err diag.Diagnostics) bool {
	foundPermissionsError := false
	for _, v := range err {
		if strings.Contains(v.Summary, "403") ||
			strings.Contains(v.Summary, "501") {
			foundPermissionsError = true
		} else {
			return false
		}
	}
	return foundPermissionsError
}

var logAttrInfo = "\nTo continue exporting other resources in spite of this error, set the 'log_permission_errors' attribute to 'true'"

func addLogAttrInfoToErrorSummary(err diag.Diagnostics) diag.Diagnostics {
	for i, v := range err {
		if strings.Contains(v.Summary, "403") ||
			strings.Contains(v.Summary, "501") {
			err[i].Summary += logAttrInfo
		}
	}
	return err
}

func (g *GenesysCloudResourceExporter) getResourcesForType(resType string, schemaProvider *schema.Provider, exporter *resourceExporter.ResourceExporter, meta interface{}) ([]resourceExporter.ResourceInfo, diag.Diagnostics) {
	tflog.Debug(g.ctx, fmt.Sprintf("Starting export for resource type: %s", resType))

	// Use thread-safe method to get resource map size
	lenResources := exporter.GetSanitizedResourceMapSize()
	tflog.Info(g.ctx, fmt.Sprintf("Found %d resources for type %s", lenResources, resType))

	if lenResources == 0 {
		tflog.Debug(g.ctx, fmt.Sprintf("No resources found for type %s, returning empty slice", resType))
		return []resourceExporter.ResourceInfo{}, nil
	}

	// Buffer channels to prevent goroutine leaks
	errorsChan := make(chan ResourceErrorInfo, lenResources)
	resourceChan := make(chan resourceExporter.ResourceInfo, lenResources)
	tflog.Trace(g.ctx, fmt.Sprintf("Created buffered channels for %d resources", lenResources))

	res := schemaProvider.ResourcesMap[resType]
	if res == nil {
		tflog.Error(g.ctx, fmt.Sprintf("Resource type %v not defined in schema provider", resType))
		return nil, diag.Errorf("Resource type %v not defined", resType)
	}
	tflog.Trace(g.ctx, fmt.Sprintf("Successfully retrieved resource schema for type %s", resType))

	exportComputed := g.exportComputed
	tflog.Debug(g.ctx, fmt.Sprintf("Export computed setting: %t", exportComputed))

	// Create a context with cancellation for this operation
	ctx, cancel := context.WithCancel(g.ctx)
	defer cancel()
	tflog.Trace(g.ctx, fmt.Sprintf("Created context with cancellation for resource type %s", resType))

	var wg sync.WaitGroup
	wg.Add(lenResources)
	tflog.Trace(g.ctx, fmt.Sprintf("Initialized WaitGroup with %d resources", lenResources))

	// Track resources to remove to avoid race conditions
	var toRemove []string
	var toRemoveMutex sync.Mutex
	tflog.Trace(g.ctx, fmt.Sprintf("Initialized resource removal tracking for type %s", resType))

	// Get a copy of the resource map to avoid race conditions during iteration
	resourceMap := exporter.GetSanitizedResourceMap()
	tflog.Trace(g.ctx, fmt.Sprintf("Retrieved sanitized resource map with %d entries for type %s", len(resourceMap), resType))

	// Create a semaphore to limit concurrent operations
	sem := make(chan struct{}, g.maxConcurrentOps)
	tflog.Trace(g.ctx, fmt.Sprintf("Using semaphore to limit concurrent operations"))

	for id, resMeta := range resourceMap {
		tflog.Trace(g.ctx, fmt.Sprintf("Starting goroutine for resource ID: %s, BlockLabel: %s", id, resMeta.BlockLabel))

		go func(id string, resMeta *resourceExporter.ResourceMeta) {
			defer wg.Done()
			tflog.Debug(g.ctx, fmt.Sprintf("Starting processing for resource ID: %s, BlockLabel: %s", id, resMeta.BlockLabel))

			// Acquire semaphore slot or return if context is cancelled or timeout
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }() // Release semaphore when done
			case <-ctx.Done():
				tflog.Trace(g.ctx, fmt.Sprintf("Context cancelled for resource ID: %s", id))
				return
			}

			tflog.Trace(g.ctx, fmt.Sprintf("Started processing for resource: %s.%s (%s)", resType, resMeta.BlockLabel, id))

			fetchResourceState := func() error {
				tflog.Trace(g.ctx, fmt.Sprintf("Starting fetchResourceState for resource ID: %s", id))

				resourceCtx, resourceCancel := context.WithCancel(ctx)
				defer resourceCancel()
				tflog.Trace(g.ctx, fmt.Sprintf("Created resource context for ID: %s", id))

				ctyType := res.CoreConfigSchema().ImpliedType()
				tflog.Trace(g.ctx, fmt.Sprintf("Retrieved CTY type for resource ctyType: %v", ctyType))

				tflog.Trace(g.ctx, fmt.Sprintf("Calling getResourceState for resource ID: %s", id))
				instanceState, err := g.getResourceState(resourceCtx, res, id, resMeta, meta)

				if err != nil {
					tflog.Error(g.ctx, fmt.Sprintf("Error while fetching read context type %s and instance %s : %v", resType, id, err))
					return fmt.Errorf("Failed to get state for %s instance %s: %v", resType, id, err)
				}

				if instanceState == nil {
					tflog.Warn(g.ctx, fmt.Sprintf("Resource %s no longer exists. Skipping.", resMeta.BlockLabel))
					toRemoveMutex.Lock()
					toRemove = append(toRemove, id)
					toRemoveMutex.Unlock()
					tflog.Debug(g.ctx, fmt.Sprintf("Added resource ID %s to removal list", id))
					return nil
				}
				tflog.Info(g.ctx, fmt.Sprintf("Successfully retrieved instance state for resource ID: %s", id))

				// Export the resource as a data resource
				if exporter.ExportAsDataFunc != nil {
					tflog.Trace(g.ctx, fmt.Sprintf("Checking if resource should be exported as data source for ID: %s", id))
					sdkConfig := g.meta.(*provider.ProviderMeta).ClientConfig
					exportAsData, err := exporter.ExportAsDataFunc(g.ctx, sdkConfig, instanceState.Attributes)
					if err != nil {
						tflog.Error(g.ctx, fmt.Sprintf("Error in ExportAsDataFunc for resource ID %s: %v", id, err))
						return fmt.Errorf("an error has occurred while trying to export as a data resource block for %s::%s : %v", resType, resMeta.BlockLabel, err)
					} else {
						if exportAsData {
							tflog.Debug(g.ctx, fmt.Sprintf("Resource ID %s will be exported as data source", id))
							g.addReplaceWithDatasource(resType + "::" + resMeta.BlockLabel)
						} else {
							tflog.Debug(g.ctx, fmt.Sprintf("Resource ID %s will NOT be exported as data source", id))
						}
					}
				} else {
					tflog.Debug(g.ctx, fmt.Sprintf("No ExportAsDataFunc defined for resource ID: %s", id))
				}

				blockType := ""

				if g.isDataSource(resType, resMeta.BlockLabel, resMeta.OriginalLabel) {
					attributes := make(map[string]string)
					g.exMutex.Lock()
					resData := schemaProvider.DataSourcesMap[resType]
					g.exMutex.Unlock()

					if resData == nil {
						return fmt.Errorf("DataSource type %v not defined", resType)
					}

					ctyType = resData.CoreConfigSchema().ImpliedType()
					for attr, _ := range resData.SchemaMap() {
						key, val := exporter.DataResolver(instanceState, attr)
						attributes[key] = val
					}
					instanceState.Attributes = attributes
					blockType = "data"
				}

				for resAttribute, resSchema := range res.Schema {
					// Remove any computed attributes if export computed exporter config not set
					if resSchema.Computed == true && !exportComputed {
						delete(instanceState.Attributes, resAttribute)
						continue
					}
					// Remove any computed read-only attributes from being exported regardless of exporter config
					// because they cannot be set by a user when reapplying the configuration in a different org
					if resSchema.Computed == true && resSchema.Optional == false {
						delete(instanceState.Attributes, resAttribute)
						continue
					}
				}
				tflog.Debug(g.ctx, fmt.Sprintf("Finished processing schema attributes for resource ID: %s", id))

				tflog.Trace(g.ctx, fmt.Sprintf("Creating ResourceInfo for resource ID: %s", id))
				select {
				case resourceChan <- resourceExporter.ResourceInfo{
					State:         instanceState,
					BlockLabel:    resMeta.BlockLabel,
					Type:          resType,
					CtyType:       ctyType,
					BlockType:     blockType,
					OriginalLabel: resMeta.OriginalLabel,
				}:
					tflog.Trace(g.ctx, fmt.Sprintf("Successfully sent ResourceInfo to channel for resource ID: %s", id))
				case <-ctx.Done():
					tflog.Warn(g.ctx, fmt.Sprintf("Context canceled while sending ResourceInfo for resource ID: %s", id))
					return fmt.Errorf("context cancelled")
				}
				tflog.Trace(g.ctx, fmt.Sprintf("Successfully sent ResourceInfo to channel for resource ID: %s", id))

				return nil
			}

			// Improved retry logic with exponential backoff
			// Allows retries up to three times before reporting error
			maxRetries := 3
			var lastErr error
			tflog.Debug(g.ctx, fmt.Sprintf("Starting retry logic for resource ID: %s (max retries: %d)", id, maxRetries))

			for attempt := 0; attempt < maxRetries; attempt++ {
				tflog.Debug(g.ctx, fmt.Sprintf("Attempt %d/%d for resource ID: %s", attempt+1, maxRetries, id))

				select {
				case <-ctx.Done():
					tflog.Warn(g.ctx, fmt.Sprintf("Context cancelled during retry attempt %d for resource ID: %s", attempt+1, id))
					return
				default:
				}

				err := fetchResourceState()

				// Return immediately if no errors
				if err == nil {
					tflog.Info(g.ctx, fmt.Sprintf("Successfully processed resource ID: %s on attempt %d", id, attempt+1))
					return // Success
				}

				lastErr = err

				tflog.Error(g.ctx, fmt.Sprintf("Error on attempt %d for resource ID %s: %v", attempt+1, id, err))

				if util.IsTimeoutError(err) {
					tflog.Debug(g.ctx, fmt.Sprintf("Timeout error detected for resource ID %s on attempt %d", id, attempt+1))

					// On timeout errors, try to add new client connections to the existing pool
					if attempt == 0 {
						tflog.Debug(g.ctx, fmt.Sprintf("Attempting to add new client connections to pool for resource ID: %s", id))

						// Get the current version from provider meta
						version := "1.0.0" // Default version
						if providerMeta, ok := g.meta.(*provider.ProviderMeta); ok && providerMeta.Version != "" {
							version = providerMeta.Version
						}

						// Call the pool's adjustment method if pool is initialized
						if provider.SdkClientPool != nil {
							provider.SdkClientPool.AdjustPoolForTimeout(version)
						}

						// Add a small delay to allow the new clients to be available
						tflog.Debug(g.ctx, fmt.Sprintf("Waiting 5 seconds for new clients to be available for resource ID: %s", id))
						select {
						case <-time.After(5 * time.Second):
							tflog.Debug(g.ctx, fmt.Sprintf("Pool adjustment delay completed for resource ID: %s", id))
						case <-ctx.Done():
							tflog.Warn(g.ctx, fmt.Sprintf("[getResourcesForType] Context cancelled during pool adjustment delay for resource ID: %s", id))
							return
						}
					}
					// Exponential backoff for retryable errors
					if attempt < maxRetries-1 {
						backoffDuration := time.Duration(1<<attempt) * time.Second
						tflog.Info(g.ctx, fmt.Sprintf("Retrying resource %s (attempt %d/%d) after %v backoff", id, attempt+1, maxRetries, backoffDuration))

						select {
						case <-time.After(backoffDuration):
							tflog.Debug(g.ctx, fmt.Sprintf("Backoff completed for resource ID: %s", id))
						case <-ctx.Done():
							tflog.Warn(g.ctx, fmt.Sprintf("Context cancelled during backoff for resource ID: %s", id))
							return
						}
					}
				} else {
					tflog.Error(g.ctx, fmt.Sprintf("Non-retryable error for resource ID %s, sending to error channel", id))
					// Non-retryable error, send to error channel
					select {
					case errorsChan <- ResourceErrorInfo{
						ResourceType:  resType,
						ResourceID:    id,
						ResourceLabel: resMeta.BlockLabel,
						ErrorMessage:  fmt.Sprintf("Non-retryable error: %v", err),
						IsTimeout:     false,
					}:
						tflog.Trace(g.ctx, fmt.Sprintf("Successfully sent error to channel for resource ID: %s", id))
					case <-ctx.Done():
						tflog.Warn(g.ctx, fmt.Sprintf("Context cancelled while sending error for resource ID: %s", id))
						return
					}
					return
				}
			}

			// If we get here, all retries failed
			if lastErr != nil {
				tflog.Trace(g.ctx, fmt.Sprintf("All retries failed for resource ID %s, sending final error to channel", id))

				select {
				case errorsChan <- ResourceErrorInfo{
					ResourceType:  resType,
					ResourceID:    id,
					ResourceLabel: resMeta.BlockLabel,
					ErrorMessage:  fmt.Sprintf("Failed after %d retries: %v", maxRetries, lastErr),
					IsTimeout:     util.IsTimeoutError(lastErr),
				}:
					tflog.Trace(g.ctx, fmt.Sprintf("Successfully sent final error to channel for resource ID: %s", id))
				case <-ctx.Done():
					tflog.Warn(g.ctx, fmt.Sprintf("Context cancelled while sending final error for resource ID: %s", id))
					return
				}

			}

		}(id, resMeta)
	}

	tflog.Trace(g.ctx, fmt.Sprintf("Started all goroutines for resource type %s, waiting for completion", resType))

	// Wait for all goroutines to complete with timeout
	wg.Wait()
	tflog.Trace(g.ctx, fmt.Sprintf("All goroutines completed for resource type %s", resType))

	close(resourceChan)
	close(errorsChan)
	tflog.Trace(g.ctx, fmt.Sprintf("Closed resource and error channels for resource type %s", resType))

	// Collect all resources
	var resources []resourceExporter.ResourceInfo
	tflog.Trace(g.ctx, fmt.Sprintf("Collecting resources from channel for resource type %s", resType))
	for r := range resourceChan {
		tflog.Trace(g.ctx, fmt.Sprintf("Collected resource: Type=%s, BlockLabel=%s, ID=%s", r.Type, r.BlockLabel, r.State.ID))
		resources = append(resources, r)
	}
	tflog.Trace(g.ctx, fmt.Sprintf("Collected %d resources from channel for resource type %s", len(resources), resType))

	// Remove resources that weren't found using thread-safe method
	tflog.Debug(g.ctx, fmt.Sprintf("Processing %d resources for removal", len(toRemove)))
	for _, id := range toRemove {
		tflog.Debug(g.ctx, fmt.Sprintf("Removing resource %v from export map", id))
		exporter.RemoveFromSanitizedResourceMap(id)
	}

	// Track resources that errored
	tflog.Trace(g.ctx, fmt.Sprintf("Collecting errors from channel for resource type %s", resType))
	var erroredResources []ResourceErrorInfo
	for erroredResource := range errorsChan {
		tflog.Debug(g.ctx, fmt.Sprintf("Collected error: %v", erroredResource))
		erroredResources = append(erroredResources, erroredResource)
	}

	tflog.Warn(g.ctx, fmt.Sprintf("Collected %d errors for resource type %s", len(erroredResources), resType))
	// Store errored resources in the exporter for later reporting
	if len(erroredResources) > 0 {
		g.resourceErrorsMutex.Lock()
		if g.resourceErrors == nil {
			g.resourceErrors = make(map[string][]ResourceErrorInfo)
		}
		g.resourceErrors[resType] = erroredResources
		g.resourceErrorsMutex.Unlock()
		tflog.Warn(g.ctx, fmt.Sprintf("Export completed for %s with %d errors out of %d resources", resType, len(erroredResources), lenResources))
	} else {
		tflog.Info(g.ctx, fmt.Sprintf("Export completed successfully for %s: %d resources successfully exported", resType, len(resources)))
	}

	return resources, nil
}

func (g *GenesysCloudResourceExporter) getResourceState(ctx context.Context, resource *schema.Resource, resID string, resMeta *resourceExporter.ResourceMeta, meta interface{}) (*terraform.InstanceState, diag.Diagnostics) {
	tflog.Trace(g.ctx, fmt.Sprintf("Starting to get resource state for ID: %s, BlockLabel: %s", resID, resMeta.BlockLabel))

	// If defined, pass the full ID through the import method to generate a readable state
	instanceState := &terraform.InstanceState{ID: resMeta.IdPrefix + resID}
	tflog.Trace(g.ctx, fmt.Sprintf("Created initial instance state with ID: %s", instanceState.ID))

	tflog.Trace(g.ctx, fmt.Sprintf("Created resource mutex for ID: %s", resID))
	g.resourceStateMutex.Lock()
	tflog.Trace(g.ctx, fmt.Sprintf("Acquired mutex lock for resource ID: %s", resID))
	resourceData := resource.Data(instanceState)
	g.resourceStateMutex.Unlock()
	tflog.Trace(g.ctx, fmt.Sprintf("Released mutex lock for resource ID: %s", resID))
	tflog.Trace(g.ctx, fmt.Sprintf("Created resource data for ID: %s", resID))

	if resource.Importer != nil && resource.Importer.StateContext != nil {
		tflog.Trace(g.ctx, fmt.Sprintf("Resource has importer with StateContext, calling for ID: %s", resID))
		resourceDataArr, err := resource.Importer.StateContext(ctx, resourceData, meta)
		if err != nil {
			tflog.Error(g.ctx, fmt.Sprintf("Error with resource Importer for id %s: %v", resID, err))
			return nil, diag.FromErr(err)
		}
		if len(resourceDataArr) > 0 {
			tflog.Trace(g.ctx, fmt.Sprintf("Importer returned %d resource data entries for ID: %s", len(resourceDataArr), resID))
			instanceState = resourceDataArr[0].State()
			tflog.Trace(g.ctx, fmt.Sprintf("Updated instance state from importer for ID: %s", resID))
		} else {
			tflog.Trace(g.ctx, fmt.Sprintf("Importer returned no resource data entries for ID: %s", resID))
		}
	} else {
		tflog.Debug(g.ctx, fmt.Sprintf("Resource has no importer or StateContext for ID: %s", resID))
	}

	g.resourceStateMutex.Lock()
	tflog.Trace(g.ctx, fmt.Sprintf("Acquiring mutex lock for RefreshWithoutUpgrade for ID: %s", resID))
	state, err := resource.RefreshWithoutUpgrade(ctx, instanceState, meta)
	g.resourceStateMutex.Unlock()
	tflog.Trace(g.ctx, fmt.Sprintf("Released mutex lock after RefreshWithoutUpgrade for ID: %s", resID))

	if err != nil {
		tflog.Error(g.ctx, fmt.Sprintf("Error during RefreshWithoutUpgrade for resource %s: %v", resID, err))
		if strings.Contains(fmt.Sprintf("%v", err), "API Error: 404") ||
			strings.Contains(fmt.Sprintf("%v", err), "API Error: 410") {
			tflog.Info(g.ctx, fmt.Sprintf("Resource not found (404/410 error) for ID: %s, returning nil", resID))
			return nil, nil
		}
		tflog.Error(g.ctx, fmt.Sprintf("Non-404/410 error during RefreshWithoutUpgrade for resource %s: %v", resID, err))
		return nil, err
	}

	if state == nil || state.ID == "" {
		// Resource no longer exists
		tflog.Trace(g.ctx, fmt.Sprintf("Empty State for resource %s, state: %v", resID, state))
		return nil, nil
	}

	tflog.Debug(g.ctx, fmt.Sprintf("Successfully retrieved state for resource %s with ID: %s", resID, state.ID))
	return state, nil
}

func correctCustomFunctions(config string) string {
	config = correctInterpolatedFileShaFunctions(config)
	return correctDependsOn(config, true)
}

// find & replace ${filesha256(\"...\")} with ${filesha256("...")}
func correctInterpolatedFileShaFunctions(config string) string {
	correctedConfig := config
	re := regexp.MustCompile(`\$\{filesha256\(\\"[^\}]*\}`)
	matches := re.FindAllString(config, -1)
	for _, match := range matches {
		correctedMatch := strings.Replace(match, `\"`, `"`, -1)
		correctedConfig = strings.Replace(correctedConfig, match, correctedMatch, -1)
	}
	return correctedConfig
}

// terraform doesn't accept quotes references in HCL https://discuss.hashicorp.com/t/terraform-0-12-14-released/3898
// Added a corrected HCL during export and also for JSON export
func correctDependsOn(config string, isHcl bool) string {
	correctedConfig := config
	re := regexp.MustCompile(`"\$dep\$([^$]+)\$dep\$"`)
	matches := re.FindAllString(config, -1)

	for _, match := range matches {
		value := re.FindStringSubmatch(match)
		if len(value) == 2 {
			if !isHcl {
				correctedConfig = strings.Replace(correctedConfig, match, fmt.Sprintf(`"%s"`, value[1]), -1)
			} else {
				correctedConfig = strings.Replace(correctedConfig, match, value[1], -1)
			}
		}
	}

	return correctedConfig
}

func (g *GenesysCloudResourceExporter) sanitizeDataConfigMap(
	configMap map[string]interface{}) {

	for key, val := range configMap {
		if key == "id" {
			// Strip off IDs from the root data source
			delete(configMap, key)
		}
		if val == nil {
			delete(configMap, key)
		}
	}
}

// Removes empty and zero-valued attributes from the JSON config.
// Map attributes are removed by setting them to null, as the Terraform
// attribute syntax requires attributes be set to null
// that would otherwise be optional in nested block form:
// https://www.terraform.io/docs/language/attr-as-blocks.html#arbitrary-expressions-with-argument-syntax
func (g *GenesysCloudResourceExporter) sanitizeConfigMap(
	resource resourceExporter.ResourceInfo,
	configMap map[string]interface{},
	prevAttr string,
	exporters map[string]*resourceExporter.ResourceExporter, //Map of all exporters
	exportingState bool,
	exportFormat string,
	parentKey bool) ([]unresolvableAttributeInfo, bool) {
	resourceType := resource.Type
	resourceLabel := resource.BlockLabel
	resourceBlockType := resource.BlockType
	exporter := exporters[resourceType] //Get the specific export that we will be working with
	resourceId := resource.State.ID
	unresolvableAttrs := make([]unresolvableAttributeInfo, 0)

	for attributeConfigKey, val := range configMap {
		fullAttributePath := attributeConfigKey
		wildcardAttr := "*"
		if prevAttr != "" {
			fullAttributePath = prevAttr + "." + attributeConfigKey
			wildcardAttr = prevAttr + "." + "*"
		}

		// Identify configMap for the parent resource and add depends_on for the parent resource
		if parentKey {
			if fullAttributePath == "id" {
				g.addDependsOnValues(val.(string), configMap)
			}
		}

		if fullAttributePath == "id" {
			// Strip off IDs from the root resource
			delete(configMap, fullAttributePath)
			continue
		}

		if exporter.IsAttributeExcluded(fullAttributePath) {
			// Excluded. Remove from the config.
			configMap[attributeConfigKey] = nil
			continue
		}

		if exporter.IsAttributeE164(fullAttributePath) {
			if _, ok := configMap[attributeConfigKey].(string); !ok {
				continue
			}
			configMap[attributeConfigKey] = sanitizeE164Number(configMap[attributeConfigKey].(string))
		}

		if exporter.IsAttributeRrule(fullAttributePath) {
			if _, ok := configMap[attributeConfigKey].(string); !ok {
				continue
			}
			configMap[attributeConfigKey] = sanitizeRrule(configMap[attributeConfigKey].(string))
		}

		if exporter.RemoveFieldIfSelfReferential(resourceId, fullAttributePath, attributeConfigKey, configMap) {
			// Remove if self-referential
			configMap[attributeConfigKey] = nil
			continue
		}

		switch val.(type) {
		case map[string]interface{}:
			// Maps are sanitized in-place
			currMap := val.(map[string]interface{})
			_, res := g.sanitizeConfigMap(resource, val.(map[string]interface{}), fullAttributePath, exporters, exportingState, exportFormat, false)
			if !res || len(currMap) == 0 {
				// Remove empty maps or maps indicating they should be removed
				configMap[attributeConfigKey] = nil
			}
		case []interface{}:
			if arr := g.sanitizeConfigArray(resource, val.([]interface{}), fullAttributePath, exporters, exportingState, exportFormat); len(arr) > 0 {
				configMap[attributeConfigKey] = arr
			} else {
				// Remove empty arrays
				configMap[attributeConfigKey] = nil
			}
		case string:
			// Check if string contains nested Ref Attributes (can occur if the string is escaped json)
			if _, ok := exporter.ContainsNestedRefAttrs(fullAttributePath); ok {
				resolvedJsonString, err := g.resolveRefAttributesInJsonString(fullAttributePath, val.(string), exporter, exporters, exportingState)
				if err != nil {
					tflog.Error(g.ctx, err.Error())
				} else {
					keys := strings.Split(fullAttributePath, ".")
					configMap[keys[len(keys)-1]] = resolvedJsonString
					break
				}
			}

			// Check if we are on a reference attribute and update as needed
			refSettings := exporter.GetRefAttrSettings(fullAttributePath)
			if refSettings == nil {
				// Check for wildcard attribute indicating all attributes in the map
				refSettings = exporter.GetRefAttrSettings(wildcardAttr)
			}

			if refSettings != nil {
				configMap[attributeConfigKey] = g.resolveReference(refSettings, val.(string), exporters, exportingState)
			} else {
				configMap[attributeConfigKey] = escapeString(val.(string))
			}

			// custom function to resolve the field to a data source depending on the value
			g.resolveValueToDataSource(exporter, configMap, fullAttributePath, val)
		}

		if attr, ok := attrInUnResolvableAttrs(attributeConfigKey, exporter.UnResolvableAttributes); ok {
			if resourceBlockType != "data" {
				varReference := fmt.Sprintf("%s_%s_%s", resourceType, resourceLabel, attributeConfigKey)
				unresolvableAttrs = append(unresolvableAttrs, unresolvableAttributeInfo{
					ResourceType:  resourceType,
					ResourceLabel: resourceLabel,
					Name:          attributeConfigKey,
					Schema:        attr,
				})
				if properties, ok := attr.Elem.(*schema.Resource); ok {
					propertiesMap := make(map[string]interface{})
					for k := range properties.Schema {
						propertiesMap[k] = fmt.Sprintf("${var.%s.%s}", varReference, k)
					}
					configMap[attributeConfigKey] = propertiesMap
				} else {
					configMap[attributeConfigKey] = fmt.Sprintf("${var.%s}", varReference)
				}
			}
		}

		// The plugin SDK does not yet have a concept of "null" for unset attributes, so they are saved in state as their "zero value".
		// This can cause invalid config files due to including attributes with limits that don't allow for zero values, so we remove
		// those attributes from the config by default. Attributes can opt-out of this behavior by being added to a ResourceExporter's
		// AllowZeroValues list.
		if !exporter.AllowForZeroValues(fullAttributePath) && !exporter.AllowForZeroValuesInMap(prevAttr) {
			removeZeroValues(attributeConfigKey, configMap[attributeConfigKey], configMap)
		}

		// Nil arrays will be turned into empty arrays if they're defined in AllowEmptyArrays.
		// We do this after the initial sanitization of empty arrays to nil
		// so this will cover both cases where the attribute on the state is: null or [].
		if exporter.AllowForEmptyArrays(fullAttributePath) {
			if configMap[attributeConfigKey] == nil {
				configMap[attributeConfigKey] = []interface{}{}
			}
		}

		//If the exporter as has customer resolver for an attribute, invoke it.
		if refAttrCustomResolver, ok := exporter.CustomAttributeResolver[fullAttributePath]; ok {
			tflog.Debug(g.ctx, fmt.Sprintf("Custom resolver invoked for attribute: %s", fullAttributePath))
			if resolverFunc := refAttrCustomResolver.ResolverFunc; resolverFunc != nil {
				if err := resolverFunc(configMap, exporters, resourceLabel); err != nil {
					tflog.Error(g.ctx, fmt.Sprintf("An error has occurred while trying invoke a custom resolver for attribute %s: %v", fullAttributePath, err))
				}
			}
		}

		if g.matchesExportFormat("/.*"+formatHCL+".*/") && exporter.IsJsonEncodable(fullAttributePath) {
			if vStr, ok := configMap[attributeConfigKey].(string); ok {
				decodedData, err := getDecodedData(vStr, fullAttributePath)
				if err != nil {
					tflog.Error(g.ctx, fmt.Sprintf("Error decoding JSON string: %v\n", err))
					configMap[attributeConfigKey] = vStr
				} else {
					uid := uuid.NewString()
					attributesDecoded[uid] = decodedData
					configMap[attributeConfigKey] = uid
				}
			}
		}
	}

	if exporter.RemoveFieldIfMissing(prevAttr, configMap) {
		// Missing some inner attributes causes the outer object to be removed
		return unresolvableAttrs, false
	}

	return unresolvableAttrs, true
}

// resolveValueToDataSource invokes a custom resolver method to add a data source to the export and
// update an attribute to reference the data source
func (g *GenesysCloudResourceExporter) resolveValueToDataSource(exporter *resourceExporter.ResourceExporter, configMap map[string]any, attribute string, originalValue any) {
	// return if ResolveToDataSourceFunc does not exist for this attribute
	refAttrCustomResolver, ok := exporter.CustomAttributeResolver[attribute]
	if !ok {
		return
	}
	resolveToDataSourceFunc := refAttrCustomResolver.ResolveToDataSourceFunc
	if resolveToDataSourceFunc == nil {
		return
	}

	sdkConfig := g.meta.(*provider.ProviderMeta).ClientConfig
	dataSourceType, dataSourceLabel, dataSourceConfig, resolve := resolveToDataSourceFunc(configMap, originalValue, sdkConfig)
	if !resolve {
		return
	}

	if g.dataSourceTypesMaps[dataSourceType] == nil {
		g.dataSourceTypesMaps[dataSourceType] = make(ResourceJSONMaps)
	}

	// add the data source to the export if it hasn't already been added
	if _, ok := g.dataSourceTypesMaps[dataSourceType][dataSourceLabel]; ok {
		return
	}
	g.dataSourceTypesMaps[dataSourceType][dataSourceLabel] = dataSourceConfig
}

func attrInUnResolvableAttrs(a string, myMap map[string]*schema.Schema) (*schema.Schema, bool) {
	for k, v := range myMap {
		if k == a {
			return v, true
		}
	}
	return nil, false
}

func removeZeroValues(key string, val interface{}, configMap util.JsonMap) {
	if val == nil || reflect.TypeOf(val).String() == "bool" {
		return
	}
	if reflect.ValueOf(val).IsZero() {
		configMap[key] = nil
	}
}

// Identify the parent config map and if the resources have further dependent resources add a new attribute depends_on
func (g *GenesysCloudResourceExporter) addDependsOnValues(key string, configMap util.JsonMap) {
	list, exists := g.dependsList[key]
	if !exists {
		return
	}

	// Build a quick lookup map from resource ID to resource
	resourceMap := make(map[string]resourceExporter.ResourceInfo)
	for _, resource := range g.resources {
		resourceMap[resource.State.ID] = resource
	}

	resource, found := resourceMap[key]
	if found && g.isDataSource(resource.Type, resource.BlockLabel, resource.OriginalLabel) {
		return
	}

	resourceDependsList := make([]string, 0)

	for _, res := range list {
		parts := strings.SplitN(res, ".", 2)
		if len(parts) != 2 {
			continue
		}
		prefix, id := parts[0], parts[1]

		resource, found := resourceMap[id]
		if !found {
			continue
		}

		resourceName := fmt.Sprintf("%s.%s", prefix, resource.BlockLabel)
		if g.isDataSource(resource.Type, resource.BlockLabel, resource.OriginalLabel) {
			resourceName = "data." + resourceName
		}

		resourceDependsList = append(resourceDependsList, fmt.Sprintf("$dep$%s$dep$", resourceName))
	}

	if len(resourceDependsList) > 0 {
		configMap["depends_on"] = resourceDependsList
	}
}

func escapeString(strValue string) string {
	// Check for any '${' or '%{' in the exported string and escape them
	// https://www.terraform.io/docs/language/expressions/strings.html#escape-sequences
	escapedVal := strings.ReplaceAll(strValue, "${", "$${")
	escapedVal = strings.ReplaceAll(escapedVal, "%{", "%%{")
	return escapedVal
}

func (g *GenesysCloudResourceExporter) sanitizeConfigArray(
	resource resourceExporter.ResourceInfo,
	anArray []interface{},
	currAttr string,
	exporters map[string]*resourceExporter.ResourceExporter,
	exportingState bool,
	exportFormat string) []interface{} {
	resourceType := resource.Type
	exporter := exporters[resourceType]
	result := []interface{}{}
	for _, val := range anArray {
		switch val.(type) {
		case map[string]interface{}:
			// Only include in the result if sanitizeConfigMap returns true and the map is not empty
			currMap := val.(map[string]interface{})
			_, res := g.sanitizeConfigMap(resource, currMap, currAttr, exporters, exportingState, exportFormat, false)
			if res && len(currMap) > 0 {
				result = append(result, val)
			}
		case []interface{}:
			if arr := g.sanitizeConfigArray(resource, val.([]interface{}), currAttr, exporters, exportingState, exportFormat); len(arr) > 0 {
				result = append(result, arr)
			}
		case string:
			// Check if we are on a reference attribute and update value in array

			if refSettings := exporter.GetRefAttrSettings(currAttr); refSettings != nil {
				referenceVal := g.resolveReference(refSettings, val.(string), exporters, exportingState)
				if referenceVal != "" {
					result = append(result, referenceVal)
				}
			} else {
				result = append(result, escapeString(val.(string)))
			}
		default:
			result = append(result, val)
		}
	}
	return result
}

func (g *GenesysCloudResourceExporter) populateConfigExcluded(exporters map[string]*resourceExporter.ResourceExporter, configExcluded []string) diag.Diagnostics {
	for _, excluded := range configExcluded {
		matchFound := false
		resourceIdx := strings.Index(excluded, ".")
		if resourceIdx == -1 {
			return diag.Errorf("Invalid excluded_attribute %s", excluded)
		}

		if len(excluded) == resourceIdx {
			return diag.Errorf("excluded_attributes value %s does not contain an attribute", excluded)
		}

		resourceTypePattern := excluded[:resourceIdx]
		// identify all the resource types which match the regex
		exporter := exporters[resourceTypePattern]
		if exporter == nil {
			for resourceType, exporter1 := range exporters {
				match, _ := regexp.MatchString(resourceTypePattern, resourceType)

				if match {
					excludedAttr := excluded[resourceIdx+1:]
					exporter1.AddExcludedAttribute(excludedAttr)
					tflog.Info(g.ctx, fmt.Sprintf("Excluding attribute %s on %s resources.", excludedAttr, resourceTypePattern))
					matchFound = true
					continue
				}
			}

			if !matchFound {
				if g.addDependsOn {
					excludedAttr := excluded[resourceIdx+1:]
					tflog.Warn(g.ctx, fmt.Sprintf("Ignoring exclude attribute %s on %s resources. Since exporter is not retrieved", excludedAttr, resourceTypePattern))
					continue
				} else {
					return diag.Errorf("Resource %s in excluded_attributes is not being exported.", resourceTypePattern)
				}
			}
		} else {
			excludedAttr := excluded[resourceIdx+1:]
			exporter.AddExcludedAttribute(excludedAttr)
			tflog.Info(g.ctx, fmt.Sprintf("Excluding attribute %s on %s resources.", excludedAttr, resourceTypePattern))
		}
	}
	return nil
}

func (g *GenesysCloudResourceExporter) resolveReference(refSettings *resourceExporter.RefAttrSettings, refID string, exporters map[string]*resourceExporter.ResourceExporter, exportingState bool) string {
	if lists.ItemInSlice(refID, refSettings.AltValues) {
		// This is not actually a reference to another object. Keep the value
		return refID
	}

	if exporters[refSettings.RefType] != nil {
		// Get the sanitized label from the ID returned as a reference expression
		if idMetaMap := exporters[refSettings.RefType].SanitizedResourceMap; idMetaMap != nil {
			meta := idMetaMap[refID]
			if meta != nil && meta.BlockLabel != "" {
				if g.isDataSource(refSettings.RefType, meta.BlockLabel, meta.OriginalLabel) && g.resourceIdExists(refID, nil) {
					return fmt.Sprintf("${%s.%s.%s.id}", "data", refSettings.RefType, meta.BlockLabel)
				}
				if g.resourceIdExists(refID, nil) {
					return fmt.Sprintf("${%s.%s.id}", refSettings.RefType, meta.BlockLabel)
				}
			}
		}
	}
	g.buildSecondDepsMutex.Lock()
	defer g.buildSecondDepsMutex.Unlock()

	if g.buildSecondDeps == nil || len(g.buildSecondDeps) == 0 {
		g.buildSecondDeps = make(map[string][]string)
	}
	guidList, exists := g.buildSecondDeps[refSettings.RefType]
	if exists {
		// Check if refID already exists in list
		present := false
		for _, element := range guidList {
			if element == refID {
				present = true // String found in the slice
			}
		}
		if !present {
			g.buildSecondDeps[refSettings.RefType] = append(guidList, refID)
		}
	} else {
		g.buildSecondDeps[refSettings.RefType] = []string{refID}
	}

	if exportingState {
		// Don't remove unmatched IDs when exporting state. This will keep existing config in an org
		return refID
	}
	// No match found. Remove the value from the config since we do not have a reference to use
	return ""
}

func (g *GenesysCloudResourceExporter) resourceIdExists(refID string, existingResources []resourceExporter.ResourceInfo) bool {
	if !g.addDependsOn {
		return true
	}
	if existingResources != nil {
		for _, resource := range existingResources {
			if refID == resource.State.ID {
				return true
			}
		}
	}
	// Thread-safe access to resources
	resources := g.getResources()
	for _, resource := range resources {
		if refID == resource.State.ID {
			return true
		}
	}
	tflog.Warn(g.ctx, fmt.Sprintf("Resource present in sanitizedConfigMap and not present in resources section %v", refID))
	return false
}

func (g *GenesysCloudResourceExporter) isDataSource(resType string, resLabel, originalLabel string) bool {
	// Thread-safe access to replaceWithDatasource
	g.replaceWithDatasourceMutex.Lock()
	defer g.replaceWithDatasourceMutex.Unlock()
	return g.containsElementUnsafe(g.replaceWithDatasource, resType, resLabel, originalLabel)
}

// containsElementUnsafe is not thread-safe and should only be called with proper locking
func (g *GenesysCloudResourceExporter) containsElementUnsafe(elements []string, resType, resLabel, originalLabel string) bool {
	for _, element := range elements {
		if element == resType+"::"+resLabel || fetchByRegex(element, resType, resLabel, originalLabel) {
			return true
		}
	}
	return false
}

func fetchByRegex(fullString string, resType string, resLabel, originalLabel string) bool {
	if strings.Contains(fullString, "::") && strings.Split(fullString, "::")[0] == resType {
		i := strings.Index(fullString, "::")
		regexStr := fullString[i+2:]

		match := matchRegex(regexStr, resLabel)
		// If filter label matches original label
		if match {
			return match
		}

		if originalLabel != "" {
			// If filter label matches original label
			match := matchRegex(regexStr, originalLabel)
			return match
		}
	}
	return false
}

func matchRegex(regexStr string,
	label string) bool {
	sanitizer := resourceExporter.NewSanitizerProvider()
	match, _ := regexp.MatchString(regexStr, label)
	if match {
		return match
	}
	sanitizedMatch, _ := regexp.MatchString(regexStr, sanitizer.S.SanitizeResourceBlockLabel(label))
	return sanitizedMatch
}

func (g *GenesysCloudResourceExporter) verifyTerraformState() diag.Diagnostics {

	if exists := featureToggles.StateComparisonTrue(); exists {
		if g.matchesExportFormat("/.*" + formatHCL + ".*/") {
			tfstatePath, _ := getFilePath(g.d, defaultTfStateFile)
			hclExporter := NewTfStateExportReader(tfstatePath, g.exportDirPath)
			hclExporter.compareExportAndTFState()
		}
	}

	return nil
}

func (g *GenesysCloudResourceExporter) matchesExportFormat(formats ...string) bool {
	// Normalize format first
	exportFormat := g.exportFormat
	if exportFormat == formatHCLJSON {
		exportFormat = formatJSONHCL
	}

	// Check against normalized format
	for _, format := range formats {
		if strings.HasPrefix(format, "/") && strings.HasSuffix(format, "/") {
			pattern := strings.Trim(format, "/")
			regex, err := regexp.Compile(pattern)
			if err != nil {
				tflog.Warn(g.ctx, fmt.Sprintf("Invalid regex pattern: %s", pattern))
				continue
			}
			if regex.MatchString(exportFormat) {
				return true
			}
		} else if exportFormat == format {
			return true
		}
	}
	return false
}

// Helper methods for thread-safe access to shared state
func (g *GenesysCloudResourceExporter) addReplaceWithDatasource(item string) {
	g.replaceWithDatasourceMutex.Lock()
	defer g.replaceWithDatasourceMutex.Unlock()
	g.replaceWithDatasource = append(g.replaceWithDatasource, item)
}

func (g *GenesysCloudResourceExporter) addResources(newResources []resourceExporter.ResourceInfo) {
	g.resourcesMutex.Lock()
	defer g.resourcesMutex.Unlock()
	g.resources = append(g.resources, newResources...)
}

func (g *GenesysCloudResourceExporter) addUnresolvedAttrs(attrs []unresolvableAttributeInfo) {
	g.unresolvedAttrsMutex.Lock()
	defer g.unresolvedAttrsMutex.Unlock()
	g.unresolvedAttrs = append(g.unresolvedAttrs, attrs...)
}

func (g *GenesysCloudResourceExporter) setResourceTypesMaps(maps map[string]ResourceJSONMaps) {
	g.resourceTypesMapsMutex.Lock()
	defer g.resourceTypesMapsMutex.Unlock()
	g.resourceTypesMaps = maps
}

func (g *GenesysCloudResourceExporter) setDataSourceTypesMaps(maps map[string]ResourceJSONMaps) {
	g.dataSourceTypesMapsMutex.Lock()
	defer g.dataSourceTypesMapsMutex.Unlock()
	g.dataSourceTypesMaps = maps
}

func (g *GenesysCloudResourceExporter) getResourceTypesMaps() map[string]ResourceJSONMaps {
	g.resourceTypesMapsMutex.RLock()
	defer g.resourceTypesMapsMutex.RUnlock()
	return g.resourceTypesMaps
}

func (g *GenesysCloudResourceExporter) getDataSourceTypesMaps() map[string]ResourceJSONMaps {
	g.dataSourceTypesMapsMutex.RLock()
	defer g.dataSourceTypesMapsMutex.RUnlock()
	return g.dataSourceTypesMaps
}

func (g *GenesysCloudResourceExporter) getResources() []resourceExporter.ResourceInfo {
	g.resourcesMutex.Lock()
	defer g.resourcesMutex.Unlock()
	return g.resources
}

func (g *GenesysCloudResourceExporter) getUnresolvedAttrs() []unresolvableAttributeInfo {
	g.unresolvedAttrsMutex.Lock()
	defer g.unresolvedAttrsMutex.Unlock()
	return g.unresolvedAttrs
}
