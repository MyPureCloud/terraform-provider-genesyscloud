package genesyscloud

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"hash/fnv"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	defaultTfJSONFile  = "genesyscloud.tf.json"
	defaultTfStateFile = "terraform.tfstate"
)

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
		The config file is named '%s', and the state file is named '%s'.
		`, defaultTfJSONFile, defaultTfStateFile),

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
	filePath, diagErr := getFilePath(d, defaultTfJSONFile)
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
	for _, resource := range resources {
		jsonResult, diagErr := instanceStateToJSONMap(resource.State, resource.CtyType)
		if diagErr != nil {
			return diagErr
		}

		// Removes zero values and sets proper reference expressions
		sanitizeConfigMap(resource.Type, jsonResult, "", exporters, includeStateFile)

		if resourceTypeJSONMaps[resource.Type] == nil {
			resourceTypeJSONMaps[resource.Type] = make(map[string]jsonMap)
		}

		if len(resourceTypeJSONMaps[resource.Type][resource.Name]) > 0 {
			algorithm := fnv.New32()
			algorithm.Write([]byte(uuid.NewString()))
			resource.Name = resource.Name + "_" + strconv.FormatUint(uint64(algorithm.Sum32()), 10)
		}

		resourceTypeJSONMaps[resource.Type][resource.Name] = jsonResult
	}

	providerSource := sourceForVersion(version)
	if includeStateFile {
		if err := writeTfState(ctx, resources, d, providerSource); err != nil {
			return err
		}
	}

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

	if err := writeConfig(rootJSONObject, filePath); err != nil {
		return err
	}

	d.SetId(filePath)

	return nil
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
	return nil
}

func getFilePath(d *schema.ResourceData, filename string) (string, diag.Diagnostics) {
	directory := d.Get("directory").(string)
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

	// Cancel remaining goroutines if an error occurs
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resource := provider.ResourcesMap[resType]
	if resource == nil {
		return nil, diag.Errorf("Resource type %s not defined", resType)
	}

	ctyType := resource.CoreConfigSchema().ImpliedType()

	var wg sync.WaitGroup
	wg.Add(lenResources)
	for id, resMeta := range exporter.SanitizedResourceMap {
		go func(id string, resMeta *ResourceMeta) {
			defer wg.Done()

			// This calls into the resource's ReadContext method which
			// will block until it can acquire a pooled client config object.
			instanceState, err := getResourceState(ctx, resource, id, resMeta, meta)
			if err != nil {
				errorChan <- diag.Errorf("Failed to get state for %s instance %s: %v", resType, id, err)
				cancel() // Stop other requests
				return
			}

			if instanceState == nil {
				log.Printf("Resource %s no longer exists. Skipping.", resMeta.Name)
				removeChan <- id // Mark for removal from the map
				return
			}

			resourceChan <- resourceInfo{
				State:   instanceState,
				Name:    resMeta.Name,
				Type:    resType,
				CtyType: ctyType,
			}
		}(id, resMeta)
	}

	go func() {
		wg.Wait()
		close(resourceChan)
		close(removeChan)
	}()

	var resources []resourceInfo
	for resource := range resourceChan {
		resources = append(resources, resource)
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

// Removes empty and zero-valued attributes from the JSON config.
// Map attributes are removed by setting them to null, as the Terraform
// attribute syntax requires attributes be set to null
// that would otherwise be optional in nested block form:
// https://www.terraform.io/docs/language/attr-as-blocks.html#arbitrary-expressions-with-argument-syntax
func sanitizeConfigMap(
	resourceType string,
	configMap map[string]interface{},
	prevAttr string,
	exporters map[string]*ResourceExporter,
	exportingState bool) bool {

	exporter := exporters[resourceType]
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
			if !sanitizeConfigMap(resourceType, val.(map[string]interface{}), currAttr, exporters, exportingState) || len(currMap) == 0 {
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
		return false
	}

	return true
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
			if sanitizeConfigMap(resourceType, currMap, currAttr, exporters, exportingState) && len(currMap) > 0 {
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
