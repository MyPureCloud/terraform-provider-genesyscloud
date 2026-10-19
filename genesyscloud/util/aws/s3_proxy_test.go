package aws

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	localStackEnv "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/aws/localstack/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnitDownloadFileUsesProviderProxy verifies S3 reads for s3:// file paths are sent through the provider proxy block
func TestUnitDownloadFileUsesProviderProxy(t *testing.T) {
	const upstreamHost = "upstream.invalid"

	var mu sync.Mutex
	var seenHosts []string
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seenHosts = append(seenHosts, r.URL.Host)
		mu.Unlock()
		_, _ = w.Write([]byte("flow: yaml"))
	}))
	defer proxyServer.Close()

	proxyURL, err := url.Parse(proxyServer.URL)
	require.NoError(t, err)
	host, port, err := net.SplitHostPort(proxyURL.Host)
	require.NoError(t, err)

	defaultCfg := platformclientv2.GetDefaultConfiguration()
	orig := defaultCfg.ProxyConfiguration
	t.Cleanup(func() { defaultCfg.ProxyConfiguration = orig })
	defaultCfg.ProxyConfiguration = &platformclientv2.ProxyConfiguration{Protocol: "http", Host: host, Port: port}

	// Static credentials, region and endpoint so no real AWS lookup happens
	t.Setenv(localStackEnv.UseLocalStackEnvVar, "false")
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ENDPOINT_URL_S3", "http://"+upstreamHost)
	t.Setenv("AWS_CONFIG_FILE", "/dev/null")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", "/dev/null")

	reader, err := DownloadFile(context.Background(), "my-bucket", "flows/inbound.yaml")

	require.NoError(t, err)
	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "flow: yaml", string(data))
	require.Len(t, seenHosts, 1, "expected the S3 request to go through the proxy")
	assert.True(t, strings.HasSuffix(seenHosts[0], upstreamHost), "unexpected upstream host %q", seenHosts[0])
}
