package architect_datatable

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *architectDatatableProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createOrUpdateArchitectDatatableFunc func(ctx context.Context, p *architectDatatableProxy, createAction bool, datatable *Datatable) (*Datatable, *platformclientv2.APIResponse, error)
type deleteArchitectDatatableFunc func(ctx context.Context, p *architectDatatableProxy, datatableId string) (*platformclientv2.APIResponse, error)
type getArchitectDatatableFunc func(ctx context.Context, p *architectDatatableProxy, datatableId string, expanded string) (*Datatable, *platformclientv2.APIResponse, error)
type getAllArchitectDatatableFunc func(ctx context.Context, p *architectDatatableProxy) (*[]platformclientv2.Datatable, *platformclientv2.APIResponse, error)

type architectDatatableProxy struct {
	clientConfig                         *platformclientv2.Configuration
	architectApi                         *platformclientv2.ArchitectApi
	createOrUpdateArchitectDatatableAttr createOrUpdateArchitectDatatableFunc
	getArchitectDatatableAttr            getArchitectDatatableFunc
	getAllArchitectDatatableAttr         getAllArchitectDatatableFunc
	deleteArchitectDatatableAttr         deleteArchitectDatatableFunc
}

func newArchitectDatatableProxy(clientConfig *platformclientv2.Configuration) *architectDatatableProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	return &architectDatatableProxy{
		clientConfig:                         clientConfig,
		architectApi:                         api,
		createOrUpdateArchitectDatatableAttr: createOrUpdateArchitectDatatableFn,
		getArchitectDatatableAttr:            getArchitectDatatableFn,
		getAllArchitectDatatableAttr:         getAllArchitectDatatableFn,
		deleteArchitectDatatableAttr:         deleteArchitectDatatableFn,
	}
}

func getArchitectDatatableProxy(clientConfig *platformclientv2.Configuration) *architectDatatableProxy {
	if internalProxy == nil {
		internalProxy = newArchitectDatatableProxy(clientConfig)
	}

	return internalProxy
}

func (p *architectDatatableProxy) createArchitectDatatable(ctx context.Context, datatable *Datatable) (*Datatable, *platformclientv2.APIResponse, error) {
	return p.createOrUpdateArchitectDatatableAttr(ctx, p, true, datatable)
}

func (p *architectDatatableProxy) updateArchitectDatatable(ctx context.Context, datatable *Datatable) (*Datatable, *platformclientv2.APIResponse, error) {
	return p.createOrUpdateArchitectDatatableAttr(ctx, p, false, datatable)
}

func (p *architectDatatableProxy) getArchitectDatatable(ctx context.Context, id string, expanded string) (*Datatable, *platformclientv2.APIResponse, error) {
	return p.getArchitectDatatableAttr(ctx, p, id, expanded)
}

func (p *architectDatatableProxy) getAllArchitectDatatable(ctx context.Context) (*[]platformclientv2.Datatable, *platformclientv2.APIResponse, error) {
	return p.getAllArchitectDatatableAttr(ctx, p)
}

func (p *architectDatatableProxy) deleteArchitectDatatable(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteArchitectDatatableAttr(ctx, p, id)
}

func createOrUpdateArchitectDatatableFn(ctx context.Context, p *architectDatatableProxy, createAction bool, datatable *Datatable) (*Datatable, *platformclientv2.APIResponse, error) {
	apiClient := &p.architectApi.Configuration.APIClient
	action := http.MethodPost

	// create path and map variables
	path := p.architectApi.Configuration.BasePath + "/api/v2/flows/datatables"

	if !createAction {
		action = http.MethodPut
		path += "/" + *datatable.Id
	}

	headerParams := make(map[string]string)

	// add default headers if any
	for key := range p.architectApi.Configuration.DefaultHeader {
		headerParams[key] = p.architectApi.Configuration.DefaultHeader[key]
	}

	headerParams["Authorization"] = "Bearer " + p.architectApi.Configuration.AccessToken
	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload *Datatable
	response, err := apiClient.CallAPI(path, action, datatable, headerParams, nil, nil, "", nil)

	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		err = json.Unmarshal([]byte(response.RawBody), &successPayload)
	}

	return successPayload, response, err
}

func getArchitectDatatableFn(ctx context.Context, p *architectDatatableProxy, datatableId string, expanded string) (*Datatable, *platformclientv2.APIResponse, error) {
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

func deleteArchitectDatatableFn(ctx context.Context, p *architectDatatableProxy, datatableId string) (*platformclientv2.APIResponse, error) {
	return p.architectApi.DeleteFlowsDatatable(datatableId, true)
}

func getAllArchitectDatatableFn(ctx context.Context, p *architectDatatableProxy) (*[]platformclientv2.Datatable, *platformclientv2.APIResponse, error) {
	var totalRecords []platformclientv2.Datatable

	const pageSize = 100
	tables, resp, getErr := p.architectApi.GetFlowsDatatables("", 1, pageSize, "", "", nil, "")
	if getErr != nil {
		return &totalRecords, resp, getErr
	}

	if tables.Entities == nil || len(*tables.Entities) == 0 {
		return &totalRecords, resp, nil
	}

	totalRecords = append(totalRecords, *tables.Entities...)

	for pageNum := 2; pageNum <= *tables.PageCount; pageNum++ {
		tables, resp, getErr := p.architectApi.GetFlowsDatatables("", pageNum, pageSize, "", "", nil, "")
		if getErr != nil {
			return &totalRecords, resp, getErr
		}

		if tables.Entities == nil || len(*tables.Entities) == 0 {
			break
		}

		totalRecords = append(totalRecords, *tables.Entities...)
	}

	return &totalRecords, resp, nil
}
