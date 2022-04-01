package genesyscloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
)

func resourceProcessAutomationTrigger() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Process Automation Trigger.",

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
				Description:  "Topic name that will fire trigger",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"enabled": {
				Description: "Whether or not the trigger should be fired on events",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"target_id": {
				Description: "Id of the target to invoke when the trigger is fired",
				Type:        schema.TypeString,
				Required:    true,
			},
			"target_type": {
				Description:      "Type of target to invoke when the trigger fires",
				Type:             schema.TypeString,
				Required:         true,
			},
            "match_criteria": {
                Description:      "JSON schema that defines the match criteria of the trigger.",
                Type:             schema.TypeString,
                Required:         false,
                Optional:         true,
                DiffSuppressFunc: suppressEquivalentJsonDiffs,
            },
			"event_ttl_seconds": {
                Description:      "How old an event can be to fire the trigger",
                Type:             schema.TypeInt,
                Optional:         true,
				Required:         false,
            },
		},
	}
}

func createProcessAutomationTrigger(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	topic_name := d.Get("topic_name").(string)
	enabled := d.Get("enabled").(bool)
	eventTTLSeconds := d.Get("event_ttl_seconds").(int)

    //make target of trigger
	target := buildTriggerTarget(d)

	//make match criteria for the trigger
	matchCriteria := d.Get("match_criteria").(string)
    matchCriteriaVal, err := jsonStringToInterface(matchCriteria)
    if err != nil {
        return diag.Errorf("Failed to parse contract input %s: %v", matchCriteria, err)
    }

	sdkConfig := meta.(*providerMeta).ClientConfig
    integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Creating process automation trigger %s", name)

	diagErr := retryWhen(isStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		trigger, resp, err := postProcessAutomationTrigger(&ProcessAutomationTrigger{
            TopicName:       &topic_name,
            Name:            &name,
            Target:          target,
            MatchCriteria:   &matchCriteriaVal,
            Enabled:         &enabled,
            EventTTLSeconds: &eventTTLSeconds,
		}, integAPI)
		if err != nil {
			return resp, diag.Errorf("Failed to create process automation trigger %s: %s: topicName:%s, Name:%s, Target:%+v, MatchCriteria:%v, Enabled:%t, EventTTLSeconds:%d", name, err, topic_name, name, target, matchCriteriaVal, enabled, eventTTLSeconds)
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
	sdkConfig := meta.(*providerMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Reading process automation trigger %s", d.Id())

	return withRetriesForRead(ctx, 30*time.Second, d, func() *resource.RetryError {
	    trigger, resp, getErr := getProcessAutomationTrigger(d.Id(), integAPI)
	    if getErr != nil {
            if isStatus404(resp) {
                return resource.RetryableError(fmt.Errorf("Failed to read integration action %s: %s", d.Id(), getErr))
            }
            return resource.NonRetryableError(fmt.Errorf("Failed to read integration action %s: %s", d.Id(), getErr))
        }

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

        if trigger.Target != nil {
            d.Set("target_id", *trigger.Target.Id)
        } else {
            d.Set("target_id", nil)
        }

        if trigger.Target != nil {
            d.Set("target_type", *trigger.Target.Type)
        } else {
            d.Set("target_type", nil)
        }

        if trigger.MatchCriteria != nil {
            input, err := flattenActionContract(*trigger.MatchCriteria)
            if err != nil {
                return resource.NonRetryableError(fmt.Errorf("%v", err))
            }
            d.Set("match_criteria", input)
        } else {
            d.Set("match_criteria", nil)
        }

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

        log.Printf("Read process automation trigger %s %s", d.Id(), *trigger.Name)
        return nil
	})
}

func updateProcessAutomationTrigger(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    name := d.Get("name").(string)
	enabled := d.Get("enabled").(bool)
	eventTTLSeconds := d.Get("event_ttl_seconds").(int)

    //make target of trigger
	target := buildTriggerTarget(d)

	//make match criteria for the trigger
    matchCriteria := d.Get("match_criteria").(string)
    matchCriteriaVal, err := jsonStringToInterface(matchCriteria)
    if err != nil {
        return diag.Errorf("Failed to parse match criteria %s: %v", matchCriteria, err)
    }

	sdkConfig := meta.(*providerMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Updating process automation trigger %s", name)

	diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get the latest trigger version to send with PATCH
		trigger, resp, getErr := getProcessAutomationTrigger(d.Id(), integAPI)
		if getErr != nil {
			return resp, diag.Errorf("Failed to read process automation trigger %s: %s", d.Id(), getErr)
		}

		_, _, err := putProcessAutomationTrigger(d.Id(), &UpdateTriggerInput{
			Name:               &name,
			Enabled:            &enabled,
			EventTTLSeconds:    &eventTTLSeconds,
            Target:             target,
			MatchCriteria:      &matchCriteriaVal,
			Version:            trigger.Version,
		}, integAPI)
		if err != nil {
			return resp, diag.Errorf("Failed to update process automation trigger %s: %s", name, err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated process automation trigger %s", name)
	time.Sleep(5 * time.Second)
	return readProcessAutomationTrigger(ctx, d, meta)
}

func removeProcessAutomationTrigger(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    name := d.Get("name").(string)

    sdkConfig := meta.(*providerMeta).ClientConfig
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
                // Integration action deleted
                log.Printf("Deleted Integration action %s", d.Id())
                return nil
            }
            return resource.NonRetryableError(fmt.Errorf("Error deleting process automation trigger %s: %s", d.Id(), err))
        }
        return resource.RetryableError(fmt.Errorf("process automation trigger %s still exists", d.Id()))
    })
}

func buildTriggerTarget(d *schema.ResourceData) (*Target){
	target_id := d.Get("target_id").(string)
	target_type := d.Get("target_type").(string)

    return &Target{
    		Type:   &target_type,
    		Id:     &target_id,
    	}
}

func flattenMatchCriteria(schema interface{}) (string, diag.Diagnostics) {
	if schema == nil {
		return "", nil
	}
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return "", diag.Errorf("Error marshalling match criteria %v: %v", schema, err)
	}
	return string(schemaBytes), nil
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

type ProcessAutomationTrigger struct {
	Id              *string             `json:"id,omitempty"`
	TopicName       *string             `json:"topicName,omitempty"`
	Name            *string             `json:"name,omitempty"`
	Target          *Target             `json:"target,omitempty"`
	MatchCriteria   *interface{}    `json:"matchCriteria,omitempty"`
	Enabled         *bool               `json:"enabled,omitempty"`
	EventTTLSeconds *int                `json:"eventTTLSeconds,omitempty"`
	Version         *int                `json:"version,omitempty"`
}

type UpdateTriggerInput struct {
	Name            *string             `json:"name,omitempty"`
	Target          *Target             `json:"target,omitempty"`
	MatchCriteria   *interface{}        `json:"matchCriteria,omitempty"`
	Enabled         *bool               `json:"enabled,omitempty"`
	EventTTLSeconds *int                `json:"eventTTLSeconds,omitempty"`
	Version         *int                `json:"version,omitempty"`
}

type Target struct {
	Type    *string `json:"type,omitempty"`
	Id      *string  `json:"id,omitempty"`
}

type MatchCriteria struct {
	JsonPath    *string     `json:"jsonPath,omitempty"`
	Operator    *string     `json:"operator,omitempty"`
	Value       *string     `json:"value,omitempty"`
}