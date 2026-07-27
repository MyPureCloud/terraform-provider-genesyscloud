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

type getRequiredByCountFunc func(ctx context.Context, p *AxonProxy, resourceType, entityType, entityID string) (int, *platformclientv2.APIResponse, error)

type AxonProxy struct {
	clientConfig           *platformclientv2.Configuration
	getRequiredByCountAttr getRequiredByCountFunc
}

func NewAxonProxy(clientConfig *platformclientv2.Configuration) *AxonProxy {
	return &AxonProxy{
		clientConfig:           clientConfig,
		getRequiredByCountAttr: getRequiredByCountFn,
	}
}

func (p *AxonProxy) GetRequiredByCount(ctx context.Context, resourceType, entityType, entityID string) (int, *platformclientv2.APIResponse, error) {
	return p.getRequiredByCountAttr(ctx, p, resourceType, entityType, entityID)
}

func getRequiredByCountFn(ctx context.Context, p *AxonProxy, resourceType, entityType, entityID string) (int, *platformclientv2.APIResponse, error) {
	c := customapi.NewClient(p.clientConfig, resourceType)
	path := "/api/v2/dependencies/type/" + url.PathEscape(entityType) + "/id/" + url.PathEscape(entityID) + "/connections/requiredbycounts"

	result, resp, err := customapi.Do[DependencyCount](ctx, c, customapi.MethodGet, path, nil, nil)
	if err != nil {
		return 0, resp, err
	}
	return result.EstimatedCount, resp, nil
}
