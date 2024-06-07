package outbound_contact_list_contact

import (
	"context"
	"github.com/mypurecloud/platform-client-sdk-go/v130/platformclientv2"
)

var internalProxy *contactProxy

type createContactFunc func(ctx context.Context, p *contactProxy, contactListId string, contact platformclientv2.Writabledialercontact, priority, clearSystemData, doNotQueue bool) ([]platformclientv2.Dialercontact, *platformclientv2.APIResponse, error)
type readContactByIdFunc func(ctx context.Context, p *contactProxy, contactListId, contactId string) (*platformclientv2.Dialercontact, *platformclientv2.APIResponse, error)
type updateContactFunc func(ctx context.Context, p *contactProxy, contactListId string, contactId string, contact platformclientv2.Dialercontact) (*platformclientv2.Dialercontact, *platformclientv2.APIResponse, error)
type deleteContactFunc func(ctx context.Context, p *contactProxy, contactListId, contactId string) (*platformclientv2.APIResponse, error)

type contactProxy struct {
	clientConfig        *platformclientv2.Configuration
	outboundApi         *platformclientv2.OutboundApi
	createContactAttr   createContactFunc
	readContactByIdAttr readContactByIdFunc
	updateContactAttr   updateContactFunc
	deleteContactAttr   deleteContactFunc
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
	}
}

func getContactProxy(clientConfig *platformclientv2.Configuration) *contactProxy {
	if internalProxy == nil {
		internalProxy = newContactProxy(clientConfig)
	}

	return internalProxy
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

func createContactFn(_ context.Context, p *contactProxy, contactListId string, contact platformclientv2.Writabledialercontact, priority, clearSystemData, doNotQueue bool) ([]platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	return p.outboundApi.PostOutboundContactlistContacts(contactListId, []platformclientv2.Writabledialercontact{contact}, priority, clearSystemData, doNotQueue)
}

func readContactByIdFn(_ context.Context, p *contactProxy, contactListId, contactId string) (*platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	return p.outboundApi.GetOutboundContactlistContact(contactListId, contactId)
}

func updateContactFn(_ context.Context, p *contactProxy, contactListId, contactId string, contact platformclientv2.Dialercontact) (*platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	return p.outboundApi.PutOutboundContactlistContact(contactListId, contactId, contact)
}

func deleteContactFn(_ context.Context, p *contactProxy, contactListId, contactId string) (*platformclientv2.APIResponse, error) {
	return p.outboundApi.DeleteOutboundContactlistContact(contactListId, contactId)
}
