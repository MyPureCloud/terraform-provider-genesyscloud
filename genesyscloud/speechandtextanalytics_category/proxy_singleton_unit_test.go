package speechandtextanalytics_category

import (
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/mrmo"
	"github.com/stretchr/testify/assert"
)

// TestUnitProxyIsolatedPerConfigInMRMOMode ensures each client config gets its own proxy when MRMO mode is active,
// so a call made with org B's config never reuses the proxy bound to org A.
func TestUnitProxyIsolatedPerConfigInMRMOMode(t *testing.T) {
	mrmo.Reset()
	t.Setenv(mrmo.MRMO_CXASCODE_INTEGRATION_ENABLED, "true")
	oldProxy := internalProxy
	internalProxy = nil
	t.Cleanup(func() {
		internalProxy = oldProxy
		t.Setenv(mrmo.MRMO_CXASCODE_INTEGRATION_ENABLED, "")
		mrmo.Reset()
	})

	cfgA := &platformclientv2.Configuration{BasePath: "https://org-a.invalid"}
	cfgB := &platformclientv2.Configuration{BasePath: "https://org-b.invalid"}

	proxyA := getCategoryProxy(cfgA)
	proxyB := getCategoryProxy(cfgB)

	assert.NotSame(t, proxyA, proxyB)
	assert.Same(t, cfgA, proxyA.clientConfig)
	assert.Same(t, cfgB, proxyB.clientConfig)
}
