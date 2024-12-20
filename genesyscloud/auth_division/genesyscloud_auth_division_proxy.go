package auth_division

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

var internalProxy *authDivisionProxy

type getAllAuthDivisionFunc func(ctx context.Context, p *authDivisionProxy, name string) (*[]platformclientv2.Authzdivision, *platformclientv2.APIResponse, error)
type createAuthDivisionFunc func(ctx context.Context, p *authDivisionProxy, authzDivision *platformclientv2.Authzdivision) (*platformclientv2.Authzdivision, *platformclientv2.APIResponse, error)
type getAuthDivisionIdByNameFunc func(ctx context.Context, p *authDivisionProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type getAuthDivisionByIdFunc func(ctx context.Context, p *authDivisionProxy, id string, objectCount, checkCache bool) (*platformclientv2.Authzdivision, *platformclientv2.APIResponse, error)
type updateAuthDivisionFunc func(ctx context.Context, p *authDivisionProxy, id string, authzDivision *platformclientv2.Authzdivision) (*platformclientv2.Authzdivision, *platformclientv2.APIResponse, error)
type deleteAuthDivisionFunc func(ctx context.Context, p *authDivisionProxy, id string, force bool) (*platformclientv2.APIResponse, error)

type authDivisionProxy struct {
	clientConfig                *platformclientv2.Configuration
	authorizationApi            *platformclientv2.AuthorizationApi
	createAuthDivisionAttr      createAuthDivisionFunc
	getAllAuthDivisionAttr      getAllAuthDivisionFunc
	getAuthDivisionIdByNameAttr getAuthDivisionIdByNameFunc
	getAuthDivisionByIdAttr     getAuthDivisionByIdFunc
	updateAuthDivisionAttr      updateAuthDivisionFunc
	deleteAuthDivisionAttr      deleteAuthDivisionFunc
	authDivisionCache           rc.CacheInterface[platformclientv2.Authzdivision]
}

// newAuthDivisionProxy initializes the auth division proxy with all of the data needed to communicate with Genesys Cloud
func newAuthDivisionProxy(clientConfig *platformclientv2.Configuration) *authDivisionProxy {
	api := platformclientv2.NewAuthorizationApiWithConfig(clientConfig)
	authDivisionCache := rc.NewResourceCache[platformclientv2.Authzdivision]()
	return &authDivisionProxy{
		clientConfig:                clientConfig,
		authorizationApi:            api,
		createAuthDivisionAttr:      createAuthDivisionFn,
		getAllAuthDivisionAttr:      getAllAuthDivisionFn,
		getAuthDivisionIdByNameAttr: getAuthDivisionIdByNameFn,
		getAuthDivisionByIdAttr:     getAuthDivisionByIdFn,
		updateAuthDivisionAttr:      updateAuthDivisionFn,
		deleteAuthDivisionAttr:      deleteAuthDivisionFn,
		authDivisionCache:           authDivisionCache,
	}
}

// getAuthDivisionProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getAuthDivisionProxy(clientConfig *platformclientv2.Configuration) *authDivisionProxy {
	if internalProxy == nil {
		internalProxy = newAuthDivisionProxy(clientConfig)
	}

	return internalProxy
}

func (p *authDivisionProxy) getAllAuthDivision(ctx context.Context, name string) (*[]platformclientv2.Authzdivision, *platformclientv2.APIResponse, error) {
	return p.getAllAuthDivisionAttr(ctx, p, name)
}

func (p *authDivisionProxy) createAuthDivision(ctx context.Context, authDivision *platformclientv2.Authzdivision) (*platformclientv2.Authzdivision, *platformclientv2.APIResponse, error) {
	return p.createAuthDivisionAttr(ctx, p, authDivision)
}

func (p *authDivisionProxy) getAuthDivisionById(ctx context.Context, id string, objectCount, checkCache bool) (*platformclientv2.Authzdivision, *platformclientv2.APIResponse, error) {
	return p.getAuthDivisionByIdAttr(ctx, p, id, objectCount, checkCache)
}

func (p *authDivisionProxy) getAuthDivisionIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getAuthDivisionIdByNameAttr(ctx, p, name)
}

func (p *authDivisionProxy) updateAuthDivision(ctx context.Context, id string, authDivision *platformclientv2.Authzdivision) (*platformclientv2.Authzdivision, *platformclientv2.APIResponse, error) {
	return p.updateAuthDivisionAttr(ctx, p, id, authDivision)
}

func (p *authDivisionProxy) deleteAuthDivision(ctx context.Context, id string, force bool) (*platformclientv2.APIResponse, error) {
	return p.deleteAuthDivisionAttr(ctx, p, id, force)
}

func getAllAuthDivisionFn(ctx context.Context, p *authDivisionProxy, name string) (*[]platformclientv2.Authzdivision, *platformclientv2.APIResponse, error) {
	var allAuthzDivisions []platformclientv2.Authzdivision
	const pageSize = 100

	authzDivisions, resp, err := p.authorizationApi.GetAuthorizationDivisions(pageSize, 1, "", nil, "", "", false, nil, name)
	if err != nil {
		return nil, resp, err
	}

	if authzDivisions.Entities == nil || len(*authzDivisions.Entities) == 0 {
		return &allAuthzDivisions, resp, nil
	}
	allAuthzDivisions = append(allAuthzDivisions, *authzDivisions.Entities...)

	for pageNum := 2; pageNum <= *authzDivisions.PageCount; pageNum++ {
		authzDivisions, resp, err := p.authorizationApi.GetAuthorizationDivisions(pageSize, pageNum, "", nil, "", "", false, nil, name)
		if err != nil {
			return nil, resp, err
		}

		if authzDivisions.Entities == nil || len(*authzDivisions.Entities) == 0 {
			break
		}
		allAuthzDivisions = append(allAuthzDivisions, *authzDivisions.Entities...)
	}

	for _, div := range allAuthzDivisions {
		rc.SetCache(p.authDivisionCache, *div.Id, div)
	}

	return &allAuthzDivisions, resp, nil
}

// createAuthDivisionFn is an implementation function for creating a Genesys Cloud auth division
func createAuthDivisionFn(ctx context.Context, p *authDivisionProxy, authDivision *platformclientv2.Authzdivision) (*platformclientv2.Authzdivision, *platformclientv2.APIResponse, error) {
	return p.authorizationApi.PostAuthorizationDivisions(*authDivision)
}

func getAuthDivisionByIdFn(ctx context.Context, p *authDivisionProxy, id string, objectCount, checkCache bool) (*platformclientv2.Authzdivision, *platformclientv2.APIResponse, error) {
	if checkCache {
		div := rc.GetCacheItem(p.authDivisionCache, id)
		if div != nil {
			return div, nil, nil
		}
	}

	return p.authorizationApi.GetAuthorizationDivision(id, objectCount)
}

func getAuthDivisionIdByNameFn(ctx context.Context, p *authDivisionProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	notFoundError := fmt.Errorf("unable to find auth division with name %s", name)

	authzDivisions, resp, err := getAllAuthDivisionFn(ctx, p, name)
	if err != nil {
		return "", resp, false, err
	}

	if authzDivisions == nil || len(*authzDivisions) == 0 {
		return "", resp, true, notFoundError
	}

	for _, authzDivision := range *authzDivisions {
		if *authzDivision.Name == name {
			log.Printf("Retrieved the auth division id %s by name %s", *authzDivision.Id, name)
			return *authzDivision.Id, resp, false, nil
		}
	}

	return "", resp, true, notFoundError
}

func updateAuthDivisionFn(ctx context.Context, p *authDivisionProxy, id string, authDivision *platformclientv2.Authzdivision) (*platformclientv2.Authzdivision, *platformclientv2.APIResponse, error) {
	return p.authorizationApi.PutAuthorizationDivision(id, *authDivision)
}

func deleteAuthDivisionFn(ctx context.Context, p *authDivisionProxy, id string, force bool) (*platformclientv2.APIResponse, error) {
	resp, err := p.authorizationApi.DeleteAuthorizationDivision(id, force)
	if err != nil {
		return resp, err
	}
	rc.DeleteCacheItem(p.authDivisionCache, id)
	return resp, nil
}
