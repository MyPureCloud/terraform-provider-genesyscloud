package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v67/platformclientv2"
)

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
			"evaluation_form": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        evaluationForm,
			},
			"user": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        user,
			},
		},
	}

	meteredEvaluationAssignment = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"evaluation_context_id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"evaluators": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        user,
			},
			"max_number_evaluations": {
				Description: "",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"evaluation_form": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        evaluationForm,
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
			"evaluation_context_id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"evaluators": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        user,
			},
			"max_number_evaluations": {
				Description: "",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"evaluation_form": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        evaluationForm,
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
			"calibrator": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        user,
			},
			"evaluators": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        user,
			},
			"evaluation_form": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        evaluationForm,
			},
			"expert_evaluator": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        user,
			},
		},
	}

	surveyAssignment = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"survey_form": {
				Description: "The survey form used for this survey.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        publishedSurveyFormReference,
			},
			"flow": {
				Description: "The URI reference to the flow associated with this survey.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        domainEntityRef,
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
			"integration": {
				Description: "The aws-s3-recording-bulk-actions-integration that the policy uses for exports.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        domainEntityRef,
			},
			"should_export_screen_recordings": {
				Description: "True if the policy should export screen recordings in addition to the other conversation media.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
		},
	}

	evaluationForm = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The globally unique identifier for the object.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "The media retention policy name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"modified_date": {
				Description: "Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"published": {
				Description: "",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"context_id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"question_groups": {
				Description: "A list of question groups",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        evaluationQuestionGroup,
			},
			// This field causes an InvalidInitCycle exception
			// "published_versions": {
			// 	Description: "",
			// 	Type:        schema.TypeList,
			//  MaxItems: 1,
			// 	Optional:    true,
			// 	Elem:        domainEntityListingEvaluationForm,
			// },
		},
	}

	user = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"division": {
				Description: "The division to which this entity belongs.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        division,
			},
			"chat": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        chat,
			},
			"department": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"email": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"addresses": {
				Description: "Email addresses and phone numbers for this user",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        contact,
			},
			"title": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"username": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			// This field causes an InvalidInitCycle exception
			// "manager": {
			// 	Description: "",
			// 	Type:        schema.TypeList,
			//  MaxItems:    1,
			// 	Optional:    true,
			// 	Elem:        user,
			// },
			"images": {
				Description: "Email addresses and phone numbers for this user",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem:        userImage,
			},
			"version": {
				Description: "Required when updating a user, this value should be the current version of the user. The current version can be obtained with a GET on the user before doing a PATCH.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"certifications": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"biography": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        biography,
			},
			"employer_info": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        employerInfo,
			},
			"acd_auto_answer": {
				Description: "acd auto answer",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"last_token_issued": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        OAuthLastTokenIssued,
			},
		},
	}

	serviceLevel = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"percentage": {
				Description: "The desired Service Level. A value between 0 and 1.",
				Type:        schema.TypeFloat,
				Optional:    true,
				Computed:    true,
			},
			"duration_ms": {
				Description: "Service Level target in milliseconds.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
		},
	}

	mediaSetting = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"alerting_timeout_seconds": {
				Description: "",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"service_level": {
				Description: "Service Level target in milliseconds.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        serviceLevel,
			},
		},
	}

	queue = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"division": {
				Description: "The division to which this entity belongs.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        division,
			},
			"description": {
				Description: "The queue description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"date_created": {
				Description: "The date the queue was created. Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"date_modified": {
				Description: "The date of the last modification to the queue. Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"modified_by": {
				Description: "The ID of the user that last modified the queue.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"created_by": {
				Description: "The ID of the user that created the queue.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"media_settings": {
				Description: "The media settings for the queue.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"call": {
							Description: "",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Elem:        mediaSetting,
						},
						"callback": {
							Description: "",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Elem:        mediaSetting,
						},
						"chat": {
							Description: "",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Elem:        mediaSetting,
						},
						"email": {
							Description: "",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Elem:        mediaSetting,
						},
						"message": {
							Description: "",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Elem:        mediaSetting,
						},
						"social_expression": {
							Description: "",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Elem:        mediaSetting,
						},
						"video_comm": {
							Description: "",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Elem:        mediaSetting,
						},
					},
				},
				// Valid keys: CALL, CALLBACK, CHAT, EMAIL, MESSAGE, SOCIAL_EXPRESSION, VIDEO_COMM
			},
			"routing_rules": {
				Description: "The routing rules for the queue, used for routing to known or preferred agents.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        routingRule,
			},
			"bullseye": {
				Description: "The bulls-eye settings for the queue.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        bullseye,
			},
			"acw_settings": {
				Description: "The ACW settings for the queue.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        acwSettings,
			},
			"skill_evaluation_method": {
				Description:  "The skill evaluation method to use when routing conversations.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"NONE", "BEST", "ALL", ""}, false),
			},
			"queue_flow": {
				Description: "The in-queue flow to use for call conversations waiting in queue.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        domainEntityRef,
			},
			"whisper_prompt": {
				Description: "The prompt used for whisper on the queue, if configured.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        domainEntityRef,
			},
			"auto_answer_only": {
				Description: "Specifies whether the configured whisper should play for all ACD calls, or only for those which are auto-answered.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"enable_transcription": {
				Description: "Indicates whether voice transcription is enabled for this queue.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"enable_manual_assignment": {
				Description: "Indicates whether manual assignment is enabled for this queue.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"calling_party_name": {
				Description: "The name to use for caller identification for outbound calls from this queue.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"calling_party_number": {
				Description: "The phone number to use for caller identification for outbound calls from this queue.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"outbound_messaging_addresses": {
				Description: "The messaging addresses for the queue.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        queueMessagingAddresses,
			},
			"outbound_email_address": {
				Description: "The messaging addresses for the queue.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        queueEmailAddress,
			},
		},
	}

	publishedSurveyFormReference = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"context_id": {
				Description: "The context id of this form.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	domainEntityRef = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"self_uri": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	// This schema var causes an InvalidInitCycle exception
	// domainEntityListingEvaluationForm = &schema.Resource{
	// 	Schema: map[string]*schema.Schema{
	// 		"entities": {
	// 			Description: "",
	// 			Type:        schema.TypeList,
	// 			Optional:    true,
	// 			Elem:        evaluationForm,
	// 		},
	// 		"page_size": {
	// 			Description: "",
	// 			Type:        schema.TypeInt,
	// 			Optional:    true,
	// 		},
	// 		"page_number": {
	// 			Description: "",
	// 			Type:        schema.TypeInt,
	// 			Optional:    true,
	// 		},
	// 		"total": {
	// 			Description: "",
	// 			Type:        schema.TypeInt,
	// 			Optional:    true,
	// 		},
	// 		"first_uri": {
	// 			Description: "",
	// 			Type:        schema.TypeString,
	// 			Optional:    true,
	// 		},
	// 		"self_uri": {
	// 			Description: "",
	// 			Type:        schema.TypeString,
	// 			Optional:    true,
	// 		},
	// 		"next_uri": {
	// 			Description: "",
	// 			Type:        schema.TypeString,
	// 			Optional:    true,
	// 		},
	// 		"previous_uri": {
	// 			Description: "",
	// 			Type:        schema.TypeString,
	// 			Optional:    true,
	// 		},
	// 		"last_uri": {
	// 			Description: "",
	// 			Type:        schema.TypeString,
	// 			Optional:    true,
	// 		},
	// 		"page_count": {
	// 			Description: "",
	// 			Type:        schema.TypeInt,
	// 			Optional:    true,
	// 		},
	// 	},
	// }

	evaluationQuestionGroup = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"type": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"default_answers_to_highest": {
				Description: "",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Default:     nil,
			},
			"default_answers_to_na": {
				Description: "",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Default:     nil,
			},
			"na_enabled": {
				Description: "",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Default:     nil,
			},
			"weight": {
				Description: "",
				Type:        schema.TypeFloat,
				Optional:    true,
				Computed:    true,
			},
			"manual_weight": {
				Description: "",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Default:     nil,
			},
			"questions": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem:        evaluationQuestion,
			},
			"visibility_condition": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        visibilityCondition,
			},
		},
	}

	evaluationQuestion = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"text": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"help_text": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"type": {
				Description:  "",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"multipleChoiceQuestion", "freeTextQuestion", "npsQuestion", "readOnlyTextBlockQuestion"}, false),
			},
			"na_enabled": {
				Description: "",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"comments_required": {
				Description: "",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"visibility_condition": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        visibilityCondition,
			},
			"answer_options": {
				Description: "Options from which to choose an answer for this question. Only used by Multiple Choice type questions.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        answerOption,
			},
			"is_kill": {
				Description: "",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"is_critical": {
				Description: "",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}

	visibilityCondition = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"combining_operation": {
				Description:  "",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"AND", "OR"}, false),
			},
			"predicates": {
				Description: "A list of strings, each representing the location in the form of the Answer Option to depend on. In the format of \"/form/questionGroup/{questionGroupIndex}/question/{questionIndex}/answer/{answerIndex}\" or, to assume the current question group, \"../question/{questionIndex}/answer/{answerIndex}\". Note: Indexes are zero-based",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	answerOption = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"text": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"value": {
				Description: "",
				Type:        schema.TypeInt,
				Optional:    true,
			},
		},
	}

	inboundRoute = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"pattern": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"queue": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        domainEntityRef,
			},
			"priority": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"skills": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        domainEntityRef,
			},
			"language": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        domainEntityRef,
			},
			"from_name": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"from_email": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"flow": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        domainEntityRef,
			},
			// This field causes an InvalidInitCycle exception
			// "reply_email_address": {
			// 	Description: "",
			// 	Type:        schema.TypeString,
			// 	Optional:    true,
			// 	Elem:        queueEmailAddress,
			// },
			"auto_bcc": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        emailAddress,
			},
			"spam_flow": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        domainEntityRef,
			},
		},
	}

	emailAddress = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"email": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
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

	division = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"self_uri": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	chat = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"jabber_id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	contact = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"address": {
				Description: "Email address or phone number for this contact type",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"media_type": {
				Description:  "",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"PHONE", "EMAIL", "SMS"}, false),
			},
			"type": {
				Description:  "",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"PRIMARY", "WORK", "WORK2", "WORK3", "WORK4", "HOME", "MOBILE", "MAIN", "OTHER"}, false),
			},
			"extension": {
				Description: "Use internal extension instead of address. Mutually exclusive with the address field.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"country_code": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"integration": {
				Description: "Integration tag value if this number is associated with an external integration.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	userImage = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"resolution": {
				Description: "Height and/or width of image. ex: 640x480 or x128",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"image_uri": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	biography = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"biography": {
				Description: "Personal detailed description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"interests": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"hobbies": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"spouse": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"education": {
				Description: "User education details",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        education,
			},
		},
	}

	employerInfo = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"official_name": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"employee_id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"employee_type": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"date_hire": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	OAuthLastTokenIssued = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"date_issued": {
				Description: "Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	routingRule = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"operator": {
				Description:  "matching operator. MEETS_THRESHOLD matches any agent with a score at or above the rule's threshold. ANY matches all specified agents, regardless of score.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"MEETS_THRESHOLD", "ANY"}, false),
			},
			"threshold": {
				Description: "threshold required for routing attempt (generally an agent score). may be null for operator ANY.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"wait_seconds": {
				Description: "seconds to wait in this rule before moving to the next",
				Type:        schema.TypeFloat,
				Optional:    true,
			},
		},
	}

	bullseye = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"rings": {
				Description: "The bullseye rings configured for this queue.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        ring,
			},
		},
	}

	acwSettings = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"wrapup_prompt": {
				Description:  "This field controls how the UI prompts the agent for a wrapup.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"MANDATORY", "OPTIONAL", "MANDATORY_TIMEOUT", "MANDATORY_FORCED_TIMEOUT", "AGENT_REQUESTED"}, false),
			},
			"timeout_ms": {
				Description: "The amount of time the agent can stay in ACW (Min: 1 sec, Max: 1 day). Can only be used when ACW is MANDATORY_TIMEOUT or MANDATORY_FORCED_TIMEOUT.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
		},
	}

	queueMessagingAddresses = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"sms_address": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        domainEntityRef,
			},
		},
	}

	queueEmailAddress = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"domain": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        domainEntityRef,
			},
			"route": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        inboundRoute,
			},
		},
	}

	wrapupCode = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "The wrap-up code name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"date_created": {
				Description: "Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"date_modified": {
				Description: "Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"modified_by": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"created_by": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}

	language = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The language name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "The language name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"date_modified": {
				Description: "Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"state": {
				Description:  "",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive", "deleted", ""}, false),
			},
			"version": {
				Description: "",
				Type:        schema.TypeString,
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
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	education = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"school": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"field_of_study": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"notes": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"date_start": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"date_end": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	ring = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"expansion_criteria": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        expansionCriterium,
			},
			"actions": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        actions,
			},
		},
	}

	expansionCriterium = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"threshold": {
				Description: "",
				Type:        schema.TypeFloat,
				Optional:    true,
			},
		},
	}

	skillsToRemove = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"id": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"self_uri": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	actions = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"skills_to_remove": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        skillsToRemove,
			},
		},
	}

	callMediaPolicyConditions = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"for_users": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        user,
			},
			"date_ranges": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Default:     nil,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"for_queues": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        queue,
			},
			"wrapup_codes": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        wrapupCode,
			},
			"languages": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        language,
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
				// Valid values: "INBOUND", "OUTBOUND"
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
			"for_users": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        user,
			},
			"date_ranges": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Default:     nil,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"for_queues": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        queue,
			},
			"wrapup_codes": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        wrapupCode,
			},
			"languages": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        language,
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
			"for_users": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        user,
			},
			"date_ranges": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Default:     nil,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"for_queues": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        queue,
			},
			"wrapup_codes": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        wrapupCode,
			},
			"languages": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        language,
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
			"for_users": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        user,
			},
			"date_ranges": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Default:     nil,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"for_queues": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        queue,
			},
			"wrapup_codes": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        wrapupCode,
			},
			"languages": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        language,
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
			"for_users": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        user,
			},
			"directions": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Default:     nil,
				Elem:        &schema.Schema{Type: schema.TypeString},
				// Valid values: "INBOUND", "OUTBOUND"
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
				// Valid values "CALL", "CHAT"
			},
			"for_queues": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        queue,
			},
			"duration": {
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        durationCondition,
			},
			"wrapup_codes": {
				Description: "",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        wrapupCode,
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
)

func resourceMediaRetentionPolicy() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Media Retention Policies",
		CreateContext: createWithPooledClient(createMediaRetentionPolicy),
		ReadContext:   readWithPooledClient(readMediaRetentionPolicy),
		UpdateContext: updateWithPooledClient(updateMediaRetentionPolicy),
		DeleteContext: deleteWithPooledClient(deleteMediaRetentionPolicy),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The policy name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"modified_date": {
				Description: "Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"created_date": {
				Description: "Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"order": {
				Description: "",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"description": {
				Description: "",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enabled": {
				Description: "",
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
				Description: "",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        policyErrors,
			},
		},
	}
}

func getAllMediaRetentionPolicies(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	recordingAPI := platformclientv2.NewRecordingApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		retentionPolicies, _, getErr := recordingAPI.GetRecordingMediaretentionpolicies(pageSize, pageNum, "", []string{}, "", "", "", true, false, false)

		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of media retention policys %v", getErr)
		}

		if retentionPolicies.Entities == nil || len(*retentionPolicies.Entities) == 0 {
			break
		}

		for _, retentionPolicy := range *retentionPolicies.Entities {
			resources[*retentionPolicy.Id] = &ResourceMeta{Name: *retentionPolicy.Name}
		}
	}

	return resources, nil
}

func buildVisibilityCondition(visibilityCondition []interface{}) *platformclientv2.Visibilitycondition {
	if visibilityCondition == nil || len(visibilityCondition) <= 0 {
		return nil
	}

	visibilityConditionMap := visibilityCondition[0].(map[string]interface{})

	combiningOperation := visibilityConditionMap["combining_operation"].(string)
	predicates := visibilityConditionMap["predicates"].([]interface{})

	return &platformclientv2.Visibilitycondition{
		CombiningOperation: &combiningOperation,
		Predicates:         &predicates,
	}
}

func flattenPolicyVisibilityCondition(visibilityCondition *platformclientv2.Visibilitycondition) []interface{} {
	if visibilityCondition == nil {
		return nil
	}

	visibilityConditionMap := make(map[string]interface{})
	if visibilityCondition.CombiningOperation != nil {
		visibilityConditionMap["combining_operation"] = *visibilityCondition.CombiningOperation
	}
	if visibilityCondition.Predicates != nil {
		visibilityConditionMap["predicates"] = interfaceListToStrings(*visibilityCondition.Predicates)
	}

	return []interface{}{visibilityConditionMap}
}

func buildAnswerOptions(options []interface{}) *[]platformclientv2.Answeroption {

	answerOptions := make([]platformclientv2.Answeroption, 0)

	for _, option := range options {
		optionMap := option.(map[string]interface{})
		id := optionMap["id"].(string)
		text := optionMap["text"].(string)
		value := optionMap["value"].(int)

		answerOptions = append(answerOptions, platformclientv2.Answeroption{
			Id:    &id,
			Text:  &text,
			Value: &value,
		})
	}

	return &answerOptions
}

func flattenPolicyAnswerOptions(answerOptions *[]platformclientv2.Answeroption) []interface{} {
	if answerOptions == nil {
		return nil
	}

	answerOptionsList := []interface{}{}

	for _, answerOption := range *answerOptions {
		answerOptionMap := make(map[string]interface{})
		if answerOption.Text != nil {
			answerOptionMap["text"] = *answerOption.Text
		}
		if answerOption.Value != nil {
			answerOptionMap["value"] = *answerOption.Value
		}

		answerOptionsList = append(answerOptionsList, answerOptionMap)
	}
	return answerOptionsList
}

func buildEvaluationQuestions(questions []interface{}) *[]platformclientv2.Evaluationquestion {

	evaluationQuestions := make([]platformclientv2.Evaluationquestion, 0)

	for _, question := range questions {
		questionMap := question.(map[string]interface{})
		id := questionMap["id"].(string)
		text := questionMap["text"].(string)
		helpText := questionMap["help_text"].(string)
		varType := questionMap["type"].(string)
		naEnabled := questionMap["na_enabled"].(bool)
		commentsRequired := questionMap["comments_required"].(bool)
		isKill := questionMap["is_kill"].(bool)
		isCritical := questionMap["is_critical"].(bool)

		evaluationQuestions = append(evaluationQuestions, platformclientv2.Evaluationquestion{
			Id:                  &id,
			Text:                &text,
			HelpText:            &helpText,
			VarType:             &varType,
			NaEnabled:           &naEnabled,
			CommentsRequired:    &commentsRequired,
			VisibilityCondition: buildVisibilityCondition(questionMap["visibility_condition"].([]interface{})),
			AnswerOptions:       buildAnswerOptions(questionMap["answer_options"].([]interface{})),
			IsKill:              &isKill,
			IsCritical:          &isCritical,
		})
	}

	return &evaluationQuestions
}

func flattenEvaluationQuestions(questions *[]platformclientv2.Evaluationquestion) []interface{} {
	if questions == nil {
		return nil
	}

	questionList := []interface{}{}

	for _, question := range *questions {
		questionMap := make(map[string]interface{})
		if question.Text != nil {
			questionMap["text"] = *question.Text
		}
		if question.HelpText != nil {
			questionMap["help_text"] = *question.HelpText
		}
		if question.NaEnabled != nil {
			questionMap["na_enabled"] = *question.NaEnabled
		}
		if question.CommentsRequired != nil {
			questionMap["comments_required"] = *question.CommentsRequired
		}
		if question.IsKill != nil {
			questionMap["is_kill"] = *question.IsKill
		}
		if question.IsCritical != nil {
			questionMap["is_critical"] = *question.IsCritical
		}
		if question.VisibilityCondition != nil {
			questionMap["visibility_condition"] = flattenPolicyVisibilityCondition(question.VisibilityCondition)
		}
		if question.AnswerOptions != nil {
			questionMap["answer_options"] = flattenPolicyAnswerOptions(question.AnswerOptions)
		}

		questionList = append(questionList, questionMap)
	}
	return questionList
}

func buildQuestionGroups(groups []interface{}) *[]platformclientv2.Evaluationquestiongroup {

	questionGroups := make([]platformclientv2.Evaluationquestiongroup, 0)

	for _, group := range groups {
		groupMap := group.(map[string]interface{})
		id := groupMap["id"].(string)
		name := groupMap["name"].(string)
		varType := groupMap["type"].(string)
		defaultAnswersToHighest := groupMap["default_answers_to_highest"].(bool)
		defaultAnswersToNA := groupMap["default_answers_to_na"].(bool)
		naEnabled := groupMap["na_enabled"].(bool)
		weight := float32(groupMap["weight"].(float64))
		manualWeight := groupMap["manual_weight"].(bool)

		questionGroups = append(questionGroups, platformclientv2.Evaluationquestiongroup{
			Id:                      &id,
			Name:                    &name,
			VarType:                 &varType,
			DefaultAnswersToHighest: &defaultAnswersToHighest,
			DefaultAnswersToNA:      &defaultAnswersToNA,
			NaEnabled:               &naEnabled,
			Weight:                  &weight,
			ManualWeight:            &manualWeight,
			Questions:               buildEvaluationQuestions(groupMap["questions"].([]interface{})),
			VisibilityCondition:     buildVisibilityCondition(groupMap["visibility_condition"].([]interface{})),
		})
	}

	return &questionGroups
}

func flattenPolicyQuestionGroups(questionGroups *[]platformclientv2.Evaluationquestiongroup) []interface{} {
	if questionGroups == nil {
		return nil
	}

	questionGroupList := []interface{}{}

	for _, questionGroup := range *questionGroups {
		questionGroupMap := make(map[string]interface{})
		if questionGroup.Name != nil {
			questionGroupMap["name"] = *questionGroup.Name
		}
		if questionGroup.DefaultAnswersToHighest != nil {
			questionGroupMap["default_answers_to_highest"] = *questionGroup.DefaultAnswersToHighest
		}
		if questionGroup.DefaultAnswersToNA != nil {
			questionGroupMap["default_answers_to_na"] = *questionGroup.DefaultAnswersToNA
		}
		if questionGroup.NaEnabled != nil {
			questionGroupMap["na_enabled"] = *questionGroup.NaEnabled
		}
		if questionGroup.Weight != nil {
			questionGroupMap["weight"] = *questionGroup.Weight
		}
		if questionGroup.ManualWeight != nil {
			questionGroupMap["manual_weight"] = *questionGroup.ManualWeight
		}
		if questionGroup.Questions != nil {
			questionGroupMap["questions"] = flattenEvaluationQuestions(questionGroup.Questions)
		}
		if questionGroup.VisibilityCondition != nil {
			questionGroupMap["visibility_condition"] = flattenPolicyVisibilityCondition(questionGroup.VisibilityCondition)
		}

		questionGroupList = append(questionGroupList, questionGroupMap)
	}
	return questionGroupList
}

func buildEvaluationForm(evaluationForm []interface{}) *platformclientv2.Evaluationform {
	if evaluationForm == nil || len(evaluationForm) <= 0 {
		return nil
	}

	evaluationFormMap := evaluationForm[0].(map[string]interface{})

	name := evaluationFormMap["name"].(string)
	published := evaluationFormMap["published"].(bool)
	contextId := evaluationFormMap["context_id"].(string)

	form := &platformclientv2.Evaluationform{
		Name:           &name,
		Published:      &published,
		ContextId:      &contextId,
		QuestionGroups: buildQuestionGroups(evaluationFormMap["question_groups"].([]interface{})),
	}

	if evaluationFormMap["modified_date"].(string) != "" {
		modifiedDateString := evaluationFormMap["modified_date"].(string)
		modifiedDate, modifiedErr := time.Parse("2006-01-02T15:04:05-0700", modifiedDateString)
		if modifiedErr == nil {
			form.ModifiedDate = &modifiedDate
		}
	}

	return form
}

func flattenEvaluationForm(evaluationForm *platformclientv2.Evaluationform) []interface{} {
	if evaluationForm == nil {
		return nil
	}

	evaluationFormMap := make(map[string]interface{})
	if evaluationForm.Name != nil {
		evaluationFormMap["name"] = *evaluationForm.Name
	}
	if evaluationForm.ModifiedDate != nil && len(evaluationForm.ModifiedDate.String()) > 0 {
		temp := *evaluationForm.ModifiedDate
		evaluationFormMap["modified_date"] = temp.String()
	}

	if evaluationForm.Published != nil {
		evaluationFormMap["published"] = *evaluationForm.Published
	}
	if evaluationForm.ContextId != nil {
		evaluationFormMap["context_id"] = *evaluationForm.ContextId
	}
	if evaluationForm.QuestionGroups != nil {
		evaluationFormMap["question_groups"] = flattenPolicyQuestionGroups(evaluationForm.QuestionGroups)
	}

	return []interface{}{evaluationFormMap}
}

func buildDivision(div []interface{}) *platformclientv2.Division {
	if div == nil || len(div) <= 0 {
		return nil
	}

	divMap := div[0].(map[string]interface{})
	name := divMap["name"].(string)
	id := divMap["id"].(string)
	selfUri := divMap["self_uri"].(string)

	return &platformclientv2.Division{
		Name:    &name,
		Id:      &id,
		SelfUri: &selfUri,
	}
}

func flattenDivision(division *platformclientv2.Division) []interface{} {
	if division == nil {
		return nil
	}

	divisionMap := make(map[string]interface{})
	if division.Name != nil {
		divisionMap["name"] = *division.Name
	}
	if division.Id != nil {
		divisionMap["id"] = *division.Id
	}
	if division.SelfUri != nil {
		divisionMap["self_uri"] = *division.SelfUri
	}

	return []interface{}{divisionMap}
}

func buildChat(chat []interface{}) *platformclientv2.Chat {
	if chat == nil || len(chat) <= 0 {
		return nil
	}

	chatMap := chat[0].(map[string]interface{})

	jabberId := chatMap["jabber_id"].(string)

	return &platformclientv2.Chat{
		JabberId: &jabberId,
	}
}

func flattenChat(chat *platformclientv2.Chat) []interface{} {
	if chat == nil {
		return nil
	}

	chatMap := make(map[string]interface{})
	if chat.JabberId != nil {
		chatMap["jabber_id"] = *chat.JabberId
	}

	return []interface{}{chatMap}
}

func buildAddresses(addresses []interface{}) *[]platformclientv2.Contact {
	contacts := make([]platformclientv2.Contact, 0)

	for _, contact := range addresses {
		contactMap := contact.(map[string]interface{})
		address := contactMap["address"].(string)
		mediaType := contactMap["media_type"].(string)
		varType := contactMap["type"].(string)
		extension := contactMap["extension"].(string)
		countryCode := contactMap["country_code"].(string)
		integration := contactMap["integration"].(string)

		contacts = append(contacts, platformclientv2.Contact{
			Address:     &address,
			MediaType:   &mediaType,
			VarType:     &varType,
			Extension:   &extension,
			CountryCode: &countryCode,
			Integration: &integration,
		})
	}

	return &contacts
}

func flattenAddresses(addresses *[]platformclientv2.Contact) []interface{} {
	if addresses == nil {
		return nil
	}

	addressList := []interface{}{}

	for _, address := range *addresses {
		addressMap := make(map[string]interface{})
		if address.Address != nil {
			addressMap["address"] = *address.Address
		}
		if address.MediaType != nil {
			addressMap["media_type"] = *address.MediaType
		}
		if address.VarType != nil {
			addressMap["type"] = *address.VarType
		}
		if address.Extension != nil {
			addressMap["extension"] = *address.Extension
		}
		if address.CountryCode != nil {
			addressMap["country_code"] = *address.CountryCode
		}
		if address.Integration != nil {
			addressMap["integration"] = *address.Integration
		}

		addressList = append(addressList, addressMap)
	}
	return addressList
}

func buildImages(images []interface{}) *[]platformclientv2.Userimage {
	builtImages := make([]platformclientv2.Userimage, 0)

	for _, image := range images {
		imageMap := image.(map[string]interface{})
		resolution := imageMap["resolution"].(string)
		imageUri := imageMap["image_uri"].(string)

		builtImages = append(builtImages, platformclientv2.Userimage{
			Resolution: &resolution,
			ImageUri:   &imageUri,
		})
	}

	return &builtImages
}

func flattenImages(images *[]platformclientv2.Userimage) []interface{} {
	if images == nil {
		return nil
	}

	imageList := []interface{}{}

	for _, image := range *images {
		imageMap := make(map[string]interface{})
		if image.Resolution != nil {
			imageMap["resolution"] = *image.Resolution
			imageMap["image_uri"] = *image.ImageUri
		}

		imageList = append(imageList, imageMap)
	}
	return imageList
}

func buildEducation(education []interface{}) *[]platformclientv2.Education {
	builtEducation := make([]platformclientv2.Education, 0)

	for _, e := range education {
		educationMap := e.(map[string]interface{})
		school := educationMap["school"].(string)
		fieldOfStudy := educationMap["field_of_study"].(string)
		notes := educationMap["notes"].(string)
		dateStartString := educationMap["date_start"].(string)
		dateEndString := educationMap["date_end"].(string)

		temp := platformclientv2.Education{
			School:       &school,
			FieldOfStudy: &fieldOfStudy,
			Notes:        &notes,
		}

		dateStart, startErr := time.Parse("2006-01-02T15:04:05-0700", dateStartString)
		if startErr == nil {
			temp.DateStart = &dateStart
		}

		dateEnd, endErr := time.Parse("2006-01-02T15:04:05-0700", dateEndString)
		if endErr == nil {
			temp.DateEnd = &dateEnd
		}

		builtEducation = append(builtEducation, temp)
	}

	return &builtEducation
}

func flattenEducation(education *[]platformclientv2.Education) []interface{} {
	if education == nil {
		return nil
	}

	educationList := []interface{}{}

	for _, e := range *education {
		educationMap := make(map[string]interface{})
		if e.School != nil {
			educationMap["school"] = *e.School
		}
		if e.FieldOfStudy != nil {
			educationMap["field_of_study"] = *e.FieldOfStudy
		}
		if e.Notes != nil {
			educationMap["notes"] = *e.Notes
		}
		if e.DateStart != nil && len(e.DateStart.String()) > 0 {
			temp := *e.DateStart
			educationMap["date_start"] = temp.String()
		}
		if e.DateEnd != nil && len(e.DateEnd.String()) > 0 {
			temp := *e.DateEnd
			educationMap["date_end"] = temp.String()
		}

		educationList = append(educationList, educationMap)
	}

	return educationList
}

func buildBiography(biography []interface{}) *platformclientv2.Biography {
	if biography == nil || len(biography) <= 0 {
		return nil
	}

	biographyMap := biography[0].(map[string]interface{})

	biographyDescription := biographyMap["biography"].(string)
	interests := biographyMap["interests"].([]string)
	hobbies := biographyMap["hobbies"].([]string)

	spouse := biographyMap["spouse"].(string)

	return &platformclientv2.Biography{
		Biography: &biographyDescription,
		Interests: &interests,
		Hobbies:   &hobbies,
		Spouse:    &spouse,
		Education: buildEducation(biographyMap["education"].([]interface{})),
	}
}

func flattenBiography(biography *platformclientv2.Biography) []interface{} {
	if biography == nil {
		return nil
	}

	biographyMap := make(map[string]interface{})
	if biography.Biography != nil {
		biographyMap["biography"] = *biography.Biography
	}
	if biography.Interests != nil {
		biographyMap["interests"] = *biography.Interests
	}
	if biography.Hobbies != nil {
		biographyMap["hobbies"] = *biography.Hobbies
	}
	if biography.Spouse != nil {
		biographyMap["spouse"] = *biography.Spouse
	}
	if biography.Education != nil {
		biographyMap["education"] = flattenEducation(biography.Education)
	}

	return []interface{}{biographyMap}
}

func buildEmployerInfo(employerInfo []interface{}) *platformclientv2.Employerinfo {
	if employerInfo == nil || len(employerInfo) <= 0 {
		return nil
	}

	employerInfoMap := employerInfo[0].(map[string]interface{})

	officialName := employerInfoMap["official_name"].(string)
	employeeId := employerInfoMap["employee_id"].(string)
	employeeType := employerInfoMap["employee_type"].(string)
	dateHire := employerInfoMap["date_hire"].(string)

	return &platformclientv2.Employerinfo{
		OfficialName: &officialName,
		EmployeeId:   &employeeId,
		EmployeeType: &employeeType,
		DateHire:     &dateHire,
	}
}

func flattenEmployerInfo(info *platformclientv2.Employerinfo) []interface{} {
	if info == nil {
		return nil
	}

	infoMap := make(map[string]interface{})
	if info.OfficialName != nil {
		infoMap["official_name"] = *info.OfficialName
	}
	if info.EmployeeId != nil {
		infoMap["employee_id"] = *info.EmployeeId
	}
	if info.EmployeeType != nil {
		infoMap["employee_type"] = *info.EmployeeId
	}
	if info.DateHire != nil {
		infoMap["date_hire"] = *info.DateHire
	}

	return []interface{}{infoMap}
}

func buildLastTokenIssued(lastTokenIssued []interface{}) *platformclientv2.Oauthlasttokenissued {
	if lastTokenIssued == nil || len(lastTokenIssued) <= 0 {
		return nil
	}

	lastTokenIssuedMap := lastTokenIssued[0].(map[string]interface{})
	dateIssuedString := lastTokenIssuedMap["date_issued"].(string)
	temp := platformclientv2.Oauthlasttokenissued{}

	dateIssued, issuedErr := time.Parse("2006-01-02T15:04:05-0700", dateIssuedString)
	if issuedErr == nil {
		temp.DateIssued = &dateIssued
	}

	return &temp
}

func flattenLastTokenIssued(lastTokenIssued *platformclientv2.Oauthlasttokenissued) []interface{} {
	if lastTokenIssued == nil {
		return nil
	}

	lastTokenIssuedMap := make(map[string]interface{})
	if lastTokenIssued.DateIssued != nil && len(lastTokenIssued.DateIssued.String()) > 0 {
		temp := *lastTokenIssued.DateIssued
		lastTokenIssuedMap["date_issued"] = temp.String()
	}

	return []interface{}{lastTokenIssuedMap}
}

func buildUser(user []interface{}) *platformclientv2.User {
	if user == nil || len(user) <= 0 {
		return nil
	}

	userMap := user[0].(map[string]interface{})
	id := userMap["id"].(string)
	name := userMap["name"].(string)
	department := userMap["department"].(string)
	email := userMap["email"].(string)
	title := userMap["title"].(string)
	username := userMap["username"].(string)
	version := userMap["version"].(int)
	acdAutoAnswer := userMap["acd_auto_answer"].(bool)

	tempUser := &platformclientv2.User{
		Id:              &id,
		Name:            &name,
		Division:        buildDivision(userMap["division"].([]interface{})),
		Chat:            buildChat(userMap["chat"].([]interface{})),
		Department:      &department,
		Email:           &email,
		Addresses:       buildAddresses(userMap["addresses"].([]interface{})),
		Title:           &title,
		Username:        &username,
		Images:          buildImages(userMap["images"].([]interface{})),
		Version:         &version,
		Biography:       buildBiography(userMap["biography"].([]interface{})),
		EmployerInfo:    buildEmployerInfo(userMap["employer_info"].([]interface{})),
		AcdAutoAnswer:   &acdAutoAnswer,
		LastTokenIssued: buildLastTokenIssued(userMap["last_token_issued"].([]interface{})),
	}

	certifications, cErr := userMap["certifications"].([]string)
	if !cErr {
		tempUser.Certifications = &certifications
	}

	return tempUser
}

func flattenUser(user *platformclientv2.User) []interface{} {
	if user == nil {
		return nil
	}

	userMap := make(map[string]interface{})
	if user.Name != nil {
		userMap["name"] = *user.Name
	}
	if user.Division != nil {
		userMap["division"] = flattenDivision(user.Division)
	}
	if user.Chat != nil {
		userMap["chat"] = flattenChat(user.Chat)
	}
	if user.Department != nil {
		userMap["department"] = *user.Department
	}
	if user.Email != nil {
		userMap["email"] = *user.Email
	}
	if user.Addresses != nil {
		userMap["addresses"] = flattenAddresses(user.Addresses)
	}
	if user.Title != nil {
		userMap["title"] = *user.Title
	}
	if user.Username != nil {
		userMap["username"] = *user.Username
	}
	if user.Images != nil {
		userMap["images"] = flattenImages(user.Images)
	}
	if user.Version != nil {
		userMap["version"] = *user.Version
	}
	if user.Certifications != nil {
		userMap["certifications"] = *user.Certifications
	}
	if user.Biography != nil {
		userMap["biography"] = flattenBiography(user.Biography)
	}
	if user.EmployerInfo != nil {
		userMap["employer_info"] = flattenEmployerInfo(user.EmployerInfo)
	}
	if user.AcdAutoAnswer != nil {
		userMap["acd_auto_answer"] = *user.AcdAutoAnswer
	}
	if user.LastTokenIssued != nil {
		userMap["last_token_issued"] = flattenLastTokenIssued(user.LastTokenIssued)
	}

	return []interface{}{userMap}
}

func buildEvaluationAssignments(evaluations []interface{}) *[]platformclientv2.Evaluationassignment {
	assignEvaluations := make([]platformclientv2.Evaluationassignment, 0)

	for _, assignEvaluation := range evaluations {
		assignEvaluationMap := assignEvaluation.(map[string]interface{})
		assignEvaluations = append(assignEvaluations, platformclientv2.Evaluationassignment{
			EvaluationForm: buildEvaluationForm(assignEvaluationMap["evaluation_form"].([]interface{})),
			User:           buildUser(assignEvaluationMap["user"].([]interface{})),
		})
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
		if assignment.EvaluationForm != nil {
			assignmentMap["evaluation_form"] = flattenEvaluationForm(assignment.EvaluationForm)
		}
		if assignment.User != nil {
			assignmentMap["user"] = flattenUser(assignment.User)
		}

		evaluationAssignments = append(evaluationAssignments, assignmentMap)
	}
	return evaluationAssignments
}

func buildUsers(users []interface{}) *[]platformclientv2.User {
	builtUsers := make([]platformclientv2.User, 0)

	for _, user := range users {
		builtUsers = append(builtUsers, *buildUser([]interface{}{user}))
	}

	return &builtUsers
}

func flattenUsers(users *[]platformclientv2.User) []interface{} {
	if users == nil {
		return nil
	}

	userList := []interface{}{}

	for _, user := range *users {
		userMap := make(map[string]interface{})
		if user.Name != nil {
			userMap["name"] = *user.Name
		}
		if user.Division != nil {
			userMap["division"] = flattenDivision(user.Division)
		}
		if user.Chat != nil {
			userMap["chat"] = flattenChat(user.Chat)
		}
		if user.Department != nil {
			userMap["department"] = *user.Department
		}
		if user.Email != nil {
			userMap["email"] = *user.Email
		}
		if user.Addresses != nil {
			userMap["addresses"] = flattenAddresses(user.Addresses)
		}
		if user.Title != nil {
			userMap["title"] = *user.Title
		}
		if user.Username != nil {
			userMap["username"] = *user.Username
		}
		if user.Images != nil {
			userMap["images"] = flattenImages(user.Images)
		}
		if user.Version != nil {
			userMap["version"] = *user.Version
		}
		if user.Certifications != nil {
			userMap["certifications"] = *user.Certifications
		}
		if user.Biography != nil {
			userMap["biography"] = flattenBiography(user.Biography)
		}
		if user.EmployerInfo != nil {
			userMap["employer_info"] = flattenEmployerInfo(user.EmployerInfo)
		}
		if user.AcdAutoAnswer != nil {
			userMap["acd_auto_answer"] = *user.AcdAutoAnswer
		}
		if user.LastTokenIssued != nil {
			userMap["last_token_issued"] = flattenLastTokenIssued(user.LastTokenIssued)
		}
		userList = append(userList, userMap)
	}
	return userList
}

func buildTimeInterval(timeInterval []interface{}) *platformclientv2.Timeinterval {
	if timeInterval == nil || len(timeInterval) <= 0 {
		return nil
	}

	timeIntervalMap := timeInterval[0].(map[string]interface{})

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
		assignmentMap := assignment.(map[string]interface{})
		evaluationContextId := assignmentMap["evaluation_context_id"].(string)
		maxNumberEvaluations := assignmentMap["max_number_evaluations"].(int)
		assignToActiveUser := assignmentMap["assign_to_active_user"].(bool)

		meteredAssignments = append(meteredAssignments, platformclientv2.Meteredevaluationassignment{
			EvaluationContextId:  &evaluationContextId,
			Evaluators:           buildUsers(assignmentMap["evaluators"].([]interface{})),
			MaxNumberEvaluations: &maxNumberEvaluations,
			EvaluationForm:       buildEvaluationForm(assignmentMap["evaluation_form"].([]interface{})),
			AssignToActiveUser:   &assignToActiveUser,
			TimeInterval:         buildTimeInterval(assignmentMap["time_interval"].([]interface{})),
		})
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
		if assignment.EvaluationContextId != nil {
			assignmentMap["evaluation_context_id"] = *assignment.EvaluationContextId
		}
		if assignment.Evaluators != nil {
			assignmentMap["evaluators"] = flattenUsers(assignment.Evaluators)
		}
		if assignment.MaxNumberEvaluations != nil {
			assignmentMap["max_number_evaluations"] = *assignment.MaxNumberEvaluations
		}
		if assignment.EvaluationForm != nil {
			assignmentMap["evaluation_form"] = flattenEvaluationForm(assignment.EvaluationForm)
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
		assignmentMap := assignment.(map[string]interface{})
		evaluationContextId := assignmentMap["evaluation_context_id"].(string)
		maxNumberEvaluations := assignmentMap["max_number_evaluations"].(int)
		timeZone := assignmentMap["time_zone"].(string)

		meteredAssignments = append(meteredAssignments, platformclientv2.Meteredassignmentbyagent{
			EvaluationContextId:  &evaluationContextId,
			Evaluators:           buildUsers(assignmentMap["evaluators"].([]interface{})),
			MaxNumberEvaluations: &maxNumberEvaluations,
			EvaluationForm:       buildEvaluationForm(assignmentMap["evaluation_form"].([]interface{})),
			TimeInterval:         buildTimeInterval(assignmentMap["time_interval"].([]interface{})),
			TimeZone:             &timeZone,
		})
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
		if assignment.EvaluationContextId != nil {
			assignmentMap["evaluation_context_id"] = *assignment.EvaluationContextId
		}
		if assignment.Evaluators != nil {
			assignmentMap["evaluators"] = flattenUsers(assignment.Evaluators)
		}
		if assignment.MaxNumberEvaluations != nil {
			assignmentMap["max_number_evaluations"] = *assignment.MaxNumberEvaluations
		}
		if assignment.EvaluationForm != nil {
			assignmentMap["evaluation_form"] = flattenEvaluationForm(assignment.EvaluationForm)
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
		assignmentMap := assignment.(map[string]interface{})
		calibrationAssignments = append(calibrationAssignments, platformclientv2.Calibrationassignment{
			Calibrator:      buildUser(assignmentMap["calibrator"].([]interface{})),
			Evaluators:      buildUsers(assignmentMap["evaluators"].([]interface{})),
			EvaluationForm:  buildEvaluationForm(assignmentMap["evaluation_form"].([]interface{})),
			ExpertEvaluator: buildUser(assignmentMap["expert_evaluator"].([]interface{})),
		})
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
			assignmentMap["calibrator"] = flattenUser(assignment.Calibrator)
		}
		if assignment.Evaluators != nil {
			assignmentMap["evaluators"] = flattenUsers(assignment.Evaluators)
		}
		if assignment.EvaluationForm != nil {
			assignmentMap["evaluation_form"] = flattenEvaluationForm(assignment.EvaluationForm)
		}
		if assignment.ExpertEvaluator != nil {
			assignmentMap["expert_evaluator"] = flattenUser(assignment.ExpertEvaluator)
		}

		calibrationAssignments = append(calibrationAssignments, assignmentMap)
	}
	return calibrationAssignments
}

func buildPublishedSurveyFormReference(publishedSurveyFormReference []interface{}) *platformclientv2.Publishedsurveyformreference {
	if publishedSurveyFormReference == nil || len(publishedSurveyFormReference) <= 0 {
		return nil
	}

	referenceMap := publishedSurveyFormReference[0].(map[string]interface{})
	name := referenceMap["name"].(string)
	contextId := referenceMap["context_id"].(string)

	return &platformclientv2.Publishedsurveyformreference{
		Name:      &name,
		ContextId: &contextId,
	}
}

func flattenPublishedSurveyFormReference(reference *platformclientv2.Publishedsurveyformreference) []interface{} {
	if reference == nil {
		return nil
	}

	referenceMap := make(map[string]interface{})
	if reference.Name != nil {
		referenceMap["name"] = *reference.Name
	}
	if reference.ContextId != nil {
		referenceMap["context_id"] = *reference.ContextId
	}

	return []interface{}{referenceMap}
}

func buildDomainEntityRef(domainEntityRef []interface{}) *platformclientv2.Domainentityref {
	if domainEntityRef == nil || len(domainEntityRef) <= 0 {
		return nil
	}

	domainEntityRefMap := domainEntityRef[0].(map[string]interface{})

	id := domainEntityRefMap["id"].(string)
	name := domainEntityRefMap["name"].(string)
	selfUri := domainEntityRefMap["self_uri"].(string)

	return &platformclientv2.Domainentityref{
		Id:      &id,
		Name:    &name,
		SelfUri: &selfUri,
	}
}

func flattenDomainEntityRef(ref *platformclientv2.Domainentityref) []interface{} {
	if ref == nil {
		return nil
	}

	refMap := make(map[string]interface{})
	if ref.Id != nil {
		refMap["id"] = *ref.Id
	}
	if ref.Name != nil {
		refMap["name"] = *ref.Name
	}
	if ref.SelfUri != nil {
		refMap["self_uri"] = *ref.SelfUri
	}

	return []interface{}{refMap}
}

func flattenDomainEntityRefs(refs *[]platformclientv2.Domainentityref) []interface{} {
	if refs == nil {
		return nil
	}

	refList := []interface{}{}

	for _, ref := range *refs {
		refList = append(refList, flattenDomainEntityRef(&ref))
	}
	return refList
}

func buildAssignSurveys(assignments []interface{}) *[]platformclientv2.Surveyassignment {
	surveyAssignments := make([]platformclientv2.Surveyassignment, 0)

	for _, assignment := range assignments {
		assignmentMap := assignment.(map[string]interface{})
		sendingUser := assignmentMap["sending_user"].(string)
		sendingDomain := assignmentMap["sending_domain"].(string)
		inviteTimeInterval := assignmentMap["invite_time_interval"].(string)

		surveyAssignments = append(surveyAssignments, platformclientv2.Surveyassignment{
			SurveyForm:         buildPublishedSurveyFormReference(assignmentMap["survey_form"].([]interface{})),
			Flow:               buildDomainEntityRef(assignmentMap["flow"].([]interface{})),
			InviteTimeInterval: &inviteTimeInterval,
			SendingUser:        &sendingUser,
			SendingDomain:      &sendingDomain,
		})
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
			assignmentMap["survey_form"] = flattenPublishedSurveyFormReference(assignment.SurveyForm)
		}
		if assignment.Flow != nil {
			assignmentMap["flow"] = flattenDomainEntityRef(assignment.Flow)
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

	archiveRetentionMap := archiveRetention[0].(map[string]interface{})

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

	deleteRetentionMap := deleteRetention[0].(map[string]interface{})

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

	retentionDurationMap := retentionDuration[0].(map[string]interface{})

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

	initiateScreenRecordingMap := initiateScreenRecording[0].(map[string]interface{})
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
		transcriptionMap := transcription.(map[string]interface{})
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

	integrationExportMap := integrationExport[0].(map[string]interface{})
	shouldExportScreenRecordings := integrationExportMap["should_export_screen_recordings"].(bool)

	return &platformclientv2.Integrationexport{
		Integration:                  buildDomainEntityRef(integrationExportMap["integration"].([]interface{})),
		ShouldExportScreenRecordings: &shouldExportScreenRecordings,
	}
}

func flattenIntegrationExport(integrationExport *platformclientv2.Integrationexport) []interface{} {
	if integrationExport == nil {
		return nil
	}

	integrationExportMap := make(map[string]interface{})
	if integrationExport.Integration != nil {
		integrationExportMap["integration"] = flattenDomainEntityRef(integrationExport.Integration)
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

	actionsMap := actions[0].(map[string]interface{})

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

func buildServiceLevel(serviceLevel []interface{}) *platformclientv2.Servicelevel {
	if serviceLevel == nil || len(serviceLevel) <= 0 {
		return nil
	}

	serviceLevelMap := serviceLevel[0].(map[string]interface{})
	percentage := serviceLevelMap["percentage"].(float64)
	durationMs := serviceLevelMap["duration_ms"].(int)

	return &platformclientv2.Servicelevel{
		Percentage: &percentage,
		DurationMs: &durationMs,
	}
}

func flattenServiceLevel(serviceLevel *platformclientv2.Servicelevel) []interface{} {
	if serviceLevel == nil {
		return nil
	}

	serviceLevelMap := make(map[string]interface{})
	if serviceLevel.Percentage != nil {
		serviceLevelMap["percentage"] = *serviceLevel.Percentage
	}
	if serviceLevel.DurationMs != nil {
		serviceLevelMap["duration_ms"] = *serviceLevel.DurationMs
	}

	return []interface{}{serviceLevelMap}
}

func buildMediaSettings(settings []interface{}) *map[string]platformclientv2.Mediasetting {
	if settings == nil || len(settings) <= 0 {
		return nil
	}

	mediaSettings := make(map[string]platformclientv2.Mediasetting)
	settingsMap := settings[0].(map[string]interface{})

	for k, v := range settingsMap {
		vMap := v.(map[string]interface{})
		alertingTimeoutSeconds := vMap["alerting_timeout_seconds"].(int)
		camelCaseKey := toCamelCase(k)
		mediaSettings[camelCaseKey] = platformclientv2.Mediasetting{
			AlertingTimeoutSeconds: &alertingTimeoutSeconds,
			ServiceLevel:           buildServiceLevel(vMap["service_level"].([]interface{})),
		}
	}

	return &mediaSettings
}

func flattenMediaSettings(settings map[string]platformclientv2.Mediasetting) []interface{} {
	if settings == nil {
		return nil
	}

	mediaSettingsMap := make(map[string]interface{})

	for k, v := range settings {
		setting := make(map[string]interface{})
		if v.AlertingTimeoutSeconds != nil {
			setting["alerting_timeout_seconds"] = *v.AlertingTimeoutSeconds
		}
		if v.ServiceLevel != nil {
			setting["service_level"] = flattenServiceLevel(v.ServiceLevel)
		}
		snakeCaseKey := toSnakeCase(k)
		mediaSettingsMap[snakeCaseKey] = []interface{}{setting}
	}
	return []interface{}{mediaSettingsMap}
}

func buildRoutingRules(rules []interface{}) *[]platformclientv2.Routingrule {
	routingRules := make([]platformclientv2.Routingrule, 0)

	for _, rule := range rules {
		ruleMap := rule.(map[string]interface{})
		operator := ruleMap["operator"].(string)
		threshold := ruleMap["threshold"].(int)
		waitSeconds := ruleMap["wait_seconds"].(float64)

		routingRules = append(routingRules, platformclientv2.Routingrule{
			Operator:    &operator,
			Threshold:   &threshold,
			WaitSeconds: &waitSeconds,
		})
	}

	return &routingRules
}

func flattenRoutingRules2(rules *[]platformclientv2.Routingrule) []interface{} {
	if rules == nil {
		return nil
	}

	RulesList := []interface{}{}

	for _, rule := range *rules {
		ruleMap := make(map[string]interface{})
		if rule.Operator != nil {
			ruleMap["operator"] = *rule.Operator
		}
		if rule.Threshold != nil {
			ruleMap["threshold"] = *rule.Threshold
		}
		if rule.WaitSeconds != nil {
			ruleMap["wait_seconds"] = *rule.WaitSeconds
		}

		RulesList = append(RulesList, ruleMap)
	}

	return RulesList
}

func buildExpansionCriteria(criteria []interface{}) *[]platformclientv2.Expansioncriterium {
	expansionCriteria := make([]platformclientv2.Expansioncriterium, 0)

	for _, criterium := range criteria {
		criteriumMap := criterium.(map[string]interface{})
		varType := criteriumMap["type"].(string)
		threshold := criteriumMap["threshold"].(float64)

		expansionCriteria = append(expansionCriteria, platformclientv2.Expansioncriterium{
			VarType:   &varType,
			Threshold: &threshold,
		})
	}

	return &expansionCriteria
}

func flattenExpansionCriteria(criteria *[]platformclientv2.Expansioncriterium) []map[string]interface{} {
	if criteria == nil {
		return nil
	}

	criteriaList := []map[string]interface{}{}

	for _, criterium := range *criteria {
		criteriumMap := make(map[string]interface{})
		if criterium.VarType != nil {
			criteriumMap["expansion_criteria"] = *criterium.VarType
		}
		if criterium.Threshold != nil {
			criteriumMap["threshold"] = *criterium.Threshold
		}

		criteriaList = append(criteriaList, criteriumMap)
	}

	return criteriaList
}

func buildSkillsToRemove(skills []interface{}) *[]platformclientv2.Skillstoremove {
	skillsToRemove := make([]platformclientv2.Skillstoremove, 0)

	for _, skill := range skills {
		skillMap := skill.(map[string]interface{})
		name := skillMap["name"].(string)
		id := skillMap["id"].(string)
		selfUri := skillMap["self_uri"].(string)

		skillsToRemove = append(skillsToRemove, platformclientv2.Skillstoremove{
			Name:    &name,
			Id:      &id,
			SelfUri: &selfUri,
		})
	}

	return &skillsToRemove
}

func flattenSkillsToRemove(skills *[]platformclientv2.Skillstoremove) []interface{} {
	if skills == nil {
		return nil
	}

	skillList := []interface{}{}

	for _, skill := range *skills {
		skillMap := make(map[string]interface{})

		if skill.Name != nil {
			skillMap["name"] = *skill.Name
		}
		if skill.Id != nil {
			skillMap["id"] = *skill.Id
		}
		if skill.SelfUri != nil {
			skillMap["self_uri"] = *skill.SelfUri
		}

		skillList = append(skillList, skillMap)
	}

	return skillList
}

func buildActions(actions []interface{}) *platformclientv2.Actions {
	if actions == nil || len(actions) <= 0 {
		return nil
	}

	actionsMap := actions[0].(map[string]interface{})
	return &platformclientv2.Actions{
		SkillsToRemove: buildSkillsToRemove(actionsMap["skills_to_remove"].([]interface{})),
	}
}

func flattenActions(actions *platformclientv2.Actions) []interface{} {
	if actions == nil {
		return nil
	}

	actionsMap := make(map[string]interface{})
	if actions.SkillsToRemove != nil {
		actionsMap["skills_to_remove"] = flattenSkillsToRemove(actions.SkillsToRemove)
	}

	return []interface{}{actionsMap}
}

func buildRings(rings []interface{}) *[]platformclientv2.Ring {
	builtRings := make([]platformclientv2.Ring, 0)

	for _, ring := range rings {
		ringMap := ring.(map[string]interface{})
		builtRings = append(builtRings, platformclientv2.Ring{
			ExpansionCriteria: buildExpansionCriteria(ringMap["expansion_criteria"].([]interface{})),
			Actions:           buildActions(ringMap["actions"].([]interface{})),
		})
	}

	return &builtRings
}

func flattenRings(rings *[]platformclientv2.Ring) []interface{} {
	if rings == nil {
		return nil
	}

	ringList := []interface{}{}

	for _, ring := range *rings {
		ringMap := make(map[string]interface{})
		if ring.ExpansionCriteria != nil {
			ringMap["expansion_criteria"] = flattenExpansionCriteria(ring.ExpansionCriteria)
		}
		if ring.Actions != nil {
			ringMap["actions"] = flattenActions(ring.Actions)
		}

		ringList = append(ringList, ringMap)
	}

	return ringList
}

func buildBullseye(bullseye []interface{}) *platformclientv2.Bullseye {
	if bullseye == nil || len(bullseye) <= 0 {
		return nil
	}

	bullseyeMap := bullseye[0].(map[string]interface{})

	return &platformclientv2.Bullseye{
		Rings: buildRings(bullseyeMap["rings"].([]interface{})),
	}
}

func flattenBullseye(bullseye *platformclientv2.Bullseye) []interface{} {
	if bullseye == nil {
		return nil
	}

	bullseyeMap := make(map[string]interface{})
	if bullseye.Rings != nil {
		bullseyeMap["rings"] = flattenRings(bullseye.Rings)
	}

	return []interface{}{bullseyeMap}
}

func buildAcwSettings(acwSettings []interface{}) *platformclientv2.Acwsettings {
	if acwSettings == nil || len(acwSettings) <= 0 {
		return nil
	}

	acwSettingsMap := acwSettings[0].(map[string]interface{})

	wrapupPrompt := acwSettingsMap["wrapup_prompt"].(string)
	timeoutMs := acwSettingsMap["timeout_ms"].(int)

	return &platformclientv2.Acwsettings{
		WrapupPrompt: &wrapupPrompt,
		TimeoutMs:    &timeoutMs,
	}
}

func flattenAcwSettings(settings *platformclientv2.Acwsettings) []interface{} {
	if settings == nil {
		return nil
	}

	settingsMap := make(map[string]interface{})
	if settings.WrapupPrompt != nil {
		settingsMap["wrapup_prompt"] = *settings.WrapupPrompt
	}
	if settings.TimeoutMs != nil {
		settingsMap["timeout_ms"] = *settings.TimeoutMs
	}

	return []interface{}{settingsMap}
}

func buildQueueMessagingAddresses(queueMessagingAddresses []interface{}) *platformclientv2.Queuemessagingaddresses {
	if queueMessagingAddresses == nil || len(queueMessagingAddresses) <= 0 {
		return nil
	}

	addressesMap := queueMessagingAddresses[0].(map[string]interface{})

	return &platformclientv2.Queuemessagingaddresses{
		SmsAddress: buildDomainEntityRef(addressesMap["sms_address"].([]interface{})),
	}
}

func flattenQueueMessagingAddresses(queueMessagingAddresses *platformclientv2.Queuemessagingaddresses) []interface{} {
	if queueMessagingAddresses == nil {
		return nil
	}

	queueMessagingAddressesMap := make(map[string]interface{})
	if queueMessagingAddresses.SmsAddress != nil {
		queueMessagingAddressesMap["sms_address"] = flattenDomainEntityRef(queueMessagingAddresses.SmsAddress)
	}

	return []interface{}{queueMessagingAddressesMap}
}

func buildDomainEntityRefs(refs []interface{}) *[]platformclientv2.Domainentityref {
	domainEntityRefs := make([]platformclientv2.Domainentityref, 0)

	for _, ref := range refs {
		domainEntityRefs = append(domainEntityRefs, *buildDomainEntityRef([]interface{}{ref}))
	}

	return &domainEntityRefs
}

func buildEmailAddresses(addresses []interface{}) *[]platformclientv2.Emailaddress {
	emailAddresses := make([]platformclientv2.Emailaddress, 0)

	for _, address := range addresses {
		addressMap := address.(map[string]interface{})
		email := addressMap["email"].(string)
		name := addressMap["name"].(string)

		emailAddresses = append(emailAddresses, platformclientv2.Emailaddress{
			Email: &email,
			Name:  &name,
		})
	}

	return &emailAddresses
}

func flattenEmailAddresses(addresses *[]platformclientv2.Emailaddress) []interface{} {
	if addresses == nil {
		return nil
	}

	AddressList := []interface{}{}

	for _, address := range *addresses {
		addressMap := make(map[string]interface{})
		if address.Email != nil {
			addressMap["email"] = *address.Email
		}
		if address.Name != nil {
			addressMap["name"] = *address.Name
		}

		AddressList = append(AddressList, addressMap)
	}

	return AddressList
}

func buildInboundRoute(inboundRoute []interface{}) *platformclientv2.Inboundroute {
	if inboundRoute == nil || len(inboundRoute) <= 0 {
		return nil
	}

	inboundRouteMap := inboundRoute[0].(map[string]interface{})

	id := inboundRouteMap["id"].(string)
	name := inboundRouteMap["name"].(string)
	pattern := inboundRouteMap["pattern"].(string)
	priority := inboundRouteMap["priority"].(int)
	fromName := inboundRouteMap["from_name"].(string)
	fromEmail := inboundRouteMap["from_email"].(string)
	selfUri := inboundRouteMap["self_uri"].(string)

	return &platformclientv2.Inboundroute{
		Id:        &id,
		Name:      &name,
		Pattern:   &pattern,
		Queue:     buildDomainEntityRef(inboundRouteMap["queue"].([]interface{})),
		Priority:  &priority,
		Skills:    buildDomainEntityRefs(inboundRouteMap["skills"].([]interface{})),
		Language:  buildDomainEntityRef(inboundRouteMap["language"].([]interface{})),
		FromName:  &fromName,
		FromEmail: &fromEmail,
		Flow:      buildDomainEntityRef(inboundRouteMap["flow"].([]interface{})),
		AutoBcc:   buildEmailAddresses(inboundRouteMap["auto_bcc"].([]interface{})),
		SpamFlow:  buildDomainEntityRef(inboundRouteMap["spam_flow"].([]interface{})),
		SelfUri:   &selfUri,
	}
}

func flattenInboundRoute(inboundRoute *platformclientv2.Inboundroute) []interface{} {
	if inboundRoute == nil {
		return nil
	}

	inboundRouteMap := make(map[string]interface{})
	if inboundRoute.Id != nil {
		inboundRouteMap["id"] = *inboundRoute.Id
	}
	if inboundRoute.Name != nil {
		inboundRouteMap["name"] = *inboundRoute.Name
	}
	if inboundRoute.Pattern != nil {
		inboundRouteMap["pattern"] = *inboundRoute.Pattern
	}
	if inboundRoute.Queue != nil {
		inboundRouteMap["queue"] = flattenDomainEntityRef(inboundRoute.Queue)
	}
	if inboundRoute.Priority != nil {
		inboundRouteMap["priority"] = *inboundRoute.Priority
	}
	if inboundRoute.Skills != nil {
		inboundRouteMap["skills"] = flattenDomainEntityRefs(inboundRoute.Skills)
	}
	if inboundRoute.Language != nil {
		inboundRouteMap["language"] = flattenDomainEntityRef(inboundRoute.Language)
	}
	if inboundRoute.FromName != nil {
		inboundRouteMap["from_name"] = *inboundRoute.FromName
	}
	if inboundRoute.FromEmail != nil {
		inboundRouteMap["from_email"] = *inboundRoute.FromEmail
	}
	if inboundRoute.Flow != nil {
		inboundRouteMap["flow"] = flattenDomainEntityRef(inboundRoute.Flow)
	}
	if inboundRoute.AutoBcc != nil {
		inboundRouteMap["flow"] = flattenEmailAddresses(inboundRoute.AutoBcc)
	}
	if inboundRoute.SpamFlow != nil {
		inboundRouteMap["spam_flow"] = flattenDomainEntityRef(inboundRoute.SpamFlow)
	}

	return []interface{}{inboundRouteMap}
}

func buildQueueEmailAddress(queueEmailAddress []interface{}) *platformclientv2.Queueemailaddress {
	if queueEmailAddress == nil || len(queueEmailAddress) <= 0 {
		return nil
	}

	addressMap := queueEmailAddress[0].(map[string]interface{})

	return &platformclientv2.Queueemailaddress{
		Domain: buildDomainEntityRef(addressMap["domain"].([]interface{})),
		Route:  buildInboundRoute(addressMap["route"].([]interface{})),
	}
}

func flattenQueueEmailAddress2(queueEmailAddress *platformclientv2.Queueemailaddress) []interface{} {
	if queueEmailAddress == nil {
		return nil
	}

	queueEmailAddressMap := make(map[string]interface{})
	if queueEmailAddress.Domain != nil {
		queueEmailAddressMap["domain"] = flattenDomainEntityRef(queueEmailAddress.Domain)
	}
	if queueEmailAddress.Route != nil {
		queueEmailAddressMap["route"] = flattenInboundRoute(queueEmailAddress.Route)
	}

	return []interface{}{queueEmailAddressMap}
}

func buildQueue(queue []interface{}) *platformclientv2.Queue {
	if queue == nil || len(queue) <= 0 {
		return nil
	}

	queueMap := queue[0].(map[string]interface{})
	id := queueMap["id"].(string)
	name := queueMap["name"].(string)
	description := queueMap["description"].(string)
	dateCreatedString := queueMap["date_created"].(string)
	dateModifiedString := queueMap["date_modified"].(string)
	modifiedBy := queueMap["modified_by"].(string)
	createdBy := queueMap["created_by"].(string)
	autoAnswerOnly := queueMap["auto_answer_only"].(bool)
	enableTranscription := queueMap["enable_transcription"].(bool)
	enableManualAssignment := queueMap["enable_manual_assignment"].(bool)
	callingPartyName := queueMap["calling_party_name"].(string)
	callingPartyNumber := queueMap["calling_party_number"].(string)

	formattedQueue := platformclientv2.Queue{
		Id:                         &id,
		Name:                       &name,
		Division:                   buildDivision(queueMap["division"].([]interface{})),
		Description:                &description,
		ModifiedBy:                 &modifiedBy,
		CreatedBy:                  &createdBy,
		MediaSettings:              buildMediaSettings(queueMap["media_settings"].([]interface{})),
		RoutingRules:               buildRoutingRules(queueMap["routing_rules"].([]interface{})),
		Bullseye:                   buildBullseye(queueMap["bullseye"].([]interface{})),
		AcwSettings:                buildAcwSettings(queueMap["acw_settings"].([]interface{})),
		QueueFlow:                  buildDomainEntityRef(queueMap["queue_flow"].([]interface{})),
		WhisperPrompt:              buildDomainEntityRef(queueMap["whisper_prompt"].([]interface{})),
		AutoAnswerOnly:             &autoAnswerOnly,
		EnableTranscription:        &enableTranscription,
		EnableManualAssignment:     &enableManualAssignment,
		CallingPartyName:           &callingPartyName,
		CallingPartyNumber:         &callingPartyNumber,
		OutboundMessagingAddresses: buildQueueMessagingAddresses(queueMap["outbound_messaging_addresses"].([]interface{})),
		OutboundEmailAddress:       buildQueueEmailAddress(queueMap["outbound_email_address"].([]interface{})),
	}

	dateCreated, createdErr := time.Parse("2006-01-02T15:04:05-0700", dateCreatedString)
	if createdErr == nil {
		formattedQueue.DateCreated = &dateCreated
	}

	dateModified, modifiedErr := time.Parse("2006-01-02T15:04:05-0700", dateModifiedString)
	if modifiedErr == nil {
		formattedQueue.DateModified = &dateModified
	}

	skillEvaluationMethod := queueMap["skill_evaluation_method"].(string)
	if len(skillEvaluationMethod) > 0 {
		formattedQueue.SkillEvaluationMethod = &skillEvaluationMethod
	}

	return &formattedQueue
}

func buildQueues(queues []interface{}) *[]platformclientv2.Queue {
	builtQueues := make([]platformclientv2.Queue, 0)

	for _, queue := range queues {
		builtQueues = append(builtQueues, *buildQueue([]interface{}{queue}))
	}

	return &builtQueues
}

func flattenQueues(queues *[]platformclientv2.Queue) []interface{} {
	if queues == nil {
		return nil
	}

	queueList := []interface{}{}

	for _, queue := range *queues {
		queueMap := make(map[string]interface{})
		if queue.Id != nil {
			queueMap["id"] = *queue.Id
		}
		if queue.Name != nil {
			queueMap["name"] = *queue.Name
		}
		if queue.Division != nil {
			queueMap["division"] = flattenDivision(queue.Division)
		}
		if queue.Description != nil {
			queueMap["description"] = *queue.Description
		}
		if queue.DateCreated != nil && len(queue.DateCreated.String()) > 0 {
			temp := *queue.DateCreated
			queueMap["date_created"] = temp.String()
		}
		if queue.DateModified != nil && len(queue.DateModified.String()) > 0 {
			temp := *queue.DateModified
			queueMap["date_modified"] = temp.String()
		}
		if queue.ModifiedBy != nil {
			queueMap["modified_by"] = *queue.ModifiedBy
		}
		if queue.CreatedBy != nil {
			queueMap["created_by"] = *queue.CreatedBy
		}
		if queue.MediaSettings != nil {
			queueMap["media_settings"] = flattenMediaSettings(*queue.MediaSettings)
		}
		if queue.RoutingRules != nil {
			queueMap["routing_rules"] = flattenRoutingRules2(queue.RoutingRules)
		}
		if queue.Bullseye != nil {
			queueMap["bullseye"] = flattenBullseye(queue.Bullseye)
		}
		if queue.AcwSettings != nil {
			queueMap["acw_settings"] = flattenAcwSettings(queue.AcwSettings)
		}
		if queue.SkillEvaluationMethod != nil {
			queueMap["skill_evaluation_method"] = *queue.SkillEvaluationMethod
		}
		if queue.QueueFlow != nil {
			queueMap["queue_flow"] = flattenDomainEntityRef(queue.QueueFlow)
		}
		if queue.WhisperPrompt != nil {
			queueMap["whisper_prompt"] = flattenDomainEntityRef(queue.WhisperPrompt)
		}
		if queue.AutoAnswerOnly != nil {
			queueMap["auto_answer_only"] = *queue.AutoAnswerOnly
		}
		if queue.EnableTranscription != nil {
			queueMap["enable_transcription"] = *queue.EnableTranscription
		}
		if queue.EnableManualAssignment != nil {
			queueMap["enable_manual_assignment"] = *queue.EnableManualAssignment
		}
		if queue.CallingPartyName != nil {
			queueMap["calling_party_name"] = *queue.CallingPartyName
		}
		if queue.CallingPartyNumber != nil {
			queueMap["calling_party_number"] = *queue.CallingPartyNumber
		}
		if queue.OutboundMessagingAddresses != nil {
			queueMap["outbound_messaging_addresses"] = flattenQueueMessagingAddresses(queue.OutboundMessagingAddresses)
		}
		if queue.OutboundEmailAddress != nil {
			queueMap["outbound_email_address"] = flattenQueueEmailAddress2(queue.OutboundEmailAddress)
		}
		queueList = append(queueList, queueMap)
	}
	return queueList
}

func buildWrapupCodes(codes []interface{}) *[]platformclientv2.Wrapupcode {
	wrapupCodes := make([]platformclientv2.Wrapupcode, 0)

	for _, code := range codes {
		codeMap := code.(map[string]interface{})
		id := codeMap["id"].(string)
		name := codeMap["name"].(string)
		dateCreatedString := codeMap["date_created"].(string)
		dateModifiedString := codeMap["date_modified"].(string)
		modifiedBy := codeMap["modified_by"].(string)
		createdBy := codeMap["created_by"].(string)

		wrapupCode := platformclientv2.Wrapupcode{
			Id:         &id,
			Name:       &name,
			ModifiedBy: &modifiedBy,
			CreatedBy:  &createdBy,
		}

		dateCreated, createdErr := time.Parse("2006-01-02T15:04:05-0700", dateCreatedString)
		if createdErr == nil {
			wrapupCode.DateCreated = &dateCreated
		}

		dateModified, modifiedErr := time.Parse("2006-01-02T15:04:05-0700", dateModifiedString)
		if modifiedErr == nil {
			wrapupCode.DateModified = &dateModified
		}

		wrapupCodes = append(wrapupCodes, wrapupCode)
	}

	return &wrapupCodes
}

func flattenWrapupCodes(codes *[]platformclientv2.Wrapupcode) []interface{} {
	if codes == nil {
		return nil
	}

	codeList := []interface{}{}

	for _, code := range *codes {
		codeMap := make(map[string]interface{})
		if code.Id != nil {
			codeMap["id"] = *code.Id
		}
		if code.Name != nil {
			codeMap["name"] = *code.Name
		}
		if code.DateCreated != nil && len(code.DateCreated.String()) > 0 {
			temp := *code.DateCreated
			codeMap["date_created"] = temp.String()
		}
		if code.DateModified != nil && len(code.DateModified.String()) > 0 {
			temp := *code.DateModified
			codeMap["date_modified"] = temp.String()
		}
		if code.ModifiedBy != nil {
			codeMap["modified_by"] = *code.ModifiedBy
		}
		if code.CreatedBy != nil {
			codeMap["created_by"] = *code.CreatedBy
		}

		codeList = append(codeList, codeMap)
	}

	return codeList
}

func buildLanguages(languages []interface{}) *[]platformclientv2.Language {
	builtLanguages := make([]platformclientv2.Language, 0)

	for _, language := range languages {
		languageMap := language.(map[string]interface{})
		id := languageMap["id"].(string)
		name := languageMap["name"].(string)
		dateModifiedString := languageMap["date_modified"].(string)
		state := languageMap["state"].(string)
		version := languageMap["version"].(string)

		temp := platformclientv2.Language{
			Id:      &id,
			Name:    &name,
			State:   &state,
			Version: &version,
		}

		dateModified, modifiedErr := time.Parse("2006-01-02T15:04:05-0700", dateModifiedString)
		if modifiedErr == nil {
			temp.DateModified = &dateModified
		}

		builtLanguages = append(builtLanguages, temp)
	}

	return &builtLanguages
}

func flattenLanguages(languages *[]platformclientv2.Language) []interface{} {
	if languages == nil {
		return nil
	}

	LanguageList := []interface{}{}

	for _, language := range *languages {
		languageMap := make(map[string]interface{})
		if language.Id != nil {
			languageMap["id"] = *language.Id
		}
		if language.Name != nil {
			languageMap["name"] = *language.Name
		}
		if language.DateModified != nil && len(language.DateModified.String()) > 0 {
			temp := *language.DateModified
			languageMap["date_modified"] = temp.String()
		}
		if language.State != nil {
			languageMap["state"] = *language.State
		}
		if language.Version != nil {
			languageMap["version"] = *language.Version
		}

		LanguageList = append(LanguageList, languageMap)
	}

	return LanguageList
}

func buildTimeSlots(slots []interface{}) *[]platformclientv2.Timeslot {
	timeSlots := make([]platformclientv2.Timeslot, 0)

	for _, slot := range slots {
		slotMap := slot.(map[string]interface{})
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

	timeAllowedMap := timeAllowed[0].(map[string]interface{})

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

	durationConditionMap := durationCondition[0].(map[string]interface{})

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

	conditionsMap := callMediaPolicyConditions[0].(map[string]interface{})
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

	return &platformclientv2.Callmediapolicyconditions{
		ForUsers:    buildUsers(conditionsMap["for_users"].([]interface{})),
		DateRanges:  &dateRanges,
		ForQueues:   buildQueues(conditionsMap["for_queues"].([]interface{})),
		WrapupCodes: buildWrapupCodes(conditionsMap["wrapup_codes"].([]interface{})),
		Languages:   buildLanguages(conditionsMap["languages"].([]interface{})),
		TimeAllowed: buildTimeAllowed(conditionsMap["time_allowed"].([]interface{})),
		Directions:  &directions,
		Duration:    buildDurationCondition(conditionsMap["duration"].([]interface{})),
	}
}

func buildChatMediaPolicyConditions(chatMediaPolicyConditions []interface{}) *platformclientv2.Chatmediapolicyconditions {
	if chatMediaPolicyConditions == nil || len(chatMediaPolicyConditions) <= 0 {
		return nil
	}

	conditionsMap := chatMediaPolicyConditions[0].(map[string]interface{})
	dateRanges := make([]string, 0)
	for _, v := range conditionsMap["date_ranges"].([]interface{}) {
		dateRange := fmt.Sprintf("%v", v)
		dateRanges = append(dateRanges, dateRange)
	}

	return &platformclientv2.Chatmediapolicyconditions{
		ForUsers:    buildUsers(conditionsMap["for_users"].([]interface{})),
		DateRanges:  &dateRanges,
		ForQueues:   buildQueues(conditionsMap["for_queues"].([]interface{})),
		WrapupCodes: buildWrapupCodes(conditionsMap["wrapup_codes"].([]interface{})),
		Languages:   buildLanguages(conditionsMap["languages"].([]interface{})),
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
		conditionsMap["for_users"] = flattenUsers(conditions.ForUsers)
	}
	if conditions.DateRanges != nil {
		conditionsMap["date_ranges"] = *conditions.DateRanges
	}
	if conditions.ForQueues != nil {
		conditionsMap["for_queues"] = flattenQueues(conditions.ForQueues)
	}
	if conditions.WrapupCodes != nil {
		conditionsMap["wrapup_codes"] = flattenWrapupCodes(conditions.WrapupCodes)
	}
	if conditions.Languages != nil {
		conditionsMap["languages"] = flattenLanguages(conditions.Languages)
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

	conditionsMap := emailMediaPolicyConditions[0].(map[string]interface{})
	dateRanges := make([]string, 0)
	for _, v := range conditionsMap["date_ranges"].([]interface{}) {
		dateRange := fmt.Sprintf("%v", v)
		dateRanges = append(dateRanges, dateRange)
	}

	return &platformclientv2.Emailmediapolicyconditions{
		ForUsers:    buildUsers(conditionsMap["for_users"].([]interface{})),
		DateRanges:  &dateRanges,
		ForQueues:   buildQueues(conditionsMap["for_queues"].([]interface{})),
		WrapupCodes: buildWrapupCodes(conditionsMap["wrapup_codes"].([]interface{})),
		Languages:   buildLanguages(conditionsMap["languages"].([]interface{})),
		TimeAllowed: buildTimeAllowed(conditionsMap["time_allowed"].([]interface{})),
	}
}

func flattenEmailMediaPolicyConditions(conditions *platformclientv2.Emailmediapolicyconditions) []interface{} {
	if conditions == nil {
		return nil
	}

	conditionsMap := make(map[string]interface{})
	if conditions.ForUsers != nil {
		conditionsMap["for_users"] = flattenUsers(conditions.ForUsers)
	}
	if conditions.DateRanges != nil {
		conditionsMap["date_ranges"] = *conditions.DateRanges
	}
	if conditions.ForQueues != nil {
		conditionsMap["for_queues"] = flattenQueues(conditions.ForQueues)
	}
	if conditions.WrapupCodes != nil {
		conditionsMap["wrapup_codes"] = flattenWrapupCodes(conditions.WrapupCodes)
	}
	if conditions.Languages != nil {
		conditionsMap["languages"] = flattenLanguages(conditions.Languages)
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

	conditionsMap := messageMediaPolicyConditions[0].(map[string]interface{})

	dateRanges := make([]string, 0)
	for _, v := range conditionsMap["date_ranges"].([]interface{}) {
		dateRange := fmt.Sprintf("%v", v)
		dateRanges = append(dateRanges, dateRange)
	}

	return &platformclientv2.Messagemediapolicyconditions{
		ForUsers:    buildUsers(conditionsMap["for_users"].([]interface{})),
		DateRanges:  &dateRanges,
		ForQueues:   buildQueues(conditionsMap["for_queues"].([]interface{})),
		WrapupCodes: buildWrapupCodes(conditionsMap["wrapup_codes"].([]interface{})),
		Languages:   buildLanguages(conditionsMap["languages"].([]interface{})),
		TimeAllowed: buildTimeAllowed(conditionsMap["time_allowed"].([]interface{})),
	}
}

func flattenMessageMediaPolicyConditions(conditions *platformclientv2.Messagemediapolicyconditions) []interface{} {
	if conditions == nil {
		return nil
	}

	conditionsMap := make(map[string]interface{})
	if conditions.ForUsers != nil {
		conditionsMap["for_users"] = flattenUsers(conditions.ForUsers)
	}
	if conditions.DateRanges != nil {
		conditionsMap["date_ranges"] = *conditions.DateRanges
	}
	if conditions.ForQueues != nil {
		conditionsMap["for_queues"] = flattenQueues(conditions.ForQueues)
	}
	if conditions.WrapupCodes != nil {
		conditionsMap["wrapup_codes"] = flattenWrapupCodes(conditions.WrapupCodes)
	}
	if conditions.Languages != nil {
		conditionsMap["languages"] = flattenLanguages(conditions.Languages)
	}
	if conditions.TimeAllowed != nil {
		conditionsMap["time_allowed"] = flattenTimeAllowed(conditions.TimeAllowed)
	}

	return []interface{}{conditionsMap}
}

func flattenCallMediaPolicyConditions(conditions *platformclientv2.Callmediapolicyconditions) []interface{} {
	if conditions == nil {
		return nil
	}

	conditionsMap := make(map[string]interface{})
	if conditions.ForUsers != nil {
		conditionsMap["for_users"] = flattenUsers(conditions.ForUsers)
	}
	if conditions.DateRanges != nil {
		conditionsMap["date_ranges"] = *conditions.DateRanges
	}
	if conditions.Directions != nil {
		conditionsMap["directions"] = *conditions.Directions
	}
	if conditions.ForQueues != nil {
		conditionsMap["for_queues"] = flattenQueues(conditions.ForQueues)
	}
	if conditions.WrapupCodes != nil {
		conditionsMap["wrapup_codes"] = flattenWrapupCodes(conditions.WrapupCodes)
	}
	if conditions.Languages != nil {
		conditionsMap["languages"] = flattenLanguages(conditions.Languages)
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

	policyMap := callMediaPolicy[0].(map[string]interface{})
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

	policyMap := chatMediaPolicy[0].(map[string]interface{})

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

	policyMap := emailMediaPolicy[0].(map[string]interface{})

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

	policyMap := messageMediaPolicy[0].(map[string]interface{})

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
	mediaPolicies := platformclientv2.Mediapolicies{}
	if len(d.Get("media_policies").([]interface{})) <= 0 {
		return nil
	}

	temp := d.Get("media_policies").([]interface{})[0].(map[string]interface{})
	if callPolicy := temp["call_policy"]; callPolicy != nil {
		mediaPolicies.CallPolicy = buildCallMediaPolicy(callPolicy.([]interface{}))
	}

	if chatPolicy := temp["chat_policy"]; chatPolicy != nil {
		mediaPolicies.ChatPolicy = buildChatMediaPolicy(chatPolicy.([]interface{}))
	}

	if emailPolicy := temp["email_policy"]; emailPolicy != nil {
		mediaPolicies.EmailPolicy = buildEmailMediaPolicy(emailPolicy.([]interface{}))
	}

	if messagePolicy := temp["message_policy"]; messagePolicy != nil {
		mediaPolicies.MessagePolicy = buildMessageMediaPolicy(messagePolicy.([]interface{}))
	}

	return &mediaPolicies

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
	if conditions, ok := d.GetOk("conditions"); ok && len(conditions.([]interface{})) > 0 {
		conditionsMap := conditions.([]interface{})[0].(map[string]interface{})

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

		return &platformclientv2.Policyconditions{
			ForUsers:    buildUsers(conditionsMap["for_users"].([]interface{})),
			Directions:  &directions,
			DateRanges:  &dateRanges,
			MediaTypes:  &mediaTypes,
			ForQueues:   buildQueues(conditionsMap["for_queues"].([]interface{})),
			Duration:    buildDurationCondition(conditionsMap["duration"].([]interface{})),
			WrapupCodes: buildWrapupCodes(conditionsMap["wrapup_codes"].([]interface{})),
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
		conditionsMap["for_users"] = flattenUsers(conditions.ForUsers)
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
		conditionsMap["for_queues"] = flattenQueues(conditions.ForQueues)
	}
	if conditions.Duration != nil {
		conditionsMap["duration"] = flattenDurationCondition(conditions.Duration)
	}
	if conditions.WrapupCodes != nil {
		conditionsMap["wrapup_codes"] = flattenWrapupCodes(conditions.WrapupCodes)
	}
	if conditions.TimeAllowed != nil {
		conditionsMap["time_allowed"] = flattenTimeAllowed(conditions.TimeAllowed)
	}

	return []interface{}{conditionsMap}
}

func buildPolicyActions2(d *schema.ResourceData) *platformclientv2.Policyactions {

	if actions, ok := d.GetOk("actions"); ok {
		actionsMap := actions.([]interface{})[0].(map[string]interface{})
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
		paramMap := param.(map[string]interface{})
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
		messageMap := message.(map[string]interface{})
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
		errorsMap := errors.([]interface{})[0].(map[string]interface{})
		return &platformclientv2.Policyerrors{
			PolicyErrorMessages: buildPolicyErrorMessages(errorsMap["policy_error_messages"].([]interface{})),
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
	sdkConfig := meta.(*providerMeta).ClientConfig
	recordingAPI := platformclientv2.NewRecordingApiWithConfig(sdkConfig)

	log.Printf("Reading media retention policy %s", d.Id())

	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		retentionPolicy, resp, getErr := recordingAPI.GetRecordingMediaretentionpolicy(d.Id())
		// fmt.Printf("media retention policy GET response data %v", retentionPolicy)

		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("failed to read media retention policy %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("failed to read media retention policy %s: %s", d.Id(), getErr))
		}

		if retentionPolicy.Name != nil {
			d.Set("name", *retentionPolicy.Name)
		}
		if retentionPolicy.ModifiedDate != nil {
			tempModifiedDate := *retentionPolicy.ModifiedDate
			d.Set("modified_date", tempModifiedDate.String())
		}
		if retentionPolicy.CreatedDate != nil {
			tempCreatedDate := *retentionPolicy.CreatedDate
			d.Set("created_date", tempCreatedDate.String())
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

		return nil
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
	sdkConfig := meta.(*providerMeta).ClientConfig
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

	if d.Get("modified_date").(string) != "" {
		modifiedDateString := d.Get("modified_date").(string)
		modifiedDate, modifiedErr := time.Parse("2006-01-02T15:04:05-0700", modifiedDateString)
		if modifiedErr == nil {
			reqBody.ModifiedDate = &modifiedDate
		}
	}

	if d.Get("created_date").(string) != "" {
		createdDateString := d.Get("created_date").(string)
		createdDate, createdErr := time.Parse("2006-01-02T15:04:05-0700", createdDateString)
		if createdErr == nil {
			reqBody.CreatedDate = &createdDate
		}
	}
	log.Printf("Creating media retention policy %s", name)
	// fmt.Printf("media retention policy POST request body %v", reqBody)

	policy, apiResponse, err := recordingAPI.PostRecordingMediaretentionpolicies(reqBody)

	fmt.Printf("Media retention policy creation status %#v", apiResponse.Status)

	if err != nil {
		return diag.Errorf("Failed to create media retention policy %s: %s", name, err)
	}

	// Make sure form is properly created
	time.Sleep(2 * time.Second)
	policyId := policy.Id
	d.SetId(*policyId)
	log.Printf("Created media retention policy %s %s", name, *policy.Id)
	return readMediaRetentionPolicy(ctx, d, meta)
}

func updateMediaRetentionPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()
	name := d.Get("name").(string)
	order := d.Get("order").(int)
	description := d.Get("description").(string)
	enabled := d.Get("enabled").(bool)

	mediaPolicies := buildMediaPolicies(d)
	conditions := buildConditions(d)
	actions := buildPolicyActions2(d)
	policyErrors := buildPolicyErrors(d)

	sdkConfig := meta.(*providerMeta).ClientConfig
	recordingAPI := platformclientv2.NewRecordingApiWithConfig(sdkConfig)

	reqBody := platformclientv2.Policy{
		Id:            &id,
		Name:          &name,
		Order:         &order,
		Description:   &description,
		Enabled:       &enabled,
		MediaPolicies: mediaPolicies,
		Conditions:    conditions,
		Actions:       actions,
		PolicyErrors:  policyErrors,
	}

	if d.Get("modified_date").(string) != "" {
		modifiedDateString := d.Get("modified_date").(string)
		modifiedDate, modifiedErr := time.Parse("2006-01-02T15:04:05-0700", modifiedDateString)
		if modifiedErr == nil {
			reqBody.ModifiedDate = &modifiedDate
		}
	}

	if d.Get("created_date").(string) != "" {
		createdDateString := d.Get("created_date").(string)
		createdDate, createdErr := time.Parse("2006-01-02T15:04:05-0700", createdDateString)
		if createdErr == nil {
			reqBody.CreatedDate = &createdDate
		}
	}

	log.Printf("Updating media retention policy %s", name)
	policy, _, err := recordingAPI.PutRecordingMediaretentionpolicy(d.Id(), reqBody)
	if err != nil {
		return diag.Errorf("Failed to update media retention policy %s", name)
	}

	log.Printf("Updated media retention policy %s %s", name, *policy.Id)
	time.Sleep(5 * time.Second)
	return readMediaRetentionPolicy(ctx, d, meta)
}

func mediaRetentionPolicyExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllMediaRetentionPolicies),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
		AllowZeroValues:  []string{"media_policies.conditions.actions.order.modified_date.created_date"},
	}
}

func deleteMediaRetentionPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	recordingAPI := platformclientv2.NewRecordingApiWithConfig(sdkConfig)

	log.Printf("Deleting media retention policy %s", name)
	if _, err := recordingAPI.DeleteRecordingMediaretentionpolicy(d.Id()); err != nil {
		return diag.Errorf("Failed to delete media retention policy %s: %s", name, err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := recordingAPI.GetRecordingMediaretentionpolicy(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// media retention policy deleted
				log.Printf("Deleted media retention policy %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error deleting media retention policy %s: %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("media retention policy %s still exists", d.Id()))
	})
}
