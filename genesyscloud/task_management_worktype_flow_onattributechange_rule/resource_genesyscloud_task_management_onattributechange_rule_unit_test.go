package task_management_worktype_flow_onattributechange_rule

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/stretchr/testify/assert"
)

func setupReadTest(t *testing.T, rule *platformclientv2.Workitemonattributechangerule) (*schema.ResourceData, *provider.ProviderMeta) {
	proxyObj := &taskManagementOnAttributeChangeRuleProxy{}
	proxyObj.getTaskManagementOnAttributeChangeRuleByIdAttr = func(ctx context.Context, p *taskManagementOnAttributeChangeRuleProxy, worktypeId string, id string) (*platformclientv2.Workitemonattributechangerule, *platformclientv2.APIResponse, error) {
		assert.Equal(t, *rule.Worktype.Id, worktypeId)
		assert.Equal(t, *rule.Id, id)
		return rule, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxyObj
	t.Cleanup(func() { internalProxy = nil })

	d := schema.TestResourceDataRaw(t, ResourceTaskManagementOnAttributeChangeRule().Schema, map[string]interface{}{})
	d.SetId(ComposeWorktypeBasedTerraformId(*rule.Worktype.Id, *rule.Id))
	return d, &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
}

// Regression test for WORKITEMS-9668: a rule using condition groups (match) has a nil condition.
// Read must return an error diagnostic instead of panicking.
func TestUnitResourceOnAttributeChangeRuleReadConditionGroupsReturnsError(t *testing.T) {
	worktypeId, ruleId := uuid.NewString(), uuid.NewString()
	rule := &platformclientv2.Workitemonattributechangerule{
		Id:        &ruleId,
		Name:      platformclientv2.String("condition group rule"),
		Worktype:  &platformclientv2.Worktypereference{Id: &worktypeId},
		Condition: nil,
	}
	d, meta := setupReadTest(t, rule)

	diags := readTaskManagementOnAttributeChangeRule(context.Background(), d, meta)

	assert.True(t, diags.HasError())
	assert.True(t, strings.Contains(diags[0].Summary, "uses condition groups (match)"), diags[0].Summary)
	assert.True(t, strings.Contains(diags[0].Summary, ruleId), diags[0].Summary)
	assert.Empty(t, d.Get("condition").([]interface{}), "no condition should be written to state")
}

func TestUnitResourceOnAttributeChangeRuleRead(t *testing.T) {
	worktypeId, ruleId, newStatusId := uuid.NewString(), uuid.NewString(), uuid.NewString()
	rule := &platformclientv2.Workitemonattributechangerule{
		Id:       &ruleId,
		Name:     platformclientv2.String("status rule"),
		Worktype: &platformclientv2.Worktypereference{Id: &worktypeId},
		Condition: &platformclientv2.Workitemonattributechangecondition{
			Attribute: platformclientv2.String("statusId"),
			NewValue:  &newStatusId,
		},
	}
	d, meta := setupReadTest(t, rule)

	diags := readTaskManagementOnAttributeChangeRule(context.Background(), d, meta)

	assert.False(t, diags.HasError(), diags)
	assert.Equal(t, "status rule", d.Get("name").(string))
	assert.Equal(t, worktypeId, d.Get("worktype_id").(string))
	assert.Equal(t, "statusId", d.Get("condition.0.attribute").(string))
	assert.Equal(t, ComposeWorktypeBasedTerraformId(worktypeId, newStatusId), d.Get("condition.0.new_value").(string))
}

func TestUnitFlattenSdkConditionNilCondition(t *testing.T) {
	assert.Nil(t, flattenSdkCondition(nil))
	assert.Nil(t, flattenSdkCondition(&platformclientv2.Workitemonattributechangerule{}))
}
