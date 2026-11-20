package task_management_worktype_flow_onattributechange_rule

import (
	"testing"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const statusRefType = "genesyscloud_task_management_worktype_status"

var conditionValueAttrs = []string{"condition.new_value", "condition.old_value"}

// The rule -> worktype_status dependency must be discoverable from static exporter metadata.
func TestUnitExporterDeclaresStatusDependencyStatically(t *testing.T) {
	exporter := TaskManagementOnAttributeChangeRuleExporter()

	for _, attr := range conditionValueAttrs {
		t.Run(attr, func(t *testing.T) {
			decls := exporter.GetConditionalRefAttrs(attr)
			require.Len(t, decls, 1)
			assert.Equal(t, statusRefType, decls[0].RefType)
			assert.Equal(t, "attribute", decls[0].SiblingAttr)
			assert.Equal(t, "statusId", decls[0].SiblingValue)
		})
	}
}

// Unconditional RefAttrs take precedence over ResolveRefTypeFunc at export time, so the condition values
// must not be declared there.
func TestUnitExporterRefAttrsUnchanged(t *testing.T) {
	exporter := TaskManagementOnAttributeChangeRuleExporter()

	assert.Equal(t, map[string]*resourceExporter.RefAttrSettings{
		"worktype_id": {RefType: "genesyscloud_task_management_worktype"},
	}, exporter.RefAttrs)
}

func TestUnitExporterConditionValueResolution(t *testing.T) {
	exporter := TaskManagementOnAttributeChangeRuleExporter()

	tests := []struct {
		name      string
		configMap map[string]interface{}
		want      string
	}{
		{"statusId resolves to worktype_status", map[string]interface{}{"attribute": "statusId", "new_value": "abc"}, statusRefType},
		{"other attribute is not a reference", map[string]interface{}{"attribute": "priority", "new_value": "abc"}, ""},
		{"case differs is not a reference", map[string]interface{}{"attribute": "StatusId"}, ""},
		{"missing attribute", map[string]interface{}{"new_value": "abc"}, ""},
		{"non-string attribute", map[string]interface{}{"attribute": 5}, ""},
		{"nil attribute", map[string]interface{}{"attribute": nil}, ""},
		{"nil config map", nil, ""},
	}

	for _, attr := range conditionValueAttrs {
		resolver := exporter.CustomAttributeResolver[attr]
		require.NotNil(t, resolver, attr)
		require.NotNil(t, resolver.ResolveRefTypeFunc, attr)

		for _, tt := range tests {
			t.Run(attr+"/"+tt.name, func(t *testing.T) {
				got, err := resolver.ResolveRefTypeFunc(tt.configMap)
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.want, exporter.ResolveConditionalRefType(attr, tt.configMap))
			})
		}
	}
}

func TestUnitExporterUndeclaredAttributeHasNoConditionalRef(t *testing.T) {
	exporter := TaskManagementOnAttributeChangeRuleExporter()

	assert.Nil(t, exporter.GetConditionalRefAttrs("condition.attribute"))
	assert.Empty(t, exporter.ResolveConditionalRefType("condition.attribute", map[string]interface{}{"attribute": "statusId"}))
}
