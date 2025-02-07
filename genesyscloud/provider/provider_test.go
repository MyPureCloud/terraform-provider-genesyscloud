package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProvider(t *testing.T) {
	if err := New("0.1.0", make(map[string]*schema.Resource), make(map[string]*schema.Resource))().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

// testProviderConfig creates a ResourceData with default test values
func testProviderConfig(t *testing.T) *schema.ResourceData {
	return schema.TestResourceDataRaw(t, ProviderSchema(), map[string]interface{}{
		"access_token":          "test-token",
		"oauthclient_id":        "test-client-id",
		"oauthclient_secret":    "test-client-secret",
		"environment":           "mypurecloud.com",
		AttrTokenPoolSize:       5,
		AttrTokenAcquireTimeout: "1s",
		AttrTokenInitTimeout:    "5s",
		AttrSdkClientPoolDebug:  false,
	})
}

// testProviderConfigCustom allows overriding default values
func testProviderConfigCustom(t *testing.T, customValues map[string]interface{}) *schema.ResourceData {
	defaultValues := map[string]interface{}{
		"access_token":          "test-token",
		"client_id":             "test-client-id",
		"client_secret":         "test-client-secret",
		"environment":           "mypurecloud.com",
		AttrTokenPoolSize:       5,
		AttrTokenAcquireTimeout: "1s",
		AttrTokenInitTimeout:    "5s",
		AttrSdkClientPoolDebug:  false,
	}

	// Merge custom values with defaults
	for k, v := range customValues {
		defaultValues[k] = v
	}

	return schema.TestResourceDataRaw(t, ProviderSchema(), defaultValues)
}
