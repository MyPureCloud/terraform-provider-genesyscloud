package guides

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

var internalProxy *guideProxy

type getAllGuidesFunc func(ctx context.Context, p *guideProxy, name string) (*[]Guide, *platformclientv2.APIResponse, error)
type createGuideFunc func(ctx context.Context, p *guideProxy, guide *Guide) (*Guide, *platformclientv2.APIResponse, error)
type getGuideByIdFunc func(ctx context.Context, p *guideProxy, id string) (*Guide, *platformclientv2.APIResponse, error)
type getGuideByNameFunc func(ctx context.Context, p *guideProxy, name string) (string, bool, *platformclientv2.APIResponse, error)

type guideProxy struct {
	clientConfig *platformclientv2.Configuration
	//guideApi        *platformclientv2.GuideApi
	getAllGuidesAttr   getAllGuidesFunc
	createGuideAttr    createGuideFunc
	getGuideByIdAttr   getGuideByIdFunc
	getGuideByNameAttr getGuideByNameFunc
}

func newGuideProxy(clientConfig *platformclientv2.Configuration) *guideProxy {
	//api := platformclientv2.NewGuideApiWithConfig(clientConfig)
	return &guideProxy{
		clientConfig: clientConfig,
		//guideApi:        api,
		getAllGuidesAttr:   getAllGuidesFn,
		createGuideAttr:    createGuideFn,
		getGuideByIdAttr:   getGuideByIdFn,
		getGuideByNameAttr: getGuideByNameFn,
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

func (p *guideProxy) createGuide(ctx context.Context, guide *Guide) (*Guide, *platformclientv2.APIResponse, error) {
	return p.createGuideAttr(ctx, p, guide)
}

func (p *guideProxy) getGuideById(ctx context.Context, id string) (*Guide, *platformclientv2.APIResponse, error) {
	return p.getGuideByIdAttr(ctx, p, id)
}

func (p *guideProxy) getGuideByName(ctx context.Context, name string) (string, bool, *platformclientv2.APIResponse, error) {
	return p.getGuideByNameAttr(ctx, p, name)
}

// GetAll Functions

func getAllGuidesFn(ctx context.Context, p *guideProxy, name string) (*[]Guide, *platformclientv2.APIResponse, error) {
	return sdkGetAllGuidesFn(ctx, p, name)
}

func sdkGetAllGuidesFn(ctx context.Context, p *guideProxy, name string) (*[]Guide, *platformclientv2.APIResponse, error) {
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("GET", "https://apicentral.genesys.cloud/api-explorer#get-api-v2-guides", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("GENESYS_ACCESS_TOKEN"))

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading response: %v", err)
	}

	// Check status code
	apiResp := &platformclientv2.APIResponse{
		StatusCode: resp.StatusCode,
		Response:   resp,
	}

	if resp.StatusCode != 200 {
		return nil, apiResp, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response into Guide struct
	var guides []Guide
	if err := json.Unmarshal(respBody, &guides); err != nil {
		return nil, apiResp, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &guides, apiResp, nil
}

// Create Functions

func createGuideFn(ctx context.Context, p *guideProxy, guide *Guide) (*Guide, *platformclientv2.APIResponse, error) {
	return sdkPostGuide(guide)
}

func sdkPostGuide(body *Guide) (*Guide, *platformclientv2.APIResponse, error) {
	client := &http.Client{}

	// Convert body to JSON
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling guide: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", "https://api.mypurecloud.com/api/v2/guides", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("GENESYS_ACCESS_TOKEN"))

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading response: %v", err)
	}

	// Check status code
	apiResp := &platformclientv2.APIResponse{
		StatusCode: resp.StatusCode,
		Response:   resp,
	}

	if resp.StatusCode != 200 {
		return nil, apiResp, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response into Guide struct
	var guide Guide
	if err := json.Unmarshal(respBody, &guide); err != nil {
		return nil, apiResp, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &guide, apiResp, nil
}

// Read Functions

func getGuideByIdFn(ctx context.Context, p *guideProxy, id string) (*Guide, *platformclientv2.APIResponse, error) {
	return sdkGetGuideById(id)
}

func sdkGetGuideById(id string) (*Guide, *platformclientv2.APIResponse, error) {
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("GET", "https://api.mypurecloud.com/api/v2/guides/"+id, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("GENESYS_ACCESS_TOKEN"))

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading response: %v", err)
	}

	// Check status code
	apiResp := &platformclientv2.APIResponse{
		StatusCode: resp.StatusCode,
		Response:   resp,
	}

	if resp.StatusCode != 200 {
		return nil, apiResp, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response into Guide struct
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

	if guides == nil || len(*guides) == 0 {
		return "", true, resp, fmt.Errorf("no guides found with name: %s", name)
	}

	for _, guide := range *guides {
		if *guide.Name == name {
			return *guide.Id, false, resp, nil
		}
	}

	return "", false, resp, fmt.Errorf("unable to find guide with name %s", name)
}
