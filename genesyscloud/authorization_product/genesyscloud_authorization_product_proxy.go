package authorization_product

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_authorization_product_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *authProductProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAuthorizationProductFunc func(ctx context.Context, p *authProductProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)

// authProductProxy contains all of the methods that call genesys cloud APIs.
type authProductProxy struct {
	clientConfig                *platformclientv2.Configuration
	authApi                     *platformclientv2.AuthorizationApi
	getAuthorizationProductAttr getAuthorizationProductFunc
}

// newauthProductProxy initializes the authorization product proxy with all of the data needed to communicate with Genesys Cloud
func newauthProductProxy(clientConfig *platformclientv2.Configuration) *authProductProxy {
	api := platformclientv2.NewAuthorizationApiWithConfig(clientConfig)
	return &authProductProxy{
		clientConfig:                clientConfig,
		authApi:                     api,
		getAuthorizationProductAttr: getAuthorizationProductFn,
	}
}

// getauthProductProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getauthProductProxy(clientConfig *platformclientv2.Configuration) *authProductProxy {
	if internalProxy == nil {
		internalProxy = newauthProductProxy(clientConfig)
	}
	return internalProxy
}

// getAuthorizationProduct returns a single Genesys Cloud authorization product by a name
func (p *authProductProxy) getAuthorizationProduct(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getAuthorizationProductAttr(ctx, p, name)
}

// getAuthorizationProductFn is an implementation of the function to get a Genesys Cloud authorization product by name
func getAuthorizationProductFn(ctx context.Context, p *authProductProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	authProducts, apiResponse, err := p.authApi.GetAuthorizationProducts()
	if err != nil {
		return "", true, apiResponse, fmt.Errorf("error requesting Auth Product %s: %s", name, err)
	}

	if authProducts.Entities == nil || len(*authProducts.Entities) == 0 {
		return "", false, apiResponse, fmt.Errorf("No Auth Products found with name %s", name)
	}

	for _, entity := range *authProducts.Entities {
		if *entity.Id == name {
			return *entity.Id, false, apiResponse, nil
		}
	}
	return "", false, apiResponse, fmt.Errorf("no Auth Product found with name %s", name)
}
