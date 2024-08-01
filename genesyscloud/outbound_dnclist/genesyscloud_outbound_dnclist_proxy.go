package outbound_dnclist

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
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

// outboundDnclistProxy contains all the methods that call genesys cloud APIs
type outboundDnclistProxy struct {
	clientConfig                    *platformclientv2.Configuration
	outboundApi                     *platformclientv2.OutboundApi
	createOutboundDnclistAttr       createOutboundDnclistFunc
	getAllOutboundDnclistAttr       getAllOutboundDnclistFunc
	getOutboundDnclistByIdAttr      getOutboundDnclistByIdFunc
	getOutboundDnclistByNameAttr    getOutboundDnclistByNameFunc
	updateOutboundDnclistAttr       updateOutboundDnclistFunc
	deleteOutboundDnclistAttr       deleteOutboundDnclistFunc
	uploadPhoneEntriesToDncListAttr uploadPhoneEntriesToDncListFunc
}

// newOutboundDnclistProxy initializes the dnclist proxy with the data needed for communication with the genesys cloud
func newOutboundDnclistProxy(clientConfig *platformclientv2.Configuration) *outboundDnclistProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &outboundDnclistProxy{
		clientConfig:                    clientConfig,
		outboundApi:                     api,
		createOutboundDnclistAttr:       createOutboundDnclistFn,
		getAllOutboundDnclistAttr:       getAllOutboundDnclistFn,
		getOutboundDnclistByIdAttr:      getOutboundDnclistByIdFn,
		getOutboundDnclistByNameAttr:    getOutboundDnclistByNameFn,
		updateOutboundDnclistAttr:       updateOutboundDnclistFn,
		deleteOutboundDnclistAttr:       deleteOutboundDnclistFn,
		uploadPhoneEntriesToDncListAttr: uploadPhoneEntriesToDncListFn,
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

func createOutboundDnclistFn(ctx context.Context, p *outboundDnclistProxy, dnclist *platformclientv2.Dnclistcreate) (*platformclientv2.Dnclist, *platformclientv2.APIResponse, error) {
	list, resp, err := p.outboundApi.PostOutboundDnclists(*dnclist)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to create dnclist: %s", err)
	}
	return list, resp, nil
}

func getAllOutboundDnclistFn(ctx context.Context, p *outboundDnclistProxy) (*[]platformclientv2.Dnclist, *platformclientv2.APIResponse, error) {
	var allDnclists []platformclientv2.Dnclist
	const pageSize = 100

	dnclists, resp, err := p.outboundApi.GetOutboundDnclists(false, false, pageSize, 1, true, "", "", "", []string{}, "", "")
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get dnclists: %v", err)
	}

	if dnclists.Entities == nil || len(*dnclists.Entities) == 0 {
		return &allDnclists, resp, nil
	}

	for _, dnclist := range *dnclists.Entities {
		allDnclists = append(allDnclists, dnclist)
	}

	var response *platformclientv2.APIResponse
	for pageNum := 2; pageNum <= *dnclists.PageCount; pageNum++ {
		dnclists, resp, err := p.outboundApi.GetOutboundDnclists(false, false, pageSize, pageNum, true, "", "", "", []string{}, "", "")
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get dnclists: %v", err)
		}
		response = resp
		if dnclists.Entities == nil || len(*dnclists.Entities) == 0 {
			break
		}

		for _, dnclist := range *dnclists.Entities {
			log.Printf("Dealing with dnclist: %s", *dnclist.Id)
			allDnclists = append(allDnclists, dnclist)
		}
	}
	return &allDnclists, response, nil
}

func updateOutboundDnclistFn(ctx context.Context, p *outboundDnclistProxy, dnclistId string, dnclist *platformclientv2.Dnclist) (*platformclientv2.Dnclist, *platformclientv2.APIResponse, error) {
	dnclist, resp, err := p.outboundApi.GetOutboundDnclist(dnclistId, false, false)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get dnc list by id %s", err)
	}

	outboundDncList, resp, err := p.outboundApi.PutOutboundDnclist(dnclistId, *dnclist)
	if err != nil {
		return nil, resp, fmt.Errorf("error updating outbound dnc list %s", err)
	}
	return outboundDncList, resp, nil
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
		// POST /api/v2/outbound/dnclists/{dncListId}/phonenumbers
		response, err := p.outboundApi.PostOutboundDnclistPhonenumbers(*dncList.Id, phoneNumbers, entryMap["expiration_date"].(string))
		if err != nil {
			return response, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to upload phone numbers to Outbound DNC list %s: %s", *dncList.Name, err), response)
		}
		resp = response
		log.Printf("Uploaded phone numbers to DNC list %s", *dncList.Name)
	}
	return resp, nil
}

func deleteOutboundDnclistFn(ctx context.Context, p *outboundDnclistProxy, dnclistId string) (*platformclientv2.APIResponse, error) {
	resp, err := p.outboundApi.DeleteOutboundDnclist(dnclistId)
	if err != nil {
		return resp, fmt.Errorf("failed to delete dnc list %s", err)
	}
	return resp, nil
}

func getOutboundDnclistByIdFn(ctx context.Context, p *outboundDnclistProxy, dnclistId string) (*platformclientv2.Dnclist, *platformclientv2.APIResponse, error) {
	dnclist, resp, err := p.outboundApi.GetOutboundDnclist(dnclistId, false, false)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to retrieve dnc list by id %s: %s", dnclistId, err)
	}
	return dnclist, resp, nil
}

func getOutboundDnclistByNameFn(ctx context.Context, p *outboundDnclistProxy, name string) (dnclistId string, retryable bool, response *platformclientv2.APIResponse, err error) {
	dnclists, resp, err := getAllOutboundDnclistFn(ctx, p)
	if err != nil {
		return "", false, resp, fmt.Errorf("Error searching outbound dnc list %s: %s", name, err)
	}

	var dnclist platformclientv2.Dnclist
	for _, dnclistSdk := range *dnclists {
		if *dnclistSdk.Name == name {
			log.Printf("Retrieved the dnc list id %s by name %s", *dnclistSdk.Id, name)
			dnclist = dnclistSdk
			return *dnclist.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("Unable to find dnc list with name %s", name)
}
