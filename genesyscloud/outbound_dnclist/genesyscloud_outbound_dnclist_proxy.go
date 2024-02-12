package outbound_dnclist

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
	"log"
)

var internalProxy *outboundDnclistProxy

// type definitions for each func on our proxy
type createOutboundDnclistFunc func(ctx context.Context, p *outboundDnclistProxy, dnclist *platformclientv2.Dnclistcreate) (*platformclientv2.Dnclist, int, error)
type getAllOutboundDnclistFunc func(ctx context.Context, p *outboundDnclistProxy) (*[]platformclientv2.Dnclist, int, error)
type getOutboundDnclistByIdFunc func(ctx context.Context, p *outboundDnclistProxy, dnclistId string) (dnclist *platformclientv2.Dnclist, responseCode int, err error)
type getOutboundDnclistByNameFunc func(ctx context.Context, p *outboundDnclistProxy, name string) (dnclistId string, retryable bool, err error)
type updateOutboundDnclistFunc func(ctx context.Context, p *outboundDnclistProxy, dnclistId string, dnclist *platformclientv2.Dnclist) (*platformclientv2.Dnclist, int, error)
type deleteOutboundDnclistFunc func(ctx context.Context, p *outboundDnclistProxy, dnclistId string) (responseCode int, err error)

// outboundDnclistProxy contains all of the methods that call genesys cloud APIs
type outboundDnclistProxy struct {
	clientConfig                 *platformclientv2.Configuration
	outboundApi                  *platformclientv2.OutboundApi
	createOutboundDnclistAttr    createOutboundDnclistFunc
	getAllOutboundDnclistAttr    getAllOutboundDnclistFunc
	getOutboundDnclistByIdAttr   getOutboundDnclistByIdFunc
	getOutboundDnclistByNameAttr getOutboundDnclistByNameFunc
	updateOutboundDnclistAttr    updateOutboundDnclistFunc
	deleteOutboundDnclistAttr    deleteOutboundDnclistFunc
}

// newOutboundDnclistProxy initializes the dnclist proxy with the data needed for communication with the genesys cloud
func newOutboundDnclistProxy(clientConfig *platformclientv2.Configuration) *outboundDnclistProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &outboundDnclistProxy{
		clientConfig:                 clientConfig,
		outboundApi:                  api,
		createOutboundDnclistAttr:    createOutboundDnclistFn,
		getAllOutboundDnclistAttr:    getAllOutboundDnclistFn,
		getOutboundDnclistByIdAttr:   getOutboundDnclistByIdFn,
		getOutboundDnclistByNameAttr: getOutboundDnclistByNameFn,
		updateOutboundDnclistAttr:    updateOutboundDnclistFn,
		deleteOutboundDnclistAttr:    deleteOutboundDnclistFn,
	}
}

func getOutboundDnclistProxy(clientConfig *platformclientv2.Configuration) *outboundDnclistProxy {
	if internalProxy == nil {
		internalProxy = newOutboundDnclistProxy(clientConfig)
	}

	return internalProxy
}

// createOutboundDnclist creates a Genesys Cloud Outbound Dnclist
func (p *outboundDnclistProxy) createOutboundDnclist(ctx context.Context, dnclist *platformclientv2.Dnclistcreate) (*platformclientv2.Dnclist, int, error) {
	return p.createOutboundDnclistAttr(ctx, p, dnclist)
}

// getAllOutboundDnclist retrieves all Genesys Cloud Outbound Dnclists
func (p *outboundDnclistProxy) getAllOutboundDnclist(ctx context.Context) (*[]platformclientv2.Dnclist, int, error) {
	return p.getAllOutboundDnclistAttr(ctx, p)
}

// getOutboundDnclistById returns a single Genesys Cloud Outbound Dnclist by Id
func (p *outboundDnclistProxy) getOutboundDnclistById(ctx context.Context, dnclistId string) (dnclist *platformclientv2.Dnclist, statusCode int, err error) {
	return p.getOutboundDnclistByIdAttr(ctx, p, dnclistId)
}

// getOutboundDnclistByName returns a single Genesys Cloud Outbound Dnclist by a name
func (p *outboundDnclistProxy) getOutboundDnclistByName(ctx context.Context, name string) (dnclistId string, retryable bool, err error) {
	return p.getOutboundDnclistByNameAttr(ctx, p, name)
}

// updateOutboundDnclist updates a Genesys Cloud Outbound Dnclist
func (p *outboundDnclistProxy) updateOutboundDnclist(ctx context.Context, dnclistId string, dnclist *platformclientv2.Dnclist) (*platformclientv2.Dnclist, int, error) {
	return p.updateOutboundDnclistAttr(ctx, p, dnclistId, dnclist)
}

// deleteOutboundDnclist deletes a Genesys Cloud Outbound Dnclist by Id
func (p *outboundDnclistProxy) deleteOutboundDnclist(ctx context.Context, dnclistId string) (statusCode int, err error) {
	return p.deleteOutboundDnclistAttr(ctx, p, dnclistId)
}

func createOutboundDnclistFn(ctx context.Context, p *outboundDnclistProxy, dnclist *platformclientv2.Dnclistcreate) (*platformclientv2.Dnclist, int, error) {
	list, resp, err := p.outboundApi.PostOutboundDnclists(*dnclist)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create dnclist: %s", err)
	}
	return list, resp.StatusCode, nil
}

func getAllOutboundDnclistFn(ctx context.Context, p *outboundDnclistProxy) (*[]platformclientv2.Dnclist, int, error) {
	var allDnclists []platformclientv2.Dnclist

	dnclists, resp, err := p.outboundApi.GetOutboundDnclists(false, false, 100, 1, true, "", "", "", []string{}, "", "")
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to get dnclists: %v", err)
	}

	if dnclists.Entities == nil || len(*dnclists.Entities) == 0 {
		return &allDnclists, resp.StatusCode, nil
	}

	for _, dnclist := range *dnclists.Entities {
		allDnclists = append(allDnclists, dnclist)
	}

	var statusCode int
	for pageNum := 2; pageNum <= *dnclists.PageCount; pageNum++ {
		const pageSize = 100
		dnclists, resp, err := p.outboundApi.GetOutboundDnclists(false, false, pageSize, pageNum, true, "", "", "", []string{}, "", "")
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to get dnclists: %v", err)
		}
		statusCode = resp.StatusCode
		if dnclists.Entities == nil || len(*dnclists.Entities) == 0 {
			break
		}

		for _, dnclist := range *dnclists.Entities {
			log.Printf("Dealing with dnclist: %s", *dnclist.Id)
			allDnclists = append(allDnclists, dnclist)
		}
	}
	return &allDnclists, statusCode, nil
}

func uploadPhoneEntriesToDncList(p *outboundDnclistProxy, dncList *platformclientv2.Dnclist, entry interface{}) (*platformclientv2.APIResponse, diag.Diagnostics) {
	var phoneNumbers []string
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
			return response, diag.Errorf("Failed to upload phone numbers to Outbound DNC list %s: %s", *dncList.Name, err)
		}
		log.Printf("Uploaded phone numbers to DNC list %s", *dncList.Name)
	}
	return nil, nil
}
