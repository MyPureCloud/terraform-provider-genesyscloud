package routing_skill_group

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *routingSkillGroupsProxy

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
	}
}

// getRoutingSkillGroupsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getRoutingSkillGroupsProxy(clientConfig *platformclientv2.Configuration) *routingSkillGroupsProxy {
	if internalProxy == nil {
		internalProxy = newRoutingSkillGroupsProxy(clientConfig)
	}
	return internalProxy
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
	var (
		allSkillGroups []platformclientv2.Skillgroupdefinition
		pageSize       = 100
		after          string
		err            error
		response       *platformclientv2.APIResponse
	)

	for i := 0; ; i++ {
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
	return &allSkillGroups, response, nil
}

// createRoutingSkillGroupsFn is an implementation function for creating a Genesys Cloud routing skill groups
func createRoutingSkillGroupsFn(ctx context.Context, p *routingSkillGroupsProxy, routingSkillGroups *platformclientv2.Skillgroupwithmemberdivisions) (*platformclientv2.Skillgroupwithmemberdivisions, *platformclientv2.APIResponse, error) {
	return p.routingApi.PostRoutingSkillgroups(*routingSkillGroups)
}

// getRoutingSkillGroupsByIdFn is an implementation of the function to get a Genesys Cloud routing skill groups by Id
func getRoutingSkillGroupsByIdFn(ctx context.Context, p *routingSkillGroupsProxy, id string) (*platformclientv2.Skillgroup, *platformclientv2.APIResponse, error) {
	return p.routingApi.GetRoutingSkillgroup(id)
}

// getRoutingSkillGroupsIdByNameFn is an implementation of the function to get a Genesys Cloud routing skill groups by name
func getRoutingSkillGroupsIdByNameFn(ctx context.Context, p *routingSkillGroupsProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
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
	return p.routingApi.PatchRoutingSkillgroup(id, *routingSkillGroups)
}

// deleteRoutingSkillGroupsFn is an implementation function for deleting a Genesys Cloud routing skill groups
func deleteRoutingSkillGroupsFn(ctx context.Context, p *routingSkillGroupsProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.routingApi.DeleteRoutingSkillgroup(id)
}

func createRoutingSkillGroupsMemberDivisionFn(ctx context.Context, p *routingSkillGroupsProxy, id string, reqBody platformclientv2.Skillgroupmemberdivisions) (*platformclientv2.APIResponse, error) {
	return p.routingApi.PostRoutingSkillgroupMembersDivisions(id, reqBody)
}

func getRoutingSkillGroupsMemberDivisonFn(ctx context.Context, p *routingSkillGroupsProxy, id string) (*platformclientv2.Skillgroupmemberdivisionlist, *platformclientv2.APIResponse, error) {
	return p.routingApi.GetRoutingSkillgroupMembersDivisions(id, "")
}
