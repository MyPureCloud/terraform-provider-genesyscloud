package provider

import (
	"net/http"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitProxyURL(t *testing.T) {
	tests := []struct {
		name string
		pc   *platformclientv2.ProxyConfiguration
		want string
	}{
		{
			name: "no auth",
			pc:   &platformclientv2.ProxyConfiguration{Protocol: "http", Host: "proxy.local", Port: "3128"},
			want: "http://proxy.local:3128",
		},
		{
			name: "with auth",
			pc: &platformclientv2.ProxyConfiguration{Protocol: "https", Host: "proxy.local", Port: "8443",
				Auth: &platformclientv2.Auth{UserName: "user", Password: "p@ss"}},
			want: "https://user:p%40ss@proxy.local:8443",
		},
		{
			name: "auth missing password is ignored",
			pc: &platformclientv2.ProxyConfiguration{Protocol: "http", Host: "proxy.local", Port: "3128",
				Auth: &platformclientv2.Auth{UserName: "user"}},
			want: "http://proxy.local:3128",
		},
		{
			name: "auth missing username is ignored",
			pc: &platformclientv2.ProxyConfiguration{Protocol: "http", Host: "proxy.local", Port: "3128",
				Auth: &platformclientv2.Auth{Password: "secret"}},
			want: "http://proxy.local:3128",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := proxyURL(tt.pc)
			require.NotNil(t, u)
			assert.Equal(t, tt.want, u.String())
		})
	}
}

func TestUnitNewHTTPClientWithoutProxy(t *testing.T) {
	cfg := platformclientv2.NewConfiguration()
	cfg.ProxyConfiguration = nil

	client := NewHTTPClient(cfg)

	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok, "expected *http.Transport, got %T", client.Transport)
	req, err := http.NewRequest(http.MethodPut, "https://bucket.s3.amazonaws.com/upload", nil)
	require.NoError(t, err)
	got, err := transport.Proxy(req)
	want, wantErr := http.ProxyFromEnvironment(req)
	assert.Equal(t, wantErr, err)
	assert.Equal(t, want, got, "without a proxy block the HTTP(S)_PROXY env vars should decide")
}

func TestUnitNewHTTPClientWithProxy(t *testing.T) {
	cfg := platformclientv2.NewConfiguration()
	cfg.ProxyConfiguration = &platformclientv2.ProxyConfiguration{
		Protocol: "http",
		Host:     "proxy.local",
		Port:     "3128",
		Auth:     &platformclientv2.Auth{UserName: "user", Password: "secret"},
	}

	client := NewHTTPClient(cfg)

	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok, "expected *http.Transport, got %T", client.Transport)
	require.NotNil(t, transport.Proxy)

	req, err := http.NewRequest(http.MethodPut, "https://bucket.s3.amazonaws.com/upload", nil)
	require.NoError(t, err)
	got, err := transport.Proxy(req)
	require.NoError(t, err)
	assert.Equal(t, "http://user:secret@proxy.local:3128", got.String())
}

func TestUnitNewHTTPClientReusesTransport(t *testing.T) {
	cfg := platformclientv2.NewConfiguration()
	cfg.ProxyConfiguration = &platformclientv2.ProxyConfiguration{Protocol: "http", Host: "proxy.local", Port: "3128"}

	first := NewHTTPClient(cfg)
	second := NewHTTPClient(cfg)

	assert.Same(t, first.Transport, second.Transport, "transport should be cached so connections are reused")
}

func TestUnitNewHTTPClientNilConfigUsesDefault(t *testing.T) {
	defaultCfg := platformclientv2.GetDefaultConfiguration()
	orig := defaultCfg.ProxyConfiguration
	t.Cleanup(func() { defaultCfg.ProxyConfiguration = orig })

	defaultCfg.ProxyConfiguration = &platformclientv2.ProxyConfiguration{Protocol: "http", Host: "default.proxy", Port: "8080"}

	client := NewHTTPClient(nil)

	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok, "expected *http.Transport, got %T", client.Transport)
	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)
	got, err := transport.Proxy(req)
	require.NoError(t, err)
	assert.Equal(t, "http://default.proxy:8080", got.String())
}

func TestUnitProxyFuncUsesProxyBlock(t *testing.T) {
	cfg := platformclientv2.NewConfiguration()
	cfg.ProxyConfiguration = &platformclientv2.ProxyConfiguration{Protocol: "http", Host: "proxy.local", Port: "3128"}

	req, err := http.NewRequest(http.MethodGet, "https://bucket.s3.amazonaws.com/key", nil)
	require.NoError(t, err)
	got, err := ProxyFunc(cfg)(req)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "http://proxy.local:3128", got.String())
}

func TestUnitProxyFuncFallsBackToEnvironment(t *testing.T) {
	cfg := platformclientv2.NewConfiguration()
	cfg.ProxyConfiguration = nil

	req, err := http.NewRequest(http.MethodGet, "https://bucket.s3.amazonaws.com/key", nil)
	require.NoError(t, err)
	got, err := ProxyFunc(cfg)(req)
	want, wantErr := http.ProxyFromEnvironment(req)

	assert.Equal(t, wantErr, err)
	assert.Equal(t, want, got, "without a proxy block the HTTP(S)_PROXY env vars should decide")
}
