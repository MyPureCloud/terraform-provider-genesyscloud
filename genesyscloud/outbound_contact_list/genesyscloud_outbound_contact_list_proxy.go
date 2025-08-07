package outbound_contact_list

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

/*
The genesyscloud_outbound_contact_list_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

var internalProxy *OutboundContactlistProxy

var contactListCache = rc.NewResourceCache[platformclientv2.Contactlist]()

// Type definitions for each func on our proxy so we can easily mock them out later
type createOutboundContactlistFunc func(ctx context.Context, p *OutboundContactlistProxy, contactList *platformclientv2.Contactlist) (*platformclientv2.Contactlist, *platformclientv2.APIResponse, error)
type getAllOutboundContactlistFunc func(ctx context.Context, p *OutboundContactlistProxy, name string) (*[]platformclientv2.Contactlist, *platformclientv2.APIResponse, error)
type getOutboundContactlistIdByNameFunc func(ctx context.Context, p *OutboundContactlistProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getOutboundContactlistByIdFunc func(ctx context.Context, p *OutboundContactlistProxy, id string) (contactList *platformclientv2.Contactlist, response *platformclientv2.APIResponse, err error)
type getOutboundContactlistContactRecordLengthFunc func(ctx context.Context, p *OutboundContactlistProxy, contactListId string) (int, *platformclientv2.APIResponse, error)
type updateOutboundContactlistFunc func(ctx context.Context, p *OutboundContactlistProxy, id string, contactList *platformclientv2.Contactlist) (*platformclientv2.Contactlist, *platformclientv2.APIResponse, error)
type deleteOutboundContactlistFunc func(ctx context.Context, p *OutboundContactlistProxy, id string) (response *platformclientv2.APIResponse, err error)
type uploadContactListBulkContactsFunc func(ctx context.Context, p *OutboundContactlistProxy, contactListId, filepath, contactIdName string) (respBytes []byte, err error)
type clearContactListContactsFunc func(ctx context.Context, p *OutboundContactlistProxy, contactListId string) (*platformclientv2.APIResponse, error)
type getContactListContactsExportUrlFunc func(ctx context.Context, p *OutboundContactlistProxy, contactListId string) (exportUrl string, resp *platformclientv2.APIResponse, error error)
type initiateContactListContactsExportFunc func(ctx context.Context, p *OutboundContactlistProxy, contactListId string) (resp *platformclientv2.APIResponse, error error)

// OutboundContactListProxy defines the interface for outbound contact list operations
type OutboundContactlistProxy struct {
	clientConfig                                  *platformclientv2.Configuration
	outboundApi                                   *platformclientv2.OutboundApi
	createOutboundContactlistAttr                 createOutboundContactlistFunc
	getAllOutboundContactlistAttr                 getAllOutboundContactlistFunc
	getOutboundContactlistIdByNameAttr            getOutboundContactlistIdByNameFunc
	getOutboundContactlistByIdAttr                getOutboundContactlistByIdFunc
	getOutboundContactlistContactRecordLengthAttr getOutboundContactlistContactRecordLengthFunc
	updateOutboundContactlistAttr                 updateOutboundContactlistFunc
	deleteOutboundContactlistAttr                 deleteOutboundContactlistFunc
	uploadContactListBulkContactsAttr             uploadContactListBulkContactsFunc
	clearContactListContactsAttr                  clearContactListContactsFunc
	basePath                                      string
	accessToken                                   string
	getContactListContactsExportUrlAttr           getContactListContactsExportUrlFunc
	initiateContactListContactsExportAttr         initiateContactListContactsExportFunc
	contactListCache                              rc.CacheInterface[platformclientv2.Contactlist]
}

// newOutboundContactlistProxy initializes the outbound contactlist proxy with all of the data needed to communicate with Genesys Cloud
func newOutboundContactlistProxy(clientConfig *platformclientv2.Configuration) *OutboundContactlistProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &OutboundContactlistProxy{
		clientConfig:                                  clientConfig,
		outboundApi:                                   api,
		createOutboundContactlistAttr:                 createOutboundContactlistFn,
		getAllOutboundContactlistAttr:                 getAllOutboundContactlistFn,
		getOutboundContactlistIdByNameAttr:            getOutboundContactlistIdByNameFn,
		getOutboundContactlistByIdAttr:                getOutboundContactlistByIdFn,
		getOutboundContactlistContactRecordLengthAttr: getOutboundContactlistContactRecordLengthFn,
		updateOutboundContactlistAttr:                 updateOutboundContactlistFn,
		deleteOutboundContactlistAttr:                 deleteOutboundContactlistFn,
		uploadContactListBulkContactsAttr:             uploadContactListBulkContactsFn,
		clearContactListContactsAttr:                  clearContactListContactsFn,
		basePath:                                      strings.Replace(api.Configuration.BasePath, "api", "apps", -1),
		accessToken:                                   api.Configuration.AccessToken,
		getContactListContactsExportUrlAttr:           getContactListContactsExportUrlFn,
		initiateContactListContactsExportAttr:         initiateContactListContactsExportFn,
		contactListCache:                              contactListCache,
	}
}

// GetOutboundContactlistProxy returns a proxy struct that implements the outbound contact list operations interface.
// It provides abstraction for the Genesys Cloud outbound contact list API operations.
//
// Parameters:
//   - sdkConfig: The Genesys Cloud SDK configuration containing authentication and connection settings
//
// Returns:
//   - An implementation of the outbound contact list proxy interface
//
// Example Usage:
//
//	sdkConfig := &platformclientv2.Configuration{
//		BasePath:           "https://api.mypurecloud.com",
//		DefaultHeader:      make(map[string]string),
//		UserAgent:         "terraform-provider-genesyscloud",
//	}
//
//	proxy := GetOutboundContactlistProxy(sdkConfig)
//	contactList, err := proxy.GetOutboundContactList(contactListId)
func GetOutboundContactlistProxy(clientConfig *platformclientv2.Configuration) *OutboundContactlistProxy {
	if internalProxy == nil {
		internalProxy = newOutboundContactlistProxy(clientConfig)
	}
	return internalProxy
}

// createOutboundContactlist creates a Genesys Cloud outbound contactlist
func (p *OutboundContactlistProxy) createOutboundContactlist(ctx context.Context, outboundContactlist *platformclientv2.Contactlist) (*platformclientv2.Contactlist, *platformclientv2.APIResponse, error) {
	return p.createOutboundContactlistAttr(ctx, p, outboundContactlist)
}

// GetAllOutboundContactlist retrieves all Genesys Cloud outbound contactlists
func (p *OutboundContactlistProxy) GetAllOutboundContactlist(ctx context.Context) (*[]platformclientv2.Contactlist, *platformclientv2.APIResponse, error) {
	return p.getAllOutboundContactlistAttr(ctx, p, "")
}

// getOutboundContactlistIdByName returns a single Genesys Cloud outbound contactlist by a name
func (p *OutboundContactlistProxy) getOutboundContactlistIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getOutboundContactlistIdByNameAttr(ctx, p, name)
}

// GetOutboundContactlistById returns a single Genesys Cloud outbound contactlist by Id
func (p *OutboundContactlistProxy) GetOutboundContactlistById(ctx context.Context, id string) (outboundContactlist *platformclientv2.Contactlist, response *platformclientv2.APIResponse, err error) {
	return p.getOutboundContactlistByIdAttr(ctx, p, id)
}

// getOutboundContactlistContactRecordLength returns the total record count of contacts on a contact list
func (p *OutboundContactlistProxy) getOutboundContactlistContactRecordLength(ctx context.Context, contactListId string) (int, *platformclientv2.APIResponse, error) {
	return p.getOutboundContactlistContactRecordLengthAttr(ctx, p, contactListId)
}

// updateOutboundContactlist updates a Genesys Cloud outbound contactlist
func (p *OutboundContactlistProxy) updateOutboundContactlist(ctx context.Context, id string, outboundContactlist *platformclientv2.Contactlist) (*platformclientv2.Contactlist, *platformclientv2.APIResponse, error) {
	return p.updateOutboundContactlistAttr(ctx, p, id, outboundContactlist)
}

// deleteOutboundContactlist deletes a Genesys Cloud outbound contactlist by Id
func (p *OutboundContactlistProxy) deleteOutboundContactlist(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteOutboundContactlistAttr(ctx, p, id)
}

// uploadContactListBulkContacts uploads a Genesys Cloud outbound contactlist
func (p *OutboundContactlistProxy) uploadContactListBulkContacts(ctx context.Context, contactListId, filepath, contactIdName string) (respBytes []byte, err error) {
	return p.uploadContactListBulkContactsAttr(ctx, p, contactListId, filepath, contactIdName)
}

// clearContactListContacts clears all of the contacts in a contact list
func (p *OutboundContactlistProxy) clearContactListContacts(ctx context.Context, contactListId string) (*platformclientv2.APIResponse, error) {
	return p.clearContactListContactsAttr(ctx, p, contactListId)
}

// initiateContactListContactsExport initiates the export for a contact list
func (p *OutboundContactlistProxy) initiateContactListContactsExport(ctx context.Context, contactListId string) (resp *platformclientv2.APIResponse, err error) {
	return p.initiateContactListContactsExportAttr(ctx, p, contactListId)
}

// getContactListContactsExportUrl gets the export url for a contact list (this is just the URL itself, no authorization included)
func (p *OutboundContactlistProxy) getContactListContactsExportUrl(ctx context.Context, contactListId string) (exportUrl string, resp *platformclientv2.APIResponse, err error) {
	return p.getContactListContactsExportUrlAttr(ctx, p, contactListId)
}

// createOutboundContactlistFn is an implementation function for creating a Genesys Cloud outbound contactlist
func createOutboundContactlistFn(ctx context.Context, p *OutboundContactlistProxy, outboundContactlist *platformclientv2.Contactlist) (*platformclientv2.Contactlist, *platformclientv2.APIResponse, error) {
	return p.outboundApi.PostOutboundContactlists(*outboundContactlist)
}

// getAllOutboundContactlistFn is the implementation for retrieving all outbound contactlist in Genesys Cloud
func getAllOutboundContactlistFn(ctx context.Context, p *OutboundContactlistProxy, name string) (*[]platformclientv2.Contactlist, *platformclientv2.APIResponse, error) {
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
func getOutboundContactlistIdByNameFn(ctx context.Context, p *OutboundContactlistProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
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
func getOutboundContactlistByIdFn(ctx context.Context, p *OutboundContactlistProxy, id string) (outboundContactlist *platformclientv2.Contactlist, response *platformclientv2.APIResponse, err error) {
	if contactList := rc.GetCacheItem(p.contactListCache, id); contactList != nil {
		return contactList, nil, nil
	}
	if tfexporter_state.IsExporterActive() {
		log.Printf("Could not read contact list '%s' from cache. Reading from the API...", id)
	}
	return p.outboundApi.GetOutboundContactlist(id, false, false)
}

// getOutboundContactlistContactRecordLengthFn is an implementation of the function to return a total count of contacts in a contact list
func getOutboundContactlistContactRecordLengthFn(ctx context.Context, p *OutboundContactlistProxy, contactListId string) (recordLength int, response *platformclientv2.APIResponse, err error) {
	blankReqBody := platformclientv2.Contactlistingrequest{}
	contactListing, resp, err := p.outboundApi.PostOutboundContactlistContactsSearch(contactListId, blankReqBody)
	if err != nil {
		return 0, resp, err
	}
	return int(*contactListing.ContactsCount), resp, nil
}

// updateOutboundContactlistFn is an implementation of the function to update a Genesys Cloud outbound contactlist
func updateOutboundContactlistFn(ctx context.Context, p *OutboundContactlistProxy, id string, outboundContactlist *platformclientv2.Contactlist) (*platformclientv2.Contactlist, *platformclientv2.APIResponse, error) {
	contactList, resp, err := p.outboundApi.GetOutboundContactlist(id, false, false)
	if err != nil {
		return nil, resp, err
	}

	outboundContactlist.Version = contactList.Version
	return p.outboundApi.PutOutboundContactlist(id, *outboundContactlist)
}

// deleteOutboundContactlistFn is an implementation function for deleting a Genesys Cloud outbound contactlist
func deleteOutboundContactlistFn(ctx context.Context, p *OutboundContactlistProxy, id string) (response *platformclientv2.APIResponse, err error) {
	resp, err := p.outboundApi.DeleteOutboundContactlist(id)
	if err != nil {
		return resp, err
	}
	rc.DeleteCacheItem(p.contactListCache, id)
	return resp, nil
}

// uploadContactListBulkContactsFn uploads a CSV file to S3 of contacts
func uploadContactListBulkContactsFn(ctx context.Context, p *OutboundContactlistProxy, contactListId, filePath, contactIdColumnName string) ([]byte, error) {
	formData, err := createBulkOutboundContactsFormData(filePath, contactListId, contactIdColumnName)
	if err != nil {
		return nil, err
	}

	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + p.accessToken

	s3Uploader := files.NewS3Uploader(nil, formData, nil, headers, "POST", p.basePath+"/uploads/v2/contactlist")
	respBytes, err := s3Uploader.Upload()
	return respBytes, err
}

func clearContactListContactsFn(_ context.Context, p *OutboundContactlistProxy, contactListId string) (*platformclientv2.APIResponse, error) {
	resp, err := p.outboundApi.PostOutboundContactlistClear(contactListId)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// initiateContactListContactsExportFn is an implementation function for retrieving the export URL for a contact list's contacts
func initiateContactListContactsExportFn(_ context.Context, p *OutboundContactlistProxy, contactListId string) (resp *platformclientv2.APIResponse, err error) {
	var (
		body platformclientv2.Contactsexportrequest
	)

	_, resp, err = p.outboundApi.PostOutboundContactlistExport(contactListId, body)

	if err != nil {
		return resp, fmt.Errorf("error calling PostOutboundContactlistExport with error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("error calling PostOutboundContactlistExport with status: %v", resp.Status)
	}
	return resp, nil
}

// getContactListContactsExportUrlFn is an implementation function for retrieving the export URL for a contact list's contacts
func getContactListContactsExportUrlFn(_ context.Context, p *OutboundContactlistProxy, contactListId string) (exportUrl string, resp *platformclientv2.APIResponse, err error) {

	data, resp, err := p.outboundApi.GetOutboundContactlistExport(contactListId, "")

	if err != nil {
		return "", resp, fmt.Errorf("error calling GetOutboundContactlistExport with error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", resp, fmt.Errorf("error calling GetOutboundContactlistExport with status: %v", resp.Status)
	}

	return *data.Uri, resp, nil
}

// createBulkOutboundContactsFormData creates the form data attributes to create a bulk upload of contacts in Genesys Cloud
func createBulkOutboundContactsFormData(filePath, contactListId, contactIdColumnName string) (map[string]io.Reader, error) {
	fileReader, _, err := files.DownloadOrOpenFile(context.Background(), filePath, S3Enabled)
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
