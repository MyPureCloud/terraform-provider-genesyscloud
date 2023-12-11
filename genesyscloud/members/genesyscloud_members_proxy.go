package members

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v116/platformclientv2"
)

/*
The genesyscloud_members_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *membersProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createMembersFunc func(ctx context.Context, p *membersProxy, teamId string, members platformclientv2.Teammembers) (*platformclientv2.Teammemberaddlistingresponse, error)
type getMembersByIdFunc func(ctx context.Context, p *membersProxy, teamId string) (members *platformclientv2.Teammemberentitylisting, statusCode int, err error)
type deleteMembersFunc func(ctx context.Context, p *membersProxy, teamId string, memberId string) (responseCode int, err error)

// membersProxy contains all of the methods that call genesys cloud APIs.
type membersProxy struct {
	clientConfig       *platformclientv2.Configuration
	teamsApi           *platformclientv2.TeamsApi
	createMembersAttr  createMembersFunc
	getMembersByIdAttr getMembersByIdFunc
	deleteMembersAttr  deleteMembersFunc
}

// newMembersProxy initializes the members proxy with all of the data needed to communicate with Genesys Cloud
func newMembersProxy(clientConfig *platformclientv2.Configuration) *membersProxy {
	api := platformclientv2.NewTeamsApiWithConfig(clientConfig)
	return &membersProxy{
		clientConfig:       clientConfig,
		teamsApi:           api,
		createMembersAttr:  createMembersFn,
		getMembersByIdAttr: getMembersByIdFn,
		deleteMembersAttr:  deleteMembersFn,
	}
}

// getMembersProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getMembersProxy(clientConfig *platformclientv2.Configuration) *membersProxy {
	if internalProxy == nil {
		internalProxy = newMembersProxy(clientConfig)
	}

	return internalProxy
}

// createMembers creates a Genesys Cloud members
func (p *membersProxy) createMembers(ctx context.Context, teamId string, members platformclientv2.Teammembers) (*platformclientv2.Teammemberaddlistingresponse, error) {
	return p.createMembersAttr(ctx, p, teamId, members)
}

// getMembersById returns a single Genesys Cloud members by Id
func (p *membersProxy) getMembersById(ctx context.Context, teamId string) (members *platformclientv2.Teammemberentitylisting, statusCode int, err error) {
	return p.getMembersByIdAttr(ctx, p, teamId)
}

// deleteMembers deletes a Genesys Cloud members by Id
func (p *membersProxy) deleteMembers(ctx context.Context, teamId string, memberId string) (statusCode int, err error) {
	return p.deleteMembersAttr(ctx, p, teamId, memberId)
}

// createMembersFn is an implementation function for creating a Genesys Cloud members
func createMembersFn(ctx context.Context, p *membersProxy, teamId string, members platformclientv2.Teammembers) (*platformclientv2.Teammemberaddlistingresponse, error) {

	teamListingResponse, _, err := p.teamsApi.PostTeamMembers(teamId, members)
	if err != nil {
		return nil, fmt.Errorf("Failed to create members: %s", err)
	}

	return teamListingResponse, nil
}

// getMembersByIdFn is an implementation of the function to get a Genesys Cloud members by Id
func getMembersByIdFn(ctx context.Context, p *membersProxy, teamId string) (members *platformclientv2.Teammemberentitylisting, statusCode int, err error) {
	teamMemberEntityListing, resp, err := p.teamsApi.GetTeamMembers(teamId, 100, "", "", "")
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to retrieve members by id %s: %s", teamId, err)
	}

	return teamMemberEntityListing, resp.StatusCode, nil
}

// deleteMembersFn is an implementation function for deleting a Genesys Cloud members
func deleteMembersFn(ctx context.Context, p *membersProxy, teamId string, memberIds string) (statusCode int, err error) {
	resp, err := p.teamsApi.DeleteTeamMembers(teamId, memberIds)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Failed to delete members: %s", err)
	}

	return resp.StatusCode, nil
}
