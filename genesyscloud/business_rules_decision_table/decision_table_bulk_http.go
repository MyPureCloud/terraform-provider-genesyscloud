package business_rules_decision_table

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

// Limits from Genesys Cloud Public API for decision table row bulk operations.
const (
	maxBulkDecisionTableRowsAdd    = 15
	maxBulkDecisionTableRowsUpdate = 15
	maxBulkDecisionTableRowsRemove = 49
)

type bulkAddDecisionTableRowsBody struct {
	Rows []platformclientv2.Createdecisiontablerowrequest `json:"rows"`
}

type bulkRemoveDecisionTableRowsBody struct {
	RowIds []string `json:"rowIds"`
}

type bulkUpdateDecisionTableRowBody struct {
	RowId   string                                                      `json:"rowId"`
	Inputs  *map[string]platformclientv2.Decisiontablerowparametervalue `json:"inputs,omitempty"`
	Outputs *map[string]platformclientv2.Decisiontablerowparametervalue `json:"outputs,omitempty"`
}

type bulkUpdateDecisionTableRowsBody struct {
	Rows []bulkUpdateDecisionTableRowBody `json:"rows"`
}

func decisionTableRowsBulkPath(tableId string, version int, suffix string) string {
	v := strconv.Itoa(version)
	return "/api/v2/businessrules/decisiontables/" + tableId + "/versions/" + v + "/rows/bulk/" + suffix
}

func callDecisionTableBulkAPI(ctx context.Context, p *BusinessRulesDecisionTableProxy, method string, path string, body interface{}) (*platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	apiClient := &p.clientConfig.APIClient
	fullPath := p.clientConfig.BasePath + path

	headerParams := make(map[string]string)
	for key := range p.clientConfig.DefaultHeader {
		headerParams[key] = p.clientConfig.DefaultHeader[key]
	}
	headerParams["Authorization"] = "Bearer " + p.clientConfig.AccessToken
	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	response, err := apiClient.CallAPI(fullPath, method, body, headerParams, nil, nil, "", nil, "")
	if err != nil {
		return response, err
	}
	if response == nil {
		return nil, errors.New("nil API response from bulk decision table rows call")
	}
	if response.Error != nil {
		return response, errors.New(response.ErrorMessage)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		msg := string(response.RawBody)
		if msg == "" {
			msg = response.Status
		}
		return response, fmt.Errorf("decision table rows bulk API returned status %d: %s", response.StatusCode, msg)
	}
	return response, nil
}

func bulkAddDecisionTableRowsFn(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, rows []platformclientv2.Createdecisiontablerowrequest) (*platformclientv2.APIResponse, error) {
	if len(rows) == 0 {
		return &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	if len(rows) > maxBulkDecisionTableRowsAdd {
		return nil, fmt.Errorf("bulk add exceeds max of %d rows per request (got %d)", maxBulkDecisionTableRowsAdd, len(rows))
	}
	body := bulkAddDecisionTableRowsBody{Rows: rows}
	return callDecisionTableBulkAPI(ctx, p, http.MethodPost, decisionTableRowsBulkPath(tableId, version, "add"), body)
}

func bulkRemoveDecisionTableRowsFn(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, rowIds []string) (*platformclientv2.APIResponse, error) {
	if len(rowIds) == 0 {
		return &platformclientv2.APIResponse{StatusCode: http.StatusNoContent}, nil
	}
	if len(rowIds) > maxBulkDecisionTableRowsRemove {
		return nil, fmt.Errorf("bulk remove exceeds max of %d row IDs per request (got %d)", maxBulkDecisionTableRowsRemove, len(rowIds))
	}
	body := bulkRemoveDecisionTableRowsBody{RowIds: rowIds}
	return callDecisionTableBulkAPI(ctx, p, http.MethodPost, decisionTableRowsBulkPath(tableId, version, "remove"), body)
}

func bulkUpdateDecisionTableRowsFn(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, rows []bulkUpdateDecisionTableRowBody) (*platformclientv2.APIResponse, error) {
	if len(rows) == 0 {
		return &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	if len(rows) > maxBulkDecisionTableRowsUpdate {
		return nil, fmt.Errorf("bulk update exceeds max of %d rows per request (got %d)", maxBulkDecisionTableRowsUpdate, len(rows))
	}
	body := bulkUpdateDecisionTableRowsBody{Rows: rows}
	return callDecisionTableBulkAPI(ctx, p, http.MethodPost, decisionTableRowsBulkPath(tableId, version, "update"), body)
}
