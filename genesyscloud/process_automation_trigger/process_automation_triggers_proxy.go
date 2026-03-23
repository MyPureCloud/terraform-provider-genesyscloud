package process_automation_trigger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	customapi "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/custom_api_client"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

func postProcessAutomationTrigger(pat *ProcessAutomationTrigger, api *platformclientv2.IntegrationsApi) (*ProcessAutomationTrigger, *platformclientv2.APIResponse, error) {
	body, err := pat.toJSONBody()
	if err != nil {
		return nil, nil, err
	}

	c := customapi.NewClient(api.Configuration, ResourceType)
	result, resp, err := customapi.Do[ProcessAutomationTrigger](context.Background(), c, customapi.MethodPost, "/api/v2/processAutomation/triggers", body, nil)
	if err != nil {
		return nil, resp, err
	}
	log.Printf("Process automation trigger created with Id %s and correlationId: %s", *result.Id, resp.CorrelationID)
	return result, resp, nil
}

func getProcessAutomationTrigger(triggerId string, api *platformclientv2.IntegrationsApi) (*ProcessAutomationTrigger, *platformclientv2.APIResponse, error) {
	c := customapi.NewClient(api.Configuration, ResourceType)
	rawBody, resp, err := customapi.DoRaw(context.Background(), c, customapi.MethodGet, "/api/v2/processAutomation/triggers/"+triggerId, nil, nil)
	if err != nil {
		return nil, resp, err
	}

	// Custom unmarshaling needed to preserve matchCriteria as raw JSON
	apiResp := &platformclientv2.APIResponse{
		RawBody:       rawBody,
		StatusCode:    resp.StatusCode,
		CorrelationID: resp.CorrelationID,
		Response:      resp.Response,
	}
	result, err := NewProcessAutomationFromPayload(apiResp)
	if err != nil {
		return nil, resp, err
	}
	return result, resp, nil
}

func putProcessAutomationTrigger(triggerId string, pat *ProcessAutomationTrigger, api *platformclientv2.IntegrationsApi) (*ProcessAutomationTrigger, *platformclientv2.APIResponse, error) {
	body, err := pat.toJSONBody()
	if err != nil {
		return nil, nil, err
	}

	c := customapi.NewClient(api.Configuration, ResourceType)
	result, resp, err := customapi.Do[ProcessAutomationTrigger](context.Background(), c, customapi.MethodPut, "/api/v2/processAutomation/triggers/"+triggerId, body, nil)
	if err != nil {
		return nil, resp, err
	}
	log.Printf("Process automation trigger updated with Id %s and correlationId: %s", *result.Id, resp.CorrelationID)
	return result, resp, nil
}

func deleteProcessAutomationTrigger(triggerId string, api *platformclientv2.IntegrationsApi) (*platformclientv2.APIResponse, error) {
	c := customapi.NewClient(api.Configuration, ResourceType)
	return customapi.DoNoResponse(context.Background(), c, customapi.MethodDelete, "/api/v2/processAutomation/triggers/"+triggerId, nil, nil)
}

func getAllProcessAutomationTriggersResourceMap(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	resources := make(resourceExporter.ResourceIDMetaMap)

	relativePath := "/api/v2/processAutomation/triggers"

	for {
		processAutomationTriggers, resp, getErr := getAllProcessAutomationTriggers(ctx, clientConfig, relativePath)

		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get page of process automation triggers: %v", getErr), resp)
		}

		if processAutomationTriggers.Entities == nil || len(*processAutomationTriggers.Entities) == 0 {
			break
		}

		for _, trigger := range *processAutomationTriggers.Entities {
			resources[*trigger.Id] = &resourceExporter.ResourceMeta{BlockLabel: *trigger.Name}
		}

		if processAutomationTriggers.NextUri == nil {
			break
		}

		relativePath = *processAutomationTriggers.NextUri
	}

	return resources, nil
}

// toJSONBody converts a ProcessAutomationTrigger to a map suitable for CallAPI's postBody.
func (p *ProcessAutomationTrigger) toJSONBody() (map[string]interface{}, error) {
	jsonStr, err := p.toJSONString()
	if err != nil {
		return nil, err
	}
	var jsonMap map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &jsonMap)
	return jsonMap, err
}
