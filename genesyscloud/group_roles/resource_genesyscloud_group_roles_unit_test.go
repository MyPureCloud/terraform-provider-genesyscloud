package group_roles

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v204/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rolesSet(roles ...map[string]interface{}) *schema.Set {
	items := make([]interface{}, 0, len(roles))
	for _, r := range roles {
		items = append(items, r)
	}
	return schema.NewSet(schema.HashResource(RoleAssignmentResource), items)
}

// TestUnitGetExistingAndConfigGrantsUsesConfigHomeDivision verifies that a role with no
// division_ids is granted in the home division of the org the passed config points at.
func TestUnitGetExistingAndConfigGrantsUsesConfigHomeDivision(t *testing.T) {
	target := &platformclientv2.Configuration{BasePath: "https://api.mypurecloud.ie", AccessToken: "target-token"}

	var got *platformclientv2.Configuration
	restore := util.SetHomeDivisionFetcherForTesting(func(c *platformclientv2.Configuration) (string, error) {
		got = c
		return "target-home", nil
	})
	defer restore()

	config := rolesSet(map[string]interface{}{
		"role_id":      "role-1",
		"division_ids": schema.NewSet(schema.HashString, []interface{}{}),
	})

	existing, configGrants, err := getExistingAndConfigGrants(target, nil, config)
	require.NoError(t, err)
	assert.Empty(t, existing)
	assert.Equal(t, []string{"role-1:target-home"}, configGrants)
	assert.Same(t, target, got, "lookup must use the per-call config, not the SDK default")
}

// TestUnitGetExistingAndConfigGrantsHomeDivisionError verifies the lookup error is returned
// rather than an empty config grant list, which the update path would read as "remove all grants".
func TestUnitGetExistingAndConfigGrantsHomeDivisionError(t *testing.T) {
	restore := util.SetHomeDivisionFetcherForTesting(func(c *platformclientv2.Configuration) (string, error) {
		return "", errors.New("401 - No authentication bearer token specified")
	})
	defer restore()

	existingGrant := platformclientv2.Authzgrant{
		Role:     &platformclientv2.Authzgrantrole{Id: platformclientv2.String("role-1")},
		Division: &platformclientv2.Authzdivision{Id: platformclientv2.String("div-1")},
	}
	config := rolesSet(map[string]interface{}{
		"role_id":      "role-1",
		"division_ids": schema.NewSet(schema.HashString, []interface{}{"div-1"}),
	})

	_, configGrants, err := getExistingAndConfigGrants(&platformclientv2.Configuration{ClientID: "c"}, []platformclientv2.Authzgrant{existingGrant}, config)
	require.Error(t, err)
	assert.Nil(t, configGrants)
}
