package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover and extract metadata from resource schema files",
	Long:  "Scans the specified path for resource schema files and extracts metadata annotations",
	RunE:  runDiscover,
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export metadata in various formats",
	Long:  "Exports discovered metadata in Markdown, JSON, or CSV format",
	RunE:  runExport,
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate metadata completeness",
	Long:  "Validates that all resources have proper metadata annotations",
	RunE:  runValidate,
}

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Generate metadata template for a resource",
	Long:  "Generates a metadata template that can be added to a resource schema file",
	RunE:  runTemplate,
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a documentation report in all formats",
	Long:  "Generates a comprehensive report in markdown, JSON, and CSV formats for documentation publishing",
	RunE:  runReport,
}

// Command flags
var (
	discoverPath     string
	exportPath       string
	exportFormat     string
	exportOutput     string
	validatePath     string
	templateResource string
	templatePackage  string
	reportPath       string
	reportOutput     string
)

func init() {
	// Discover command flags
	discoverCmd.Flags().StringVarP(&discoverPath, "path", "p", "./genesyscloud", "Path to scan for resource schema files")

	// Export command flags
	exportCmd.Flags().StringVarP(&exportPath, "path", "p", "./genesyscloud", "Path to scan for resource schema files")
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "markdown", "Output format (markdown, json, csv)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file (defaults to stdout)")

	// Validate command flags
	validateCmd.Flags().StringVarP(&validatePath, "path", "p", "./genesyscloud", "Path to scan for resource schema files")

	// Template command flags
	templateCmd.Flags().StringVarP(&templateResource, "resource", "r", "", "Resource type name (required)")
	templateCmd.Flags().StringVarP(&templatePackage, "package", "p", "", "Package name (required)")
	templateCmd.MarkFlagRequired("resource")
	templateCmd.MarkFlagRequired("package")

	// Report command flags
	reportCmd.Flags().StringVarP(&reportPath, "path", "p", "./genesyscloud", "Path to scan for resource schema files")
	reportCmd.Flags().StringVarP(&reportOutput, "output", "o", "./resource-annotation-report", "Output location and base name for the report")
}

func runDiscover(cmd *cobra.Command, args []string) error {
	fmt.Printf("Discovering resources in: %s\n", discoverPath)

	// Use the real discovery framework
	discovery := NewResourceDiscovery(discoverPath)
	metadata, err := discovery.DiscoverResources()
	if err != nil {
		return fmt.Errorf("failed to discover resources: %w", err)
	}

	fmt.Printf("Discovered %d resources with metadata\n", len(metadata))

	// Display results in table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Resource Type\tPackage\tTeam\tChat Room\tDescription")
	fmt.Fprintln(w, "-------------\t-------\t----\t---------\t-----------")

	for _, m := range metadata {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			m.ResourceType,
			m.PackageName,
			m.TeamName,
			m.TeamChatRoom,
			m.Description)
	}
	w.Flush()

	return nil
}

func runExport(cmd *cobra.Command, args []string) error {
	discovery := NewResourceDiscovery(exportPath)
	metadata, err := discovery.DiscoverResources()
	if err != nil {
		return fmt.Errorf("failed to discover resources: %w", err)
	}

	// Convert to CLI format
	var cliMetadata []ResourceMetadata
	for _, m := range metadata {
		cliMetadata = append(cliMetadata, ResourceMetadata{
			ResourceType: m.ResourceType,
			PackageName:  m.PackageName,
			TeamName:     m.TeamName,
			TeamChatRoom: m.TeamChatRoom,
			Description:  m.Description,
		})
	}

	var output io.Writer

	if exportOutput != "" {
		file, err := os.Create(exportOutput)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		output = file
	} else {
		output = os.Stdout
	}

	switch strings.ToLower(exportFormat) {
	case "markdown":
		return exportMarkdown(cliMetadata, output)
	case "json":
		return exportJSON(cliMetadata, output)
	case "csv":
		return exportCSV(cliMetadata, output)
	default:
		return fmt.Errorf("unsupported format: %s", exportFormat)
	}
}

func runValidate(cmd *cobra.Command, args []string) error {
	fmt.Printf("Validating metadata in: %s\n", validatePath)

	// Use the real discovery framework
	discovery := NewResourceDiscovery(validatePath)

	// Get all schema files using the findAllSchemaFiles method
	allSchemaFiles, err := discovery.findAllSchemaFiles()
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	var validResources []string
	var invalidResources []string

	// Check each file
	for _, file := range allSchemaFiles {
		metadata, _ := discovery.extractMetadataFromFile(file)

		if metadata != nil && metadata.TeamName != "" {
			validResources = append(validResources, metadata.ResourceType)
		} else {
			// Extract filename from path for invalid resources
			parts := strings.Split(file, "/")
			filename := parts[len(parts)-1] // Get last part (filename)
			invalidResources = append(invalidResources, filename)
		}
	}

	// Print results
	fmt.Printf("\nValidation Results:\n")
	fmt.Printf("==================\n")
	fmt.Printf("Total schema files: %d\n", len(allSchemaFiles))
	fmt.Printf("Valid resources: %d\n", len(validResources))

	for _, resource := range validResources {
		fmt.Printf("  %s\n", resource)
	}

	if len(invalidResources) > 0 {
		fmt.Printf("\nInvalid resources: %d\n", len(invalidResources))
		for _, resource := range invalidResources {
			fmt.Printf("  %s (missing team annotation)\n", resource)
		}
		return fmt.Errorf("validation failed: %d resources missing metadata", len(invalidResources))
	}

	fmt.Println("\nAll resources have valid metadata")
	return nil
}

func runTemplate(cmd *cobra.Command, args []string) error {
	template := fmt.Sprintf(`// Resource Metadata Template for %s
// Add these comments to your schema file:
// @team: [TEAM_NAME]
// @chat: [TEAM_CHAT_ROOM]
// @description: [RESOURCE_DESCRIPTION]
// Or use build tags:
//go:build team=[TEAM_NAME] chat=[TEAM_CHAT_ROOM]
// Example:
// @team: Platform Team
// @chat: #platform-team
// @description: Manages %s resources and their configurations
// Place these comments at the top of your schema file, after the package declaration.
`, templateResource, templateResource)

	fmt.Println(template)
	return nil
}

func runReport(cmd *cobra.Command, args []string) error {
	discovery := NewResourceDiscovery(reportPath)
	metadata, err := discovery.DiscoverResources()
	if err != nil {
		return fmt.Errorf("failed to discover resources: %w", err)
	}

	var fullMetadata []ResourceMetadata
	for _, m := range metadata {
		fullMetadata = append(fullMetadata, ResourceMetadata{
			ResourceType: m.ResourceType,
			PackageName:  m.PackageName,
			TeamName:     m.TeamName,
			TeamChatRoom: m.TeamChatRoom,
			Description:  m.Description,
		})
	}

	reportBaseName := filepath.Base(reportOutput)
	reportPath := filepath.Dir(reportOutput)

	var jsonBuffer bytes.Buffer
	if err := exportJSON(fullMetadata, &jsonBuffer); err != nil {
		return fmt.Errorf("failed to export JSON: %w", err)
	}
	jsonMarkdownPath := filepath.Join(reportPath, reportBaseName+".json.md")
	jsonMarkdownFile, err := os.Create(jsonMarkdownPath)
	if err != nil {
		return fmt.Errorf("failed to create JSON markdown file: %w", err)
	}
	defer jsonMarkdownFile.Close()

	jsonHeader := fmt.Sprintf(`---
title: Resource Support Directory (JSON)
order: 3
---
`)

	if _, err := jsonMarkdownFile.WriteString(jsonHeader); err != nil {
		return fmt.Errorf("failed to write JSON markdown header: %w", err)
	}

	if _, err := jsonMarkdownFile.WriteString("```json\n"); err != nil {
		return fmt.Errorf("failed to write JSON code block start: %w", err)
	}

	if _, err := jsonMarkdownFile.Write(jsonBuffer.Bytes()); err != nil {
		return fmt.Errorf("failed to write JSON content: %w", err)
	}

	if _, err := jsonMarkdownFile.WriteString("\n```\n"); err != nil {
		return fmt.Errorf("failed to write JSON code block end: %w", err)
	}

	var csvBuffer bytes.Buffer
	if err := exportCSV(fullMetadata, &csvBuffer); err != nil {
		return fmt.Errorf("failed to export CSV: %w", err)
	}
	csvMarkdownPath := filepath.Join(reportPath, reportBaseName+".csv.md")
	csvMarkdownFile, err := os.Create(csvMarkdownPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV markdown file: %w", err)
	}
	defer csvMarkdownFile.Close()

	csvHeader := fmt.Sprintf(`---
title: Resource Support Directory (CSV)
order: 4
---
`)

	if _, err := csvMarkdownFile.WriteString(csvHeader); err != nil {
		return fmt.Errorf("failed to write CSV markdown header: %w", err)
	}

	if _, err := csvMarkdownFile.WriteString("```csv\n"); err != nil {
		return fmt.Errorf("failed to write CSV code block start: %w", err)
	}

	if _, err := csvMarkdownFile.Write(csvBuffer.Bytes()); err != nil {
		return fmt.Errorf("failed to write CSV content: %w", err)
	}

	if _, err := csvMarkdownFile.WriteString("\n```\n"); err != nil {
		return fmt.Errorf("failed to write CSV code block end: %w", err)
	}

	reportFilePath := filepath.Join(reportPath, reportBaseName+".md")
	reportFile, err := os.Create(reportFilePath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer reportFile.Close()

	reportHeader := `---
title: Resource Support Directory
order: 2
---

This report contains the information and contact details for the teams that are responsible for the resources in the CX as Code project.

---

`

	if _, err := reportFile.WriteString(reportHeader); err != nil {
		return fmt.Errorf("failed to write report header: %w", err)
	}

	if err := exportMarkdown(fullMetadata, reportFile); err != nil {
		return fmt.Errorf("failed to export markdown: %w", err)
	}

	reportFooter := fmt.Sprintf(`
---

## About This Report

This report is automatically generated from resource metadata annotations in the [Terraform provider codebase](https://github.com/MyPureCloud/terraform-provider-genesyscloud).

**Last Updated**: %s

**Total Resources**: %d
`, getCurrentDate(), len(fullMetadata))

	if _, err := reportFile.WriteString(reportFooter); err != nil {
		return fmt.Errorf("failed to write report footer: %w", err)
	}

	indexPath := filepath.Join(reportPath, "index.md")
	indexFile, err := os.Create(indexPath)
	if err != nil {
		return fmt.Errorf("failed to create index file: %w", err)
	}
	defer indexFile.Close()

	indexContent := fmt.Sprintf(`---
title: Overview
group: Resource Support Reports
order: 1
---

This directory contains resource annotation reports generated from the [Terraform provider codebase](https://github.com/MyPureCloud/terraform-provider-genesyscloud). These reports provide information about team ownership and contact details for resources in the CX as Code Terraform Provider.

## Report Contents

Report contains:
- Resource type identifiers
- Package names
- Team ownership information
- Genesys Cloud Chat room contact details
- Resource descriptions

## Last Updated

Reports are automatically generated.

**Last Updated**: %s

**Total Resources**: %d
`, getCurrentDate(), len(fullMetadata))

	if _, err := indexFile.WriteString(indexContent); err != nil {
		return fmt.Errorf("failed to write index file: %w", err)
	}

	fmt.Printf("Report generated successfully!\n")
	fmt.Printf("  Index:  %s\n", indexPath)
	fmt.Printf("  Report: %s\n", reportFilePath)
	fmt.Printf("  JSON:   %s\n", jsonMarkdownPath)
	fmt.Printf("  CSV:    %s\n", csvMarkdownPath)

	return nil
}

func getCurrentDate() string {
	return time.Now().Format("2006-01-02")
}

// Export functions
func exportMarkdown(annotations []ResourceMetadata, output io.Writer) error {
	fmt.Fprintln(output, "| Resource Type | Package | Team | Genesys Cloud Chat Room | Description |")
	fmt.Fprintln(output, "|--------------|:--------:|------|:-----------------------------:|-------------|")

	for _, a := range annotations {
		resourceType := strings.ReplaceAll(a.ResourceType, "_", "\\_")
		packageName := fmt.Sprintf("`%s`", a.PackageName)
		chatRoom := strings.TrimPrefix(a.TeamChatRoom, "#")

		teamName := a.TeamName
		if teamName == "" {
			teamName = "Unknown Team"
		}

		if chatRoom == "" {
			chatRoom = "Unknown Chat Room"
		}

		description := a.Description
		if description == "" {
			description = "N/A"
		}

		fmt.Fprintf(output, "| %s | %s | %s | %s | %s |\n",
			resourceType,
			packageName,
			teamName,
			chatRoom,
			description)
	}

	return nil
}

func exportJSON(annotations []ResourceMetadata, output io.Writer) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(annotations)
}

func exportCSV(annotations []ResourceMetadata, output io.Writer) error {
	writer := csv.NewWriter(output)
	defer writer.Flush()

	if err := writer.Write([]string{"Resource Type", "Package", "Team", "Chat Room", "Description"}); err != nil {
		return err
	}

	for _, m := range annotations {
		if err := writer.Write([]string{
			m.ResourceType,
			m.PackageName,
			m.TeamName,
			m.TeamChatRoom,
			m.Description,
		}); err != nil {
			return err
		}
	}

	return nil
}

// ResourceMetadata represents the metadata structure for the CLI tool
type ResourceMetadata struct {
	ResourceType string `json:"resource_type"`
	PackageName  string `json:"package_name"`
	TeamName     string `json:"team_name"`
	TeamChatRoom string `json:"team_chat_room"`
	Description  string `json:"description"`
}

// ResourceDiscovery provides functionality to discover and extract metadata from resources
type ResourceDiscovery struct {
	basePath string
}

// NewResourceDiscovery creates a new resource discovery instance(Constructor function)
func NewResourceDiscovery(basePath string) *ResourceDiscovery {
	return &ResourceDiscovery{
		basePath: basePath,
	}
}

// findAllSchemaFiles scans the directory and returns all schema file paths
func (d *ResourceDiscovery) findAllSchemaFiles() ([]string, error) {
	var schemaFiles []string

	err := filepath.Walk(d.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Look for schema files (specifically resource schema files in genesyscloud directory)
		if strings.HasSuffix(path, "_schema.go") && strings.Contains(path, "genesyscloud") {
			schemaFiles = append(schemaFiles, path)
		}

		return nil
	})

	return schemaFiles, err
}

// DiscoverResources scans the codebase for resource schema files and extracts metadata
func (d *ResourceDiscovery) DiscoverResources() ([]ResourceMetadata, error) {
	var allMetadata []ResourceMetadata

	// Get all schema files using the reusable method
	schemaFiles, err := d.findAllSchemaFiles()
	if err != nil {
		return nil, err
	}

	// Extract metadata from each file
	for _, path := range schemaFiles {
		metadata, err := d.extractMetadataFromFile(path)
		if err != nil {
			// Log error but continue scanning
			fmt.Printf("Warning: Failed to extract metadata from %s: %v\n", path, err)
			continue
		}

		if metadata != nil {
			allMetadata = append(allMetadata, *metadata)
		}
	}

	return allMetadata, nil
}

// extractMetadataFromFile extracts metadata from a single schema file
func (d *ResourceDiscovery) extractMetadataFromFile(filePath string) (*ResourceMetadata, error) {
	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Split into lines
	lines := strings.Split(string(content), "\n")

	// Extract package name from first few lines
	var packageName string
	var resourceType string
	var teamName string
	var teamChatRoom string
	var description string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Extract package name
		if strings.HasPrefix(line, "package ") {
			packageName = strings.TrimSpace(strings.TrimPrefix(line, "package "))
		}

		// Extract resource type (look for const ResourceType = "...")
		if strings.Contains(line, "ResourceType") && strings.Contains(line, "=") && strings.Contains(line, `"`) {
			// Handle both: const ResourceType = "value" and ResourceType = "value"
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				resourceType = strings.Trim(strings.TrimSpace(parts[1]), `"`)
				// Remove any trailing comments or semicolons
				if idx := strings.Index(resourceType, "//"); idx != -1 {
					resourceType = strings.TrimSpace(resourceType[:idx])
				}
				resourceType = strings.Trim(resourceType, `"`)
			}
		}

		// Extract comment-based annotations
		if strings.HasPrefix(line, "// @team:") {
			teamName = strings.TrimSpace(strings.TrimPrefix(line, "// @team:"))
		} else if strings.HasPrefix(line, "// @chat:") {
			teamChatRoom = strings.TrimSpace(strings.TrimPrefix(line, "// @chat:"))
		} else if strings.HasPrefix(line, "// @description:") {
			description = strings.TrimSpace(strings.TrimPrefix(line, "// @description:"))
		}

		// Extract build tag annotations
		if strings.HasPrefix(line, "//go:build ") {
			annotations := parseBuildTags(line)
			if teamName == "" && annotations["team"] != "" {
				teamName = annotations["team"]
			}
			if teamChatRoom == "" && annotations["chat"] != "" {
				teamChatRoom = annotations["chat"]
			}
		}
	}

	// If no team name is found, this might not be a resource we want to track
	if teamName == "" {
		return nil, nil
	}

	if resourceType == "" {
		return nil, fmt.Errorf("could not determine resource type from file")
	}

	return &ResourceMetadata{
		ResourceType: resourceType,
		PackageName:  packageName,
		TeamName:     teamName,
		TeamChatRoom: teamChatRoom,
		Description:  description,
	}, nil
}

// parseBuildTags parses build tag annotations
func parseBuildTags(buildTag string) map[string]string {
	annotations := make(map[string]string)

	// Remove build tag prefix
	buildTag = strings.TrimPrefix(buildTag, "//go:build ")

	// Parse key-value pairs
	pairs := strings.Split(buildTag, " ")
	for _, pair := range pairs {
		if strings.Contains(pair, "=") {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				annotations[key] = value
			}
		}
	}

	return annotations
}
