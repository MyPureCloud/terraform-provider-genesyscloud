package speechandtextanalytics_category

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

/*
The genesyscloud_speechandtextanalytics_category_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *categoryProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createCategoryFunc func(ctx context.Context, p *categoryProxy, category *platformclientv2.Categoryrequest) (*platformclientv2.Stacategory, *platformclientv2.APIResponse, error)
type getAllCategoriesFunc func(ctx context.Context, p *categoryProxy) (*[]platformclientv2.Stacategory, *platformclientv2.APIResponse, error)
type getCategoryIdByNameFunc func(ctx context.Context, p *categoryProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getCategoryByIdFunc func(ctx context.Context, p *categoryProxy, id string) (category *platformclientv2.Stacategory, response *platformclientv2.APIResponse, err error)
type updateCategoryFunc func(ctx context.Context, p *categoryProxy, id string, category *platformclientv2.Categoryrequest) (*platformclientv2.Stacategory, *platformclientv2.APIResponse, error)
type deleteCategoryFunc func(ctx context.Context, p *categoryProxy, id string) (response *platformclientv2.APIResponse, err error)

// categoryProxy contains all of the methods that call genesys cloud APIs.
type categoryProxy struct {
	clientConfig            *platformclientv2.Configuration
	speechTextAnalyticsApi  *platformclientv2.SpeechTextAnalyticsApi
	createCategoryAttr      createCategoryFunc
	getAllCategoriesAttr    getAllCategoriesFunc
	getCategoryIdByNameAttr getCategoryIdByNameFunc
	getCategoryByIdAttr     getCategoryByIdFunc
	updateCategoryAttr      updateCategoryFunc
	deleteCategoryAttr      deleteCategoryFunc
}

// newCategoryProxy initializes the category proxy with all of the data needed to communicate with Genesys Cloud
func newCategoryProxy(clientConfig *platformclientv2.Configuration) *categoryProxy {
	api := platformclientv2.NewSpeechTextAnalyticsApiWithConfig(clientConfig)
	return &categoryProxy{
		clientConfig:            clientConfig,
		speechTextAnalyticsApi:  api,
		createCategoryAttr:      createCategoryFn,
		getAllCategoriesAttr:    getAllCategoriesFn,
		getCategoryIdByNameAttr: getCategoryIdByNameFn,
		getCategoryByIdAttr:     getCategoryByIdFn,
		updateCategoryAttr:      updateCategoryFn,
		deleteCategoryAttr:      deleteCategoryFn,
	}
}

// getCategoryProxy acts as a singleton for the internalProxy. It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getCategoryProxy(clientConfig *platformclientv2.Configuration) *categoryProxy {
	if internalProxy == nil {
		internalProxy = newCategoryProxy(clientConfig)
	}
	return internalProxy
}

// createCategory creates a Genesys Cloud category
func (p *categoryProxy) createCategory(ctx context.Context, category *platformclientv2.Categoryrequest) (*platformclientv2.Stacategory, *platformclientv2.APIResponse, error) {
	return p.createCategoryAttr(ctx, p, category)
}

// getAllCategories retrieves all Genesys Cloud categories
func (p *categoryProxy) getAllCategories(ctx context.Context) (*[]platformclientv2.Stacategory, *platformclientv2.APIResponse, error) {
	return p.getAllCategoriesAttr(ctx, p)
}

// getCategoryIdByName returns a single Genesys Cloud category by name
func (p *categoryProxy) getCategoryIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getCategoryIdByNameAttr(ctx, p, name)
}

// getCategoryById returns a single Genesys Cloud category by Id
func (p *categoryProxy) getCategoryById(ctx context.Context, id string) (category *platformclientv2.Stacategory, response *platformclientv2.APIResponse, err error) {
	return p.getCategoryByIdAttr(ctx, p, id)
}

// updateCategory updates a Genesys Cloud category
func (p *categoryProxy) updateCategory(ctx context.Context, id string, category *platformclientv2.Categoryrequest) (*platformclientv2.Stacategory, *platformclientv2.APIResponse, error) {
	return p.updateCategoryAttr(ctx, p, id, category)
}

// deleteCategory deletes a Genesys Cloud category
func (p *categoryProxy) deleteCategory(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteCategoryAttr(ctx, p, id)
}

// createCategoryFn is an implementation function for creating a Genesys Cloud category
func createCategoryFn(ctx context.Context, p *categoryProxy, category *platformclientv2.Categoryrequest) (*platformclientv2.Stacategory, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	createdCategory, resp, err := p.speechTextAnalyticsApi.PostSpeechandtextanalyticsCategories(*category)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to create category: %s", err)
	}
	return createdCategory, resp, nil
}

// getAllCategoriesFn is the implementation for retrieving all categories in Genesys Cloud
func getAllCategoriesFn(ctx context.Context, p *categoryProxy) (*[]platformclientv2.Stacategory, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	var allCategories []platformclientv2.Stacategory
	const pageSize = 100

	categories, resp, err := p.speechTextAnalyticsApi.GetSpeechandtextanalyticsCategories(pageSize, 1, "", "", "", nil)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get categories: %v", err)
	}
	if categories.Entities == nil || len(*categories.Entities) == 0 {
		return &allCategories, resp, nil
	}
	allCategories = append(allCategories, *categories.Entities...)

	pageCount := 1
	if categories.PageCount != nil {
		pageCount = *categories.PageCount
	}

	for pageNum := 2; pageNum <= pageCount; pageNum++ {
		categories, resp, err = p.speechTextAnalyticsApi.GetSpeechandtextanalyticsCategories(pageSize, pageNum, "", "", "", nil)
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get categories: %v", err)
		}
		if categories.Entities == nil || len(*categories.Entities) == 0 {
			break
		}
		allCategories = append(allCategories, *categories.Entities...)
	}

	return &allCategories, resp, nil
}

// getCategoryIdByNameFn is an implementation of the function to get a Genesys Cloud category by name
func getCategoryIdByNameFn(ctx context.Context, p *categoryProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	categories, resp, getErr := p.getAllCategories(ctx)
	if getErr != nil {
		return "", false, resp, getErr
	}

	if categories == nil || len(*categories) == 0 {
		return "", true, resp, fmt.Errorf("no category found with name %s", name)
	}

	for _, category := range *categories {
		if category.Name != nil && *category.Name == name && category.Id != nil {
			log.Printf("Retrieved the category id %s by name %s", *category.Id, name)
			return *category.Id, false, resp, nil
		}
	}

	return "", true, resp, fmt.Errorf("unable to find category with name %s", name)
}

// getCategoryByIdFn is an implementation of the function to get a Genesys Cloud category by Id
func getCategoryByIdFn(ctx context.Context, p *categoryProxy, id string) (category *platformclientv2.Stacategory, response *platformclientv2.APIResponse, err error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	category, response, err = p.speechTextAnalyticsApi.GetSpeechandtextanalyticsCategory(id)
	if err != nil {
		return nil, response, fmt.Errorf("failed to retrieve category by id %s: %s", id, err)
	}
	return category, response, nil
}

// updateCategoryFn is an implementation of the function to update a Genesys Cloud category
func updateCategoryFn(ctx context.Context, p *categoryProxy, id string, category *platformclientv2.Categoryrequest) (*platformclientv2.Stacategory, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	updatedCategory, resp, err := p.speechTextAnalyticsApi.PutSpeechandtextanalyticsCategory(id, *category)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update category %s: %s", id, err)
	}
	return updatedCategory, resp, nil
}

// deleteCategoryFn is an implementation function for deleting a Genesys Cloud category
func deleteCategoryFn(ctx context.Context, p *categoryProxy, id string) (response *platformclientv2.APIResponse, err error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	response, err = p.speechTextAnalyticsApi.DeleteSpeechandtextanalyticsCategory(id)
	if err != nil {
		return response, fmt.Errorf("failed to delete category %s: %s", id, err)
	}
	return response, nil
}
