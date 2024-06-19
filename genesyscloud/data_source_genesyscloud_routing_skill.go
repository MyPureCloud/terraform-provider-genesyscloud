package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"
)

// The context is now added without Timeout ,
// since the warming up of cache will take place for the first Datasource registered during a Terraform Apply.
func dataSourceRoutingSkill() *schema.Resource {
	return &schema.Resource{
		Description:        "Data source for Genesys Cloud Routing Skills. Select a skill by name.",
		ReadWithoutTimeout: provider.ReadWithPooledClient(dataSourceRoutingSkillRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Skill name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

var (
	dataSourceRoutingSkillCache *rc.DataSourceCache
)

func dataSourceRoutingSkillRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig

	key := d.Get("name").(string)

	if dataSourceRoutingSkillCache == nil {
		dataSourceRoutingSkillCache = rc.NewDataSourceCache(sdkConfig, hydrateRoutingSkillCacheFn, getSkillByNameFn)
	}

	queueId, err := rc.RetrieveId(dataSourceRoutingSkillCache, "genesyscloud_routing_skill", key, ctx)

	if err != nil {
		return err
	}

	d.SetId(queueId)
	return nil
}

func hydrateRoutingSkillCacheFn(c *rc.DataSourceCache) error {
	log.Printf("hydrating cache for data source genesyscloud_routing_skill")

	routingApi := platformclientv2.NewRoutingApiWithConfig(c.ClientConfig)
	const pageSize = 100
	skills, _, getErr := routingApi.GetRoutingSkills(pageSize, 1, "", nil)

	if getErr != nil {
		return fmt.Errorf("failed to get page of skills: %v", getErr)
	}

	if skills.Entities == nil || len(*skills.Entities) == 0 {
		return nil
	}

	for _, skill := range *skills.Entities {
		c.Cache[*skill.Name] = *skill.Id
	}

	for pageNum := 2; pageNum <= *skills.PageCount; pageNum++ {

		log.Printf("calling cache for data source genesyscloud_routing_skill")

		skills, _, getErr := routingApi.GetRoutingSkills(pageSize, pageNum, "", nil)
		log.Printf("calling cache for data source genesyscloud_routing_skill %v", pageNum)
		if getErr != nil {
			return fmt.Errorf("failed to get page of skills: %v", getErr)
		}

		if skills.Entities == nil || len(*skills.Entities) == 0 {
			break
		}

		// Add ids to cache
		for _, skill := range *skills.Entities {
			c.Cache[*skill.Name] = *skill.Id
		}
	}

	log.Printf("cache hydration completed for data source genesyscloud_routing_skill")

	return nil
}

func getSkillByNameFn(c *rc.DataSourceCache, name string, ctx context.Context) (string, diag.Diagnostics) {
	const pageSize = 100
	skillId := ""
	routingAPI := platformclientv2.NewRoutingApiWithConfig(c.ClientConfig)

	// Find first non-deleted skill by name. Retry in case new skill is not yet indexed by search
	diag := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			skills, resp, getErr := routingAPI.GetRoutingSkills(pageSize, pageNum, name, nil)
			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_skill", fmt.Sprintf("error requesting skill %s | error: %s", name, getErr), resp))
			}

			if skills.Entities == nil || len(*skills.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_skill", fmt.Sprintf("no routing skills found with name %s", name), resp))
			}

			for _, skill := range *skills.Entities {
				if skill.Name != nil && *skill.Name == name &&
					skill.State != nil && *skill.State != "deleted" {
					skillId = *skill.Id
					return nil
				}
			}
		}
	})
	return skillId, diag
}
