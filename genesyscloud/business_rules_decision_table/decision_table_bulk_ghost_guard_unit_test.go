package business_rules_decision_table

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func unitTestDuplicateRowResponse(index int) *platformclientv2.APIResponse {
	body := []byte(`{"message":"Duplicate decision table rows found [{\"id\":\"ghost-row\",\"index\":` + strconv.Itoa(index) + `}]",` +
		`"code":"decision.table.duplicate.row","status":409,` +
		`"messageWithParams":"Duplicate decision table rows found {duplicateRows}",` +
		`"messageParams":{"duplicateRows":"[{\"id\":\"ghost-row\",\"index\":` + strconv.Itoa(index) + `}]"}}`)
	return &platformclientv2.APIResponse{StatusCode: http.StatusConflict, RawBody: body}
}

func unitTestRulePersistenceConflictResponse() *platformclientv2.APIResponse {
	body := []byte(`{"message":"Transaction canceled","code":"RULE_PERSISTENCE_CONFLICT","status":409}`)
	return &platformclientv2.APIResponse{StatusCode: http.StatusConflict, RawBody: body}
}

func unitTestBulkAddRow() platformclientv2.Createdecisiontablerowrequest {
	return platformclientv2.Createdecisiontablerowrequest{}
}

func TestUnitExpectedChunkIndexRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		baseRowCount int
		chunkStart   int
		chunkSize    int
		wantFirst    int
		wantLast     int
	}{
		{name: "create first chunk", baseRowCount: 0, chunkStart: 0, chunkSize: 15, wantFirst: 1, wantLast: 15},
		{name: "create second chunk", baseRowCount: 0, chunkStart: 15, chunkSize: 15, wantFirst: 16, wantLast: 30},
		{name: "update after kept rows", baseRowCount: 3, chunkStart: 0, chunkSize: 2, wantFirst: 4, wantLast: 5},
		{name: "update second add chunk", baseRowCount: 3, chunkStart: 2, chunkSize: 2, wantFirst: 6, wantLast: 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first, last := expectedChunkIndexRange(tt.baseRowCount, tt.chunkStart, tt.chunkSize)
			assert.Equal(t, tt.wantFirst, first)
			assert.Equal(t, tt.wantLast, last)
		})
	}
}

// TestUnitIsGhostChunkDuplicate is the bulk equivalent of the former
// TestUnitIsDuplicateRowAtIndex: a matching-index-range 409 duplicate.row is a
// ghost chunk; genuine duplicates, RULE_PERSISTENCE_CONFLICT, and bare/malformed
// responses are not skipped.
func TestUnitIsGhostChunkDuplicate(t *testing.T) {
	t.Parallel()

	dupBody := func(id string, index int) []byte {
		return []byte(`{"message":"Duplicate decision table rows found [{\"id\":\"` + id + `\",\"index\":` + strconv.Itoa(index) + `}]",` +
			`"code":"decision.table.duplicate.row","status":409,` +
			`"messageWithParams":"Duplicate decision table rows found {duplicateRows}",` +
			`"messageParams":{"duplicateRows":"[{\"id\":\"` + id + `\",\"index\":` + strconv.Itoa(index) + `}]"}}`)
	}

	tests := []struct {
		name         string
		resp         *platformclientv2.APIResponse
		baseRowCount int
		chunkStart   int
		chunkSize    int
		want         bool
	}{
		{
			name:         "duplicate index in chunk range -> ghost",
			resp:         &platformclientv2.APIResponse{StatusCode: 409, RawBody: dupBody("ghost", 16)},
			baseRowCount: 0,
			chunkStart:   15,
			chunkSize:    15,
			want:         true,
		},
		{
			name:         "duplicate index before chunk range -> genuine duplicate",
			resp:         &platformclientv2.APIResponse{StatusCode: 409, RawBody: dupBody("existing", 3)},
			baseRowCount: 3,
			chunkStart:   0,
			chunkSize:    2,
			want:         false,
		},
		{
			name:         "RULE_PERSISTENCE_CONFLICT 409 -> not ghost",
			resp:         unitTestRulePersistenceConflictResponse(),
			baseRowCount: 0,
			chunkStart:   0,
			chunkSize:    15,
			want:         false,
		},
		{
			name:         "nil response -> not ghost",
			resp:         nil,
			baseRowCount: 0,
			chunkStart:   0,
			chunkSize:    15,
			want:         false,
		},
		{
			name:         "empty body -> not ghost",
			resp:         &platformclientv2.APIResponse{StatusCode: 409},
			baseRowCount: 0,
			chunkStart:   0,
			chunkSize:    15,
			want:         false,
		},
		{
			name:         "non-409 status -> not ghost",
			resp:         &platformclientv2.APIResponse{StatusCode: 504, RawBody: dupBody("ghost", 1)},
			baseRowCount: 0,
			chunkStart:   0,
			chunkSize:    1,
			want:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isGhostChunkDuplicate(tt.resp, tt.baseRowCount, tt.chunkStart, tt.chunkSize))
		})
	}
}

// TestUnitBulkAddConvertedRowsSkipsGhostChunkDuplicate verifies create-path bulk
// add treats a matching-index-range 409 as an already-created ghost chunk.
func TestUnitBulkAddConvertedRowsSkipsGhostChunkDuplicate(t *testing.T) {
	const (
		tableID      = "table-bulk-ghost"
		version      = 1
		baseRowCount = 0
		expectedIdx  = 16 // first row of second chunk on create
	)

	var bulkCalls int
	proxy := &BusinessRulesDecisionTableProxy{}
	proxy.bulkAddDecisionTableRowsAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, ver int, rows []platformclientv2.Createdecisiontablerowrequest) (*platformclientv2.APIResponse, error) {
		bulkCalls++
		if bulkCalls == 1 {
			return &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
		}
		return unitTestDuplicateRowResponse(expectedIdx), assert.AnError
	}

	rows := make([]platformclientv2.Createdecisiontablerowrequest, 16)
	for i := range rows {
		rows[i] = unitTestBulkAddRow()
	}

	err := bulkAddConvertedRows(context.Background(), proxy, tableID, version, rows, 15, baseRowCount)
	assert.NoError(t, err)
	assert.Equal(t, 2, bulkCalls, "ghost duplicate should be treated as success without further retries")
}

// TestUnitBulkAddConvertedRowsFailsOnGenuineDuplicate verifies a 409 duplicate.row
// at an earlier index is not skipped during bulk add.
func TestUnitBulkAddConvertedRowsFailsOnGenuineDuplicate(t *testing.T) {
	const (
		tableID      = "table-bulk-genuine-dup"
		version      = 2
		baseRowCount = 3
	)

	proxy := &BusinessRulesDecisionTableProxy{}
	proxy.bulkAddDecisionTableRowsAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, ver int, rows []platformclientv2.Createdecisiontablerowrequest) (*platformclientv2.APIResponse, error) {
		return unitTestDuplicateRowResponse(1), assert.AnError
	}

	err := bulkAddConvertedRows(context.Background(), proxy, tableID, version, []platformclientv2.Createdecisiontablerowrequest{unitTestBulkAddRow()}, 15, baseRowCount)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to bulk add rows")
}

// TestUnitBulkAddConvertedRowsFailsOnRulePersistenceConflict verifies a 409
// RULE_PERSISTENCE_CONFLICT is not treated as a ghost chunk.
func TestUnitBulkAddConvertedRowsFailsOnRulePersistenceConflict(t *testing.T) {
	proxy := &BusinessRulesDecisionTableProxy{}
	proxy.bulkAddDecisionTableRowsAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, ver int, rows []platformclientv2.Createdecisiontablerowrequest) (*platformclientv2.APIResponse, error) {
		return unitTestRulePersistenceConflictResponse(), assert.AnError
	}

	err := bulkAddConvertedRows(context.Background(), proxy, "table-conflict", 1, []platformclientv2.Createdecisiontablerowrequest{unitTestBulkAddRow()}, 15, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to bulk add rows")
}

// TestUnitBulkAddConvertedRowsSkipsGhostChunkOnUpdatePath verifies the update-path
// bulk add uses priorRowCount as the index base (equivalent to the former
// TestUnitApplyRowChangesSkipsGhostRowDuplicate).
func TestUnitBulkAddConvertedRowsSkipsGhostChunkOnUpdatePath(t *testing.T) {
	const (
		tableID       = "table-update-bulk-ghost"
		version       = 2
		priorRowCount = 3
		expectedIndex = priorRowCount + 1
	)

	var bulkCalls int
	proxy := &BusinessRulesDecisionTableProxy{}
	proxy.bulkAddDecisionTableRowsAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, ver int, rows []platformclientv2.Createdecisiontablerowrequest) (*platformclientv2.APIResponse, error) {
		bulkCalls++
		return unitTestDuplicateRowResponse(expectedIndex), assert.AnError
	}

	err := bulkAddConvertedRows(context.Background(), proxy, tableID, version, []platformclientv2.Createdecisiontablerowrequest{unitTestBulkAddRow()}, 15, priorRowCount)
	assert.NoError(t, err)
	assert.Equal(t, 1, bulkCalls)
}
