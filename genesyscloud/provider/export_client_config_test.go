package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v191/platformclientv2"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextWithExportClientConfig(t *testing.T) {
	t.Parallel()

	cfg := &platformclientv2.Configuration{}
	ctx := ContextWithExportClientConfig(context.Background(), cfg)

	got, ok := ExportClientConfigFromContext(ctx)
	require.True(t, ok)
	assert.Same(t, cfg, got)
}

func TestExportClientConfigFromContext_nilAndMissing(t *testing.T) {
	t.Parallel()

	_, ok := ExportClientConfigFromContext(nil)
	assert.False(t, ok)

	_, ok = ExportClientConfigFromContext(context.Background())
	assert.False(t, ok)

	ctx := ContextWithExportClientConfig(context.Background(), nil)
	_, ok = ExportClientConfigFromContext(ctx)
	assert.False(t, ok)
}

func TestResolveClientConfigForContext_prefersExportContext(t *testing.T) {
	t.Parallel()

	cfg := &platformclientv2.Configuration{BasePath: "export-context-path"}
	ctx := ContextWithExportClientConfig(context.Background(), cfg)

	got, release, diags := resolveClientConfigForContext(ctx)
	require.Nil(t, diags)
	require.NotNil(t, got)
	assert.Same(t, cfg, got)
	release()
}

func TestGetAllWithPooledClient_usesExportContextConfig(t *testing.T) {
	t.Parallel()

	exportCfg := &platformclientv2.Configuration{BasePath: "export-only-path"}
	ctx := ContextWithExportClientConfig(context.Background(), exportCfg)

	var seen *platformclientv2.Configuration
	wrapped := GetAllWithPooledClient(func(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
		seen = clientConfig
		return make(resourceExporter.ResourceIDMetaMap), nil
	})

	_, diags := wrapped(ctx)
	require.Nil(t, diags)
	assert.Same(t, exportCfg, seen)
}

func TestGetAllWithPooledClientCustom_usesExportContextConfig(t *testing.T) {
	t.Parallel()

	exportCfg := &platformclientv2.Configuration{BasePath: "export-custom-path"}
	ctx := ContextWithExportClientConfig(context.Background(), exportCfg)

	var seen *platformclientv2.Configuration
	wrapped := GetAllWithPooledClientCustom(func(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, []string, diag.Diagnostics) {
		seen = clientConfig
		return make(resourceExporter.ResourceIDMetaMap), nil, nil, nil
	})

	_, _, _, diags := wrapped(ctx)
	require.Nil(t, diags)
	assert.Same(t, exportCfg, seen)
}
