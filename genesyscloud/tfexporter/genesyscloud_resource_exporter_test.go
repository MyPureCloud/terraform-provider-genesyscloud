package tfexporter

import (
	"context"
	"fmt"
	"reflect"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"

	"testing"

	dependentconsumers "terraform-provider-genesyscloud/genesyscloud/dependent_consumers"
)

type PostProcessHclBytesTestCase struct {
	original   string
	expected   string
	decodedMap map[string]string
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
	testResourceName := "test_res_name"
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
		filterType:         IncludeResources,
		resourceTypeFilter: IncludeFilterByResourceType,
		resourceFilter:     IncludeFilterResourceByRegex,
		exportAsHCL:        true,
		exporters: &map[string]*resourceExporter.ResourceExporter{
			testResourceType: testExporter,
		},
		resources: []resourceExporter.ResourceInfo{
			{
				Name: testResourceName,
				Type: testResourceType,
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

	configMap := testResourceExporter.resourceTypesMaps[testResourceType][testResourceName]

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
		Name:     "example::::resource",
		IdPrefix: "prefix_",
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
		//assert.Equal(t, targetName, name)
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
	name := "genesyscloud_resource_queue"
	resourceType := "example_type"

	// Create an instance of ResourceInfo
	resourceInfo := &resourceExporter.ResourceInfo{
		State: state,
		Name:  name,
		Type:  resourceType,
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
		Name:     "example resource1",
		IdPrefix: "prefix_",
	}

	// Create an instance of ResourceIDMetaMap and add the meta to it
	result := resourceExporter.ResourceIDMetaMap{
		"queue_resources_1": meta,
		"queue_resources_2": &resourceExporter.ResourceMeta{
			Name:     "example resource2",
			IdPrefix: "prefix_",
		},
	}

	// Test case 1: When the name is found in the filter
	name := "Resource2"
	filter := []string{"Resource1::queue_resources", "Resource2::queue_resources_2"}

	expectedResult := resourceExporter.ResourceIDMetaMap{
		"queue_resources_2": &resourceExporter.ResourceMeta{
			Name:     "example resource2",
			IdPrefix: "prefix_",
		},
	}
	actualResult := FilterResourceById(result, name, filter)

	if !reflect.DeepEqual(actualResult, expectedResult) {
		t.Errorf("Expected result: %v, but got: %v", expectedResult, actualResult)
	}

	// Test case 2: When the name is not found in the filter
	name = "Resource4"
	filter = []string{"Resource1::", "Resource2::"}

	expectedResult = result // The result should remain unchanged
	actualResult = FilterResourceById(result, name, filter)

	if !reflect.DeepEqual(actualResult, expectedResult) {
		t.Errorf("Expected result: %v, but got: %v", expectedResult, actualResult)
	}
}

func TestUnitTfExportTestExcludeAttributes(t *testing.T) {

	gre := &GenesysCloudResourceExporter{
		exportAsHCL:          false,
		splitFilesByResource: true,
	}

	m1 := map[string]*resourceExporter.ResourceExporter{
		"exporter1": &resourceExporter.ResourceExporter{AllowZeroValues: []string{"key1", "key2"}},
		"exporter2": &resourceExporter.ResourceExporter{AllowZeroValues: []string{"key3", "key4"}},
		"exporter3": &resourceExporter.ResourceExporter{AllowZeroValues: []string{"key3", "key4"}},
	}

	filter := []string{"e*.name"}

	// Call the function
	gre.populateConfigExcluded(m1, filter)
	name := "name"
	// Check if the exporters in the result have the expected keys
	for _, exporter := range m1 {

		attributes := exporter.ExcludedAttributes

		for _, atribute := range attributes {
			if atribute != name {
				t.Errorf("Attribute %s not excluded in exporter", name)
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
		originalValueOfScriptId         = "1234"
		scriptResourceId                = "genesyscloud_script"
		defaultOutboundScriptName       = "Default Outbound Script"
		defaultOutboundScriptResourceId = "Default_Outbound_Script"
	)

	// set up
	g := setupGenesysCloudResourceExporter(t)

	resolverFunc := func(configMap map[string]any, value any, sdkConfig *platformclientv2.Configuration) (string, string, map[string]any, bool) {
		configMap["script_id"] = fmt.Sprintf(`${data.%s.%s.id}`, scriptResourceId, defaultOutboundScriptResourceId)
		dataSourceConfig := make(map[string]any)
		dataSourceConfig["name"] = defaultOutboundScriptName
		return scriptResourceId, defaultOutboundScriptResourceId, dataSourceConfig, true
	}
	attrCustomResolver := make(map[string]*resourceExporter.RefAttrCustomResolver)
	attrCustomResolver["script_id"] = &resourceExporter.RefAttrCustomResolver{ResolveToDataSourceFunc: resolverFunc}
	exporter := &resourceExporter.ResourceExporter{
		CustomAttributeResolver: attrCustomResolver,
	}

	configMap := getMockCampaignConfig(originalValueOfScriptId)

	// invoke - expecting script data source to be added to export
	g.resolveValueToDataSource(exporter, configMap, "script_id", originalValueOfScriptId)

	if _, ok := g.dataSourceTypesMaps[scriptResourceId]; !ok {
		t.Errorf("expected key '%s' to exist in dataSourceTypesMaps", scriptResourceId)
	}

	if _, ok := g.dataSourceTypesMaps[scriptResourceId][defaultOutboundScriptResourceId]; !ok {
		t.Errorf("expected dataSourceTypesMaps['%s'] to hold nested key '%s'", scriptResourceId, defaultOutboundScriptResourceId)
	}

	dataSourceConfig := g.dataSourceTypesMaps[scriptResourceId][defaultOutboundScriptResourceId]
	nameInDataSource, ok := dataSourceConfig["name"].(string)
	if !ok {
		t.Errorf("expected the data source config to contain key 'name'")
	}
	if nameInDataSource != defaultOutboundScriptName {
		t.Errorf("expected data source name to be '%s', got '%s'", defaultOutboundScriptName, nameInDataSource)
	}

	hclBlocks, ok := g.resourceTypesHCLBlocks[scriptResourceId]
	if !ok {
		t.Errorf("expected resourceTypesHCLBlocks to contain key '%s'", scriptResourceId)
	}
	if len(hclBlocks) == 0 {
		t.Errorf("expected length of resourceTypesHCLBlocks to not be zero")
	}

	// set up
	resolverFunc = func(configMap map[string]any, value any, sdkConfig *platformclientv2.Configuration) (string, string, map[string]any, bool) {
		return "", "", nil, false
	}
	g.dataSourceTypesMaps = make(map[string]resourceJSONMaps)
	g.resourceTypesHCLBlocks = make(map[string]resourceHCLBlock)
	attrCustomResolver["script_id"] = &resourceExporter.RefAttrCustomResolver{ResolveToDataSourceFunc: resolverFunc}
	exporter = &resourceExporter.ResourceExporter{
		CustomAttributeResolver: attrCustomResolver,
	}

	// invoke - not expecting script data source to be added to export
	g.resolveValueToDataSource(exporter, configMap, "script_id", originalValueOfScriptId)

	if _, ok := g.dataSourceTypesMaps[scriptResourceId]; ok {
		t.Errorf("expected key '%s' to not exist in dataSourceTypesMaps", scriptResourceId)
	}

	if _, ok := g.resourceTypesHCLBlocks[scriptResourceId]; ok {
		t.Errorf("expected key '%s' to not exist in resourceTypesHCLBlocks map", scriptResourceId)
	}
}

func setupGenesysCloudResourceExporter(t *testing.T) *GenesysCloudResourceExporter {
	exportMap := map[string]interface{}{
		"export_as_hcl":                false,
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
	g.resourceTypesHCLBlocks = make(map[string]resourceHCLBlock)
	g.exportAsHCL = true
	return g
}

func getMockCampaignConfig(originalValueOfScriptId string) map[string]any {
	config := make(map[string]any)

	config["name"] = "Mock Campaign"
	config["script_id"] = originalValueOfScriptId

	return config
}
