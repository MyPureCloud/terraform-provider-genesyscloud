package guide_jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

var internalProxy *guideJobsProxy

type createGuideJobFunc func(ctx context.Context, p *guideJobsProxy, guideJob *GenerateGuideContentRequest) (*JobResponse, *platformclientv2.APIResponse, error)
type getGuideJobByIdFunc func(ctx context.Context, p *guideJobsProxy, id string) (*JobResponse, *platformclientv2.APIResponse, error)

type guideJobsProxy struct {
	clientConfig        *platformclientv2.Configuration
	createGuideJobAttr  createGuideJobFunc
	getGuideJobByIdAttr getGuideJobByIdFunc
}

func newGuideJobsProxy(clientConfig *platformclientv2.Configuration) *guideJobsProxy {
	return &guideJobsProxy{
		clientConfig:        clientConfig,
		createGuideJobAttr:  createGuideJobFn,
		getGuideJobByIdAttr: getGuideJobByIdFn,
	}
}

func getGuideJobsProxy(config *platformclientv2.Configuration) *guideJobsProxy {
	if internalProxy == nil {
		internalProxy = newGuideJobsProxy(config)
	}
	return internalProxy
}

func (p *guideJobsProxy) createGuideJob(ctx context.Context, guideJob *GenerateGuideContentRequest) (*JobResponse, *platformclientv2.APIResponse, error) {
	return p.createGuideJobAttr(ctx, p, guideJob)
}

func (p *guideJobsProxy) getGuideJobById(ctx context.Context, id string) (*JobResponse, *platformclientv2.APIResponse, error) {
	return p.getGuideJobByIdAttr(ctx, p, id)
}

// Create Functions

func createGuideJobFn(ctx context.Context, p *guideJobsProxy, guideJob *GenerateGuideContentRequest) (*JobResponse, *platformclientv2.APIResponse, error) {
	return sdkCreateGuideJob(ctx, p, guideJob)
}

func sdkCreateGuideJob(ctx context.Context, p *guideJobsProxy, guideJob *GenerateGuideContentRequest) (*JobResponse, *platformclientv2.APIResponse, error) {
	client := &http.Client{}
	action := http.MethodPost
	baseURL := p.clientConfig.BasePath + "/api/v2/guides/jobs"

	jsonBody, err := json.Marshal(guideJob)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling guide job | error: %w", err)
	}

	req, err := http.NewRequest(action, baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, nil, fmt.Errorf("error creating guide job request | error: %w", err)
	}

	req = setRequestHeader(req, p)

	respBody, resp, err := callAPI(ctx, client, req)
	if err != nil {
		return nil, resp, err
	}

	var job JobResponse
	err = json.Unmarshal(respBody, &job)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshaling guide job | error: %w", err)
	}

	return &job, resp, nil
}

// Read Functions

func getGuideJobByIdFn(ctx context.Context, p *guideJobsProxy, id string) (*JobResponse, *platformclientv2.APIResponse, error) {
	return sdkGetGuideJobById(ctx, p, id)
}

func sdkGetGuideJobById(ctx context.Context, p *guideJobsProxy, id string) (*JobResponse, *platformclientv2.APIResponse, error) {
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
		return nil, nil, fmt.Errorf("error unmarshaling guide job | error: %w", err)
	}

	return &job, resp, nil
}

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
