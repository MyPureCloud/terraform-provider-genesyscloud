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
	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"

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
	tests := []struct {
		enableDependencyResolution bool
		allowDependencyResolution  ExporterDependencyResolutionDecision
		expected                   bool
	}{
		{true, ExporterDependencyResolutionDecision(true), true},
		{true, ExporterDependencyResolutionDecision(false), false},
		{false, ExporterDependencyResolutionDecision(true), false},
		{false, ExporterDependencyResolutionDecision(false), false},
	}

	for _, test := range tests {
		result := computeDependsOn(test.enableDependencyResolution, test.allowDependencyResolution)
		if result != test.expected {
			t.Errorf("computeDependsOn(%v, %v) = %v; want %v",
				test.enableDependencyResolution, test.allowDependencyResolution, result, test.expected)
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
		maxConcurrentOps:   10,
		provider: &schema.Provider{
			ResourcesMap: map[string]*schema.Resource{
				testResourceType: testResource,
			},
		},
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

// TestUnitRemoveAllNilNestedBlocks directly exercises removeAllNilNestedBlocks to confirm it
// only removes blocks that are entirely nil, and leaves every other shape (real values, maps
// with real entries like credentials, AllowEmptyArrays-style non-nil empty slices, and plain
// scalar/string arrays) completely untouched. This backs the GitHub issue #2417 fix that
// suppresses an empty `config {}` shell without disturbing unrelated exporter behaviors.
func TestUnitRemoveAllNilNestedBlocks(t *testing.T) {
	t.Run("all-nil single block is removed", func(t *testing.T) {
		configMap := map[string]interface{}{
			"integration_type": "purecloud-data-actions",
			"config": []interface{}{
				map[string]interface{}{
					"name":        nil,
					"properties":  nil,
					"advanced":    nil,
					"notes":       nil,
					"credentials": nil,
				},
			},
		}
		removeAllNilNestedBlocks(configMap)
		assert.NotContains(t, configMap, "config", "all-nil config block should be removed entirely")
		assert.Equal(t, "purecloud-data-actions", configMap["integration_type"], "sibling scalar attribute must be untouched")
	})

	t.Run("block with a real map value (credentials) is kept", func(t *testing.T) {
		configMap := map[string]interface{}{
			"config": []interface{}{
				map[string]interface{}{
					"name":       nil,
					"properties": nil,
					"advanced":   nil,
					"notes":      nil,
					"credentials": map[string]interface{}{
						"pureCloudOAuthClient": "some-credential-guid",
					},
				},
			},
		}
		removeAllNilNestedBlocks(configMap)
		require.Contains(t, configMap, "config", "block holding a real credentials map must be kept")
		cfg := configMap["config"].([]interface{})[0].(map[string]interface{})
		assert.Equal(t, map[string]interface{}{"pureCloudOAuthClient": "some-credential-guid"}, cfg["credentials"])
	})

	t.Run("block with a real string value (notes) is kept", func(t *testing.T) {
		configMap := map[string]interface{}{
			"config": []interface{}{
				map[string]interface{}{
					"name":       nil,
					"properties": nil,
					"advanced":   nil,
					"notes":      "user notes",
				},
			},
		}
		removeAllNilNestedBlocks(configMap)
		require.Contains(t, configMap, "config")
		cfg := configMap["config"].([]interface{})[0].(map[string]interface{})
		assert.Equal(t, "user notes", cfg["notes"])
	})

	t.Run("AllowEmptyArrays non-nil empty slice keeps the parent block", func(t *testing.T) {
		// Mirrors what AllowForEmptyArrays produces during sanitize: configMap[key] = []interface{}{}
		configMap := map[string]interface{}{
			"nested": []interface{}{
				map[string]interface{}{
					"computed_field": nil,
					"arr_attr":       []interface{}{}, // non-nil empty slice, not nil
				},
			},
		}
		removeAllNilNestedBlocks(configMap)
		require.Contains(t, configMap, "nested", "block retaining a non-nil AllowEmptyArrays slice must be kept")
		nested := configMap["nested"].([]interface{})[0].(map[string]interface{})
		assert.NotNil(t, nested["arr_attr"])
		assert.Len(t, nested["arr_attr"], 0)
	})

	t.Run("plain scalar array is left untouched even if empty", func(t *testing.T) {
		configMap := map[string]interface{}{
			"tags": []interface{}{},
		}
		removeAllNilNestedBlocks(configMap)
		assert.Contains(t, configMap, "tags", "non-block (scalar) arrays must never be touched by this function")
	})

	t.Run("plain scalar array with values is left untouched", func(t *testing.T) {
		configMap := map[string]interface{}{
			"tags": []interface{}{"a", "b"},
		}
		removeAllNilNestedBlocks(configMap)
		assert.Equal(t, []interface{}{"a", "b"}, configMap["tags"])
	})

	t.Run("mixed list: one all-nil element removed, one real element kept", func(t *testing.T) {
		configMap := map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{"a": nil, "b": nil},
				map[string]interface{}{"a": "value", "b": nil},
			},
		}
		removeAllNilNestedBlocks(configMap)
		items := configMap["items"].([]interface{})
		require.Len(t, items, 1, "only the all-nil element should be dropped")
		assert.Equal(t, "value", items[0].(map[string]interface{})["a"])
	})

	t.Run("nested empty map inside a block collapses the parent too", func(t *testing.T) {
		configMap := map[string]interface{}{
			"outer": []interface{}{
				map[string]interface{}{
					"inner": map[string]interface{}{
						"leaf": nil,
					},
				},
			},
		}
		removeAllNilNestedBlocks(configMap)
		assert.NotContains(t, configMap, "outer", "a block whose only child (also emptied) leaves it all-nil should be removed")
	})

	t.Run("resource with no nested blocks at all is unaffected", func(t *testing.T) {
		configMap := map[string]interface{}{
			"integration_type": "purecloud-data-actions",
			"intended_state":   "ENABLED",
		}
		removeAllNilNestedBlocks(configMap)
		assert.Equal(t, "purecloud-data-actions", configMap["integration_type"])
		assert.Equal(t, "ENABLED", configMap["intended_state"])
	})
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
	// Reset singleton and trigger proxyOnce.Do first
	dependentconsumers.InternalProxy = nil
	_ = dependentconsumers.GetDependentConsumerProxy(nil) // Trigger proxyOnce.Do

	meta := &resourceExporter.ResourceMeta{
		BlockLabel: "example::::resource",
		IdPrefix:   "prefix_",
	}

	// Create an instance of ResourceIDMetaMap and add the meta to it
	resources := resourceExporter.ResourceIDMetaMap{
		"queue resources": meta,
	}

	dependencyStruct := &resourceExporter.DependencyResource{
		DependsMap:        map[string][]string{"1": {"example.resource"}},
		CyclicDependsList: nil,
	}

	// Mock the GetAllWithPooledClient to return directly without calling SDK pool
	getAllPooledFn := func(ctx context.Context, method provider.GetCustomConfigFunc) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, []string, diag.Diagnostics) {
		// Return mock data directly without calling the method
		return resources, dependencyStruct, nil, nil
	}

	dependencyProxy := &dependentconsumers.DependentConsumerProxy{
		GetPooledClientAttr: getAllPooledFn,
		ClientConfig:        &platformclientv2.Configuration{},
	}

	dependentconsumers.InternalProxy = dependencyProxy
	defer func() { dependentconsumers.InternalProxy = nil }()

	ctx := context.Background()

	gre := &GenesysCloudResourceExporter{
		ctx:               ctx,
		dependsList:       make(map[string][]string),
		flowResourcesList: []string{},
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
	// Test that removeExcludedAttrsFromMap correctly nils out excluded attributes
	configMap := util.JsonMap{
		"name":        "test-resource",
		"description": "a description",
		"nested": map[string]interface{}{
			"inner_attr": "value",
			"keep_attr":  "keep",
		},
	}

	excludedAttrs := []string{"name", "nested.inner_attr"}
	removeExcludedAttrsFromMap(configMap, excludedAttrs, "")

	if configMap["name"] != nil {
		t.Errorf("Expected 'name' to be nil, got %v", configMap["name"])
	}
	if configMap["description"] != "a description" {
		t.Errorf("Expected 'description' to be unchanged, got %v", configMap["description"])
	}
	nested := configMap["nested"].(map[string]interface{})
	if nested["inner_attr"] != nil {
		t.Errorf("Expected 'nested.inner_attr' to be nil, got %v", nested["inner_attr"])
	}
	if nested["keep_attr"] != "keep" {
		t.Errorf("Expected 'nested.keep_attr' to be unchanged, got %v", nested["keep_attr"])
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
	g.dataSourceTypesMaps = make(map[string]ResourceJSONMaps)
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
	g, diagErr := NewGenesysCloudResourceExporter(context.TODO(), resourceData, providerMeta, IncludeResources, AllowDependencyResolution)
	if diagErr != nil {
		t.Errorf("%v", diagErr)
	}
	g.dataSourceTypesMaps = make(map[string]ResourceJSONMaps)
	g.exportFormat = "hcl"
	return g
}

func getMockCampaignConfig(originalValueOfScriptId string) map[string]any {
	config := make(map[string]any)

	config["name"] = "Mock Campaign"
	config["script_id"] = originalValueOfScriptId

	return config
}

func TestUnitContainsElement(t *testing.T) {
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
			name:           "Type-only match",
			elements:       []string{"resourceType"},
			resType:        "resourceType",
			resLabel:       "anyLabel",
			originalLabel:  "",
			expectedResult: true,
		},
		{
			name:           "Type-only does not match other types",
			elements:       []string{"resourceType"},
			resType:        "otherResourceType",
			resLabel:       "anyLabel",
			originalLabel:  "",
			expectedResult: false,
		},
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

func TestUnitGetResourceStateRemovesComputedAttributes(t *testing.T) {

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
				mockResourceType,
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

func TestUnitGenesysCloudResourceExporter_buildResourceConfigMap(t *testing.T) {
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
					"export_omit_unresolved_refs": {
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
					"export_omit_unresolved_refs":        false,
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
					"export_omit_unresolved_refs": {
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
					"export_omit_unresolved_refs":        false,
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
					"export_omit_unresolved_refs": {
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
					"export_omit_unresolved_refs":        false,
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
func TestUnitGenesysCloudResourceExporter_buildResourceConfigMap_WithCustomFileWriter(t *testing.T) {
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
		"export_omit_unresolved_refs": {
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
		"export_omit_unresolved_refs":        false,
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

func TestUnitBuildResourceConfigMapExcludesSchemaBasedAttributes(t *testing.T) {
	resourceType := "test_resource"
	resourceLabel := "test_label"

	tests := []struct {
		name             string
		exportComputed   bool
		exportDeprecated bool
		resourceSchema   map[string]*schema.Schema
		stateAttrs       map[string]string
		checkConfigMap   func(*testing.T, util.JsonMap)
	}{
		{
			name:             "Top-level computed attribute removed from config map",
			exportComputed:   false,
			exportDeprecated: true,
			resourceSchema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"computed_field": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
			},
			stateAttrs: map[string]string{
				"name":           "my-resource",
				"computed_field": "server-generated",
			},
			checkConfigMap: func(t *testing.T, configMap util.JsonMap) {
				assert.Equal(t, "my-resource", configMap["name"])
				assert.Nil(t, configMap["computed_field"])
			},
		},
		{
			name:             "Nested computed attribute removed from config map",
			exportComputed:   false,
			exportDeprecated: true,
			resourceSchema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"value": {
								Type:     schema.TypeString,
								Required: true,
							},
							"internal_id": {
								Type:     schema.TypeString,
								Computed: true,
								Optional: false,
							},
						},
					},
				},
			},
			stateAttrs: map[string]string{
				"name":                   "my-resource",
				"settings.#":             "1",
				"settings.0.value":       "hello",
				"settings.0.internal_id": "abc-123",
			},
			checkConfigMap: func(t *testing.T, configMap util.JsonMap) {
				assert.Equal(t, "my-resource", configMap["name"])
				settings, ok := configMap["settings"].([]interface{})
				require.True(t, ok, "settings should be a list")
				require.Len(t, settings, 1)
				settingsMap, ok := settings[0].(map[string]interface{})
				require.True(t, ok, "settings[0] should be a map")
				assert.Equal(t, "hello", settingsMap["value"])
				assert.Nil(t, settingsMap["internal_id"])
			},
		},
		{
			name:             "Deprecated nested attribute removed from config map",
			exportComputed:   true,
			exportDeprecated: false,
			resourceSchema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"config": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"new_setting": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"old_setting": {
								Type:       schema.TypeString,
								Optional:   true,
								Deprecated: "Use new_setting",
							},
						},
					},
				},
			},
			stateAttrs: map[string]string{
				"name":                 "my-resource",
				"config.#":             "1",
				"config.0.new_setting": "modern",
				"config.0.old_setting": "legacy",
			},
			checkConfigMap: func(t *testing.T, configMap util.JsonMap) {
				assert.Equal(t, "my-resource", configMap["name"])
				configBlock, ok := configMap["config"].([]interface{})
				require.True(t, ok, "config should be a list")
				require.Len(t, configBlock, 1)
				configBlockMap, ok := configBlock[0].(map[string]interface{})
				require.True(t, ok, "config[0] should be a map")
				assert.Equal(t, "modern", configBlockMap["new_setting"])
				assert.Nil(t, configBlockMap["old_setting"])
			},
		},
		{
			name:             "Deeply nested computed attribute removed from config map",
			exportComputed:   false,
			exportDeprecated: true,
			resourceSchema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"outer": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"inner": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"keep_me": {
											Type:     schema.TypeString,
											Required: true,
										},
										"remove_me": {
											Type:     schema.TypeString,
											Computed: true,
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
			},
			stateAttrs: map[string]string{
				"name":                      "my-resource",
				"outer.#":                   "1",
				"outer.0.inner.#":           "1",
				"outer.0.inner.0.keep_me":   "important",
				"outer.0.inner.0.remove_me": "ephemeral",
			},
			checkConfigMap: func(t *testing.T, configMap util.JsonMap) {
				assert.Equal(t, "my-resource", configMap["name"])
				outer, ok := configMap["outer"].([]interface{})
				require.True(t, ok)
				require.Len(t, outer, 1)
				outerMap := outer[0].(map[string]interface{})
				inner, ok := outerMap["inner"].([]interface{})
				require.True(t, ok)
				require.Len(t, inner, 1)
				innerMap := inner[0].(map[string]interface{})
				assert.Equal(t, "important", innerMap["keep_me"])
				assert.Nil(t, innerMap["remove_me"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Build the test resource schema
			testResource := &schema.Resource{Schema: tt.resourceSchema}

			// Build a provider with the resource in ResourcesMap
			testProvider := &schema.Provider{
				ResourcesMap: map[string]*schema.Resource{
					resourceType: testResource,
				},
			}

			// Build the exporter
			exporters := map[string]*resourceExporter.ResourceExporter{
				resourceType: {},
			}

			d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				"export_format":                      {Type: schema.TypeString},
				"split_files_by_resource":            {Type: schema.TypeBool},
				"log_permission_errors":              {Type: schema.TypeBool},
				"add_depends_on":                     {Type: schema.TypeBool},
				"include_state_file":                 {Type: schema.TypeBool},
				"version":                            {Type: schema.TypeString},
				"provider_registry":                  {Type: schema.TypeString},
				"export_dir_path":                    {Type: schema.TypeString},
				"ignore_cyclic_dependencies":         {Type: schema.TypeBool},
				"export_computed":                    {Type: schema.TypeBool},
				"export_deprecated":                  {Type: schema.TypeBool},
				"export_omit_unresolved_refs":        {Type: schema.TypeBool},
				"use_legacy_architect_flow_exporter": {Type: schema.TypeBool},
			}, map[string]interface{}{
				"export_format":                      "hcl",
				"split_files_by_resource":            false,
				"log_permission_errors":              false,
				"add_depends_on":                     false,
				"include_state_file":                 false,
				"version":                            "1.0.0",
				"provider_registry":                  "test",
				"export_dir_path":                    "/tmp/test",
				"ignore_cyclic_dependencies":         false,
				"export_computed":                    tt.exportComputed,
				"export_deprecated":                  tt.exportDeprecated,
				"export_omit_unresolved_refs":        false,
				"use_legacy_architect_flow_exporter": false,
			})

			g := NewThreadSafeGenesysCloudResourceExporter(d, ctx, nil, testProvider, &exporters)
			g.exportComputed = tt.exportComputed
			g.exportDeprecated = tt.exportDeprecated

			// Add the resource
			g.addResources([]resourceExporter.ResourceInfo{
				{
					State: &terraform.InstanceState{
						ID:         "test-id",
						Attributes: tt.stateAttrs,
					},
					BlockLabel:    resourceLabel,
					OriginalLabel: resourceLabel,
					Type:          resourceType,
					CtyType:       testResource.CoreConfigSchema().ImpliedType(),
					BlockType:     "resource",
				},
			})

			diags := g.buildResourceConfigMap()
			require.False(t, diags.HasError(), "buildResourceConfigMap returned errors: %v", diags)

			resourceMaps := g.getResourceTypesMaps()
			require.Contains(t, resourceMaps, resourceType)
			require.Contains(t, resourceMaps[resourceType], resourceLabel)

			tt.checkConfigMap(t, resourceMaps[resourceType][resourceLabel])
		})
	}
}

func TestUnitSanitizeConfigMapOmitUnresolvedRefs(t *testing.T) {
	resourceType := "test_resource"
	guid := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

	exporter := &resourceExporter.ResourceExporter{
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"contact_list_id": {RefType: "genesyscloud_outbound_contact_list"},
		},
		CustomAttributeResolver: map[string]*resourceExporter.RefAttrCustomResolver{
			"contact_list_id": resourceExporter.OmitUnresolvedRefResolver(),
		},
	}
	exporters := map[string]*resourceExporter.ResourceExporter{
		resourceType:                         exporter,
		"genesyscloud_outbound_contact_list": {SanitizedResourceMap: map[string]*resourceExporter.ResourceMeta{}},
	}

	resource := resourceExporter.ResourceInfo{
		Type:       resourceType,
		BlockLabel: "test_label",
		BlockType:  "resource",
		State:      &terraform.InstanceState{ID: "ruleset-id"},
	}

	g := setupGenesysCloudResourceExporter(t)

	t.Run("keeps unresolved GUID when export_omit_unresolved_refs is false", func(t *testing.T) {
		g.exportOmitUnresolvedRefs = false
		configMap := map[string]interface{}{
			"name":            "test",
			"contact_list_id": guid,
		}
		_, ok := g.sanitizeConfigMap(resource, configMap, "", exporters, false, "hcl", true)
		require.True(t, ok)
		assert.Equal(t, guid, configMap["contact_list_id"])
	})

	t.Run("omits unresolved GUID when export_omit_unresolved_refs is true", func(t *testing.T) {
		g.exportOmitUnresolvedRefs = true
		configMap := map[string]interface{}{
			"name":            "test",
			"contact_list_id": guid,
		}
		_, ok := g.sanitizeConfigMap(resource, configMap, "", exporters, false, "hcl", true)
		require.True(t, ok)
		_, exists := configMap["contact_list_id"]
		assert.False(t, exists)
	})

	t.Run("keeps resolved reference when export_omit_unresolved_refs is true", func(t *testing.T) {
		g.exportOmitUnresolvedRefs = true
		resolvedRef := "${genesyscloud_outbound_contact_list.example.id}"
		configMap := map[string]interface{}{
			"name":            "test",
			"contact_list_id": resolvedRef,
		}
		_, ok := g.sanitizeConfigMap(resource, configMap, "", exporters, false, "hcl", true)
		require.True(t, ok)
		assert.Equal(t, resolvedRef, configMap["contact_list_id"])
	})
}

func TestUnitCollectSchemaBasedExcludedAttributes(t *testing.T) {
	resourceType := "genesyscloud_test_resource"

	tests := []struct {
		name             string
		exportComputed   bool
		exportDeprecated bool
		schemaMap        map[string]*schema.Schema
		prefix           string
		expected         []string
	}{
		{
			name:             "Top-level computed attribute excluded when exportComputed is false",
			exportComputed:   false,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"computed_field": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
				"normal_field": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
			expected: []string{"computed_field"},
		},
		{
			name:             "Top-level computed attribute NOT excluded when exportComputed is true",
			exportComputed:   true,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"computed_field": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
				"normal_field": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
			expected: []string{},
		},
		{
			name:             "Read-only computed attribute always excluded even when exportComputed is true",
			exportComputed:   true,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"readonly_computed": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: false,
				},
				"normal_field": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
			expected: []string{"readonly_computed"},
		},
		{
			name:             "Deprecated attribute excluded when exportDeprecated is false",
			exportComputed:   true,
			exportDeprecated: false,
			schemaMap: map[string]*schema.Schema{
				"old_field": {
					Type:       schema.TypeString,
					Optional:   true,
					Deprecated: "Use new_field instead",
				},
				"new_field": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
			expected: []string{"old_field"},
		},
		{
			name:             "Deprecated attribute NOT excluded when exportDeprecated is true",
			exportComputed:   true,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"old_field": {
					Type:       schema.TypeString,
					Optional:   true,
					Deprecated: "Use new_field instead",
				},
			},
			expected: []string{},
		},
		{
			name:             "Nested computed attribute in a list block",
			exportComputed:   false,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"parent_block": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"child_computed": {
								Type:     schema.TypeString,
								Computed: true,
								Optional: true,
							},
							"child_normal": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
			},
			expected: []string{"parent_block.child_computed"},
		},
		{
			name:             "Nested read-only computed attribute in a set block",
			exportComputed:   true,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"settings": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"server_generated_id": {
								Type:     schema.TypeString,
								Computed: true,
								Optional: false,
							},
							"value": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
			},
			expected: []string{"settings.server_generated_id"},
		},
		{
			name:             "Nested deprecated attribute",
			exportComputed:   true,
			exportDeprecated: false,
			schemaMap: map[string]*schema.Schema{
				"config": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"legacy_setting": {
								Type:       schema.TypeString,
								Optional:   true,
								Deprecated: "Use modern_setting instead",
							},
							"modern_setting": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
			},
			expected: []string{"config.legacy_setting"},
		},
		{
			name:             "Deeply nested computed attribute (3 levels)",
			exportComputed:   false,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"level1": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"level2": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"level3_computed": {
											Type:     schema.TypeString,
											Computed: true,
											Optional: true,
										},
										"level3_normal": {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
							"level2_normal": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
			},
			expected: []string{"level1.level2.level3_computed"},
		},
		{
			name:             "Multiple exclusions at different nesting levels",
			exportComputed:   false,
			exportDeprecated: false,
			schemaMap: map[string]*schema.Schema{
				"top_computed": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
				"top_deprecated": {
					Type:       schema.TypeString,
					Optional:   true,
					Deprecated: "removed",
				},
				"block": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"nested_readonly": {
								Type:     schema.TypeString,
								Computed: true,
								Optional: false,
							},
							"nested_deprecated": {
								Type:       schema.TypeInt,
								Optional:   true,
								Deprecated: "no longer used",
							},
							"nested_normal": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"normal_field": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
			expected: []string{"top_computed", "top_deprecated", "block.nested_readonly", "block.nested_deprecated"},
		},
		{
			name:             "List with simple Elem (no nested schema) is not recursed",
			exportComputed:   false,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"tags": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"computed_tags": {
					Type:     schema.TypeList,
					Computed: true,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
			},
			expected: []string{"computed_tags"},
		},
		{
			name:             "Map type with computed flag",
			exportComputed:   false,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"metadata": {
					Type:     schema.TypeMap,
					Computed: true,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"labels": {
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
			},
			expected: []string{"metadata"},
		},
		{
			name:             "Prefix is applied correctly",
			exportComputed:   false,
			exportDeprecated: true,
			prefix:           "outer_block",
			schemaMap: map[string]*schema.Schema{
				"inner_computed": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
				"inner_normal": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
			expected: []string{"outer_block.inner_computed"},
		},
		{
			name:             "No attributes excluded when all are normal",
			exportComputed:   true,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"description": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
			expected: []string{},
		},
		{
			name:             "Empty schema returns no exclusions",
			exportComputed:   false,
			exportDeprecated: false,
			schemaMap:        map[string]*schema.Schema{},
			expected:         []string{},
		},
		{
			name:             "Computed parent block with non-computed child is recursed, child preserved",
			exportComputed:   false,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"computed_block": {
					Type:     schema.TypeList,
					Computed: true,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"child_attr": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
			},
			// The parent is a nested block flagged Computed, but its child is user-settable.
			// We must recurse instead of excluding the whole block (GitHub issue #2417), so the
			// non-computed child is preserved and nothing is excluded.
			expected: nil,
		},
		{
			name:             "Computed parent block with mixed children excludes only computed leaves",
			exportComputed:   false,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				// Mirrors genesyscloud_integration.config: block itself Computed, some children
				// user-set (notes/credentials) and some computed (name/properties/advanced).
				"config": {
					Type:     schema.TypeList,
					Computed: true,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name":        {Type: schema.TypeString, Optional: true, Computed: true},
							"notes":       {Type: schema.TypeString, Optional: true},
							"properties":  {Type: schema.TypeString, Optional: true, Computed: true},
							"advanced":    {Type: schema.TypeString, Optional: true, Computed: true},
							"credentials": {Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
						},
					},
				},
			},
			// The block is preserved; only the computed leaf children are excluded.
			expected: []string{"config.name", "config.properties", "config.advanced"},
		},
		{
			name:             "Computed parent block with ALL computed children collapses to the block path",
			exportComputed:   false,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				// Every child is computed, so the block would export as an empty `settings {}`.
				// We collapse to excluding the whole block instead of emitting an empty shell.
				"settings": {
					Type:     schema.TypeList,
					Computed: true,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"generated_a": {Type: schema.TypeString, Optional: true, Computed: true},
							"generated_b": {Type: schema.TypeString, Optional: true, Computed: true},
						},
					},
				},
			},
			// Whole block excluded (not "settings.generated_a"/"settings.generated_b").
			expected: []string{"settings"},
		},
		{
			// Regression test for PR #2554 review comment: the block-collapse behavior must be
			// scoped to export_computed=false. Under export_computed=true (even with
			// export_deprecated=false), a block whose only child is deprecated must NOT collapse -
			// only the deprecated leaf is excluded, matching pre-fix behavior for this setting.
			name:             "All-deprecated block does NOT collapse when exportComputed is true",
			exportComputed:   true,
			exportDeprecated: false,
			schemaMap: map[string]*schema.Schema{
				"block": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"old_field": {Type: schema.TypeString, Optional: true, Deprecated: "use new_field"},
						},
					},
				},
			},
			expected: []string{"block.old_field"},
		},
		{
			// Regression test for PR #2554 review comment: a block whose only child is a pure
			// read-only computed field (e.g. workforcemanagement_businessunits' metadata.version)
			// is always excluded regardless of export_computed, but under the DEFAULT settings
			// (export_computed=true) the block itself must NOT collapse - only pre-existing
			// behavior (empty shell) is preserved on this path.
			name:             "All-readonly-computed block does NOT collapse when exportComputed is true (default)",
			exportComputed:   true,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"metadata": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"version": {Type: schema.TypeInt, Computed: true},
						},
					},
				},
			},
			expected: []string{"metadata.version"},
		},
		{
			// Companion case: the SAME all-readonly-computed block DOES collapse to the block
			// path when export_computed=false, since that's the scenario the collapse behavior
			// is meant to cover.
			name:             "All-readonly-computed block collapses when exportComputed is false",
			exportComputed:   false,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"metadata": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"version": {Type: schema.TypeInt, Computed: true},
						},
					},
				},
			},
			expected: []string{"metadata"},
		},
		{
			// A nested BLOCK itself (not just a leaf) that is pure read-only (Computed, not
			// Optional, not Required) is excluded in full without recursing into its children,
			// even when a child is a normal user-settable field, and regardless of exportComputed.
			// This is intentional and predates the #2417 fix: unlike an Optional+Computed block,
			// a purely-Computed block can never be written by a user in HCL at all, so there is
			// nothing to preserve by recursing into it.
			name:             "Purely read-only BLOCK (Computed, not Optional) is excluded whole, even with a normal child, exportComputed true",
			exportComputed:   true,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"readonly_block": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"child": {Type: schema.TypeString, Optional: true},
						},
					},
				},
			},
			expected: []string{"readonly_block"},
		},
		{
			// Same purely read-only BLOCK case, but with exportComputed=false - behavior is
			// identical, confirming this exclusion does not depend on exportComputed at all.
			name:             "Purely read-only BLOCK (Computed, not Optional) is excluded whole, exportComputed false",
			exportComputed:   false,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"readonly_block": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"child": {Type: schema.TypeString, Optional: true},
						},
					},
				},
			},
			expected: []string{"readonly_block"},
		},
		{
			// Mixed reasons INSIDE a block that is itself Optional+Computed: one child computed,
			// one child deprecated, one child a normal survivor. Only the two problem children
			// should be excluded; the block and the surviving normal child must remain.
			name:             "Computed parent block with computed AND deprecated children, one normal child survives",
			exportComputed:   false,
			exportDeprecated: false,
			schemaMap: map[string]*schema.Schema{
				"block": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"computed_child":   {Type: schema.TypeString, Optional: true, Computed: true},
							"deprecated_child": {Type: schema.TypeString, Optional: true, Deprecated: "old"},
							"normal_child":     {Type: schema.TypeString, Optional: true},
						},
					},
				},
			},
			expected: []string{"block.computed_child", "block.deprecated_child"},
		},
		{
			// Same shape as above, but ALL children are excluded (computed + deprecated only,
			// no normal survivor) with export_computed=false AND export_deprecated=false. The
			// block should collapse to the whole block path.
			name:             "Computed parent block with computed AND deprecated children, none survive, collapses",
			exportComputed:   false,
			exportDeprecated: false,
			schemaMap: map[string]*schema.Schema{
				"block": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"computed_child":   {Type: schema.TypeString, Optional: true, Computed: true},
							"deprecated_child": {Type: schema.TypeString, Optional: true, Deprecated: "old"},
						},
					},
				},
			},
			expected: []string{"block"},
		},
		{
			// A block that is Optional+Computed, with a nested SUB-block as its only child, where
			// that sub-block itself becomes fully excluded (collapses to its own path). The outer
			// block should then ALSO collapse, since its only child (the sub-block path) is excluded.
			name:             "Two levels of Optional+Computed blocks, both all-computed, both collapse",
			exportComputed:   false,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"outer": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"inner": {
								Type:     schema.TypeList,
								Optional: true,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"generated": {Type: schema.TypeString, Optional: true, Computed: true},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"outer"},
		},
		{
			// Same two-level nesting, but the INNER block has one normal child. The inner block
			// should NOT collapse (child survives), and therefore the OUTER block must also NOT
			// collapse, since its only child ("inner") is not in the excluded set.
			name:             "Two levels of Optional+Computed blocks, inner has a surviving child, neither collapses",
			exportComputed:   false,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"outer": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"inner": {
								Type:     schema.TypeList,
								Optional: true,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"generated": {Type: schema.TypeString, Optional: true, Computed: true},
										"notes":     {Type: schema.TypeString, Optional: true},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"outer.inner.generated"},
		},
		{
			// A block that is Optional (NOT computed) but ends up with all children excluded
			// (all computed leaves, export_computed=false). Confirms the collapse behavior does
			// not require the container itself to be Computed - it only cares whether every
			// child ends up excluded.
			name:             "Non-computed Optional block with all-computed children collapses when exportComputed false",
			exportComputed:   false,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"plain_block": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"generated_a": {Type: schema.TypeString, Optional: true, Computed: true},
							"generated_b": {Type: schema.TypeString, Optional: true, Computed: true},
						},
					},
				},
			},
			expected: []string{"plain_block"},
		},
		{
			// Same non-computed Optional block, but export_computed=true: children survive
			// (Optional+Computed leaves are kept when exportComputed=true), so nothing at all
			// should be excluded and the block stays fully intact.
			name:             "Non-computed Optional block with all-computed children: nothing excluded when exportComputed true",
			exportComputed:   true,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"plain_block": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"generated_a": {Type: schema.TypeString, Optional: true, Computed: true},
							"generated_b": {Type: schema.TypeString, Optional: true, Computed: true},
						},
					},
				},
			},
			expected: nil,
		},
		{
			// A block (TypeSet, not TypeList) that is Optional+Computed with all-computed
			// children - confirms the collapse logic applies to TypeSet blocks too, not just
			// TypeList.
			name:             "TypeSet Optional+Computed block with all-computed children collapses",
			exportComputed:   false,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"tag_set": {
					Type:     schema.TypeSet,
					Optional: true,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"generated": {Type: schema.TypeString, Optional: true, Computed: true},
						},
					},
				},
			},
			expected: []string{"tag_set"},
		},
		{
			name:             "Deeply nested with mixed exclusion reasons",
			exportComputed:   false,
			exportDeprecated: false,
			schemaMap: map[string]*schema.Schema{
				"routing": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"rules": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"priority": {
											Type:     schema.TypeInt,
											Required: true,
										},
										"internal_id": {
											Type:     schema.TypeString,
											Computed: true,
											Optional: false,
										},
										"old_weight": {
											Type:       schema.TypeInt,
											Optional:   true,
											Deprecated: "Use weight instead",
										},
									},
								},
							},
							"timeout": {
								Type:     schema.TypeInt,
								Computed: true,
								Optional: true,
							},
						},
					},
				},
			},
			expected: []string{"routing.rules.internal_id", "routing.rules.old_weight", "routing.timeout"},
		},
		{
			name:             "Read-only computed excluded before computed+optional check when exportComputed is false",
			exportComputed:   false,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"read_only_id": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: false,
				},
				"computed_optional": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
				"normal": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
			expected: []string{"read_only_id", "computed_optional"},
		},
		{
			name:             "Read-only computed excluded but computed+optional kept when exportComputed is true",
			exportComputed:   true,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"read_only_id": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: false,
				},
				"computed_optional": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
				"normal": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
			expected: []string{"read_only_id"},
		},
		{
			name:             "Nested read-only computed always excluded regardless of exportComputed",
			exportComputed:   true,
			exportDeprecated: true,
			schemaMap: map[string]*schema.Schema{
				"block": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"server_id": {
								Type:     schema.TypeString,
								Computed: true,
								Optional: false,
							},
							"default_value": {
								Type:     schema.TypeString,
								Computed: true,
								Optional: true,
							},
							"user_value": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
			},
			// server_id is read-only computed (always excluded), default_value is computed+optional (kept when exportComputed=true)
			expected: []string{"block.server_id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GenesysCloudResourceExporter{
				ctx:              context.Background(),
				exportComputed:   tt.exportComputed,
				exportDeprecated: tt.exportDeprecated,
			}

			result := g.collectSchemaBasedExcludedAttributes(resourceType, tt.schemaMap, tt.prefix)

			if tt.expected == nil || len(tt.expected) == 0 {
				assert.Empty(t, result)
			} else {
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

// Test error handling in instanceStateToMap
func TestUnitGenesysCloudResourceExporter_buildResourceConfigMap_InstanceStateError(t *testing.T) {
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
		"export_omit_unresolved_refs": {
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
		"export_omit_unresolved_refs":        false,
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
