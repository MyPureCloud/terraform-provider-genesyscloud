package idp_adfs

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v129/platformclientv2"
)

/*
The genesyscloud_idp_adfs_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *idpAdfsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllIdpAdfsFunc func(ctx context.Context, p *idpAdfsProxy) (*[]platformclientv2.Adfs, error)
type getIdpAdfsIdByNameFunc func(ctx context.Context, p *idpAdfsProxy, name string) (id string, retryable bool, err error)
type getIdpAdfsByIdFunc func(ctx context.Context, p *idpAdfsProxy, id string) (aDFS *platformclientv2.Adfs, statusCode int, err error)
type updateIdpAdfsFunc func(ctx context.Context, p *idpAdfsProxy, id string, aDFS *platformclientv2.Adfs) (statusCode int, err error)
type deleteIdpAdfsFunc func(ctx context.Context, p *idpAdfsProxy, id string) (statusCode int, err error)

// idpAdfsProxy contains all of the methods that call genesys cloud APIs.
type idpAdfsProxy struct {
	clientConfig           *platformclientv2.Configuration
	identityProviderApi    *platformclientv2.IdentityProviderApi
	getAllIdpAdfsAttr      getAllIdpAdfsFunc
	getIdpAdfsIdByNameAttr getIdpAdfsIdByNameFunc
	getIdpAdfsByIdAttr     getIdpAdfsByIdFunc
	updateIdpAdfsAttr      updateIdpAdfsFunc
	deleteIdpAdfsAttr      deleteIdpAdfsFunc
}

// newIdpAdfsProxy initializes the idp adfs proxy with all of the data needed to communicate with Genesys Cloud
func newIdpAdfsProxy(clientConfig *platformclientv2.Configuration) *idpAdfsProxy {
	api := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	return &idpAdfsProxy{
		clientConfig:           clientConfig,
		identityProviderApi:    api,
		getAllIdpAdfsAttr:      getAllIdpAdfsFn,
		getIdpAdfsIdByNameAttr: getIdpAdfsIdByNameFn,
		getIdpAdfsByIdAttr:     getIdpAdfsByIdFn,
		updateIdpAdfsAttr:      updateIdpAdfsFn,
		deleteIdpAdfsAttr:      deleteIdpAdfsFn,
	}
}

// getIdpAdfsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getIdpAdfsProxy(clientConfig *platformclientv2.Configuration) *idpAdfsProxy {
	if internalProxy == nil {
		internalProxy = newIdpAdfsProxy(clientConfig)
	}

	return internalProxy
}

// getIdpAdfs retrieves all Genesys Cloud idp adfs
func (p *idpAdfsProxy) getAllIdpAdfs(ctx context.Context) (*[]platformclientv2.Adfs, error) {
	return p.getAllIdpAdfsAttr(ctx, p)
}

// getIdpAdfsIdByName returns a single Genesys Cloud idp adfs by a name
func (p *idpAdfsProxy) getIdpAdfsIdByName(ctx context.Context, name string) (id string, retryable bool, err error) {
	return p.getIdpAdfsIdByNameAttr(ctx, p, name)
}

// getIdpAdfsById returns a single Genesys Cloud idp adfs by Id
func (p *idpAdfsProxy) getIdpAdfsById(ctx context.Context, id string) (idpAdfs *platformclientv2.Adfs, statusCode int, err error) {
	return p.getIdpAdfsByIdAttr(ctx, p, id)
}

// updateIdpAdfs updates a Genesys Cloud idp adfs
func (p *idpAdfsProxy) updateIdpAdfs(ctx context.Context, id string, idpAdfs *platformclientv2.Adfs) (statusCode int, err error) {
	return p.updateIdpAdfsAttr(ctx, p, id, idpAdfs)
}

// deleteIdpAdfs deletes a Genesys Cloud idp adfs by Id
func (p *idpAdfsProxy) deleteIdpAdfs(ctx context.Context, id string) (statusCode int, err error) {
	return p.deleteIdpAdfsAttr(ctx, p, id)
}

// getAllIdpAdfsFn is the implementation for retrieving all idp adfs in Genesys Cloud
func getAllIdpAdfsFn(ctx context.Context, p *idpAdfsProxy) (*[]platformclientv2.Adfs, error) {
	var allADFSs []platformclientv2.Adfs
	const pageSize = 100

	aDFSs, _, err := p.identityProviderApi.GetIdentityprovidersAdfs()
	if err != nil {
		return nil, fmt.Errorf("Failed to get a d f s: %v", err)
	}
	// previously aDFSs.Entities && len(*aDFSs.Entities)
	if aDFSs == nil {
		return &allADFSs, nil
	}
	//for _, aDFS := range *aDFSs {
	allADFSs = append(allADFSs, *aDFSs)
	//}

	// since no page count field and single response gets returned
	// for pageNum := 2; pageNum <= *aDFSs.PageCount; pageNum++ {
	// 	aDFSs, _, err := p.identityProviderApi.GetIdentityprovidersAdfs()
	// 	if err != nil {
	// 		return nil, fmt.Errorf("Failed to get a d f s: %v", err)
	// 	}

	// 	if aDFSs == nil || len(*aDFSs.Entities) == 0 {
	// 		break
	// 	}

	// 	for _, aDFS := range *aDFSs.Entities {
	// 		allADFSs = append(allADFSs, aDFS)
	// 	}
	// }

	return &allADFSs, nil
}

// getIdpAdfsIdByNameFn is an implementation of the function to get a Genesys Cloud idp adfs by name
func getIdpAdfsIdByNameFn(ctx context.Context, p *idpAdfsProxy, name string) (id string, retryable bool, err error) {
	aDFSs, _, err := p.identityProviderApi.GetIdentityprovidersAdfs()
	if err != nil {
		return "", false, err
	}

	if aDFSs == nil {
		return "", true, fmt.Errorf("No idp adfs found with name %s", name)
	}

	//for _, aDFS := range *aDFSs.Entities {
	if *aDFSs.Name == name {
		log.Printf("Retrieved the idp adfs id %s by name %s", *aDFSs.Id, name)
		return *aDFSs.Id, false, nil
	}
	//}

	return "", true, fmt.Errorf("Unable to find idp adfs with name %s", name)
}

// getIdpAdfsByIdFn is an implementation of the function to get a Genesys Cloud idp adfs by Id
func getIdpAdfsByIdFn(ctx context.Context, p *idpAdfsProxy, id string) (idpAdfs *platformclientv2.Adfs, statusCode int, err error) {
	// we dont have api where we can pass ID to GET request
	aDFS, resp, err := p.identityProviderApi.GetIdentityprovidersAdfs()
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to retrieve idp adfs by id %s: %s", id, err)
	}

	return aDFS, resp.StatusCode, nil
}

// updateIdpAdfsFn is an implementation of the function to update a Genesys Cloud idp adfs
func updateIdpAdfsFn(ctx context.Context, p *idpAdfsProxy, id string, idpAdfs *platformclientv2.Adfs) (statusCode int, err error) {
	_, resp, err := p.identityProviderApi.PutIdentityprovidersAdfs(*idpAdfs)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Failed to update idp adfs: %s", err)
	}
	return resp.StatusCode, nil
}

// deleteIdpAdfsFn is an implementation function for deleting a Genesys Cloud idp adfs
func deleteIdpAdfsFn(ctx context.Context, p *idpAdfsProxy, id string) (statusCode int, err error) {
	_, resp, err := p.identityProviderApi.DeleteIdentityprovidersAdfs()
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Failed to delete idp adfs: %s", err)
	}

	return resp.StatusCode, nil
}
