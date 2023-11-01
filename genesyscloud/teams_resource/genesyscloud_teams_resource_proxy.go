package teams_resource

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
	"log"
)

/*
The genesyscloud_teams_resource_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *teamsResourceProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createTeamsResourceFunc func(ctx context.Context, p *teamsResourceProxy, team *platformclientv2.Team) (*platformclientv2.Team, error)
type getAllTeamsResourceFunc func(ctx context.Context, p *teamsResourceProxy) (*[]platformclientv2.Team, error)
type getTeamsResourceIdByNameFunc func(ctx context.Context, p *teamsResourceProxy, name string) (id string, retryable bool, err error)
type getTeamsResourceByIdFunc func(ctx context.Context, p *teamsResourceProxy, id string) (team *platformclientv2.Team, responseCode int, err error)
type updateTeamsResourceFunc func(ctx context.Context, p *teamsResourceProxy, id string, team *platformclientv2.Team) (*platformclientv2.Team, error)
type deleteTeamsResourceFunc func(ctx context.Context, p *teamsResourceProxy, id string) (responseCode int, err error)

// teamsResourceProxy contains all of the methods that call genesys cloud APIs.
type teamsResourceProxy struct {
	clientConfig                 *platformclientv2.Configuration
	teamsApi                     *platformclientv2.TeamsApi
	createTeamsResourceAttr      createTeamsResourceFunc
	getAllTeamsResourceAttr      getAllTeamsResourceFunc
	getTeamsResourceIdByNameAttr getTeamsResourceIdByNameFunc
	getTeamsResourceByIdAttr     getTeamsResourceByIdFunc
	updateTeamsResourceAttr      updateTeamsResourceFunc
	deleteTeamsResourceAttr      deleteTeamsResourceFunc
}

// newTeamsResourceProxy initializes the teams resource proxy with all of the data needed to communicate with Genesys Cloud
func newTeamsResourceProxy(clientConfig *platformclientv2.Configuration) *teamsResourceProxy {
	api := platformclientv2.NewTeamsApiWithConfig(clientConfig)
	return &teamsResourceProxy{
		clientConfig:                 clientConfig,
		teamsApi:                     api,
		createTeamsResourceAttr:      createTeamsResourceFn,
		getAllTeamsResourceAttr:      getAllTeamsResourceFn,
		getTeamsResourceIdByNameAttr: getTeamsResourceIdByNameFn,
		getTeamsResourceByIdAttr:     getTeamsResourceByIdFn,
		updateTeamsResourceAttr:      updateTeamsResourceFn,
		deleteTeamsResourceAttr:      deleteTeamsResourceFn,
	}
}

// getTeamsResourceProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getTeamsResourceProxy(clientConfig *platformclientv2.Configuration) *teamsResourceProxy {
	if internalProxy == nil {
		internalProxy = newTeamsResourceProxy(clientConfig)
	}

	return internalProxy
}

// createTeamsResource creates a Genesys Cloud teams resource
func (p *teamsResourceProxy) createTeamsResource(ctx context.Context, teamsResource *platformclientv2.Team) (*platformclientv2.Team, error) {
	return p.createTeamsResourceAttr(ctx, p, teamsResource)
}

// getTeamsResource retrieves all Genesys Cloud teams resource
func (p *teamsResourceProxy) getAllTeamsResource(ctx context.Context) (*[]platformclientv2.Team, error) {
	return p.getAllTeamsResourceAttr(ctx, p)
}

// getTeamsResourceIdByName returns a single Genesys Cloud teams resource by a name
func (p *teamsResourceProxy) getTeamsResourceIdByName(ctx context.Context, name string) (id string, retryable bool, err error) {
	return p.getTeamsResourceIdByNameAttr(ctx, p, name)
}

// getTeamsResourceById returns a single Genesys Cloud teams resource by Id
func (p *teamsResourceProxy) getTeamsResourceById(ctx context.Context, id string) (teamsResource *platformclientv2.Team, statusCode int, err error) {
	return p.getTeamsResourceByIdAttr(ctx, p, id)
}

// updateTeamsResource updates a Genesys Cloud teams resource
func (p *teamsResourceProxy) updateTeamsResource(ctx context.Context, id string, teamsResource *platformclientv2.Team) (*platformclientv2.Team, error) {
	return p.updateTeamsResourceAttr(ctx, p, id, teamsResource)
}

// deleteTeamsResource deletes a Genesys Cloud teams resource by Id
func (p *teamsResourceProxy) deleteTeamsResource(ctx context.Context, id string) (statusCode int, err error) {
	return p.deleteTeamsResourceAttr(ctx, p, id)
}

// createTeamsResourceFn is an implementation function for creating a Genesys Cloud teams resource
func createTeamsResourceFn(ctx context.Context, p *teamsResourceProxy, teamsResource *platformclientv2.Team) (*platformclientv2.Team, error) {
	team, _, err := p.teamsApi.PostTeams(*teamsResource)
	if err != nil {
		return nil, fmt.Errorf("Failed to create teams resource: %s", err)
	}

	return team, nil
}

// getAllTeamsResourceFn is the implementation for retrieving all teams resource in Genesys Cloud
func getAllTeamsResourceFn(ctx context.Context, p *teamsResourceProxy) (*[]platformclientv2.Team, error) {
	var allTeams []platformclientv2.Team

	teams, _, err := p.teamsApi.GetTeams()
	if err != nil {
		return nil, fmt.Errorf("Failed to get team: %v", err)
	}
	for _, team := range *teams.Entities {
		allTeams = append(allTeams, team)
	}

	for pageNum := 2; pageNum <= *teams.PageCount; pageNum++ {
		const pageSize = 100

		teams, _, err := p.teamsApi.GetTeams()
		if err != nil {
			return nil, fmt.Errorf("Failed to get team: %v", err)
		}

		if teams.Entities == nil || len(*teams.Entities) == 0 {
			break
		}

		for _, team := range *teams.Entities {
			allTeams = append(allTeams, team)
		}
	}

	return &allTeams, nil
}

// getTeamsResourceIdByNameFn is an implementation of the function to get a Genesys Cloud teams resource by name
func getTeamsResourceIdByNameFn(ctx context.Context, p *teamsResourceProxy, name string) (id string, retryable bool, err error) {
	const pageNum = 1
	const pageSize = 100
	teams, _, err := p.teamsApi.GetTeams()
	if err != nil {
		return "", false, fmt.Errorf("Error searching teams resource %s: %s", name, err)
	}

	if teams.Entities == nil || len(*teams.Entities) == 0 {
		return "", true, fmt.Errorf("No teams resource found with name %s", name)
	}

	var team platformclientv2.Team
	for _, teamSdk := range *teams.Entities {
		if *team.Name == name {
			log.Printf("Retrieved the teams resource id %s by name %s", *teamSdk.Id, name)
			team = teamSdk
			return *team.Id, false, nil
		}
	}

	return "", false, fmt.Errorf("Unable to find teams resource with name %s", name)
}

// getTeamsResourceByIdFn is an implementation of the function to get a Genesys Cloud teams resource by Id
func getTeamsResourceByIdFn(ctx context.Context, p *teamsResourceProxy, id string) (teamsResource *platformclientv2.Team, statusCode int, err error) {
	team, resp, err := p.teamsApi.GetTeam(id)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to retrieve teams resource by id %s: %s", id, err)
	}

	return team, resp.StatusCode, nil
}

// updateTeamsResourceFn is an implementation of the function to update a Genesys Cloud teams resource
func updateTeamsResourceFn(ctx context.Context, p *teamsResourceProxy, id string, teamsResource *platformclientv2.Team) (*platformclientv2.Team, error) {
	team, _, err := getTeamsResourceByIdFn(ctx, p, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get teams resource by id %s: %s", id, err)
	}

	team.Version = teamsResource.Version
	team, _, err = p.teamsApi.PatchTeam(id, *team)
	if err != nil {
		return nil, fmt.Errorf("Failed to update teams resource: %s", err)
	}
	return team, nil
}

// deleteTeamsResourceFn is an implementation function for deleting a Genesys Cloud teams resource
func deleteTeamsResourceFn(ctx context.Context, p *teamsResourceProxy, id string) (statusCode int, err error) {
	resp, err := p.teamsApi.DeleteTeam(id)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Failed to delete teams resource: %s", err)
	}

	return resp.StatusCode, nil
}
