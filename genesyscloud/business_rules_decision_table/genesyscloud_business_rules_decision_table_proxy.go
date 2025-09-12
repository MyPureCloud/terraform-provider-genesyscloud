package business_rules_decision_table

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
The genesyscloud_business_rules_decision_table_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

var businessRulesDecisionTableCache = rc.NewResourceCache[platformclientv2.Decisiontable]()
var internalProxy *BusinessRulesDecisionTableProxy

// Function type definitions for composition pattern
type createBusinessRulesDecisionTableFunc func(ctx context.Context, p *BusinessRulesDecisionTableProxy, createRequest *platformclientv2.Createdecisiontablerequest) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error)
type getBusinessRulesDecisionTableFunc func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error)
type updateBusinessRulesDecisionTableFunc func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, updateRequest *platformclientv2.Updatedecisiontablerequest) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error)
type deleteBusinessRulesDecisionTableFunc func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.APIResponse, error)
type getAllBusinessRulesDecisionTablesFunc func(ctx context.Context, p *BusinessRulesDecisionTableProxy, name string) (*platformclientv2.Decisiontablelisting, *platformclientv2.APIResponse, error)
type getBusinessRulesDecisionTablesByNameFunc func(ctx context.Context, p *BusinessRulesDecisionTableProxy, name string) (tables *[]platformclientv2.Decisiontable, retryable bool, resp *platformclientv2.APIResponse, err error)
type getBusinessRulesDecisionTableVersionFunc func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, versionNumber int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error)

// BusinessRulesDecisionTableProxy contains all the methods that call genesys cloud APIs.
type BusinessRulesDecisionTableProxy struct {
	clientConfig     *platformclientv2.Configuration
	businessRulesApi *platformclientv2.BusinessRulesApi

	createBusinessRulesDecisionTableAttr     createBusinessRulesDecisionTableFunc
	getBusinessRulesDecisionTableAttr        getBusinessRulesDecisionTableFunc
	updateBusinessRulesDecisionTableAttr     updateBusinessRulesDecisionTableFunc
	deleteBusinessRulesDecisionTableAttr     deleteBusinessRulesDecisionTableFunc
	getAllBusinessRulesDecisionTablesAttr    getAllBusinessRulesDecisionTablesFunc
	getBusinessRulesDecisionTablesByNameAttr getBusinessRulesDecisionTablesByNameFunc
	getBusinessRulesDecisionTableVersionAttr getBusinessRulesDecisionTableVersionFunc

	BusinessRulesDecisionTableCache rc.CacheInterface[platformclientv2.Decisiontable]

	// Provider fields for testing
	queueLookupProvider  QueueLookupProvider
	schemaLookupProvider SchemaLookupProvider
}

// newBusinessRulesDecisionTableProxy initializes the business rules decision table proxy with all the data needed to communicate with Genesys Cloud
func newBusinessRulesDecisionTableProxy(clientConfig *platformclientv2.Configuration) *BusinessRulesDecisionTableProxy {
	api := platformclientv2.NewBusinessRulesApiWithConfig(clientConfig)

	return &BusinessRulesDecisionTableProxy{
		clientConfig:     clientConfig,
		businessRulesApi: api,

		createBusinessRulesDecisionTableAttr:     createBusinessRulesDecisionTableFn,
		getBusinessRulesDecisionTableAttr:        getBusinessRulesDecisionTableFn,
		updateBusinessRulesDecisionTableAttr:     updateBusinessRulesDecisionTableFn,
		deleteBusinessRulesDecisionTableAttr:     deleteBusinessRulesDecisionTableFn,
		getAllBusinessRulesDecisionTablesAttr:    getAllBusinessRulesDecisionTablesFn,
		getBusinessRulesDecisionTablesByNameAttr: getBusinessRulesDecisionTablesByNameFn,
		getBusinessRulesDecisionTableVersionAttr: getBusinessRulesDecisionTableVersionFn,

		BusinessRulesDecisionTableCache: businessRulesDecisionTableCache,
	}
}

// getBusinessRulesDecisionTableProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getBusinessRulesDecisionTableProxy(clientConfig *platformclientv2.Configuration) *BusinessRulesDecisionTableProxy {
	if internalProxy == nil {
		internalProxy = newBusinessRulesDecisionTableProxy(clientConfig)
	}
	return internalProxy
}

// Method implementations that delegate to the function attributes
func (p *BusinessRulesDecisionTableProxy) createBusinessRulesDecisionTable(ctx context.Context, createRequest *platformclientv2.Createdecisiontablerequest) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
	return p.createBusinessRulesDecisionTableAttr(ctx, p, createRequest)
}

func (p *BusinessRulesDecisionTableProxy) getBusinessRulesDecisionTable(ctx context.Context, tableId string) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
	return p.getBusinessRulesDecisionTableAttr(ctx, p, tableId)
}

func (p *BusinessRulesDecisionTableProxy) updateBusinessRulesDecisionTable(ctx context.Context, tableId string, updateRequest *platformclientv2.Updatedecisiontablerequest) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
	return p.updateBusinessRulesDecisionTableAttr(ctx, p, tableId, updateRequest)
}

func (p *BusinessRulesDecisionTableProxy) deleteBusinessRulesDecisionTable(ctx context.Context, tableId string) (*platformclientv2.APIResponse, error) {
	return p.deleteBusinessRulesDecisionTableAttr(ctx, p, tableId)
}

func (p *BusinessRulesDecisionTableProxy) getAllBusinessRulesDecisionTables(ctx context.Context, name string) (*platformclientv2.Decisiontablelisting, *platformclientv2.APIResponse, error) {
	return p.getAllBusinessRulesDecisionTablesAttr(ctx, p, name)
}

// getBusinessRulesDecisionTablesByName returns Genesys Cloud business rules decision tables by name
func (p *BusinessRulesDecisionTableProxy) getBusinessRulesDecisionTablesByName(ctx context.Context, name string) (tables *[]platformclientv2.Decisiontable, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getBusinessRulesDecisionTablesByNameAttr(ctx, p, name)
}

// getBusinessRulesDecisionTableVersion retrieves a specific decision table version
func (p *BusinessRulesDecisionTableProxy) getBusinessRulesDecisionTableVersion(ctx context.Context, tableId string, versionNumber int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
	return p.getBusinessRulesDecisionTableVersionAttr(ctx, p, tableId, versionNumber)
}

// Provider getter and setter methods for testing
func (p *BusinessRulesDecisionTableProxy) GetQueueLookupProvider() QueueLookupProvider {
	return p.queueLookupProvider
}

func (p *BusinessRulesDecisionTableProxy) SetQueueLookupProvider(provider QueueLookupProvider) {
	p.queueLookupProvider = provider
}

func (p *BusinessRulesDecisionTableProxy) GetSchemaLookupProvider() SchemaLookupProvider {
	return p.schemaLookupProvider
}

func (p *BusinessRulesDecisionTableProxy) SetSchemaLookupProvider(provider SchemaLookupProvider) {
	p.schemaLookupProvider = provider
}

// Function implementations that make the actual API calls
func createBusinessRulesDecisionTableFn(ctx context.Context, p *BusinessRulesDecisionTableProxy, createRequest *platformclientv2.Createdecisiontablerequest) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
	return p.businessRulesApi.PostBusinessrulesDecisiontables(*createRequest)
}

func getBusinessRulesDecisionTableFn(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
	// Check cache first
	businessRulesDecisionTable := rc.GetCacheItem(p.BusinessRulesDecisionTableCache, tableId)
	if businessRulesDecisionTable != nil {
		return businessRulesDecisionTable, nil, nil
	}

	// If not in cache, make API call
	table, resp, err := p.businessRulesApi.GetBusinessrulesDecisiontable(tableId)
	if err == nil && table != nil {
		// Cache the successful response (dereference pointer to store value)
		rc.SetCache(p.BusinessRulesDecisionTableCache, tableId, *table)
	}
	return table, resp, err
}

func updateBusinessRulesDecisionTableFn(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, updateRequest *platformclientv2.Updatedecisiontablerequest) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
	table, resp, err := p.businessRulesApi.PatchBusinessrulesDecisiontable(tableId, *updateRequest)
	if err == nil && table != nil {
		// Update cache with new data after successful update (dereference pointer to store value)
		rc.SetCache(p.BusinessRulesDecisionTableCache, tableId, *table)
	}
	return table, resp, err
}

func deleteBusinessRulesDecisionTableFn(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.APIResponse, error) {
	resp, err := p.businessRulesApi.DeleteBusinessrulesDecisiontable(tableId, false)
	if err == nil {
		// Remove from cache after successful deletion
		rc.DeleteCacheItem(p.BusinessRulesDecisionTableCache, tableId)
	}
	return resp, err
}

func getAllBusinessRulesDecisionTablesFn(ctx context.Context, p *BusinessRulesDecisionTableProxy, name string) (*platformclientv2.Decisiontablelisting, *platformclientv2.APIResponse, error) {
	var allTables []platformclientv2.Decisiontable
	pageSize := "100"
	after := ""
	var response *platformclientv2.APIResponse

	for {
		// API signature: GetBusinessrulesDecisiontables(after string, pageSize string, divisionIds []string, name string)
		tables, resp, err := p.businessRulesApi.GetBusinessrulesDecisiontables(after, pageSize, nil, name)
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get business rules decision tables: %v", err)
		}
		response = resp

		if tables.Entities != nil {
			allTables = append(allTables, *tables.Entities...)
		}

		// Check if there are more pages by looking at NextUri
		// If NextUri is nil or empty, we're on the last page
		if tables.NextUri == nil || *tables.NextUri == "" {
			break
		}

		// Extract the 'after' parameter from NextUri for the next iteration
		after, err = util.GetQueryParamValueFromUri(*tables.NextUri, "after")
		if err != nil {
			return nil, resp, fmt.Errorf("unable to parse after cursor from decision tables next uri: %v", err)
		}
		if after == "" {
			break
		}
	}

	// Create a new Decisiontablelisting with all collected tables
	result := &platformclientv2.Decisiontablelisting{
		Entities: &allTables,
	}

	return result, response, nil
}

// getBusinessRulesDecisionTablesByNameFn is an implementation of the function to get Genesys Cloud business rules decision tables by name
func getBusinessRulesDecisionTablesByNameFn(ctx context.Context, p *BusinessRulesDecisionTableProxy, name string) (matchingTables *[]platformclientv2.Decisiontable, retryable bool, resp *platformclientv2.APIResponse, err error) {
	finalTables := []platformclientv2.Decisiontable{}

	// Use the updated getAll function with name parameter for server-side filtering
	tables, resp, err := getAllBusinessRulesDecisionTablesFn(ctx, p, name)
	if err != nil {
		return nil, false, resp, err
	}

	if tables.Entities == nil {
		return &finalTables, true, resp, nil
	}

	// Filter for exact name matches (API does contains search, we need exact)
	for _, table := range *tables.Entities {
		if table.Name != nil && *table.Name == name {
			finalTables = append(finalTables, table)
		}
	}

	if len(finalTables) == 0 {
		return &finalTables, true, resp, nil
	}

	return &finalTables, false, resp, nil
}

func getBusinessRulesDecisionTableVersionFn(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, versionNumber int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
	return p.businessRulesApi.GetBusinessrulesDecisiontableVersion(tableId, versionNumber)
}
