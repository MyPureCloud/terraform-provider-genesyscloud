package case_management_caseplan

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
The genesyscloud_case_management_caseplan_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on our proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *caseManagementCaseplanProxy

// caseplanDataschemaKeyDefault is the only schema key name supported today for PUT
// /api/v2/casemanagement/caseplans/{caseplanId}/dataschemas/{schemaKeyName}.
const caseplanDataschemaKeyDefault = "default"

// caseplanAPIVersionLatest addresses the newest version (the draft when one exists).
const caseplanAPIVersionLatest = "latest"

// stageplanPageSize must hold every stageplan in one page: the API applies process-flow ordering per page.
const stageplanPageSize = "25"

// Type definitions for each func on our proxy so we can easily mock them out later
type createCaseManagementCaseplanFunc func(ctx context.Context, p *caseManagementCaseplanProxy, body *platformclientv2.Caseplancreate) (*platformclientv2.Caseplancreateresponse, *platformclientv2.APIResponse, error)
type getAllCaseManagementCaseplanFunc func(ctx context.Context, p *caseManagementCaseplanProxy) (*[]platformclientv2.Caseplan, *platformclientv2.APIResponse, error)
type getCaseManagementCaseplanIdByNameFunc func(ctx context.Context, p *caseManagementCaseplanProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type getCaseManagementCaseplanByIdFunc func(ctx context.Context, p *caseManagementCaseplanProxy, id string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error)
type getCaseManagementCaseplanVersionFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, versionId string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error)
type getCaseManagementCaseplanVersionDataschemasFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, versionId string) (*platformclientv2.Caseplandataschemalisting, *platformclientv2.APIResponse, error)
type getCaseManagementCaseplanVersionIntakesettingsFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, versionId string) (*platformclientv2.Intakesettingslisting, *platformclientv2.APIResponse, error)
type putCaseManagementCaseplanIntakesettingsFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, body platformclientv2.Intakesettingsupdate) (*platformclientv2.Intakesettingslisting, *platformclientv2.APIResponse, error)
type patchCaseManagementCaseplanFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, body platformclientv2.Caseplanupdate) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error)
type postCaseManagementCaseplanDataschemaFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, schemaId string) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error)
type putCaseManagementCaseplanDataschemaFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, schemaKeyName string, schemaId string) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error)
type postCaseManagementCaseplanVersionsFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error)
type publishCaseManagementCaseplanFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error)
type deleteCaseManagementCaseplanFunc func(ctx context.Context, p *caseManagementCaseplanProxy, id string) (*platformclientv2.APIResponse, error)
type listStageplansFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, versionId string) ([]platformclientv2.Stageplan, *platformclientv2.APIResponse, error)
type createStageplanFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, body platformclientv2.Stageplancreate) (*platformclientv2.Stageplan, *platformclientv2.APIResponse, error)
type patchStageplanFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, stageplanId string, body platformclientv2.Stageplanupdate) (*platformclientv2.Stageplan, *platformclientv2.APIResponse, error)
type deleteStageplanFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, stageplanId string) (*platformclientv2.APIResponse, error)
type repositionStageplanFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, stageplanId string, body platformclientv2.Stageplanreposition) (*platformclientv2.APIResponse, error)
type listStepplansFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, versionId string, stageplanId string) ([]platformclientv2.Stepplan, *platformclientv2.APIResponse, error)
type patchStepplanFunc func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, stageplanId string, stepplanId string, body platformclientv2.Stepplanupdate) (*platformclientv2.Stepplan, *platformclientv2.APIResponse, error)

// caseManagementCaseplanProxy contains all of the methods that call genesys cloud APIs.
type caseManagementCaseplanProxy struct {
	clientConfig                                       *platformclientv2.Configuration
	caseManagementApi                                  *platformclientv2.CaseManagementApi
	createCaseManagementCaseplanAttr                   createCaseManagementCaseplanFunc
	getAllCaseManagementCaseplanAttr                   getAllCaseManagementCaseplanFunc
	getCaseManagementCaseplanIdByNameAttr              getCaseManagementCaseplanIdByNameFunc
	getCaseManagementCaseplanByIdAttr                  getCaseManagementCaseplanByIdFunc
	getCaseManagementCaseplanVersionAttr               getCaseManagementCaseplanVersionFunc
	getCaseManagementCaseplanVersionDataschemasAttr    getCaseManagementCaseplanVersionDataschemasFunc
	getCaseManagementCaseplanVersionIntakesettingsAttr getCaseManagementCaseplanVersionIntakesettingsFunc
	putCaseManagementCaseplanIntakesettingsAttr        putCaseManagementCaseplanIntakesettingsFunc
	patchCaseManagementCaseplanAttr                    patchCaseManagementCaseplanFunc
	postCaseManagementCaseplanDataschemaAttr           postCaseManagementCaseplanDataschemaFunc
	putCaseManagementCaseplanDataschemaAttr            putCaseManagementCaseplanDataschemaFunc
	postCaseManagementCaseplanVersionsAttr             postCaseManagementCaseplanVersionsFunc
	publishCaseManagementCaseplanAttr                  publishCaseManagementCaseplanFunc
	deleteCaseManagementCaseplanAttr                   deleteCaseManagementCaseplanFunc
	listStageplansAttr                                 listStageplansFunc
	createStageplanAttr                                createStageplanFunc
	patchStageplanAttr                                 patchStageplanFunc
	deleteStageplanAttr                                deleteStageplanFunc
	repositionStageplanAttr                            repositionStageplanFunc
	listStepplansAttr                                  listStepplansFunc
	patchStepplanAttr                                  patchStepplanFunc
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
		getCaseManagementCaseplanVersionAttr:               getCaseManagementCaseplanVersionFn,
		getCaseManagementCaseplanVersionDataschemasAttr:    getCaseManagementCaseplanVersionDataschemasFn,
		getCaseManagementCaseplanVersionIntakesettingsAttr: getCaseManagementCaseplanVersionIntakesettingsFn,
		putCaseManagementCaseplanIntakesettingsAttr:        putCaseManagementCaseplanIntakesettingsFn,
		patchCaseManagementCaseplanAttr:                    patchCaseManagementCaseplanFn,
		postCaseManagementCaseplanDataschemaAttr:           postCaseManagementCaseplanDataschemaFn,
		putCaseManagementCaseplanDataschemaAttr:            putCaseManagementCaseplanDataschemaFn,
		postCaseManagementCaseplanVersionsAttr:             postCaseManagementCaseplanVersionsFn,
		publishCaseManagementCaseplanAttr:                  publishCaseManagementCaseplanFn,
		deleteCaseManagementCaseplanAttr:                   deleteCaseManagementCaseplanFn,
		listStageplansAttr:                                 listStageplansFn,
		createStageplanAttr:                                createStageplanFn,
		patchStageplanAttr:                                 patchStageplanFn,
		deleteStageplanAttr:                                deleteStageplanFn,
		repositionStageplanAttr:                            repositionStageplanFn,
		listStepplansAttr:                                  listStepplansFn,
		patchStepplanAttr:                                  patchStepplanFn,
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

// getCaseManagementCaseplanById returns the latest version of a caseplan, including latest/published version numbers
func (p *caseManagementCaseplanProxy) getCaseManagementCaseplanById(ctx context.Context, id string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
	return p.getCaseManagementCaseplanByIdAttr(ctx, p, id)
}

// getCaseManagementCaseplanVersion returns a caseplan version (number, "latest" or "published")
func (p *caseManagementCaseplanProxy) getCaseManagementCaseplanVersion(ctx context.Context, caseplanId string, versionId string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
	return p.getCaseManagementCaseplanVersionAttr(ctx, p, caseplanId, versionId)
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

func (p *caseManagementCaseplanProxy) postCaseManagementCaseplanDataschema(ctx context.Context, caseplanId string, schemaId string) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error) {
	return p.postCaseManagementCaseplanDataschemaAttr(ctx, p, caseplanId, schemaId)
}

func (p *caseManagementCaseplanProxy) putCaseManagementCaseplanDataschema(ctx context.Context, caseplanId string, schemaKeyName string, schemaId string) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error) {
	return p.putCaseManagementCaseplanDataschemaAttr(ctx, p, caseplanId, schemaKeyName, schemaId)
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

// listStageplans returns the stageplans of a caseplan version in process-flow order
func (p *caseManagementCaseplanProxy) listStageplans(ctx context.Context, caseplanId string, versionId string) ([]platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
	return p.listStageplansAttr(ctx, p, caseplanId, versionId)
}

func (p *caseManagementCaseplanProxy) createStageplan(ctx context.Context, caseplanId string, body platformclientv2.Stageplancreate) (*platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
	return p.createStageplanAttr(ctx, p, caseplanId, body)
}

func (p *caseManagementCaseplanProxy) patchStageplan(ctx context.Context, caseplanId string, stageplanId string, body platformclientv2.Stageplanupdate) (*platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
	return p.patchStageplanAttr(ctx, p, caseplanId, stageplanId, body)
}

func (p *caseManagementCaseplanProxy) deleteStageplan(ctx context.Context, caseplanId string, stageplanId string) (*platformclientv2.APIResponse, error) {
	return p.deleteStageplanAttr(ctx, p, caseplanId, stageplanId)
}

func (p *caseManagementCaseplanProxy) repositionStageplan(ctx context.Context, caseplanId string, stageplanId string, body platformclientv2.Stageplanreposition) (*platformclientv2.APIResponse, error) {
	return p.repositionStageplanAttr(ctx, p, caseplanId, stageplanId, body)
}

func (p *caseManagementCaseplanProxy) listStepplans(ctx context.Context, caseplanId string, versionId string, stageplanId string) ([]platformclientv2.Stepplan, *platformclientv2.APIResponse, error) {
	return p.listStepplansAttr(ctx, p, caseplanId, versionId, stageplanId)
}

func (p *caseManagementCaseplanProxy) patchStepplan(ctx context.Context, caseplanId string, stageplanId string, stepplanId string, body platformclientv2.Stepplanupdate) (*platformclientv2.Stepplan, *platformclientv2.APIResponse, error) {
	return p.patchStepplanAttr(ctx, p, caseplanId, stageplanId, stepplanId, body)
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

		if listing.NextUri == nil || *listing.NextUri == "" {
			break
		}
		nextAfter, err := util.GetQueryParamValueFromUri(*listing.NextUri, "after")
		if err != nil {
			return nil, resp, fmt.Errorf("unable to parse after cursor from caseplans next uri: %w", err)
		}
		if nextAfter == "" || nextAfter == after {
			break
		}
		after = nextAfter
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

func getCaseManagementCaseplanVersionFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, versionId string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.GetCasemanagementCaseplanVersion(caseplanId, versionId)
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

func postCaseManagementCaseplanDataschemaFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, schemaId string) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.PostCasemanagementCaseplanDataschemas(caseplanId, platformclientv2.Caseplandataschemarequest{Id: &schemaId})
}

func putCaseManagementCaseplanDataschemaFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, schemaKeyName string, schemaId string) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.PutCasemanagementCaseplanDataschema(caseplanId, schemaKeyName, platformclientv2.Caseplandataschemarequest{Id: &schemaId})
}

func postCaseManagementCaseplanVersionsFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.PostCasemanagementCaseplanVersions(caseplanId)
}

func publishCaseManagementCaseplanFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.PostCasemanagementCaseplanPublish(caseplanId)
}

func deleteCaseManagementCaseplanFn(ctx context.Context, p *caseManagementCaseplanProxy, id string) (*platformclientv2.APIResponse, error) {
	_, resp, err := p.caseManagementApi.DeleteCasemanagementCaseplan(id)
	return resp, err
}

func listStageplansFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, versionId string) ([]platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
	var all []platformclientv2.Stageplan
	after := ""
	for {
		listing, resp, err := p.caseManagementApi.GetCasemanagementCaseplanVersionStageplans(caseplanId, versionId, "", after, stageplanPageSize, nil)
		if err != nil {
			return nil, resp, err
		}
		if listing == nil || listing.Entities == nil {
			return all, resp, nil
		}
		all = append(all, *listing.Entities...)
		next, done, err := nextAfterCursor(listing.NextUri, after)
		if err != nil {
			return nil, resp, err
		}
		if done {
			return all, resp, nil
		}
		after = next
	}
}

func createStageplanFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, body platformclientv2.Stageplancreate) (*platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.PostCasemanagementCaseplanStageplans(caseplanId, body)
}

func patchStageplanFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, stageplanId string, body platformclientv2.Stageplanupdate) (*platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.PatchCasemanagementCaseplanStageplan(caseplanId, stageplanId, body)
}

func deleteStageplanFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, stageplanId string) (*platformclientv2.APIResponse, error) {
	_, resp, err := p.caseManagementApi.DeleteCasemanagementCaseplanStageplan(caseplanId, stageplanId)
	return resp, err
}

func repositionStageplanFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, stageplanId string, body platformclientv2.Stageplanreposition) (*platformclientv2.APIResponse, error) {
	_, resp, err := p.caseManagementApi.PostCasemanagementCaseplanStageplanReposition(caseplanId, stageplanId, body)
	return resp, err
}

func listStepplansFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, versionId string, stageplanId string) ([]platformclientv2.Stepplan, *platformclientv2.APIResponse, error) {
	var all []platformclientv2.Stepplan
	after := ""
	for {
		listing, resp, err := p.caseManagementApi.GetCasemanagementCaseplanVersionStageplanStepplans(caseplanId, versionId, stageplanId, "", after, stageplanPageSize, nil)
		if err != nil {
			return nil, resp, err
		}
		if listing == nil || listing.Entities == nil {
			return all, resp, nil
		}
		all = append(all, *listing.Entities...)
		next, done, err := nextAfterCursor(listing.NextUri, after)
		if err != nil {
			return nil, resp, err
		}
		if done {
			return all, resp, nil
		}
		after = next
	}
}

func patchStepplanFn(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, stageplanId string, stepplanId string, body platformclientv2.Stepplanupdate) (*platformclientv2.Stepplan, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.PatchCasemanagementCaseplanStageplanStepplan(caseplanId, stageplanId, stepplanId, body)
}

// nextAfterCursor extracts the "after" cursor from a listing nextUri. done is true when there is no further page.
func nextAfterCursor(nextUri *string, current string) (next string, done bool, err error) {
	if nextUri == nil || *nextUri == "" {
		return "", true, nil
	}
	next, err = util.GetQueryParamValueFromUri(*nextUri, "after")
	if err != nil {
		return "", true, fmt.Errorf("unable to parse after cursor from next uri: %w", err)
	}
	if next == "" || next == current {
		return "", true, nil
	}
	return next, false, nil
}
