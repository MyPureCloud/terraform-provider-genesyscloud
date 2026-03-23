package architect_datatable

import (
	"context"

	customapi "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/custom_api_client"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
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
	customApiClient                      *customapi.Client
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
		customApiClient:                      customapi.NewClient(clientConfig, ResourceType),
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
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	method := customapi.MethodPost
	path := "/api/v2/flows/datatables"

	if !createAction {
		method = customapi.MethodPut
		path += "/" + *datatable.Id
	}

	return customapi.Do[Datatable](ctx, p.customApiClient, method, path, datatable, nil)
}

func getArchitectDatatableFn(ctx context.Context, p *architectDatatableProxy, datatableId string, expanded string) (*Datatable, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	queryParams := customapi.NewQueryParams(map[string]string{"expand": expanded})

	return customapi.Do[Datatable](ctx, p.customApiClient, customapi.MethodGet, "/api/v2/flows/datatables/"+datatableId, nil, queryParams)
}

func deleteArchitectDatatableFn(ctx context.Context, p *architectDatatableProxy, datatableId string) (*platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	return p.architectApi.DeleteFlowsDatatable(datatableId, true)
}

func getAllArchitectDatatableFn(ctx context.Context, p *architectDatatableProxy) (*[]platformclientv2.Datatable, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

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
