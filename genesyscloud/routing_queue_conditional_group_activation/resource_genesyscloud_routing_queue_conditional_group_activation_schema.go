package routing_queue_conditional_group_activation

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const ResourceType = "genesyscloud_routing_queue_conditional_group_activation"

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceRoutingQueueConditionalGroupActivation())
	regInstance.RegisterExporter(ResourceType, RoutingQueueConditionalGroupActivationExporter())
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

	cgaSimpleMetric = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"metric": {
				Description:  "The queue metric being evaluated. Valid values: EstimatedWaitTime, ServiceLevel, IdleAgentCount.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"EstimatedWaitTime", "ServiceLevel", "IdleAgentCount"}, false),
			},
			"queue_id": {
				Description: "The queue being evaluated for this rule. If null, the current queue will be used.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}

	cgaCondition = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"simple_metric": {
				Description: "Instructs this condition to evaluate a simple queue-level metric.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Elem:        cgaSimpleMetric,
			},
			"operator": {
				Description:  "The operator used to compare the actual value against the threshold value. Valid values: GreaterThan, GreaterThanOrEqualTo, LessThan, LessThanOrEqualTo, EqualTo, NotEqualTo.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"GreaterThan", "LessThan", "GreaterThanOrEqualTo", "LessThanOrEqualTo", "EqualTo", "NotEqualTo"}, false),
			},
			"value": {
				Description: "The threshold value, beyond which a rule evaluates as true.",
				Type:        schema.TypeFloat,
				Required:    true,
			},
		},
	}
)

func ResourceRoutingQueueConditionalGroupActivation() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud routing queue conditional group activation rules",

		CreateContext: provider.CreateWithPooledClient(createRoutingQueueConditionalGroupActivation),
		ReadContext:   provider.ReadWithPooledClient(readRoutingQueueConditionalGroupActivation),
		UpdateContext: provider.UpdateWithPooledClient(updateRoutingQueueConditionalGroupActivation),
		DeleteContext: provider.DeleteWithPooledClient(deleteRoutingQueueConditionalGroupActivation),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"queue_id": {
				Description: "Id of the routing queue to which the conditional group activation rules belong.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"pilot_rule": {
				Description: "The pilot rule for this queue, which executes periodically to determine queue health.",
				Type:        schema.TypeList,
				Optional:    true,
				MinItems:    1,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"condition_expression": {
							Description: "A string expression that defines the relationships of conditions in this rule.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"conditions": {
							Description: "The list of conditions used in this rule.",
							Type:        schema.TypeList,
							Required:    true,
							MinItems:    1,
							MaxItems:    10,
							Elem:        cgaCondition,
						},
					},
				},
			},
			"rules": {
				Description: "The set of rules to be periodically executed on the queue (if the pilot rule evaluates as true or there is no pilot rule).",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				MaxItems:    5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"condition_expression": {
							Description: "A string expression that defines the relationships of conditions in this rule.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"conditions": {
							Description: "The list of conditions used in this rule.",
							Type:        schema.TypeList,
							Required:    true,
							MinItems:    1,
							MaxItems:    10,
							Elem:        cgaCondition,
						},
						"groups": {
							Description: "The group(s) to activate if the rule evaluates as true.",
							Type:        schema.TypeList,
							Required:    true,
							MinItems:    1,
							MaxItems:    5,
							Elem:        memberGroupResource,
						},
					},
				},
			},
		},
	}
}

func RoutingQueueConditionalGroupActivationExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRoutingQueueConditionalGroupActivation),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"queue_id": {RefType: "genesyscloud_routing_queue"},
			"pilot_rule.conditions.simple_metric.queue_id": {RefType: "genesyscloud_routing_queue"},
			"rules.conditions.simple_metric.queue_id":      {RefType: "genesyscloud_routing_queue"},
		},
		CustomAttributeResolver: map[string]*resourceExporter.RefAttrCustomResolver{
			"rules.groups.member_group_id": {ResolverFunc: resourceExporter.MemberGroupsResolver},
		},
	}
}
