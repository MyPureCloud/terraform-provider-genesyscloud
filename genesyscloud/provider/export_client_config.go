package provider

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
)

type exportClientConfigKey struct{}

// ContextWithExportClientConfig attaches an SDK client configuration to ctx for MRMO and
// other standalone export callers. Export paths resolve config from ctx at call time instead
// of process-wide MRMO global state.
func ContextWithExportClientConfig(ctx context.Context, clientConfig *platformclientv2.Configuration) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if clientConfig == nil {
		return ctx
	}
	return context.WithValue(ctx, exportClientConfigKey{}, clientConfig)
}

// ExportClientConfigFromContext returns the SDK client configuration attached by
// ContextWithExportClientConfig.
func ExportClientConfigFromContext(ctx context.Context) (*platformclientv2.Configuration, bool) {
	if ctx == nil {
		return nil, false
	}
	clientConfig, ok := ctx.Value(exportClientConfigKey{}).(*platformclientv2.Configuration)
	return clientConfig, ok && clientConfig != nil
}
