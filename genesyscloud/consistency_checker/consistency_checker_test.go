package consistency_checker

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestFilterMapSlice_skipsNilNestedBlocks(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"key":   "userName",
			"value": "user@example.com",
		},
		nil,
	}

	result := filterMapSlice(input).([]interface{})

	assert.Len(t, result, 1)
	assert.Equal(t, "user@example.com", result[0].(map[string]interface{})["value"])
}

func TestFilterMapSlice_nilFirstElementWithMaps(t *testing.T) {
	input := []interface{}{
		nil,
		map[string]interface{}{
			"error_code": "TEST",
		},
	}

	result := filterMapSlice(input).([]interface{})

	assert.Len(t, result, 1)
	assert.Equal(t, "TEST", result[0].(map[string]interface{})["error_code"])
}

func TestCompareValuesAtPath_supportsArbitrarilyNestedLists(t *testing.T) {
	state := []interface{}{
		map[string]interface{}{
			"tools": []interface{}{
				map[string]interface{}{
					"inputs": []interface{}{
						map[string]interface{}{
							"required": true,
						},
					},
				},
			},
		},
	}
	path := []string{"0", "tools", "0", "inputs", "0", "required"}

	assert.True(t, compareValuesAtPath(state, true, path))
	assert.False(t, compareValuesAtPath(state, false, path))
}

func TestCheckState_ignoredPathDoesNotRetry(t *testing.T) {
	resourceSchema := map[string]*schema.Schema{
		"definition": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"tools": {
					Type:     schema.TypeList,
					Required: true,
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{
						"context_retention_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					}},
				},
			}},
		},
	}

	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
		"definition": []interface{}{
			map[string]interface{}{
				"tools": []interface{}{
					map[string]interface{}{"context_retention_enabled": true},
				},
			},
		},
	})
	d.SetId("consistency-check-ignored-path")

	checker := NewConsistencyCheck(
		context.Background(),
		d,
		nil,
		&schema.Resource{Schema: resourceSchema},
		3,
		"test_resource",
	)

	err := d.Set("definition", []interface{}{
		map[string]interface{}{
			"tools": []interface{}{
				map[string]interface{}{"context_retention_enabled": false},
			},
		},
	})
	assert.NoError(t, err)
	assert.Nil(t, checker.CheckState(d, "definition.0.tools.0.context_retention_enabled"))
}
