package business_rules_decision_table

import (
	"context"
	"strconv"
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
	assert.Equal(t, 8*time.Minute, *resource.Timeouts.Delete, "Delete timeout should be 8m")
}

// TestUnitDecisionTableDefaultsToComputed verifies the mutually-exclusive defaults_to
// fields (value, values, special) are Optional+Computed. Because a column sets exactly
// one of them and the legacy SDK materializes the unused siblings as empty, Computed is
// what keeps the null -> "" reconciliation from being flagged as an inconsistency (the
// "produced an invalid plan / unexpected new value" warning) and avoids a phantom
// ForceNew diff on columns.
func TestUnitDecisionTableDefaultsToComputed(t *testing.T) {
	defaultsTo := defaultsToSchemaFunc()

	for _, field := range []string{"value", "values", "special"} {
		s, ok := defaultsTo.Schema[field]
		if !ok {
			t.Fatalf("defaults_to schema missing field %q", field)
		}
		assert.True(t, s.Optional, "defaults_to.%s should be Optional", field)
		assert.True(t, s.Computed, "defaults_to.%s should be Computed (mutually-exclusive sibling owned by provider)", field)
	}
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

// TestUnitIsDuplicateRowAtIndex verifies the add-rows idempotency guard: a 409
// "decision.table.duplicate.row" whose reported index matches the row currently
// being added is treated as an already-created row (the 504 ghost / non-idempotent
// retry case) and skipped, while genuine duplicates, other errors, and malformed
// responses are not skipped.
func TestUnitIsDuplicateRowAtIndex(t *testing.T) {
	dupBody := func(id string, index int) []byte {
		return []byte(`{"message":"Duplicate decision table rows found [{\"id\":\"` + id + `\",\"index\":` + strconv.Itoa(index) + `}]",` +
			`"code":"decision.table.duplicate.row","status":409,` +
			`"messageWithParams":"Duplicate decision table rows found {duplicateRows}",` +
			`"messageParams":{"duplicateRows":"[{\"id\":\"` + id + `\",\"index\":` + strconv.Itoa(index) + `}]"}}`)
	}

	tests := []struct {
		name      string
		resp      *platformclientv2.APIResponse
		rowNumber int
		want      bool
	}{
		{
			name:      "409 duplicate at matching index -> skip",
			resp:      &platformclientv2.APIResponse{StatusCode: 409, RawBody: dupBody("bb8d8418", 5409)},
			rowNumber: 5409,
			want:      true,
		},
		{
			name:      "409 duplicate at earlier index (genuine config dup) -> do not skip",
			resp:      &platformclientv2.APIResponse{StatusCode: 409, RawBody: dupBody("bb8d8418", 10)},
			rowNumber: 5409,
			want:      false,
		},
		{
			name:      "409 with different error code -> do not skip",
			resp:      &platformclientv2.APIResponse{StatusCode: 409, RawBody: []byte(`{"code":"some.other.conflict","status":409}`)},
			rowNumber: 5409,
			want:      false,
		},
		{
			name:      "non-409 status -> do not skip",
			resp:      &platformclientv2.APIResponse{StatusCode: 500, RawBody: dupBody("bb8d8418", 5409)},
			rowNumber: 5409,
			want:      false,
		},
		{
			name:      "nil response -> do not skip",
			resp:      nil,
			rowNumber: 5409,
			want:      false,
		},
		{
			name:      "malformed body -> do not skip",
			resp:      &platformclientv2.APIResponse{StatusCode: 409, RawBody: []byte(`not-json`)},
			rowNumber: 5409,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isDuplicateRowAtIndex(tt.resp, tt.rowNumber))
		})
	}
}
