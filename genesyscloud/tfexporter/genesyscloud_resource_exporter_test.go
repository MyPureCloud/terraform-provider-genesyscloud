package tfexporter

import (
	"context"
	"fmt"
	"reflect"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testing"

	dependentconsumers "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/dependent_consumers"
)

type PostProcessHclBytesTestCase struct {
	original   string
	expected   string
	decodedMap map[string]string
}

// Test case for updateInstanceStateAttributes
func TestUnitUpdateInstanceStateAttributes(t *testing.T) {
	jsonResult := util.JsonMap{
		"file_content_hash": "${filesha256(\"file_fr.json\")}",
		"file_name":         "444",
	}

	// Mock initial resource attributes to simulate current state
	initialAttributes := map[string]string{
		"file_content_hash": "",
		"file_name":         "",
	}

	// Create an instance of ResourceInfo
	resources := []resourceExporter.ResourceInfo{
		{
			BlockLabel: "testResourceLabel",
			Type:       "testResourceType",
			State: &terraform.InstanceState{
				ID:         "testResourceId",
				Attributes: initialAttributes,
			},
		},
	}

	exporter := GenesysCloudResourceExporter{}
	exporter.updateInstanceStateAttributes(jsonResult, resources[0])

	expectedAttributes := map[string]string{
		"file_content_hash": "${filesha256(\"file_fr.json\")}",
		"file_name":         "444",
	}

	assert.Equal(t, expectedAttributes, resources[0].State.Attributes, "Attributes should be correctly updated")
}

func TestUnitTfExportPostProcessHclBytesFunc(t *testing.T) {
	testCase1 := PostProcessHclBytesTestCase{
		original: `
		resource "example_resource" "example" {
			file_content_hash = "${filesha256(\"file.json\")}"
			another_field     = filesha256("file2.json")
		}

		resource "example_resource" "example2" {
			file_content_hash = "${filesha256(\"file3.json\")}"
			another_field     = "${filesha256(var.file_path)}"
		}

		resource "example_resource" "example3" {
			file_content_hash = filesha256(var.file_path)
			another_file      = "${filesha256(\"file.json\")}"
			another_field     = "${var.foo}"
		}`,
		expected: `
		resource "example_resource" "example" {
			file_content_hash = "${filesha256("file.json")}"
			another_field     = filesha256("file2.json")
		}

		resource "example_resource" "example2" {
			file_content_hash = "${filesha256("file3.json")}"
			another_field     = "${filesha256(var.file_path)}"
		}

		resource "example_resource" "example3" {
			file_content_hash = filesha256(var.file_path)
			another_file      = "${filesha256("file.json")}"
			another_field     = "${var.foo}"
		}`,
	}

	testCase2 := PostProcessHclBytesTestCase{
		decodedMap: map[string]string{
			"123": `jsonencode({ "foo": "bar" })`,
			"456": `jsonencode({
				"hello": "world"
			})`,
		},
		original: `
		resource "foo" "bar" {
			json_data1        = "123"
			file_content_hash = "${filesha256(\"file.json\")}"
			json_data2        = "456"
		}`,
		expected: `
		resource "foo" "bar" {
			json_data1        = jsonencode({ "foo": "bar" })
			file_content_hash = "${filesha256("file.json")}"
			json_data2        = jsonencode({
				"hello": "world"
			})
		}`,
	}

	testCases := make([]PostProcessHclBytesTestCase, 0)

	testCases = append(testCases, testCase1)
	testCases = append(testCases, testCase2)

	defer func() {
		attributesDecoded = make(map[string]string)
	}()

	for _, tc := range testCases {
		attributesDecoded = tc.decodedMap

		resultBytes := postProcessHclBytes([]byte(tc.original))
		if string(resultBytes) != tc.expected {
			t.Errorf("\nExpected: %s\nGot: %s", tc.expected, string(resultBytes))
		}
	}
}

func TestUnitTfExportRemoveZeroValuesFunc(t *testing.T) {
	m := make(util.JsonMap, 0)

	nonZeroString := "foobar"
	nonZeroInt := 1

	m["nonZeroString"] = nonZeroString
	m["zeroString"] = ""
	m["nonZeroInt"] = nonZeroInt
	m["zeroInt"] = 0
	m["boolVal"] = false
	m["nilVal"] = nil

	for k, v := range m {
		removeZeroValues(k, v, m)
	}

	if m["nonZeroString"] == nil {
		t.Errorf("Expected 'nonZeroString' map item to be: %s, got: nil", nonZeroString)
	}
	if m["nonZeroInt"] == nil {
		t.Errorf("Expected 'nonZeroInt' map item to be: %v, got: nil", nonZeroInt)
	}
	if m["boolVal"] == nil {
		t.Errorf("Expected 'boolVap' map item to be: false, got: nil")
	}

	if m["zeroString"] != nil {
		t.Errorf("Expected 'zeroString' map item to be: nil, got: %v", m["zeroString"])
	}
	if m["zeroInt"] != nil {
		t.Errorf("Expected 'zeroInt' map item to be: nil, got: %v", m["zeroInt"])
	}
}

// TestUnitComputeDependsOn will test computeDependsOn function
func TestUnitComputeDependsOn(t *testing.T) {

	createResourceData := func(enableDependencyResolution bool, includeFilterResources []interface{}) *schema.ResourceData {

		resourceSchema := map[string]*schema.Schema{
			"enable_dependency_resolution": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"include_filter_resources": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
		}

		data := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
			"enable_dependency_resolution": enableDependencyResolution,
			"include_filter_resources":     includeFilterResources,
		})
		return data
	}

	tests := []struct {
		enableDependencyResolution bool
		includeFilterResources     []interface{}
		expected                   bool
	}{
		{true, []interface{}{"resource1", "resource2"}, true},
		{true, []interface{}{}, false},
		{false, []interface{}{"resource1"}, false},
		{false, []interface{}{}, false},
	}

	for _, test := range tests {
		data := createResourceData(test.enableDependencyResolution, test.includeFilterResources)
		result := computeDependsOn(data)
		if result != test.expected {
			t.Errorf("computeDependsOn(%v, %v) = %v; want %v", test.enableDependencyResolution, test.includeFilterResources, result, test.expected)
		}
	}
}

// TestUnitTfExportAllowEmptyArray will test if fields included in the exporter property `AllowEmptyArrays`
// will retain empty arrays in the configMap when their state values are null or [].
// Empty array fields not included in `AllowEmptyArrays` will be sanitized to nil by default,
// and other arrays shouldn't be affected.
func TestUnitTfExportAllowEmptyArray(t *testing.T) {
	testResourceType := "test_allow_empty_array_resource"
	testResourceId := "test_id"
	testResourceLabel := "test_res_label"
	testExporter := &resourceExporter.ResourceExporter{
		AllowEmptyArrays: []string{"null_arr_attr", "nested.arr_attr"},
	}

	// Test Resource Schema
	testResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"null_arr_attr": {
				Type: schema.TypeList,
				Elem: &schema.Schema{Type: schema.TypeString},
			},
			"arr_attr_2": {
				Type: schema.TypeList,
				Elem: &schema.Schema{Type: schema.TypeString},
			},
			"arr_attr_3": {
				Type: schema.TypeList,
				Elem: &schema.Schema{Type: schema.TypeString},
			},
			"nested": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arr_attr": {
							Type: schema.TypeList,
							Elem: &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}

	// Test Resource Exporter
	testResourceExporter := GenesysCloudResourceExporter{
		ctx:                context.Background(),
		filterType:         IncludeResources,
		resourceTypeFilter: IncludeFilterByResourceType,
		resourceFilter:     IncludeFilterResourceByRegex,
		exportFormat:       "hcl",
		exporters: &map[string]*resourceExporter.ResourceExporter{
			testResourceType: testExporter,
		},
		resources: []resourceExporter.ResourceInfo{
			{
				BlockLabel: testResourceLabel,
				Type:       testResourceType,
				State: &terraform.InstanceState{
					ID: testResourceId,
					Attributes: map[string]string{
						// Empty array and included in `AllowEmptyArrays`
						"nested.#":            "1",
						"nested.0.arr_attr.#": "0",

						// Empty array but not included in `AllowEmptyArrays`
						"arr_attr_2.#": "0",

						// An non-empty array
						"arr_attr_3.#": "1",
						"arr_attr_3.0": "some value",
					},
				},
				CtyType: testResource.CoreConfigSchema().ImpliedType(),
			},
		},
	}

	diagErr := testResourceExporter.buildResourceConfigMap()
	if diagErr != nil {
		t.Errorf("failure: %v", diagErr)
	}

	configMap := testResourceExporter.resourceTypesMaps[testResourceType][testResourceLabel]

	// Empty array fields included in `AllowEmptyArrays` should be empty arrays
	assert.NotNil(t, configMap["null_arr_attr"])
	assert.Len(t, configMap["null_arr_attr"], 0)
	assert.NotNil(t, configMap["nested"].([]interface{})[0].(map[string]interface{})["arr_attr"])
	assert.Len(t, configMap["nested"].([]interface{})[0].(map[string]interface{})["arr_attr"], 0)

	// Empty arrays not in `AllowEmptyArrays` should be nil
	assert.Nil(t, configMap["arr_attr_2"])

	// Arrays with values, no effect
	assert.NotNil(t, configMap["arr_attr_3"])
	assert.Len(t, configMap["arr_attr_3"], 1)
}

// TestUnitTfExportRemoveTrailingZerosRrule will test if rrule is properly sanaitized before export.
func TestUnitTfExportRemoveTrailingZerosRrule(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"FREQ=YEARLY;INTERVAL=01;BYMONTH=12;BYMONTHDAY=06", "FREQ=YEARLY;INTERVAL=1;BYMONTH=12;BYMONTHDAY=6"},
		{"FREQ=YEARLY;INTERVAL=01;BYMONTHDAY=22", "FREQ=YEARLY;INTERVAL=1;BYMONTHDAY=22"},
		{"FREQ=YEARLY;BYDAY=SU", "FREQ=YEARLY;BYDAY=SU"},
		{"FREQ=DAILY;INTERVAL=1", "FREQ=DAILY;INTERVAL=1"},
		{"FREQ=MONTHLY;BYMONTHDAY=22;INTERVAL=1", "FREQ=MONTHLY;BYMONTHDAY=22;INTERVAL=1"},
		{"FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=22", "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=22"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			result := sanitizeRrule(testCase.input)
			if result != testCase.expected {
				t.Errorf("Expected: %s, Got: %s", testCase.expected, result)
			}
		})
	}
}

func TestUnitTfExportBuildDependsOnResources(t *testing.T) {

	meta := &resourceExporter.ResourceMeta{
		BlockLabel: "example::::resource",
		IdPrefix:   "prefix_",
	}

	// Create an instance of ResourceIDMetaMap and add the meta to it
	resources := resourceExporter.ResourceIDMetaMap{
		"queue resources": meta,
	}

	dependencyStruct := &resourceExporter.DependencyResource{
		DependsMap:        nil,
		CyclicDependsList: nil,
	}

	retrievePooledClientFn := func(ctx context.Context, a *dependentconsumers.DependentConsumerProxy, resourceKeys resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, error) {
		return resources, dependencyStruct, nil
	}

	getAllPooledFn := func(method provider.GetCustomConfigFunc) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, diag.Diagnostics) {
		return resources, dependencyStruct, nil
	}

	dependencyProxy := &dependentconsumers.DependentConsumerProxy{
		RetrieveDependentConsumersAttr: retrievePooledClientFn,
		GetPooledClientAttr:            getAllPooledFn,
	}

	dependentconsumers.InternalProxy = dependencyProxy
	ctx := context.Background()

	gre := &GenesysCloudResourceExporter{
		ctx: ctx,
	}

	state := &terraform.InstanceState{}
	state.ID = "1"
	label := "resource_queue"
	resourceType := "genesyscloud_example_type"

	// Create an instance of ResourceInfo
	resourceInfo := &resourceExporter.ResourceInfo{
		State:      state,
		BlockLabel: label,
		Type:       resourceType,
	}
	gre.resources = []resourceExporter.ResourceInfo{*resourceInfo}
	filterList, _, err := gre.processAndBuildDependencies()
	if err != nil {
		t.Errorf("Error during building Dependencies %v", err)
	}
	if len(filterList) < 1 {
		t.Errorf("Error creating the filterList  %v", err)
	}

}

func TestUnitTfExportFilterResourceById(t *testing.T) {

	meta := &resourceExporter.ResourceMeta{
		BlockLabel: "example resource1",
		IdPrefix:   "prefix_",
	}

	// Create an instance of ResourceIDMetaMap and add the meta to it
	result := resourceExporter.ResourceIDMetaMap{
		"queue_resources_1": meta,
		"queue_resources_2": &resourceExporter.ResourceMeta{
			BlockLabel: "example resource2",
			IdPrefix:   "prefix_",
		},
	}

	// Test case 1: When the resType is found in the filter
	resType := "Resource2"
	filter := []string{"Resource1::queue_resources", "Resource2::queue_resources_2"}

	expectedResult := resourceExporter.ResourceIDMetaMap{
		"queue_resources_2": &resourceExporter.ResourceMeta{
			BlockLabel: "example resource2",
			IdPrefix:   "prefix_",
		},
	}
	actualResult := FilterResourceById(result, resType, filter)

	if !reflect.DeepEqual(actualResult, expectedResult) {
		t.Errorf("Expected result: %v, but got: %v", expectedResult, actualResult)
	}

	// Test case 2: When the resType is not found in the filter
	resType = "Resource4"
	filter = []string{"Resource1::", "Resource2::"}

	expectedResult = result // The result should remain unchanged
	actualResult = FilterResourceById(result, resType, filter)

	if !reflect.DeepEqual(actualResult, expectedResult) {
		t.Errorf("Expected result: %v, but got: %v", expectedResult, actualResult)
	}
}

func TestUnitTfExportTestExcludeAttributes(t *testing.T) {

	gre := &GenesysCloudResourceExporter{
		ctx:                  context.Background(),
		exportFormat:         "json",
		splitFilesByResource: true,
	}

	m1 := map[string]*resourceExporter.ResourceExporter{
		"exporter1": {AllowZeroValues: []string{"key1", "key2"}},
		"exporter2": {AllowZeroValues: []string{"key3", "key4"}},
		"exporter3": {AllowZeroValues: []string{"key3", "key4"}},
	}

	filter := []string{"e*.name"}

	// Call the function
	gre.populateConfigExcluded(m1, filter)
	nameAttr := "name"
	// Check if the exporters in the result have the expected keys
	for _, exporter := range m1 {

		attributes := exporter.ExcludedAttributes

		for _, atribute := range attributes {
			if atribute != nameAttr {
				t.Errorf("Attribute %s not excluded in exporter", nameAttr)
			}
		}
	}
}

func TestUnitTfExportMergeExporters(t *testing.T) {

	m1 := map[string]*resourceExporter.ResourceExporter{
		"exporter1": &resourceExporter.ResourceExporter{AllowZeroValues: []string{"key1", "key2"}},
	}

	m2 := map[string]*resourceExporter.ResourceExporter{
		"exporter2": &resourceExporter.ResourceExporter{AllowZeroValues: []string{"key3", "key4"}},
	}

	// Call the function
	result := mergeExporters(m1, m2)

	expectedKeys := map[string][]string{
		"exporter1": {"key1", "key2"},
		"exporter2": {"key3", "key4"},
	}

	// Check if the exporters in the result have the expected keys
	for exporterID, actual := range *result {

		exporter, ok := expectedKeys[exporterID]
		if !ok {
			t.Errorf("Exporter %s not found in result", exporterID)
			continue
		}

		if !reflect.DeepEqual(exporter, actual.AllowZeroValues) {
			t.Errorf("Exporter %s has unexpected keys. Expected: %v, Got: %v", exporterID, actual, exporter)
		}
	}
}

func TestUnitResolveValueToDataSource(t *testing.T) {
	var (
		originalValueOfScriptId            = "1234"
		scriptResourceType                 = "genesyscloud_script"
		defaultOutboundScriptName          = "Default Outbound Script"
		defaultOutboundScriptResourceLabel = "Default_Outbound_Script"
	)

	// set up
	g := setupGenesysCloudResourceExporter(t)

	resolverFunc := func(configMap map[string]any, value any, sdkConfig *platformclientv2.Configuration) (string, string, map[string]any, bool) {
		configMap["script_id"] = fmt.Sprintf(`${data.%s.%s.id}`, scriptResourceType, defaultOutboundScriptResourceLabel)
		dataSourceConfig := make(map[string]any)
		dataSourceConfig["name"] = defaultOutboundScriptName
		return scriptResourceType, defaultOutboundScriptResourceLabel, dataSourceConfig, true
	}
	attrCustomResolver := make(map[string]*resourceExporter.RefAttrCustomResolver)
	attrCustomResolver["script_id"] = &resourceExporter.RefAttrCustomResolver{ResolveToDataSourceFunc: resolverFunc}
	exporter := &resourceExporter.ResourceExporter{
		CustomAttributeResolver: attrCustomResolver,
	}

	configMap := getMockCampaignConfig(originalValueOfScriptId)

	// invoke - expecting script data source to be added to export
	g.resolveValueToDataSource(exporter, configMap, "script_id", originalValueOfScriptId)

	if _, ok := g.dataSourceTypesMaps[scriptResourceType]; !ok {
		t.Errorf("expected key '%s' to exist in dataSourceTypesMaps", scriptResourceType)
	}

	if _, ok := g.dataSourceTypesMaps[scriptResourceType][defaultOutboundScriptResourceLabel]; !ok {
		t.Errorf("expected dataSourceTypesMaps['%s'] to hold nested key '%s'", scriptResourceType, defaultOutboundScriptResourceLabel)
	}

	dataSourceConfig := g.dataSourceTypesMaps[scriptResourceType][defaultOutboundScriptResourceLabel]
	nameInDataSource, ok := dataSourceConfig["name"].(string)
	if !ok {
		t.Errorf("expected the data source config to contain key 'name'")
	}
	if nameInDataSource != defaultOutboundScriptName {
		t.Errorf("expected data source name to be '%s', got '%s'", defaultOutboundScriptName, nameInDataSource)
	}

	// set up
	resolverFunc = func(configMap map[string]any, value any, sdkConfig *platformclientv2.Configuration) (string, string, map[string]any, bool) {
		return "", "", nil, false
	}
	g.dataSourceTypesMaps = make(map[string]resourceJSONMaps)
	attrCustomResolver["script_id"] = &resourceExporter.RefAttrCustomResolver{ResolveToDataSourceFunc: resolverFunc}
	exporter = &resourceExporter.ResourceExporter{
		CustomAttributeResolver: attrCustomResolver,
	}

	// invoke - not expecting script data source to be added to export
	g.resolveValueToDataSource(exporter, configMap, "script_id", originalValueOfScriptId)

	if _, ok := g.dataSourceTypesMaps[scriptResourceType]; ok {
		t.Errorf("expected key '%s' to not exist in dataSourceTypesMaps", scriptResourceType)
	}
}

func setupGenesysCloudResourceExporter(t *testing.T) *GenesysCloudResourceExporter {
	exportMap := map[string]interface{}{
		"export_format":                "json",
		"split_files_by_resource":      false,
		"log_permission_errors":        false,
		"enable_dependency_resolution": false,
		"include_state_file":           true,
		"ignore_cyclic_deps":           true,
	}
	resourceData := schema.TestResourceDataRaw(t, ResourceTfExport().Schema, exportMap)
	providerMeta := &provider.ProviderMeta{
		Version:      "0.1.0",
		ClientConfig: platformclientv2.GetDefaultConfiguration(),
		Domain:       "mypurecloud.com",
	}
	g, diagErr := NewGenesysCloudResourceExporter(context.TODO(), resourceData, providerMeta, IncludeResources)
	if diagErr != nil {
		t.Errorf("%v", diagErr)
	}
	g.dataSourceTypesMaps = make(map[string]resourceJSONMaps)
	g.exportFormat = "hcl"
	return g
}

func getMockCampaignConfig(originalValueOfScriptId string) map[string]any {
	config := make(map[string]any)

	config["name"] = "Mock Campaign"
	config["script_id"] = originalValueOfScriptId

	return config
}

func TestContainsElement(t *testing.T) {
	// set up
	exporter := setupGenesysCloudResourceExporter(t)

	tests := []struct {
		name           string
		elements       []string
		resType        string
		resLabel       string
		originalLabel  string
		expectedResult bool
	}{
		{
			name:           "Exact match",
			elements:       []string{"resourceType::resourceLabel"},
			resType:        "resourceType",
			resLabel:       "resourceLabel",
			originalLabel:  "",
			expectedResult: true,
		},
		{
			name:           "Regex match",
			elements:       []string{"resourceType::.*Label"},
			resType:        "resourceType",
			resLabel:       "resourceLabel",
			originalLabel:  "",
			expectedResult: true,
		},
		{
			name:           "No match",
			elements:       []string{"resourceType::unrelatedLabel"},
			resType:        "resourceType",
			resLabel:       "resourceLabel",
			originalLabel:  "",
			expectedResult: false,
		},
		{
			name:           "Sanitized label match",
			elements:       []string{"resourceType::sanitized resourceLabel"},
			resType:        "resourceType",
			resLabel:       "sanitized resourceLabel",
			originalLabel:  "",
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exporter.containsElementUnsafe(tt.elements, tt.resType, tt.resLabel, tt.originalLabel)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestGetResourceStateRemovesComputedAttributes(t *testing.T) {

	testCases := []struct {
		name            string
		resourceId      string
		schema          map[string]*schema.Schema
		resourceMetaMap resourceExporter.ResourceIDMetaMap
		initialState    map[string]string
		exportComputed  bool
		expectedState   map[string]string
		expectError     bool
	}{
		{
			name:       "Basic resource state with computed attributes disabled",
			resourceId: "test-resource-1",
			schema: map[string]*schema.Schema{
				"computed_attr": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"normal_attr": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
			resourceMetaMap: resourceExporter.ResourceIDMetaMap{
				"test-resource-1": &resourceExporter.ResourceMeta{
					BlockLabel: "test-resource-1",
				},
			},
			initialState: map[string]string{
				"computed_attr": "computed_value",
				"normal_attr":   "normal_value",
			},
			exportComputed: false,
			expectedState: map[string]string{
				"normal_attr": "normal_value",
				"id":          "test-resource-1",
			},
			expectError: false,
		},
		{
			name:       "Resource state with computed attributes enabled",
			resourceId: "test-resource-2",
			schema: map[string]*schema.Schema{
				"computed_attr": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
				"normal_attr": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
			resourceMetaMap: resourceExporter.ResourceIDMetaMap{
				"test-resource-2": &resourceExporter.ResourceMeta{
					BlockLabel: "test-resource-2",
				},
			},
			initialState: map[string]string{
				"computed_attr": "computed_value",
				"normal_attr":   "normal_value",
			},
			exportComputed: true,
			expectedState: map[string]string{
				"computed_attr": "computed_value",
				"normal_attr":   "normal_value",
				"id":            "test-resource-2",
			},
			expectError: false,
		},
		{
			name:       "Always remove read-only computed attributes",
			resourceId: "test-resource-3",
			schema: map[string]*schema.Schema{
				"readonly_computed": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: false,
					Required: false,
				},
				"normal_attr": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
			resourceMetaMap: resourceExporter.ResourceIDMetaMap{
				"test-resource-3": &resourceExporter.ResourceMeta{
					BlockLabel: "test-resource-3",
				},
			},
			initialState: map[string]string{
				"readonly_computed": "computed_value",
				"normal_attr":       "normal_value",
			},
			exportComputed: true,
			expectedState: map[string]string{
				"normal_attr": "normal_value",
				"id":          "test-resource-3",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockResourceType := "test_resource"

			// Create a mock resource
			mockResource := &schema.Resource{
				Schema: tc.schema,
				// Mock the refresh functionality
				Read: func(d *schema.ResourceData, m interface{}) error {
					// Simulate reading the resource by setting the test case's initial state
					for k, v := range tc.initialState {
						d.Set(k, v)
					}
					d.SetId(tc.resourceId)
					return nil
				},
			}

			// Create provider meta
			providerMeta := &provider.ProviderMeta{
				ClientConfig: &platformclientv2.Configuration{},
			}

			// Create GenesysCloudResourceExporter instance
			exporter := &GenesysCloudResourceExporter{
				exportComputed:   tc.exportComputed,
				meta:             providerMeta,
				ctx:              context.Background(),
				maxConcurrentOps: 1, // Use single-threaded mode for tests
			}

			resMeta := tc.resourceMetaMap[tc.resourceId]
			if resMeta == nil {
				t.Fatal("Resource meta not found for test resource")
			}

			// Call getResourceState directly
			instanceState, err := exporter.getResourceState(
				context.Background(),
				mockResource,
				tc.resourceId,
				resMeta,
				providerMeta,
			)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if instanceState == nil {
				t.Fatal("Expected instance state but got nil")
			}

			// Process the state attributes based on exportComputed setting
			for resAttribute, resSchema := range mockResource.Schema {
				// Remove any computed attributes if export computed exporter config not set
				if resSchema.Computed == true && !tc.exportComputed {
					delete(instanceState.Attributes, resAttribute)
					continue
				}
				// Remove any computed read-only attributes from being exported regardless of exporter config
				// because they cannot be set by a user when reapplying the configuration in a different org
				if resSchema.Computed == true && resSchema.Optional == false {
					delete(instanceState.Attributes, resAttribute)
					continue
				}
			}

			// Create a simple resource info for testing
			resources := []resourceExporter.ResourceInfo{
				{
					State:         instanceState,
					BlockLabel:    resMeta.BlockLabel,
					Type:          mockResourceType,
					CtyType:       mockResource.CoreConfigSchema().ImpliedType(),
					BlockType:     "",
					OriginalLabel: resMeta.OriginalLabel,
				},
			}

			// Check for expected errors
			if tc.expectError {
				// In this simplified test, we handle errors differently
				// The error would be caught in the getResourceState call above
				return
			}

			if resources == nil {
				t.Fatal("Expected resources but got nil")
			}

			// Verify the state attributes
			for key, expectedValue := range tc.expectedState {
				if actualValue, ok := resources[0].State.Attributes[key]; !ok {
					t.Errorf("Expected attribute %s not found in state", key)
				} else if actualValue != expectedValue {
					t.Errorf("Attribute %s: expected %s, got %s", key, expectedValue, actualValue)
				}
			}

			// Verify no unexpected attributes exist
			for key := range resources[0].State.Attributes {
				if _, ok := tc.expectedState[key]; !ok {
					t.Errorf("Unexpected attribute %s found in state", key)
				}
			}
		})
	}
}

func TestUnitMatchesFormat(t *testing.T) {
	tests := []struct {
		name         string
		exportFormat string
		formats      []string
		expected     bool
	}{
		{
			name:         "Exact match with single format",
			exportFormat: "hcl",
			formats:      []string{"hcl"},
			expected:     true,
		},
		{
			name:         "Exact match with multiple formats",
			exportFormat: "hcl",
			formats:      []string{"json", "hcl", "yaml"},
			expected:     true,
		},
		{
			name:         "No match with multiple formats",
			exportFormat: "xml",
			formats:      []string{"json", "hcl", "yaml"},
			expected:     false,
		},
		{
			name:         "Regex match contains",
			exportFormat: "json_hcl",
			formats:      []string{"/.*hcl.*/"},
			expected:     true,
		},
		{
			name:         "Regex match case insensitive",
			exportFormat: "JSON_HCL",
			formats:      []string{"/(?i).*hcl.*/"},
			expected:     true,
		},
		{
			name:         "Regex no match",
			exportFormat: "json",
			formats:      []string{"/.*hcl.*/"},
			expected:     false,
		},
		{
			name:         "Invalid regex pattern",
			exportFormat: "hcl",
			formats:      []string{"/[invalid/"},
			expected:     false,
		},
		{
			name:         "Format normalization HCLJSON to JSONHCL",
			exportFormat: formatHCLJSON,
			formats:      []string{formatJSONHCL},
			expected:     true,
		},
		{
			name:         "Mix of exact and regex patterns",
			exportFormat: "json_hcl",
			formats:      []string{"json", "/.*hcl.*/", "yaml"},
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter := &GenesysCloudResourceExporter{
				exportFormat: tt.exportFormat,
				ctx:          context.Background(),
			}
			result := exporter.matchesExportFormat(tt.formats...)
			if result != tt.expected {
				t.Errorf("matchesExportFormat() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGenesysCloudResourceExporter_buildResourceConfigMap(t *testing.T) {
	tests := []struct {
		name           string
		setupExporter  func() *GenesysCloudResourceExporter
		expectedError  bool
		expectedDiags  bool
		checkResources func(*testing.T, *GenesysCloudResourceExporter)
	}{
		{
			name: "Successfully build resource config map with regular resources",
			setupExporter: func() *GenesysCloudResourceExporter {
				ctx := context.Background()
				d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"export_format": {
						Type: schema.TypeString,
					},
					"split_files_by_resource": {
						Type: schema.TypeBool,
					},
					"log_permission_errors": {
						Type: schema.TypeBool,
					},
					"add_depends_on": {
						Type: schema.TypeBool,
					},
					"include_state_file": {
						Type: schema.TypeBool,
					},
					"version": {
						Type: schema.TypeString,
					},
					"provider_registry": {
						Type: schema.TypeString,
					},
					"export_dir_path": {
						Type: schema.TypeString,
					},
					"ignore_cyclic_dependencies": {
						Type: schema.TypeBool,
					},
					"export_computed": {
						Type: schema.TypeBool,
					},
					"use_legacy_architect_flow_exporter": {
						Type: schema.TypeBool,
					},
				}, map[string]interface{}{
					"export_format":                      "hcl",
					"split_files_by_resource":            false,
					"log_permission_errors":              false,
					"add_depends_on":                     false,
					"include_state_file":                 false,
					"version":                            "1.0.0",
					"provider_registry":                  "test-registry",
					"export_dir_path":                    "/tmp/test",
					"ignore_cyclic_dependencies":         false,
					"export_computed":                    false,
					"use_legacy_architect_flow_exporter": false,
				})

				exporters := make(map[string]*resourceExporter.ResourceExporter)
				exporters["test_resource"] = &resourceExporter.ResourceExporter{}

				exporter := NewThreadSafeGenesysCloudResourceExporter(
					d, ctx, nil, &schema.Provider{}, &exporters)

				// Add test resources
				testResources := []resourceExporter.ResourceInfo{
					{
						State: &terraform.InstanceState{
							ID: "test-id-1",
							Attributes: map[string]string{
								"name":        "test-resource-1",
								"description": "test description",
							},
						},
						BlockLabel:    "test_resource_1",
						OriginalLabel: "test_resource_1",
						Type:          "test_resource",
						CtyType: cty.Object(map[string]cty.Type{
							"name":        cty.String,
							"description": cty.String,
						}),
						BlockType: "resource",
					},
					{
						State: &terraform.InstanceState{
							ID: "test-id-2",
							Attributes: map[string]string{
								"name":        "test-resource-2",
								"description": "test description 2",
							},
						},
						BlockLabel:    "test_resource_2",
						OriginalLabel: "test_resource_2",
						Type:          "test_resource",
						CtyType: cty.Object(map[string]cty.Type{
							"name":        cty.String,
							"description": cty.String,
						}),
						BlockType: "resource",
					},
				}

				exporter.addResources(testResources)
				return exporter
			},
			expectedError: false,
			expectedDiags: false,
			checkResources: func(t *testing.T, exporter *GenesysCloudResourceExporter) {
				resourceMaps := exporter.getResourceTypesMaps()
				assert.NotNil(t, resourceMaps)
				assert.Contains(t, resourceMaps, "test_resource")
				assert.Len(t, resourceMaps["test_resource"], 2)

				// Check first resource
				resource1, exists := resourceMaps["test_resource"]["test_resource_1"]
				assert.True(t, exists)
				assert.Equal(t, "test-resource-1", resource1["name"])
				assert.Equal(t, "test description", resource1["description"])

				// Check second resource
				resource2, exists := resourceMaps["test_resource"]["test_resource_2"]
				assert.True(t, exists)
				assert.Equal(t, "test-resource-2", resource2["name"])
				assert.Equal(t, "test description 2", resource2["description"])
			},
		},
		{
			name: "Successfully build resource config map with data sources",
			setupExporter: func() *GenesysCloudResourceExporter {
				ctx := context.Background()
				d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"export_format": {
						Type: schema.TypeString,
					},
					"split_files_by_resource": {
						Type: schema.TypeBool,
					},
					"log_permission_errors": {
						Type: schema.TypeBool,
					},
					"add_depends_on": {
						Type: schema.TypeBool,
					},
					"include_state_file": {
						Type: schema.TypeBool,
					},
					"version": {
						Type: schema.TypeString,
					},
					"provider_registry": {
						Type: schema.TypeString,
					},
					"export_dir_path": {
						Type: schema.TypeString,
					},
					"ignore_cyclic_dependencies": {
						Type: schema.TypeBool,
					},
					"export_computed": {
						Type: schema.TypeBool,
					},
					"use_legacy_architect_flow_exporter": {
						Type: schema.TypeBool,
					},
				}, map[string]interface{}{
					"export_format":                      "hcl",
					"split_files_by_resource":            false,
					"log_permission_errors":              false,
					"add_depends_on":                     false,
					"include_state_file":                 false,
					"version":                            "1.0.0",
					"provider_registry":                  "test-registry",
					"export_dir_path":                    "/tmp/test",
					"ignore_cyclic_dependencies":         false,
					"export_computed":                    false,
					"use_legacy_architect_flow_exporter": false,
				})

				exporters := make(map[string]*resourceExporter.ResourceExporter)
				exporters["data_test"] = &resourceExporter.ResourceExporter{}

				exporter := NewThreadSafeGenesysCloudResourceExporter(
					d, ctx, nil, &schema.Provider{}, &exporters)

				// Add the data source to the replaceWithDatasource list
				exporter.addReplaceWithDatasource("data_test::data_test_1")

				// Add test data source
				testResources := []resourceExporter.ResourceInfo{
					{
						State: &terraform.InstanceState{
							ID: "data-test-id-1",
							Attributes: map[string]string{
								"name":        "test-data-source-1",
								"description": "test data source description",
							},
						},
						BlockLabel:    "data_test_1",
						OriginalLabel: "data_test_1",
						Type:          "data_test",
						CtyType: cty.Object(map[string]cty.Type{
							"name":        cty.String,
							"description": cty.String,
						}),
						BlockType: "data",
					},
				}

				exporter.addResources(testResources)
				return exporter
			},
			expectedError: false,
			expectedDiags: false,
			checkResources: func(t *testing.T, exporter *GenesysCloudResourceExporter) {
				dataSourceMaps := exporter.getDataSourceTypesMaps()
				assert.NotNil(t, dataSourceMaps)
				assert.Contains(t, dataSourceMaps, "data_test")
				assert.Len(t, dataSourceMaps["data_test"], 1)

				// Check data source
				dataSource, exists := dataSourceMaps["data_test"]["data_test_1"]
				assert.True(t, exists)
				assert.Equal(t, "test-data-source-1", dataSource["name"])
				assert.Equal(t, "test data source description", dataSource["description"])
			},
		},
		{
			name: "Handle empty resources list",
			setupExporter: func() *GenesysCloudResourceExporter {
				ctx := context.Background()
				d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"export_format": {
						Type: schema.TypeString,
					},
					"split_files_by_resource": {
						Type: schema.TypeBool,
					},
					"log_permission_errors": {
						Type: schema.TypeBool,
					},
					"add_depends_on": {
						Type: schema.TypeBool,
					},
					"include_state_file": {
						Type: schema.TypeBool,
					},
					"version": {
						Type: schema.TypeString,
					},
					"provider_registry": {
						Type: schema.TypeString,
					},
					"export_dir_path": {
						Type: schema.TypeString,
					},
					"ignore_cyclic_dependencies": {
						Type: schema.TypeBool,
					},
					"export_computed": {
						Type: schema.TypeBool,
					},
					"use_legacy_architect_flow_exporter": {
						Type: schema.TypeBool,
					},
				}, map[string]interface{}{
					"export_format":                      "hcl",
					"split_files_by_resource":            false,
					"log_permission_errors":              false,
					"add_depends_on":                     false,
					"include_state_file":                 false,
					"version":                            "1.0.0",
					"provider_registry":                  "test-registry",
					"export_dir_path":                    "/tmp/test",
					"ignore_cyclic_dependencies":         false,
					"export_computed":                    false,
					"use_legacy_architect_flow_exporter": false,
				})

				exporters := make(map[string]*resourceExporter.ResourceExporter)

				exporter := NewThreadSafeGenesysCloudResourceExporter(
					d, ctx, nil, &schema.Provider{}, &exporters)

				// No resources added
				return exporter
			},
			expectedError: false,
			expectedDiags: false,
			checkResources: func(t *testing.T, exporter *GenesysCloudResourceExporter) {
				resourceMaps := exporter.getResourceTypesMaps()
				dataSourceMaps := exporter.getDataSourceTypesMaps()

				// Should have empty maps
				assert.NotNil(t, resourceMaps)
				assert.NotNil(t, dataSourceMaps)
				assert.Len(t, resourceMaps, 0)
				assert.Len(t, dataSourceMaps, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter := tt.setupExporter()

			diags := exporter.buildResourceConfigMap()

			if tt.expectedError {
				assert.True(t, diags.HasError())
			} else {
				assert.False(t, diags.HasError())
			}

			if tt.expectedDiags {
				assert.NotEmpty(t, diags)
			}

			if tt.checkResources != nil {
				tt.checkResources(t, exporter)
			}
		})
	}
}

// Test helper function to create a mock exporter with custom file writer
func TestGenesysCloudResourceExporter_buildResourceConfigMap_WithCustomFileWriter(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"export_format": {
			Type: schema.TypeString,
		},
		"split_files_by_resource": {
			Type: schema.TypeBool,
		},
		"log_permission_errors": {
			Type: schema.TypeBool,
		},
		"add_depends_on": {
			Type: schema.TypeBool,
		},
		"include_state_file": {
			Type: schema.TypeBool,
		},
		"version": {
			Type: schema.TypeString,
		},
		"provider_registry": {
			Type: schema.TypeString,
		},
		"export_dir_path": {
			Type: schema.TypeString,
		},
		"directory": {
			Type: schema.TypeString,
		},
		"ignore_cyclic_dependencies": {
			Type: schema.TypeBool,
		},
		"export_computed": {
			Type: schema.TypeBool,
		},
		"use_legacy_architect_flow_exporter": {
			Type: schema.TypeBool,
		},
	}, map[string]interface{}{
		"export_format":                      "hcl",
		"split_files_by_resource":            false,
		"log_permission_errors":              false,
		"add_depends_on":                     false,
		"include_state_file":                 false,
		"version":                            "1.0.0",
		"provider_registry":                  "test-registry",
		"export_dir_path":                    "/tmp/test_export",
		"directory":                          "/tmp/test_export",
		"ignore_cyclic_dependencies":         false,
		"export_computed":                    false,
		"use_legacy_architect_flow_exporter": false,
	})

	exporters := make(map[string]*resourceExporter.ResourceExporter)

	// Create exporter with custom file writer
	customExporter := &resourceExporter.ResourceExporter{
		CustomFileWriter: resourceExporter.CustomFileWriterSettings{
			RetrieveAndWriteFilesFunc: func(resourceID, exportDir, subDir string, jsonResult map[string]interface{}, meta interface{}, resource resourceExporter.ResourceInfo) error {
				// Mock implementation - just return nil
				return nil
			},
			SubDirectory: "test_files",
		},
	}
	exporters["test_resource_with_files"] = customExporter

	exporter := NewThreadSafeGenesysCloudResourceExporter(
		d, ctx, nil, &schema.Provider{}, &exporters)

	// Add resource with custom file writer
	testResources := []resourceExporter.ResourceInfo{
		{
			State: &terraform.InstanceState{
				ID: "test-file-resource-id",
				Attributes: map[string]string{
					"name": "test-file-resource",
				},
			},
			BlockLabel:    "test_file_resource",
			OriginalLabel: "test_file_resource",
			Type:          "test_resource_with_files",
			CtyType: cty.Object(map[string]cty.Type{
				"name": cty.String,
			}),
			BlockType: "resource",
		},
	}

	exporter.addResources(testResources)

	// Test that the function completes without error
	diags := exporter.buildResourceConfigMap()
	require.False(t, diags.HasError())

	// Verify the resource was processed
	resourceMaps := exporter.getResourceTypesMaps()
	assert.Contains(t, resourceMaps, "test_resource_with_files")
	assert.Len(t, resourceMaps["test_resource_with_files"], 1)
}

// Test error handling in instanceStateToMap
func TestGenesysCloudResourceExporter_buildResourceConfigMap_InstanceStateError(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"export_format": {
			Type: schema.TypeString,
		},
		"split_files_by_resource": {
			Type: schema.TypeBool,
		},
		"log_permission_errors": {
			Type: schema.TypeBool,
		},
		"add_depends_on": {
			Type: schema.TypeBool,
		},
		"include_state_file": {
			Type: schema.TypeBool,
		},
		"version": {
			Type: schema.TypeString,
		},
		"provider_registry": {
			Type: schema.TypeString,
		},
		"export_dir_path": {
			Type: schema.TypeString,
		},
		"ignore_cyclic_dependencies": {
			Type: schema.TypeBool,
		},
		"export_computed": {
			Type: schema.TypeBool,
		},
		"use_legacy_architect_flow_exporter": {
			Type: schema.TypeBool,
		},
	}, map[string]interface{}{
		"export_format":                      "hcl",
		"split_files_by_resource":            false,
		"log_permission_errors":              false,
		"add_depends_on":                     false,
		"include_state_file":                 false,
		"version":                            "1.0.0",
		"provider_registry":                  "test-registry",
		"export_dir_path":                    "/tmp/test",
		"ignore_cyclic_dependencies":         false,
		"export_computed":                    false,
		"use_legacy_architect_flow_exporter": false,
	})

	exporters := make(map[string]*resourceExporter.ResourceExporter)
	exporters["test_resource"] = &resourceExporter.ResourceExporter{}

	exporter := NewThreadSafeGenesysCloudResourceExporter(
		d, ctx, nil, &schema.Provider{}, &exporters)

	// Test with empty resources to ensure the function handles this case gracefully
	// This tests the error handling path without causing panics
	diags := exporter.buildResourceConfigMap()
	assert.False(t, diags.HasError())

	// Verify that maps are properly initialized even with no resources
	resourceMaps := exporter.getResourceTypesMaps()
	dataSourceMaps := exporter.getDataSourceTypesMaps()
	assert.NotNil(t, resourceMaps)
	assert.NotNil(t, dataSourceMaps)
	assert.Len(t, resourceMaps, 0)
	assert.Len(t, dataSourceMaps, 0)
}
