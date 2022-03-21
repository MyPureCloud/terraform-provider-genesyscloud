package genesyscloud

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
)

const (
	defaultTfJSONFile  = "genesyscloud.tf.json"
	defaultTfHCLFile   = "genesyscloud.tf"
	defaultTfVarsFile  = "terraform.tfvars"
	defaultTfStateFile = "terraform.tfstate"
)

// Used to store the TF config block as a string so that it can be ignored when testing the exported HCL config file.
var terraformHCLBlock string

type unresolvableAttributeInfo struct {
	ResourceType string
	ResourceName string
	Name         string
	Schema       *schema.Schema
}

func validateSubStringInSlice(valid []string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return warnings, errors
		}

		for _, b := range valid {
			if strings.Contains(v, b) {
				return warnings, errors
			}
		}

		if !stringInSlice(v, valid) || !subStringInSlice(v, valid) {
			errors = append(errors, fmt.Errorf("string %s not in slice", v))
			return warnings, errors
		}

		if !subStringInSlice(v, valid) {
			errors = append(errors, fmt.Errorf("substring %s not in slice", v))
			return warnings, errors
		}

		return warnings, errors
	}
}

func resourceTfExport() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf(`
		Genesys Cloud Resource to export Terraform config and (optionally) tfstate files to a local directory. 
		The config file is named '%s' or '%s', and the state file is named '%s'.
		`, defaultTfJSONFile, defaultTfHCLFile, defaultTfStateFile),

		CreateContext: createTfExport,
		ReadContext:   readTfExport,
		DeleteContext: deleteTfExport,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"directory": {
				Description: "Directory where the config and state files will be exported.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "./genesyscloud",
				ForceNew:    true,
			},
			"resource_types": {
				Description: "Resource types to export, e.g. 'genesyscloud_user'. Defaults to all exportable types.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateSubStringInSlice(getAvailableExporterTypes()),
				},
				ForceNew: true,
			},
			"include_state_file": {
				Description: "Export a 'terraform.tfstate' file along with the config file. This can be used for orgs to begin managing existing resources with terraform.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
			},
			"export_as_hcl": {
				Description: "Export the config as HCL.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
			},
			"exclude_attributes": {
				Description: "Attributes to exclude from the config when exporting resources. Each value should be of the form {resource_name}.{attribute}, e.g. 'genesyscloud_user.skills'. Excluded attributes must be optional.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				ForceNew:    true,
			},
		},
	}
}

type resourceInfo struct {
	State   *terraform.InstanceState
	Name    string
	Type    string
	CtyType cty.Type
}

func createTfExport(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var defaultFileName string
	exportAsHCL := d.Get("export_as_hcl").(bool)

	if exportAsHCL {
		defaultFileName = defaultTfHCLFile
	} else {
		defaultFileName = defaultTfJSONFile
	}

	filePath, diagErr := getFilePath(d, defaultFileName)
	if diagErr != nil {
		return diagErr
	}

	tfVarsFilePath, diagErr := getFilePath(d, defaultTfVarsFile)
	if diagErr != nil {
		return diagErr
	}

	version := meta.(*providerMeta).Version

	var filter []string
	if resourceTypes, ok := d.GetOk("resource_types"); ok {
		filter = interfaceListToStrings(resourceTypes.([]interface{}))
	}
	exporters := getResourceExporters(filter)

	newFilter := make([]string, 0)
	for _, f := range filter {
		if strings.Contains(f, "::") {
			newFilter = append(newFilter, f)
		}
	}

	if len(exporters) == 0 {
		return diag.Errorf("No valid resource types to export.")
	}

	if excludedAttrs, ok := d.GetOk("exclude_attributes"); ok {
		if diagErr := populateConfigExcluded(exporters, interfaceListToStrings(excludedAttrs.([]interface{}))); diagErr != nil {
			return diagErr
		}
	}

	diagErr = buildSanitizedResourceMaps(exporters, newFilter)
	if diagErr != nil {
		return diagErr
	}

	includeStateFile := d.Get("include_state_file").(bool)
	provider := New(version)()

	// Read the instance data from each exporter
	var resources []resourceInfo
	for resType, exporter := range exporters {
		typeResources, err := getResourcesForType(resType, provider, exporter, meta)
		if err != nil {
			return err
		}
		resources = append(resources, typeResources...)
	}

	// Generate the JSON config map
	resourceTypeJSONMaps := make(map[string]map[string]jsonMap)
	resourceTypeHCLBlocks := make([][]byte, 0)
	unresolvedAttrs := make([]unresolvableAttributeInfo, 0)
	for _, resource := range resources {
		jsonResult, diagErr := instanceStateToJSONMap(resource.State, resource.CtyType)
		if diagErr != nil {
			return diagErr
		}

		if resourceTypeJSONMaps[resource.Type] == nil {
			resourceTypeJSONMaps[resource.Type] = make(map[string]jsonMap)
		}

		if len(resourceTypeJSONMaps[resource.Type][resource.Name]) > 0 {
			algorithm := fnv.New32()
			algorithm.Write([]byte(uuid.NewString()))
			resource.Name = resource.Name + "_" + strconv.FormatUint(uint64(algorithm.Sum32()), 10)
		}

		// Removes zero values and sets proper reference expressions
		unresolved, _ := sanitizeConfigMap(resource.Type, resource.Name, jsonResult, "", exporters, includeStateFile)
		if len(unresolved) > 0 {
			unresolvedAttrs = append(unresolvedAttrs, unresolved...)
		}

		resourceTypeHCLBlocks = append(resourceTypeHCLBlocks, instanceStateToHCLBlock(resource.Type, resource.Name, jsonResult))
		resourceTypeJSONMaps[resource.Type][resource.Name] = jsonResult
	}

	providerSource := sourceForVersion(version)
	if includeStateFile {
		if err := writeTfState(ctx, resources, d, providerSource); err != nil {
			return err
		}
	}

	var err diag.Diagnostics
	if exportAsHCL {
		err = exportHCLConfig(resourceTypeHCLBlocks, unresolvedAttrs, providerSource, version, filePath, tfVarsFilePath)
	} else {
		err = exportJSONConfig(resourceTypeJSONMaps, unresolvedAttrs, providerSource, version, filePath, tfVarsFilePath)
	}
	if err != nil {
		return diagErr
	}

	d.SetId(filePath)

	return nil
}

func exportHCLConfig(
	resourceTypeHCLBlocksSlice [][]byte,
	unresolvedAttrs []unresolvableAttributeInfo,
	providerSource,
	version,
	filePath,
	tfVarsFilePath string) diag.Diagnostics {
	rootFile := hclwrite.NewEmptyFile()
	rootBody := rootFile.Body()
	tfBlock := rootBody.AppendNewBlock("terraform", nil)
	requiredProvidersBlock := tfBlock.Body().AppendNewBlock("required_providers", nil)
	requiredProvidersBlock.Body().SetAttributeValue("genesyscloud", zclconfCty.ObjectVal(map[string]zclconfCty.Value{
		"source":  zclconfCty.StringVal(providerSource),
		"version": zclconfCty.StringVal(version),
	}))
	terraformHCLBlock = fmt.Sprintf("%s", rootFile.Bytes())

	if len(resourceTypeHCLBlocksSlice) > 0 {
		// prepend terraform block
		first := resourceTypeHCLBlocksSlice[0]
		resourceTypeHCLBlocksSlice[0] = rootFile.Bytes()
		resourceTypeHCLBlocksSlice = append(resourceTypeHCLBlocksSlice, first)
	} else {
		// no resources exist - prepend terraform block alone
		resourceTypeHCLBlocksSlice = append(resourceTypeHCLBlocksSlice, rootFile.Bytes())
	}

	if len(unresolvedAttrs) > 0 {
		mFile := hclwrite.NewEmptyFile()
		tfVars := make(map[string]interface{})
		keys := make(map[string]string)
		for _, attr := range unresolvedAttrs {
			mBody := mFile.Body()
			key := fmt.Sprintf("%s_%s_%s", attr.ResourceType, attr.ResourceName, attr.Name)
			if keys[key] != "" {
				continue
			}
			keys[key] = key

			variableBlock := mBody.AppendNewBlock("variable", []string{key})

			if attr.Schema.Description != "" {
				variableBlock.Body().SetAttributeValue("description", zclconfCty.StringVal(attr.Schema.Description))
			}
			if attr.Schema.Default != nil {
				variableBlock.Body().SetAttributeValue("default", getCtyValue(attr.Schema.Default))
			}
			if attr.Schema.Sensitive {
				variableBlock.Body().SetAttributeValue("sensitive", zclconfCty.BoolVal(attr.Schema.Sensitive))
			}

			tfVars[key] = determineVarValue(attr.Schema)
		}

		resourceTypeHCLBlocksSlice = append(resourceTypeHCLBlocksSlice, [][]byte{mFile.Bytes()}...)
		if err := writeTfVars(tfVars, tfVarsFilePath); err != nil {
			return err
		}
	}

	return writeHCLToFile(resourceTypeHCLBlocksSlice, filePath)
}

func exportJSONConfig(
	resourceTypeJSONMaps map[string]map[string]jsonMap,
	unresolvedAttrs []unresolvableAttributeInfo,
	providerSource,
	version,
	filePath,
	tfVarsFilePath string) diag.Diagnostics {
	rootJSONObject := jsonMap{
		"resource": resourceTypeJSONMaps,
		"terraform": jsonMap{
			"required_providers": jsonMap{
				"genesyscloud": jsonMap{
					"source":  providerSource,
					"version": version,
				},
			},
		},
	}

	if len(unresolvedAttrs) > 0 {
		tfVars := make(map[string]interface{})
		variable := make(map[string]jsonMap)
		for _, attr := range unresolvedAttrs {
			key := fmt.Sprintf("%s_%s_%s", attr.ResourceType, attr.ResourceName, attr.Name)
			variable[key] = make(jsonMap)
			tfVars[key] = make(jsonMap)
			variable[key]["description"] = attr.Schema.Description
			if variable[key]["description"] == "" {
				variable[key]["description"] = fmt.Sprintf("%s value for resource %s of type %s", attr.Name, attr.ResourceName, attr.ResourceType)
			}

			variable[key]["sensitive"] = attr.Schema.Sensitive
			if attr.Schema.Default != nil {
				variable[key]["default"] = attr.Schema.Default
			}

			tfVars[key] = determineVarValue(attr.Schema)

			variable[key]["type"] = determineVarType(attr.Schema)
		}
		rootJSONObject["variable"] = variable
		if err := writeTfVars(tfVars, tfVarsFilePath); err != nil {
			return err
		}
	}

	return writeConfig(rootJSONObject, filePath)
}

func instanceStateToHCLBlock(resType, resName string, json jsonMap) []byte {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	block := rootBody.AppendNewBlock("resource", []string{resType, resName})
	body := block.Body()

	addBody(body, json)

	newCopy := strings.Replace(fmt.Sprintf("%s", f.Bytes()), "$${", "${", -1)
	return []byte(newCopy)
}

func addBody(body *hclwrite.Body, json jsonMap) {
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

func buildSanitizedResourceMaps(exporters map[string]*ResourceExporter, filter []string) diag.Diagnostics {
	errorChan := make(chan diag.Diagnostics)
	wgDone := make(chan bool)

	// Cancel remaining goroutines if an error occurs
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	for name, exporter := range exporters {
		wg.Add(1)
		go func(name string, exporter *ResourceExporter) {
			defer wg.Done()
			log.Printf("Getting all resources for type %s", name)
			err := exporter.loadSanitizedResourceMap(ctx, name, filter)
			if err != nil {
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

func getResourcesForType(resType string, provider *schema.Provider, exporter *ResourceExporter, meta interface{}) ([]resourceInfo, diag.Diagnostics) {
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
		go func(id string, resMeta *ResourceMeta) {
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

func getResourceState(ctx context.Context, resource *schema.Resource, resID string, resMeta *ResourceMeta, meta interface{}) (*terraform.InstanceState, diag.Diagnostics) {
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
		if strings.Contains(fmt.Sprintf("%v", err), "API Error: 404") {
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

func instanceStateToJSONMap(state *terraform.InstanceState, ctyType cty.Type) (jsonMap, diag.Diagnostics) {
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

func writeHCLToFile(bytes [][]byte, path string) diag.Diagnostics {
	// clear contents
	_ = ioutil.WriteFile(path, nil, os.ModePerm)
	for _, v := range bytes {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return diag.Errorf("Error opening/creating file %s: %v", path, err)
		}
		if _, err := f.Write(v); err != nil {
			return diag.Errorf("Error writing file %s: %v", path, err)
		}

		_, _ = f.Write([]byte("\n"))

		if err := f.Close(); err != nil {
			return diag.Errorf("Error closing file %s: %v", path, err)
		}
	}
	return nil
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
		log.Println(cliError)
		return nil
	}

	// exec.CommandContext does not auto-resolve symlinks
	fileInfo, err := os.Lstat(tfpath)
	if err != nil {
		log.Println(cliError)
		return nil
	}
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		tfpath, err = os.Readlink(tfpath)
		if err != nil {
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
	if err := cmd.Run(); err != nil {
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
	tfVarsStr = fmt.Sprintf("// This file has been autogenerated. The following properties could not be retrieved from the API or would not make sense in a different org e.g. Edge IDs" +
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
	exporters map[string]*ResourceExporter,
	exportingState bool) ([]unresolvableAttributeInfo, bool) {
	exporter := exporters[resourceType]
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

		if exporter.isAttributeExcluded(currAttr) {
			// Excluded. Remove from the config.
			configMap[key] = nil
			continue
		}

		switch val.(type) {
		case map[string]interface{}:
			// Maps are sanitized in-place
			currMap := val.(map[string]interface{})
			_, res := sanitizeConfigMap(resourceType, "", val.(map[string]interface{}), currAttr, exporters, exportingState)
			if !res || len(currMap) == 0 {
				// Remove empty maps or maps indicating they should be removed
				configMap[key] = nil
			}
		case []interface{}:
			if arr := sanitizeConfigArray(resourceType, val.([]interface{}), currAttr, exporters, exportingState); len(arr) > 0 {
				configMap[key] = arr
			} else {
				// Remove empty arrays
				configMap[key] = nil
			}
		case string:
			// Check if we are on a reference attribute and update as needed
			refSettings := exporter.getRefAttrSettings(currAttr)
			if refSettings == nil {
				// Check for wildcard attribute indicating all attributes in the map
				refSettings = exporter.getRefAttrSettings(wildcardAttr)
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
		if !exporter.allowZeroValues(currAttr) {
			removeZeroValues(key, configMap[key], configMap)
		}
	}

	if exporter.removeIfMissing(prevAttr, configMap) {
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

func removeZeroValues(key string, val interface{}, configMap jsonMap) {
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
	exporters map[string]*ResourceExporter,
	exportingState bool) []interface{} {
	exporter := exporters[resourceType]
	result := []interface{}{}
	for _, val := range anArray {
		switch val.(type) {
		case map[string]interface{}:
			// Only include in the result if sanitizeConfigMap returns true and the map is not empty
			currMap := val.(map[string]interface{})
			_, res := sanitizeConfigMap(resourceType, "", currMap, currAttr, exporters, exportingState)
			if res && len(currMap) > 0 {
				result = append(result, val)
			}
		case []interface{}:
			if arr := sanitizeConfigArray(resourceType, val.([]interface{}), currAttr, exporters, exportingState); len(arr) > 0 {
				result = append(result, arr)
			}
		case string:
			// Check if we are on a reference attribute and update value in array
			if refSettings := exporter.getRefAttrSettings(currAttr); refSettings != nil {
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

func resolveReference(refSettings *RefAttrSettings, refID string, exporters map[string]*ResourceExporter, exportingState bool) string {
	if stringInSlice(refID, refSettings.AltValues) {
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

func populateConfigExcluded(exporters map[string]*ResourceExporter, configExcluded []string) diag.Diagnostics {
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
		exporter.addExcludedAttribute(excludedAttr)
		log.Printf("Excluding attribute %s on %s resources.", excludedAttr, resourceName)
	}
	return nil
}
