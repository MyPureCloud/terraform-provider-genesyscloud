// Package routing_language contains temporary export utilities for Plugin Framework routing language resource.
//
// IMPORTANT: This file contains migration scaffolding that converts SDK types to flat
// attribute maps for the legacy exporter's dependency resolution logic.
//
// TODO: Remove this entire file once all resources are migrated to Plugin Framework
// and the exporter is updated to work natively with Framework types (Phase 2).
// This is Phase 1 temporary code - resource-specific implementation.
//
// File: genesyscloud/routing_language/resource_genesyscloud_routing_language_export_utils.go

package routing_language

import (
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

// buildLanguageAttributes creates a flat attribute map from SDK language object for export.
// This function converts the SDK language object to a flat map matching SDKv2 InstanceState format.
//
// Parameters:
//   - language: Language object from API
//
// Returns:
//   - map[string]string: Flat attribute map with all language attributes
//
// Attribute Map Format (matching SDKv2 InstanceState):
//   - "id" = language ID
//   - "name" = language name
//
// Note: Routing language is a simple resource with only id and name attributes.
// No nested attributes or dependency references.
func buildLanguageAttributes(language *platformclientv2.Language) map[string]string {
	attributes := make(map[string]string)

	// Basic attributes
	if language.Id != nil {
		attributes["id"] = *language.Id
	}
	if language.Name != nil {
		attributes["name"] = *language.Name
	}

	return attributes
}
