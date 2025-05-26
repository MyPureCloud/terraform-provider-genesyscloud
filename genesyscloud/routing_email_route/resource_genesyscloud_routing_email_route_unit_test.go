package routing_email_route

import (
	"context"
	"fmt"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

// TestUnitGetAllRoutingEmailRoutes verifies that errors are returned properly and most importantly that no nil pointer
// exceptions occur when nothing objects are returned from getAllRoutingEmailRoutes
func TestUnitGetAllRoutingEmailRoutes(t *testing.T) {
	internalProxyCopy := internalProxy
	defer func() {
		internalProxy = internalProxyCopy
	}()

	sdkConfig := platformclientv2.GetDefaultConfiguration()
	ctx := context.Background()

	mockGetAllFunc := func(_ context.Context, _ *routingEmailRouteProxy, _, _ string) (*map[string][]platformclientv2.Inboundroute, *platformclientv2.APIResponse, error) {
		return nil, nil, nil
	}
	internalProxy = &routingEmailRouteProxy{
		getAllRoutingEmailRouteAttr: mockGetAllFunc,
	}

	resourceMap, diagErr := getAllRoutingEmailRoutes(ctx, sdkConfig)
	if diagErr != nil {
		t.Errorf("Expected diagnostic error to be nil, got '%v'", diagErr)
	}

	if resourceMap == nil {
		t.Errorf("Expected resource map to not be nil")
	}

	if len(resourceMap) > 0 {
		t.Errorf("Expected resource map length to be 0, got %d", len(resourceMap))
	}

	mockGetAllFunc = func(_ context.Context, _ *routingEmailRouteProxy, _, _ string) (*map[string][]platformclientv2.Inboundroute, *platformclientv2.APIResponse, error) {
		return nil, nil, fmt.Errorf("mock error")
	}
	internalProxy = &routingEmailRouteProxy{
		getAllRoutingEmailRouteAttr: mockGetAllFunc,
	}

	resourceMap, diagErr = getAllRoutingEmailRoutes(ctx, sdkConfig)
	if diagErr == nil {
		t.Errorf("Expected diagnostics error to not be nil")
	}

	if resourceMap != nil {
		t.Errorf("Expected resource map to be nil, got %v", resourceMap)
	}

	if len(resourceMap) > 0 {
		t.Errorf("Expected resource map length to be 0, got %d", len(resourceMap))
	}
}
