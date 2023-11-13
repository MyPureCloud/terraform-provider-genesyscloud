package telephony_providers_edges_did_pool

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

/*
The genesyscloud_telephony_providers_edges_did_pool_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.

Each proxy implementation:

1.  Should provide a private package level variable that holds an instance of a proxy class.
2.  A New... constructor function  to initialize the proxy object.  This constructor should only be used within
    the proxy.
3.  A get private constructor function that the classes in the package can be used to retrieve
    the proxy.  This proxy should check to see if the package level proxy instance is nil and
    should initialize it, otherwise it should return the instance
4.  Type definitions for each function that will be used in the proxy.  We use composition here
    so that we can easily provide mocks for testing.
5.  A struct for the proxy that holds an attribute for each function type.
6.  Wrapper methods on each of the elements on the struct.
7.  Function implementations for each function type definition.

*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *telephonyDidPoolProxy

// Type definitions for each func on our proxy, so we can easily mock them out later
type createTelephonyDidPool func(ctx context.Context, t *telephonyDidPoolProxy, didPool *platformclientv2.Didpool) (*platformclientv2.Didpool, error)
type getTelephonyDidPoolById func(context.Context, *telephonyDidPoolProxy, string) (didPool *platformclientv2.Didpool, respCode int, err error)
type updateTelephonyDidPool func(context.Context, *telephonyDidPoolProxy, string, *platformclientv2.Didpool) (*platformclientv2.Didpool, error)
type deleteTelephonyDidPool func(context.Context, *telephonyDidPoolProxy, string) error
type getTelephonyDidPoolIdByStartAndEndNumber func(ctx context.Context, t *telephonyDidPoolProxy, start, end string) (id string, retryable bool, err error)
type getAllTelephonyDidPools func(context.Context, *telephonyDidPoolProxy) (*[]platformclientv2.Didpool, error)

// telephonyDidPoolProxy contains all methods that call genesys cloud APIs.
type telephonyDidPoolProxy struct {
	clientConfig                                 *platformclientv2.Configuration
	telephonyApi                                 *platformclientv2.TelephonyProvidersEdgeApi
	createTelephonyDidPoolAttr                   createTelephonyDidPool
	getTelephonyDidPoolByIdAttr                  getTelephonyDidPoolById
	updateEdgesDidPoolAttr                       updateTelephonyDidPool
	deleteTelephonyDidPoolAttr                   deleteTelephonyDidPool
	getTelephonyDidPoolIdByStartAndEndNumberAttr getTelephonyDidPoolIdByStartAndEndNumber
	getAllTelephonyDidPoolsAttr                  getAllTelephonyDidPools
}

// newTelephonyProvidersEdgesDidPoolProxy initializes the proxy with all data needed to communicate with Genesys Cloud
func newTelephonyProvidersEdgesDidPoolProxy(clientConfig *platformclientv2.Configuration) *telephonyDidPoolProxy {
	api := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(clientConfig)
	return &telephonyDidPoolProxy{
		clientConfig:                                 clientConfig,
		telephonyApi:                                 api,
		createTelephonyDidPoolAttr:                   createTelephonyDidPoolFn,
		getTelephonyDidPoolByIdAttr:                  getTelephonyDidPoolByIdFn,
		updateEdgesDidPoolAttr:                       updateEdgesDidPoolFn,
		deleteTelephonyDidPoolAttr:                   deleteTelephonyDidPoolFn,
		getTelephonyDidPoolIdByStartAndEndNumberAttr: getTelephonyDidPoolIdByStartAndEndNumberFn,
		getAllTelephonyDidPoolsAttr:                  getAllTelephonyDidPoolsFn,
	}
}

// getTelephonyDidPoolProxy acts as a singleton for the internalProxy. It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getTelephonyDidPoolProxy(clientConfig *platformclientv2.Configuration) *telephonyDidPoolProxy {
	if internalProxy == nil {
		internalProxy = newTelephonyProvidersEdgesDidPoolProxy(clientConfig)
	}
	return internalProxy
}

// createTelephonyDidPool creates a Genesys Cloud did pool
func (t *telephonyDidPoolProxy) createTelephonyDidPool(ctx context.Context, didPool *platformclientv2.Didpool) (*platformclientv2.Didpool, error) {
	return t.createTelephonyDidPoolAttr(ctx, t, didPool)
}

// getTelephonyDidPoolById reads a Genesys Cloud did pool by id
func (t *telephonyDidPoolProxy) getTelephonyDidPoolById(ctx context.Context, id string) (*platformclientv2.Didpool, int, error) {
	return t.getTelephonyDidPoolByIdAttr(ctx, t, id)
}

// updateTelephonyDidPool update a Genesys Cloud did pool
func (t *telephonyDidPoolProxy) updateTelephonyDidPool(ctx context.Context, id string, didPool *platformclientv2.Didpool) (*platformclientv2.Didpool, error) {
	return t.updateEdgesDidPoolAttr(ctx, t, id, didPool)
}

// deleteTelephonyDidPool delete a Genesys Cloud did pool
func (t *telephonyDidPoolProxy) deleteTelephonyDidPool(ctx context.Context, id string) error {
	return t.deleteTelephonyDidPoolAttr(ctx, t, id)
}

// getTelephonyDidPoolIdByStartAndEndNumber find a Genesys Cloud did pool id using the start and end number
func (t *telephonyDidPoolProxy) getTelephonyDidPoolIdByStartAndEndNumber(ctx context.Context, start, end string) (string, bool, error) {
	return t.getTelephonyDidPoolIdByStartAndEndNumberAttr(ctx, t, start, end)
}

// getAllTelephonyDidPools read all Genesys Cloud did pools
func (t *telephonyDidPoolProxy) getAllTelephonyDidPools(ctx context.Context) (*[]platformclientv2.Didpool, error) {
	return t.getAllTelephonyDidPoolsAttr(ctx, t)
}

// createTelephonyDidPoolFn is an implementation function for creating a Genesys Cloud did pool
func createTelephonyDidPoolFn(_ context.Context, t *telephonyDidPoolProxy, didPool *platformclientv2.Didpool) (*platformclientv2.Didpool, error) {
	postDidPool, _, err := t.telephonyApi.PostTelephonyProvidersEdgesDidpools(*didPool)
	if err != nil {
		return nil, err
	}
	return postDidPool, nil
}

// getTelephonyDidPoolByIdFn is an implementation function for reading a Genesys Cloud did pool by ID
func getTelephonyDidPoolByIdFn(_ context.Context, t *telephonyDidPoolProxy, id string) (*platformclientv2.Didpool, int, error) {
	didPool, resp, err := t.telephonyApi.GetTelephonyProvidersEdgesDidpool(id)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return didPool, resp.StatusCode, nil
}

// updateEdgesDidPoolFn is an implementation function for updating a Genesys Cloud did pool
func updateEdgesDidPoolFn(_ context.Context, t *telephonyDidPoolProxy, id string, didPool *platformclientv2.Didpool) (*platformclientv2.Didpool, error) {
	updatedDidPool, _, err := t.telephonyApi.PutTelephonyProvidersEdgesDidpool(id, *didPool)
	if err != nil {
		return nil, err
	}
	return updatedDidPool, nil
}

// deleteTelephonyDidPoolFn is an implementation function for deleting a Genesys Cloud did pool
func deleteTelephonyDidPoolFn(_ context.Context, t *telephonyDidPoolProxy, id string) error {
	_, err := t.telephonyApi.DeleteTelephonyProvidersEdgesDidpool(id)
	return err
}

// getAllTelephonyDidPoolsFn is an implementation function for reading all Genesys Cloud did pools
func getAllTelephonyDidPoolsFn(_ context.Context, t *telephonyDidPoolProxy) (*[]platformclientv2.Didpool, error) {
	var (
		allDidPools []platformclientv2.Didpool
		pageCount   int
		pageNum     = 1
	)
	const pageSize = 100

	didPools, _, getErr := t.telephonyApi.GetTelephonyProvidersEdgesDidpools(pageSize, pageNum, "", nil)
	if getErr != nil {
		return nil, getErr
	}
	pageCount = *didPools.PageCount

	if didPools.Entities != nil && len(*didPools.Entities) > 0 {
		allDidPools = append(allDidPools, *didPools.Entities...)
	}

	if pageCount < 2 {
		return &allDidPools, nil
	}

	for pageNum := 2; pageNum <= pageCount; pageNum++ {
		didPools, _, getErr := t.telephonyApi.GetTelephonyProvidersEdgesDidpools(pageSize, pageNum, "", nil)
		if getErr != nil {
			return nil, getErr
		}

		if didPools.Entities == nil || len(*didPools.Entities) == 0 {
			break
		}

		allDidPools = append(allDidPools, *didPools.Entities...)
	}
	return &allDidPools, nil
}

// getTelephonyDidPoolIdByStartAndEndNumberFn is an implementation function for finding a Genesys Cloud did pool using the start and end number
func getTelephonyDidPoolIdByStartAndEndNumberFn(ctx context.Context, t *telephonyDidPoolProxy, start, end string) (string, bool, error) {
	allDidPools, err := getAllTelephonyDidPoolsFn(ctx, t)
	if err != nil {
		return "", false, fmt.Errorf("failed to read did pools: %v", err)
	}
	for _, didPool := range *allDidPools {
		if didPool.StartPhoneNumber != nil && *didPool.StartPhoneNumber == start &&
			didPool.EndPhoneNumber != nil && *didPool.EndPhoneNumber == end &&
			didPool.State != nil && *didPool.State != "deleted" {
			return *didPool.Id, false, nil
		}
	}
	return "", true, fmt.Errorf("failed to find DID pool with start phone number '%s' and end phone number '%s'", start, end)
}
