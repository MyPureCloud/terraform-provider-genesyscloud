package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestDiscoverCommand(t *testing.T) {
	// Create test data directory
	testDataDir := "./testdata"
	if err := os.MkdirAll(testDataDir, 0755); err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}
	defer os.RemoveAll(testDataDir)

	// Create a test schema file with annotations
	testSchemaFile := filepath.Join(testDataDir, "test_schema.go")
	testContent := `package test_package

// @team: Test Team
// @chat: #test-team
// @description: Test resource

const ResourceType = "genesyscloud_test"

func ResourceTest() *schema.Resource {
    return &schema.Resource{}
}
`
	if err := os.WriteFile(testSchemaFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test schema file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run discover command
	discoverCmd.SetArgs([]string{"--path", testDataDir})
	err := discoverCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("Discover command failed: %v", err)
	}

	// Check that output contains expected content
	if !strings.Contains(output, "Discovering resources in:") {
		t.Error("Expected output to contain discovery message")
	}

	if !strings.Contains(output, "Discovered") {
		t.Error("Expected output to contain discovery results")
	}
}

func TestExportCommand_Markdown(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run export command
	exportCmd.SetArgs([]string{"--format", "markdown"})
	err := exportCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("Export command failed: %v", err)
	}

	// Check markdown output
	if !strings.Contains(output, "# Genesys Cloud Terraform Provider - Resource Metadata") {
		t.Error("Expected markdown header")
	}

	if !strings.Contains(output, "| Resource Type | Package | Team | Chat Room | Description |") {
		t.Error("Expected markdown table header")
	}

	if !strings.Contains(output, "genesyscloud_flow") {
		t.Error("Expected resource type in output")
	}
}

func TestExportCommand_JSON(t *testing.T) {
	// Create test data directory
	testDataDir := "./testdata_export"
	if err := os.MkdirAll(testDataDir, 0755); err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}
	defer os.RemoveAll(testDataDir)

	// Create a test schema file with annotations
	testSchemaFile := filepath.Join(testDataDir, "test_schema.go")
	testContent := `package test_package

// @team: Test Team
// @chat: #test-team
// @description: Test resource

const ResourceType = "genesyscloud_test"

func ResourceTest() *schema.Resource {
    return &schema.Resource{}
}
`
	if err := os.WriteFile(testSchemaFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test schema file: %v", err)
	}

	// Note: The export command uses a hardcoded path, so we test with real data
	// In a production environment, we'd want to make the path configurable

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run export command
	exportCmd.SetArgs([]string{"--format", "json"})
	err := exportCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("Export command failed: %v", err)
	}

	// Parse JSON output
	var metadata []ResourceMetadata
	if err := json.Unmarshal([]byte(output), &metadata); err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
	}

	if len(metadata) == 0 {
		t.Error("Expected metadata in JSON output")
	}

	// Check that we have at least one resource (the actual data may vary)
	if metadata[0].ResourceType == "" {
		t.Error("Expected resource type to be set")
	}
}

func TestExportCommand_CSV(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run export command
	exportCmd.SetArgs([]string{"--format", "csv"})
	err := exportCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("Export command failed: %v", err)
	}

	// Check CSV output
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Error("Expected CSV output with header and data")
	}

	// Check header
	if !strings.Contains(lines[0], "Resource Type,Package,Team,Chat Room,Description") {
		t.Error("Expected CSV header")
	}

	// Check data
	if !strings.Contains(output, "genesyscloud_flow") {
		t.Error("Expected resource data in CSV")
	}
}

func TestExportCommand_InvalidFormat(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Run export command with invalid format
	exportCmd.SetArgs([]string{"--format", "invalid"})
	err := exportCmd.Execute()

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err == nil {
		t.Error("Expected error for invalid format")
	}

	if !strings.Contains(output, "unsupported format") {
		t.Error("Expected error message about unsupported format")
	}
}

func TestValidateCommand(t *testing.T) {
	// Capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Run validate command
	validateCmd.SetArgs([]string{"--path", "./testdata"})
	err := validateCmd.Execute()

	// Restore stdout and stderr
	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Check validation output
	if !strings.Contains(output, "Validating metadata in:") {
		t.Error("Expected validation message")
	}

	if !strings.Contains(output, "Valid resources:") {
		t.Error("Expected validation results")
	}

	// The mock validation is expected to fail, so we expect an error
	if err == nil {
		t.Error("Expected validation to fail with mock data")
	}
}

func TestTemplateCommand(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run template command
	templateCmd.SetArgs([]string{"--resource", "genesyscloud_test", "--package", "test_package"})
	err := templateCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("Template command failed: %v", err)
	}

	// Check template output
	if !strings.Contains(output, "Resource Metadata Template for genesyscloud_test") {
		t.Error("Expected template header")
	}

	if !strings.Contains(output, "@team:") {
		t.Error("Expected team annotation in template")
	}

	if !strings.Contains(output, "@chat:") {
		t.Error("Expected chat annotation in template")
	}
}

func newTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Generate metadata template for a resource",
		Long:  "Generates a metadata template that can be added to a resource schema file",
		RunE:  runTemplate,
	}
	cmd.Flags().StringVarP(&templateResource, "resource", "r", "", "Resource type name (required)")
	cmd.Flags().StringVarP(&templatePackage, "package", "p", "", "Package name (required)")
	cmd.MarkFlagRequired("resource")
	cmd.MarkFlagRequired("package")
	return cmd
}

func TestTemplateCommand_MissingRequiredFlags(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Create a new root command and attach a fresh templateCmd
	rootCmd := &cobra.Command{Use: "testroot"}
	rootCmd.AddCommand(newTemplateCmd())

	// Run root command with template subcommand and no flags
	rootCmd.SetArgs([]string{"template"})
	err := rootCmd.Execute()

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// The command should fail due to missing required flags
	if err == nil {
		t.Error("Expected error for missing required flags")
		return
	}

	// Check for error message about required flags
	if !strings.Contains(output, "required") && !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected error message about required flags, got: %v, output: %s", err, output)
	}
}

func TestExportFunctions(t *testing.T) {
	metadata := []ResourceMetadata{
		{
			ResourceType: "genesyscloud_test",
			PackageName:  "test_package",
			TeamName:     "Test Team",
			TeamChatRoom: "#test-team",
			Description:  "Test resource",
		},
	}

	// Test markdown export
	var buf bytes.Buffer
	err := exportMarkdown(metadata, &buf)
	if err != nil {
		t.Errorf("Markdown export failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "# Genesys Cloud Terraform Provider - Resource Metadata") {
		t.Error("Expected markdown header")
	}

	if !strings.Contains(output, "genesyscloud_test") {
		t.Error("Expected resource type in markdown")
	}

	// Test JSON export
	buf.Reset()
	err = exportJSON(metadata, &buf)
	if err != nil {
		t.Errorf("JSON export failed: %v", err)
	}

	var parsedMetadata []ResourceMetadata
	if err := json.Unmarshal(buf.Bytes(), &parsedMetadata); err != nil {
		t.Errorf("Failed to parse JSON: %v", err)
	}

	if len(parsedMetadata) != 1 {
		t.Error("Expected one metadata item")
	}

	// Test CSV export
	buf.Reset()
	err = exportCSV(metadata, &buf)
	if err != nil {
		t.Errorf("CSV export failed: %v", err)
	}

	output = buf.String()
	if !strings.Contains(output, "Resource Type,Package,Team,Chat Room,Description") {
		t.Error("Expected CSV header")
	}

	if !strings.Contains(output, "genesyscloud_test") {
		t.Error("Expected resource type in CSV")
	}
}

func TestCommandFlags(t *testing.T) {
	// Test discover command flags
	if discoverCmd.Flags().Lookup("path") == nil {
		t.Error("Expected path flag on discover command")
	}

	// Test export command flags
	if exportCmd.Flags().Lookup("format") == nil {
		t.Error("Expected format flag on export command")
	}

	if exportCmd.Flags().Lookup("output") == nil {
		t.Error("Expected output flag on export command")
	}

	// Test validate command flags
	if validateCmd.Flags().Lookup("path") == nil {
		t.Error("Expected path flag on validate command")
	}

	// Test template command flags
	if templateCmd.Flags().Lookup("resource") == nil {
		t.Error("Expected resource flag on template command")
	}

	if templateCmd.Flags().Lookup("package") == nil {
		t.Error("Expected package flag on template command")
	}
}

func TestMain(m *testing.M) {
	// Setup test environment
	os.Exit(m.Run())
}
