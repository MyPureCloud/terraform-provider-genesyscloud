package tfexporter

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	zclconfCty "github.com/zclconf/go-cty/cty"
)

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
