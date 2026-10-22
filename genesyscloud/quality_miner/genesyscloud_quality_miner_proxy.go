package quality_miner

import (
	"context"
	"fmt"
	"log"
	"net/url"

	customapi "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/custom_api_client"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

const minersBasePath = "/api/v2/languageunderstanding/miners"

var internalProxy *qualityMinerProxy

var qualityMinerCache = rc.NewResourceCache[platformclientv2.Miner]()

type getAllQualityMinersFunc func(ctx context.Context, p *qualityMinerProxy) (*[]platformclientv2.Miner, *platformclientv2.APIResponse, error)
type createQualityMinerFunc func(ctx context.Context, p *qualityMinerProxy, miner *platformclientv2.Miner) (*platformclientv2.Miner, *platformclientv2.APIResponse, error)
type getQualityMinerByIdFunc func(ctx context.Context, p *qualityMinerProxy, id string) (*platformclientv2.Miner, *platformclientv2.APIResponse, error)
type getQualityMinerIdByNameFunc func(ctx context.Context, p *qualityMinerProxy, name string, minerType string) (string, *platformclientv2.APIResponse, bool, error)
type deleteQualityMinerFunc func(ctx context.Context, p *qualityMinerProxy, id string) (*platformclientv2.APIResponse, error)

type qualityMinerProxy struct {
	clientConfig                *platformclientv2.Configuration
	languageUnderstandingApi    *platformclientv2.LanguageUnderstandingApi
	customApiClient             *customapi.Client
	getAllQualityMinersAttr     getAllQualityMinersFunc
	createQualityMinerAttr      createQualityMinerFunc
	getQualityMinerByIdAttr     getQualityMinerByIdFunc
	getQualityMinerIdByNameAttr getQualityMinerIdByNameFunc
	deleteQualityMinerAttr      deleteQualityMinerFunc
	qualityMinerCache           rc.CacheInterface[platformclientv2.Miner]
}

func newQualityMinerProxy(clientConfig *platformclientv2.Configuration) *qualityMinerProxy {
	api := platformclientv2.NewLanguageUnderstandingApiWithConfig(clientConfig)

	return &qualityMinerProxy{
		clientConfig:                clientConfig,
		languageUnderstandingApi:    api,
		customApiClient:             customapi.NewClient(clientConfig, ResourceType),
		getAllQualityMinersAttr:     getAllQualityMinersFn,
		createQualityMinerAttr:      createQualityMinerFn,
		getQualityMinerByIdAttr:     getQualityMinerByIdFn,
		getQualityMinerIdByNameAttr: getQualityMinerIdByNameFn,
		deleteQualityMinerAttr:      deleteQualityMinerFn,
		qualityMinerCache:           qualityMinerCache,
	}
}

func getQualityMinerProxy(clientConfig *platformclientv2.Configuration) *qualityMinerProxy {
	if internalProxy == nil {
		internalProxy = newQualityMinerProxy(clientConfig)
	}
	return internalProxy
}

// getAllQualityMiners retrieves all Genesys Cloud miners (across every type)
func (p *qualityMinerProxy) getAllQualityMiners(ctx context.Context) (*[]platformclientv2.Miner, *platformclientv2.APIResponse, error) {
	return p.getAllQualityMinersAttr(ctx, p)
}

func (p *qualityMinerProxy) createQualityMiner(ctx context.Context, miner *platformclientv2.Miner) (*platformclientv2.Miner, *platformclientv2.APIResponse, error) {
	return p.createQualityMinerAttr(ctx, p, miner)
}

// getQualityMinerById returns a single Genesys Cloud miner by Id
func (p *qualityMinerProxy) getQualityMinerById(ctx context.Context, id string) (*platformclientv2.Miner, *platformclientv2.APIResponse, error) {
	return p.getQualityMinerByIdAttr(ctx, p, id)
}

// getQualityMinerIdByName returns a single Genesys Cloud miner Id by name, optionally narrowed by type
func (p *qualityMinerProxy) getQualityMinerIdByName(ctx context.Context, name string, minerType string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getQualityMinerIdByNameAttr(ctx, p, name, minerType)
}

func (p *qualityMinerProxy) deleteQualityMiner(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteQualityMinerAttr(ctx, p, id)
}

func getAllQualityMinersFn(ctx context.Context, p *qualityMinerProxy) (*[]platformclientv2.Miner, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	var miners []platformclientv2.Miner
	var lastResp *platformclientv2.APIResponse

	path := minersBasePath + "?minerType="
	for path != "" {
		// The generated SDK list method takes no cursor param, so we follow NextUri via the custom API client to page through all miners.
		listing, resp, err := customapi.Do[platformclientv2.Minerlisting](ctx, p.customApiClient, customapi.MethodGet, path, nil, url.Values{})
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get miners: %v", err)
		}
		lastResp = resp

		if listing.Entities != nil {
			miners = append(miners, *listing.Entities...)
		}

		if listing.NextUri == nil || *listing.NextUri == "" {
			break
		}
		path = *listing.NextUri
	}

	for _, miner := range miners {
		if miner.Id != nil {
			rc.SetCache(p.qualityMinerCache, *miner.Id, miner)
		}
	}

	return &miners, lastResp, nil
}

// createQualityMinerFn is an implementation function for creating a Genesys Cloud miner
func createQualityMinerFn(ctx context.Context, p *qualityMinerProxy, miner *platformclientv2.Miner) (*platformclientv2.Miner, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	return p.languageUnderstandingApi.PostLanguageunderstandingMiners(*miner)
}

// getQualityMinerByIdFn is an implementation of the function to get a Genesys Cloud miner by Id
func getQualityMinerByIdFn(ctx context.Context, p *qualityMinerProxy, id string) (*platformclientv2.Miner, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	if miner := rc.GetCacheItem(p.qualityMinerCache, id); miner != nil {
		return miner, nil, nil
	}
	return p.languageUnderstandingApi.GetLanguageunderstandingMiner(id)
}

// getQualityMinerIdByNameFn finds a miner Id by name. When minerType is non-empty, only miners of that type
// are considered, which disambiguates miners of different types that share a name.
func getQualityMinerIdByNameFn(ctx context.Context, p *qualityMinerProxy, name string, minerType string) (string, *platformclientv2.APIResponse, bool, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	miners, resp, err := p.getAllQualityMiners(ctx)
	if err != nil {
		return "", resp, false, err
	}

	if miners == nil || len(*miners) == 0 {
		return "", resp, true, fmt.Errorf("no miner found with name %s", name)
	}

	for _, miner := range *miners {
		if miner.Name == nil || *miner.Name != name {
			continue
		}
		if minerType != "" && (miner.MinerType == nil || *miner.MinerType != minerType) {
			continue
		}
		log.Printf("Retrieved the miner id %s by name %s", *miner.Id, name)
		return *miner.Id, resp, false, nil
	}
	return "", resp, true, fmt.Errorf("unable to find miner with name %s", name)
}

// deleteQualityMinerFn is an implementation function for deleting a Genesys Cloud miner
func deleteQualityMinerFn(ctx context.Context, p *qualityMinerProxy, id string) (*platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	resp, err := p.languageUnderstandingApi.DeleteLanguageunderstandingMiner(id)
	if err != nil {
		return resp, err
	}
	rc.DeleteCacheItem(p.qualityMinerCache, id)
	return resp, nil
}
