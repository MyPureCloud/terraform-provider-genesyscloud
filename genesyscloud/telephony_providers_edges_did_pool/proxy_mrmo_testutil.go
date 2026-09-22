package telephony_providers_edges_did_pool

import (
	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
)

// ResetDidPoolProxyForTest clears the package-level proxy singleton between tests.
func ResetDidPoolProxyForTest() {
	internalProxy = nil
}

// DidPoolProxyClientConfigForTest returns the client config bound to a proxy instance.
func DidPoolProxyClientConfigForTest(clientConfig *platformclientv2.Configuration) *platformclientv2.Configuration {
	return getTelephonyDidPoolProxy(clientConfig).clientConfig
}
