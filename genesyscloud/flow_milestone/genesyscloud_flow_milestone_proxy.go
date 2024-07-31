package flow_milestone

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_flow_milestone_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *flowMilestoneProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createFlowMilestoneFunc func(ctx context.Context, p *flowMilestoneProxy, flowMilestone *platformclientv2.Flowmilestone) (*platformclientv2.Flowmilestone, *platformclientv2.APIResponse, error)
type getAllFlowMilestoneFunc func(ctx context.Context, p *flowMilestoneProxy) (*[]platformclientv2.Flowmilestone, *platformclientv2.APIResponse, error)
type getFlowMilestoneIdByNameFunc func(ctx context.Context, p *flowMilestoneProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getFlowMilestoneByIdFunc func(ctx context.Context, p *flowMilestoneProxy, id string) (flowMilestone *platformclientv2.Flowmilestone, response *platformclientv2.APIResponse, err error)
type updateFlowMilestoneFunc func(ctx context.Context, p *flowMilestoneProxy, id string, flowMilestone *platformclientv2.Flowmilestone) (*platformclientv2.Flowmilestone, *platformclientv2.APIResponse, error)
type deleteFlowMilestoneFunc func(ctx context.Context, p *flowMilestoneProxy, id string) (response *platformclientv2.APIResponse, err error)

// flowMilestoneProxy contains all of the methods that call genesys cloud APIs.
type flowMilestoneProxy struct {
	clientConfig                 *platformclientv2.Configuration
	architectApi                 *platformclientv2.ArchitectApi
	createFlowMilestoneAttr      createFlowMilestoneFunc
	getAllFlowMilestoneAttr      getAllFlowMilestoneFunc
	getFlowMilestoneIdByNameAttr getFlowMilestoneIdByNameFunc
	getFlowMilestoneByIdAttr     getFlowMilestoneByIdFunc
	updateFlowMilestoneAttr      updateFlowMilestoneFunc
	deleteFlowMilestoneAttr      deleteFlowMilestoneFunc
}

// newFlowMilestoneProxy initializes the flow milestone proxy with all of the data needed to communicate with Genesys Cloud
func newFlowMilestoneProxy(clientConfig *platformclientv2.Configuration) *flowMilestoneProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	return &flowMilestoneProxy{
		clientConfig:                 clientConfig,
		architectApi:                 api,
		createFlowMilestoneAttr:      createFlowMilestoneFn,
		getAllFlowMilestoneAttr:      getAllFlowMilestoneFn,
		getFlowMilestoneIdByNameAttr: getFlowMilestoneIdByNameFn,
		getFlowMilestoneByIdAttr:     getFlowMilestoneByIdFn,
		updateFlowMilestoneAttr:      updateFlowMilestoneFn,
		deleteFlowMilestoneAttr:      deleteFlowMilestoneFn,
	}
}

// getFlowMilestoneProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getFlowMilestoneProxy(clientConfig *platformclientv2.Configuration) *flowMilestoneProxy {
	if internalProxy == nil {
		internalProxy = newFlowMilestoneProxy(clientConfig)
	}
	return internalProxy
}

// createFlowMilestone creates a Genesys Cloud flow milestone
func (p *flowMilestoneProxy) createFlowMilestone(ctx context.Context, flowMilestone *platformclientv2.Flowmilestone) (*platformclientv2.Flowmilestone, *platformclientv2.APIResponse, error) {
	return p.createFlowMilestoneAttr(ctx, p, flowMilestone)
}

// getFlowMilestone retrieves all Genesys Cloud flow milestone
func (p *flowMilestoneProxy) getAllFlowMilestone(ctx context.Context) (*[]platformclientv2.Flowmilestone, *platformclientv2.APIResponse, error) {
	return p.getAllFlowMilestoneAttr(ctx, p)
}

// getFlowMilestoneIdByName returns a single Genesys Cloud flow milestone by a name
func (p *flowMilestoneProxy) getFlowMilestoneIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getFlowMilestoneIdByNameAttr(ctx, p, name)
}

// getFlowMilestoneById returns a single Genesys Cloud flow milestone by Id
func (p *flowMilestoneProxy) getFlowMilestoneById(ctx context.Context, id string) (flowMilestone *platformclientv2.Flowmilestone, response *platformclientv2.APIResponse, err error) {
	return p.getFlowMilestoneByIdAttr(ctx, p, id)
}

// updateFlowMilestone updates a Genesys Cloud flow milestone
func (p *flowMilestoneProxy) updateFlowMilestone(ctx context.Context, id string, flowMilestone *platformclientv2.Flowmilestone) (*platformclientv2.Flowmilestone, *platformclientv2.APIResponse, error) {
	return p.updateFlowMilestoneAttr(ctx, p, id, flowMilestone)
}

// deleteFlowMilestone deletes a Genesys Cloud flow milestone by Id
func (p *flowMilestoneProxy) deleteFlowMilestone(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteFlowMilestoneAttr(ctx, p, id)
}

// createFlowMilestoneFn is an implementation function for creating a Genesys Cloud flow milestone
func createFlowMilestoneFn(ctx context.Context, p *flowMilestoneProxy, flowMilestone *platformclientv2.Flowmilestone) (*platformclientv2.Flowmilestone, *platformclientv2.APIResponse, error) {
	flowMilestone, resp, err := p.architectApi.PostFlowsMilestones(*flowMilestone)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to create flow milestone: %s ", err)
	}
	return flowMilestone, resp, nil
}

// getAllFlowMilestoneFn is the implementation for retrieving all flow milestone in Genesys Cloud
func getAllFlowMilestoneFn(ctx context.Context, p *flowMilestoneProxy) (*[]platformclientv2.Flowmilestone, *platformclientv2.APIResponse, error) {
	var allFlowMilestones []platformclientv2.Flowmilestone
	const pageSize = 100

	flowMilestones, resp, err := p.architectApi.GetFlowsMilestones(1, pageSize, "", "", nil, "", "", "", nil)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get flow milestone: %v", err)
	}
	if flowMilestones.Entities == nil || len(*flowMilestones.Entities) == 0 {
		return &allFlowMilestones, resp, nil
	}
	for _, flowMilestone := range *flowMilestones.Entities {
		allFlowMilestones = append(allFlowMilestones, flowMilestone)
	}

	for pageNum := 2; pageNum <= *flowMilestones.PageCount; pageNum++ {
		flowMilestones, resp, err := p.architectApi.GetFlowsMilestones(pageNum, pageSize, "", "", nil, "", "", "", nil)
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get flow milestone: %v", err)
		}

		if flowMilestones.Entities == nil || len(*flowMilestones.Entities) == 0 {
			break
		}

		for _, flowMilestone := range *flowMilestones.Entities {
			allFlowMilestones = append(allFlowMilestones, flowMilestone)
		}
	}
	return &allFlowMilestones, resp, nil
}

// getFlowMilestoneIdByNameFn is an implementation of the function to get a Genesys Cloud flow milestone by name
func getFlowMilestoneIdByNameFn(ctx context.Context, p *flowMilestoneProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	flowMilestones, resp, err := p.architectApi.GetFlowsMilestones(1, 100, "", "", nil, name, "", "", nil)
	if err != nil {
		return "", false, resp, err
	}

	if flowMilestones.Entities == nil || len(*flowMilestones.Entities) == 0 {
		return "", true, resp, fmt.Errorf("No flow milestone found with name %s", name)
	}

	for _, flowMilestoneSdk := range *flowMilestones.Entities {
		if *flowMilestoneSdk.Name == name {
			log.Printf("Retrieved the flow milestone id %s by name %s", *flowMilestoneSdk.Id, name)
			return *flowMilestoneSdk.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("Unable to find flow milestone with name %s", name)
}

// getFlowMilestoneByIdFn is an implementation of the function to get a Genesys Cloud flow milestone by Id
func getFlowMilestoneByIdFn(ctx context.Context, p *flowMilestoneProxy, id string) (flowMilestone *platformclientv2.Flowmilestone, response *platformclientv2.APIResponse, err error) {
	flowMilestone, resp, err := p.architectApi.GetFlowsMilestone(id)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to retrieve flow milestone by id %s: %s", id, err)
	}
	return flowMilestone, resp, nil
}

// updateFlowMilestoneFn is an implementation of the function to update a Genesys Cloud flow milestone
func updateFlowMilestoneFn(ctx context.Context, p *flowMilestoneProxy, id string, flowMilestone *platformclientv2.Flowmilestone) (*platformclientv2.Flowmilestone, *platformclientv2.APIResponse, error) {
	flowMilestone, resp, err := p.architectApi.PutFlowsMilestone(id, *flowMilestone)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to update flow milestone: %s", err)
	}
	return flowMilestone, resp, nil
}

// deleteFlowMilestoneFn is an implementation function for deleting a Genesys Cloud flow milestone
func deleteFlowMilestoneFn(ctx context.Context, p *flowMilestoneProxy, id string) (response *platformclientv2.APIResponse, err error) {
	_, resp, err := p.architectApi.DeleteFlowsMilestone(id)
	if err != nil {
		return resp, fmt.Errorf("Failed to delete flow milestone: %s", err)
	}
	return resp, nil
}
