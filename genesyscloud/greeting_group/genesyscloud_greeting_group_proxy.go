package greeting_group

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)

var internalProxy *greetingProxy

type getAllGreetingsFunc func(ctx context.Context, p *greetingProxy) (*[]platformclientv2.Greeting, *platformclientv2.APIResponse, error)
type getGroupGreetingByIdFunc func(ctx context.Context, p *greetingProxy, groupId string, id string) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error)
type updateGroupGreetingFunc func(ctx context.Context, p *greetingProxy, greetingId string, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error)
type createGroupGreetingFunc func(ctx context.Context, p *greetingProxy, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error)
type deleteGroupGreetingFunc func(ctx context.Context, p *greetingProxy, id string) (*platformclientv2.APIResponse, error)

type greetingProxy struct {
	clientConfig        *platformclientv2.Configuration
	greetingsApi        *platformclientv2.GreetingsApi
	groupsApi           *platformclientv2.GroupsApi
	getAllGreetingsAttr getAllGreetingsFunc
	createGreetingAttr  createGroupGreetingFunc
	getGreetingByIdAttr getGroupGreetingByIdFunc
	updateGreetingAttr  updateGroupGreetingFunc
	deleteGreetingAttr  deleteGroupGreetingFunc
}

func newGreetingProxy(clientConfig *platformclientv2.Configuration) *greetingProxy {
	api := platformclientv2.NewGreetingsApiWithConfig(clientConfig)
	groupsApi := platformclientv2.NewGroupsApiWithConfig(clientConfig)
	return &greetingProxy{
		clientConfig:        clientConfig,
		greetingsApi:        api,
		groupsApi:           groupsApi,
		getAllGreetingsAttr: getAllGreetingsFn,
		createGreetingAttr:  createGroupGreetingFn,
		getGreetingByIdAttr: getGroupGreetingByIdFn,
		updateGreetingAttr:  updateGroupGreetingFn,
		deleteGreetingAttr:  deleteGroupGreetingFn,
	}
}

func getGreeetingProxy(clientConfig *platformclientv2.Configuration) *greetingProxy {
	if internalProxy == nil {
		internalProxy = newGreetingProxy(clientConfig)
	}

	return internalProxy
}

func (p *greetingProxy) getAllGreetings(ctx context.Context) (*[]platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.getAllGreetingsAttr(ctx, p)
}
func (p *greetingProxy) createGroupGreeting(ctx context.Context, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.createGreetingAttr(ctx, p, body)
}
func (p *greetingProxy) getGroupGreetingById(ctx context.Context, groupId string, id string) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.getGreetingByIdAttr(ctx, p, groupId, id)
}
func (p *greetingProxy) updateGroupGreeting(ctx context.Context, greetingID string, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.updateGreetingAttr(ctx, p, greetingID, body)
}
func (p *greetingProxy) deleteGroupGreeting(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteGreetingAttr(ctx, p, id)
}
func getAllGreetingsFn(ctx context.Context, p *greetingProxy) (*[]platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	var allGreetings []platformclientv2.Greeting
	const pageSize = 100
	allGroups, resp, err := getAllGroupsFn(ctx, p)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get groups %s", err)
	}

	for _, group := range *allGroups {
		groupGreetings, resp, err := p.greetingsApi.GetGroupGreetings(*group.Id, pageSize, 1)
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get greetings for group %s: %s", *group.Id, err)
		}
		if groupGreetings.Entities != nil {
			allGreetings = append(allGreetings, *groupGreetings.Entities...)
		}
		for pageNum := 2; pageNum <= *groupGreetings.PageCount; pageNum++ {
			groupGreetings, resp, err := p.greetingsApi.GetGroupGreetings(*group.Id, pageSize, pageNum)
			if err != nil {
				return nil, resp, fmt.Errorf("failed to get greetings for group %s: %s", *group.Id, err)
			}
			if groupGreetings.Entities != nil {
				allGreetings = append(allGreetings, *groupGreetings.Entities...)
			}
		}
	}
	return &allGreetings, resp, nil
}

func createGroupGreetingFn(ctx context.Context, p *greetingProxy, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.greetingsApi.PostGroupGreetings(*body.Owner.Id, *body)
}
func getGroupGreetingByIdFn(ctx context.Context, p *greetingProxy, groupId string, id string) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	if groupId == "" {
		return p.greetingsApi.GetGreeting(id)
	}
	return getGreetingFromGroup(ctx, p, groupId, id)
}
func getGreetingFromGroup(ctx context.Context, p *greetingProxy, groupId string, id string) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	groupGreetings, resp, err := p.greetingsApi.GetGroupGreetings(groupId, pageSize, 1)
	if err != nil {
		return nil, resp, err
	}
	if greeting := findGreetingInGroupEntities(groupGreetings.Entities, id); greeting != nil {
		return greeting, resp, nil
	}

	for pageNum := 2; pageNum <= *groupGreetings.PageCount; pageNum++ {
		groupGreetings, resp, err = p.greetingsApi.GetGroupGreetings(groupId, pageSize, pageNum)
		if err != nil {
			return nil, resp, err
		}
		if greeting := findGreetingInGroupEntities(groupGreetings.Entities, id); greeting != nil {
			return greeting, resp, nil
		}
	}

	return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, fmt.Errorf("greeting %s not found for group %s", id, groupId)
}

func findGreetingInGroupEntities(entities *[]platformclientv2.Greeting, greetingId string) *platformclientv2.Greeting {
	if entities == nil {
		return nil
	}
	for _, entity := range *entities {
		if entity.Id != nil && *entity.Id == greetingId {
			return &entity
		}
	}
	return nil
}
func updateGroupGreetingFn(ctx context.Context, p *greetingProxy, greetingId string, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.greetingsApi.PutGreeting(greetingId, *body)
}
func deleteGroupGreetingFn(ctx context.Context, p *greetingProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.greetingsApi.DeleteGreeting(id)
}
func getAllGroupsFn(ctx context.Context, p *greetingProxy) (*[]platformclientv2.Group, *platformclientv2.APIResponse, error) {
	var allGroups []platformclientv2.Group
	const pageSize = 100

	groups, resp, err := p.groupsApi.GetGroups(pageSize, 1, nil, nil, "")
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get groups %s", err)
	}

	if groups.Entities == nil || len(*groups.Entities) == 0 {
		return &allGroups, resp, nil
	}
	allGroups = append(allGroups, *groups.Entities...)

	for pageNum := 2; pageNum <= *groups.PageCount; pageNum++ {
		groups, resp, err := p.groupsApi.GetGroups(pageSize, pageNum, nil, nil, "")
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get groups %s", err)
		}

		if groups.Entities == nil || len(*groups.Entities) == 0 {
			return &allGroups, resp, nil
		}

		allGroups = append(allGroups, *groups.Entities...)
	}
	return &allGroups, resp, nil
}
