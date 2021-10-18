package genesyscloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
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

func setToStringList(strSet *schema.Set) *[]string {
	interfaceList := strSet.List()
	strList := make([]string, len(interfaceList))
	for i, s := range interfaceList {
		strList[i] = s.(string)
	}
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
