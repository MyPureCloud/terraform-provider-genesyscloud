package genesyscloud

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v89/platformclientv2"
)

var (
	sdkConfig *platformclientv2.Configuration
)

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactories = map[string]func() (*schema.Provider, error){
	"genesyscloud": func() (*schema.Provider, error) {
		return New("0.1.0")(), nil
	},
}

func TestProvider(t *testing.T) {
	if err := New("0.1.0")().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("GENESYSCLOUD_OAUTHCLIENT_ID"); v == "" {
		t.Fatal("Missing env GENESYSCLOUD_OAUTHCLIENT_ID")
	}
	if v := os.Getenv("GENESYSCLOUD_OAUTHCLIENT_SECRET"); v == "" {
		t.Fatal("Missing env GENESYSCLOUD_OAUTHCLIENT_SECRET")
	}
	if v := os.Getenv("GENESYSCLOUD_REGION"); v == "" {
		os.Setenv("GENESYSCLOUD_REGION", "dca") // Default to dev environment
	}
}

func authorizeSdk() error {
	// Create new config
	sdkConfig = platformclientv2.GetDefaultConfiguration()

	sdkConfig.BasePath = getRegionBasePath(os.Getenv("GENESYSCLOUD_REGION"))

	err := sdkConfig.AuthorizeClientCredentials(os.Getenv("GENESYSCLOUD_OAUTHCLIENT_ID"), os.Getenv("GENESYSCLOUD_OAUTHCLIENT_SECRET"))
	if err != nil {
		return err
	}

	return nil
}
