package routing_email_domain

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
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

func TestUnitFlattenGraphApiSettings(t *testing.T) {
	integrationID := "6572c166-70dc-4ea7-b410-cabe2ee3e4c6"
	status := "Active"

	flat := flattenGraphApiSettings(&platformclientv2.Graphapisettings{
		Integration: &platformclientv2.Domainentityref{Id: &integrationID},
		Status:      &status,
	})
	if len(flat) != 1 {
		t.Fatalf("expected 1 flattened block, got %d", len(flat))
	}

	settingsMap, ok := flat[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", flat[0])
	}
	if settingsMap["integration_id"] != integrationID {
		t.Fatalf("expected integration_id %q, got %v", integrationID, settingsMap["integration_id"])
	}
	if settingsMap["status"] != status {
		t.Fatalf("expected status %q, got %v", status, settingsMap["status"])
	}

	if flattenGraphApiSettings(nil) != nil {
		t.Fatalf("expected nil for nil settings")
	}
}

func TestUnitFlattenImapSettings(t *testing.T) {
	integrationID := "imap-integration-id"
	status := "AwaitingFolders"

	flat := flattenImapSettings(&platformclientv2.Imapsettings{
		Integration: &platformclientv2.Domainentityref{Id: &integrationID},
		Status:      &status,
	})
	if len(flat) != 1 {
		t.Fatalf("expected 1 flattened block, got %d", len(flat))
	}

	settingsMap, ok := flat[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", flat[0])
	}
	if settingsMap["integration_id"] != integrationID {
		t.Fatalf("expected integration_id %q, got %v", integrationID, settingsMap["integration_id"])
	}
	if settingsMap["status"] != status {
		t.Fatalf("expected status %q, got %v", status, settingsMap["status"])
	}

	if flattenImapSettings(nil) != nil {
		t.Fatalf("expected nil for nil settings")
	}
}
