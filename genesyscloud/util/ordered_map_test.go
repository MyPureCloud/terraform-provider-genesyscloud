package util

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderedMap_UnmarshalJSON_PreservesOrder(t *testing.T) {
	jsonStr := `{"zebra":"z","apple":"a","mango":"m","banana":"b"}`

	var om OrderedMap[string]
	err := json.Unmarshal([]byte(jsonStr), &om)
	assert.NoError(t, err)
	assert.Equal(t, []string{"zebra", "apple", "mango", "banana"}, om.Keys())
	assert.Equal(t, 4, om.Len())

	val, ok := om.Get("apple")
	assert.True(t, ok)
	assert.Equal(t, "a", val)
}

func TestOrderedMap_MarshalJSON_PreservesOrder(t *testing.T) {
	om := NewOrderedMap[string]()
	om.Set("zebra", "z")
	om.Set("apple", "a")
	om.Set("mango", "m")

	data, err := json.Marshal(om)
	assert.NoError(t, err)
	assert.Equal(t, `{"zebra":"z","apple":"a","mango":"m"}`, string(data))
}

func TestOrderedMap_RoundTrip(t *testing.T) {
	jsonStr := `{"third":"3","first":"1","second":"2"}`

	var om OrderedMap[string]
	err := json.Unmarshal([]byte(jsonStr), &om)
	assert.NoError(t, err)

	data, err := json.Marshal(&om)
	assert.NoError(t, err)
	assert.Equal(t, jsonStr, string(data))
}

func TestOrderedMap_SetOverwrite(t *testing.T) {
	om := NewOrderedMap[string]()
	om.Set("a", "1")
	om.Set("b", "2")
	om.Set("a", "3")

	assert.Equal(t, []string{"a", "b"}, om.Keys())
	val, _ := om.Get("a")
	assert.Equal(t, "3", val)
}

func TestOrderedMap_Delete(t *testing.T) {
	om := NewOrderedMap[string]()
	om.Set("a", "1")
	om.Set("b", "2")
	om.Set("c", "3")
	om.Delete("b")

	assert.Equal(t, []string{"a", "c"}, om.Keys())
	assert.Equal(t, 2, om.Len())
	_, ok := om.Get("b")
	assert.False(t, ok)
}

func TestOrderedMap_DeleteNonExistent(t *testing.T) {
	om := NewOrderedMap[string]()
	om.Set("a", "1")
	om.Delete("z")
	assert.Equal(t, 1, om.Len())
}

type testStruct struct {
	Name  *string `json:"name,omitempty"`
	Value *int    `json:"value,omitempty"`
}

func TestOrderedMap_UnmarshalJSON_StructValues(t *testing.T) {
	jsonStr := `{"second":{"name":"b","value":2},"first":{"name":"a","value":1}}`

	var om OrderedMap[testStruct]
	err := json.Unmarshal([]byte(jsonStr), &om)
	assert.NoError(t, err)
	assert.Equal(t, []string{"second", "first"}, om.Keys())

	val, ok := om.Get("first")
	assert.True(t, ok)
	assert.Equal(t, "a", *val.Name)
	assert.Equal(t, 1, *val.Value)
}

func TestOrderedMap_EmbeddedInStruct(t *testing.T) {
	type wrapper struct {
		Props *OrderedMap[string] `json:"props,omitempty"`
	}

	jsonStr := `{"props":{"z":"1","a":"2"}}`
	var w wrapper
	err := json.Unmarshal([]byte(jsonStr), &w)
	assert.NoError(t, err)
	assert.Equal(t, []string{"z", "a"}, w.Props.Keys())

	data, err := json.Marshal(&w)
	assert.NoError(t, err)
	assert.Equal(t, jsonStr, string(data))
}
