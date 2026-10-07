package exporter_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/mrmo"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/mrmo/exporter"
	didpool "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_did_pool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMRMO_ExportThenApplyUsesDistinctClientConfigs(t *testing.T) {
	mrmo.Reset()
	t.Setenv(mrmo.MRMO_CXASCODE_INTEGRATION_ENABLED, "true")
	t.Cleanup(func() {
		t.Setenv(mrmo.MRMO_CXASCODE_INTEGRATION_ENABLED, "")
		mrmo.Reset()
		didpool.ResetDidPoolProxyForTest()
	})

	sourceCfg := &platformclientv2.Configuration{BasePath: "https://source.example"}
	targetCfg := &platformclientv2.Configuration{BasePath: "https://target.example"}

	didpool.ResetDidPoolProxyForTest()
	exportProxyCfg := didpool.DidPoolProxyClientConfigForTest(sourceCfg)

	didpool.ResetDidPoolProxyForTest()
	applyProxyCfg := didpool.DidPoolProxyClientConfigForTest(targetCfg)

	require.NotNil(t, exportProxyCfg)
	require.NotNil(t, applyProxyCfg)
	assert.Same(t, sourceCfg, exportProxyCfg)
	assert.Same(t, targetCfg, applyProxyCfg)
	assert.NotSame(t, exportProxyCfg, applyProxyCfg)
}

// newOAuthStub returns a stub OAuth token endpoint for CreateClientConfig to authorize
// against. AuthorizeClientCredentials POSTs to BasePath + "/oauth/token".
func newOAuthStub(t *testing.T, token string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"bearer","expires_in":86400}`, token)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestMRMO_CreateClientConfigFeedsDistinctProxyConfigs is the end-to-end form of
// TestMRMO_ExportThenApplyUsesDistinctClientConfigs above.
//
// That test hand-builds its two Configuration values, so it verifies the proxy layer
// honours whatever configs it is handed while leaving the factory outside the assertion.
// This one obtains both configs from exporter.CreateClientConfig, which is where the
// isolation actually has to start: when that factory returned the SDK's sync.Once
// singleton, the second call re-authorized the first caller's config and the export and
// apply proxies silently shared one org's credentials.
func TestMRMO_CreateClientConfigFeedsDistinctProxyConfigs(t *testing.T) {
	mrmo.Reset()
	t.Setenv(mrmo.MRMO_CXASCODE_INTEGRATION_ENABLED, "true")
	t.Cleanup(func() {
		t.Setenv(mrmo.MRMO_CXASCODE_INTEGRATION_ENABLED, "")
		mrmo.Reset()
		didpool.ResetDidPoolProxyForTest()
	})

	sourceSrv := newOAuthStub(t, "source-token")
	targetSrv := newOAuthStub(t, "target-token")

	sourceCfg, err := exporter.CreateClientConfig(exporter.Credentials{
		ClientId:         "source-client",
		ClientSecret:     "source-secret",
		Region:           "us-east-1",
		BasePathOverride: sourceSrv.URL,
	})
	require.NoError(t, err)
	require.NotNil(t, sourceCfg)

	targetCfg, err := exporter.CreateClientConfig(exporter.Credentials{
		ClientId:         "target-client",
		ClientSecret:     "target-secret",
		Region:           "us-east-1",
		BasePathOverride: targetSrv.URL,
	})
	require.NoError(t, err)
	require.NotNil(t, targetCfg)

	require.NotSame(t, sourceCfg, targetCfg,
		"CreateClientConfig must not hand the export and apply paths the same *Configuration")

	didpool.ResetDidPoolProxyForTest()
	exportProxyCfg := didpool.DidPoolProxyClientConfigForTest(sourceCfg)

	didpool.ResetDidPoolProxyForTest()
	applyProxyCfg := didpool.DidPoolProxyClientConfigForTest(targetCfg)

	require.NotNil(t, exportProxyCfg)
	require.NotNil(t, applyProxyCfg)
	assert.Same(t, sourceCfg, exportProxyCfg)
	assert.Same(t, targetCfg, applyProxyCfg)
	assert.NotSame(t, exportProxyCfg, applyProxyCfg)

	// The point of the isolation: the two proxies must be authenticated to different orgs.
	assert.Equal(t, "source-token", exportProxyCfg.AccessToken,
		"the export proxy is using the apply org's access token")
	assert.Equal(t, "target-token", applyProxyCfg.AccessToken)
	assert.NotEqual(t, exportProxyCfg.BasePath, applyProxyCfg.BasePath)
}
