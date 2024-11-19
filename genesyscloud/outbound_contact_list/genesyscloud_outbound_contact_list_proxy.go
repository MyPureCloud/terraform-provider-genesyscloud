package outbound_contact_list

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"terraform-provider-genesyscloud/genesyscloud/tfexporter_state"

	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The genesyscloud_outbound_contact_list_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

var contactListCache = rc.NewResourceCache[platformclientv2.Contactlist]()

// Type definitions for each func on our proxy so we can easily mock them out later
type createOutboundContactlistFunc func(ctx context.Context, p *outboundContactlistProxy, contactList *platformclientv2.Contactlist) (*platformclientv2.Contactlist, *platformclientv2.APIResponse, error)
type getAllOutboundContactlistFunc func(ctx context.Context, p *outboundContactlistProxy, name string) (*[]platformclientv2.Contactlist, *platformclientv2.APIResponse, error)
type getOutboundContactlistIdByNameFunc func(ctx context.Context, p *outboundContactlistProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getOutboundContactlistByIdFunc func(ctx context.Context, p *outboundContactlistProxy, id string) (contactList *platformclientv2.Contactlist, response *platformclientv2.APIResponse, err error)
type updateOutboundContactlistFunc func(ctx context.Context, p *outboundContactlistProxy, id string, contactList *platformclientv2.Contactlist) (*platformclientv2.Contactlist, *platformclientv2.APIResponse, error)
type deleteOutboundContactlistFunc func(ctx context.Context, p *outboundContactlistProxy, id string) (response *platformclientv2.APIResponse, err error)

// outboundContactlistProxy contains all of the methods that call genesys cloud APIs.
type outboundContactlistProxy struct {
	clientConfig                       *platformclientv2.Configuration
	outboundApi                        *platformclientv2.OutboundApi
	createOutboundContactlistAttr      createOutboundContactlistFunc
	getAllOutboundContactlistAttr      getAllOutboundContactlistFunc
	getOutboundContactlistIdByNameAttr getOutboundContactlistIdByNameFunc
	getOutboundContactlistByIdAttr     getOutboundContactlistByIdFunc
	updateOutboundContactlistAttr      updateOutboundContactlistFunc
	deleteOutboundContactlistAttr      deleteOutboundContactlistFunc
	contactListCache                   rc.CacheInterface[platformclientv2.Contactlist]
}

// newOutboundContactlistProxy initializes the outbound contactlist proxy with all of the data needed to communicate with Genesys Cloud
func newOutboundContactlistProxy(clientConfig *platformclientv2.Configuration) *outboundContactlistProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &outboundContactlistProxy{
		clientConfig:                       clientConfig,
		outboundApi:                        api,
		createOutboundContactlistAttr:      createOutboundContactlistFn,
		getAllOutboundContactlistAttr:      getAllOutboundContactlistFn,
		getOutboundContactlistIdByNameAttr: getOutboundContactlistIdByNameFn,
		getOutboundContactlistByIdAttr:     getOutboundContactlistByIdFn,
		updateOutboundContactlistAttr:      updateOutboundContactlistFn,
		deleteOutboundContactlistAttr:      deleteOutboundContactlistFn,
		contactListCache:                   contactListCache,
	}
}

func getOutboundContactlistProxy(clientConfig *platformclientv2.Configuration) *outboundContactlistProxy {
	return newOutboundContactlistProxy(clientConfig)
}

// createOutboundContactlist creates a Genesys Cloud outbound contactlist
func (p *outboundContactlistProxy) createOutboundContactlist(ctx context.Context, outboundContactlist *platformclientv2.Contactlist) (*platformclientv2.Contactlist, *platformclientv2.APIResponse, error) {
	return p.createOutboundContactlistAttr(ctx, p, outboundContactlist)
}

// getOutboundContactlist retrieves all Genesys Cloud outbound contactlist
func (p *outboundContactlistProxy) getAllOutboundContactlist(ctx context.Context) (*[]platformclientv2.Contactlist, *platformclientv2.APIResponse, error) {
	return p.getAllOutboundContactlistAttr(ctx, p, "")
}

// getOutboundContactlistIdByName returns a single Genesys Cloud outbound contactlist by a name
func (p *outboundContactlistProxy) getOutboundContactlistIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getOutboundContactlistIdByNameAttr(ctx, p, name)
}

// getOutboundContactlistById returns a single Genesys Cloud outbound contactlist by Id
func (p *outboundContactlistProxy) getOutboundContactlistById(ctx context.Context, id string) (outboundContactlist *platformclientv2.Contactlist, response *platformclientv2.APIResponse, err error) {
	return p.getOutboundContactlistByIdAttr(ctx, p, id)
}

// updateOutboundContactlist updates a Genesys Cloud outbound contactlist
func (p *outboundContactlistProxy) updateOutboundContactlist(ctx context.Context, id string, outboundContactlist *platformclientv2.Contactlist) (*platformclientv2.Contactlist, *platformclientv2.APIResponse, error) {
	return p.updateOutboundContactlistAttr(ctx, p, id, outboundContactlist)
}

// deleteOutboundContactlist deletes a Genesys Cloud outbound contactlist by Id
func (p *outboundContactlistProxy) deleteOutboundContactlist(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteOutboundContactlistAttr(ctx, p, id)
}

// createOutboundContactlistFn is an implementation function for creating a Genesys Cloud outbound contactlist
func createOutboundContactlistFn(ctx context.Context, p *outboundContactlistProxy, outboundContactlist *platformclientv2.Contactlist) (*platformclientv2.Contactlist, *platformclientv2.APIResponse, error) {
	return p.outboundApi.PostOutboundContactlists(*outboundContactlist)
}

// getAllOutboundContactlistFn is the implementation for retrieving all outbound contactlist in Genesys Cloud
func getAllOutboundContactlistFn(ctx context.Context, p *outboundContactlistProxy, name string) (*[]platformclientv2.Contactlist, *platformclientv2.APIResponse, error) {
	var allContactlists []platformclientv2.Contactlist
	const pageSize = 100

	contactLists, resp, err := p.outboundApi.GetOutboundContactlists(false, false, pageSize, 1, true, "", name, []string{}, []string{}, "", "")
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get page of contact list: %v", err)
	}

	if contactLists.Entities == nil || len(*contactLists.Entities) == 0 {
		return &allContactlists, resp, nil
	}

	allContactlists = append(allContactlists, *contactLists.Entities...)

	for pageNum := 2; pageNum <= *contactLists.PageCount; pageNum++ {
		contactLists, resp, err := p.outboundApi.GetOutboundContactlists(false, false, pageSize, pageNum, true, "", name, []string{}, []string{}, "", "")
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get page of contact list : %v", err)
		}

		if contactLists.Entities == nil || len(*contactLists.Entities) == 0 {
			break
		}

		allContactlists = append(allContactlists, *contactLists.Entities...)
	}

	for _, contactList := range allContactlists {
		rc.SetCache(p.contactListCache, *contactList.Id, contactList)
	}

	return &allContactlists, resp, nil
}

// getOutboundContactlistIdByNameFn is an implementation of the function to get a Genesys Cloud outbound contactlist by name
func getOutboundContactlistIdByNameFn(ctx context.Context, p *outboundContactlistProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	contactLists, resp, err := getAllOutboundContactlistFn(ctx, p, name)
	if err != nil {
		return "", false, resp, fmt.Errorf("error searching outbound contact list  %s: %s", name, err)
	}

	for _, contactList := range *contactLists {
		if *contactList.Name == name {
			log.Printf("Retrieved the contact list id %s by name %s", *contactList.Id, name)
			return *contactList.Id, false, resp, nil
		}
	}

	return "", true, resp, fmt.Errorf("no contact lists found with the name '%s'", name)
}

// getOutboundContactlistByIdFn is an implementation of the function to get a Genesys Cloud outbound contactlist by Id
func getOutboundContactlistByIdFn(ctx context.Context, p *outboundContactlistProxy, id string) (outboundContactlist *platformclientv2.Contactlist, response *platformclientv2.APIResponse, err error) {
	if contactList := rc.GetCacheItem(p.contactListCache, id); contactList != nil {
		return contactList, nil, nil
	}
	if tfexporter_state.IsExporterActive() {
		log.Printf("Could not read contact list '%s' from cache. Reading from the API...", id)
	}
	return p.outboundApi.GetOutboundContactlist(id, false, false)
}

// updateOutboundContactlistFn is an implementation of the function to update a Genesys Cloud outbound contactlist
func updateOutboundContactlistFn(ctx context.Context, p *outboundContactlistProxy, id string, outboundContactlist *platformclientv2.Contactlist) (*platformclientv2.Contactlist, *platformclientv2.APIResponse, error) {
	contactList, resp, err := p.outboundApi.GetOutboundContactlist(id, false, false)
	if err != nil {
		return nil, resp, err
	}

	outboundContactlist.Version = contactList.Version
	return p.outboundApi.PutOutboundContactlist(id, *outboundContactlist)
}

// deleteOutboundContactlistFn is an implementation function for deleting a Genesys Cloud outbound contactlist
func deleteOutboundContactlistFn(ctx context.Context, p *outboundContactlistProxy, id string) (response *platformclientv2.APIResponse, err error) {
	resp, err := p.outboundApi.DeleteOutboundContactlist(id)
	if err != nil {
		return resp, err
	}
	rc.DeleteCacheItem(p.contactListCache, id)
	return resp, nil
}
