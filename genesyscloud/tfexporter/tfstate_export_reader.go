package tfexporter

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	files "terraform-provider-genesyscloud/genesyscloud/util/files"
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

	var resourceTypes []string

	// Traverse the directory
	err := filepath.Walk(exporterDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the current item is a file
		if !info.IsDir() && filepath.Ext(path) == ".tf" {
			// Process the Terraform configuration
			if err := processTerraformFile(path, resourceTypes); err != nil {
				log.Printf("Error processing %s: %v\n", path, err)
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	resourcesFromTf := readTfState(tfStateDirectory)

	differences := compareStates(resourceTypes, resourcesFromTf)

	if len(differences) == 0 {
		log.Printf("The state and Exporter are consistent.")
	} else {
		log.Printf("The state and Exporter have differences:")
		diagErr := files.WriteToFile([]byte(strings.Join(differences, "\n")), filepath.Join(t.exporterDirectoryPath, "TFStateInconsistencies.txt"))

		if diagErr != nil {
			log.Printf("Error WritingFile %s: %v\n", t.exporterDirectoryPath, diagErr)
		}
	}
	return nil
}

func compareStates(resourceTypes, resourcesFromTf []string) []string {
	differences := []string{}

	diffResourceTypes := lists.SliceDifference(resourceTypes, resourcesFromTf)

	if len(diffResourceTypes) > 0 {
		differences = append(differences, fmt.Sprintf("Elements present in TFState but not in Exporter"))
		differences = append(differences, diffResourceTypes...)
	}

	diffTFState := lists.SliceDifference(resourcesFromTf, resourceTypes)

	if len(diffTFState) > 0 {
		differences = append(differences, fmt.Sprintf("Elements present in Exporter but not in TFState"))
		differences = append(differences, diffTFState...)
	}

	return differences
}

func processTerraformFile(path string, resourceTypes []string) error {
	// Create a new HCL parser
	parser := hclparse.NewParser()

	// Parse the Terraform file
	content, diag := parser.ParseHCLFile(path)
	if diag.HasErrors() {
		return fmt.Errorf("error parsing HCL in %s: %v", path, diag)
	}

	log.Printf("bytess %v", string(content.Bytes))
	body, _ := content.Body.(*hclsyntax.Body)

	for _, block := range body.Blocks {

		if len(block.Labels) > 1 {
			resourceType := &block.Labels[0]
			resourceName := &block.Labels[1]
			resourceTypes = append(resourceTypes, *resourceType+"."+*resourceName)
		}

	}
	return nil
}

func readTfState(path string) []string {
	// Read the JSON file
	jsonFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Error reading JSON file:", err)
		return nil
	}

	// Parse JSON
	var jsonData map[string]interface{}
	err = json.Unmarshal(jsonFile, &jsonData)
	if err != nil {
		log.Printf("Error parsing JSON:", err)
		return nil
	}

	names := extractResourceTypes(jsonData)

	return names
}

func extractResourceTypes(data map[string]interface{}) []string {
	var resourceTypesFromTf []string
	resources, ok := data["resources"].([]interface{})
	if !ok {
		log.Printf("Error: resources not found in JSON")
		return resourceTypesFromTf
	}

	for _, resource := range resources {
		resourceMap, ok := resource.(map[string]interface{})
		if !ok {
			log.Printf("Error: invalid resource format in JSON")
			continue
		}

		resourceType, ok := resourceMap["type"].(string)
		if !ok {
			log.Printf("Error: name attribute not found in resource")
			continue
		}

		name, ok := resourceMap["name"].(string)
		if !ok {
			log.Printf("Error: name attribute not found in resource %v")
			continue
		}
		resourceTypesFromTf = append(resourceTypesFromTf, resourceType+"."+name)

	}
	return resourceTypesFromTf
}
