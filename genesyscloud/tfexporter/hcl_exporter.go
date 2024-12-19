package tfexporter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	zclconfCty "github.com/zclconf/go-cty/cty"
)

/*
   This file contains all of the functions used to export HCL functions.
*/

const resourceHCLFileExt = "tf"

type HCLExporter struct {
	resourceTypesJSONMaps map[string]resourceJSONMaps
	dataSourceTypesMaps   map[string]resourceJSONMaps
	unresolvedAttrs       []unresolvableAttributeInfo
	providerSource        string
	version               string
	dirPath               string
	splitFilesByResource  bool
}

func NewHClExporter(resourceTypesJSONMaps map[string]resourceJSONMaps, dataSourceTypesMaps map[string]resourceJSONMaps, unresolvedAttrs []unresolvableAttributeInfo, providerSource string, version string, dirPath string, splitFilesByResource bool) *HCLExporter {
	hclExporter := &HCLExporter{
		resourceTypesJSONMaps: resourceTypesJSONMaps,
		dataSourceTypesMaps:   dataSourceTypesMaps,
		unresolvedAttrs:       unresolvedAttrs,
		providerSource:        providerSource,
		version:               version,
		dirPath:               dirPath,
		splitFilesByResource:  splitFilesByResource,
	}
	return hclExporter
}

func (h *HCLExporter) exportHCLConfig() diag.Diagnostics {
	providerBlock := createHCLProviderBlock(h.providerSource, h.version)
	variablesBlock := createHCLVariablesBlock(h.unresolvedAttrs)

	hclBlocks := make(map[string][][]byte, 0)

	// Data resources
	for resDataType, dataJSONMap := range h.dataSourceTypesMaps {

		// Output the data resources in a sorted fashion
		blockLabels := make([]string, 0)
		for resDataLabel, _ := range dataJSONMap {
			blockLabels = append(blockLabels, resDataLabel)
		}
		sort.Strings(blockLabels)
		for _, blockLabel := range blockLabels {
			resDataJson := dataJSONMap[blockLabel]
			hclBlock := instanceStateToHCLBlock(resDataType, blockLabel, resDataJson, true)
			hclBlocks[resDataType] = append(hclBlocks[resDataType], hclBlock)
		}
	}

	// Resources
	for resType, resJSONMap := range h.resourceTypesJSONMaps {

		// Output the resources in a sorted fashion
		blockLabels := make([]string, 0)
		for resLabel, _ := range resJSONMap {
			blockLabels = append(blockLabels, resLabel)
		}
		sort.Strings(blockLabels)
		for _, resLabel := range blockLabels {
			resJson := resJSONMap[resLabel]
			hclBlock := instanceStateToHCLBlock(resType, resLabel, resJson, false)
			hclBlocks[resType] = append(hclBlocks[resType], hclBlock)
		}
	}

	if h.splitFilesByResource {
		// Provider file
		providerHCLFilePath := filepath.Join(h.dirPath, defaultTfHCLProviderFile)
		if providerHCLFilePath == "" {
			return diag.Errorf("Failed to create file path %s", providerHCLFilePath)
		}
		if diagErr := writeHCLToFile([][]byte{providerBlock}, providerHCLFilePath); diagErr != nil {
			return diagErr
		}

		// Variables file
		variablesHCLFilePath := filepath.Join(h.dirPath, defaultTfHCLVariablesFile)
		if variablesHCLFilePath == "" {
			return diag.Errorf("Failed to create file path %s", variablesHCLFilePath)
		}
		if diagErr := writeHCLToFile([][]byte{variablesBlock}, variablesHCLFilePath); diagErr != nil {
			return diagErr
		}

		// Resources files
		for resType, hclContent := range hclBlocks {
			resourceHCLFilePath := filepath.Join(h.dirPath, fmt.Sprintf("%s.%s", resType, resourceHCLFileExt))
			if resourceHCLFilePath == "" {
				return diag.Errorf("Failed to create file path %s", resourceHCLFilePath)
			}
			if diagErr := writeHCLToFile(hclContent, resourceHCLFilePath); diagErr != nil {
				return diagErr
			}
		}

	} else {
		// Single file export
		allBlockSlice := make([][]byte, 0)
		allBlockSlice = append(allBlockSlice, providerBlock)

		// Sort resource types
		resourceTypes := make([]string, 0)
		for resType, _ := range hclBlocks {
			resourceTypes = append(resourceTypes, resType)
		}
		sort.Strings(resourceTypes)
		for _, resType := range resourceTypes {
			hclContent := hclBlocks[resType]
			allBlockSlice = append(allBlockSlice, hclContent...)
		}

		allBlockSlice = append(allBlockSlice, variablesBlock)

		hclFilePath := filepath.Join(h.dirPath, defaultTfHCLFile)
		if hclFilePath == "" {
			return diag.Errorf("Failed to create file path %s", hclFilePath)
		}
		if diagErr := writeHCLToFile(allBlockSlice, hclFilePath); diagErr != nil {
			return diagErr
		}
	}

	// Optional tfvars file creation for unresolved attributes
	if len(h.unresolvedAttrs) > 0 {
		tfVars := make(map[string]interface{})
		keys := make(map[string]string)
		for _, attr := range h.unresolvedAttrs {
			key := createUnresolvedAttrKey(attr)
			if keys[key] != "" {
				continue
			}
			keys[key] = key

			tfVars[key] = determineVarValue(attr.Schema)
		}

		tfVarsFilePath := filepath.Join(h.dirPath, defaultTfVarsFile)
		if tfVarsFilePath == "" {
			return diag.Errorf("Failed to create tfvars file path %s", tfVarsFilePath)
		}
		if diagErr := writeTfVars(tfVars, tfVarsFilePath); diagErr != nil {
			return diagErr
		}
	}

	return nil
}

// Create the  HCL block for terraform and the genesyscloud provider
func createHCLProviderBlock(providerSource string, version string) []byte {
	rootFile := hclwrite.NewEmptyFile()
	rootBody := rootFile.Body()
	tfBlock := rootBody.AppendNewBlock("terraform", nil)
	requiredProvidersBlock := tfBlock.Body().AppendNewBlock("required_providers", nil)
	requiredProvidersBlock.Body().SetAttributeValue("genesyscloud", zclconfCty.ObjectVal(map[string]zclconfCty.Value{
		"source":  zclconfCty.StringVal(providerSource),
		"version": zclconfCty.StringVal(version),
	}))

	// side effect assign to terraformHCLBlock. This is for testing.
	terraformHCLBlock = string(rootFile.Bytes())

	return rootFile.Bytes()
}

// Create HCL variable blocks for the unresolved attributes
func createHCLVariablesBlock(unresolvedAttrs []unresolvableAttributeInfo) []byte {
	mFile := hclwrite.NewEmptyFile()
	keys := make(map[string]string)
	sort.Slice(unresolvedAttrs, func(i, j int) bool {
		return unresolvedAttrs[i].ResourceType < unresolvedAttrs[j].ResourceType ||
			(unresolvedAttrs[i].ResourceType == unresolvedAttrs[j].ResourceType &&
				unresolvedAttrs[i].ResourceLabel < unresolvedAttrs[j].ResourceLabel)
	})
	for _, attr := range unresolvedAttrs {
		mBody := mFile.Body()
		key := createUnresolvedAttrKey(attr)
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
	}

	return mFile.Bytes()
}

func postProcessHclBytes(resource []byte) []byte {
	resourceStr := string(resource)
	for placeholderId, val := range attributesDecoded {
		resourceStr = strings.Replace(resourceStr, fmt.Sprintf("\"%s\"", placeholderId), val, -1)
	}

	resourceStr = correctCustomFunctions(resourceStr)
	return []byte(resourceStr)
}

func writeHCLToFile(bytes [][]byte, path string) diag.Diagnostics {
	// clear contents
	_ = os.WriteFile(path, nil, os.ModePerm)
	for _, v := range bytes {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return diag.Errorf("Error opening/creating file %s: %v", path, err)
		}

		v = postProcessHclBytes(v)

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

func instanceStateToHCLBlock(resType, resLabel string, json util.JsonMap, isDataSource bool) []byte {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	var block *hclwrite.Block
	if isDataSource {
		block = rootBody.AppendNewBlock("data", []string{resType, resLabel})
	} else {
		block = rootBody.AppendNewBlock("resource", []string{resType, resLabel})
	}

	body := block.Body()

	sortedJson := sortJSONMap(json)
	addBody(body, sortedJson)

	newCopy := strings.Replace(string(f.Bytes()), "$${", "${", -1)
	return []byte(newCopy)
}

func addBody(body *hclwrite.Body, json util.JsonMap) {
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
	} else if vMapInter, ok := v.([]string); ok {
		var values []zclconfCty.Value
		for _, s := range vMapInter {
			values = append(values, zclconfCty.StringVal(s))
		}
		value = zclconfCty.ListVal(values)
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

	nestedBlock := false
	for _, val := range v {
		// k { ... }
		if valMap, ok := val.(map[string]interface{}); ok {
			block := body.AppendNewBlock(k, nil)
			for key, value := range valMap {
				addValue(block.Body(), key, value)
			}
			nestedBlock = true
			// k = [ ... ]
		} else {
			listItems = append(listItems, getCtyValue(val))
			nestedBlock = false
		}
	}
	if len(listItems) > 0 {
		body.SetAttributeValue(k, zclconfCty.ListVal(listItems))
	} else if len(listItems) == 0 && !nestedBlock {
		body.SetAttributeValue(k, zclconfCty.ListValEmpty(zclconfCty.NilType))
	}
}
