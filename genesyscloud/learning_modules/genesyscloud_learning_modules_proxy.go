package learning_modules

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

/*
The genesyscloud_learning_modules_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *learningModulesProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createLearningModuleFunc func(ctx context.Context, p *learningModulesProxy, learningModule *platformclientv2.Learningmodulerequest) (*platformclientv2.Learningmodule, *platformclientv2.APIResponse, error)
type publishLearningModuleFunc func(ctx context.Context, p *learningModulesProxy, id string) (*platformclientv2.Learningmodulepublishresponse, *platformclientv2.APIResponse, error)
type getAllLearningModulesFunc func(ctx context.Context, p *learningModulesProxy, searchTerm string) (*[]platformclientv2.Learningmodule, *platformclientv2.APIResponse, error)
type getLearningModuleIdByNameFunc func(ctx context.Context, p *learningModulesProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getLearningModuleByIdFunc func(ctx context.Context, p *learningModulesProxy, id string) (learningModule *platformclientv2.Learningmodule, response *platformclientv2.APIResponse, err error)
type updateLearningModuleFunc func(ctx context.Context, p *learningModulesProxy, id string, learningModule *platformclientv2.Learningmodulerequest) (*platformclientv2.Learningmodule, *platformclientv2.APIResponse, error)
type deleteLearningModuleFunc func(ctx context.Context, p *learningModulesProxy, id string) (response *platformclientv2.APIResponse, err error)

// learningModulesProxy contains all of the methods that call genesys cloud APIs.
type learningModulesProxy struct {
	clientConfig                  *platformclientv2.Configuration
	learningApi                   *platformclientv2.LearningApi
	createLearningModuleAttr      createLearningModuleFunc
	publishLearningModuleAttr     publishLearningModuleFunc
	getAllLearningModulesAttr     getAllLearningModulesFunc
	getLearningModuleIdByNameAttr getLearningModuleIdByNameFunc
	getLearningModuleByIdAttr     getLearningModuleByIdFunc
	updateLearningModuleAttr      updateLearningModuleFunc
	deleteLearningModuleAttr      deleteLearningModuleFunc
}

// newLearningModulesProxy initializes the learning modules proxy with all of the data needed to communicate with Genesys Cloud
func newLearningModulesProxy(clientConfig *platformclientv2.Configuration) *learningModulesProxy {
	api := platformclientv2.NewLearningApiWithConfig(clientConfig)
	return &learningModulesProxy{
		clientConfig:                  clientConfig,
		learningApi:                   api,
		createLearningModuleAttr:      createLearningModuleFn,
		publishLearningModuleAttr:     publishLearningModuleFn,
		getAllLearningModulesAttr:     getAllLearningModulesFn,
		getLearningModuleIdByNameAttr: getLearningModuleIdByNameFn,
		getLearningModuleByIdAttr:     getLearningModuleByIdFn,
		updateLearningModuleAttr:      updateLearningModuleFn,
		deleteLearningModuleAttr:      deleteLearningModuleFn,
	}
}

// getLearningModulesProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getLearningModulesProxy(clientConfig *platformclientv2.Configuration) *learningModulesProxy {
	if internalProxy == nil {
		internalProxy = newLearningModulesProxy(clientConfig)
	}
	return internalProxy
}

// createLearningModule creates a Genesys Cloud learning module
func (p *learningModulesProxy) createLearningModule(ctx context.Context, learningModule *platformclientv2.Learningmodulerequest) (*platformclientv2.Learningmodule, *platformclientv2.APIResponse, error) {
	return p.createLearningModuleAttr(ctx, p, learningModule)
}

// publishLearningModule publishes a Genesys Cloud learning module
func (p *learningModulesProxy) publishLearningModule(ctx context.Context, id string) (*platformclientv2.Learningmodulepublishresponse, *platformclientv2.APIResponse, error) {
	return p.publishLearningModuleAttr(ctx, p, id)
}

// getAllLearningModules retrieves all Genesys Cloud learning modules
func (p *learningModulesProxy) getAllLearningModules(ctx context.Context, searchTerm string) (*[]platformclientv2.Learningmodule, *platformclientv2.APIResponse, error) {
	return p.getAllLearningModulesAttr(ctx, p, searchTerm)
}

// getLearningModuleIdByName returns a single Genesys Cloud learning module by name
func (p *learningModulesProxy) getLearningModuleIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getLearningModuleIdByNameAttr(ctx, p, name)
}

// getLearningModuleById returns a single Genesys Cloud learning module by Id
func (p *learningModulesProxy) getLearningModuleById(ctx context.Context, id string) (learningModule *platformclientv2.Learningmodule, response *platformclientv2.APIResponse, err error) {
	return p.getLearningModuleByIdAttr(ctx, p, id)
}

// updateLearningModule updates a Genesys Cloud learning module
func (p *learningModulesProxy) updateLearningModule(ctx context.Context, id string, learningModule *platformclientv2.Learningmodulerequest) (*platformclientv2.Learningmodule, *platformclientv2.APIResponse, error) {
	return p.updateLearningModuleAttr(ctx, p, id, learningModule)
}

// deleteLearningModule deletes a Genesys Cloud learning module by Id
func (p *learningModulesProxy) deleteLearningModule(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteLearningModuleAttr(ctx, p, id)
}

// createLearningModuleFn is an implementation function for creating a Genesys Cloud learning module
func createLearningModuleFn(ctx context.Context, p *learningModulesProxy, learningModule *platformclientv2.Learningmodulerequest) (*platformclientv2.Learningmodule, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	module, resp, err := p.learningApi.PostLearningModules(*learningModule)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to create learning module: %s", err)
	}
	return module, resp, nil
}

// publishLearningModuleFn is an implementation function for publishing a Genesys Cloud learning module
func publishLearningModuleFn(ctx context.Context, p *learningModulesProxy, id string) (*platformclientv2.Learningmodulepublishresponse, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	request := platformclientv2.Learningmodulepublishrequest{
		TermsAndConditionsAccepted: platformclientv2.Bool(true),
	}
	module, resp, err := p.learningApi.PostLearningModulePublish(id, request)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to publish learning module: %s", err)
	}
	return module, resp, nil
}

// getAllLearningModulesFn is the implementation for retrieving all learning modules in Genesys Cloud
func getAllLearningModulesFn(ctx context.Context, p *learningModulesProxy, searchTerm string) (*[]platformclientv2.Learningmodule, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	var allModules []platformclientv2.Learningmodule
	const isArchived = true
	types := []string{}
	const pageSize = 100
	const pageNumber = 1
	const sortOrder = ""
	const sortBy = ""
	expand := []string{}
	const isPublished = ""
	statuses := []string{}
	externalIds := []string{}

	modules, resp, err := p.learningApi.GetLearningModules(
		isArchived,
		types,
		pageSize,
		pageNumber,
		sortOrder,
		sortBy,
		searchTerm,
		expand,
		isPublished,
		statuses,
		externalIds,
	)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get learning modules: %v", err)
	}
	if modules.Entities == nil || len(*modules.Entities) == 0 {
		return &allModules, resp, nil
	}
	for _, module := range *modules.Entities {
		allModules = append(allModules, module)
	}

	for pageNum := 2; pageNum <= *modules.PageCount; pageNum++ {
		modules, resp, err := p.learningApi.GetLearningModules(
			isArchived,
			types,
			pageSize,
			pageNum,
			sortOrder,
			sortBy,
			searchTerm,
			expand,
			isPublished,
			statuses,
			externalIds,
		)
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get learning modules: %v", err)
		}

		if modules.Entities == nil || len(*modules.Entities) == 0 {
			break
		}

		for _, module := range *modules.Entities {
			allModules = append(allModules, module)
		}
	}

	return &allModules, resp, nil
}

// getLearningModuleIdByNameFn is an implementation of the function to get a Genesys Cloud learning module by name
func getLearningModuleIdByNameFn(ctx context.Context, p *learningModulesProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	modules, resp, err := getAllLearningModulesFn(ctx, p, name)
	if err != nil {
		return "", false, resp, err
	}

	if modules == nil || len(*modules) == 0 {
		return "", true, resp, fmt.Errorf("No learning module found with name %s", name)
	}

	for _, module := range *modules {
		if *module.Name == name {
			log.Printf("Retrieved the learning module id %s by name %s", *module.Id, name)
			return *module.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("Unable to find learning module with name %s", name)
}

// getLearningModuleByIdFn is an implementation of the function to get a Genesys Cloud learning module by Id
func getLearningModuleByIdFn(ctx context.Context, p *learningModulesProxy, id string) (learningModule *platformclientv2.Learningmodule, response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	module, resp, err := p.learningApi.GetLearningModule(id, []string{"assessmentForm"})
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to retrieve learning module by id %s: %s", id, err)
	}
	return module, resp, nil
}

// updateLearningModuleFn is an implementation of the function to update a Genesys Cloud learning module
func updateLearningModuleFn(ctx context.Context, p *learningModulesProxy, id string, learningModule *platformclientv2.Learningmodulerequest) (*platformclientv2.Learningmodule, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	module, resp, err := p.learningApi.PutLearningModule(id, *learningModule)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to update learning module: %s", err)
	}
	return module, resp, nil
}

// deleteLearningModuleFn is an implementation function for deleting a Genesys Cloud learning module
func deleteLearningModuleFn(ctx context.Context, p *learningModulesProxy, id string) (response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	resp, err := p.learningApi.DeleteLearningModule(id)
	if err != nil {
		return resp, fmt.Errorf("Failed to delete learning module: %s", err)
	}
	return resp, nil
}
