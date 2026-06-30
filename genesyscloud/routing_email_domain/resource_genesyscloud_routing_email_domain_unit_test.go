package routing_email_domain

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

func routingEmailDomainTestSchema() map[string]*schema.Schema {
	return ResourceRoutingEmailDomain().Schema
}

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

func TestUnitExpandGraphApiSettings(t *testing.T) {
	integrationID := "6572c166-70dc-4ea7-b410-cabe2ee3e4c6"
	resourceSchema := routingEmailDomainTestSchema()

	tests := []struct {
		name    string
		data    map[string]interface{}
		wantNil bool
		wantID  string
	}{
		{
			name: "block absent",
			data: map[string]interface{}{
				"domain_id": "example.com",
				"subdomain": false,
			},
			wantNil: true,
		},
		{
			name: "empty block list",
			data: map[string]interface{}{
				"domain_id":          "example.com",
				"subdomain":          false,
				"graph_api_settings": []interface{}{},
			},
			wantNil: true,
		},
		{
			name: "empty integration_id",
			data: map[string]interface{}{
				"domain_id": "example.com",
				"subdomain": false,
				"graph_api_settings": []interface{}{
					map[string]interface{}{
						"integration_id": "",
					},
				},
			},
			wantNil: true,
		},
		{
			name: "valid integration_id",
			data: map[string]interface{}{
				"domain_id": "example.com",
				"subdomain": false,
				"graph_api_settings": []interface{}{
					map[string]interface{}{
						"integration_id": integrationID,
					},
				},
			},
			wantNil: false,
			wantID:  integrationID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.data)
			got := expandGraphApiSettings(d)

			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %+v", got)
				}
				return
			}

			if got == nil || got.Integration == nil || got.Integration.Id == nil {
				t.Fatalf("expected Graph API settings with integration id, got %+v", got)
			}
			if *got.Integration.Id != tt.wantID {
				t.Fatalf("expected integration id %q, got %q", tt.wantID, *got.Integration.Id)
			}
		})
	}
}

func TestUnitExpandImapSettings(t *testing.T) {
	integrationID := "imap-integration-id"
	resourceSchema := routingEmailDomainTestSchema()

	tests := []struct {
		name    string
		data    map[string]interface{}
		wantNil bool
		wantID  string
	}{
		{
			name: "block absent",
			data: map[string]interface{}{
				"domain_id": "example.com",
				"subdomain": false,
			},
			wantNil: true,
		},
		{
			name: "empty block list",
			data: map[string]interface{}{
				"domain_id":     "example.com",
				"subdomain":     false,
				"imap_settings": []interface{}{},
			},
			wantNil: true,
		},
		{
			name: "empty integration_id",
			data: map[string]interface{}{
				"domain_id": "example.com",
				"subdomain": false,
				"imap_settings": []interface{}{
					map[string]interface{}{
						"integration_id": "",
					},
				},
			},
			wantNil: true,
		},
		{
			name: "valid integration_id",
			data: map[string]interface{}{
				"domain_id": "example.com",
				"subdomain": false,
				"imap_settings": []interface{}{
					map[string]interface{}{
						"integration_id": integrationID,
					},
				},
			},
			wantNil: false,
			wantID:  integrationID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.data)
			got := expandImapSettings(d)

			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %+v", got)
				}
				return
			}

			if got == nil || got.Integration == nil || got.Integration.Id == nil {
				t.Fatalf("expected IMAP settings with integration id, got %+v", got)
			}
			if *got.Integration.Id != tt.wantID {
				t.Fatalf("expected integration id %q, got %q", tt.wantID, *got.Integration.Id)
			}
		})
	}
}

func TestUnitCreateRoutingEmailDomain_AlreadyExists_AdoptsExisting(t *testing.T) {
	domainName := "acdemailplaysandboxeuscedee1"
	existingFullID := "acdemailplaysandboxeuscedee1.mypurecloud.com"
	subdomain := true

	// Mock proxy that simulates the "already exists" 400 error on create,
	// but returns the domain on lookup by name.
	mockProxy := &routingEmailDomainProxy{
		createRoutingEmailDomainAttr: func(_ context.Context, _ *routingEmailDomainProxy, _ *platformclientv2.Inbounddomaincreaterequest) (*platformclientv2.Inbounddomain, *platformclientv2.APIResponse, error) {
			return nil, &platformclientv2.APIResponse{
				StatusCode:   http.StatusBadRequest,
				ErrorMessage: "The inbound domain already exists (f24218ed-277b-441b-b6a6-b66c63a34c51)",
			}, fmt.Errorf("API Error: 400 - The inbound domain already exists")
		},
		getRoutingEmailDomainIdByNameAttr: func(_ context.Context, _ *routingEmailDomainProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
			return existingFullID, &platformclientv2.APIResponse{StatusCode: 200}, false, nil
		},
		getRoutingEmailDomainByIdAttr: func(_ context.Context, _ *routingEmailDomainProxy, id string) (*platformclientv2.Inbounddomain, *platformclientv2.APIResponse, error) {
			return &platformclientv2.Inbounddomain{
				Id:        &existingFullID,
				SubDomain: &subdomain,
			}, &platformclientv2.APIResponse{StatusCode: 200}, nil
		},
	}

	// Temporarily replace the internal proxy singleton
	oldProxy := internalProxy
	internalProxy = mockProxy
	defer func() { internalProxy = oldProxy }()

	resourceSchema := routingEmailDomainTestSchema()
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
		"domain_id": domainName,
		"subdomain": true,
	})

	meta := &provider.ProviderMeta{
		ClientConfig: &platformclientv2.Configuration{},
	}

	diags := createRoutingEmailDomain(context.Background(), d, meta)
	if diags != nil && diags.HasError() {
		t.Fatalf("expected no error, got diagnostics: %v", diags)
	}

	if d.Id() != existingFullID {
		t.Fatalf("expected resource ID to be %q, got %q", existingFullID, d.Id())
	}
}

func TestUnitCreateRoutingEmailDomain_AlreadyExists_LookupFails_ReturnsError(t *testing.T) {
	domainName := "acdemailplaysandboxeuscedee1"

	// Mock proxy that simulates the "already exists" 400 error on create,
	// AND fails the lookup with a non-retryable error.
	mockProxy := &routingEmailDomainProxy{
		createRoutingEmailDomainAttr: func(_ context.Context, _ *routingEmailDomainProxy, _ *platformclientv2.Inbounddomaincreaterequest) (*platformclientv2.Inbounddomain, *platformclientv2.APIResponse, error) {
			return nil, &platformclientv2.APIResponse{
				StatusCode:   http.StatusBadRequest,
				ErrorMessage: "The inbound domain already exists (f24218ed-277b-441b-b6a6-b66c63a34c51)",
			}, fmt.Errorf("API Error: 400 - The inbound domain already exists")
		},
		getRoutingEmailDomainIdByNameAttr: func(_ context.Context, _ *routingEmailDomainProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
			return "", &platformclientv2.APIResponse{StatusCode: 200}, false, fmt.Errorf("unable to find routing email domain with name %s", name)
		},
	}

	oldProxy := internalProxy
	internalProxy = mockProxy
	defer func() { internalProxy = oldProxy }()

	resourceSchema := routingEmailDomainTestSchema()
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
		"domain_id": domainName,
		"subdomain": true,
	})

	meta := &provider.ProviderMeta{
		ClientConfig: &platformclientv2.Configuration{},
	}

	diags := createRoutingEmailDomain(context.Background(), d, meta)
	if diags == nil || !diags.HasError() {
		t.Fatalf("expected error diagnostic when lookup fails, got none")
	}
}
