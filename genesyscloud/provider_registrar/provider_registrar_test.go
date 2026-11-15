package provider_registrar

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// The tfstate exporter only marks top-level Sensitive attributes in sensitive_attributes (see addSensitiveAttributesToState).
func TestUnitNoNestedSensitiveAttributes(t *testing.T) {
	resources, _ := GetProviderResources()

	var check func(path string, attrs map[string]*schema.Schema)
	check = func(path string, attrs map[string]*schema.Schema) {
		for name, attr := range attrs {
			if r, ok := attr.Elem.(*schema.Resource); ok {
				for nestedName, nested := range r.Schema {
					if nested.Sensitive {
						t.Errorf("%s.%s.%s is a nested sensitive attribute; addSensitiveAttributesToState must be extended to mark it", path, name, nestedName)
					}
				}
				check(path+"."+name, r.Schema)
			}
		}
	}
	for resourceType, r := range resources {
		check(resourceType, r.Schema)
	}
}
