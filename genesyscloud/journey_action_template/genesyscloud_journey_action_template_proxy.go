package journey_action_template

import (
	"context"
	"fmt"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

/*
The file genesyscloud_journey_action_template_proxy.go manages the interaction between our software and
the Genesys Cloud SDK. Within this file, we define proxy structures and methods.
We employ a technique called composition for each function on the proxy. This means that each function
is built by combining smaller, independent parts. One advantage of this approach is that it allows us
to isolate and test individual functions more easily. For testing purposes, we can replace or
simulate these smaller parts, known as stubs, to ensure that each function behaves correctly in different scenarios.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *journeyActionTemplateProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createJourneyActionTemplateFunc func(ctx context.Context, p *journeyActionTemplateProxy, template *platformclientv2.Actiontemplate) (*platformclientv2.Actiontemplate, *platformclientv2.APIResponse, error)
type getAllJourneyActionTemplatesFunc func(ctx context.Context, p *journeyActionTemplateProxy) (*[]platformclientv2.Actiontemplate, *platformclientv2.APIResponse, error)
type getJourneyActionTemplateIdByNameFunc func(ctx context.Context, p *journeyActionTemplateProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getJourneyActionTemplateByIdFunc func(ctx context.Context, p *journeyActionTemplateProxy, id string) (template *platformclientv2.Actiontemplate, response *platformclientv2.APIResponse, err error)
type updateJourneyActionTemplateFunc func(ctx context.Context, p *journeyActionTemplateProxy, id string, template *platformclientv2.Patchactiontemplate) (*platformclientv2.Actiontemplate, *platformclientv2.APIResponse, error)
type deleteJourneyActionTemplateFunc func(ctx context.Context, p *journeyActionTemplateProxy, id string) (*platformclientv2.APIResponse, error)

/*
The journeyActionTemplateProxy struct holds all the methods responsible for making calls to
the Genesys Cloud APIs. This means that within this struct, you'll find all the functions designed
to interact directly with the various features and services offered by Genesys Cloud,
enabling this terraform provider software to perform tasks like retrieving data, updating information,
or triggering actions within the Genesys Cloud environment.
*/
type journeyActionTemplateProxy struct {
	clientConfig                         *platformclientv2.Configuration
	journeyApi                           *platformclientv2.JourneyApi
	createJourneyActionTemplateAttr      createJourneyActionTemplateFunc
	getAllJourneyActionTemplatesAttr     getAllJourneyActionTemplatesFunc
	getJourneyActionTemplateIdByNameAttr getJourneyActionTemplateIdByNameFunc
	getJourneyActionTemplateByIdAttr     getJourneyActionTemplateByIdFunc
	updateJourneyActionTemplateAttr      updateJourneyActionTemplateFunc
	deleteJourneyActionTemplateAttr      deleteJourneyActionTemplateFunc
	templateCache                        rc.CacheInterface[platformclientv2.Actiontemplate]
}

/*
The function newJourneyActionTemplateProxy sets up the journey action template proxy by providing it
with all the necessary information to communicate effectively with Genesys Cloud.
This includes configuring the proxy with the required data and settings so that it can interact
seamlessly with the Genesys Cloud platform.
*/
func newJourneyActionTemplateProxy(clientConfig *platformclientv2.Configuration) *journeyActionTemplateProxy {
	api := platformclientv2.NewJourneyApiWithConfig(clientConfig)
	templateCache := rc.NewResourceCache[platformclientv2.Actiontemplate]()

	return &journeyActionTemplateProxy{
		clientConfig:                         clientConfig,
		journeyApi:                           api,
		templateCache:                        templateCache,
		createJourneyActionTemplateAttr:      createJourneyActionTemplateFn,
		getAllJourneyActionTemplatesAttr:     getAllJourneyActionTemplatesFn,
		getJourneyActionTemplateIdByNameAttr: getJourneyActionTemplateIdByNameFn,
		getJourneyActionTemplateByIdAttr:     getJourneyActionTemplateByIdFn,
		updateJourneyActionTemplateAttr:      updateJourneyActionTemplateFn,
		deleteJourneyActionTemplateAttr:      deleteJourneyActionTemplateFn,
	}
}

/*
The function getJourneyActionTemplateProxy serves a dual purpose: first, it functions as a singleton for
the internalProxy, meaning it ensures that only one instance of the internalProxy exists. Second,
it enables us to proxy our tests by allowing us to directly set the internalProxy package variable.
This ensures consistency and control in managing the internalProxy across our codebase, while also
facilitating efficient testing by providing a straightforward way to substitute the proxy for testing purposes.
*/
func getJourneyActionTemplateProxy(clientConfig *platformclientv2.Configuration) *journeyActionTemplateProxy {
	if internalProxy == nil {
		internalProxy = newJourneyActionTemplateProxy(clientConfig)
	}
	return internalProxy
}

// createJourneyActionTemplate creates a Genesys Cloud journey action template
func (p *journeyActionTemplateProxy) createJourneyActionTemplate(ctx context.Context, template *platformclientv2.Actiontemplate) (*platformclientv2.Actiontemplate, *platformclientv2.APIResponse, error) {
	return p.createJourneyActionTemplateAttr(ctx, p, template)
}

// getAllJourneyActionTemplates retrieves all Genesys Cloud journey action templates
func (p *journeyActionTemplateProxy) getAllJourneyActionTemplates(ctx context.Context) (*[]platformclientv2.Actiontemplate, *platformclientv2.APIResponse, error) {
	return p.getAllJourneyActionTemplatesAttr(ctx, p)
}

// getJourneyActionTemplateIdByName returns a single Genesys Cloud journey action template by name
func (p *journeyActionTemplateProxy) getJourneyActionTemplateIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getJourneyActionTemplateIdByNameAttr(ctx, p, name)
}

// getJourneyActionTemplateById returns a single Genesys Cloud journey action template by Id
func (p *journeyActionTemplateProxy) getJourneyActionTemplateById(ctx context.Context, id string) (template *platformclientv2.Actiontemplate, response *platformclientv2.APIResponse, err error) {
	if template := rc.GetCacheItem(p.templateCache, id); template != nil {
		return template, nil, nil
	}
	return p.getJourneyActionTemplateByIdAttr(ctx, p, id)
}

// updateJourneyActionTemplate updates a Genesys Cloud journey action template
func (p *journeyActionTemplateProxy) updateJourneyActionTemplate(ctx context.Context, id string, template *platformclientv2.Patchactiontemplate) (*platformclientv2.Actiontemplate, *platformclientv2.APIResponse, error) {
	return p.updateJourneyActionTemplateAttr(ctx, p, id, template)
}

// deleteJourneyActionTemplate deletes a Genesys Cloud journey action template by Id
func (p *journeyActionTemplateProxy) deleteJourneyActionTemplate(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteJourneyActionTemplateAttr(ctx, p, id)
}

// createJourneyActionTemplateFn is an implementation function for creating a Genesys Cloud journey action template
func createJourneyActionTemplateFn(ctx context.Context, p *journeyActionTemplateProxy, template *platformclientv2.Actiontemplate) (*platformclientv2.Actiontemplate, *platformclientv2.APIResponse, error) {
	actionTemplate, resp, err := p.journeyApi.PostJourneyActiontemplates(*template)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to create journey action template: %s", err)
	}
	return actionTemplate, resp, nil
}

// getAllJourneyActionTemplatesFn is the implementation for retrieving all journey action templates in Genesys Cloud
func getAllJourneyActionTemplatesFn(ctx context.Context, p *journeyActionTemplateProxy) (*[]platformclientv2.Actiontemplate, *platformclientv2.APIResponse, error) {
	var allTemplates []platformclientv2.Actiontemplate
	const pageSize = 100

	templates, resp, err := p.journeyApi.GetJourneyActiontemplates(1, pageSize, "", "", "", nil, "")
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get journey action templates: %s", err)
	}

	if templates.Entities == nil || len(*templates.Entities) == 0 {
		return &allTemplates, resp, nil
	}

	allTemplates = append(allTemplates, *templates.Entities...)

	for pageNum := 2; pageNum <= *templates.PageCount; pageNum++ {
		templates, resp, err := p.journeyApi.GetJourneyActiontemplates(pageNum, pageSize, "", "", "", nil, "")
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get journey action templates page %d: %s", pageNum, err)
		}
		if templates.Entities == nil || len(*templates.Entities) == 0 {
			break
		}

		allTemplates = append(allTemplates, *templates.Entities...)
	}

	// Cache the action templates for later use
	for _, template := range allTemplates {
		rc.SetCache(p.templateCache, *template.Id, template)
	}

	return &allTemplates, resp, nil
}

// getJourneyActionTemplateIdByNameFn is an implementation function for getting a journey action template by name
func getJourneyActionTemplateIdByNameFn(ctx context.Context, p *journeyActionTemplateProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	templates, resp, err := p.getAllJourneyActionTemplates(ctx)
	if err != nil {
		return "", false, resp, err
	}

	if templates == nil || len(*templates) == 0 {
		return "", true, resp, fmt.Errorf("No journey action template found with name %s", name)
	}

	for _, template := range *templates {
		if *template.Name == name {
			return *template.Id, false, resp, nil
		}
	}

	return "", true, resp, fmt.Errorf("No journey action template found with name %s", name)
}

// getJourneyActionTemplateByIdFn is an implementation function for getting a journey action template by ID
func getJourneyActionTemplateByIdFn(ctx context.Context, p *journeyActionTemplateProxy, id string) (template *platformclientv2.Actiontemplate, response *platformclientv2.APIResponse, err error) {
	template, resp, err := p.journeyApi.GetJourneyActiontemplate(id)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get journey action template %s: %s", id, err)
	}
	return template, resp, nil
}

// updateJourneyActionTemplateFn is an implementation function for updating a journey action template
func updateJourneyActionTemplateFn(ctx context.Context, p *journeyActionTemplateProxy, id string, template *platformclientv2.Patchactiontemplate) (*platformclientv2.Actiontemplate, *platformclientv2.APIResponse, error) {
	templateResp, resp, err := p.journeyApi.PatchJourneyActiontemplate(id, *template)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to update journey action template %s: %s", id, err)
	}
	return templateResp, resp, nil
}

// deleteJourneyActionTemplateFn is an implementation function for deleting a journey action template
func deleteJourneyActionTemplateFn(ctx context.Context, p *journeyActionTemplateProxy, id string) (*platformclientv2.APIResponse, error) {
	resp, err := p.journeyApi.DeleteJourneyActiontemplate(id, true)
	if err != nil {
		return resp, fmt.Errorf("Failed to delete journey action template %s: %s", id, err)
	}
	return resp, nil
}
