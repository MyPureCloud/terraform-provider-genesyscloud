package routing_skill

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/stretchr/testify/assert"
)

// domainSkillUsersJson is the response fixture from the public API integration tests for
// GET /api/v2/routing/skills/{skillId}/users (public-api-v2 PR #18535, domain-skill-users.json)
const domainSkillUsersJson = `{
  "entities": [
    {"id":"55bc6be3-078c-4a49-a4e6-1e05776ed7e8","selfUri":"/api/v2/users/55bc6be3-078c-4a49-a4e6-1e05776ed7e8"},
    {"id":"7f163c3d-7f7a-43fd-8286-dae9792e8c2b","selfUri":"/api/v2/users/7f163c3d-7f7a-43fd-8286-dae9792e8c2b"},
    {"id":"7f163c3d-7f7a-43fd-8286-dae9792e8c2d","selfUri":"/api/v2/users/7f163c3d-7f7a-43fd-8286-dae9792e8c2d"}
  ]
}`

func buildSkillUserReferences(ids ...string) *[]platformclientv2.Userreference {
	users := make([]platformclientv2.Userreference, 0, len(ids))
	for _, id := range ids {
		users = append(users, platformclientv2.Userreference{
			Id:      platformclientv2.String(id),
			SelfUri: platformclientv2.String("/api/v2/users/" + id),
		})
	}
	return &users
}

func getSkillUserIdsFromState(d *schema.ResourceData) []string {
	ids := make([]string, 0)
	for _, id := range d.Get("user_ids").(*schema.Set).List() {
		ids = append(ids, id.(string))
	}
	sort.Strings(ids)
	return ids
}

// setTestRoutingSkillProxy injects proxy and clears the name lookup cache. The returned func restores both.
func setTestRoutingSkillProxy(proxy *routingSkillProxy) func() {
	internalProxy = proxy
	dataSourceRoutingSkillCache = nil
	return func() {
		internalProxy = nil
		dataSourceRoutingSkillCache = nil
	}
}

func readRoutingSkillUsersDataSource(t *testing.T, config map[string]interface{}) (*schema.ResourceData, error) {
	d := schema.TestResourceDataRaw(t, DataSourceRoutingSkillUsers().Schema, config)
	meta := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	diags := dataSourceRoutingSkillUsersRead(context.Background(), d, meta)
	if diags.HasError() {
		return d, fmt.Errorf("%s: %v", diags[0].Summary, diags)
	}
	return d, nil
}

func TestUnitDataSourceRoutingSkillUsersSchema(t *testing.T) {
	assert.NoError(t, DataSourceRoutingSkillUsers().InternalValidate(nil, false))
}

func TestUnitRoutingSkillUserListingUnmarshal(t *testing.T) {
	var listing routingSkillUserListing
	assert.NoError(t, json.Unmarshal([]byte(domainSkillUsersJson), &listing))
	assert.NotNil(t, listing.Entities)
	assert.Equal(t, 3, len(*listing.Entities))
	assert.Equal(t, "55bc6be3-078c-4a49-a4e6-1e05776ed7e8", *(*listing.Entities)[0].Id)
	assert.Equal(t, "/api/v2/users/7f163c3d-7f7a-43fd-8286-dae9792e8c2d", *(*listing.Entities)[2].SelfUri)
}

func TestUnitDataSourceRoutingSkillUsersReadById(t *testing.T) {
	skillId := uuid.NewString()
	userIds := []string{uuid.NewString(), uuid.NewString(), uuid.NewString()}
	sort.Strings(userIds)

	proxy := &routingSkillProxy{}
	proxy.getRoutingSkillUsersAttr = func(ctx context.Context, p *routingSkillProxy, id string) (*[]platformclientv2.Userreference, *platformclientv2.APIResponse, error) {
		assert.Equal(t, skillId, id)
		return buildSkillUserReferences(userIds...), &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.getAllRoutingSkillsAttr = func(ctx context.Context, p *routingSkillProxy, name string) (*[]platformclientv2.Routingskill, *platformclientv2.APIResponse, error) {
		t.Fatal("skills should not be listed when skill_id is set")
		return nil, nil, nil
	}
	defer setTestRoutingSkillProxy(proxy)()

	d, err := readRoutingSkillUsersDataSource(t, map[string]interface{}{"skill_id": skillId})
	assert.NoError(t, err)
	assert.Equal(t, skillId, d.Id())
	assert.Equal(t, skillId, d.Get("skill_id").(string))
	assert.Equal(t, userIds, getSkillUserIdsFromState(d))
}

func TestUnitDataSourceRoutingSkillUsersReadByName(t *testing.T) {
	skillId := uuid.NewString()
	skillName := "Terraform Skill " + uuid.NewString()
	userId := uuid.NewString()

	proxy := &routingSkillProxy{}
	proxy.getAllRoutingSkillsAttr = func(ctx context.Context, p *routingSkillProxy, name string) (*[]platformclientv2.Routingskill, *platformclientv2.APIResponse, error) {
		skills := []platformclientv2.Routingskill{
			{Id: platformclientv2.String(uuid.NewString()), Name: platformclientv2.String("Other Skill")},
			{Id: platformclientv2.String(skillId), Name: platformclientv2.String(skillName)},
		}
		return &skills, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.getRoutingSkillUsersAttr = func(ctx context.Context, p *routingSkillProxy, id string) (*[]platformclientv2.Userreference, *platformclientv2.APIResponse, error) {
		assert.Equal(t, skillId, id)
		return buildSkillUserReferences(userId), &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	defer setTestRoutingSkillProxy(proxy)()

	d, err := readRoutingSkillUsersDataSource(t, map[string]interface{}{"skill_name": skillName})
	assert.NoError(t, err)
	assert.Equal(t, skillId, d.Id())
	assert.Equal(t, skillId, d.Get("skill_id").(string))
	assert.Equal(t, skillName, d.Get("skill_name").(string))
	assert.Equal(t, []string{userId}, getSkillUserIdsFromState(d))
}

func TestUnitDataSourceRoutingSkillUsersReadEmptySkill(t *testing.T) {
	skillId := uuid.NewString()

	proxy := &routingSkillProxy{}
	proxy.getRoutingSkillUsersAttr = func(ctx context.Context, p *routingSkillProxy, id string) (*[]platformclientv2.Userreference, *platformclientv2.APIResponse, error) {
		return buildSkillUserReferences(), &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	defer setTestRoutingSkillProxy(proxy)()

	d, err := readRoutingSkillUsersDataSource(t, map[string]interface{}{"skill_id": skillId})
	assert.NoError(t, err)
	assert.Equal(t, skillId, d.Id())
	assert.Equal(t, 0, d.Get("user_ids").(*schema.Set).Len())
}

func TestUnitDataSourceRoutingSkillUsersReadSkillNotFound(t *testing.T) {
	skillId := uuid.NewString()

	proxy := &routingSkillProxy{}
	proxy.getRoutingSkillUsersAttr = func(ctx context.Context, p *routingSkillProxy, id string) (*[]platformclientv2.Userreference, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, fmt.Errorf("The skill ID could not be found.")
	}
	defer setTestRoutingSkillProxy(proxy)()

	d, err := readRoutingSkillUsersDataSource(t, map[string]interface{}{"skill_id": skillId})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("routing skill %s not found", skillId))
	assert.Equal(t, "", d.Id())
}

func TestUnitDataSourceRoutingSkillUsersReadApiError(t *testing.T) {
	skillId := uuid.NewString()

	proxy := &routingSkillProxy{}
	proxy.getRoutingSkillUsersAttr = func(ctx context.Context, p *routingSkillProxy, id string) (*[]platformclientv2.Userreference, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusForbidden}, fmt.Errorf("missing permission")
	}
	defer setTestRoutingSkillProxy(proxy)()

	d, err := readRoutingSkillUsersDataSource(t, map[string]interface{}{"skill_id": skillId})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), skillId)
	assert.Contains(t, err.Error(), "missing permission")
	assert.Equal(t, "", d.Id())
}

func TestUnitDataSourceRoutingSkillUsersReadUnknownSkillName(t *testing.T) {
	skillName := "Unknown Skill " + uuid.NewString()

	proxy := &routingSkillProxy{}
	proxy.getAllRoutingSkillsAttr = func(ctx context.Context, p *routingSkillProxy, name string) (*[]platformclientv2.Routingskill, *platformclientv2.APIResponse, error) {
		skills := []platformclientv2.Routingskill{
			{Id: platformclientv2.String(uuid.NewString()), Name: platformclientv2.String("Other Skill")},
		}
		return &skills, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.getRoutingSkillIdByNameAttr = func(ctx context.Context, p *routingSkillProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
		// Non-retryable so the test does not wait out the retry timeout
		return "", &platformclientv2.APIResponse{StatusCode: http.StatusOK}, false, fmt.Errorf("no routing skills found with name '%s'", name)
	}
	proxy.getRoutingSkillUsersAttr = func(ctx context.Context, p *routingSkillProxy, id string) (*[]platformclientv2.Userreference, *platformclientv2.APIResponse, error) {
		t.Fatal("users should not be requested when the skill name cannot be resolved")
		return nil, nil, nil
	}
	defer setTestRoutingSkillProxy(proxy)()

	d, err := readRoutingSkillUsersDataSource(t, map[string]interface{}{"skill_name": skillName})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), skillName)
	assert.Equal(t, "", d.Id())
}
