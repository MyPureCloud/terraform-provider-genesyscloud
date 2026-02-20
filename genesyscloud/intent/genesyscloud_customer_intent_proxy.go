package customer_intent

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v178/platformclientv2"
	"log"
)

/*
The genesyscloud_customer_intent_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *customerIntentProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createCustomerIntentFunc func(ctx context.Context, p *customerIntentProxy, customerIntentResponse *platformclientv2.Customerintentresponse) (*platformclientv2.Customerintentresponse, *platformclientv2.APIResponse, error)
type getAllCustomerIntentFunc func(ctx context.Context, p *customerIntentProxy) (*[]platformclientv2.Customerintentresponse, *platformclientv2.APIResponse, error)
type getCustomerIntentIdByNameFunc func(ctx context.Context, p *customerIntentProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type getCustomerIntentByIdFunc func(ctx context.Context, p *customerIntentProxy, id string) (*platformclientv2.Customerintentresponse, *platformclientv2.APIResponse, error)
type updateCustomerIntentFunc func(ctx context.Context, p *customerIntentProxy, id string, customerIntentResponse *platformclientv2.Customerintentresponse) (*platformclientv2.Customerintentresponse, *platformclientv2.APIResponse, error)
type deleteCustomerIntentFunc func(ctx context.Context, p *customerIntentProxy, id string) (*platformclientv2.APIResponse, error)

// customerIntentProxy contains all of the methods that call genesys cloud APIs.
type customerIntentProxy struct {
	clientConfig                  *platformclientv2.Configuration
	intentsApi                    *platformclientv2.IntentsApi
	createCustomerIntentAttr      createCustomerIntentFunc
	getAllCustomerIntentAttr      getAllCustomerIntentFunc
	getCustomerIntentIdByNameAttr getCustomerIntentIdByNameFunc
	getCustomerIntentByIdAttr     getCustomerIntentByIdFunc
	updateCustomerIntentAttr      updateCustomerIntentFunc
	deleteCustomerIntentAttr      deleteCustomerIntentFunc
}

// newCustomerIntentProxy initializes the customer intent proxy with all of the data needed to communicate with Genesys Cloud
func newCustomerIntentProxy(clientConfig *platformclientv2.Configuration) *customerIntentProxy {
	api := platformclientv2.NewIntentsApiWithConfig(clientConfig)
	return &customerIntentProxy{
		clientConfig:                  clientConfig,
		intentsApi:                    api,
		createCustomerIntentAttr:      createCustomerIntentFn,
		getAllCustomerIntentAttr:      getAllCustomerIntentFn,
		getCustomerIntentIdByNameAttr: getCustomerIntentIdByNameFn,
		getCustomerIntentByIdAttr:     getCustomerIntentByIdFn,
		updateCustomerIntentAttr:      updateCustomerIntentFn,
		deleteCustomerIntentAttr:      deleteCustomerIntentFn,
	}
}

// getCustomerIntentProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getCustomerIntentProxy(clientConfig *platformclientv2.Configuration) *customerIntentProxy {
	if internalProxy == nil {
		internalProxy = newCustomerIntentProxy(clientConfig)
	}

	return internalProxy
}

// createCustomerIntent creates a Genesys Cloud customer intent
func (p *customerIntentProxy) createCustomerIntent(ctx context.Context, customerIntent *platformclientv2.Customerintentresponse) (*platformclientv2.Customerintentresponse, *platformclientv2.APIResponse, error) {
	return p.createCustomerIntentAttr(ctx, p, customerIntent)
}

// getCustomerIntent retrieves all Genesys Cloud customer intent
func (p *customerIntentProxy) getAllCustomerIntent(ctx context.Context) (*[]platformclientv2.Customerintentresponse, *platformclientv2.APIResponse, error) {
	return p.getAllCustomerIntentAttr(ctx, p)
}

// getCustomerIntentIdByName returns a single Genesys Cloud customer intent by a name
func (p *customerIntentProxy) getCustomerIntentIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getCustomerIntentIdByNameAttr(ctx, p, name)
}

// getCustomerIntentById returns a single Genesys Cloud customer intent by Id
func (p *customerIntentProxy) getCustomerIntentById(ctx context.Context, id string) (*platformclientv2.Customerintentresponse, *platformclientv2.APIResponse, error) {
	return p.getCustomerIntentByIdAttr(ctx, p, id)
}

// updateCustomerIntent updates a Genesys Cloud customer intent
func (p *customerIntentProxy) updateCustomerIntent(ctx context.Context, id string, customerIntent *platformclientv2.Customerintentresponse) (*platformclientv2.Customerintentresponse, *platformclientv2.APIResponse, error) {
	return p.updateCustomerIntentAttr(ctx, p, id, customerIntent)
}

// deleteCustomerIntent deletes a Genesys Cloud customer intent by Id
func (p *customerIntentProxy) deleteCustomerIntent(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteCustomerIntentAttr(ctx, p, id)
}

// createCustomerIntentFn is an implementation function for creating a Genesys Cloud customer intent
func createCustomerIntentFn(ctx context.Context, p *customerIntentProxy, customerIntent *platformclientv2.Customerintentresponse) (*platformclientv2.Customerintentresponse, *platformclientv2.APIResponse, error) {
	// Convert Customerintentresponse to Customerintent for the API call
	body := platformclientv2.Customerintent{
		Name:        customerIntent.Name,
		Description: customerIntent.Description,
		ExpiryTime:  customerIntent.ExpiryTime,
	}
	// Add CategoryId if Category is set
	if customerIntent.Category != nil && customerIntent.Category.Id != nil {
		body.CategoryId = customerIntent.Category.Id
	}
	return p.intentsApi.PostIntentsCustomerintents(body)
}

// getAllCustomerIntentFn is the implementation for retrieving all customer intent in Genesys Cloud
func getAllCustomerIntentFn(ctx context.Context, p *customerIntentProxy) (*[]platformclientv2.Customerintentresponse, *platformclientv2.APIResponse, error) {
	var allCustomerIntentResponses []platformclientv2.Customerintentresponse
	const pageSize = 100

	customerIntentResponses, resp, err := p.intentsApi.GetIntentsCustomerintents(pageSize, 1, "", "")
	if err != nil {
		return nil, resp, err
	}
	if customerIntentResponses.Entities == nil || len(*customerIntentResponses.Entities) == 0 {
		return &allCustomerIntentResponses, resp, nil
	}
	for _, customerIntentResponse := range *customerIntentResponses.Entities {
		allCustomerIntentResponses = append(allCustomerIntentResponses, customerIntentResponse)
	}

	for pageNum := 2; pageNum <= *customerIntentResponses.PageCount; pageNum++ {
		customerIntentResponses, _, err := p.intentsApi.GetIntentsCustomerintents(pageSize, pageNum, "", "")
		if err != nil {
			return nil, resp, err
		}

		if customerIntentResponses.Entities == nil || len(*customerIntentResponses.Entities) == 0 {
			break
		}

		for _, customerIntentResponse := range *customerIntentResponses.Entities {
			allCustomerIntentResponses = append(allCustomerIntentResponses, customerIntentResponse)
		}
	}

	return &allCustomerIntentResponses, resp, nil
}

// getCustomerIntentIdByNameFn is an implementation of the function to get a Genesys Cloud customer intent by name
func getCustomerIntentIdByNameFn(ctx context.Context, p *customerIntentProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	customerIntentResponses, resp, err := p.intentsApi.GetIntentsCustomerintents(100, 1, name, "")
	if err != nil {
		return "", resp, false, err
	}

	if customerIntentResponses.Entities == nil || len(*customerIntentResponses.Entities) == 0 {
		return "", resp, true, err
	}

	for _, customerIntentResponse := range *customerIntentResponses.Entities {
		if *customerIntentResponse.Name == name {
			log.Printf("Retrieved the customer intent id %s by name %s", *customerIntentResponse.Id, name)
			return *customerIntentResponse.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("Unable to find customer intent with name %s", name)
}

// getCustomerIntentByIdFn is an implementation of the function to get a Genesys Cloud customer intent by Id
func getCustomerIntentByIdFn(ctx context.Context, p *customerIntentProxy, id string) (*platformclientv2.Customerintentresponse, *platformclientv2.APIResponse, error) {
	return p.intentsApi.GetIntentsCustomerintent(id)
}

// updateCustomerIntentFn is an implementation of the function to update a Genesys Cloud customer intent
func updateCustomerIntentFn(ctx context.Context, p *customerIntentProxy, id string, customerIntent *platformclientv2.Customerintentresponse) (*platformclientv2.Customerintentresponse, *platformclientv2.APIResponse, error) {
	// Convert Customerintentresponse to Customerintentpatch for the API call
	body := platformclientv2.Customerintentpatch{
		Name:        customerIntent.Name,
		Description: customerIntent.Description,
		ExpiryTime:  customerIntent.ExpiryTime,
	}
	// Add CategoryId if Category is set
	if customerIntent.Category != nil && customerIntent.Category.Id != nil {
		body.CategoryId = customerIntent.Category.Id
	}
	return p.intentsApi.PatchIntentsCustomerintent(id, body)
}

// deleteCustomerIntentFn is an implementation function for deleting a Genesys Cloud customer intent
func deleteCustomerIntentFn(ctx context.Context, p *customerIntentProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.intentsApi.DeleteIntentsCustomerintent(id)
}
