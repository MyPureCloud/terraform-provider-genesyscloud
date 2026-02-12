package examples

import (
	lists "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
)

func GetIgnoredResources() []string {
	return []string{
		"genesyscloud_bcp_tf_exporter",
	}
}

func RemoveIgnoredResources(resources []string) []string {
	ignoredResources := GetIgnoredResources()
	for _, ignoredResource := range ignoredResources {
		resources = lists.RemoveStringFromSlice(ignoredResource, resources)
	}
	return resources
}
