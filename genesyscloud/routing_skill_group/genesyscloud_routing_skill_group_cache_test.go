package routing_skill_group

import (
	"context"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitGetRoutingSkillGroupsByIdCacheHit(t *testing.T) {
	tfexporter_state.ActivateExporterState()

	skillGroupID := "skill-group-cache-test"
	cached := platformclientv2.Skillgroup{
		Id:   platformclientv2.String(skillGroupID),
		Name: platformclientv2.String("Cached Skill Group"),
	}
	rc.SetCache(skillGroupCache, skillGroupID, cached)

	skillGroup, _, err := getRoutingSkillGroupsByIdFn(context.Background(), &routingSkillGroupsProxy{skillGroupCache: skillGroupCache}, skillGroupID)
	require.NoError(t, err)
	require.NotNil(t, skillGroup)
	assert.Equal(t, "Cached Skill Group", *skillGroup.Name)
}

func TestUnitGetAllRoutingSkillGroupsListCacheHit(t *testing.T) {
	tfexporter_state.ActivateExporterState()

	listKey := ""
	cached := []platformclientv2.Skillgroupdefinition{
		{Id: platformclientv2.String("skill-group-list-1"), Name: platformclientv2.String("Skill Group 1")},
	}
	rc.SetCache(skillGroupListCache, listKey, cached)

	skillGroups, _, err := getAllRoutingSkillGroupsFn(context.Background(), &routingSkillGroupsProxy{}, "")
	require.NoError(t, err)
	require.Len(t, *skillGroups, 1)
	assert.Equal(t, "skill-group-list-1", *(*skillGroups)[0].Id)
}

func TestUnitGetRoutingSkillGroupsMemberDivisionsCacheHit(t *testing.T) {
	tfexporter_state.ActivateExporterState()

	skillGroupID := "skill-group-member-divisions-cache-test"
	cached := platformclientv2.Skillgroupmemberdivisionlist{
		Entities: &[]platformclientv2.Division{
			{Id: platformclientv2.String("division-1")},
			{Id: platformclientv2.String("division-2")},
		},
	}
	rc.SetCache(skillGroupMemberDivisionsCache, skillGroupID, cached)

	memberDivisions, _, err := getRoutingSkillGroupsMemberDivisonFn(context.Background(), &routingSkillGroupsProxy{}, skillGroupID)
	require.NoError(t, err)
	require.NotNil(t, memberDivisions)
	require.NotNil(t, memberDivisions.Entities)
	require.Len(t, *memberDivisions.Entities, 2)
	assert.Equal(t, "division-1", *(*memberDivisions.Entities)[0].Id)
}

func TestUnitStoreSkillGroupInCache(t *testing.T) {
	tfexporter_state.ActivateExporterState()

	skillGroupID := "skill-group-write-through-test"
	skillGroup := platformclientv2.Skillgroup{
		Id:   platformclientv2.String(skillGroupID),
		Name: platformclientv2.String("Write Through Skill Group"),
	}

	storeSkillGroupInCache(&skillGroup)

	cached := rc.GetCacheItem(skillGroupCache, skillGroupID)
	require.NotNil(t, cached)
	assert.Equal(t, "Write Through Skill Group", *cached.Name)
}
