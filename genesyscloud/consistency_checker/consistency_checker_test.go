package consistency_checker

import (
	"testing"

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
