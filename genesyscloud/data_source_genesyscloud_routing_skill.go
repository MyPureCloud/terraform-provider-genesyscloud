package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v72/platformclientv2"
)

func dataSourceRoutingSkill() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Routing Skills. Select a skill by name.",
		ReadContext: readWithPooledClient(dataSourceRoutingSkillRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Skill name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceRoutingSkillRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Find first non-deleted skill by name. Retry in case new skill is not yet indexed by search
	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			skills, _, getErr := routingAPI.GetRoutingSkills(pageSize, pageNum, name, nil)
			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("error requesting skill %s: %s", name, getErr))
			}

			if skills.Entities == nil || len(*skills.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("no routing skills found with name %s", name))
			}

			for _, skill := range *skills.Entities {
				if skill.Name != nil && *skill.Name == name &&
					skill.State != nil && *skill.State != "deleted" {
					d.SetId(*skill.Id)
					return nil
				}
			}
		}
	})
}
