package routing_email_domain

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

func TestUnitRoutingEmailDomainExporter_DataSourceResolver_UsesInstanceID(t *testing.T) {
	instanceID := "delltechnologies.mypurecloud.com"
	state := &terraform.InstanceState{
		ID: instanceID,
		Attributes: map[string]string{
			"domain_id":        "delltechnologies", // subdomain prefix in state
			"subdomain":        "true",
			"id":               "attribute-id-should-not-win",
			"mail_from_domain": "",
		},
	}

	key, val := RoutingEmailDomainExporter().DataResolver(state, "name")
	if key != "name" {
		t.Fatalf("expected key 'name', got %q", key)
	}
	if val != instanceID {
		t.Fatalf("expected data source name to equal instance ID %q, got %q", instanceID, val)
	}
}

func TestUnitGetRoutingEmailDomainIdByName_CaseInsensitive(t *testing.T) {
	lowerID := "delltechnologies.mypurecloud.com"
	proxy := &routingEmailDomainProxy{
		getAllRoutingEmailDomainsAttr: func(_ context.Context, _ *routingEmailDomainProxy) (*[]platformclientv2.Inbounddomain, *platformclientv2.APIResponse, error) {
			domains := []platformclientv2.Inbounddomain{
				{Id: &lowerID},
			}
			return &domains, &platformclientv2.APIResponse{StatusCode: 200}, nil
		},
	}

	gotID, _, retryable, err := getRoutingEmailDomainIdByNameFn(context.Background(), proxy, "DellTechnologies.MyPureCloud.Com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if retryable {
		t.Fatalf("expected retryable=false, got true")
	}
	if gotID != lowerID {
		t.Fatalf("expected id %q, got %q", lowerID, gotID)
	}
}

func TestUnitGetRoutingEmailDomainIdByName_SubdomainPrefixMatch(t *testing.T) {
	fullID := "delltechnologies.mypurecloud.com"
	subdomain := true
	proxy := &routingEmailDomainProxy{
		getAllRoutingEmailDomainsAttr: func(_ context.Context, _ *routingEmailDomainProxy) (*[]platformclientv2.Inbounddomain, *platformclientv2.APIResponse, error) {
			domains := []platformclientv2.Inbounddomain{
				{Id: &fullID, SubDomain: &subdomain},
			}
			return &domains, &platformclientv2.APIResponse{StatusCode: 200}, nil
		},
	}

	gotID, _, retryable, err := getRoutingEmailDomainIdByNameFn(context.Background(), proxy, "DellTechnologies")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if retryable {
		t.Fatalf("expected retryable=false, got true")
	}
	if gotID != fullID {
		t.Fatalf("expected id %q, got %q", fullID, gotID)
	}
}
