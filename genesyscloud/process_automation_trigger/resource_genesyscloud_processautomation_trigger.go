package process_automation_trigger

import (
	"context"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"fmt"
	"log"

	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	ResourceType = "genesyscloud_processautomation_trigger"
)

var (
	workflowTargetSettings = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"data_format": {
				Description: "The data format to use when invoking target.",
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"Json",
					"TopLevelPrimitives",
				}, false),
			},
		},
	}
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
			"workflow_target_settings": {
				Description: "Optional config for the target. Until the feature gets enabled will always operate in TopLevelPrimitives mode.",
				Type:        schema.TypeSet,
				Required:    false,
				Optional:    true,
				MaxItems:    1,
				Elem:        workflowTargetSettings,
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
func ResourceProcessAutomationTrigger() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Process Automation Trigger`,

		CreateContext: provider.CreateWithPooledClient(createProcessAutomationTrigger),
		ReadContext:   provider.ReadWithPooledClient(readProcessAutomationTrigger),
		UpdateContext: provider.UpdateWithPooledClient(updateProcessAutomationTrigger),
		DeleteContext: provider.DeleteWithPooledClient(removeProcessAutomationTrigger),
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
				Description: "Match criteria that controls when the trigger will fire. NOTE: The match_criteria field type has changed from a complex object to a string. This was done to allow for complex JSON object definitions.",
				Type:        schema.TypeString,
				Optional:    true,
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

func ProcessAutomationTriggerExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllProcessAutomationTriggersResourceMap),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
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
	matchingCriteria := d.Get("match_criteria").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	if eventTTLSeconds > 0 && delayBySeconds > 0 {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Only one of event_ttl_seconds or delay_by_seconds can be set."), fmt.Errorf("event_ttl_seconds and delay_by_seconds are both set"))
	}

	log.Printf("Creating process automation trigger %s", name)

	triggerInput := &ProcessAutomationTrigger{
		TopicName:     &topic_name,
		Name:          &name,
		Target:        buildTarget(d),
		MatchCriteria: &matchingCriteria,
		Enabled:       &enabled,
		Description:   &description,
	}

	if eventTTLSeconds > 0 {
		triggerInput.EventTTLSeconds = &eventTTLSeconds
	}

	if delayBySeconds > 0 {
		triggerInput.DelayBySeconds = &delayBySeconds
	}

	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		trigger, resp, err := postProcessAutomationTrigger(triggerInput, integAPI)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create process automation trigger %s error: %s", name, err), resp)
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
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceProcessAutomationTrigger(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading process automation trigger %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		trigger, resp, getErr := getProcessAutomationTrigger(d.Id(), integAPI)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read process automation trigger %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to process read automation trigger %s | error: %s", d.Id(), getErr), resp))
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

		d.Set("match_criteria", trigger.MatchCriteria)
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
		return cc.CheckState(d)
	})
}

func updateProcessAutomationTrigger(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	enabled := d.Get("enabled").(bool)
	eventTTLSeconds := d.Get("event_ttl_seconds").(int)
	delayBySeconds := d.Get("delay_by_seconds").(int)
	description := d.Get("description").(string)
	matchingCriteria := d.Get("match_criteria").(string)

	topic_name := d.Get("topic_name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Updating process automation trigger %s", name)

	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get the latest trigger version to send with PATCH
		trigger, resp, getErr := getProcessAutomationTrigger(d.Id(), integAPI)
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read process automation trigger %s error: %s", d.Id(), getErr), resp)
		}

		if eventTTLSeconds > 0 && delayBySeconds > 0 {
			return resp, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Only one of event_ttl_seconds or delay_by_seconds can be set."), fmt.Errorf("event_ttl_seconds and delay_by_seconds are both set"))
		}

		triggerInput := &ProcessAutomationTrigger{
			TopicName:     &topic_name,
			Name:          &name,
			Enabled:       &enabled,
			Target:        buildTarget(d),
			MatchCriteria: &matchingCriteria,
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
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update process automation trigger %s error: %s", name, err), resp)
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

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Deleting process automation trigger %s", name)

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		resp, err := deleteProcessAutomationTrigger(d.Id(), integAPI)

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("process automation trigger already deleted %s", d.Id())
				return nil
			}
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("process automation trigger %s still exists", d.Id()), resp))
		}
		return nil
	})
}

func buildTarget(d *schema.ResourceData) *Target {
	if target := d.Get("target"); target != nil {
		if targetList := target.(*schema.Set).List(); len(targetList) > 0 {
			targetMap := targetList[0].(map[string]interface{})

			targetType := targetMap["type"].(string)
			id := targetMap["id"].(string)

			target := &Target{
				Type: &targetType,
				Id:   &id,
			}

			workflowTargetSettingsInput := targetMap["workflow_target_settings"].(*schema.Set).List()

			if len(workflowTargetSettingsInput) > 0 {
				workflowTargetSettingsInputMap := workflowTargetSettingsInput[0].(map[string]interface{})
				dataFormat := workflowTargetSettingsInputMap["data_format"].(string)
				if dataFormat == "" {
					return target
				}
				target.WorkflowTargetSettings = &WorkflowTargetSettings{
					DataFormat: &dataFormat,
				}
			}

			return target
		}
	}

	return &Target{}
}

func flattenTarget(inputTarget *Target) *schema.Set {
	if inputTarget == nil {
		return nil
	}

	targetSet := schema.NewSet(schema.HashResource(target), []interface{}{})

	flattendedTarget := make(map[string]interface{})
	flattendedTarget["id"] = *inputTarget.Id
	flattendedTarget["type"] = *inputTarget.Type

	if inputTarget.WorkflowTargetSettings != nil {
		worklfowTargetSettingsSet := schema.NewSet(schema.HashResource(target), []interface{}{})
		flattendedWorkflowTargetSettings := make(map[string]interface{})
		flattendedWorkflowTargetSettings["data_format"] = *inputTarget.WorkflowTargetSettings.DataFormat
		worklfowTargetSettingsSet.Add(flattendedWorkflowTargetSettings)

		flattendedTarget["workflow_target_settings"] = worklfowTargetSettingsSet
	}
	targetSet.Add(flattendedTarget)

	return targetSet
}
