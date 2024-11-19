package telephony_provider_edges_trunkbasesettings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

var internalProxy *trunkbaseSettingProxy
var trunkBaseCache = rc.NewResourceCache[platformclientv2.Trunkbase]()

// Type definitions for each func on our proxy so we can easily mock them out later
type getTrunkBaseSettingByIdFunc func(ctx context.Context, p *trunkbaseSettingProxy, id string) (*platformclientv2.Trunkbase, *platformclientv2.APIResponse, error)
type getAllTrunkBaseSettingsFunc func(ctx context.Context, p *trunkbaseSettingProxy, name string) (*[]platformclientv2.Trunkbase, *platformclientv2.APIResponse, error)
type createTrunkBaseSettingFunc func(ctx context.Context, p *trunkbaseSettingProxy, trunkBaseSetting platformclientv2.Trunkbase) (*platformclientv2.Trunkbase, *platformclientv2.APIResponse, error)
type updateTrunkBaseSettingFunc func(ctx context.Context, p *trunkbaseSettingProxy, id string, trunkBaseSetting platformclientv2.Trunkbase) (*platformclientv2.Trunkbase, *platformclientv2.APIResponse, error)
type deleteTrunkBaseSettingFunc func(ctx context.Context, p *trunkbaseSettingProxy, id string) (*platformclientv2.APIResponse, error)

type trunkbaseSettingProxy struct {
	clientConfig *platformclientv2.Configuration
	edgesApi     *platformclientv2.TelephonyProvidersEdgeApi

	getTrunkBaseSettingByIdAttr getTrunkBaseSettingByIdFunc
	getAllTrunkBaseSettingsAttr getAllTrunkBaseSettingsFunc
	createTrunkBaseSettingAttr  createTrunkBaseSettingFunc
	updateTrunkBaseSettingAttr  updateTrunkBaseSettingFunc
	deleteTrunkBaseSettingAttr  deleteTrunkBaseSettingFunc
	trunkBaseCache              rc.CacheInterface[platformclientv2.Trunkbase]
}

// initializes the  proxy with all of the data needed to communicate with Genesys Cloud
func newTrunkBaseSettingProxy(clientConfig *platformclientv2.Configuration) *trunkbaseSettingProxy {
	edgesApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(clientConfig)
	return &trunkbaseSettingProxy{
		clientConfig:                clientConfig,
		edgesApi:                    edgesApi,
		getTrunkBaseSettingByIdAttr: getTrunkBaseSettingByIdFn,
		createTrunkBaseSettingAttr:  createTrunkBaseSettingFn,
		updateTrunkBaseSettingAttr:  updateTrunkBaseSettingFn,
		deleteTrunkBaseSettingAttr:  deleteTrunkBaseSettingFn,
		getAllTrunkBaseSettingsAttr: getAllTrunkBaseSettingsFn,
		trunkBaseCache:              trunkBaseCache,
	}
}

//	getTrunkBaseSettingProxy acts as a singleton to for the internalProxy.  It also ensures
//
// that we can still proxy our tests by directly setting internalProxy package variable
func getTrunkBaseSettingProxy(clientConfig *platformclientv2.Configuration) *trunkbaseSettingProxy {
	if internalProxy == nil {
		internalProxy = newTrunkBaseSettingProxy(clientConfig)
	}

	return internalProxy
}

func (p *trunkbaseSettingProxy) GetTrunkBaseSettingById(ctx context.Context, trunkBaseSettingId string) (*platformclientv2.Trunkbase, *platformclientv2.APIResponse, error) {
	return p.getTrunkBaseSettingByIdAttr(ctx, p, trunkBaseSettingId)
}

func (p *trunkbaseSettingProxy) GetAllTrunkBaseSetting(ctx context.Context) (*[]platformclientv2.Trunkbase, *platformclientv2.APIResponse, error) {
	return p.getAllTrunkBaseSettingsAttr(ctx, p, "")
}

func (p *trunkbaseSettingProxy) GetAllTrunkBaseSettingWithName(ctx context.Context, name string) (*[]platformclientv2.Trunkbase, *platformclientv2.APIResponse, error) {
	return p.getAllTrunkBaseSettingsAttr(ctx, p, name)
}
func (p *trunkbaseSettingProxy) CreateTrunkBaseSetting(ctx context.Context, trunkBaseSetting platformclientv2.Trunkbase) (*platformclientv2.Trunkbase, *platformclientv2.APIResponse, error) {
	return p.createTrunkBaseSettingAttr(ctx, p, trunkBaseSetting)
}

func (p *trunkbaseSettingProxy) UpdateTrunkBaseSetting(ctx context.Context, trunkbaseSettingId string, trunkBaseSetting platformclientv2.Trunkbase) (*platformclientv2.Trunkbase, *platformclientv2.APIResponse, error) {
	return p.updateTrunkBaseSettingAttr(ctx, p, trunkbaseSettingId, trunkBaseSetting)
}

func (p *trunkbaseSettingProxy) DeleteTrunkBaseSetting(ctx context.Context, trunkbaseSettingId string) (*platformclientv2.APIResponse, error) {
	rc.DeleteCacheItem(p.trunkBaseCache, trunkbaseSettingId)
	return p.deleteTrunkBaseSettingAttr(ctx, p, trunkbaseSettingId)
}

func getTrunkBaseSettingByIdFn(ctx context.Context, p *trunkbaseSettingProxy, trunkBaseSettingId string) (*platformclientv2.Trunkbase, *platformclientv2.APIResponse, error) {
	tb := rc.GetCacheItem(p.trunkBaseCache, trunkBaseSettingId)
	if tb != nil {
		return tb, nil, nil
	}

	tb, resp, err := p.edgesApi.GetTelephonyProvidersEdgesTrunkbasesetting(trunkBaseSettingId, true)
	return tb, resp, err
}

func getAllTrunkBaseSettingsFn(ctx context.Context, p *trunkbaseSettingProxy, name string) (*[]platformclientv2.Trunkbase, *platformclientv2.APIResponse, error) {
	trunkbaseSlice := make([]platformclientv2.Trunkbase, 0, 10)
	util.WithRetries(ctx, 5*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100.
			trunkBaseSettings, resp, getErr := getTelephonyProvidersEdgesTrunkbasesettings(p, pageNum, pageSize, name)

			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error requesting trunk base settings %s | error: %s", name, getErr), resp))
			}

			if trunkBaseSettings.Entities == nil || len(*trunkBaseSettings.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No trunkBaseSettings found with name %s", name), resp))
			}

			for _, trunkBaseSetting := range *trunkBaseSettings.Entities {
				if trunkBaseSetting.State != nil && *trunkBaseSetting.State != "deleted" {
					if name == "" {
						rc.SetCache(p.trunkBaseCache, *trunkBaseSetting.Id, trunkBaseSetting)
					}

					trunkbaseSlice = append(trunkbaseSlice, trunkBaseSetting)
				}
			}

			if name == "" {
				break
			}

		}
		return nil
	})

	return &trunkbaseSlice, nil, nil
}

// The SDK function is too cumbersome because of the various boolean query parameters.
// This function was written in order to leave them out and make a single API call
func getTelephonyProvidersEdgesTrunkbasesettings(p *trunkbaseSettingProxy, pageNumber int, pageSize int, name string) (*platformclientv2.Trunkbaseentitylisting, *platformclientv2.APIResponse, error) {
	headerParams := make(map[string]string)
	if p.clientConfig.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + p.clientConfig.AccessToken
	}
	// add default headers if any
	for key := range p.clientConfig.DefaultHeader {
		headerParams[key] = p.clientConfig.DefaultHeader[key]
	}

	queryParams := make(map[string]string)
	queryParams["pageNumber"] = p.clientConfig.APIClient.ParameterToString(pageNumber, "")
	queryParams["pageSize"] = p.clientConfig.APIClient.ParameterToString(pageSize, "")
	if name != "" {
		queryParams["name"] = p.clientConfig.APIClient.ParameterToString(name, "")
	}

	// to determine the Content-Type header
	httpContentTypes := []string{"application/json"}

	// set Content-Type header
	httpContentType := p.clientConfig.APIClient.SelectHeaderContentType(httpContentTypes)
	if httpContentType != "" {
		headerParams["Content-Type"] = httpContentType
	}

	// set Accept header
	httpHeaderAccept := p.clientConfig.APIClient.SelectHeaderAccept([]string{
		"application/json",
	})
	if httpHeaderAccept != "" {
		headerParams["Accept"] = httpHeaderAccept
	}
	var successPayload *platformclientv2.Trunkbaseentitylisting
	path := p.clientConfig.BasePath + "/api/v2/telephony/providers/edges/trunkbasesettings"
	response, err := p.clientConfig.APIClient.CallAPI(path, http.MethodGet, nil, headerParams, queryParams, nil, "", nil, "")
	if err != nil {
		return nil, nil, err
	}

	if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		err = json.Unmarshal(response.RawBody, &successPayload)
	}

	return successPayload, response, err
}

func createTrunkBaseSettingFn(ctx context.Context, p *trunkbaseSettingProxy, trunkBaseSetting platformclientv2.Trunkbase) (*platformclientv2.Trunkbase, *platformclientv2.APIResponse, error) {
	return p.edgesApi.PostTelephonyProvidersEdgesTrunkbasesettings(trunkBaseSetting)
}

func updateTrunkBaseSettingFn(ctx context.Context, p *trunkbaseSettingProxy, trunkbaseSettingId string, trunkBaseSetting platformclientv2.Trunkbase) (*platformclientv2.Trunkbase, *platformclientv2.APIResponse, error) {
	return p.edgesApi.PutTelephonyProvidersEdgesTrunkbasesetting(trunkbaseSettingId, trunkBaseSetting)
}

func deleteTrunkBaseSettingFn(ctx context.Context, p *trunkbaseSettingProxy, trunkBaseSettingId string) (*platformclientv2.APIResponse, error) {
	return p.edgesApi.DeleteTelephonyProvidersEdgesTrunkbasesetting(trunkBaseSettingId)
}
