package integration_action

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnitUploadIntegrationActionDraftFunctionUsesProviderProxy verifies both the signed URL request and the
// S3 upload are sent through the provider proxy block rather than directly
func TestUnitUploadIntegrationActionDraftFunctionUsesProviderProxy(t *testing.T) {
	const upstreamHost = "upstream.invalid"

	var mu sync.Mutex
	var seen []string
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.Method+" "+r.URL.Host+r.URL.Path)
		mu.Unlock()
		_, _ = io.Copy(io.Discard, r.Body)
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"url":"http://` + upstreamHost + `/s3/function.zip","headers":{}}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer proxyServer.Close()

	proxyURL, err := url.Parse(proxyServer.URL)
	require.NoError(t, err)
	host, port, err := net.SplitHostPort(proxyURL.Host)
	require.NoError(t, err)

	cfg := platformclientv2.NewConfiguration()
	cfg.BasePath = "http://" + upstreamHost
	cfg.AccessToken = "token"
	cfg.ProxyConfiguration = &platformclientv2.ProxyConfiguration{Protocol: "http", Host: host, Port: port}

	zipPath := filepath.Join(t.TempDir(), "function.zip")
	require.NoError(t, os.WriteFile(zipPath, []byte("zip-bytes"), 0600))

	p := &integrationActionsProxy{clientConfig: cfg}
	resp, err := uploadIntegrationActionDraftFunctionFn(context.Background(), p, "action-id", zipPath)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, []string{
		"POST " + upstreamHost + "/api/v2/integrations/actions/action-id/draft/function/upload",
		"PUT " + upstreamHost + "/s3/function.zip",
	}, seen)
}
