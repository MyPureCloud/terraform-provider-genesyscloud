package tfexporter

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	r_registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"
)

/*
   This file contains all of the logic associated wite the process of exporting a file.
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
	includeStateFile       bool
	version                string
	provider               *schema.Provider
	exportDirPath          string
	exporters              *map[string]*resourceExporter.ResourceExporter
	resources              []resourceInfo
	resourceTypesHCLBlocks map[string]resourceHCLBlock
	resourceTypesMaps      map[string]resourceJSONMaps
	unresolvedAttrs        []unresolvableAttributeInfo
	d                      *schema.ResourceData
	ctx                    context.Context
	meta                   interface{}
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
		providerResources, providerDataSources = r_registrar.GetResources()
	}

	gre := &GenesysCloudResourceExporter{
		exportAsHCL:          d.Get("export_as_hcl").(bool),
		splitFilesByResource: d.Get("split_files_by_resource").(bool),
		logPermissionErrors:  d.Get("log_permission_errors").(bool),
		filterType:           filterType,
		includeStateFile:     d.Get("include_state_file").(bool),
		version:              meta.(*gcloud.ProviderMeta).Version,
		provider:             gcloud.New(meta.(*gcloud.ProviderMeta).Version, providerResources, providerDataSources)(),
		d:                    d,
		ctx:                  ctx,
		meta:                 meta,
	}

	err := gre.setUpExportDirPath()
	if err != nil {
		return nil, err
	}

	//Setting up the filter
	configureExporterType(ctx, d, gre, filterType)
	return gre, nil
}

func (g *GenesysCloudResourceExporter) Export() (diagErr diag.Diagnostics) {
	// Step #1 Retrieve the exporters we are have registered and have been requested by the user
	g.retrieveExporters()

	// Step #2 Retrieve all of the individual resources we are going to export
	diagErr = g.retrieveSanitizedResourceMaps()
	if diagErr != nil {
		return diagErr
	}

	// Step #3 Build a list of exporters that have an attribute we want to exclude
	if excludedAttrs, ok := g.d.GetOk("exclude_attributes"); ok {
		if diagErr := populateConfigExcluded(*g.exporters, lists.InterfaceListToStrings(excludedAttrs.([]interface{}))); diagErr != nil {
			return diagErr
		}
	}

	// Step #4 Retrieve the individual genesys cloud object instances
	diagErr = g.retrieveGenesysCloudObjectInstances()
	if diagErr != nil {
		return diagErr
	}

	// Step #5 Convert the Genesys Cloud resources to neutral format (e.g. map of maps)
	diagErr = g.buildResourceConfigMap()
	if diagErr != nil {
		return diagErr
	}

	// Step #6 Write the terraform state file along with either the HCL or JSON
	diagErr = g.generateOutputFiles()
	if diagErr != nil {
		return diagErr
	}

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

// retrieveExporters will return a list of all the registered exporters. If the resource_type on the exporter contains any elements, only the defined
// elements in the resource_type attribute will be returned.
func (g *GenesysCloudResourceExporter) retrieveExporters() {
	log.Printf("Retrieving exporters list")
	exports := resourceExporter.GetResourceExporters()

	if g.resourceTypeFilter != nil && g.filterList != nil {
		exports = g.resourceTypeFilter(exports, *g.filterList)
	}

	g.exporters = &exports

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
			//
			typeResources, err := getResourcesForType(resType, g.provider, exporter, g.meta)
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
	g.resourceTypesHCLBlocks = make(map[string]resourceHCLBlock, 0)
	g.unresolvedAttrs = make([]unresolvableAttributeInfo, 0)

	for _, resource := range g.resources {
		jsonResult, diagErr := g.instanceStateToMap(resource.State, resource.CtyType)
		if diagErr != nil {
			return diagErr
		}

		if g.resourceTypesMaps[resource.Type] == nil {
			g.resourceTypesMaps[resource.Type] = make(resourceJSONMaps)
		}

		if len(g.resourceTypesMaps[resource.Type][resource.Name]) > 0 {
			algorithm := fnv.New32()
			algorithm.Write([]byte(uuid.NewString()))
			resource.Name = resource.Name + "_" + strconv.FormatUint(uint64(algorithm.Sum32()), 10)
		}

		// Removes zero values and sets proper reference expressions
		unresolved, _ := sanitizeConfigMap(resource.Type, resource.Name, jsonResult, "", *g.exporters, g.includeStateFile, g.exportAsHCL)
		if len(unresolved) > 0 {
			g.unresolvedAttrs = append(g.unresolvedAttrs, unresolved...)
		}

		exporters := *g.exporters
		exporter := *exporters[resource.Type]
		if resourceFilesWriterFunc := exporter.CustomFileWriter.RetrieveAndWriteFilesFunc; resourceFilesWriterFunc != nil {
			exportDir, _ := getFilePath(g.d, "")
			err := resourceFilesWriterFunc(resource.State.ID, exportDir, exporter.CustomFileWriter.SubDirectory, jsonResult, g.meta)
			if err != nil {
				log.Printf("An error has occured while trying invoking the RetrieveAndWriteFilesFunc for resource type %s: %v", resource.Type, err)
			}
		}

		if g.exportAsHCL {
			if _, ok := g.resourceTypesHCLBlocks[resource.Type]; !ok {
				g.resourceTypesHCLBlocks[resource.Type] = make(resourceHCLBlock, 0)
			}
			g.resourceTypesHCLBlocks[resource.Type] = append(g.resourceTypesHCLBlocks[resource.Type], instanceStateToHCLBlock(resource.Type, resource.Name, jsonResult))
		}

		g.resourceTypesMaps[resource.Type][resource.Name] = jsonResult
	}

	return nil
}

func (g *GenesysCloudResourceExporter) instanceStateToMap(state *terraform.InstanceState, ctyType cty.Type) (gcloud.JsonMap, diag.Diagnostics) {
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
		jsonExporter := NewJsonExporter(g.resourceTypesMaps, g.unresolvedAttrs, providerSource, g.version, g.exportDirPath, g.splitFilesByResource)
		err = jsonExporter.exportJSONConfig()
	}
	if err != nil {
		return err
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
				log.Print("log_permission_errors = true. Resuming export...")
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

func getResourcesForType(resType string, provider *schema.Provider, exporter *resourceExporter.ResourceExporter, meta interface{}) ([]resourceInfo, diag.Diagnostics) {
	lenResources := len(exporter.SanitizedResourceMap)
	errorChan := make(chan diag.Diagnostics, lenResources)
	resourceChan := make(chan resourceInfo, lenResources)
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
				if err != nil {
					errString := fmt.Sprintf("Failed to get state for %s instance %s: %v", resType, id, err)
					return fmt.Errorf(errString)
				}

				if instanceState == nil {
					log.Printf("Resource %s no longer exists. Skipping.", resMeta.Name)
					removeChan <- id // Mark for removal from the map
					return nil
				}

				resourceChan <- resourceInfo{
					State:   instanceState,
					Name:    resMeta.Name,
					Type:    resType,
					CtyType: ctyType,
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

	var resources []resourceInfo
	for r := range resourceChan {
		resources = append(resources, r)
	}

	// Remove resources that weren't found in this pass
	for id := range removeChan {
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

func writeToFile(bytes []byte, path string) diag.Diagnostics {
	err := os.WriteFile(path, bytes, os.ModePerm)
	if err != nil {
		return diag.Errorf("Error writing file %s: %v", path, err)
	}
	return nil
}

// Removes empty and zero-valued attributes from the JSON config.
// Map attributes are removed by setting them to null, as the Terraform
// attribute syntax requires attributes be set to null
// that would otherwise be optional in nested block form:
// https://www.terraform.io/docs/language/attr-as-blocks.html#arbitrary-expressions-with-argument-syntax
func sanitizeConfigMap(
	resourceType string,
	resourceName string,
	configMap map[string]interface{},
	prevAttr string,
	exporters map[string]*resourceExporter.ResourceExporter, //Map of all of the exporters
	exportingState bool,
	exportingAsHCL bool) ([]unresolvableAttributeInfo, bool) {
	exporter := exporters[resourceType] //Get the specific export that we will be working with

	unresolvableAttrs := make([]unresolvableAttributeInfo, 0)

	for key, val := range configMap {
		currAttr := key
		wildcardAttr := "*"
		if prevAttr != "" {
			currAttr = prevAttr + "." + key
			wildcardAttr = prevAttr + "." + "*"
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
			if phoneNumber, ok := configMap[key].(string); !ok || phoneNumber == "" {
				continue
			}
			configMap[key] = sanitizeE164Number(configMap[key].(string))
			continue
		}

		switch val.(type) {
		case map[string]interface{}:
			// Maps are sanitized in-place
			currMap := val.(map[string]interface{})
			_, res := sanitizeConfigMap(resourceType, "", val.(map[string]interface{}), currAttr, exporters, exportingState, exportingAsHCL)
			if !res || len(currMap) == 0 {
				// Remove empty maps or maps indicating they should be removed
				configMap[key] = nil
			}
		case []interface{}:
			if arr := sanitizeConfigArray(resourceType, val.([]interface{}), currAttr, exporters, exportingState, exportingAsHCL); len(arr) > 0 {
				configMap[key] = arr
			} else {
				// Remove empty arrays
				configMap[key] = nil
			}
		case string:
			// Check if string contains nested Ref Attributes (can occur if the string is escaped json)
			if _, ok := exporter.ContainsNestedRefAttrs(currAttr); ok {
				resolvedJsonString, err := resolveRefAttributesInJsonString(currAttr, val.(string), exporter, exporters, exportingState)
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
				configMap[key] = resolveReference(refSettings, val.(string), exporters, exportingState)
			} else {
				configMap[key] = escapeString(val.(string))
			}
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
		if !exporter.AllowForZeroValues(currAttr) {
			removeZeroValues(key, configMap[key], configMap)
		}

		//If the exporter as has customer resolver for an attribute, invoke it.
		if refAttrCustomResolver, ok := exporter.CustomAttributeResolver[currAttr]; ok {
			log.Printf("Custom resolver invoked for attribute: %s", currAttr)
			err := refAttrCustomResolver.ResolverFunc(configMap, exporters)

			if err != nil {
				log.Printf("An error has occurred while trying invoke a custom resolver for attribute %s", currAttr)
			}
		}

		// Check if the exporter has custom flow resolver (Only applicable for flow resource)
		if refAttrCustomFlowResolver, ok := exporter.CustomFlowResolver[currAttr]; ok {
			log.Printf("Custom resolver invoked for attribute: %s", currAttr)
			varReference := fmt.Sprintf("%s_%s_%s", resourceType, resourceName, "filepath")
			err := refAttrCustomFlowResolver.ResolverFunc(configMap, varReference)

			if err != nil {
				log.Printf("An error has occurred while trying invoke a custom resolver for attribute %s", currAttr)
			}
		}

		if exportingAsHCL && exporter.IsJsonEncodable(currAttr) {
			if vStr, ok := configMap[key].(string); ok {
				decodedData, err := getDecodedData(vStr, currAttr)
				if err != nil {
					log.Printf("error decoding json string: %v\n", err)
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

func attrInUnResolvableAttrs(a string, myMap map[string]*schema.Schema) (*schema.Schema, bool) {
	for k, v := range myMap {
		if k == a {
			return v, true
		}
	}
	return nil, false
}

func removeZeroValues(key string, val interface{}, configMap gcloud.JsonMap) {
	if val == nil || reflect.TypeOf(val).String() == "bool" {
		return
	}
	if reflect.ValueOf(val).IsZero() {
		configMap[key] = nil
	}
}

func escapeString(strValue string) string {
	// Check for any '${' or '%{' in the exported string and escape them
	// https://www.terraform.io/docs/language/expressions/strings.html#escape-sequences
	escapedVal := strings.ReplaceAll(strValue, "${", "$${")
	escapedVal = strings.ReplaceAll(escapedVal, "%{", "%%{")
	return escapedVal
}

func sanitizeConfigArray(
	resourceType string,
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
			_, res := sanitizeConfigMap(resourceType, "", currMap, currAttr, exporters, exportingState, exportingAsHCL)
			if res && len(currMap) > 0 {
				result = append(result, val)
			}
		case []interface{}:
			if arr := sanitizeConfigArray(resourceType, val.([]interface{}), currAttr, exporters, exportingState, exportingAsHCL); len(arr) > 0 {
				result = append(result, arr)
			}
		case string:
			// Check if we are on a reference attribute and update value in array
			if refSettings := exporter.GetRefAttrSettings(currAttr); refSettings != nil {
				referenceVal := resolveReference(refSettings, val.(string), exporters, exportingState)
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

func populateConfigExcluded(exporters map[string]*resourceExporter.ResourceExporter, configExcluded []string) diag.Diagnostics {
	for _, excluded := range configExcluded {
		resourceIdx := strings.Index(excluded, ".")
		if resourceIdx == -1 {
			return diag.Errorf("Invalid excluded_attribute %s", excluded)
		}

		if len(excluded) == resourceIdx {
			return diag.Errorf("excluded_attributes value %s does not contain an attribute", excluded)
		}

		resourceName := excluded[:resourceIdx]
		exporter := exporters[resourceName]
		if exporter == nil {
			return diag.Errorf("Resource %s in excluded_attributes is not being exported.", resourceName)
		}
		excludedAttr := excluded[resourceIdx+1:]
		exporter.AddExcludedAttribute(excludedAttr)
		log.Printf("Excluding attribute %s on %s resources.", excludedAttr, resourceName)
	}
	return nil
}
