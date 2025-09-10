package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// NewMuxedProvider builds a Protocol v6 factory function suitable for tf6server.Serve.
// It combines:
//   - SDKv2 provider (native v5)  → upgraded to v6 via tf5to6server
//   - Framework provider (native v6)
func NewMuxedProvider(
	version string,
	providerResources map[string]*schema.Resource,
	providerDataSources map[string]*schema.Resource,
) func() (func() tfprotov6.ProviderServer, error) {
	return func() (func() tfprotov6.ProviderServer, error) {
		ctx := context.Background()

		// --- SDKv2 side (native v5) → wrap to Protocol v6 via method value ---
		sdkv2Provider := NewSDKv2Provider(version, providerResources, providerDataSources)()
		// IMPORTANT: pass the method value, not ServeOpts
		upgradedV6, err := tf5to6server.UpgradeServer(ctx, sdkv2Provider.GRPCProvider)
		if err != nil {
			return nil, err
		}

		// For now, we'll only use the SDKv2 provider (upgraded to v6)
		// The Framework provider will be added later when we start migrating resources
		// This avoids schema mismatch issues during the initial muxing setup

		// Return just the upgraded SDKv2 provider for now
		return func() tfprotov6.ProviderServer { return upgradedV6 }, nil
	}
}
