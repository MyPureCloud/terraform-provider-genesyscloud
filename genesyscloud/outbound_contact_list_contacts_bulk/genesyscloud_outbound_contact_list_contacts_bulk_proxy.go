package outbound_contact_list_contacts_bulk

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
	"terraform-provider-genesyscloud/genesyscloud/util/files"

	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

var contactCache = rc.NewResourceCache[platformclientv2.Dialercontact]()

type ContactEntry struct {
	ContactList *platformclientv2.Contactlist
	Contact     *[]platformclientv2.Dialercontact
}
type createBulkContactFunc func(ctx context.Context, p *contactProxy, contactListId string, contact platformclientv2.Writabledialercontact, priority, clearSystemData, doNotQueue bool) ([]platformclientv2.Dialercontact, *platformclientv2.APIResponse, error)
type readBulkContactByIdFunc func(ctx context.Context, p *contactProxy, contactListId, contactId string) (*platformclientv2.Dialercontact, *platformclientv2.APIResponse, error)
type updateBulkContactFunc func(ctx context.Context, p *contactProxy, contactListId string, contactId string, contact platformclientv2.Dialercontact) (*platformclientv2.Dialercontact, *platformclientv2.APIResponse, error)
type deleteBulkContactFunc func(ctx context.Context, p *contactProxy, contactListId, contactId string) (*platformclientv2.APIResponse, error)
type getAllBulkContactsFunc func(ctx context.Context, p *contactProxy) ([]ContactEntry, *platformclientv2.APIResponse, error)
type getContactListContactsExportUrlFunc func(ctx context.Context, p *contactProxy, contactListId string) (exportUrl string, resp *platformclientv2.APIResponse, error error)
type contactProxy struct {
	clientConfig                        *platformclientv2.Configuration
	outboundApi                         *platformclientv2.OutboundApi
	createBulkContactAttr               createBulkContactFunc
	readBulkContactByIdAttr             readBulkContactByIdFunc
	updateBulkContactAttr               updateBulkContactFunc
	deleteBulkContactAttr               deleteBulkContactFunc
	getAllBulkContactsAttr              getAllBulkContactsFunc
	contactCache                        rc.CacheInterface[platformclientv2.Dialercontact]
	basePath                            string
	accessToken                         string
	getContactListContactsExportUrlAttr getContactListContactsExportUrlFunc
}

func newBulkContactProxy(clientConfig *platformclientv2.Configuration) *contactProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &contactProxy{
		clientConfig:                        clientConfig,
		outboundApi:                         api,
		createBulkContactAttr:               createBulkContactFn,
		readBulkContactByIdAttr:             readBulkContactByIdFn,
		updateBulkContactAttr:               updateBulkContactFn,
		deleteBulkContactAttr:               deleteBulkContactFn,
		getAllBulkContactsAttr:              getAllBulkContactsFn,
		contactCache:                        contactCache,
		basePath:                            strings.Replace(api.Configuration.BasePath, "api", "apps", -1),
		accessToken:                         api.Configuration.AccessToken,
		getContactListContactsExportUrlAttr: getContactListContactsExportUrlFn,
	}
}

func getBulkContactsProxy(clientConfig *platformclientv2.Configuration) *contactProxy {
	return newBulkContactProxy(clientConfig)
}

func (p *contactProxy) createBulkContacts(ctx context.Context, contactListId string, contact platformclientv2.Writabledialercontact, priority, clearSystemData, doNotQueue bool) ([]platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	return p.createBulkContactAttr(ctx, p, contactListId, contact, priority, clearSystemData, doNotQueue)
}

func (p *contactProxy) readBulkContactsById(ctx context.Context, contactListId, contactId string) (*platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	return p.readBulkContactByIdAttr(ctx, p, contactListId, contactId)
}

func (p *contactProxy) updateBulkContacts(ctx context.Context, contactListId, contactId string, contact platformclientv2.Dialercontact) (*platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	return p.updateBulkContactAttr(ctx, p, contactListId, contactId, contact)
}

func (p *contactProxy) deleteBulkContacts(ctx context.Context, contactListId, contactId string) (*platformclientv2.APIResponse, error) {
	return p.deleteBulkContactAttr(ctx, p, contactListId, contactId)
}

func (p *contactProxy) getAllBulkContacts(ctx context.Context) ([]ContactEntry, *platformclientv2.APIResponse, error) {
	return p.getAllBulkContactsAttr(ctx, p)
}

func createBulkContactFn(_ context.Context, p *contactProxy, contactListId string, contact platformclientv2.Writabledialercontact, priority, clearSystemData, doNotQueue bool) ([]platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	return p.outboundApi.PostOutboundContactlistContacts(contactListId, []platformclientv2.Writabledialercontact{contact}, priority, clearSystemData, doNotQueue)
}

func readBulkContactByIdFn(_ context.Context, p *contactProxy, contactListId, contactId string) (*platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	if contact := rc.GetCacheItem(p.contactCache, buildComplexContactId(contactListId, contactId)); contact != nil {
		return contact, nil, nil
	}
	if tfexporter_state.IsExporterActive() {
		log.Printf("Could not read contact '%s' from cache (Contact list '%s'). Reading from the API...", contactId, contactListId)
	}
	return p.outboundApi.GetOutboundContactlistContact(contactListId, contactId)
}

func updateBulkContactFn(_ context.Context, p *contactProxy, contactListId, contactId string, contact platformclientv2.Dialercontact) (*platformclientv2.Dialercontact, *platformclientv2.APIResponse, error) {
	return p.outboundApi.PutOutboundContactlistContact(contactListId, contactId, contact)
}

func deleteBulkContactFn(_ context.Context, p *contactProxy, contactListId, contactId string) (*platformclientv2.APIResponse, error) {
	resp, err := p.outboundApi.DeleteOutboundContactlistContact(contactListId, contactId)
	if err != nil {
		return resp, err
	}
	rc.DeleteCacheItem(p.contactCache, buildComplexContactId(contactListId, contactId))
	return resp, nil
}

func getAllBulkContactsFn(ctx context.Context, p *contactProxy) ([]ContactEntry, *platformclientv2.APIResponse, error) {
	var allContacts []ContactEntry

	contactLists, resp, err := p.getAllContactLists(ctx)
	if err != nil {
		return allContacts, resp, err
	}

	for _, contactList := range contactLists {
		contacts, resp, err := p.getContactsByContactListId(ctx, *contactList.Id)
		if err != nil {
			return nil, resp, err
		}
		contactEntry := ContactEntry{
			ContactList: &contactList,
			Contact:     &contacts,
		}
		allContacts = append(allContacts, contactEntry)
		for _, contact := range contacts {
			rc.SetCache(p.contactCache, buildComplexContactId(*contactList.Id, *contact.Id), contact)
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

func (p *contactProxy) getAllContactLists(_ context.Context) ([]platformclientv2.Contactlist, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	var pageNum = 1
	var allContactLists []platformclientv2.Contactlist

	contactListConfigs, resp, getErr := p.outboundApi.GetOutboundContactlists(false, false, pageSize, pageNum, true, "", "", []string{}, []string{}, "", "")
	if getErr != nil {
		return nil, resp, getErr
	}
	if contactListConfigs.Entities == nil || len(*contactListConfigs.Entities) == 0 {
		return nil, nil, nil
	}
	for _, cl := range *contactListConfigs.Entities {
		allContactLists = append(allContactLists, cl)
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
			allContactLists = append(allContactLists, cl)
		}
	}

	return allContactLists, nil, nil
}

// createBulkOutboundContactsFormData creates the form data attributes to create a bulk upload of contacts in Genesys Cloud
func (p *contactProxy) createBulkOutboundContactsFormData(filePath, contactListId, contactIdColumnName string) (map[string]io.Reader, error) {
	fileReader, _, err := files.DownloadOrOpenFile(filePath)
	if err != nil {
		return nil, err
	}
	// The form data structure follows the Genesys Cloud API specification for uploading contact lists as CSV files
	// See full documentation at: https://developer.genesys.cloud/routing/outbound/uploadcontactlists
	formData := make(map[string]io.Reader)
	formData["file"] = fileReader
	formData["fileType"] = strings.NewReader("contactlist")
	formData["id"] = strings.NewReader(contactListId)
	formData["contact-id-name"] = strings.NewReader(contactIdColumnName)
	return formData, nil
}

// uploadBulkOutboundContactsFile uploads a CSV file to S3 of contacts
// For creates, scriptId should be an empty string
func (p *contactProxy) uploadBulkOutboundContactsFile(filePath, contactListId, contactIdColumnName string) ([]byte, error) {
	formData, err := p.createBulkOutboundContactsFormData(filePath, contactListId, contactIdColumnName)
	if err != nil {
		return nil, err
	}

	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + p.accessToken

	s3Uploader := files.NewS3Uploader(nil, formData, nil, headers, "POST", p.basePath+"/uploads/v2/contactlist")
	resp, err := s3Uploader.Upload()
	return resp, err
}

// getContactListContactsExportUrlFn retrieves the export URL for a contact list's contacts
func getContactListContactsExportUrlFn(_ context.Context, p *contactProxy, contactListId string) (exportUrl string, resp *platformclientv2.APIResponse, err error) {
	var (
		body platformclientv2.Contactsexportrequest
	)

	_, resp, err = p.outboundApi.PostOutboundContactlistExport(contactListId, body)

	if err != nil {
		return "", resp, fmt.Errorf("error calling PostOutboundContactlistExport with error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", resp, fmt.Errorf("error calling PostOutboundContactlistExport with status: %v", resp.Status)
	}

	data, resp, err := p.outboundApi.GetOutboundContactlistExport(contactListId, "")

	if err != nil {
		return "", resp, fmt.Errorf("error calling GetOutboundContactlistExport with error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", resp, fmt.Errorf("error calling GetOutboundContactlistExport with status: %v", resp.Status)
	}

	return *data.Uri, resp, nil
}
