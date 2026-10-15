package resource_cache

import (
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/mrmo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheForProxy_usesSharedCacheWhenMRMOInactive(t *testing.T) {
	t.Parallel()

	shared := NewResourceCache[string]()
	cached := CacheForProxy(shared)

	require.NotNil(t, cached)
	assert.Same(t, shared, cached)
}

func TestCacheForProxy_returnsFreshCacheWhenMRMOActive(t *testing.T) {
	mrmo.Reset()
	t.Setenv(mrmo.MRMO_CXASCODE_INTEGRATION_ENABLED, "true")
	t.Cleanup(func() {
		t.Setenv(mrmo.MRMO_CXASCODE_INTEGRATION_ENABLED, "")
		mrmo.Reset()
	})

	shared := NewResourceCache[string]()
	exportCache := CacheForProxy(shared)
	applyCache := CacheForProxy(shared)

	require.NotNil(t, exportCache)
	require.NotNil(t, applyCache)
	assert.NotSame(t, shared, exportCache)
	assert.NotSame(t, exportCache, applyCache)
}
