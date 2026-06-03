package exporter

import (
	"context"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v188/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/mrmo"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	providerRegistrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider_registrar"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParallelExportByTypeContextIsolation(t *testing.T) {
	t.Parallel()

	const workers = 6
	type configResult struct {
		path   string
		errMsg string
	}

	results := make(chan configResult, workers)
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(idx int) {
			defer wg.Done()
			cfg := &platformclientv2.Configuration{BasePath: string(rune('A' + idx))}
			ctx := provider.ContextWithExportClientConfig(context.Background(), cfg)

			var seen *platformclientv2.Configuration
			wrapped := provider.GetAllWithPooledClient(func(callCtx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
				seen = clientConfig
				got, ok := provider.ExportClientConfigFromContext(callCtx)
				if !ok {
					return nil, diag.Errorf("missing export config in context")
				}
				if got != clientConfig {
					return nil, diag.Errorf("context config mismatch")
				}
				return make(resourceExporter.ResourceIDMetaMap), nil
			})

			_, diags := wrapped(ctx)
			if diags.HasError() {
				results <- configResult{errMsg: diags[0].Summary}
				return
			}
			if seen == nil {
				results <- configResult{errMsg: "nil client config"}
				return
			}
			results <- configResult{path: seen.BasePath}
		}(i)
	}

	wg.Wait()
	close(results)

	seenPaths := make(map[string]struct{})
	for result := range results {
		require.Empty(t, result.errMsg, result.errMsg)
		seenPaths[result.path] = struct{}{}
	}
	assert.Len(t, seenPaths, workers)
}

func TestParallelClonedExportersDoNotMutateRegistrySingleton(t *testing.T) {
	t.Parallel()

	providerRegistrar.GetProviderResources()
	resourceType := "genesyscloud_outbound_attempt_limit"

	singletonBefore := providerRegistrar.GetResourceExporterByResourceType(resourceType)
	require.NotNil(t, singletonBefore)
	beforeMap := singletonBefore.GetSanitizedResourceMap()
	beforeFilter := singletonBefore.FilterResource

	const workers = 10
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(idx int) {
			defer wg.Done()
			clone := providerRegistrar.GetClonedResourceExporterByResourceType(resourceType)
			require.NotNil(t, clone)
			assert.NotSame(t, singletonBefore, clone)

			clone.FilterResource = func(resourceIdMetaMap resourceExporter.ResourceIDMetaMap, resourceType string, filter []string) resourceExporter.ResourceIDMetaMap {
				return resourceIdMetaMap
			}
			clone.SetSanitizedResourceMap(resourceExporter.ResourceIDMetaMap{
				"id": {BlockLabel: string(rune('a' + idx))},
			})
		}(i)
	}
	wg.Wait()

	assert.Equal(t, beforeMap, singletonBefore.GetSanitizedResourceMap())
	if beforeFilter == nil {
		assert.Nil(t, singletonBefore.FilterResource)
	} else {
		assert.NotNil(t, singletonBefore.FilterResource)
	}
}

func TestMRMOActivateNotRequiredWhenContextConfigPresent(t *testing.T) {
	mrmo.Reset()
	t.Setenv(mrmo.MRMO_CXASCODE_INTEGRATION_ENABLED, "")

	cfg := &platformclientv2.Configuration{BasePath: "ctx-only"}
	ctx := provider.ContextWithExportClientConfig(context.Background(), cfg)

	var seen *platformclientv2.Configuration
	wrapped := provider.GetAllWithPooledClient(func(callCtx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
		seen = clientConfig
		return make(resourceExporter.ResourceIDMetaMap), nil
	})

	_, diags := wrapped(ctx)
	require.Nil(t, diags)
	assert.Same(t, cfg, seen)
}
