package provider

import (
	"context"
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestMain(m *testing.M) {
	// Setup
	cleanup := func() {
		if sigChan != nil {
			signal.Stop(sigChan)
			close(sigChan)
		}
		if SdkClientPool != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_ = SdkClientPool.Close(ctx)
		}
	}

	// Run tests
	code := m.Run()

	// Cleanup
	cleanup()
	os.Exit(code)
}

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
		AttrTokenAcquireTimeout: "1m",
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
