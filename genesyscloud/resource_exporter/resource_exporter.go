package resource_exporter

import (
	"context"
	"regexp"
	"strings"
	"sync"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"

	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

var resourceExporters map[string]*ResourceExporter
var resourceExporterMapMutex = sync.RWMutex{}

type ResourceMeta struct {
	// Name of the resource to be used in exports
	Name string

	// Prefix to add to the ID when reading state
	IdPrefix string
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
	State        *terraform.InstanceState
	Name         string
	Type         string
	CtyType      cty.Type
	ResourceType string
}

// RefAttrCustomResolver allows the definition of a custom resolver for an exporter.
type RefAttrCustomResolver struct {
	ResolverFunc            func(map[string]interface{}, map[string]*ResourceExporter, string) error
	ResolveToDataSourceFunc func(map[string]interface{}, any, *platformclientv2.Configuration) (string, string, map[string]interface{}, bool)
}

// CustomFlowResolver allows the definition of a custom resolver for an exporter.
type CustomFlowResolver struct {
	ResolverFunc func(map[string]interface{}, string) error
}

type CustomFileWriterSettings struct {
	// Custom function for dumping data/media stored in an object in a sub directory along
	// with the exported config. For example: prompt audio files, csv data, jps/pngs
	RetrieveAndWriteFilesFunc func(string, string, string, map[string]interface{}, interface{}) error

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

type DependencyResource struct {
	DependsMap        map[string][]string
	CyclicDependsList []string
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

	// Map of resource id->names. This is set after a call to loadSanitizedResourceMap
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

	CustomFlowResolver map[string]*CustomFlowResolver

	//This a placeholder filter out specific resources from a filter.
	FilterResource func(ResourceIDMetaMap, string, []string) ResourceIDMetaMap
	// Attributes that are mentioned with custom exports like e164 numbers,rrule  should be ensured to export in the correct format (remove hyphens, whitespace, etc.)
	CustomValidateExports map[string][]string
	mutex                 sync.RWMutex
}

func (r *ResourceExporter) LoadSanitizedResourceMap(ctx context.Context, name string, filter []string) diag.Diagnostics {
	result, err := r.GetResourcesFunc(ctx)
	if err != nil {
		return err
	}

	if r.FilterResource != nil {
		result = r.FilterResource(result, name, filter)
	}

	// Lock the Resource Map as it is accessed by goroutines
	r.mutex.Lock()
	r.SanitizedResourceMap = result
	r.mutex.Unlock()

	sanitizer := NewSanitizerProvider()
	sanitizer.S.Sanitize(r.SanitizedResourceMap)

	return nil
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

// Resource names must only contain alphanumeric chars, underscores, or dashes
// https://www.terraform.io/docs/language/syntax/configuration.html#identifiers
var unsafeNameChars = regexp.MustCompile(`[^0-9A-Za-z_-]`)

// Resource names must start with a letter or underscore
// https://www.terraform.io/docs/language/syntax/configuration.html#identifiers
var unsafeNameStartingChars = regexp.MustCompile(`[^A-Za-z_]`)

func RegisterExporter(exporterName string, resourceExporter *ResourceExporter) {
	resourceExporterMapMutex.Lock()
	defer resourceExporterMapMutex.Unlock()
	resourceExporters[exporterName] = resourceExporter
}

func SetRegisterExporter(resources map[string]*ResourceExporter) {
	resourceExporterMapMutex.Lock()
	defer resourceExporterMapMutex.Unlock()
	resourceExporters = resources
}

var (
	ExportAsData          []string
	dsMutex               sync.Mutex
	resourceNameSanitizer = NewSanitizerProvider()
)

// The AddDataSourceItems function adds resources to the ExportAsData []string and are formatted correctly
// The ExportAsData will be checked in the genesyscloud_resource_exporter to determine resources to be exported as data source
func AddDataSourceItems(resourceName, itemName string) {
	exportName := resourceName + "::" + resourceNameSanitizer.S.SanitizeResourceName(itemName)
	addDataSourceItemstoExport(exportName)
}

func addDataSourceItemstoExport(name string) {
	dsMutex.Lock()
	defer dsMutex.Unlock()
	ExportAsData = append(ExportAsData, name)
}
