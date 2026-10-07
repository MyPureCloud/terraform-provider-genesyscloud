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
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
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

// validateExportByTypeInput validates the export input
func validateExportByTypeInput(input ExportByTypeInput) error {
	if input.ResourceType == "" {
		return errors.New("'ResourceType' is a required field")
	}
	if input.GenerateOutputFiles && input.Directory == "" {
		return errors.New("'Directory' is a required field when 'GenerateOutputFiles' is set to true")
	}
	return nil
}

// generateDefaults generates the default values for the export input
func generateDefaults(input *BaseExportInput) {
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

	// Copy the SDK's default Configuration rather than returning it.
	//
	// platformclientv2.GetDefaultConfiguration is a sync.Once singleton, and two orgs'
	// configurations coexist in one process on the MRMO path: export reads the source org
	// while apply writes the target org, and pairing bootstrap resolves both orgs' Home
	// divisions back to back. Returning the singleton means the second caller
	// re-authorizes the first caller's configuration, and both silently end up pointing
	// at whichever org was configured last.
	//
	// The SDK's own constructors cannot be used per call either: NewConfiguration spawns a
	// permanent periodicConfigUpdater goroutine, and both it and NewConfigurationWithConfigFile
	// mutate package-level state — the trace/debug/error loggers via configureLogging, and
	// global viper state via updateConfigFromFile. A struct copy avoids all of that, and
	// Configuration holds no locks so copying it is safe.
	base := platformclientv2.GetDefaultConfiguration()
	config := *base

	// A struct copy shares map headers, so clone them: a per-org header or API-key write
	// must not bleed across tenants. LoggingConfiguration and RetryConfiguration stay
	// shared deliberately — they hold no per-org state, are read-only on this path, and
	// LoggingConfiguration's fields are unexported so it cannot be rebuilt outside the SDK
	// (the API client dereferences it, so it must be non-nil).
	config.DefaultHeader = cloneStringMap(base.DefaultHeader)
	config.APIKey = cloneStringMap(base.APIKey)
	config.APIKeyPrefix = cloneStringMap(base.APIKeyPrefix)

	// Never inherit another org's identity from the singleton.
	config.AccessToken = ""
	config.RefreshToken = ""
	config.OAuthToken = ""
	config.ClientID = ""
	config.ClientSecret = ""
	config.UserName = ""
	config.Password = ""

	if creds.BasePathOverride != "" {
		config.BasePath = creds.BasePathOverride
	} else {
		config.BasePath = provider.GetRegionBasePath(creds.Region)
	}

	// Rebind the API client so it references this copy instead of the singleton.
	// NewAPIClient reuses the existing HTTP client, so the connection pool is preserved.
	config.APIClient = platformclientv2.NewAPIClient(&config)

	err = config.AuthorizeClientCredentials(creds.ClientId, creds.ClientSecret)
	return &config, err
}

// cloneStringMap returns an independent copy of m, never nil. Used so configs produced
// by CreateClientConfig do not share mutable header/API-key maps with one another or
// with the SDK's default Configuration.
func cloneStringMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// createExportResourceData generates the export resource config that the genesyscloud tf exporter will use
func createExportResourceData(s map[string]*schema.Schema, input BaseExportInput) *schema.ResourceData {
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
