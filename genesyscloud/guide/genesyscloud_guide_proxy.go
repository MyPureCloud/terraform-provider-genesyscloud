package guide

import (
	"context"
	"fmt"

	customapi "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/custom_api_client"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
)

var internalProxy *guideProxy

type getAllGuidesFunc func(ctx context.Context, p *guideProxy, name string) (*[]Guide, *platformclientv2.APIResponse, error)
type createGuideFunc func(ctx context.Context, p *guideProxy, guide *CreateGuide) (*Guide, *platformclientv2.APIResponse, error)
type getGuideByIdFunc func(ctx context.Context, p *guideProxy, id string) (*Guide, *platformclientv2.APIResponse, error)
type getGuideByNameFunc func(ctx context.Context, p *guideProxy, name string) (string, bool, *platformclientv2.APIResponse, error)
type deleteGuideFunc func(ctx context.Context, p *guideProxy, id string) (*DeleteObjectJob, *platformclientv2.APIResponse, error)
type getDeleteJobStatusByIdFunc func(ctx context.Context, p *guideProxy, id string, guideId string) (*DeleteObjectJob, *platformclientv2.APIResponse, error)

type guideProxy struct {
	clientConfig               *platformclientv2.Configuration
	customApiClient            *customapi.Client
	getAllGuidesAttr           getAllGuidesFunc
	createGuideAttr            createGuideFunc
	getGuideByIdAttr           getGuideByIdFunc
	getGuideByNameAttr         getGuideByNameFunc
	deleteGuideAttr            deleteGuideFunc
	getDeleteJobStatusByIdAttr getDeleteJobStatusByIdFunc
	guideCache                 rc.CacheInterface[Guide]
}

func newGuideProxy(clientConfig *platformclientv2.Configuration) *guideProxy {
	guideCache := rc.NewResourceCache[Guide]()
	return &guideProxy{
		clientConfig:               clientConfig,
		customApiClient:            customapi.NewClient(clientConfig, ResourceType),
		getAllGuidesAttr:           getAllGuidesFn,
		createGuideAttr:            createGuideFn,
		getGuideByIdAttr:           getGuideByIdFn,
		getGuideByNameAttr:         getGuideByNameFn,
		deleteGuideAttr:            deleteGuideFn,
		getDeleteJobStatusByIdAttr: getDeleteJobStatusByIdFn,
		guideCache:                 guideCache,
	}
}
func getGuideProxy(clientConfig *platformclientv2.Configuration) *guideProxy {
	if internalProxy == nil {
		internalProxy = newGuideProxy(clientConfig)
	}
	return internalProxy
}

func (p *guideProxy) getAllGuides(ctx context.Context, name string) (*[]Guide, *platformclientv2.APIResponse, error) {
	return p.getAllGuidesAttr(ctx, p, name)
}
func (p *guideProxy) createGuide(ctx context.Context, guide *CreateGuide) (*Guide, *platformclientv2.APIResponse, error) {
	return p.createGuideAttr(ctx, p, guide)
}
func (p *guideProxy) getGuideById(ctx context.Context, id string) (*Guide, *platformclientv2.APIResponse, error) {
	return p.getGuideByIdAttr(ctx, p, id)
}
func (p *guideProxy) getGuideByName(ctx context.Context, name string) (string, bool, *platformclientv2.APIResponse, error) {
	return p.getGuideByNameAttr(ctx, p, name)
}
func (p *guideProxy) deleteGuide(ctx context.Context, id string) (*DeleteObjectJob, *platformclientv2.APIResponse, error) {
	return p.deleteGuideAttr(ctx, p, id)
}
func (p *guideProxy) getDeleteJobStatusById(ctx context.Context, id string, guideId string) (*DeleteObjectJob, *platformclientv2.APIResponse, error) {
	return p.getDeleteJobStatusByIdAttr(ctx, p, id, guideId)
}

// GetAll Functions

func getAllGuidesFn(ctx context.Context, p *guideProxy, name string) (*[]Guide, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	return sdkGetAllGuidesFn(ctx, p, name)
}

func sdkGetAllGuidesFn(ctx context.Context, p *guideProxy, name string) (*[]Guide, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	var allGuides []Guide
	queryParams := customapi.NewQueryParams(map[string]string{"pageSize": "100", "pageNumber": "1"})
	if name != "" {
		queryParams.Set("name", name)
	}

	guides, resp, err := customapi.Do[GuideEntityListing](ctx, p.customApiClient, customapi.MethodGet, "/api/v2/guides", nil, queryParams)
	if err != nil {
		return nil, resp, err
	}
	if guides.Entities == nil {
		return &allGuides, resp, nil
	}
	allGuides = append(allGuides, *guides.Entities...)

	for pageNum := 2; pageNum <= *guides.PageCount; pageNum++ {
		queryParams.Set("pageNumber", fmt.Sprintf("%v", pageNum))
		pageGuides, resp, err := customapi.Do[GuideEntityListing](ctx, p.customApiClient, customapi.MethodGet, "/api/v2/guides", nil, queryParams)
		if err != nil {
			return nil, resp, err
		}
		if pageGuides.Entities != nil {
			allGuides = append(allGuides, *pageGuides.Entities...)
		}
	}

	for _, guide := range allGuides {
		rc.SetCache(p.guideCache, *guide.Id, guide)
	}
	return &allGuides, resp, nil
}

func createGuideFn(ctx context.Context, p *guideProxy, guide *CreateGuide) (*Guide, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return customapi.Do[Guide](ctx, p.customApiClient, customapi.MethodPost, "/api/v2/guides", guide, nil)
}

func getGuideByIdFn(ctx context.Context, p *guideProxy, id string) (*Guide, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	if guide := rc.GetCacheItem(p.guideCache, id); guide != nil {
		return guide, nil, nil
	}
	return customapi.Do[Guide](ctx, p.customApiClient, customapi.MethodGet, "/api/v2/guides/"+id, nil, nil)
}

func getGuideByNameFn(ctx context.Context, p *guideProxy, name string) (string, bool, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	guides, resp, err := getAllGuidesFn(ctx, p, name)
	if err != nil {
		return "", false, resp, err
	}

	if guides == nil || len(*guides) == 0 {
		return "", true, resp, fmt.Errorf("no guide found with name: %s", name)
	}

	for _, guide := range *guides {
		if guide.Name != nil && *guide.Name == name {
			if guide.Id != nil {
				return *guide.Id, false, resp, nil
			}
			return "", false, resp, fmt.Errorf("guide found but has nil ID: %s", name)
		}
	}

	return "", false, resp, fmt.Errorf("unable to find guide with name %s", name)
}

func deleteGuideFn(ctx context.Context, p *guideProxy, id string) (*DeleteObjectJob, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	jobResponse, resp, err := customapi.Do[DeleteObjectJob](ctx, p.customApiClient, customapi.MethodDelete, "/api/v2/guides/"+id+"/jobs", nil, nil)
	if err != nil {
		return nil, resp, err
	}
	jobResponse.GuideId = id
	return jobResponse, resp, nil
}

func getDeleteJobStatusByIdFn(ctx context.Context, p *guideProxy, jobId string, guideId string) (*DeleteObjectJob, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	jobResponse, resp, err := customapi.Do[DeleteObjectJob](ctx, p.customApiClient, customapi.MethodGet, "/api/v2/guides/"+guideId+"/jobs/"+jobId, nil, nil)
	if err != nil {
		return nil, resp, err
	}
	jobResponse.GuideId = guideId
	if jobResponse.Status == "Succeeded" {
		rc.DeleteCacheItem(p.guideCache, guideId)
	}
	return jobResponse, resp, nil
}
