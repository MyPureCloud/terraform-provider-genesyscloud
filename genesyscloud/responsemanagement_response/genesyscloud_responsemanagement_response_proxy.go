package responsemanagement_response

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_responsemanagement_response_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *responsemanagementResponseProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createResponsemanagementResponseFunc func(ctx context.Context, p *responsemanagementResponseProxy, response *platformclientv2.Response) (responseManagementResponse *platformclientv2.Response, resp *platformclientv2.APIResponse, err error)
type getAllResponsemanagementResponseFunc func(ctx context.Context, p *responsemanagementResponseProxy, libraryId string) (*[]platformclientv2.Response, *platformclientv2.APIResponse, error)
type getResponsemanagementResponseIdByNameFunc func(ctx context.Context, p *responsemanagementResponseProxy, name string, libraryId string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error)
type getResponsemanagementResponseByIdFunc func(ctx context.Context, p *responsemanagementResponseProxy, id string) (response *platformclientv2.Response, resp *platformclientv2.APIResponse, err error)
type updateResponsemanagementResponseFunc func(ctx context.Context, p *responsemanagementResponseProxy, id string, response *platformclientv2.Response) (*platformclientv2.Response, *platformclientv2.APIResponse, error)
type deleteResponsemanagementResponseFunc func(ctx context.Context, p *responsemanagementResponseProxy, id string) (resp *platformclientv2.APIResponse, err error)

// responsemanagementResponseProxy contains all of the methods that call genesys cloud APIs.
type responsemanagementResponseProxy struct {
	clientConfig                              *platformclientv2.Configuration
	responseManagementApi                     *platformclientv2.ResponseManagementApi
	createResponsemanagementResponseAttr      createResponsemanagementResponseFunc
	getAllResponsemanagementResponseAttr      getAllResponsemanagementResponseFunc
	getResponsemanagementResponseIdByNameAttr getResponsemanagementResponseIdByNameFunc
	getResponsemanagementResponseByIdAttr     getResponsemanagementResponseByIdFunc
	updateResponsemanagementResponseAttr      updateResponsemanagementResponseFunc
	deleteResponsemanagementResponseAttr      deleteResponsemanagementResponseFunc
}

// newResponsemanagementResponseProxy initializes the responsemanagement response proxy with all of the data needed to communicate with Genesys Cloud
func newResponsemanagementResponseProxy(clientConfig *platformclientv2.Configuration) *responsemanagementResponseProxy {
	api := platformclientv2.NewResponseManagementApiWithConfig(clientConfig)
	return &responsemanagementResponseProxy{
		clientConfig:                              clientConfig,
		responseManagementApi:                     api,
		createResponsemanagementResponseAttr:      createResponsemanagementResponseFn,
		getAllResponsemanagementResponseAttr:      getAllResponsemanagementResponseFn,
		getResponsemanagementResponseIdByNameAttr: getResponsemanagementResponseIdByNameFn,
		getResponsemanagementResponseByIdAttr:     getResponsemanagementResponseByIdFn,
		updateResponsemanagementResponseAttr:      updateResponsemanagementResponseFn,
		deleteResponsemanagementResponseAttr:      deleteResponsemanagementResponseFn,
	}
}

// getResponsemanagementResponseProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getResponsemanagementResponseProxy(clientConfig *platformclientv2.Configuration) *responsemanagementResponseProxy {
	if internalProxy == nil {
		internalProxy = newResponsemanagementResponseProxy(clientConfig)
	}
	return internalProxy
}

// createResponsemanagementResponse creates a Genesys Cloud responsemanagement response
func (p *responsemanagementResponseProxy) createResponsemanagementResponse(ctx context.Context, responsemanagementResponse *platformclientv2.Response) (response *platformclientv2.Response, resp *platformclientv2.APIResponse, err error) {
	return p.createResponsemanagementResponseAttr(ctx, p, responsemanagementResponse)
}

// getResponsemanagementResponse retrieves all Genesys Cloud responsemanagement response
func (p *responsemanagementResponseProxy) getAllResponsemanagementResponse(ctx context.Context) (*[]platformclientv2.Response, *platformclientv2.APIResponse, error) {
	return p.getAllResponsemanagementResponseAttr(ctx, p, "")
}

// getResponsemanagementResponseIdByName returns a single Genesys Cloud responsemanagement response by a name
func (p *responsemanagementResponseProxy) getResponsemanagementResponseIdByName(ctx context.Context, name string, libraryId string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getResponsemanagementResponseIdByNameAttr(ctx, p, name, libraryId)
}

// getResponsemanagementResponseById returns a single Genesys Cloud responsemanagement response by Id
func (p *responsemanagementResponseProxy) getResponsemanagementResponseById(ctx context.Context, id string) (responsemanagementResponse *platformclientv2.Response, resp *platformclientv2.APIResponse, err error) {
	return p.getResponsemanagementResponseByIdAttr(ctx, p, id)
}

// updateResponsemanagementResponse updates a Genesys Cloud responsemanagement response
func (p *responsemanagementResponseProxy) updateResponsemanagementResponse(ctx context.Context, id string, responsemanagementResponse *platformclientv2.Response) (*platformclientv2.Response, *platformclientv2.APIResponse, error) {
	return p.updateResponsemanagementResponseAttr(ctx, p, id, responsemanagementResponse)
}

// deleteResponsemanagementResponse deletes a Genesys Cloud responsemanagement response by Id
func (p *responsemanagementResponseProxy) deleteResponsemanagementResponse(ctx context.Context, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.deleteResponsemanagementResponseAttr(ctx, p, id)
}

// createResponsemanagementResponseFn is an implementation function for creating a Genesys Cloud responsemanagement response
func createResponsemanagementResponseFn(ctx context.Context, p *responsemanagementResponseProxy, responsemanagementResponse *platformclientv2.Response) (*platformclientv2.Response, *platformclientv2.APIResponse, error) {
	response, resp, err := p.responseManagementApi.PostResponsemanagementResponses(*responsemanagementResponse, "")
	if err != nil {
		return nil, resp, err
	}
	return response, resp, nil
}

// getAllResponsemanagementResponseFn is the implementation for retrieving all responsemanagement response in Genesys Cloud
func getAllResponsemanagementResponseFn(ctx context.Context, p *responsemanagementResponseProxy, libraryId string) (*[]platformclientv2.Response, *platformclientv2.APIResponse, error) {
	var allResponseManagementResponses []platformclientv2.Response
	const pageSize = 100

	if libraryId != "" {
		responses, resp, getErr := p.responseManagementApi.GetResponsemanagementResponses(libraryId, 1, pageSize, "")
		if getErr != nil {
			return nil, resp, fmt.Errorf("Error requesting page of Responsemanagement Response: %s", getErr)
		}

		if responses.Entities == nil || len(*responses.Entities) == 0 {
			return &allResponseManagementResponses, resp, nil
		}

		for _, response := range *responses.Entities {
			allResponseManagementResponses = append(allResponseManagementResponses, response)
		}

		return &allResponseManagementResponses, resp, nil
	}

	libraries, resp, getErr := p.responseManagementApi.GetResponsemanagementLibraries(1, pageSize, "", "")
	if getErr != nil {
		return nil, resp, fmt.Errorf("Error requesting page of Responsemanagement library: %s", getErr)
	}
	if libraries.Entities == nil || len(*libraries.Entities) == 0 {
		return &allResponseManagementResponses, resp, nil
	}

	for _, library := range *libraries.Entities {
		for pageNum := 1; ; pageNum++ {
			responses, resp, getErr := p.responseManagementApi.GetResponsemanagementResponses(*library.Id, pageNum, pageSize, "")
			if getErr != nil {
				return nil, resp, fmt.Errorf("Error requesting page of Responsemanagement Response: %s", getErr)
			}

			if responses.Entities == nil || len(*responses.Entities) == 0 {
				break
			}

			for _, response := range *responses.Entities {
				allResponseManagementResponses = append(allResponseManagementResponses, response)
			}
		}
	}

	for pageNum := 2; pageNum <= *libraries.PageCount; pageNum++ {
		libraries, resp, getErr := p.responseManagementApi.GetResponsemanagementLibraries(pageNum, pageSize, "", "")

		if getErr != nil {
			return nil, resp, fmt.Errorf("Error requesting page of Responsemanagement library: %s", getErr)
		}
		if libraries.Entities == nil || len(*libraries.Entities) == 0 {
			break
		}

		for _, library := range *libraries.Entities {
			for pageNum := 1; ; pageNum++ {
				responses, resp, getErr := p.responseManagementApi.GetResponsemanagementResponses(*library.Id, pageNum, pageSize, "")
				if getErr != nil {
					return nil, resp, fmt.Errorf("Error requesting page of Responsemanagement Response: %s", getErr)
				}

				if responses.Entities == nil || len(*responses.Entities) == 0 {
					break
				}

				for _, response := range *responses.Entities {
					allResponseManagementResponses = append(allResponseManagementResponses, response)
				}
			}
		}
	}

	return &allResponseManagementResponses, resp, nil
}

// getResponsemanagementResponseIdByNameFn is an implementation of the function to get a Genesys Cloud responsemanagement response by name
func getResponsemanagementResponseIdByNameFn(ctx context.Context, p *responsemanagementResponseProxy, name string, libraryId string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	responses, resp, err := getAllResponsemanagementResponseFn(ctx, p, libraryId)
	if err != nil {
		return "", false, resp, fmt.Errorf("Error searching response management responses %s: %s", name, err)
	}

	var response platformclientv2.Response
	for _, responseSdk := range *responses {
		if *responseSdk.Name == name {
			log.Printf("Retrieved the response management response %s by name %s", *responseSdk.Id, name)
			response = responseSdk
			return *response.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("Unable to find response management response with name %s", name)
}

// getResponsemanagementResponseByIdFn is an implementation of the function to get a Genesys Cloud responsemanagement response by Id
func getResponsemanagementResponseByIdFn(ctx context.Context, p *responsemanagementResponseProxy, id string) (responsemanagementResponse *platformclientv2.Response, resp *platformclientv2.APIResponse, err error) {
	mamangementResponse, resp, err := p.responseManagementApi.GetResponsemanagementResponse(id, "")
	if err != nil {
		return nil, resp, err
	}
	return mamangementResponse, resp, nil
}

// updateResponsemanagementResponseFn is an implementation of the function to update a Genesys Cloud responsemanagement response
func updateResponsemanagementResponseFn(ctx context.Context, p *responsemanagementResponseProxy, id string, response *platformclientv2.Response) (*platformclientv2.Response, *platformclientv2.APIResponse, error) {
	responsemanagementResponse, resp, err := p.responseManagementApi.GetResponsemanagementResponse(id, "")
	if err != nil {
		return nil, resp, err
	}

	response.Version = responsemanagementResponse.Version
	responsemanagementResponse, resp, updateErr := p.responseManagementApi.PutResponsemanagementResponse(id, *response, "")
	if updateErr != nil {
		return nil, resp, updateErr
	}

	return responsemanagementResponse, resp, nil
}

// deleteResponsemanagementResponseFn is an implementation function for deleting a Genesys Cloud responsemanagement response
func deleteResponsemanagementResponseFn(ctx context.Context, p *responsemanagementResponseProxy, id string) (*platformclientv2.APIResponse, error) {
	resp, err := p.responseManagementApi.DeleteResponsemanagementResponse(id)
	if err != nil {
		return resp, err
	}
	return resp, err
}
