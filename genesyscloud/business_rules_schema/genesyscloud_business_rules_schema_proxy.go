package business_rules_schema

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

/*
The genesyscloud_business_rules_schema_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *businessRulesSchemaProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createBusinessRulesSchemaFunc func(ctx context.Context, p *businessRulesSchemaProxy, schema *platformclientv2.Dataschema) (*platformclientv2.Dataschema, *platformclientv2.APIResponse, error)
type getAllBusinessRulesSchemaFunc func(ctx context.Context, p *businessRulesSchemaProxy) (*[]platformclientv2.Dataschema, *platformclientv2.APIResponse, error)
type getBusinessRulesSchemasByNameFunc func(ctx context.Context, p *businessRulesSchemaProxy, name string) (schemas *[]platformclientv2.Dataschema, retryable bool, resp *platformclientv2.APIResponse, err error)
type getBusinessRulesSchemaByIdFunc func(ctx context.Context, p *businessRulesSchemaProxy, id string) (schema *platformclientv2.Dataschema, response *platformclientv2.APIResponse, err error)
type updateBusinessRulesSchemaFunc func(ctx context.Context, p *businessRulesSchemaProxy, id string, schema *platformclientv2.Dataschema) (*platformclientv2.Dataschema, *platformclientv2.APIResponse, error)
type deleteBusinessRulesSchemaFunc func(ctx context.Context, p *businessRulesSchemaProxy, id string) (response *platformclientv2.APIResponse, err error)
type getBusinessRulesSchemaDeletedStatusFunc func(ctx context.Context, p *businessRulesSchemaProxy, schemaId string) (isDeleted bool, resp *platformclientv2.APIResponse, err error)

// businessRulesSchemaProxy contains all of the methods that call genesys cloud APIs.
type businessRulesSchemaProxy struct {
	clientConfig                            *platformclientv2.Configuration
	businessRulesApi                        *platformclientv2.BusinessRulesApi
	createBusinessRulesSchemaAttr           createBusinessRulesSchemaFunc
	getAllBusinessRulesSchemaAttr           getAllBusinessRulesSchemaFunc
	getBusinessRulesSchemasByNameAttr       getBusinessRulesSchemasByNameFunc
	getBusinessRulesSchemaByIdAttr          getBusinessRulesSchemaByIdFunc
	updateBusinessRulesSchemaAttr           updateBusinessRulesSchemaFunc
	deleteBusinessRulesSchemaAttr           deleteBusinessRulesSchemaFunc
	getBusinessRulesSchemaDeletedStatusAttr getBusinessRulesSchemaDeletedStatusFunc
	businessRulesSchemaCache                rc.CacheInterface[platformclientv2.Dataschema]
}

// newBusinessRulesSchemaProxy initializes the business rules schema proxy with all of the data needed to communicate with Genesys Cloud
func newBusinessRulesSchemaProxy(clientConfig *platformclientv2.Configuration) *businessRulesSchemaProxy {
	api := platformclientv2.NewBusinessRulesApiWithConfig(clientConfig)
	businessRulesSchemaCache := rc.NewResourceCache[platformclientv2.Dataschema]()

	return &businessRulesSchemaProxy{
		clientConfig:                            clientConfig,
		businessRulesApi:                        api,
		createBusinessRulesSchemaAttr:           createBusinessRulesSchemaFn,
		getAllBusinessRulesSchemaAttr:           getAllBusinessRulesSchemaFn,
		getBusinessRulesSchemasByNameAttr:       getBusinessRulesSchemasByNameFn,
		getBusinessRulesSchemaByIdAttr:          getBusinessRulesSchemaByIdFn,
		updateBusinessRulesSchemaAttr:           updateBusinessRulesSchemaFn,
		deleteBusinessRulesSchemaAttr:           deleteBusinessRulesSchemaFn,
		getBusinessRulesSchemaDeletedStatusAttr: getBusinessRulesSchemaDeletedStatusFn,
		businessRulesSchemaCache:                businessRulesSchemaCache,
	}
}

// getBusinessRulesSchemaProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getBusinessRulesSchemaProxy(clientConfig *platformclientv2.Configuration) *businessRulesSchemaProxy {
	if internalProxy == nil {
		internalProxy = newBusinessRulesSchemaProxy(clientConfig)
	}
	return internalProxy
}

// createBusinessRulesSchema creates a Genesys Cloud business rules schema
func (p *businessRulesSchemaProxy) createBusinessRulesSchema(ctx context.Context, schema *platformclientv2.Dataschema) (*platformclientv2.Dataschema, *platformclientv2.APIResponse, error) {
	return p.createBusinessRulesSchemaAttr(ctx, p, schema)
}

// getAllBusinessRulesSchema retrieves all Genesys Cloud business rules schemas
func (p *businessRulesSchemaProxy) getAllBusinessRulesSchema(ctx context.Context) (*[]platformclientv2.Dataschema, *platformclientv2.APIResponse, error) {
	return p.getAllBusinessRulesSchemaAttr(ctx, p)
}

// getBusinessRulesSchemasByName returns a single Genesys Cloud business rules schema by a name
func (p *businessRulesSchemaProxy) getBusinessRulesSchemasByName(ctx context.Context, name string) (schemas *[]platformclientv2.Dataschema, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getBusinessRulesSchemasByNameAttr(ctx, p, name)
}

// getBusinessRulesSchemaById returns a single Genesys Cloud business rules schema by Id
func (p *businessRulesSchemaProxy) getBusinessRulesSchemaById(ctx context.Context, id string) (schema *platformclientv2.Dataschema, resp *platformclientv2.APIResponse, err error) {
	return p.getBusinessRulesSchemaByIdAttr(ctx, p, id)
}

// updateBusinessRulesSchema updates a Genesys Cloud business rules schema
func (p *businessRulesSchemaProxy) updateBusinessRulesSchema(ctx context.Context, id string, schemaUpdate *platformclientv2.Dataschema) (*platformclientv2.Dataschema, *platformclientv2.APIResponse, error) {
	return p.updateBusinessRulesSchemaAttr(ctx, p, id, schemaUpdate)
}

// deleteBusinessRulesSchema deletes a Genesys Cloud business rules schema by Id
func (p *businessRulesSchemaProxy) deleteBusinessRulesSchema(ctx context.Context, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.deleteBusinessRulesSchemaAttr(ctx, p, id)
}

// getBusinessRulesSchemaDeletedStatus gets the deleted status of a Genesys Cloud business rules schema
func (p *businessRulesSchemaProxy) getBusinessRulesSchemaDeletedStatus(ctx context.Context, schemaId string) (isDeleted bool, resp *platformclientv2.APIResponse, err error) {
	return p.getBusinessRulesSchemaDeletedStatusAttr(ctx, p, schemaId)
}

// createBusinessRulesSchemaFn is an implementation function for creating a Genesys Cloud business rules schema
func createBusinessRulesSchemaFn(ctx context.Context, p *businessRulesSchemaProxy, schema *platformclientv2.Dataschema) (*platformclientv2.Dataschema, *platformclientv2.APIResponse, error) {
	log.Printf("Creating business rules schema: %s", *schema.Name)
	createdSchema, resp, err := p.businessRulesApi.PostBusinessrulesSchemas(*schema)
	log.Printf("Completed call to create business rules schema %s with status code %d, correlation id %s", *schema.Name, resp.StatusCode, resp.CorrelationID)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to create business rules schema: %s", err)
	}
	return createdSchema, resp, nil
}

// getAllBusinessRulesSchemaFn is the implementation for retrieving all business rules schemas in Genesys Cloud
func getAllBusinessRulesSchemaFn(ctx context.Context, p *businessRulesSchemaProxy) (*[]platformclientv2.Dataschema, *platformclientv2.APIResponse, error) {
	// NOTE: At the time of implementation (Preview API) retrieving schemas does not have any sort of pagination.
	// It seemingly will return all schemas in one call. This might have to be updated as there may be some
	// undocumented limit or if there would be changes to the API call before release.

	schemas, resp, err := p.businessRulesApi.GetBusinessrulesSchemas()
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get all business rules schemas: %v", err)
	}
	if schemas.Entities == nil || *schemas.Total == 0 {
		return &([]platformclientv2.Dataschema{}), resp, nil
	}
	return schemas.Entities, resp, nil
}

// getBusinessRulesSchemasByNameFn is an implementation of the function to get a Genesys Cloud business rules schemas by name
func getBusinessRulesSchemasByNameFn(ctx context.Context, p *businessRulesSchemaProxy, name string) (matchingSchemas *[]platformclientv2.Dataschema, retryable bool, resp *platformclientv2.APIResponse, err error) {
	finalSchemas := []platformclientv2.Dataschema{}

	schemas, resp, err := p.getAllBusinessRulesSchema(ctx)
	if err != nil {
		return nil, false, resp, err
	}

	for _, schema := range *schemas {
		if schema.Name != nil && *schema.Name == name {
			finalSchemas = append(finalSchemas, schema)
		}
	}

	if len(finalSchemas) == 0 {
		return nil, true, resp, fmt.Errorf("no business rules schema found with name %s", name)
	}
	return &finalSchemas, false, resp, nil
}

// getBusinessRulesSchemaByIdFn is an implementation of the function to get a Genesys Cloud business rules schema by Id
func getBusinessRulesSchemaByIdFn(ctx context.Context, p *businessRulesSchemaProxy, id string) (schema *platformclientv2.Dataschema, resp *platformclientv2.APIResponse, err error) {
	businessRulesSchema := rc.GetCacheItem(p.businessRulesSchemaCache, id)
	if businessRulesSchema != nil {
		return businessRulesSchema, nil, nil
	}
	return p.businessRulesApi.GetBusinessrulesSchema(id)
}

// updateBusinessRulesSchemaFn is an implementation of the function to update a Genesys Cloud business rules schema
func updateBusinessRulesSchemaFn(ctx context.Context, p *businessRulesSchemaProxy, id string, schemaUpdate *platformclientv2.Dataschema) (*platformclientv2.Dataschema, *platformclientv2.APIResponse, error) {
	schema, resp, err := p.businessRulesApi.PutBusinessrulesSchema(id, *schemaUpdate)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update business rules schema: %s", err)
	}
	return schema, resp, nil
}

// deleteBusinessRulesSchemaFn is an implementation function for deleting a Genesys Cloud business rules schema
func deleteBusinessRulesSchemaFn(ctx context.Context, p *businessRulesSchemaProxy, id string) (resp *platformclientv2.APIResponse, err error) {
	resp, err = p.businessRulesApi.DeleteBusinessrulesSchema(id)
	if err != nil {
		return resp, fmt.Errorf("failed to delete business rules schema: %s", err)
	}
	rc.DeleteCacheItem(p.businessRulesSchemaCache, id)
	return resp, nil
}

// getBusinessRulesSchemaDeletedStatusFn is an implementation function to get the 'deleted' status of a Genesys Cloud business rules schema
func getBusinessRulesSchemaDeletedStatusFn(ctx context.Context, p *businessRulesSchemaProxy, schemaId string) (isDeleted bool, resp *platformclientv2.APIResponse, err error) {
	apiClient := &p.clientConfig.APIClient
	// create path and map variables
	path := p.clientConfig.BasePath + "/api/v2/businessrules/schemas/" + schemaId

	headerParams := make(map[string]string)
	queryParams := make(map[string]string)

	// oauth required
	if p.clientConfig.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + p.clientConfig.AccessToken
	}
	// add default headers if any
	for key := range p.clientConfig.DefaultHeader {
		headerParams[key] = p.clientConfig.DefaultHeader[key]
	}

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload map[string]interface{}
	response, err := apiClient.CallAPI(path, http.MethodGet, nil, headerParams, queryParams, nil, "", nil, "")
	if err != nil {
		return false, response, fmt.Errorf("failed to get business rules schema %s: %v", schemaId, err)
	}
	if response.Error != nil {
		return false, response, fmt.Errorf("failed to get business rules schema %s: %v", schemaId, errors.New(response.ErrorMessage))
	}

	err = json.Unmarshal([]byte(response.RawBody), &successPayload)
	if err != nil {
		return false, response, fmt.Errorf("failed to get deleted status of %s: %v", schemaId, err)
	}

	// Manually query for the 'deleted' property because it is removed when
	// response JSON body becomes SDK Dataschema object.
	if isDeleted, ok := successPayload["deleted"].(bool); ok {
		return isDeleted, response, nil
	}

	return false, response, fmt.Errorf("failed to get deleted status of %s: %v", schemaId, err)
}
