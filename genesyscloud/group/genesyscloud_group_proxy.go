package group

import (
	"context"
	"fmt"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *groupProxy

type createGroupFunc func(ctx context.Context, p *groupProxy, group *platformclientv2.Groupcreate) (*platformclientv2.Group, *platformclientv2.APIResponse, error)
type getAllGroupFunc func(ctx context.Context, p *groupProxy) (*[]platformclientv2.Group, *platformclientv2.APIResponse, error)
type updateGroupFunc func(ctx context.Context, p *groupProxy, id string, group *platformclientv2.Groupupdate) (*platformclientv2.Group, *platformclientv2.APIResponse, error)
type getGroupByIdFunc func(ctx context.Context, p *groupProxy, id string) (*platformclientv2.Group, *platformclientv2.APIResponse, error)
type addGroupMembersFunc func(ctx context.Context, p *groupProxy, id string, members *platformclientv2.Groupmembersupdate) (*interface{}, *platformclientv2.APIResponse, error)
type deleteGroupMembersFunc func(ctx context.Context, p *groupProxy, id string, members string) (*interface{}, *platformclientv2.APIResponse, error)
type getGroupMembersFunc func(ctx context.Context, p *groupProxy, id string) (*[]string, *platformclientv2.APIResponse, error)
type getGroupByNameFunc func(ctx context.Context, p *groupProxy, name string) (*platformclientv2.Groupssearchresponse, *platformclientv2.APIResponse, error)
type deleteGroupFunc func(ctx context.Context, p *groupProxy, id string) (*platformclientv2.APIResponse, error)

type groupProxy struct {
	clientConfig           *platformclientv2.Configuration
	groupsApi              *platformclientv2.GroupsApi
	createGroupAttr        createGroupFunc
	getAllGroupAttr        getAllGroupFunc
	updateGroupAttr        updateGroupFunc
	deleteGroupAttr        deleteGroupFunc
	getGroupByNameAttr     getGroupByNameFunc
	getGroupByIdAttr       getGroupByIdFunc
	addGroupMembersAttr    addGroupMembersFunc
	deleteGroupMembersAttr deleteGroupMembersFunc
	getGroupMembersAttr    getGroupMembersFunc
	groupCache             rc.CacheInterface[platformclientv2.Group]
}

func newGroupProxy(clientConfig *platformclientv2.Configuration) *groupProxy {
	api := platformclientv2.NewGroupsApiWithConfig(clientConfig)
	groupCache := rc.NewResourceCache[platformclientv2.Group]()
	return &groupProxy{
		clientConfig:           clientConfig,
		groupsApi:              api,
		createGroupAttr:        createGroupFn,
		getAllGroupAttr:        getAllGroupFn,
		updateGroupAttr:        updateGroupFn,
		deleteGroupAttr:        deleteGroupFn,
		getGroupByNameAttr:     getGroupByNameFn,
		getGroupByIdAttr:       getGroupByIdFn,
		addGroupMembersAttr:    addGroupMembersFn,
		deleteGroupMembersAttr: deleteGroupMembersFn,
		getGroupMembersAttr:    getGroupMembersFn,
		groupCache:             groupCache,
	}
}

func getGroupProxy(clientConfig *platformclientv2.Configuration) *groupProxy {
	if internalProxy == nil {
		internalProxy = newGroupProxy(clientConfig)
	}

	return internalProxy
}

func (p *groupProxy) createGroup(ctx context.Context, group *platformclientv2.Groupcreate) (*platformclientv2.Group, *platformclientv2.APIResponse, error) {
	return p.createGroupAttr(ctx, p, group)
}

func (p *groupProxy) updateGroup(ctx context.Context, id string, group *platformclientv2.Groupupdate) (*platformclientv2.Group, *platformclientv2.APIResponse, error) {
	return p.updateGroupAttr(ctx, p, id, group)
}

func (p *groupProxy) deleteGroup(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteGroupAttr(ctx, p, id)
}

func (p *groupProxy) getAllGroups(ctx context.Context) (*[]platformclientv2.Group, *platformclientv2.APIResponse, error) {
	return p.getAllGroupAttr(ctx, p)
}

func (p *groupProxy) getGroupById(ctx context.Context, id string) (*platformclientv2.Group, *platformclientv2.APIResponse, error) {
	return p.getGroupByIdAttr(ctx, p, id)
}

func (p *groupProxy) addGroupMembers(ctx context.Context, id string, members *platformclientv2.Groupmembersupdate) (*interface{}, *platformclientv2.APIResponse, error) {
	return p.addGroupMembersAttr(ctx, p, id, members)
}

func (p *groupProxy) deleteGroupMembers(ctx context.Context, id string, members string) (*interface{}, *platformclientv2.APIResponse, error) {
	return p.deleteGroupMembersAttr(ctx, p, id, members)
}

func (p *groupProxy) getGroupMembers(ctx context.Context, id string) (*[]string, *platformclientv2.APIResponse, error) {
	return p.getGroupMembersAttr(ctx, p, id)
}

func (p *groupProxy) getGroupsByName(ctx context.Context, name string) (*platformclientv2.Groupssearchresponse, *platformclientv2.APIResponse, error) {
	return p.getGroupByNameAttr(ctx, p, name)
}

func createGroupFn(_ context.Context, p *groupProxy, group *platformclientv2.Groupcreate) (*platformclientv2.Group, *platformclientv2.APIResponse, error) {
	return p.groupsApi.PostGroups(*group)
}

func updateGroupFn(_ context.Context, p *groupProxy, id string, group *platformclientv2.Groupupdate) (*platformclientv2.Group, *platformclientv2.APIResponse, error) {
	return p.groupsApi.PutGroup(id, *group)
}

func deleteGroupFn(_ context.Context, p *groupProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.groupsApi.DeleteGroup(id)
}

func getGroupByIdFn(_ context.Context, p *groupProxy, id string) (*platformclientv2.Group, *platformclientv2.APIResponse, error) {
	group := rc.GetCacheItem(p.groupCache, id)
	if group != nil {
		return group, nil, nil
	}
	return p.groupsApi.GetGroup(id)
}

func addGroupMembersFn(_ context.Context, p *groupProxy, id string, members *platformclientv2.Groupmembersupdate) (*interface{}, *platformclientv2.APIResponse, error) {
	return p.groupsApi.PostGroupMembers(id, *members)
}

func deleteGroupMembersFn(_ context.Context, p *groupProxy, id string, members string) (*interface{}, *platformclientv2.APIResponse, error) {
	return p.groupsApi.DeleteGroupMembers(id, members)
}

func getGroupMembersFn(_ context.Context, p *groupProxy, id string) (*[]string, *platformclientv2.APIResponse, error) {
	members, response, err := p.groupsApi.GetGroupIndividuals(id)

	if err != nil {
		return nil, response, err
	}

	var existingMembers []string
	if members.Entities != nil {
		for _, member := range *members.Entities {
			existingMembers = append(existingMembers, *member.Id)
		}
	}
	return &existingMembers, nil, nil
}

func getGroupByNameFn(_ context.Context, p *groupProxy, name string) (*platformclientv2.Groupssearchresponse, *platformclientv2.APIResponse, error) {
	exactSearchType := "EXACT"
	nameField := "name"
	nameStr := name

	searchCriteria := platformclientv2.Groupsearchcriteria{
		VarType: &exactSearchType,
		Value:   &nameStr,
		Fields:  &[]string{nameField},
	}

	groups, resp, getErr := p.groupsApi.PostGroupsSearch(platformclientv2.Groupsearchrequest{
		Query: &[]platformclientv2.Groupsearchcriteria{searchCriteria},
	})

	return groups, resp, getErr
}

func getAllGroupFn(_ context.Context, p *groupProxy) (*[]platformclientv2.Group, *platformclientv2.APIResponse, error) {
	var allGroups []platformclientv2.Group
	const pageSize = 100

	groups, resp, getErr := p.groupsApi.GetGroups(pageSize, 1, nil, nil, "")
	if getErr != nil {
		return nil, resp, fmt.Errorf("failed to get first page of groups: %v", getErr)
	}

	allGroups = append(allGroups, *groups.Entities...)

	for pageNum := 2; pageNum <= *groups.PageCount; pageNum++ {
		groups, resp, getErr := p.groupsApi.GetGroups(pageSize, pageNum, nil, nil, "")
		if getErr != nil {
			return nil, resp, fmt.Errorf("failed to get page of groups: %v", getErr)
		}
		allGroups = append(allGroups, *groups.Entities...)
	}

	for _, group := range allGroups {
		rc.SetCache(p.groupCache, *group.Id, group)
	}

	return &allGroups, nil, nil
}
