package business_rules_decision_table

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mypurecloud/platform-client-sdk-go/v191/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

// Limits from Genesys Cloud Public API for decision table row bulk operations.
const (
	maxBulkDecisionTableRowsAdd    = 15
	maxBulkDecisionTableRowsUpdate = 15
	maxBulkDecisionTableRowsRemove = 49
)

func getBulkChunkLimits() (add, update, remove int) {
	return maxBulkDecisionTableRowsAdd, maxBulkDecisionTableRowsUpdate, maxBulkDecisionTableRowsRemove
}

func bulkAddDecisionTableRowsFn(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, rows []platformclientv2.Createdecisiontablerowrequest) (*platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	if len(rows) == 0 {
		return &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	if len(rows) > maxBulkDecisionTableRowsAdd {
		return nil, fmt.Errorf("bulk add exceeds max of %d rows per request (got %d)", maxBulkDecisionTableRowsAdd, len(rows))
	}

	body := platformclientv2.Bulkadddecisiontablerowsrequest{Rows: &rows}
	result, resp, err := p.businessRulesApi.PostBusinessrulesDecisiontableVersionRowsBulkAdd(tableId, version, body)
	if err != nil {
		return resp, err
	}
	if result != nil && result.TotalCreated != nil && *result.TotalCreated != len(rows) {
		return resp, fmt.Errorf("bulk add returned totalCreated=%d, expected %d", *result.TotalCreated, len(rows))
	}
	return resp, nil
}

func bulkRemoveDecisionTableRowsFn(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, rowIds []string) (*platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	if len(rowIds) == 0 {
		return &platformclientv2.APIResponse{StatusCode: http.StatusNoContent}, nil
	}
	if len(rowIds) > maxBulkDecisionTableRowsRemove {
		return nil, fmt.Errorf("bulk remove exceeds max of %d row IDs per request (got %d)", maxBulkDecisionTableRowsRemove, len(rowIds))
	}

	body := platformclientv2.Bulkdeletedecisiontablerowsrequest{RowIds: &rowIds}
	return p.businessRulesApi.PostBusinessrulesDecisiontableVersionRowsBulkRemove(tableId, version, body)
}

func bulkUpdateDecisionTableRowsFn(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, rows []platformclientv2.Row) (*platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	if len(rows) == 0 {
		return &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	if len(rows) > maxBulkDecisionTableRowsUpdate {
		return nil, fmt.Errorf("bulk update exceeds max of %d rows per request (got %d)", maxBulkDecisionTableRowsUpdate, len(rows))
	}

	body := platformclientv2.Bulkupdatedecisiontablerowsrequest{Rows: &rows}
	result, resp, err := p.businessRulesApi.PostBusinessrulesDecisiontableVersionRowsBulkUpdate(tableId, version, body)
	if err != nil {
		return resp, err
	}
	if result != nil && result.TotalUpdated != nil && *result.TotalUpdated != len(rows) {
		return resp, fmt.Errorf("bulk update returned totalUpdated=%d, expected %d", *result.TotalUpdated, len(rows))
	}
	return resp, nil
}
