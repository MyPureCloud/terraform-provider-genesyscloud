package routing_queue

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitFlattenBullseyeRings_IncludesMemberGroupsOnSixthRing(t *testing.T) {
	id1 := "sg-ring1-a"
	id2 := "sg-ring1-b"
	id3 := "sg-ring1-c"
	id4 := "sg-ring1-d"
	id5 := "sg-ring1-e"
	id6 := "sg-ring6"
	sg := "SKILLGROUP"
	timeouts := []float64{10, 20, 2, 2, 7200, 2}

	rings := make([]platformclientv2.Ring, 6)
	for i := 0; i < 6; i++ {
		t := timeouts[i]
		rings[i] = platformclientv2.Ring{
			ExpansionCriteria: &[]platformclientv2.Expansioncriterium{{
				VarType:   &bullseyeExpansionTypeTimeout,
				Threshold: &t,
			}},
		}
	}
	rings[0].MemberGroups = &[]platformclientv2.Membergroup{
		{Id: &id1, VarType: &sg},
		{Id: &id2, VarType: &sg},
		{Id: &id3, VarType: &sg},
		{Id: &id4, VarType: &sg},
		{Id: &id5, VarType: &sg},
	}
	rings[5].MemberGroups = &[]platformclientv2.Membergroup{
		{Id: &id6, VarType: &sg},
	}

	flattened := flattenBullseyeRings(&rings)
	assert.Equal(t, 6, len(flattened))

	ring0 := flattened[0].(map[string]interface{})
	ring5 := flattened[5].(map[string]interface{})
	mg0 := ring0["member_groups"].(*schema.Set)
	mg5 := ring5["member_groups"].(*schema.Set)
	assert.Equal(t, 5, mg0.Len(), "ring 1 should have 5 skill groups")
	assert.Equal(t, 1, mg5.Len(), "ring 6 should have 1 skill group")

	found := false
	for _, raw := range mg5.List() {
		m := raw.(map[string]interface{})
		if m["member_group_id"] == id6 {
			found = true
			assert.Equal(t, sg, m["member_group_type"])
		}
	}
	assert.True(t, found, "ring 6 skill group id should be present")
}
