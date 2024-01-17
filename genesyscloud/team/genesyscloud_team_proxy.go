package team

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

/*
The genesyscloud_team_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *teamProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createTeamFunc func(ctx context.Context, p *teamProxy, team *platformclientv2.Team) (*platformclientv2.Team, error)
type getAllTeamFunc func(ctx context.Context, p *teamProxy, name string) (*[]platformclientv2.Team, error)
type getTeamIdByNameFunc func(ctx context.Context, p *teamProxy, name string) (id string, retryable bool, err error)
type getTeamByIdFunc func(ctx context.Context, p *teamProxy, id string) (team *platformclientv2.Team, responseCode int, err error)
type updateTeamFunc func(ctx context.Context, p *teamProxy, id string, team *platformclientv2.Team) (*platformclientv2.Team, error)
type deleteTeamFunc func(ctx context.Context, p *teamProxy, id string) (responseCode int, err error)
type createMembersFunc func(ctx context.Context, p *teamProxy, teamId string, members platformclientv2.Teammembers) (*platformclientv2.Teammemberaddlistingresponse, error)
type getMembersByIdFunc func(ctx context.Context, p *teamProxy, teamId string) (members *[]platformclientv2.Userreferencewithname, err error)
type deleteMembersFunc func(ctx context.Context, p *teamProxy, teamId string, memberId string) (responseCode int, err error)

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
func (p *teamProxy) createTeam(ctx context.Context, team *platformclientv2.Team) (*platformclientv2.Team, error) {
	return p.createTeamAttr(ctx, p, team)
}

// getTeam retrieves all Genesys Cloud team
func (p *teamProxy) getAllTeam(ctx context.Context, name string) (*[]platformclientv2.Team, error) {
	return p.getAllTeamAttr(ctx, p, name)
}

// getTeamIdByName returns a single Genesys Cloud team by a name
func (p *teamProxy) getTeamIdByName(ctx context.Context, name string) (id string, retryable bool, err error) {
	return p.getTeamIdByNameAttr(ctx, p, name)
}

// getTeamById returns a single Genesys Cloud team by Id
func (p *teamProxy) getTeamById(ctx context.Context, id string) (team *platformclientv2.Team, statusCode int, err error) {
	return p.getTeamByIdAttr(ctx, p, id)
}

// updateTeam updates a Genesys Cloud team
func (p *teamProxy) updateTeam(ctx context.Context, id string, team *platformclientv2.Team) (*platformclientv2.Team, error) {
	return p.updateTeamAttr(ctx, p, id, team)
}

// deleteTeam deletes a Genesys Cloud team by Id
func (p *teamProxy) deleteTeam(ctx context.Context, id string) (statusCode int, err error) {
	return p.deleteTeamAttr(ctx, p, id)
}

// createMembers creates a Genesys Cloud members
func (p *teamProxy) createMembers(ctx context.Context, teamId string, members platformclientv2.Teammembers) (*platformclientv2.Teammemberaddlistingresponse, error) {
	return p.createMembersAttr(ctx, p, teamId, members)
}

// getMembersById returns a single Genesys Cloud members by Id
func (p *teamProxy) getMembersById(ctx context.Context, teamId string) (members *[]platformclientv2.Userreferencewithname, err error) {
	return p.getMembersByIdAttr(ctx, p, teamId)
}

// deleteMembers deletes a Genesys Cloud members by Id
func (p *teamProxy) deleteMembers(ctx context.Context, teamId string, memberId string) (statusCode int, err error) {
	return p.deleteMembersAttr(ctx, p, teamId, memberId)
}

// createTeamFn is an implementation function for creating a Genesys Cloud team
func createTeamFn(ctx context.Context, p *teamProxy, team *platformclientv2.Team) (*platformclientv2.Team, error) {
	team, _, err := p.teamsApi.PostTeams(*team)
	if err != nil {
		return nil, fmt.Errorf("Failed to create team: %s", err)
	}

	return team, nil
}

// getAllTeamFn is the implementation for retrieving all team in Genesys Cloud
func getAllTeamFn(ctx context.Context, p *teamProxy, name string) (*[]platformclientv2.Team, error) {
	var (
		after    string
		allTeams []platformclientv2.Team
	)

	const pageSize = 100
	for i := 0; ; i++ {

		teams, _, getErr := p.teamsApi.GetTeams(pageSize, name, after, "", "")
		if getErr != nil {
			return nil, fmt.Errorf("Unable to find team entities %s", getErr)
		}

		if teams.Entities == nil || len(*teams.Entities) == 0 {
			break
		}

		allTeams = append(allTeams, *teams.Entities...)

		if teams.NextUri == nil || *teams.NextUri == "" {
			break
		}

		u, err := url.Parse(*teams.NextUri)
		if err != nil {
			return nil, fmt.Errorf("Unable to find team entities %s", err)
		}

		m, _ := url.ParseQuery(u.RawQuery)
		if afterSlice, ok := m["after"]; ok && len(afterSlice) > 0 {
			after = afterSlice[0]
		}
		if after == "" {
			break
		}
	}

	return &allTeams, nil

}

// getTeamIdByNameFn is an implementation of the function to get a Genesys Cloud team by name
func getTeamIdByNameFn(ctx context.Context, p *teamProxy, name string) (id string, retryable bool, err error) {
	teams, err := getAllTeamFn(ctx, p, name)
	if err != nil {
		return "", false, err
	}

	if len(*teams) == 0 {
		return "", true, fmt.Errorf("No team found with name %s", name)
	}

	for _, team := range *teams {
		if *team.Name == name {
			log.Printf("Retrieved the team id %s by name %s", *team.Id, name)
			return *team.Id, false, nil
		}
	}

	return "", true, fmt.Errorf("Unable to find team with name %s", name)
}

// getTeamByIdFn is an implementation of the function to get a Genesys Cloud team by Id
func getTeamByIdFn(ctx context.Context, p *teamProxy, id string) (team *platformclientv2.Team, statusCode int, err error) {
	team, resp, err := p.teamsApi.GetTeam(id)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to retrieve team by id %s: %s", id, err)
	}

	return team, resp.StatusCode, nil
}

// updateTeamFn is an implementation of the function to update a Genesys Cloud team
func updateTeamFn(ctx context.Context, p *teamProxy, id string, team *platformclientv2.Team) (*platformclientv2.Team, error) {
	team, _, err := p.teamsApi.PatchTeam(id, *team)
	if err != nil {
		return nil, fmt.Errorf("Failed to update team: %s", err)
	}
	return team, nil
}

// deleteTeamFn is an implementation function for deleting a Genesys Cloud team
func deleteTeamFn(ctx context.Context, p *teamProxy, id string) (statusCode int, err error) {
	resp, err := p.teamsApi.DeleteTeam(id)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Failed to delete team: %s", err)
	}

	return resp.StatusCode, nil
}

// createMembersFn is an implementation function for creating a Genesys Cloud members
func createMembersFn(ctx context.Context, p *teamProxy, teamId string, members platformclientv2.Teammembers) (*platformclientv2.Teammemberaddlistingresponse, error) {

	teamListingResponse, _, err := p.teamsApi.PostTeamMembers(teamId, members)
	if err != nil {
		return nil, fmt.Errorf("Failed to create members: %s", err)
	}

	return teamListingResponse, nil
}

// getMembersByIdFn is an implementation of the function to get a Genesys Cloud members by Id
func getMembersByIdFn(ctx context.Context, p *teamProxy, teamId string) (members *[]platformclientv2.Userreferencewithname, err error) {

	var (
		after      string
		allMembers []platformclientv2.Userreferencewithname
	)
	const pageSize = 100

	for i := 0; ; i++ {
		members, _, getErr := p.teamsApi.GetTeamMembers(teamId, pageSize, "", after, "")
		if getErr != nil {
			return nil, fmt.Errorf("Unable to find member entities %s", getErr)
		}
		if members.Entities == nil || len(*members.Entities) == 0 {
			break
		}
		allMembers = append(allMembers, *members.Entities...)
		if members.NextUri == nil || *members.NextUri == "" {
			break
		}
		u, err := url.Parse(*members.NextUri)
		if err != nil {
			return nil, fmt.Errorf("Unable to find member entities %s", err)
		}
		m, _ := url.ParseQuery(u.RawQuery)
		if afterSlice, ok := m["after"]; ok && len(afterSlice) > 0 {
			after = afterSlice[0]
		}
		if after == "" {
			break
		}
	}
	return &allMembers, nil
}

// deleteMembersFn is an implementation function for deleting a Genesys Cloud members
func deleteMembersFn(ctx context.Context, p *teamProxy, teamId string, memberIds string) (statusCode int, err error) {
	resp, err := p.teamsApi.DeleteTeamMembers(teamId, memberIds)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Failed to delete members: %s", err)
	}

	return resp.StatusCode, nil
}
