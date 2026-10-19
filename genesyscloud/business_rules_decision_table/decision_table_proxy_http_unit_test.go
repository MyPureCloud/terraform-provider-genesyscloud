package business_rules_decision_table

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const proxyTestUpstreamHost = "upstream.invalid"

// newProxiedDecisionTableProxy returns a proxy whose client config routes through a fake forward proxy,
// and a func returning the "METHOD host/path" of every request that proxy received
func newProxiedDecisionTableProxy(t *testing.T, body string) (*BusinessRulesDecisionTableProxy, func() []string) {
	t.Helper()
	var mu sync.Mutex
	var seen []string
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.Method+" "+r.URL.Host+r.URL.Path)
		mu.Unlock()
		_, _ = io.Copy(io.Discard, r.Body)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(proxyServer.Close)

	proxyURL, err := url.Parse(proxyServer.URL)
	require.NoError(t, err)
	host, port, err := net.SplitHostPort(proxyURL.Host)
	require.NoError(t, err)

	cfg := platformclientv2.NewConfiguration()
	cfg.BasePath = "http://" + proxyTestUpstreamHost
	cfg.ProxyConfiguration = &platformclientv2.ProxyConfiguration{Protocol: "http", Host: host, Port: port}

	return &BusinessRulesDecisionTableProxy{clientConfig: cfg}, func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), seen...)
	}
}

func TestUnitUploadDecisionTableImportFileUsesProviderProxy(t *testing.T) {
	p, seen := newProxiedDecisionTableProxy(t, "")

	err := uploadDecisionTableImportFileFn(context.Background(), p, "http://"+proxyTestUpstreamHost+"/s3/import.csv", map[string]string{}, []byte("a,b\n"))

	require.NoError(t, err)
	assert.Equal(t, []string{"PUT " + proxyTestUpstreamHost + "/s3/import.csv"}, seen())
}

func TestUnitDownloadDecisionTableExportUsesProviderProxy(t *testing.T) {
	p, seen := newProxiedDecisionTableProxy(t, "a,b\n1,2\n")

	data, err := downloadDecisionTableExportFn(context.Background(), p, "/api/v2/downloads/export-id")

	require.NoError(t, err)
	assert.Equal(t, "a,b\n1,2\n", string(data))
	assert.Equal(t, []string{"GET " + proxyTestUpstreamHost + "/api/v2/downloads/export-id"}, seen())
}
