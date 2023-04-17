package genesyscloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v96/platformclientv2"
)

type ProcessAutomationTrigger struct {
	Id              *string          `json:"id,omitempty"`
	TopicName       *string          `json:"topicName,omitempty"`
	Name            *string          `json:"name,omitempty"`
	Target          *Target          `json:"target,omitempty"`
	MatchCriteria   *[]MatchCriteria `json:"matchCriteria,omitempty"`
	Enabled         *bool            `json:"enabled,omitempty"`
	EventTTLSeconds *int             `json:"eventTTLSeconds,omitempty"`
	DelayBySeconds  *int             `json:"delayBySeconds,omitempty"`
	Version         *int             `json:"version,omitempty"`
	Description     *string          `json:"description,omitempty"`
}

type UpdateTriggerInput struct {
	TopicName       *string          `json:"topicName,omitempty"`
	Name            *string          `json:"name,omitempty"`
	Target          *Target          `json:"target,omitempty"`
	MatchCriteria   *[]MatchCriteria `json:"matchCriteria,omitempty"`
	Enabled         *bool            `json:"enabled,omitempty"`
	EventTTLSeconds *int             `json:"eventTTLSeconds,omitempty"`
	DelayBySeconds  *int             `json:"delayBySeconds,omitempty"`
	Version         *int             `json:"version,omitempty"`
	Description     *string          `json:"description,omitempty"`
}

type MatchCriteria struct {
	JsonPath *string   `json:"jsonPath,omitempty"`
	Operator *string   `json:"operator,omitempty"`
	Value    *string   `json:"value,omitempty"`
	Values   *[]string `json:"values,omitempty"`
}

type Target struct {
	Type *string `json:"type,omitempty"`
	Id   *string `json:"id,omitempty"`
}

var (
	target = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description: "Type of the target the trigger is configured to hit",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"Workflow",
				}, false),
			},
			"id": {
				Description: "Id of the target the trigger is configured to hit",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}

	matchCriteria = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"json_path": {
				Description: "The json path of the topic event to be compared to match criteria value",
				Type:        schema.TypeString,
				Required:    true,
			},
			"operator": {
				Description: "The operator used to compare the json path against the value of the match criteria",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"GreaterThanOrEqual",
					"LessThanOrEqual",
					"Equal",
					"NotEqual",
					"LessThan",
					"GreaterThan",
					"NotIn",
					"In",
					"Contains",
					"All",
					"Exists",
					"Size",
				}, false),
			},
			"value": {
				Description: "Value the jsonPath is compared against",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"values": {
				Description: "Values the jsonPath are compared against",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
)

/*
NOTE:
This resource currently does not use the Go SDk and instead makes API calls directly.
The Go SDK can not properly handle process automation triggers due the value and values
attributes in the matchCriteria object being listed as JsonNode in the swagger docs.
A JsonNode is a placeholder type with no nested values which creates problems in Go
because it can't properly determine a type for the value/values field.
*/
func resourceProcessAutomationTrigger() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Process Automation Trigger

**NOTE: This component is currently in beta. If you wish to use this provider make sure your client has the correct permissions**`,

		CreateContext: createWithPooledClient(createProcessAutomationTrigger),
		ReadContext:   readWithPooledClient(readProcessAutomationTrigger),
		UpdateContext: updateWithPooledClient(updateProcessAutomationTrigger),
		DeleteContext: deleteWithPooledClient(removeProcessAutomationTrigger),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "Name of the Trigger",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"topic_name": {
				Description:  "Topic name that will fire trigger. Changing the topic_name attribute will cause the processautomation_trigger object to be dropped and recreated with a new ID. ",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"enabled": {
				Description: "Whether or not the trigger should be fired on events",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"target": {
				Description: "Target the trigger will invoke when fired",
				Type:        schema.TypeSet,
				Optional:    false,
				Required:    true,
				MaxItems:    1,
				Elem:        target,
			},
			"match_criteria": {
				Description: "Match criteria that controls when the trigger will fire.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        matchCriteria,
			},
			"event_ttl_seconds": {
				Description:  "How old an event can be to fire the trigger. Must be an number greater than or equal to 10. Only one of event_ttl_seconds or delay_by_seconds can be set.",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(10),
			},
			"delay_by_seconds": {
				Description:  "How long to delay processing of a trigger after an event passes the match criteria. Must be an number between 60 and 900 inclusive. Only one of event_ttl_seconds or delay_by_seconds can be set.",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(60, 900),
			},
			"description": {
				Description:  "A description of the trigger",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
		},
	}
}

func processAutomationTriggerExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllProcessAutomationTriggersResourceMap),
		RefAttrs: map[string]*RefAttrSettings{
			"target.id": {RefType: "genesyscloud_flow"},
		},
	}
}

func createProcessAutomationTrigger(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	topic_name := d.Get("topic_name").(string)
	enabled := d.Get("enabled").(bool)
	eventTTLSeconds := d.Get("event_ttl_seconds").(int)
	delayBySeconds := d.Get("delay_by_seconds").(int)
	description := d.Get("description").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	if eventTTLSeconds > 0 && delayBySeconds > 0 {
		return diag.Errorf("Only one of event_ttl_seconds or delay_by_seconds can be set.")
	}

	log.Printf("Creating process automation trigger %s", name)

	triggerInput := &ProcessAutomationTrigger{
		TopicName:     &topic_name,
		Name:          &name,
		Target:        buildTarget(d),
		MatchCriteria: buildMatchCriteria(d),
		Enabled:       &enabled,
		Description:   &description,
	}

	if eventTTLSeconds > 0 {
		triggerInput.EventTTLSeconds = &eventTTLSeconds
	}

	if delayBySeconds > 0 {
		triggerInput.DelayBySeconds = &delayBySeconds
	}

	diagErr := retryWhen(isStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		trigger, resp, err := postProcessAutomationTrigger(triggerInput, integAPI)

		if err != nil {
			return resp, diag.Errorf("Failed to create process automation trigger %s: %s", name, err)
		}
		d.SetId(*trigger.Id)

		log.Printf("Created process automation trigger %s %s", name, *trigger.Id)
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return readProcessAutomationTrigger(ctx, d, meta)
}

func readProcessAutomationTrigger(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Reading process automation trigger %s", d.Id())

	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		trigger, resp, getErr := getProcessAutomationTrigger(d.Id(), integAPI)
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read process automation trigger %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to process read automation trigger %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceProcessAutomationTrigger())

		if trigger.Name != nil {
			d.Set("name", *trigger.Name)
		} else {
			d.Set("name", nil)
		}

		if trigger.TopicName != nil {
			d.Set("topic_name", *trigger.TopicName)
		} else {
			d.Set("topic_name", nil)
		}

		d.Set("match_criteria", flattenMatchCriteria(trigger.MatchCriteria))
		d.Set("target", flattenTarget(trigger.Target))

		if trigger.Enabled != nil {
			d.Set("enabled", *trigger.Enabled)
		} else {
			d.Set("enabled", nil)
		}

		if trigger.EventTTLSeconds != nil {
			d.Set("event_ttl_seconds", *trigger.EventTTLSeconds)
		} else {
			d.Set("event_ttl_seconds", nil)
		}

		if trigger.DelayBySeconds != nil {
			d.Set("delay_by_seconds", *trigger.DelayBySeconds)
		} else {
			d.Set("delay_by_seconds", nil)
		}

		if trigger.Description != nil {
			d.Set("description", *trigger.Description)
		} else {
			d.Set("description", nil)
		}

		log.Printf("Read process automation trigger %s %s", d.Id(), *trigger.Name)
		return cc.CheckState()
	})
}

func updateProcessAutomationTrigger(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	enabled := d.Get("enabled").(bool)
	eventTTLSeconds := d.Get("event_ttl_seconds").(int)
	delayBySeconds := d.Get("delay_by_seconds").(int)
	description := d.Get("description").(string)

	topic_name := d.Get("topic_name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Updating process automation trigger %s", name)

	diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get the latest trigger version to send with PATCH
		trigger, resp, getErr := getProcessAutomationTrigger(d.Id(), integAPI)
		if getErr != nil {
			return resp, diag.Errorf("Failed to read process automation trigger %s: %s", d.Id(), getErr)
		}

		if eventTTLSeconds > 0 && delayBySeconds > 0 {
			return resp, diag.Errorf("Only one of event_ttl_seconds or delay_by_seconds can be set.")
		}

		triggerInput := &UpdateTriggerInput{
			TopicName:     &topic_name,
			Name:          &name,
			Enabled:       &enabled,
			Target:        buildTarget(d),
			MatchCriteria: buildMatchCriteria(d),
			Version:       trigger.Version,
			Description:   &description,
		}

		if eventTTLSeconds > 0 {
			triggerInput.EventTTLSeconds = &eventTTLSeconds
		}

		if delayBySeconds > 0 {
			triggerInput.DelayBySeconds = &delayBySeconds
		}

		_, putResp, err := putProcessAutomationTrigger(d.Id(), triggerInput, integAPI)

		if err != nil {
			return putResp, diag.Errorf("Failed to update process automation trigger %s: %s", name, err)
		}
		return putResp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated process automation trigger %s", name)
	return readProcessAutomationTrigger(ctx, d, meta)
}

func removeProcessAutomationTrigger(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Deleting process automation trigger %s", name)
	resp, err := deleteProcessAutomationTrigger(d.Id(), integAPI)
	if err != nil {
		if isStatus404(resp) {
			log.Printf("process automation trigger already deleted %s", d.Id())
			return nil
		}
		return diag.Errorf("Failed to delete process automation trigger %s: %s", d.Id(), err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := getProcessAutomationTrigger(d.Id(), integAPI)
		if err != nil {
			if isStatus404(resp) {
				// process Automation trigger deleted
				log.Printf("Deleted process Automation trigger %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting process automation trigger %s: %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("process automation trigger %s still exists", d.Id()))
	})
}

func buildTarget(d *schema.ResourceData) *Target {
	if target := d.Get("target"); target != nil {
		if targetList := target.(*schema.Set).List(); len(targetList) > 0 {
			targetMap := targetList[0].(map[string]interface{})

			targetType := targetMap["type"].(string)
			id := targetMap["id"].(string)

			return &Target{
				Type: &targetType,
				Id:   &id,
			}
		}
	}

	return &Target{}
}

func buildMatchCriteria(d *schema.ResourceData) *[]MatchCriteria {
	if matchCriteriaVal := d.Get("match_criteria"); matchCriteriaVal != nil {
		if matchCriteriaList := matchCriteriaVal.(*schema.Set).List(); len(matchCriteriaList) > 0 {

			var matchCriteriaObjectList []MatchCriteria = []MatchCriteria{}

			for item := 0; item < len(matchCriteriaList); item++ {
				matchCriteriaMap := matchCriteriaList[item].(map[string]interface{})

				jsonPath := matchCriteriaMap["json_path"].(string)
				operator := matchCriteriaMap["operator"].(string)
				value := matchCriteriaMap["value"].(string)
				valuesInt := matchCriteriaMap["values"].([]interface{})

				values := make([]string, len(valuesInt))
				for i, v := range valuesInt {
					values[i] = fmt.Sprint(v)
				}

				if len(values) < 1 {
					criteria := MatchCriteria{
						JsonPath: &jsonPath,
						Operator: &operator,
						Value:    &value,
					}

					matchCriteriaObjectList = append(matchCriteriaObjectList, criteria)
				} else {
					criteria := MatchCriteria{
						JsonPath: &jsonPath,
						Operator: &operator,
						Values:   &values,
					}

					matchCriteriaObjectList = append(matchCriteriaObjectList, criteria)
				}

			}

			return &matchCriteriaObjectList
		}
	}
	return &[]MatchCriteria{}
}

func flattenTarget(inputTarget *Target) *schema.Set {
	if inputTarget == nil {
		return nil
	}

	targetSet := schema.NewSet(schema.HashResource(target), []interface{}{})

	flattendedTarget := make(map[string]interface{})
	flattendedTarget["id"] = *inputTarget.Id
	flattendedTarget["type"] = *inputTarget.Type
	targetSet.Add(flattendedTarget)

	return targetSet
}

func flattenMatchCriteria(inputMatchCriteria *[]MatchCriteria) *schema.Set {
	if inputMatchCriteria == nil {
		return nil
	}

	matchCriteriaSet := schema.NewSet(schema.HashResource(matchCriteria), []interface{}{})
	for _, sdkMatchCriteria := range *inputMatchCriteria {
		flattendedMatchCriteria := make(map[string]interface{})
		flattendedMatchCriteria["json_path"] = *sdkMatchCriteria.JsonPath
		flattendedMatchCriteria["operator"] = *sdkMatchCriteria.Operator

		if sdkMatchCriteria.Value != nil {
			flattendedMatchCriteria["value"] = *sdkMatchCriteria.Value
		}

		if sdkMatchCriteria.Values != nil {

			t := *sdkMatchCriteria.Values
			s := make([]interface{}, len(t))
			for i, v := range t {
				s[i] = v
			}
			flattendedMatchCriteria["values"] = s
		}
		matchCriteriaSet.Add(flattendedMatchCriteria)
	}
	return matchCriteriaSet
}

func postProcessAutomationTrigger(body *ProcessAutomationTrigger, api *platformclientv2.IntegrationsApi) (*ProcessAutomationTrigger, *platformclientv2.APIResponse, error) {
	apiClient := &api.Configuration.APIClient

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
	response, err := apiClient.CallAPI(path, http.MethodPost, body, headerParams, nil, nil, "", nil)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		err = json.Unmarshal([]byte(response.RawBody), &successPayload)
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

	var successPayload *ProcessAutomationTrigger
	response, err := apiClient.CallAPI(path, http.MethodGet, nil, headerParams, nil, nil, "", nil)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		err = json.Unmarshal([]byte(response.RawBody), &successPayload)
	}
	return successPayload, response, err
}

func putProcessAutomationTrigger(triggerId string, updateInput *UpdateTriggerInput, api *platformclientv2.IntegrationsApi) (*ProcessAutomationTrigger, *platformclientv2.APIResponse, error) {
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

	var successPayload *ProcessAutomationTrigger
	response, err := apiClient.CallAPI(path, http.MethodPut, updateInput, headerParams, nil, nil, "", nil)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		err = json.Unmarshal([]byte(response.RawBody), &successPayload)
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

func getAllProcessAutomationTriggersResourceMap(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)

	// create path and map variables
	path := integAPI.Configuration.BasePath + "/api/v2/processAutomation/triggers"

	for pageNum := 1; ; pageNum++ {
		processAutomationTriggers, _, getErr := getAllProcessAutomationTriggers(path, integAPI)

		if getErr != nil {
			return nil, diag.Errorf("failed to get page of process automation triggers: %v", getErr)
		}

		if processAutomationTriggers.Entities == nil || len(*processAutomationTriggers.Entities) == 0 {
			break
		}

		for _, trigger := range *processAutomationTriggers.Entities {
			resources[*trigger.Id] = &ResourceMeta{Name: *trigger.Name}
		}

		if processAutomationTriggers.NextUri == nil {
			break
		}

		path = integAPI.Configuration.BasePath + *processAutomationTriggers.NextUri
	}

	return resources, nil
}
