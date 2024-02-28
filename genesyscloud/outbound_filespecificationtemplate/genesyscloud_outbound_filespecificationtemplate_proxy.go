package outbound_filespecificationtemplate

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
)

/*
The genesyscloud_outbound_filespecificationtemplate_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *outboundFilespecificationtemplateProxy

// Type definitions for each func on our proxy, so we can easily mock them out later
type createOutboundFilespecificationtemplateFunc func(ctx context.Context, p *outboundFilespecificationtemplateProxy, fileSpecificationTemplate *platformclientv2.Filespecificationtemplate) (*platformclientv2.Filespecificationtemplate, error)
type getAllOutboundFilespecificationtemplateFunc func(ctx context.Context, p *outboundFilespecificationtemplateProxy, name string) (*[]platformclientv2.Filespecificationtemplate, error)
type getOutboundFilespecificationtemplateIdByNameFunc func(ctx context.Context, p *outboundFilespecificationtemplateProxy, name string) (id string, retryable bool, err error)
type getOutboundFilespecificationtemplateByIdFunc func(ctx context.Context, p *outboundFilespecificationtemplateProxy, id string) (fileSpecificationTemplate *platformclientv2.Filespecificationtemplate, responseCode int, err error)
type updateOutboundFilespecificationtemplateFunc func(ctx context.Context, p *outboundFilespecificationtemplateProxy, id string, fileSpecificationTemplate *platformclientv2.Filespecificationtemplate) (*platformclientv2.Filespecificationtemplate, error)
type deleteOutboundFilespecificationtemplateFunc func(ctx context.Context, p *outboundFilespecificationtemplateProxy, id string) (response *platformclientv2.APIResponse, err error)

// outboundFilespecificationtemplateProxy contains all the methods that call genesys cloud APIs.
type outboundFilespecificationtemplateProxy struct {
	clientConfig                                     *platformclientv2.Configuration
	outboundApi                                      *platformclientv2.OutboundApi
	createOutboundFilespecificationtemplateAttr      createOutboundFilespecificationtemplateFunc
	getAllOutboundFilespecificationtemplateAttr      getAllOutboundFilespecificationtemplateFunc
	getOutboundFilespecificationtemplateIdByNameAttr getOutboundFilespecificationtemplateIdByNameFunc
	getOutboundFilespecificationtemplateByIdAttr     getOutboundFilespecificationtemplateByIdFunc
	updateOutboundFilespecificationtemplateAttr      updateOutboundFilespecificationtemplateFunc
	deleteOutboundFilespecificationtemplateAttr      deleteOutboundFilespecificationtemplateFunc
}

// newOutboundFilespecificationtemplateProxy initializes the outbound filespecificationtemplate proxy
// with all the data needed to communicate with Genesys Cloud
func newOutboundFilespecificationtemplateProxy(clientConfig *platformclientv2.Configuration) *outboundFilespecificationtemplateProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &outboundFilespecificationtemplateProxy{
		clientConfig: clientConfig,
		outboundApi:  api,
		createOutboundFilespecificationtemplateAttr:      createOutboundFilespecificationtemplateFn,
		getAllOutboundFilespecificationtemplateAttr:      getAllOutboundFilespecificationtemplateFn,
		getOutboundFilespecificationtemplateIdByNameAttr: getOutboundFilespecificationtemplateIdByNameFn,
		getOutboundFilespecificationtemplateByIdAttr:     getOutboundFilespecificationtemplateByIdFn,
		updateOutboundFilespecificationtemplateAttr:      updateOutboundFilespecificationtemplateFn,
		deleteOutboundFilespecificationtemplateAttr:      deleteOutboundFilespecificationtemplateFn,
	}
}

// getOutboundFilespecificationtemplateProxy acts as a singleton to for the internalProxy. It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getOutboundFilespecificationtemplateProxy(clientConfig *platformclientv2.Configuration) *outboundFilespecificationtemplateProxy {
	if internalProxy == nil {
		internalProxy = newOutboundFilespecificationtemplateProxy(clientConfig)
	}

	return internalProxy
}

// createOutboundFilespecificationtemplate creates a Genesys Cloud outbound filespecificationtemplate
func (p *outboundFilespecificationtemplateProxy) createOutboundFilespecificationtemplate(ctx context.Context, outboundFilespecificationtemplate *platformclientv2.Filespecificationtemplate) (*platformclientv2.Filespecificationtemplate, error) {
	return p.createOutboundFilespecificationtemplateAttr(ctx, p, outboundFilespecificationtemplate)
}

// getAllOutboundFilespecificationtemplate retrieves all Genesys Cloud outbound filespecificationtemplate
func (p *outboundFilespecificationtemplateProxy) getAllOutboundFilespecificationtemplate(ctx context.Context) (*[]platformclientv2.Filespecificationtemplate, error) {
	return p.getAllOutboundFilespecificationtemplateAttr(ctx, p, "")
}

// getOutboundFilespecificationtemplateIdByName returns a single Genesys Cloud outbound filespecificationtemplate by name
func (p *outboundFilespecificationtemplateProxy) getOutboundFilespecificationtemplateIdByName(ctx context.Context, name string) (id string, retryable bool, err error) {
	return p.getOutboundFilespecificationtemplateIdByNameAttr(ctx, p, name)
}

// getOutboundFilespecificationtemplateById returns a single Genesys Cloud outbound filespecificationtemplate by id
func (p *outboundFilespecificationtemplateProxy) getOutboundFilespecificationtemplateById(ctx context.Context, id string) (outboundFilespecificationtemplate *platformclientv2.Filespecificationtemplate, statusCode int, err error) {
	return p.getOutboundFilespecificationtemplateByIdAttr(ctx, p, id)
}

// updateOutboundFilespecificationtemplate updates a Genesys Cloud outbound filespecificationtemplate
func (p *outboundFilespecificationtemplateProxy) updateOutboundFilespecificationtemplate(ctx context.Context, id string, outboundFilespecificationtemplate *platformclientv2.Filespecificationtemplate) (*platformclientv2.Filespecificationtemplate, error) {
	return p.updateOutboundFilespecificationtemplateAttr(ctx, p, id, outboundFilespecificationtemplate)
}

// deleteOutboundFilespecificationtemplate deletes a Genesys Cloud outbound filespecificationtemplate by id
func (p *outboundFilespecificationtemplateProxy) deleteOutboundFilespecificationtemplate(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteOutboundFilespecificationtemplateAttr(ctx, p, id)
}

// createOutboundFilespecificationtemplateFn is an implementation function
// for creating a Genesys Cloud outbound filespecificationtemplate
func createOutboundFilespecificationtemplateFn(ctx context.Context, p *outboundFilespecificationtemplateProxy, outboundFilespecificationtemplate *platformclientv2.Filespecificationtemplate) (*platformclientv2.Filespecificationtemplate, error) {
	fst, _, err := p.outboundApi.PostOutboundFilespecificationtemplates(*outboundFilespecificationtemplate)
	if err != nil {
		return nil, fmt.Errorf("Failed to create file specification template %s", err)
	}

	return fst, nil
}

// getAllOutboundFilespecificationtemplateFn is the implementation for retrieving
// all outbound filespecificationtemplate in Genesys Cloud
func getAllOutboundFilespecificationtemplateFn(ctx context.Context, p *outboundFilespecificationtemplateProxy, name string) (*[]platformclientv2.Filespecificationtemplate, error) {
	var allFileSpecificationTemplates []platformclientv2.Filespecificationtemplate
	const pageSize = 100

	fileSpecificationTemplates, _, err := p.outboundApi.GetOutboundFilespecificationtemplates(
		pageSize, 1, true, "", name, "", "")
	if err != nil {
		return nil, fmt.Errorf("Failed to get file specification templates: %v", err)
	}

	if fileSpecificationTemplates.Entities == nil || len(*fileSpecificationTemplates.Entities) == 0 {
		return &allFileSpecificationTemplates, nil
	}

	for _, fileSpecificationTemplate := range *fileSpecificationTemplates.Entities {
		allFileSpecificationTemplates = append(allFileSpecificationTemplates, fileSpecificationTemplate)
	}

	for pageNum := 2; pageNum <= *fileSpecificationTemplates.PageCount; pageNum++ {
		fileSpecificationTemplates, _, err := p.outboundApi.GetOutboundFilespecificationtemplates(
			pageSize, pageNum, true, "", "", "", "")
		if err != nil {
			return nil, fmt.Errorf("Failed to get file specification templates: %v", err)
		}

		if fileSpecificationTemplates.Entities == nil || len(*fileSpecificationTemplates.Entities) == 0 {
			break
		}

		for _, fileSpecificationTemplate := range *fileSpecificationTemplates.Entities {
			allFileSpecificationTemplates = append(allFileSpecificationTemplates, fileSpecificationTemplate)
		}
	}

	return &allFileSpecificationTemplates, nil
}

// getOutboundFilespecificationtemplateIdByNameFn is an implementation of the function
// to get a Genesys Cloud outbound filespecificationtemplate by name
func getOutboundFilespecificationtemplateIdByNameFn(ctx context.Context, p *outboundFilespecificationtemplateProxy, name string) (id string, retryable bool, err error) {
	fileSpecificationTemplates, err := getAllOutboundFilespecificationtemplateFn(ctx, p, name)

	if err != nil {
		return "", false, err
	}

	if len(*fileSpecificationTemplates) == 0 {
		return "", true, fmt.Errorf("No file specification template found with name %s", name)
	}

	for _, fileSpecificationTemplate := range *fileSpecificationTemplates {
		if *fileSpecificationTemplate.Name == name {
			log.Printf("Retrieved the outbound file specification template id %s by name %s", *fileSpecificationTemplate.Id, name)
			return *fileSpecificationTemplate.Id, false, nil
		}
	}

	return "", true, fmt.Errorf("Unable to find outbound file specification template with name %s", name)
}

// getOutboundFilespecificationtemplateByIdFn is an implementation of the function
// to get a Genesys Cloud outbound filespecificationtemplate by id
func getOutboundFilespecificationtemplateByIdFn(ctx context.Context, p *outboundFilespecificationtemplateProxy, id string) (outboundFilespecificationtemplate *platformclientv2.Filespecificationtemplate, statusCode int, err error) {
	fst, resp, err := p.outboundApi.GetOutboundFilespecificationtemplate(id)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to retrieve file specification template by id %s: %s", id, err)
	}

	return fst, resp.StatusCode, nil
}

// updateOutboundFilespecificationtemplateFn is an implementation of the function
// to update a Genesys Cloud outbound filespecificationtemplate
func updateOutboundFilespecificationtemplateFn(ctx context.Context, p *outboundFilespecificationtemplateProxy, id string, outboundFilespecificationtemplate *platformclientv2.Filespecificationtemplate) (*platformclientv2.Filespecificationtemplate, error) {
	fst, _, err := getOutboundFilespecificationtemplateByIdFn(ctx, p, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to file specification template by id %s: %s", id, err)
	}

	outboundFilespecificationtemplate.Version = fst.Version
	fileSpecificationTemplate, _, err := p.outboundApi.PutOutboundFilespecificationtemplate(id, *outboundFilespecificationtemplate)
	if err != nil {
		return nil, fmt.Errorf("Failed to update file specification template: %s", err)
	}
	return fileSpecificationTemplate, nil
}

// deleteOutboundFilespecificationtemplateFn is an implementation function for
// deleting a Genesys Cloud outbound filespecificationtemplate
func deleteOutboundFilespecificationtemplateFn(ctx context.Context, p *outboundFilespecificationtemplateProxy, id string) (response *platformclientv2.APIResponse, err error) {
	return p.outboundApi.DeleteOutboundFilespecificationtemplate(id)
}
