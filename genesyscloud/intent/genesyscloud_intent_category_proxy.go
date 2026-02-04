package intent_category

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v178/platformclientv2"
	"log"
)

/*
The genesyscloud_intent_category_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *intentCategoryProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createIntentCategoryFunc func(ctx context.Context, p *intentCategoryProxy, intentsCategory *platformclientv2.Intentscategory) (*platformclientv2.Intentscategory, *platformclientv2.APIResponse, error)
type getAllIntentCategoryFunc func(ctx context.Context, p *intentCategoryProxy) (*[]platformclientv2.Intentscategory, *platformclientv2.APIResponse, error)
type getIntentCategoryIdByNameFunc func(ctx context.Context, p *intentCategoryProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type getIntentCategoryByIdFunc func(ctx context.Context, p *intentCategoryProxy, id string) (*platformclientv2.Intentscategory, *platformclientv2.APIResponse, error)
type updateIntentCategoryFunc func(ctx context.Context, p *intentCategoryProxy, id string, intentsCategory *platformclientv2.Intentscategory) (*platformclientv2.Intentscategory, *platformclientv2.APIResponse, error)
type deleteIntentCategoryFunc func(ctx context.Context, p *intentCategoryProxy, id string) (*platformclientv2.APIResponse, error)

// intentCategoryProxy contains all of the methods that call genesys cloud APIs.
type intentCategoryProxy struct {
	clientConfig                  *platformclientv2.Configuration
	intentsApi                    *platformclientv2.IntentsApi
	createIntentCategoryAttr      createIntentCategoryFunc
	getAllIntentCategoryAttr      getAllIntentCategoryFunc
	getIntentCategoryIdByNameAttr getIntentCategoryIdByNameFunc
	getIntentCategoryByIdAttr     getIntentCategoryByIdFunc
	updateIntentCategoryAttr      updateIntentCategoryFunc
	deleteIntentCategoryAttr      deleteIntentCategoryFunc
}

// newIntentCategoryProxy initializes the intent category proxy with all of the data needed to communicate with Genesys Cloud
func newIntentCategoryProxy(clientConfig *platformclientv2.Configuration) *intentCategoryProxy {
	api := platformclientv2.NewIntentsApiWithConfig(clientConfig)
	return &intentCategoryProxy{
		clientConfig:                  clientConfig,
		intentsApi:                    api,
		createIntentCategoryAttr:      createIntentCategoryFn,
		getAllIntentCategoryAttr:      getAllIntentCategoryFn,
		getIntentCategoryIdByNameAttr: getIntentCategoryIdByNameFn,
		getIntentCategoryByIdAttr:     getIntentCategoryByIdFn,
		updateIntentCategoryAttr:      updateIntentCategoryFn,
		deleteIntentCategoryAttr:      deleteIntentCategoryFn,
	}
}

// getIntentCategoryProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getIntentCategoryProxy(clientConfig *platformclientv2.Configuration) *intentCategoryProxy {
	if internalProxy == nil {
		internalProxy = newIntentCategoryProxy(clientConfig)
	}

	return internalProxy
}

// createIntentCategory creates a Genesys Cloud intent category
func (p *intentCategoryProxy) createIntentCategory(ctx context.Context, intentCategory *platformclientv2.Intentscategory) (*platformclientv2.Intentscategory, *platformclientv2.APIResponse, error) {
	return p.createIntentCategoryAttr(ctx, p, intentCategory)
}

// getIntentCategory retrieves all Genesys Cloud intent category
func (p *intentCategoryProxy) getAllIntentCategory(ctx context.Context) (*[]platformclientv2.Intentscategory, *platformclientv2.APIResponse, error) {
	return p.getAllIntentCategoryAttr(ctx, p)
}

// getIntentCategoryIdByName returns a single Genesys Cloud intent category by a name
func (p *intentCategoryProxy) getIntentCategoryIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getIntentCategoryIdByNameAttr(ctx, p, name)
}

// getIntentCategoryById returns a single Genesys Cloud intent category by Id
func (p *intentCategoryProxy) getIntentCategoryById(ctx context.Context, id string) (*platformclientv2.Intentscategory, *platformclientv2.APIResponse, error) {
	return p.getIntentCategoryByIdAttr(ctx, p, id)
}

// updateIntentCategory updates a Genesys Cloud intent category
func (p *intentCategoryProxy) updateIntentCategory(ctx context.Context, id string, intentCategory *platformclientv2.Intentscategory) (*platformclientv2.Intentscategory, *platformclientv2.APIResponse, error) {
	return p.updateIntentCategoryAttr(ctx, p, id, intentCategory)
}

// deleteIntentCategory deletes a Genesys Cloud intent category by Id
func (p *intentCategoryProxy) deleteIntentCategory(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteIntentCategoryAttr(ctx, p, id)
}

// createIntentCategoryFn is an implementation function for creating a Genesys Cloud intent category
func createIntentCategoryFn(ctx context.Context, p *intentCategoryProxy, intentCategory *platformclientv2.Intentscategory) (*platformclientv2.Intentscategory, *platformclientv2.APIResponse, error) {
	return p.intentsApi.PostIntentsCategories(*intentCategory)
}

// getAllIntentCategoryFn is the implementation for retrieving all intent category in Genesys Cloud
func getAllIntentCategoryFn(ctx context.Context, p *intentCategoryProxy) (*[]platformclientv2.Intentscategory, *platformclientv2.APIResponse, error) {
	var allIntentsCategorys []platformclientv2.Intentscategory
	const pageSize = 100

	intentsCategorys, resp, err := p.intentsApi.GetIntentsCategories(pageSize, 1, "")
	if err != nil {
		return nil, resp, err
	}
	if intentsCategorys.Entities == nil || len(*intentsCategorys.Entities) == 0 {
		return &allIntentsCategorys, resp, nil
	}
	for _, intentsCategory := range *intentsCategorys.Entities {
		allIntentsCategorys = append(allIntentsCategorys, intentsCategory)
	}

	for pageNum := 2; pageNum <= *intentsCategorys.PageCount; pageNum++ {
		intentsCategorys, _, err := p.intentsApi.GetIntentsCategories(pageSize, pageNum, "")
		if err != nil {
			return nil, resp, err
		}

		if intentsCategorys.Entities == nil || len(*intentsCategorys.Entities) == 0 {
			break
		}

		for _, intentsCategory := range *intentsCategorys.Entities {
			allIntentsCategorys = append(allIntentsCategorys, intentsCategory)
		}
	}

	return &allIntentsCategorys, resp, nil
}

// getIntentCategoryIdByNameFn is an implementation of the function to get a Genesys Cloud intent category by name
func getIntentCategoryIdByNameFn(ctx context.Context, p *intentCategoryProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	const pageSize = 100
	intentsCategorys, resp, err := p.intentsApi.GetIntentsCategories(pageSize, 1, "")
	if err != nil {
		return "", resp, false, err
	}

	if intentsCategorys.Entities == nil || len(*intentsCategorys.Entities) == 0 {
		return "", resp, true, err
	}

	for _, intentsCategory := range *intentsCategorys.Entities {
		if *intentsCategory.Name == name {
			log.Printf("Retrieved the intent category id %s by name %s", *intentsCategory.Id, name)
			return *intentsCategory.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("Unable to find intent category with name %s", name)
}

// getIntentCategoryByIdFn is an implementation of the function to get a Genesys Cloud intent category by Id
func getIntentCategoryByIdFn(ctx context.Context, p *intentCategoryProxy, id string) (*platformclientv2.Intentscategory, *platformclientv2.APIResponse, error) {
	return p.intentsApi.GetIntentsCategory(id)
}

// updateIntentCategoryFn is an implementation of the function to update a Genesys Cloud intent category
func updateIntentCategoryFn(ctx context.Context, p *intentCategoryProxy, id string, intentCategory *platformclientv2.Intentscategory) (*platformclientv2.Intentscategory, *platformclientv2.APIResponse, error) {
	// Convert Intentscategory to Intentscategorypatch for the PATCH API
	patch := platformclientv2.Intentscategorypatch{
		Name:        intentCategory.Name,
		Description: intentCategory.Description,
	}
	return p.intentsApi.PatchIntentsCategory(id, patch)
}

// deleteIntentCategoryFn is an implementation function for deleting a Genesys Cloud intent category
func deleteIntentCategoryFn(ctx context.Context, p *intentCategoryProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.intentsApi.DeleteIntentsCategory(id)
}
