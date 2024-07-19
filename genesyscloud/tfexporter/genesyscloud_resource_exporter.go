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
	dependentconsumers "terraform-provider-genesyscloud/genesyscloud/dependent_consumers"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	rRegistrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/util"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/stringmap"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mohae/deepcopy"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
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
	ResourceType string
	ResourceName string
	Name         string
	Schema       *schema.Schema
}

type GenesysCloudResourceExporter struct {
	configExporter         Exporter
	filterType             ExporterFilterType
	resourceTypeFilter     ExporterResourceTypeFilter
	resourceFilter         ExporterResourceFilter
	filterList             *[]string
	exportAsHCL            bool
	splitFilesByResource   bool
	logPermissionErrors    bool
	addDependsOn           bool
	replaceWithDatasource  []string
	includeStateFile       bool
	version                string
	provider               *schema.Provider
	exportDirPath          string
	exporters              *map[string]*resourceExporter.ResourceExporter
	resources              []resourceExporter.ResourceInfo
	resourceTypesHCLBlocks map[string]resourceHCLBlock
	resourceTypesMaps      map[string]resourceJSONMaps
	dataSourceTypesMaps    map[string]resourceJSONMaps
	unresolvedAttrs        []unresolvableAttributeInfo
	d                      *schema.ResourceData
	ctx                    context.Context
	meta                   interface{}
	dependsList            map[string][]string
	buildSecondDeps        map[string][]string
	exMutex                sync.RWMutex
	cyclicDependsList      []string
	ignoreCyclicDeps       bool
	flowResourcesList      []string
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
		gre.resourceFilter = FilterResourceByName            //Setting up the resource filters
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
		exportAsHCL:          d.Get("export_as_hcl").(bool),
		splitFilesByResource: d.Get("split_files_by_resource").(bool),
		logPermissionErrors:  d.Get("log_permission_errors").(bool),
		addDependsOn:         computeDependsOn(d),
		filterType:           filterType,
		includeStateFile:     d.Get("include_state_file").(bool),
		ignoreCyclicDeps:     d.Get("ignore_cyclic_deps").(bool),
		version:              meta.(*provider.ProviderMeta).Version,
		provider:             provider.New(meta.(*provider.ProviderMeta).Version, providerResources, providerDataSources)(),
		d:                    d,
		ctx:                  ctx,
		meta:                 meta,
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
	diagErr = g.retrieveExporters()
	if diagErr != nil {
		return diagErr
	}
	// Step #2 Retrieve all the individual resources we are going to export
	diagErr = g.retrieveSanitizedResourceMaps()
	if diagErr != nil {
		return diagErr
	}

	// Step #3 Retrieve the individual genesys cloud object instances
	diagErr = g.retrieveGenesysCloudObjectInstances()
	if diagErr != nil {
		return diagErr
	}

	// Step #4 export dependent resources for the flows
	diagErr = g.buildAndExportDependsOnResourcesForFlows()
	if diagErr != nil {
		return diagErr
	}

	// Step #5 Convert the Genesys Cloud resources to neutral format (e.g. map of maps)
	diagErr = g.buildResourceConfigMap()
	if diagErr != nil {
		return diagErr
	}

	// Step #6 export dependents for other resources
	diagErr = g.buildAndExportDependentResources()
	if diagErr != nil {
		return diagErr
	}

	// Step #7 Write the terraform state file along with either the HCL or JSON
	diagErr = g.generateOutputFiles()
	if diagErr != nil {
		return diagErr
	}

	// step #8 Verify the terraform state file with Exporter Resources
	g.verifyTerraformState()

	return nil
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

// Removes the ::resource_name from the resource_types list
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
	log.Printf("Retrieving Genesys Cloud objects from Genesys Cloud")
	// Retrieves data on each individual Genesys Cloud object from each registered exporter

	errorChan := make(chan diag.Diagnostics)
	wgDone := make(chan bool)
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(g.ctx)
	defer cancel()
	// We use concurrency here to spin off each exporter type and getting the data
	for resType, exporter := range *g.exporters {
		wg.Add(1)
		go func(resType string, exporter *resourceExporter.ResourceExporter) {
			defer wg.Done()

			log.Printf("Getting exported resources for [%s]", resType)
			typeResources, err := g.getResourcesForType(resType, g.provider, exporter, g.meta)

			if err != nil {
				select {
				case <-ctx.Done():
				case errorChan <- err:
				}
				cancel()
				return
			}
			g.resources = append(g.resources, typeResources...)
		}(resType, exporter)
	}

	go func() {
		wg.Wait()
		close(wgDone)
	}()

	// Wait until either WaitGroup is done or an error is received
	select {
	case <-wgDone:
	case err := <-errorChan:
		return err
	}

	return nil
}

// buildResourceConfigMap Builds a map of all the Terraform resources data returned for each resource
func (g *GenesysCloudResourceExporter) buildResourceConfigMap() diag.Diagnostics {
	log.Printf("Build Genesys Cloud Resources Map")
	g.resourceTypesMaps = make(map[string]resourceJSONMaps)
	g.dataSourceTypesMaps = make(map[string]resourceJSONMaps)
	g.resourceTypesHCLBlocks = make(map[string]resourceHCLBlock, 0)
	g.unresolvedAttrs = make([]unresolvableAttributeInfo, 0)

	for _, resource := range g.resources {
		jsonResult, diagErr := g.instanceStateToMap(resource.State, resource.CtyType)
		isDataSource := g.isDataSource(resource.Type, resource.Name)
		if diagErr != nil {
			return diagErr
		}

		if g.resourceTypesMaps[resource.Type] == nil {
			g.resourceTypesMaps[resource.Type] = make(resourceJSONMaps)
		}

		if len(g.resourceTypesMaps[resource.Type][resource.Name]) > 0 || len(g.dataSourceTypesMaps[resource.Type][resource.Name]) > 0 {
			algorithm := fnv.New32()
			algorithm.Write([]byte(uuid.NewString()))
			resource.Name = resource.Name + "_" + strconv.FormatUint(uint64(algorithm.Sum32()), 10)
			g.updateSanitiseMap(*g.exporters, resource)
		}

		if !isDataSource {
			// Removes zero values and sets proper reference expressions
			unresolved, _ := g.sanitizeConfigMap(resource.Type, resource.Name, jsonResult, "", *g.exporters, g.includeStateFile, g.exportAsHCL, true)
			if len(unresolved) > 0 {
				g.unresolvedAttrs = append(g.unresolvedAttrs, unresolved...)
			}
		} else {
			g.sanitizeDataConfigMap(jsonResult)
		}

		// TODO put this in separate call
		exporters := *g.exporters
		if resourceFilesWriterFunc := exporters[resource.Type].CustomFileWriter.RetrieveAndWriteFilesFunc; resourceFilesWriterFunc != nil {
			exportDir, _ := getFilePath(g.d, "")
			if err := resourceFilesWriterFunc(resource.State.ID, exportDir, exporters[resource.Type].CustomFileWriter.SubDirectory, jsonResult, g.meta); err != nil {
				log.Printf("An error has occurred while trying invoking the RetrieveAndWriteFilesFunc for resource type %s: %v", resource.Type, err)
			}
		}

		if g.exportAsHCL {
			if _, ok := g.resourceTypesHCLBlocks[resource.Type]; !ok {
				g.resourceTypesHCLBlocks[resource.Type] = make(resourceHCLBlock, 0)
			}
			g.resourceTypesHCLBlocks[resource.Type] = append(g.resourceTypesHCLBlocks[resource.Type], instanceStateToHCLBlock(resource.Type, resource.Name, jsonResult, isDataSource))
		}

		if isDataSource {
			if g.dataSourceTypesMaps[resource.Type] == nil {
				g.dataSourceTypesMaps[resource.Type] = make(resourceJSONMaps)
			}
			g.dataSourceTypesMaps[resource.Type][resource.Name] = jsonResult
		} else {
			g.resourceTypesMaps[resource.Type][resource.Name] = jsonResult
		}

	}

	return nil
}

func (g *GenesysCloudResourceExporter) updateSanitiseMap(exporters map[string]*resourceExporter.ResourceExporter, //Map of all of the exporters
	resource resourceExporter.ResourceInfo) {
	if exporters[resource.Type] != nil {
		// Get the sanitized name from the ID returned as a reference expression
		if idMetaMap := exporters[resource.Type].SanitizedResourceMap; idMetaMap != nil {
			if meta := idMetaMap[resource.State.ID]; meta != nil && meta.Name != "" {
				meta.Name = resource.Name
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
func (g *GenesysCloudResourceExporter) generateOutputFiles() diag.Diagnostics {
	providerSource := g.sourceForVersion(g.version)
	if g.includeStateFile {
		t := NewTFStateWriter(g.ctx, g.resources, g.d, providerSource)
		if err := t.writeTfState(); err != nil {
			return err
		}
	}

	var err diag.Diagnostics
	if g.exportAsHCL {
		hclExporter := NewHClExporter(g.resourceTypesHCLBlocks, g.unresolvedAttrs, providerSource, g.version, g.exportDirPath, g.splitFilesByResource)
		err = hclExporter.exportHCLConfig()
	} else {
		jsonExporter := NewJsonExporter(g.resourceTypesMaps, g.dataSourceTypesMaps, g.unresolvedAttrs, providerSource, g.version, g.exportDirPath, g.splitFilesByResource)
		err = jsonExporter.exportJSONConfig()
	}

	if err != nil {
		return err
	}

	if g.cyclicDependsList != nil && len(g.cyclicDependsList) > 0 {
		err = files.WriteToFile([]byte(strings.Join(g.cyclicDependsList, "\n")), filepath.Join(g.exportDirPath, "cyclicDepends.txt"))

		if err != nil {
			return err
		}
	}

	err = g.generateZipForExporter()
	if err != nil {
		return err
	}

	return nil
}

func (g *GenesysCloudResourceExporter) generateZipForExporter() diag.Diagnostics {
	zipFileName := "../archive_genesyscloud_tf_export" + uuid.NewString() + ".zip"
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

				resource := strings.Split(meta.Name, "::::")
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
	g.buildResourceConfigMap()
	g.exportAndResolveDependencyAttributes()
	g.appendResources(uniqueResources)
	g.appendResources(existingResources)
	g.exporters = mergeExporters(existingExporters, *mergeExporters(depExporters, *g.exporters))

	return nil
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
		if diagErr != nil {
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
		if err != nil {
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

func (g *GenesysCloudResourceExporter) sourceForVersion(version string) string {
	providerSource := "registry.terraform.io/mypurecloud/genesyscloud"
	if g.version == "0.1.0" {
		// Force using local dev version by providing a unique repo URL
		providerSource = "genesys.com/mypurecloud/genesyscloud"
	}
	return providerSource
}

func (g *GenesysCloudResourceExporter) appendResources(resourcesToAdd []resourceExporter.ResourceInfo) {

	existingResources := g.copyResource()

	for _, resourceToAdd := range resourcesToAdd {
		// Check if the resource with the same ID already exists
		duplicate := false
		for _, existingResource := range g.resources {
			if existingResource.State.ID == resourceToAdd.State.ID && existingResource.Type == resourceToAdd.Type {
				duplicate = true
				break
			}
		}

		// No duplicate found, append the resource
		if !duplicate {
			existingResources = append(existingResources, resourceToAdd)
		}
	}

	g.copyResourceAddtoG(existingResources)
}

func (g *GenesysCloudResourceExporter) buildSanitizedResourceMaps(exporters map[string]*resourceExporter.ResourceExporter, filter []string, logErrors bool) diag.Diagnostics {
	errorChan := make(chan diag.Diagnostics)
	wgDone := make(chan bool)
	// Cancel remaining goroutines if an error occurs
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	for name, exporter := range exporters {
		wg.Add(1)
		go func(name string, exporter *resourceExporter.ResourceExporter) {
			defer wg.Done()
			log.Printf("Getting all resources for type %s", name)
			exporter.FilterResource = g.resourceFilter

			err := exporter.LoadSanitizedResourceMap(ctx, name, filter)

			// Used in tests
			if mockError != nil {
				err = mockError
			}
			if containsPermissionsErrorOnly(err) && logErrors {
				log.Printf("%v", err[0].Summary)
				log.Printf("Logging permission error for %s. Resuming export...", name)
				return
			}
			if err != nil {
				if !logErrors {
					err = addLogAttrInfoToErrorSummary(err)
				}
				select {
				case <-ctx.Done():
				case errorChan <- err:
				}
				cancel()
				return
			}
			log.Printf("Found %d resources for type %s", len(exporter.SanitizedResourceMap), name)
		}(name, exporter)
	}

	go func() {
		wg.Wait()
		log.Print(`Finished building sanitized resource maps`)
		close(wgDone)
	}()

	// Wait until either WaitGroup is done or an error is received
	select {
	case <-wgDone:
		return nil
	case err := <-errorChan:
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

func (g *GenesysCloudResourceExporter) getResourcesForType(resType string, provider *schema.Provider, exporter *resourceExporter.ResourceExporter, meta interface{}) ([]resourceExporter.ResourceInfo, diag.Diagnostics) {
	lenResources := len(exporter.SanitizedResourceMap)
	errorChan := make(chan diag.Diagnostics, lenResources)
	resourceChan := make(chan resourceExporter.ResourceInfo, lenResources)
	removeChan := make(chan string, lenResources)

	res := provider.ResourcesMap[resType]

	if res == nil {
		return nil, diag.Errorf("Resource type %v not defined", resType)
	}

	ctyType := res.CoreConfigSchema().ImpliedType()
	var wg sync.WaitGroup
	wg.Add(lenResources)
	for id, resMeta := range exporter.SanitizedResourceMap {
		go func(id string, resMeta *resourceExporter.ResourceMeta) {
			defer wg.Done()
			fetchResourceState := func() error {
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Minute)
				defer cancel()
				// This calls into the resource's ReadContext method which
				// will block until it can acquire a pooled client config object.
				instanceState, err := getResourceState(ctx, res, id, resMeta, meta)

				resourceType := ""
				if g.isDataSource(resType, resMeta.Name) {
					g.exMutex.Lock()
					res = provider.DataSourcesMap[resType]
					g.exMutex.Unlock()

					if res == nil {
						return fmt.Errorf("DataSource type %v not defined", resType)
					}

					schemaMap := res.SchemaMap()

					attributes := make(map[string]string)

					for attr, _ := range schemaMap {
						if value, ok := instanceState.Attributes[attr]; ok {
							attributes[attr] = value
						}
					}
					instanceState.Attributes = attributes
					resourceType = "data."
				}

				if err != nil {
					errString := fmt.Sprintf("Failed to get state for %s instance %s: %v", resType, id, err)
					return fmt.Errorf(errString)
				}

				if instanceState == nil {
					log.Printf("Resource %s no longer exists. Skipping.", resMeta.Name)
					removeChan <- id // Mark for removal from the map
					return nil
				}

				resourceChan <- resourceExporter.ResourceInfo{
					State:        instanceState,
					Name:         resMeta.Name,
					Type:         resType,
					CtyType:      ctyType,
					ResourceType: resourceType,
				}

				return nil
			}

			isTimeoutError := func(err error) bool {
				return strings.Contains(fmt.Sprintf("%v", err), "timeout while waiting for state to become") ||
					strings.Contains(fmt.Sprintf("%v", err), "context deadline exceeded")
			}

			var err error
			for ok := true; ok; ok = isTimeoutError(err) {
				err = fetchResourceState()
				if err == nil {
					return
				}
				if !isTimeoutError(err) {
					errorChan <- diag.Errorf("Failed to get state for %s instance %s: %v", resType, id, err)
				}
			}
		}(id, resMeta)
	}

	go func() {
		wg.Wait()
		close(resourceChan)
		close(removeChan)
	}()

	var resources []resourceExporter.ResourceInfo
	for r := range resourceChan {
		resources = append(resources, r)
	}

	// Remove resources that weren't found in this pass
	for id := range removeChan {
		log.Printf("Deleted resource %v", id)
		delete(exporter.SanitizedResourceMap, id)
	}

	// Return the first error if one was received
	select {
	case err := <-errorChan:
		return nil, err
	default:
		return resources, nil
	}
}

func getResourceState(ctx context.Context, resource *schema.Resource, resID string, resMeta *resourceExporter.ResourceMeta, meta interface{}) (*terraform.InstanceState, diag.Diagnostics) {
	// If defined, pass the full ID through the import method to generate a readable state
	instanceState := &terraform.InstanceState{ID: resMeta.IdPrefix + resID}
	if resource.Importer != nil && resource.Importer.StateContext != nil {
		resourceDataArr, err := resource.Importer.StateContext(ctx, resource.Data(instanceState), meta)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		if len(resourceDataArr) > 0 {
			instanceState = resourceDataArr[0].State()
		}
	}

	state, err := resource.RefreshWithoutUpgrade(ctx, instanceState, meta)
	if err != nil {
		if strings.Contains(fmt.Sprintf("%v", err), "API Error: 404") ||
			strings.Contains(fmt.Sprintf("%v", err), "API Error: 410") {
			return nil, nil
		}
		return nil, err
	}
	if state == nil || state.ID == "" {
		// Resource no longer exists
		return nil, nil
	}
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
	resourceType string,
	resourceName string,
	configMap map[string]interface{},
	prevAttr string,
	exporters map[string]*resourceExporter.ResourceExporter, //Map of all exporters
	exportingState bool,
	exportingAsHCL bool,
	parentKey bool) ([]unresolvableAttributeInfo, bool) {
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
			_, res := g.sanitizeConfigMap(resourceType, resourceName, val.(map[string]interface{}), currAttr, exporters, exportingState, exportingAsHCL, false)
			if !res || len(currMap) == 0 {
				// Remove empty maps or maps indicating they should be removed
				configMap[key] = nil
			}
		case []interface{}:
			if arr := g.sanitizeConfigArray(resourceType, resourceName, val.([]interface{}), currAttr, exporters, exportingState, exportingAsHCL); len(arr) > 0 {
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
			varReference := fmt.Sprintf("%s_%s_%s", resourceType, resourceName, key)
			unresolvableAttrs = append(unresolvableAttrs, unresolvableAttributeInfo{
				ResourceType: resourceType,
				ResourceName: resourceName,
				Name:         key,
				Schema:       attr,
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
				if err := resolverFunc(configMap, exporters, resourceName); err != nil {
					log.Printf("An error has occurred while trying invoke a custom resolver for attribute %s: %v", currAttr, err)
				}
			}
		}

		// Check if the exporter has custom flow resolver (Only applicable for flow resource)
		if refAttrCustomFlowResolver, ok := exporter.CustomFlowResolver[currAttr]; ok {
			log.Printf("Custom resolver invoked for attribute: %s", currAttr)
			varReference := fmt.Sprintf("%s_%s_%s", resourceType, resourceName, "filepath")
			if err := refAttrCustomFlowResolver.ResolverFunc(configMap, varReference); err != nil {
				log.Printf("An error has occurred while trying invoke a custom resolver for attribute %s: %v", currAttr, err)
			}
		}

		if exportingAsHCL && exporter.IsJsonEncodable(currAttr) {
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
	dataSourceType, dataSourceId, dataSourceConfig, resolve := resolveToDataSourceFunc(configMap, originalValue, sdkConfig)
	if !resolve {
		return
	}

	if g.dataSourceTypesMaps[dataSourceType] == nil {
		g.dataSourceTypesMaps[dataSourceType] = make(resourceJSONMaps)
	}

	// add the data source to the export if it hasn't already been added
	if _, ok := g.dataSourceTypesMaps[dataSourceType][dataSourceId]; ok {
		return
	}
	g.dataSourceTypesMaps[dataSourceType][dataSourceId] = dataSourceConfig
	if g.exportAsHCL {
		if _, ok := g.resourceTypesHCLBlocks[dataSourceType]; !ok {
			g.resourceTypesHCLBlocks[dataSourceType] = make(resourceHCLBlock, 0)
		}
		g.resourceTypesHCLBlocks[dataSourceType] = append(g.resourceTypesHCLBlocks[dataSourceType], instanceStateToHCLBlock(dataSourceType, dataSourceId, dataSourceConfig, true))
	}
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

	resourceDependsList := make([]string, 0)
	if exists {
		for _, res := range list {
			for _, resource := range g.resources {
				if resource.State.ID == strings.Split(res, ".")[1] {
					resourceDependsList = append(resourceDependsList, fmt.Sprintf("$dep$%s$dep$", strings.Split(res, ".")[0]+"."+resource.Name))
				}
			}
		}
		if len(resourceDependsList) > 0 {
			configMap["depends_on"] = resourceDependsList
		}

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
	resourceType string,
	resourceName string,
	anArray []interface{},
	currAttr string,
	exporters map[string]*resourceExporter.ResourceExporter,
	exportingState bool,
	exportingAsHCL bool) []interface{} {
	exporter := exporters[resourceType]
	result := []interface{}{}
	for _, val := range anArray {
		switch val.(type) {
		case map[string]interface{}:
			// Only include in the result if sanitizeConfigMap returns true and the map is not empty
			currMap := val.(map[string]interface{})
			_, res := g.sanitizeConfigMap(resourceType, resourceName, currMap, currAttr, exporters, exportingState, exportingAsHCL, false)
			if res && len(currMap) > 0 {
				result = append(result, val)
			}
		case []interface{}:
			if arr := g.sanitizeConfigArray(resourceType, resourceName, val.([]interface{}), currAttr, exporters, exportingState, exportingAsHCL); len(arr) > 0 {
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

		resourceName := excluded[:resourceIdx]
		// identify all the resource names which match the regex
		exporter := exporters[resourceName]
		if exporter == nil {
			for name, exporter1 := range exporters {
				match, _ := regexp.MatchString(resourceName, name)

				if match {
					excludedAttr := excluded[resourceIdx+1:]
					exporter1.AddExcludedAttribute(excludedAttr)
					log.Printf("Excluding attribute %s on %s resources.", excludedAttr, resourceName)
					matchFound = true
					continue
				}
			}

			if !matchFound {
				if g.addDependsOn {
					excludedAttr := excluded[resourceIdx+1:]
					log.Printf("Ignoring exclude attribute %s on %s resources. Since exporter is not retrieved", excludedAttr, resourceName)
					continue
				} else {
					return diag.Errorf("Resource %s in excluded_attributes is not being exported.", resourceName)
				}
			}
		} else {
			excludedAttr := excluded[resourceIdx+1:]
			exporter.AddExcludedAttribute(excludedAttr)
			log.Printf("Excluding attribute %s on %s resources.", excludedAttr, resourceName)
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
		// Get the sanitized name from the ID returned as a reference expression
		if idMetaMap := exporters[refSettings.RefType].SanitizedResourceMap; idMetaMap != nil {
			if meta := idMetaMap[refID]; meta != nil && meta.Name != "" {

				if g.isDataSource(refSettings.RefType, meta.Name) && g.resourceIdExists(refID, nil) {
					return fmt.Sprintf("${%s.%s.%s.id}", "data", refSettings.RefType, meta.Name)
				}
				if g.resourceIdExists(refID, nil) {
					return fmt.Sprintf("${%s.%s.id}", refSettings.RefType, meta.Name)
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

func (g *GenesysCloudResourceExporter) isDataSource(resType string, name string) bool {
	return g.containsElement(resourceExporter.ExportAsData, resType, name) || g.containsElement(g.replaceWithDatasource, resType, name)
}

func (g *GenesysCloudResourceExporter) containsElement(elements []string, resType, name string) bool {

	for _, element := range elements {
		if element == resType+"::"+name || fetchByRegex(element, resType, name) {
			return true
		}
	}
	return false
}

func fetchByRegex(fullName string, resType string, name string) bool {
	if strings.Contains(fullName, "::") && strings.Split(fullName, "::")[0] == resType {
		i := strings.Index(fullName, "::")
		regexStr := fullName[i+2:]
		match, _ := regexp.MatchString(regexStr, name)
		return match
	}
	return false
}

func (g *GenesysCloudResourceExporter) verifyTerraformState() diag.Diagnostics {

	if exists := featureToggles.StateComparisonTrue(); exists {
		if g.exportAsHCL {
			tfstatePath, _ := getFilePath(g.d, defaultTfStateFile)
			hclExporter := NewTfStateExportReader(tfstatePath, g.exportDirPath)
			hclExporter.compareExportAndTFState()
		}
	}

	return nil
}
