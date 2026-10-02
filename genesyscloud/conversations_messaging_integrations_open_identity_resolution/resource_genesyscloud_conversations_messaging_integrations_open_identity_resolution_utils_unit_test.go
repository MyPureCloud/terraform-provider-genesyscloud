package conversations_messaging_integrations_open_identity_resolution

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
	externalSourceId := uuid.NewString()

	tests := []struct {
		name     string
		config   *platformclientv2.Openmessagingidentityresolutionconfig
		expected bool
	}{
		{
			name: "resolve true without division and external source",
			config: &platformclientv2.Openmessagingidentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
			},
			expected: true,
		},
		{
			name: "resolve true with division * and without external source",
			config: &platformclientv2.Openmessagingidentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
				Division:          &platformclientv2.Writablestarrabledivision{Id: &star},
			},
			expected: true,
		},
		{
			name: "resolve true with external source",
			config: &platformclientv2.Openmessagingidentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
				ExternalSource:    &platformclientv2.Identityresolutionexternalsource{Id: &externalSourceId},
			},
			expected: false,
		},
		{
			name: "resolve true with division * and external source",
			config: &platformclientv2.Openmessagingidentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
				Division:          &platformclientv2.Writablestarrabledivision{Id: &star},
				ExternalSource:    &platformclientv2.Identityresolutionexternalsource{Id: &externalSourceId},
			},
			expected: false,
		},
		{
			name: "resolve false",
			config: &platformclientv2.Openmessagingidentityresolutionconfig{
				ResolveIdentities: &resolveFalse,
			},
			expected: false,
		},
		{
			name: "specific division and external source",
			config: &platformclientv2.Openmessagingidentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
				Division:          &platformclientv2.Writablestarrabledivision{Id: &divisionId},
				ExternalSource:    &platformclientv2.Identityresolutionexternalsource{Id: &externalSourceId},
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

func TestUnitConversationsMessagingIntegrationsIdentityResolutionExporterRefAttrs(t *testing.T) {
	exporter := ConversationsMessagingIntegrationsOpenIdentityResolutionExporter()

	assert.Equal(t, "genesyscloud_conversations_messaging_integrations_open", exporter.RefAttrs["open_integration_id"].RefType)
	assert.Equal(t, "genesyscloud_auth_division", exporter.RefAttrs["division_id"].RefType)
	assert.Equal(t, []string{"*"}, exporter.RefAttrs["division_id"].AltValues)
	assert.Equal(t, "genesyscloud_externalcontacts_external_source", exporter.RefAttrs["external_source_id"].RefType)
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
