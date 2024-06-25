package employeeperformance_externalmetrics_definitions

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_employeeperformance_externalmetrics_definition_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *employeeperformanceExternalmetricsDefinitionProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createEmployeeperformanceExternalmetricsDefinitionFunc func(ctx context.Context, p *employeeperformanceExternalmetricsDefinitionProxy, domainOrganizationRole *platformclientv2.Externalmetricdefinitioncreaterequest) (*platformclientv2.Externalmetricdefinition, *platformclientv2.APIResponse, error)
type getAllEmployeeperformanceExternalmetricsDefinitionFunc func(ctx context.Context, p *employeeperformanceExternalmetricsDefinitionProxy) (*[]platformclientv2.Externalmetricdefinition, *platformclientv2.APIResponse, error)
type getEmployeeperformanceExternalmetricsDefinitionIdByNameFunc func(ctx context.Context, p *employeeperformanceExternalmetricsDefinitionProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getEmployeeperformanceExternalmetricsDefinitionByIdFunc func(ctx context.Context, p *employeeperformanceExternalmetricsDefinitionProxy, id string) (domainOrganizationRole *platformclientv2.Externalmetricdefinition, response *platformclientv2.APIResponse, err error)
type updateEmployeeperformanceExternalmetricsDefinitionFunc func(ctx context.Context, p *employeeperformanceExternalmetricsDefinitionProxy, id string, domainOrganizationRole *platformclientv2.Externalmetricdefinitionupdaterequest) (*platformclientv2.Externalmetricdefinition, *platformclientv2.APIResponse, error)
type deleteEmployeeperformanceExternalmetricsDefinitionFunc func(ctx context.Context, p *employeeperformanceExternalmetricsDefinitionProxy, id string) (response *platformclientv2.APIResponse, err error)

// employeeperformanceExternalmetricsDefinitionProxy contains all of the methods that call genesys cloud APIs.
type employeeperformanceExternalmetricsDefinitionProxy struct {
	clientConfig                                                *platformclientv2.Configuration
	gamificationApi                                             *platformclientv2.GamificationApi
	createEmployeeperformanceExternalmetricsDefinitionAttr      createEmployeeperformanceExternalmetricsDefinitionFunc
	getAllEmployeeperformanceExternalmetricsDefinitionAttr      getAllEmployeeperformanceExternalmetricsDefinitionFunc
	getEmployeeperformanceExternalmetricsDefinitionIdByNameAttr getEmployeeperformanceExternalmetricsDefinitionIdByNameFunc
	getEmployeeperformanceExternalmetricsDefinitionByIdAttr     getEmployeeperformanceExternalmetricsDefinitionByIdFunc
	updateEmployeeperformanceExternalmetricsDefinitionAttr      updateEmployeeperformanceExternalmetricsDefinitionFunc
	deleteEmployeeperformanceExternalmetricsDefinitionAttr      deleteEmployeeperformanceExternalmetricsDefinitionFunc
}

// newEmployeeperformanceExternalmetricsDefinitionProxy initializes the employeeperformance externalmetrics definition proxy with all of the data needed to communicate with Genesys Cloud
func newEmployeeperformanceExternalmetricsDefinitionProxy(clientConfig *platformclientv2.Configuration) *employeeperformanceExternalmetricsDefinitionProxy {
	api := platformclientv2.NewGamificationApiWithConfig(clientConfig)
	return &employeeperformanceExternalmetricsDefinitionProxy{
		clientConfig:    clientConfig,
		gamificationApi: api,
		createEmployeeperformanceExternalmetricsDefinitionAttr:      createEmployeeperformanceExternalmetricsDefinitionFn,
		getAllEmployeeperformanceExternalmetricsDefinitionAttr:      getAllEmployeeperformanceExternalmetricsDefinitionFn,
		getEmployeeperformanceExternalmetricsDefinitionIdByNameAttr: getEmployeeperformanceExternalmetricsDefinitionIdByNameFn,
		getEmployeeperformanceExternalmetricsDefinitionByIdAttr:     getEmployeeperformanceExternalmetricsDefinitionByIdFn,
		updateEmployeeperformanceExternalmetricsDefinitionAttr:      updateEmployeeperformanceExternalmetricsDefinitionFn,
		deleteEmployeeperformanceExternalmetricsDefinitionAttr:      deleteEmployeeperformanceExternalmetricsDefinitionFn,
	}
}

// getEmployeeperformanceExternalmetricsDefinitionProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getEmployeeperformanceExternalmetricsDefinitionProxy(clientConfig *platformclientv2.Configuration) *employeeperformanceExternalmetricsDefinitionProxy {
	if internalProxy == nil {
		internalProxy = newEmployeeperformanceExternalmetricsDefinitionProxy(clientConfig)
	}
	return internalProxy
}

// createEmployeeperformanceExternalmetricsDefinition creates a Genesys Cloud employeeperformance externalmetrics definition
func (p *employeeperformanceExternalmetricsDefinitionProxy) createEmployeeperformanceExternalmetricsDefinition(ctx context.Context, employeeperformanceExternalmetricsDefinition *platformclientv2.Externalmetricdefinitioncreaterequest) (*platformclientv2.Externalmetricdefinition, *platformclientv2.APIResponse, error) {
	return p.createEmployeeperformanceExternalmetricsDefinitionAttr(ctx, p, employeeperformanceExternalmetricsDefinition)
}

// getEmployeeperformanceExternalmetricsDefinition retrieves all Genesys Cloud employeeperformance externalmetrics definition
func (p *employeeperformanceExternalmetricsDefinitionProxy) getAllEmployeeperformanceExternalmetricsDefinition(ctx context.Context) (*[]platformclientv2.Externalmetricdefinition, *platformclientv2.APIResponse, error) {
	return p.getAllEmployeeperformanceExternalmetricsDefinitionAttr(ctx, p)
}

// getEmployeeperformanceExternalmetricsDefinitionIdByName returns a single Genesys Cloud employeeperformance externalmetrics definition by a name
func (p *employeeperformanceExternalmetricsDefinitionProxy) getEmployeeperformanceExternalmetricsDefinitionIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getEmployeeperformanceExternalmetricsDefinitionIdByNameAttr(ctx, p, name)
}

// getEmployeeperformanceExternalmetricsDefinitionById returns a single Genesys Cloud employeeperformance externalmetrics definition by Id
func (p *employeeperformanceExternalmetricsDefinitionProxy) getEmployeeperformanceExternalmetricsDefinitionById(ctx context.Context, id string) (employeeperformanceExternalmetricsDefinition *platformclientv2.Externalmetricdefinition, response *platformclientv2.APIResponse, err error) {
	return p.getEmployeeperformanceExternalmetricsDefinitionByIdAttr(ctx, p, id)
}

// updateEmployeeperformanceExternalmetricsDefinition updates a Genesys Cloud employeeperformance externalmetrics definition
func (p *employeeperformanceExternalmetricsDefinitionProxy) updateEmployeeperformanceExternalmetricsDefinition(ctx context.Context, id string, employeeperformanceExternalmetricsDefinition *platformclientv2.Externalmetricdefinitionupdaterequest) (*platformclientv2.Externalmetricdefinition, *platformclientv2.APIResponse, error) {
	return p.updateEmployeeperformanceExternalmetricsDefinitionAttr(ctx, p, id, employeeperformanceExternalmetricsDefinition)
}

// deleteEmployeeperformanceExternalmetricsDefinition deletes a Genesys Cloud employeeperformance externalmetrics definition by Id
func (p *employeeperformanceExternalmetricsDefinitionProxy) deleteEmployeeperformanceExternalmetricsDefinition(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteEmployeeperformanceExternalmetricsDefinitionAttr(ctx, p, id)
}

// createEmployeeperformanceExternalmetricsDefinitionFn is an implementation function for creating a Genesys Cloud employeeperformance externalmetrics definition
func createEmployeeperformanceExternalmetricsDefinitionFn(ctx context.Context, p *employeeperformanceExternalmetricsDefinitionProxy, employeeperformanceExternalmetricsDefinition *platformclientv2.Externalmetricdefinitioncreaterequest) (*platformclientv2.Externalmetricdefinition, *platformclientv2.APIResponse, error) {
	definition, resp, err := p.gamificationApi.PostEmployeeperformanceExternalmetricsDefinitions(*employeeperformanceExternalmetricsDefinition)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to create employeeperformance externalmetrics definition: %s", err)
	}
	return definition, resp, nil
}

// getAllEmployeeperformanceExternalmetricsDefinitionFn is the implementation for retrieving all employeeperformance externalmetrics definition in Genesys Cloud
func getAllEmployeeperformanceExternalmetricsDefinitionFn(ctx context.Context, p *employeeperformanceExternalmetricsDefinitionProxy) (*[]platformclientv2.Externalmetricdefinition, *platformclientv2.APIResponse, error) {
	var allDefinitions []platformclientv2.Externalmetricdefinition
	const pageSize = 100

	definitions, resp, err := p.gamificationApi.GetEmployeeperformanceExternalmetricsDefinitions(pageSize, 1)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get domain organization role: %v", err)
	}
	if definitions.Entities == nil || len(*definitions.Entities) == 0 {
		return &allDefinitions, resp, nil
	}
	for _, definition := range *definitions.Entities {
		allDefinitions = append(allDefinitions, definition)
	}

	for pageNum := 2; pageNum <= *definitions.PageCount; pageNum++ {
		definitions, resp, err := p.gamificationApi.GetEmployeeperformanceExternalmetricsDefinitions(pageSize, pageNum)
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get domain organization role: %v", err)
		}

		if definitions.Entities == nil || len(*definitions.Entities) == 0 {
			break
		}

		for _, definition := range *definitions.Entities {
			allDefinitions = append(allDefinitions, definition)
		}
	}

	return &allDefinitions, resp, nil
}

// getEmployeeperformanceExternalmetricsDefinitionIdByNameFn is an implementation of the function to get a Genesys Cloud employeeperformance externalmetrics definition by name
func getEmployeeperformanceExternalmetricsDefinitionIdByNameFn(ctx context.Context, p *employeeperformanceExternalmetricsDefinitionProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	definitions, resp, err := getAllEmployeeperformanceExternalmetricsDefinitionFn(ctx, p)
	if err != nil {
		return "", false, resp, err
	}

	if definitions == nil || len(*definitions) == 0 {
		return "", true, resp, fmt.Errorf("No employeeperformance externalmetrics definition found with name %s", name)
	}

	for _, definition := range *definitions {
		if *definition.Name == name {
			log.Printf("Retrieved the employeeperformance externalmetrics definition id %s by name %s", *definition.Id, name)
			return *definition.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("Unable to find employeeperformance externalmetrics definition with name %s", name)
}

// getEmployeeperformanceExternalmetricsDefinitionByIdFn is an implementation of the function to get a Genesys Cloud employeeperformance externalmetrics definition by Id
func getEmployeeperformanceExternalmetricsDefinitionByIdFn(ctx context.Context, p *employeeperformanceExternalmetricsDefinitionProxy, id string) (employeeperformanceExternalmetricsDefinition *platformclientv2.Externalmetricdefinition, response *platformclientv2.APIResponse, err error) {
	definition, resp, err := p.gamificationApi.GetEmployeeperformanceExternalmetricsDefinition(id)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to retrieve employeeperformance externalmetrics definition by id %s: %s", id, err)
	}
	return definition, resp, nil
}

// updateEmployeeperformanceExternalmetricsDefinitionFn is an implementation of the function to update a Genesys Cloud employeeperformance externalmetrics definition
func updateEmployeeperformanceExternalmetricsDefinitionFn(ctx context.Context, p *employeeperformanceExternalmetricsDefinitionProxy, id string, employeeperformanceExternalmetricsDefinition *platformclientv2.Externalmetricdefinitionupdaterequest) (*platformclientv2.Externalmetricdefinition, *platformclientv2.APIResponse, error) {
	definition, resp, err := p.gamificationApi.PatchEmployeeperformanceExternalmetricsDefinition(id, *employeeperformanceExternalmetricsDefinition)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to update employeeperformance externalmetrics definition: %s", err)
	}
	return definition, resp, nil
}

// deleteEmployeeperformanceExternalmetricsDefinitionFn is an implementation function for deleting a Genesys Cloud employeeperformance externalmetrics definition
func deleteEmployeeperformanceExternalmetricsDefinitionFn(ctx context.Context, p *employeeperformanceExternalmetricsDefinitionProxy, id string) (response *platformclientv2.APIResponse, err error) {
	resp, err := p.gamificationApi.DeleteEmployeeperformanceExternalmetricsDefinition(id)
	if err != nil {
		return resp, fmt.Errorf("Failed to delete employeeperformance externalmetrics definition: %s", err)
	}
	return resp, nil
}
