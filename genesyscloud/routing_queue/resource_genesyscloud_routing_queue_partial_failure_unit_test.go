package routing_queue

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/stretchr/testify/assert"
)

// Verifies partial failure after updateQueueWrapupCodes: consistency entry is cleared, state is
// refreshed from the API when sync succeeds, and diagnostics include the wrapup error.
func TestUnitUpdateRoutingQueuePartialFailureWrapupSyncsStateAndClearsConsistencyCheck(t *testing.T) {
	tID := uuid.NewString()
	queueName := "queue-before-sync"
	testRoutingQueue := generateRoutingQueueData(tID, queueName)
	testRoutingQueue.CannedResponseLibraries = nil

	apiQueue := convertCreateQueuetoQueue(testRoutingQueue)
	syncedName := "queue-after-api-sync"
	apiQueueSynced := *apiQueue
	apiQueueSynced.Name = &syncedName

	var getByIDCalls int
	wrapupListCalls := 0

	queueProxy := &RoutingQueueProxy{}
	queueProxy.createRoutingQueueWrapupCodeAttr = func(ctx context.Context, p *RoutingQueueProxy, queueId string, body []platformclientv2.Wrapupcodereference) ([]platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	queueProxy.deleteRoutingQueueWrapupCodeAttr = func(ctx context.Context, p *RoutingQueueProxy, queueId, codeId string) (*platformclientv2.APIResponse, error) {
		return &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	queueProxy.getRoutingQueueByIdAttr = func(ctx context.Context, p *RoutingQueueProxy, id string, checkCache bool) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
		getByIDCalls++
		assert.Equal(t, tID, id)
		switch getByIDCalls {
		case 1:
			return apiQueue, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
		case 2:
			return &apiQueueSynced, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
		default:
			return apiQueue, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
		}
	}

	queueProxy.updateRoutingQueueAttr = func(ctx context.Context, p *RoutingQueueProxy, id string, routingQueue *platformclientv2.Queuerequest) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	queueProxy.getAllRoutingQueueWrapupCodesAttr = func(ctx context.Context, p *RoutingQueueProxy, queueId string) (*[]platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
		wrapupListCalls++
		if wrapupListCalls == 1 {
			return nil, &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("forced wrapup codes list failure")
		}
		empty := []platformclientv2.Wrapupcode{}
		return &empty, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	err := setRoutingQueueUnitTestsEnvVar()
	if err != nil {
		t.Skipf("failed to set env variable %s: %s", unitTestsAreActiveEnv, err.Error())
	}
	internalProxy = queueProxy
	defer func() {
		internalProxy = nil
		_ = unsetRoutingQueueUnitTestsEnvVar()
	}()

	ctx := context.Background()
	meta := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceRoutingQueue().Schema
	resourceDataMap := buildRoutingQueueResourceMap(tID, queueName, testRoutingQueue)
	resourceDataMap["ignore_members"] = true
	resourceDataMap["wrapup_codes"] = []interface{}{uuid.NewString()}

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tID)

	_ = consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingQueue(), constants.ConsistencyChecks(), ResourceType)
	assert.True(t, consistency_checker.ConsistencyCheckExists(tID), "precondition: consistency check should be registered")

	diags := updateRoutingQueue(ctx, d, meta)
	assert.True(t, diags.HasError())
	assert.False(t, consistency_checker.ConsistencyCheckExists(tID), "consistency check should be removed after partial failure")

	msg := diagSummary(diags)
	assert.Contains(t, msg, "Failed to query wrapup codes", "expected original wrapup failure diagnostic")
	assert.Equal(t, syncedName, d.Get("name").(string), "resource name should match API after syncRoutingQueueStateFromAPI")
	assert.GreaterOrEqual(t, wrapupListCalls, 2, "wrapup list should be queried for update then again during sync")
	assert.GreaterOrEqual(t, getByIDCalls, 2, "getRoutingQueueById should run for addCGRAndOEA then for sync")
}

// Verifies diagnostics aggregate the wrapup failure and a subsequent sync read failure.
func TestUnitUpdateRoutingQueuePartialFailureWrapupIncludesSyncReadError(t *testing.T) {
	tID := uuid.NewString()
	queueName := "queue-wrapup-fail"
	testRoutingQueue := generateRoutingQueueData(tID, queueName)
	testRoutingQueue.CannedResponseLibraries = nil

	apiQueue := convertCreateQueuetoQueue(testRoutingQueue)

	var getByIDCalls int
	wrapupListCalls := 0

	queueProxy := &RoutingQueueProxy{}
	queueProxy.createRoutingQueueWrapupCodeAttr = func(ctx context.Context, p *RoutingQueueProxy, queueId string, body []platformclientv2.Wrapupcodereference) ([]platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	queueProxy.deleteRoutingQueueWrapupCodeAttr = func(ctx context.Context, p *RoutingQueueProxy, queueId, codeId string) (*platformclientv2.APIResponse, error) {
		return &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	queueProxy.getRoutingQueueByIdAttr = func(ctx context.Context, p *RoutingQueueProxy, id string, checkCache bool) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
		getByIDCalls++
		assert.Equal(t, tID, id)
		if getByIDCalls == 1 {
			return apiQueue, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
		}
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("forced sync read failure")
	}

	queueProxy.updateRoutingQueueAttr = func(ctx context.Context, p *RoutingQueueProxy, id string, routingQueue *platformclientv2.Queuerequest) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	queueProxy.getAllRoutingQueueWrapupCodesAttr = func(ctx context.Context, p *RoutingQueueProxy, queueId string) (*[]platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
		wrapupListCalls++
		if wrapupListCalls == 1 {
			return nil, &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("forced wrapup codes list failure")
		}
		empty := []platformclientv2.Wrapupcode{}
		return &empty, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	err := setRoutingQueueUnitTestsEnvVar()
	if err != nil {
		t.Skipf("failed to set env variable %s: %s", unitTestsAreActiveEnv, err.Error())
	}
	internalProxy = queueProxy
	defer func() {
		internalProxy = nil
		_ = unsetRoutingQueueUnitTestsEnvVar()
	}()

	ctx := context.Background()
	meta := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceRoutingQueue().Schema
	resourceDataMap := buildRoutingQueueResourceMap(tID, queueName, testRoutingQueue)
	resourceDataMap["ignore_members"] = true
	resourceDataMap["wrapup_codes"] = []interface{}{uuid.NewString()}

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tID)

	_ = consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingQueue(), constants.ConsistencyChecks(), ResourceType)

	diags := updateRoutingQueue(ctx, d, meta)
	assert.True(t, diags.HasError())
	assert.False(t, consistency_checker.ConsistencyCheckExists(tID))

	msg := diagSummary(diags)
	assert.Contains(t, msg, "Failed to query wrapup codes")
	assert.Contains(t, msg, "Failed to read queue")
	assert.GreaterOrEqual(t, wrapupListCalls, 1)
	assert.Equal(t, 2, getByIDCalls)
}

func diagSummary(diags diag.Diagnostics) string {
	var b strings.Builder
	for _, d := range diags {
		b.WriteString(d.Summary)
		b.WriteString(" ")
		b.WriteString(d.Detail)
		b.WriteString("; ")
	}
	return b.String()
}
