package architect_datatable_row

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *architectDatatableRowProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getArchitectDatatableFunc func(ctx context.Context, p *architectDatatableRowProxy, datatableId string, expanded string) (*Datatable, *platformclientv2.APIResponse, error)
type getAllArchitectDatatableFunc func(ctx context.Context, p *architectDatatableRowProxy) (*[]platformclientv2.Datatable, error)
type getAllArchitectDatatableRowsFunc func(ctx context.Context, p *architectDatatableRowProxy, tableId string) (*[]map[string]interface{}, error)
type getArchitectDatatableRowFunc func(ctx context.Context, p *architectDatatableRowProxy, tableId string, key string) (*map[string]interface{}, *platformclientv2.APIResponse, error)
type createArchitectDatatableRowFunc func(ctx context.Context, p *architectDatatableRowProxy, tableId string, row *map[string]interface{}) (*map[string]interface{}, *platformclientv2.APIResponse, error)
type updateArchitectDatatableRowFunc func(ctx context.Context, p *architectDatatableRowProxy, tableId string, key string, row *map[string]interface{}) (*map[string]interface{}, *platformclientv2.APIResponse, error)
type deleteArchitectDatatableRowFunc func(ctx context.Context, p *architectDatatableRowProxy, tableId string, rowId string) (*platformclientv2.APIResponse, error)

type architectDatatableRowProxy struct {
	clientConfig                     *platformclientv2.Configuration
	architectApi                     *platformclientv2.ArchitectApi
	createArchitectDatatableRowAttr  createArchitectDatatableRowFunc
	getArchitectDatatableAttr        getArchitectDatatableFunc
	getAllArchitectDatatableAttr     getAllArchitectDatatableFunc
	getAllArchitectDatatableRowsAttr getAllArchitectDatatableRowsFunc
	getArchitectDatatableRowAttr     getArchitectDatatableRowFunc
	updateArchitectDatatableRowAttr  updateArchitectDatatableRowFunc
	deleteArchitectDatatableRowAttr  deleteArchitectDatatableRowFunc
}

func newArchitectDatatableRowProxy(clientConfig *platformclientv2.Configuration) *architectDatatableRowProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	return &architectDatatableRowProxy{
		clientConfig:                     clientConfig,
		architectApi:                     api,
		getArchitectDatatableAttr:        getArchitectDatatableFn,
		getAllArchitectDatatableAttr:     getAllArchitectDatatableFn,
		getAllArchitectDatatableRowsAttr: getAllArchitectDatatableRowsFn,
		getArchitectDatatableRowAttr:     getArchitectDataTableRowFn,
		createArchitectDatatableRowAttr:  createArchitectDatatableRowFn,
		updateArchitectDatatableRowAttr:  updateArchitectDatatableRowFn,
		deleteArchitectDatatableRowAttr:  deleteArchitectDatatableRowFn,
	}
}

func getArchitectDatatableRowProxy(clientConfig *platformclientv2.Configuration) *architectDatatableRowProxy {
	if internalProxy == nil {
		internalProxy = newArchitectDatatableRowProxy(clientConfig)
	}

	return internalProxy
}

func (p *architectDatatableRowProxy) getArchitectDatatable(ctx context.Context, id string, expanded string) (*Datatable, *platformclientv2.APIResponse, error) {
	return p.getArchitectDatatableAttr(ctx, p, id, expanded)
}

func (p *architectDatatableRowProxy) getAllArchitectDatatable(ctx context.Context) (*[]platformclientv2.Datatable, error) {
	return p.getAllArchitectDatatableAttr(ctx, p)
}

func (p *architectDatatableRowProxy) getAllArchitectDatatableRows(ctx context.Context, tableId string) (*[]map[string]interface{}, error) {
	return p.getAllArchitectDatatableRowsAttr(ctx, p, tableId)
}

func (p *architectDatatableRowProxy) getArchitectDatatableRow(ctx context.Context, tableId string, key string) (*map[string]interface{}, *platformclientv2.APIResponse, error) {
	return p.getArchitectDatatableRowAttr(ctx, p, tableId, key)
}

func (p *architectDatatableRowProxy) createArchitectDatatableRow(ctx context.Context, tableId string, row *map[string]interface{}) (*map[string]interface{}, *platformclientv2.APIResponse, error) {
	return p.createArchitectDatatableRowAttr(ctx, p, tableId, row)
}

func (p *architectDatatableRowProxy) updateArchitectDatatableRow(ctx context.Context, tableId string, key string, row *map[string]interface{}) (*map[string]interface{}, *platformclientv2.APIResponse, error) {
	return p.updateArchitectDatatableRowAttr(ctx, p, tableId, key, row)
}

func (p *architectDatatableRowProxy) deleteArchitectDatatableRow(ctx context.Context, tableId string, rowId string) (*platformclientv2.APIResponse, error) {
	return p.deleteArchitectDatatableRowAttr(ctx, p, tableId, rowId)
}

func getAllArchitectDatatableFn(ctx context.Context, p *architectDatatableRowProxy) (*[]platformclientv2.Datatable, error) {
	var totalRecords []platformclientv2.Datatable

	const pageSize = 100
	tables, _, getErr := p.architectApi.GetFlowsDatatables("", 1, pageSize, "", "", nil, "")
	if getErr != nil {
		return &totalRecords, getErr
	}

	if tables.Entities == nil || len(*tables.Entities) == 0 {
		return &totalRecords, nil
	}

	for _, table := range *tables.Entities {
		totalRecords = append(totalRecords, table)
	}

	for pageNum := 2; pageNum <= *tables.PageCount; pageNum++ {
		tables, _, getErr := p.architectApi.GetFlowsDatatables("", pageNum, pageSize, "", "", nil, "")
		if getErr != nil {
			return &totalRecords, getErr
		}

		if tables.Entities == nil || len(*tables.Entities) == 0 {
			break
		}

		for _, table := range *tables.Entities {
			totalRecords = append(totalRecords, table)
		}
	}

	return &totalRecords, nil
}

func getArchitectDatatableFn(ctx context.Context, p *architectDatatableRowProxy, datatableId string, expanded string) (*Datatable, *platformclientv2.APIResponse, error) {
	apiClient := &p.architectApi.Configuration.APIClient

	// create path and map variables
	path := p.architectApi.Configuration.BasePath + "/api/v2/flows/datatables/" + datatableId

	headerParams := make(map[string]string)
	queryParams := make(map[string]string)

	// oauth required
	if p.architectApi.Configuration.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + p.architectApi.Configuration.AccessToken
	}
	// add default headers if any
	for key := range p.architectApi.Configuration.DefaultHeader {
		headerParams[key] = p.architectApi.Configuration.DefaultHeader[key]
	}

	queryParams["expand"] = apiClient.ParameterToString(expanded, "")

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload *Datatable
	response, err := apiClient.CallAPI(path, http.MethodGet, nil, headerParams, queryParams, nil, "", nil)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		err = json.Unmarshal(response.RawBody, &successPayload)
	}
	return successPayload, response, err
}

func getAllArchitectDatatableRowsFn(ctx context.Context, p *architectDatatableRowProxy, tableId string) (*[]map[string]interface{}, error) {
	var resources []map[string]interface{}
	const pageSize = 100

	rows, _, getErr := p.architectApi.GetFlowsDatatableRows(tableId, 1, pageSize, false, "")
	if getErr != nil {
		return nil, getErr
	}

	if rows.Entities == nil || len(*rows.Entities) == 0 {
		return &resources, nil
	}

	for _, row := range *rows.Entities {
		resources = append(resources, row)
	}

	for pageNum := 2; pageNum <= *rows.PageCount; pageNum++ {

		rows, _, getErr := p.architectApi.GetFlowsDatatableRows(tableId, pageNum, pageSize, false, "")
		if getErr != nil {
			return nil, getErr
		}

		if rows.Entities == nil || len(*rows.Entities) == 0 {
			break
		}

		for _, row := range *rows.Entities {
			resources = append(resources, row)
		}
	}

	return &resources, nil
}

func getArchitectDataTableRowFn(ctx context.Context, p *architectDatatableRowProxy, tableId string, key string) (*map[string]interface{}, *platformclientv2.APIResponse, error) {
	return p.architectApi.GetFlowsDatatableRow(tableId, key, false)
}

func createArchitectDatatableRowFn(ctx context.Context, p *architectDatatableRowProxy, tableId string, row *map[string]interface{}) (*map[string]interface{}, *platformclientv2.APIResponse, error) {
	return p.architectApi.PostFlowsDatatableRows(tableId, *row)
}

func updateArchitectDatatableRowFn(ctx context.Context, p *architectDatatableRowProxy, tableId string, key string, row *map[string]interface{}) (*map[string]interface{}, *platformclientv2.APIResponse, error) {
	return p.architectApi.PutFlowsDatatableRow(tableId, key, *row)
}

func deleteArchitectDatatableRowFn(ctx context.Context, p *architectDatatableRowProxy, tableId string, rowId string) (*platformclientv2.APIResponse, error) {
	return p.architectApi.DeleteFlowsDatatableRow(tableId, rowId)
}
