package genesyscloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	chunksProcess "terraform-provider-genesyscloud/genesyscloud/util/chunks"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

var (
	bullseyeExpansionTypeTimeout = "TIMEOUT_SECONDS"

	memberGroupResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"member_group_id": {
				Description: "ID (GUID) for Group, SkillGroup, Team",
				Type:        schema.TypeString,
				Required:    true,
			},
			"member_group_type": {
				Description:  "The type of the member group. Accepted values: TEAM, GROUP, SKILLGROUP",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"TEAM", "GROUP", "SKILLGROUP"}, false),
			},
		},
	}

	queueMediaSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"alerting_timeout_sec": {
				Description:  "Alerting timeout in seconds. Must be >= 7",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(7),
			},
			"enable_auto_answer": {
				Description: "Auto-Answer for digital channels(Email, Message)",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"service_level_percentage": {
				Description:  "The desired Service Level. A float value between 0 and 1.",
				Type:         schema.TypeFloat,
				Required:     true,
				ValidateFunc: validation.FloatBetween(0, 1),
			},
			"service_level_duration_ms": {
				Description:  "Service Level target in milliseconds. Must be >= 1000",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1000),
			},
		},
	}

	queueMemberResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"user_id": {
				Description: "User ID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"ring_num": {
				Description:  "Ring number between 1 and 6 for this user in the queue.",
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 6),
			},
		},
	}

	directRoutingResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"backup_queue_id": {
				Description: "Direct Routing default backup queue id.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"agent_wait_seconds": {
				Description: "The queue default time a Direct Routing interaction will wait for an agent before it goes to configured backup.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     60,
			},
			"wait_for_agent": {
				Description: "Boolean indicating if Direct Routing interactions should wait for the targeted agent by default.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"call_use_agent_address_outbound": {
				Description: "Boolean indicating if user Direct Routing addresses should be used outbound on behalf of queue in place of Queue address for calls.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"email_use_agent_address_outbound": {
				Description: "Boolean indicating if user Direct Routing addresses should be used outbound on behalf of queue in place of Queue address for emails.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"message_use_agent_address_outbound": {
				Description: "Boolean indicating if user Direct Routing addresses should be used outbound on behalf of queue in place of Queue address for messages.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
		},
	}
)

func getAllRoutingQueues(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	// Newly created resources often aren't returned unless there's a delay
	time.Sleep(5 * time.Second)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		queues, _, getErr := routingAPI.GetRoutingQueues(pageNum, pageSize, "", "", nil, nil, nil, false)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of queues: %v", getErr)
		}

		if queues.Entities == nil || len(*queues.Entities) == 0 {
			break
		}

		for _, queue := range *queues.Entities {
			resources[*queue.Id] = &resourceExporter.ResourceMeta{Name: *queue.Name}
		}
	}

	return resources, nil
}

func RoutingQueueExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllRoutingQueues),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id":                              {RefType: "genesyscloud_auth_division"},
			"queue_flow_id":                            {RefType: "genesyscloud_flow"},
			"email_in_queue_flow_id":                   {RefType: "genesyscloud_flow"},
			"message_in_queue_flow_id":                 {RefType: "genesyscloud_flow"},
			"whisper_prompt_id":                        {RefType: "genesyscloud_architect_user_prompt"},
			"outbound_messaging_sms_address_id":        {},                               // Ref type not yet defined
			"default_script_ids.*":                     {RefType: "genesyscloud_script"}, // Ref type not yet defined
			"outbound_email_address.route_id":          {RefType: "genesyscloud_routing_email_route"},
			"outbound_email_address.domain_id":         {RefType: "genesyscloud_routing_email_domain"},
			"bullseye_rings.skills_to_remove":          {RefType: "genesyscloud_routing_skill"},
			"members.user_id":                          {RefType: "genesyscloud_user"},
			"wrapup_codes":                             {RefType: "genesyscloud_routing_wrapupcode"},
			"skill_groups":                             {RefType: "genesyscloud_routing_skill_group"},
			"teams":                                    {}, //Need to add this when we get teams resources implemented
			"groups":                                   {RefType: "genesyscloud_group"},
			"conditional_group_routing_rules.queue_id": {RefType: "genesyscloud_routing_queue"},
		},
		RemoveIfMissing: map[string][]string{
			"outbound_email_address": {"route_id"},
			"members":                {"user_id"},
		},
		AllowZeroValues: []string{"bullseye_rings.expansion_timeout_seconds"},
		CustomAttributeResolver: map[string]*resourceExporter.RefAttrCustomResolver{
			"bullseye_rings.member_groups.member_group_id":           {ResolverFunc: resourceExporter.MemberGroupsResolver},
			"conditional_group_routing_rules.groups.member_group_id": {ResolverFunc: resourceExporter.MemberGroupsResolver},
		},
	}
}

func ResourceRoutingQueue() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Queue",

		CreateContext: CreateWithPooledClient(createQueue),
		ReadContext:   ReadWithPooledClient(readQueue),
		UpdateContext: UpdateWithPooledClient(updateQueue),
		DeleteContext: DeleteWithPooledClient(deleteQueue),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Queue name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"division_id": {
				Description: "The division to which this queue will belong. If not set, the home division will be used.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Description: "Queue description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"media_settings_call": {
				Description: "Call media settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        queueMediaSettingsResource,
			},
			"media_settings_callback": {
				Description: "Callback media settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        queueMediaSettingsResource,
			},
			"media_settings_chat": {
				Description: "Chat media settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        queueMediaSettingsResource,
			},
			"media_settings_email": {
				Description: "Email media settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        queueMediaSettingsResource,
			},
			"media_settings_message": {
				Description: "Message media settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        queueMediaSettingsResource,
			},
			"routing_rules": {
				Description: "The routing rules for the queue, used for routing to known or preferred agents.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    6,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"operator": {
							Description:  "Matching operator (MEETS_THRESHOLD | ANY). MEETS_THRESHOLD matches any agent with a score at or above the rule's threshold. ANY matches all specified agents, regardless of score.",
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "MEETS_THRESHOLD",
							ValidateFunc: validation.StringInSlice([]string{"MEETS_THRESHOLD", "ANY"}, false),
						},
						"threshold": {
							Description: "Threshold required for routing attempt (generally an agent score). Ignored for operator ANY.",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"wait_seconds": {
							Description:  "Seconds to wait in this rule before moving to the next.",
							Type:         schema.TypeFloat,
							Optional:     true,
							Default:      5,
							ValidateFunc: validation.FloatBetween(2, 259200),
						},
					},
				},
			},
			"bullseye_rings": {
				Description: "The bullseye ring settings for the queue.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"expansion_timeout_seconds": {
							Description:  "Seconds to wait in this ring before moving to the next.",
							Type:         schema.TypeFloat,
							Required:     true,
							ValidateFunc: validation.FloatBetween(0, 259200),
						},
						"skills_to_remove": {
							Description: "Skill IDs to remove on ring exit.",
							Type:        schema.TypeSet,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"member_groups": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     memberGroupResource,
						},
					},
				},
			},
			"conditional_group_routing_rules": {
				Description: "The Conditional Group Routing settings for the queue.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"queue_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: `The ID of the queue being evaluated for this rule. For rule 1, this is always the current queue, so should not be specified.`,
						},
						"operator": {
							Description:  "The operator that compares the actual value against the condition value. Valid values: GreaterThan, GreaterThanOrEqualTo, LessThan, LessThanOrEqualTo.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"GreaterThan", "LessThan", "GreaterThanOrEqualTo", "LessThanOrEqualTo"}, false),
						},
						"metric": {
							Description: "The queue metric being evaluated. Valid values: EstimatedWaitTime, ServiceLevel",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "EstimatedWaitTime",
						},
						"condition_value": {
							Description:  "The limit value, beyond which a rule evaluates as true.",
							Type:         schema.TypeFloat,
							Required:     true,
							ValidateFunc: validation.FloatBetween(0, 259200),
						},
						"wait_seconds": {
							Description:  "The number of seconds to wait in this rule, if it evaluates as true, before evaluating the next rule. For the final rule, this is ignored, so need not be specified.",
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      2,
							ValidateFunc: validation.IntBetween(0, 259200),
						},
						"groups": {
							Type:        schema.TypeList,
							Required:    true,
							MinItems:    1,
							Description: "The group(s) to activate if the rule evaluates as true.",
							Elem:        memberGroupResource,
						},
					},
				},
			},
			"acw_wrapup_prompt": {
				Description:  "This field controls how the UI prompts the agent for a wrapup (MANDATORY | OPTIONAL | MANDATORY_TIMEOUT | MANDATORY_FORCED_TIMEOUT | AGENT_REQUESTED).",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "MANDATORY_TIMEOUT",
				ValidateFunc: validation.StringInSlice([]string{"MANDATORY", "OPTIONAL", "MANDATORY_TIMEOUT", "MANDATORY_FORCED_TIMEOUT", "AGENT_REQUESTED"}, false),
			},
			"acw_timeout_ms": {
				Description:  "The amount of time the agent can stay in ACW. Only set when ACW is MANDATORY_TIMEOUT, MANDATORY_FORCED_TIMEOUT or AGENT_REQUESTED.",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true, // Default may be set by server
				ValidateFunc: validation.IntBetween(0, 86400000),
			},
			"skill_evaluation_method": {
				Description:  "The skill evaluation method to use when routing conversations (NONE | BEST | ALL).",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "ALL",
				ValidateFunc: validation.StringInSlice([]string{"NONE", "BEST", "ALL"}, false),
			},
			"queue_flow_id": {
				Description: "The in-queue flow ID to use for call conversations waiting in queue.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"email_in_queue_flow_id": {
				Description: "The in-queue flow ID to use for email conversations waiting in queue.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"message_in_queue_flow_id": {
				Description: "The in-queue flow ID to use for message conversations waiting in queue.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"whisper_prompt_id": {
				Description: "The prompt ID used for whisper on the queue, if configured.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"auto_answer_only": {
				Description: "Specifies whether the configured whisper should play for all ACD calls, or only for those which are auto-answered.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"enable_transcription": {
				Description: "Indicates whether voice transcription is enabled for this queue.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"suppress_in_queue_call_recording": {
				Description: "Indicates whether recording in-queue calls is suppressed for this queue.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"enable_manual_assignment": {
				Description: "Indicates whether manual assignment is enabled for this queue.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"calling_party_name": {
				Description: "The name to use for caller identification for outbound calls from this queue.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"calling_party_number": {
				Description: "The phone number to use for caller identification for outbound calls from this queue.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"default_script_ids": {
				Description:      "The default script IDs for each communication type. Communication types: (CALL | CALLBACK | CHAT | COBROWSE | EMAIL | MESSAGE | SOCIAL_EXPRESSION | VIDEO | SCREENSHARE)",
				Type:             schema.TypeMap,
				ValidateDiagFunc: validateMapCommTypes,
				Optional:         true,
				Elem:             &schema.Schema{Type: schema.TypeString},
			},
			"outbound_messaging_sms_address_id": {
				Description: "The unique ID of the outbound messaging SMS address for the queue.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"outbound_email_address": {
				Description: "The outbound email address settings for this queue.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_id": {
							Description: "Unique ID of the email domain. e.g. \"test.example.com\"",
							Type:        schema.TypeString,
							Required:    true,
						},
						"route_id": {
							Description: "Unique ID of the email route.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"members": {
				Description: "Users in the queue. If not set, this resource will not manage members.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				ConfigMode:  schema.SchemaConfigModeAttr,
				Elem:        queueMemberResource,
			},
			"wrapup_codes": {
				Description: "IDs of wrapup codes assigned to this queue. If not set, this resource will not manage wrapup codes.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"direct_routing": {
				Description: "Used by the System to set Direct Routing settings for a system Direct Routing queue.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        directRoutingResource,
			},
			"skill_groups": {
				Description: "List of skill group ids assigned to the queue.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"groups": {
				Description: "List of group ids assigned to the queue",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"teams": {
				Description: "List of ids assigned to the queue",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func createQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	divisionID := d.Get("division_id").(string)
	description := d.Get("description").(string)
	skillEvaluationMethod := d.Get("skill_evaluation_method").(string)
	autoAnswerOnly := d.Get("auto_answer_only").(bool)
	enableTranscription := d.Get("enable_transcription").(bool)
	suppressInQueueCallRecording := d.Get("suppress_in_queue_call_recording").(bool)
	enableManualAssignment := d.Get("enable_manual_assignment").(bool)
	callingPartyName := d.Get("calling_party_name").(string)
	callingPartyNumber := d.Get("calling_party_number").(string)
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	skillGroups := buildMemberGroupList(d, "skill_groups", "SKILLGROUP")
	groups := buildMemberGroupList(d, "groups", "GROUP")
	teams := buildMemberGroupList(d, "teams", "TEAM")
	memberGroups := append(*skillGroups, *groups...)
	memberGroups = append(memberGroups, *teams...)

	conditionalGroupRouting, diagErr := buildSdkConditionalGroupRouting(d)
	if diagErr != nil {
		return diagErr
	}

	createQueue := platformclientv2.Createqueuerequest{
		Name:                         &name,
		Description:                  &description,
		MediaSettings:                buildSdkMediaSettings(d),
		RoutingRules:                 buildSdkRoutingRules(d),
		Bullseye:                     buildSdkBullseyeSettings(d),
		ConditionalGroupRouting:      conditionalGroupRouting,
		AcwSettings:                  buildSdkAcwSettings(d),
		SkillEvaluationMethod:        &skillEvaluationMethod,
		QueueFlow:                    BuildSdkDomainEntityRef(d, "queue_flow_id"),
		EmailInQueueFlow:             BuildSdkDomainEntityRef(d, "email_in_queue_flow_id"),
		MessageInQueueFlow:           BuildSdkDomainEntityRef(d, "message_in_queue_flow_id"),
		WhisperPrompt:                BuildSdkDomainEntityRef(d, "whisper_prompt_id"),
		AutoAnswerOnly:               &autoAnswerOnly,
		CallingPartyName:             &callingPartyName,
		CallingPartyNumber:           &callingPartyNumber,
		DefaultScripts:               buildSdkDefaultScriptsMap(d),
		OutboundMessagingAddresses:   buildSdkQueueMessagingAddresses(d),
		OutboundEmailAddress:         buildSdkQueueEmailAddress(d),
		EnableTranscription:          &enableTranscription,
		SuppressInQueueCallRecording: &suppressInQueueCallRecording,
		EnableManualAssignment:       &enableManualAssignment,
		DirectRouting:                buildSdkDirectRouting(d),
		MemberGroups:                 &memberGroups,
	}

	if divisionID != "" {
		createQueue.Division = &platformclientv2.Writabledivision{Id: &divisionID}
	}

	log.Printf("Creating queue %s", name)
	queue, _, err := routingAPI.PostRoutingQueues(createQueue)
	if err != nil {
		return diag.Errorf("Failed to create queue %s: %s", name, err)
	}
	d.SetId(*queue.Id)

	diagErr = updateQueueMembers(d, routingAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateQueueWrapupCodes(d, routingAPI)
	if diagErr != nil {
		return diagErr
	}

	return readQueue(ctx, d, meta)
}

func readQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Reading queue %s", d.Id())
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		currentQueue, resp, getErr := routingAPI.GetRoutingQueue(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read queue %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read queue %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingQueue())
		if currentQueue.Name != nil {
			d.Set("name", *currentQueue.Name)
		} else {
			d.Set("name", nil)
		}

		if currentQueue.Division != nil && currentQueue.Division.Id != nil {
			d.Set("division_id", *currentQueue.Division.Id)
		} else {
			d.Set("division_id", nil)
		}

		if currentQueue.Description != nil {
			d.Set("description", *currentQueue.Description)
		} else {
			d.Set("description", nil)
		}

		d.Set("acw_wrapup_prompt", nil)
		d.Set("acw_timeout_ms", nil)
		if currentQueue.AcwSettings != nil {
			if currentQueue.AcwSettings.WrapupPrompt != nil {
				d.Set("acw_wrapup_prompt", *currentQueue.AcwSettings.WrapupPrompt)
			}
			if currentQueue.AcwSettings.TimeoutMs != nil {
				d.Set("acw_timeout_ms", int(*currentQueue.AcwSettings.TimeoutMs))
			}
		}

		if currentQueue.SkillEvaluationMethod != nil {
			d.Set("skill_evaluation_method", *currentQueue.SkillEvaluationMethod)
		} else {
			d.Set("skill_evaluation_method", nil)
		}

		d.Set("media_settings_call", nil)
		d.Set("media_settings_callback", nil)
		d.Set("media_settings_chat", nil)
		d.Set("media_settings_email", nil)
		d.Set("media_settings_message", nil)

		if currentQueue.MediaSettings != nil {
			if currentQueue.MediaSettings.Call != nil {
				d.Set("media_settings_call", flattenMediaSetting(*currentQueue.MediaSettings.Call))
			}

			if currentQueue.MediaSettings.Callback != nil {
				d.Set("media_settings_callback", flattenMediaSettingCallback(*currentQueue.MediaSettings.Callback))
			}

			if currentQueue.MediaSettings.Chat != nil {
				d.Set("media_settings_chat", flattenMediaSetting(*currentQueue.MediaSettings.Chat))
			}

			if currentQueue.MediaSettings.Email != nil {
				d.Set("media_settings_email", flattenMediaSetting(*currentQueue.MediaSettings.Email))
			}

			if currentQueue.MediaSettings.Message != nil {
				d.Set("media_settings_message", flattenMediaSetting(*currentQueue.MediaSettings.Message))
			}

		}

		if currentQueue.RoutingRules != nil {
			d.Set("routing_rules", flattenRoutingRules(*currentQueue.RoutingRules))
		} else {
			d.Set("routing_rules", nil)
		}

		if currentQueue.Bullseye != nil && currentQueue.Bullseye.Rings != nil {
			d.Set("bullseye_rings", flattenBullseyeRings(*currentQueue.Bullseye.Rings))
		} else {
			d.Set("bullseye_rings", nil)
		}

		if currentQueue.QueueFlow != nil && currentQueue.QueueFlow.Id != nil {
			d.Set("queue_flow_id", *currentQueue.QueueFlow.Id)
		} else {
			d.Set("queue_flow_id", nil)
		}

		if currentQueue.MessageInQueueFlow != nil && currentQueue.MessageInQueueFlow.Id != nil {
			d.Set("message_in_queue_flow_id", *currentQueue.MessageInQueueFlow.Id)
		} else {
			d.Set("message_in_queue_flow_id", nil)
		}

		if currentQueue.EmailInQueueFlow != nil && currentQueue.EmailInQueueFlow.Id != nil {
			d.Set("email_in_queue_flow_id", *currentQueue.EmailInQueueFlow.Id)
		} else {
			d.Set("email_in_queue_flow_id", nil)
		}

		if currentQueue.WhisperPrompt != nil && currentQueue.WhisperPrompt.Id != nil {
			d.Set("whisper_prompt_id", *currentQueue.WhisperPrompt.Id)
		} else {
			d.Set("whisper_prompt_id", nil)
		}

		if currentQueue.AutoAnswerOnly != nil {
			d.Set("auto_answer_only", *currentQueue.AutoAnswerOnly)
		} else {
			d.Set("auto_answer_only", nil)
		}

		if currentQueue.EnableTranscription != nil {
			d.Set("enable_transcription", *currentQueue.EnableTranscription)
		} else {
			d.Set("enable_transcription", nil)
		}

		if currentQueue.SuppressInQueueCallRecording != nil {
			d.Set("suppress_in_queue_call_recording", *currentQueue.SuppressInQueueCallRecording)
		} else {
			d.Set("suppress_in_queue_call_recording", nil)
		}

		if currentQueue.EnableManualAssignment != nil {
			d.Set("enable_manual_assignment", *currentQueue.EnableManualAssignment)
		} else {
			d.Set("enable_manual_assignment", nil)
		}

		if currentQueue.CallingPartyName != nil {
			d.Set("calling_party_name", *currentQueue.CallingPartyName)
		} else {
			d.Set("calling_party_name", nil)
		}

		if currentQueue.CallingPartyNumber != nil {
			d.Set("calling_party_number", *currentQueue.CallingPartyNumber)
		} else {
			d.Set("calling_party_number", nil)
		}

		if currentQueue.DefaultScripts != nil {
			d.Set("default_script_ids", flattenDefaultScripts(*currentQueue.DefaultScripts))
		} else {
			d.Set("default_script_ids", nil)
		}

		if currentQueue.OutboundMessagingAddresses != nil && currentQueue.OutboundMessagingAddresses.SmsAddress != nil {
			d.Set("outbound_messaging_sms_address_id", *currentQueue.OutboundMessagingAddresses.SmsAddress.Id)
		} else {
			d.Set("outbound_messaging_sms_address_id", nil)
		}

		if currentQueue.OutboundEmailAddress != nil && *currentQueue.OutboundEmailAddress != nil {
			outboundEmailAddress := *currentQueue.OutboundEmailAddress
			d.Set("outbound_email_address", []interface{}{flattenQueueEmailAddress(*outboundEmailAddress)})
		} else {
			d.Set("outbound_email_address", nil)
		}

		if currentQueue.DirectRouting != nil {
			d.Set("direct_routing", []interface{}{flattenDirectRouting(*currentQueue.DirectRouting)})
		} else {
			d.Set("direct_routing", nil)
		}

		wrapupCodes, err := flattenQueueWrapupCodes(d.Id(), routingAPI)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("%v", err))
		}
		d.Set("wrapup_codes", wrapupCodes)

		members, err := flattenQueueMembers(d.Id(), "user", routingAPI)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("%v", err))
		}
		d.Set("members", members)

		skillgroup := "SKILLGROUP"
		team := "TEAM"
		group := "GROUP"

		d.Set("skill_groups", flattenQueueMemberGroupsList(currentQueue, &skillgroup))
		d.Set("teams", flattenQueueMemberGroupsList(currentQueue, &team))
		d.Set("groups", flattenQueueMemberGroupsList(currentQueue, &group))

		d.Set("conditional_group_routing_rules", flattenConditionalGroupRoutingRules(currentQueue))

		log.Printf("Done reading queue %s %s", d.Id(), *currentQueue.Name)
		return cc.CheckState()
	})
}

func updateQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	skillEvaluationMethod := d.Get("skill_evaluation_method").(string)
	autoAnswerOnly := d.Get("auto_answer_only").(bool)
	enableTranscription := d.Get("enable_transcription").(bool)
	suppressInQueueCallRecording := d.Get("suppress_in_queue_call_recording").(bool)
	enableManualAssignment := d.Get("enable_manual_assignment").(bool)
	callingPartyName := d.Get("calling_party_name").(string)
	callingPartyNumber := d.Get("calling_party_number").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	skillGroups := buildMemberGroupList(d, "skill_groups", "SKILLGROUP")
	groups := buildMemberGroupList(d, "groups", "GROUP")
	teams := buildMemberGroupList(d, "teams", "TEAM")
	memberGroups := append(*skillGroups, *groups...)
	memberGroups = append(memberGroups, *teams...)

	conditionalGroupRouting, diagErr := buildSdkConditionalGroupRouting(d)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updating queue %s", name)

	_, _, err := routingAPI.PutRoutingQueue(d.Id(), platformclientv2.Queuerequest{
		Name:                         &name,
		Description:                  &description,
		MediaSettings:                buildSdkMediaSettings(d),
		RoutingRules:                 buildSdkRoutingRules(d),
		Bullseye:                     buildSdkBullseyeSettings(d),
		ConditionalGroupRouting:      conditionalGroupRouting,
		AcwSettings:                  buildSdkAcwSettings(d),
		SkillEvaluationMethod:        &skillEvaluationMethod,
		QueueFlow:                    BuildSdkDomainEntityRef(d, "queue_flow_id"),
		EmailInQueueFlow:             BuildSdkDomainEntityRef(d, "email_in_queue_flow_id"),
		MessageInQueueFlow:           BuildSdkDomainEntityRef(d, "message_in_queue_flow_id"),
		WhisperPrompt:                BuildSdkDomainEntityRef(d, "whisper_prompt_id"),
		AutoAnswerOnly:               &autoAnswerOnly,
		CallingPartyName:             &callingPartyName,
		CallingPartyNumber:           &callingPartyNumber,
		DefaultScripts:               buildSdkDefaultScriptsMap(d),
		OutboundMessagingAddresses:   buildSdkQueueMessagingAddresses(d),
		OutboundEmailAddress:         buildSdkQueueEmailAddress(d),
		EnableTranscription:          &enableTranscription,
		SuppressInQueueCallRecording: &suppressInQueueCallRecording,
		EnableManualAssignment:       &enableManualAssignment,
		DirectRouting:                buildSdkDirectRouting(d),
		MemberGroups:                 &memberGroups,
	})

	if err != nil {
		return diag.Errorf("Error updating queue %s: %s", name, err)
	}

	diagErr = updateObjectDivision(d, "QUEUE", sdkConfig)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateQueueMembers(d, routingAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateQueueWrapupCodes(d, routingAPI)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished updating queue %s", name)
	return readQueue(ctx, d, meta)
}

func deleteQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Deleting queue %s", name)
	_, err := routingAPI.DeleteRoutingQueue(d.Id(), true)
	if err != nil {
		return diag.Errorf("Failed to delete queue %s: %s", name, err)
	}

	// Queue deletes are not immediate. Query until queue is no longer found
	// Add a delay before the first request to reduce the liklihood of public API's cache
	// re-populating the queue after the delete. Otherwise it may not expire for a minute.
	time.Sleep(5 * time.Second)

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := routingAPI.GetRoutingQueue(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Queue deleted
				log.Printf("Queue %s deleted", name)
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting queue %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("Queue %s still exists", d.Id()))
	})
}

func buildSdkMediaSettings(d *schema.ResourceData) *platformclientv2.Queuemediasettings {
	queueMediaSettings := &platformclientv2.Queuemediasettings{}

	mediaSettingsCall := d.Get("media_settings_call").([]interface{})
	if mediaSettingsCall != nil && len(mediaSettingsCall) > 0 {
		queueMediaSettings.Call = buildSdkMediaSetting(mediaSettingsCall)

	}

	mediaSettingsCallback := d.Get("media_settings_callback").([]interface{})
	if mediaSettingsCallback != nil && len(mediaSettingsCallback) > 0 {
		queueMediaSettings.Callback = buildSdkMediaSettingCallback(mediaSettingsCallback)
	}

	mediaSettingsChat := d.Get("media_settings_chat").([]interface{})
	if mediaSettingsChat != nil && len(mediaSettingsChat) > 0 {
		queueMediaSettings.Chat = buildSdkMediaSetting(mediaSettingsChat)
	}

	mediaSettingsEmail := d.Get("media_settings_email").([]interface{})
	log.Printf("The media settings email #%v", mediaSettingsEmail)
	if mediaSettingsEmail != nil && len(mediaSettingsEmail) > 0 {
		queueMediaSettings.Email = buildSdkMediaSetting(mediaSettingsEmail)
	}

	mediaSettingsMessage := d.Get("media_settings_message").([]interface{})
	if mediaSettingsMessage != nil && len(mediaSettingsMessage) > 0 {
		queueMediaSettings.Message = buildSdkMediaSetting(mediaSettingsMessage)
	}

	return queueMediaSettings
}

func buildSdkMediaSetting(settings []interface{}) *platformclientv2.Mediasettings {
	settingsMap := settings[0].(map[string]interface{})

	alertingTimeout := settingsMap["alerting_timeout_sec"].(int)
	enableAutoAnswer := settingsMap["enable_auto_answer"].(bool)
	serviceLevelPct := settingsMap["service_level_percentage"].(float64)
	serviceLevelDur := settingsMap["service_level_duration_ms"].(int)

	return &platformclientv2.Mediasettings{
		AlertingTimeoutSeconds: &alertingTimeout,
		EnableAutoAnswer:       &enableAutoAnswer,
		ServiceLevel: &platformclientv2.Servicelevel{
			Percentage: &serviceLevelPct,
			DurationMs: &serviceLevelDur,
		},
	}
}

func buildSdkMediaSettingCallback(settings []interface{}) *platformclientv2.Callbackmediasettings {
	settingsMap := settings[0].(map[string]interface{})

	alertingTimeout := settingsMap["alerting_timeout_sec"].(int)
	serviceLevelPct := settingsMap["service_level_percentage"].(float64)
	serviceLevelDur := settingsMap["service_level_duration_ms"].(int)

	return &platformclientv2.Callbackmediasettings{
		AlertingTimeoutSeconds: &alertingTimeout,
		ServiceLevel: &platformclientv2.Servicelevel{
			Percentage: &serviceLevelPct,
			DurationMs: &serviceLevelDur,
		},
	}
}

func flattenMediaSetting(settings platformclientv2.Mediasettings) []interface{} {
	settingsMap := make(map[string]interface{})
	settingsMap["alerting_timeout_sec"] = *settings.AlertingTimeoutSeconds

	if settings.EnableAutoAnswer != nil {
		settingsMap["enable_auto_answer"] = *settings.EnableAutoAnswer
	} else {
		settingsMap["enable_auto_answer"] = false
	}

	settingsMap["service_level_percentage"] = *settings.ServiceLevel.Percentage
	settingsMap["service_level_duration_ms"] = *settings.ServiceLevel.DurationMs
	return []interface{}{settingsMap}
}

func flattenMediaSettingCallback(settings platformclientv2.Callbackmediasettings) []interface{} {
	settingsMap := make(map[string]interface{})

	settingsMap["alerting_timeout_sec"] = *settings.AlertingTimeoutSeconds
	settingsMap["service_level_percentage"] = *settings.ServiceLevel.Percentage
	settingsMap["service_level_duration_ms"] = *settings.ServiceLevel.DurationMs
	return []interface{}{settingsMap}
}

func buildMemberGroupList(d *schema.ResourceData, groupKey string, groupType string) *[]platformclientv2.Membergroup {
	var memberGroups []platformclientv2.Membergroup
	if mg, ok := d.GetOk(groupKey); ok {

		for _, mgId := range mg.(*schema.Set).List() {
			id := mgId.(string)

			memberGroup := &platformclientv2.Membergroup{Id: &id, VarType: &groupType}
			memberGroups = append(memberGroups, *memberGroup)
		}
	}

	return &memberGroups
}

func buildSdkRoutingRules(d *schema.ResourceData) *[]platformclientv2.Routingrule {
	var routingRules []platformclientv2.Routingrule
	if configRoutingRules, ok := d.GetOk("routing_rules"); ok {
		for _, configRule := range configRoutingRules.([]interface{}) {
			ruleSettings, ok := configRule.(map[string]interface{})
			if !ok {
				continue
			}
			var sdkRule platformclientv2.Routingrule
			if operator, ok := ruleSettings["operator"].(string); ok {
				sdkRule.Operator = &operator
			}
			if threshold, ok := ruleSettings["threshold"]; ok {
				v := threshold.(int)
				sdkRule.Threshold = &v
			}
			if waitSeconds, ok := ruleSettings["wait_seconds"].(float64); ok {
				sdkRule.WaitSeconds = &waitSeconds
			}
			routingRules = append(routingRules, sdkRule)
		}
	}
	return &routingRules
}

func flattenRoutingRules(sdkRoutingRules []platformclientv2.Routingrule) []interface{} {
	rules := make([]interface{}, len(sdkRoutingRules))
	for i, sdkRule := range sdkRoutingRules {
		ruleSettings := make(map[string]interface{})
		if sdkRule.Operator != nil {
			ruleSettings["operator"] = *sdkRule.Operator
		}
		if sdkRule.Threshold != nil {
			ruleSettings["threshold"] = *sdkRule.Threshold
		}
		if sdkRule.WaitSeconds != nil {
			ruleSettings["wait_seconds"] = *sdkRule.WaitSeconds
		}
		rules[i] = ruleSettings
	}
	return rules
}

func buildSdkBullseyeSettings(d *schema.ResourceData) *platformclientv2.Bullseye {
	if configRings, ok := d.GetOk("bullseye_rings"); ok {
		var sdkRings []platformclientv2.Ring
		for _, configRing := range configRings.([]interface{}) {
			ringSettings, ok := configRing.(map[string]interface{})
			if !ok {
				continue
			}
			var sdkRing platformclientv2.Ring

			if waitSeconds, ok := ringSettings["expansion_timeout_seconds"].(float64); ok {
				sdkRing.ExpansionCriteria = &[]platformclientv2.Expansioncriterium{
					{
						VarType:   &bullseyeExpansionTypeTimeout,
						Threshold: &waitSeconds,
					},
				}
			}

			if skillsToRemove, ok := ringSettings["skills_to_remove"]; ok {
				skillIds := skillsToRemove.(*schema.Set).List()
				if len(skillIds) > 0 {
					sdkSkillsToRemove := make([]platformclientv2.Skillstoremove, len(skillIds))
					for i, id := range skillIds {
						skillID := id.(string)
						sdkSkillsToRemove[i] = platformclientv2.Skillstoremove{
							Id: &skillID,
						}
					}
					sdkRing.Actions = &platformclientv2.Actions{
						SkillsToRemove: &sdkSkillsToRemove,
					}
				}
			}

			if memberGroups, ok := ringSettings["member_groups"]; ok {
				memberGroupList := memberGroups.(*schema.Set).List()
				if len(memberGroupList) > 0 {

					sdkMemberGroups := make([]platformclientv2.Membergroup, len(memberGroupList))
					for i, memberGroup := range memberGroupList {
						settingsMap := memberGroup.(map[string]interface{})
						memberGroupID := settingsMap["member_group_id"].(string)
						memberGroupType := settingsMap["member_group_type"].(string)

						sdkMemberGroups[i] = platformclientv2.Membergroup{
							Id:      &memberGroupID,
							VarType: &memberGroupType,
						}
					}
					sdkRing.MemberGroups = &sdkMemberGroups
				}
			}
			sdkRings = append(sdkRings, sdkRing)
		}

		/*
			The routing queues API is a little unusual.  You can have up to six bullseye routing rings but the last one is always
			a treated as the default ring.  This means you can actually ony define a maximum of 5.  So, I have changed the behavior of this
			resource to only allow you to add 5 items and then the code always adds a 6 item (see the code below) with a default timeout of 2.
		*/
		var defaultSdkRing platformclientv2.Ring
		defaultTimeoutInt := 2
		defaultTimeoutFloat := float64(defaultTimeoutInt)
		defaultSdkRing.ExpansionCriteria = &[]platformclientv2.Expansioncriterium{
			{
				VarType:   &bullseyeExpansionTypeTimeout,
				Threshold: &defaultTimeoutFloat,
			},
		}

		sdkRings = append(sdkRings, defaultSdkRing)

		return &platformclientv2.Bullseye{Rings: &sdkRings}
	}
	return nil
}

/*
The flattenBullseyeRings function maps the data retrieved from our SDK call over to the bullseye_ring attribute within the provider.
You might notice in the code that we are always mapping all but the last item in the list of rings retrieved by the API.  The reason for this
is that when you submit a list of bullseye_rings to the API, the API will always take the last item in the list and use it to drive default behavior
This is a change from earlier parts of the API where you could define 6 bullseye rings and there would always be six.  Now when you define bullseye rings,
the public API will take the list item in the list and make it the default and it will not show up on the screen.  To get around this you needed
to always add a dumb bullseye ring block.  Now, we automatically add one for you.  We only except a maximum of 5 bullseyes_ring blocks, but we will always
remove the last block returned by the API.
*/
func flattenBullseyeRings(sdkRings []platformclientv2.Ring) []interface{} {
	rings := make([]interface{}, len(sdkRings)-1) //Sizing the target array of Rings to account for us removing the default block
	for i, sdkRing := range sdkRings {
		if i < len(sdkRings)-1 { //Checking to make sure we are do nothing with the last item in the list by skipping processing if it is defined
			ringSettings := make(map[string]interface{})
			if sdkRing.ExpansionCriteria != nil {
				for _, criteria := range *sdkRing.ExpansionCriteria {
					if *criteria.VarType == bullseyeExpansionTypeTimeout {
						ringSettings["expansion_timeout_seconds"] = *criteria.Threshold
						break
					}
				}
			}

			if sdkRing.Actions != nil && sdkRing.Actions.SkillsToRemove != nil {
				skillIds := make([]interface{}, len(*sdkRing.Actions.SkillsToRemove))
				for s, skill := range *sdkRing.Actions.SkillsToRemove {
					skillIds[s] = *skill.Id
				}
				ringSettings["skills_to_remove"] = schema.NewSet(schema.HashString, skillIds)
			}

			if sdkRing.MemberGroups != nil {
				memberGroups := schema.NewSet(schema.HashResource(memberGroupResource), []interface{}{})

				for _, memberGroup := range *sdkRing.MemberGroups {
					memberGroupMap := make(map[string]interface{})
					memberGroupMap["member_group_id"] = *memberGroup.Id
					memberGroupMap["member_group_type"] = *memberGroup.VarType

					memberGroups.Add(memberGroupMap)
				}

				ringSettings["member_groups"] = memberGroups
			}
			rings[i] = ringSettings
		}
	}
	return rings
}

func buildSdkConditionalGroupRouting(d *schema.ResourceData) (*platformclientv2.Conditionalgrouprouting, diag.Diagnostics) {
	if configRules, ok := d.GetOk("conditional_group_routing_rules"); ok {
		var sdkCGRRules []platformclientv2.Conditionalgrouproutingrule
		for i, configRules := range configRules.([]interface{}) {
			ruleSettings, ok := configRules.(map[string]interface{})
			if !ok {
				continue
			}
			var sdkCGRRule platformclientv2.Conditionalgrouproutingrule

			if waitSeconds, ok := ruleSettings["wait_seconds"].(int); ok {
				sdkCGRRule.WaitSeconds = &waitSeconds
			}

			if operator, ok := ruleSettings["operator"].(string); ok {
				sdkCGRRule.Operator = &operator
			}

			if conditionValue, ok := ruleSettings["condition_value"].(float64); ok {
				sdkCGRRule.ConditionValue = &conditionValue
			}

			if metric, ok := ruleSettings["metric"].(string); ok {
				sdkCGRRule.Metric = &metric
			}

			if queueId, ok := ruleSettings["queue_id"].(string); ok && queueId != "" {
				if i == 0 {
					return nil, diag.Errorf("For rule 1, queue_id is always assumed to be the current queue, so should not be specified.")
				}
				sdkCGRRule.Queue = &platformclientv2.Domainentityref{Id: &queueId}
			}

			if memberGroupList, ok := ruleSettings["groups"].([]interface{}); ok {
				if len(memberGroupList) > 0 {
					sdkMemberGroups := make([]platformclientv2.Membergroup, len(memberGroupList))
					for i, memberGroup := range memberGroupList {
						settingsMap, ok := memberGroup.(map[string]interface{})
						if !ok {
							continue
						}
						memberGroupID := settingsMap["member_group_id"].(string)
						memberGroupType := settingsMap["member_group_type"].(string)

						sdkMemberGroups[i] = platformclientv2.Membergroup{
							Id:      &memberGroupID,
							VarType: &memberGroupType,
						}
					}
					sdkCGRRule.Groups = &sdkMemberGroups
				}
			}
			sdkCGRRules = append(sdkCGRRules, sdkCGRRule)
		}
		rules := &sdkCGRRules
		return &platformclientv2.Conditionalgrouprouting{Rules: rules}, nil
	}
	return nil, nil
}

func flattenConditionalGroupRoutingRules(queue *platformclientv2.Queue) []interface{} {
	if queue.ConditionalGroupRouting == nil || len(*queue.ConditionalGroupRouting.Rules) == 0 {
		return nil
	}

	rules := make([]interface{}, len(*queue.ConditionalGroupRouting.Rules))
	for i, rule := range *queue.ConditionalGroupRouting.Rules {
		ruleSettings := make(map[string]interface{})

		if rule.WaitSeconds != nil {
			ruleSettings["wait_seconds"] = *rule.WaitSeconds
		}

		if rule.Operator != nil {
			ruleSettings["operator"] = *rule.Operator
		}

		if rule.ConditionValue != nil {
			ruleSettings["condition_value"] = *rule.ConditionValue
		}

		if rule.Metric != nil {
			ruleSettings["metric"] = *rule.Metric
		}

		// The first rule is assumed to apply to this queue, so queue_id should be omitted from the first rule on queue creation
		// Hence it should not be read in either
		if rule.Queue != nil && *rule.Queue.Id != *queue.Id {
			ruleSettings["queue_id"] = *rule.Queue.Id
		}

		if rule.Groups != nil {
			memberGroups := make([]interface{}, 0)
			for _, group := range *rule.Groups {
				memberGroupMap := make(map[string]interface{})
				if group.Id != nil {
					memberGroupMap["member_group_id"] = *group.Id
				}
				if group.VarType != nil {
					memberGroupMap["member_group_type"] = *group.VarType
				}
				memberGroups = append(memberGroups, memberGroupMap)
			}
			ruleSettings["groups"] = memberGroups
		}

		rules[i] = ruleSettings
	}

	return rules
}

func buildSdkAcwSettings(d *schema.ResourceData) *platformclientv2.Acwsettings {
	acwWrapupPrompt := d.Get("acw_wrapup_prompt").(string)

	acwSettings := platformclientv2.Acwsettings{
		WrapupPrompt: &acwWrapupPrompt, // Set or default
	}

	// Only set timeout for certain wrapup prompt types
	if acwWrapupPrompt == "MANDATORY_TIMEOUT" || acwWrapupPrompt == "MANDATORY_FORCED_TIMEOUT" || acwWrapupPrompt == "AGENT_REQUESTED" {
		acwTimeoutMs, hasTimeout := d.GetOk("acw_timeout_ms")
		if hasTimeout {
			timeout := acwTimeoutMs.(int)
			acwSettings.TimeoutMs = &timeout
		}
	}
	return &acwSettings
}

func buildSdkDefaultScriptsMap(d *schema.ResourceData) *map[string]platformclientv2.Script {
	if scriptIds, ok := d.GetOk("default_script_ids"); ok {
		scriptMap := scriptIds.(map[string]interface{})

		results := make(map[string]platformclientv2.Script)
		for k, v := range scriptMap {
			scriptID := v.(string)
			results[k] = platformclientv2.Script{Id: &scriptID}
		}
		return &results
	}
	return nil
}

func flattenDefaultScripts(sdkScripts map[string]platformclientv2.Script) map[string]interface{} {
	if len(sdkScripts) == 0 {
		return nil
	}

	results := make(map[string]interface{})
	for k, v := range sdkScripts {
		results[k] = *v.Id
	}
	return results
}

func validateMapCommTypes(val interface{}, _ cty.Path) diag.Diagnostics {
	if val == nil {
		return nil
	}

	commTypes := []string{"CALL", "CALLBACK", "CHAT", "COBROWSE", "EMAIL", "MESSAGE", "SOCIAL_EXPRESSION", "VIDEO", "SCREENSHARE"}
	m := val.(map[string]interface{})
	for k := range m {
		if !lists.ItemInSlice(k, commTypes) {
			return diag.Errorf("%s is an invalid communication type key.", k)
		}
	}
	return nil
}

func buildSdkQueueMessagingAddresses(d *schema.ResourceData) *platformclientv2.Queuemessagingaddresses {
	if _, ok := d.GetOk("outbound_messaging_sms_address_id"); ok {
		return &platformclientv2.Queuemessagingaddresses{
			SmsAddress: BuildSdkDomainEntityRef(d, "outbound_messaging_sms_address_id"),
		}
	}
	return nil
}

func buildSdkQueueEmailAddress(d *schema.ResourceData) *platformclientv2.Queueemailaddress {
	outboundEmailAddress := d.Get("outbound_email_address").([]interface{})
	if outboundEmailAddress != nil && len(outboundEmailAddress) > 0 {
		settingsMap := outboundEmailAddress[0].(map[string]interface{})

		domainID := settingsMap["domain_id"].(string)
		routeID := settingsMap["route_id"].(string)
		inboundRoute := &platformclientv2.Inboundroute{
			Id: &routeID,
		}
		return &platformclientv2.Queueemailaddress{
			Domain: &platformclientv2.Domainentityref{Id: &domainID},
			Route:  &inboundRoute,
		}
	}
	return nil
}

func flattenQueueEmailAddress(settings platformclientv2.Queueemailaddress) map[string]interface{} {
	settingsMap := make(map[string]interface{})
	if settings.Domain != nil {
		settingsMap["domain_id"] = *settings.Domain.Id
	}
	if settings.Route != nil {
		route := *settings.Route
		settingsMap["route_id"] = *route.Id
	}

	return settingsMap
}

func buildSdkDirectRouting(d *schema.ResourceData) *platformclientv2.Directrouting {
	directRouting := d.Get("direct_routing").([]interface{})
	if directRouting != nil && len(directRouting) > 0 {
		settingsMap := directRouting[0].(map[string]interface{})

		backupQueueID := settingsMap["backup_queue_id"].(string)
		agentWaitSeconds := settingsMap["agent_wait_seconds"].(int)
		waitForAgent := settingsMap["wait_for_agent"].(bool)

		callUseAgentAddressOutbound := settingsMap["call_use_agent_address_outbound"].(bool)
		callSettings := &platformclientv2.Directroutingmediasettings{
			UseAgentAddressOutbound: &callUseAgentAddressOutbound,
		}

		emailUseAgentAddressOutbound := settingsMap["email_use_agent_address_outbound"].(bool)
		emailSettings := &platformclientv2.Directroutingmediasettings{
			UseAgentAddressOutbound: &emailUseAgentAddressOutbound,
		}

		messageUseAgentAddressOutbound := settingsMap["message_use_agent_address_outbound"].(bool)
		messageSettings := &platformclientv2.Directroutingmediasettings{
			UseAgentAddressOutbound: &messageUseAgentAddressOutbound,
		}

		return &platformclientv2.Directrouting{
			CallMediaSettings:    callSettings,
			EmailMediaSettings:   emailSettings,
			MessageMediaSettings: messageSettings,
			WaitForAgent:         &waitForAgent,
			AgentWaitSeconds:     &agentWaitSeconds,
			BackupQueueId:        &backupQueueID,
		}
	}
	return nil
}

func flattenDirectRouting(settings platformclientv2.Directrouting) map[string]interface{} {
	settingsMap := make(map[string]interface{})
	if settings.BackupQueueId != nil {
		settingsMap["backup_queue_id"] = *settings.BackupQueueId
	}
	if settings.AgentWaitSeconds != nil {
		settingsMap["agent_wait_seconds"] = *settings.AgentWaitSeconds
	}
	if settings.WaitForAgent != nil {
		settingsMap["wait_for_agent"] = *settings.WaitForAgent
	}
	if settings.CallMediaSettings != nil {
		callSettings := *settings.CallMediaSettings
		settingsMap["call_use_agent_address_outbound"] = *callSettings.UseAgentAddressOutbound
	}
	if settings.EmailMediaSettings != nil {
		emailSettings := *settings.EmailMediaSettings
		settingsMap["email_use_agent_address_outbound"] = *emailSettings.UseAgentAddressOutbound
	}
	if settings.MessageMediaSettings != nil {
		messageSettings := *settings.MessageMediaSettings
		settingsMap["message_use_agent_address_outbound"] = *messageSettings.UseAgentAddressOutbound
	}

	return settingsMap
}

func updateQueueWrapupCodes(d *schema.ResourceData, routingAPI *platformclientv2.RoutingApi) diag.Diagnostics {
	if d.HasChange("wrapup_codes") {
		if codesConfig := d.Get("wrapup_codes"); codesConfig != nil {
			// Get existing codes
			codes, err := getRoutingQueueWrapupCodes(d.Id(), routingAPI)
			if err != nil {
				return err
			}

			var existingCodes []string
			if codes != nil {
				for _, code := range codes {
					existingCodes = append(existingCodes, *code.Id)
				}
			}
			configCodes := *lists.SetToStringList(codesConfig.(*schema.Set))

			codesToRemove := lists.SliceDifference(existingCodes, configCodes)
			if len(codesToRemove) > 0 {
				for _, codeId := range codesToRemove {
					resp, err := routingAPI.DeleteRoutingQueueWrapupcode(d.Id(), codeId)
					if err != nil {
						if IsStatus404(resp) {
							// Ignore missing queue or wrapup code
							continue
						}
						return diag.Errorf("Failed to remove wrapup code from queue %s: %s", d.Id(), err)
					}
				}
			}

			codesToAdd := lists.SliceDifference(configCodes, existingCodes)
			if len(codesToAdd) > 0 {
				err := addWrapupCodesInChunks(d.Id(), codesToAdd, routingAPI)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func addWrapupCodesInChunks(queueID string, codesToAdd []string, api *platformclientv2.RoutingApi) diag.Diagnostics {
	// API restricts wraup code adds to 100 per call
	const maxBatchSize = 100
	for i := 0; i < len(codesToAdd); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(codesToAdd) {
			end = len(codesToAdd)
		}
		var updateChunk []platformclientv2.Wrapupcodereference
		for j := i; j < end; j++ {
			updateChunk = append(updateChunk, platformclientv2.Wrapupcodereference{Id: &codesToAdd[j]})
		}

		if len(updateChunk) > 0 {
			_, _, err := api.PostRoutingQueueWrapupcodes(queueID, updateChunk)
			if err != nil {
				return diag.Errorf("Failed to update wrapup codes in queue %s: %s", queueID, err)
			}
		}
	}
	return nil
}

func getRoutingQueueWrapupCodes(queueID string, api *platformclientv2.RoutingApi) ([]platformclientv2.Wrapupcode, diag.Diagnostics) {
	const maxPageSize = 100

	var codes []platformclientv2.Wrapupcode
	for pageNum := 1; ; pageNum++ {
		codeResult, _, err := api.GetRoutingQueueWrapupcodes(queueID, maxPageSize, pageNum)
		if err != nil {
			return nil, diag.Errorf("Failed to query wrapup codes for queue %s: %s", queueID, err)
		}
		if codeResult == nil || codeResult.Entities == nil || len(*codeResult.Entities) == 0 {
			return codes, nil
		}
		for _, code := range *codeResult.Entities {
			codes = append(codes, code)
		}
	}
}

func updateQueueMembers(d *schema.ResourceData, routingAPI *platformclientv2.RoutingApi) diag.Diagnostics {
	if d.HasChange("members") {
		if members := d.Get("members"); members != nil {
			log.Printf("Updating members for Queue %s", d.Get("name"))
			newUserRingNums := make(map[string]int)
			memberList := members.(*schema.Set).List()
			newUserIds := make([]string, len(memberList))
			for i, member := range memberList {
				memberMap := member.(map[string]interface{})
				newUserIds[i] = memberMap["user_id"].(string)
				newUserRingNums[newUserIds[i]] = memberMap["ring_num"].(int)
			}

			oldSdkUsers, err := getRoutingQueueMembers(d.Id(), "", routingAPI)
			if err != nil {
				return err
			}

			oldUserIds := make([]string, len(oldSdkUsers))
			oldUserRingNums := make(map[string]int)
			for i, user := range oldSdkUsers {
				oldUserIds[i] = *user.Id
				oldUserRingNums[oldUserIds[i]] = *user.RingNumber
			}

			if len(oldUserIds) > 0 {
				usersToRemove := lists.SliceDifference(oldUserIds, newUserIds)
				err := updateMembersInChunks(d.Id(), usersToRemove, true, routingAPI)
				if err != nil {
					return err
				}
			}

			if len(newUserIds) > 0 {
				usersToAdd := lists.SliceDifference(newUserIds, oldUserIds)
				err := updateMembersInChunks(d.Id(), usersToAdd, false, routingAPI)
				if err != nil {
					return err
				}
			}

			// Check for ring numbers to update
			for userID, newNum := range newUserRingNums {
				if oldNum, found := oldUserRingNums[userID]; found {
					if newNum != oldNum {
						// Number changed. Update ring number
						err := updateQueueUserRingNum(d.Id(), userID, newNum, routingAPI)
						if err != nil {
							return err
						}
					}
				} else if newNum != 1 {
					// New queue member. Update ring num if not set to the default of 1
					err := updateQueueUserRingNum(d.Id(), userID, newNum, routingAPI)
					if err != nil {
						return err
					}
				}
			}
			log.Printf("Members updated for Queue %s", d.Get("name"))
		}
	}
	return nil
}

func updateMembersInChunks(queueID string, membersToUpdate []string, remove bool, api *platformclientv2.RoutingApi) diag.Diagnostics {

	// API restricts member adds/removes to 100 per call
	// Generic call to prepare chunks for the Update. Takes in three args
	// 1. MemberstoUpdate 2. The Entity prepare func for the update 3. Chunk Size
	if len(membersToUpdate) > 0 {
		chunks := chunksProcess.ChunkItems(membersToUpdate, platformWritableEntityFunc, 100)
		// Closure to process the chunks
		chunkProcessor := func(chunk []platformclientv2.Writableentity) diag.Diagnostics {
			_, err := api.PostRoutingQueueMembers(queueID, chunk, remove)
			if err != nil {
				return diag.Errorf("Failed to update members in queue %s: %s", queueID, err)
			}
			return nil
		}
		// Genric Function call which takes in the chunks and the processing function
		return chunksProcess.ProcessChunks(chunks, chunkProcessor)
	}
	return nil

}

func platformWritableEntityFunc(val string) platformclientv2.Writableentity {
	return platformclientv2.Writableentity{Id: &val}
}

func updateQueueUserRingNum(queueID string, userID string, ringNum int, api *platformclientv2.RoutingApi) diag.Diagnostics {
	_, err := api.PatchRoutingQueueMember(queueID, userID, platformclientv2.Queuemember{
		Id:         &userID,
		RingNumber: &ringNum,
	})
	if err != nil {
		return diag.Errorf("Failed to update ring number for queue %s user %s: %s", queueID, userID, err)
	}
	return nil
}

func getRoutingQueueMembers(queueID string, memberBy string, api *platformclientv2.RoutingApi) ([]platformclientv2.Queuemember, diag.Diagnostics) {
	const maxPageSize = 100

	var members []platformclientv2.Queuemember
	for pageNum := 1; ; pageNum++ {
		users, _, err := sdkGetRoutingQueueMembers(queueID, memberBy, pageNum, maxPageSize, api)
		if err != nil {
			return nil, diag.Errorf("Failed to query users for queue %s: %s", queueID, err)
		}
		if users == nil || users.Entities == nil || len(*users.Entities) == 0 {
			return members, nil
		}
		for _, user := range *users.Entities {
			members = append(members, user)
		}
	}
}

func sdkGetRoutingQueueMembers(queueID, memberBy string, pageNumber, pageSize int, api *platformclientv2.RoutingApi) (*platformclientv2.Queuememberentitylisting, *platformclientv2.APIResponse, error) {
	// SDK does not support nil values for boolean query params yet, so we must manually construct this HTTP request for now
	apiClient := &api.Configuration.APIClient

	// create path and map variables
	path := api.Configuration.BasePath + "/api/v2/routing/queues/{queueId}/members"
	path = strings.Replace(path, "{queueId}", fmt.Sprintf("%v", queueID), -1)

	headerParams := make(map[string]string)
	queryParams := make(map[string]string)
	formParams := url.Values{}
	var postBody interface{}
	var postFileName string
	var fileBytes []byte

	// oauth required
	if api.Configuration.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + api.Configuration.AccessToken
	}
	// add default headers if any
	for key := range api.Configuration.DefaultHeader {
		headerParams[key] = api.Configuration.DefaultHeader[key]
	}

	queryParams["pageSize"] = apiClient.ParameterToString(pageSize, "")
	queryParams["pageNumber"] = apiClient.ParameterToString(pageNumber, "")
	if memberBy != "" {
		queryParams["memberBy"] = memberBy
	}

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload *platformclientv2.Queuememberentitylisting
	response, err := apiClient.CallAPI(path, http.MethodGet, postBody, headerParams, queryParams, formParams, postFileName, fileBytes)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = fmt.Errorf(response.ErrorMessage)
	} else {
		err = json.Unmarshal([]byte(response.RawBody), &successPayload)
	}
	return successPayload, response, err
}

func flattenQueueMembers(queueID string, memberBy string, api *platformclientv2.RoutingApi) (*schema.Set, diag.Diagnostics) {
	members, err := getRoutingQueueMembers(queueID, memberBy, api)
	if err != nil {
		return nil, err
	}

	memberSet := schema.NewSet(schema.HashResource(queueMemberResource), []interface{}{})
	for _, member := range members {
		memberMap := make(map[string]interface{})
		memberMap["user_id"] = *member.Id
		memberMap["ring_num"] = *member.RingNumber
		memberSet.Add(memberMap)
	}

	return memberSet, nil
}

func flattenQueueWrapupCodes(queueID string, api *platformclientv2.RoutingApi) (*schema.Set, diag.Diagnostics) {
	const maxPageSize = 100
	var codeIds []string
	for pageNum := 1; ; pageNum++ {
		codes, _, err := api.GetRoutingQueueWrapupcodes(queueID, maxPageSize, pageNum)
		if err != nil {
			return nil, diag.Errorf("Failed to query wrapup codes for queue %s: %s", queueID, err)
		}
		if codes == nil || codes.Entities == nil || len(*codes.Entities) == 0 {
			break
		}
		for _, code := range *codes.Entities {
			codeIds = append(codeIds, *code.Id)
		}
	}

	if codeIds != nil {
		return lists.StringListToSet(codeIds), nil
	}
	return nil, nil
}

func flattenQueueMemberGroupsList(queue *platformclientv2.Queue, groupType *string) *schema.Set {
	var groupIds []string

	if queue == nil || queue.MemberGroups == nil {
		return nil
	}

	for _, memberGroup := range *queue.MemberGroups {
		if strings.Compare(*memberGroup.VarType, *groupType) == 0 {
			groupIds = append(groupIds, *memberGroup.Id)
		}
	}

	if groupIds != nil {
		return lists.StringListToSet(groupIds)
	}

	return nil
}

func GenerateRoutingQueueResource(
	resourceID string,
	name string,
	desc string,
	acwWrapupPrompt string,
	acwTimeout string,
	skillEvalMethod string,
	autoAnswerOnly string,
	callingPartyName string,
	callingPartyNumber string,
	enableTranscription string,
	suppressInQueueCallRecording string,
	enableManualAssignment string,
	nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_queue" "%s" {
		name = "%s"
		description = "%s"
		acw_wrapup_prompt = %s
		acw_timeout_ms = %s
		skill_evaluation_method = %s
		auto_answer_only = %s
		calling_party_name = %s
		calling_party_number = %s
		enable_transcription = %s
        suppress_in_queue_call_recording = %s
  		enable_manual_assignment = %s
		%s
	}
	`, resourceID,
		name,
		desc,
		acwWrapupPrompt,
		acwTimeout,
		skillEvalMethod,
		autoAnswerOnly,
		callingPartyName,
		callingPartyNumber,
		enableTranscription,
		suppressInQueueCallRecording,
		enableManualAssignment,
		strings.Join(nestedBlocks, "\n"))
}

func GenerateRoutingQueueResourceBasic(resourceID string, name string, nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_queue" "%s" {
		name = "%s"
		%s
	}
	`, resourceID, name, strings.Join(nestedBlocks, "\n"))
}

// Used when testing skills group dependencies.
func GenerateRoutingQueueResourceBasicWithDepends(resourceID string, dependsOn string, name string, nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_queue" "%s" {
		depends_on = [%s]
		name = "%s"
		%s
	}
	`, resourceID, dependsOn, name, strings.Join(nestedBlocks, "\n"))
}

func GenerateMediaSettings(attrName string, alertingTimeout string, enableAutoAnswer string, slPercent string, slDurationMs string) string {
	return fmt.Sprintf(`%s {
		alerting_timeout_sec = %s
		enable_auto_answer = %s
		service_level_percentage = %s
		service_level_duration_ms = %s
	}
	`, attrName, alertingTimeout, enableAutoAnswer, slPercent, slDurationMs)
}

func GenerateRoutingRules(operator string, threshold string, waitSeconds string) string {
	return fmt.Sprintf(`routing_rules {
		operator = "%s"
		threshold = %s
		wait_seconds = %s
	}
	`, operator, threshold, waitSeconds)
}

func GenerateDefaultScriptIDs(chat, email string) string {
	return fmt.Sprintf(`default_script_ids = {
		CHAT  = "%s"
		EMAIL = "%s"
	}`, chat, email)
}

func GenerateBullseyeSettings(expTimeout string, skillsToRemove ...string) string {
	return fmt.Sprintf(`bullseye_rings {
		expansion_timeout_seconds = %s
		skills_to_remove = [%s]
	}
	`, expTimeout, strings.Join(skillsToRemove, ", "))
}

func GenerateConditionalGroupRoutingRules(queueId, operator, metric, conditionValue, waitSeconds string, nestedBlocks ...string) string {
	return fmt.Sprintf(`conditional_group_routing_rules {
		queue_id        = %s
		operator        = "%s"
		metric          = "%s"
		condition_value = %s
		wait_seconds    = %s
		%s
	}
	`, queueId, operator, metric, conditionValue, waitSeconds, strings.Join(nestedBlocks, "\n"))
}

func GenerateConditionalGroupRoutingRuleGroup(groupId, groupType string) string {
	return fmt.Sprintf(`groups {
		member_group_id   = %s
		member_group_type = "%s"
	}
	`, groupId, groupType)
}

func GenerateBullseyeSettingsWithMemberGroup(expTimeout, memberGroupId, memberGroupType string, skillsToRemove ...string) string {
	return fmt.Sprintf(`bullseye_rings {
		expansion_timeout_seconds = %s
		skills_to_remove = [%s]
		member_groups {
			member_group_id = %s
			member_group_type = "%s"
		}
	}
	`, expTimeout, strings.Join(skillsToRemove, ", "), memberGroupId, memberGroupType)
}

func GenerateMemberBlock(userID, ringNum string) string {
	return fmt.Sprintf(`members {
		user_id = %s
		ring_num = %s
	}
	`, userID, ringNum)
}

func GenerateQueueWrapupCodes(wrapupCodes ...string) string {
	return fmt.Sprintf(`
		wrapup_codes = [%s]
	`, strings.Join(wrapupCodes, ", "))
}
