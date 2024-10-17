package consistency_checker

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unsafe"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDataToMap(d *schema.ResourceData) map[string]interface{} {
	schemaInterface := getUnexportedField(reflect.ValueOf(d).Elem().FieldByName("schema"))
	resourceSchema := schemaInterface.(map[string]*schema.Schema)

	stateMap := make(map[string]interface{})
	for k := range resourceSchema {
		stateMap[k] = d.Get(k)
	}
	return filterMap(stateMap)
}

func getUnexportedField(field reflect.Value) interface{} {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
}

func filterMap(m map[string]interface{}) map[string]interface{} {
	newM := make(map[string]interface{})
	for k, v := range m {
		switch t := v.(type) {
		case *schema.Set:
			newM[k] = filterMapSlice(t.List())
		case []interface{}:
			newM[k] = filterMapSlice(t)
		case map[string]interface{}:
			if len(t) > 0 {
				newM[k] = filterMap(t)
			}
		default:
			newM[k] = v
		}
	}

	return newM
}

func filterMapSlice(unfilteredSlice []interface{}) interface{} {
	if len(unfilteredSlice) == 0 {
		return unfilteredSlice
	}

	switch unfilteredSlice[0].(type) {
	case map[string]interface{}:
		filteredSlice := make([]interface{}, 0)
		for _, s := range unfilteredSlice {
			filteredSlice = append(filteredSlice, filterMap(s.(map[string]interface{})))
		}
		return filteredSlice
	}

	return unfilteredSlice
}

func isEmptyState(d *schema.ResourceData) *bool {
	stateString := strings.Split(d.State().String(), "\n")
	isEmpty := true
	for _, s := range stateString {
		if len(s) == 0 {
			continue
		}
		sSplit := strings.Split(s, " ")
		attribute := sSplit[0]
		if attribute == "ID" ||
			attribute == "Tainted" ||
			strings.HasSuffix(s, ".# = 0") ||
			strings.HasSuffix(attribute, "id") {
			continue
		}
		isEmpty = false
		break
	}
	return &isEmpty
}

// populateComputedProperties will find any properties in a resource marked as computed
func populateComputedProperties(resource *schema.Resource, attributes, blocks *[]string, parent string) {
	for attr, attrSchema := range resource.Schema {
		var fullName string
		if parent == "" {
			fullName = attr
		} else {
			fullName = parent + "." + attr
		}

		if attrSchema.Type == schema.TypeSet || attrSchema.Type == schema.TypeList {
			if attrSchema.Computed == true {
				*blocks = append(*blocks, fullName)
			}

			switch elem := attrSchema.Elem.(type) {
			case *schema.Resource:
				populateComputedProperties(elem, attributes, blocks, fullName)
			default:
				continue
			}
		} else {
			if attrSchema.Computed == true {
				*attributes = append(*attributes, fullName)
			}
		}
	}
}

// compareStateMaps compares two state maps and returns true if they are equal, accounting for re-ordering blocks
func compareStateMaps(originalState, currentState map[string]interface{}) bool {
	return compareValues(originalState, currentState)
}

func compareValues(v1, v2 interface{}) bool {
	if reflect.DeepEqual(v1, v2) {
		return true
	}

	switch v1 := v1.(type) {
	case map[string]interface{}:
		if v2, ok := v2.(map[string]interface{}); ok {
			return compareMaps(v1, v2)
		}
		return false
	case []interface{}:
		if v2, ok := v2.([]interface{}); ok {
			return compareSlices(v1, v2)
		}
		return false
	default:
		return false
	}
}

// compareMaps compares two maps recursively.
func compareMaps(map1, map2 map[string]interface{}) bool {
	if len(map1) != len(map2) {
		return false
	}
	for k, v1 := range map1 {
		v2, ok := map2[k]
		if !ok {
			return false
		}
		if !compareValues(v1, v2) {
			return false
		}
	}
	return true
}

// compareSlices compares two slices regardless of order.
func compareSlices(slice1, slice2 []interface{}) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	sortedSlice1 := make([]interface{}, len(slice1))
	sortedSlice2 := make([]interface{}, len(slice2))

	copy(sortedSlice1, slice1)
	copy(sortedSlice2, slice2)

	sort.Slice(sortedSlice1, func(i, j int) bool {
		return fmt.Sprintf("%v", sortedSlice1[i]) < fmt.Sprintf("%v", sortedSlice1[j])
	})
	sort.Slice(sortedSlice2, func(i, j int) bool {
		return fmt.Sprintf("%v", sortedSlice2[i]) < fmt.Sprintf("%v", sortedSlice2[j])
	})

	return reflect.DeepEqual(sortedSlice1, sortedSlice2)
}
