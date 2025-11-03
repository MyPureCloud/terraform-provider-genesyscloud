package flow_outcome

import (
	"context"
	"net/http"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitFlowOutcomeCreate(t *testing.T) {
	var (
		name        = "Unit Test Flow Outcome"
		description = "Test description"
		divisionId  = uuid.NewString()
		outcomeId   = uuid.NewString()
	)

	proxy := &flowOutcomeProxy{}

	proxy.getFlowOutcomeIdByNameAttr = func(ctx context.Context, p *flowOutcomeProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
		return "", true, nil, nil
	}

	proxy.createFlowOutcomeAttr = func(ctx context.Context, p *flowOutcomeProxy, flowOutcome *platformclientv2.Flowoutcome) (*platformclientv2.Flowoutcome, *platformclientv2.APIResponse, error) {
		assert.Equal(t, name, *flowOutcome.Name, "flowOutcome.Name check failed in create")
		assert.Equal(t, description, *flowOutcome.Description, "flowOutcome.Description check failed in create")
		assert.Equal(t, divisionId, *flowOutcome.Division.Id, "flowOutcome.Division.Id check failed in create")

		createdOutcome := &platformclientv2.Flowoutcome{
			Id:          &outcomeId,
			Name:        &name,
			Description: &description,
			Division:    &platformclientv2.Writabledivision{Id: &divisionId},
		}

		return createdOutcome, nil, nil
	}

	proxy.getFlowOutcomeByIdAttr = func(ctx context.Context, p *flowOutcomeProxy, id string) (flowOutcome *platformclientv2.Flowoutcome, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, outcomeId, id)
		outcome := &platformclientv2.Flowoutcome{
			Id:          &outcomeId,
			Name:        &name,
			Description: &description,
			Division:    &platformclientv2.Writabledivision{Id: &divisionId},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return outcome, apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceFlowOutcome().Schema
	resourceDataMap := buildFlowOutcomeResourceMap(outcomeId, name, description, divisionId)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(outcomeId)

	diag := createFlowOutcome(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, outcomeId, d.Id())
}

func TestUnitFlowOutcomeRead(t *testing.T) {
	var (
		name        = "Unit Test Flow Outcome"
		description = "Test description"
		divisionId  = uuid.NewString()
		outcomeId   = uuid.NewString()
	)

	proxy := &flowOutcomeProxy{}

	proxy.getFlowOutcomeByIdAttr = func(ctx context.Context, p *flowOutcomeProxy, id string) (flowOutcome *platformclientv2.Flowoutcome, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, outcomeId, id)
		outcome := &platformclientv2.Flowoutcome{
			Id:          &outcomeId,
			Name:        &name,
			Description: &description,
			Division:    &platformclientv2.Writabledivision{Id: &divisionId},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return outcome, apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceFlowOutcome().Schema
	resourceDataMap := buildFlowOutcomeResourceMap(outcomeId, name, description, divisionId)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(outcomeId)

	diag := readFlowOutcome(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, outcomeId, d.Id())
	assert.Equal(t, name, d.Get("name").(string))
	assert.Equal(t, description, d.Get("description").(string))
	assert.Equal(t, divisionId, d.Get("division_id").(string))
}

func TestUnitFlowOutcomeUpdate(t *testing.T) {
	var (
		name        = "Unit Test Flow Outcome Updated"
		description = "Updated description"
		divisionId  = uuid.NewString()
		outcomeId   = uuid.NewString()
	)

	proxy := &flowOutcomeProxy{}

	proxy.updateFlowOutcomeAttr = func(ctx context.Context, p *flowOutcomeProxy, id string, flowOutcome *platformclientv2.Flowoutcome) (*platformclientv2.Flowoutcome, *platformclientv2.APIResponse, error) {
		assert.Equal(t, outcomeId, id)
		assert.Equal(t, name, *flowOutcome.Name, "flowOutcome.Name check failed in update")
		assert.Equal(t, description, *flowOutcome.Description, "flowOutcome.Description check failed in update")
		assert.Equal(t, divisionId, *flowOutcome.Division.Id, "flowOutcome.Division.Id check failed in update")

		updatedOutcome := &platformclientv2.Flowoutcome{
			Id:          &outcomeId,
			Name:        &name,
			Description: &description,
			Division:    &platformclientv2.Writabledivision{Id: &divisionId},
		}

		return updatedOutcome, nil, nil
	}

	proxy.getFlowOutcomeByIdAttr = func(ctx context.Context, p *flowOutcomeProxy, id string) (flowOutcome *platformclientv2.Flowoutcome, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, outcomeId, id)
		outcome := &platformclientv2.Flowoutcome{
			Id:          &outcomeId,
			Name:        &name,
			Description: &description,
			Division:    &platformclientv2.Writabledivision{Id: &divisionId},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return outcome, apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceFlowOutcome().Schema
	resourceDataMap := buildFlowOutcomeResourceMap(outcomeId, name, description, divisionId)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(outcomeId)

	diag := updateFlowOutcome(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, outcomeId, d.Id())
}

func TestUnitDataSourceFlowOutcomeRead(t *testing.T) {
	targetId := uuid.NewString()
	targetName := "MyTargetId"
	proxy := &flowOutcomeProxy{}
	proxy.getFlowOutcomeIdByNameAttr = func(ctx context.Context, p *flowOutcomeProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, targetName, name)
		return targetId, false, nil, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()
	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := DataSourceFlowOutcome().Schema

	resourceDataMap := map[string]interface{}{
		"name": targetName,
	}

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	dataSourceFlowOutcomeRead(ctx, d, gcloud)
	assert.Equal(t, targetId, d.Id())
}

func TestUnitResourceFlowOutcomeGetAll(t *testing.T) {
	var (
		name1        = "Unit Test Flow Outcome"
		description1 = "Test description"
		divisionId1  = uuid.NewString()
		outcomeId1   = uuid.NewString()

		name2        = "Second Flow Outcome"
		description2 = "Second description"
		divisionId2  = uuid.NewString()
		outcomeId2   = uuid.NewString()
	)

	proxy := &flowOutcomeProxy{}

	proxy.getAllFlowOutcomeAttr = func(ctx context.Context, p *flowOutcomeProxy) (*[]platformclientv2.Flowoutcome, *platformclientv2.APIResponse, error) {
		var outcomes []platformclientv2.Flowoutcome

		outcome1 := &platformclientv2.Flowoutcome{
			Id:          &outcomeId1,
			Name:        &name1,
			Description: &description1,
			Division:    &platformclientv2.Writabledivision{Id: &divisionId1},
		}

		outcome2 := &platformclientv2.Flowoutcome{
			Id:          &outcomeId2,
			Name:        &name2,
			Description: &description2,
			Division:    &platformclientv2.Writabledivision{Id: &divisionId2},
		}
		outcomes = append(outcomes, *outcome1)
		outcomes = append(outcomes, *outcome2)

		return &outcomes, nil, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()
	ctx := context.Background()

	exportedResource, diag := getAllAuthFlowOutcomes(ctx, &platformclientv2.Configuration{})
	assert.Equal(t, true, len(exportedResource) == 2)
	assert.Equal(t, false, diag.HasError())
}

func buildFlowOutcomeResourceMap(id string, name string, description string, divisionId string) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"id":          id,
		"name":        name,
		"description": description,
		"division_id": divisionId,
	}
	return resourceDataMap
}
