package guide_jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
	"io"
	"net/http"
	"os"
)

var internalProxy *guideJobsProxy

type createGuideJobFunc func(ctx context.Context, p *guideJobsProxy, guideJob *GenerateGuideContentRequest) (*GuideJob, *platformclientv2.APIResponse, error)
type getGuideJobByIdFunc func(ctx context.Context, p *guideJobsProxy, id string) (guideJob *GuideJob, resp *platformclientv2.APIResponse, err error)
type deleteGuideJobFunc func(ctx context.Context, p *guideJobsProxy, id string) (resp *platformclientv2.APIResponse, err error)

type guideJobsProxy struct {
	clientConfig *platformclientv2.Configuration
	// TODO: implement api
	createGuideJobAttr  createGuideJobFunc
	getGuideJobByIdAttr getGuideJobByIdFunc
	deleteGuideJobAttr  deleteGuideJobFunc
}

func newGuideJobsProxy(clientConfig *platformclientv2.Configuration) *guideJobsProxy {
	// TODO: implement api
	return &guideJobsProxy{
		clientConfig:        clientConfig,
		createGuideJobAttr:  createGuideJobFn,
		getGuideJobByIdAttr: getGuideJobByIdFn,
		deleteGuideJobAttr:  deleteGuideJobFn,
	}
}

func getGuideJobsProxy(config *platformclientv2.Configuration) *guideJobsProxy {
	if internalProxy == nil {
		internalProxy = newGuideJobsProxy(config)
	}
	return internalProxy
}

func (p *guideJobsProxy) createGuideJob(ctx context.Context, guideJob *GenerateGuideContentRequest) (*GuideJob, *platformclientv2.APIResponse, error) {
	return p.createGuideJobAttr(ctx, p, guideJob)
}

func (p *guideJobsProxy) getGuideJobById(ctx context.Context, id string) (guideJob *GuideJob, resp *platformclientv2.APIResponse, err error) {
	return p.getGuideJobByIdAttr(ctx, p, id)
}

func (p *guideJobsProxy) deleteGuideJob(ctx context.Context, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.deleteGuideJobAttr(ctx, p, id)
}

// Create Functions

func createGuideJobFn(ctx context.Context, p *guideJobsProxy, guideJob *GenerateGuideContentRequest) (*GuideJob, *platformclientv2.APIResponse, error) {
	return sdkCreateGuideJob(ctx, p, guideJob)
}

func sdkCreateGuideJob(ctx context.Context, p *guideJobsProxy, guideJob *GenerateGuideContentRequest) (*GuideJob, *platformclientv2.APIResponse, error) {
	client := &http.Client{}

	jsonBody, err := json.Marshal(guideJob)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling guide job | error: %w", err)
	}

	req, err := http.NewRequest("POST", "XXXXXXXXX", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, nil, fmt.Errorf("error creating guide job request | error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error making request | error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading api response | error: %w", err)
	}

	apiResponse := &platformclientv2.APIResponse{
		StatusCode: resp.StatusCode,
		Response:   resp,
	}

	if resp.StatusCode != 200 {
		return nil, apiResponse, fmt.Errorf("error creating guide job, status code: %d, body: %s", resp.StatusCode, respBody)
	}

	var job GuideJob
	err = json.Unmarshal(respBody, &job)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshaling guide job | error: %w", err)
	}

	return &job, apiResponse, nil
}

// Read Functions

func getGuideJobByIdFn(ctx context.Context, p *guideJobsProxy, id string) (*GuideJob, *platformclientv2.APIResponse, error) {
	return sdkGetGuideJobById(ctx, p, id)
}

func sdkGetGuideJobById(ctx context.Context, p *guideJobsProxy, id string) (*GuideJob, *platformclientv2.APIResponse, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "XXXXXXXXX", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating guide job request | error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("GENESYS_ACCESS_TOKEN"))

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error making request | error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading api response | error: %w", err)
	}

	apiResponse := &platformclientv2.APIResponse{
		StatusCode: resp.StatusCode,
		Response:   resp,
	}

	if resp.StatusCode != 200 {
		return nil, apiResponse, fmt.Errorf("error creating guide job, status code: %d, body: %s", resp.StatusCode, respBody)
	}

	var job GuideJob
	err = json.Unmarshal(respBody, &job)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshaling guide job | error: %w", err)
	}

	return &job, apiResponse, nil
}

// Delete Functions

func deleteGuideJobFn(ctx context.Context, p *guideJobsProxy, id string) (*platformclientv2.APIResponse, error) {
	return sdkDeleteGuideJob(ctx, p, id)
}

func sdkDeleteGuideJob(ctx context.Context, p *guideJobsProxy, id string) (*platformclientv2.APIResponse, error) {
	client := &http.Client{}

	req, err := http.NewRequest("DELETE", "XXXXXXXXX", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating guide job request | error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("GENESYS_ACCESS_TOKEN"))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request | error: %w", err)
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading api response | error: %w", err)
	}

	apiResponse := &platformclientv2.APIResponse{
		StatusCode: resp.StatusCode,
		Response:   resp,
	}

	if resp.StatusCode != 200 {
		return apiResponse, fmt.Errorf("error deleting guide job, status code: %d, body: %s", resp.StatusCode, respBody)
	}

	return apiResponse, nil
}
