// Package routing_wrapupcode contains temporary export utilities for Plugin Framework routing wrapupcode resource.
//
// IMPORTANT: This file contains migration scaffolding that converts SDK types to flat
// attribute maps for the legacy exporter's dependency resolution logic.
//
// TODO: Remove this entire file once all resources are migrated to Plugin Framework
// and the exporter is updated to work natively with Framework types (Phase 2).
// This is Phase 1 temporary code - resource-specific implementation.
//
// File: genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_export_utils.go

package routing_wrapupcode

import (
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

// buildWrapupcodeAttributes creates a flat attribute map from SDK wrapupcode object for export.
// This function converts the SDK wrapupcode object to a flat map matching SDKv2 InstanceState format.
//
// Parameters:
//   - wrapupcode: Wrapupcode object from API
//
// Returns:
//   - map[string]string: Flat attribute map with all wrapupcode attributes
//
// Attribute Map Format (matching SDKv2 InstanceState):
//   - "id" = wrapupcode ID
//   - "name" = wrapupcode name
//   - "division_id" = division ID (dependency reference)
//   - "description" = wrapupcode description
//
// Note: Unlike user resource, wrapupcode is simple with no nested attributes or additional API calls.
func buildWrapupcodeAttributes(wrapupcode *platformclientv2.Wrapupcode) map[string]string {
	attributes := make(map[string]string)

	// Basic attributes
	if wrapupcode.Id != nil {
		attributes["id"] = *wrapupcode.Id
	}
	if wrapupcode.Name != nil {
		attributes["name"] = *wrapupcode.Name
	}
	if wrapupcode.Description != nil {
		attributes["description"] = *wrapupcode.Description
	}

	// ⭐ CRITICAL: Dependency reference (used by exporter for dependency resolution)
	if wrapupcode.Division != nil && wrapupcode.Division.Id != nil {
		attributes["division_id"] = *wrapupcode.Division.Id
	}

	return attributes
}
