package guide

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
	"io"
	"net/http"
	"net/url"
)

var internalProxy *guideProxy

type getAllGuidesFunc func(ctx context.Context, p *guideProxy, name string) (*GuideEntityListing, *platformclientv2.APIResponse, error)
type createGuideFunc func(ctx context.Context, p *guideProxy, guide *Createguide) (*Guide, *platformclientv2.APIResponse, error)
type getGuideByIdFunc func(ctx context.Context, p *guideProxy, id string) (*Guide, *platformclientv2.APIResponse, error)
type getGuideByNameFunc func(ctx context.Context, p *guideProxy, name string) (string, bool, *platformclientv2.APIResponse, error)
type deleteGuideFunc func(ctx context.Context, p *guideProxy, id string) (*platformclientv2.APIResponse, error)

type guideProxy struct {
	clientConfig       *platformclientv2.Configuration
	getAllGuidesAttr   getAllGuidesFunc
	createGuideAttr    createGuideFunc
	getGuideByIdAttr   getGuideByIdFunc
	getGuideByNameAttr getGuideByNameFunc
	deleteGuideAttr    deleteGuideFunc
}

func newGuideProxy(clientConfig *platformclientv2.Configuration) *guideProxy {
	return &guideProxy{
		clientConfig:       clientConfig,
		getAllGuidesAttr:   getAllGuidesFn,
		createGuideAttr:    createGuideFn,
		getGuideByIdAttr:   getGuideByIdFn,
		getGuideByNameAttr: getGuideByNameFn,
		deleteGuideAttr:    deleteGuideFn,
	}
}
func getGuideProxy(clientConfig *platformclientv2.Configuration) *guideProxy {
	if internalProxy == nil {
		internalProxy = newGuideProxy(clientConfig)
	}
	return internalProxy
}

func (p *guideProxy) getAllGuides(ctx context.Context, name string) (*GuideEntityListing, *platformclientv2.APIResponse, error) {
	return p.getAllGuidesAttr(ctx, p, name)
}
func (p *guideProxy) createGuide(ctx context.Context, guide *Createguide) (*Guide, *platformclientv2.APIResponse, error) {
	return p.createGuideAttr(ctx, p, guide)
}
func (p *guideProxy) getGuideById(ctx context.Context, id string) (*Guide, *platformclientv2.APIResponse, error) {
	return p.getGuideByIdAttr(ctx, p, id)
}
func (p *guideProxy) getGuideByName(ctx context.Context, name string) (string, bool, *platformclientv2.APIResponse, error) {
	return p.getGuideByNameAttr(ctx, p, name)
}
func (p *guideProxy) deleteGuide(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteGuideAttr(ctx, p, id)
}

// GetAll Functions

func getAllGuidesFn(ctx context.Context, p *guideProxy, name string) (*GuideEntityListing, *platformclientv2.APIResponse, error) {
	return sdkGetAllGuidesFn(ctx, p, name)
}

func sdkGetAllGuidesFn(_ context.Context, p *guideProxy, name string) (*GuideEntityListing, *platformclientv2.APIResponse, error) {
	client := &http.Client{}
	action := http.MethodGet
	baseURL := "https://api.inintca.com/api/v2/guides"

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

	u.RawQuery = q.Encode()

	req, err := http.NewRequest(action, u.String(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	req = setRequestHeader(req, p)

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
	if resp.StatusCode != 200 {
		return nil, apiResp, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var successPayload GuideEntityListing
	if err := json.Unmarshal([]byte(respBody), &successPayload); err != nil {
		return nil, apiResp, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &successPayload, apiResp, nil
}

// Create Functions

func createGuideFn(_ context.Context, p *guideProxy, guide *Createguide) (*Guide, *platformclientv2.APIResponse, error) {
	return sdkPostGuide(p, guide)
}

func sdkPostGuide(p *guideProxy, body *Createguide) (*Guide, *platformclientv2.APIResponse, error) {
	client := &http.Client{}
	action := http.MethodPost
	baseURL := "https://api.inintca.com/api/v2/guides"

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling guide: %v", err)
	}

	// Create request
	req, err := http.NewRequest(action, baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req = setRequestHeader(req, p)

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
	if resp.StatusCode != 200 {
		return nil, apiResp, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var guide Guide
	if err := json.Unmarshal(respBody, &guide); err != nil {
		return nil, apiResp, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &guide, apiResp, nil
}

// Read Functions

func getGuideByIdFn(_ context.Context, p *guideProxy, id string) (*Guide, *platformclientv2.APIResponse, error) {
	return sdkGetGuideById(p, id)
}

func sdkGetGuideById(p *guideProxy, id string) (*Guide, *platformclientv2.APIResponse, error) {
	client := &http.Client{}
	action := http.MethodGet
	baseURL := "https://api.inintca.com/api/v2/guides/" + id

	req, err := http.NewRequest(action, baseURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	req = setRequestHeader(req, p)

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
	if resp.StatusCode != 200 {
		return nil, apiResp, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var guide Guide
	if err := json.Unmarshal(respBody, &guide); err != nil {
		return nil, apiResp, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &guide, apiResp, nil
}

// Get By Name Functions

func getGuideByNameFn(ctx context.Context, p *guideProxy, name string) (string, bool, *platformclientv2.APIResponse, error) {
	guides, resp, err := getAllGuidesFn(ctx, p, name)
	if err != nil {
		return "", false, resp, err
	}

	if guides == nil || len(*guides.Entities) == 0 {
		return "", true, resp, fmt.Errorf("no guide found with name: %s", name)
	}

	for _, guide := range *guides.Entities {
		if *guide.Name == name {
			return *guide.Id, false, resp, nil
		}
	}

	return "", false, resp, fmt.Errorf("unable to find guide with name %s", name)
}

// Delete Functions

func deleteGuideFn(ctx context.Context, p *guideProxy, id string) (*platformclientv2.APIResponse, error) {
	return sdkDeleteGuide(p, id)
}

func sdkDeleteGuide(p *guideProxy, id string) (*platformclientv2.APIResponse, error) {
	client := &http.Client{}
	action := http.MethodDelete
	baseURL := "https://api.inintca.com/api/v2/guides/" + id + "/jobs"

	req, err := http.NewRequest(action, baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req = setRequestHeader(req, p)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	apiResp := &platformclientv2.APIResponse{
		StatusCode: resp.StatusCode,
		Response:   resp,
	}
	// Delete API Returns multiple successful StatusCodes
	// Check if returned code falls outside that scope
	if resp.StatusCode >= 400 {
		return apiResp, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return apiResp, nil
}
