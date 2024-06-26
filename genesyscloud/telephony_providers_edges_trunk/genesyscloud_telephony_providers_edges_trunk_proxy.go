package telephony_providers_edges_trunk

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

//generate a proxy for telephony_providers_edges_trunk

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *trunkProxy

// Type definitions for each func on our proxy so we can easily mock them out later

type getTrunkByIdFunc func(ctx context.Context, p *trunkProxy, id string) (*platformclientv2.Trunk, *platformclientv2.APIResponse, error)
type getAllTrunksFunc func(ctx context.Context, p *trunkProxy, pageNum int, pageSize int) (*platformclientv2.Trunkentitylisting, *platformclientv2.APIResponse, error)
type getTrunkBaseSettingsFunc func(ctx context.Context, p *trunkProxy, trunkBaseSettingsId string) (*platformclientv2.Trunkbase, *platformclientv2.APIResponse, error)
type getEdgeFunc func(ctx context.Context, p *trunkProxy, edgeId string) (*platformclientv2.Edge, *platformclientv2.APIResponse, error)
type putEdgeFunc func(ctx context.Context, p *trunkProxy, edgeId string, edge platformclientv2.Edge) (*platformclientv2.Edge, *platformclientv2.APIResponse, error)
type getEdgeGroupFunc func(ctx context.Context, p *trunkProxy, edgeGroupId string) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error)
type putEdgeGroupFunc func(ctx context.Context, p *trunkProxy, edgeGroupId string, edgeGroup platformclientv2.Edgegroup) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error)

// Proxy contains all of the methods that call genesys cloud APIs.
type trunkProxy struct {
	clientConfig *platformclientv2.Configuration
	edgesApi     *platformclientv2.TelephonyProvidersEdgeApi

	getTrunkByIdAttr         getTrunkByIdFunc
	getAllTrunksAttr         getAllTrunksFunc
	getTrunkBaseSettingsAttr getTrunkBaseSettingsFunc
	getEdgeAttr              getEdgeFunc
	putEdgeAttr              putEdgeFunc
	getEdgeGroupAttr         getEdgeGroupFunc
	putEdgeGroupAttr         putEdgeGroupFunc
}

// initializes the  proxy with all of the data needed to communicate with Genesys Cloud
func newTrunkProxy(clientConfig *platformclientv2.Configuration) *trunkProxy {
	edgesApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(clientConfig)
	return &trunkProxy{
		clientConfig:             clientConfig,
		edgesApi:                 edgesApi,
		getTrunkByIdAttr:         getTrunkByIdFn,
		getAllTrunksAttr:         getAllTrunksFn,
		getEdgeAttr:              getEdgeFn,
		putEdgeAttr:              putEdgeFn,
		getEdgeGroupAttr:         getEdgeGroupFn,
		putEdgeGroupAttr:         putEdgeGroupFn,
		getTrunkBaseSettingsAttr: getTrunkBaseSettingsFn,
	}
}

// getTeamProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getTrunkProxy(clientConfig *platformclientv2.Configuration) *trunkProxy {
	if internalProxy == nil {
		internalProxy = newTrunkProxy(clientConfig)
	}

	return internalProxy
}

func (p *trunkProxy) getEdge(ctx context.Context, edgeId string) (*platformclientv2.Edge, *platformclientv2.APIResponse, error) {
	return p.getEdgeAttr(ctx, p, edgeId)
}

func (p *trunkProxy) putEdge(ctx context.Context, edgeId string, edge platformclientv2.Edge) (*platformclientv2.Edge, *platformclientv2.APIResponse, error) {
	return p.putEdgeAttr(ctx, p, edgeId, edge)
}

func (p *trunkProxy) getEdgeGroup(ctx context.Context, edgeGroupId string) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
	return p.getEdgeGroupAttr(ctx, p, edgeGroupId)
}

func (p *trunkProxy) putEdgeGroup(ctx context.Context, edgeGroupId string, edgeGroup platformclientv2.Edgegroup) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
	return p.putEdgeGroupAttr(ctx, p, edgeGroupId, edgeGroup)
}

func (p *trunkProxy) getTrunkBaseSettings(ctx context.Context, trunkBaseSettingsId string) (*platformclientv2.Trunkbase, *platformclientv2.APIResponse, error) {
	return p.getTrunkBaseSettingsAttr(ctx, p, trunkBaseSettingsId)
}

func (p *trunkProxy) getTrunkById(ctx context.Context, id string) (*platformclientv2.Trunk, *platformclientv2.APIResponse, error) {
	return p.getTrunkByIdAttr(ctx, p, id)
}

func (p *trunkProxy) getAllTrunks(ctx context.Context, pageNum int, pageSize int) (*platformclientv2.Trunkentitylisting, *platformclientv2.APIResponse, error) {
	return p.getAllTrunksAttr(ctx, p, pageNum, pageSize)
}

func getEdgeFn(ctx context.Context, p *trunkProxy, edgeId string) (*platformclientv2.Edge, *platformclientv2.APIResponse, error) {
	return p.edgesApi.GetTelephonyProvidersEdge(edgeId, nil)
}

func putEdgeFn(ctx context.Context, p *trunkProxy, edgeId string, edge platformclientv2.Edge) (*platformclientv2.Edge, *platformclientv2.APIResponse, error) {
	return p.edgesApi.PutTelephonyProvidersEdge(edgeId, edge)
}

func getEdgeGroupFn(ctx context.Context, p *trunkProxy, edgeGroupId string) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
	return p.edgesApi.GetTelephonyProvidersEdgesEdgegroup(edgeGroupId, nil)
}

func putEdgeGroupFn(ctx context.Context, p *trunkProxy, edgeGroupId string, edgeGroup platformclientv2.Edgegroup) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
	return p.edgesApi.PutTelephonyProvidersEdgesEdgegroup(edgeGroupId, edgeGroup)
}

func getTrunkBaseSettingsFn(ctx context.Context, p *trunkProxy, trunkBaseSettingsId string) (*platformclientv2.Trunkbase, *platformclientv2.APIResponse, error) {
	return p.edgesApi.GetTelephonyProvidersEdgesTrunkbasesetting(trunkBaseSettingsId, true)
}

func getTrunkByIdFn(ctx context.Context, p *trunkProxy, trunkBaseSettingsId string) (*platformclientv2.Trunk, *platformclientv2.APIResponse, error) {
	return p.edgesApi.GetTelephonyProvidersEdgesTrunk(trunkBaseSettingsId)
}

func getAllTrunksFn(ctx context.Context, p *trunkProxy, pageNum int, pageSize int) (*platformclientv2.Trunkentitylisting, *platformclientv2.APIResponse, error) {

	return p.edgesApi.GetTelephonyProvidersEdgesTrunks(pageNum, pageSize, "", "", "", "", "")
}
