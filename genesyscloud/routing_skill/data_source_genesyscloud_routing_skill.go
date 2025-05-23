package routing_skill

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var dataSourceRoutingSkillCache *rc.DataSourceCache

func dataSourceRoutingSkillRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	key := d.Get("name").(string)

	if dataSourceRoutingSkillCache == nil {
		dataSourceRoutingSkillCache = rc.NewDataSourceCache(sdkConfig, hydrateRoutingSkillCacheFn, getSkillByNameFn)
	}

	queueId, err := rc.RetrieveId(dataSourceRoutingSkillCache, ResourceType, key, ctx)
	if err != nil {
		return err
	}

	d.SetId(queueId)
	return nil
}

func hydrateRoutingSkillCacheFn(c *rc.DataSourceCache, ctx context.Context) error {
	log.Printf("hydrating cache for data source %s", ResourceType)
	proxy := getRoutingSkillProxy(c.ClientConfig)

	skills, resp, getErr := proxy.getAllRoutingSkills(ctx, "")
	if getErr != nil {
		return fmt.Errorf("failed to get page of skills: %v %v", getErr, resp)
	}

	if skills == nil || len(*skills) == 0 {
		log.Printf("no skills returned. Cache will remain empty")
		return nil
	}

	for _, skill := range *skills {
		c.Cache[*skill.Name] = *skill.Id
	}

	log.Printf("cache hydration completed for data source %s", ResourceType)

	return nil
}

func getSkillByNameFn(c *rc.DataSourceCache, name string, ctx context.Context) (string, diag.Diagnostics) {
	skillId := ""
	proxy := getRoutingSkillProxy(c.ClientConfig)

	// Find first non-deleted skill by name. Retry in case new skill is not yet indexed by search
	diag := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		skill, resp, retryable, getErr := proxy.getRoutingSkillIdByName(ctx, name)
		if getErr != nil {
			diagnosticError := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error requesting skill %s | error: %s", name, getErr), resp)
			if !retryable {
				return retry.NonRetryableError(diagnosticError)
			}
			return retry.RetryableError(diagnosticError)
		}

		skillId = skill
		return nil
	})
	return skillId, diag
}
