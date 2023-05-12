package tfexporter

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	zclconfCty "github.com/zclconf/go-cty/cty"
)

/*
   This file contains all of the functions used to export HCL functions.
*/

type HCLExporter struct {
	resourceTypeHCLBlocksSlice [][]byte
	unresolvedAttrs            []unresolvableAttributeInfo
	providerSource             string
	version                    string
	filePath                   string
	tfVarsFilePath             string
}

func NewHClExporter(resourceTypeHCLBlocksSlice [][]byte, unresolvedAttrs []unresolvableAttributeInfo, providerSource string, version string, filePath string, tfVarsFilePath string) *HCLExporter {
	hclExporter := &HCLExporter{
		resourceTypeHCLBlocksSlice: resourceTypeHCLBlocksSlice,
		unresolvedAttrs:            unresolvedAttrs,
		providerSource:             providerSource,
		version:                    version,
		filePath:                   filePath,
		tfVarsFilePath:             tfVarsFilePath,
	}
	return hclExporter
}

func (h *HCLExporter) exportHCLConfig() diag.Diagnostics {

	rootFile := hclwrite.NewEmptyFile()
	rootBody := rootFile.Body()
	tfBlock := rootBody.AppendNewBlock("terraform", nil)
	requiredProvidersBlock := tfBlock.Body().AppendNewBlock("required_providers", nil)
	requiredProvidersBlock.Body().SetAttributeValue("genesyscloud", zclconfCty.ObjectVal(map[string]zclconfCty.Value{
		"source":  zclconfCty.StringVal(h.providerSource),
		"version": zclconfCty.StringVal(h.version),
	}))
	terraformHCLBlock = fmt.Sprintf("%s", rootFile.Bytes())

	if len(h.resourceTypeHCLBlocksSlice) > 0 {
		// prepend terraform block
		first := h.resourceTypeHCLBlocksSlice[0]
		h.resourceTypeHCLBlocksSlice[0] = rootFile.Bytes()
		h.resourceTypeHCLBlocksSlice = append(h.resourceTypeHCLBlocksSlice, first)
	} else {
		// no resources exist - prepend terraform block alone
		h.resourceTypeHCLBlocksSlice = append(h.resourceTypeHCLBlocksSlice, rootFile.Bytes())
	}

	if len(h.unresolvedAttrs) > 0 {
		mFile := hclwrite.NewEmptyFile()
		tfVars := make(map[string]interface{})
		keys := make(map[string]string)
		for _, attr := range h.unresolvedAttrs {
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

		h.resourceTypeHCLBlocksSlice = append(h.resourceTypeHCLBlocksSlice, [][]byte{mFile.Bytes()}...)
		if err := writeTfVars(tfVars, h.tfVarsFilePath); err != nil {
			return err
		}
	}

	return writeHCLToFile(h.resourceTypeHCLBlocksSlice, h.filePath)
}

func postProcessHclBytes(resource []byte) []byte {
	resourceStr := string(resource)
	for placeholderId, val := range attributesDecoded {
		resourceStr = strings.Replace(resourceStr, fmt.Sprintf("\"%s\"", placeholderId), val, -1)
	}

	resourceStr = correctInterpolatedFileShaFunctions(resourceStr)

	return []byte(resourceStr)
}

func writeHCLToFile(bytes [][]byte, path string) diag.Diagnostics {
	// clear contents
	_ = ioutil.WriteFile(path, nil, os.ModePerm)
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
