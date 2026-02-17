// Package provider implements the Genesys Cloud Terraform provider with support for both
// SDKv2 and Framework provider architectures through muxing.
//
// This file contains the muxing logic that combines:
//   - Legacy SDKv2 resources and data sources (upgraded to Protocol v6)
//   - Modern Framework resources and data sources (native Protocol v6)
//
// The muxing allows for gradual migration from SDKv2 to Framework while maintaining
// backward compatibility and enabling new Framework features.
package provider

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// NewMuxedProvider builds a Protocol v6 factory function suitable for tf6server.Serve.
// It combines:
//   - SDKv2 provider (native v5)  → upgraded to v6 via tf5to6server
//   - Framework provider (native v6)
//
// Parameters:
//   - version: Provider version string
//   - providerResources: Map of SDKv2 resource names to resource implementations
//   - providerDataSources: Map of SDKv2 data source names to data source implementations
//   - frameworkResources: Map of Framework resource names to resource factory functions
//   - frameworkDataSources: Map of Framework data source names to data source factory functions
//
// Returns:
//   - A factory function that creates a muxed tfprotov6.ProviderServer
//   - An error if the muxing fails
//
// The function automatically detects if Framework resources/data sources are present.
// If none are found, it returns an SDKv2-only provider for optimal performance.
func NewMuxedProvider(
	version string,
	providerResources map[string]*schema.Resource,
	providerDataSources map[string]*schema.Resource,
	frameworkResources map[string]func() resource.Resource,
	frameworkDataSources map[string]func() datasource.DataSource,
) func() (func() tfprotov6.ProviderServer, error) {
	return func() (func() tfprotov6.ProviderServer, error) {
		ctx := context.Background()

		// Create SDKv2 provider and upgrade to v6
		sdkv2Provider := NewSDKv2Provider(version, providerResources, providerDataSources)()
		upgradedV6, err := tf5to6server.UpgradeServer(ctx, sdkv2Provider.GRPCProvider)
		if err != nil {
			log.Printf("[ERROR] Failed to upgrade SDKv2 provider to v6: %v", err)
			return nil, err
		}

		// Check if we have any Framework resources/datasources to mux
		hasFrameworkResources := len(frameworkResources) > 0
		hasFrameworkDataSources := len(frameworkDataSources) > 0

		if !hasFrameworkResources && !hasFrameworkDataSources {
			log.Printf("[INFO] No Framework resources/datasources found, using SDKv2 provider only")
			return func() tfprotov6.ProviderServer { return upgradedV6 }, nil
		}

		// Create Framework provider factory
		frameworkProviderFactory := NewFrameworkProvider(version, frameworkResources, frameworkDataSources)

		// Create muxed server
		log.Printf("[INFO] Creating muxed provider with %d Framework resources and %d Framework datasources",
			len(frameworkResources), len(frameworkDataSources))

		muxServer, err := tf6muxserver.NewMuxServer(ctx,
			func() tfprotov6.ProviderServer { return upgradedV6 },
			func() tfprotov6.ProviderServer {
				return providerserver.NewProtocol6(frameworkProviderFactory())()
			},
		)
		if err != nil {
			log.Printf("[ERROR] Failed to create mux server: %v", err)
			return nil, err
		}

		return func() tfprotov6.ProviderServer { return muxServer }, nil
	}
}
