package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"
)

func dataSourceRoutingSkillGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Routing Skills Groups. Select a skill group by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceRoutingSkillGroupRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Skill group name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceRoutingSkillGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		resp *platformclientv2.APIResponse
		err  error
	)

	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	name := d.Get("name").(string)

	retryErr := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		var skillGroups []platformclientv2.Skillgroupdefinition
		skillGroups, resp, err = getAllSkillGroupsByName(routingAPI, name)
		if err != nil {
			return retry.NonRetryableError(err)
		}

		for _, skillGroup := range skillGroups {
			if name == *skillGroup.Name {
				d.SetId(*skillGroup.Id)
				return nil
			}
		}

		return retry.RetryableError(fmt.Errorf("failed to find skill group with name '%s'", name))
	})

	if retryErr != nil {
		msg := fmt.Sprintf("failed to read skill group by name %s: %v", name, retryErr)
		if resp != nil {
			return util.BuildAPIDiagnosticError(getSkillGroupResourceName(), msg, resp)
		}
		return util.BuildDiagnosticError(getSkillGroupResourceName(), msg, fmt.Errorf("%v", retryErr))
	}

	return nil
}

func getAllSkillGroupsByName(api *platformclientv2.RoutingApi, name string) ([]platformclientv2.Skillgroupdefinition, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	var (
		after          string
		allSkillGroups []platformclientv2.Skillgroupdefinition
	)

	for {
		skillGroups, resp, err := api.GetRoutingSkillgroups(pageSize, name, after, "")
		if err != nil {
			return nil, resp, err
		}

		if skillGroups.Entities == nil || len(*skillGroups.Entities) == 0 {
			break
		}

		allSkillGroups = append(allSkillGroups, *skillGroups.Entities...)

		if skillGroups.NextUri == nil || *skillGroups.NextUri == "" {
			break
		}

		after, err = util.GetQueryParamValueFromUri(*skillGroups.NextUri, "after")
		if err != nil {
			return nil, nil, fmt.Errorf("error parsing NextUri '%s' while reading skill groups: %v", *skillGroups.NextUri, err)
		}
		if after == "" {
			break
		}
	}

	return allSkillGroups, nil, nil
}
