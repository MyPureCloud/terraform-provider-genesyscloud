package routing_queue_conditional_group_routing

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const resourceName = "genesyscloud_routing_queue_conditional_group_routing"

// SetRegistrar registers all the resources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceRoutingQueueConditionalGroupRouting())
	regInstance.RegisterExporter(resourceName, RoutingQueueConditionalGroupRoutingExporter())
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
)

// ResourceRoutingQueueConditionalGroupRouting registers the genesyscloud_routing_queue_conditional_group_routing resource with Terraform
func ResourceRoutingQueueConditionalGroupRouting() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud routing queue conditional group routing rules",

		CreateContext: provider.CreateWithPooledClient(createRoutingQueueConditionalRoutingGroup),
		ReadContext:   provider.ReadWithPooledClient(readRoutingQueueConditionalRoutingGroup),
		UpdateContext: provider.UpdateWithPooledClient(updateRoutingQueueConditionalRoutingGroup),
		DeleteContext: provider.DeleteWithPooledClient(deleteRoutingQueueConditionalRoutingGroup),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"queue_id": {
				Description: "Id of the routing queue to which the rules belong",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"rules": {
				Description: "The Conditional Group Routing settings for the queue.",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				MaxItems:    5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"evaluated_queue_id": {
							Description: "The queue being evaluated for this rule. For rule 1, this is always the current queue, so should not be specified.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"operator": {
							Description:  "The operator that compares the actual value against the condition value. Valid values: GreaterThan, GreaterThanOrEqualTo, LessThan, LessThanOrEqualTo.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"GreaterThan", "LessThan", "GreaterThanOrEqualTo", "LessThanOrEqualTo"}, false),
						},
						"metric": {
							Description:  "The queue metric being evaluated. Valid values: EstimatedWaitTime, ServiceLevel.",
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "EstimatedWaitTime",
							ValidateFunc: validation.StringInSlice([]string{"EstimatedWaitTime", "ServiceLevel"}, false),
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
		},
	}
}

func RoutingQueueConditionalGroupRoutingExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthRoutingQueueConditionalGroup),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"queue_id":                 {RefType: "genesyscloud_routing_queue"},
			"rules.evaluated_queue_id": {RefType: "genesyscloud_routing_queue"},
		},
		CustomAttributeResolver: map[string]*resourceExporter.RefAttrCustomResolver{
			"rules.groups.member_group_id": {ResolverFunc: resourceExporter.MemberGroupsResolver},
			"rules.condition_value":        {ResolverFunc: resourceExporter.ConditionValueResolver},
		},
	}
}
