package guide_version

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

var internalProxy *guideVersionProxy

type GetAllGuidesFunc func(ctx context.Context, p *guideVersionProxy) (*[]Guide, *platformclientv2.APIResponse, error)
type createGuideVersionFunc func(ctx context.Context, p *guideVersionProxy, guideVersion *CreateGuideVersionRequest, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error)
type getGuideVersionByIdFunc func(ctx context.Context, p *guideVersionProxy, id string, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error)
type updateGuideVersionFunc func(ctx context.Context, p *guideVersionProxy, id string, guideId string, guideVersion *UpdateGuideVersion) (*VersionResponse, *platformclientv2.APIResponse, error)
type publishGuideVersionFunc func(ctx context.Context, p *guideVersionProxy, body *GuideVersionPublishJobRequest) (*VersionJobResponse, *platformclientv2.APIResponse, error)
type getGuideVersionPublishJobStatusFunc func(ctx context.Context, p *guideVersionProxy, versionId, jobId, guideId string) (*VersionJobResponse, *platformclientv2.APIResponse, error)
type getGuideByIdFunc func(ctx context.Context, p *guideVersionProxy, id string) (*Guide, *platformclientv2.APIResponse, error)
type guideVersionProxy struct {
	clientConfig                        *platformclientv2.Configuration
	GetAllGuidesAttr                    GetAllGuidesFunc
	createGuideVersionAttr              createGuideVersionFunc
	getGuideVersionByIdAttr             getGuideVersionByIdFunc
	updateGuideVersionAttr              updateGuideVersionFunc
	publishGuideVersionAttr             publishGuideVersionFunc
	getGuideVersionPublishJobStatusAttr getGuideVersionPublishJobStatusFunc
	getGuideByIdAttr                    getGuideByIdFunc
}

func newGuideVersionProxy(clientConfig *platformclientv2.Configuration) *guideVersionProxy {
	return &guideVersionProxy{
		clientConfig:                        clientConfig,
		GetAllGuidesAttr:                    GetAllGuidesFn,
		createGuideVersionAttr:              createGuideVersionFn,
		getGuideVersionByIdAttr:             getGuideVersionByIdFn,
		updateGuideVersionAttr:              updateGuideVersionFn,
		publishGuideVersionAttr:             publishGuideVersionFn,
		getGuideVersionPublishJobStatusAttr: getGuideVersionPublishJobStatusFn,
		getGuideByIdAttr:                    getGuideByIdFn,
	}
}
func getGuideVersionProxy(clientConfig *platformclientv2.Configuration) *guideVersionProxy {
	if internalProxy == nil {
		internalProxy = newGuideVersionProxy(clientConfig)
	}
	return internalProxy
}

func (p *guideVersionProxy) GetAllGuides(ctx context.Context) (*[]Guide, *platformclientv2.APIResponse, error) {
	return p.GetAllGuidesAttr(ctx, p)
}
func (p *guideVersionProxy) createGuideVersion(ctx context.Context, guideVersion *CreateGuideVersionRequest, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error) {
	return p.createGuideVersionAttr(ctx, p, guideVersion, guideId)
}
func (p *guideVersionProxy) getGuideVersionById(ctx context.Context, id string, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error) {
	return p.getGuideVersionByIdAttr(ctx, p, id, guideId)
}
func (p *guideVersionProxy) updateGuideVersion(ctx context.Context, id string, guideId string, guideVersion *UpdateGuideVersion) (*VersionResponse, *platformclientv2.APIResponse, error) {
	return p.updateGuideVersionAttr(ctx, p, id, guideId, guideVersion)
}
func (p *guideVersionProxy) publishGuideVersion(ctx context.Context, body *GuideVersionPublishJobRequest) (*VersionJobResponse, *platformclientv2.APIResponse, error) {
	return p.publishGuideVersionAttr(ctx, p, body)
}
func (p *guideVersionProxy) getGuideVersionPublishJobStatus(ctx context.Context, versionId, jobId, guideId string) (*VersionJobResponse, *platformclientv2.APIResponse, error) {
	return p.getGuideVersionPublishJobStatusAttr(ctx, p, versionId, jobId, guideId)
}
func (p *guideVersionProxy) getGuideById(ctx context.Context, id string) (*Guide, *platformclientv2.APIResponse, error) {
	return p.getGuideByIdAttr(ctx, p, id)
}

// GetAll Functions

func GetAllGuidesFn(ctx context.Context, p *guideVersionProxy) (*[]Guide, *platformclientv2.APIResponse, error) {
	return sdkGetAllGuides(ctx, p)
}

func sdkGetAllGuides(_ context.Context, p *guideVersionProxy) (*[]Guide, *platformclientv2.APIResponse, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	action := http.MethodGet
	baseURL := p.clientConfig.BasePath + "/api/v2/guides"
	var allGuides []Guide

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing URL: %v", err)
	}

	q := u.Query()
	q.Add("pageSize", "100")
	q.Add("pageNumber", "1")

	u.RawQuery = q.Encode()

	req, err := http.NewRequest(action, u.String(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	req = buildRequestHeader(req, p)

	body, resp, err := callAPI(client, req)
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

	if guides.PageCount != nil && *guides.PageCount > 1 {
		for pageNum := 2; pageNum <= *guides.PageCount; pageNum++ {
			q.Set("pageNumber", fmt.Sprintf("%v", pageNum))
			req.URL.RawQuery = q.Encode()

			body, resp, err = callAPI(client, req)
			if err != nil {
				return nil, resp, fmt.Errorf("error fetching page %d: %v", pageNum, err)
			}

			var respBody GuideEntityListing
			if err := json.Unmarshal([]byte(body), &respBody); err != nil {
				return nil, resp, fmt.Errorf("error unmarshaling response for page %d: %v", pageNum, err)
			}

			if respBody.Entities != nil {
				allGuides = append(allGuides, *respBody.Entities...)
			}
		}
	}

	log.Printf("Successfully retrieved %d guides", len(allGuides))
	return &allGuides, resp, nil
}

// Create Functions

func createGuideVersionFn(ctx context.Context, p *guideVersionProxy, guideVersion *CreateGuideVersionRequest, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error) {
	return sdkPostGuideVersion(guideVersion, p, guideId)
}

func sdkPostGuideVersion(body *CreateGuideVersionRequest, p *guideVersionProxy, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error) {
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

	req = buildRequestHeader(req, p)

	respBody, apiResp, err := callAPI(client, req)
	if err != nil {
		return nil, apiResp, fmt.Errorf("error calling API: %v", err)
	}

	var guideVersion VersionResponse
	if err := json.Unmarshal(respBody, &guideVersion); err != nil {
		return nil, apiResp, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &guideVersion, apiResp, nil
}

// Read Functions

func getGuideVersionByIdFn(ctx context.Context, p *guideVersionProxy, id string, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error) {
	return sdkGetGuideVersionById(p, id, guideId)
}

func sdkGetGuideVersionById(p *guideVersionProxy, id string, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	action := http.MethodGet
	baseURL := p.clientConfig.BasePath + "/api/v2/guides/" + guideId + "/versions/" + id

	req, err := http.NewRequest(action, baseURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	req = buildRequestHeader(req, p)

	respBody, apiResp, err := callAPI(client, req)
	if err != nil {
		return nil, apiResp, fmt.Errorf("error calling API: %v", err)
	}

	var guideVersion VersionResponse
	if err := json.Unmarshal(respBody, &guideVersion); err != nil {
		return nil, apiResp, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &guideVersion, apiResp, nil
}

// Update Functions

func updateGuideVersionFn(ctx context.Context, p *guideVersionProxy, id string, guideId string, guideVersion *UpdateGuideVersion) (*VersionResponse, *platformclientv2.APIResponse, error) {
	return sdkUpdateGuideVersion(p, id, guideId, guideVersion)
}

func sdkUpdateGuideVersion(p *guideVersionProxy, id string, guideId string, body *UpdateGuideVersion) (*VersionResponse, *platformclientv2.APIResponse, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	action := http.MethodPatch
	baseURL := p.clientConfig.BasePath + "/api/v2/guides/" + guideId + "/versions/" + id

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling guide version: %v", err)
	}

	req, err := http.NewRequest(action, baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	req = buildRequestHeader(req, p)

	respBody, apiResp, err := callAPI(client, req)
	if err != nil {
		return nil, apiResp, fmt.Errorf("error calling API: %v", err)
	}

	var guideVersion VersionResponse
	if err := json.Unmarshal(respBody, &guideVersion); err != nil {
		return nil, apiResp, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &guideVersion, apiResp, nil
}

// Helper api call function to be removed once endpoints are public
func callAPI(client *http.Client, req *http.Request) ([]byte, *platformclientv2.APIResponse, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading response: %v", err)
	}

	apiResp := &platformclientv2.APIResponse{
		StatusCode: resp.StatusCode,
		Response:   resp,
	}

	if resp.StatusCode >= 400 {
		return nil, apiResp, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, apiResp, nil
}

// Functions to publish the guide version

func publishGuideVersionFn(ctx context.Context, p *guideVersionProxy, body *GuideVersionPublishJobRequest) (*VersionJobResponse, *platformclientv2.APIResponse, error) {
	return sdkPublishGuideVersion(p, body)
}

func sdkPublishGuideVersion(p *guideVersionProxy, body *GuideVersionPublishJobRequest) (*VersionJobResponse, *platformclientv2.APIResponse, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	action := http.MethodPost
	baseURL := p.clientConfig.BasePath + "/api/v2/guides/" + body.GuideId + "/versions/" + body.VersionId + "/jobs"

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling guide version: %v", err)
	}

	req, err := http.NewRequest(action, baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	req = buildRequestHeader(req, p)

	respBody, apiResp, err := callAPI(client, req)
	if err != nil {
		return nil, apiResp, fmt.Errorf("error calling API: %v", err)
	}

	if apiResp.StatusCode == 202 {
		if len(respBody) == 0 {
			log.Println("Received 202 with empty body - job started successfully")
			return nil, apiResp, nil
		}
	}

	if len(respBody) == 0 {
		return nil, apiResp, fmt.Errorf("empty response body")
	}

	var jobResponse VersionJobResponse
	if err := json.Unmarshal(respBody, &jobResponse); err != nil {
		return nil, apiResp, fmt.Errorf("error unmarshaling response: %v, body: %s", err, string(respBody))
	}

	return &jobResponse, apiResp, nil
}

func getGuideVersionPublishJobStatusFn(ctx context.Context, p *guideVersionProxy, versionId, jobId, guideId string) (*VersionJobResponse, *platformclientv2.APIResponse, error) {
	return sdkGetGuideVersionPublishJobStatus(p, versionId, jobId, guideId)
}

func sdkGetGuideVersionPublishJobStatus(p *guideVersionProxy, versionId, jobId string, guideId string) (*VersionJobResponse, *platformclientv2.APIResponse, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	action := http.MethodGet
	baseURL := p.clientConfig.BasePath + "/api/v2/guides/" + guideId + "/versions/" + versionId + "/jobs/" + jobId

	req, err := http.NewRequest(action, baseURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	req = buildRequestHeader(req, p)

	respBody, apiResp, err := callAPI(client, req)
	if err != nil {
		return nil, apiResp, fmt.Errorf("error calling API: %v", err)
	}

	var guideVersion VersionJobResponse
	if err := json.Unmarshal(respBody, &guideVersion); err != nil {
		return nil, apiResp, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &guideVersion, apiResp, nil
}

func getGuideByIdFn(ctx context.Context, p *guideVersionProxy, id string) (*Guide, *platformclientv2.APIResponse, error) {
	return sdkGetGuideById(p, id)
}
func sdkGetGuideById(p *guideVersionProxy, id string) (*Guide, *platformclientv2.APIResponse, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	action := http.MethodGet
	baseURL := p.clientConfig.BasePath + "/api/v2/guides/" + id

	req, err := http.NewRequest(action, baseURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	req = buildRequestHeader(req, p)

	respBody, apiResp, err := callAPI(client, req)
	if err != nil {
		return nil, apiResp, fmt.Errorf("error calling API: %v", err)
	}

	var guide Guide
	if err := json.Unmarshal(respBody, &guide); err != nil {
		return nil, apiResp, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &guide, apiResp, nil
}
