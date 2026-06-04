package routing_email_route

import (
	"testing"
)

func TestRoutingEmailRouteExporter_RemoveIfMissingKeepsSelfReferenceReplyAddress(t *testing.T) {
	exporter := RoutingEmailRouteExporter()
	if exporter == nil {
		t.Fatal("expected RoutingEmailRouteExporter to be defined")
	}

	replyConfig := map[string]interface{}{
		"self_reference_route": true,
		"route_id":             nil,
	}

	if exporter.RemoveFieldIfMissing("reply_email_address", replyConfig) {
		t.Fatal("expected reply_email_address block to be kept when self_reference_route is true")
	}
}

func TestRoutingEmailRouteExporter_HasSelfReferenceRouteResolver(t *testing.T) {
	exporter := RoutingEmailRouteExporter()
	for _, attr := range []string{
		"reply_email_address.route_id",
		"reply_email_address.self_reference_route",
	} {
		resolver, ok := exporter.CustomAttributeResolver[attr]
		if !ok || resolver == nil || resolver.ResolverFunc == nil {
			t.Fatalf("expected %s custom resolver to be configured", attr)
		}
	}
}
