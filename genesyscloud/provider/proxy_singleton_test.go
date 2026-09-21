package provider

import (
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/mrmo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubProxy struct {
	clientConfig *platformclientv2.Configuration
}

func newStubProxy(clientConfig *platformclientv2.Configuration) *stubProxy {
	return &stubProxy{clientConfig: clientConfig}
}

func TestSingletonOrFresh_usesSingletonWhenMRMOInactive(t *testing.T) {
	t.Parallel()

	var slot *stubProxy
	sourceCfg := &platformclientv2.Configuration{BasePath: "source"}
	targetCfg := &platformclientv2.Configuration{BasePath: "target"}

	first := SingletonOrFresh(&slot, sourceCfg, newStubProxy)
	second := SingletonOrFresh(&slot, targetCfg, newStubProxy)

	require.NotNil(t, first)
	assert.Same(t, first, second)
	assert.Same(t, sourceCfg, first.clientConfig)
}

func TestSingletonOrFresh_returnsFreshProxyWhenMRMOActive(t *testing.T) {
	mrmo.Reset()
	t.Setenv(mrmo.MRMO_CXASCODE_INTEGRATION_ENABLED, "true")
	t.Cleanup(func() {
		t.Setenv(mrmo.MRMO_CXASCODE_INTEGRATION_ENABLED, "")
		mrmo.Reset()
	})

	var slot *stubProxy
	sourceCfg := &platformclientv2.Configuration{BasePath: "source"}
	targetCfg := &platformclientv2.Configuration{BasePath: "target"}

	exportProxy := SingletonOrFresh(&slot, sourceCfg, newStubProxy)
	applyProxy := SingletonOrFresh(&slot, targetCfg, newStubProxy)

	require.NotNil(t, exportProxy)
	require.NotNil(t, applyProxy)
	assert.NotSame(t, exportProxy, applyProxy)
	assert.Same(t, sourceCfg, exportProxy.clientConfig)
	assert.Same(t, targetCfg, applyProxy.clientConfig)
}
