package tfexporter

import (
	"encoding/json"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"
)

/*
This file contains all of the functions used to compare TFstate and Exporter functions.
*/
type TfStateExportReader struct {
	tfStateDirectoryPath  string
	exporterDirectoryPath string
}

func NewTfStateExportReader(tfStateDirectoryPath string, exporterDirectoryPath string) *TfStateExportReader {
	tfStateExportReader := &TfStateExportReader{
		tfStateDirectoryPath:  tfStateDirectoryPath,
		exporterDirectoryPath: exporterDirectoryPath,
	}
	return tfStateExportReader
}

func (t *TfStateExportReader) compareExportAndTFState() diag.Diagnostics {

	tfStateDirectory := t.tfStateDirectoryPath
	exporterDirectory := t.exporterDirectoryPath

	resourceTypes := readExporterInstances(exporterDirectory)
	resourcesFromTf := readTfState(tfStateDirectory)

	compareStatesAndWriteFile(resourceTypes, resourcesFromTf, exporterDirectory)

	return nil
}

func compareStatesAndWriteFile(resourceTypes, resourcesFromTf []string, exporterDirectoryPath string) {
	diffResourceTypes := lists.SliceDifference(resourceTypes, resourcesFromTf)
	jsonData := make(map[string]interface{})
	if len(diffResourceTypes) > 0 {
		exporterJSON := createResourceJSON("Elements present in TFState but not in Exporter", diffResourceTypes)
		jsonData["MissingExporterResources"] = exporterJSON
	}

	diffTFState := lists.SliceDifference(resourcesFromTf, resourceTypes)
	if len(diffResourceTypes) > 0 {
		tfStateJSON := createResourceJSON("Elements present in Exporter but not in TFState", diffTFState)
		jsonData["MissingTfStateResources"] = tfStateJSON
	}

	if len(jsonData) > 0 {
		jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
			log.Printf("Error Marshalling Json %s: %v\n", jsonData, err)
			return
		}
		log.Printf("The state and Exporter have differences:")
		diagErr := files.WriteToFile(jsonBytes, filepath.Join(exporterDirectoryPath, "TFStateInconsistencies.txt"))

		if diagErr != nil {
			log.Printf("Error WritingFile %s: %v\n", exporterDirectoryPath, diagErr)
		}
		return
	} else {
		log.Printf("The state and Exporter are consistent.")
	}
}

func createResourceJSON(description string, resources []string) map[string]interface{} {
	resourceMaps := make([]map[string]string, len(resources))
	for i, res := range resources {
		resourceMaps[i] = map[string]string{"name": res}
	}

	// Create JSON structure for resources
	json := map[string]interface{}{
		"description": description,
		"resources":   resourceMaps,
	}

	return json
}

func processTerraformFile(path string, resourceTypes []string) []string {
	// Create a new HCL parser
	parser := hclparse.NewParser()

	// Parse the Terraform file
	content, diag := parser.ParseHCLFile(path)
	if diag.HasErrors() {
		log.Printf("error parsing exxport tf in %s: %v", path, diag)
		return nil
	}

	body, _ := content.Body.(*hclsyntax.Body)

	for _, block := range body.Blocks {

		if len(block.Labels) > 1 {
			resourceType := &block.Labels[0]
			resourceName := &block.Labels[1]
			resourceTypes = append(resourceTypes, *resourceType+"."+*resourceName)
		}

	}
	return resourceTypes
}

func readExporterInstances(exporterDirectory string) []string {
	var resourceTypes []string
	err := filepath.Walk(exporterDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error processing %s: %v\n", path, err)
			return nil
		}

		// Check if the current item is a file
		if !info.IsDir() && filepath.Ext(path) == ".tf" {
			// Process the Terraform configuration
			if resourceTypes = processTerraformFile(path, resourceTypes); err != nil {
				log.Printf("Error processing %s: %v\n", path, err)
				return nil
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return resourceTypes
}

func readTfState(path string) []string {
	// Read the JSON file
	jsonFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Error reading TF State File: %v", err)
		return nil
	}

	// Parse JSON
	var jsonData map[string]interface{}
	err = json.Unmarshal(jsonFile, &jsonData)
	if err != nil {
		log.Printf("Error parsing TF State File: %v", err)
		return nil
	}

	names := extractResourceTypes(jsonData)

	return names
}

func extractResourceTypes(data map[string]interface{}) []string {
	var resourceTypesFromTf []string
	resources, ok := data["resources"].([]interface{})
	if !ok {
		log.Printf("Error: resources not found in TF State File")
		return resourceTypesFromTf
	}

	for _, resource := range resources {
		resourceMap, ok := resource.(map[string]interface{})
		if !ok {
			log.Printf("Error: invalid resource format in TF State File")
			continue
		}

		resourceType, ok := resourceMap["type"].(string)
		if !ok {
			log.Printf("Error: Type attribute not found in resource %v", resource)
			continue
		}

		name, ok := resourceMap["name"].(string)
		if !ok {
			log.Printf("Error: name attribute not found in resource %v", resource)
			continue
		}
		resourceTypesFromTf = append(resourceTypesFromTf, resourceType+"."+name)

	}
	return resourceTypesFromTf
}
