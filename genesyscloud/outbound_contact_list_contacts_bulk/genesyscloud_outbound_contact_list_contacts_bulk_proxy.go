package outbound_contact_list_contacts_bulk

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
	contactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	"terraform-provider-genesyscloud/genesyscloud/util/files"

	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

var internalProxy *contactsBulkProxy

type uploadContactListBulkContactsFunc func(ctx context.Context, p *contactsBulkProxy, contactListId, filepath, contactIdName string) (respBytes []byte, err error)
type uploadContactListTemplateBulkContactsFunc func(ctx context.Context, p *contactsBulkProxy, contactListTemplateId, filepath, contactIdName, listNamePrefix, divisionIdForTargetContactLists string) (respBytes []byte, err error)
type readContactListAndRecordLengthByIdFunc func(ctx context.Context, p *contactsBulkProxy, contactListId string) (contactList *platformclientv2.Contactlist, recordLength int, response *platformclientv2.APIResponse, err error)
type clearContactListBulkContactsFunc func(ctx context.Context, p *contactsBulkProxy, contactListId string) (*platformclientv2.APIResponse, error)
type getAllContactListsFunc func(ctx context.Context, p *contactsBulkProxy) (*[]platformclientv2.Contactlist, *platformclientv2.APIResponse, error)
type getContactListContactsExportUrlFunc func(ctx context.Context, p *contactsBulkProxy, contactListId string) (exportUrl string, resp *platformclientv2.APIResponse, error error)
type contactsBulkProxy struct {
	clientConfig                             *platformclientv2.Configuration
	outboundApi                              *platformclientv2.OutboundApi
	uploadContactListBulkContactsAttr        uploadContactListBulkContactsFunc
	uploadContactListTemplateBulkContactAttr uploadContactListTemplateBulkContactsFunc
	readContactListAndRecordLengthByIdAttr   readContactListAndRecordLengthByIdFunc
	clearContactListBulkContactsAttr         clearContactListBulkContactsFunc
	getAllContactListsAttr                   getAllContactListsFunc
	basePath                                 string
	accessToken                              string
	getContactListContactsExportUrlAttr      getContactListContactsExportUrlFunc
	contactListProxy                         contactList.OutboundContactlistProxy
}

func newBulkContactProxy(clientConfig *platformclientv2.Configuration) *contactsBulkProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	outboundContactListProxy := contactList.GetOutboundContactlistProxy(clientConfig)
	return &contactsBulkProxy{
		clientConfig:                             clientConfig,
		outboundApi:                              api,
		uploadContactListBulkContactsAttr:        uploadContactListBulkContactsFn,
		uploadContactListTemplateBulkContactAttr: uploadContactListTemplateBulkContactFn,
		readContactListAndRecordLengthByIdAttr:   readContactListAndRecordLengthByIdFn,
		clearContactListBulkContactsAttr:         clearContactListBulkContactFn,
		getAllContactListsAttr:                   getAllBulkContactListsFn,
		basePath:                                 strings.Replace(api.Configuration.BasePath, "api", "apps", -1),
		accessToken:                              api.Configuration.AccessToken,
		getContactListContactsExportUrlAttr:      getContactListContactsExportUrlFn,
		contactListProxy:                         *outboundContactListProxy,
	}
}

func getBulkContactsProxy(clientConfig *platformclientv2.Configuration) *contactsBulkProxy {
	if internalProxy == nil {
		internalProxy = newBulkContactProxy(clientConfig)
	}
	return internalProxy
}

func (p *contactsBulkProxy) uploadContactListBulkContacts(ctx context.Context, contactListId, filepath, contactIdName string) (respBytes []byte, err error) {
	return p.uploadContactListBulkContactsAttr(ctx, p, contactListId, filepath, contactIdName)
}

func (p *contactsBulkProxy) uploadContactListTemplateBulkContacts(ctx context.Context, contactListTemplateId, filepath, contactIdName, listNamePrefix, divisionIdForTargetContactLists string) (respBytes []byte, err error) {
	return p.uploadContactListTemplateBulkContactAttr(ctx, p, contactListTemplateId, filepath, contactIdName, listNamePrefix, divisionIdForTargetContactLists)
}

func (p *contactsBulkProxy) readContactListAndRecordLengthById(ctx context.Context, contactListId string) (contactList *platformclientv2.Contactlist, recordsCount int, response *platformclientv2.APIResponse, err error) {
	return p.readContactListAndRecordLengthByIdAttr(ctx, p, contactListId)
}

func (p *contactsBulkProxy) clearContactListBulkContacts(ctx context.Context, contactListId string) (*platformclientv2.APIResponse, error) {
	return p.clearContactListBulkContactsAttr(ctx, p, contactListId)
}

func (p *contactsBulkProxy) getAllContactLists(ctx context.Context) (*[]platformclientv2.Contactlist, *platformclientv2.APIResponse, error) {
	return p.getAllContactListsAttr(ctx, p)
}

func (p *contactsBulkProxy) getContactListContactsExportUrl(ctx context.Context, contactListId string) (exportUrl string, resp *platformclientv2.APIResponse, err error) {
	return p.getContactListContactsExportUrlAttr(ctx, p, contactListId)
}

func (p contactsBulkProxy) getCSVRecordCount(filepath string) (int, error) {
	// Open file up and read the record count
	reader, file, err := files.DownloadOrOpenFile(filepath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Count the number of records in the CSV file
	csvReader := csv.NewReader(reader)
	recordCount := 0
	for {
		_, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
		recordCount++
	}

	// Subtract 1 to account for header row
	if recordCount > 0 {
		recordCount--
	}

	return recordCount, nil
}

// getContactListContactsExportUrlFn retrieves the export URL for a contact list's contacts
func getContactListContactsExportUrlFn(_ context.Context, p *contactsBulkProxy, contactListId string) (exportUrl string, resp *platformclientv2.APIResponse, err error) {
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

// createBulkOutboundContactsFormData creates the form data attributes to create a bulk upload of contacts in Genesys Cloud
func createBulkOutboundContactsFormData(filePath, contactListId, contactIdColumnName string) (map[string]io.Reader, error) {
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

// uploadContactListBulkContactsFn uploads a CSV file to S3 of contacts
func uploadContactListBulkContactsFn(ctx context.Context, p *contactsBulkProxy, contactListId, filePath, contactIdColumnName string) ([]byte, error) {
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

// createBulkOutboundContactsFormData creates the form data attributes to create a bulk upload of contacts in Genesys Cloud for contact list templates
func createBulkOutboundContactsTemplateFormData(filePath, contactListTemplateId, contactIdColumnName, listNamePrefix, divisionIdForTargetContactLists string) (map[string]io.Reader, error) {
	fileReader, _, err := files.DownloadOrOpenFile(filePath)
	if err != nil {
		return nil, err
	}
	// The form data structure follows the Genesys Cloud API specification for uploading contact lists as CSV files based on a template
	// See full documentation at: https://developer.genesys.cloud/routing/outbound/contactListBuilder
	formData := make(map[string]io.Reader)
	formData["file"] = fileReader
	formData["fileType"] = strings.NewReader("contactlist")
	formData["importTemplateId"] = strings.NewReader(contactListTemplateId)
	formData["contact-id-name"] = strings.NewReader(contactIdColumnName)
	formData["listNamePrefix"] = strings.NewReader(listNamePrefix)
	formData["divisionIdForTargetContactLists"] = strings.NewReader(divisionIdForTargetContactLists)
	return formData, nil
}

// uploadContactListTemplateBulkContactFn uploads a CSV file to S3 of contacts based on a contact list template
func uploadContactListTemplateBulkContactFn(ctx context.Context, p *contactsBulkProxy, filepath, contactListTemplateId, contactIdColumnName, listNamePrefix, divisionIdForTargetContactLists string) ([]byte, error) {
	formData, err := createBulkOutboundContactsTemplateFormData(filepath, contactListTemplateId, contactIdColumnName, listNamePrefix, divisionIdForTargetContactLists)
	if err != nil {
		return nil, err
	}

	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + p.accessToken

	s3Uploader := files.NewS3Uploader(nil, formData, nil, headers, "POST", p.basePath+"/uploads/v2/contactlisttemplate")
	respBytes, err := s3Uploader.Upload()
	return respBytes, err
}

func readContactListAndRecordLengthByIdFn(ctx context.Context, p *contactsBulkProxy, contactListId string) (contactList *platformclientv2.Contactlist, record_count int, response *platformclientv2.APIResponse, err error) {
	contactList, resp, err := p.contactListProxy.GetOutboundContactlistById(ctx, contactListId)
	if err != nil {
		return nil, 0, resp, err
	}
	blankReqBody := platformclientv2.Contactlistingrequest{}
	contactListing, resp, err := p.outboundApi.PostOutboundContactlistContactsSearch(contactListId, blankReqBody)
	if err != nil {
		return nil, 0, resp, err
	}
	return contactList, *contactListing.ContactsCount, resp, err
}

func clearContactListBulkContactFn(_ context.Context, p *contactsBulkProxy, contactListId string) (*platformclientv2.APIResponse, error) {
	resp, err := p.outboundApi.PostOutboundContactlistClear(contactListId)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func getAllBulkContactListsFn(ctx context.Context, p *contactsBulkProxy) (*[]platformclientv2.Contactlist, *platformclientv2.APIResponse, error) {
	contactLists, resp, err := p.contactListProxy.GetAllOutboundContactlist(ctx)
	if err != nil {
		return nil, resp, err
	}

	return contactLists, nil, nil
}
