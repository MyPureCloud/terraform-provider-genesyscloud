package routing_queue

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestUnitStateUpgraderRoutingQueueV2ToV3(t *testing.T) {
	memberSet := schema.NewSet(schema.HashResource(queueMemberResource), []interface{}{
		map[string]interface{}{"user_id": "user-b", "ring_num": 1},
		map[string]interface{}{"user_id": "user-a", "ring_num": 2},
	})
	wrapupSet := schema.NewSet(schema.HashString, []interface{}{"code-z", "code-a"})

	rawState := map[string]interface{}{
		"members":      memberSet,
		"wrapup_codes": wrapupSet,
	}

	upgraded, err := stateUpgraderRoutingQueueV2ToV3(context.Background(), rawState, nil)
	assert.NoError(t, err)

	members, ok := upgraded["members"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, members, 2)

	wrapupCodes, ok := upgraded["wrapup_codes"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{"code-a", "code-z"}, wrapupCodes)
}

func TestUnitOrganizeStringIdsForRead(t *testing.T) {
	schemaList := []string{"a", "b", "c"}
	apiList := []string{"c", "a", "b"}
	assert.Equal(t, schemaList, organizeStringIdsForRead(schemaList, apiList))

	apiChanged := []string{"a", "b"}
	assert.Equal(t, apiChanged, organizeStringIdsForRead(schemaList, apiChanged))
}

func TestUnitOrganizeMembersForRead(t *testing.T) {
	schemaMembers := []interface{}{
		map[string]interface{}{"user_id": "u1", "ring_num": 1},
		map[string]interface{}{"user_id": "u2", "ring_num": 2},
	}
	apiMembers := []interface{}{
		map[string]interface{}{"user_id": "u2", "ring_num": 2},
		map[string]interface{}{"user_id": "u1", "ring_num": 1},
	}
	assert.Equal(t, schemaMembers, organizeMembersForRead(schemaMembers, apiMembers))

	apiChanged := []interface{}{
		map[string]interface{}{"user_id": "u1", "ring_num": 1},
	}
	assert.Equal(t, apiChanged, organizeMembersForRead(schemaMembers, apiChanged))
}

func TestUnitOrganizeAllOutboundEmailAddressesForRead(t *testing.T) {
	schemaAddresses := []interface{}{
		map[string]interface{}{"domain_id": "d.example.com", "route_id": "r1"},
		map[string]interface{}{"domain_id": "d.example.com", "route_id": "r2"},
	}

	// Same set, API returns a different order -> keep config order (no perpetual diff).
	apiReordered := []interface{}{
		map[string]interface{}{"domain_id": "d.example.com", "route_id": "r2"},
		map[string]interface{}{"domain_id": "d.example.com", "route_id": "r1"},
	}
	assert.Equal(t, schemaAddresses, organizeAllOutboundEmailAddressesForRead(schemaAddresses, apiReordered))

	// Set actually changed -> use the API result.
	apiChanged := []interface{}{
		map[string]interface{}{"domain_id": "d.example.com", "route_id": "r1"},
	}
	assert.Equal(t, apiChanged, organizeAllOutboundEmailAddressesForRead(schemaAddresses, apiChanged))
}
