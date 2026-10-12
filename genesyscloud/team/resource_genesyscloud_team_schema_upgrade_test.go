package team

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestUnitUpgradeTeamStateV1(t *testing.T) {
	memberSet := schema.NewSet(schema.HashString, []interface{}{"user-z", "user-a"})
	rawState := map[string]interface{}{
		"member_ids": memberSet,
	}

	upgraded, err := upgradeTeamStateV1(context.Background(), rawState, nil)
	assert.NoError(t, err)

	memberIds, ok := upgraded["member_ids"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{"user-a", "user-z"}, memberIds)
}

func TestUnitOrganizeMemberIdsForRead(t *testing.T) {
	schemaList := []string{"a", "b", "c"}
	apiList := []string{"c", "a", "b"}
	assert.Equal(t, schemaList, organizeMemberIdsForRead(schemaList, apiList))

	apiChanged := []string{"a", "b"}
	assert.Equal(t, apiChanged, organizeMemberIdsForRead(schemaList, apiChanged))
}
