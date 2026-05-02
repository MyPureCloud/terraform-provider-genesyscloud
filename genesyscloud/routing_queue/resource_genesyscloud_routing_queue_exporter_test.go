package routing_queue

import (
	"testing"

	architectFlow "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
)

func TestRoutingQueueExporter_InactivityTimeoutFlowIdIsRefAttr(t *testing.T) {
	exporter := RoutingQueueExporter()
	if exporter == nil || exporter.RefAttrs == nil {
		t.Fatalf("expected exporter RefAttrs to be defined")
	}

	key := "media_settings_message.inactivity_timeout_settings.flow_id"
	settings, ok := exporter.RefAttrs[key]
	if !ok || settings == nil {
		t.Fatalf("expected RefAttrs to include %q", key)
	}
	if settings.RefType != architectFlow.ResourceType {
		t.Fatalf("expected %q RefType %q, got %q", key, architectFlow.ResourceType, settings.RefType)
	}
}
