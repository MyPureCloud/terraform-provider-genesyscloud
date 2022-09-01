package genesyscloud

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func subStringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.Contains(b, a) {
			return true
		}
	}
	return false
}

// difference returns the elements in a that aren't in b
func sliceDifference(a, b []string) []string {
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

func stringListToSet(list []string) *schema.Set {
	interfaceList := make([]interface{}, len(list))
	for i, v := range list {
		interfaceList[i] = v
	}
	return schema.NewSet(schema.HashString, interfaceList)
}

func stringListToInterfaceList(list []string) []interface{} {
	interfaceList := make([]interface{}, len(list))
	for i, v := range list {
		interfaceList[i] = v
	}
	return interfaceList
}

func setToStringList(strSet *schema.Set) *[]string {
	interfaceList := strSet.List()
	strList := interfaceListToStrings(interfaceList)
	return &strList
}

func interfaceListToStrings(interfaceList []interface{}) []string {
	strs := make([]string, len(interfaceList))
	for i, val := range interfaceList {
		strs[i] = val.(string)
	}
	return strs
}

func buildSdkStringList(d *schema.ResourceData, attrName string) *[]string {
	if val, ok := d.GetOk(attrName); ok {
		return setToStringList(val.(*schema.Set))
	}
	return nil
}

func buildSdkStringListFromMapEntry(d map[string]interface{}, attrName string) *[]string {
	child := d[attrName]
	if child != nil {
		return setToStringList(child.(*schema.Set))
	}
	return nil
}

func buildSdkGenericListFirstElement[T interface{}](d *schema.ResourceData, attrName string, elementBuilder func(map[string]interface{}) *T) *T {
	child := d.Get(attrName).(*schema.Set).List()
	if len(child) > 0 {
		return elementBuilder(child[0].(map[string]interface{}))
	}
	return elementBuilder(nil)
}

func buildSdkGenericList[T interface{}](d map[string]interface{}, attrName string, elementBuilder func(map[string]interface{}) *T) *[]T {
	child := d[attrName]
	if child != nil {
		list := child.(*schema.Set).List()
		sdkList := make([]T, len(list))
		for i, element := range list {
			sdkList[i] = *elementBuilder(element.(map[string]interface{}))
		}
		return &sdkList
	}
	return nil
}

func flattenGenericList[T interface{}](resourceList *[]T, elementFlattener func(resource *T) map[string]interface{}) []map[string]interface{} {
	if resourceList == nil {
		return nil
	}

	var resultList []map[string]interface{}

	for _, resource := range *resourceList {
		resultList = append(resultList, elementFlattener(&resource))
	}
	return resultList
}

func flattenGenericAsList[T interface{}](resource *T, elementFlattener func(resource *T) map[string]interface{}) *[]map[string]interface{} {
	if resource == nil {
		return nil
	}

	flattened := elementFlattener(resource)
	if flattened != nil {
		return &[]map[string]interface{}{flattened}
	}

	return nil
}
