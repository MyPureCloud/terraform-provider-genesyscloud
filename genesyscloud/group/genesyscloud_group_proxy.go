package group

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
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

var groupCache = rc.NewResourceCache[platformclientv2.Group]()

func newGroupProxy(clientConfig *platformclientv2.Configuration) *groupProxy {
	api := platformclientv2.NewGroupsApiWithConfig(clientConfig)
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

func createGroupFn(_ context.Context, p *groupProxy, group *platformclientv2.Groupcreate) (*platformclientv2.Group, *platformclientv2.APIResponse, error) {
	return p.groupsApi.PostGroups(*group)
}

func updateGroupFn(_ context.Context, p *groupProxy, id string, group *platformclientv2.Groupupdate) (*platformclientv2.Group, *platformclientv2.APIResponse, error) {
	return callUpdateGroupApi(id, group, p.clientConfig)
}

func deleteGroupFn(_ context.Context, p *groupProxy, id string) (*platformclientv2.APIResponse, error) {
	resp, err := p.groupsApi.DeleteGroup(id)
	if err != nil {
		return resp, err
	}
	rc.DeleteCacheItem(p.groupCache, id)
	return nil, nil
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

func callUpdateGroupApi(groupId string, body *platformclientv2.Groupupdate, sdkConfig *platformclientv2.Configuration) (*platformclientv2.Group, *platformclientv2.APIResponse, error) {
	api := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	var httpMethod = "PUT"
	// create path and map variables
	path := api.Configuration.BasePath + "/api/v2/groups/{groupId}"
	path = strings.Replace(path, "{groupId}", url.PathEscape(fmt.Sprintf("%v", groupId)), -1)

	// Converting the groupupdate req to a JSON byte arrays
	b, err := json.Marshal(body)
	if err != nil {
		return nil, nil, err
	}
	gpJson := string(b)

	// Converting the JSON string to a Golang Map
	var gpMap map[string]interface{}
	err = json.Unmarshal([]byte(gpJson), &gpMap)
	if err != nil {
		return nil, nil, err
	}

	// Set the value for ownerIds as empty array if found nil
	ownerIds := gpMap["ownerIds"]
	if ownerIds == nil {
		gpMap["ownerIds"] = []string{}
	}

	headerParams := make(map[string]string)
	queryParams := make(map[string]string)
	formParams := url.Values{}
	var postFileName string
	var fileBytes []byte

	// authentication (PureCloud OAuth) required
	// oauth required
	if api.Configuration.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + api.Configuration.AccessToken
	}
	// add default headers if any
	for key := range api.Configuration.DefaultHeader {
		headerParams[key] = api.Configuration.DefaultHeader[key]
	}

	// Find an replace keys that were altered to avoid clashes with go keywords
	correctedQueryParams := make(map[string]string)
	for k, v := range queryParams {
		if k == "varType" {
			correctedQueryParams["type"] = v
			continue
		}
		correctedQueryParams[k] = v
	}
	queryParams = correctedQueryParams

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	// Convert the golang map to jsonBytes and to string
	jsonBytes, _ := json.Marshal(gpMap)
	jsonStr := string(jsonBytes)

	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &jsonMap)

	var successPayload *platformclientv2.Group
	response, err := api.Configuration.APIClient.CallAPI(path, httpMethod, jsonMap, headerParams, queryParams, formParams, postFileName, fileBytes, "other")
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if err == nil && response.Error != nil {
		err = errors.New(response.ErrorMessage)
	}
	return successPayload, response, err
}
