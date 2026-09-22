package exporter_test

import (
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/mrmo"
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
