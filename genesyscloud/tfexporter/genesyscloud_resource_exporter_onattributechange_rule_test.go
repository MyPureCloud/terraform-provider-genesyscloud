package tfexporter

import (
	"testing"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	onAttributeChangeRule "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype_flow_onattributechange_rule"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Export-time resolution of condition.new_value / condition.old_value must only treat the values as worktype
// status references when condition.attribute == "statusId", using the real rule exporter configuration.
func TestUnitSanitizeConfigMapOnAttributeChangeRuleConditionValues(t *testing.T) {
	const (
		statusType = "genesyscloud_task_management_worktype_status"
		statusID   = "11111111-1111-1111-1111-111111111111"
	)

	ruleExporter := onAttributeChangeRule.TaskManagementOnAttributeChangeRuleExporter()
	statusExporter := &resourceExporter.ResourceExporter{
		SanitizedResourceMap: resourceExporter.ResourceIDMetaMap{
			statusID: {BlockLabel: "in_progress"},
		},
	}
	exporters := map[string]*resourceExporter.ResourceExporter{
		onAttributeChangeRule.ResourceType: ruleExporter,
		statusType:                         statusExporter,
	}
	resource := resourceExporter.ResourceInfo{
		Type:       onAttributeChangeRule.ResourceType,
		BlockLabel: "rule",
		BlockType:  "resource",
		State:      &terraform.InstanceState{ID: "rule-id"},
	}
	g := setupGenesysCloudResourceExporter(t)

	sanitizeCondition := func(attribute interface{}) map[string]interface{} {
		condition := map[string]interface{}{"new_value": statusID, "old_value": statusID}
		if attribute != nil {
			condition["attribute"] = attribute
		}
		configMap := map[string]interface{}{
			"name":      "rule",
			"condition": []interface{}{condition},
		}
		_, ok := g.sanitizeConfigMap(resource, configMap, "", exporters, false, "hcl", true)
		require.True(t, ok)
		return configMap["condition"].([]interface{})[0].(map[string]interface{})
	}

	t.Run("statusId values become worktype_status references", func(t *testing.T) {
		condition := sanitizeCondition("statusId")
		assert.Equal(t, "${"+statusType+".in_progress.id}", condition["new_value"])
		assert.Equal(t, "${"+statusType+".in_progress.id}", condition["old_value"])
	})

	t.Run("other attributes are left as plain values", func(t *testing.T) {
		condition := sanitizeCondition("priority")
		assert.Equal(t, statusID, condition["new_value"])
		assert.Equal(t, statusID, condition["old_value"])
	})

	t.Run("missing or non-string attribute is left as plain values", func(t *testing.T) {
		for _, attribute := range []interface{}{nil, 5} {
			condition := sanitizeCondition(attribute)
			assert.Equal(t, statusID, condition["new_value"])
			assert.Equal(t, statusID, condition["old_value"])
		}
	})
}
