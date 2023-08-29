package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

type EvaluationFormQuestionGroupStruct struct {
	Name                    string
	DefaultAnswersToHighest bool
	DefaultAnswersToNA      bool
	NaEnabled               bool
	Weight                  float32
	ManualWeight            bool
	Questions               []EvaluationFormQuestionStruct
	VisibilityCondition     VisibilityConditionStruct
}

type EvaluationFormStruct struct {
	Name           string
	Published      bool
	QuestionGroups []EvaluationFormQuestionGroupStruct
}

type EvaluationFormQuestionStruct struct {
	Text                string
	HelpText            string
	NaEnabled           bool
	CommentsRequired    bool
	IsKill              bool
	IsCritical          bool
	VisibilityCondition VisibilityConditionStruct
	AnswerOptions       []AnswerOptionStruct
}

type AnswerOptionStruct struct {
	Text  string
	Value int
}

type VisibilityConditionStruct struct {
	CombiningOperation string
	Predicates         []string
}

var (
	mediaPolicies = &schema.Resource{
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

	chatMediaPolicy = &schema.Resource{
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

	callMediaPolicy = &schema.Resource{
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

	emailMediaPolicy = &schema.Resource{
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

	messageMediaPolicy = &schema.Resource{
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

	policyActions = &schema.Resource{
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

	evaluationAssignment = &schema.Resource{
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

	meteredEvaluationAssignment = &schema.Resource{
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
				Elem:        timeInterval,
			},
		},
	}

	meteredAssignmentByAgent = &schema.Resource{
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
				Elem:        timeInterval,
			},
			"time_zone": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	calibrationAssignment = &schema.Resource{
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

	surveyAssignment = &schema.Resource{
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

	retentionDuration = &schema.Resource{
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

	initiateScreenRecording = &schema.Resource{
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

	mediaTranscription = &schema.Resource{
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

	integrationExport = &schema.Resource{
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

	timeInterval = &schema.Resource{
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
			"hours": {
				Description: "",
				Type:        schema.TypeInt,
				Optional:    true,
			},
		},
	}

	policyErrors = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"policy_error_messages": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        policyErrorMessage,
			},
		},
	}

	policyErrorMessage = &schema.Resource{
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

	userParam = &schema.Resource{
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

	archiveRetention = &schema.Resource{
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

	deleteRetention = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"days": {
				Description: "",
				Type:        schema.TypeInt,
				Optional:    true,
			},
		},
	}

	timeSlot = &schema.Resource{
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

	timeAllowed = &schema.Resource{
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

	durationCondition = &schema.Resource{
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

	callMediaPolicyConditions = &schema.Resource{
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

	chatMediaPolicyConditions = &schema.Resource{
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

	emailMediaPolicyConditions = &schema.Resource{
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

	messageMediaPolicyConditions = &schema.Resource{
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

	policyConditions = &schema.Resource{
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

	qualityAPI = platformclientv2.NewQualityApi()
)

func ResourceMediaRetentionPolicy() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Media Retention Policies",
		CreateContext: CreateWithPooledClient(createMediaRetentionPolicy),
		ReadContext:   ReadWithPooledClient(readMediaRetentionPolicy),
		UpdateContext: UpdateWithPooledClient(updateMediaRetentionPolicy),
		DeleteContext: DeleteWithPooledClient(deleteMediaRetentionPolicy),
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

func getAllMediaRetentionPolicies(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	recordingAPI := platformclientv2.NewRecordingApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		retentionPolicies, _, getErr := recordingAPI.GetRecordingMediaretentionpolicies(pageSize, pageNum, "", []string{}, "", "", "", true, false, false, 365)

		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of media retention policies %v", getErr)
		}

		if retentionPolicies.Entities == nil || len(*retentionPolicies.Entities) == 0 {
			break
		}

		for _, retentionPolicy := range *retentionPolicies.Entities {
			resources[*retentionPolicy.Id] = &resourceExporter.ResourceMeta{Name: *retentionPolicy.Name}
		}
	}

	return resources, nil
}

func buildEvaluationAssignments(evaluations []interface{}) *[]platformclientv2.Evaluationassignment {
	assignEvaluations := make([]platformclientv2.Evaluationassignment, 0)

	for _, assignEvaluation := range evaluations {
		assignEvaluationMap, ok := assignEvaluation.(map[string]interface{})
		if !ok {
			continue
		}
		evaluationFormId := assignEvaluationMap["evaluation_form_id"].(string)
		userId := assignEvaluationMap["user_id"].(string)
		assignment := platformclientv2.Evaluationassignment{}

		// if evaluation form id is present, get the context id and build the evaluation form
		if evaluationFormId != "" {
			form, _, err := qualityAPI.GetQualityFormsEvaluation(evaluationFormId)
			if err != nil {
				log.Fatalf("failed to read evaluation form %s: %s", evaluationFormId, err)
			} else {
				evaluationFormContextId := form.ContextId
				assignment.EvaluationForm = &platformclientv2.Evaluationform{Id: &evaluationFormId, ContextId: evaluationFormContextId}
			}
		}
		if userId != "" {
			assignment.User = &platformclientv2.User{Id: &userId}
		}
		assignEvaluations = append(assignEvaluations, assignment)
	}

	return &assignEvaluations
}

func flattenEvaluationAssignments(assignments *[]platformclientv2.Evaluationassignment) []interface{} {
	if assignments == nil {
		return nil
	}

	evaluationAssignments := []interface{}{}
	for _, assignment := range *assignments {
		assignmentMap := make(map[string]interface{})

		// if form is present in the response, assign the most recent unpublished version id to align with evaluation form resource behavior for export purposes.
		if assignment.EvaluationForm != nil {
			formId := *assignment.EvaluationForm.Id
			formVersions, _, err := qualityAPI.GetQualityFormsEvaluationVersions(formId, 25, 1, "desc")
			if err != nil {
				log.Fatalf("Failed to get evaluation form versions %s", *assignment.EvaluationForm.Name)
			} else if formVersions.Entities == nil || len(*formVersions.Entities) == 0 {
				log.Fatalf("No versions found for form %s", formId)
			} else {
				formId = *(*formVersions.Entities)[0].Id
			}

			assignmentMap["evaluation_form_id"] = formId
		}
		if assignment.User != nil {
			assignmentMap["user_id"] = *assignment.User.Id
		}
		evaluationAssignments = append(evaluationAssignments, assignmentMap)
	}
	return evaluationAssignments
}

func buildTimeInterval(timeInterval []interface{}) *platformclientv2.Timeinterval {
	if timeInterval == nil || len(timeInterval) <= 0 {
		return nil
	}

	timeIntervalMap, ok := timeInterval[0].(map[string]interface{})
	if !ok {
		return nil
	}

	months := timeIntervalMap["months"].(int)
	weeks := timeIntervalMap["weeks"].(int)
	days := timeIntervalMap["days"].(int)
	hours := timeIntervalMap["hours"].(int)

	return &platformclientv2.Timeinterval{
		Months: &months,
		Weeks:  &weeks,
		Days:   &days,
		Hours:  &hours,
	}
}

func flattenTimeInterval(timeInterval *platformclientv2.Timeinterval) []interface{} {
	if timeInterval == nil {
		return nil
	}

	timeIntervalMap := make(map[string]interface{})
	if timeInterval.Months != nil {
		timeIntervalMap["months"] = *timeInterval.Months
	}
	if timeInterval.Weeks != nil {
		timeIntervalMap["weeks"] = *timeInterval.Weeks
	}
	if timeInterval.Days != nil {
		timeIntervalMap["days"] = *timeInterval.Days
	}
	if timeInterval.Hours != nil {
		timeIntervalMap["hours"] = *timeInterval.Hours
	}

	return []interface{}{timeIntervalMap}
}

func buildAssignMeteredEvaluations(assignments []interface{}) *[]platformclientv2.Meteredevaluationassignment {
	meteredAssignments := make([]platformclientv2.Meteredevaluationassignment, 0)

	for _, assignment := range assignments {
		assignmentMap, ok := assignment.(map[string]interface{})
		if !ok {
			continue
		}
		maxNumberEvaluations := assignmentMap["max_number_evaluations"].(int)
		assignToActiveUser := assignmentMap["assign_to_active_user"].(bool)
		evaluationFormId := assignmentMap["evaluation_form_id"].(string)
		evaluatorIds := assignmentMap["evaluator_ids"].([]interface{})

		idStrings := make([]string, 0)
		for _, evaluatorId := range evaluatorIds {
			idStrings = append(idStrings, fmt.Sprintf("%v", evaluatorId))
		}

		evaluators := make([]platformclientv2.User, 0)
		for _, evaluatorId := range idStrings {
			evaluator := evaluatorId
			evaluators = append(evaluators, platformclientv2.User{Id: &evaluator})
		}

		temp := platformclientv2.Meteredevaluationassignment{
			Evaluators:           &evaluators,
			MaxNumberEvaluations: &maxNumberEvaluations,
			AssignToActiveUser:   &assignToActiveUser,
			TimeInterval:         buildTimeInterval(assignmentMap["time_interval"].([]interface{})),
		}

		// if evaluation form id is present, get the context id and build the evaluation form
		if evaluationFormId != "" {
			form, _, err := qualityAPI.GetQualityFormsEvaluation(evaluationFormId)
			if err != nil {
				log.Fatalf("failed to read media evaluation form %s: %s", evaluationFormId, err)
			} else {
				evaluationFormContextId := form.ContextId
				temp.EvaluationForm = &platformclientv2.Evaluationform{Id: &evaluationFormId, ContextId: evaluationFormContextId}
			}
		}
		meteredAssignments = append(meteredAssignments, temp)
	}

	return &meteredAssignments
}

func flattenAssignMeteredEvaluations(assignments *[]platformclientv2.Meteredevaluationassignment) []interface{} {
	if assignments == nil {
		return nil
	}

	meteredAssignments := []interface{}{}
	for _, assignment := range *assignments {
		assignmentMap := make(map[string]interface{})
		if assignment.Evaluators != nil {
			evaluatorIds := make([]string, 0)
			for _, evaluator := range *assignment.Evaluators {
				evaluatorIds = append(evaluatorIds, *evaluator.Id)
			}
			assignmentMap["evaluator_ids"] = evaluatorIds
		}
		if assignment.MaxNumberEvaluations != nil {
			assignmentMap["max_number_evaluations"] = *assignment.MaxNumberEvaluations
		}
		// if form is present in the response, assign the most recent unpublished version id to align with evaluation form resource behavior for export purposes.
		if assignment.EvaluationForm != nil {
			formId := *assignment.EvaluationForm.Id
			formVersions, _, err := qualityAPI.GetQualityFormsEvaluationVersions(formId, 25, 1, "desc")
			if err != nil {
				log.Fatalf("Failed to get evaluation form versions %s", *assignment.EvaluationForm.Name)
			} else if formVersions.Entities == nil || len(*formVersions.Entities) == 0 {
				log.Fatalf("No versions found for form %s", formId)
			} else {
				formId = *(*formVersions.Entities)[0].Id
			}

			assignmentMap["evaluation_form_id"] = formId
		}
		if assignment.AssignToActiveUser != nil {
			assignmentMap["assign_to_active_user"] = *assignment.AssignToActiveUser
		}
		if assignment.TimeInterval != nil {
			assignmentMap["time_interval"] = flattenTimeInterval(assignment.TimeInterval)
		}

		meteredAssignments = append(meteredAssignments, assignmentMap)
	}
	return meteredAssignments
}

func buildAssignMeteredAssignmentByAgent(assignments []interface{}) *[]platformclientv2.Meteredassignmentbyagent {
	meteredAssignments := make([]platformclientv2.Meteredassignmentbyagent, 0)
	for _, assignment := range assignments {
		assignmentMap, ok := assignment.(map[string]interface{})
		if !ok {
			continue
		}
		maxNumberEvaluations := assignmentMap["max_number_evaluations"].(int)
		timeZone := assignmentMap["time_zone"].(string)
		evaluationFormId := assignmentMap["evaluation_form_id"].(string)
		evaluatorIds := assignmentMap["evaluator_ids"].([]interface{})

		idStrings := make([]string, 0)
		for _, evaluatorId := range evaluatorIds {
			idStrings = append(idStrings, fmt.Sprintf("%v", evaluatorId))
		}

		evaluators := make([]platformclientv2.User, 0)
		for _, evaluatorId := range idStrings {
			evaluator := evaluatorId
			evaluators = append(evaluators, platformclientv2.User{Id: &evaluator})
		}

		temp := platformclientv2.Meteredassignmentbyagent{
			Evaluators:           &evaluators,
			MaxNumberEvaluations: &maxNumberEvaluations,
			TimeInterval:         buildTimeInterval(assignmentMap["time_interval"].([]interface{})),
			TimeZone:             &timeZone,
		}

		// if evaluation form id is present, get the context id and build the evaluation form
		if evaluationFormId != "" {
			form, _, err := qualityAPI.GetQualityFormsEvaluation(evaluationFormId)
			if err != nil {
				log.Fatalf("failed to read evaluation form %s: %s", evaluationFormId, err)
			} else {
				evaluationFormContextId := form.ContextId
				temp.EvaluationForm = &platformclientv2.Evaluationform{Id: &evaluationFormId, ContextId: evaluationFormContextId}
			}
		}

		meteredAssignments = append(meteredAssignments, temp)
	}

	return &meteredAssignments
}

func flattenAssignMeteredAssignmentByAgent(assignments *[]platformclientv2.Meteredassignmentbyagent) []interface{} {
	if assignments == nil {
		return nil
	}

	meteredAssignments := []interface{}{}
	for _, assignment := range *assignments {
		assignmentMap := make(map[string]interface{})
		if assignment.Evaluators != nil {
			evaluatorIds := make([]string, 0)
			for _, evaluator := range *assignment.Evaluators {
				evaluatorIds = append(evaluatorIds, *evaluator.Id)
			}
			assignmentMap["evaluator_ids"] = evaluatorIds
		}
		if assignment.MaxNumberEvaluations != nil {
			assignmentMap["max_number_evaluations"] = *assignment.MaxNumberEvaluations
		}
		// if form is present in the response, assign the most recent unpublished version id to align with evaluation form resource behavior for export purposes.
		if assignment.EvaluationForm != nil {
			formId := *assignment.EvaluationForm.Id
			formVersions, _, err := qualityAPI.GetQualityFormsEvaluationVersions(formId, 25, 1, "desc")
			if err != nil {
				log.Fatalf("Failed to get evaluation form versions %s", *assignment.EvaluationForm.Name)
			} else if formVersions.Entities == nil || len(*formVersions.Entities) == 0 {
				log.Fatalf("No versions found for form %s", formId)
			} else {
				formId = *(*formVersions.Entities)[0].Id
			}

			assignmentMap["evaluation_form_id"] = formId
		}
		if assignment.TimeInterval != nil {
			assignmentMap["time_interval"] = flattenTimeInterval(assignment.TimeInterval)
		}
		if assignment.TimeZone != nil {
			assignmentMap["time_zone"] = *assignment.TimeZone
		}

		meteredAssignments = append(meteredAssignments, assignmentMap)
	}
	return meteredAssignments
}

func buildAssignCalibrations(assignments []interface{}) *[]platformclientv2.Calibrationassignment {
	calibrationAssignments := make([]platformclientv2.Calibrationassignment, 0)

	for _, assignment := range assignments {
		assignmentMap, ok := assignment.(map[string]interface{})
		if !ok {
			continue
		}
		evaluationFormId := assignmentMap["evaluation_form_id"].(string)
		calibratorId := assignmentMap["calibrator_id"].(string)
		expertEvaluatorId := assignmentMap["expert_evaluator_id"].(string)
		evaluatorIds := assignmentMap["evaluator_ids"].([]interface{})

		idStrings := make([]string, 0)
		for _, evaluatorId := range evaluatorIds {
			idStrings = append(idStrings, fmt.Sprintf("%v", evaluatorId))
		}

		evaluators := make([]platformclientv2.User, 0)
		for _, evaluatorId := range idStrings {
			id := evaluatorId
			evaluators = append(evaluators, platformclientv2.User{Id: &id})
		}

		temp := platformclientv2.Calibrationassignment{
			Evaluators: &evaluators,
		}

		// if evaluation form id is present, get the context id and build the evaluation form
		if evaluationFormId != "" {
			form, _, err := qualityAPI.GetQualityFormsEvaluation(evaluationFormId)
			if err != nil {
				log.Fatalf("failed to read evaluation form %s: %s", evaluationFormId, err)
			} else {
				evaluationFormContextId := form.ContextId
				temp.EvaluationForm = &platformclientv2.Evaluationform{Id: &evaluationFormId, ContextId: evaluationFormContextId}
			}
		}

		if calibratorId != "" {
			temp.Calibrator = &platformclientv2.User{Id: &calibratorId}
		}
		if expertEvaluatorId != "" {
			temp.ExpertEvaluator = &platformclientv2.User{Id: &expertEvaluatorId}
		}

		calibrationAssignments = append(calibrationAssignments, temp)
	}

	return &calibrationAssignments
}

func flattenAssignCalibrations(assignments *[]platformclientv2.Calibrationassignment) []interface{} {
	if assignments == nil {
		return nil
	}

	calibrationAssignments := []interface{}{}
	for _, assignment := range *assignments {
		assignmentMap := make(map[string]interface{})
		if assignment.Calibrator != nil {
			assignmentMap["calibrator_id"] = *assignment.Calibrator.Id
		}
		if assignment.Evaluators != nil {
			evaluatorIds := make([]string, 0)
			for _, evaluator := range *assignment.Evaluators {
				evaluatorIds = append(evaluatorIds, *evaluator.Id)
			}
			assignmentMap["evaluator_ids"] = evaluatorIds
		}
		// if form is present in the response, assign the most recent unpublished version id to align with evaluation form resource behavior for export purposes.
		if assignment.EvaluationForm != nil {
			formId := *assignment.EvaluationForm.Id
			formVersions, _, err := qualityAPI.GetQualityFormsEvaluationVersions(formId, 25, 1, "desc")
			if err != nil {
				log.Fatalf("Failed to get evaluation form versions %s", *assignment.EvaluationForm.Name)
			} else if formVersions.Entities == nil || len(*formVersions.Entities) == 0 {
				log.Fatalf("No versions found for form %s", formId)
			} else {
				formId = *(*formVersions.Entities)[0].Id
			}

			assignmentMap["evaluation_form_id"] = formId
		}
		if assignment.ExpertEvaluator != nil {
			assignmentMap["expert_evaluator_id"] = *assignment.ExpertEvaluator.Id
		}

		calibrationAssignments = append(calibrationAssignments, assignmentMap)
	}
	return calibrationAssignments
}

func buildDomainEntityRef(idVal string) *platformclientv2.Domainentityref {
	if idVal == "nil" {
		return nil
	}

	return &platformclientv2.Domainentityref{
		Id: &idVal,
	}
}

func buildAssignSurveys(assignments []interface{}) *[]platformclientv2.Surveyassignment {
	surveyAssignments := make([]platformclientv2.Surveyassignment, 0)

	for _, assignment := range assignments {
		assignmentMap, ok := assignment.(map[string]interface{})
		if !ok {
			continue
		}
		sendingUser := assignmentMap["sending_user"].(string)
		sendingDomain := assignmentMap["sending_domain"].(string)
		inviteTimeInterval := assignmentMap["invite_time_interval"].(string)
		surveyFormName := assignmentMap["survey_form_name"].(string)

		temp := platformclientv2.Surveyassignment{
			Flow:               buildDomainEntityRef(assignmentMap["flow_id"].(string)),
			InviteTimeInterval: &inviteTimeInterval,
			SendingUser:        &sendingUser,
			SendingDomain:      &sendingDomain,
		}

		// If a survey form name is provided, get the context id and build the published survey form reference
		if surveyFormName != "" {
			const pageNum = 1
			const pageSize = 100
			forms, _, getErr := qualityAPI.GetQualityFormsSurveys(pageSize, pageNum, "", "", "", "", surveyFormName, "desc")
			if getErr != nil {
				log.Fatalf("Error requesting survey forms %s: %s", surveyFormName, getErr)
			} else if forms.Entities == nil || len(*forms.Entities) == 0 {
				log.Fatalf("No survey forms found with name %s", surveyFormName)
			} else {
				surveyFormReference := platformclientv2.Publishedsurveyformreference{Name: &surveyFormName, ContextId: (*forms.Entities)[0].ContextId}
				temp.SurveyForm = &surveyFormReference
			}
		}

		surveyAssignments = append(surveyAssignments, temp)
	}

	return &surveyAssignments
}

func flattenAssignSurveys(assignments *[]platformclientv2.Surveyassignment) []interface{} {
	if assignments == nil {
		return nil
	}

	surveyAssignments := []interface{}{}

	for _, assignment := range *assignments {
		assignmentMap := make(map[string]interface{})
		if assignment.SurveyForm != nil {
			assignmentMap["survey_form_name"] = *assignment.SurveyForm.Name
		}
		if assignment.Flow != nil {
			assignmentMap["flow_id"] = *assignment.Flow.Id
		}
		if assignment.InviteTimeInterval != nil {
			assignmentMap["invite_time_interval"] = *assignment.InviteTimeInterval
		}
		if assignment.SendingUser != nil {
			assignmentMap["sending_user"] = *assignment.SendingUser
		}
		if assignment.SendingDomain != nil {
			assignmentMap["sending_domain"] = *assignment.SendingDomain
		}

		surveyAssignments = append(surveyAssignments, assignmentMap)
	}
	return surveyAssignments
}

func buildArchiveRetention(archiveRetention []interface{}) *platformclientv2.Archiveretention {
	if archiveRetention == nil || len(archiveRetention) <= 0 {
		return nil
	}

	archiveRetentionMap, ok := archiveRetention[0].(map[string]interface{})
	if !ok {
		return nil
	}

	days := archiveRetentionMap["days"].(int)
	storageMedium := archiveRetentionMap["storage_medium"].(string)

	return &platformclientv2.Archiveretention{
		Days:          &days,
		StorageMedium: &storageMedium,
	}
}

func flattenArchiveRetention(archiveRetention *platformclientv2.Archiveretention) []interface{} {
	if archiveRetention == nil {
		return nil
	}

	archiveRetentionMap := make(map[string]interface{})
	if archiveRetention.Days != nil {
		archiveRetentionMap["days"] = *archiveRetention.Days
	}
	if archiveRetention.StorageMedium != nil {
		archiveRetentionMap["storage_medium"] = *archiveRetention.StorageMedium
	}

	return []interface{}{archiveRetentionMap}
}

func buildDeleteRetention(deleteRetention []interface{}) *platformclientv2.Deleteretention {
	if deleteRetention == nil || len(deleteRetention) <= 0 {
		return nil
	}

	deleteRetentionMap, ok := deleteRetention[0].(map[string]interface{})
	if !ok {
		return nil
	}

	days := deleteRetentionMap["days"].(int)

	return &platformclientv2.Deleteretention{
		Days: &days,
	}
}

func flattenDeleteRetention(deleteRetention *platformclientv2.Deleteretention) []interface{} {
	if deleteRetention == nil {
		return nil
	}

	deleteRetentionMap := make(map[string]interface{})
	if deleteRetention.Days != nil {
		deleteRetentionMap["days"] = *deleteRetention.Days
	}

	return []interface{}{deleteRetentionMap}
}

func buildRetentionDuration(retentionDuration []interface{}) *platformclientv2.Retentionduration {
	if retentionDuration == nil || len(retentionDuration) <= 0 {
		return nil
	}

	retentionDurationMap, ok := retentionDuration[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &platformclientv2.Retentionduration{
		ArchiveRetention: buildArchiveRetention(retentionDurationMap["archive_retention"].([]interface{})),
		DeleteRetention:  buildDeleteRetention(retentionDurationMap["delete_retention"].([]interface{})),
	}
}

func flattenRetentionDuration(retentionDuration *platformclientv2.Retentionduration) []interface{} {
	if retentionDuration == nil {
		return nil
	}

	retentionDurationMap := make(map[string]interface{})
	if retentionDuration.ArchiveRetention != nil {
		retentionDurationMap["archive_retention"] = flattenArchiveRetention(retentionDuration.ArchiveRetention)
	}
	if retentionDuration.DeleteRetention != nil {
		retentionDurationMap["delete_retention"] = flattenDeleteRetention(retentionDuration.DeleteRetention)
	}

	return []interface{}{retentionDurationMap}
}

func buildInitiateScreenRecording(initiateScreenRecording []interface{}) *platformclientv2.Initiatescreenrecording {
	if initiateScreenRecording == nil || len(initiateScreenRecording) <= 0 {
		return nil
	}

	initiateScreenRecordingMap, ok := initiateScreenRecording[0].(map[string]interface{})
	if !ok {
		return nil
	}
	recordACW := initiateScreenRecordingMap["record_acw"].(bool)

	return &platformclientv2.Initiatescreenrecording{
		RecordACW:        &recordACW,
		ArchiveRetention: buildArchiveRetention(initiateScreenRecordingMap["archive_retention"].([]interface{})),
		DeleteRetention:  buildDeleteRetention(initiateScreenRecordingMap["delete_retention"].([]interface{})),
	}
}

func flattenInitiateScreenRecording(recording *platformclientv2.Initiatescreenrecording) []interface{} {
	if recording == nil {
		return nil
	}

	recordingMap := make(map[string]interface{})
	if recording.RecordACW != nil {
		recordingMap["record_acw"] = *recording.RecordACW
	}
	if recording.ArchiveRetention != nil {
		recordingMap["archive_retention"] = flattenArchiveRetention(recording.ArchiveRetention)
	}
	if recording.DeleteRetention != nil {
		recordingMap["delete_retention"] = flattenDeleteRetention(recording.DeleteRetention)
	}

	return []interface{}{recordingMap}
}

func buildMediaTranscriptions(transcriptions []interface{}) *[]platformclientv2.Mediatranscription {
	mediaTranscriptions := make([]platformclientv2.Mediatranscription, 0)

	for _, transcription := range transcriptions {
		transcriptionMap, ok := transcription.(map[string]interface{})
		if !ok {
			continue
		}
		displayName := transcriptionMap["display_name"].(string)
		transcriptionProvider := transcriptionMap["transcription_provider"].(string)
		integrationId := transcriptionMap["integration_id"].(string)

		mediaTranscriptions = append(mediaTranscriptions, platformclientv2.Mediatranscription{
			DisplayName:           &displayName,
			TranscriptionProvider: &transcriptionProvider,
			IntegrationId:         &integrationId,
		})
	}

	return &mediaTranscriptions
}

func flattenMediaTranscriptions(transcriptions *[]platformclientv2.Mediatranscription) []interface{} {
	if transcriptions == nil {
		return nil
	}

	mediaTranscriptions := []interface{}{}

	for _, transcription := range *transcriptions {
		transcriptionMap := make(map[string]interface{})
		if transcription.DisplayName != nil {
			transcriptionMap["display_name"] = *transcription.DisplayName
		}
		if transcription.TranscriptionProvider != nil {
			transcriptionMap["transcription_provider"] = *transcription.TranscriptionProvider
		}
		if transcription.IntegrationId != nil {
			transcriptionMap["integration_id"] = *transcription.IntegrationId
		}

		mediaTranscriptions = append(mediaTranscriptions, transcriptionMap)
	}

	return mediaTranscriptions
}

func buildIntegrationExport(integrationExport []interface{}) *platformclientv2.Integrationexport {
	if integrationExport == nil || len(integrationExport) <= 0 {
		return nil
	}

	integrationExportMap, ok := integrationExport[0].(map[string]interface{})
	if !ok {
		return nil
	}
	shouldExportScreenRecordings := integrationExportMap["should_export_screen_recordings"].(bool)

	return &platformclientv2.Integrationexport{
		Integration:                  buildDomainEntityRef(integrationExportMap["integration_id"].(string)),
		ShouldExportScreenRecordings: &shouldExportScreenRecordings,
	}
}

func flattenIntegrationExport(integrationExport *platformclientv2.Integrationexport) []interface{} {
	if integrationExport == nil {
		return nil
	}

	integrationExportMap := make(map[string]interface{})
	if integrationExport.Integration != nil {
		integrationExportMap["integration_id"] = *integrationExport.Integration.Id
	}
	if integrationExport.ShouldExportScreenRecordings != nil {
		integrationExportMap["should_export_screen_recordings"] = *integrationExport.ShouldExportScreenRecordings
	}

	return []interface{}{integrationExportMap}
}

func buildPolicyActions(actions []interface{}) *platformclientv2.Policyactions {
	if actions == nil || len(actions) <= 0 {
		return nil
	}

	actionsMap, ok := actions[0].(map[string]interface{})
	if !ok {
		return nil
	}

	retainRecording := actionsMap["retain_recording"].(bool)
	deleteRecording := actionsMap["delete_recording"].(bool)
	alwaysDelete := actionsMap["always_delete"].(bool)

	return &platformclientv2.Policyactions{
		RetainRecording:                &retainRecording,
		DeleteRecording:                &deleteRecording,
		AlwaysDelete:                   &alwaysDelete,
		AssignEvaluations:              buildEvaluationAssignments(actionsMap["assign_evaluations"].([]interface{})),
		AssignMeteredEvaluations:       buildAssignMeteredEvaluations(actionsMap["assign_metered_evaluations"].([]interface{})),
		AssignMeteredAssignmentByAgent: buildAssignMeteredAssignmentByAgent(actionsMap["assign_metered_assignment_by_agent"].([]interface{})),
		AssignCalibrations:             buildAssignCalibrations(actionsMap["assign_calibrations"].([]interface{})),
		AssignSurveys:                  buildAssignSurveys(actionsMap["assign_surveys"].([]interface{})),
		RetentionDuration:              buildRetentionDuration(actionsMap["retention_duration"].([]interface{})),
		InitiateScreenRecording:        buildInitiateScreenRecording(actionsMap["initiate_screen_recording"].([]interface{})),
		MediaTranscriptions:            buildMediaTranscriptions(actionsMap["media_transcriptions"].([]interface{})),
		IntegrationExport:              buildIntegrationExport(actionsMap["integration_export"].([]interface{})),
	}
}

func flattenPolicyActions(actions *platformclientv2.Policyactions) []interface{} {
	if actions == nil || reflect.DeepEqual(platformclientv2.Policyactions{}, *actions) {
		return nil
	}

	actionsMap := make(map[string]interface{})
	if actions.RetainRecording != nil {
		actionsMap["retain_recording"] = *actions.RetainRecording
	}
	if actions.DeleteRecording != nil {
		actionsMap["delete_recording"] = *actions.DeleteRecording
	}
	if actions.AlwaysDelete != nil {
		actionsMap["always_delete"] = *actions.AlwaysDelete
	}
	if actions.AssignEvaluations != nil {
		actionsMap["assign_evaluations"] = flattenEvaluationAssignments(actions.AssignEvaluations)
	}
	if actions.AssignMeteredEvaluations != nil {
		actionsMap["assign_metered_evaluations"] = flattenAssignMeteredEvaluations(actions.AssignMeteredEvaluations)
	}
	if actions.AssignMeteredAssignmentByAgent != nil {
		actionsMap["assign_metered_assignment_by_agent"] = flattenAssignMeteredAssignmentByAgent(actions.AssignMeteredAssignmentByAgent)
	}
	if actions.AssignCalibrations != nil {
		actionsMap["assign_calibrations"] = flattenAssignCalibrations(actions.AssignCalibrations)
	}
	if actions.AssignSurveys != nil {
		actionsMap["assign_surveys"] = flattenAssignSurveys(actions.AssignSurveys)
	}
	if actions.RetentionDuration != nil {
		actionsMap["retention_duration"] = flattenRetentionDuration(actions.RetentionDuration)
	}
	if actions.InitiateScreenRecording != nil {
		actionsMap["initiate_screen_recording"] = flattenInitiateScreenRecording(actions.InitiateScreenRecording)
	}
	if actions.MediaTranscriptions != nil {
		actionsMap["media_transcriptions"] = flattenMediaTranscriptions(actions.MediaTranscriptions)
	}
	if actions.IntegrationExport != nil {
		actionsMap["integration_export"] = flattenIntegrationExport(actions.IntegrationExport)
	}

	return []interface{}{actionsMap}
}

func buildTimeSlots(slots []interface{}) *[]platformclientv2.Timeslot {
	timeSlots := make([]platformclientv2.Timeslot, 0)

	for _, slot := range slots {
		slotMap, ok := slot.(map[string]interface{})
		if !ok {
			continue
		}
		startTime := slotMap["start_time"].(string)
		stopTime := slotMap["stop_time"].(string)
		day := slotMap["day"].(int)

		timeSlots = append(timeSlots, platformclientv2.Timeslot{
			StartTime: &startTime,
			StopTime:  &stopTime,
			Day:       &day,
		})
	}

	return &timeSlots
}

func flattenTimeSlots(slots *[]platformclientv2.Timeslot) []interface{} {
	if slots == nil {
		return nil
	}

	slotList := []interface{}{}

	for _, slot := range *slots {
		slotMap := make(map[string]interface{})
		if slot.StartTime != nil {
			slotMap["start_time"] = *slot.StartTime
		}
		if slot.StopTime != nil {
			slotMap["stop_time"] = *slot.StopTime
		}
		if slot.Day != nil {
			slotMap["day"] = *slot.Day
		}

		slotList = append(slotList, slotMap)
	}

	return slotList
}

func buildTimeAllowed(timeAllowed []interface{}) *platformclientv2.Timeallowed {
	if timeAllowed == nil || len(timeAllowed) <= 0 {
		return nil
	}

	timeAllowedMap, ok := timeAllowed[0].(map[string]interface{})
	if !ok {
		return nil
	}

	timeZoneId := timeAllowedMap["time_zone_id"].(string)
	empty := timeAllowedMap["empty"].(bool)

	return &platformclientv2.Timeallowed{
		TimeSlots:  buildTimeSlots(timeAllowedMap["time_slots"].([]interface{})),
		TimeZoneId: &timeZoneId,
		Empty:      &empty,
	}
}

func flattenTimeAllowed(timeAllowed *platformclientv2.Timeallowed) []interface{} {
	if timeAllowed == nil {
		return nil
	}

	timeAllowedMap := make(map[string]interface{})
	if timeAllowed.TimeSlots != nil {
		timeAllowedMap["time_slots"] = flattenTimeSlots(timeAllowed.TimeSlots)
	}
	if timeAllowed.TimeZoneId != nil {
		timeAllowedMap["time_zone_id"] = *timeAllowed.TimeZoneId
	}
	if timeAllowed.Empty != nil {
		timeAllowedMap["empty"] = *timeAllowed.Empty
	}

	return []interface{}{timeAllowedMap}
}

func buildDurationCondition(durationCondition []interface{}) *platformclientv2.Durationcondition {
	if durationCondition == nil || len(durationCondition) <= 0 {
		return nil
	}

	durationConditionMap, ok := durationCondition[0].(map[string]interface{})
	if !ok {
		return nil
	}

	durationTarget := durationConditionMap["duration_target"].(string)
	durationOperator := durationConditionMap["duration_operator"].(string)
	durationRange := durationConditionMap["duration_range"].(string)
	durationMode := durationConditionMap["duration_mode"].(string)

	return &platformclientv2.Durationcondition{
		DurationTarget:   &durationTarget,
		DurationOperator: &durationOperator,
		DurationRange:    &durationRange,
		DurationMode:     &durationMode,
	}
}

func flattenDurationCondition(durationCondition *platformclientv2.Durationcondition) []interface{} {
	if durationCondition == nil {
		return nil
	}

	durationConditionMap := make(map[string]interface{})
	if durationCondition.DurationTarget != nil {
		durationConditionMap["duration_target"] = *durationCondition.DurationTarget
	}
	if durationCondition.DurationOperator != nil {
		durationConditionMap["duration_operator"] = *durationCondition.DurationOperator
	}
	if durationCondition.DurationRange != nil {
		durationConditionMap["duration_range"] = *durationCondition.DurationRange
	}
	if durationCondition.DurationMode != nil {
		durationConditionMap["duration_mode"] = *durationCondition.DurationMode
	}

	return []interface{}{durationConditionMap}
}

func buildCallMediaPolicyConditions(callMediaPolicyConditions []interface{}) *platformclientv2.Callmediapolicyconditions {
	if callMediaPolicyConditions == nil || len(callMediaPolicyConditions) <= 0 {
		return nil
	}

	conditionsMap, ok := callMediaPolicyConditions[0].(map[string]interface{})
	if !ok {
		return nil
	}

	directions := make([]string, 0)
	for _, v := range conditionsMap["directions"].([]interface{}) {
		direction := fmt.Sprintf("%v", v)
		directions = append(directions, direction)
	}

	dateRanges := make([]string, 0)
	for _, v := range conditionsMap["date_ranges"].([]interface{}) {
		dateRange := fmt.Sprintf("%v", v)
		dateRanges = append(dateRanges, dateRange)
	}

	forUserIds := conditionsMap["for_user_ids"].([]interface{})
	idStrings := make([]string, 0)
	for _, id := range forUserIds {
		idStrings = append(idStrings, fmt.Sprintf("%v", id))
	}

	forUsers := make([]platformclientv2.User, 0)
	for _, id := range idStrings {
		userId := id
		forUsers = append(forUsers, platformclientv2.User{Id: &userId})
	}

	wrapupCodeIds := conditionsMap["wrapup_code_ids"].([]interface{})
	wrapupCodeIdStrings := make([]string, 0)
	for _, id := range wrapupCodeIds {
		wrapupCodeIdStrings = append(wrapupCodeIdStrings, fmt.Sprintf("%v", id))
	}

	wrapupCodes := make([]platformclientv2.Wrapupcode, 0)
	for _, id := range wrapupCodeIdStrings {
		wrapupId := id
		wrapupCodes = append(wrapupCodes, platformclientv2.Wrapupcode{Id: &wrapupId})
	}

	languageIds := conditionsMap["language_ids"].([]interface{})
	languageIdStrings := make([]string, 0)
	for _, id := range languageIds {
		languageIdStrings = append(languageIdStrings, fmt.Sprintf("%v", id))
	}

	languages := make([]platformclientv2.Language, 0)
	for _, id := range languageIdStrings {
		languageId := id
		languages = append(languages, platformclientv2.Language{Id: &languageId})
	}

	forQueueIds := conditionsMap["for_queue_ids"].([]interface{})
	queueIdStrings := make([]string, 0)
	for _, id := range forQueueIds {
		queueIdStrings = append(queueIdStrings, fmt.Sprintf("%v", id))
	}

	forQueues := make([]platformclientv2.Queue, 0)
	for _, id := range queueIdStrings {
		queueId := id
		forQueues = append(forQueues, platformclientv2.Queue{Id: &queueId})
	}

	return &platformclientv2.Callmediapolicyconditions{
		ForUsers:    &forUsers,
		DateRanges:  &dateRanges,
		ForQueues:   &forQueues,
		WrapupCodes: &wrapupCodes,
		Languages:   &languages,
		TimeAllowed: buildTimeAllowed(conditionsMap["time_allowed"].([]interface{})),
		Directions:  &directions,
		Duration:    buildDurationCondition(conditionsMap["duration"].([]interface{})),
	}
}

func flattenCallMediaPolicyConditions(conditions *platformclientv2.Callmediapolicyconditions) []interface{} {
	if conditions == nil {
		return nil
	}

	conditionsMap := make(map[string]interface{})
	if conditions.ForUsers != nil {
		userIds := make([]string, 0)
		for _, user := range *conditions.ForUsers {
			userIds = append(userIds, *user.Id)
		}
		conditionsMap["for_user_ids"] = userIds
	}
	if conditions.DateRanges != nil {
		conditionsMap["date_ranges"] = *conditions.DateRanges
	}
	if conditions.Directions != nil {
		conditionsMap["directions"] = *conditions.Directions
	}
	if conditions.ForQueues != nil {
		queueIds := make([]string, 0)
		for _, queue := range *conditions.ForQueues {
			queueIds = append(queueIds, *queue.Id)
		}
		conditionsMap["for_queue_ids"] = queueIds
	}
	if conditions.WrapupCodes != nil {
		wrapupCodeIds := make([]string, 0)
		for _, code := range *conditions.WrapupCodes {
			wrapupCodeIds = append(wrapupCodeIds, *code.Id)
		}
		conditionsMap["wrapup_code_ids"] = wrapupCodeIds
	}
	if conditions.Languages != nil {
		languageIds := make([]string, 0)
		for _, code := range *conditions.Languages {
			languageIds = append(languageIds, *code.Id)
		}
		conditionsMap["language_ids"] = languageIds
	}
	if conditions.TimeAllowed != nil {
		conditionsMap["time_allowed"] = flattenTimeAllowed(conditions.TimeAllowed)
	}

	return []interface{}{conditionsMap}
}

func buildChatMediaPolicyConditions(chatMediaPolicyConditions []interface{}) *platformclientv2.Chatmediapolicyconditions {
	if chatMediaPolicyConditions == nil || len(chatMediaPolicyConditions) <= 0 {
		return nil
	}

	conditionsMap, ok := chatMediaPolicyConditions[0].(map[string]interface{})
	if !ok {
		return nil
	}

	dateRanges := make([]string, 0)
	for _, v := range conditionsMap["date_ranges"].([]interface{}) {
		dateRange := fmt.Sprintf("%v", v)
		dateRanges = append(dateRanges, dateRange)
	}

	forUserIds := conditionsMap["for_user_ids"].([]interface{})
	idStrings := make([]string, 0)
	for _, id := range forUserIds {
		idStrings = append(idStrings, fmt.Sprintf("%v", id))
	}

	forUsers := make([]platformclientv2.User, 0)
	for _, id := range idStrings {
		userId := id
		forUsers = append(forUsers, platformclientv2.User{Id: &userId})
	}

	wrapupCodeIds := conditionsMap["wrapup_code_ids"].([]interface{})
	wrapupCodeIdStrings := make([]string, 0)
	for _, id := range wrapupCodeIds {
		wrapupCodeIdStrings = append(wrapupCodeIdStrings, fmt.Sprintf("%v", id))
	}

	wrapupCodes := make([]platformclientv2.Wrapupcode, 0)
	for _, id := range wrapupCodeIdStrings {
		wrapupId := id
		wrapupCodes = append(wrapupCodes, platformclientv2.Wrapupcode{Id: &wrapupId})
	}

	languageIds := conditionsMap["language_ids"].([]interface{})
	languageIdStrings := make([]string, 0)
	for _, id := range languageIds {
		languageIdStrings = append(languageIdStrings, fmt.Sprintf("%v", id))
	}

	languages := make([]platformclientv2.Language, 0)
	for _, id := range languageIdStrings {
		languageId := id
		languages = append(languages, platformclientv2.Language{Id: &languageId})
	}

	forQueueIds := conditionsMap["for_queue_ids"].([]interface{})
	queueIdStrings := make([]string, 0)
	for _, id := range forQueueIds {
		queueIdStrings = append(queueIdStrings, fmt.Sprintf("%v", id))
	}

	forQueues := make([]platformclientv2.Queue, 0)
	for _, id := range queueIdStrings {
		queueId := id
		forQueues = append(forQueues, platformclientv2.Queue{Id: &queueId})
	}

	return &platformclientv2.Chatmediapolicyconditions{
		ForUsers:    &forUsers,
		DateRanges:  &dateRanges,
		ForQueues:   &forQueues,
		WrapupCodes: &wrapupCodes,
		Languages:   &languages,
		TimeAllowed: buildTimeAllowed(conditionsMap["time_allowed"].([]interface{})),
		Duration:    buildDurationCondition(conditionsMap["duration"].([]interface{})),
	}
}

func flattenChatMediaPolicyConditions(conditions *platformclientv2.Chatmediapolicyconditions) []interface{} {
	if conditions == nil {
		return nil
	}

	conditionsMap := make(map[string]interface{})
	if conditions.ForUsers != nil {
		userIds := make([]string, 0)
		for _, user := range *conditions.ForUsers {
			userIds = append(userIds, *user.Id)
		}
		conditionsMap["for_user_ids"] = userIds
	}
	if conditions.DateRanges != nil {
		conditionsMap["date_ranges"] = *conditions.DateRanges
	}
	if conditions.ForQueues != nil {
		queueIds := make([]string, 0)
		for _, queue := range *conditions.ForQueues {
			queueIds = append(queueIds, *queue.Id)
		}
		conditionsMap["for_queue_ids"] = queueIds
	}
	if conditions.WrapupCodes != nil {
		wrapupCodeIds := make([]string, 0)
		for _, code := range *conditions.WrapupCodes {
			wrapupCodeIds = append(wrapupCodeIds, *code.Id)
		}
		conditionsMap["wrapup_code_ids"] = wrapupCodeIds
	}
	if conditions.Languages != nil {
		languageIds := make([]string, 0)
		for _, code := range *conditions.Languages {
			languageIds = append(languageIds, *code.Id)
		}
		conditionsMap["language_ids"] = languageIds
	}
	if conditions.TimeAllowed != nil {
		conditionsMap["time_allowed"] = flattenTimeAllowed(conditions.TimeAllowed)
	}
	if conditions.Duration != nil {
		conditionsMap["duration"] = flattenDurationCondition(conditions.Duration)
	}

	return []interface{}{conditionsMap}
}

func buildEmailMediaPolicyConditions(emailMediaPolicyConditions []interface{}) *platformclientv2.Emailmediapolicyconditions {
	if emailMediaPolicyConditions == nil || len(emailMediaPolicyConditions) <= 0 {
		return nil
	}

	conditionsMap, ok := emailMediaPolicyConditions[0].(map[string]interface{})
	if !ok {
		return nil
	}

	dateRanges := make([]string, 0)
	for _, v := range conditionsMap["date_ranges"].([]interface{}) {
		dateRange := fmt.Sprintf("%v", v)
		dateRanges = append(dateRanges, dateRange)
	}

	forUserIds := conditionsMap["for_user_ids"].([]interface{})
	idStrings := make([]string, 0)
	for _, id := range forUserIds {
		idStrings = append(idStrings, fmt.Sprintf("%v", id))
	}

	forUsers := make([]platformclientv2.User, 0)
	for _, id := range idStrings {
		userId := id
		forUsers = append(forUsers, platformclientv2.User{Id: &userId})
	}

	wrapupCodeIds := conditionsMap["wrapup_code_ids"].([]interface{})
	wrapupCodeIdStrings := make([]string, 0)
	for _, id := range wrapupCodeIds {
		wrapupCodeIdStrings = append(wrapupCodeIdStrings, fmt.Sprintf("%v", id))
	}

	wrapupCodes := make([]platformclientv2.Wrapupcode, 0)
	for _, id := range wrapupCodeIdStrings {
		wrapupId := id
		wrapupCodes = append(wrapupCodes, platformclientv2.Wrapupcode{Id: &wrapupId})
	}

	languageIds := conditionsMap["language_ids"].([]interface{})
	languageIdStrings := make([]string, 0)
	for _, id := range languageIds {
		languageIdStrings = append(languageIdStrings, fmt.Sprintf("%v", id))
	}

	languages := make([]platformclientv2.Language, 0)
	for _, id := range languageIdStrings {
		languageId := id
		languages = append(languages, platformclientv2.Language{Id: &languageId})
	}

	forQueueIds := conditionsMap["for_queue_ids"].([]interface{})
	queueIdStrings := make([]string, 0)
	for _, id := range forQueueIds {
		queueIdStrings = append(queueIdStrings, fmt.Sprintf("%v", id))
	}

	forQueues := make([]platformclientv2.Queue, 0)
	for _, id := range queueIdStrings {
		queueId := id
		forQueues = append(forQueues, platformclientv2.Queue{Id: &queueId})
	}

	return &platformclientv2.Emailmediapolicyconditions{
		ForUsers:    &forUsers,
		DateRanges:  &dateRanges,
		ForQueues:   &forQueues,
		WrapupCodes: &wrapupCodes,
		Languages:   &languages,
		TimeAllowed: buildTimeAllowed(conditionsMap["time_allowed"].([]interface{})),
	}
}

func flattenEmailMediaPolicyConditions(conditions *platformclientv2.Emailmediapolicyconditions) []interface{} {
	if conditions == nil {
		return nil
	}

	conditionsMap := make(map[string]interface{})
	if conditions.ForUsers != nil {
		userIds := make([]string, 0)
		for _, user := range *conditions.ForUsers {
			userIds = append(userIds, *user.Id)
		}
		conditionsMap["for_user_ids"] = userIds
	}
	if conditions.DateRanges != nil {
		conditionsMap["date_ranges"] = *conditions.DateRanges
	}
	if conditions.ForQueues != nil {
		queueIds := make([]string, 0)
		for _, queue := range *conditions.ForQueues {
			queueIds = append(queueIds, *queue.Id)
		}
		conditionsMap["for_queue_ids"] = queueIds
	}
	if conditions.WrapupCodes != nil {
		wrapupCodeIds := make([]string, 0)
		for _, code := range *conditions.WrapupCodes {
			wrapupCodeIds = append(wrapupCodeIds, *code.Id)
		}
		conditionsMap["wrapup_code_ids"] = wrapupCodeIds
	}
	if conditions.Languages != nil {
		languageIds := make([]string, 0)
		for _, code := range *conditions.Languages {
			languageIds = append(languageIds, *code.Id)
		}
		conditionsMap["language_ids"] = languageIds
	}
	if conditions.TimeAllowed != nil {
		conditionsMap["time_allowed"] = flattenTimeAllowed(conditions.TimeAllowed)
	}

	return []interface{}{conditionsMap}
}

func buildMessageMediaPolicyConditions(messageMediaPolicyConditions []interface{}) *platformclientv2.Messagemediapolicyconditions {
	if messageMediaPolicyConditions == nil || len(messageMediaPolicyConditions) <= 0 {
		return nil
	}

	conditionsMap, ok := messageMediaPolicyConditions[0].(map[string]interface{})
	if !ok {
		return nil
	}

	dateRanges := make([]string, 0)
	for _, v := range conditionsMap["date_ranges"].([]interface{}) {
		dateRange := fmt.Sprintf("%v", v)
		dateRanges = append(dateRanges, dateRange)
	}

	forUserIds := conditionsMap["for_user_ids"].([]interface{})
	idStrings := make([]string, 0)
	for _, id := range forUserIds {
		idStrings = append(idStrings, fmt.Sprintf("%v", id))
	}

	forUsers := make([]platformclientv2.User, 0)
	for _, id := range idStrings {
		userId := id
		forUsers = append(forUsers, platformclientv2.User{Id: &userId})
	}

	wrapupCodeIds := conditionsMap["wrapup_code_ids"].([]interface{})
	wrapupCodeIdStrings := make([]string, 0)
	for _, id := range wrapupCodeIds {
		wrapupCodeIdStrings = append(wrapupCodeIdStrings, fmt.Sprintf("%v", id))
	}

	wrapupCodes := make([]platformclientv2.Wrapupcode, 0)
	for _, id := range wrapupCodeIdStrings {
		wrapupId := id
		wrapupCodes = append(wrapupCodes, platformclientv2.Wrapupcode{Id: &wrapupId})
	}

	languageIds := conditionsMap["language_ids"].([]interface{})
	languageIdStrings := make([]string, 0)
	for _, id := range languageIds {
		languageIdStrings = append(languageIdStrings, fmt.Sprintf("%v", id))
	}

	languages := make([]platformclientv2.Language, 0)
	for _, id := range languageIdStrings {
		languageId := id
		languages = append(languages, platformclientv2.Language{Id: &languageId})
	}

	forQueueIds := conditionsMap["for_queue_ids"].([]interface{})
	queueIdStrings := make([]string, 0)
	for _, id := range forQueueIds {
		queueIdStrings = append(queueIdStrings, fmt.Sprintf("%v", id))
	}

	forQueues := make([]platformclientv2.Queue, 0)
	for _, id := range queueIdStrings {
		queueId := id
		forQueues = append(forQueues, platformclientv2.Queue{Id: &queueId})
	}

	return &platformclientv2.Messagemediapolicyconditions{
		ForUsers:    &forUsers,
		DateRanges:  &dateRanges,
		ForQueues:   &forQueues,
		WrapupCodes: &wrapupCodes,
		Languages:   &languages,
		TimeAllowed: buildTimeAllowed(conditionsMap["time_allowed"].([]interface{})),
	}
}

func flattenMessageMediaPolicyConditions(conditions *platformclientv2.Messagemediapolicyconditions) []interface{} {
	if conditions == nil {
		return nil
	}

	conditionsMap := make(map[string]interface{})
	if conditions.ForUsers != nil {
		userIds := make([]string, 0)
		for _, user := range *conditions.ForUsers {
			userIds = append(userIds, *user.Id)
		}
		conditionsMap["for_user_ids"] = userIds
	}
	if conditions.DateRanges != nil {
		conditionsMap["date_ranges"] = *conditions.DateRanges
	}
	if conditions.ForQueues != nil {
		queueIds := make([]string, 0)
		for _, queue := range *conditions.ForQueues {
			queueIds = append(queueIds, *queue.Id)
		}
		conditionsMap["for_queue_ids"] = queueIds
	}
	if conditions.WrapupCodes != nil {
		wrapupCodeIds := make([]string, 0)
		for _, code := range *conditions.WrapupCodes {
			wrapupCodeIds = append(wrapupCodeIds, *code.Id)
		}
		conditionsMap["wrapup_code_ids"] = wrapupCodeIds
	}
	if conditions.Languages != nil {
		languageIds := make([]string, 0)
		for _, code := range *conditions.Languages {
			languageIds = append(languageIds, *code.Id)
		}
		conditionsMap["language_ids"] = languageIds
	}
	if conditions.TimeAllowed != nil {
		conditionsMap["time_allowed"] = flattenTimeAllowed(conditions.TimeAllowed)
	}

	return []interface{}{conditionsMap}
}

func buildCallMediaPolicy(callMediaPolicy []interface{}) *platformclientv2.Callmediapolicy {
	if callMediaPolicy == nil || len(callMediaPolicy) <= 0 {
		return nil
	}

	policyMap, ok := callMediaPolicy[0].(map[string]interface{})
	if !ok {
		return nil
	}
	return &platformclientv2.Callmediapolicy{
		Actions:    buildPolicyActions(policyMap["actions"].([]interface{})),
		Conditions: buildCallMediaPolicyConditions(policyMap["conditions"].([]interface{})),
	}
}

func flattenCallMediaPolicy(chatMediaPolicy *platformclientv2.Callmediapolicy) []interface{} {
	if chatMediaPolicy == nil {
		return nil
	}

	chatMediaPolicyMap := make(map[string]interface{})
	if chatMediaPolicy.Actions != nil {
		chatMediaPolicyMap["actions"] = flattenPolicyActions(chatMediaPolicy.Actions)
	}
	if chatMediaPolicy.Conditions != nil {
		chatMediaPolicyMap["conditions"] = flattenCallMediaPolicyConditions(chatMediaPolicy.Conditions)
	}

	return []interface{}{chatMediaPolicyMap}
}

func buildChatMediaPolicy(chatMediaPolicy []interface{}) *platformclientv2.Chatmediapolicy {
	if chatMediaPolicy == nil || len(chatMediaPolicy) <= 0 {
		return nil
	}

	policyMap, ok := chatMediaPolicy[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &platformclientv2.Chatmediapolicy{
		Actions:    buildPolicyActions(policyMap["actions"].([]interface{})),
		Conditions: buildChatMediaPolicyConditions(policyMap["conditions"].([]interface{})),
	}
}

func flattenChatMediaPolicy(chatMediaPolicy *platformclientv2.Chatmediapolicy) []interface{} {
	if chatMediaPolicy == nil {
		return nil
	}

	chatMediaPolicyMap := make(map[string]interface{})
	if chatMediaPolicy.Actions != nil {
		chatMediaPolicyMap["actions"] = flattenPolicyActions(chatMediaPolicy.Actions)
	}
	if chatMediaPolicy.Conditions != nil {
		chatMediaPolicyMap["conditions"] = flattenChatMediaPolicyConditions(chatMediaPolicy.Conditions)
	}

	return []interface{}{chatMediaPolicyMap}
}

func buildEmailMediaPolicy(emailMediaPolicy []interface{}) *platformclientv2.Emailmediapolicy {
	if emailMediaPolicy == nil || len(emailMediaPolicy) <= 0 {
		return nil
	}

	policyMap, ok := emailMediaPolicy[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &platformclientv2.Emailmediapolicy{
		Actions:    buildPolicyActions(policyMap["actions"].([]interface{})),
		Conditions: buildEmailMediaPolicyConditions(policyMap["conditions"].([]interface{})),
	}
}

func flattenEmailMediaPolicy(emailMediaPolicy *platformclientv2.Emailmediapolicy) []interface{} {
	if emailMediaPolicy == nil {
		return nil
	}

	emailMediaPolicyMap := make(map[string]interface{})
	if emailMediaPolicy.Actions != nil {
		emailMediaPolicyMap["actions"] = flattenPolicyActions(emailMediaPolicy.Actions)
	}
	if emailMediaPolicy.Conditions != nil {
		emailMediaPolicyMap["conditions"] = flattenEmailMediaPolicyConditions(emailMediaPolicy.Conditions)
	}

	return []interface{}{emailMediaPolicyMap}
}

func buildMessageMediaPolicy(messageMediaPolicy []interface{}) *platformclientv2.Messagemediapolicy {
	if messageMediaPolicy == nil || len(messageMediaPolicy) <= 0 {
		return nil
	}

	policyMap, ok := messageMediaPolicy[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &platformclientv2.Messagemediapolicy{
		Actions:    buildPolicyActions(policyMap["actions"].([]interface{})),
		Conditions: buildMessageMediaPolicyConditions(policyMap["conditions"].([]interface{})),
	}
}

func flattenMessageMediaPolicy(messageMediaPolicy *platformclientv2.Messagemediapolicy) []interface{} {
	if messageMediaPolicy == nil {
		return nil
	}

	messageMediaPolicyMap := make(map[string]interface{})
	if messageMediaPolicy.Actions != nil {
		messageMediaPolicyMap["actions"] = flattenPolicyActions(messageMediaPolicy.Actions)
	}
	if messageMediaPolicy.Conditions != nil {
		messageMediaPolicyMap["conditions"] = flattenMessageMediaPolicyConditions(messageMediaPolicy.Conditions)
	}

	return []interface{}{messageMediaPolicyMap}
}

func buildMediaPolicies(d *schema.ResourceData) *platformclientv2.Mediapolicies {
	sdkMediaPolicies := platformclientv2.Mediapolicies{}

	if mediaPolicies, ok := d.Get("media_policies").([]interface{}); ok && len(mediaPolicies) > 0 {
		mediaPoliciesMap, ok := mediaPolicies[0].(map[string]interface{})
		if !ok {
			return nil
		}
		if callPolicy := mediaPoliciesMap["call_policy"]; callPolicy != nil {
			sdkMediaPolicies.CallPolicy = buildCallMediaPolicy(callPolicy.([]interface{}))
		}

		if chatPolicy := mediaPoliciesMap["chat_policy"]; chatPolicy != nil {
			sdkMediaPolicies.ChatPolicy = buildChatMediaPolicy(chatPolicy.([]interface{}))
		}

		if emailPolicy := mediaPoliciesMap["email_policy"]; emailPolicy != nil {
			sdkMediaPolicies.EmailPolicy = buildEmailMediaPolicy(emailPolicy.([]interface{}))
		}

		if messagePolicy := mediaPoliciesMap["message_policy"]; messagePolicy != nil {
			sdkMediaPolicies.MessagePolicy = buildMessageMediaPolicy(messagePolicy.([]interface{}))
		}
	}

	return &sdkMediaPolicies
}

func flattenMediaPolicies(mediaPolicies *platformclientv2.Mediapolicies) []interface{} {
	if mediaPolicies == nil {
		return nil
	}

	mediaPoliciesMap := make(map[string]interface{})
	if mediaPolicies.CallPolicy != nil {
		mediaPoliciesMap["call_policy"] = flattenCallMediaPolicy(mediaPolicies.CallPolicy)
	}
	if mediaPolicies.ChatPolicy != nil {
		mediaPoliciesMap["chat_policy"] = flattenChatMediaPolicy(mediaPolicies.ChatPolicy)
	}
	if mediaPolicies.EmailPolicy != nil {
		mediaPoliciesMap["email_policy"] = flattenEmailMediaPolicy(mediaPolicies.EmailPolicy)
	}
	if mediaPolicies.MessagePolicy != nil {
		mediaPoliciesMap["message_policy"] = flattenMessageMediaPolicy(mediaPolicies.MessagePolicy)
	}

	return []interface{}{mediaPoliciesMap}
}

func buildConditions(d *schema.ResourceData) *platformclientv2.Policyconditions {
	if conditions, ok := d.Get("conditions").([]interface{}); ok && len(conditions) > 0 {
		conditionsMap, ok := conditions[0].(map[string]interface{})
		if !ok {
			return nil
		}

		directions := make([]string, 0)
		for _, v := range conditionsMap["directions"].([]interface{}) {
			direction := fmt.Sprintf("%v", v)
			directions = append(directions, direction)
		}

		dateRanges := make([]string, 0)
		for _, v := range conditionsMap["date_ranges"].([]interface{}) {
			dateRange := fmt.Sprintf("%v", v)
			dateRanges = append(dateRanges, dateRange)
		}

		mediaTypes := make([]string, 0)
		for _, v := range conditionsMap["media_types"].([]interface{}) {
			mediaType := fmt.Sprintf("%v", v)
			mediaTypes = append(mediaTypes, mediaType)
		}

		forUserIds := conditionsMap["for_user_ids"].([]interface{})
		idStrings := make([]string, 0)
		for _, id := range forUserIds {
			idStrings = append(idStrings, fmt.Sprintf("%v", id))
		}

		forUsers := make([]platformclientv2.User, 0)
		for _, id := range idStrings {
			userId := id
			forUsers = append(forUsers, platformclientv2.User{Id: &userId})
		}

		wrapupCodeIds := conditionsMap["wrapup_code_ids"].([]interface{})
		wrapupCodeIdStrings := make([]string, 0)
		for _, id := range wrapupCodeIds {
			wrapupCodeIdStrings = append(wrapupCodeIdStrings, fmt.Sprintf("%v", id))
		}

		wrapupCodes := make([]platformclientv2.Wrapupcode, 0)
		for _, id := range wrapupCodeIdStrings {
			wrapupId := id
			wrapupCodes = append(wrapupCodes, platformclientv2.Wrapupcode{Id: &wrapupId})
		}

		forQueueIds := conditionsMap["for_queue_ids"].([]interface{})
		queueIdStrings := make([]string, 0)
		for _, id := range forQueueIds {
			queueIdStrings = append(queueIdStrings, fmt.Sprintf("%v", id))
		}

		forQueues := make([]platformclientv2.Queue, 0)
		for _, id := range queueIdStrings {
			queueId := id
			forQueues = append(forQueues, platformclientv2.Queue{Id: &queueId})
		}

		return &platformclientv2.Policyconditions{
			ForUsers:    &forUsers,
			Directions:  &directions,
			DateRanges:  &dateRanges,
			MediaTypes:  &mediaTypes,
			ForQueues:   &forQueues,
			Duration:    buildDurationCondition(conditionsMap["duration"].([]interface{})),
			WrapupCodes: &wrapupCodes,
			TimeAllowed: buildTimeAllowed(conditionsMap["time_allowed"].([]interface{})),
		}
	}

	return nil
}

func flattenConditions(conditions *platformclientv2.Policyconditions) []interface{} {
	if conditions == nil || reflect.DeepEqual(platformclientv2.Policyconditions{}, *conditions) {
		return nil
	}

	conditionsMap := make(map[string]interface{})
	if conditions.ForUsers != nil {
		userIds := make([]string, 0)
		for _, user := range *conditions.ForUsers {
			userIds = append(userIds, *user.Id)
		}
		conditionsMap["for_user_ids"] = userIds
	}
	if conditions.Directions != nil {
		conditionsMap["directions"] = *conditions.Directions
	}
	if conditions.DateRanges != nil {
		conditionsMap["date_ranges"] = *conditions.DateRanges
	}
	if conditions.MediaTypes != nil {
		conditionsMap["media_types"] = *conditions.MediaTypes
	}
	if conditions.ForQueues != nil {
		queueIds := make([]string, 0)
		for _, queue := range *conditions.ForQueues {
			queueIds = append(queueIds, *queue.Id)
		}
		conditionsMap["for_queue_ids"] = queueIds
	}
	if conditions.Duration != nil {
		conditionsMap["duration"] = flattenDurationCondition(conditions.Duration)
	}
	if conditions.WrapupCodes != nil {
		wrapupCodeIds := make([]string, 0)
		for _, code := range *conditions.WrapupCodes {
			wrapupCodeIds = append(wrapupCodeIds, *code.Id)
		}
		conditionsMap["wrapup_code_ids"] = wrapupCodeIds
	}
	if conditions.TimeAllowed != nil {
		conditionsMap["time_allowed"] = flattenTimeAllowed(conditions.TimeAllowed)
	}

	return []interface{}{conditionsMap}
}

func buildPolicyActions2(d *schema.ResourceData) *platformclientv2.Policyactions {

	if actions, ok := d.Get("actions").([]interface{}); ok && len(actions) > 0 {
		actionsMap, ok := actions[0].(map[string]interface{})
		if !ok {
			return nil
		}
		retainRecording := actionsMap["retain_recording"].(bool)
		deleteRecording := actionsMap["delete_recording"].(bool)
		alwaysDelete := actionsMap["always_delete"].(bool)

		return &platformclientv2.Policyactions{
			RetainRecording:                &retainRecording,
			DeleteRecording:                &deleteRecording,
			AlwaysDelete:                   &alwaysDelete,
			AssignEvaluations:              buildEvaluationAssignments(actionsMap["assign_evaluations"].([]interface{})),
			AssignMeteredEvaluations:       buildAssignMeteredEvaluations(actionsMap["assign_metered_evaluations"].([]interface{})),
			AssignMeteredAssignmentByAgent: buildAssignMeteredAssignmentByAgent(actionsMap["assign_metered_assignment_by_agent"].([]interface{})),
			AssignCalibrations:             buildAssignCalibrations(actionsMap["assign_calibrations"].([]interface{})),
			AssignSurveys:                  buildAssignSurveys(actionsMap["assign_surveys"].([]interface{})),
			RetentionDuration:              buildRetentionDuration(actionsMap["retention_duration"].([]interface{})),
			InitiateScreenRecording:        buildInitiateScreenRecording(actionsMap["initiate_screen_recording"].([]interface{})),
			MediaTranscriptions:            buildMediaTranscriptions(actionsMap["media_transcriptions"].([]interface{})),
			IntegrationExport:              buildIntegrationExport(actionsMap["integration_export"].([]interface{})),
		}
	}

	return nil
}

func buildUserParams(params []interface{}) *[]platformclientv2.Userparam {
	userParams := make([]platformclientv2.Userparam, 0)

	for _, param := range params {
		paramMap, ok := param.(map[string]interface{})
		if !ok {
			continue
		}
		key := paramMap["key"].(string)
		value := paramMap["value"].(string)

		userParams = append(userParams, platformclientv2.Userparam{
			Key:   &key,
			Value: &value,
		})
	}

	return &userParams
}

func flattenUserParams(params *[]platformclientv2.Userparam) []interface{} {
	if params == nil {
		return nil
	}

	paramList := []interface{}{}

	for _, param := range *params {
		paramMap := make(map[string]interface{})
		if param.Key != nil {
			paramMap["key"] = *param.Key
		}
		if param.Value != nil {
			paramMap["value"] = *param.Value
		}

		paramList = append(paramList, paramMap)
	}

	return paramList
}

func buildPolicyErrorMessages(messages []interface{}) *[]platformclientv2.Policyerrormessage {
	policyErrorMessages := make([]platformclientv2.Policyerrormessage, 0)

	for _, message := range messages {
		messageMap, ok := message.(map[string]interface{})
		if !ok {
			continue
		}
		statusCode := messageMap["status_code"].(int)
		userMessage := messageMap["user_message"]
		userParamsMessage := messageMap["user_params_message"].(string)
		errorCode := messageMap["error_code"].(string)
		correlationId := messageMap["correlation_id"].(string)
		insertDateString := messageMap["insert_date"].(string)

		temp := platformclientv2.Policyerrormessage{
			StatusCode:        &statusCode,
			UserMessage:       &userMessage,
			UserParamsMessage: &userParamsMessage,
			ErrorCode:         &errorCode,
			CorrelationId:     &correlationId,
			UserParams:        buildUserParams(messageMap["user_params"].([]interface{})),
		}

		insertDate, insertErr := time.Parse("2006-01-02T15:04:05-0700", insertDateString)
		if insertErr == nil {
			temp.InsertDate = &insertDate
		}

		policyErrorMessages = append(policyErrorMessages, temp)
	}

	return &policyErrorMessages
}

func flattenPolicyErrorMessages(errorMessages *[]platformclientv2.Policyerrormessage) []interface{} {
	if errorMessages == nil {
		return nil
	}

	errorMessageList := []interface{}{}

	for _, errorMessage := range *errorMessages {
		errorMessageMap := make(map[string]interface{})
		if errorMessage.StatusCode != nil {
			errorMessageMap["status_code"] = *errorMessage.StatusCode
		}
		if errorMessage.UserMessage != nil {
			errorMessageMap["user_message"] = *errorMessage.UserMessage
		}
		if errorMessage.UserParamsMessage != nil {
			errorMessageMap["user_params_message"] = *errorMessage.UserParamsMessage
		}
		if errorMessage.ErrorCode != nil {
			errorMessageMap["error_code"] = *errorMessage.ErrorCode
		}
		if errorMessage.CorrelationId != nil {
			errorMessageMap["correlation_id"] = *errorMessage.CorrelationId
		}
		if errorMessage.InsertDate != nil && len(errorMessage.InsertDate.String()) > 0 {
			temp := *errorMessage.InsertDate
			errorMessageMap["insert_date"] = temp.String()
		}
		if errorMessage.UserParams != nil {
			errorMessageMap["user_params"] = flattenUserParams(errorMessage.UserParams)
		}

		errorMessageList = append(errorMessageList, errorMessageMap)
	}

	return errorMessageList
}

func buildPolicyErrors(d *schema.ResourceData) *platformclientv2.Policyerrors {
	if errors, ok := d.GetOk("policy_errors"); ok {
		if errorsList, ok := errors.([]interface{}); ok || len(errorsList) > 0 {
			errorsMap, ok := errorsList[0].(map[string]interface{})
			if !ok {
				return nil
			}
			return &platformclientv2.Policyerrors{
				PolicyErrorMessages: buildPolicyErrorMessages(errorsMap["policy_error_messages"].([]interface{})),
			}
		}
	}

	return nil
}

func flattenPolicyErrors(policyErrors *platformclientv2.Policyerrors) []interface{} {
	if policyErrors == nil {
		return nil
	}

	policyErrorsMap := make(map[string]interface{})
	if policyErrors.PolicyErrorMessages != nil {
		policyErrorsMap["policy_error_messages"] = flattenPolicyErrorMessages(policyErrors.PolicyErrorMessages)
	}

	return []interface{}{policyErrorsMap}
}

func readMediaRetentionPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	recordingAPI := platformclientv2.NewRecordingApiWithConfig(sdkConfig)

	log.Printf("Reading media retention policy %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		retentionPolicy, resp, getErr := recordingAPI.GetRecordingMediaretentionpolicy(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read media retention policy %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read media retention policy %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceSurveyForm())
		if retentionPolicy.Name != nil {
			d.Set("name", *retentionPolicy.Name)
		}
		if retentionPolicy.Order != nil {
			d.Set("order", *retentionPolicy.Order)
		}
		if retentionPolicy.Description != nil {
			d.Set("description", *retentionPolicy.Description)
		}
		if retentionPolicy.Enabled != nil {
			d.Set("enabled", *retentionPolicy.Enabled)
		}
		if retentionPolicy.MediaPolicies != nil {
			d.Set("media_policies", flattenMediaPolicies(retentionPolicy.MediaPolicies))
		}
		if retentionPolicy.Conditions != nil {
			d.Set("conditions", flattenConditions(retentionPolicy.Conditions))
		}
		if retentionPolicy.Actions != nil {
			d.Set("actions", flattenPolicyActions(retentionPolicy.Actions))
		}
		if retentionPolicy.PolicyErrors != nil {
			d.Set("policy_errors", flattenPolicyErrors(retentionPolicy.PolicyErrors))
		}

		return cc.CheckState()
	})
}

func createMediaRetentionPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	order := d.Get("order").(int)
	description := d.Get("description").(string)
	enabled := d.Get("enabled").(bool)
	mediaPolicies := buildMediaPolicies(d)
	conditions := buildConditions(d)
	actions := buildPolicyActions2(d)
	policyErrors := buildPolicyErrors(d)
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	recordingAPI := platformclientv2.NewRecordingApiWithConfig(sdkConfig)

	reqBody := platformclientv2.Policycreate{
		Name:          &name,
		Order:         &order,
		Description:   &description,
		Enabled:       &enabled,
		MediaPolicies: mediaPolicies,
		Conditions:    conditions,
		Actions:       actions,
		PolicyErrors:  policyErrors,
	}

	log.Printf("Creating media retention policy %s", name)
	policy, apiResponse, err := recordingAPI.PostRecordingMediaretentionpolicies(reqBody)
	log.Printf("Media retention policy creation status %#v", apiResponse.Status)

	if err != nil {
		return diag.Errorf("Failed to create media retention policy %s: %s", name, err)
	}

	// Make sure form is properly created
	policyId := policy.Id
	d.SetId(*policyId)
	log.Printf("Created media retention policy %s %s", name, *policy.Id)
	return readMediaRetentionPolicy(ctx, d, meta)
}

func updateMediaRetentionPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	order := d.Get("order").(int)
	description := d.Get("description").(string)
	enabled := d.Get("enabled").(bool)

	mediaPolicies := buildMediaPolicies(d)
	conditions := buildConditions(d)
	actions := buildPolicyActions2(d)
	policyErrors := buildPolicyErrors(d)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	recordingAPI := platformclientv2.NewRecordingApiWithConfig(sdkConfig)

	reqBody := platformclientv2.Policy{
		Name:          &name,
		Order:         &order,
		Description:   &description,
		Enabled:       &enabled,
		MediaPolicies: mediaPolicies,
		Conditions:    conditions,
		Actions:       actions,
		PolicyErrors:  policyErrors,
	}

	log.Printf("Updating media retention policy %s", name)
	policy, _, err := recordingAPI.PutRecordingMediaretentionpolicy(d.Id(), reqBody)
	if err != nil {
		return diag.Errorf("Failed to update media retention policy %s: %s", name, err)
	}

	log.Printf("Updated media retention policy %s %s", name, *policy.Id)
	return readMediaRetentionPolicy(ctx, d, meta)
}

func MediaRetentionPolicyExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllMediaRetentionPolicies),
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

func deleteMediaRetentionPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	recordingAPI := platformclientv2.NewRecordingApiWithConfig(sdkConfig)

	log.Printf("Deleting media retention policy %s", name)
	if _, err := recordingAPI.DeleteRecordingMediaretentionpolicy(d.Id()); err != nil {
		return diag.Errorf("Failed to delete media retention policy %s: %s", name, err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := recordingAPI.GetRecordingMediaretentionpolicy(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// media retention policy deleted
				log.Printf("Deleted media retention policy %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting media retention policy %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("media retention policy %s still exists", d.Id()))
	})
}
