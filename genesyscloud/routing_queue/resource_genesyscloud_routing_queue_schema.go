package routing_queue

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const resourceName = "genesyscloud_routing_queue"

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceRoutingQueue())
	regInstance.RegisterDataSource(resourceName, DataSourceRoutingQueue())
	regInstance.RegisterExporter(resourceName, RoutingQueueExporter())
}

var (
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

	agentOwnedRoutingResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"enable_agent_owned_callbacks": {
				Description: "Enable Agent Owned Callbacks",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"max_owned_callback_hours": {
				Description: "Auto End Delay Seconds Must be >= 7",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"max_owned_callback_delay_hours": {
				Description: "Max Owned Call Back Delay Hours >= 7",
				Type:        schema.TypeInt,
				Required:    true,
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
			"auto_end_delay_seconds": {
				Description: "Auto End Delay Seconds.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"auto_dial_delay_seconds": {
				Description: "Auto Dial Delay Seconds.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"enable_auto_answer": {
				Description: "Auto-Answer for digital channels(Email, Message)",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"enable_auto_dial_and_end": {
				Description: "Auto Dail and End",
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
				Description: "Direct Routing default backup queue id (if none supplied this queue will be used as backup).",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
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

func ResourceRoutingQueue() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Queue",

		CreateContext: provider.CreateWithPooledClient(createQueue),
		ReadContext:   provider.ReadWithPooledClient(readQueue),
		UpdateContext: provider.UpdateWithPooledClient(updateQueue),
		DeleteContext: provider.DeleteWithPooledClient(deleteQueue),
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
			"agent_owned_routing": {
				Description: "Agent Owned Routing.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        agentOwnedRoutingResource,
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
				Deprecated:  "conditional_group_routing_rules is deprecated in genesyscloud_routing_queue. CGR is now a standalone resource, please set ENABLE_STANDALONE_CGR in your environment variables to enable and use genesyscloud_routing_queue_conditional_group_routing",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"queue_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: `The ID of the queue being evaluated for this rule. For rule 1, this is always be the current queue, so no queue id should be specified for the first rule.`,
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
			"on_hold_prompt_id": {
				Description: "The audio to be played when calls on this queue are on hold. If not configured, the default on-hold music will play.",
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
			"scoring_method": {
				Description:  "The Scoring Method for the queue. Defaults to TimestampAndPriority.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "TimestampAndPriority",
				ValidateFunc: validation.StringInSlice([]string{"TimestampAndPriority", "PriorityOnly"}, false),
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
				Deprecated:  "outbound_email_address is deprecated in genesyscloud_routing_queue. OEA is now a standalone resource, please set ENABLE_STANDALONE_EMAIL_ADDRESS in your environment variables to enable and use genesyscloud_routing_queue_outbound_email_address",
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
				Description: "Users in the queue. If not set, this resource will not manage members. If a user is already assigned to this queue via a group, attempting to assign them using this field will cause an error to be thrown.",
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

func RoutingQueueExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRoutingQueues),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id":                              {RefType: "genesyscloud_auth_division"},
			"queue_flow_id":                            {RefType: "genesyscloud_flow"},
			"email_in_queue_flow_id":                   {RefType: "genesyscloud_flow"},
			"message_in_queue_flow_id":                 {RefType: "genesyscloud_flow"},
			"whisper_prompt_id":                        {RefType: "genesyscloud_architect_user_prompt"},
			"on_hold_prompt_id":                        {RefType: "genesyscloud_architect_user_prompt"},
			"outbound_messaging_sms_address_id":        {},                               // Ref type not yet defined
			"default_script_ids.*":                     {RefType: "genesyscloud_script"}, // Ref type not yet defined
			"outbound_email_address.route_id":          {RefType: "genesyscloud_routing_email_route"},
			"outbound_email_address.domain_id":         {RefType: "genesyscloud_routing_email_domain"},
			"bullseye_rings.skills_to_remove":          {RefType: "genesyscloud_routing_skill"},
			"members.user_id":                          {RefType: "genesyscloud_user"},
			"wrapup_codes":                             {RefType: "genesyscloud_routing_wrapupcode"},
			"skill_groups":                             {RefType: "genesyscloud_routing_skill_group"},
			"teams":                                    {RefType: "genesyscloud_team"},
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

func DataSourceRoutingQueue() *schema.Resource {
	return &schema.Resource{
		Description:        "Data source for Genesys Cloud Routing Queues. Select a queue by name.",
		ReadWithoutTimeout: provider.ReadWithPooledClient(dataSourceRoutingQueueRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Queue name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
