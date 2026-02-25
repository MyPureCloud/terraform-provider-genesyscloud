package outbound_dnclist

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
	lists "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

var internalProxy *outboundDnclistProxy

// type definitions for each func on our proxy
type createOutboundDnclistFunc func(ctx context.Context, p *outboundDnclistProxy, dnclist *platformclientv2.Dnclistcreate) (*platformclientv2.Dnclist, *platformclientv2.APIResponse, error)
type getAllOutboundDnclistFunc func(ctx context.Context, p *outboundDnclistProxy) (*[]platformclientv2.Dnclist, *platformclientv2.APIResponse, error)
type getOutboundDnclistByIdFunc func(ctx context.Context, p *outboundDnclistProxy, dnclistId string) (*platformclientv2.Dnclist, *platformclientv2.APIResponse, error)
type getOutboundDnclistByNameFunc func(ctx context.Context, p *outboundDnclistProxy, name string) (dnclistId string, retryable bool, response *platformclientv2.APIResponse, err error)
type updateOutboundDnclistFunc func(ctx context.Context, p *outboundDnclistProxy, dnclistId string, dnclist *platformclientv2.Dnclist) (*platformclientv2.Dnclist, *platformclientv2.APIResponse, error)
type deleteOutboundDnclistFunc func(ctx context.Context, p *outboundDnclistProxy, dnclistId string) (*platformclientv2.APIResponse, error)
type uploadPhoneEntriesToDncListFunc func(p *outboundDnclistProxy, dncList *platformclientv2.Dnclist, entry interface{}) (*platformclientv2.APIResponse, diag.Diagnostics)
type deleteOutboundDnclistPhoneEntriesFunc func(ctx context.Context, p *outboundDnclistProxy, dnclistId string, expiredOnly bool) (*platformclientv2.APIResponse, error)
type initiateOutboundDnclistExportFunc func(ctx context.Context, p *outboundDnclistProxy, dncListId string) (*platformclientv2.Domainentityref, *platformclientv2.APIResponse, error)
type getOutboundDnclistExportFunc func(ctx context.Context, p *outboundDnclistProxy, dncListId string, download string) (*platformclientv2.Exporturi, *platformclientv2.APIResponse, error)
type getOutboundDnclistEntriesFunc func(ctx context.Context, p *outboundDnclistProxy, dncListId string) ([]interface{}, *platformclientv2.APIResponse, error)

// outboundDnclistProxy contains all the methods that call genesys cloud APIs
type outboundDnclistProxy struct {
	clientConfig                          *platformclientv2.Configuration
	outboundApi                           *platformclientv2.OutboundApi
	createOutboundDnclistAttr             createOutboundDnclistFunc
	getAllOutboundDnclistAttr             getAllOutboundDnclistFunc
	getOutboundDnclistByIdAttr            getOutboundDnclistByIdFunc
	getOutboundDnclistByNameAttr          getOutboundDnclistByNameFunc
	updateOutboundDnclistAttr             updateOutboundDnclistFunc
	deleteOutboundDnclistAttr             deleteOutboundDnclistFunc
	uploadPhoneEntriesToDncListAttr       uploadPhoneEntriesToDncListFunc
	deleteOutboundDnclistPhoneEntriesAttr deleteOutboundDnclistPhoneEntriesFunc
	initiateOutboundDnclistExportAttr     initiateOutboundDnclistExportFunc
	getOutboundDnclistExportAttr          getOutboundDnclistExportFunc
	getOutboundDnclistEntriesAttr         getOutboundDnclistEntriesFunc
}

// newOutboundDnclistProxy initializes the dnclist proxy with the data needed for communication with the genesys cloud
func newOutboundDnclistProxy(clientConfig *platformclientv2.Configuration) *outboundDnclistProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &outboundDnclistProxy{
		clientConfig:                          clientConfig,
		outboundApi:                           api,
		createOutboundDnclistAttr:             createOutboundDnclistFn,
		getAllOutboundDnclistAttr:             getAllOutboundDnclistFn,
		getOutboundDnclistByIdAttr:            getOutboundDnclistByIdFn,
		getOutboundDnclistByNameAttr:          getOutboundDnclistByNameFn,
		updateOutboundDnclistAttr:             updateOutboundDnclistFn,
		deleteOutboundDnclistAttr:             deleteOutboundDnclistFn,
		uploadPhoneEntriesToDncListAttr:       uploadPhoneEntriesToDncListFn,
		deleteOutboundDnclistPhoneEntriesAttr: deleteOutboundDnclistPhoneEntriesFn,
		initiateOutboundDnclistExportAttr:     initiateOutboundDnclistExportFn,
		getOutboundDnclistExportAttr:          getOutboundDnclistExportFn,
		getOutboundDnclistEntriesAttr:         getOutboundDnclistEntriesFn,
	}
}

func getOutboundDnclistProxy(clientConfig *platformclientv2.Configuration) *outboundDnclistProxy {
	if internalProxy == nil {
		internalProxy = newOutboundDnclistProxy(clientConfig)
	}
	return internalProxy
}

// createOutboundDnclist creates a Genesys Cloud Outbound Dnclist
func (p *outboundDnclistProxy) createOutboundDnclist(ctx context.Context, dnclist *platformclientv2.Dnclistcreate) (*platformclientv2.Dnclist, *platformclientv2.APIResponse, error) {
	return p.createOutboundDnclistAttr(ctx, p, dnclist)
}

// getAllOutboundDnclist retrieves all Genesys Cloud Outbound Dnclists
func (p *outboundDnclistProxy) getAllOutboundDnclist(ctx context.Context) (*[]platformclientv2.Dnclist, *platformclientv2.APIResponse, error) {
	return p.getAllOutboundDnclistAttr(ctx, p)
}

// getOutboundDnclistById returns a single Genesys Cloud Outbound Dnclist by Id
func (p *outboundDnclistProxy) getOutboundDnclistById(ctx context.Context, dnclistId string) (*platformclientv2.Dnclist, *platformclientv2.APIResponse, error) {
	return p.getOutboundDnclistByIdAttr(ctx, p, dnclistId)
}

// getOutboundDnclistByName returns a single Genesys Cloud Outbound Dnclist by a name
func (p *outboundDnclistProxy) getOutboundDnclistByName(ctx context.Context, name string) (dnclistId string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getOutboundDnclistByNameAttr(ctx, p, name)
}

// updateOutboundDnclist updates a Genesys Cloud Outbound Dnclist
func (p *outboundDnclistProxy) updateOutboundDnclist(ctx context.Context, dnclistId string, dnclist *platformclientv2.Dnclist) (*platformclientv2.Dnclist, *platformclientv2.APIResponse, error) {
	return p.updateOutboundDnclistAttr(ctx, p, dnclistId, dnclist)
}

// deleteOutboundDnclist deletes a Genesys Cloud Outbound Dnclist by Id
func (p *outboundDnclistProxy) deleteOutboundDnclist(ctx context.Context, dnclistId string) (*platformclientv2.APIResponse, error) {
	return p.deleteOutboundDnclistAttr(ctx, p, dnclistId)
}

func (p *outboundDnclistProxy) uploadPhoneEntriesToDncList(dncList *platformclientv2.Dnclist, entry interface{}) (*platformclientv2.APIResponse, diag.Diagnostics) {
	return p.uploadPhoneEntriesToDncListAttr(p, dncList, entry)
}

// deleteOutboundDnclistPhoneEntries deletes phone entries from a Genesys Cloud Outbound Dnclist
func (p *outboundDnclistProxy) deleteOutboundDnclistPhoneEntries(ctx context.Context, dnclistId string, expiredOnly bool) (*platformclientv2.APIResponse, error) {
	return p.deleteOutboundDnclistPhoneEntriesAttr(ctx, p, dnclistId, expiredOnly)
}

// initiateOutboundDnclistExport initiates the export for a Genesys Cloud Outbound Dnclist
func (p *outboundDnclistProxy) initiateOutboundDnclistExport(ctx context.Context, dncListId string) (*platformclientv2.Domainentityref, *platformclientv2.APIResponse, error) {
	return p.initiateOutboundDnclistExportAttr(ctx, p, dncListId)
}

// getOutboundDnclistExport retrieves the export status or download URI for a Genesys Cloud Outbound Dnclist
func (p *outboundDnclistProxy) getOutboundDnclistExport(ctx context.Context, dncListId string, download string) (*platformclientv2.Exporturi, *platformclientv2.APIResponse, error) {
	return p.getOutboundDnclistExportAttr(ctx, p, dncListId, download)
}

// getOutboundDnclistEntries retrieves the phone entries from a Genesys Cloud Outbound Dnclist
func (p *outboundDnclistProxy) getOutboundDnclistEntries(ctx context.Context, dncListId string) ([]interface{}, *platformclientv2.APIResponse, error) {
	return p.getOutboundDnclistEntriesAttr(ctx, p, dncListId)
}

func createOutboundDnclistFn(ctx context.Context, p *outboundDnclistProxy, dnclist *platformclientv2.Dnclistcreate) (*platformclientv2.Dnclist, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	return p.outboundApi.PostOutboundDnclists(*dnclist)
}

func getAllOutboundDnclistFn(ctx context.Context, p *outboundDnclistProxy) (*[]platformclientv2.Dnclist, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	var allDnclists []platformclientv2.Dnclist
	const pageSize = 100

	dnclists, resp, err := p.outboundApi.GetOutboundDnclists(false, false, pageSize, 1, true, "", "", "", []string{}, "", "")
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get dnclists: %v", err)
	}

	if dnclists.Entities == nil || len(*dnclists.Entities) == 0 {
		return &allDnclists, resp, nil
	}

	allDnclists = append(allDnclists, *dnclists.Entities...)

	var response *platformclientv2.APIResponse
	for pageNum := 2; pageNum <= *dnclists.PageCount; pageNum++ {
		dnclists, resp, err := p.outboundApi.GetOutboundDnclists(false, false, pageSize, pageNum, true, "", "", "", []string{}, "", "")
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get dnclists: %v", err)
		}
		response = resp
		if dnclists.Entities == nil || len(*dnclists.Entities) == 0 {
			break
		}

		allDnclists = append(allDnclists, *dnclists.Entities...)
	}
	return &allDnclists, response, nil
}

func updateOutboundDnclistFn(ctx context.Context, p *outboundDnclistProxy, dnclistId string, dnclist *platformclientv2.Dnclist) (*platformclientv2.Dnclist, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	return p.outboundApi.PutOutboundDnclist(dnclistId, *dnclist)
}

func uploadPhoneEntriesToDncListFn(p *outboundDnclistProxy, dncList *platformclientv2.Dnclist, entry interface{}) (*platformclientv2.APIResponse, diag.Diagnostics) {
	var phoneNumbers []string
	var resp *platformclientv2.APIResponse
	if entryMap, ok := entry.(map[string]interface{}); ok && len(entryMap) > 0 {
		if phoneNumbersList := entryMap["phone_numbers"].([]interface{}); phoneNumbersList != nil {
			for _, number := range phoneNumbersList {
				phoneNumbers = append(phoneNumbers, number.(string))
			}
		}
		log.Printf("Uploading phone numbers to DNC list %s", *dncList.Name)
		log.Printf("XXPhone numbers: %v", phoneNumbers)
		log.Printf("XXExpiration date: %v", entryMap["expiration_date"])
		// POST /api/v2/outbound/dnclists/{dncListId}/phonenumbers
		response, err := p.outboundApi.PostOutboundDnclistPhonenumbers(*dncList.Id, phoneNumbers, entryMap["expiration_date"].(string))
		if err != nil {
			return response, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to upload phone numbers to Outbound DNC list %s: %s", *dncList.Name, err), response)
		}
		resp = response
		log.Printf("Uploaded phone numbers to DNC list %s", *dncList.Name)
	}
	return resp, nil
}

func deleteOutboundDnclistPhoneEntriesFn(ctx context.Context, p *outboundDnclistProxy, dnclistId string, expiredOnly bool) (*platformclientv2.APIResponse, error) {
	return p.outboundApi.DeleteOutboundDnclistPhonenumbers(dnclistId, expiredOnly)
}

func deleteOutboundDnclistFn(ctx context.Context, p *outboundDnclistProxy, dnclistId string) (*platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	return p.outboundApi.DeleteOutboundDnclist(dnclistId)
}

func getOutboundDnclistByIdFn(ctx context.Context, p *outboundDnclistProxy, dnclistId string) (*platformclientv2.Dnclist, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	return p.outboundApi.GetOutboundDnclist(dnclistId, false, false)
}

func getOutboundDnclistByNameFn(ctx context.Context, p *outboundDnclistProxy, name string) (dnclistId string, retryable bool, response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	dnclists, resp, err := getAllOutboundDnclistFn(ctx, p)
	if err != nil {
		return "", false, resp, fmt.Errorf("error searching outbound dnc list %s: %s", name, err)
	}
	for _, dnclistSdk := range *dnclists {
		if *dnclistSdk.Name == name {
			log.Printf("Retrieved the dnc list id %s by name %s", *dnclistSdk.Id, name)
			return *dnclistSdk.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("unable to find dnc list with name %s", name)
}

func initiateOutboundDnclistExportFn(ctx context.Context, p *outboundDnclistProxy, dncListId string) (*platformclientv2.Domainentityref, *platformclientv2.APIResponse, error) {
	return p.outboundApi.PostOutboundDnclistExport(dncListId)
}

func getOutboundDnclistExportFn(ctx context.Context, p *outboundDnclistProxy, dncListId string, download string) (*platformclientv2.Exporturi, *platformclientv2.APIResponse, error) {
	return p.outboundApi.GetOutboundDnclistExport(dncListId, download)
}

func getOutboundDnclistEntriesFn(ctx context.Context, p *outboundDnclistProxy, dncListId string) ([]interface{}, *platformclientv2.APIResponse, error) {
	data, resp, err := p.getOutboundDnclistExport(ctx, dncListId, "")
	if err != nil {
		if util.IsStatus400(resp) {
			return nil, resp, fmt.Errorf("Export not ready yet for Outbound DNC list %s: %s", dncListId, err)
		}
		return nil, resp, fmt.Errorf("Failed to retrieve export URI for Outbound DNC list %s: %s", dncListId, err)
	}

	if data == nil || data.Uri == nil {
		return nil, resp, fmt.Errorf("Export URI is not found for Outbound DNC list %s", dncListId)
	}

	records, err := files.DownloadAndReadCSVFromURI(*data.Uri, p.clientConfig.AccessToken)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to download and read CSV for Outbound DNC list %s: %s", dncListId, err)
	}

	log.Printf("Downloaded records for Outbound DNC list %s: %d records", dncListId, len(records))
	for i, record := range records {
		log.Printf("  Record %d: %v", i, record)
	}

	// If only header row exists (no data), return empty list
	if len(records) <= 1 {
		return []interface{}{}, resp, nil
	}

	entries := make(map[string][]string)
	for _, record := range records[1:] {
		entries[record[2]] = append(entries[record[2]], record[0])
	}

	entriesList := make([]interface{}, 0)
	for expirationDate, phoneNumbers := range entries {
		entriesList = append(entriesList, map[string]interface{}{
			"expiration_date": expirationDate,
			"phone_numbers":   lists.StringListToInterfaceList(phoneNumbers),
		})
	}

	return entriesList, resp, nil
}
