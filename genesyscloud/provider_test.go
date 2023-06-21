package genesyscloud

import (
	"os"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v102/platformclientv2"
)

var (
	sdkConfig *platformclientv2.Configuration
)

func TestProvider(t *testing.T) {
	if err := New("0.1.0")().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
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
