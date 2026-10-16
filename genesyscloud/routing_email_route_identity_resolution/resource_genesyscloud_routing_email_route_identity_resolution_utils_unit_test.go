package routing_email_route_identity_resolution

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitIsDefaultIdentityResolutionConfig(t *testing.T) {
	resolveTrue := true
	resolveFalse := false
	star := "*"
	divisionId := uuid.NewString()

	tests := []struct {
		name     string
		config   *platformclientv2.Routeidentityresolutionconfig
		expected bool
	}{
		{
			name: "resolve true without division",
			config: &platformclientv2.Routeidentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
			},
			expected: true,
		},
		{
			name: "resolve true with division *",
			config: &platformclientv2.Routeidentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
				Division:          &platformclientv2.Writablestarrabledivision{Id: &star},
			},
			expected: true,
		},
		{
			name: "resolve false",
			config: &platformclientv2.Routeidentityresolutionconfig{
				ResolveIdentities: &resolveFalse,
			},
			expected: false,
		},
		{
			name: "specific division",
			config: &platformclientv2.Routeidentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
				Division:          &platformclientv2.Writablestarrabledivision{Id: &divisionId},
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

func TestUnitRoutingEmailRouteIdentityResolutionExporterRefAttrs(t *testing.T) {
	exporter := RoutingEmailRouteIdentityResolutionExporter()

	assert.Equal(t, "genesyscloud_routing_email_domain", exporter.RefAttrs["domain_name"].RefType)
	assert.Equal(t, "genesyscloud_routing_email_route", exporter.RefAttrs["route_id"].RefType)
	assert.Equal(t, "genesyscloud_auth_division", exporter.RefAttrs["division_id"].RefType)
	assert.Equal(t, []string{"*"}, exporter.RefAttrs["division_id"].AltValues)
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
