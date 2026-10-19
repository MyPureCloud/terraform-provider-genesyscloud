package files

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// upstreamHost is never resolvable, so a request only succeeds if it is sent through the proxy
const upstreamHost = "upstream.invalid"

type fakeProxy struct {
	mu       sync.Mutex
	requests []*http.Request
	authz    []string
}

func (f *fakeProxy) recorded() ([]*http.Request, []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]*http.Request(nil), f.requests...), append([]string(nil), f.authz...)
}

// startFakeProxy starts an HTTP forward proxy that answers every request itself with body
// and points the default SDK configuration's proxy block at it for the duration of the test.
func startFakeProxy(t *testing.T, auth *platformclientv2.Auth, body string) *fakeProxy {
	t.Helper()
	fp := &fakeProxy{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fp.mu.Lock()
		fp.requests = append(fp.requests, r.Clone(context.Background()))
		fp.authz = append(fp.authz, r.Header.Get("Proxy-Authorization"))
		fp.mu.Unlock()
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	host, port, err := net.SplitHostPort(serverURL.Host)
	require.NoError(t, err)

	cfg := platformclientv2.GetDefaultConfiguration()
	orig := cfg.ProxyConfiguration
	cfg.ProxyConfiguration = &platformclientv2.ProxyConfiguration{
		Protocol: "http",
		Host:     host,
		Port:     port,
		Auth:     auth,
	}
	t.Cleanup(func() { cfg.ProxyConfiguration = orig })
	return fp
}

func assertSentViaProxy(t *testing.T, fp *fakeProxy, method, path string) {
	t.Helper()
	reqs, _ := fp.recorded()
	require.Len(t, reqs, 1, "expected exactly one request through the proxy")
	assert.Equal(t, method, reqs[0].Method)
	assert.Equal(t, upstreamHost, reqs[0].URL.Host, "proxy should receive the absolute upstream URL")
	assert.Equal(t, path, reqs[0].URL.Path)
}

func TestUnitS3UploaderUsesProviderProxy(t *testing.T) {
	fp := startFakeProxy(t, nil, "ok")

	uploader := NewS3Uploader(strings.NewReader("flow: yaml"), nil, nil, map[string]string{}, http.MethodPut, "http://"+upstreamHost+"/s3/presigned")
	_, err := uploader.Upload()

	require.NoError(t, err)
	assertSentViaProxy(t, fp, http.MethodPut, "/s3/presigned")
}

func TestUnitS3UploaderSendsProxyAuth(t *testing.T) {
	fp := startFakeProxy(t, &platformclientv2.Auth{UserName: "user", Password: "secret"}, "ok")

	uploader := NewS3Uploader(strings.NewReader("flow: yaml"), nil, nil, map[string]string{}, http.MethodPut, "http://"+upstreamHost+"/s3/presigned")
	_, err := uploader.Upload()

	require.NoError(t, err)
	_, authz := fp.recorded()
	require.Len(t, authz, 1)
	// base64("user:secret")
	assert.Equal(t, "Basic dXNlcjpzZWNyZXQ=", authz[0])
}

func TestUnitDownloadOrOpenFileUsesProviderProxy(t *testing.T) {
	fp := startFakeProxy(t, nil, "file contents")

	reader, _, err := DownloadOrOpenFile(context.Background(), "http://"+upstreamHost+"/files/flow.yaml", false)

	require.NoError(t, err)
	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "file contents", string(data))
	assertSentViaProxy(t, fp, http.MethodGet, "/files/flow.yaml")
}

func TestUnitDownloadExportFileUsesProviderProxy(t *testing.T) {
	fp := startFakeProxy(t, nil, "exported")
	dir := t.TempDir()

	_, err := DownloadExportFileWithAccessToken(dir, "export.yaml", "http://"+upstreamHost+"/exports/export.yaml", "token")

	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(dir, "export.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "exported", string(data))
	assertSentViaProxy(t, fp, http.MethodGet, "/exports/export.yaml")
}

func TestUnitDownloadAndReadCSVFromURIUsesProviderProxy(t *testing.T) {
	fp := startFakeProxy(t, nil, "a,b\n1,2\n")

	records, err := DownloadAndReadCSVFromURI("http://"+upstreamHost+"/dnc.csv", "token")

	require.NoError(t, err)
	assert.Equal(t, [][]string{{"a", "b"}, {"1", "2"}}, records)
	assertSentViaProxy(t, fp, http.MethodGet, "/dnc.csv")
}
