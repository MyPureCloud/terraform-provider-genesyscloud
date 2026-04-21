package case_management_caseplan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"
)

/*
The genesyscloud_case_management_caseplan_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on our proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *caseManagementCaseplanProxy

// caseplanDataschemaKeyDefault is the only schema key name supported today for PUT/DELETE
// /api/v2/casemanagement/caseplans/{caseplanId}/dataschemas/{schemaKeyName}.
const caseplanDataschemaKeyDefault = "default"

// caseplanDataschemaPostBody is the JSON body for POST /caseplans/{caseplanId}/dataschemas (add to draft).
type caseplanDataschemaPostBody struct {
	Id string `json:"id"`
}

// Type definitions for each func on our proxy so we can easily mock them out later
type createCaseManagementCaseplanFunc func(ctx context.Context, p *caseManagementCaseplanProxy, body *platformclientv2.Caseplancreate) (*platformclientv2.Caseplancreateresponse, *platformclientv2.APIResponse, error)
type getAllCaseManagementCaseplanFunc func(ctx context.Context, p *caseManagementCaseplanProxy) (*[]platformclientv2.Caseplan, *platformclientv2.APIResponse, error)
type getCaseManagementCaseplanIdByNameFunc func(ctx context.Context, p *caseManagementCaseplanProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type getCaseManagementCaseplanByIdFunc func(ctx context.Context, p *caseManagementCaseplanProxy, id string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error)
type getCaseManagementCaseplanVersionDataschemasFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, versionId string) (*platformclientv2.Caseplandataschemalisting, *platformclientv2.APIResponse, error)
type getCaseManagementCaseplanVersionIntakesettingsFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, versionId string) (*platformclientv2.Intakesettingslisting, *platformclientv2.APIResponse, error)
type putCaseManagementCaseplanIntakesettingsFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, body platformclientv2.Intakesettingsupdate) (*platformclientv2.Intakesettingslisting, *platformclientv2.APIResponse, error)
type patchCaseManagementCaseplanFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, body platformclientv2.Caseplanupdate) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error)
type postCaseManagementCaseplanDataschemaFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, body caseplanDataschemaPostBody) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error)
type putCaseManagementCaseplanDataschemaFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, schemaKeyName string, body platformclientv2.Caseplandataschema) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error)
type deleteCaseManagementCaseplanDataschemaFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, schemaKeyName string) (*platformclientv2.APIResponse, error)
type postCaseManagementCaseplanVersionsFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error)
type publishCaseManagementCaseplanFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error)
type deleteCaseManagementCaseplanFunc func(ctx context.Context, p *caseManagementCaseplanProxy, id string) (*platformclientv2.APIResponse, error)

// caseManagementCaseplanProxy contains all of the methods that call genesys cloud APIs.
type caseManagementCaseplanProxy struct {
	clientConfig                                       *platformclientv2.Configuration
	caseManagementApi                                  *platformclientv2.CaseManagementApi
	createCaseManagementCaseplanAttr                   createCaseManagementCaseplanFunc
	getAllCaseManagementCaseplanAttr                   getAllCaseManagementCaseplanFunc
	getCaseManagementCaseplanIdByNameAttr              getCaseManagementCaseplanIdByNameFunc
	getCaseManagementCaseplanByIdAttr                  getCaseManagementCaseplanByIdFunc
	getCaseManagementCaseplanVersionDataschemasAttr    getCaseManagementCaseplanVersionDataschemasFunc
	getCaseManagementCaseplanVersionIntakesettingsAttr getCaseManagementCaseplanVersionIntakesettingsFunc
	putCaseManagementCaseplanIntakesettingsAttr        putCaseManagementCaseplanIntakesettingsFunc
	patchCaseManagementCaseplanAttr                    patchCaseManagementCaseplanFunc
	postCaseManagementCaseplanDataschemaAttr           postCaseManagementCaseplanDataschemaFunc
	putCaseManagementCaseplanDataschemaAttr            putCaseManagementCaseplanDataschemaFunc
	deleteCaseManagementCaseplanDataschemaAttr         deleteCaseManagementCaseplanDataschemaFunc
	postCaseManagementCaseplanVersionsAttr             postCaseManagementCaseplanVersionsFunc
	publishCaseManagementCaseplanAttr                  publishCaseManagementCaseplanFunc
	deleteCaseManagementCaseplanAttr                   deleteCaseManagementCaseplanFunc
}

// newCaseManagementCaseplanProxy initializes the case management caseplan proxy with all of the data needed to communicate with Genesys Cloud
func newCaseManagementCaseplanProxy(clientConfig *platformclientv2.Configuration) *caseManagementCaseplanProxy {
	api := platformclientv2.NewCaseManagementApiWithConfig(clientConfig)
	return &caseManagementCaseplanProxy{
		clientConfig:                                       clientConfig,
		caseManagementApi:                                  api,
		createCaseManagementCaseplanAttr:                   createCaseManagementCaseplanFn,
		getAllCaseManagementCaseplanAttr:                   getAllCaseManagementCaseplanFn,
		getCaseManagementCaseplanIdByNameAttr:              getCaseManagementCaseplanIdByNameFn,
		getCaseManagementCaseplanByIdAttr:                  getCaseManagementCaseplanByIdFn,
		getCaseManagementCaseplanVersionDataschemasAttr:    getCaseManagementCaseplanVersionDataschemasFn,
		getCaseManagementCaseplanVersionIntakesettingsAttr: getCaseManagementCaseplanVersionIntakesettingsFn,
		putCaseManagementCaseplanIntakesettingsAttr:        putCaseManagementCaseplanIntakesettingsFn,
		patchCaseManagementCaseplanAttr:                    patchCaseManagementCaseplanFn,
		postCaseManagementCaseplanDataschemaAttr:           postCaseManagementCaseplanDataschemaFn,
		putCaseManagementCaseplanDataschemaAttr:            putCaseManagementCaseplanDataschemaFn,
		deleteCaseManagementCaseplanDataschemaAttr:         deleteCaseManagementCaseplanDataschemaFn,
		postCaseManagementCaseplanVersionsAttr:             postCaseManagementCaseplanVersionsFn,
		publishCaseManagementCaseplanAttr:                  publishCaseManagementCaseplanFn,
		deleteCaseManagementCaseplanAttr:                   deleteCaseManagementCaseplanFn,
	}
}

// getCaseManagementCaseplanProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getCaseManagementCaseplanProxy(clientConfig *platformclientv2.Configuration) *caseManagementCaseplanProxy {
	if internalProxy == nil {
		internalProxy = newCaseManagementCaseplanProxy(clientConfig)
	}

	return internalProxy
}

// createCaseManagementCaseplan creates a Genesys Cloud case management caseplan
func (p *caseManagementCaseplanProxy) createCaseManagementCaseplan(ctx context.Context, body *platformclientv2.Caseplancreate) (*platformclientv2.Caseplancreateresponse, *platformclientv2.APIResponse, error) {
	return p.createCaseManagementCaseplanAttr(ctx, p, body)
}

// getAllCaseManagementCaseplan retrieves all Genesys Cloud case management caseplan
func (p *caseManagementCaseplanProxy) getAllCaseManagementCaseplan(ctx context.Context) (*[]platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
	return p.getAllCaseManagementCaseplanAttr(ctx, p)
}

// getCaseManagementCaseplanIdByName returns a single Genesys Cloud case management caseplan by a name
func (p *caseManagementCaseplanProxy) getCaseManagementCaseplanIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getCaseManagementCaseplanIdByNameAttr(ctx, p, name)
}

// getCaseManagementCaseplanById returns a single Genesys Cloud case management caseplan by Id
func (p *caseManagementCaseplanProxy) getCaseManagementCaseplanById(ctx context.Context, id string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
	return p.getCaseManagementCaseplanByIdAttr(ctx, p, id)
}

func (p *caseManagementCaseplanProxy) getCaseManagementCaseplanVersionDataschemas(ctx context.Context, caseplanId string, versionId string) (*platformclientv2.Caseplandataschemalisting, *platformclientv2.APIResponse, error) {
	return p.getCaseManagementCaseplanVersionDataschemasAttr(ctx, p, caseplanId, versionId)
}

func (p *caseManagementCaseplanProxy) getCaseManagementCaseplanVersionIntakesettings(ctx context.Context, caseplanId string, versionId string) (*platformclientv2.Intakesettingslisting, *platformclientv2.APIResponse, error) {
	return p.getCaseManagementCaseplanVersionIntakesettingsAttr(ctx, p, caseplanId, versionId)
}

func (p *caseManagementCaseplanProxy) putCaseManagementCaseplanIntakesettings(ctx context.Context, caseplanId string, body platformclientv2.Intakesettingsupdate) (*platformclientv2.Intakesettingslisting, *platformclientv2.APIResponse, error) {
	return p.putCaseManagementCaseplanIntakesettingsAttr(ctx, p, caseplanId, body)
}

func (p *caseManagementCaseplanProxy) patchCaseManagementCaseplan(ctx context.Context, caseplanId string, body platformclientv2.Caseplanupdate) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
	return p.patchCaseManagementCaseplanAttr(ctx, p, caseplanId, body)
}

func (p *caseManagementCaseplanProxy) postCaseManagementCaseplanDataschema(ctx context.Context, caseplanId string, body caseplanDataschemaPostBody) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error) {
	return p.postCaseManagementCaseplanDataschemaAttr(ctx, p, caseplanId, body)
}

func (p *caseManagementCaseplanProxy) putCaseManagementCaseplanDataschema(ctx context.Context, caseplanId string, schemaKeyName string, body platformclientv2.Caseplandataschema) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error) {
	return p.putCaseManagementCaseplanDataschemaAttr(ctx, p, caseplanId, schemaKeyName, body)
}

func (p *caseManagementCaseplanProxy) deleteCaseManagementCaseplanDataschema(ctx context.Context, caseplanId string, schemaKeyName string) (*platformclientv2.APIResponse, error) {
	return p.deleteCaseManagementCaseplanDataschemaAttr(ctx, p, caseplanId, schemaKeyName)
}

func (p *caseManagementCaseplanProxy) postCaseManagementCaseplanVersions(ctx context.Context, caseplanId string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
	return p.postCaseManagementCaseplanVersionsAttr(ctx, p, caseplanId)
}

func (p *caseManagementCaseplanProxy) publishCaseManagementCaseplan(ctx context.Context, caseplanId string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
	return p.publishCaseManagementCaseplanAttr(ctx, p, caseplanId)
}

// deleteCaseManagementCaseplan deletes a Genesys Cloud case management caseplan by Id
func (p *caseManagementCaseplanProxy) deleteCaseManagementCaseplan(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteCaseManagementCaseplanAttr(ctx, p, id)
}

func createCaseManagementCaseplanFn(ctx context.Context, p *caseManagementCaseplanProxy, body *platformclientv2.Caseplancreate) (*platformclientv2.Caseplancreateresponse, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.PostCasemanagementCaseplans(*body)
}

func listAllCaseplans(p *caseManagementCaseplanProxy) ([]platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
	var all []platformclientv2.Caseplan
	after := ""
	const pageSize = 100
	var lastResp *platformclientv2.APIResponse

	for {
		listing, resp, err := p.caseManagementApi.GetCasemanagementCaseplans(after, pageSize, "", "")
		lastResp = resp
		if err != nil {
			return nil, resp, err
		}
		if listing == nil || listing.Entities == nil || len(*listing.Entities) == 0 {
			break
		}
		entities := *listing.Entities
		all = append(all, entities...)
		if len(entities) < pageSize {
			break
		}
		last := entities[len(entities)-1]
		if last.Id == nil || *last.Id == "" {
			break
		}
		next := *last.Id
		if next == after {
			break
		}
		after = next
	}

	return all, lastResp, nil
}

func getAllCaseManagementCaseplanFn(ctx context.Context, p *caseManagementCaseplanProxy) (*[]platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
	all, lastResp, err := listAllCaseplans(p)
	if err != nil {
		return nil, lastResp, err
	}
	return &all, lastResp, nil
}

func getCaseManagementCaseplanIdByNameFn(ctx context.Context, p *caseManagementCaseplanProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	all, resp, err := listAllCaseplans(p)
	if err != nil {
		return "", resp, false, err
	}
	for i := range all {
		caseplan := all[i]
		if caseplan.Name != nil && *caseplan.Name == name && caseplan.Id != nil {
			return *caseplan.Id, resp, false, nil
		}
	}
	return "", resp, true, fmt.Errorf("unable to find case management caseplan with name %s", name)
}

func getCaseManagementCaseplanByIdFn(ctx context.Context, p *caseManagementCaseplanProxy, id string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.GetCasemanagementCaseplan(id)
}

func getCaseManagementCaseplanVersionDataschemasFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, versionId string) (*platformclientv2.Caseplandataschemalisting, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.GetCasemanagementCaseplanVersionDataschemas(caseplanId, versionId)
}

func getCaseManagementCaseplanVersionIntakesettingsFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, versionId string) (*platformclientv2.Intakesettingslisting, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.GetCasemanagementCaseplanVersionIntakesettings(caseplanId, versionId)
}

func putCaseManagementCaseplanIntakesettingsFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, body platformclientv2.Intakesettingsupdate) (*platformclientv2.Intakesettingslisting, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.PutCasemanagementCaseplanIntakesettings(caseplanId, body)
}

func patchCaseManagementCaseplanFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, body platformclientv2.Caseplanupdate) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.PatchCasemanagementCaseplan(caseplanId, body)
}

// postCaseManagementCaseplanDataschemaFn calls POST /api/v2/casemanagement/caseplans/{caseplanId}/dataschemas with body {"id":"..."}.
// Not generated in platform-client-sdk-go v186.
func postCaseManagementCaseplanDataschemaFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, body caseplanDataschemaPostBody) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error) {
	cfg := p.caseManagementApi.Configuration
	path := cfg.BasePath + "/api/v2/casemanagement/caseplans/{caseplanId}/dataschemas"
	path = strings.Replace(path, "{caseplanId}", url.PathEscape(fmt.Sprintf("%v", caseplanId)), -1)

	headerParams := casemanagementAPIHeaders(cfg)
	headerParams["Content-Type"] = cfg.APIClient.SelectHeaderContentType([]string{"application/json"})
	if accept := cfg.APIClient.SelectHeaderAccept([]string{"application/json"}); accept != "" {
		headerParams["Accept"] = accept
	}
	postBody := &body
	response, err := cfg.APIClient.CallAPI(path, "POST", postBody, headerParams, map[string]string{}, url.Values{}, "", nil, "other")
	if err != nil {
		return nil, response, err
	}
	if response.Error != nil {
		return nil, response, errors.New(response.ErrorMessage)
	}
	var successPayload *platformclientv2.Caseplandataschema
	if response.HasBody && len(response.RawBody) > 0 {
		if err = json.Unmarshal(response.RawBody, &successPayload); err != nil {
			return nil, response, err
		}
	}
	return successPayload, response, nil
}

// putCaseManagementCaseplanDataschemaFn calls PUT /api/v2/casemanagement/caseplans/{caseplanId}/dataschemas/{schemaKeyName}.
// Not generated in platform-client-sdk-go v186. Today schemaKeyName is always "default".
func putCaseManagementCaseplanDataschemaFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, schemaKeyName string, body platformclientv2.Caseplandataschema) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error) {
	cfg := p.caseManagementApi.Configuration
	path := cfg.BasePath + "/api/v2/casemanagement/caseplans/{caseplanId}/dataschemas/{schemaKeyName}"
	path = strings.Replace(path, "{caseplanId}", url.PathEscape(fmt.Sprintf("%v", caseplanId)), -1)
	path = strings.Replace(path, "{schemaKeyName}", url.PathEscape(fmt.Sprintf("%v", schemaKeyName)), -1)

	headerParams := casemanagementAPIHeaders(cfg)
	headerParams["Content-Type"] = cfg.APIClient.SelectHeaderContentType([]string{"application/json"})
	if accept := cfg.APIClient.SelectHeaderAccept([]string{"application/json"}); accept != "" {
		headerParams["Accept"] = accept
	}
	postBody := &body
	response, err := cfg.APIClient.CallAPI(path, "PUT", postBody, headerParams, map[string]string{}, url.Values{}, "", nil, "other")
	if err != nil {
		return nil, response, err
	}
	if response.Error != nil {
		return nil, response, errors.New(response.ErrorMessage)
	}
	var successPayload *platformclientv2.Caseplandataschema
	if response.HasBody && len(response.RawBody) > 0 {
		if err = json.Unmarshal(response.RawBody, &successPayload); err != nil {
			return nil, response, err
		}
	}
	return successPayload, response, nil
}

func postCaseManagementCaseplanVersionsFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.PostCasemanagementCaseplanVersions(caseplanId)
}

func deleteCaseManagementCaseplanDataschemaFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, schemaKeyName string) (*platformclientv2.APIResponse, error) {
	cfg := p.caseManagementApi.Configuration
	path := cfg.BasePath + "/api/v2/casemanagement/caseplans/{caseplanId}/dataschemas/{schemaKeyName}"
	path = strings.Replace(path, "{caseplanId}", url.PathEscape(fmt.Sprintf("%v", caseplanId)), -1)
	path = strings.Replace(path, "{schemaKeyName}", url.PathEscape(fmt.Sprintf("%v", schemaKeyName)), -1)

	headerParams := casemanagementAPIHeaders(cfg)
	if accept := cfg.APIClient.SelectHeaderAccept([]string{"application/json"}); accept != "" {
		headerParams["Accept"] = accept
	}
	response, err := cfg.APIClient.CallAPI(path, "DELETE", nil, headerParams, map[string]string{}, url.Values{}, "", nil, "other")
	if err != nil {
		return response, err
	}
	if response.Error != nil {
		return response, errors.New(response.ErrorMessage)
	}
	return response, nil
}

func casemanagementAPIHeaders(cfg *platformclientv2.Configuration) map[string]string {
	headerParams := make(map[string]string)
	if cfg.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + cfg.AccessToken
	}
	for key := range cfg.DefaultHeader {
		headerParams[key] = cfg.DefaultHeader[key]
	}
	return headerParams
}

func publishCaseManagementCaseplanFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.PostCasemanagementCaseplanPublish(caseplanId)
}

func deleteCaseManagementCaseplanFn(ctx context.Context, p *caseManagementCaseplanProxy, id string) (*platformclientv2.APIResponse, error) {
	_, resp, err := p.caseManagementApi.DeleteCasemanagementCaseplan(id)
	return resp, err
}
