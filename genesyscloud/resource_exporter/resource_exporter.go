package resource_exporter

import (
	"context"
	"regexp"
	"strings"
	"sync"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"

	lists "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

var resourceExporters map[string]*ResourceExporter
var resourceExporterMapMutex = sync.RWMutex{}

type ResourceMeta struct {
	// BlockLabel of the resource to be used in exports
	BlockLabel string

	// Prefix to add to the ID when reading state
	IdPrefix string

	// BlockHash represents a unique identifier generated from the resource's distinguishing attributes,
	// explicitly excluding IDs and calculated fields to enable cross-org resource correlation.
	// Important:
	//   * Use util.QuickHashFields() to generate this hash
	//   * Only include fields that uniquely identify the resource WITHOUT using its ID
	//   * Only include fields that would match if the resource exists in another org
	//   * DO NOT include fields that are likely to be updated or modified as the resource evolves
	//   * Do NOT include fields that are calculated (i.e., createdDate)
	//   * ID fields and calculated field must be excluded from the hash calculation because:
	//       - IDs and calculated values prevent correlation of resources across different exports and
	//         make it impossible to compare equivalent resources between orgs
	BlockHash string

	// Represents the unsanitized version of the BlockLabel
	OriginalLabel string

	// ExportAttributes contains the complete flat attribute map for Plugin Framework resources.
	// This field enables dependency resolution for PF resources by providing all attribute values
	// (including dependency references like division_id, skill_id, etc.) to the exporter.
	//
	// Format: Flat map matching Terraform InstanceState format with dot notation and indices
	//   Example: "addresses.0.phone_numbers.0.extension_pool_id" = "pool-guid-123"
	//
	// Usage:
	//   - Populated by GetAll functions (e.g., GetAllUsersSDK) during resource discovery
	//   - Consumed by exporter to create InstanceState for dependency resolution
	//   - Optional field (nil for SDKv2 resources and PF resources not yet migrated)
	//
	// Backward Compatibility:
	//   - If nil or empty, exporter falls back to stub behavior (id + name only)
	//   - No breaking changes to existing code
	//
	// TODO: This is a temporary migration helper for Phase 1 (resource-specific implementations).
	// Phase 2 will introduce a generic Framework-to-InstanceState converter that works for all
	// PF resources automatically. Once Phase 2 is complete and all resources are migrated,
	// this field may be deprecated or repurposed for the generic converter's output.
	ExportAttributes map[string]string

	// LazyFetchAttributes is an optional callback for fetching export attributes on-demand.
	// This callback is invoked AFTER filtering, only for resources that will be exported.
	// This is more efficient than ExportAttributes for filtered exports.
	//
	// The callback receives a context and should return:
	//   - map[string]string: Complete flat attribute map for dependency resolution
	//   - error: Error if fetch fails (exporter will fall back to basic attributes)
	//
	// Usage pattern:
	//   1. GetAll function sets this callback (captures resource ID in closure)
	//   2. Exporter filters resources
	//   3. Exporter calls callback only for filtered resources
	//   4. Callback fetches full details and builds attribute map
	//
	// Performance benefit:
	//   - Filtered export (10 users): 31 API calls vs 6,448 (99.5% reduction)
	//   - Time: 3 seconds vs 5+ minutes timeout
	//
	// TODO: Phase 2 - Remove when generic Framework-to-InstanceState converter is implemented
	LazyFetchAttributes func(ctx context.Context) (map[string]string, error)
}

// ResourceIDMetaMap is a map of IDs to ResourceMeta
type ResourceIDMetaMap map[string]*ResourceMeta

type GetAllCustomResourcesFunc func(context.Context) (ResourceIDMetaMap, *DependencyResource, diag.Diagnostics)

// GetAllResourcesFunc is a method that returns all resource IDs
type GetAllResourcesFunc func(context.Context) (ResourceIDMetaMap, diag.Diagnostics)

// RefAttrSettings contains behavior settings for references
type RefAttrSettings struct {

	// Referenced resource type
	RefType string

	// Values that may be set that should not be treated as IDs
	AltValues []string
}

type ResourceInfo struct {
	State         *terraform.InstanceState
	BlockLabel    string
	OriginalLabel string
	Type          string
	CtyType       cty.Type
	BlockType     string
	IsFramework   bool // TODO: Phase 2 - Remove after all resources migrated to Framework
}

// DataSourceResolver allows the definition of a custom resolver for an exporter.
type DataSourceResolver struct {
	ResolverFunc func(map[string]interface{}, string) error
}

// RefAttrCustomResolver allows the definition of a custom resolver for an exporter.
type RefAttrCustomResolver struct {
	ResolverFunc            func(configMap map[string]interface{}, exporters map[string]*ResourceExporter, resourceLabel string) error
	ResolveToDataSourceFunc func(configMap map[string]interface{}, originalValue any, sdkConfig *platformclientv2.Configuration) (string, string, map[string]interface{}, bool)
}

type CustomFileWriterSettings struct {
	// Custom function for dumping data/media stored in an object in a sub directory along
	// with the exported config. For example: prompt audio files, csv data, jps/pngs
	RetrieveAndWriteFilesFunc func(string, string, string, map[string]interface{}, interface{}, ResourceInfo) error

	// Sub directory within export folder in which to write files retrieved by RetrieveAndWriteFilesFunc
	// For example, the user_prompt resource defines SubDirectory as "audio", so the prompt audio files will
	// be written to genesyscloud_tf_export.directory/audio/
	// The logic for retrieving and writing data to this dir should be defined in RetrieveAndWriteFilesFunc
	SubDirectory string
}

type JsonEncodeRefAttr struct {
	// The outer key
	Attr string

	// The RefAttr nested inside the json data
	NestedAttr string
}

type DataAttr struct {
	Attr string
}

type ResourceAttr struct {
	Attr string
}

type DependencyResource struct {
	DependsMap        map[string][]string
	CyclicDependsList []string
}

// ResourceExporter is an interface to implement for resources that can be exported
type ResourceExporter struct {

	// Method to load all resource IDs for a given resource.
	// Returned map key should be the ID and the value should be a label to use for the resource.
	// Label will be sanitized with part of the ID appended, so it is not required that they be unique
	GetResourcesFunc GetAllResourcesFunc

	// A map of resource attributes to types that they reference
	// Attributes in nested objects can be defined with a '.' separator
	RefAttrs map[string]*RefAttrSettings

	// AllowZeroValues is a list of attributes that should allow zero values in the export.
	// By default zero values are removed from the config due to lack of "null" support in the plugin SDK
	AllowZeroValues []string

	// AllowZeroValuesInMap is a list of attributes that are maps. Adding a map attribute to this list indicates to
	// the exporter that the values within said map should not be cleaned up if they are zero values
	AllowZeroValuesInMap []string

	// AllowEmptyArrays is a list of List attributes that should allow empty arrays in export.
	// By default, empty arrays are removed but some array attributes may be required in the schema
	// or depending on the API behavior better presented explicitly in the API as empty arrays.
	// If the state has this as null or empty array, then the attribute will be returned as an empty array.
	AllowEmptyArrays []string

	// Some of our dependencies can not be exported properly because they have interdependencies between attributes.  You can
	// define a map of custom attribute resolvers with an exporter.  See resource_genesyscloud_routing_queue for an example of how to define this.
	// NOTE: CustomAttributeResolvers should be the exception and not the norm so use them when you have to do logic that will help you
	// resolve to the write reference
	CustomAttributeResolver map[string]*RefAttrCustomResolver

	// RemoveIfMissing is a map of attributes to a list of inner object attributes.
	// When all specified inner attributes are missing from an object, that object is removed
	RemoveIfMissing map[string][]string

	// Map of resource id->labels. This is set after a call to loadSanitizedResourceMap
	SanitizedResourceMap ResourceIDMetaMap
	// List of attributes to exclude from config. This is set by the export configuration.
	ExcludedAttributes []string

	// Map of attributes that cannot be resolved. E.g. edge Ids which are locked to an org or properties that cannot be retrieved from the API
	UnResolvableAttributes map[string]*schema.Schema

	// List of attributes which can and should be exported in a jsonencode object rather than as a long escaped string of JSON data.
	JsonEncodeAttributes []string

	// Attributes that are jsonencode objects, and that contain nested RefAttrs
	EncodedRefAttrs map[*JsonEncodeRefAttr]*RefAttrSettings

	CustomFileWriter CustomFileWriterSettings

	ExportAsDataFunc func(context.Context, *platformclientv2.Configuration, map[string]string) (bool, error)

	// used when the names of the attributes in Datasource and Resource schema does not match,
	//gives you the flexibility to match them and use when a resource need to be replaced as datasource.
	DataSourceResolver map[*DataAttr]*ResourceAttr

	//This a placeholder filter out specific resources from a filter.
	FilterResource func(resourceIdMetaMap ResourceIDMetaMap, resourceType string, filter []string) ResourceIDMetaMap
	// Attributes that are mentioned with custom exports like e164 numbers,rrule  should be ensured to export in the correct format (remove hyphens, whitespace, etc.)
	CustomValidateExports map[string][]string
	mutex                 sync.RWMutex
}

func (r *ResourceExporter) LoadSanitizedResourceMap(ctx context.Context, resourceType string, filter []string) diag.Diagnostics {
	result, err := r.GetResourcesFunc(ctx)
	if err != nil {
		return err
	}

	if r.FilterResource != nil {
		result = r.FilterResource(result, resourceType, filter)
	}

	// Lock the Resource Map as it is accessed by goroutines
	r.mutex.Lock()
	r.SanitizedResourceMap = result
	r.mutex.Unlock()

	sanitizer := NewSanitizerProvider()
	sanitizer.S.Sanitize(r.SanitizedResourceMap)

	return nil
}

// Thread-safe methods for accessing SanitizedResourceMap
func (r *ResourceExporter) GetSanitizedResourceMap() ResourceIDMetaMap {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.SanitizedResourceMap
}

func (r *ResourceExporter) SetSanitizedResourceMap(resourceMap ResourceIDMetaMap) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.SanitizedResourceMap = resourceMap
}

func (r *ResourceExporter) RemoveFromSanitizedResourceMap(id string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.SanitizedResourceMap, id)
}

func (r *ResourceExporter) GetSanitizedResourceMapSize() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.SanitizedResourceMap)
}

func (r *ResourceExporter) GetRefAttrSettings(attribute string) *RefAttrSettings {
	if r.RefAttrs == nil {
		return nil
	}
	return r.RefAttrs[attribute]
}

func (r *ResourceExporter) GetNestedRefAttrSettings(attribute string) *RefAttrSettings {
	for key, val := range r.EncodedRefAttrs {
		if key.NestedAttr == attribute {
			return val
		}
	}
	return nil
}

func (r *ResourceExporter) ContainsNestedRefAttrs(attribute string) ([]string, bool) {
	var nestedAttributes []string
	for key := range r.EncodedRefAttrs {
		if key.Attr == attribute {
			nestedAttributes = append(nestedAttributes, key.NestedAttr)
		}
	}
	return nestedAttributes, len(nestedAttributes) > 0
}

func (r *ResourceExporter) DataResolver(instanceState *terraform.InstanceState, attr string) (string, string) {
	for key, val := range r.DataSourceResolver {
		if key.Attr == attr {
			value := r.fetchFromInstanceState(instanceState, val.Attr)
			if value != "" {
				return attr, value
			}
		}
	}
	if value, ok := instanceState.Attributes[attr]; ok {
		return attr, value
	}
	return "", ""
}

func (r *ResourceExporter) fetchFromInstanceState(instanceState *terraform.InstanceState, pattern string) string {
	re := regexp.MustCompile(pattern)
	for key, val := range instanceState.Attributes {
		if re.MatchString(key) {
			return val
		}
	}
	return ""
}

func (r *ResourceExporter) AllowForZeroValues(attribute string) bool {
	return lists.ItemInSlice(attribute, r.AllowZeroValues)
}

func (r *ResourceExporter) AllowForZeroValuesInMap(attribute string) bool {
	return lists.ItemInSlice(attribute, r.AllowZeroValuesInMap)
}

func (r *ResourceExporter) AllowForEmptyArrays(attribute string) bool {
	return lists.ItemInSlice(attribute, r.AllowEmptyArrays)
}

func (r *ResourceExporter) IsJsonEncodable(attribute string) bool {
	return lists.ItemInSlice(attribute, r.JsonEncodeAttributes)
}

func (r *ResourceExporter) IsAttributeE164(attribute string) bool {
	values, exists := r.CustomValidateExports["E164"]
	if !exists {
		return false
	}
	return lists.ItemInSlice(attribute, values)
}

func (r *ResourceExporter) IsAttributeRrule(attribute string) bool {
	values, exists := r.CustomValidateExports["rrule"]
	if !exists {
		return false
	}
	return lists.ItemInSlice(attribute, values)
}

func (r *ResourceExporter) AddExcludedAttribute(attribute string) {
	r.ExcludedAttributes = append(r.ExcludedAttributes, attribute)
}

func (r *ResourceExporter) IsAttributeExcluded(attribute string) bool {
	for _, excluded := range r.ExcludedAttributes {
		// Excluded if attributes match, or the specified attribute is nested in the excluded attribute
		if excluded == attribute || strings.HasPrefix(attribute, excluded+".") {
			return true
		}
	}
	return false
}

func (r *ResourceExporter) RemoveFieldIfMissing(attribute string, config map[string]interface{}) bool {
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

func GetResourceExporters() map[string]*ResourceExporter {

	//Make a Copy of the Map
	exportCopy := make(map[string]*ResourceExporter, len(resourceExporters))

	for k, v := range resourceExporters {
		exportCopy[k] = v
	}
	return exportCopy
}

// terraform-provider-genesyscloud/genesyscloud/tfexporter
func GetAvailableExporterTypes() []string {
	exporters := GetResourceExporters()
	types := make([]string, len(exporters))
	i := 0
	for k := range exporters {
		types[i] = k
		i++
	}
	return types
}

func escapeRune(s string) string {
	// Always replace with an underscore for readability. The appended hash will help ensure uniqueness
	return "_"
}

// Resource labels must only contain alphanumeric chars, underscores, or dashes
// https://www.terraform.io/docs/language/syntax/configuration.html#identifiers
var unsafeLabelChars = regexp.MustCompile(`[^0-9A-Za-z_-]`)

// Resource labels must start with a letter or underscore
// https://www.terraform.io/docs/language/syntax/configuration.html#identifiers
var unsafeLabelStartingChars = regexp.MustCompile(`[^A-Za-z_]`)

func RegisterExporter(exporterLabel string, resourceExporter *ResourceExporter) {
	resourceExporterMapMutex.Lock()
	defer resourceExporterMapMutex.Unlock()
	resourceExporters[exporterLabel] = resourceExporter
}

func SetRegisterExporter(resources map[string]*ResourceExporter) {
	resourceExporterMapMutex.Lock()
	defer resourceExporterMapMutex.Unlock()
	resourceExporters = resources
}
