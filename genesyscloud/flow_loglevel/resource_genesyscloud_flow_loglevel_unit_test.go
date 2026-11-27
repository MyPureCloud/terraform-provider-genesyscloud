package flow_loglevel

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	architectFlow "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/stretchr/testify/assert"
)

// newReadTestProxy returns a proxy whose get-by-ID records the flow ID it was asked for.
func newReadTestProxy(requestedFlowId *string, level string) *flowLogLevelProxy {
	return &flowLogLevelProxy{
		getFlowLogLevelByIdAttr: func(_ context.Context, _ *flowLogLevelProxy, flowId string) (*platformclientv2.Flowsettingsresponse, *platformclientv2.APIResponse, error) {
			*requestedFlowId = flowId
			return &platformclientv2.Flowsettingsresponse{
				Id:                      platformclientv2.String(flowId),
				LogLevelCharacteristics: &platformclientv2.Flowloglevel{Level: platformclientv2.String(level)},
			}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
		},
	}
}

// An exported or imported resource starts with only its ID, which is the flow ID. The read must
// fetch that flow and record flow_id, or the export has no flow_id to apply.
func TestUnitResourceFlowLogLevelReadFromIdOnly(t *testing.T) {
	const flowId = "11111111-2222-3333-4444-555555555555"
	var requested string
	internalProxy = newReadTestProxy(&requested, "All")
	defer func() { internalProxy = nil }()

	d := schema.TestResourceDataRaw(t, ResourceFlowLoglevel().Schema, map[string]interface{}{})
	d.SetId(flowId)

	diags := readFlowLogLevel(context.Background(), d, &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}})

	assert.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, flowId, requested)
	assert.Equal(t, flowId, d.Get("flow_id"))
	assert.Equal(t, "All", d.Get("flow_log_level"))
}

// The read addresses the flow by the resource ID and sets flow_id from it, as group_roles does.
func TestUnitResourceFlowLogLevelReadUsesResourceId(t *testing.T) {
	const flowId = "11111111-2222-3333-4444-555555555555"
	var requested string
	internalProxy = newReadTestProxy(&requested, "Base")
	defer func() { internalProxy = nil }()

	d := schema.TestResourceDataRaw(t, ResourceFlowLoglevel().Schema, map[string]interface{}{
		"flow_id":        flowId,
		"flow_log_level": "Base",
	})
	d.SetId(flowId)

	diags := readFlowLogLevel(context.Background(), d, &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}})

	assert.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, flowId, requested)
	assert.Equal(t, flowId, d.Get("flow_id"))
}

// Create sets the resource ID to the flow ID itself, so it does not depend on the response's ID.
func TestUnitResourceFlowLogLevelCreateSetsIdToFlowId(t *testing.T) {
	const flowId = "11111111-2222-3333-4444-555555555555"
	var requested, created string
	proxy := newReadTestProxy(&requested, "All")
	proxy.createFlowLogLevelAttr = func(_ context.Context, _ *flowLogLevelProxy, id string, _ *platformclientv2.Flowloglevelrequest) (*platformclientv2.Flowsettingsresponse, *platformclientv2.APIResponse, error) {
		created = id
		// No Id in the response.
		return &platformclientv2.Flowsettingsresponse{}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	d := schema.TestResourceDataRaw(t, ResourceFlowLoglevel().Schema, map[string]interface{}{
		"flow_id":        flowId,
		"flow_log_level": "All",
	})

	diags := createFlowLogLevel(context.Background(), d, &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}})

	assert.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, flowId, created)
	assert.Equal(t, flowId, d.Id())
	assert.Equal(t, flowId, requested)
}

// Update and delete address the flow by the resource ID, and update does not depend on the response's ID.
func TestUnitResourceFlowLogLevelUpdateAndDeleteUseResourceId(t *testing.T) {
	const flowId = "11111111-2222-3333-4444-555555555555"
	var requested, updated, deleted string
	proxy := newReadTestProxy(&requested, "All")
	proxy.updateFlowLogLevelAttr = func(_ context.Context, _ *flowLogLevelProxy, id string, _ *platformclientv2.Flowloglevelrequest) (*platformclientv2.Flowsettingsresponse, *platformclientv2.APIResponse, error) {
		updated = id
		// No Id in the response.
		return &platformclientv2.Flowsettingsresponse{}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.deleteFlowLogLevelByIdAttr = func(_ context.Context, _ *flowLogLevelProxy, id string) (*platformclientv2.APIResponse, error) {
		deleted = id
		// Deleted from here on, so the delete's follow-up read sees a 404.
		proxy.getFlowLogLevelByIdAttr = func(_ context.Context, _ *flowLogLevelProxy, _ string) (*platformclientv2.Flowsettingsresponse, *platformclientv2.APIResponse, error) {
			return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, fmt.Errorf("not found")
		}
		return &platformclientv2.APIResponse{StatusCode: http.StatusNoContent}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	meta := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceFlowLoglevel().Schema, map[string]interface{}{
		"flow_id":        flowId,
		"flow_log_level": "All",
	})
	d.SetId(flowId)

	assert.False(t, updateFlowLogLevel(context.Background(), d, meta).HasError())
	assert.Equal(t, flowId, updated)
	assert.False(t, deleteFlowLogLevel(context.Background(), d, meta).HasError())
	assert.Equal(t, flowId, deleted)
}

// flow_id references the flow, so exports resolve it to the flow rather than hard-coding its GUID.
func TestUnitFlowLogLevelExporterFlowIdIsFlowReference(t *testing.T) {
	ref, ok := FlowLogLevelExporter().RefAttrs["flow_id"]
	if assert.True(t, ok, "flow_id must be a RefAttr") {
		assert.Equal(t, architectFlow.ResourceType, ref.RefType)
	}
}
