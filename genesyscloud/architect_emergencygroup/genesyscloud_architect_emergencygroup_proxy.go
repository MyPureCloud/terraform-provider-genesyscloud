package architect_emergencygroup

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *architectEmergencyGroupProxy

type createArchitectEmergencyGroupFunc func(ctx context.Context, p *architectEmergencyGroupProxy, emergencyGroup platformclientv2.Emergencygroup) (*platformclientv2.Emergencygroup, *platformclientv2.APIResponse, error)
type getAllArchitectEmergencyGroupFunc func(ctx context.Context, p *architectEmergencyGroupProxy) (*[]platformclientv2.Emergencygroup, *platformclientv2.APIResponse, error)
type getArchitectEmergencyGroupFunc func(ctx context.Context, p *architectEmergencyGroupProxy, emergencyGroupId string) (emergencyGroup *platformclientv2.Emergencygroup, apiResponse *platformclientv2.APIResponse, err error)
type updateArchitectEmergencyGroupFunc func(ctx context.Context, p *architectEmergencyGroupProxy, emergencyGroupId string, emergencyGroup platformclientv2.Emergencygroup) (*platformclientv2.Emergencygroup, *platformclientv2.APIResponse, error)
type deleteArchitectEmergencyGroupFunc func(ctx context.Context, p *architectEmergencyGroupProxy, emergencyGroupId string) (*platformclientv2.APIResponse, error)
type getArchitectEmergencyGroupIdByNameFunc func(ctx context.Context, p *architectEmergencyGroupProxy, name string) (emergencyGroup *platformclientv2.Emergencygrouplisting, apiResponse *platformclientv2.APIResponse, err error)

type architectEmergencyGroupProxy struct {
	clientConfig                           *platformclientv2.Configuration
	architectApi                           *platformclientv2.ArchitectApi
	createArchitectEmergencyGroupAttr      createArchitectEmergencyGroupFunc
	getAllArchitectEmergencyGroupAttr      getAllArchitectEmergencyGroupFunc
	getArchitectEmergencyGroupAttr         getArchitectEmergencyGroupFunc
	getArchitectEmergencyGroupIdByNameAttr getArchitectEmergencyGroupIdByNameFunc
	updateArchitectEmergencyGroupAttr      updateArchitectEmergencyGroupFunc
	deleteArchitectEmergencyGroupAttr      deleteArchitectEmergencyGroupFunc
}

func newArchitectEmergencyGroupProxy(clientConfig *platformclientv2.Configuration) *architectEmergencyGroupProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	return &architectEmergencyGroupProxy{
		clientConfig:                           clientConfig,
		architectApi:                           api,
		createArchitectEmergencyGroupAttr:      createArchitectEmergencyGroupFn,
		getAllArchitectEmergencyGroupAttr:      getAllArchitectEmergencyGroupFn,
		getArchitectEmergencyGroupAttr:         getArchitectEmergencyGroupFn,
		updateArchitectEmergencyGroupAttr:      updateArchitectEmergencyGroupFn,
		getArchitectEmergencyGroupIdByNameAttr: getArchitectEmergencyGroupIdByNameFn,
		deleteArchitectEmergencyGroupAttr:      deleteArchitectEmergencyGroupFn,
	}
}

func getArchitectEmergencyGroupProxy(clientConfig *platformclientv2.Configuration) *architectEmergencyGroupProxy {
	if internalProxy == nil {
		internalProxy = newArchitectEmergencyGroupProxy(clientConfig)
	}

	return internalProxy
}

func (p *architectEmergencyGroupProxy) getAllArchitectEmergencyGroups(ctx context.Context) (*[]platformclientv2.Emergencygroup, *platformclientv2.APIResponse, error) {
	return p.getAllArchitectEmergencyGroupAttr(ctx, p)
}

func (p *architectEmergencyGroupProxy) getArchitectEmergencyGroup(ctx context.Context, emergencyGroupId string) (*platformclientv2.Emergencygroup, *platformclientv2.APIResponse, error) {
	return p.getArchitectEmergencyGroupAttr(ctx, p, emergencyGroupId)
}

func (p *architectEmergencyGroupProxy) updateArchitectEmergencyGroup(ctx context.Context, emergencyGroupId string, emergencyGroup platformclientv2.Emergencygroup) (*platformclientv2.Emergencygroup, *platformclientv2.APIResponse, error) {
	return p.updateArchitectEmergencyGroupAttr(ctx, p, emergencyGroupId, emergencyGroup)
}

func (p *architectEmergencyGroupProxy) getArchitectEmergencyGroupIdByName(ctx context.Context, name string) (emergencyGroup *platformclientv2.Emergencygrouplisting, apiResponse *platformclientv2.APIResponse, err error) {
	return p.getArchitectEmergencyGroupIdByNameAttr(ctx, p, name)
}

func (p *architectEmergencyGroupProxy) deleteArchitectEmergencyGroup(ctx context.Context, emergencyGroupId string) (*platformclientv2.APIResponse, error) {
	return p.deleteArchitectEmergencyGroupAttr(ctx, p, emergencyGroupId)
}

func (p *architectEmergencyGroupProxy) createArchitectEmergencyGroup(ctx context.Context, emergencyGroup platformclientv2.Emergencygroup) (*platformclientv2.Emergencygroup, *platformclientv2.APIResponse, error) {
	return p.createArchitectEmergencyGroupAttr(ctx, p, emergencyGroup)
}

func getAllArchitectEmergencyGroupFn(ctx context.Context, p *architectEmergencyGroupProxy) (*[]platformclientv2.Emergencygroup, *platformclientv2.APIResponse, error) {
	var totalRecords []platformclientv2.Emergencygroup

	const pageSize = 100

	emergencyGroupConfigs, resp, getErr := p.architectApi.GetArchitectEmergencygroups(1, pageSize, "", "", "")

	if getErr != nil {
		return nil, resp, fmt.Errorf("Failed to get page of emergency group configs: %v", getErr)
	}

	if emergencyGroupConfigs.Entities == nil || len(*emergencyGroupConfigs.Entities) == 0 {
		return &totalRecords, nil, nil
	}

	totalRecords = append(totalRecords, *emergencyGroupConfigs.Entities...)

	for pageNum := 2; pageNum <= *emergencyGroupConfigs.PageCount; pageNum++ {
		emergencyGroupConfigs, resp, getErr := p.architectApi.GetArchitectEmergencygroups(pageNum, pageSize, "", "", "")

		if getErr != nil {
			return nil, resp, fmt.Errorf("Failed to get page of emergency group configs: %v", getErr)
		}

		if emergencyGroupConfigs.Entities == nil || len(*emergencyGroupConfigs.Entities) == 0 {
			break
		}

		totalRecords = append(totalRecords, *emergencyGroupConfigs.Entities...)
	}

	return &totalRecords, nil, nil
}

func getArchitectEmergencyGroupFn(ctx context.Context, p *architectEmergencyGroupProxy, emergencyGroupId string) (emergencyGroup *platformclientv2.Emergencygroup, apiResponse *platformclientv2.APIResponse, err error) {
	return p.architectApi.GetArchitectEmergencygroup(emergencyGroupId)
}

func getArchitectEmergencyGroupIdByNameFn(ctx context.Context, p *architectEmergencyGroupProxy, name string) (emergencyGroup *platformclientv2.Emergencygrouplisting, apiResponse *platformclientv2.APIResponse, err error) {
	const pageNum = 1
	const pageSize = 100

	return p.architectApi.GetArchitectEmergencygroups(pageNum, pageSize, "", "", name)
}

func updateArchitectEmergencyGroupFn(ctx context.Context, p *architectEmergencyGroupProxy, emergencyGroupId string, emergencyGroup platformclientv2.Emergencygroup) (*platformclientv2.Emergencygroup, *platformclientv2.APIResponse, error) {
	return p.architectApi.PutArchitectEmergencygroup(emergencyGroupId, emergencyGroup)
}

func deleteArchitectEmergencyGroupFn(ctx context.Context, p *architectEmergencyGroupProxy, emergencyGroupId string) (*platformclientv2.APIResponse, error) {
	return p.architectApi.DeleteArchitectEmergencygroup(emergencyGroupId)
}

func createArchitectEmergencyGroupFn(ctx context.Context, p *architectEmergencyGroupProxy, emergencyGroup platformclientv2.Emergencygroup) (*platformclientv2.Emergencygroup, *platformclientv2.APIResponse, error) {
	return p.architectApi.PostArchitectEmergencygroups(emergencyGroup)
}
