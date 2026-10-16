package architect_ivr_identity_resolution

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
		config   *platformclientv2.Ivridentityresolutionconfig
		expected bool
	}{
		{
			name: "resolve true without division",
			config: &platformclientv2.Ivridentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
			},
			expected: true,
		},
		{
			name: "resolve true with division *",
			config: &platformclientv2.Ivridentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
				Division:          &platformclientv2.Writablestarrabledivision{Id: &star},
			},
			expected: true,
		},
		{
			name: "resolve false",
			config: &platformclientv2.Ivridentityresolutionconfig{
				ResolveIdentities: &resolveFalse,
			},
			expected: false,
		},
		{
			name: "specific division",
			config: &platformclientv2.Ivridentityresolutionconfig{
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

func TestUnitArchitectIvrIdentityResolutionExporterRefAttrs(t *testing.T) {
	exporter := ArchitectIvrIdentityResolutionExporter()

	assert.Equal(t, "genesyscloud_architect_ivr", exporter.RefAttrs["ivr_id"].RefType)
	assert.Equal(t, "genesyscloud_auth_division", exporter.RefAttrs["division_id"].RefType)
	assert.Equal(t, []string{"*"}, exporter.RefAttrs["division_id"].AltValues)
}

func TestUnitSuppressUnassignedDivisionDiff(t *testing.T) {
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
