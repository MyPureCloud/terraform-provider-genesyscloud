package routing_queue_conditional_group_routing

import (
	"context"
	"net/http"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitResourceRoutingQueueConditionalGroupRoutingUpdate(t *testing.T) {
	tQueueId := uuid.NewString()
	tRules := generateRuleData()
	tId := tQueueId + "/rules"

	if !featureToggles.CSGToggleExists() {
		t.Skipf("Skipping because %s env variable is not set", featureToggles.CSGToggleName())
	}

	groupRoutingProxy := &routingQueueConditionalGroupRoutingProxy{}
	groupRoutingProxy.updateRoutingQueueConditionRoutingAttr = func(ctx context.Context, p *routingQueueConditionalGroupRoutingProxy, queueId string, rules *[]platformclientv2.Conditionalgrouproutingrule) (*[]platformclientv2.Conditionalgrouproutingrule, *platformclientv2.APIResponse, error) {
		equal := cmp.Equal(tRules, *rules)
		assert.Equal(t, true, equal, "rules not equal to expected value in update: %s", cmp.Diff(tRules, *rules))

		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return rules, &apiResponse, nil
	}

	groupRoutingProxy.getRoutingQueueConditionRoutingAttr = func(ctx context.Context, p *routingQueueConditionalGroupRoutingProxy, queueId string) (*[]platformclientv2.Conditionalgrouproutingrule, *platformclientv2.APIResponse, error) {
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &tRules, &apiResponse, nil
	}

	internalProxy = groupRoutingProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceRoutingQueueConditionalGroupRouting().Schema

	//Setup a map of values
	resourceDataMap := buildConditionalGroupRoutingResourceMap(tQueueId, &tRules)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := updateRoutingQueueConditionalRoutingGroup(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError(), diag)
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceRoutingQueueConditionalGroupRoutingRead(t *testing.T) {
	tQueueId := uuid.NewString()
	tRules := generateRuleData()
	tId := tQueueId + "/rules"

	if !featureToggles.CSGToggleExists() {
		t.Skipf("Skipping because %s env variable is not set", featureToggles.CSGToggleName())
	}

	groupRoutingProxy := &routingQueueConditionalGroupRoutingProxy{}

	groupRoutingProxy.getRoutingQueueConditionRoutingAttr = func(ctx context.Context, p *routingQueueConditionalGroupRoutingProxy, queueId string) (*[]platformclientv2.Conditionalgrouproutingrule, *platformclientv2.APIResponse, error) {
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &tRules, &apiResponse, nil
	}

	internalProxy = groupRoutingProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceRoutingQueueConditionalGroupRouting().Schema

	//Setup a map of values
	resourceDataMap := buildConditionalGroupRoutingResourceMap(tQueueId, &tRules)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readRoutingQueueConditionalRoutingGroup(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError(), diag)

	assert.Equal(t, tId, d.Id())
	rules, err := buildConditionalGroupRouting(d.Get("rules").([]interface{}))
	assert.Equal(t, err, nil)
	equal := cmp.Equal(tRules, rules)
	assert.Equal(t, true, equal, "rules not equal to expected value in read: %s", cmp.Diff(tRules, rules))
}

func generateRuleData() []platformclientv2.Conditionalgrouproutingrule {
	groupMember1 := platformclientv2.Membergroup{
		Id:      platformclientv2.String(uuid.NewString()),
		VarType: platformclientv2.String("TEAM"),
	}
	groupMember2 := platformclientv2.Membergroup{
		Id:      platformclientv2.String(uuid.NewString()),
		VarType: platformclientv2.String("SKILLGROUP"),
	}
	groupMember3 := platformclientv2.Membergroup{
		Id:      platformclientv2.String(uuid.NewString()),
		VarType: platformclientv2.String("GROUP"),
	}
	group1 := []platformclientv2.Membergroup{groupMember1, groupMember2, groupMember3}

	rule1 := platformclientv2.Conditionalgrouproutingrule{
		Metric:         platformclientv2.String("test"),
		Operator:       platformclientv2.String("GreaterThan"),
		ConditionValue: platformclientv2.Float64(2345),
		Groups:         &group1,
		WaitSeconds:    platformclientv2.Int(5432),
	}

	rules := []platformclientv2.Conditionalgrouproutingrule{rule1}

	return rules
}

func buildConditionalGroupRoutingResourceMap(queueId string, rules *[]platformclientv2.Conditionalgrouproutingrule) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"queue_id": queueId,
		"rules":    flattenConditionalGroupRouting(rules),
	}

	return resourceDataMap
}
