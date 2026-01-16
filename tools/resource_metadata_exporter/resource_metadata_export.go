package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version information
	Version = "0.1.0"
	Commit  = "dev"
	Date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "resource-metadata-exporter",
		Short: "Export and manage resource metadata for Genesys Cloud Terraform Provider",
		Long: `A command-line tool for managing resource metadata in the Genesys Cloud Terraform Provider.
This tool provides functionality to:
- Discover and extract metadata from resource schema files
- Export metadata in multiple formats (Markdown, JSON, CSV)
- Validate metadata completeness
- Generate metadata templates for new resources
Examples:
  resource-metadata-exporter discover --path ./genesyscloud
  resource-metadata-exporter export --format markdown --output metadata.md
  resource-metadata-exporter validate --path ./genesyscloud`,
		Version: Version,
	}

	// Add subcommands
	rootCmd.AddCommand(discoverCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(templateCmd)

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}