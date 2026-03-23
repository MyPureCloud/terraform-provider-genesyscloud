package guide_version

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	customapi "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/custom_api_client"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
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
	customApiClient                     *customapi.Client
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
		customApiClient:                     customapi.NewClient(clientConfig, ResourceType),
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
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return sdkGetAllGuides(ctx, p)
}

func sdkGetAllGuides(ctx context.Context, p *guideVersionProxy) (*[]Guide, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	var allGuides []Guide
	queryParams := customapi.NewQueryParams(map[string]string{"pageSize": "100", "pageNumber": "1"})

	guides, resp, err := customapi.Do[GuideEntityListing](ctx, p.customApiClient, customapi.MethodGet, "/api/v2/guides", nil, queryParams)
	if err != nil {
		return nil, resp, err
	}
	if guides.Entities == nil {
		return &allGuides, resp, nil
	}
	allGuides = append(allGuides, *guides.Entities...)

	if guides.PageCount != nil && *guides.PageCount > 1 {
		for pageNum := 2; pageNum <= *guides.PageCount; pageNum++ {
			queryParams.Set("pageNumber", fmt.Sprintf("%v", pageNum))
			pageGuides, _, err := customapi.Do[GuideEntityListing](ctx, p.customApiClient, customapi.MethodGet, "/api/v2/guides", nil, queryParams)
			if err != nil {
				return nil, resp, fmt.Errorf("error fetching page %d: %v", pageNum, err)
			}
			if pageGuides.Entities != nil {
				allGuides = append(allGuides, *pageGuides.Entities...)
			}
		}
	}
	log.Printf("Successfully retrieved %d guides", len(allGuides))
	return &allGuides, resp, nil
}

// Create Functions

func createGuideVersionFn(ctx context.Context, p *guideVersionProxy, guideVersion *CreateGuideVersionRequest, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return sdkPostGuideVersion(ctx, guideVersion, p, guideId)
}

func sdkPostGuideVersion(ctx context.Context, body *CreateGuideVersionRequest, p *guideVersionProxy, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return customapi.Do[VersionResponse](ctx, p.customApiClient, customapi.MethodPost, "/api/v2/guides/"+guideId+"/versions", body, nil)
}

// Read Functions

func getGuideVersionByIdFn(ctx context.Context, p *guideVersionProxy, id string, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return sdkGetGuideVersionById(ctx, p, id, guideId)
}

func sdkGetGuideVersionById(ctx context.Context, p *guideVersionProxy, id string, guideId string) (*VersionResponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return customapi.Do[VersionResponse](ctx, p.customApiClient, customapi.MethodGet, "/api/v2/guides/"+guideId+"/versions/"+id, nil, nil)
}

// Update Functions

func updateGuideVersionFn(ctx context.Context, p *guideVersionProxy, id string, guideId string, guideVersion *UpdateGuideVersion) (*VersionResponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return sdkUpdateGuideVersion(ctx, p, id, guideId, guideVersion)
}

func sdkUpdateGuideVersion(ctx context.Context, p *guideVersionProxy, id string, guideId string, body *UpdateGuideVersion) (*VersionResponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return customapi.Do[VersionResponse](ctx, p.customApiClient, customapi.MethodPatch, "/api/v2/guides/"+guideId+"/versions/"+id, body, nil)
}

// Functions to publish the guide version

func publishGuideVersionFn(ctx context.Context, p *guideVersionProxy, body *GuideVersionPublishJobRequest) (*VersionJobResponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return sdkPublishGuideVersion(ctx, p, body)
}

func sdkPublishGuideVersion(ctx context.Context, p *guideVersionProxy, body *GuideVersionPublishJobRequest) (*VersionJobResponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	path := "/api/v2/guides/" + body.GuideId + "/versions/" + body.VersionId + "/jobs"
	rawBody, resp, err := customapi.DoRaw(ctx, p.customApiClient, customapi.MethodPost, path, body, nil)
	if err != nil {
		return nil, resp, err
	}
	if resp.StatusCode == 202 && len(rawBody) == 0 {
		log.Println("Received 202 with empty body - job started successfully")
		return nil, resp, nil
	}
	if len(rawBody) == 0 {
		return nil, resp, fmt.Errorf("empty response body")
	}
	var jobResponse VersionJobResponse
	if err := json.Unmarshal(rawBody, &jobResponse); err != nil {
		return nil, resp, fmt.Errorf("error unmarshaling response: %v, body: %s", err, string(rawBody))
	}
	return &jobResponse, resp, nil
}

func getGuideVersionPublishJobStatusFn(ctx context.Context, p *guideVersionProxy, versionId, jobId, guideId string) (*VersionJobResponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return sdkGetGuideVersionPublishJobStatus(ctx, p, versionId, jobId, guideId)
}

func sdkGetGuideVersionPublishJobStatus(ctx context.Context, p *guideVersionProxy, versionId, jobId string, guideId string) (*VersionJobResponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return customapi.Do[VersionJobResponse](ctx, p.customApiClient, customapi.MethodGet, "/api/v2/guides/"+guideId+"/versions/"+versionId+"/jobs/"+jobId, nil, nil)
}

func getGuideByIdFn(ctx context.Context, p *guideVersionProxy, id string) (*Guide, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return sdkGetGuideById(ctx, p, id)
}

func sdkGetGuideById(ctx context.Context, p *guideVersionProxy, id string) (*Guide, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return customapi.Do[Guide](ctx, p.customApiClient, customapi.MethodGet, "/api/v2/guides/"+id, nil, nil)
}
