package recording_media_retention_policy

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
The genesyscloud_recording_media_retention_policy_schema.go should hold four types of functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the genesyscloud_recording_media_retention_policy resource.
3.  The datasource schema definitions for the genesyscloud_recording_media_retention_policy datasource.
4.  The resource exporter configuration for the genesyscloud_recording_media_retention_policy exporter.
*/

const resourceName = "genesyscloud_recording_media_retention_policy"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, DataSourceRecordingMediaRetentionPolicy())
	l.RegisterResource(resourceName, ResourceMediaRetentionPolicy())
	l.RegisterExporter(resourceName, MediaRetentionPolicyExporter())
}

// ResourceMediaRetentionPolicy registers the genesyscloud_recording_media_retention_policy resource with Terraform
func ResourceMediaRetentionPolicy() *schema.Resource {
	timeSlot := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"start_time": {
				Description: "start time in xx:xx:xx.xxx format",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"stop_time": {
				Description: "stop time in xx:xx:xx.xxx format",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"day": {
				Description: "Day for this time slot, Monday = 1 ... Sunday = 7",
				Type:        schema.TypeInt,
				Optional:    true,
			},
		},
	}

	timeAllowed := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"time_slots": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        timeSlot,
			},
			"time_zone_id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"empty": {
				Description: "",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}

	durationCondition := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"duration_target": {
				Description:  "",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"DURATION", "DURATION_RANGE"}, false),
			},
			"duration_operator": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"duration_range": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"duration_mode": {
				Description:  "",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Between", "Over", "Under"}, false),
			},
		},
	}

	callMediaPolicyConditions := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"for_user_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"date_ranges": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Default:     nil,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"for_queue_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"wrapup_code_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"language_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"time_allowed": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        timeAllowed,
			},
			"directions": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Default:     nil,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"duration": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        durationCondition,
			},
		},
	}

	chatMediaPolicyConditions := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"for_user_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"date_ranges": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Default:     nil,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"for_queue_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"wrapup_code_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"language_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"time_allowed": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        timeAllowed,
			},
			"duration": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        durationCondition,
			},
		},
	}

	emailMediaPolicyConditions := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"for_user_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"date_ranges": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Default:     nil,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"for_queue_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"wrapup_code_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"language_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"time_allowed": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        timeAllowed,
			},
		},
	}

	messageMediaPolicyConditions := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"for_user_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"date_ranges": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Default:     nil,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"for_queue_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"wrapup_code_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"language_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"time_allowed": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        timeAllowed,
			},
		},
	}

	policyConditions := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"for_user_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"directions": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Default:     nil,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"date_ranges": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Default:     nil,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"media_types": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"for_queue_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"duration": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        durationCondition,
			},
			"wrapup_code_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"time_allowed": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        timeAllowed,
			},
		},
	}

	userParam := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"value": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	archiveRetention := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"days": {
				Description: "",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"storage_medium": {
				Description:  "",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"CLOUDARCHIVE"}, false)},
		},
	}

	deleteRetention := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"days": {
				Description: "",
				Type:        schema.TypeInt,
				Optional:    true,
			},
		},
	}

	policyErrorMessage := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"status_code": {
				Description: "",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"user_message": {
				Description: "",
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
			"user_params_message": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"error_code": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"correlation_id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"user_params": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        userParam,
			},
			"insert_date": {
				Description: "Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	agentTimeInterval := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"months": {
				Description: "",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"weeks": {
				Description: "",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"days": {
				Description: "",
				Type:        schema.TypeInt,
				Optional:    true,
			},
		},
	}

	evalTimeInterval := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"hours": {
				Description: "",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"days": {
				Description: "",
				Type:        schema.TypeInt,
				Optional:    true,
			},
		},
	}

	policyErrors := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"policy_error_messages": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        policyErrorMessage,
			},
		},
	}

	evaluationAssignment := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"evaluation_form_id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"user_id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	meteredEvaluationAssignment := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"evaluator_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"max_number_evaluations": {
				Description: "",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"evaluation_form_id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"assign_to_active_user": {
				Description: "",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"time_interval": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        evalTimeInterval,
			},
		},
	}

	meteredAssignmentByAgent := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"evaluator_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"max_number_evaluations": {
				Description: "",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"evaluation_form_id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"time_interval": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        agentTimeInterval,
			},
			"time_zone": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	calibrationAssignment := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"calibrator_id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"evaluator_ids": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"evaluation_form_id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"expert_evaluator_id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	surveyAssignment := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"survey_form_name": {
				Description: "The survey form used for this survey.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"flow_id": {
				Description: "The UUID reference to the flow associated with this survey.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"invite_time_interval": {
				Description: "An ISO 8601 repeated interval consisting of the number of repetitions, the start datetime, and the interval (e.g. R2/2018-03-01T13:00:00Z/P1M10DT2H30M). Total duration must not exceed 90 days.",
				Type:        schema.TypeString,
				Default:     "R1/P0M",
				Optional:    true,
			},
			"sending_user": {
				Description: "User together with sendingDomain used to send email, null to use no-reply",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"sending_domain": {
				Description: "Validated email domain, required",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}

	retentionDuration := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"archive_retention": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        archiveRetention,
			},
			"delete_retention": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        deleteRetention,
			},
		},
	}

	initiateScreenRecording := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"record_acw": {
				Description: "",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"archive_retention": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        archiveRetention,
			},
			"delete_retention": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        deleteRetention,
			},
		},
	}

	mediaTranscription := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"display_name": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"transcription_provider": {
				Description:  "",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"VOCI", "CALLJOURNEY"}, false),
			},
			"integration_id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	integrationExport := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"integration_id": {
				Description: "The aws-s3-recording-bulk-actions-integration that the policy uses for exports.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"should_export_screen_recordings": {
				Description: "True if the policy should export screen recordings in addition to the other conversation media.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
		},
	}

	policyActions := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"retain_recording": {
				Description: "true to retain the recording associated with the conversation.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"delete_recording": {
				Description: "true to delete the recording associated with the conversation. If retainRecording = true, this will be ignored.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"always_delete": {
				Description: "true to delete the recording associated with the conversation regardless of the values of retainRecording or deleteRecording.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"assign_evaluations": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        evaluationAssignment,
			},
			"assign_metered_evaluations": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        meteredEvaluationAssignment,
			},
			"assign_metered_assignment_by_agent": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        meteredAssignmentByAgent,
			},
			"assign_calibrations": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        calibrationAssignment,
			},
			"assign_surveys": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        surveyAssignment,
			},
			"retention_duration": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        retentionDuration,
			},
			"initiate_screen_recording": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        initiateScreenRecording,
			},
			"media_transcriptions": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        mediaTranscription,
			},
			"integration_export": {
				Description: "Policy action for exporting recordings using an integration to 3rd party s3.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        integrationExport,
			},
		},
	}

	callMediaPolicy := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"actions": {
				Description: "Actions applied when specified conditions are met",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        policyActions,
			},
			"conditions": {
				Description: "Conditions for when actions should be applied",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        callMediaPolicyConditions,
			},
		},
	}

	chatMediaPolicy := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"actions": {
				Description: "Actions applied when specified conditions are met",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        policyActions,
			},
			"conditions": {
				Description: "Conditions for when actions should be applied",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        chatMediaPolicyConditions,
			},
		},
	}

	emailMediaPolicy := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"actions": {
				Description: "Actions applied when specified conditions are met",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        policyActions,
			},
			"conditions": {
				Description: "Conditions for when actions should be applied",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        emailMediaPolicyConditions,
			},
		},
	}

	messageMediaPolicy := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"actions": {
				Description: "Actions applied when specified conditions are met",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        policyActions,
			},
			"conditions": {
				Description: "Conditions for when actions should be applied",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        messageMediaPolicyConditions,
			},
		},
	}

	mediaPolicies := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"call_policy": {
				Description: "Conditions and actions for calls",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        callMediaPolicy,
			},
			"chat_policy": {
				Description: "Conditions and actions for calls",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        chatMediaPolicy,
			},
			"email_policy": {
				Description: "Conditions and actions for calls",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        emailMediaPolicy,
			},
			"message_policy": {
				Description: "Conditions and actions for calls",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        messageMediaPolicy,
			},
		},
	}

	return &schema.Resource{
		Description:   "Genesys Cloud Media Retention Policies",
		CreateContext: provider.CreateWithPooledClient(createMediaRetentionPolicy),
		ReadContext:   provider.ReadWithPooledClient(readMediaRetentionPolicy),
		UpdateContext: provider.UpdateWithPooledClient(updateMediaRetentionPolicy),
		DeleteContext: provider.DeleteWithPooledClient(deleteMediaRetentionPolicy),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The policy name. Changing the policy_name attribute will cause the recording_media_retention_policy to be dropped and recreated with a new ID.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"order": {
				Description: "The ordinal number for the policy",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"description": {
				Description: "The description for the policy",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enabled": {
				Description: "The policy will be enabled if true, otherwise it will be disabled",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"media_policies": {
				Description: "Conditions and actions per media type",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        mediaPolicies,
			},
			"conditions": {
				Description: "Conditions",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        policyConditions,
			},
			"actions": {
				Description: "Actions",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        policyActions,
			},
			"policy_errors": {
				Description: "A list of errors in the policy configuration",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        policyErrors,
			},
		},
	}
}

// MediaRetentionPolicyExporter returns the resourceExporter object used to hold the genesyscloud_recording_media_retention_policy exporter's config
func MediaRetentionPolicyExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllMediaRetentionPolicies),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"media_policies.chat_policy.conditions.for_queue_ids":                                         {RefType: "genesyscloud_routing_queue", AltValues: []string{"*"}},
			"media_policies.call_policy.conditions.for_queue_ids":                                         {RefType: "genesyscloud_routing_queue", AltValues: []string{"*"}},
			"media_policies.message_policy.conditions.for_queue_ids":                                      {RefType: "genesyscloud_routing_queue", AltValues: []string{"*"}},
			"media_policies.email_policy.conditions.for_queue_ids":                                        {RefType: "genesyscloud_routing_queue", AltValues: []string{"*"}},
			"conditions.for_queue_ids":                                                                    {RefType: "genesyscloud_routing_queue", AltValues: []string{"*"}},
			"media_policies.call_policy.conditions.for_user_ids":                                          {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.chat_policy.conditions.for_user_ids":                                          {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.email_policy.conditions.for_user_ids":                                         {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.message_policy.conditions.for_user_ids":                                       {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"conditions.for_user_ids":                                                                     {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.call_policy.actions.assign_evaluations.evaluation_form_id":                    {RefType: "genesyscloud_quality_forms_evaluation"},
			"media_policies.call_policy.actions.assign_calibrations.evaluation_form_id":                   {RefType: "genesyscloud_quality_forms_evaluation"},
			"media_policies.call_policy.actions.assign_metered_evaluations.evaluation_form_id":            {RefType: "genesyscloud_quality_forms_evaluation"},
			"media_policies.call_policy.actions.assign_metered_assignment_by_agent.evaluation_form_id":    {RefType: "genesyscloud_quality_forms_evaluation"},
			"media_policies.chat_policy.actions.assign_evaluations.evaluation_form_id":                    {RefType: "genesyscloud_quality_forms_evaluation"},
			"media_policies.chat_policy.actions.assign_calibrations.evaluation_form_id":                   {RefType: "genesyscloud_quality_forms_evaluation"},
			"media_policies.chat_policy.actions.assign_metered_evaluations.evaluation_form_id":            {RefType: "genesyscloud_quality_forms_evaluation"},
			"media_policies.chat_policy.actions.assign_metered_assignment_by_agent.evaluation_form_id":    {RefType: "genesyscloud_quality_forms_evaluation"},
			"media_policies.message_policy.actions.assign_evaluations.evaluation_form_id":                 {RefType: "genesyscloud_quality_forms_evaluation"},
			"media_policies.message_policy.actions.assign_calibrations.evaluation_form_id":                {RefType: "genesyscloud_quality_forms_evaluation"},
			"media_policies.message_policy.actions.assign_metered_evaluations.evaluation_form_id":         {RefType: "genesyscloud_quality_forms_evaluation"},
			"media_policies.message_policy.actions.assign_metered_assignment_by_agent.evaluation_form_id": {RefType: "genesyscloud_quality_forms_evaluation"},
			"media_policies.email_policy.actions.assign_evaluations.evaluation_form_id":                   {RefType: "genesyscloud_quality_forms_evaluation"},
			"media_policies.email_policy.actions.assign_calibrations.evaluation_form_id":                  {RefType: "genesyscloud_quality_forms_evaluation"},
			"media_policies.email_policy.actions.assign_metered_evaluations.evaluation_form_id":           {RefType: "genesyscloud_quality_forms_evaluation"},
			"media_policies.email_policy.actions.assign_metered_assignment_by_agent.evaluation_form_id":   {RefType: "genesyscloud_quality_forms_evaluation"},
			"actions.assign_evaluations.evaluation_form_id":                                               {RefType: "genesyscloud_quality_forms_evaluation"},
			"actions.assign_calibrations.evaluation_form_id":                                              {RefType: "genesyscloud_quality_forms_evaluation"},
			"actions.assign_metered_evaluations.evaluation_form_id":                                       {RefType: "genesyscloud_quality_forms_evaluation"},
			"actions.assign_metered_assignment_by_agent.evaluation_form_id":                               {RefType: "genesyscloud_quality_forms_evaluation"},
			"media_policies.call_policy.actions.assign_evaluations.evaluator_ids":                         {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.call_policy.actions.assign_calibrations.evaluator_ids":                        {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.call_policy.actions.assign_metered_evaluations.evaluator_ids":                 {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.call_policy.actions.assign_metered_assignment_by_agent.evaluator_ids":         {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.chat_policy.actions.assign_evaluations.evaluator_ids":                         {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.chat_policy.actions.assign_calibrations.evaluator_ids":                        {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.chat_policy.actions.assign_metered_evaluations.evaluator_ids":                 {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.chat_policy.actions.assign_metered_assignment_by_agent.evaluator_ids":         {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.message_policy.actions.assign_evaluations.evaluator_ids":                      {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.message_policy.actions.assign_calibrations.evaluator_ids":                     {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.message_policy.actions.assign_metered_evaluations.evaluator_ids":              {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.message_policy.actions.assign_metered_assignment_by_agent.evaluator_ids":      {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.email_policy.actions.assign_evaluations.evaluator_ids":                        {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.email_policy.actions.assign_calibrations.evaluator_ids":                       {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.email_policy.actions.assign_metered_evaluations.evaluator_ids":                {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.email_policy.actions.assign_metered_assignment_by_agent.evaluator_ids":        {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"actions.assign_evaluations.evaluator_ids":                                                    {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"actions.assign_calibrations.evaluator_ids":                                                   {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"actions.assign_metered_evaluations.evaluator_ids":                                            {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"actions.assign_metered_assignment_by_agent.evaluator_ids":                                    {RefType: "genesyscloud_user", AltValues: []string{"*"}},
			"media_policies.call_policy.actions.assign_calibrations.calibrator_id":                        {RefType: "genesyscloud_user"},
			"media_policies.chat_policy.actions.assign_calibrations.calibrator_id":                        {RefType: "genesyscloud_user"},
			"media_policies.message_policy.actions.assign_calibrations.calibrator_id":                     {RefType: "genesyscloud_user"},
			"media_policies.email_policy.actions.assign_calibrations.calibrator_id":                       {RefType: "genesyscloud_user"},
			"media_policies.call_policy.actions.assign_calibrations.expert_evaluator_id":                  {RefType: "genesyscloud_user"},
			"media_policies.chat_policy.actions.assign_calibrations.expert_evaluator_id":                  {RefType: "genesyscloud_user"},
			"media_policies.message_policy.actions.assign_calibrations.expert_evaluator_id":               {RefType: "genesyscloud_user"},
			"media_policies.email_policy.actions.assign_calibrations.expert_evaluator_id":                 {RefType: "genesyscloud_user"},
			"actions.assign_calibrations.expert_evaluator_id":                                             {RefType: "genesyscloud_user"},
			"media_policies.call_policy.conditions.language_ids":                                          {RefType: "genesyscloud_routing_language", AltValues: []string{"*"}},
			"media_policies.chat_policy.conditions.language_ids":                                          {RefType: "genesyscloud_routing_language", AltValues: []string{"*"}},
			"media_policies.message_policy.conditions.language_ids":                                       {RefType: "genesyscloud_routing_language", AltValues: []string{"*"}},
			"media_policies.email_policy.conditions.language_ids":                                         {RefType: "genesyscloud_routing_language", AltValues: []string{"*"}},
			"media_policies.call_policy.conditions.wrapup_code_ids":                                       {RefType: "genesyscloud_routing_wrapupcode", AltValues: []string{"*"}},
			"media_policies.chat_policy.conditions.wrapup_code_ids":                                       {RefType: "genesyscloud_routing_wrapupcode", AltValues: []string{"*"}},
			"media_policies.message_policy.conditions.wrapup_code_ids":                                    {RefType: "genesyscloud_routing_wrapupcode", AltValues: []string{"*"}},
			"media_policies.email_policy.conditions.wrapup_code_ids":                                      {RefType: "genesyscloud_routing_wrapupcode", AltValues: []string{"*"}},
			"conditions.wrapup_code_ids":                                                                  {RefType: "genesyscloud_routing_wrapupcode", AltValues: []string{"*"}},
			"media_policies.call_policy.actions.integration_export.integration_id":                        {RefType: "genesyscloud_integration"},
			"media_policies.chat_policy.actions.integration_export.integration_id":                        {RefType: "genesyscloud_integration"},
			"media_policies.message_policy.actions.integration_export.integration_id":                     {RefType: "genesyscloud_integration"},
			"media_policies.email_policy.actions.integration_export.integration_id":                       {RefType: "genesyscloud_integration"},
			"actions.media_transcriptions.integration_id":                                                 {RefType: "genesyscloud_integration"},
			"media_policies.call_policy.actions.assign_surveys.flow_id":                                   {RefType: "genesyscloud_flow"},
			"media_policies.chat_policy.actions.assign_surveys.flow_id":                                   {RefType: "genesyscloud_flow"},
			"media_policies.message_policy.actions.assign_surveys.flow_id":                                {RefType: "genesyscloud_flow"},
			"media_policies.email_policy.actions.assign_surveys.flow_id":                                  {RefType: "genesyscloud_flow"},
			"actions.assign_surveys.flow_id":                                                              {RefType: "genesyscloud_flow"},
			"media_policies.call_policy.actions.assign_evaluations.user_id":                               {RefType: "genesyscloud_user"},
			"media_policies.chat_policy.actions.assign_evaluations.user_id":                               {RefType: "genesyscloud_user"},
			"media_policies.message_policy.actions.assign_evaluations.user_id":                            {RefType: "genesyscloud_user"},
			"media_policies.email_policy.actions.assign_evaluations.user_id":                              {RefType: "genesyscloud_user"},
			"actions.assign_evaluations.user_id":                                                          {RefType: "genesyscloud_user"},
		},
		AllowZeroValues: []string{"order"},
		RemoveIfMissing: map[string][]string{
			"":               {"conditions", "actions"},
			"media_policies": {"call_policy", "chat_policy", "message_policy", "email_policy"},
		},
	}
}

// DataSourceRecordingMediaRetentionPolicy registers the genesyscloud_recording_media_retention_policy data source
func DataSourceRecordingMediaRetentionPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud media retention policy. Select a policy by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceRecordingMediaRetentionPolicyRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Media retention policy name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
