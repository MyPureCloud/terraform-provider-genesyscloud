package outbound_campaignrule

import (
	"context"
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

	campaignRulePoxy := &outboundCampaignruleProxy{}
	campaignRulePoxy.getOutboundCampaignruleByIdAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, id string) (*platformclientv2.Campaignrule, int, error) {
		assert.Equal(t, tId, id)
		campaignRule := &testCampaignRule

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return campaignRule, apiResponse.StatusCode, nil
	}

	campaignRulePoxy.createOutboundCampaignruleAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, campaignRule *platformclientv2.Campaignrule) (*platformclientv2.Campaignrule, error) {
		assert.Equal(t, testCampaignRule.Name, *campaignRule.Name, "campaignRule.Name check failed in create createOutboundCampaignruleAttr")

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
	resourceDataMap := buildCampaignRuleResourceMap(tId, *testCampaignRule.Name)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := createOutboundCampaignRule(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceOutboundCampaignruleRead(t *testing.T) {
	tId := uuid.NewString()
	testCampaignRule := generateCampaignRuleData(tId, "")

	campaignRulePoxy := &outboundCampaignruleProxy{}

	campaignRulePoxy.getOutboundCampaignruleByIdAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, id string) (*platformclientv2.Campaignrule, int, error) {
		assert.Equal(t, tId, id)
		campaignRule := &testCampaignRule

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return campaignRule, apiResponse.StatusCode, nil
	}

	internalProxy = campaignRulePoxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &gcloud.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceOutboundCampaignrule().Schema

	//Setup a map of values
	resourceDataMap := buildCampaignRuleResourceMap(tId, *testCampaignRule.Name)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readOutboundCampaignRule(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
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
		assert.Equal(t, testCampaignRule.Name, *campaignRule.Name, "campaignRule.Name check failed in create createOutboundCampaignruleAttr")

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
	resourceDataMap := buildCampaignRuleResourceMap(tId, *testCampaignRule.Name)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := updateOutboundCampaignRule(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceOutboundCampaignruleDelete(t *testing.T) {
	tId := uuid.NewString()
	//tName := "Updated campaign rule name"
	testCampaignRule := generateCampaignRuleData(tId, "")

	campaignRulePoxy := &outboundCampaignruleProxy{}

	campaignRulePoxy.deleteOutboundCampaignruleAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, id string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	campaignRulePoxy.getOutboundCampaignruleByIdAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, id string) (*platformclientv2.Campaignrule, int, error) {
		assert.Equal(t, tId, id)
		campaignRule := &testCampaignRule

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return campaignRule, apiResponse.StatusCode, nil
	}

	internalProxy = campaignRulePoxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &gcloud.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceOutboundCampaignrule().Schema

	//Setup a map of values
	resourceDataMap := buildCampaignRuleResourceMap(tId, *testCampaignRule.Name)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := deleteOutboundCampaignRule(ctx, d, gcloud)
	assert.Nil(t, diag)
	assert.Equal(t, tId, d.Id())
}

func generateCampaignRuleData(id string, name string) platformclientv2.Campaignrule {
	tEnabled := false
	tMatchAnyConditions := true
	//var campaigns []platformclientv2.Domainentityref
	//var sequences []platformclientv2.Domainentityref
	//for i := 0; i <= 3; i++ {
	//	campaigns[i] = generateRandomDomainEntityRef()
	//	sequences[i] = generateRandomDomainEntityRef()
	//}
	//campaignRuleEntities := platformclientv2.Campaignruleentities{
	//	Campaigns: &campaigns,
	//	Sequences: &sequences,
	//}

	return platformclientv2.Campaignrule{
		Id:                 &id,
		Name:               &name,
		Enabled:            &tEnabled,
		MatchAnyConditions: &tMatchAnyConditions,
		//CampaignRuleEntities: &campaignRuleEntities,
	}
}

func generateRandomDomainEntityRef() platformclientv2.Domainentityref {
	id := uuid.NewString()
	return platformclientv2.Domainentityref{
		Id: &id,
	}
}

func buildCampaignRuleResourceMap(tId string, tName string) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"id":   tId,
		"name": tName,
	}
	return resourceDataMap
}
