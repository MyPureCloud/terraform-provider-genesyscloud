package axon

import (
	"context"
	"log"
	"net/url"

	customapi "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/custom_api_client"

	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"
)

// Following would normally be defined in resource_genesyscloud_dependencies_schema.go
const ResourceType = "genesyscloud_dependencies"

// DependencyCount is the response of the requiredbycounts dependencies endpoint.
type DependencyCount struct {
	EstimatedCount int `json:"estimatedCount"`
}

// EntityConnection is a single dependency edge to another entity.
type EntityConnection struct {
	EntityType string `json:"entityType"`
	EntityID   string `json:"entityId"`
}

// EntityConnectionPagedResponse is the paged response of the requires/requiredby endpoints.
type EntityConnectionPagedResponse struct {
	EntityID    string             `json:"entityId"`
	EntityType  string             `json:"entityType"`
	Connections []EntityConnection `json:"connections"`
	SelfURI     string             `json:"selfUri"`
	NextURI     string             `json:"nextUri"`
	PreviousURI string             `json:"previousUri"`
}

type getRequiredByCountFunc func(ctx context.Context, p *AxonProxy, entityType, entityID string) (int, error)
type getRequiresFunc func(ctx context.Context, p *AxonProxy, entityType, entityID string, queryParams url.Values) (*EntityConnectionPagedResponse, *platformclientv2.APIResponse, error)
type getRequiredByFunc func(ctx context.Context, p *AxonProxy, entityType, entityID string, queryParams url.Values) (*EntityConnectionPagedResponse, *platformclientv2.APIResponse, error)

type AxonProxy struct {
	clientConfig           *platformclientv2.Configuration
	getRequiredByCountAttr getRequiredByCountFunc
	getRequiresAttr        getRequiresFunc
	getRequiredByAttr      getRequiredByFunc
}

func NewAxonProxy(clientConfig *platformclientv2.Configuration) *AxonProxy {
	return &AxonProxy{
		clientConfig:           clientConfig,
		getRequiredByCountAttr: getRequiredByCountFn,
		getRequiresAttr:        getRequiresFn,
		getRequiredByAttr:      getRequiredByFn,
	}
}

func (p *AxonProxy) GetRequiredByCount(ctx context.Context, entityType, entityID string) (int, error) {
	count, err := p.getRequiredByCountAttr(ctx, p, entityType, entityID)
	return count, err
}

// getRequires pages the outgoing connections (entities this entity requires).
func (p *AxonProxy) getRequires(ctx context.Context, entityType, entityID string, queryParams url.Values) (*EntityConnectionPagedResponse, *platformclientv2.APIResponse, error) {
	return p.getRequiresAttr(ctx, p, entityType, entityID, queryParams)
}

// getRequiredBy pages the incoming connections (entities that require this entity).
func (p *AxonProxy) getRequiredBy(ctx context.Context, entityType, entityID string, queryParams url.Values) (*EntityConnectionPagedResponse, *platformclientv2.APIResponse, error) {
	return p.getRequiredByAttr(ctx, p, entityType, entityID, queryParams)
}

func getRequiredByCountFn(ctx context.Context, p *AxonProxy, entityType, entityID string) (int, error) {
	c := customapi.NewClient(p.clientConfig, ResourceType)
	path := "/api/v2/dependencies/type/" + url.PathEscape(entityType) + "/id/" + url.PathEscape(entityID) + "/connections/requiredbycounts"

	result, resp, err := customapi.Do[DependencyCount](ctx, c, customapi.MethodGet, path, nil, nil)
	if err != nil {
		if resp.StatusCode == 404 {
			// Return 0 w/o error if the entity doesn't exist
			return 0, nil
		} else {
			log.Printf("Failed to get dependency count for %s %s: %v", entityType, entityID, err)
			// log.Print(resp)
			return 0, err
		}
	}
	return result.EstimatedCount, nil
}

func getRequiresFn(ctx context.Context, p *AxonProxy, entityType, entityID string, queryParams url.Values) (*EntityConnectionPagedResponse, *platformclientv2.APIResponse, error) {
	c := customapi.NewClient(p.clientConfig, ResourceType)
	path := "/api/v2/dependencies/type/" + url.PathEscape(entityType) + "/id/" + url.PathEscape(entityID) + "/connections/requires"

	return customapi.Do[EntityConnectionPagedResponse](ctx, c, customapi.MethodGet, path, nil, queryParams)
}

func getRequiredByFn(ctx context.Context, p *AxonProxy, entityType, entityID string, queryParams url.Values) (*EntityConnectionPagedResponse, *platformclientv2.APIResponse, error) {
	c := customapi.NewClient(p.clientConfig, ResourceType)
	path := "/api/v2/dependencies/type/" + url.PathEscape(entityType) + "/id/" + url.PathEscape(entityID) + "/connections/requiredby"

	return customapi.Do[EntityConnectionPagedResponse](ctx, c, customapi.MethodGet, path, nil, queryParams)
}
