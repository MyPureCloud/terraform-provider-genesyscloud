package telephony_providers_edges_did

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_telephony_providers_edges_did_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.

Each proxy implementation:

1.  Should provide a private package level variable that holds a instance of a proxy class.
2.  A New... constructor function  to initialize the proxy object.  This constructor should only be used within
    the proxy.
3.  A get private constructor function that the classes in the package can be used to to retrieve
    the proxy.  This proxy should check to see if the package level proxy instance is nil and
    should initialize it, otherwise it should return the instance
4.  Type definitions for each function that will be used in the proxy.  We use composition here
    so that we can easily provide mocks for testing.
5.  A struct for the proxy that holds an attribute for each function type.
6.  Wrapper methods on each of the elements on the struct.
7.  Function implementations for each function type definition.

*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *telephonyProvidersEdgesDidProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getTelephonyProvidersEdgesDidIdByDidFunc func(ctx context.Context, t *telephonyProvidersEdgesDidProxy, did string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error)

// telephonyProvidersEdgesDidProxy contains all of the methods that call genesys cloud APIs.
type telephonyProvidersEdgesDidProxy struct {
	clientConfig                             *platformclientv2.Configuration
	telephonyApi                             *platformclientv2.TelephonyProvidersEdgeApi
	getTelephonyProvidersEdgesDidIdByDidAttr getTelephonyProvidersEdgesDidIdByDidFunc
}

// newTelephonyProvidersEdgesDidProxy initializes the proxy with all data needed to communicate with Genesys Cloud
func newTelephonyProvidersEdgesDidProxy(clientConfig *platformclientv2.Configuration) *telephonyProvidersEdgesDidProxy {
	api := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(clientConfig)
	return &telephonyProvidersEdgesDidProxy{
		clientConfig:                             clientConfig,
		telephonyApi:                             api,
		getTelephonyProvidersEdgesDidIdByDidAttr: getTelephonyProvidersEdgesDidIdByDidFn,
	}
}

// getTelephonyProvidersEdgesDidProxy acts as a singleton for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getTelephonyProvidersEdgesDidProxy(clientConfig *platformclientv2.Configuration) *telephonyProvidersEdgesDidProxy {
	if internalProxy == nil {
		internalProxy = newTelephonyProvidersEdgesDidProxy(clientConfig)
	}
	return internalProxy
}

// getTelephonyProvidersEdgesDidIdByDid gets a Genesys Cloud telephony DID ID by DID number
func (t *telephonyProvidersEdgesDidProxy) getTelephonyProvidersEdgesDidIdByDid(ctx context.Context, did string) (string, bool, *platformclientv2.APIResponse, error) {
	return t.getTelephonyProvidersEdgesDidIdByDidAttr(ctx, t, did)
}

// getTelephonyProvidersEdgesDidIdByDidFn is an implementation function for getting a telephony DID ID by DID number.
func getTelephonyProvidersEdgesDidIdByDidFn(_ context.Context, t *telephonyProvidersEdgesDidProxy, did string) (string, bool, *platformclientv2.APIResponse, error) {
	const pageSize = 100

	pageNum := 1
	dids, resp, getErr := t.telephonyApi.GetTelephonyProvidersEdgesDids(pageSize, pageNum, "", "", did, "", "", nil)
	if getErr != nil {
		return "", false, resp, fmt.Errorf("error requesting list of DIDs: %s", getErr)
	}

	for pageNum := 1; pageNum <= *dids.PageCount; pageNum++ {
		dids, resp, getErr := t.telephonyApi.GetTelephonyProvidersEdgesDids(pageSize, pageNum, "", "", did, "", "", nil)
		if getErr != nil {
			return "", false, resp, fmt.Errorf("error requesting list of DIDs: %s", getErr)
		}
		if dids.Entities == nil || len(*dids.Entities) == 0 {
			break
		}
		for _, entity := range *dids.Entities {
			if *entity.PhoneNumber == did {
				return *entity.Id, false, resp, nil
			}
		}
	}
	return "", true, resp, fmt.Errorf("failed to find ID of did number '%s'", did)
}
