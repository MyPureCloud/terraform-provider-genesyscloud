package outbound_contactlistfilter

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
	"log"
)

/*
The genesyscloud_outbound_contactlistfilter_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *outboundContactlistfilterProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createOutboundContactlistfilterFunc func(ctx context.Context, p *outboundContactlistfilterProxy, contactListFilter *platformclientv2.Contactlistfilter) (*platformclientv2.Contactlistfilter, error)
type getAllOutboundContactlistfilterFunc func(ctx context.Context, p *outboundContactlistfilterProxy, name string) (*[]platformclientv2.Contactlistfilter, error)
type getOutboundContactlistfilterIdByNameFunc func(ctx context.Context, p *outboundContactlistfilterProxy, name string) (id string, retryable bool, err error)
type getOutboundContactlistfilterByIdFunc func(ctx context.Context, p *outboundContactlistfilterProxy, id string) (contactListFilter *platformclientv2.Contactlistfilter, responseCode int, err error)
type updateOutboundContactlistfilterFunc func(ctx context.Context, p *outboundContactlistfilterProxy, id string, contactListFilter *platformclientv2.Contactlistfilter) (*platformclientv2.Contactlistfilter, error)
type deleteOutboundContactlistfilterFunc func(ctx context.Context, p *outboundContactlistfilterProxy, id string) (response *platformclientv2.APIResponse, err error)

// outboundContactlistfilterProxy contains all of the methods that call genesys cloud APIs.
type outboundContactlistfilterProxy struct {
	clientConfig                             *platformclientv2.Configuration
	outboundApi                              *platformclientv2.OutboundApi
	createOutboundContactlistfilterAttr      createOutboundContactlistfilterFunc
	getAllOutboundContactlistfilterAttr      getAllOutboundContactlistfilterFunc
	getOutboundContactlistfilterIdByNameAttr getOutboundContactlistfilterIdByNameFunc
	getOutboundContactlistfilterByIdAttr     getOutboundContactlistfilterByIdFunc
	updateOutboundContactlistfilterAttr      updateOutboundContactlistfilterFunc
	deleteOutboundContactlistfilterAttr      deleteOutboundContactlistfilterFunc
}

// newOutboundContactlistfilterProxy initializes the outbound contactlistfilter proxy with all of the data needed to communicate with Genesys Cloud
func newOutboundContactlistfilterProxy(clientConfig *platformclientv2.Configuration) *outboundContactlistfilterProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &outboundContactlistfilterProxy{
		clientConfig:                             clientConfig,
		outboundApi:                              api,
		createOutboundContactlistfilterAttr:      createOutboundContactlistfilterFn,
		getAllOutboundContactlistfilterAttr:      getAllOutboundContactlistfilterFn,
		getOutboundContactlistfilterIdByNameAttr: getOutboundContactlistfilterIdByNameFn,
		getOutboundContactlistfilterByIdAttr:     getOutboundContactlistfilterByIdFn,
		updateOutboundContactlistfilterAttr:      updateOutboundContactlistfilterFn,
		deleteOutboundContactlistfilterAttr:      deleteOutboundContactlistfilterFn,
	}
}

// getOutboundContactlistfilterProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getOutboundContactlistfilterProxy(clientConfig *platformclientv2.Configuration) *outboundContactlistfilterProxy {
	if internalProxy == nil {
		internalProxy = newOutboundContactlistfilterProxy(clientConfig)
	}

	return internalProxy
}

// createOutboundContactlistfilter creates a Genesys Cloud outbound contactlistfilter
func (p *outboundContactlistfilterProxy) createOutboundContactlistfilter(ctx context.Context, outboundContactlistfilter *platformclientv2.Contactlistfilter) (*platformclientv2.Contactlistfilter, error) {
	return p.createOutboundContactlistfilterAttr(ctx, p, outboundContactlistfilter)
}

// getOutboundContactlistfilter retrieves all Genesys Cloud outbound contactlistfilter
func (p *outboundContactlistfilterProxy) getAllOutboundContactlistfilter(ctx context.Context) (*[]platformclientv2.Contactlistfilter, error) {
	return p.getAllOutboundContactlistfilterAttr(ctx, p, "")
}

// getOutboundContactlistfilterIdByName returns a single Genesys Cloud outbound contactlistfilter by a name
func (p *outboundContactlistfilterProxy) getOutboundContactlistfilterIdByName(ctx context.Context, name string) (id string, retryable bool, err error) {
	return p.getOutboundContactlistfilterIdByNameAttr(ctx, p, name)
}

// getOutboundContactlistfilterById returns a single Genesys Cloud outbound contactlistfilter by Id
func (p *outboundContactlistfilterProxy) getOutboundContactlistfilterById(ctx context.Context, id string) (outboundContactlistfilter *platformclientv2.Contactlistfilter, statusCode int, err error) {
	return p.getOutboundContactlistfilterByIdAttr(ctx, p, id)
}

// updateOutboundContactlistfilter updates a Genesys Cloud outbound contactlistfilter
func (p *outboundContactlistfilterProxy) updateOutboundContactlistfilter(ctx context.Context, id string, outboundContactlistfilter *platformclientv2.Contactlistfilter) (*platformclientv2.Contactlistfilter, error) {
	return p.updateOutboundContactlistfilterAttr(ctx, p, id, outboundContactlistfilter)
}

// deleteOutboundContactlistfilter deletes a Genesys Cloud outbound contactlistfilter by Id
func (p *outboundContactlistfilterProxy) deleteOutboundContactlistfilter(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteOutboundContactlistfilterAttr(ctx, p, id)
}

// createOutboundContactlistfilterFn is an implementation function for creating a Genesys Cloud outbound contactlistfilter
func createOutboundContactlistfilterFn(ctx context.Context, p *outboundContactlistfilterProxy, outboundContactlistfilter *platformclientv2.Contactlistfilter) (*platformclientv2.Contactlistfilter, error) {
	contactListFilter, _, err := p.outboundApi.PostOutboundContactlistfilters(*outboundContactlistfilter)
	if err != nil {
		return nil, err
	}

	return contactListFilter, nil
}

// getAllOutboundContactlistfilterFn is the implementation for retrieving all outbound contactlistfilter in Genesys Cloud
func getAllOutboundContactlistfilterFn(ctx context.Context, p *outboundContactlistfilterProxy, name string) (*[]platformclientv2.Contactlistfilter, error) {
	var allContactlistfilters []platformclientv2.Contactlistfilter
	const pageSize = 100

	contactListFilters, _, err := p.outboundApi.GetOutboundContactlistfilters(pageSize, 1, true, "", name, "", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get page of contact list filter: %v", err)
	}

	if contactListFilters.Entities == nil || len(*contactListFilters.Entities) == 0 {
		return &allContactlistfilters, nil
	}

	for _, contactListFilter := range *contactListFilters.Entities {
		allContactlistfilters = append(allContactlistfilters, contactListFilter)
	}

	for pageNum := 2; pageNum <= *contactListFilters.PageCount; pageNum++ {
		contactListFilters, _, err := p.outboundApi.GetOutboundContactlistfilters(pageSize, pageNum, true, "", name, "", "", "")
		if err != nil {
			return nil, fmt.Errorf("failed to get page of contact list filter: %v", err)
		}

		if contactListFilters.Entities == nil || len(*contactListFilters.Entities) == 0 {
			break
		}

		for _, contactListFilter := range *contactListFilters.Entities {
			allContactlistfilters = append(allContactlistfilters, contactListFilter)
		}
	}

	return &allContactlistfilters, nil
}

// getOutboundContactlistfilterIdByNameFn is an implementation of the function to get a Genesys Cloud outbound contactlistfilter by name
func getOutboundContactlistfilterIdByNameFn(ctx context.Context, p *outboundContactlistfilterProxy, name string) (id string, retryable bool, err error) {
	contactListFilters, err := getAllOutboundContactlistfilterFn(ctx, p, name)
	if err != nil {
		return "", false, fmt.Errorf("error searching outbound contact list filter %s: %s", name, err)
	}

	var filter platformclientv2.Contactlistfilter
	for _, contactListFilter := range *contactListFilters {
		if *contactListFilter.Name == name {
			log.Printf("Retrieved the contact list filter id %s by name %s", *contactListFilter.Id, name)
			filter = contactListFilter
			return *filter.Id, false, nil
		}
	}

	return "", true, nil
}

// getOutboundContactlistfilterByIdFn is an implementation of the function to get a Genesys Cloud outbound contactlistfilter by Id
func getOutboundContactlistfilterByIdFn(ctx context.Context, p *outboundContactlistfilterProxy, id string) (outboundContactlistfilter *platformclientv2.Contactlistfilter, statusCode int, err error) {
	contactListFilter, resp, err := p.outboundApi.GetOutboundContactlistfilter(id)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return contactListFilter, resp.StatusCode, nil
}

// updateOutboundContactlistfilterFn is an implementation of the function to update a Genesys Cloud outbound contactlistfilter
func updateOutboundContactlistfilterFn(ctx context.Context, p *outboundContactlistfilterProxy, id string, outboundContactlistfilter *platformclientv2.Contactlistfilter) (*platformclientv2.Contactlistfilter, error) {
	contactListFilter, _, err := p.outboundApi.GetOutboundContactlistfilter(id)
	if err != nil {
		return nil, err
	}

	outboundContactlistfilter.Version = contactListFilter.Version
	outboundContactlistfilter, _, updateErr := p.outboundApi.PutOutboundContactlistfilter(id, *outboundContactlistfilter)
	if updateErr != nil {
		return nil, updateErr
	}

	return outboundContactlistfilter, nil
}

// deleteOutboundContactlistfilterFn is an implementation function for deleting a Genesys Cloud outbound contactlistfilter
func deleteOutboundContactlistfilterFn(ctx context.Context, p *outboundContactlistfilterProxy, id string) (response *platformclientv2.APIResponse, err error) {
	return p.outboundApi.DeleteOutboundContactlistfilter(id)
}
