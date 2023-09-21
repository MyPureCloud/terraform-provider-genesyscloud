package external_contacts

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

/*
The genesyscloud_externalcontacts_contact_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.

Each proxy implementation:

1.  Should provide a private package level variable that holds a instance of a proxy class.
2.  A New... constructor function  to initialize the proxy object.  This constructor should only be used within
    the proxy.
3.  A get private constructor function that the classes in the package can be used to to retrieve
    the proxy.  This proxy should check to see if the package level proxy instance is nil and
    should initialize it, otherwise it should return the instance
4.  Type definitions for each function that will be used in the proxy.  We use composition here
    so that we can easily provide mocks for testing.
5.  A struct for the proxy that holds an attribute for each function type.
6.  Wrapper methods on each of the elements on the struct.
7.  Function implementations for each function type definition.

*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *externalContactsContactsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllExternalContactsFunc func(ctx context.Context, p *externalContactsContactsProxy) (*[]platformclientv2.Externalcontact, error)
type createExternalContactFunc func(ctx context.Context, p *externalContactsContactsProxy, externalContact *platformclientv2.Externalcontact) (*platformclientv2.Externalcontact, error)
type deleteExternalContactFunc func(ctx context.Context, p *externalContactsContactsProxy, externalContactId string) (responseCode int, err error)
type getExternalContactByIdFunc func(ctx context.Context, p *externalContactsContactsProxy, externalContactId string) (externalContact *platformclientv2.Externalcontact, responseCode int, err error)
type getExternalContactIdBySearchFunc func(ctx context.Context, p *externalContactsContactsProxy, search string) (externalContactId string, retryable bool, err error)
type updateExternalContactFunc func(ctx context.Context, p *externalContactsContactsProxy, externalContactId string, externalContact *platformclientv2.Externalcontact) (*platformclientv2.Externalcontact, error)

// externalContactsContactsProxy contains all of the methods that call genesys cloud APIs.
type externalContactsContactsProxy struct {
	clientConfig                     *platformclientv2.Configuration
	externalContactsApi              *platformclientv2.ExternalContactsApi
	getAllExternalContactsAttr       getAllExternalContactsFunc
	createExternalContactAttr        createExternalContactFunc
	deleteExternalContactByIdAttr    deleteExternalContactFunc
	getExternalContactByIdAttr       getExternalContactByIdFunc
	getExternalContactIdBySearchAttr getExternalContactIdBySearchFunc
	updateExternalContactAttr        updateExternalContactFunc
}

// newExternalContactsContactsProxy initializes the External Contacts proxy with all of the data needed to communicate with Genesys Cloud
func newExternalContactsContactsProxy(clientConfig *platformclientv2.Configuration) *externalContactsContactsProxy {
	api := platformclientv2.NewExternalContactsApiWithConfig(clientConfig)
	return &externalContactsContactsProxy{
		clientConfig:                     clientConfig,
		externalContactsApi:              api,
		getAllExternalContactsAttr:       getAllExternalContactsFn,
		createExternalContactAttr:        createExternalContactFn,
		getExternalContactByIdAttr:       getExternalContactByIdFn,
		deleteExternalContactByIdAttr:    deleteExternalContactsFn,
		getExternalContactIdBySearchAttr: getExternalContactIdBySearchFn,
		updateExternalContactAttr:        updateExternalContactFn,
	}
}

// getExternalContactsContactsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getExternalContactsContactsProxy(clientConfig *platformclientv2.Configuration) *externalContactsContactsProxy {
	if internalProxy == nil {
		internalProxy = newExternalContactsContactsProxy(clientConfig)
	}

	return internalProxy
}

// getAllExternalContacts retrieves all Genesys Cloud External Contacts
func (p *externalContactsContactsProxy) getAllExternalContacts(ctx context.Context) (*[]platformclientv2.Externalcontact, error) {
	return p.getAllExternalContactsAttr(ctx, p)
}

// createExternalContact creates a Genesys Cloud External Contact
func (p *externalContactsContactsProxy) createExternalContact(ctx context.Context, externalContact *platformclientv2.Externalcontact) (*platformclientv2.Externalcontact, error) {
	return p.createExternalContactAttr(ctx, p, externalContact)
}

// DeleteExternalContact deletes a Genesys Cloud External Contact by Id
func (p *externalContactsContactsProxy) deleteExternalContactId(ctx context.Context, externalContactId string) (int, error) {
	return p.deleteExternalContactByIdAttr(ctx, p, externalContactId)
}

// getExternalContactById returns a single Genesys Cloud External Contact by Id
func (p *externalContactsContactsProxy) getExternalContactById(ctx context.Context, externalContactId string) (*platformclientv2.Externalcontact, int, error) {
	return p.getExternalContactByIdAttr(ctx, p, externalContactId)
}

// getExternalContactIdBySearch returns a single Genesys Cloud External Contact by a search term
func (p *externalContactsContactsProxy) getExternalContactIdBySearch(ctx context.Context, search string) (externalContactId string, retryable bool, err error) {
	return p.getExternalContactIdBySearchAttr(ctx, p, search)
}

// updateExternalContact updates a Genesys Cloud External Contact
func (p *externalContactsContactsProxy) updateExternalContact(ctx context.Context, externalContactId string, externalContact *platformclientv2.Externalcontact) (*platformclientv2.Externalcontact, error) {
	return p.updateExternalContactAttr(ctx, p, externalContactId, externalContact)
}

// getAllExternalContactsFn is the implementation for retrieving all external contacts in Genesys Cloud
func getAllExternalContactsFn(ctx context.Context, p *externalContactsContactsProxy) (*[]platformclientv2.Externalcontact, error) {
	var allExternalContacts []platformclientv2.Externalcontact

	cursor := ""
	for {
		externalContacts, _, err := p.externalContactsApi.GetExternalcontactsScanContacts(100, cursor)
		if err != nil {
			return nil, fmt.Errorf("Failed to get external contacts: %v", err)
		}

		if externalContacts.Entities == nil || len(*externalContacts.Entities) == 0 {
			break
		}

		for _, externalContact := range *externalContacts.Entities {
			allExternalContacts = append(allExternalContacts, externalContact)
		}

		if externalContacts.Cursors == nil || externalContacts.Cursors.After == nil {
			break
		}
		cursor = *externalContacts.Cursors.After
	}

	return &allExternalContacts, nil
}

// createExternalContactFn is an implementation function for creating a Genesys Cloud External Contact
func createExternalContactFn(ctx context.Context, p *externalContactsContactsProxy, externalContact *platformclientv2.Externalcontact) (*platformclientv2.Externalcontact, error) {
	contact, _, err := p.externalContactsApi.PostExternalcontactsContacts(*externalContact)
	if err != nil {
		return nil, fmt.Errorf("Failed to create external contact: %s", err)
	}

	return contact, nil
}

// deleteExternalContactsFn is an implementation function for deleting a Genesys Cloud External Contact
func deleteExternalContactsFn(ctx context.Context, p *externalContactsContactsProxy, externalContactId string) (int, error) {
	_, resp, err := p.externalContactsApi.DeleteExternalcontactsContact(externalContactId)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Failed to delete external contact: %s", err)
	}

	return resp.StatusCode, nil
}

// getExternalContactByIdFn is an implementation of the function to get a Genesys Cloud External Contact by Id
func getExternalContactByIdFn(ctx context.Context, p *externalContactsContactsProxy, externalContactId string) (*platformclientv2.Externalcontact, int, error) {
	externalContact, resp, err := p.externalContactsApi.GetExternalcontactsContact(externalContactId, nil)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to retrieve external contact by id %s: %s", externalContactId, err)
	}

	return externalContact, 0, nil
}

// getExternalContactIdBySearchFn is an implementation of the function to get a Genesys Cloud External contact by a search team
func getExternalContactIdBySearchFn(ctx context.Context, p *externalContactsContactsProxy, search string) (externalContactId string, retryable bool, err error) {
	const pageNum = 1
	const pageSize = 100
	contacts, _, err := p.externalContactsApi.GetExternalcontactsContacts(pageSize, pageNum, search, "", nil)
	if err != nil {
		return "", false, fmt.Errorf("Error searching external contact %s: %s", search, err)
	}

	if contacts.Entities == nil || len(*contacts.Entities) == 0 {
		return "", true, fmt.Errorf("No external contact found with search %s", search)
	}

	if len(*contacts.Entities) > 1 {
		return "", false, fmt.Errorf("Too many values returned in look for external contact.  Unable to choose 1 external contact.  Please refine search and continue.")
	}

	log.Printf("Retrieved the external contact search id %s by name %s", *(*contacts.Entities)[0].Id, search)
	contact := (*contacts.Entities)[0]
	return *contact.Id, false, nil
}

// updateExternalContactFn is an implementation of the function to update a Genesys Cloud external contact
func updateExternalContactFn(ctx context.Context, p *externalContactsContactsProxy, externalContactId string, externalContact *platformclientv2.Externalcontact) (*platformclientv2.Externalcontact, error) {
	externalContact, _, err := p.externalContactsApi.PutExternalcontactsContact(externalContactId, *externalContact)
	if err != nil {
		return nil, fmt.Errorf("Failed to update external contact: %s", err)
	}
	return externalContact, nil
}
