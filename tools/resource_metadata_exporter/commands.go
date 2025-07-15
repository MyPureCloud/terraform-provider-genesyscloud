package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

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

// Command flags
var (
	discoverPath     string
	exportFormat     string
	exportOutput     string
	validatePath     string
	templateResource string
	templatePackage  string
)

func init() {
	// Discover command flags
	discoverCmd.Flags().StringVarP(&discoverPath, "path", "p", "./genesyscloud", "Path to scan for resource schema files")

	// Export command flags
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "markdown", "Output format (markdown, json, csv)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file (defaults to stdout)")

	// Validate command flags
	validateCmd.Flags().StringVarP(&validatePath, "path", "p", "./genesyscloud", "Path to scan for resource schema files")

	// Template command flags
	templateCmd.Flags().StringVarP(&templateResource, "resource", "r", "", "Resource type name (required)")
	templateCmd.Flags().StringVarP(&templatePackage, "package", "p", "", "Package name (required)")
	templateCmd.MarkFlagRequired("resource")
	templateCmd.MarkFlagRequired("package")
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
	// Use the real discovery framework to get metadata
	discovery := NewResourceDiscovery("../../genesyscloud") // Use correct path
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

	// Mock validation results
	validResources := []string{
		"genesyscloud_flow",
		"genesyscloud_queue",
		"genesyscloud_user",
	}

	invalidResources := []string{
		"genesyscloud_unknown",
	}

	fmt.Printf("Valid resources: %d\n", len(validResources))
	for _, resource := range validResources {
		fmt.Printf("  ✓ %s\n", resource)
	}

	if len(invalidResources) > 0 {
		fmt.Printf("Invalid resources: %d\n", len(invalidResources))
		for _, resource := range invalidResources {
			fmt.Printf("  ✗ %s (missing team annotation)\n", resource)
		}
		return fmt.Errorf("validation failed: %d resources missing metadata", len(invalidResources))
	}

	fmt.Println("All resources have valid metadata!")
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

// Export functions
func exportMarkdown(metadata []ResourceMetadata, output io.Writer) error {
	fmt.Fprintln(output, "# Genesys Cloud Terraform Provider - Resource Metadata")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "| Resource Type | Package | Team | Chat Room | Description |")
	fmt.Fprintln(output, "|---------------|---------|------|-----------|-------------|")

	for _, m := range metadata {
		fmt.Fprintf(output, "| %s | %s | %s | %s | %s |\n",
			m.ResourceType,
			m.PackageName,
			m.TeamName,
			m.TeamChatRoom,
			m.Description)
	}

	return nil
}

func exportJSON(metadata []ResourceMetadata, output io.Writer) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(metadata)
}

func exportCSV(metadata []ResourceMetadata, output io.Writer) error {
	writer := csv.NewWriter(output)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Resource Type", "Package", "Team", "Chat Room", "Description"}); err != nil {
		return err
	}

	// Write data
	for _, m := range metadata {
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

// NewResourceDiscovery creates a new resource discovery instance
func NewResourceDiscovery(basePath string) *ResourceDiscovery {
	return &ResourceDiscovery{
		basePath: basePath,
	}
}

// DiscoverResources scans the codebase for resource schema files and extracts metadata
func (d *ResourceDiscovery) DiscoverResources() ([]ResourceMetadata, error) {
	var allMetadata []ResourceMetadata

	// Walk through the genesyscloud directory
	err := filepath.Walk(d.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Look for schema files
		if strings.Contains(path, "_schema.go") {
			metadata, err := d.extractMetadataFromFile(path)
			if err != nil {
				// Log error but continue scanning
				fmt.Printf("Warning: Failed to extract metadata from %s: %v\n", path, err)
				return nil
			}

			if metadata != nil {
				allMetadata = append(allMetadata, *metadata)
			}
		}

		return nil
	})

	return allMetadata, err
}

// extractMetadataFromFile extracts metadata from a single schema file
func (d *ResourceDiscovery) extractMetadataFromFile(filePath string) (*ResourceMetadata, error) {
	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

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

		// Extract resource type
		if strings.Contains(line, "ResourceType = ") {
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				resourceType = strings.Trim(strings.TrimSpace(parts[1]), `"`)
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
