package outbound_contact_list_template

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_outbound_contact_list_template_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *outboundContactlisttemplateProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createOutboundContactlisttemplateFunc func(ctx context.Context, p *outboundContactlisttemplateProxy, Contactlisttemplate *platformclientv2.Contactlisttemplate) (*platformclientv2.Contactlisttemplate, *platformclientv2.APIResponse, error)
type getAllOutboundContactlisttemplateFunc func(ctx context.Context, p *outboundContactlisttemplateProxy, name string) (*[]platformclientv2.Contactlisttemplate, *platformclientv2.APIResponse, error)
type getOutboundContactlisttemplateIdByNameFunc func(ctx context.Context, p *outboundContactlisttemplateProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getOutboundContactlisttemplateByIdFunc func(ctx context.Context, p *outboundContactlisttemplateProxy, id string) (Contactlisttemplate *platformclientv2.Contactlisttemplate, response *platformclientv2.APIResponse, err error)
type updateOutboundContactlisttemplateFunc func(ctx context.Context, p *outboundContactlisttemplateProxy, id string, Contactlisttemplate *platformclientv2.Contactlisttemplate) (*platformclientv2.Contactlisttemplate, *platformclientv2.APIResponse, error)
type deleteOutboundContactlisttemplateFunc func(ctx context.Context, p *outboundContactlisttemplateProxy, id string) (response *platformclientv2.APIResponse, err error)

// outboundContactlisttemplateProxy contains all of the methods that call genesys cloud APIs.
type outboundContactlisttemplateProxy struct {
	clientConfig                               *platformclientv2.Configuration
	outboundApi                                *platformclientv2.OutboundApi
	createOutboundContactlisttemplateAttr      createOutboundContactlisttemplateFunc
	getAllOutboundContactlisttemplateAttr      getAllOutboundContactlisttemplateFunc
	getOutboundContactlisttemplateIdByNameAttr getOutboundContactlisttemplateIdByNameFunc
	getOutboundContactlisttemplateByIdAttr     getOutboundContactlisttemplateByIdFunc
	updateOutboundContactlisttemplateAttr      updateOutboundContactlisttemplateFunc
	deleteOutboundContactlisttemplateAttr      deleteOutboundContactlisttemplateFunc
}

// newOutboundContactlisttemplateProxy initializes the outbound Contactlisttemplate proxy with all of the data needed to communicate with Genesys Cloud
func newOutboundContactlisttemplateProxy(clientConfig *platformclientv2.Configuration) *outboundContactlisttemplateProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &outboundContactlisttemplateProxy{
		clientConfig:                               clientConfig,
		outboundApi:                                api,
		createOutboundContactlisttemplateAttr:      createOutboundContactlisttemplateFn,
		getAllOutboundContactlisttemplateAttr:      getAllOutboundContactlisttemplateFn,
		getOutboundContactlisttemplateIdByNameAttr: getOutboundContactlisttemplateIdByNameFn,
		getOutboundContactlisttemplateByIdAttr:     getOutboundContactlisttemplateByIdFn,
		updateOutboundContactlisttemplateAttr:      updateOutboundContactlisttemplateFn,
		deleteOutboundContactlisttemplateAttr:      deleteOutboundContactlisttemplateFn,
	}
}

// getOutboundContactlisttemplateProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getOutboundContactlisttemplateProxy(clientConfig *platformclientv2.Configuration) *outboundContactlisttemplateProxy {
	if internalProxy == nil {
		internalProxy = newOutboundContactlisttemplateProxy(clientConfig)
	}
	return internalProxy
}

// createOutboundContactlisttemplate creates a Genesys Cloud outbound Contactlisttemplate
func (p *outboundContactlisttemplateProxy) createOutboundContactlisttemplate(ctx context.Context, outboundContactlisttemplate *platformclientv2.Contactlisttemplate) (*platformclientv2.Contactlisttemplate, *platformclientv2.APIResponse, error) {
	return p.createOutboundContactlisttemplateAttr(ctx, p, outboundContactlisttemplate)
}

// getOutboundContactlisttemplate retrieves all Genesys Cloud outbound Contactlisttemplate
func (p *outboundContactlisttemplateProxy) getAllOutboundContactlisttemplate(ctx context.Context) (*[]platformclientv2.Contactlisttemplate, *platformclientv2.APIResponse, error) {
	return p.getAllOutboundContactlisttemplateAttr(ctx, p, "")
}

// getOutboundContactlisttemplateIdByName returns a single Genesys Cloud outbound Contactlisttemplate by a name
func (p *outboundContactlisttemplateProxy) getOutboundContactlisttemplateIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getOutboundContactlisttemplateIdByNameAttr(ctx, p, name)
}

// getOutboundContactlisttemplateById returns a single Genesys Cloud outbound Contactlisttemplate by Id
func (p *outboundContactlisttemplateProxy) getOutboundContactlisttemplateById(ctx context.Context, id string) (outboundContactlisttemplate *platformclientv2.Contactlisttemplate, response *platformclientv2.APIResponse, err error) {
	return p.getOutboundContactlisttemplateByIdAttr(ctx, p, id)
}

// updateOutboundContactlisttemplate updates a Genesys Cloud outbound Contactlisttemplate
func (p *outboundContactlisttemplateProxy) updateOutboundContactlisttemplate(ctx context.Context, id string, outboundContactlisttemplate *platformclientv2.Contactlisttemplate) (*platformclientv2.Contactlisttemplate, *platformclientv2.APIResponse, error) {
	return p.updateOutboundContactlisttemplateAttr(ctx, p, id, outboundContactlisttemplate)
}

// deleteOutboundContactlisttemplate deletes a Genesys Cloud outbound Contactlisttemplate by Id
func (p *outboundContactlisttemplateProxy) deleteOutboundContactlisttemplate(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteOutboundContactlisttemplateAttr(ctx, p, id)
}

// createOutboundContactlisttemplateFn is an implementation function for creating a Genesys Cloud outbound Contactlisttemplate
func createOutboundContactlisttemplateFn(ctx context.Context, p *outboundContactlisttemplateProxy, outboundContactlisttemplate *platformclientv2.Contactlisttemplate) (*platformclientv2.Contactlisttemplate, *platformclientv2.APIResponse, error) {
	Contactlisttemplate, resp, err := p.outboundApi.PostOutboundContactlisttemplates(*outboundContactlisttemplate)
	if err != nil {
		return nil, resp, err
	}
	return Contactlisttemplate, resp, nil
}

// getAllOutboundContactlisttemplateFn is the implementation for retrieving all outbound Contactlisttemplate in Genesys Cloud
func getAllOutboundContactlisttemplateFn(ctx context.Context, p *outboundContactlisttemplateProxy, name string) (*[]platformclientv2.Contactlisttemplate, *platformclientv2.APIResponse, error) {
	var allContactlisttemplates []platformclientv2.Contactlisttemplate
	const pageSize = 100

	Contactlisttemplates, resp, err := p.outboundApi.GetOutboundContactlisttemplates(pageSize, 1, true, "", name, "", "")
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get page of contact list template: %v", err)
	}

	if Contactlisttemplates.Entities == nil || len(*Contactlisttemplates.Entities) == 0 {
		return &allContactlisttemplates, resp, nil
	}

	for _, Contactlisttemplate := range *Contactlisttemplates.Entities {
		allContactlisttemplates = append(allContactlisttemplates, Contactlisttemplate)
	}

	for pageNum := 2; pageNum <= *Contactlisttemplates.PageCount; pageNum++ {
		Contactlisttemplates, resp, err := p.outboundApi.GetOutboundContactlisttemplates(pageSize, pageNum, true, "", name, "", "")
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get page of contact list template: %v", err)
		}

		if Contactlisttemplates.Entities == nil || len(*Contactlisttemplates.Entities) == 0 {
			break
		}

		for _, Contactlisttemplate := range *Contactlisttemplates.Entities {
			allContactlisttemplates = append(allContactlisttemplates, Contactlisttemplate)
		}
	}

	return &allContactlisttemplates, resp, nil
}

// getOutboundContactlisttemplateIdByNameFn is an implementation of the function to get a Genesys Cloud outbound Contactlisttemplate by name
func getOutboundContactlisttemplateIdByNameFn(ctx context.Context, p *outboundContactlisttemplateProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	Contactlisttemplates, resp, err := getAllOutboundContactlisttemplateFn(ctx, p, name)
	if err != nil {
		return "", false, resp, fmt.Errorf("error searching outbound contact list template  %s: %s", name, err)
	}

	var list platformclientv2.Contactlisttemplate
	for _, Contactlisttemplate := range *Contactlisttemplates {
		if *Contactlisttemplate.Name == name {
			log.Printf("Retrieved the contact list template id %s by name %s", *Contactlisttemplate.Id, name)
			list = Contactlisttemplate
			return *list.Id, false, resp, nil
		}
	}

	return "", true, resp, nil
}

// getOutboundContactlisttemplateByIdFn is an implementation of the function to get a Genesys Cloud outbound Contactlisttemplate by Id
func getOutboundContactlisttemplateByIdFn(ctx context.Context, p *outboundContactlisttemplateProxy, id string) (outboundContactlisttemplate *platformclientv2.Contactlisttemplate, response *platformclientv2.APIResponse, err error) {
	Contactlisttemplate, resp, err := p.outboundApi.GetOutboundContactlisttemplate(id)
	if err != nil {
		return nil, resp, err
	}
	return Contactlisttemplate, resp, nil
}

// updateOutboundContactlisttemplateFn is an implementation of the function to update a Genesys Cloud outbound Contactlisttemplate
func updateOutboundContactlisttemplateFn(ctx context.Context, p *outboundContactlisttemplateProxy, id string, outboundContactlisttemplate *platformclientv2.Contactlisttemplate) (*platformclientv2.Contactlisttemplate, *platformclientv2.APIResponse, error) {
	Contactlisttemplate, resp, err := p.outboundApi.GetOutboundContactlisttemplate(id)
	if err != nil {
		return nil, resp, err
	}

	outboundContactlisttemplate.Version = Contactlisttemplate.Version
	outboundContactlisttemplate, resp, updateErr := p.outboundApi.PutOutboundContactlisttemplate(id, *outboundContactlisttemplate)
	if updateErr != nil {
		return nil, resp, updateErr
	}
	return outboundContactlisttemplate, resp, nil
}

// deleteOutboundContactlisttemplateFn is an implementation function for deleting a Genesys Cloud outbound Contactlisttemplate
func deleteOutboundContactlisttemplateFn(ctx context.Context, p *outboundContactlisttemplateProxy, id string) (response *platformclientv2.APIResponse, err error) {
	return p.outboundApi.DeleteOutboundContactlisttemplate(id)
}
