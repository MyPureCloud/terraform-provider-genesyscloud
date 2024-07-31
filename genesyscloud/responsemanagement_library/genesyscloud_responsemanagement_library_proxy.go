package responsemanagement_library

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_responsemanagement_library_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *responsemanagementLibraryProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createResponsemanagementLibraryFunc func(ctx context.Context, p *responsemanagementLibraryProxy, library *platformclientv2.Library) (*platformclientv2.Library, *platformclientv2.APIResponse, error)
type getAllResponsemanagementLibraryFunc func(ctx context.Context, p *responsemanagementLibraryProxy, name string) (*[]platformclientv2.Library, *platformclientv2.APIResponse, error)
type getResponsemanagementLibraryIdByNameFunc func(ctx context.Context, p *responsemanagementLibraryProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getResponsemanagementLibraryByIdFunc func(ctx context.Context, p *responsemanagementLibraryProxy, id string) (library *platformclientv2.Library, response *platformclientv2.APIResponse, err error)
type updateResponsemanagementLibraryFunc func(ctx context.Context, p *responsemanagementLibraryProxy, id string, library *platformclientv2.Library) (*platformclientv2.Library, *platformclientv2.APIResponse, error)
type deleteResponsemanagementLibraryFunc func(ctx context.Context, p *responsemanagementLibraryProxy, id string) (response *platformclientv2.APIResponse, err error)

// responsemanagementLibraryProxy contains all of the methods that call genesys cloud APIs.
type responsemanagementLibraryProxy struct {
	clientConfig                             *platformclientv2.Configuration
	responseManagementApi                    *platformclientv2.ResponseManagementApi
	createResponsemanagementLibraryAttr      createResponsemanagementLibraryFunc
	getAllResponsemanagementLibraryAttr      getAllResponsemanagementLibraryFunc
	getResponsemanagementLibraryIdByNameAttr getResponsemanagementLibraryIdByNameFunc
	getResponsemanagementLibraryByIdAttr     getResponsemanagementLibraryByIdFunc
	updateResponsemanagementLibraryAttr      updateResponsemanagementLibraryFunc
	deleteResponsemanagementLibraryAttr      deleteResponsemanagementLibraryFunc
}

// newResponsemanagementLibraryProxy initializes the responsemanagement library proxy with all of the data needed to communicate with Genesys Cloud
func newResponsemanagementLibraryProxy(clientConfig *platformclientv2.Configuration) *responsemanagementLibraryProxy {
	api := platformclientv2.NewResponseManagementApiWithConfig(clientConfig)
	return &responsemanagementLibraryProxy{
		clientConfig:                             clientConfig,
		responseManagementApi:                    api,
		createResponsemanagementLibraryAttr:      createResponsemanagementLibraryFn,
		getAllResponsemanagementLibraryAttr:      getAllResponsemanagementLibraryFn,
		getResponsemanagementLibraryIdByNameAttr: getResponsemanagementLibraryIdByNameFn,
		getResponsemanagementLibraryByIdAttr:     getResponsemanagementLibraryByIdFn,
		updateResponsemanagementLibraryAttr:      updateResponsemanagementLibraryFn,
		deleteResponsemanagementLibraryAttr:      deleteResponsemanagementLibraryFn,
	}
}

// getResponsemanagementLibraryProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getResponsemanagementLibraryProxy(clientConfig *platformclientv2.Configuration) *responsemanagementLibraryProxy {
	if internalProxy == nil {
		internalProxy = newResponsemanagementLibraryProxy(clientConfig)
	}
	return internalProxy
}

// createResponsemanagementLibrary creates a Genesys Cloud responsemanagement library
func (p *responsemanagementLibraryProxy) createResponsemanagementLibrary(ctx context.Context, responsemanagementLibrary *platformclientv2.Library) (*platformclientv2.Library, *platformclientv2.APIResponse, error) {
	return p.createResponsemanagementLibraryAttr(ctx, p, responsemanagementLibrary)
}

// getResponsemanagementLibrary retrieves all Genesys Cloud responsemanagement library
func (p *responsemanagementLibraryProxy) getAllResponsemanagementLibrary(ctx context.Context) (*[]platformclientv2.Library, *platformclientv2.APIResponse, error) {
	return p.getAllResponsemanagementLibraryAttr(ctx, p, "")
}

// getResponsemanagementLibraryIdByName returns a single Genesys Cloud responsemanagement library by a name
func (p *responsemanagementLibraryProxy) getResponsemanagementLibraryIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getResponsemanagementLibraryIdByNameAttr(ctx, p, name)
}

// getResponsemanagementLibraryById returns a single Genesys Cloud responsemanagement library by Id
func (p *responsemanagementLibraryProxy) getResponsemanagementLibraryById(ctx context.Context, id string) (responsemanagementLibrary *platformclientv2.Library, response *platformclientv2.APIResponse, err error) {
	return p.getResponsemanagementLibraryByIdAttr(ctx, p, id)
}

// updateResponsemanagementLibrary updates a Genesys Cloud responsemanagement library
func (p *responsemanagementLibraryProxy) updateResponsemanagementLibrary(ctx context.Context, id string, responsemanagementLibrary *platformclientv2.Library) (*platformclientv2.Library, *platformclientv2.APIResponse, error) {
	return p.updateResponsemanagementLibraryAttr(ctx, p, id, responsemanagementLibrary)
}

// deleteResponsemanagementLibrary deletes a Genesys Cloud responsemanagement library by Id
func (p *responsemanagementLibraryProxy) deleteResponsemanagementLibrary(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteResponsemanagementLibraryAttr(ctx, p, id)
}

// createResponsemanagementLibraryFn is an implementation function for creating a Genesys Cloud responsemanagement library
func createResponsemanagementLibraryFn(ctx context.Context, p *responsemanagementLibraryProxy, responsemanagementLibrary *platformclientv2.Library) (*platformclientv2.Library, *platformclientv2.APIResponse, error) {
	library, resp, err := p.responseManagementApi.PostResponsemanagementLibraries(*responsemanagementLibrary)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to create responsemanagement library: %s", err)
	}
	return library, resp, nil
}

// getAllResponsemanagementLibraryFn is the implementation for retrieving all responsemanagement library in Genesys Cloud
func getAllResponsemanagementLibraryFn(ctx context.Context, p *responsemanagementLibraryProxy, name string) (*[]platformclientv2.Library, *platformclientv2.APIResponse, error) {
	var allLibrarys []platformclientv2.Library
	const pageSize = 100

	librarys, resp, err := p.responseManagementApi.GetResponsemanagementLibraries(1, pageSize, "", name)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get library: %v", err)
	}
	if librarys.Entities == nil || len(*librarys.Entities) == 0 {
		return &allLibrarys, resp, nil
	}
	for _, library := range *librarys.Entities {
		allLibrarys = append(allLibrarys, library)
	}

	for pageNum := 2; pageNum <= *librarys.PageCount; pageNum++ {
		librarys, resp, err := p.responseManagementApi.GetResponsemanagementLibraries(pageNum, pageSize, "", name)
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get library: %v", err)
		}

		if librarys.Entities == nil || len(*librarys.Entities) == 0 {
			break
		}

		for _, library := range *librarys.Entities {
			allLibrarys = append(allLibrarys, library)
		}
	}
	return &allLibrarys, resp, nil
}

// getResponsemanagementLibraryIdByNameFn is an implementation of the function to get a Genesys Cloud responsemanagement library by name
func getResponsemanagementLibraryIdByNameFn(ctx context.Context, p *responsemanagementLibraryProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	librarys, resp, err := getAllResponsemanagementLibraryFn(ctx, p, name)
	if err != nil {
		return "", false, resp, err
	}

	if librarys == nil || len(*librarys) == 0 {
		return "", true, resp, fmt.Errorf("No responsemanagement library found with name %s", name)
	}

	for _, library := range *librarys {
		if *library.Name == name {
			log.Printf("Retrieved the responsemanagement library id %s by name %s", *library.Id, name)
			return *library.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("Unable to find responsemanagement library with name %s", name)
}

// getResponsemanagementLibraryByIdFn is an implementation of the function to get a Genesys Cloud responsemanagement library by Id
func getResponsemanagementLibraryByIdFn(ctx context.Context, p *responsemanagementLibraryProxy, id string) (responsemanagementLibrary *platformclientv2.Library, response *platformclientv2.APIResponse, err error) {
	library, resp, err := p.responseManagementApi.GetResponsemanagementLibrary(id)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to retrieve responsemanagement library by id %s: %s", id, err)
	}
	return library, resp, nil
}

// updateResponsemanagementLibraryFn is an implementation of the function to update a Genesys Cloud responsemanagement library
func updateResponsemanagementLibraryFn(ctx context.Context, p *responsemanagementLibraryProxy, id string, responsemanagementLibrary *platformclientv2.Library) (*platformclientv2.Library, *platformclientv2.APIResponse, error) {
	lib, resp, err := getResponsemanagementLibraryByIdFn(ctx, p, id)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to update responsemanagement library: %s", err)
	}

	responsemanagementLibrary.Version = lib.Version
	library, resp, err := p.responseManagementApi.PutResponsemanagementLibrary(id, *responsemanagementLibrary)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to update responsemanagement library: %s", err)
	}
	return library, resp, nil
}

// deleteResponsemanagementLibraryFn is an implementation function for deleting a Genesys Cloud responsemanagement library
func deleteResponsemanagementLibraryFn(ctx context.Context, p *responsemanagementLibraryProxy, id string) (response *platformclientv2.APIResponse, err error) {
	resp, err := p.responseManagementApi.DeleteResponsemanagementLibrary(id)
	if err != nil {
		return resp, fmt.Errorf("Failed to delete responsemanagement library: %s", err)
	}
	return resp, nil
}
