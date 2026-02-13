package exporter

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"errors"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

// validateExportInput validates the export input
func validateExportInput(input ExportInput) error {
	if input.ResourceType == "" {
		return errors.New("'ResourceType' is a required field")
	}
	if input.EntityId == "" {
		return errors.New("'EntityId' is a required field")
	}
	if input.GenerateOutputFiles && input.Directory == "" {
		return errors.New("'Directory' is a required field when 'GenerateOutputFiles' is set to true")
	}
	return nil
}

// generateDefaults generates the default values for the export input
func generateDefaults(input *ExportInput) {
	// Setting Directory to a default value if GenerateOutputFiles is false
	// This is a precaution to ensure that, if output folder are unexpectedly generated, they will be written
	// to a uniquely named folder in the temp directory.
	if !input.GenerateOutputFiles && input.Directory == "" {
		defaultValue := filepath.Join(os.TempDir(), "mrmo_"+uuid.NewString())
		log.Println("Setting 'Directory' to ", defaultValue)
		input.Directory = defaultValue
	}
}

// CreateClientConfig creates the client config for the export
func CreateClientConfig(creds Credentials) (_ *platformclientv2.Configuration, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("createClientConfig: %w", err)
		}
	}()

	if creds.ClientId == "" || creds.ClientSecret == "" || creds.Region == "" {
		return nil, fmt.Errorf("insufficient client information provided")
	}

	config := platformclientv2.GetDefaultConfiguration()
	if creds.BasePathOverride != "" {
		config.BasePath = creds.BasePathOverride
	} else {
		config.BasePath = provider.GetRegionBasePath(creds.Region)
	}

	err = config.AuthorizeClientCredentials(creds.ClientId, creds.ClientSecret)
	return config, err
}

// createExportResourceData generates the export resource config that the genesyscloud tf exporter will use
func createExportResourceData(s map[string]*schema.Schema, input ExportInput) *schema.ResourceData {
	config := map[string]any{
		"directory":                          input.Directory,
		"include_state_file":                 input.IncludeStateFile,
		"export_format":                      "json",
		"include_filter_resources":           []any{input.ResourceType},
		"use_legacy_architect_flow_exporter": false,
	}

	var t testing.T
	return schema.TestResourceDataRaw(&t, s, config)
}
