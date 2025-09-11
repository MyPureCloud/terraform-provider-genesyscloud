package guide

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"

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
	return sdkGetAllGuidesFn(ctx, p, name)
}

func sdkGetAllGuidesFn(ctx context.Context, p *guideProxy, name string) (*[]Guide, *platformclientv2.APIResponse, error) {
	client := &http.Client{}
	action := http.MethodGet
	baseURL := p.clientConfig.BasePath + "/api/v2/guides"
	var allGuides []Guide

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing URL: %v", err)
	}

	q := u.Query()
	if name != "" {
		q.Add("name", name)
	}
	q.Add("pageSize", "100")
	q.Add("pageNumber", "1")

	u.RawQuery = q.Encode()

	req, err := createHTTPRequest(action, u.String(), nil, p)
	if err != nil {
		return nil, nil, err
	}

	body, resp, err := callAPI(ctx, client, req)
	if err != nil {
		return nil, resp, err
	}

	var guides GuideEntityListing
	if err := unmarshalResponse(body, &guides); err != nil {
		return nil, resp, err
	}

	if guides.Entities == nil {
		return &allGuides, resp, nil
	}

	allGuides = append(allGuides, *guides.Entities...)

	for pageNum := 2; pageNum <= *guides.PageCount; pageNum++ {
		q.Set("pageNumber", fmt.Sprintf("%v", pageNum))
		req.URL.RawQuery = q.Encode()

		body, resp, err = callAPI(ctx, client, req)
		if err != nil {
			return nil, resp, err
		}

		var respBody GuideEntityListing
		if err := unmarshalResponse(body, &respBody); err != nil {
			return nil, resp, err
		}

		if respBody.Entities != nil {
			allGuides = append(allGuides, *respBody.Entities...)
		}
	}

	for _, guide := range allGuides {
		rc.SetCache(p.guideCache, *guide.Id, guide)
	}

	return &allGuides, resp, nil
}

func createGuideFn(ctx context.Context, p *guideProxy, guide *CreateGuide) (*Guide, *platformclientv2.APIResponse, error) {
	baseURL := p.clientConfig.BasePath + "/api/v2/guides"
	return makeAPIRequest[Guide](ctx, http.MethodPost, baseURL, guide, p)
}

func getGuideByIdFn(ctx context.Context, p *guideProxy, id string) (*Guide, *platformclientv2.APIResponse, error) {
	if guide := rc.GetCacheItem(p.guideCache, id); guide != nil {
		return guide, nil, nil
	}
	baseURL := p.clientConfig.BasePath + "/api/v2/guides/" + id
	return makeAPIRequest[Guide](ctx, http.MethodGet, baseURL, nil, p)
}

func getGuideByNameFn(ctx context.Context, p *guideProxy, name string) (string, bool, *platformclientv2.APIResponse, error) {
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
	baseURL := p.clientConfig.BasePath + "/api/v2/guides/" + id + "/jobs"
	jobResponse, resp, err := makeAPIRequest[DeleteObjectJob](ctx, http.MethodDelete, baseURL, nil, p)
	if err != nil {
		return nil, resp, err
	}
	jobResponse.GuideId = id
	return jobResponse, resp, nil
}

func getDeleteJobStatusByIdFn(ctx context.Context, p *guideProxy, jobId string, guideId string) (*DeleteObjectJob, *platformclientv2.APIResponse, error) {
	baseURL := p.clientConfig.BasePath + "/api/v2/guides/" + guideId + "/jobs/" + jobId
	jobResponse, resp, err := makeAPIRequest[DeleteObjectJob](ctx, http.MethodGet, baseURL, nil, p)
	if err != nil {
		return nil, resp, err
	}
	jobResponse.GuideId = guideId

	if jobResponse.Status == "Succeeded" {
		rc.DeleteCacheItem(p.guideCache, guideId)
	}

	return jobResponse, resp, nil
}

// makeAPIRequest performs a complete API request for any of the guide endpoints
func makeAPIRequest[T any](ctx context.Context, method, url string, requestBody interface{}, p *guideProxy) (*T, *platformclientv2.APIResponse, error) {
	var req *http.Request
	var err error

	if requestBody != nil {
		req, err = marshalAndCreateRequest(method, url, requestBody, p)
	} else {
		req, err = createHTTPRequest(method, url, nil, p)
	}

	if err != nil {
		return nil, nil, err
	}

	client := &http.Client{}
	respBody, resp, err := callAPI(ctx, client, req)
	if err != nil {
		return nil, resp, err
	}

	var result T
	if err := unmarshalResponse(respBody, &result); err != nil {
		return nil, resp, err
	}

	return &result, resp, nil
}

// callAPI is a helper function which will be removed when the endpoints are public
func callAPI(ctx context.Context, client *http.Client, req *http.Request) ([]byte, *platformclientv2.APIResponse, error) {
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading response: %v", err)
	}

	response := &platformclientv2.APIResponse{
		StatusCode: resp.StatusCode,
		Response:   resp,
	}

	if resp.StatusCode >= 400 {
		return nil, response, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, response, nil
}
