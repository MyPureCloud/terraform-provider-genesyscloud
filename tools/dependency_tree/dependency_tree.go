package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	providerRegistrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider_registrar"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ResourceNode represents a resource in the dependency tree
type ResourceNode struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Dependencies []string `json:"dependencies,omitempty"`
}

// DependencyTree represents the complete tree of resources and their dependencies
type DependencyTree struct {
	Version   string         `json:"version"`
	Resources []ResourceNode `json:"resources"`
}

// Types for Sorting by resource type
type ByResourceType []ResourceNode

func (a ByResourceType) Len() int           { return len(a) }
func (a ByResourceType) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByResourceType) Less(i, j int) bool { return a[i].Type < a[j].Type }

func main() {
	// Get output path from command line args or use default
	outputPath := "public/data/"
	ext := ".json"
	filename := "dependency_tree"
	version := "latest"

	if len(os.Args) > 1 {
		outputPath = os.Args[1]
	}

	if len(os.Args) > 2 {
		version = os.Args[2]
	}

	// Ensure directory exists
	dir := filepath.Dir(outputPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
		os.Exit(1)
	}

	versionedFilename := fmt.Sprintf("%s-%s%s", filename, version, ext)
	outputPath = filepath.Join(dir, versionedFilename)

	exporters := providerRegistrar.GetResourceExporters()

	dependencyTree := buildDependencyTree(exporters)

	dependencyTree.Version = version

	// Convert to JSON
	jsonData, err := json.MarshalIndent(dependencyTree, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling dependency tree to JSON: %s\n", err)
		os.Exit(1)
	}

	// Write to file
	err = os.WriteFile(outputPath, jsonData, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to output file: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Dependency tree written to: %s\n", outputPath)
}

// buildDependencyTree creates a tree structure of resources and their dependencies
func buildDependencyTree(exporters map[string]*resourceExporter.ResourceExporter) DependencyTree {
	resourceNodes := []ResourceNode{}

	// Process each resource and its dependencies
	for resourceType, exporter := range exporters {
		node := ResourceNode{
			Name:         buildResourceName(resourceType),
			Type:         resourceType,
			Dependencies: []string{},
		}

		// Get dependencies from RefAttrs in the exporter
		for _, refAttr := range exporter.RefAttrs {
			if refAttr.RefType != "" && refAttr.RefType != resourceType {
				// Check if this dependency is already in the list
				isDuplicate := false
				for _, dep := range node.Dependencies {
					if dep == refAttr.RefType {
						isDuplicate = true
						break
					}
				}

				if !isDuplicate {
					node.Dependencies = append(node.Dependencies, refAttr.RefType)
				}
			}
		}

		resourceNodes = append(resourceNodes, node)
	}

	sort.Sort(ByResourceType(resourceNodes))

	// Sort the dependencies within each node
	for i := range resourceNodes {
		sort.Strings(resourceNodes[i].Dependencies)
	}

	return DependencyTree{
		Resources: resourceNodes,
	}
}

// buildResourceName formats a resource type as a readable name
func buildResourceName(resourceType string) string {
	// Strip off "genesyscloud_" from the beginning of the resource type
	name := strings.TrimPrefix(resourceType, "genesyscloud_")

	// Replace underscores with spaces
	name = strings.ReplaceAll(name, "_", " ")

	// Capitalize the first letter of each word
	name = cases.Title(language.English).String(name)

	return name
}
