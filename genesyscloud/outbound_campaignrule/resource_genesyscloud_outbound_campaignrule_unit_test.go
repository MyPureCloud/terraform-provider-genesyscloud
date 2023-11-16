package outbound_campaignrule

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
	"github.com/stretchr/testify/assert"
	"net/http"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"testing"
)

func TestUnitResourceOutboundCampaignruleCreate(t *testing.T) {
	tId := uuid.NewString()
	tName := "campaign rule name"
	testCampaignRule := generateCampaignRuleData(tId, tName)

	campaignRuleProxy := &outboundCampaignruleProxy{}
	campaignRuleProxy.getOutboundCampaignruleByIdAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, id string) (*platformclientv2.Campaignrule, int, error) {
		assert.Equal(t, tId, id)
		campaignRule := &testCampaignRule

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return campaignRule, apiResponse.StatusCode, nil
	}

	campaignRuleProxy.createOutboundCampaignruleAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, campaignRule *platformclientv2.Campaignrule) (*platformclientv2.Campaignrule, error) {
		assert.Equal(t, *testCampaignRule.Name, *campaignRule.Name, "campaignRule.Name check failed in create createOutboundCampaignruleAttr")
		assert.Equal(t, *testCampaignRule.Enabled, *campaignRule.Enabled, "campaignRule.Enabled check failed in create createOutboundCampaignruleAttr")
		assert.Equal(t, *testCampaignRule.MatchAnyConditions, *campaignRule.MatchAnyConditions, "campaignRule.MatchAnyConditions check failed in create createOutboundCampaignruleAttr")
		assert.EqualValues(t, testCampaignRule.CampaignRuleEntities.Campaigns, campaignRule.CampaignRuleEntities.Campaigns, "campaignRule.CampaignRuleEntities.Campaigns check failed in create createOutboundCampaignruleAttr")
		assert.EqualValues(t, testCampaignRule.CampaignRuleEntities.Sequences, campaignRule.CampaignRuleEntities.Sequences, "campaignRule.CampaignRuleEntities.Sequences check failed in create createOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].Id, *(*campaignRule.CampaignRuleConditions)[0].Id, "campaignRule.CampaignRuleConditions[0].Id check failed in create createOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].ConditionType, *(*campaignRule.CampaignRuleConditions)[0].ConditionType, "campaignRule.CampaignRuleConditions[0].Id check failed in create createOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].Parameters.Priority, *(*campaignRule.CampaignRuleConditions)[0].Parameters.Priority, "campaignRule.CampaignRuleConditions[0].Parameters.Priority check failed in create createOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].Parameters.Operator, *(*campaignRule.CampaignRuleConditions)[0].Parameters.Operator, "campaignRule.CampaignRuleConditions[0].Parameters.Operator check failed in create createOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].Parameters.Value, *(*campaignRule.CampaignRuleConditions)[0].Parameters.Value, "campaignRule.CampaignRuleConditions[0].Parameters.Value check failed in create createOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].Parameters.DialingMode, *(*campaignRule.CampaignRuleConditions)[0].Parameters.DialingMode, "campaignRule.CampaignRuleConditions[0].Parameters.DialingMode check failed in create createOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].Id, *(*campaignRule.CampaignRuleActions)[0].Id, "campaignRule.CampaignRuleActions[0].Id check failed in create createOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].ActionType, *(*campaignRule.CampaignRuleActions)[0].ActionType, "campaignRule.CampaignRuleActions[0].ActionType check failed in create createOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].Parameters.Operator, *(*campaignRule.CampaignRuleActions)[0].Parameters.Operator, "campaignRule.CampaignRuleActions[0].Parameters.Operator check failed in create createOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].Parameters.Priority, *(*campaignRule.CampaignRuleActions)[0].Parameters.Priority, "campaignRule.CampaignRuleActions[0].Parameters.Priority check failed in create createOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].Parameters.Value, *(*campaignRule.CampaignRuleActions)[0].Parameters.Value, "campaignRule.CampaignRuleActions[0].Parameters.Value check failed in create createOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].Parameters.DialingMode, *(*campaignRule.CampaignRuleActions)[0].Parameters.DialingMode, "campaignRule.CampaignRuleActions[0].Parameters.DialingMode check failed in create createOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].CampaignRuleActionEntities.UseTriggeringEntity, *(*campaignRule.CampaignRuleActions)[0].CampaignRuleActionEntities.UseTriggeringEntity, "campaignRule.CampaignRuleActions[0].CampaignRuleActionEntities.UseTriggeringEntity check failed in create createOutboundCampaignruleAttr")
		assert.EqualValues(t, *(*testCampaignRule.CampaignRuleActions)[0].CampaignRuleActionEntities.Sequences, *(*campaignRule.CampaignRuleActions)[0].CampaignRuleActionEntities.Sequences, "campaignRule.CampaignRuleActions[0].CampaignRuleActionEntities.Sequences check failed in create createOutboundCampaignruleAttr")
		assert.EqualValues(t, *(*testCampaignRule.CampaignRuleActions)[0].CampaignRuleActionEntities.Campaigns, *(*campaignRule.CampaignRuleActions)[0].CampaignRuleActionEntities.Campaigns, "campaignRule.CampaignRuleActions[0].CampaignRuleActionEntities.Campaigns check failed in create createOutboundCampaignruleAttr")

		campaignRule.Id = &tId

		return campaignRule, nil
	}

	internalProxy = campaignRuleProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &gcloud.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceOutboundCampaignrule().Schema

	//Setup a map of values
	resourceDataMap := buildCampaignRuleResourceMap(tId, *testCampaignRule.Name, *testCampaignRule.Enabled, *testCampaignRule.MatchAnyConditions, *testCampaignRule.CampaignRuleEntities, *testCampaignRule.CampaignRuleConditions, *testCampaignRule.CampaignRuleActions)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := createOutboundCampaignRule(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceOutboundCampaignruleRead(t *testing.T) {
	tId := uuid.NewString()
	tName := "campaign rule name"
	testCampaignRule := generateCampaignRuleData(tId, tName)

	campaignRuleProxy := &outboundCampaignruleProxy{}

	campaignRuleProxy.getOutboundCampaignruleByIdAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, id string) (*platformclientv2.Campaignrule, int, error) {
		assert.Equal(t, tId, id)
		campaignRule := &testCampaignRule

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return campaignRule, apiResponse.StatusCode, nil
	}

	internalProxy = campaignRuleProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &gcloud.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceOutboundCampaignrule().Schema

	//Setup a map of values
	resourceDataMap := buildCampaignRuleResourceMap(tId, *testCampaignRule.Name, *testCampaignRule.Enabled, *testCampaignRule.MatchAnyConditions, *testCampaignRule.CampaignRuleEntities, *testCampaignRule.CampaignRuleConditions, *testCampaignRule.CampaignRuleActions)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readOutboundCampaignRule(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, *testCampaignRule.Name, d.Get("name").(string))
	assert.Equal(t, *testCampaignRule.Enabled, d.Get("enabled").(bool))
	assert.Equal(t, *testCampaignRule.MatchAnyConditions, d.Get("match_any_conditions").(bool))
	campaignRuleEntities := buildCampaignRuleEntities(d.Get("campaign_rule_entities").(*schema.Set))
	assert.EqualValues(t, testCampaignRule.CampaignRuleEntities.Campaigns, campaignRuleEntities.Campaigns)
	assert.EqualValues(t, testCampaignRule.CampaignRuleEntities.Sequences, campaignRuleEntities.Sequences)
	campaignRuleConditions := buildCampaignRuleConditions(d.Get("campaign_rule_conditions").([]interface{}))
	assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].Id, *(*campaignRuleConditions)[0].Id)
	assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].ConditionType, *(*campaignRuleConditions)[0].ConditionType)
	assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].Parameters.Value, *(*campaignRuleConditions)[0].Parameters.Value)
	assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].Parameters.Priority, *(*campaignRuleConditions)[0].Parameters.Priority)
	assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].Parameters.Operator, *(*campaignRuleConditions)[0].Parameters.Operator)
	assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].Parameters.DialingMode, *(*campaignRuleConditions)[0].Parameters.DialingMode)
	campaignRuleActions := buildCampaignRuleAction(d.Get("campaign_rule_actions").([]interface{}))
	assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].Id, *(*campaignRuleActions)[0].Id)
	assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].ActionType, *(*campaignRuleActions)[0].ActionType)
	assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].Parameters.Value, *(*campaignRuleActions)[0].Parameters.Value)
	assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].Parameters.Operator, *(*campaignRuleActions)[0].Parameters.Operator)
	assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].Parameters.Priority, *(*campaignRuleActions)[0].Parameters.Priority)
	assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].Parameters.DialingMode, *(*campaignRuleActions)[0].Parameters.DialingMode)
	assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].CampaignRuleActionEntities.UseTriggeringEntity, *(*campaignRuleActions)[0].CampaignRuleActionEntities.UseTriggeringEntity)
	assert.EqualValues(t, *(*testCampaignRule.CampaignRuleActions)[0].CampaignRuleActionEntities.Campaigns, *(*campaignRuleActions)[0].CampaignRuleActionEntities.Campaigns)
	assert.EqualValues(t, *(*testCampaignRule.CampaignRuleActions)[0].CampaignRuleActionEntities.Sequences, *(*campaignRuleActions)[0].CampaignRuleActionEntities.Sequences)
}

func TestUnitResourceOutboundCampaignruleUpdate(t *testing.T) {
	tId := uuid.NewString()
	tName := "Updated campaign rule name"
	testCampaignRule := generateCampaignRuleData(tId, tName)

	campaignRulePoxy := &outboundCampaignruleProxy{}
	campaignRulePoxy.getOutboundCampaignruleByIdAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, id string) (*platformclientv2.Campaignrule, int, error) {
		assert.Equal(t, tId, id)
		campaignRule := &testCampaignRule

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return campaignRule, apiResponse.StatusCode, nil
	}

	campaignRulePoxy.updateOutboundCampaignruleAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, id string, campaignRule *platformclientv2.Campaignrule) (*platformclientv2.Campaignrule, error) {
		assert.Equal(t, *testCampaignRule.Name, *campaignRule.Name, "campaignRule.Name check failed in update updateOutboundCampaignruleAttr")
		assert.Equal(t, *testCampaignRule.Enabled, *campaignRule.Enabled, "campaignRule.Enabled check failed in update updateOutboundCampaignruleAttr")
		assert.Equal(t, *testCampaignRule.MatchAnyConditions, *campaignRule.MatchAnyConditions, "campaignRule.MatchAnyConditions check failed in update updateOutboundCampaignruleAttr")
		assert.EqualValues(t, testCampaignRule.CampaignRuleEntities.Campaigns, campaignRule.CampaignRuleEntities.Campaigns, "campaignRule.CampaignRuleEntities.Campaigns check failed in update updateOutboundCampaignruleAttr")
		assert.EqualValues(t, testCampaignRule.CampaignRuleEntities.Sequences, campaignRule.CampaignRuleEntities.Sequences, "campaignRule.CampaignRuleEntities.Sequences check failed in update updateOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].Id, *(*campaignRule.CampaignRuleConditions)[0].Id, "campaignRule.CampaignRuleConditions[0].Id check failed in update updateOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].ConditionType, *(*campaignRule.CampaignRuleConditions)[0].ConditionType, "campaignRule.CampaignRuleConditions[0].Id check failed in update updateOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].Parameters.Priority, *(*campaignRule.CampaignRuleConditions)[0].Parameters.Priority, "campaignRule.CampaignRuleConditions[0].Parameters.Priority check failed in update updateOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].Parameters.Operator, *(*campaignRule.CampaignRuleConditions)[0].Parameters.Operator, "campaignRule.CampaignRuleConditions[0].Parameters.Operator check failed in update updateOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].Parameters.Value, *(*campaignRule.CampaignRuleConditions)[0].Parameters.Value, "campaignRule.CampaignRuleConditions[0].Parameters.Value check failed in update updateOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleConditions)[0].Parameters.DialingMode, *(*campaignRule.CampaignRuleConditions)[0].Parameters.DialingMode, "campaignRule.CampaignRuleConditions[0].Parameters.DialingMode check failed in update updateOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].Id, *(*campaignRule.CampaignRuleActions)[0].Id, "campaignRule.CampaignRuleActions[0].Id check failed in update updateOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].ActionType, *(*campaignRule.CampaignRuleActions)[0].ActionType, "campaignRule.CampaignRuleActions[0].ActionType check failed in update updateOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].Parameters.Operator, *(*campaignRule.CampaignRuleActions)[0].Parameters.Operator, "campaignRule.CampaignRuleActions[0].Parameters.Operator check failed in update updateOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].Parameters.Priority, *(*campaignRule.CampaignRuleActions)[0].Parameters.Priority, "campaignRule.CampaignRuleActions[0].Parameters.Priority check failed in update updateOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].Parameters.Value, *(*campaignRule.CampaignRuleActions)[0].Parameters.Value, "campaignRule.CampaignRuleActions[0].Parameters.Value check failed in update updateOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].Parameters.DialingMode, *(*campaignRule.CampaignRuleActions)[0].Parameters.DialingMode, "campaignRule.CampaignRuleActions[0].Parameters.DialingMode check failed in update updateOutboundCampaignruleAttr")
		assert.Equal(t, *(*testCampaignRule.CampaignRuleActions)[0].CampaignRuleActionEntities.UseTriggeringEntity, *(*campaignRule.CampaignRuleActions)[0].CampaignRuleActionEntities.UseTriggeringEntity, "campaignRule.CampaignRuleActions[0].CampaignRuleActionEntities.UseTriggeringEntity check failed in update updateOutboundCampaignruleAttr")
		assert.EqualValues(t, *(*testCampaignRule.CampaignRuleActions)[0].CampaignRuleActionEntities.Sequences, *(*campaignRule.CampaignRuleActions)[0].CampaignRuleActionEntities.Sequences, "campaignRule.CampaignRuleActions[0].CampaignRuleActionEntities.Sequences check failed in update updateOutboundCampaignruleAttr")
		assert.EqualValues(t, *(*testCampaignRule.CampaignRuleActions)[0].CampaignRuleActionEntities.Campaigns, *(*campaignRule.CampaignRuleActions)[0].CampaignRuleActionEntities.Campaigns, "campaignRule.CampaignRuleActions[0].CampaignRuleActionEntities.Campaigns check failed in update updateOutboundCampaignruleAttr")

		campaignRule.Id = &tId

		return campaignRule, nil
	}

	internalProxy = campaignRulePoxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &gcloud.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceOutboundCampaignrule().Schema

	//Setup a map of values
	resourceDataMap := buildCampaignRuleResourceMap(tId, *testCampaignRule.Name, *testCampaignRule.Enabled, *testCampaignRule.MatchAnyConditions, *testCampaignRule.CampaignRuleEntities, *testCampaignRule.CampaignRuleConditions, *testCampaignRule.CampaignRuleActions)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := updateOutboundCampaignRule(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, *testCampaignRule.Name, d.Get("name").(string))
}

func TestUnitResourceOutboundCampaignruleDelete(t *testing.T) {
	tId := uuid.NewString()
	tName := "campaign rule name"
	testCampaignRule := generateCampaignRuleData(tId, tName)

	campaignRulePoxy := &outboundCampaignruleProxy{}

	campaignRulePoxy.deleteOutboundCampaignruleAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, id string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	campaignRulePoxy.getOutboundCampaignruleByIdAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, id string) (*platformclientv2.Campaignrule, int, error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}
		err := fmt.Errorf("Unable to find targeted IVR: %s", id)
		return nil, apiResponse.StatusCode, err
	}

	internalProxy = campaignRulePoxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &gcloud.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceOutboundCampaignrule().Schema

	//Setup a map of values
	resourceDataMap := buildCampaignRuleResourceMap(tId, *testCampaignRule.Name, *testCampaignRule.Enabled, *testCampaignRule.MatchAnyConditions, *testCampaignRule.CampaignRuleEntities, *testCampaignRule.CampaignRuleConditions, *testCampaignRule.CampaignRuleActions)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := deleteOutboundCampaignRule(ctx, d, gcloud)
	assert.Nil(t, diag)
	assert.Equal(t, tId, d.Id())
}

func generateCampaignRuleData(id string, name string) platformclientv2.Campaignrule {
	// Create campaign rule entity
	campaignsEntities := make([]platformclientv2.Domainentityref, 3)
	sequencesEntities := make([]platformclientv2.Domainentityref, 3)
	for i := 0; i < 3; i++ {
		campaignsEntities[i] = generateRandomDomainEntityRef()
		sequencesEntities[i] = generateRandomDomainEntityRef()
	}
	campaignRuleEntities := platformclientv2.Campaignruleentities{
		Campaigns: &campaignsEntities,
		Sequences: &sequencesEntities,
	}

	// Create campaign rule conditions
	parameterCondition := platformclientv2.Campaignruleparameters{
		Operator:    platformclientv2.String("lessThan"),
		Value:       platformclientv2.String("0.5"),
		DialingMode: platformclientv2.String("preview"),
		Priority:    platformclientv2.String("2"),
	}
	campaignRuleCondition := platformclientv2.Campaignrulecondition{
		Id:            platformclientv2.String(uuid.NewString()),
		ConditionType: platformclientv2.String("campaignProgress"),
		Parameters:    &parameterCondition,
	}
	campaignRuleConditions := []platformclientv2.Campaignrulecondition{campaignRuleCondition}

	// Create campaign rule actions
	parameterAction := platformclientv2.Campaignruleparameters{
		Operator:    platformclientv2.String("lessThan"),
		Value:       platformclientv2.String("0.5"),
		DialingMode: platformclientv2.String("preview"),
		Priority:    platformclientv2.String("2"),
	}
	campaignsActions := make([]platformclientv2.Domainentityref, 3)
	sequencesActions := make([]platformclientv2.Domainentityref, 3)
	for i := 0; i < 3; i++ {
		campaignsActions[i] = generateRandomDomainEntityRef()
		sequencesActions[i] = generateRandomDomainEntityRef()
	}
	actionEntities := platformclientv2.Campaignruleactionentities{
		UseTriggeringEntity: platformclientv2.Bool(false),
		Campaigns:           &campaignsActions,
		Sequences:           &sequencesActions,
	}
	campaignRuleAction := platformclientv2.Campaignruleaction{
		Id:                         platformclientv2.String(uuid.NewString()),
		ActionType:                 platformclientv2.String("turnOnCampaign"),
		Parameters:                 &parameterAction,
		CampaignRuleActionEntities: &actionEntities,
	}
	campaignRuleActions := []platformclientv2.Campaignruleaction{campaignRuleAction}

	return platformclientv2.Campaignrule{
		Id:                     &id,
		Name:                   &name,
		Enabled:                platformclientv2.Bool(false),
		MatchAnyConditions:     platformclientv2.Bool(true),
		CampaignRuleEntities:   &campaignRuleEntities,
		CampaignRuleConditions: &campaignRuleConditions,
		CampaignRuleActions:    &campaignRuleActions,
	}
}

func generateRandomDomainEntityRef() platformclientv2.Domainentityref {
	id := uuid.NewString()
	return platformclientv2.Domainentityref{
		Id: &id,
	}
}

// tCampaignRuleConditions interface{}, tCampaignRuleActions interface{}
func buildCampaignRuleResourceMap(tId string, tName string, tEnabled bool, tMatchAnyConditions bool, tCampaignRuleEntities platformclientv2.Campaignruleentities, tCampaignRuleConditions []platformclientv2.Campaignrulecondition, tCampaignRuleActions []platformclientv2.Campaignruleaction) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"id":                       tId,
		"name":                     tName,
		"enabled":                  tEnabled,
		"match_any_conditions":     tMatchAnyConditions,
		"campaign_rule_entities":   generateCampaignruleEntityInterface(&tCampaignRuleEntities),
		"campaign_rule_conditions": flattenCampaignRuleConditions(&tCampaignRuleConditions),
		"campaign_rule_actions":    flattenCampaignRuleAction(&tCampaignRuleActions, generateActionEntities),
	}
	return resourceDataMap
}

func generateCampaignruleEntityInterface(campaignRuleEntities *platformclientv2.Campaignruleentities) []interface{} {
	var (
		campaignRuleEntitiesMap = make(map[string]interface{})
		campaigns               []interface{}
		sequences               []interface{}
	)

	if campaignRuleEntities.Campaigns != nil {
		for _, v := range *campaignRuleEntities.Campaigns {
			campaigns = append(campaigns, *v.Id)
		}
	}

	if campaignRuleEntities.Sequences != nil {
		for _, v := range *campaignRuleEntities.Sequences {
			sequences = append(sequences, *v.Id)
		}
	}

	campaignRuleEntitiesMap["campaign_ids"] = campaigns
	campaignRuleEntitiesMap["sequence_ids"] = sequences

	return []interface{}{campaignRuleEntitiesMap}
}

func generateActionEntities(entities *platformclientv2.Campaignruleactionentities) []interface{} {
	var (
		campaigns   []interface{}
		sequences   []interface{}
		entitiesMap = make(map[string]interface{})
	)

	if entities == nil {
		return nil
	}

	if entities.Campaigns != nil {
		for _, campaign := range *entities.Campaigns {
			campaigns = append(campaigns, *campaign.Id)
		}
	}

	if entities.Sequences != nil {
		for _, sequence := range *entities.Sequences {
			sequences = append(sequences, *sequence.Id)
		}
	}

	entitiesMap["campaign_ids"] = campaigns
	entitiesMap["sequence_ids"] = sequences
	entitiesMap["use_triggering_entity"] = *entities.UseTriggeringEntity

	return []interface{}{entitiesMap}
}
