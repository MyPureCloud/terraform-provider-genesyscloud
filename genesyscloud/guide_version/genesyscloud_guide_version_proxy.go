package guide_version

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
	"io"
	"net/http"
)

var internalProxy *guideVersionProxy

type createGuideVersionFunc func(ctx context.Context, p *guideVersionProxy, guideVersion *CreateGuideVersionRequest, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error)
type getGuideVersionByIdFunc func(ctx context.Context, p *guideVersionProxy, id string, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error)
type updateGuideVersionFunc func(ctx context.Context, p *guideVersionProxy, id string, guideId string, guideVersion *UpdateGuideVersion) (*VersionResponse, *platformclientv2.APIResponse, error)

type guideVersionProxy struct {
	clientConfig            *platformclientv2.Configuration
	createGuideVersionAttr  createGuideVersionFunc
	getGuideVersionByIdAttr getGuideVersionByIdFunc
	updateGuideVersionAttr  updateGuideVersionFunc
}

func newGuideVersionProxy(clientConfig *platformclientv2.Configuration) *guideVersionProxy {
	return &guideVersionProxy{
		clientConfig:            clientConfig,
		createGuideVersionAttr:  createGuideVersionFn,
		getGuideVersionByIdAttr: getGuideVersionByIdFn,
		updateGuideVersionAttr:  updateGuideVersionFn,
	}
}
func getGuideVersionProxy(clientConfig *platformclientv2.Configuration) *guideVersionProxy {
	if internalProxy == nil {
		internalProxy = newGuideVersionProxy(clientConfig)
	}
	return internalProxy
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

// Create Functions

func createGuideVersionFn(ctx context.Context, p *guideVersionProxy, guideVersion *CreateGuideVersionRequest, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error) {
	return sdkPostGuideVersion(guideVersion, p, guideId)
}

func sdkPostGuideVersion(body *CreateGuideVersionRequest, p *guideVersionProxy, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error) {
	client := &http.Client{}
	action := http.MethodPost
	baseURL := "https://api.inintca.com/api/v2/guides/" + guideId + "/versions"

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
	client := &http.Client{}
	action := http.MethodGet
	baseURL := "https://api.inintca.com/api/v2/guides/" + guideId + "/versions/" + id

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
	client := &http.Client{}
	action := http.MethodPatch
	baseURL := "https://api.inintca.com/api/v2/guides/" + guideId + "/versions/" + id

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
	if resp.StatusCode != 200 {
		return nil, apiResp, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, apiResp, nil
}
