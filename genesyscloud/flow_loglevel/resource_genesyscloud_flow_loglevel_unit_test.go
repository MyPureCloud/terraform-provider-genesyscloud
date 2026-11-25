package flow_loglevel

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
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

// A configured resource keeps reading the flow named by flow_id, as before.
func TestUnitResourceFlowLogLevelReadPrefersFlowId(t *testing.T) {
	const flowId = "11111111-2222-3333-4444-555555555555"
	var requested string
	internalProxy = newReadTestProxy(&requested, "Base")
	defer func() { internalProxy = nil }()

	d := schema.TestResourceDataRaw(t, ResourceFlowLoglevel().Schema, map[string]interface{}{
		"flow_id":        flowId,
		"flow_log_level": "Base",
	})
	d.SetId("some-other-id")

	diags := readFlowLogLevel(context.Background(), d, &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}})

	assert.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, flowId, requested)
	assert.Equal(t, flowId, d.Get("flow_id"))
}
