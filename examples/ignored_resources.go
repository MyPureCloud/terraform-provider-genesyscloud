package examples

import (
	lists "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
)

// GetIgnoredResources returns a list of resources that should be ignored when generating the api docs. These are resource types that do not expect to have examples, and should not be included in public documentation
func GetIgnoredResources() []string {
	return []string{
		"genesyscloud_bcp_tf_exporter",
		"genesyscloud_externalusers_identity",
	}
}

func RemoveIgnoredResources(resources []string) []string {
	ignoredResources := GetIgnoredResources()
	for _, ignoredResource := range ignoredResources {
		resources = lists.RemoveStringFromSlice(ignoredResource, resources)
	}
	return resources
}
