package util

import (
	"context"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SetResourceContext extracts resource metadata from ResourceData and adds it to the context.
// This function extracts both resource ID and name as separate values.
// If either value is unavailable, it will be set to "unavailable".
// This context can then be used by SDK debug hooks to include resource information in logs.
func SetResourceContext(ctx context.Context, d *schema.ResourceData, resourceType string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	resourceId, resourceName := extractResourceIdAndName(d)

	return provider.WithResourceContext(ctx, resourceType, resourceId, resourceName)
}

// extractResourceIdAndName extracts both ID and name from ResourceData.
// Returns "unavailable" for any missing values.
func extractResourceIdAndName(d *schema.ResourceData) (resourceId, resourceName string) {
	// Get resource ID
	if d != nil && d.Id() != "" {
		resourceId = d.Id()
	} else {
		resourceId = "unavailable"
	}

	// Get resource name
	if d != nil {
		if name, ok := d.GetOk("name"); ok {
			if nameStr, ok := name.(string); ok && nameStr != "" {
				resourceName = nameStr
			} else {
				resourceName = "unavailable"
			}
		} else {
			resourceName = "unavailable"
		}
	} else {
		resourceName = "unavailable"
	}

	return resourceId, resourceName
}
