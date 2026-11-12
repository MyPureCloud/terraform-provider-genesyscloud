package routing_skill

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const UsersResourceType = "genesyscloud_routing_skill_users"

func DataSourceRoutingSkillUsers() *schema.Resource {
	return &schema.Resource{
		Description:        "Data source for listing the users assigned to a Genesys Cloud routing skill. Select the skill by ID or by name.",
		ReadWithoutTimeout: provider.ReadWithPooledClient(dataSourceRoutingSkillUsersRead),
		Schema: map[string]*schema.Schema{
			"skill_id": {
				Description:  "ID of the routing skill. At least one of `skill_id` or `skill_name` must be set. If both are set, `skill_id` is used.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"skill_id", "skill_name"},
				ValidateFunc: validation.IsUUID,
			},
			"skill_name": {
				Description:  "Name of the routing skill. At least one of `skill_id` or `skill_name` must be set. Ignored if `skill_id` is set.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"skill_id", "skill_name"},
			},
			"user_ids": {
				Description: "IDs of the users assigned to the routing skill.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}
