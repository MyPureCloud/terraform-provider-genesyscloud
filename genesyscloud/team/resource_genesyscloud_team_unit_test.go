package team

// build
import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"

	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
	"github.com/stretchr/testify/assert"
)

/** Unit Test **/
func TestUnitResourceTeamRead(t *testing.T) {
	tId := uuid.NewString()
	tName := "My Unit Test Team"
	tDescription := "My Unit Test Team"
	tDivisionId := uuid.NewString()

	teamProxyobj := &teamProxy{}

	teamProxyobj.getTeamByIdAttr = func(ctx context.Context, p *teamProxy, id string) (team *platformclientv2.Team, resp *platformclientv2.APIResponse, err error) {
		assert.Equal(t, tId, id)
		teamObj := &platformclientv2.Team{
			Name:        &tName,
			Description: &tDescription,
			Division:    &platformclientv2.Writabledivision{Id: &tDivisionId},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return teamObj, apiResponse, nil
	}
	teamProxyobj.getMembersByIdAttr = func(ctx context.Context, p *teamProxy, teamId string) (members *[]platformclientv2.Userreferencewithname, resp *platformclientv2.APIResponse, err error) {
		return nil, nil, nil
	}

	internalProxy = teamProxyobj
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gc := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceTeam().Schema

	resourceDataMap := buildTeamResourceMap(tId, tName, tDescription, tDivisionId)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readTeam(ctx, d, gc)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, tName, d.Get("name").(string))
	assert.Equal(t, tName, d.Get("description").(string))
	assert.Equal(t, tDivisionId, d.Get("division_id").(string))

}

func TestUnitResourceTeamDelete(t *testing.T) {

	tId := uuid.NewString()
	tName := "My Unit Test Team"
	tDescription := "My Unit Test Team"
	tDivisionId := uuid.NewString()

	teamProxyobj := &teamProxy{}

	teamProxyobj.deleteTeamAttr = func(ctx context.Context, p *teamProxy, id string) (resp *platformclientv2.APIResponse, err error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	teamProxyobj.getTeamByIdAttr = func(ctx context.Context, p *teamProxy, id string) (team *platformclientv2.Team, resp *platformclientv2.APIResponse, err error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}
		return nil, apiResponse, fmt.Errorf("Unable to find the team: %s", id)
	}

	internalProxy = teamProxyobj
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gc := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceTeam().Schema

	resourceDataMap := buildTeamResourceMap(tId, tName, tDescription, tDivisionId)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := deleteTeam(ctx, d, gc)
	assert.Nil(t, diag)
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceTeamCreate(t *testing.T) {

	tId := uuid.NewString()
	tName := "My Unit Test Team"
	tDescription := "My Unit Test Team"
	tDivisionId := uuid.NewString()

	teamProxyobj := &teamProxy{}

	teamProxyobj.getTeamByIdAttr = func(ctx context.Context, p *teamProxy, id string) (team *platformclientv2.Team, resp *platformclientv2.APIResponse, err error) {
		assert.Equal(t, tId, id)
		teamObj := &platformclientv2.Team{
			Name:        &tName,
			Description: &tDescription,
			Division:    &platformclientv2.Writabledivision{Id: &tDivisionId},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return teamObj, apiResponse, nil
	}

	teamProxyobj.createTeamAttr = func(ctx context.Context, p *teamProxy, team *platformclientv2.Team) (*platformclientv2.Team, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tName, *team.Name, "team.Name check failed in create team")
		assert.Equal(t, tDescription, *team.Description, "team.Description check failed in create team")
		assert.Equal(t, tDivisionId, *team.Division.Id, "team.Division.Id check failed in create team")

		team.Id = &tId

		return team, nil, nil
	}

	teamProxyobj.getMembersByIdAttr = func(ctx context.Context, p *teamProxy, teamId string) (members *[]platformclientv2.Userreferencewithname, resp *platformclientv2.APIResponse, err error) {
		return nil, nil, nil
	}

	internalProxy = teamProxyobj
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gc := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceTeam().Schema

	resourceDataMap := buildTeamResourceMap(tId, tName, tDescription, tDivisionId)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := createTeam(ctx, d, gc)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceTeamUpdate(t *testing.T) {

	tId := uuid.NewString()
	tName := "My Unit Test Team"
	tDescription := "My Updated Unit Test Team"
	tDivisionId := uuid.NewString()

	teamProxyobj := &teamProxy{}

	teamProxyobj.getTeamByIdAttr = func(ctx context.Context, p *teamProxy, id string) (team *platformclientv2.Team, resp *platformclientv2.APIResponse, err error) {
		assert.Equal(t, tId, id)
		teamObj := &platformclientv2.Team{
			Name:        &tName,
			Description: &tDescription,
			Division:    &platformclientv2.Writabledivision{Id: &tDivisionId},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return teamObj, apiResponse, nil
	}

	teamProxyobj.updateTeamAttr = func(ctx context.Context, p *teamProxy, id string, team *platformclientv2.Team) (*platformclientv2.Team, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tName, *team.Name, "team.Name check failed in create team")
		assert.Equal(t, tDescription, *team.Description, "team.Description check failed in create team")
		assert.Equal(t, tDivisionId, *team.Division.Id, "team.Division.Id check failed in create team")

		team.Id = &tId

		return team, nil, nil
	}

	teamProxyobj.getMembersByIdAttr = func(ctx context.Context, p *teamProxy, teamId string) (members *[]platformclientv2.Userreferencewithname, resp *platformclientv2.APIResponse, err error) {
		return nil, nil, nil
	}

	internalProxy = teamProxyobj
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gc := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceTeam().Schema

	resourceDataMap := buildTeamResourceMap(tId, tName, tDescription, tDivisionId)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := updateTeam(ctx, d, gc)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, tDescription, d.Get("description").(string))
}

func TestUnitResourceTeamGetAll(t *testing.T) {
	tIdFirst := uuid.NewString()
	tNameFirst := "My Unit Test Team"
	tDescriptionFirst := "My Unit Test Team"
	tDivisionIdFirst := uuid.NewString()

	tIdSecond := uuid.NewString()
	tNameSecond := "My Unit Test Team 2"
	tDescriptionSecond := "My Unit Test Team 2"
	tDivisionIdSecond := uuid.NewString()

	teamProxyobj := &teamProxy{}

	teamProxyobj.getAllTeamAttr = func(ctx context.Context, p *teamProxy, name string) (*[]platformclientv2.Team, *platformclientv2.APIResponse, error) {
		var allTeams []platformclientv2.Team

		teamObjFirst := &platformclientv2.Team{
			Id:          &tIdFirst,
			Name:        &tNameFirst,
			Description: &tDescriptionFirst,
			Division:    &platformclientv2.Writabledivision{Id: &tDivisionIdFirst},
		}

		teamObjSecond := &platformclientv2.Team{
			Id:          &tIdSecond,
			Name:        &tNameSecond,
			Description: &tDescriptionSecond,
			Division:    &platformclientv2.Writabledivision{Id: &tDivisionIdSecond},
		}

		allTeams = append(allTeams, *teamObjFirst)
		allTeams = append(allTeams, *teamObjSecond)

		return &allTeams, nil, nil
	}

	internalProxy = teamProxyobj
	defer func() { internalProxy = nil }()

	ctx := context.Background()

	exportedResource, diag := getAllAuthTeams(ctx, &platformclientv2.Configuration{})
	assert.Equal(t, true, len(exportedResource) > 0)
	assert.Equal(t, false, diag.HasError())

}

func buildTeamResourceMap(tId string, tName string, tDescription string, tDivisionId string) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"id":          tId,
		"name":        tName,
		"description": tDescription,
		"division_id": tDivisionId,
	}
	return resourceDataMap
}
