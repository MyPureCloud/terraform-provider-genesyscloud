package outbound_contact_list_contact

import (
	"context"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"terraform-provider-genesyscloud/genesyscloud/tfexporter_state"

	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

var contactCache = rc.NewResourceCache[platformclientv2.Dialercontact]()

type createContactFunc func(ctx context.Context, p *contactProxy, contactListId string, contact platformclientv2.Writabledialercontact, priority, clearSystemData, doNotQueue bool) ([]platformclientv2.Dialercontact, *platformclientv2.APIResponse, error)
type readContactByIdFunc func(ctx context.Context, p *contactProxy, contactListId, contactId string) (*platformclientv2.Dialercontact, *platformclientv2.APIResponse, error)
type updateContactFunc func(ctx context.Context, p *contactProxy, contactListId string, contactId string, contact platformclientv2.Dialercontact) (*platformclientv2.Dialercontact, *platformclientv2.APIResponse, error)
type deleteContactFunc func(ctx context.Context, p *contactProxy, contactListId, contactId string) (*platformclientv2.APIResponse, error)
type getAllContactsFunc func(ctx context.Context, p *contactProxy) ([]platformclientv2.Dialercontact, *platformclientv2.APIResponse, error)

type contactProxy struct {
	clientConfig        *platformclientv2.Configuration
	outboundApi         *platformclientv2.OutboundApi
	createContactAttr   createContactFunc
	readContactByIdAttr readContactByIdFunc
	updateContactAttr   updateContactFunc
	deleteContactAttr   deleteContactFunc
	getAllContactsAttr  getAllContactsFunc
	contactCache        rc.CacheInterface[platformclientv2.Dialercontact]
}

func newContactProxy(clientConfig *platformclientv2.Configuration) *contactProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &contactProxy{
		clientConfig:        clientConfig,
		outboundApi:         api,
		createContactAttr:   createContactFn,
		readContactByIdAttr: readContactByIdFn,
		updateContactAttr:   updateContactFn,
		deleteContactAttr:   deleteContactFn,
		getAllContactsAttr:  getAllContactsFn,
		contactCache:        contactCache,
	}
}

func getContactProxy(clientConfig *platformclientv2.Configuration) *contactProxy {
	return newContactProxy(clientConfig)
}

func (p *contactProxy) createContact(ctx context.Context, contactListId string, contact platformclientv2.Writabledialercontact, priority, clearSystemData, doNotQueue bool) ([]platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	return p.createContactAttr(ctx, p, contactListId, contact, priority, clearSystemData, doNotQueue)
}

func (p *contactProxy) readContactById(ctx context.Context, contactListId, contactId string) (*platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	return p.readContactByIdAttr(ctx, p, contactListId, contactId)
}

func (p *contactProxy) updateContact(ctx context.Context, contactListId, contactId string, contact platformclientv2.Dialercontact) (*platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	return p.updateContactAttr(ctx, p, contactListId, contactId, contact)
}

func (p *contactProxy) deleteContact(ctx context.Context, contactListId, contactId string) (*platformclientv2.APIResponse, error) {
	return p.deleteContactAttr(ctx, p, contactListId, contactId)
}

func (p *contactProxy) getAllContacts(ctx context.Context) ([]platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	return p.getAllContactsAttr(ctx, p)
}

func createContactFn(_ context.Context, p *contactProxy, contactListId string, contact platformclientv2.Writabledialercontact, priority, clearSystemData, doNotQueue bool) ([]platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	return p.outboundApi.PostOutboundContactlistContacts(contactListId, []platformclientv2.Writabledialercontact{contact}, priority, clearSystemData, doNotQueue)
}

func readContactByIdFn(_ context.Context, p *contactProxy, contactListId, contactId string) (*platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	if contact := rc.GetCacheItem(p.contactCache, createComplexContact(contactListId, contactId)); contact != nil {
		return contact, nil, nil
	}
	if tfexporter_state.IsExporterActive() {
		log.Printf("Could not read contact '%s' from cache (Contact list '%s'). Reading from the API...", contactId, contactListId)
	}
	return p.outboundApi.GetOutboundContactlistContact(contactListId, contactId)
}

func updateContactFn(_ context.Context, p *contactProxy, contactListId, contactId string, contact platformclientv2.Dialercontact) (*platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	return p.outboundApi.PutOutboundContactlistContact(contactListId, contactId, contact)
}

func deleteContactFn(_ context.Context, p *contactProxy, contactListId, contactId string) (*platformclientv2.APIResponse, error) {
	resp, err := p.outboundApi.DeleteOutboundContactlistContact(contactListId, contactId)
	if err != nil {
		return resp, err
	}
	rc.DeleteCacheItem(p.contactCache, createComplexContact(contactListId, contactId))
	return resp, nil
}

func getAllContactsFn(ctx context.Context, p *contactProxy) ([]platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	var allContacts []platformclientv2.Dialercontact
	contactMatrix := make(map[string][]platformclientv2.Dialercontact)

	contactListIds, resp, err := p.getAllContactListIds(ctx)
	if err != nil {
		return allContacts, resp, err
	}

	for _, contactListId := range contactListIds {
		contacts, resp, err := p.getContactsByContactListId(ctx, contactListId)
		if err != nil {
			return nil, resp, err
		}
		allContacts = append(allContacts, contacts...)
		contactMatrix[contactListId] = contacts
	}

	for contactListId, contactListContacts := range contactMatrix {
		for _, contact := range contactListContacts {
			rc.SetCache(p.contactCache, createComplexContact(contactListId, *contact.Id), contact)
		}
	}

	return allContacts, nil, nil
}

func (p *contactProxy) getContactsByContactListId(_ context.Context, contactListId string) ([]platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	var (
		pageNum     = 1
		pageSize    = 100
		allContacts []platformclientv2.Dialercontact
	)

	body := platformclientv2.Contactlistingrequest{
		PageNumber: &pageNum,
		PageSize:   &pageSize,
	}

	data, resp, err := p.outboundApi.PostOutboundContactlistContactsSearch(contactListId, body)
	if err != nil {
		return nil, resp, err
	}
	if data == nil || data.Entities == nil || len(*data.Entities) == 0 {
		return nil, nil, nil
	}
	allContacts = append(allContacts, *data.Entities...)

	if data.PageCount == nil {
		return allContacts, nil, nil
	}

	for pageNum = 2; pageNum <= *data.PageCount; pageNum++ {
		body.PageNumber = &pageNum
		data, resp, err = p.outboundApi.PostOutboundContactlistContactsSearch(contactListId, body)
		if err != nil {
			return nil, resp, err
		}
		if data == nil || data.Entities == nil || len(*data.Entities) == 0 {
			break
		}
		allContacts = append(allContacts, *data.Entities...)
	}

	return allContacts, nil, nil
}

func (p *contactProxy) getAllContactListIds(_ context.Context) ([]string, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	var pageNum = 1
	var allContactListIds []string

	contactListConfigs, resp, getErr := p.outboundApi.GetOutboundContactlists(false, false, pageSize, pageNum, true, "", "", []string{}, []string{}, "", "")
	if getErr != nil {
		return nil, resp, getErr
	}
	if contactListConfigs.Entities == nil || len(*contactListConfigs.Entities) == 0 {
		return nil, nil, nil
	}
	for _, cl := range *contactListConfigs.Entities {
		allContactListIds = append(allContactListIds, *cl.Id)
	}

	for pageNum := 2; pageNum <= *contactListConfigs.PageCount; pageNum++ {
		contactListConfigs, resp, getErr := p.outboundApi.GetOutboundContactlists(false, false, pageSize, pageNum, true, "", "", []string{}, []string{}, "", "")
		if getErr != nil {
			return nil, resp, getErr
		}
		if contactListConfigs.Entities == nil || len(*contactListConfigs.Entities) == 0 {
			break
		}
		for _, cl := range *contactListConfigs.Entities {
			allContactListIds = append(allContactListIds, *cl.Id)
		}
	}

	return allContactListIds, nil, nil
}
