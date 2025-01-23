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
			BlockLabel: "contacts_test-contact-list",
			State: &terraform.InstanceState{
				Attributes: make(map[string]string),
			},
		}

		// Mock the file download function
		origDownloadFile := files.DownloadExportFileWithAccessToken
		files.DownloadExportFileWithAccessToken = func(directory, filename, url, accessToken string) error {
			fullPath := path.Join(directory, filename)
			if err := os.MkdirAll(directory, os.ModePerm); err != nil {
				return err
			}
			return os.WriteFile(fullPath, []byte("test content"), 0644)
		}
		defer func() { files.DownloadExportFileWithAccessToken = origDownloadFile }()

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
		files.DownloadExportFileWithAccessToken = func(directory, filename, url, accessToken string) error {
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
