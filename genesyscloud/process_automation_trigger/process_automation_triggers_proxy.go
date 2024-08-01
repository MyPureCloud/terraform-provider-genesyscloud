package process_automation_trigger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func postProcessAutomationTrigger(pat *ProcessAutomationTrigger, api *platformclientv2.IntegrationsApi) (*ProcessAutomationTrigger, *platformclientv2.APIResponse, error) {
	apiClient := &api.Configuration.APIClient
	jsonStr, err := pat.toJSONString()
	if err != nil {
		return nil, nil, err
	}

	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &jsonMap)

	// create path and map variables
	path := api.Configuration.BasePath + "/api/v2/processAutomation/triggers"

	// add default headers if any
	headerParams := make(map[string]string)

	for key := range api.Configuration.DefaultHeader {
		headerParams[key] = api.Configuration.DefaultHeader[key]
	}

	headerParams["Authorization"] = "Bearer " + api.Configuration.AccessToken
	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload *ProcessAutomationTrigger
	response, err := apiClient.CallAPI(path, http.MethodPost, jsonMap, headerParams, nil, nil, "", nil)

	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {

		err = errors.New(response.ErrorMessage)
	} else {
		err = json.Unmarshal([]byte(response.RawBody), &successPayload)
		log.Printf("Process automation trigger created with Id %s and correlationId: %s", *successPayload.Id, response.CorrelationID)
	}

	return successPayload, response, err
}

func getProcessAutomationTrigger(triggerId string, api *platformclientv2.IntegrationsApi) (*ProcessAutomationTrigger, *platformclientv2.APIResponse, error) {
	apiClient := &api.Configuration.APIClient

	// create path and map variables
	path := api.Configuration.BasePath + "/api/v2/processAutomation/triggers/" + triggerId

	headerParams := make(map[string]string)

	// oauth required
	if api.Configuration.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + api.Configuration.AccessToken
	}
	// add default headers if any
	for key := range api.Configuration.DefaultHeader {
		headerParams[key] = api.Configuration.DefaultHeader[key]
	}

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	response, err := apiClient.CallAPI(path, http.MethodGet, nil, headerParams, nil, nil, "", nil)
	if response.Error != nil {
		err = errors.New(response.ErrorMessage)
		return nil, nil, err
	}

	successPayload, err := NewProcessAutomationFromPayload(response)
	if err != nil {
		return nil, response, err
	}

	return successPayload, response, err
}

func putProcessAutomationTrigger(triggerId string, pat *ProcessAutomationTrigger, api *platformclientv2.IntegrationsApi) (*ProcessAutomationTrigger, *platformclientv2.APIResponse, error) {
	apiClient := &api.Configuration.APIClient
	jsonStr, err := pat.toJSONString()
	if err != nil {
		return nil, nil, err
	}

	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &jsonMap)

	// create path and map variables
	path := api.Configuration.BasePath + "/api/v2/processAutomation/triggers/" + triggerId
	headerParams := make(map[string]string)

	// oauth required
	if api.Configuration.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + api.Configuration.AccessToken
	}
	// add default headers if any
	for key := range api.Configuration.DefaultHeader {
		headerParams[key] = api.Configuration.DefaultHeader[key]
	}

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload *ProcessAutomationTrigger
	response, err := apiClient.CallAPI(path, http.MethodPut, jsonMap, headerParams, nil, nil, "", nil)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		err = json.Unmarshal([]byte(response.RawBody), &successPayload)
		log.Printf("Process automation trigger updated with Id %s and correlationId: %s", *successPayload.Id, response.CorrelationID)
	}
	return successPayload, response, err
}

func deleteProcessAutomationTrigger(triggerId string, api *platformclientv2.IntegrationsApi) (*platformclientv2.APIResponse, error) {
	apiClient := &api.Configuration.APIClient

	// create path and map variables
	path := api.Configuration.BasePath + "/api/v2/processAutomation/triggers/" + triggerId

	headerParams := make(map[string]string)

	// oauth required
	if api.Configuration.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + api.Configuration.AccessToken
	}
	// add default headers if any
	for key := range api.Configuration.DefaultHeader {
		headerParams[key] = api.Configuration.DefaultHeader[key]
	}

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	response, err := apiClient.CallAPI(path, http.MethodDelete, nil, headerParams, nil, nil, "", nil)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	}

	return response, err
}

func getAllProcessAutomationTriggersResourceMap(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)

	// create path and map variables
	path := integAPI.Configuration.BasePath + "/api/v2/processAutomation/triggers"

	for pageNum := 1; ; pageNum++ {
		processAutomationTriggers, resp, getErr := getAllProcessAutomationTriggers(path, integAPI)

		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("failed to get page of process automation triggers: %v", getErr), resp)
		}

		if processAutomationTriggers.Entities == nil || len(*processAutomationTriggers.Entities) == 0 {
			break
		}

		for _, trigger := range *processAutomationTriggers.Entities {
			resources[*trigger.Id] = &resourceExporter.ResourceMeta{Name: *trigger.Name}
		}

		if processAutomationTriggers.NextUri == nil {
			break
		}

		path = integAPI.Configuration.BasePath + *processAutomationTriggers.NextUri
	}

	return resources, nil
}
