package lists

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ItemInSlice[T comparable](a T, list []T) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func RemoveStringFromSlice(value string, slice []string) []string {
	s := make([]string, 0)
	for _, v := range slice {
		if v != value {
			s = append(s, v)
		}
	}
	return s
}

func SubStringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.Contains(b, a) {
			return true
		}
	}
	return false
}

// difference returns the elements in a that aren't in b
func SliceDifference(a, b []string) []string {
	var diff []string
	if len(a) == 0 {
		return diff
	}
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

// Returns true if a and b are equivalent, ignoring the ordering of the items.
func ListsAreEquivalent(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, aItem := range a {
		matchFound := false
		for _, bItem := range b {
			if bItem == aItem {
				matchFound = true
				break
			}
		}
		if !matchFound {
			return false
		}
	}
	return true
}

func StringListToSet(list []string) *schema.Set {
	interfaceList := make([]interface{}, len(list))
	for i, v := range list {
		interfaceList[i] = v
	}
	return schema.NewSet(schema.HashString, interfaceList)
}

func StringListToSetOrNil(list *[]string) *schema.Set {
	if list == nil {
		return nil
	}
	return StringListToSet(*list)
}

func StringListToInterfaceList(list []string) []interface{} {
	interfaceList := make([]interface{}, len(list))
	for i, v := range list {
		interfaceList[i] = v
	}
	return interfaceList
}

func SetToStringList(strSet *schema.Set) *[]string {
	interfaceList := strSet.List()
	strList := InterfaceListToStrings(interfaceList)
	return &strList
}

func InterfaceListToStrings(interfaceList []interface{}) []string {
	strs := make([]string, len(interfaceList))
	for i, val := range interfaceList {
		strs[i] = val.(string)
	}
	return strs
}

func BuildSdkStringList(d *schema.ResourceData, attrName string) *[]string {
	if val, ok := d.GetOk(attrName); ok {
		return SetToStringList(val.(*schema.Set))
	}
	return nil
}

func BuildSdkStringListFromInterfaceArray(d *schema.ResourceData, attrName string) *[]string {
	var stringArray []string
	if val, ok := d.GetOk(attrName); ok {
		if valArray, ok := val.([]interface{}); ok {
			stringArray = InterfaceListToStrings(valArray)
		}
	}
	return &stringArray
}

func FlattenList[T interface{}](resourceList *[]T, elementFlattener func(resource *T) map[string]interface{}) *[]map[string]interface{} {
	if resourceList == nil {
		return nil
	}

	var resultList []map[string]interface{}

	for _, resource := range *resourceList {
		resultList = append(resultList, elementFlattener(&resource))
	}
	return &resultList
}

func FlattenAsList[T interface{}](resource *T, elementFlattener func(resource *T) map[string]interface{}) *[]map[string]interface{} {
	if resource == nil {
		return nil
	}

	flattened := elementFlattener(resource)
	if flattened != nil {
		return &[]map[string]interface{}{flattened}
	}

	return nil
}

func NilToEmptyList[T interface{}](list *[]T) *[]T {
	if list == nil {
		emptyArray := []T{}
		return &emptyArray
	}
	return list
}
