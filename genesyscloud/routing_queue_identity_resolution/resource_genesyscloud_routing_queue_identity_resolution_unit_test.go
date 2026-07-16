package routing_queue_identity_resolution

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitResourceRoutingQueueIdentityResolutionUpdate(t *testing.T) {
	tQueueId := uuid.NewString()

	proxy := &routingQueueIdentityResolutionProxy{}
	proxy.putRoutingQueueIdentityResolutionAttr = func(ctx context.Context, p *routingQueueIdentityResolutionProxy, queueId string, config platformclientv2.Identityresolutionqueueconfig) (*platformclientv2.Identityresolutionqueueconfig, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tQueueId, queueId)
		assert.NotNil(t, config.CallOnBehalfOfQueue)
		assert.NotNil(t, config.CallOnBehalfOfQueue.ResolveIdentities)
		assert.Equal(t, false, *config.CallOnBehalfOfQueue.ResolveIdentities)

		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &config, &apiResponse, nil
	}

	proxy.getRoutingQueueIdentityResolutionAttr = func(ctx context.Context, p *routingQueueIdentityResolutionProxy, queueId string) (*platformclientv2.Identityresolutionqueueconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Identityresolutionqueueconfig{
			CallOnBehalfOfQueue: &platformclientv2.Outboundqueueidentityresolutionconfig{
				ResolveIdentities: &resolveIdentities,
			},
		}, &apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceRoutingQueueIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tQueueId, false, "")

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tQueueId)

	diag := updateRoutingQueueIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Equal(t, tQueueId, d.Id())
	assert.Equal(t, tQueueId, d.Get("queue_id").(string))
}

func TestUnitResourceRoutingQueueIdentityResolutionRead(t *testing.T) {
	tQueueId := uuid.NewString()
	tDivisionId := uuid.NewString()

	proxy := &routingQueueIdentityResolutionProxy{}
	proxy.getRoutingQueueIdentityResolutionAttr = func(ctx context.Context, p *routingQueueIdentityResolutionProxy, queueId string) (*platformclientv2.Identityresolutionqueueconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Identityresolutionqueueconfig{
			CallOnBehalfOfQueue: &platformclientv2.Outboundqueueidentityresolutionconfig{
				ResolveIdentities: &resolveIdentities,
				Division: &platformclientv2.Writablestarrabledivision{
					Id: &tDivisionId,
				},
			},
		}, &apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceRoutingQueueIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tQueueId, false, tDivisionId)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tQueueId)

	diag := readRoutingQueueIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Equal(t, tQueueId, d.Id())
	assert.Equal(t, tQueueId, d.Get("queue_id").(string))

	blocks := d.Get("call_on_behalf_of_queue").([]interface{})
	assert.Len(t, blocks, 1)
	block := blocks[0].(map[string]interface{})
	assert.Equal(t, false, block["resolve_identities"])
	assert.Equal(t, tDivisionId, block["division_id"])
}

func TestUnitResourceRoutingQueueIdentityResolutionReadDefaultConfig(t *testing.T) {
	tQueueId := uuid.NewString()
	allDivisions := "*"
	resolveTrue := true

	tests := []struct {
		name   string
		config *platformclientv2.Identityresolutionqueueconfig
	}{
		{
			name: "resolve true without division",
			config: &platformclientv2.Identityresolutionqueueconfig{
				CallOnBehalfOfQueue: &platformclientv2.Outboundqueueidentityresolutionconfig{
					ResolveIdentities: &resolveTrue,
				},
			},
		},
		{
			name: "resolve true with all divisions",
			config: &platformclientv2.Identityresolutionqueueconfig{
				CallOnBehalfOfQueue: &platformclientv2.Outboundqueueidentityresolutionconfig{
					ResolveIdentities: &resolveTrue,
					Division: &platformclientv2.Writablestarrabledivision{
						Id: &allDivisions,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			proxy := &routingQueueIdentityResolutionProxy{}
			proxy.getRoutingQueueIdentityResolutionAttr = func(_ context.Context, _ *routingQueueIdentityResolutionProxy, queueId string) (*platformclientv2.Identityresolutionqueueconfig, *platformclientv2.APIResponse, error) {
				assert.Equal(t, tQueueId, queueId)
				apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
				return test.config, &apiResponse, nil
			}

			internalProxy = proxy
			defer func() { internalProxy = nil }()

			ctx := context.Background()
			gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
			resourceSchema := ResourceRoutingQueueIdentityResolution().Schema
			resourceDataMap := buildIdentityResolutionResourceMap(tQueueId, true, "")

			d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
			d.SetId(tQueueId)

			diag := readRoutingQueueIdentityResolution(ctx, d, gcloud)
			assert.False(t, diag.HasError(), diag)
			assert.Equal(t, tQueueId, d.Id())
			assert.Equal(t, tQueueId, d.Get("queue_id").(string))

			blocks := d.Get("call_on_behalf_of_queue").([]interface{})
			assert.Len(t, blocks, 1)
			block := blocks[0].(map[string]interface{})
			assert.Equal(t, true, block["resolve_identities"])
			if divisionId, ok := block["division_id"].(string); ok {
				assert.True(t, divisionId == "" || divisionId == "*", "expected division_id to be omitted or all divisions, got %q", divisionId)
			}
		})
	}
}

func TestUnitResourceRoutingQueueIdentityResolutionDelete(t *testing.T) {
	tQueueId := uuid.NewString()

	proxy := &routingQueueIdentityResolutionProxy{}
	proxy.getRoutingQueueByIdAttr = func(ctx context.Context, p *routingQueueIdentityResolutionProxy, queueId string) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Queue{Id: &tQueueId}, &apiResponse, nil
	}

	proxy.putRoutingQueueIdentityResolutionAttr = func(ctx context.Context, p *routingQueueIdentityResolutionProxy, queueId string, config platformclientv2.Identityresolutionqueueconfig) (*platformclientv2.Identityresolutionqueueconfig, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tQueueId, queueId)
		assert.NotNil(t, config.CallOnBehalfOfQueue)
		assert.NotNil(t, config.CallOnBehalfOfQueue.ResolveIdentities)
		assert.Equal(t, true, *config.CallOnBehalfOfQueue.ResolveIdentities)
		assert.Nil(t, config.CallOnBehalfOfQueue.Division)

		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &config, &apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceRoutingQueueIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tQueueId, false, "")

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tQueueId)

	diag := deleteRoutingQueueIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
}

func TestUnitGetAllRoutingQueueIdentityResolution(t *testing.T) {
	defaultQueueId := uuid.NewString()
	defaultQueueName := "Default Queue"
	customQueueId := uuid.NewString()
	customQueueName := "Custom Queue"
	notFoundQueueId := uuid.NewString()
	notFoundQueueName := "Not Found Queue"

	resolveTrue := true
	resolveFalse := false
	divisionId := uuid.NewString()

	queueProxy := &routingQueue.RoutingQueueProxy{}
	queueProxy.GetAllRoutingQueuesAttr = func(_ context.Context, _ *routingQueue.RoutingQueueProxy, _ string, _ bool) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error) {
		return &[]platformclientv2.Queue{
			{Id: &defaultQueueId, Name: &defaultQueueName},
			{Id: &customQueueId, Name: &customQueueName},
			{Id: &notFoundQueueId, Name: &notFoundQueueName},
		}, nil, nil
	}

	proxy := &routingQueueIdentityResolutionProxy{
		routingQueueProxy: queueProxy,
	}
	proxy.getRoutingQueueIdentityResolutionAttr = func(_ context.Context, _ *routingQueueIdentityResolutionProxy, queueId string) (*platformclientv2.Identityresolutionqueueconfig, *platformclientv2.APIResponse, error) {
		switch queueId {
		case defaultQueueId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
			return &platformclientv2.Identityresolutionqueueconfig{
				CallOnBehalfOfQueue: &platformclientv2.Outboundqueueidentityresolutionconfig{
					ResolveIdentities: &resolveTrue,
				},
			}, &apiResponse, nil
		case customQueueId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
			return &platformclientv2.Identityresolutionqueueconfig{
				CallOnBehalfOfQueue: &platformclientv2.Outboundqueueidentityresolutionconfig{
					ResolveIdentities: &resolveFalse,
					Division: &platformclientv2.Writablestarrabledivision{
						Id: &divisionId,
					},
				},
			}, &apiResponse, nil
		case notFoundQueueId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusNotFound}
			return nil, &apiResponse, fmt.Errorf("not found")
		default:
			t.Fatalf("unexpected queue id %s", queueId)
			return nil, nil, nil
		}
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	resources, diag := getAllRoutingQueueIdentityResolution(ctx, &platformclientv2.Configuration{})

	assert.False(t, diag.HasError())
	assert.Len(t, resources, 1)
	assert.Contains(t, resources, customQueueId)
	assert.Equal(t, customQueueName+"-identity-resolution", resources[customQueueId].BlockLabel)
}

func TestUnitGetAllRoutingQueueIdentityResolutionListError(t *testing.T) {
	queueProxy := &routingQueue.RoutingQueueProxy{}
	queueProxy.GetAllRoutingQueuesAttr = func(_ context.Context, _ *routingQueue.RoutingQueueProxy, _ string, _ bool) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error) {
		return nil, nil, fmt.Errorf("mock list error")
	}

	proxy := &routingQueueIdentityResolutionProxy{
		routingQueueProxy: queueProxy,
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	resources, diag := getAllRoutingQueueIdentityResolution(ctx, &platformclientv2.Configuration{})

	assert.True(t, diag.HasError())
	assert.Nil(t, resources)
}

func buildIdentityResolutionResourceMap(queueId string, resolveIdentities bool, divisionId string) map[string]interface{} {
	block := map[string]interface{}{
		"resolve_identities": resolveIdentities,
	}
	if divisionId != "" {
		block["division_id"] = divisionId
	}

	return map[string]interface{}{
		"queue_id": queueId,
		"call_on_behalf_of_queue": []interface{}{
			block,
		},
	}
}
