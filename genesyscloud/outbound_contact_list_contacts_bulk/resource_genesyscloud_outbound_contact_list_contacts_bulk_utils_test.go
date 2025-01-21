package outbound_contact_list_contacts_bulk

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
	testrunner "terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

func TestBulkContactsExporterResolver(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()
	subDir := "test_subdir"

	// Mock the config map
	configMap := map[string]interface{}{
		"contact_list_id": "test-contact-list",
	}

	t.Run("successful export", func(t *testing.T) {
		// Create test proxy with our test implementation
		testProxy := &contactsBulkProxy{
			getContactListContactsExportUrlAttr: func(_ context.Context, p *contactsBulkProxy, contactListId string) (string, *platformclientv2.APIResponse, error) {
				return "http://test-url.com/export", nil, nil
			},
		}

		// Set the internal proxy to our test proxy
		internalProxy = testProxy

		// Mock the provider meta
		mockMeta := &provider.ProviderMeta{
			ClientConfig: &platformclientv2.Configuration{},
		}

		// Mock resource info
		mockResource := resourceExporter.ResourceInfo{
			State: &terraform.InstanceState{
				Attributes: make(map[string]string),
			},
		}

		// Mock the file download function
		origDownloadFile := files.DownloadExportFile
		files.DownloadExportFile = func(directory, filename, url string) error {
			fullPath := path.Join(directory, filename)
			if err := os.MkdirAll(directory, os.ModePerm); err != nil {
				return err
			}
			return os.WriteFile(fullPath, []byte("test content"), 0644)
		}
		defer func() { files.DownloadExportFile = origDownloadFile }()

		// Test the function
		err := BulkContactsExporterResolver("test-id", tempDir, subDir, configMap, mockMeta, mockResource)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify the filepath was set in configMap
		expectedPath := path.Join(tempDir, subDir, "contacts_test-contact-list.csv")
		if configMap["filepath"] != expectedPath {
			t.Errorf("Expected filepath %s, got %s", expectedPath, configMap["filepath"])
		}

		// Verify file_content_hash was set
		if mockResource.State.Attributes["file_content_hash"] == "" {
			t.Error("Expected file_content_hash to be set")
		}
	})

	t.Run("export url error", func(t *testing.T) {
		// Create test proxy with error case
		testProxy := &contactsBulkProxy{
			getContactListContactsExportUrlAttr: func(_ context.Context, p *contactsBulkProxy, contactListId string) (string, *platformclientv2.APIResponse, error) {
				return "", nil, fmt.Errorf("failed to get export URL")
			},
		}

		// Set the internal proxy to our test proxy
		internalProxy = testProxy

		mockMeta := &provider.ProviderMeta{
			ClientConfig: &platformclientv2.Configuration{},
		}

		mockResource := resourceExporter.ResourceInfo{
			State: &terraform.InstanceState{
				Attributes: make(map[string]string),
			},
		}

		err := BulkContactsExporterResolver("test-id", tempDir, subDir, configMap, mockMeta, mockResource)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("download error", func(t *testing.T) {
		// Create test proxy
		testProxy := &contactsBulkProxy{
			getContactListContactsExportUrlAttr: func(_ context.Context, p *contactsBulkProxy, contactListId string) (string, *platformclientv2.APIResponse, error) {
				return "http://test-url.com/export", nil, nil
			},
		}

		// Set the internal proxy to our test proxy
		internalProxy = testProxy

		mockMeta := &provider.ProviderMeta{
			ClientConfig: &platformclientv2.Configuration{},
		}

		mockResource := resourceExporter.ResourceInfo{
			State: &terraform.InstanceState{
				Attributes: make(map[string]string),
			},
		}

		// Mock download failure
		files.DownloadExportFile = func(directory, filename, url string) error {
			return fmt.Errorf("download failed")
		}

		err := BulkContactsExporterResolver("test-id", tempDir, subDir, configMap, mockMeta, mockResource)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	// Clean up after all tests
	internalProxy = nil
}

func TestFileContentHashChanged(t *testing.T) {
	t.Parallel()

	t.Run("hash changes when file content changes", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		filepath := path.Join(tmpDir, "test.csv")

		// Create initial file
		if err := os.WriteFile(filepath, []byte("initial content"), 0644); err != nil {
			t.Fatalf("Failed to write initial file: %v", err)
		}

		initialHash, err := getFileContentHash(filepath)
		if err != nil {
			t.Fatalf("Failed to get initial hash: %v", err)
		}

		provider := testrunner.GenerateTestProvider("test_resource",
			map[string]*schema.Schema{
				"filepath": {
					Type:     schema.TypeString,
					Required: true,
				},
				"file_content_hash": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
			customdiff.ComputedIf("file_content_hash", fileContentHashChanged),
		)

		// Write new content to file
		if err := os.WriteFile(filepath, []byte("changed content"), 0644); err != nil {
			t.Fatalf("Failed to write new file content: %v", err)
		}

		diff, err := testrunner.GenerateTestDiff(
			provider,
			"test_resource",
			map[string]string{
				"filepath":          filepath,
				"file_content_hash": initialHash,
			},
			map[string]string{
				"filepath": filepath,
			},
		)

		if err != nil {
			t.Fatalf("Diff failed with error: %s", err)
		}

		if diff == nil {
			t.Error("Expected a diff when file content changes, got nil")
		} else if !diff.Attributes["file_content_hash"].NewComputed {
			t.Error("file_content_hash is not marked as NewComputed when file content changes")
		}
	})

	t.Run("hash unchanged when file content remains same", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		filepath := path.Join(tmpDir, "test.csv")
		content := []byte("test content")

		// Create file with fixed content
		if err := os.WriteFile(filepath, content, 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		hash, err := getFileContentHash(filepath)
		if err != nil {
			t.Fatalf("Failed to get hash: %v", err)
		}

		provider := testrunner.GenerateTestProvider(
			"test_resource",
			map[string]*schema.Schema{
				"filepath": {
					Type:     schema.TypeString,
					Required: true,
				},
				"file_content_hash": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
			customdiff.ComputedIf("file_content_hash", fileContentHashChanged),
		)

		diff, err := testrunner.GenerateTestDiff(
			provider,
			"test_resource",
			map[string]string{
				"filepath":          filepath,
				"file_content_hash": hash,
			},
			map[string]string{
				"filepath": filepath,
			},
		)

		if err != nil {
			t.Fatalf("Diff failed with error: %s", err)
		}

		if diff != nil {
			t.Error("Expected no diff when file content remains the same")
		}
	})
}

func TestUnitGetFileContentHash(t *testing.T) {
	// Create a temporary test file
	tempContent := []byte("test content")
	tempFile, err := os.CreateTemp("", "test_file_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up after test

	// Write content to temp file
	if err := os.WriteFile(tempFile.Name(), tempContent, 0644); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Test successful case
	t.Run("successful hash", func(t *testing.T) {
		hash, err := getFileContentHash(tempFile.Name())
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if hash == "" {
			t.Error("Expected non-empty hash")
		}
		// Known hash for "test content"
		expectedHash := "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72"
		if hash != expectedHash {
			t.Errorf("Expected hash %s, got %s", expectedHash, hash)
		}
	})

	// Test non-existent file
	t.Run("non-existent file", func(t *testing.T) {
		hash, err := getFileContentHash("non_existent_file.txt")
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
		if hash != "" {
			t.Errorf("Expected empty hash for error case, got %s", hash)
		}
	})
}

func TestUnitBuildBulkContactId(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"standard case": {
			input:    "test123",
			expected: "test123_contacts_bulk",
		},
		"empty string": {
			input:    "",
			expected: "_contacts_bulk",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := buildBulkContactId(tc.input)
			if result != tc.expected {
				t.Errorf("buildBulkContactId() = %v, want %v", result, tc.expected)
			}
		})
	}
}
