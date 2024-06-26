package architect_datatable_row

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mitchellh/mapstructure"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *architectDatatableRowProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getArchitectDatatableFunc func(ctx context.Context, p *architectDatatableRowProxy, datatableId string, expanded string) (*Datatable, *platformclientv2.APIResponse, error)
type getAllArchitectDatatableFunc func(ctx context.Context, p *architectDatatableRowProxy) (*[]platformclientv2.Datatable, *platformclientv2.APIResponse, error)
type getAllArchitectDatatableRowsFunc func(ctx context.Context, p *architectDatatableRowProxy, tableId string) (*[]map[string]interface{}, *platformclientv2.APIResponse, error)
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
	dataTableRowCache                rc.CacheInterface[map[string]interface{}]
	dataTableCache                   rc.CacheInterface[Datatable]
}

func newArchitectDatatableRowProxy(clientConfig *platformclientv2.Configuration) *architectDatatableRowProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	dataTableRowCache := rc.NewResourceCache[map[string]interface{}]()
	dataTableCache := rc.NewResourceCache[Datatable]()
	return &architectDatatableRowProxy{
		clientConfig:                     clientConfig,
		architectApi:                     api,
		dataTableRowCache:                dataTableRowCache,
		dataTableCache:                   dataTableCache,
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

func (p *architectDatatableRowProxy) getAllArchitectDatatable(ctx context.Context) (*[]platformclientv2.Datatable, *platformclientv2.APIResponse, error) {
	return p.getAllArchitectDatatableAttr(ctx, p)
}

func (p *architectDatatableRowProxy) getAllArchitectDatatableRows(ctx context.Context, tableId string) (*[]map[string]interface{}, *platformclientv2.APIResponse, error) {
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

func getAllArchitectDatatableFn(_ context.Context, p *architectDatatableRowProxy) (*[]platformclientv2.Datatable, *platformclientv2.APIResponse, error) {
	var totalRecords []platformclientv2.Datatable

	const pageSize = 100
	tables, apiResponse, getErr := p.architectApi.GetFlowsDatatables("", 1, pageSize, "", "", nil, "")
	if getErr != nil {
		return &totalRecords, apiResponse, getErr
	}

	if tables.Entities == nil || len(*tables.Entities) == 0 {
		return &totalRecords, apiResponse, nil
	}

	for _, table := range *tables.Entities {
		totalRecords = append(totalRecords, table)
		rc.SetCache(p.dataTableCache, *table.Id, *ConvertDatatable(table))
	}

	for pageNum := 2; pageNum <= *tables.PageCount; pageNum++ {
		tables, apiResponse, getErr := p.architectApi.GetFlowsDatatables("", pageNum, pageSize, "", "", nil, "")
		if getErr != nil {
			return &totalRecords, apiResponse, getErr
		}

		if tables.Entities == nil || len(*tables.Entities) == 0 {
			break
		}

		for _, table := range *tables.Entities {
			totalRecords = append(totalRecords, table)
			rc.SetCache(p.dataTableCache, *table.Id, *ConvertDatatable(table))
		}
	}
	return &totalRecords, apiResponse, nil
}

func ConvertDatatable(master platformclientv2.Datatable) *Datatable {
	var datatable Datatable
	err := mapstructure.Decode(master, &datatable)
	if err != nil {
		log.Printf("Error converting the DataTable for id %v, error: %v", *master.Id, err)
		return nil
	}
	return &datatable
}

func getArchitectDatatableFn(_ context.Context, p *architectDatatableRowProxy, datatableId string, expanded string) (*Datatable, *platformclientv2.APIResponse, error) {

	eg := rc.GetCacheItem(p.dataTableCache, datatableId)
	if eg != nil {
		return eg, nil, nil
	}

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

func getAllArchitectDatatableRowsFn(_ context.Context, p *architectDatatableRowProxy, tableId string) (*[]map[string]interface{}, *platformclientv2.APIResponse, error) {
	var resources []map[string]interface{}
	const pageSize = 100

	rows, apiResponse, getErr := p.architectApi.GetFlowsDatatableRows(tableId, 1, pageSize, false, "")
	if getErr != nil {
		return nil, apiResponse, getErr
	}

	if rows.Entities == nil || len(*rows.Entities) == 0 {
		return &resources, apiResponse, nil
	}

	for _, row := range *rows.Entities {
		resources = append(resources, row)
		if keyVal, ok := row["key"]; ok {
			rc.SetCache(p.dataTableRowCache, tableId+"_"+keyVal.(string), row)
		}
	}

	for pageNum := 2; pageNum <= *rows.PageCount; pageNum++ {

		rows, apiResponse, getErr := p.architectApi.GetFlowsDatatableRows(tableId, pageNum, pageSize, false, "")
		if getErr != nil {
			return nil, apiResponse, getErr
		}

		if rows.Entities == nil || len(*rows.Entities) == 0 {
			break
		}

		for _, row := range *rows.Entities {
			resources = append(resources, row)
			if keyVal, ok := row["key"]; ok {
				rc.SetCache(p.dataTableRowCache, tableId+"_"+keyVal.(string), row)
			}
		}
	}
	return &resources, apiResponse, nil
}

func getArchitectDataTableRowFn(_ context.Context, p *architectDatatableRowProxy, tableId string, key string) (*map[string]interface{}, *platformclientv2.APIResponse, error) {
	eg := rc.GetCacheItem(p.dataTableRowCache, tableId+"_"+key)
	if eg != nil {
		return eg, nil, nil
	}
	return p.architectApi.GetFlowsDatatableRow(tableId, key, false)
}

func createArchitectDatatableRowFn(_ context.Context, p *architectDatatableRowProxy, tableId string, row *map[string]interface{}) (*map[string]interface{}, *platformclientv2.APIResponse, error) {
	return p.architectApi.PostFlowsDatatableRows(tableId, *row)
}

func updateArchitectDatatableRowFn(_ context.Context, p *architectDatatableRowProxy, tableId string, key string, row *map[string]interface{}) (*map[string]interface{}, *platformclientv2.APIResponse, error) {
	return p.architectApi.PutFlowsDatatableRow(tableId, key, *row)
}

func deleteArchitectDatatableRowFn(_ context.Context, p *architectDatatableRowProxy, tableId string, rowId string) (*platformclientv2.APIResponse, error) {
	return p.architectApi.DeleteFlowsDatatableRow(tableId, rowId)
}
