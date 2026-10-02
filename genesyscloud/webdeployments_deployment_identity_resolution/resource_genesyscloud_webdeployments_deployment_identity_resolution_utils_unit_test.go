package webdeployments_deployment_identity_resolution

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	externalSource "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/external_contacts_external_source"
	webDeployment "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/webdeployments_deployment"
	"github.com/stretchr/testify/assert"
)

func automerge(authenticatedWebMessaging, webTracking bool) *platformclientv2.Identityresolutionautomergeconfig {
	return &platformclientv2.Identityresolutionautomergeconfig{
		AuthenticatedWebMessaging: platformclientv2.Bool(authenticatedWebMessaging),
		WebTracking:               platformclientv2.Bool(webTracking),
	}
}

func TestUnitIsDefaultAutomergeConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   *platformclientv2.Identityresolutionautomergeconfig
		expected bool
	}{
		{name: "nil config", config: nil, expected: true},
		{name: "both off", config: automerge(false, false), expected: true},
		{name: "authenticated web messaging on", config: automerge(true, false), expected: false},
		{name: "web tracking on", config: automerge(false, true), expected: false},
		{name: "both on", config: automerge(true, true), expected: false},
		{
			name:     "both nil",
			config:   &platformclientv2.Identityresolutionautomergeconfig{},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, isDefaultAutomergeConfig(test.config))
		})
	}
}

func TestUnitIsDefaultIdentityResolutionConfig(t *testing.T) {
	star := "*"
	divisionId := uuid.NewString()
	externalSourceId := uuid.NewString()

	tests := []struct {
		name     string
		config   *platformclientv2.Deploymentidentityresolutionconfig
		expected bool
	}{
		{name: "nil config", config: nil, expected: true},
		{
			name: "resolve true, nothing else set",
			config: &platformclientv2.Deploymentidentityresolutionconfig{
				ResolveIdentities: platformclientv2.Bool(true),
			},
			expected: true,
		},
		{
			name: "resolve nil is treated as the default",
			config: &platformclientv2.Deploymentidentityresolutionconfig{
				Automerge: automerge(false, false),
			},
			expected: true,
		},
		{
			name: "division returned as the unassigned star sentinel",
			config: &platformclientv2.Deploymentidentityresolutionconfig{
				ResolveIdentities: platformclientv2.Bool(true),
				Division:          &platformclientv2.Writablestarrabledivision{Id: &star},
				Automerge:         automerge(false, false),
			},
			expected: true,
		},
		{
			name: "resolve false",
			config: &platformclientv2.Deploymentidentityresolutionconfig{
				ResolveIdentities: platformclientv2.Bool(false),
			},
			expected: false,
		},
		{
			name: "specific division",
			config: &platformclientv2.Deploymentidentityresolutionconfig{
				ResolveIdentities: platformclientv2.Bool(true),
				Division:          &platformclientv2.Writablestarrabledivision{Id: &divisionId},
			},
			expected: false,
		},
		{
			name: "external source set",
			config: &platformclientv2.Deploymentidentityresolutionconfig{
				ResolveIdentities: platformclientv2.Bool(true),
				ExternalSource:    &platformclientv2.Identityresolutionexternalsource{Id: &externalSourceId},
			},
			expected: false,
		},
		{
			name: "automerge enabled for one channel",
			config: &platformclientv2.Deploymentidentityresolutionconfig{
				ResolveIdentities: platformclientv2.Bool(true),
				Automerge:         automerge(false, true),
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, isDefaultIdentityResolutionConfig(test.config))
		})
	}
}

// TestUnitBuildDeploymentIdentityResolutionConfig pins the PUT contract: only real
// values are sent, and the unassigned division sentinel is never serialised as an
// empty division object (contacts-service answers 422).
func TestUnitBuildDeploymentIdentityResolutionConfig(t *testing.T) {
	divisionId := uuid.NewString()
	externalSourceId := uuid.NewString()

	tests := []struct {
		name                    string
		divisionId              string
		externalSourceId        string
		expectDivisionNil       bool
		expectExternalSourceNil bool
	}{
		{name: "both set", divisionId: divisionId, externalSourceId: externalSourceId},
		{name: "both omitted", expectDivisionNil: true, expectExternalSourceNil: true},
		{name: "division explicit star", divisionId: "*", expectDivisionNil: true, expectExternalSourceNil: true},
		{name: "division only", divisionId: divisionId, expectExternalSourceNil: true},
		{name: "external source only", externalSourceId: externalSourceId, expectDivisionNil: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resourceMap := map[string]interface{}{
				"deployment_id":      uuid.NewString(),
				"resolve_identities": false,
			}
			if test.divisionId != "" {
				resourceMap["division_id"] = test.divisionId
			}
			if test.externalSourceId != "" {
				resourceMap["external_source_id"] = test.externalSourceId
			}

			d := schema.TestResourceDataRaw(t, ResourceWebDeploymentIdentityResolution().Schema, resourceMap)

			config, err := buildDeploymentIdentityResolutionConfig(d)
			assert.NoError(t, err)
			assert.NotNil(t, config.ResolveIdentities)
			assert.False(t, *config.ResolveIdentities)

			if test.expectDivisionNil {
				assert.Nil(t, config.Division)
			} else {
				assert.NotNil(t, config.Division)
				assert.Equal(t, test.divisionId, *config.Division.Id)
			}

			if test.expectExternalSourceNil {
				assert.Nil(t, config.ExternalSource)
			} else {
				assert.NotNil(t, config.ExternalSource)
				assert.Equal(t, test.externalSourceId, *config.ExternalSource.Id)
			}
		})
	}
}

// TestUnitBuildAutomergeConfig also guards the schema keys: a typo in either map key
// silently sends false instead of the configured value.
func TestUnitBuildAutomergeConfig(t *testing.T) {
	tests := []struct {
		name                          string
		block                         []interface{}
		expectAuthenticatedWebMessage bool
		expectWebTracking             bool
	}{
		{
			name:                          "block omitted defaults to automerging off",
			block:                         []interface{}{},
			expectAuthenticatedWebMessage: false,
			expectWebTracking:             false,
		},
		{
			name: "both flags on",
			block: []interface{}{map[string]interface{}{
				"authenticated_web_messaging": true,
				"web_tracking":                true,
			}},
			expectAuthenticatedWebMessage: true,
			expectWebTracking:             true,
		},
		{
			name: "authenticated web messaging on only",
			block: []interface{}{map[string]interface{}{
				"authenticated_web_messaging": true,
				"web_tracking":                false,
			}},
			expectAuthenticatedWebMessage: true,
			expectWebTracking:             false,
		},
		{
			name: "web tracking on only",
			block: []interface{}{map[string]interface{}{
				"authenticated_web_messaging": false,
				"web_tracking":                true,
			}},
			expectAuthenticatedWebMessage: false,
			expectWebTracking:             true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config, err := buildAutomergeConfig(test.block)
			assert.NoError(t, err)
			assert.NotNil(t, config)
			assert.NotNil(t, config.AuthenticatedWebMessaging)
			assert.NotNil(t, config.WebTracking)
			assert.Equal(t, test.expectAuthenticatedWebMessage, *config.AuthenticatedWebMessaging)
			assert.Equal(t, test.expectWebTracking, *config.WebTracking)
		})
	}
}

func TestUnitBuildDefaultDeploymentIdentityResolutionConfig(t *testing.T) {
	config := buildDefaultDeploymentIdentityResolutionConfig()

	assert.NotNil(t, config.ResolveIdentities)
	assert.True(t, *config.ResolveIdentities, "destroy must never PUT resolve_identities=false")
	assert.Nil(t, config.Division)
	assert.Nil(t, config.ExternalSource)
	assert.NotNil(t, config.Automerge, "automerge is reset explicitly rather than relying on PUT-replace")
	assert.False(t, *config.Automerge.AuthenticatedWebMessaging)
	assert.False(t, *config.Automerge.WebTracking)
}

// TestUnitFlattenAutomergeConfig is the convergence contract. The API returns an
// all-false automerge object for every deployment, configured or not, so the block is
// only written to state when the resource already has one or automerging is actually on.
func TestUnitFlattenAutomergeConfig(t *testing.T) {
	tests := []struct {
		name                          string
		config                        *platformclientv2.Identityresolutionautomergeconfig
		declared                      bool
		expectOmitted                 bool
		expectAuthenticatedWebMessage bool
		expectWebTracking             bool
	}{
		{name: "nil config, block undeclared", config: nil, declared: false, expectOmitted: true},
		{name: "all off, block undeclared", config: automerge(false, false), declared: false, expectOmitted: true},
		{
			name:              "web tracking on, block undeclared is surfaced as drift",
			config:            automerge(false, true),
			declared:          false,
			expectWebTracking: true,
		},
		{
			name:                          "both on, block undeclared is surfaced as drift",
			config:                        automerge(true, true),
			declared:                      false,
			expectAuthenticatedWebMessage: true,
			expectWebTracking:             true,
		},
		{
			name:          "all off, block declared is kept",
			config:        automerge(false, false),
			declared:      true,
			expectOmitted: false,
		},
		{
			name:                          "authenticated web messaging on, block declared",
			config:                        automerge(true, false),
			declared:                      true,
			expectAuthenticatedWebMessage: true,
		},
		{
			name:          "nil config, block declared is kept as all off",
			config:        nil,
			declared:      true,
			expectOmitted: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flattened := flattenAutomergeConfig(test.config, test.declared)

			if test.expectOmitted {
				assert.Nil(t, flattened)
				return
			}

			assert.Len(t, flattened, 1)
			assert.Equal(t, test.expectAuthenticatedWebMessage, flattened[0]["authenticated_web_messaging"])
			assert.Equal(t, test.expectWebTracking, flattened[0]["web_tracking"])
		})
	}
}

// TestUnitPlanConvergence drives the SDK's own diff to prove each state/config pair a
// read can produce plans empty. The last case is what the sticky flatten prevents.
func TestUnitPlanConvergence(t *testing.T) {
	r := ResourceWebDeploymentIdentityResolution()

	tests := []struct {
		name       string
		state      map[string]string
		config     map[string]interface{}
		expectPlan bool
	}{
		{
			name: "no automerge block on either side",
			state: map[string]string{
				"deployment_id": "dep-1", "resolve_identities": "true", "automerge_config.#": "0",
			},
			config: map[string]interface{}{"deployment_id": "dep-1", "resolve_identities": true},
		},
		{
			name: "explicit all-false block is kept in state",
			state: map[string]string{
				"deployment_id": "dep-1", "resolve_identities": "true", "automerge_config.#": "1",
				"automerge_config.0.authenticated_web_messaging": "false",
				"automerge_config.0.web_tracking":                "false",
			},
			config: map[string]interface{}{
				"deployment_id": "dep-1", "resolve_identities": true,
				"automerge_config": []interface{}{map[string]interface{}{
					"authenticated_web_messaging": false, "web_tracking": false,
				}},
			},
		},
		{
			name: "one flag on",
			state: map[string]string{
				"deployment_id": "dep-1", "resolve_identities": "true", "automerge_config.#": "1",
				"automerge_config.0.authenticated_web_messaging": "false",
				"automerge_config.0.web_tracking":                "true",
			},
			config: map[string]interface{}{
				"deployment_id": "dep-1", "resolve_identities": true,
				"automerge_config": []interface{}{map[string]interface{}{
					"authenticated_web_messaging": false, "web_tracking": true,
				}},
			},
		},
		{
			name: "explicit star division against a cleared state value",
			state: map[string]string{
				"deployment_id": "dep-1", "resolve_identities": "true",
				"division_id": "", "automerge_config.#": "0",
			},
			config: map[string]interface{}{
				"deployment_id": "dep-1", "resolve_identities": true, "division_id": "*",
			},
		},
		{
			name: "undeclared block present in state still plans a removal",
			state: map[string]string{
				"deployment_id": "dep-1", "resolve_identities": "true", "automerge_config.#": "1",
				"automerge_config.0.authenticated_web_messaging": "false",
				"automerge_config.0.web_tracking":                "false",
			},
			config:     map[string]interface{}{"deployment_id": "dep-1", "resolve_identities": true},
			expectPlan: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := &terraform.InstanceState{ID: "dep-1", Attributes: test.state}
			diff, err := r.SimpleDiff(context.Background(), state, terraform.NewResourceConfigRaw(test.config), nil)
			assert.NoError(t, err)

			hasPlan := diff != nil && len(diff.Attributes) > 0
			assert.Equal(t, test.expectPlan, hasPlan, "diff: %v", diff)
		})
	}
}

// TestUnitAuthenticatedWebMessagingDowngraded pins the one silent normalization this
// resource guards against. The platform answers 200 and stores false, so this is the
// only place the refusal is visible.
func TestUnitAuthenticatedWebMessagingDowngraded(t *testing.T) {
	requestedOn := &platformclientv2.Deploymentidentityresolutionconfig{
		ResolveIdentities: platformclientv2.Bool(true),
		Automerge:         automerge(true, true),
	}

	tests := []struct {
		name      string
		requested *platformclientv2.Deploymentidentityresolutionconfig
		stored    *platformclientv2.Deploymentidentityresolutionconfig
		expected  bool
	}{
		{
			name:      "requested on and stored on",
			requested: requestedOn,
			stored: &platformclientv2.Deploymentidentityresolutionconfig{
				Automerge: automerge(true, true),
			},
			expected: false,
		},
		{
			name:      "requested on but stored off",
			requested: requestedOn,
			stored: &platformclientv2.Deploymentidentityresolutionconfig{
				Automerge: automerge(false, true),
			},
			expected: true,
		},
		{
			name:      "requested on but whole automerge object dropped",
			requested: requestedOn,
			stored: &platformclientv2.Deploymentidentityresolutionconfig{
				ResolveIdentities: platformclientv2.Bool(true),
			},
			expected: true,
		},
		{
			name: "requested off is never a downgrade",
			requested: &platformclientv2.Deploymentidentityresolutionconfig{
				Automerge: automerge(false, true),
			},
			stored: &platformclientv2.Deploymentidentityresolutionconfig{
				Automerge: automerge(false, true),
			},
			expected: false,
		},
		{
			name: "no automerge requested",
			requested: &platformclientv2.Deploymentidentityresolutionconfig{
				ResolveIdentities: platformclientv2.Bool(true),
			},
			stored:   &platformclientv2.Deploymentidentityresolutionconfig{},
			expected: false,
		},
		{name: "nil requested", requested: nil, stored: requestedOn, expected: false},
		{name: "nil stored", requested: requestedOn, stored: nil, expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, authenticatedWebMessagingDowngraded(test.requested, test.stored))
		})
	}
}

func TestUnitWebDeploymentIdentityResolutionExporterRefAttrs(t *testing.T) {
	exporter := WebDeploymentIdentityResolutionExporter()

	assert.Equal(t, webDeployment.ResourceType, exporter.RefAttrs["deployment_id"].RefType)
	assert.Equal(t, authDivision.ResourceType, exporter.RefAttrs["division_id"].RefType)
	assert.Equal(t, []string{"*"}, exporter.RefAttrs["division_id"].AltValues)
	assert.Equal(t, externalSource.ResourceType, exporter.RefAttrs["external_source_id"].RefType)
}

func TestUnitIsUnassignedDivisionId(t *testing.T) {
	assert.True(t, isUnassignedDivisionId(""))
	assert.True(t, isUnassignedDivisionId("*"))
	assert.False(t, isUnassignedDivisionId(uuid.NewString()))
}

func TestUnitSuppressUnassignedDivisionIdDiff(t *testing.T) {
	tests := []struct {
		name     string
		old, new string
		suppress bool
	}{
		{name: "empty to *", old: "", new: "*", suppress: true},
		{name: "* to empty", old: "*", new: "", suppress: true},
		{name: "empty to empty", old: "", new: "", suppress: true},
		{name: "* to *", old: "*", new: "*", suppress: true},
		{name: "empty to uuid", old: "", new: uuid.NewString(), suppress: false},
		{name: "uuid to *", old: uuid.NewString(), new: "*", suppress: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.suppress, suppressUnassignedDivisionIdDiff("division_id", test.old, test.new, nil))
		})
	}
}

func TestUnitResourceWebDeploymentIdentityResolutionSchema(t *testing.T) {
	r := ResourceWebDeploymentIdentityResolution()
	assert.NoError(t, r.InternalValidate(nil, true))

	automergeSchema := r.Schema["automerge_config"].Elem.(*schema.Resource).Schema
	assert.True(t, automergeSchema["authenticated_web_messaging"].Required,
		"both flags are required inside the block so omitted never means false by accident")
	assert.True(t, automergeSchema["web_tracking"].Required)
	assert.False(t, r.Schema["automerge_config"].Computed,
		"automerge_config must not be Computed or removing the block can never reset it")
}
