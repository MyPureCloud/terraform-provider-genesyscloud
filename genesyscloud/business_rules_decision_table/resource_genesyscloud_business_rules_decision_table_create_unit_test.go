package business_rules_decision_table

import (
	"context"
	"testing"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v188/platformclientv2"
	"github.com/stretchr/testify/assert"
)

// TestUnitDecisionTableSchemaTimeouts verifies the resource exposes the longer
// Create/Update timeouts (needed for large tables that add rows one POST at a
// time) while keeping the Read and Delete timeouts.
func TestUnitDecisionTableSchemaTimeouts(t *testing.T) {
	resource := ResourceBusinessRulesDecisionTable()
	if resource.Timeouts == nil {
		t.Fatal("resource Timeouts block is nil")
	}

	assert.NotNil(t, resource.Timeouts.Create, "Create timeout should be set")
	assert.NotNil(t, resource.Timeouts.Update, "Update timeout should be set")
	assert.NotNil(t, resource.Timeouts.Read, "Read timeout should be set")
	assert.NotNil(t, resource.Timeouts.Delete, "Delete timeout should be set")

	assert.Equal(t, 120*time.Minute, *resource.Timeouts.Create, "Create timeout should be 120m")
	assert.Equal(t, 120*time.Minute, *resource.Timeouts.Update, "Update timeout should be 120m")
	assert.Equal(t, 8*time.Minute, *resource.Timeouts.Read, "Read timeout should be 8m")
	assert.Equal(t, 8*time.Minute, *resource.Timeouts.Delete, "Delete timeout should be 10m")
}

// TestUnitRollbackDecisionTableUsesFreshContext verifies that the create-failure
// rollback deletes the partial table on a fresh, non-expired context, even when
// the original request context is already cancelled (the create-timeout case).
// This guards against the orphaned-table bug where the rollback reused the
// expired request context and the DELETE was never sent.
func TestUnitRollbackDecisionTableUsesFreshContext(t *testing.T) {
	const tableId = "table-rollback-123"

	var (
		deleteCalled   bool
		deletedTableId string
		ctxHadError    error
		ctxHadDeadline bool
	)

	proxy := &BusinessRulesDecisionTableProxy{}
	proxy.deleteBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, id string) (*platformclientv2.APIResponse, error) {
		deleteCalled = true
		deletedTableId = id
		ctxHadError = ctx.Err()
		_, ctxHadDeadline = ctx.Deadline()
		return &platformclientv2.APIResponse{StatusCode: 200}, nil
	}

	rollbackDecisionTable(tableId, proxy)

	assert.True(t, deleteCalled, "rollback should invoke the delete")
	assert.Equal(t, tableId, deletedTableId, "rollback should delete the correct table")
	assert.NoError(t, ctxHadError, "rollback context must be live (not expired/cancelled)")
	assert.True(t, ctxHadDeadline, "rollback context should carry its own deadline")
}

// TestUnitRollbackDecisionTableNotBlockedByExpiredRequestContext verifies that an
// already-expired request context does not affect the rollback, since the
// rollback derives a brand new context from context.Background().
func TestUnitRollbackDecisionTableNotBlockedByExpiredRequestContext(t *testing.T) {
	expiredCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	cancel()
	time.Sleep(1 * time.Millisecond)
	assert.Error(t, expiredCtx.Err(), "sanity: request context should be expired")

	var ctxErrSeen error
	proxy := &BusinessRulesDecisionTableProxy{}
	proxy.deleteBusinessRulesDecisionTableAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, id string) (*platformclientv2.APIResponse, error) {
		ctxErrSeen = ctx.Err()
		return &platformclientv2.APIResponse{StatusCode: 200}, nil
	}

	rollbackDecisionTable("table-xyz", proxy)

	assert.NoError(t, ctxErrSeen, "rollback should run on a fresh context regardless of the expired request context")
}
