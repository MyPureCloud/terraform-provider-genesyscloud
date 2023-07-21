package external_contacts

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

var externalContactsContactsProxy *ExternalContactsContactsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllExternalContactsFunc func(ctx context.Context, p *ExternalContactsContactsProxy) (*[]platformclientv2.Externalcontact, error)
type createExternalContactFunc func(ctx context.Context, p *ExternalContactsContactsProxy, externalContact *platformclientv2.Externalcontact) (*platformclientv2.Externalcontact, error)
type deleteExternalContactFunc func(ctx context.Context, p *ExternalContactsContactsProxy, externalContactId string) (responseCode int, err error)
type getExternalContactByIdFunc func(ctx context.Context, p *ExternalContactsContactsProxy, externalContactId string) (externalContact *platformclientv2.Externalcontact, responseCode int, err error)
type getExternalContactIdBySearchFunc func(ctx context.Context, p *ExternalContactsContactsProxy, search string) (externalContactId string, retryable bool, err error)
type updateExternalContactFunc func(ctx context.Context, p *ExternalContactsContactsProxy, externalContactId string, externalContact *platformclientv2.Externalcontact) (*platformclientv2.Externalcontact, error)

// ExternalContactsContactsProxy contains all of the methods that call genesys cloud APIs.
type ExternalContactsContactsProxy struct {
	clientConfig                 *platformclientv2.Configuration
	externalContactsApi          *platformclientv2.ExternalContactsApi
	getAllExternalContacts       getAllExternalContactsFunc
	createExternalContact        createExternalContactFunc
	deleteExternalContactById    deleteExternalContactFunc
	getExternalContactById       getExternalContactByIdFunc
	getExternalContactIdBySearch getExternalContactIdBySearchFunc
	updateExternalContact        updateExternalContactFunc
}

// newScriptsProxy initializes the Scripts proxy with all of the data needed to communicate with Genesys Cloud
func newExternalContactsContactsProxy(clientConfig *platformclientv2.Configuration) *ExternalContactsContactsProxy {
	api := platformclientv2.NewExternalContactsApiWithConfig(clientConfig)
	return &ExternalContactsContactsProxy{
		clientConfig:                 clientConfig,
		externalContactsApi:          api,
		getAllExternalContacts:       getAllExternalContactsFn,
		createExternalContact:        createExternalContactFn,
		getExternalContactById:       getExternalContactByIdFn,
		deleteExternalContactById:    deleteExternalContactsFn,
		getExternalContactIdBySearch: getExternalContactIdBySearchFn,
		updateExternalContact:        updateExternalContactFn,
	}
}

// GetExternalContactsContactsProxy acts as a singleton to externalContactsContactsProxy.  It also ensures
// that we can still proxy our tests by directly setting externalContactsContactsProxy package variable
func GetExternalContactsContactsProxy(clientConfig *platformclientv2.Configuration) *ExternalContactsContactsProxy {
	if externalContactsContactsProxy == nil {
		externalContactsContactsProxy = newExternalContactsContactsProxy(clientConfig)
	}

	return externalContactsContactsProxy
}

// GetAllExternalContacts retrieves all Genesys Cloud External Contacts
func (p *ExternalContactsContactsProxy) GetAllExternalContacts(ctx context.Context) (*[]platformclientv2.Externalcontact, error) {
	return p.getAllExternalContacts(ctx, p)
}

// CreateExternalContact creates a Genesys Cloud External Contact
func (p *ExternalContactsContactsProxy) CreateExternalContact(ctx context.Context, externalContact *platformclientv2.Externalcontact) (*platformclientv2.Externalcontact, error) {
	return p.createExternalContact(ctx, p, externalContact)
}

// DeleteExternalContact deletes a Genesys Cloud External Contact by Id
func (p *ExternalContactsContactsProxy) DeleteExternalContactId(ctx context.Context, externalContactId string) (int, error) {
	return p.deleteExternalContactById(ctx, p, externalContactId)
}

// GetExternalContactById returns a single Genesys Cloud External Contact by Id
func (p *ExternalContactsContactsProxy) GetExternalContactById(ctx context.Context, externalContactId string) (*platformclientv2.Externalcontact, int, error) {
	return p.getExternalContactById(ctx, p, externalContactId)
}

// GetExternalContactIdBySearch returns a single Genesys Cloud External Contact by a search term
func (p *ExternalContactsContactsProxy) GetExternalContactIdBySearch(ctx context.Context, search string) (externalContactId string, retryable bool, err error) {
	return p.getExternalContactIdBySearch(ctx, p, search)
}

// UpdateExternalContact updates a Genesys Cloud External Contact
func (p *ExternalContactsContactsProxy) UpdateExternalContact(ctx context.Context, externalContactId string, externalContact *platformclientv2.Externalcontact) (*platformclientv2.Externalcontact, error) {
	return p.updateExternalContact(ctx, p, externalContactId, externalContact)
}

// getAllExternalContactsFn is the implementation for retrieving all external contacts in Genesys Cloud
func getAllExternalContactsFn(ctx context.Context, p *ExternalContactsContactsProxy) (*[]platformclientv2.Externalcontact, error) {
	var allExternalContacts []platformclientv2.Externalcontact

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100

		// The /api/v2/externalcontacts/contacts endpoint can only retrieve 10K records total.
		// We put a constraint in make sure we never pull then 10,000 records.
		if pageNum > 10 {
			fmt.Printf("*******************************************************\n")
			fmt.Printf("*              Warning                                *\n")
			fmt.Printf("*******************************************************\n")
			fmt.Printf("*                                                     *\n")
			fmt.Printf("* The External Contacts API can only retrieve 10,000  *\n")
			fmt.Printf("* records. Capping the number of External Contacts    *\n")
			fmt.Printf("* exported to 10,000.                                 *\n")
			fmt.Printf("*                                                     *\n")
			fmt.Printf("*******************************************************\n")
			return &allExternalContacts, nil
		}

		externalContacts, _, err := p.externalContactsApi.GetExternalcontactsContacts(pageSize, pageNum, "", "", nil)
		if err != nil {
			return nil, fmt.Errorf("Failed to get external contacts: %v", err)
		}

		if externalContacts.Entities == nil || len(*externalContacts.Entities) == 0 {
			break
		}

		for _, externalContact := range *externalContacts.Entities {
			log.Printf("Dealing with external contact id : %s", *externalContact.Id)
			allExternalContacts = append(allExternalContacts, externalContact)
		}
	}

	return &allExternalContacts, nil
}

// createExternalContactFn Is an implementation function for creating a Genesys Cloud External Contact
func createExternalContactFn(ctx context.Context, p *ExternalContactsContactsProxy, externalContact *platformclientv2.Externalcontact) (*platformclientv2.Externalcontact, error) {
	contact, _, err := p.externalContactsApi.PostExternalcontactsContacts(*externalContact)
	if err != nil {
		return nil, fmt.Errorf("Failed to create external contact: %s", err)
	}

	return contact, nil
}

// deleteExternalContactsFn Is an implementation function for deleting a Gensys Cloud External Contact
func deleteExternalContactsFn(ctx context.Context, p *ExternalContactsContactsProxy, externalContactId string) (int, error) {
	_, resp, err := p.externalContactsApi.DeleteExternalcontactsContact(externalContactId)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Failed to delete external contact: %s", err)
	}

	return resp.StatusCode, nil
}

// getExternalContactByIdFn is an implementation of the function to get a Genesys Cloud External Contact by Id
func getExternalContactByIdFn(ctx context.Context, p *ExternalContactsContactsProxy, externalContactId string) (*platformclientv2.Externalcontact, int, error) {
	externalContact, resp, err := p.externalContactsApi.GetExternalcontactsContact(externalContactId, nil)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to retrieve external contact by id %s: %s", externalContactId, err)
	}

	return externalContact, 0, nil
}

// getExternalContactIdBySearchFn is an implementation of the function to get a Genesys Cloud External contact by a search team
func getExternalContactIdBySearchFn(ctx context.Context, p *ExternalContactsContactsProxy, search string) (externalContactId string, retryable bool, err error) {
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
func updateExternalContactFn(ctx context.Context, p *ExternalContactsContactsProxy, externalContactId string, externalContact *platformclientv2.Externalcontact) (*platformclientv2.Externalcontact, error) {

	externalContact, _, err := p.externalContactsApi.PutExternalcontactsContact(externalContactId, *externalContact)
	if err != nil {
		return nil, fmt.Errorf("Failed to update external contact: %s", err)
	}
	return externalContact, nil
}
