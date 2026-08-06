package routing_queue_identity_resolution

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitIsDefaultIdentityResolutionConfig(t *testing.T) {
	resolveTrue := true
	resolveFalse := false
	star := "*"
	divisionId := uuid.NewString()

	tests := []struct {
		name     string
		config   *platformclientv2.Identityresolutionqueueconfig
		expected bool
	}{
		{
			name: "resolve true without division",
			config: &platformclientv2.Identityresolutionqueueconfig{
				CallOnBehalfOfQueue: &platformclientv2.Outboundqueueidentityresolutionconfig{
					ResolveIdentities: &resolveTrue,
				},
			},
			expected: true,
		},
		{
			name: "resolve true with division *",
			config: &platformclientv2.Identityresolutionqueueconfig{
				CallOnBehalfOfQueue: &platformclientv2.Outboundqueueidentityresolutionconfig{
					ResolveIdentities: &resolveTrue,
					Division:          &platformclientv2.Writablestarrabledivision{Id: &star},
				},
			},
			expected: true,
		},
		{
			name: "resolve false",
			config: &platformclientv2.Identityresolutionqueueconfig{
				CallOnBehalfOfQueue: &platformclientv2.Outboundqueueidentityresolutionconfig{
					ResolveIdentities: &resolveFalse,
				},
			},
			expected: false,
		},
		{
			name: "specific division",
			config: &platformclientv2.Identityresolutionqueueconfig{
				CallOnBehalfOfQueue: &platformclientv2.Outboundqueueidentityresolutionconfig{
					ResolveIdentities: &resolveTrue,
					Division:          &platformclientv2.Writablestarrabledivision{Id: &divisionId},
				},
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

func TestUnitBuildCallOnBehalfOfQueueConfig(t *testing.T) {
	divisionId := uuid.NewString()

	tests := []struct {
		name               string
		divisionId         string
		expectDivisionNil  bool
		expectedDivisionId string
	}{
		{name: "specific division", divisionId: divisionId, expectedDivisionId: divisionId},
		{name: "omitted division", expectDivisionNil: true},
		{name: "explicit *", divisionId: "*", expectDivisionNil: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			block := map[string]interface{}{
				"resolve_identities": false,
			}
			if test.divisionId != "" {
				block["division_id"] = test.divisionId
			}

			config, err := buildCallOnBehalfOfQueueConfig([]interface{}{block})
			assert.NoError(t, err)
			assert.NotNil(t, config.ResolveIdentities)
			assert.False(t, *config.ResolveIdentities)

			if test.expectDivisionNil {
				assert.Nil(t, config.Division)
				return
			}
			assert.NotNil(t, config.Division)
			assert.Equal(t, test.expectedDivisionId, *config.Division.Id)
		})
	}
}

func TestUnitFlattenCallOnBehalfOfQueue(t *testing.T) {
	resolveTrue := true
	star := "*"
	divisionId := uuid.NewString()

	tests := []struct {
		name               string
		config             *platformclientv2.Outboundqueueidentityresolutionconfig
		expectResolve      bool
		expectDivisionId   string
		expectDivisionOmit bool
	}{
		{
			name:               "nil config defaults resolve true",
			config:             nil,
			expectResolve:      true,
			expectDivisionOmit: true,
		},
		{
			name: "nil resolve_identities defaults to true",
			config: &platformclientv2.Outboundqueueidentityresolutionconfig{
				ResolveIdentities: nil,
			},
			expectResolve:      true,
			expectDivisionOmit: true,
		},
		{
			name: "division * omitted from state",
			config: &platformclientv2.Outboundqueueidentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
				Division:          &platformclientv2.Writablestarrabledivision{Id: &star},
			},
			expectResolve:      true,
			expectDivisionOmit: true,
		},
		{
			name: "specific division kept",
			config: &platformclientv2.Outboundqueueidentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
				Division:          &platformclientv2.Writablestarrabledivision{Id: &divisionId},
			},
			expectResolve:    true,
			expectDivisionId: divisionId,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flattened := flattenCallOnBehalfOfQueue(test.config)
			assert.Len(t, flattened, 1)
			block := flattened[0].(map[string]interface{})
			assert.Equal(t, test.expectResolve, block["resolve_identities"])

			divisionId, hasDivision := block["division_id"]
			if test.expectDivisionOmit {
				assert.False(t, hasDivision)
				return
			}
			assert.True(t, hasDivision)
			assert.Equal(t, test.expectDivisionId, divisionId)
		})
	}
}

func TestUnitRoutingQueueIdentityResolutionExporterRefAttrs(t *testing.T) {
	exporter := RoutingQueueIdentityResolutionExporter()

	assert.Equal(t, "genesyscloud_routing_queue", exporter.RefAttrs["queue_id"].RefType)
	assert.Equal(t, "genesyscloud_auth_division", exporter.RefAttrs["call_on_behalf_of_queue.division_id"].RefType)
	assert.Equal(t, []string{"*"}, exporter.RefAttrs["call_on_behalf_of_queue.division_id"].AltValues)
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
