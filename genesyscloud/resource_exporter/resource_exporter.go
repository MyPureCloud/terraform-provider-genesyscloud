package resource_exporter

import (
	"context"
	"fmt"
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists" 
)

var resourceExporters map[string]*ResourceExporter
var resourceExporterMapMutex = sync.RWMutex{}

func init() {
	resourceExporters = make(map[string]*ResourceExporter)
}

func RegisterExporter(exporterName string, resourceExporter *ResourceExporter) {
	//resourceMapMutex.Lock()
	resourceExporters[exporterName] = resourceExporter
	//resourceMapMutex.Unlock()
}

func SetRegisterExporter(resources map[string]*ResourceExporter) {
	//resourceMapMutex.Lock()
	resourceExporters = resources
	//resourceMapMutex.Unlock()
}



type ResourceMeta struct {
	// Name of the resource to be used in exports
	Name string

	// Prefix to add to the ID when reading state
	IdPrefix string
}

// resource_exporter.ResourceIDMetaMap is a map of IDs to ResourceMeta
type ResourceIDMetaMap map[string]*ResourceMeta

// GetAllResourcesFunc is a method that returns all resource IDs
type GetAllResourcesFunc func(context.Context) (ResourceIDMetaMap, diag.Diagnostics)

// RefAttrSettings contains behavior settings for references
type RefAttrSettings struct {

	// Referenced resource type
	RefType string

	// Values that may be set that should not be treated as IDs
	AltValues []string
}

// Allows the definition of a custom resolver for an exporter.
type RefAttrCustomResolver struct {
	ResolverFunc func(map[string]interface{}, map[string]*ResourceExporter) error
}

// Allows the definition of a custom resolver for an exporter.
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
}

func (r *ResourceExporter) LoadSanitizedResourceMap(ctx context.Context, name string, filter []string) diag.Diagnostics {
	result, err := r.GetResourcesFunc(ctx)
	if err != nil {
		return err
	}

	if lists.SubStringInSlice(fmt.Sprintf("%v::", name), filter) {
		result = filterResources(result, name, filter)
	}

	r.SanitizedResourceMap = result
	sanitizeResourceNames(r.SanitizedResourceMap)
	return nil
}

func filterResources(result ResourceIDMetaMap, name string, filter []string) ResourceIDMetaMap {
	names := make([]string, 0)
	for _, f := range filter {
		n := fmt.Sprintf("%v::", name)
		if strings.Contains(f, n) {
			names = append(names, strings.Replace(f, n, "", 1))
		}
	}

	newResult := make(ResourceIDMetaMap)
	for _, name := range names {
		for k, v := range result {
			if v.Name == name {
				newResult[k] = v
			}
		}
	}
	return newResult
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
	for key, _ := range r.EncodedRefAttrs {
		if key.Attr == attribute {
			nestedAttributes = append(nestedAttributes, key.NestedAttr)
		}
	}
	return nestedAttributes, len(nestedAttributes) > 0
}

func (r *ResourceExporter) AllowForZeroValues(attribute string) bool {
	return lists.StringInSlice(attribute, r.AllowZeroValues)
}

func (r *ResourceExporter) IsJsonEncodable(attribute string) bool {
	return lists.StringInSlice(attribute, r.JsonEncodeAttributes)
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

func GetResourceExporters(filter []string) map[string]*ResourceExporter {

	exportCopy := make(map[string]*ResourceExporter, len(resourceExporters))

	for k, v := range resourceExporters {
		exportCopy[k] = v
	}
	

	// Include all if no filters
	if len(filter) > 0 {
		for resType := range exportCopy {
			if !lists.StringInSlice(resType, FormatFilter(filter)) {
				delete(exportCopy, resType)
			}
		}
	}
	return exportCopy
}

// Removes the ::resource_name from the resource_types list
func FormatFilter(filter []string) []string {
	newFilter := make([]string, 0)
	for _, str := range filter {
		newFilter = append(newFilter, strings.Split(str, "::")[0])
	}
	return newFilter
}
//terraform-provider-genesyscloud/genesyscloud/tfexporter
func GetAvailableExporterTypes() []string {
	exporters := GetResourceExporters(nil)
	log.Println("exporters")
	log.Println(exporters)
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

func sanitizeResourceNames(idMetaMap ResourceIDMetaMap) {
	for _, meta := range idMetaMap {
		meta.Name = SanitizeResourceName(meta.Name)
	}
}

func SanitizeResourceName(inputName string) string {
	name := unsafeNameChars.ReplaceAllStringFunc(inputName, escapeRune)
	if name != inputName {
		// Append a hash of the original name to ensure uniqueness for similar names
		// and that equivalent names are consistent across orgs
		algorithm := fnv.New32()
		algorithm.Write([]byte(inputName))
		name = name + "_" + strconv.FormatUint(uint64(algorithm.Sum32()), 10)
	}
	if unicode.IsDigit(rune(name[0])) {
		// Terraform does not allow names to begin with a number. Prefix with an underscore instead
		name = "_" + name
	}

	return name
}
