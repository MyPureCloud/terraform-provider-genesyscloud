package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/examples"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
)

// Method to insert the contents of each resource's apis.md file into the markdown documentation
func main() {
	fmt.Println("Updating APIs in docs...")
	const (
		resourceFolder = "docs/resources"
		exampleFolder  = "examples/resources"
		apiDocsTag     = "**No APIs**"
	)

	missingExamples := []string{}
	ignoredExamples := examples.GetIgnoredResources()

	files, err := ioutil.ReadDir("docs/resources")
	if err != nil {
		log.Fatalf("Failed to read folder %s", resourceFolder)
	}

	for _, file := range files {

		shortResourceName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		resourceName := fmt.Sprintf("genesyscloud_%s", shortResourceName)
		fullResourceFilePath := fmt.Sprintf("%s/%s", resourceFolder, file.Name())

		// Remove any docs generated for ignored examples
		if lists.ItemInSlice(resourceName, ignoredExamples) {
			os.Remove(fullResourceFilePath)
			continue
		}

		// If no examples are provided, note, and alert at end
		examplesDir := fmt.Sprintf("%s/%s", exampleFolder, resourceName)
		if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
			log.Printf("No examples found! %s", resourceName)
			missingExamples = append(missingExamples, shortResourceName)
		}

		// Open and read the apis.md file for this resource
		apiFileName := fmt.Sprintf("%s/apis.md", examplesDir)
		apisFile, err := os.Open(apiFileName)
		if err != nil {
			fmt.Printf("Missing APIs file: %s\n", apiFileName)
			continue
		}
		defer apisFile.Close()

		apiFileBytes, err := ioutil.ReadAll(apisFile)
		if err != nil {
			fmt.Printf("Couldn't read bytes from %s\n", apiFileName)
			continue
		}

		//open the doc file
		docFile, err := os.OpenFile(fmt.Sprintf("%s/%s", resourceFolder, file.Name()), os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("Couldn't open file: %s\n", file.Name())
			continue
		}
		defer docFile.Close()

		docFileBytes, err := ioutil.ReadAll(docFile)
		if err != nil {
			fmt.Printf("Couldn't read bytes from %s\n", file.Name())
			continue
		}

		// Replace the **No APIs** line with the apis.md file
		newBytes := bytes.Replace(docFileBytes, []byte(apiDocsTag), apiFileBytes, 1)
		docFile.Truncate(0)
		docFile.WriteAt(newBytes, 0)
		fmt.Printf("Updated APIs in doc file: %s\n", file.Name())
	}

	fmt.Println()
	fmt.Printf("The following resources were explicitly ignored, and so no docs were generated: %v", ignoredExamples)
	fmt.Println()
	fmt.Printf("The following resources did not have any examples, and so docs without examples or APIs were generated: %v", missingExamples)
}
