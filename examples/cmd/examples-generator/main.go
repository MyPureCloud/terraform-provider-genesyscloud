// cmd/examples-generator/main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/examples"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider_registrar"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"
)

func main() {
	var (
		outputDir = flag.String("output", "./generated-hcl", "Output directory for generated HCL files")
		resources = flag.String("resources", "", "Comma-separated list of resource types to generate (default: all)")
		exclude   = flag.String("exclude", "genesyscloud_tf_export", "Comma-separated list of resource types to exclude")
		combined  = flag.Bool("combined", false, "Generate a single combined HCL file instead of individual files")
		help      = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if err := generateHCL(*outputDir, *resources, *exclude, *combined); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("Examples Generator - Generate Terraform HCL from Genesys Cloud provider examples")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  examples-generator [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -output string")
	fmt.Println("        Output directory for generated HCL files (default \"./generated-hcl\")")
	fmt.Println("  -resources string")
	fmt.Println("        Comma-separated list of resource types to generate (default: all)")
	fmt.Println("  -exclude string")
	fmt.Println("        Comma-separated list of resource types to exclude (default \"genesyscloud_tf_export\")")
	fmt.Println("  -combined")
	fmt.Println("        Generate a single combined HCL file instead of individual files")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Generate HCL for all resources")
	fmt.Println("  examples-generator")
	fmt.Println()
	fmt.Println("  # Generate HCL for specific resources")
	fmt.Println("  examples-generator -resources genesyscloud_routing_skill,genesyscloud_routing_queue")
	fmt.Println()
	fmt.Println("  # Exclude additional resources")
	fmt.Println("  examples-generator -exclude genesyscloud_tf_export,genesyscloud_user")
	fmt.Println()
	fmt.Println("  # Generate a single combined file")
	fmt.Println("  examples-generator -combined -output ./my-config.tf")
}

func generateHCL(outputDir, resourcesFlag, excludeFlag string, combined bool) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	var resourceTypes []string
	if resourcesFlag != "" {
		resourceTypes = strings.Split(resourcesFlag, ",")
		for i, r := range resourceTypes {
			resourceTypes[i] = strings.TrimSpace(r)
		}
	} else {
		resourceTypes = provider_registrar.GetResourceTypeNames()
	}

	// Apply exclusions
	if excludeFlag != "" {
		excludeList := strings.Split(excludeFlag, ",")
		excludeMap := make(map[string]bool)
		for _, exclude := range excludeList {
			excludeMap[strings.TrimSpace(exclude)] = true
		}

		filtered := make([]string, 0, len(resourceTypes))
		for _, resourceType := range resourceTypes {
			if !excludeMap[resourceType] {
				filtered = append(filtered, resourceType)
			}
		}
		resourceTypes = filtered
	}

	sort.Strings(resourceTypes)
	fmt.Printf("Generating HCL for %d resource types...\n", len(resourceTypes))

	if combined {
		return generateCombinedHCL(outputDir, resourceTypes)
	}
	return generateIndividualHCL(outputDir, resourceTypes)
}

func generateIndividualHCL(outputDir string, resourceTypes []string) error {
	// Generate provider.tf file
	if err := generateProviderFile(outputDir); err != nil {
		return fmt.Errorf("failed to generate provider file: %w", err)
	}

	successCount := 0
	errorCount := 0

	for _, resourceType := range resourceTypes {
		if err := generateSingleResourceHCL(outputDir, resourceType); err != nil {
			fmt.Printf("❌ %s: %v\n", resourceType, err)
			errorCount++
		} else {
			fmt.Printf("✅ %s\n", resourceType)
			successCount++
		}
	}

	fmt.Printf("\nCompleted: %d successful, %d errors\n", successCount, errorCount)
	return nil
}

func generateSingleResourceHCL(outputDir, resourceType string) error {
	exampleDir := filepath.Join(testrunner.RootDir, "examples", "resources", resourceType)

	if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
		return fmt.Errorf("example directory not found: %s", exampleDir)
	}

	newExample := examples.NewExample()
	processedState := examples.NewProcessedExampleState()
	example, err := newExample.LoadExampleWithDependencies(filepath.Join(exampleDir, "resource.tf"), processedState)
	if err != nil {
		return fmt.Errorf("failed to load example: %w", err)
	}

	if err := copyWorkingDirFiles(example, outputDir); err != nil {
		return fmt.Errorf("failed to copy working dir files: %w", err)
	}
	// Update working directory paths to point to output directory
	updateWorkingDirPaths(example, outputDir)

	hclContent, err := example.GenerateOutput()
	if err != nil {
		return fmt.Errorf("failed to generate HCL: %w", err)
	}

	outputFile := filepath.Join(outputDir, fmt.Sprintf("%s.tf", resourceType))
	if err := os.WriteFile(outputFile, []byte(hclContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func generateCombinedHCL(outputDir string, resourceTypes []string) error {
	// Generate provider.tf file
	if err := generateProviderFile(outputDir); err != nil {
		return fmt.Errorf("failed to generate provider file: %w", err)
	}

	combinedExample := examples.NewExample()
	processedState := examples.NewProcessedExampleState()

	successCount := 0
	errorCount := 0

	for _, resourceType := range resourceTypes {
		exampleDir := filepath.Join(testrunner.RootDir, "examples", "resources", resourceType)

		if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
			fmt.Printf("⚠️  %s: example directory not found\n", resourceType)
			errorCount++
			continue
		}

		resourceExample, err := examples.NewExample().LoadExampleWithDependencies(filepath.Join(exampleDir, "resource.tf"), processedState)
		if err != nil {
			fmt.Printf("❌ %s: %v\n", resourceType, err)
			errorCount++
			continue
		}

		combinedExample.Resources = append(combinedExample.Resources, resourceExample.Resources...)
		if resourceExample.Locals != nil {
			if combinedExample.Locals == nil {
				combinedExample.Locals = examples.NewLocals()
			}
			combinedExample.Locals.Merge(resourceExample.Locals)
		}

		fmt.Printf("✅ %s\n", resourceType)
		successCount++
	}

	if err := copyWorkingDirFiles(combinedExample, outputDir); err != nil {
		return fmt.Errorf("failed to copy working dir files: %w", err)
	}
	// Update working directory paths to point to output directory
	updateWorkingDirPaths(combinedExample, outputDir)

	hclContent, err := combinedExample.GenerateOutput()
	if err != nil {
		return fmt.Errorf("failed to generate combined HCL: %w", err)
	}

	outputFile := filepath.Join(outputDir, "combined.tf")
	if err := os.WriteFile(outputFile, []byte(hclContent), 0644); err != nil {
		return fmt.Errorf("failed to write combined file: %w", err)
	}

	fmt.Printf("\nGenerated combined HCL: %s\n", outputFile)
	fmt.Printf("Completed: %d successful, %d errors\n", successCount, errorCount)
	return nil
}

func copyWorkingDirFiles(example *examples.Example, outputDir string) error {
	if example.Locals == nil || len(example.Locals.WorkingDir) == 0 {
		return nil
	}

	for _, srcDir := range example.Locals.WorkingDir {
		if err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || strings.HasSuffix(path, ".tf") {
				return err
			}

			relPath, _ := filepath.Rel(srcDir, path)
			destPath := filepath.Join(outputDir, relPath)

			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}

			src, err := os.Open(path)
			if err != nil {
				return err
			}
			defer src.Close()

			dst, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer dst.Close()

			_, err = io.Copy(dst, src)
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}

func updateWorkingDirPaths(example *examples.Example, outputDir string) {
	if example.Locals == nil || len(example.Locals.WorkingDir) == 0 {
		return
	}

	for key := range example.Locals.WorkingDir {
		example.Locals.WorkingDir[key] = outputDir
	}
}

func getLatestProviderVersion() (string, error) {
	defaultVersion := "0.1.0"
	resp, err := http.Get("https://api.github.com/repos/MyPureCloud/terraform-provider-genesyscloud/releases/latest")
	if err != nil {
		return fmt.Sprintf(">= %s", defaultVersion), nil // fallback
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Sprintf(">= %s", defaultVersion), nil // fallback
	}

	// Remove 'v' prefix if present
	version := strings.TrimPrefix(release.TagName, "v")
	return ">= " + version, nil
}

func generateProviderFile(outputDir string) error {
	version, _ := getLatestProviderVersion()

	providerContent := `terraform {
  required_version = ">= 1.0.0"
  required_providers {
    genesyscloud = {
      source  = "mypurecloud/genesyscloud"
      version = "` + version + `"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.7.2"
    }
    time = {
      source  = "hashicorp/time"
      version = "0.13.1"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}
`

	providerFile := filepath.Join(outputDir, "provider.tf")
	return os.WriteFile(providerFile, []byte(providerContent), 0644)
}
