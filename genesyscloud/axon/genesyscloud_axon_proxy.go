package axon

import (
	"context"
	"net/url"

	customapi "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/custom_api_client"

	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"
)

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

type getRequiredByCountFunc func(ctx context.Context, p *axonProxy, resourceType, entityType, entityID string) (int, *platformclientv2.APIResponse, error)
type getRequiresFunc func(ctx context.Context, p *axonProxy, resourceType, entityType, entityID string, queryParams url.Values) (*EntityConnectionPagedResponse, *platformclientv2.APIResponse, error)
type getRequiredByFunc func(ctx context.Context, p *axonProxy, resourceType, entityType, entityID string, queryParams url.Values) (*EntityConnectionPagedResponse, *platformclientv2.APIResponse, error)

type axonProxy struct {
	clientConfig           *platformclientv2.Configuration
	getRequiredByCountAttr getRequiredByCountFunc
	getRequiresAttr        getRequiresFunc
	getRequiredByAttr      getRequiredByFunc
}

func newAxonProxy(clientConfig *platformclientv2.Configuration) *axonProxy {
	return &axonProxy{
		clientConfig:           clientConfig,
		getRequiredByCountAttr: getRequiredByCountFn,
		getRequiresAttr:        getRequiresFn,
		getRequiredByAttr:      getRequiredByFn,
	}
}

func (p *axonProxy) getRequiredByCount(ctx context.Context, resourceType, entityType, entityID string) (int, *platformclientv2.APIResponse, error) {
	return p.getRequiredByCountAttr(ctx, p, resourceType, entityType, entityID)
}

// getRequires pages the outgoing connections (entities this entity requires).
func (p *axonProxy) getRequires(ctx context.Context, resourceType, entityType, entityID string, queryParams url.Values) (*EntityConnectionPagedResponse, *platformclientv2.APIResponse, error) {
	return p.getRequiresAttr(ctx, p, resourceType, entityType, entityID, queryParams)
}

// getRequiredBy pages the incoming connections (entities that require this entity).
func (p *axonProxy) getRequiredBy(ctx context.Context, resourceType, entityType, entityID string, queryParams url.Values) (*EntityConnectionPagedResponse, *platformclientv2.APIResponse, error) {
	return p.getRequiredByAttr(ctx, p, resourceType, entityType, entityID, queryParams)
}

func getRequiredByCountFn(ctx context.Context, p *axonProxy, resourceType, entityType, entityID string) (int, *platformclientv2.APIResponse, error) {
	c := customapi.NewClient(p.clientConfig, resourceType)
	path := "/api/v2/dependencies/type/" + url.PathEscape(entityType) + "/id/" + url.PathEscape(entityID) + "/connections/requiredbycounts"

	result, resp, err := customapi.Do[DependencyCount](ctx, c, customapi.MethodGet, path, nil, nil)
	if err != nil {
		return 0, resp, err
	}
	return result.EstimatedCount, resp, nil
}

func getRequiresFn(ctx context.Context, p *axonProxy, resourceType, entityType, entityID string, queryParams url.Values) (*EntityConnectionPagedResponse, *platformclientv2.APIResponse, error) {
	c := customapi.NewClient(p.clientConfig, resourceType)
	path := "/api/v2/dependencies/type/" + url.PathEscape(entityType) + "/id/" + url.PathEscape(entityID) + "/connections/requires"

	return customapi.Do[EntityConnectionPagedResponse](ctx, c, customapi.MethodGet, path, nil, queryParams)
}

func getRequiredByFn(ctx context.Context, p *axonProxy, resourceType, entityType, entityID string, queryParams url.Values) (*EntityConnectionPagedResponse, *platformclientv2.APIResponse, error) {
	c := customapi.NewClient(p.clientConfig, resourceType)
	path := "/api/v2/dependencies/type/" + url.PathEscape(entityType) + "/id/" + url.PathEscape(entityID) + "/connections/requiredby"

	return customapi.Do[EntityConnectionPagedResponse](ctx, c, customapi.MethodGet, path, nil, queryParams)
}
