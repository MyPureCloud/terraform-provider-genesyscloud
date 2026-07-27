package axon

import (
	"context"
	"log"
	"net/url"

	customapi "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/custom_api_client"

	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"
)

// DependencyCount is the response of the requiredbycounts dependencies endpoint.
type DependencyCount struct {
	EstimatedCount int `json:"estimatedCount"`
}

type getRequiredByCountFunc func(ctx context.Context, p *AxonProxy, entityType, entityID string) (int, error)

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

func (p *AxonProxy) GetRequiredByCount(ctx context.Context, entityType, entityID string) (int, error) {
	count, err := p.getRequiredByCountAttr(ctx, p, entityType, entityID)
	return count, err
}

func getRequiredByCountFn(ctx context.Context, p *AxonProxy, entityType, entityID string) (int, error) {
	c := customapi.NewClient(p.clientConfig, "dependencies")
	path := "/api/v2/dependencies/type/" + url.PathEscape(entityType) + "/id/" + url.PathEscape(entityID) + "/connections/requiredbycounts"

	result, resp, err := customapi.Do[DependencyCount](ctx, c, customapi.MethodGet, path, nil, nil)
	if err != nil {
		if resp.StatusCode == 404 {
			// Return 0 if the entity doesn't exist
			return 0, nil
		} else {
			log.Printf("Failed to get dependency count for %s %s: %v", entityType, entityID, err)
			log.Print(resp)
			return 0, err
		}
	}
	return result.EstimatedCount, nil
}
