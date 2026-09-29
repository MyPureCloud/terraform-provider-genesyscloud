package provider

import (
	"context"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
)

type exportClientConfigKey struct{}

type customRetryTimeoutKey struct{}

// ContextWithCustomRetryTimeout attaches a per-call custom retry timeout override to
// ctx, for MRMO and other standalone callers that invoke resource Create/Update/
// Delete/Read context functions directly and run many of them concurrently in one
// process. GetCustomRetryTimeout otherwise falls back to the process-wide
// GENESYSCLOUD_CUSTOM_RETRY_TIMEOUT env var, which concurrent callers cannot safely
// share — the same problem ContextWithExportClientConfig solves for client config.
func ContextWithCustomRetryTimeout(ctx context.Context, timeout time.Duration) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, customRetryTimeoutKey{}, timeout)
}

// CustomRetryTimeoutFromContext returns the retry timeout attached by
// ContextWithCustomRetryTimeout.
func CustomRetryTimeoutFromContext(ctx context.Context) (time.Duration, bool) {
	if ctx == nil {
		return 0, false
	}
	timeout, ok := ctx.Value(customRetryTimeoutKey{}).(time.Duration)
	return timeout, ok
}

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
