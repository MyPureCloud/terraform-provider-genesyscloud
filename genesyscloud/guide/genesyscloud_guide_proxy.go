package guide

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
)

var internalProxy *guideProxy

type getAllGuidesFunc func(ctx context.Context, p *guideProxy, name string) (*[]Guide, *platformclientv2.APIResponse, error)
type createGuideFunc func(ctx context.Context, p *guideProxy, guide *CreateGuide) (*Guide, *platformclientv2.APIResponse, error)
type getGuideByIdFunc func(ctx context.Context, p *guideProxy, id string) (*Guide, *platformclientv2.APIResponse, error)
type getGuideByNameFunc func(ctx context.Context, p *guideProxy, name string) (string, bool, *platformclientv2.APIResponse, error)
type deleteGuideFunc func(ctx context.Context, p *guideProxy, id string) (*DeleteObjectJob, *platformclientv2.APIResponse, error)
type getDeleteJobStatusByIdFunc func(ctx context.Context, p *guideProxy, id string, guideId string) (*DeleteObjectJob, *platformclientv2.APIResponse, error)

type createGuideJobFunc func(ctx context.Context, p *guideProxy, guideJob *GenerateGuideContentRequest) (*JobResponse, *platformclientv2.APIResponse, error)
type getGuideJobByIdFunc func(ctx context.Context, p *guideProxy, id string) (*JobResponse, *platformclientv2.APIResponse, error)

// Guide Version
type createGuideVersionFunc func(ctx context.Context, p *guideProxy, guideVersion *CreateGuideVersionRequest, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error)

type guideProxy struct {
	clientConfig               *platformclientv2.Configuration
	getAllGuidesAttr           getAllGuidesFunc
	createGuideAttr            createGuideFunc
	getGuideByIdAttr           getGuideByIdFunc
	getGuideByNameAttr         getGuideByNameFunc
	deleteGuideAttr            deleteGuideFunc
	getDeleteJobStatusByIdAttr getDeleteJobStatusByIdFunc
	createGuideJobAttr         createGuideJobFunc
	getGuideJobByIdAttr        getGuideJobByIdFunc
	createGuideVersionAttr     createGuideVersionFunc
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
		createGuideJobAttr:         createGuideJobFn,
		getGuideJobByIdAttr:        getGuideJobByIdFn,
		createGuideVersionAttr:     createGuideVersionFn,
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
func (p *guideProxy) createGuideJob(ctx context.Context, guideJob *GenerateGuideContentRequest) (*JobResponse, *platformclientv2.APIResponse, error) {
	return p.createGuideJobAttr(ctx, p, guideJob)
}
func (p *guideProxy) getGuideJobById(ctx context.Context, id string) (*JobResponse, *platformclientv2.APIResponse, error) {
	return p.getGuideJobByIdAttr(ctx, p, id)
}
func (p *guideProxy) createGuideVersion(ctx context.Context, guideVersion *CreateGuideVersionRequest, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error) {
	return p.createGuideVersionAttr(ctx, p, guideVersion, guideId)
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

	// Create URL with query parameters
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

	req, err := http.NewRequest(action, u.String(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	req = setRequestHeader(req, p)

	body, resp, err := callAPI(ctx, client, req)
	if err != nil {
		return nil, resp, err
	}

	var guides GuideEntityListing
	if err := json.Unmarshal([]byte(body), &guides); err != nil {
		return nil, resp, fmt.Errorf("error unmarshaling response: %v", err)
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
		if err := json.Unmarshal([]byte(body), &respBody); err != nil {
			return nil, resp, fmt.Errorf("error unmarshaling response: %v", err)
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

// Create Functions

func createGuideFn(ctx context.Context, p *guideProxy, guide *CreateGuide) (*Guide, *platformclientv2.APIResponse, error) {
	return sdkPostGuide(ctx, p, guide)
}

func sdkPostGuide(ctx context.Context, p *guideProxy, body *CreateGuide) (*Guide, *platformclientv2.APIResponse, error) {
	client := &http.Client{}
	action := http.MethodPost
	baseURL := p.clientConfig.BasePath + "/api/v2/guides"

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling guide: %v", err)
	}

	req, err := http.NewRequest(action, baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	req = setRequestHeader(req, p)

	respBody, resp, err := callAPI(ctx, client, req)
	if err != nil {
		return nil, resp, err
	}

	var guide Guide
	if err := json.Unmarshal(respBody, &guide); err != nil {
		return nil, resp, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &guide, resp, nil
}

// Read Functions

func getGuideByIdFn(ctx context.Context, p *guideProxy, id string) (*Guide, *platformclientv2.APIResponse, error) {
	if guide := rc.GetCacheItem(p.guideCache, id); guide != nil {
		return guide, nil, nil
	}
	return sdkGetGuideById(ctx, p, id)
}

func sdkGetGuideById(ctx context.Context, p *guideProxy, id string) (*Guide, *platformclientv2.APIResponse, error) {
	client := &http.Client{}
	action := http.MethodGet
	baseURL := p.clientConfig.BasePath + "/api/v2/guides/" + id

	req, err := http.NewRequest(action, baseURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	req = setRequestHeader(req, p)

	respBody, resp, err := callAPI(ctx, client, req)
	if err != nil {
		return nil, resp, err
	}

	var guide Guide
	if err := json.Unmarshal(respBody, &guide); err != nil {
		return nil, resp, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &guide, resp, nil
}

// Get By Name Functions

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

// Delete Functions

func deleteGuideFn(ctx context.Context, p *guideProxy, id string) (*DeleteObjectJob, *platformclientv2.APIResponse, error) {
	return sdkDeleteGuide(ctx, p, id)
}

func sdkDeleteGuide(ctx context.Context, p *guideProxy, id string) (*DeleteObjectJob, *platformclientv2.APIResponse, error) {
	client := &http.Client{}
	action := http.MethodDelete
	baseURL := p.clientConfig.BasePath + "/api/v2/guides/" + id + "/jobs"

	req, err := http.NewRequest(action, baseURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	req = setRequestHeader(req, p)

	respBody, resp, err := callAPI(ctx, client, req)
	if err != nil {
		return nil, resp, err
	}

	var jobResponse DeleteObjectJob
	err = json.Unmarshal(respBody, &jobResponse)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshaling response: %v", err)
	}
	jobResponse.GuideId = id

	return &jobResponse, resp, nil
}

func getDeleteJobStatusByIdFn(ctx context.Context, p *guideProxy, jobId string, guideId string) (*DeleteObjectJob, *platformclientv2.APIResponse, error) {
	return sdkGetJobDeletionStatus(ctx, p, jobId, guideId)
}

func sdkGetJobDeletionStatus(ctx context.Context, p *guideProxy, jobId string, guideId string) (*DeleteObjectJob, *platformclientv2.APIResponse, error) {
	client := &http.Client{}
	action := http.MethodGet
	baseURL := p.clientConfig.BasePath + "/api/v2/guides/" + guideId + "/jobs/" + jobId

	req, err := http.NewRequest(action, baseURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	req = setRequestHeader(req, p)

	respBody, resp, err := callAPI(ctx, client, req)
	if err != nil {
		return nil, resp, err
	}

	var jobResponse DeleteObjectJob
	err = json.Unmarshal(respBody, &jobResponse)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshaling response: %v", err)
	}
	jobResponse.GuideId = guideId

	if jobResponse.Status == "Succeeded" {
		rc.DeleteCacheItem(p.guideCache, guideId)
	}

	return &jobResponse, resp, nil
}

// Create Job Functions

func createGuideJobFn(ctx context.Context, p *guideProxy, guideJob *GenerateGuideContentRequest) (*JobResponse, *platformclientv2.APIResponse, error) {
	return sdkCreateGuideJob(ctx, p, guideJob)
}

func sdkCreateGuideJob(ctx context.Context, p *guideProxy, guideJob *GenerateGuideContentRequest) (*JobResponse, *platformclientv2.APIResponse, error) {
	client := &http.Client{}
	action := http.MethodPost
	baseURL := p.clientConfig.BasePath + "/api/v2/guides/jobs"

	jsonBody, err := json.Marshal(guideJob)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling guide job: %v", err)
	}

	req, err := http.NewRequest(action, baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, nil, fmt.Errorf("error creating guide job request: %v", err)
	}

	req = setRequestHeader(req, p)

	respBody, resp, err := callAPI(ctx, client, req)
	if err != nil {
		return nil, resp, err
	}

	var job JobResponse
	err = json.Unmarshal(respBody, &job)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshaling guide job: %v", err)
	}

	return &job, resp, nil
}

// Get Job Functions

func getGuideJobByIdFn(ctx context.Context, p *guideProxy, id string) (*JobResponse, *platformclientv2.APIResponse, error) {
	return sdkGetGuideJobById(ctx, p, id)
}

func sdkGetGuideJobById(ctx context.Context, p *guideProxy, id string) (*JobResponse, *platformclientv2.APIResponse, error) {
	client := &http.Client{}
	action := http.MethodGet
	baseURL := p.clientConfig.BasePath + "/api/v2/guides/jobs/" + id

	req, err := http.NewRequest(action, baseURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	req = setRequestHeader(req, p)

	respBody, resp, err := callAPI(ctx, client, req)
	if err != nil {
		return nil, resp, err
	}

	var job JobResponse
	err = json.Unmarshal(respBody, &job)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshaling guide job: %v", err)
	}

	return &job, resp, nil
}

// Guide Version

func createGuideVersionFn(ctx context.Context, p *guideProxy, guideVersion *CreateGuideVersionRequest, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error) {
	return sdkPostGuideVersion(guideVersion, p, guideId)
}

func sdkPostGuideVersion(body *CreateGuideVersionRequest, p *guideProxy, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	action := http.MethodPost
	baseURL := p.clientConfig.BasePath + "/api/v2/guides/" + guideId + "/versions"

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling guide version: %v", err)
	}

	req, err := http.NewRequest(action, baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	req = setRequestHeader(req, p)

	respBody, apiResp, err := callAPI(context.Background(), client, req)
	if err != nil {
		return nil, apiResp, fmt.Errorf("error calling API: %v", err)
	}

	var guideVersion VersionResponse
	if err := json.Unmarshal(respBody, &guideVersion); err != nil {
		return nil, apiResp, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &guideVersion, apiResp, nil
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
