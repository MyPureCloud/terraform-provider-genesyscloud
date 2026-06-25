package group

import (
	"context"
	"fmt"
	"log"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
)

type createGroupFunc func(ctx context.Context, p *groupProxy, group *platformclientv2.Groupcreate) (*platformclientv2.Group, *platformclientv2.APIResponse, error)
type getAllGroupFunc func(ctx context.Context, p *groupProxy) (*[]platformclientv2.Group, *platformclientv2.APIResponse, error)
type updateGroupFunc func(ctx context.Context, p *groupProxy, id string, group *platformclientv2.Groupupdate) (*platformclientv2.Group, *platformclientv2.APIResponse, error)
type getGroupByIdFunc func(ctx context.Context, p *groupProxy, id string) (*platformclientv2.Group, *platformclientv2.APIResponse, error)
type addGroupMembersFunc func(ctx context.Context, p *groupProxy, id string, members *platformclientv2.Groupmembersupdate) (*interface{}, *platformclientv2.APIResponse, error)
type deleteGroupMembersFunc func(ctx context.Context, p *groupProxy, id string, members string) (*interface{}, *platformclientv2.APIResponse, error)
type getGroupMembersFunc func(ctx context.Context, p *groupProxy, id string) (*[]string, *platformclientv2.APIResponse, error)
type getGroupByNameFunc func(ctx context.Context, p *groupProxy, name string) (*platformclientv2.Groupssearchresponse, *platformclientv2.APIResponse, error)
type deleteGroupFunc func(ctx context.Context, p *groupProxy, id string) (*platformclientv2.APIResponse, error)
type updateGroupVoicemailPolicyFunc func(ctx context.Context, p *groupProxy, id string, policy *platformclientv2.Voicemailgrouppolicy) (*platformclientv2.Voicemailgrouppolicy, *platformclientv2.APIResponse, error)
type getGroupVoicemailPolicyFunc func(ctx context.Context, p *groupProxy, id string) (*platformclientv2.Voicemailgrouppolicy, *platformclientv2.APIResponse, error)

type groupProxy struct {
	clientConfig                   *platformclientv2.Configuration
	groupsApi                      *platformclientv2.GroupsApi
	voicemailApi                   *platformclientv2.VoicemailApi
	createGroupAttr                createGroupFunc
	getAllGroupAttr                getAllGroupFunc
	updateGroupAttr                updateGroupFunc
	deleteGroupAttr                deleteGroupFunc
	getGroupByNameAttr             getGroupByNameFunc
	getGroupByIdAttr               getGroupByIdFunc
	addGroupMembersAttr            addGroupMembersFunc
	deleteGroupMembersAttr         deleteGroupMembersFunc
	getGroupMembersAttr            getGroupMembersFunc
	updateGroupVoicemailPolicyAttr updateGroupVoicemailPolicyFunc
	getGroupVoicemailPolicyAttr    getGroupVoicemailPolicyFunc
	groupCache                     rc.CacheInterface[platformclientv2.Group]
}

var groupCache = rc.NewResourceCache[platformclientv2.Group]()

// groupMembersCache stores group member IDs per group during export.
var groupMembersCache = rc.NewResourceCache[[]string]()

// groupVoicemailPolicyCache stores voicemail group policies per group during export.
var groupVoicemailPolicyCache = rc.NewResourceCache[platformclientv2.Voicemailgrouppolicy]()

func invalidateGroupMembersCache(groupID string) {
	rc.DeleteCacheItem(groupMembersCache, groupID)
}

func invalidateGroupVoicemailPolicyCache(groupID string) {
	rc.DeleteCacheItem(groupVoicemailPolicyCache, groupID)
}

func invalidateGroupDetailCaches(groupID string) {
	invalidateGroupMembersCache(groupID)
	invalidateGroupVoicemailPolicyCache(groupID)
}

func newGroupProxy(clientConfig *platformclientv2.Configuration) *groupProxy {
	api := platformclientv2.NewGroupsApiWithConfig(clientConfig)
	voicemailApi := platformclientv2.NewVoicemailApiWithConfig(clientConfig)
	return &groupProxy{
		clientConfig:                   clientConfig,
		groupsApi:                      api,
		voicemailApi:                   voicemailApi,
		createGroupAttr:                createGroupFn,
		getAllGroupAttr:                getAllGroupFn,
		updateGroupAttr:                updateGroupFn,
		deleteGroupAttr:                deleteGroupFn,
		getGroupByNameAttr:             getGroupByNameFn,
		getGroupByIdAttr:               getGroupByIdFn,
		addGroupMembersAttr:            addGroupMembersFn,
		deleteGroupMembersAttr:         deleteGroupMembersFn,
		getGroupMembersAttr:            getGroupMembersFn,
		updateGroupVoicemailPolicyAttr: updateGroupVoicemailPolicyFn,
		getGroupVoicemailPolicyAttr:    getGroupVoicemailPolicyFn,
		groupCache:                     groupCache,
	}
}

func getGroupProxy(clientConfig *platformclientv2.Configuration) *groupProxy {
	return newGroupProxy(clientConfig)
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

func (p *groupProxy) updateGroupVoicemailPolicy(ctx context.Context, id string, policy *platformclientv2.Voicemailgrouppolicy) (*platformclientv2.Voicemailgrouppolicy, *platformclientv2.APIResponse, error) {
	return p.updateGroupVoicemailPolicyAttr(ctx, p, id, policy)
}

func (p *groupProxy) getGroupVoicemailPolicy(ctx context.Context, id string) (*platformclientv2.Voicemailgrouppolicy, *platformclientv2.APIResponse, error) {
	return p.getGroupVoicemailPolicyAttr(ctx, p, id)
}

func createGroupFn(ctx context.Context, p *groupProxy, group *platformclientv2.Groupcreate) (*platformclientv2.Group, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	return p.groupsApi.PostGroups(*group)
}

func updateGroupFn(ctx context.Context, p *groupProxy, id string, group *platformclientv2.Groupupdate) (*platformclientv2.Group, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	result, resp, err := p.groupsApi.PutGroup(id, *group)
	if err != nil {
		return nil, resp, err
	}

	invalidateGroupDetailCaches(id)
	if result != nil {
		rc.SetCache(p.groupCache, id, *result)
	}
	return result, resp, nil
}

func deleteGroupFn(ctx context.Context, p *groupProxy, id string) (*platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	resp, err := p.groupsApi.DeleteGroup(id)
	if err != nil {
		return resp, err
	}
	rc.DeleteCacheItem(p.groupCache, id)
	invalidateGroupDetailCaches(id)
	return nil, nil
}

func getGroupByIdFn(ctx context.Context, p *groupProxy, id string) (*platformclientv2.Group, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	group := rc.GetCacheItem(p.groupCache, id)
	if group != nil {
		return group, nil, nil
	}
	return p.groupsApi.GetGroup(id)
}

func addGroupMembersFn(ctx context.Context, p *groupProxy, id string, members *platformclientv2.Groupmembersupdate) (*interface{}, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	result, resp, err := p.groupsApi.PostGroupMembers(id, *members)
	if err != nil {
		return nil, resp, err
	}

	invalidateGroupMembersCache(id)
	return result, resp, nil
}

func deleteGroupMembersFn(ctx context.Context, p *groupProxy, id string, members string) (*interface{}, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	result, resp, err := p.groupsApi.DeleteGroupMembers(id, members)
	if err != nil {
		return nil, resp, err
	}

	invalidateGroupMembersCache(id)
	return result, resp, nil
}

func getGroupMembersFn(ctx context.Context, p *groupProxy, id string) (*[]string, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	if cached := rc.GetCacheItem(groupMembersCache, id); cached != nil {
		log.Printf("[GROUP-CACHE] Group %s: members cache hit (%d members)", id, len(*cached))
		return cached, nil, nil
	}

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

	rc.SetCache(groupMembersCache, id, existingMembers)
	return &existingMembers, nil, nil
}

func getGroupByNameFn(ctx context.Context, p *groupProxy, name string) (*platformclientv2.Groupssearchresponse, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

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

func getAllGroupFn(ctx context.Context, p *groupProxy) (*[]platformclientv2.Group, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	var allGroups []platformclientv2.Group
	const pageSize = 500

	groups, resp, getErr := p.groupsApi.GetGroups(pageSize, 1, nil, nil, "")
	if getErr != nil {
		return nil, resp, fmt.Errorf("failed to get first page of groups: %w", getErr)
	}

	if groups.Entities == nil || len(*groups.Entities) == 0 {
		return &allGroups, resp, nil
	}

	allGroups = append(allGroups, *groups.Entities...)

	totalPages := 1
	if groups.PageCount != nil {
		totalPages = *groups.PageCount
	}

	allGroups, resp, getErr = provider.FetchPagesConcurrently(ctx, ResourceType, allGroups, resp, totalPages, p.clientConfig,
		func(ctx context.Context, clientConfig *platformclientv2.Configuration, pageNum int) ([]platformclientv2.Group, *platformclientv2.APIResponse, error) {
			ctx = provider.EnsureResourceContext(ctx, ResourceType)
			pageProxy := newGroupProxy(clientConfig)
			pageGroups, pageResp, pageErr := pageProxy.groupsApi.GetGroups(pageSize, pageNum, nil, nil, "")
			if pageErr != nil {
				return nil, pageResp, fmt.Errorf("failed to get page of groups: %w", pageErr)
			}

			if pageGroups.Entities == nil || len(*pageGroups.Entities) == 0 {
				return []platformclientv2.Group{}, pageResp, nil
			}

			return *pageGroups.Entities, pageResp, nil
		},
	)
	if getErr != nil {
		return nil, resp, getErr
	}

	for _, group := range allGroups {
		rc.SetCache(p.groupCache, *group.Id, group)
	}

	return &allGroups, resp, nil
}

func updateGroupVoicemailPolicyFn(ctx context.Context, p *groupProxy, id string, policy *platformclientv2.Voicemailgrouppolicy) (*platformclientv2.Voicemailgrouppolicy, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	updatedPolicy, resp, err := p.voicemailApi.PatchVoicemailGroupPolicy(id, *policy)
	if err != nil {
		return nil, resp, err
	}

	invalidateGroupVoicemailPolicyCache(id)
	if updatedPolicy != nil {
		rc.SetCache(groupVoicemailPolicyCache, id, *updatedPolicy)
	}
	return updatedPolicy, resp, nil
}

func getGroupVoicemailPolicyFn(ctx context.Context, p *groupProxy, id string) (*platformclientv2.Voicemailgrouppolicy, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	if cached := rc.GetCacheItem(groupVoicemailPolicyCache, id); cached != nil {
		log.Printf("[GROUP-CACHE] Group %s: voicemail policy cache hit", id)
		return cached, nil, nil
	}

	policy, resp, err := p.voicemailApi.GetVoicemailGroupPolicy(id)
	if err != nil {
		return nil, resp, err
	}

	if policy != nil {
		rc.SetCache(groupVoicemailPolicyCache, id, *policy)
	}

	return policy, resp, nil
}
