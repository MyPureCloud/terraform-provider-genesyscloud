package routing_queue_identity_resolution

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitIsDefaultIdentityResolutionConfigTrue(t *testing.T) {
	resolveIdentities := true
	config := platformclientv2.Identityresolutionqueueconfig{
		CallOnBehalfOfQueue: &platformclientv2.Outboundqueueidentityresolutionconfig{
			ResolveIdentities: &resolveIdentities,
		},
	}
	assert.True(t, isDefaultIdentityResolutionConfig(&config))
}

func TestUnitIsDefaultIdentityResolutionConfigFalseResolve(t *testing.T) {
	resolveIdentities := false
	config := platformclientv2.Identityresolutionqueueconfig{
		CallOnBehalfOfQueue: &platformclientv2.Outboundqueueidentityresolutionconfig{
			ResolveIdentities: &resolveIdentities,
		},
	}
	assert.False(t, isDefaultIdentityResolutionConfig(&config))
}

func TestUnitIsDefaultIdentityResolutionConfigFalseDivision(t *testing.T) {
	resolveIdentities := true
	divisionId := uuid.NewString()
	config := platformclientv2.Identityresolutionqueueconfig{
		CallOnBehalfOfQueue: &platformclientv2.Outboundqueueidentityresolutionconfig{
			ResolveIdentities: &resolveIdentities,
			Division: &platformclientv2.Writablestarrabledivision{
				Id: &divisionId,
			},
		},
	}
	assert.False(t, isDefaultIdentityResolutionConfig(&config))
}

func TestUnitBuildCallOnBehalfOfQueueConfig(t *testing.T) {
	divisionId := uuid.NewString()
	blocks := []interface{}{
		map[string]interface{}{
			"resolve_identities": false,
			"division_id":        divisionId,
		},
	}

	config, err := buildCallOnBehalfOfQueueConfig(blocks)
	assert.NoError(t, err)
	assert.NotNil(t, config.ResolveIdentities)
	assert.False(t, *config.ResolveIdentities)
	assert.NotNil(t, config.Division)
	assert.Equal(t, divisionId, *config.Division.Id)
}

func TestUnitBuildCallOnBehalfOfQueueConfigAllDivisions(t *testing.T) {
	blocks := []interface{}{
		map[string]interface{}{
			"resolve_identities": false,
		},
	}

	config, err := buildCallOnBehalfOfQueueConfig(blocks)
	assert.NoError(t, err)
	assert.Nil(t, config.Division)
}

func TestUnitFlattenCallOnBehalfOfQueue(t *testing.T) {
	resolveIdentities := true
	divisionId := "*"
	config := &platformclientv2.Outboundqueueidentityresolutionconfig{
		ResolveIdentities: &resolveIdentities,
		Division: &platformclientv2.Writablestarrabledivision{
			Id: &divisionId,
		},
	}

	flattened := flattenCallOnBehalfOfQueue(config)
	assert.Len(t, flattened, 1)
	block := flattened[0].(map[string]interface{})
	assert.Equal(t, true, block["resolve_identities"])
	_, hasDivision := block["division_id"]
	assert.False(t, hasDivision)
}

func TestUnitFlattenCallOnBehalfOfQueueNil(t *testing.T) {
	flattened := flattenCallOnBehalfOfQueue(nil)
	assert.Len(t, flattened, 1)
	block := flattened[0].(map[string]interface{})
	assert.Equal(t, true, block["resolve_identities"])
}

func TestUnitRoutingQueueIdentityResolutionExporterRefAttrs(t *testing.T) {
	exporter := RoutingQueueIdentityResolutionExporter()

	assert.Equal(t, "genesyscloud_routing_queue", exporter.RefAttrs["queue_id"].RefType)
	assert.Equal(t, "genesyscloud_auth_division", exporter.RefAttrs["call_on_behalf_of_queue.division_id"].RefType)
	assert.Equal(t, []string{"*"}, exporter.RefAttrs["call_on_behalf_of_queue.division_id"].AltValues)
}
