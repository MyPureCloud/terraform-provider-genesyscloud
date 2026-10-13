package exporter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTokenServer returns a stub OAuth token endpoint. AuthorizeClientCredentials
// POSTs to BasePath + "/oauth/token" (the //api. -> //login. rewrite is a no-op for
// a host:port base path), so any path is accepted here.
func newTokenServer(t *testing.T, token string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"bearer","expires_in":86400}`, token)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestCreateClientConfigReturnsIndependentConfigs guards against CreateClientConfig
// handing out the SDK's process-wide default Configuration.
//
// Two orgs' configurations coexist in one process on the MRMO path: export reads the
// source org while apply writes the target org, and the pairing bootstrap resolves both
// orgs' Home divisions back to back. platformclientv2.GetDefaultConfiguration is a
// sync.Once singleton, so returning it means the second call re-authorizes the first
// caller's configuration and both callers end up pointing at whichever org was
// configured last — silently, with no error.
//
// The production incident this guards: bootstrap resolved both the source and target
// Home divisions through the same (target-authenticated) config, mapped the target org's
// Home division to itself, and took the replica claim on it. That blocked the legitimate
// source-Home mapping permanently and stalled replication for every resource in the Home
// division.
func TestCreateClientConfigReturnsIndependentConfigs(t *testing.T) {
	sourceSrv := newTokenServer(t, "source-token")
	targetSrv := newTokenServer(t, "target-token")

	sourceCfg, err := CreateClientConfig(Credentials{
		ClientId:         "source-client",
		ClientSecret:     "source-secret",
		Region:           "us-east-1",
		BasePathOverride: sourceSrv.URL,
	})
	require.NoError(t, err)
	require.NotNil(t, sourceCfg)

	targetCfg, err := CreateClientConfig(Credentials{
		ClientId:         "target-client",
		ClientSecret:     "target-secret",
		Region:           "us-east-1",
		BasePathOverride: targetSrv.URL,
	})
	require.NoError(t, err)
	require.NotNil(t, targetCfg)

	assert.NotSame(t, sourceCfg, targetCfg,
		"CreateClientConfig returned the same *Configuration twice; it must not hand out the SDK's shared default")

	assert.Equal(t, sourceSrv.URL, sourceCfg.BasePath,
		"the source config's BasePath was overwritten by the second CreateClientConfig call")
	assert.Equal(t, targetSrv.URL, targetCfg.BasePath)

	assert.Equal(t, "source-token", sourceCfg.AccessToken,
		"the source config's access token was overwritten by the second CreateClientConfig call")
	assert.Equal(t, "target-token", targetCfg.AccessToken)
}

// TestCreateClientConfigDoesNotShareHeaderMaps guards the shallow-copy hazard: a
// struct copy shares map headers, so per-config header or API-key writes would be
// visible across every config in the process.
func TestCreateClientConfigDoesNotShareHeaderMaps(t *testing.T) {
	srv := newTokenServer(t, "tok")

	first, err := CreateClientConfig(Credentials{
		ClientId: "a", ClientSecret: "b", Region: "us-east-1", BasePathOverride: srv.URL,
	})
	require.NoError(t, err)
	second, err := CreateClientConfig(Credentials{
		ClientId: "c", ClientSecret: "d", Region: "us-east-1", BasePathOverride: srv.URL,
	})
	require.NoError(t, err)

	require.NotNil(t, first.DefaultHeader)
	require.NotNil(t, second.DefaultHeader)

	first.DefaultHeader["X-Mrmo-Test"] = "source-only"
	assert.NotContains(t, second.DefaultHeader, "X-Mrmo-Test",
		"DefaultHeader map is shared between configs; a per-org header write bleeds across tenants")
}
