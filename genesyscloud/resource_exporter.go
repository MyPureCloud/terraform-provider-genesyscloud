package genesyscloud

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// ResourceIDNameMap is a map of IDs to resource names
type ResourceIDNameMap map[string]string

// GetAllResourcesFunc is a method that returns all resource IDs
type GetAllResourcesFunc func(context.Context) (ResourceIDNameMap, diag.Diagnostics)

// RefAttrSettings contains behavior settings for references
type RefAttrSettings struct {

	// Referenced resource type
	RefType string

	// Values that may be set that should not be treated as IDs
	AltValues []string
}

// ResourceExporter is an interface to implement for resources that can be exported
type ResourceExporter struct {

	// Method to load all resource IDs for a given resource.
	// Returned map key should be the ID and the value should be a name to use for the resource.
	// Names will be sanitized with part of the ID appended, so it is not required that they be unique
	GetResourcesFunc GetAllResourcesFunc

	// A map of resource attributes to types that they reference
	// Attributes in nested objects can be defined with a '.' separator
	RefAttrs map[string]*RefAttrSettings

	// AllowZeroValues is a list of attributes that should allow zero values in the export.
	// By default zero values are removed from the config due to lack of "null" support in the plugin SDK
	AllowZeroValues []string

	// RemoveIfMissing is a map of attributes to a list of inner object attributes.
	// When all specified inner attributes are missing from an object, that object is removed
	RemoveIfMissing map[string][]string

	// Map of resource id->names. This is set after a call to loadSanitizedResourceMap
	SanitizedResourceMap ResourceIDNameMap

	// List of attributes to exclude from config. This is set by the export configuration.
	ExcludedAttributes []string
}

func (r *ResourceExporter) loadSanitizedResourceMap(ctx context.Context) diag.Diagnostics {
	result, err := r.GetResourcesFunc(ctx)
	if err != nil {
		return err
	}
	r.SanitizedResourceMap = result
	sanitizeResourceNames(r.SanitizedResourceMap)
	return nil
}

func (r *ResourceExporter) getRefAttrSettings(attribute string) *RefAttrSettings {
	if r.RefAttrs == nil {
		return nil
	}
	return r.RefAttrs[attribute]
}

func (r *ResourceExporter) allowZeroValues(attribute string) bool {
	return stringInSlice(attribute, r.AllowZeroValues)
}

func (r *ResourceExporter) addExcludedAttribute(attribute string) {
	r.ExcludedAttributes = append(r.ExcludedAttributes, attribute)
}

func (r *ResourceExporter) isAttributeExcluded(attribute string) bool {
	for _, excluded := range r.ExcludedAttributes {
		// Excluded if attributes match, or the specified attribute is nested in the excluded attribute
		if excluded == attribute || strings.HasPrefix(attribute, excluded+".") {
			return true
		}
	}
	return false
}

func (r *ResourceExporter) removeIfMissing(attribute string, config map[string]interface{}) bool {
	if attrs, ok := r.RemoveIfMissing[attribute]; ok {
		// Check if all required inner attributes are missing
		missingAll := true
		for _, attr := range attrs {
			if val, foundInner := config[attr]; foundInner && val != nil {
				missingAll = false
				break
			}
		}
		return missingAll
	}
	return false
}

func getResourceExporters(filter []string) map[string]*ResourceExporter {
	exporters := map[string]*ResourceExporter{
		// Add new resources that can be exported here
		"genesyscloud_auth_division":       authDivisionExporter(),
		"genesyscloud_auth_role":           authRoleExporter(),
		"genesyscloud_group":               groupExporter(),
		"genesyscloud_group_roles":         groupRolesExporter(),
		"genesyscloud_idp_adfs":            idpAdfsExporter(),
		"genesyscloud_idp_generic":         idpGenericExporter(),
		"genesyscloud_idp_gsuite":          idpGsuiteExporter(),
		"genesyscloud_idp_okta":            idpOktaExporter(),
		"genesyscloud_idp_onelogin":        idpOneloginExporter(),
		"genesyscloud_idp_ping":            idpPingExporter(),
		"genesyscloud_idp_salesforce":      idpSalesforceExporter(),
		"genesyscloud_location":            locationExporter(),
		"genesyscloud_routing_language":    routingLanguageExporter(),
		"genesyscloud_routing_queue":       routingQueueExporter(),
		"genesyscloud_routing_skill":       routingSkillExporter(),
		"genesyscloud_routing_utilization": routingUtilizationExporter(),
		"genesyscloud_user":                userExporter(),
		"genesyscloud_user_roles":          userRolesExporter(),
	}

	// Include all if no filters
	if len(filter) > 0 {
		for resType := range exporters {
			if !stringInSlice(resType, filter) {
				delete(exporters, resType)
			}
		}
	}
	return exporters
}

func getAvailableExporterTypes() []string {
	exporters := getResourceExporters(nil)
	types := make([]string, len(exporters))
	i := 0
	for k := range exporters {
		types[i] = k
		i++
	}
	return types
}

func escapeRune(s string) string {
	return fmt.Sprintf("%02X", s)
}

// Resource names must only contain alphanumeric chars, underscores, or dashes
var unsafeNameChars = regexp.MustCompile(`[^0-9A-Za-z_-]`)

func sanitizeResourceNames(idNamesMap ResourceIDNameMap) {
	for id, name := range idNamesMap {
		name = unsafeNameChars.ReplaceAllStringFunc(name, escapeRune)
		// Append part of ID to ensure uniqueness for similar names
		if len(id) > 6 {
			name = name + "_" + id[:6]
		} else {
			name = name + "_" + id
		}
		idNamesMap[id] = name
	}
}
