package tfexporter

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	zclconfCty "github.com/zclconf/go-cty/cty"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
)

// Used to store the TF config block as a string so that it can be ignored when testing the exported HCL config file.
var (
	terraformHCLBlock string
	mockError         diag.Diagnostics

	// UID : "jsonencode({decoded representation of json string})"
	// Attrs which can be exported in jsonencode objects are populated with a UID
	// The same UID is stored as the key in attributesDecoded, with the value being the jsonencode representation of the json string.
	// When the bytes are being written to the file, the UID is found and replaced with the unquoted jsonencode object
	attributesDecoded = make(map[string]string)
)

type unresolvableAttributeInfo struct {
	ResourceType string
	ResourceName string
	Name         string
	Schema       *schema.Schema
}

const (
	defaultTfJSONFile  = "genesyscloud.tf.json"
	defaultTfHCLFile   = "genesyscloud.tf"
	defaultTfVarsFile  = "terraform.tfvars"
	defaultTfStateFile = "terraform.tfstate"
)

type GenesysCloudResourceExporter struct {
	exportAsHCL           bool
	logPermissionErrors   bool
	includeStateFile      bool
	version               string
	provider              *schema.Provider
	exportFilePath        string
	tfVarsFilePath        string
	exporters             *map[string]*gcloud.ResourceExporter
	resources             []resourceInfo
	resourceTypeHCLBlocks [][]byte
	resourceTypeJSONMaps  map[string]map[string]gcloud.JsonMap
	unresolvedAttrs       []unresolvableAttributeInfo
	d                     *schema.ResourceData
	ctx                   context.Context
	meta                  interface{}
}

func NewGenesysCloudResourceExporter(ctx context.Context, d *schema.ResourceData, meta interface{}) (GenesysCloudResourceExporter, diag.Diagnostics) {
	gre := &GenesysCloudResourceExporter{
		exportAsHCL:         d.Get("export_as_hcl").(bool),
		logPermissionErrors: d.Get("log_permission_errors").(bool),
		includeStateFile:    d.Get("include_state_file").(bool),
		version:             meta.(*gcloud.ProviderMeta).Version,
		provider:            gcloud.New(meta.(*gcloud.ProviderMeta).Version)(),
		d:                   d,
		ctx:                 ctx,
		meta:                meta,
	}
	return *gre, nil
}

func (g *GenesysCloudResourceExporter) Export() (diagErr diag.Diagnostics) {
	// Step #1 Setup Export File paths
	diagErr = g.setUpExportFilePaths()
	if diagErr != nil {
		return diagErr
	}

	// Step #2 Retrieve the exporters we are have registered and have been requested by the user
	g.retrieveExporters()

	// Step #3 Retrieve all of the individual resources we are going to export
	diagErr = g.retrieveSanitizedResourceMaps()
	if diagErr != nil {
		return diagErr
	}

	// Step #4 Build a list of exporters that have an attribute we want to exclude
	if excludedAttrs, ok := g.d.GetOk("exclude_attributes"); ok {
		if diagErr := populateConfigExcluded(*g.exporters, gcloud.InterfaceListToStrings(excludedAttrs.([]interface{}))); diagErr != nil {
			return diagErr
		}
	}

	// Step #5 Retrieve the individual genesys cloud object instances
	diagErr = g.retrieveGenesysCloudObjectInstances()
	if diagErr != nil {
		return diagErr
	}

	// Step #6 Convert the objects to map of JSON or HCL blocks
	// TODO - Need to refactor this method because we should handle JSON and HCL separately
	diagErr = g.buildJsonConfigMap()
	if diagErr != nil {
		return diagErr
	}

	// Step #7 Write the terraform state file along with either the HCL or JSON
	//  TODO this function needs to be refactored because it extremely large.
	diagErr = g.generateOutputFiles()
	if diagErr != nil {
		return diagErr
	}

	return nil
}

// SetupExportFilePaths determines whether we creating an HCL or JSON file and then returns the paths for these type of file
func (g *GenesysCloudResourceExporter) setUpExportFilePaths() (diagErr diag.Diagnostics) {
	log.Printf("Setting up file paths for export")
	var defaultFileName string

	if g.exportAsHCL {
		defaultFileName = defaultTfHCLFile
	} else {
		defaultFileName = defaultTfJSONFile
	}

	g.exportFilePath, diagErr = getFilePath(g.d, defaultFileName)
	if diagErr != nil {
		return diagErr
	}

	g.tfVarsFilePath, diagErr = getFilePath(g.d, defaultTfVarsFile)
	if diagErr != nil {
		return diagErr
	}

	return nil
}

// retrieveExporters will return a list of all the registered exporters. If the resource_type on the exporter contains any elements, only the defined
// elements in the resource_type attribute will be returned.
func (g *GenesysCloudResourceExporter) retrieveExporters() {
	log.Printf("Retrieving exporters list")
	var filter []string
	if resourceTypes, ok := g.d.GetOk("resource_types"); ok {
		filter = gcloud.InterfaceListToStrings(resourceTypes.([]interface{}))
	}

	exports := gcloud.GetResourceExporters(filter)

	g.exporters = &exports
}

// retrieveSanitizedResourceMaps will retrieve a list of all of the resources to be exported.  It will also apply a filter (e.g the :: ) and only return the specific Genesys Cloud
// resources that are specified via :: delimiter
func (g *GenesysCloudResourceExporter) retrieveSanitizedResourceMaps() (diagErr diag.Diagnostics) {
	log.Printf("Retrieving map of Genesys Cloud resources to export")
	var filter []string
	if resourceTypes, ok := g.d.GetOk("resource_types"); ok {
		filter = gcloud.InterfaceListToStrings(resourceTypes.([]interface{}))
	}

	newFilter := make([]string, 0)
	for _, f := range filter {
		if strings.Contains(f, "::") {
			newFilter = append(newFilter, f)
		}
	}

	//Retrieve a map of all of the objects we are going to build.  Apply the filter that will remove specific classes of an object
	diagErr = buildSanitizedResourceMaps(*g.exporters, newFilter, g.logPermissionErrors)
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
		go func(resType string, exporter *gcloud.ResourceExporter) {
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

// buildJsonConfigMap Builds a map of all the JSON data returned for each resource
func (g *GenesysCloudResourceExporter) buildJsonConfigMap() diag.Diagnostics {
	log.Printf("Build JSON Map and HCL Blocks")
	g.resourceTypeJSONMaps = make(map[string]map[string]gcloud.JsonMap)
	g.resourceTypeHCLBlocks = make([][]byte, 0)
	g.unresolvedAttrs = make([]unresolvableAttributeInfo, 0)

	for _, resource := range g.resources {
		jsonResult, diagErr := instanceStateToJSONMap(resource.State, resource.CtyType)
		if diagErr != nil {
			return diagErr
		}

		if g.resourceTypeJSONMaps[resource.Type] == nil {
			g.resourceTypeJSONMaps[resource.Type] = make(map[string]gcloud.JsonMap)
		}

		if len(g.resourceTypeJSONMaps[resource.Type][resource.Name]) > 0 {
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

		g.resourceTypeHCLBlocks = append(g.resourceTypeHCLBlocks, instanceStateToHCLBlock(resource.Type, resource.Name, jsonResult))
		g.resourceTypeJSONMaps[resource.Type][resource.Name] = jsonResult
	}

	return nil
}

// generateOutputFiles is used to generate the tfStateFile and either the tf export or the json based export
// TODO - This file is way overloaded  We should refactor all of this into structs with each of the state files
// being attached as individual function
func (g *GenesysCloudResourceExporter) generateOutputFiles() diag.Diagnostics {
	providerSource := sourceForVersion(g.version)
	if g.includeStateFile {
		if err := writeTfState(g.ctx, g.resources, g.d, providerSource); err != nil {
			return err
		}
	}

	var err diag.Diagnostics
	if g.exportAsHCL {
		err = exportHCLConfig(g.resourceTypeHCLBlocks, g.unresolvedAttrs, providerSource, g.version, g.exportFilePath, g.tfVarsFilePath)
	} else {
		err = exportJSONConfig(g.resourceTypeJSONMaps, g.unresolvedAttrs, providerSource, g.version, g.exportFilePath, g.tfVarsFilePath)
	}
	if err != nil {
		return err
	}

	return nil
}

/*****ADDED CODE HERE*****/
func instanceStateToHCLBlock(resType, resName string, json gcloud.JsonMap) []byte {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	block := rootBody.AppendNewBlock("resource", []string{resType, resName})
	body := block.Body()

	addBody(body, json)

	newCopy := strings.Replace(fmt.Sprintf("%s", f.Bytes()), "$${", "${", -1)
	return []byte(newCopy)
}

func addBody(body *hclwrite.Body, json gcloud.JsonMap) {
	for k, v := range json {
		addValue(body, k, v)
	}
}

func addValue(body *hclwrite.Body, k string, v interface{}) {
	if vInter, ok := v.([]interface{}); ok {
		handleInterfaceArray(body, k, vInter)
	} else {
		ctyVal := getCtyValue(v)
		if ctyVal != zclconfCty.NilVal {
			body.SetAttributeValue(k, ctyVal)
		}
	}
}

func getCtyValue(v interface{}) zclconfCty.Value {
	var value zclconfCty.Value
	if vStr, ok := v.(string); ok {
		value = zclconfCty.StringVal(vStr)
	} else if vBool, ok := v.(bool); ok {
		value = zclconfCty.BoolVal(vBool)
	} else if vInt, ok := v.(int); ok {
		value = zclconfCty.NumberIntVal(int64(vInt))
	} else if vInt32, ok := v.(int32); ok {
		value = zclconfCty.NumberIntVal(int64(vInt32))
	} else if vInt64, ok := v.(int64); ok {
		value = zclconfCty.NumberIntVal(vInt64)
	} else if vFloat32, ok := v.(float32); ok {
		value = zclconfCty.NumberFloatVal(float64(vFloat32))
	} else if vFloat64, ok := v.(float64); ok {
		value = zclconfCty.NumberFloatVal(vFloat64)
	} else if vMapInter, ok := v.(map[string]interface{}); ok {
		value = createHCLObject(vMapInter)
	} else {
		value = zclconfCty.NilVal
	}
	return value
}

// Creates hcl objects in the format name = { item1 = "", item2 = "", ... }
func createHCLObject(v map[string]interface{}) zclconfCty.Value {
	obj := make(map[string]zclconfCty.Value)
	for key, val := range v {
		ctyVal := getCtyValue(val)
		if ctyVal != zclconfCty.NilVal {
			obj[key] = ctyVal
		}
	}
	if len(obj) == 0 {
		return zclconfCty.NilVal
	}
	return zclconfCty.ObjectVal(obj)
}

func handleInterfaceArray(body *hclwrite.Body, k string, v []interface{}) {
	var listItems []zclconfCty.Value
	for _, val := range v {
		// k { ... }
		if valMap, ok := val.(map[string]interface{}); ok {
			block := body.AppendNewBlock(k, nil)
			for key, value := range valMap {
				addValue(block.Body(), key, value)
			}
			// k = [ ... ]
		} else {
			listItems = append(listItems, getCtyValue(val))
		}
	}
	if len(listItems) > 0 {
		body.SetAttributeValue(k, zclconfCty.ListVal(listItems))
	}
}

func determineVarValue(s *schema.Schema) interface{} {
	if s.Default != nil {
		if m, ok := s.Default.(map[string]string); ok {
			stringMap := make(map[string]interface{})
			for k, v := range m {
				stringMap[k] = v
			}
			return stringMap
		}

		return s.Default
	}

	switch s.Type {
	case schema.TypeString:
		return ""
	case schema.TypeInt:
		return 0
	case schema.TypeFloat:
		return 0.0
	case schema.TypeBool:
		return false
	default:
		if properties, ok := s.Elem.(*schema.Resource); ok {
			propertyMap := make(map[string]interface{})
			for k, v := range properties.Schema {
				propertyMap[k] = determineVarValue(v)
			}
			return propertyMap
		}
	}

	return nil
}

func determineVarType(s *schema.Schema) string {
	var varType string
	switch s.Type {
	case schema.TypeMap:
		if elem, ok := s.Elem.(*schema.Schema); ok {
			varType = fmt.Sprintf("map(%s)", determineVarType(elem))
		} else {
			varType = "map"
		}
	case schema.TypeBool:
		varType = "bool"
	case schema.TypeString:
		varType = "string"
	case schema.TypeList:
		fallthrough
	case schema.TypeSet:
		if elem, ok := s.Elem.(*schema.Schema); ok {
			varType = fmt.Sprintf("list(%s)", determineVarType(elem))
		} else {
			if properties, ok := s.Elem.(*schema.Resource); ok {
				propPairs := ""
				for k, v := range properties.Schema {
					propPairs = fmt.Sprintf("%s%v = %v\n", propPairs, k, determineVarType(v))
				}
				varType = fmt.Sprintf("object({%s})", propPairs)
			} else {
				varType = "object({})"
			}
		}
	case schema.TypeInt:
		fallthrough
	case schema.TypeFloat:
		varType = "number"
	}

	return varType
}

func getDecodedData(jsonString string, currAttr string) (string, error) {
	var jsonVar interface{}
	err := json.Unmarshal([]byte(jsonString), &jsonVar)
	if err != nil {
		return "", err
	}

	formattedJson, err := json.MarshalIndent(jsonVar, "", "\t")
	if err != nil {
		return "", err
	}

	formattedJsonStr := string(formattedJson)
	// fix indentation
	numOfIndents := strings.Count(currAttr, ".") + 1
	spaces := ""
	for i := 0; i < numOfIndents; i++ {
		spaces = spaces + "  "
	}
	formattedJsonStr = strings.Replace(formattedJsonStr, "\t", fmt.Sprintf("\t%v", spaces), -1)
	// add extra space before the final character (either ']' or '}')
	formattedJsonStr = fmt.Sprintf("%v%v%v", formattedJsonStr[:len(formattedJsonStr)-1], spaces, formattedJsonStr[len(formattedJsonStr)-1:])
	formattedJsonStr = fmt.Sprintf("jsonencode(%v)", formattedJsonStr)
	return formattedJsonStr, nil
}

func sourceForVersion(version string) string {
	providerSource := "registry.terraform.io/mypurecloud/genesyscloud"
	if version == "0.1.0" {
		// Force using local dev version by providing a unique repo URL
		providerSource = "genesys.com/mypurecloud/genesyscloud"
	}
	return providerSource
}

func readTfExport(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// If the output config file doesn't exist, mark the resource for creation.
	path := d.Id()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		d.SetId("")
		return nil
	}
	return nil
}

func deleteTfExport(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	configPath := d.Id()
	if _, err := os.Stat(configPath); err == nil {
		log.Printf("Deleting export config %s", configPath)
		os.Remove(configPath)
	}

	stateFile, _ := getFilePath(d, defaultTfStateFile)
	if _, err := os.Stat(stateFile); err == nil {
		log.Printf("Deleting export state %s", stateFile)
		os.Remove(stateFile)
	}

	tfVarsFile, _ := getFilePath(d, defaultTfVarsFile)
	if _, err := os.Stat(tfVarsFile); err == nil {
		log.Printf("Deleting export vars %s", tfVarsFile)
		os.Remove(tfVarsFile)
	}

	// delete left over folders e.g. prompt audio data
	dir, _ := getFilePath(d, "")
	contents, err := ioutil.ReadDir(dir)
	if err == nil {
		for _, c := range contents {
			if c.IsDir() {
				pathToLeftoverDir := path.Join(dir, c.Name())
				log.Printf("Deleting leftover directory %s", pathToLeftoverDir)
				_ = os.RemoveAll(pathToLeftoverDir)
			}
		}
	}

	return nil
}

func getFilePath(d *schema.ResourceData, filename string) (string, diag.Diagnostics) {
	directory := d.Get("directory").(string)
	if strings.HasPrefix(directory, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", diag.Errorf("Failed to evaluate home directory: %v", err)
		}
		directory = strings.Replace(directory, "~", homeDir, 1)
	}
	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		return "", diag.FromErr(err)
	}

	path := filepath.Join(directory, filename)
	if path == "" {
		return "", diag.Errorf("Failed to create file path with directory %s", directory)
	}
	return path, nil
}

func buildSanitizedResourceMaps(exporters map[string]*gcloud.ResourceExporter, filter []string, logErrors bool) diag.Diagnostics {
	errorChan := make(chan diag.Diagnostics)
	wgDone := make(chan bool)
	// Cancel remaining goroutines if an error occurs
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	for name, exporter := range exporters {
		wg.Add(1)
		go func(name string, exporter *gcloud.ResourceExporter) {
			defer wg.Done()
			log.Printf("Getting all resources for type %s", name)
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

func getResourcesForType(resType string, provider *schema.Provider, exporter *gcloud.ResourceExporter, meta interface{}) ([]resourceInfo, diag.Diagnostics) {
	lenResources := len(exporter.SanitizedResourceMap)
	errorChan := make(chan diag.Diagnostics, lenResources)
	resourceChan := make(chan resourceInfo, lenResources)
	removeChan := make(chan string, lenResources)

	res := provider.ResourcesMap[resType]
	if res == nil {
		return nil, diag.Errorf("Resource type %s not defined", resType)
	}

	ctyType := res.CoreConfigSchema().ImpliedType()

	var wg sync.WaitGroup
	wg.Add(lenResources)
	for id, resMeta := range exporter.SanitizedResourceMap {
		go func(id string, resMeta *gcloud.ResourceMeta) {
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

func getResourceState(ctx context.Context, resource *schema.Resource, resID string, resMeta *gcloud.ResourceMeta, meta interface{}) (*terraform.InstanceState, diag.Diagnostics) {
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

func instanceStateToJSONMap(state *terraform.InstanceState, ctyType cty.Type) (gcloud.JsonMap, diag.Diagnostics) {
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

func postProcessHclBytes(resource []byte) []byte {
	resourceStr := string(resource)
	for placeholderId, val := range attributesDecoded {
		resourceStr = strings.Replace(resourceStr, fmt.Sprintf("\"%s\"", placeholderId), val, -1)
	}

	resourceStr = correctInterpolatedFileShaFunctions(resourceStr)

	return []byte(resourceStr)
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
	err := ioutil.WriteFile(path, bytes, os.ModePerm)
	if err != nil {
		return diag.Errorf("Error writing file %s: %v", path, err)
	}
	return nil
}

func writeTfState(ctx context.Context, resources []resourceInfo, d *schema.ResourceData, providerSource string) diag.Diagnostics {
	stateFilePath, diagErr := getFilePath(d, defaultTfStateFile)
	if diagErr != nil {
		return diagErr
	}

	tfstate := terraform.NewState()
	for _, resource := range resources {
		resourceState := &terraform.ResourceState{
			Type:     resource.Type,
			Primary:  resource.State,
			Provider: "provider.genesyscloud",
		}
		tfstate.RootModule().Resources[resource.Type+"."+resource.Name] = resourceState
	}

	data, err := json.MarshalIndent(tfstate, "", "  ")
	if err != nil {
		return diag.Errorf("Failed to encode state as JSON: %v", err)
	}

	log.Printf("Writing export state file to %s", stateFilePath)
	if err := writeToFile(data, stateFilePath); err != nil {
		return err
	}

	// This outputs terraform state v3, and there is currently no public lib to generate v4 which is required for terraform 0.13+.
	// However, the state can be upgraded automatically by calling the terraform CLI. If this fails, just print a warning indicating
	// that the state likely needs to be upgraded manually.
	cliError := `Failed to run the terraform CLI to upgrade the generated state file. 
	The generated tfstate file will need to be upgraded manually by running the 
	following in the state file's directory:
	'terraform state replace-provider registry.terraform.io/-/genesyscloud registry.terraform.io/mypurecloud/genesyscloud'`

	tfpath, err := exec.LookPath("terraform")
	if err != nil {
		log.Println("Failed to find terraform path:", err)
		log.Println(cliError)
		return nil
	}

	// exec.CommandContext does not auto-resolve symlinks
	fileInfo, err := os.Lstat(tfpath)
	if err != nil {
		log.Println("Failed to Lstat terraform path:", err)
		log.Println(cliError)
		return nil
	}
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		tfpath, err = filepath.EvalSymlinks(tfpath)
		if err != nil {
			log.Println("Failed to resolve terraform path symlink:", err)
			log.Println(cliError)
			return nil
		}
	}

	cmd := exec.CommandContext(ctx, tfpath)
	cmd.Args = append(cmd.Args, []string{
		"state",
		"replace-provider",
		"-auto-approve",
		"-state=" + stateFilePath,
		"registry.terraform.io/-/genesyscloud",
		providerSource,
	}...)

	log.Printf("Running 'terraform state replace-provider' on %s", stateFilePath)
	if err = cmd.Run(); err != nil {
		log.Println("Failed to run command:", err)
		log.Println(cliError)
		return nil
	}
	return nil
}

func writeConfig(jsonMap map[string]interface{}, path string) diag.Diagnostics {
	dataJSONBytes, err := json.MarshalIndent(jsonMap, "", "  ")
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("Writing export config file to %s", path)
	if err := writeToFile(dataJSONBytes, path); err != nil {
		return err
	}
	return nil
}

func generateTfVarsContent(vars map[string]interface{}) string {
	tfVarsContent := ""
	for k, v := range vars {
		vStr := v
		if v == nil {
			vStr = "null"
		} else if s, ok := v.(string); ok {
			vStr = fmt.Sprintf(`"%s"`, s)
		} else if m, ok := v.(map[string]interface{}); ok {
			vStr = fmt.Sprintf(`{
	%s
}`, strings.Replace(generateTfVarsContent(m), "\n", "\n\t", -1))
		}
		newLine := ""
		if tfVarsContent != "" {
			newLine = "\n"
		}
		tfVarsContent = fmt.Sprintf("%v%s%s = %v", tfVarsContent, newLine, k, vStr)
	}

	return tfVarsContent
}

func writeTfVars(tfVars map[string]interface{}, path string) diag.Diagnostics {
	tfVarsStr := generateTfVarsContent(tfVars)
	tfVarsStr = fmt.Sprintf("// This file has been autogenerated. The following properties could not be retrieved from the API or would not make sense in a different org e.g. Edge IDs"+
		"\n// The variables contained in this file have been given default values and should be edited as necessary\n\n%s", tfVarsStr)

	log.Printf("Writing export tfvars file to %s", path)
	return writeToFile([]byte(tfVarsStr), path)
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
	exporters map[string]*gcloud.ResourceExporter, //Map of all of the exporters
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

func resolveRefAttributesInJsonString(currAttr string, currVal string, exporter *gcloud.ResourceExporter, exporters map[string]*gcloud.ResourceExporter, exportingState bool) (string, error) {
	var jsonData interface{}
	err := json.Unmarshal([]byte(currVal), &jsonData)
	if err != nil {
		return "", err
	}

	nestedAttrs, _ := exporter.ContainsNestedRefAttrs(currAttr)
	for _, value := range nestedAttrs {
		refSettings := exporter.GetNestedRefAttrSettings(value)
		if data, ok := jsonData.(map[string]interface{}); ok {
			switch data[value].(type) {
			case string:
				data[value] = resolveReference(refSettings, data[value].(string), exporters, exportingState)
			case []interface{}:
				array := data[value].([]interface{})
				for k, v := range array {
					array[k] = resolveReference(refSettings, v.(string), exporters, exportingState)
				}
				data[value] = array
			}
			jsonData = data
		}
	}
	jsonDataMarshalled, err := json.Marshal(jsonData)
	if err != nil {
		return "", err
	}
	return string(jsonDataMarshalled), nil
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
	switch val.(type) {
	case string:
		if val.(string) == "" {
			configMap[key] = nil
		}
	case int:
		if val.(int) == 0 {
			configMap[key] = nil
		}
	case float64:
		if val.(float64) == 0 {
			configMap[key] = nil
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

func sanitizeConfigArray(
	resourceType string,
	anArray []interface{},
	currAttr string,
	exporters map[string]*gcloud.ResourceExporter,
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

func resolveReference(refSettings *gcloud.RefAttrSettings, refID string, exporters map[string]*gcloud.ResourceExporter, exportingState bool) string {
	if gcloud.StringInSlice(refID, refSettings.AltValues) {
		// This is not actually a reference to another object. Keep the value
		return refID
	}

	if exporters[refSettings.RefType] != nil {
		// Get the sanitized name from the ID returned as a reference expression
		if idMetaMap := exporters[refSettings.RefType].SanitizedResourceMap; idMetaMap != nil {
			if meta := idMetaMap[refID]; meta != nil && meta.Name != "" {
				return fmt.Sprintf("${%s.%s.id}", refSettings.RefType, meta.Name)
			}
		}
	}

	if exportingState {
		// Don't remove unmatched IDs when exporting state. This will keep existing config in an org
		return refID
	}
	// No match found. Remove the value from the config since we do not have a reference to use
	return ""
}

func populateConfigExcluded(exporters map[string]*gcloud.ResourceExporter, configExcluded []string) diag.Diagnostics {
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
