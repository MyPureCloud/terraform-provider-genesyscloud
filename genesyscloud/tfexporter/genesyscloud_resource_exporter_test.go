package tfexporter

import (
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"

	"testing"
)

type PostProcessHclBytesTestCase struct {
	original   string
	expected   string
	decodedMap map[string]string
}

func TestPostProcessHclBytesFunc(t *testing.T) {
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

func TestRemoveZeroValuesFunc(t *testing.T) {
	m := make(gcloud.JsonMap, 0)

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

// TestAllowEmptyArray will test if fields included in the exporter property `AllowEmptyArrays`
// will retain empty arrays in the configMap when their state values are null or [].
// Empty array fields not included in `AllowEmptyArrays` will be sanitized to nil by default,
// and other arrays shouldn't be affected.
func TestAllowEmptyArray(t *testing.T) {
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
		resources: []resourceInfo{
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
