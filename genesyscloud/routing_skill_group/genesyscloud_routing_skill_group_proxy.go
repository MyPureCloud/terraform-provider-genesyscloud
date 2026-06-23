package routing_skill_group

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
)

var skillGroupCache = rc.NewResourceCache[platformclientv2.Skillgroup]()

// skillGroupListCache stores paginated skill group list results keyed by name during export.
var skillGroupListCache = rc.NewResourceCache[[]platformclientv2.Skillgroupdefinition]()

// skillGroupMemberDivisionsCache stores member division listings per skill group during export.
var skillGroupMemberDivisionsCache = rc.NewResourceCache[platformclientv2.Skillgroupmemberdivisionlist]()

func invalidateSkillGroupListCache() {
	if tfexporter_state.IsExporterActive() {
		skillGroupListCache = rc.NewResourceCache[[]platformclientv2.Skillgroupdefinition]()
	}
}

func storeSkillGroupInCache(skillGroup *platformclientv2.Skillgroup) {
	if skillGroup != nil && skillGroup.Id != nil {
		rc.SetCache(skillGroupCache, *skillGroup.Id, *skillGroup)
	}
}

func invalidateSkillGroupMemberDivisionsCache(skillGroupID string) {
	rc.DeleteCacheItem(skillGroupMemberDivisionsCache, skillGroupID)
}

func invalidateSkillGroupCaches(skillGroupID string) {
	rc.DeleteCacheItem(skillGroupCache, skillGroupID)
	invalidateSkillGroupMemberDivisionsCache(skillGroupID)
	invalidateSkillGroupListCache()
}

type getAllRoutingSkillGroupsFunc func(ctx context.Context, p *routingSkillGroupsProxy, name string) (*[]platformclientv2.Skillgroupdefinition, *platformclientv2.APIResponse, error)
type createRoutingSkillGroupsFunc func(ctx context.Context, p *routingSkillGroupsProxy, skillGroupWithMemberDivisions *platformclientv2.Skillgroupwithmemberdivisions) (*platformclientv2.Skillgroupwithmemberdivisions, *platformclientv2.APIResponse, error)
type getRoutingSkillGroupsByIdFunc func(ctx context.Context, p *routingSkillGroupsProxy, id string) (*platformclientv2.Skillgroup, *platformclientv2.APIResponse, error)
type getRoutingSkillGroupsIdByNameFunc func(ctx context.Context, p *routingSkillGroupsProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type updateRoutingSkillGroupsFunc func(ctx context.Context, p *routingSkillGroupsProxy, id string, skillGroupWithMemberDivisions *platformclientv2.Skillgroup) (*platformclientv2.Skillgroup, *platformclientv2.APIResponse, error)
type deleteRoutingSkillGroupsFunc func(ctx context.Context, p *routingSkillGroupsProxy, id string) (*platformclientv2.APIResponse, error)
type createRoutingSkillGroupsMemberDivisionFunc func(ctx context.Context, p *routingSkillGroupsProxy, id string, reqBody platformclientv2.Skillgroupmemberdivisions) (*platformclientv2.APIResponse, error)
type getRoutingSkillGroupsMemberDivisonFunc func(ctx context.Context, p *routingSkillGroupsProxy, id string) (*platformclientv2.Skillgroupmemberdivisionlist, *platformclientv2.APIResponse, error)

// routingSkillGroupsProxy contains all of the methods that call genesys cloud APIs.
type routingSkillGroupsProxy struct {
	clientConfig                               *platformclientv2.Configuration
	routingApi                                 *platformclientv2.RoutingApi
	createRoutingSkillGroupsAttr               createRoutingSkillGroupsFunc
	getAllRoutingSkillGroupsAttr               getAllRoutingSkillGroupsFunc
	getRoutingSkillGroupsIdByNameAttr          getRoutingSkillGroupsIdByNameFunc
	getRoutingSkillGroupsByIdAttr              getRoutingSkillGroupsByIdFunc
	updateRoutingSkillGroupsAttr               updateRoutingSkillGroupsFunc
	deleteRoutingSkillGroupsAttr               deleteRoutingSkillGroupsFunc
	createRoutingSkillGroupsMemberDivisionAttr createRoutingSkillGroupsMemberDivisionFunc
	getRoutingSkillGroupsMemberDivisonAttr     getRoutingSkillGroupsMemberDivisonFunc
	skillGroupCache                            rc.CacheInterface[platformclientv2.Skillgroup]
}

// newRoutingSkillGroupsProxy initializes the routing skill groups proxy with all of the data needed to communicate with Genesys Cloud
func newRoutingSkillGroupsProxy(clientConfig *platformclientv2.Configuration) *routingSkillGroupsProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	return &routingSkillGroupsProxy{
		clientConfig:                               clientConfig,
		routingApi:                                 api,
		createRoutingSkillGroupsAttr:               createRoutingSkillGroupsFn,
		getAllRoutingSkillGroupsAttr:               getAllRoutingSkillGroupsFn,
		getRoutingSkillGroupsIdByNameAttr:          getRoutingSkillGroupsIdByNameFn,
		getRoutingSkillGroupsByIdAttr:              getRoutingSkillGroupsByIdFn,
		updateRoutingSkillGroupsAttr:               updateRoutingSkillGroupsFn,
		deleteRoutingSkillGroupsAttr:               deleteRoutingSkillGroupsFn,
		createRoutingSkillGroupsMemberDivisionAttr: createRoutingSkillGroupsMemberDivisionFn,
		getRoutingSkillGroupsMemberDivisonAttr:     getRoutingSkillGroupsMemberDivisonFn,
		skillGroupCache:                            skillGroupCache,
	}
}

// getRoutingSkillGroupsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getRoutingSkillGroupsProxy(clientConfig *platformclientv2.Configuration) *routingSkillGroupsProxy {
	return newRoutingSkillGroupsProxy(clientConfig)
}

// getRoutingSkillGroups retrieves all Genesys Cloud routing skill groups
func (p *routingSkillGroupsProxy) getAllRoutingSkillGroups(ctx context.Context, name string) (*[]platformclientv2.Skillgroupdefinition, *platformclientv2.APIResponse, error) {
	return p.getAllRoutingSkillGroupsAttr(ctx, p, name)
}

// createRoutingSkillGroups creates a Genesys Cloud routing skill groups
func (p *routingSkillGroupsProxy) createRoutingSkillGroups(ctx context.Context, routingSkillGroups *platformclientv2.Skillgroupwithmemberdivisions) (*platformclientv2.Skillgroupwithmemberdivisions, *platformclientv2.APIResponse, error) {
	return p.createRoutingSkillGroupsAttr(ctx, p, routingSkillGroups)
}

// getRoutingSkillGroupsById returns a single Genesys Cloud routing skill groups by Id
func (p *routingSkillGroupsProxy) getRoutingSkillGroupsById(ctx context.Context, id string) (*platformclientv2.Skillgroup, *platformclientv2.APIResponse, error) {
	return p.getRoutingSkillGroupsByIdAttr(ctx, p, id)
}

// getRoutingSkillGroupsIdByName returns a single Genesys Cloud routing skill groups by a name
func (p *routingSkillGroupsProxy) getRoutingSkillGroupsIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getRoutingSkillGroupsIdByNameAttr(ctx, p, name)
}

// updateRoutingSkillGroups updates a Genesys Cloud routing skill groups
func (p *routingSkillGroupsProxy) updateRoutingSkillGroups(ctx context.Context, id string, routingSkillGroups *platformclientv2.Skillgroup) (*platformclientv2.Skillgroup, *platformclientv2.APIResponse, error) {
	return p.updateRoutingSkillGroupsAttr(ctx, p, id, routingSkillGroups)
}

// deleteRoutingSkillGroups deletes a Genesys Cloud routing skill groups by Id
func (p *routingSkillGroupsProxy) deleteRoutingSkillGroups(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteRoutingSkillGroupsAttr(ctx, p, id)
}

func (p *routingSkillGroupsProxy) createRoutingSkillGroupsMemberDivision(ctx context.Context, id string, reqBody platformclientv2.Skillgroupmemberdivisions) (*platformclientv2.APIResponse, error) {
	return p.createRoutingSkillGroupsMemberDivisionAttr(ctx, p, id, reqBody)
}

func (p *routingSkillGroupsProxy) getRoutingSkillGroupsMemberDivison(ctx context.Context, id string) (*platformclientv2.Skillgroupmemberdivisionlist, *platformclientv2.APIResponse, error) {
	return p.getRoutingSkillGroupsMemberDivisonAttr(ctx, p, id)
}

// getAllRoutingSkillGroupsFn is the implementation for retrieving all routing skill groups in Genesys Cloud
func getAllRoutingSkillGroupsFn(ctx context.Context, p *routingSkillGroupsProxy, name string) (*[]platformclientv2.Skillgroupdefinition, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	cacheKey := name
	if cached := rc.GetCacheItem(skillGroupListCache, cacheKey); cached != nil {
		log.Printf("[SKILL-GROUP-CACHE] key=%s: cache hit (%d skill groups)", cacheKey, len(*cached))
		return cached, nil, nil
	}

	var (
		allSkillGroups []platformclientv2.Skillgroupdefinition
		pageSize       = 100
		after          string
		err            error
		response       *platformclientv2.APIResponse
	)

	for {
		skillGroups, resp, getErr := p.routingApi.GetRoutingSkillgroups(pageSize, name, after, "")
		response = resp
		if getErr != nil {
			return nil, resp, fmt.Errorf("unable to get routing skill groups %s", getErr)
		}

		if skillGroups.Entities == nil || len(*skillGroups.Entities) == 0 {
			break
		}

		allSkillGroups = append(allSkillGroups, *skillGroups.Entities...)

		if skillGroups.NextUri == nil || *skillGroups.NextUri == "" {
			break
		}

		after, err = util.GetQueryParamValueFromUri(*skillGroups.NextUri, "after")
		if err != nil {
			return nil, resp, fmt.Errorf("unable to parse after cursor from skill groups next uri: %v", err)
		}
		if after == "" {
			break
		}
	}

	rc.SetCache(skillGroupListCache, cacheKey, allSkillGroups)
	log.Printf("[SKILL-GROUP-CACHE] key=%s: cached %d skill groups", cacheKey, len(allSkillGroups))

	return &allSkillGroups, response, nil
}

// createRoutingSkillGroupsFn is an implementation function for creating a Genesys Cloud routing skill groups
func createRoutingSkillGroupsFn(ctx context.Context, p *routingSkillGroupsProxy, routingSkillGroups *platformclientv2.Skillgroupwithmemberdivisions) (*platformclientv2.Skillgroupwithmemberdivisions, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	return p.routingApi.PostRoutingSkillgroups(*routingSkillGroups)
}

// getRoutingSkillGroupsByIdFn is an implementation of the function to get a Genesys Cloud routing skill groups by Id
func getRoutingSkillGroupsByIdFn(ctx context.Context, p *routingSkillGroupsProxy, id string) (*platformclientv2.Skillgroup, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	if skillGroup := rc.GetCacheItem(p.skillGroupCache, id); skillGroup != nil {
		log.Printf("[SKILL-GROUP-CACHE] Skill group %s: cache hit", id)
		return skillGroup, nil, nil
	}

	skillGroup, resp, err := p.routingApi.GetRoutingSkillgroup(id)
	if err != nil {
		return nil, resp, err
	}

	storeSkillGroupInCache(skillGroup)
	if tfexporter_state.IsExporterActive() && skillGroup != nil {
		log.Printf("[SKILL-GROUP-CACHE] Skill group %s: cached skill group", id)
	}

	return skillGroup, resp, nil
}

// getRoutingSkillGroupsIdByNameFn is an implementation of the function to get a Genesys Cloud routing skill groups by name
func getRoutingSkillGroupsIdByNameFn(ctx context.Context, p *routingSkillGroupsProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	skillGroup, resp, err := getAllRoutingSkillGroupsFn(ctx, p, name)
	if err != nil {
		return "", resp, false, err
	}

	if skillGroup == nil || len(*skillGroup) == 0 {
		return "", resp, true, fmt.Errorf("no skill groups found with name %s", name)
	}

	for _, skillGroup := range *skillGroup {
		if *skillGroup.Name == name {
			log.Printf("Retrieved the routing skill groups id %s by name %s", *skillGroup.Id, name)
			return *skillGroup.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("unable to find routing skill groups with name %s", name)
}

// updateRoutingSkillGroupsFn is an implementation of the function to update a Genesys Cloud routing skill groups
func updateRoutingSkillGroupsFn(ctx context.Context, p *routingSkillGroupsProxy, id string, routingSkillGroups *platformclientv2.Skillgroup) (*platformclientv2.Skillgroup, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	skillGroup, resp, err := p.routingApi.PatchRoutingSkillgroup(id, *routingSkillGroups)
	if err != nil {
		return nil, resp, err
	}

	invalidateSkillGroupCaches(id)
	storeSkillGroupInCache(skillGroup)
	return skillGroup, resp, nil
}

// deleteRoutingSkillGroupsFn is an implementation function for deleting a Genesys Cloud routing skill groups
func deleteRoutingSkillGroupsFn(ctx context.Context, p *routingSkillGroupsProxy, id string) (*platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	resp, err := p.routingApi.DeleteRoutingSkillgroup(id)
	if err != nil {
		return resp, err
	}

	invalidateSkillGroupCaches(id)
	return resp, nil
}

func createRoutingSkillGroupsMemberDivisionFn(ctx context.Context, p *routingSkillGroupsProxy, id string, reqBody platformclientv2.Skillgroupmemberdivisions) (*platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	resp, err := p.routingApi.PostRoutingSkillgroupMembersDivisions(id, reqBody)
	if err != nil {
		return resp, err
	}

	invalidateSkillGroupMemberDivisionsCache(id)
	rc.DeleteCacheItem(skillGroupCache, id)
	return resp, nil
}

func getRoutingSkillGroupsMemberDivisonFn(ctx context.Context, p *routingSkillGroupsProxy, id string) (*platformclientv2.Skillgroupmemberdivisionlist, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	if cached := rc.GetCacheItem(skillGroupMemberDivisionsCache, id); cached != nil {
		log.Printf("[SKILL-GROUP-CACHE] Skill group %s: member divisions cache hit", id)
		return cached, nil, nil
	}

	memberDivisions, resp, err := p.routingApi.GetRoutingSkillgroupMembersDivisions(id, "")
	if err != nil {
		return nil, resp, err
	}

	if memberDivisions != nil {
		rc.SetCache(skillGroupMemberDivisionsCache, id, *memberDivisions)
	}
	return memberDivisions, resp, nil
}
