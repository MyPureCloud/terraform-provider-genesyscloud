package util

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

// ExporterByIdsWithName builds a by-ID exporter function for resources that are fetched with a plain GET-by-ID and
// are labelled by their name, which covers most exporters. getter returns the proxy's get-by-ID call; name picks the
// label from the fetched object. Objects that no longer exist (404) are skipped.
func ExporterByIdsWithName[T any](
	resourceType string,
	getter func(*platformclientv2.Configuration) func(ctx context.Context, id string) (*T, *platformclientv2.APIResponse, error),
	name func(*T) *string,
) resourceExporter.GetResourcesByIdsFunc {
	return provider.GetByIdsWithPooledClient(func(ctx context.Context, config *platformclientv2.Configuration, ids []string) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
		get := getter(config)
		return resourceExporter.GetResourcesByIds(ctx, ids, func(ctx context.Context, id string) (*resourceExporter.ResourceMeta, diag.Diagnostics) {
			item, resp, err := get(ctx, id)
			if err != nil {
				if IsStatus404(resp) {
					return nil, nil
				}
				return nil, BuildAPIDiagnosticError(resourceType, fmt.Sprintf("Failed to get %s %s error: %s", resourceType, id, err), resp)
			}
			if item == nil || name(item) == nil {
				return nil, nil
			}
			return &resourceExporter.ResourceMeta{BlockLabel: *name(item)}, nil
		})
	})
}
