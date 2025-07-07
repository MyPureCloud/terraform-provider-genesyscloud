package tfexporter

import (
	"archive/zip"
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"log"
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mohae/deepcopy"

	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
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

type GenesysCloudResourceExporter struct {
	configExporter        Exporter
	filterType            ExporterFilterType
	resourceTypeFilter    ExporterResourceTypeFilter
	resourceFilter        ExporterResourceFilter
	filterList            *[]string
	exportFormat          string
	splitFilesByResource  bool
	logPermissionErrors   bool
	addDependsOn          bool
	replaceWithDatasource []string
	includeStateFile      bool
	version               string
	providerRegistry      string
	provider              *schema.Provider
	exportDirPath         string
	exporters             *map[string]*resourceExporter.ResourceExporter
	resources             []resourceExporter.ResourceInfo
	resourceTypesMaps     map[string]resourceJSONMaps
	dataSourceTypesMaps   map[string]resourceJSONMaps
	unresolvedAttrs       []unresolvableAttributeInfo
	d                     *schema.ResourceData
	ctx                   context.Context
	meta                  interface{}
	dependsList           map[string][]string
	buildSecondDeps       map[string][]string
	exMutex               sync.RWMutex
	cyclicDependsList     []string
	ignoreCyclicDeps      bool
	flowResourcesList     []string
	exportComputed        bool
	maxConcurrentOps      int // New field to control concurrency

	// New mutexes for protecting shared state
	replaceWithDatasourceMutex sync.Mutex
	resourcesMutex             sync.Mutex
	resourceTypesMapsMutex     sync.RWMutex
	dataSourceTypesMapsMutex   sync.RWMutex
	unresolvedAttrsMutex       sync.Mutex
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
		resourceTypesMaps:     make(map[string]resourceJSONMaps),
		dataSourceTypesMaps:   make(map[string]resourceJSONMaps),
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
	diagErr = append(diagErr, g.retrieveExporters()...)
	if diagErr.HasError() {
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
	return diagErr
}

func (g *GenesysCloudResourceExporter) setUpExportDirPath() (diagErr diag.Diagnostics) {
	log.Printf("Setting up export directory path")

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
	log.Printf("Retrieving exporters list")
	exports := resourceExporter.GetResourceExporters()

	log.Printf("Retrieving exporters list %v", g.filterList)

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

// retrieveSanitizedResourceMaps will retrieve a list of all of the resources to be exported.  It will also apply a filter (e.g the :: ) and only return the specific Genesys Cloud
// resources that are specified via :: delimiter
func (g *GenesysCloudResourceExporter) retrieveSanitizedResourceMaps() (diagErr diag.Diagnostics) {
	log.Printf("Retrieving map of Genesys Cloud resources to export")
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

	//Retrieve a map of all of the objects we are going to build.  Apply the filter that will remove specific classes of an object
	diagErr = g.buildSanitizedResourceMaps(*g.exporters, newFilter, g.logPermissionErrors)
	if diagErr != nil {
		return diagErr
	}

	//Check to see if we found any exporters.  If we did find the exporter
	if len(*g.exporters) == 0 {
		return diag.Errorf("No valid resource types to export.")
	}

	return nil
}

// retrieveGenesysCloudObjectInstances will take a list of exporters and then return the actual terraform Genesys Cloud data
func (g *GenesysCloudResourceExporter) retrieveGenesysCloudObjectInstances() diag.Diagnostics {
	log.Printf("[retrieveGenesysCloudObjectInstances] Starting to retrieve Genesys Cloud objects from Genesys Cloud")
	log.Printf("[retrieveGenesysCloudObjectInstances] Number of exporters to process: %d", len(*g.exporters))

	// Log all exporter types being processed
	for resType, exporter := range *g.exporters {
		log.Printf("[retrieveGenesysCloudObjectInstances] Exporter for %s has %d resources in SanitizedResourceMap", resType, len(exporter.SanitizedResourceMap))
	}

	// Retrieves data on each individual Genesys Cloud object from each registered exporter

	// Buffer error channel to prevent goroutine leaks or deadlocks
	errorChan := make(chan diag.Diagnostics, len(*g.exporters))
	wgDone := make(chan bool)
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(g.ctx)
	defer cancel()
	log.Printf("[retrieveGenesysCloudObjectInstances] Created context with cancellation")

	// Track successful and failed resource types
	var successfulTypes []string
	var failedTypes []string
	var statsMutex sync.Mutex

	// We use concurrency here to spin off each exporter type and getting the data
	for resType, exporter := range *g.exporters {
		log.Printf("[retrieveGenesysCloudObjectInstances] Starting processing for resource type: %s", resType)
		wg.Add(1)
		go func(resType string, exporter *resourceExporter.ResourceExporter) {
			defer wg.Done()
			log.Printf("[retrieveGenesysCloudObjectInstances] Starting goroutine for resource type: %s", resType)

			// Check if context was cancelled before processing
			select {
			case <-ctx.Done():
				log.Printf("[retrieveGenesysCloudObjectInstances] Context cancelled before processing resource type: %s", resType)
				return
			default:
			}

			log.Printf("[retrieveGenesysCloudObjectInstances] Getting exported resources for [%s]", resType)
			typeResources, err := g.getResourcesForType(resType, g.provider, exporter, g.meta)

			if err != nil {
				log.Printf("[retrieveGenesysCloudObjectInstances] ERROR getting resources for type %s: %v", resType, err)

				// Track failed resource type
				statsMutex.Lock()
				failedTypes = append(failedTypes, resType)
				statsMutex.Unlock()

				select {
				case <-ctx.Done():
					log.Printf("[retrieveGenesysCloudObjectInstances] Context cancelled while handling error for %s", resType)
				case errorChan <- err:
					log.Printf("[retrieveGenesysCloudObjectInstances] Successfully sent error to channel for resource type: %s", resType)
				default:
					log.Printf("[retrieveGenesysCloudObjectInstances] ERROR: Could not send error to channel for resource type: %s (channel full)", resType)
				}
				cancel()
				return
			}

			log.Printf("[retrieveGenesysCloudObjectInstances] Successfully retrieved %d resources for type %s", len(typeResources), resType)

			// Use thread-safe method to add resources
			if len(typeResources) > 0 {
				g.addResources(typeResources)
				log.Printf("[retrieveGenesysCloudObjectInstances] Successfully added %d resources for type %s to global resources list", len(typeResources), resType)

				// Track successful resource type
				statsMutex.Lock()
				successfulTypes = append(successfulTypes, resType)
				statsMutex.Unlock()
			} else {
				log.Printf("[retrieveGenesysCloudObjectInstances] WARNING: No resources found for type %s", resType)
			}

			log.Printf("[retrieveGenesysCloudObjectInstances] Completed processing for resource type: %s", resType)
		}(resType, exporter)
	}

	log.Printf("[retrieveGenesysCloudObjectInstances] Started all goroutines, waiting for completion")

	go func() {
		wg.Wait()
		log.Printf("[retrieveGenesysCloudObjectInstances] All resource type processing completed")
		close(wgDone)
	}()

	// Wait until either WaitGroup is done or an error is received
	select {
	case <-wgDone:
		log.Printf("[retrieveGenesysCloudObjectInstances] Successfully retrieved all Genesys Cloud object instances")
		log.Printf("[retrieveGenesysCloudObjectInstances] Summary - Successful types: %v", successfulTypes)
		log.Printf("[retrieveGenesysCloudObjectInstances] Summary - Failed types: %v", failedTypes)
		log.Printf("[retrieveGenesysCloudObjectInstances] Summary - Total successful: %d, Total failed: %d", len(successfulTypes), len(failedTypes))
		return nil
	case err := <-errorChan:
		log.Printf("[retrieveGenesysCloudObjectInstances] Error received from channel: %v", err)

		// Give other goroutines a chance to clean up
		go func() {
			<-wgDone // Wait for all goroutines to finish
			log.Printf("[retrieveGenesysCloudObjectInstances] All goroutines finished after error")
		}()

		log.Printf("[retrieveGenesysCloudObjectInstances] Returning error: %v", err)
		return err
	}
}

// buildResourceConfigMap Builds a map of all the Terraform resources data returned for each resource
func (g *GenesysCloudResourceExporter) buildResourceConfigMap() (diagnostics diag.Diagnostics) {
	log.Printf("Build Genesys Cloud Resources Map")

	// Initialize maps using thread-safe methods
	g.setResourceTypesMaps(make(map[string]resourceJSONMaps))
	g.setDataSourceTypesMaps(make(map[string]resourceJSONMaps))

	// Get resources using thread-safe method
	resources := g.getResources()

	for _, resource := range resources {
		// 1. Get instance state as JSON Map
		jsonResult, diagErr := g.instanceStateToMap(resource.State, resource.CtyType)
		if diagErr != nil {
			return diagErr
		}

		// 2. Determine if instance is a data source
		isDataSource := g.isDataSource(resource.Type, resource.BlockLabel, resource.OriginalLabel)
		if isDataSource {
			dataSourceMaps := g.getDataSourceTypesMaps()
			if dataSourceMaps[resource.Type] == nil {
				dataSourceMaps[resource.Type] = make(resourceJSONMaps)
			}
			g.setDataSourceTypesMaps(dataSourceMaps)
		} else {
			// 3. Ensure the resource type is instantiated
			resourceMaps := g.getResourceTypesMaps()
			if resourceMaps[resource.Type] == nil {
				resourceMaps[resource.Type] = make(resourceJSONMaps)
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
		configMap := make(map[string]interface{})
		for key, value := range jsonResult {
			configMap[key] = value
		}

		// 5. Sanitize the config map
		unresolvableAttrs, _ := g.sanitizeConfigMap(resource, configMap, "", *g.exporters, false, g.exportFormat, false)
		if len(unresolvableAttrs) > 0 {
			g.addUnresolvedAttrs(unresolvableAttrs)
		}

		// 6. Add the resource to the appropriate map
		if isDataSource {
			dataSourceMaps = g.getDataSourceTypesMaps()
			dataSourceMaps[resource.Type][resource.BlockLabel] = configMap
			g.setDataSourceTypesMaps(dataSourceMaps)
		} else {
			resourceMaps = g.getResourceTypesMaps()
			resourceMaps[resource.Type][resource.BlockLabel] = configMap
			g.setResourceTypesMaps(resourceMaps)
		}

		// 7. Update instance state attributes
		g.updateInstanceStateAttributes(jsonResult, resource)
	}

	log.Printf("Successfully built resource config map with %d resources", len(resources))
	return nil
}

func (g *GenesysCloudResourceExporter) customWriteAttributes(jsonResult util.JsonMap,
	resource resourceExporter.ResourceInfo) (diagnostics diag.Diagnostics) {
	exporters := *g.exporters
	if resourceFilesWriterFunc := exporters[resource.Type].CustomFileWriter.RetrieveAndWriteFilesFunc; resourceFilesWriterFunc != nil {
		exportDir, _ := getFilePath(g.d, "")
		if err := resourceFilesWriterFunc(resource.State.ID, exportDir, exporters[resource.Type].CustomFileWriter.SubDirectory, jsonResult, g.meta, resource); err != nil {
			log.Printf("An error has occurred while trying invoking the RetrieveAndWriteFilesFunc for resource type %s: %v", resource.Type, err)
			diagnostics = append(diagnostics, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Failed to invoke %s custom resolver method.", resource.Type),
				Detail:   err.Error(),
			})
		}
	}

	if len(exporters[resource.Type].CustomFlowResolver) > 0 {
		g.updateInstanceStateAttributes(jsonResult, resource)
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
			log.Printf("dependent consumers retrieved %v", resourceKeys.State.ID)
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
	log.Printf("rebuild exporters list")
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
	log.Printf("rebuild exports from exportDependentResources")

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
		if diagErr.HasError() {
			return diagErr
		}
	}
	return nil
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
			log.Printf("Deleted removeId %v", removeId)
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
			g.chainDependencies(make([]resourceExporter.ResourceInfo, 0), exp)
		}
	}
	return nil
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
						log.Printf("Resource already present in the resources. %v", guid)
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
		log.Printf("rebuild exporters list from chainDependencies")
		err := g.rebuildExports(*g.filterList)
		if err != nil {
			return err
		}
		// checks and exports if there are any dependent flow resources
		err = g.buildAndExportDependsOnResourcesForFlows()
		if err != nil {
			return err
		}

		err = g.buildResourceConfigMap()
		if err.HasError() {
			return err
		}
		//append the resources and exporters
		g.appendResources(existingResources)
		g.exporters = mergeExporters(existingExporters, *g.exporters)

		// deep copy is needed here else exporters being overridden
		existingExportersInterface := deepcopy.Copy(*g.exporters)
		existingExporters, _ = existingExportersInterface.(map[string]*resourceExporter.ResourceExporter)
		existingResources = g.resources

		// Recursive call until all the dependencies are addressed.
		return g.chainDependencies(existingResources, existingExporters)
	}
	return nil
}

func (g *GenesysCloudResourceExporter) appendResources(resourcesToAdd []resourceExporter.ResourceInfo) {
	log.Printf("Appending %d resources to existing resources", len(resourcesToAdd))

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
			log.Printf("Skipping duplicate resource: %s", key)
		}
	}

	if len(newResources) > 0 {
		// Use thread-safe method to add new resources
		g.addResources(newResources)
		log.Printf("Successfully added %d new resources", len(newResources))
	} else {
		log.Printf("No new resources to add")
	}
}

func (g *GenesysCloudResourceExporter) buildSanitizedResourceMaps(exporters map[string]*resourceExporter.ResourceExporter, filter []string, logErrors bool) diag.Diagnostics {
	log.Printf("Starting to build sanitized resource maps for %d exporters", len(exporters))

	// Buffer error channel to prevent goroutine leaks or deadlocks
	errorChan := make(chan diag.Diagnostics, len(exporters))
	wgDone := make(chan bool)

	// Cancel remaining goroutines if an error occurs
	ctx, cancel := context.WithCancel(g.ctx)
	defer cancel()

	// Create semaphore to limit concurrent operations to the configured maximum
	sem := make(chan struct{}, g.maxConcurrentOps)

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

			log.Printf("Getting all resources for type %s", resourceType)
			exporter.FilterResource = g.resourceFilter

			err := exporter.LoadSanitizedResourceMap(ctx, resourceType, filter)

			// Used in tests
			if mockError != nil {
				err = mockError
			}
			if containsPermissionsErrorOnly(err) && logErrors {
				log.Printf("%v", err[0].Summary)
				log.Printf("Logging permission error for %s. Resuming export...", resourceType)
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
			log.Printf("Found %d resources for type %s", len(exporter.SanitizedResourceMap), resourceType)
		}(resourceType, exporter)
	}

	go func() {
		wg.Wait()
		log.Print(`Finished building sanitized resource maps`)
		close(wgDone)
	}()

	// Wait until either WaitGroup is done or an error is received
	select {
	case <-wgDone:
		log.Printf("Successfully completed building sanitized resource maps")
		return nil
	case err := <-errorChan:
		// Give other goroutines a chance to clean up
		go func() {
			<-wgDone // Wait for all goroutines to finish
		}()
		log.Printf("Error occurred while building sanitized resource maps: %v", err)
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
	log.Printf("[getResourcesForType] Starting export for resource type: %s", resType)

	// Use thread-safe method to get resource map size
	lenResources := exporter.GetSanitizedResourceMapSize()
	log.Printf("[getResourcesForType] Found %d resources for type %s", lenResources, resType)

	if lenResources == 0 {
		log.Printf("[getResourcesForType] No resources found for type %s, returning empty slice", resType)
		return []resourceExporter.ResourceInfo{}, nil
	}

	// Buffer channels to prevent goroutine leaks
	errorChan := make(chan diag.Diagnostics, lenResources)
	resourceChan := make(chan resourceExporter.ResourceInfo, lenResources)
	log.Printf("[getResourcesForType] Created buffered channels for %d resources", lenResources)

	res := schemaProvider.ResourcesMap[resType]
	if res == nil {
		log.Printf("[getResourcesForType] ERROR: Resource type %v not defined in schema provider", resType)
		return nil, diag.Errorf("Resource type %v not defined", resType)
	}
	log.Printf("[getResourcesForType] Successfully retrieved resource schema for type %s", resType)

	exportComputed := g.exportComputed
	log.Printf("[getResourcesForType] Export computed setting: %t", exportComputed)

	ctx, cancel := context.WithCancel(g.ctx)
	defer cancel()
	log.Printf("[getResourcesForType] Created context with cancellation for resource type %s", resType)

	var wg sync.WaitGroup
	wg.Add(lenResources)
	log.Printf("[getResourcesForType] Initialized WaitGroup with %d resources", lenResources)

	// Track resources to remove to avoid race conditions
	var toRemove []string
	var toRemoveMutex sync.Mutex
	log.Printf("[getResourcesForType] Initialized resource removal tracking for type %s", resType)

	// Get a copy of the resource map to avoid race conditions during iteration
	resourceMap := exporter.GetSanitizedResourceMap()
	log.Printf("[getResourcesForType] Retrieved sanitized resource map with %d entries for type %s", len(resourceMap), resType)

	for id, resMeta := range resourceMap {
		log.Printf("[getResourcesForType] Starting goroutine for resource ID: %s, BlockLabel: %s", id, resMeta.BlockLabel)

		go func(id string, resMeta *resourceExporter.ResourceMeta) {
			defer wg.Done()
			log.Printf("[getResourcesForType] Starting processing for resource ID: %s, BlockLabel: %s", id, resMeta.BlockLabel)

			// Check if context was cancelled
			select {
			case <-ctx.Done():
				log.Printf("[getResourcesForType] Context cancelled for resource ID: %s", id)
				return
			default:
				log.Printf("[getResourcesForType] Context check passed for resource ID: %s", id)
			}

			fetchResourceState := func() error {
				log.Printf("[getResourcesForType] Starting fetchResourceState for resource ID: %s", id)

				// Use a shorter timeout for individual resource fetches
				resourceCtx, resourceCancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer resourceCancel()
				log.Printf("[getResourcesForType] Created resource context with 5-minute timeout for ID: %s", id)

				ctyType := res.CoreConfigSchema().ImpliedType()
				log.Printf("[getResourcesForType] Retrieved CTY type for resource ctyType: %v", ctyType)

				log.Printf("[getResourcesForType] Calling getResourceState for resource ID: %s", id)
				instanceState, err := getResourceState(resourceCtx, res, id, resMeta, meta)

				if err != nil {
					log.Printf("[getResourcesForType] Error while fetching read context type %s and instance %s : %v", resType, id, err)
					return fmt.Errorf("Failed to get state for %s instance %s: %v", resType, id, err)
				}

				if instanceState == nil {
					log.Printf("[getResourcesForType] Resource %s no longer exists. Skipping.", resMeta.BlockLabel)
					toRemoveMutex.Lock()
					toRemove = append(toRemove, id)
					toRemoveMutex.Unlock()
					log.Printf("[getResourcesForType] Added resource ID %s to removal list", id)
					return nil
				}
				log.Printf("[getResourcesForType] Successfully retrieved instance state for resource ID: %s", id)

				// Export the resource as a data resource
				if exporter.ExportAsDataFunc != nil {
					log.Printf("[getResourcesForType] Checking if resource should be exported as data source for ID: %s", id)
					sdkConfig := g.meta.(*provider.ProviderMeta).ClientConfig
					exportAsData, err := exporter.ExportAsDataFunc(g.ctx, sdkConfig, instanceState.Attributes)
					if err != nil {
						log.Printf("[getResourcesForType] Error in ExportAsDataFunc for resource ID %s: %v", id, err)
						return fmt.Errorf("an error has occurred while trying to export as a data resource block for %s::%s : %v", resType, resMeta.BlockLabel, err)
					} else {
						if exportAsData {
							log.Printf("[getResourcesForType] Resource ID %s will be exported as data source", id)
							g.addReplaceWithDatasource(resType + "::" + resMeta.BlockLabel)
						} else {
							log.Printf("[getResourcesForType] Resource ID %s will NOT be exported as data source", id)
						}
					}
				} else {
					log.Printf("[getResourcesForType] No ExportAsDataFunc defined for resource ID: %s", id)
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
				log.Printf("[getResourcesForType] Finished processing schema attributes for resource ID: %s", id)

				log.Printf("[getResourcesForType] Creating ResourceInfo for resource ID: %s", id)
				resourceChan <- resourceExporter.ResourceInfo{
					State:         instanceState,
					BlockLabel:    resMeta.BlockLabel,
					Type:          resType,
					CtyType:       ctyType,
					BlockType:     blockType,
					OriginalLabel: resMeta.OriginalLabel,
				}
				log.Printf("[getResourcesForType] Successfully sent ResourceInfo to channel for resource ID: %s", id)

				return nil
			}

			// Improved retry logic with exponential backoff
			maxRetries := 3
			var lastErr error
			log.Printf("[getResourcesForType] Starting retry logic for resource ID: %s (max retries: %d)", id, maxRetries)

			for attempt := 0; attempt < maxRetries; attempt++ {
				log.Printf("[getResourcesForType] Attempt %d/%d for resource ID: %s", attempt+1, maxRetries, id)

				select {
				case <-ctx.Done():
					log.Printf("[getResourcesForType] Context cancelled during retry attempt %d for resource ID: %s", attempt+1, id)
					return
				default:
				}

				err := fetchResourceState()
				if err == nil {
					log.Printf("[getResourcesForType] Successfully processed resource ID: %s on attempt %d", id, attempt+1)
					return // Success
				}

				lastErr = err
				log.Printf("[getResourcesForType] Error on attempt %d for resource ID %s: %v", attempt+1, id, err)

				// Check if it's a timeout error that we should retry
				isTimeoutError := func(err error) bool {
					return strings.Contains(fmt.Sprintf("%v", err), "timeout while waiting for state to become") ||
						strings.Contains(fmt.Sprintf("%v", err), "context deadline exceeded") ||
						strings.Contains(fmt.Sprintf("%v", err), "timeout")
				}

				if !isTimeoutError(err) {
					log.Printf("[getResourcesForType] Non-retryable error for resource ID %s, sending to error channel", id)
					// Non-retryable error, send to error channel
					select {
					case errorChan <- diag.Errorf("Failed to get state for %s instance %s: %v", resType, id, err):
						log.Printf("[getResourcesForType] Successfully sent error to channel for resource ID: %s", id)
					case <-ctx.Done():
						log.Printf("[getResourcesForType] Context cancelled while sending error for resource ID: %s", id)
						return
					}
					return
				}

				// Exponential backoff for retryable errors
				if attempt < maxRetries-1 {
					backoffDuration := time.Duration(1<<attempt) * time.Second
					log.Printf("[getResourcesForType] Retrying resource %s (attempt %d/%d) after %v backoff", id, attempt+1, maxRetries, backoffDuration)

					select {
					case <-time.After(backoffDuration):
						log.Printf("[getResourcesForType] Backoff completed for resource ID: %s", id)
					case <-ctx.Done():
						log.Printf("[getResourcesForType] Context cancelled during backoff for resource ID: %s", id)
						return
					}
				}
			}

			// If we get here, all retries failed
			if lastErr != nil {
				log.Printf("[getResourcesForType] All retries failed for resource ID %s, sending final error to channel", id)
				select {
				case errorChan <- diag.Errorf("Failed to get state for %s instance %s after %d retries: %v", resType, id, maxRetries, lastErr):
					log.Printf("[getResourcesForType] Successfully sent final error to channel for resource ID: %s", id)
				case <-ctx.Done():
					log.Printf("[getResourcesForType] Context cancelled while sending final error for resource ID: %s", id)
					return
				}
			}
		}(id, resMeta)
	}

	log.Printf("[getResourcesForType] Started all goroutines for resource type %s, waiting for completion", resType)

	// Wait for all goroutines to complete
	wg.Wait()
	log.Printf("[getResourcesForType] All goroutines completed for resource type %s", resType)

	close(resourceChan)
	close(errorChan)
	log.Printf("[getResourcesForType] Closed resource and error channels for resource type %s", resType)

	// Collect all resources
	var resources []resourceExporter.ResourceInfo
	log.Printf("[getResourcesForType] Collecting resources from channel for resource type %s", resType)
	for r := range resourceChan {
		log.Printf("[getResourcesForType] Collected resource: Type=%s, BlockLabel=%s, ID=%s", r.Type, r.BlockLabel, r.State.ID)
		resources = append(resources, r)
	}
	log.Printf("[getResourcesForType] Collected %d resources from channel for resource type %s", len(resources), resType)

	// Remove resources that weren't found using thread-safe method
	log.Printf("[getResourcesForType] Processing %d resources for removal", len(toRemove))
	for _, id := range toRemove {
		log.Printf("[getResourcesForType] Removing resource %v from export map", id)
		exporter.RemoveFromSanitizedResourceMap(id)
	}

	// Collect all errors
	var allErrors diag.Diagnostics
	log.Printf("[getResourcesForType] Collecting errors from channel for resource type %s", resType)
	for err := range errorChan {
		log.Printf("[getResourcesForType] Collected error: %v", err)
		allErrors = append(allErrors, err...)
	}
	log.Printf("[getResourcesForType] Collected %d errors for resource type %s", len(allErrors), resType)

	// Return errors if any occurred
	if len(allErrors) > 0 {
		log.Printf("[getResourcesForType] Export completed for %s with %d errors out of %d resources", resType, len(allErrors), lenResources)
		return resources, allErrors
	}

	log.Printf("[getResourcesForType] Export completed successfully for %s: %d resources successfully exported", resType, len(resources))
	return resources, nil
}

func getResourceState(ctx context.Context, resource *schema.Resource, resID string, resMeta *resourceExporter.ResourceMeta, meta interface{}) (*terraform.InstanceState, diag.Diagnostics) {
	log.Printf("[getResourceState] Starting to get resource state for ID: %s, BlockLabel: %s", resID, resMeta.BlockLabel)

	// If defined, pass the full ID through the import method to generate a readable state
	instanceState := &terraform.InstanceState{ID: resMeta.IdPrefix + resID}
	log.Printf("[getResourceState] Created initial instance state with ID: %s", instanceState.ID)

	resourceMutex := &sync.Mutex{}
	log.Printf("[getResourceState] Created resource mutex for ID: %s", resID)

	resourceMutex.Lock()
	log.Printf("[getResourceState] Acquired mutex lock for resource ID: %s", resID)
	resourceData := resource.Data(instanceState)
	resourceMutex.Unlock()
	log.Printf("[getResourceState] Released mutex lock for resource ID: %s", resID)
	log.Printf("[getResourceState] Created resource data for ID: %s", resID)

	if resource.Importer != nil && resource.Importer.StateContext != nil {
		log.Printf("[getResourceState] Resource has importer with StateContext, calling for ID: %s", resID)
		resourceDataArr, err := resource.Importer.StateContext(ctx, resourceData, meta)
		if err != nil {
			log.Printf("[getResourceState] Error with resource Importer for id %s: %v", resID, err)
			return nil, diag.FromErr(err)
		}
		if len(resourceDataArr) > 0 {
			log.Printf("[getResourceState] Importer returned %d resource data entries for ID: %s", len(resourceDataArr), resID)
			instanceState = resourceDataArr[0].State()
			log.Printf("[getResourceState] Updated instance state from importer for ID: %s", resID)
		} else {
			log.Printf("[getResourceState] Importer returned no resource data entries for ID: %s", resID)
		}
	} else {
		log.Printf("[getResourceState] Resource has no importer or StateContext for ID: %s", resID)
	}

	resourceMutex.Lock()
	log.Printf("[getResourceState] Acquiring mutex lock for RefreshWithoutUpgrade for ID: %s", resID)
	state, err := resource.RefreshWithoutUpgrade(ctx, instanceState, meta)
	resourceMutex.Unlock()
	log.Printf("[getResourceState] Released mutex lock after RefreshWithoutUpgrade for ID: %s", resID)

	if err != nil {
		log.Printf("[getResourceState] Error during RefreshWithoutUpgrade for resource %s: %v", resID, err)
		if strings.Contains(fmt.Sprintf("%v", err), "API Error: 404") ||
			strings.Contains(fmt.Sprintf("%v", err), "API Error: 410") {
			log.Printf("[getResourceState] Resource not found (404/410 error) for ID: %s, returning nil", resID)
			return nil, nil
		}
		log.Printf("[getResourceState] Non-404/410 error during RefreshWithoutUpgrade for resource %s: %v", resID, err)
		return nil, err
	}

	if state == nil || state.ID == "" {
		// Resource no longer exists
		log.Printf("[getResourceState] Empty State for resource %s, state: %v", resID, state)
		return nil, nil
	}

	log.Printf("[getResourceState] Successfully retrieved state for resource %s with ID: %s", resID, state.ID)
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

	unresolvableAttrs := make([]unresolvableAttributeInfo, 0)

	for key, val := range configMap {
		currAttr := key
		wildcardAttr := "*"
		if prevAttr != "" {
			currAttr = prevAttr + "." + key
			wildcardAttr = prevAttr + "." + "*"
		}

		// Identify configMap for the parent resource and add depends_on for the parent resource
		if parentKey {
			if currAttr == "id" {
				g.addDependsOnValues(val.(string), configMap)
			}
		}

		if currAttr == "id" {
			// Strip off IDs from the root resource
			delete(configMap, currAttr)
			continue
		}

		if exporter.IsAttributeExcluded(currAttr) {
			// Excluded. Remove from the config.
			configMap[key] = nil
			continue
		}

		if exporter.IsAttributeE164(currAttr) {
			if _, ok := configMap[key].(string); !ok {
				continue
			}
			configMap[key] = sanitizeE164Number(configMap[key].(string))
		}

		if exporter.IsAttributeRrule(currAttr) {
			if _, ok := configMap[key].(string); !ok {
				continue
			}
			configMap[key] = sanitizeRrule(configMap[key].(string))
		}

		switch val.(type) {
		case map[string]interface{}:
			// Maps are sanitized in-place
			currMap := val.(map[string]interface{})
			_, res := g.sanitizeConfigMap(resource, val.(map[string]interface{}), currAttr, exporters, exportingState, exportFormat, false)
			if !res || len(currMap) == 0 {
				// Remove empty maps or maps indicating they should be removed
				configMap[key] = nil
			}
		case []interface{}:
			if arr := g.sanitizeConfigArray(resource, val.([]interface{}), currAttr, exporters, exportingState, exportFormat); len(arr) > 0 {
				configMap[key] = arr
			} else {
				// Remove empty arrays
				configMap[key] = nil
			}
		case string:
			// Check if string contains nested Ref Attributes (can occur if the string is escaped json)
			if _, ok := exporter.ContainsNestedRefAttrs(currAttr); ok {
				resolvedJsonString, err := g.resolveRefAttributesInJsonString(currAttr, val.(string), exporter, exporters, exportingState)
				if err != nil {
					log.Println(err)
				} else {
					keys := strings.Split(currAttr, ".")
					configMap[keys[len(keys)-1]] = resolvedJsonString
					break
				}
			}

			// Check if we are on a reference attribute and update as needed
			refSettings := exporter.GetRefAttrSettings(currAttr)
			if refSettings == nil {
				// Check for wildcard attribute indicating all attributes in the map
				refSettings = exporter.GetRefAttrSettings(wildcardAttr)
			}

			if refSettings != nil {
				configMap[key] = g.resolveReference(refSettings, val.(string), exporters, exportingState)
			} else {
				configMap[key] = escapeString(val.(string))
			}

			// custom function to resolve the field to a data source depending on the value
			g.resolveValueToDataSource(exporter, configMap, currAttr, val)
		}

		if attr, ok := attrInUnResolvableAttrs(key, exporter.UnResolvableAttributes); ok {
			if resourceBlockType != "data" {
				varReference := fmt.Sprintf("%s_%s_%s", resourceType, resourceLabel, key)
				unresolvableAttrs = append(unresolvableAttrs, unresolvableAttributeInfo{
					ResourceType:  resourceType,
					ResourceLabel: resourceLabel,
					Name:          key,
					Schema:        attr,
				})
				if properties, ok := attr.Elem.(*schema.Resource); ok {
					propertiesMap := make(map[string]interface{})
					for k := range properties.Schema {
						propertiesMap[k] = fmt.Sprintf("${var.%s.%s}", varReference, k)
					}
					configMap[key] = propertiesMap
				} else {
					configMap[key] = fmt.Sprintf("${var.%s}", varReference)
				}
			}
		}

		// The plugin SDK does not yet have a concept of "null" for unset attributes, so they are saved in state as their "zero value".
		// This can cause invalid config files due to including attributes with limits that don't allow for zero values, so we remove
		// those attributes from the config by default. Attributes can opt-out of this behavior by being added to a ResourceExporter's
		// AllowZeroValues list.
		if !exporter.AllowForZeroValues(currAttr) && !exporter.AllowForZeroValuesInMap(prevAttr) {
			removeZeroValues(key, configMap[key], configMap)
		}

		// Nil arrays will be turned into empty arrays if they're defined in AllowEmptyArrays.
		// We do this after the initial sanitization of empty arrays to nil
		// so this will cover both cases where the attribute on the state is: null or [].
		if exporter.AllowForEmptyArrays(currAttr) {
			if configMap[key] == nil {
				configMap[key] = []interface{}{}
			}
		}

		//If the exporter as has customer resolver for an attribute, invoke it.
		if refAttrCustomResolver, ok := exporter.CustomAttributeResolver[currAttr]; ok {
			log.Printf("Custom resolver invoked for attribute: %s", currAttr)
			if resolverFunc := refAttrCustomResolver.ResolverFunc; resolverFunc != nil {
				if err := resolverFunc(configMap, exporters, resourceLabel); err != nil {
					log.Printf("An error has occurred while trying invoke a custom resolver for attribute %s: %v", currAttr, err)
				}
			}
		}

		// Check if the exporter has custom flow resolver (Only applicable for flow resource)
		if refAttrCustomFlowResolver, ok := exporter.CustomFlowResolver[currAttr]; ok {
			log.Printf("Custom resolver invoked for attribute: %s", currAttr)
			varReference := fmt.Sprintf("%s_%s_%s", resourceType, resourceLabel, "filepath")
			if err := refAttrCustomFlowResolver.ResolverFunc(configMap, varReference); err != nil {
				log.Printf("An error has occurred while trying invoke a custom resolver for attribute %s: %v", currAttr, err)
			}
		}

		if g.matchesExportFormat("/.*"+formatHCL+".*/") && exporter.IsJsonEncodable(currAttr) {
			if vStr, ok := configMap[key].(string); ok {
				decodedData, err := getDecodedData(vStr, currAttr)
				if err != nil {
					log.Printf("Error decoding JSON string: %v\n", err)
					configMap[key] = vStr
				} else {
					uid := uuid.NewString()
					attributesDecoded[uid] = decodedData
					configMap[key] = uid
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
		g.dataSourceTypesMaps[dataSourceType] = make(resourceJSONMaps)
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
					log.Printf("Excluding attribute %s on %s resources.", excludedAttr, resourceTypePattern)
					matchFound = true
					continue
				}
			}

			if !matchFound {
				if g.addDependsOn {
					excludedAttr := excluded[resourceIdx+1:]
					log.Printf("Ignoring exclude attribute %s on %s resources. Since exporter is not retrieved", excludedAttr, resourceTypePattern)
					continue
				} else {
					return diag.Errorf("Resource %s in excluded_attributes is not being exported.", resourceTypePattern)
				}
			}
		} else {
			excludedAttr := excluded[resourceIdx+1:]
			exporter.AddExcludedAttribute(excludedAttr)
			log.Printf("Excluding attribute %s on %s resources.", excludedAttr, resourceTypePattern)
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
	if g.buildSecondDeps == nil || len(g.buildSecondDeps) == 0 {
		g.buildSecondDeps = make(map[string][]string)
	}
	if g.buildSecondDeps[refSettings.RefType] != nil {
		guidList := g.buildSecondDeps[refSettings.RefType]
		present := false
		for _, element := range guidList {
			if element == refID {
				present = true // String found in the slice
			}
		}
		if !present {
			guidList = append(guidList, refID)
			g.buildSecondDeps[refSettings.RefType] = guidList
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
	if g.addDependsOn {
		if existingResources != nil {
			for _, resource := range existingResources {
				if refID == resource.State.ID {
					return true
				}
			}
		}
		for _, resource := range g.resources {
			if refID == resource.State.ID {
				return true
			}
		}
		log.Printf("Resource present in sanitizedConfigMap and not present in resources section %v", refID)
		return false
	}
	return true
}

func (g *GenesysCloudResourceExporter) isDataSource(resType string, resLabel, originalLabel string) bool {
	return g.containsElement(g.replaceWithDatasource, resType, resLabel, originalLabel)
}

func (g *GenesysCloudResourceExporter) containsElement(elements []string, resType, resLabel, originalLabel string) bool {

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
				log.Printf("Invalid regex pattern: %s", pattern)
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

func (g *GenesysCloudResourceExporter) setResourceTypesMaps(maps map[string]resourceJSONMaps) {
	g.resourceTypesMapsMutex.Lock()
	defer g.resourceTypesMapsMutex.Unlock()
	g.resourceTypesMaps = maps
}

func (g *GenesysCloudResourceExporter) setDataSourceTypesMaps(maps map[string]resourceJSONMaps) {
	g.dataSourceTypesMapsMutex.Lock()
	defer g.dataSourceTypesMapsMutex.Unlock()
	g.dataSourceTypesMaps = maps
}

func (g *GenesysCloudResourceExporter) getResourceTypesMaps() map[string]resourceJSONMaps {
	g.resourceTypesMapsMutex.RLock()
	defer g.resourceTypesMapsMutex.RUnlock()
	return g.resourceTypesMaps
}

func (g *GenesysCloudResourceExporter) getDataSourceTypesMaps() map[string]resourceJSONMaps {
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
