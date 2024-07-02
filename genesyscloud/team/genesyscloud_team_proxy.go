package team

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_team_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *teamProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createTeamFunc func(ctx context.Context, p *teamProxy, team *platformclientv2.Team) (*platformclientv2.Team, *platformclientv2.APIResponse, error)
type getAllTeamFunc func(ctx context.Context, p *teamProxy, name string) (*[]platformclientv2.Team, *platformclientv2.APIResponse, error)
type getTeamIdByNameFunc func(ctx context.Context, p *teamProxy, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error)
type getTeamByIdFunc func(ctx context.Context, p *teamProxy, id string) (team *platformclientv2.Team, response *platformclientv2.APIResponse, err error)
type updateTeamFunc func(ctx context.Context, p *teamProxy, id string, team *platformclientv2.Team) (*platformclientv2.Team, *platformclientv2.APIResponse, error)
type deleteTeamFunc func(ctx context.Context, p *teamProxy, id string) (response *platformclientv2.APIResponse, err error)
type createMembersFunc func(ctx context.Context, p *teamProxy, teamId string, members platformclientv2.Teammembers) (*platformclientv2.Teammemberaddlistingresponse, *platformclientv2.APIResponse, error)
type getMembersByIdFunc func(ctx context.Context, p *teamProxy, teamId string) (members *[]platformclientv2.Userreferencewithname, resp *platformclientv2.APIResponse, err error)
type deleteMembersFunc func(ctx context.Context, p *teamProxy, teamId string, memberId string) (response *platformclientv2.APIResponse, err error)

// teamProxy contains all of the methods that call genesys cloud APIs.
type teamProxy struct {
	clientConfig        *platformclientv2.Configuration
	teamsApi            *platformclientv2.TeamsApi
	createTeamAttr      createTeamFunc
	getAllTeamAttr      getAllTeamFunc
	getTeamIdByNameAttr getTeamIdByNameFunc
	getTeamByIdAttr     getTeamByIdFunc
	updateTeamAttr      updateTeamFunc
	deleteTeamAttr      deleteTeamFunc
	createMembersAttr   createMembersFunc
	getMembersByIdAttr  getMembersByIdFunc
	deleteMembersAttr   deleteMembersFunc
}

// newTeamProxy initializes the team proxy with all of the data needed to communicate with Genesys Cloud
func newTeamProxy(clientConfig *platformclientv2.Configuration) *teamProxy {
	api := platformclientv2.NewTeamsApiWithConfig(clientConfig)
	return &teamProxy{
		clientConfig:        clientConfig,
		teamsApi:            api,
		createTeamAttr:      createTeamFn,
		getAllTeamAttr:      getAllTeamFn,
		getTeamIdByNameAttr: getTeamIdByNameFn,
		getTeamByIdAttr:     getTeamByIdFn,
		updateTeamAttr:      updateTeamFn,
		deleteTeamAttr:      deleteTeamFn,
		createMembersAttr:   createMembersFn,
		getMembersByIdAttr:  getMembersByIdFn,
		deleteMembersAttr:   deleteMembersFn,
	}
}

// getTeamProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getTeamProxy(clientConfig *platformclientv2.Configuration) *teamProxy {
	if internalProxy == nil {
		internalProxy = newTeamProxy(clientConfig)
	}
	return internalProxy
}

// createTeam creates a Genesys Cloud team
func (p *teamProxy) createTeam(ctx context.Context, team *platformclientv2.Team) (*platformclientv2.Team, *platformclientv2.APIResponse, error) {
	return p.createTeamAttr(ctx, p, team)
}

// getTeam retrieves all Genesys Cloud team
func (p *teamProxy) getAllTeam(ctx context.Context, name string) (*[]platformclientv2.Team, *platformclientv2.APIResponse, error) {
	return p.getAllTeamAttr(ctx, p, name)
}

// getTeamIdByName returns a single Genesys Cloud team by a name
func (p *teamProxy) getTeamIdByName(ctx context.Context, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getTeamIdByNameAttr(ctx, p, name)
}

// getTeamById returns a single Genesys Cloud team by Id
func (p *teamProxy) getTeamById(ctx context.Context, id string) (team *platformclientv2.Team, resp *platformclientv2.APIResponse, err error) {
	return p.getTeamByIdAttr(ctx, p, id)
}

// updateTeam updates a Genesys Cloud team
func (p *teamProxy) updateTeam(ctx context.Context, id string, team *platformclientv2.Team) (*platformclientv2.Team, *platformclientv2.APIResponse, error) {
	return p.updateTeamAttr(ctx, p, id, team)
}

// deleteTeam deletes a Genesys Cloud team by Id
func (p *teamProxy) deleteTeam(ctx context.Context, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.deleteTeamAttr(ctx, p, id)
}

// createMembers creates a Genesys Cloud members
func (p *teamProxy) createMembers(ctx context.Context, teamId string, members platformclientv2.Teammembers) (*platformclientv2.Teammemberaddlistingresponse, *platformclientv2.APIResponse, error) {
	return p.createMembersAttr(ctx, p, teamId, members)
}

// getMembersById returns a single Genesys Cloud members by Id
func (p *teamProxy) getMembersById(ctx context.Context, teamId string) (members *[]platformclientv2.Userreferencewithname, resp *platformclientv2.APIResponse, err error) {
	return p.getMembersByIdAttr(ctx, p, teamId)
}

// deleteMembers deletes a Genesys Cloud members by Id
func (p *teamProxy) deleteMembers(ctx context.Context, teamId string, memberId string) (resp *platformclientv2.APIResponse, err error) {
	return p.deleteMembersAttr(ctx, p, teamId, memberId)
}

// createTeamFn is an implementation function for creating a Genesys Cloud team
func createTeamFn(ctx context.Context, p *teamProxy, team *platformclientv2.Team) (*platformclientv2.Team, *platformclientv2.APIResponse, error) {
	team, resp, err := p.teamsApi.PostTeams(*team)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to create team: %s", err)
	}
	return team, resp, nil
}

// getAllTeamFn is the implementation for retrieving all team in Genesys Cloud
func getAllTeamFn(ctx context.Context, p *teamProxy, name string) (*[]platformclientv2.Team, *platformclientv2.APIResponse, error) {
	var (
		after    string
		err      error
		allTeams []platformclientv2.Team
		response *platformclientv2.APIResponse
	)

	const pageSize = 100
	for i := 0; ; i++ {

		teams, resp, getErr := p.teamsApi.GetTeams(pageSize, name, after, "", "")
		response = resp
		if getErr != nil {
			return nil, resp, fmt.Errorf("Unable to find team entities %s", getErr)
		}

		if teams.Entities == nil || len(*teams.Entities) == 0 {
			break
		}

		allTeams = append(allTeams, *teams.Entities...)

		if teams.NextUri == nil || *teams.NextUri == "" {
			break
		}

		after, err = util.GetQueryParamValueFromUri(*teams.NextUri, "after")
		if err != nil {
			return nil, resp, fmt.Errorf("unable to parse after cursor from teams next uri: %v", err)
		}
		if after == "" {
			break
		}
	}
	return &allTeams, response, nil

}

// getTeamIdByNameFn is an implementation of the function to get a Genesys Cloud team by name
func getTeamIdByNameFn(ctx context.Context, p *teamProxy, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	teams, resp, err := getAllTeamFn(ctx, p, name)
	if err != nil {
		return "", false, resp, err
	}

	if len(*teams) == 0 {
		return "", true, resp, fmt.Errorf("No team found with name %s", name)
	}

	for _, team := range *teams {
		if *team.Name == name {
			log.Printf("Retrieved the team id %s by name %s", *team.Id, name)
			return *team.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("Unable to find team with name %s", name)
}

// getTeamByIdFn is an implementation of the function to get a Genesys Cloud team by Id
func getTeamByIdFn(ctx context.Context, p *teamProxy, id string) (team *platformclientv2.Team, resp *platformclientv2.APIResponse, err error) {
	team, resp, err = p.teamsApi.GetTeam(id)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to retrieve team by id %s: %s", id, err)
	}

	return team, resp, nil
}

// updateTeamFn is an implementation of the function to update a Genesys Cloud team
func updateTeamFn(ctx context.Context, p *teamProxy, id string, team *platformclientv2.Team) (*platformclientv2.Team, *platformclientv2.APIResponse, error) {
	team, resp, err := p.teamsApi.PatchTeam(id, *team)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to update team: %s", err)
	}
	return team, resp, nil
}

// deleteTeamFn is an implementation function for deleting a Genesys Cloud team
func deleteTeamFn(ctx context.Context, p *teamProxy, id string) (resp *platformclientv2.APIResponse, err error) {
	resp, err = p.teamsApi.DeleteTeam(id)
	if err != nil {
		return resp, fmt.Errorf("Failed to delete team: %s", err)
	}
	return resp, nil
}

// createMembersFn is an implementation function for creating a Genesys Cloud members
func createMembersFn(ctx context.Context, p *teamProxy, teamId string, members platformclientv2.Teammembers) (*platformclientv2.Teammemberaddlistingresponse, *platformclientv2.APIResponse, error) {
	teamListingResponse, resp, err := p.teamsApi.PostTeamMembers(teamId, members)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to create members: %s", err)
	}
	return teamListingResponse, resp, nil
}

// getMembersByIdFn is an implementation of the function to get a Genesys Cloud members by Id
func getMembersByIdFn(_ context.Context, p *teamProxy, teamId string) (*[]platformclientv2.Userreferencewithname, *platformclientv2.APIResponse, error) {
	var (
		after      string
		allMembers []platformclientv2.Userreferencewithname
		response   *platformclientv2.APIResponse
	)
	const pageSize = 100

	for {
		members, resp, getErr := p.teamsApi.GetTeamMembers(teamId, pageSize, "", after, "")
		response = resp
		if getErr != nil {
			return nil, resp, fmt.Errorf("unable to find member entities %s", getErr)
		}
		if members.Entities == nil || len(*members.Entities) == 0 {
			break
		}
		allMembers = append(allMembers, *members.Entities...)
		if members.NextUri == nil || *members.NextUri == "" {
			break
		}

		after, err := util.GetQueryParamValueFromUri(*members.NextUri, "after")
		if err != nil {
			return nil, resp, fmt.Errorf("unable to parse after cursor from members next uri: %v", err)
		}
		if after == "" {
			break
		}
	}
	return &allMembers, response, nil
}

// deleteMembersFn is an implementation function for deleting a Genesys Cloud members
func deleteMembersFn(ctx context.Context, p *teamProxy, teamId string, memberIds string) (resp *platformclientv2.APIResponse, err error) {
	resp, err = p.teamsApi.DeleteTeamMembers(teamId, memberIds)
	if err != nil {
		return resp, fmt.Errorf("Failed to delete members: %s", err)
	}
	return resp, nil
}
