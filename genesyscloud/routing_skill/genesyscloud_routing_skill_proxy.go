package routing_skill

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

type getAllRoutingSkillsFunc func(ctx context.Context, p *routingSkillProxy, name string) (*[]platformclientv2.Routingskill, *platformclientv2.APIResponse, error)
type createRoutingSkillFunc func(ctx context.Context, p *routingSkillProxy, routingSkill *platformclientv2.Routingskill) (*platformclientv2.Routingskill, *platformclientv2.APIResponse, error)
type getRoutingSkillByIdFunc func(ctx context.Context, p *routingSkillProxy, id string) (*platformclientv2.Routingskill, *platformclientv2.APIResponse, error)
type getRoutingSkillIdByNameFunc func(ctx context.Context, p *routingSkillProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type deleteRoutingSkillFunc func(ctx context.Context, p *routingSkillProxy, id string) (*platformclientv2.APIResponse, error)

var routingSkillCache = rc.NewResourceCache[platformclientv2.Routingskill]()

// routingSkillProxy contains all of the methods that call genesys cloud APIs.
type routingSkillProxy struct {
	clientConfig                *platformclientv2.Configuration
	routingApi                  *platformclientv2.RoutingApi
	createRoutingSkillAttr      createRoutingSkillFunc
	getAllRoutingSkillsAttr     getAllRoutingSkillsFunc
	getRoutingSkillIdByNameAttr getRoutingSkillIdByNameFunc
	getRoutingSkillByIdAttr     getRoutingSkillByIdFunc
	deleteRoutingSkillAttr      deleteRoutingSkillFunc
	routingSkillCache           rc.CacheInterface[platformclientv2.Routingskill]
}

// newRoutingSkillProxy initializes the routing skill proxy with all of the data needed to communicate with Genesys Cloud
func newRoutingSkillProxy(clientConfig *platformclientv2.Configuration) *routingSkillProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	return &routingSkillProxy{
		clientConfig:                clientConfig,
		routingApi:                  api,
		createRoutingSkillAttr:      createRoutingSkillFn,
		getAllRoutingSkillsAttr:     getAllRoutingSkillsFn,
		getRoutingSkillIdByNameAttr: getRoutingSkillIdByNameFn,
		getRoutingSkillByIdAttr:     getRoutingSkillByIdFn,
		deleteRoutingSkillAttr:      deleteRoutingSkillFn,
		routingSkillCache:           routingSkillCache,
	}
}

func getRoutingSkillProxy(clientConfig *platformclientv2.Configuration) *routingSkillProxy {
	return newRoutingSkillProxy(clientConfig)
}

func (p *routingSkillProxy) getAllRoutingSkills(ctx context.Context, name string) (*[]platformclientv2.Routingskill, *platformclientv2.APIResponse, error) {
	return p.getAllRoutingSkillsAttr(ctx, p, name)
}

func (p *routingSkillProxy) createRoutingSkill(ctx context.Context, routingSkill *platformclientv2.Routingskill) (*platformclientv2.Routingskill, *platformclientv2.APIResponse, error) {
	return p.createRoutingSkillAttr(ctx, p, routingSkill)
}

func (p *routingSkillProxy) getRoutingSkillById(ctx context.Context, id string) (*platformclientv2.Routingskill, *platformclientv2.APIResponse, error) {
	return p.getRoutingSkillByIdAttr(ctx, p, id)
}

func (p *routingSkillProxy) getRoutingSkillIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getRoutingSkillIdByNameAttr(ctx, p, name)
}

func (p *routingSkillProxy) deleteRoutingSkill(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteRoutingSkillAttr(ctx, p, id)
}

func getAllRoutingSkillsFn(ctx context.Context, p *routingSkillProxy, name string) (*[]platformclientv2.Routingskill, *platformclientv2.APIResponse, error) {
	var allRoutingSkills []platformclientv2.Routingskill
	const pageSize = 100

	routingSkills, resp, err := p.routingApi.GetRoutingSkills(pageSize, 1, name, nil)
	if err != nil {
		return nil, resp, err
	}

	if routingSkills.Entities == nil || len(*routingSkills.Entities) == 0 {
		return &allRoutingSkills, resp, nil
	}

	allRoutingSkills = append(allRoutingSkills, *routingSkills.Entities...)

	for pageNum := 2; pageNum <= *routingSkills.PageCount; pageNum++ {
		routingSkills, _, err := p.routingApi.GetRoutingSkills(pageSize, pageNum, name, nil)
		if err != nil {
			return nil, resp, err
		}

		if routingSkills.Entities == nil || len(*routingSkills.Entities) == 0 {
			break
		}

		allRoutingSkills = append(allRoutingSkills, *routingSkills.Entities...)

	}

	for _, skill := range allRoutingSkills {
		rc.SetCache(p.routingSkillCache, *skill.Id, skill)
	}

	return &allRoutingSkills, resp, nil
}

func createRoutingSkillFn(ctx context.Context, p *routingSkillProxy, routingSkill *platformclientv2.Routingskill) (*platformclientv2.Routingskill, *platformclientv2.APIResponse, error) {
	return p.routingApi.PostRoutingSkills(*routingSkill)
}

func getRoutingSkillByIdFn(ctx context.Context, p *routingSkillProxy, id string) (*platformclientv2.Routingskill, *platformclientv2.APIResponse, error) {
	if skill := rc.GetCacheItem(p.routingSkillCache, id); skill != nil {
		return skill, nil, nil
	}
	return p.routingApi.GetRoutingSkill(id)
}

func getRoutingSkillIdByNameFn(ctx context.Context, p *routingSkillProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	routingSkills, resp, err := getAllRoutingSkillsFn(ctx, p, name)
	if err != nil {
		return "", resp, false, err
	}

	noneFoundError := fmt.Errorf("no routing skills found with name '%s'", name)

	if routingSkills == nil || len(*routingSkills) == 0 {
		return "", resp, true, noneFoundError
	}

	for _, routingSkill := range *routingSkills {
		if *routingSkill.Name == name {
			log.Printf("Retrieved the routing skill id %s by name %s", *routingSkill.Id, name)
			return *routingSkill.Id, resp, false, nil
		}
	}

	return "", resp, true, noneFoundError
}

func deleteRoutingSkillFn(ctx context.Context, p *routingSkillProxy, id string) (*platformclientv2.APIResponse, error) {
	resp, err := p.routingApi.DeleteRoutingSkill(id)
	if err != nil {
		return resp, err
	}
	rc.DeleteCacheItem(p.routingSkillCache, id)
	return nil, nil
}
