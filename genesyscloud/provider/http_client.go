package provider

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

var (
	transportsMu sync.Mutex
	transports   = map[*platformclientv2.ProxyConfiguration]*http.Transport{}
)

// NewHTTPClient returns an *http.Client for requests made outside the SDK (e.g. presigned S3 uploads and file downloads)
// whose proxy is selected by ProxyFunc. If cfg is nil the default SDK configuration is used.
func NewHTTPClient(cfg *platformclientv2.Configuration) *http.Client {
	if cfg == nil {
		cfg = platformclientv2.GetDefaultConfiguration()
	}
	return &http.Client{Transport: transport(cfg)}
}

// ProxyFunc returns a proxy selector for an http.Transport. The provider `proxy` block is used when set,
// otherwise it falls back to the HTTP(S)_PROXY/NO_PROXY env vars, and finally to no proxy.
// If cfg is nil the default SDK configuration is used.
func ProxyFunc(cfg *platformclientv2.Configuration) func(*http.Request) (*url.URL, error) {
	if cfg == nil {
		cfg = platformclientv2.GetDefaultConfiguration()
	}
	if cfg == nil || cfg.ProxyConfiguration == nil {
		return http.ProxyFromEnvironment
	}
	return http.ProxyURL(proxyURL(cfg.ProxyConfiguration))
}

// transport returns a cached transport per proxy configuration so connections are reused across clients
func transport(cfg *platformclientv2.Configuration) *http.Transport {
	var key *platformclientv2.ProxyConfiguration
	if cfg != nil {
		key = cfg.ProxyConfiguration
	}

	transportsMu.Lock()
	defer transportsMu.Unlock()

	if tr, ok := transports[key]; ok {
		return tr
	}

	var tr *http.Transport
	if defaultTransport, ok := http.DefaultTransport.(*http.Transport); ok {
		tr = defaultTransport.Clone()
	} else {
		tr = &http.Transport{}
	}
	tr.Proxy = ProxyFunc(cfg)
	transports[key] = tr
	return tr
}

// proxyURL builds the proxy URL the same way the SDK does: auth is only included when both username and password are set
func proxyURL(pc *platformclientv2.ProxyConfiguration) *url.URL {
	u := &url.URL{
		Scheme: pc.Protocol,
		Host:   pc.Host + ":" + pc.Port,
	}
	if pc.Auth != nil && pc.Auth.UserName != "" && pc.Auth.Password != "" {
		u.User = url.UserPassword(pc.Auth.UserName, pc.Auth.Password)
	}
	return u
}
