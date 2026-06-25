package routing_email_route

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
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

// TestUnitBuildSignature verifies that buildSignature correctly maps Terraform state to the SDK struct
func TestUnitBuildSignature(t *testing.T) {
	t.Run("returns nil when signature block is absent", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, ResourceRoutingEmailRoute().Schema, map[string]interface{}{})
		result := buildSignature(d)
		if result != nil {
			t.Errorf("expected nil, got %+v", result)
		}
	})

	t.Run("maps all fields correctly", func(t *testing.T) {
		cannedID := "canned-abc-123"
		inclusionType := "Always"
		d := schema.TestResourceDataRaw(t, ResourceRoutingEmailRoute().Schema, map[string]interface{}{
			"signature": []interface{}{
				map[string]interface{}{
					"enabled":            true,
					"canned_response_id": cannedID,
					"always_included":    false,
					"inclusion_type":     inclusionType,
				},
			},
		})

		result := buildSignature(d)
		if result == nil {
			t.Fatal("expected non-nil Signature, got nil")
		}
		if result.Enabled == nil || *result.Enabled != true {
			t.Errorf("expected Enabled=true, got %v", result.Enabled)
		}
		if result.CannedResponseId == nil || *result.CannedResponseId != cannedID {
			t.Errorf("expected CannedResponseId=%q, got %v", cannedID, result.CannedResponseId)
		}
		if result.AlwaysIncluded == nil || *result.AlwaysIncluded != false {
			t.Errorf("expected AlwaysIncluded=false, got %v", result.AlwaysIncluded)
		}
		if result.InclusionType == nil || *result.InclusionType != inclusionType {
			t.Errorf("expected InclusionType=%q, got %v", inclusionType, result.InclusionType)
		}
	})

	t.Run("omits canned_response_id when empty string", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, ResourceRoutingEmailRoute().Schema, map[string]interface{}{
			"signature": []interface{}{
				map[string]interface{}{
					"enabled":            false,
					"canned_response_id": "",
					"always_included":    true,
					"inclusion_type":     "",
				},
			},
		})

		result := buildSignature(d)
		if result == nil {
			t.Fatal("expected non-nil Signature, got nil")
		}
		if result.CannedResponseId != nil {
			t.Errorf("expected CannedResponseId to be nil when empty, got %q", *result.CannedResponseId)
		}
		if result.InclusionType != nil {
			t.Errorf("expected InclusionType to be nil when empty, got %q", *result.InclusionType)
		}
	})
}

// TestUnitFlattenSignature verifies that flattenSignature correctly maps the SDK struct to Terraform state
func TestUnitFlattenSignature(t *testing.T) {
	t.Run("returns nil when signature is nil", func(t *testing.T) {
		result := flattenSignature(nil)
		if result != nil {
			t.Errorf("expected nil, got %+v", result)
		}
	})

	t.Run("maps all fields correctly", func(t *testing.T) {
		enabled := true
		cannedID := "canned-xyz-456"
		alwaysIncluded := false
		inclusionType := "FirstResponseOnly"

		sig := &platformclientv2.Signature{
			Enabled:          &enabled,
			CannedResponseId: &cannedID,
			AlwaysIncluded:   &alwaysIncluded,
			InclusionType:    &inclusionType,
		}

		result := flattenSignature(sig)
		if len(result) != 1 {
			t.Fatalf("expected slice of length 1, got %d", len(result))
		}
		sigMap, ok := result[0].(map[string]interface{})
		if !ok {
			t.Fatal("expected result[0] to be map[string]interface{}")
		}
		if sigMap["enabled"] != enabled {
			t.Errorf("expected enabled=%v, got %v", enabled, sigMap["enabled"])
		}
		if sigMap["canned_response_id"] != cannedID {
			t.Errorf("expected canned_response_id=%q, got %v", cannedID, sigMap["canned_response_id"])
		}
		if sigMap["always_included"] != alwaysIncluded {
			t.Errorf("expected always_included=%v, got %v", alwaysIncluded, sigMap["always_included"])
		}
		if sigMap["inclusion_type"] != inclusionType {
			t.Errorf("expected inclusion_type=%q, got %v", inclusionType, sigMap["inclusion_type"])
		}
	})

	t.Run("handles partially nil fields", func(t *testing.T) {
		enabled := false
		sig := &platformclientv2.Signature{
			Enabled: &enabled,
			// CannedResponseId, AlwaysIncluded, InclusionType intentionally nil
		}

		result := flattenSignature(sig)
		if len(result) != 1 {
			t.Fatalf("expected slice of length 1, got %d", len(result))
		}
		sigMap := result[0].(map[string]interface{})
		if sigMap["canned_response_id"] != nil {
			t.Errorf("expected canned_response_id to be nil, got %v", sigMap["canned_response_id"])
		}
		if sigMap["always_included"] != nil {
			t.Errorf("expected always_included to be nil, got %v", sigMap["always_included"])
		}
		if sigMap["inclusion_type"] != nil {
			t.Errorf("expected inclusion_type to be nil, got %v", sigMap["inclusion_type"])
		}
	})
}
