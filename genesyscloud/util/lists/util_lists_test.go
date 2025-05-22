package lists

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAreEquivalent(t *testing.T) {
	testCases := []struct {
		name       string
		arrayA     []string
		arrayB     []string
		equivalent bool
	}{
		{
			name:       "Empty string lists",
			arrayA:     []string{},
			arrayB:     []string{},
			equivalent: true,
		},
		{
			name:       "Equivalent lists not in same order",
			arrayA:     []string{"foo", "bar"},
			arrayB:     []string{"bar", "foo"},
			equivalent: true,
		},
		{
			name:       "Equivalent longer lists not in same order",
			arrayA:     []string{"y", "x", "foo", "bar"},
			arrayB:     []string{"x", "bar", "foo", "y"},
			equivalent: true,
		},
		{
			name:       "Lists of unequal length with same content",
			arrayA:     []string{"x", "x", "x"},
			arrayB:     []string{"x", "x"},
			equivalent: false,
		},
		{
			name:       "Lists of equal length with different content",
			arrayA:     []string{"x", "x"},
			arrayB:     []string{"x", "y"},
			equivalent: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			arrayACopy := make([]string, len(tc.arrayA))
			arrayBCopy := make([]string, len(tc.arrayB))
			copy(arrayACopy, tc.arrayA)
			copy(arrayBCopy, tc.arrayB)

			result := AreEquivalent(tc.arrayA, tc.arrayB)
			if result != tc.equivalent {
				t.Errorf("Got %v for lists %v and %v. Should have got %v.", result, tc.arrayA, tc.arrayB, tc.equivalent)
			}

			// Ensure the sort function hasn't manipulated the arrays that were passed into AreEquivalent()
			if len(tc.arrayA) != len(arrayACopy) {
				t.Errorf("arrayA has changed after going through the function. Should be %v, got %v", arrayACopy, tc.arrayA)
			}
			if len(tc.arrayB) != len(arrayBCopy) {
				t.Errorf("arrayB has changed after going through the function. Should be %v, got %v", arrayBCopy, tc.arrayB)
			}
			for k, v := range tc.arrayA {
				if v != arrayACopy[k] {
					t.Errorf("arrayA has changed after going through the function. Should be %v, got %v", arrayACopy, tc.arrayA)
				}
			}
			for k, v := range tc.arrayB {
				if v != arrayBCopy[k] {
					t.Errorf("arrayB has changed after going through the function. Should be %v, got %v", arrayBCopy, tc.arrayB)
				}
			}
		})
	}
}

func TestRemove(t *testing.T) {
	testCases := []struct {
		name          string
		originalSlice []string
		itemToRemove  string
		expectedSlice []string
	}{
		{
			name:          "Empty slice",
			originalSlice: []string{},
			itemToRemove:  "test",
			expectedSlice: []string{},
		},
		{
			name:          "Item in middle of slice",
			originalSlice: []string{"a", "b", "c"},
			itemToRemove:  "b",
			expectedSlice: []string{"a", "c"},
		},
		{
			name:          "Item at start of slice",
			originalSlice: []string{"a", "b", "c"},
			itemToRemove:  "a",
			expectedSlice: []string{"b", "c"},
		},
		{
			name:          "Item at end of slice",
			originalSlice: []string{"a", "b", "c"},
			itemToRemove:  "c",
			expectedSlice: []string{"a", "b"},
		},
		{
			name:          "Item not in slice",
			originalSlice: []string{"test1", "test2", "test3"},
			itemToRemove:  "test4",
			expectedSlice: []string{"test1", "test2", "test3"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			res := Remove(testCase.originalSlice, testCase.itemToRemove)
			if !AreEquivalent(res, testCase.expectedSlice) {
				t.Errorf("expected %v, got %v", testCase.expectedSlice, res)
			}
		})
	}
}

func TestItemInSlice(t *testing.T) {
	tests := []struct {
		name     string
		item     string
		slice    []string
		expected bool
	}{
		{
			name:     "Empty slice",
			item:     "test",
			slice:    []string{},
			expected: false,
		},
		{
			name:     "Item in slice",
			item:     "test",
			slice:    []string{"test", "other"},
			expected: true,
		},
		{
			name:     "Item not in slice",
			item:     "test",
			slice:    []string{"other", "another"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ItemInSlice(tt.item, tt.slice)
			if result != tt.expected {
				t.Errorf("ItemInSlice() = %v, want %v", result, tt.expected)
			}
		})
	}

	// Test with integers
	intTests := []struct {
		name     string
		item     int
		slice    []int
		expected bool
	}{
		{
			name:     "Empty int slice",
			item:     1,
			slice:    []int{},
			expected: false,
		},
		{
			name:     "Int in slice",
			item:     1,
			slice:    []int{1, 2, 3},
			expected: true,
		},
		{
			name:     "Int not in slice",
			item:     4,
			slice:    []int{1, 2, 3},
			expected: false,
		},
	}

	for _, tt := range intTests {
		t.Run(tt.name, func(t *testing.T) {
			result := ItemInSlice(tt.item, tt.slice)
			if result != tt.expected {
				t.Errorf("ItemInSlice() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRemoveStringFromSlice(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		slice    []string
		expected []string
	}{
		{
			name:     "Empty slice",
			value:    "test",
			slice:    []string{},
			expected: []string{},
		},
		{
			name:     "Value in slice",
			value:    "test",
			slice:    []string{"test", "other"},
			expected: []string{"other"},
		},
		{
			name:     "Value not in slice",
			value:    "test",
			slice:    []string{"other", "another"},
			expected: []string{"other", "another"},
		},
		{
			name:     "Multiple occurrences",
			value:    "test",
			slice:    []string{"test", "other", "test"},
			expected: []string{"other"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveStringFromSlice(tt.value, tt.slice)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("RemoveStringFromSlice() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSubStringInSlice(t *testing.T) {
	tests := []struct {
		name     string
		substr   string
		slice    []string
		expected bool
	}{
		{
			name:     "Empty slice",
			substr:   "test",
			slice:    []string{},
			expected: false,
		},
		{
			name:     "Substring in slice",
			substr:   "est",
			slice:    []string{"test", "other"},
			expected: true,
		},
		{
			name:     "Substring not in slice",
			substr:   "xyz",
			slice:    []string{"test", "other"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SubStringInSlice(tt.substr, tt.slice)
			if result != tt.expected {
				t.Errorf("SubStringInSlice() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContainsAnySubStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		slice    []string
		expected bool
	}{
		{
			name:     "Empty slice",
			str:      "test",
			slice:    []string{},
			expected: false,
		},
		{
			name:     "String contains substring",
			str:      "testing",
			slice:    []string{"est", "xyz"},
			expected: true,
		},
		{
			name:     "String doesn't contain substring",
			str:      "testing",
			slice:    []string{"abc", "xyz"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsAnySubStringSlice(tt.str, tt.slice)
			if result != tt.expected {
				t.Errorf("ContainsAnySubStringSlice() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSliceDifference(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected []string
	}{
		{
			name:     "Both empty",
			a:        []string{},
			b:        []string{},
			expected: nil,
		},
		{
			name:     "A empty",
			a:        []string{},
			b:        []string{"test"},
			expected: nil,
		},
		{
			name:     "B empty",
			a:        []string{"test"},
			b:        []string{},
			expected: []string{"test"},
		},
		{
			name:     "No common elements",
			a:        []string{"test1", "test2"},
			b:        []string{"test3", "test4"},
			expected: []string{"test1", "test2"},
		},
		{
			name:     "Some common elements",
			a:        []string{"test1", "test2", "test3"},
			b:        []string{"test2", "test3", "test4"},
			expected: []string{"test1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceDifference(tt.a, tt.b)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("SliceDifference() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStringListToSet(t *testing.T) {
	tests := []struct {
		name  string
		list  []string
		empty bool
	}{
		{
			name:  "Empty list",
			list:  []string{},
			empty: true,
		},
		{
			name:  "Single item",
			list:  []string{"test"},
			empty: false,
		},
		{
			name:  "Multiple items",
			list:  []string{"test1", "test2"},
			empty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringListToSet(tt.list)
			if tt.empty && result.Len() != 0 {
				t.Errorf("StringListToSet() expected empty set, got %v", result)
			}
			if !tt.empty {
				for _, item := range tt.list {
					if !result.Contains(item) {
						t.Errorf("StringListToSet() result doesn't contain %s", item)
					}
				}
			}
		})
	}
}

func TestStringListToSetOrNil(t *testing.T) {
	tests := []struct {
		name     string
		list     *[]string
		expected bool // true if result should be nil
	}{
		{
			name:     "Nil list",
			list:     nil,
			expected: true,
		},
		{
			name:     "Empty list",
			list:     &[]string{},
			expected: false,
		},
		{
			name:     "Non-empty list",
			list:     &[]string{"test"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringListToSetOrNil(tt.list)
			if tt.expected && result != nil {
				t.Errorf("StringListToSetOrNil() expected nil, got %v", result)
			}
			if !tt.expected && result == nil {
				t.Errorf("StringListToSetOrNil() expected non-nil, got nil")
			}
		})
	}
}

func TestStringListToInterfaceList(t *testing.T) {
	tests := []struct {
		name     string
		list     []string
		expected int // expected length
	}{
		{
			name:     "Empty list",
			list:     []string{},
			expected: 0,
		},
		{
			name:     "Single item",
			list:     []string{"test"},
			expected: 1,
		},
		{
			name:     "Multiple items",
			list:     []string{"test1", "test2"},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringListToInterfaceList(tt.list)
			if len(result) != tt.expected {
				t.Errorf("StringListToInterfaceList() expected length %d, got %d", tt.expected, len(result))
			}
			for i, item := range result {
				if item.(string) != tt.list[i] {
					t.Errorf("StringListToInterfaceList() item at index %d = %v, want %v", i, item, tt.list[i])
				}
			}
		})
	}
}

func TestSetToStringList(t *testing.T) {
	tests := []struct {
		name     string
		set      *schema.Set
		expected *[]string
	}{
		{
			name:     "Nil set",
			set:      nil,
			expected: nil,
		},
		{
			name:     "Empty set",
			set:      schema.NewSet(schema.HashString, []interface{}{}),
			expected: &[]string{},
		},
		{
			name:     "Set with values",
			set:      schema.NewSet(schema.HashString, []interface{}{"test1", "test2"}),
			expected: &[]string{"test1", "test2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SetToStringList(tt.set)
			if tt.expected == nil && result != nil {
				t.Errorf("SetToStringList() expected nil, got %v", result)
				return
			}
			if tt.expected != nil && result == nil {
				t.Errorf("SetToStringList() expected non-nil, got nil")
				return
			}
			if tt.expected != nil && result != nil {
				if len(*result) != len(*tt.expected) {
					t.Errorf("SetToStringList() expected length %d, got %d", len(*tt.expected), len(*result))
				}
			}
		})
	}
}

func TestInterfaceListToStrings(t *testing.T) {
	tests := []struct {
		name          string
		interfaceList []interface{}
		expected      []string
	}{
		{
			name:          "Empty list",
			interfaceList: []interface{}{},
			expected:      []string{},
		},
		{
			name:          "String values",
			interfaceList: []interface{}{"test1", "test2"},
			expected:      []string{"test1", "test2"},
		},
		{
			name:          "Mixed values",
			interfaceList: []interface{}{"test1", 123, nil, true},
			expected:      []string{"test1", "", "", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InterfaceListToStrings(tt.interfaceList)
			if len(result) != len(tt.expected) {
				t.Errorf("InterfaceListToStrings() expected length %d, got %d", len(tt.expected), len(result))
				return
			}
			for i, val := range result {
				if val != tt.expected[i] {
					t.Errorf("InterfaceListToStrings() at index %d = %v, want %v", i, val, tt.expected[i])
				}
			}
		})
	}
}

func TestBuildStringListFromSetInMap(t *testing.T) {
	testSet := schema.NewSet(schema.HashString, []interface{}{"test1", "test2"})
	emptySet := schema.NewSet(schema.HashString, []interface{}{})

	tests := []struct {
		name     string
		mapValue map[string]any
		key      string
		expected []string
	}{
		{
			name:     "Nil map",
			mapValue: nil,
			key:      "key",
			expected: nil,
		},
		{
			name:     "Key not in map",
			mapValue: map[string]any{"otherKey": testSet},
			key:      "key",
			expected: nil,
		},
		{
			name:     "Nil value in map",
			mapValue: map[string]any{"key": nil},
			key:      "key",
			expected: nil,
		},
		{
			name:     "Non-set value in map",
			mapValue: map[string]any{"key": "not a set"},
			key:      "key",
			expected: nil,
		},
		{
			name:     "Empty set in map",
			mapValue: map[string]any{"key": emptySet},
			key:      "key",
			expected: nil,
		},
		{
			name:     "Set with values",
			mapValue: map[string]any{"key": testSet},
			key:      "key",
			expected: []string{"test1", "test2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildStringListFromSetInMap(tt.mapValue, tt.key)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("BuildStringListFromSetInMap() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildSdkStringList(t *testing.T) {

	tests := []struct {
		name     string
		d        *schema.ResourceData
		attrName string
		expected *[]string
	}{
		{
			name:     "Nil ResourceData",
			d:        nil,
			attrName: "attr",
			expected: nil,
		},
		{
			name:     "Attribute not found",
			d:        &schema.ResourceData{},
			attrName: "attr",
			expected: nil,
		},
		{
			name:     "Attribute not a set",
			d:        &schema.ResourceData{},
			attrName: "attr",
			expected: nil,
		},
		{
			name: "Attribute in set",
			d: func() *schema.ResourceData {
				schemaMap := map[string]*schema.Schema{
					"attr": {
						Type:     schema.TypeSet,
						Required: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
				}

				dataMap := map[string]interface{}{
					"attr": []interface{}{"test1", "test2"},
				}
				d := schema.TestResourceDataRaw(t, schemaMap, dataMap)
				return d
			}(),
			attrName: "attr",
			expected: &[]string{"test1", "test2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildSdkStringList(tt.d, tt.attrName)
			if tt.expected == nil && result != nil {
				t.Errorf("BuildSdkStringList() expected nil, got %v", result)
			}
			if tt.expected != nil && result == nil {
				t.Errorf("BuildSdkStringList() expected non-nil, got nil")
			}
		})
	}
}

func TestNilToEmptyList(t *testing.T) {
	var nilList *[]string
	nonNilList := &[]string{"test"}
	emptyList := []string{}

	tests := []struct {
		name     string
		list     *[]string
		expected *[]string
	}{
		{
			name:     "Nil",
			list:     nil,
			expected: &emptyList,
		},
		{
			name:     "Nil list",
			list:     nilList,
			expected: &emptyList,
		},
		{
			name:     "Non-nil list",
			list:     nonNilList,
			expected: nonNilList,
		},
		{
			name:     "Empty list",
			list:     &emptyList,
			expected: &emptyList,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NilToEmptyList(tt.list)
			if result == nil {
				t.Errorf("NilToEmptyList() returned nil")
			}
			if !reflect.DeepEqual(*tt.expected, *result) {
				t.Errorf("NilToEmptyList() expected %v, got %v", tt.expected, *result)
			}
		})
	}
}

func TestConvertMapStringAnyToMapStringString(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]any
		expected map[string]string
	}{
		{
			name:     "Nil map",
			m:        nil,
			expected: nil,
		},
		{
			name:     "Empty map",
			m:        map[string]any{},
			expected: map[string]string{},
		},
		{
			name:     "String values",
			m:        map[string]any{"key1": "value1", "key2": "value2"},
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:     "Mixed values",
			m:        map[string]any{"key1": "value1", "key2": 123, "key3": nil},
			expected: map[string]string{"key1": "value1", "key2": "", "key3": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertMapStringAnyToMapStringString(tt.m)
			if tt.expected == nil && result != nil {
				t.Errorf("ConvertMapStringAnyToMapStringString() expected nil, got %v", result)
				return
			}
			if tt.expected != nil && result == nil {
				t.Errorf("ConvertMapStringAnyToMapStringString() expected non-nil, got nil")
				return
			}
			if tt.expected != nil && result != nil {
				if len(result) != len(tt.expected) {
					t.Errorf("ConvertMapStringAnyToMapStringString() expected length %d, got %d", len(tt.expected), len(result))
					return
				}
				for k, v := range tt.expected {
					if result[k] != v {
						t.Errorf("ConvertMapStringAnyToMapStringString() key %s = %v, want %v", k, result[k], v)
					}
				}
			}
		})
	}
}

func TestMap(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		fn       func(int) string
		expected []string
	}{
		{
			name:     "Nil function",
			slice:    []int{1, 2, 3},
			fn:       nil,
			expected: nil,
		},
		{
			name:     "Empty slice",
			slice:    []int{},
			fn:       func(i int) string { return "test" },
			expected: []string{},
		},
		{
			name:     "Normal case",
			slice:    []int{1, 2, 3},
			fn:       func(i int) string { return "test" + string(rune(i+'0')) },
			expected: []string{"test1", "test2", "test3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Map(tt.slice, tt.fn)
			if tt.expected == nil && result != nil {
				t.Errorf("Map() expected nil, got %v", result)
				return
			}
			if tt.expected != nil && result == nil {
				t.Errorf("Map() expected non-nil, got nil")
				return
			}
			if tt.expected != nil && result != nil {
				if len(result) != len(tt.expected) {
					t.Errorf("Map() expected length %d, got %d", len(tt.expected), len(result))
					return
				}
				for i, v := range tt.expected {
					if result[i] != v {
						t.Errorf("Map() index %d = %v, want %v", i, result[i], v)
					}
				}
			}
		})
	}
}
