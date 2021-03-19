package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Method to insert the contents of each resource's apis.md file into the markdown documentation
func main() {
	fmt.Println("Updating APIs in docs...")
	const (
		resourceFolder = "docs/resources"
		exampleFolder  = "examples/resources"
		apiDocsTag     = "**No APIs**"
	)

	files, err := ioutil.ReadDir("docs/resources")
	if err != nil {
		log.Fatalf("Failed to read folder %s", resourceFolder)
	}

	for _, file := range files {
		// Open and read the apis.md file for this resource
		apiFileName := fmt.Sprintf("%s/genesyscloud_%s/apis.md", exampleFolder, strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())))
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
}
