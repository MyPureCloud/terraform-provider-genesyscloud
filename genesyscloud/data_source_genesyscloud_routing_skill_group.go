package genesyscloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func dataSourceRoutingSkillGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Routing Skills Groups. Select a skill group by name.",
		ReadContext: ReadWithPooledClient(dataSourceRoutingSkillGroupRead),
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

	sdkConfig := m.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Find first non-deleted skill by name. Retry in case new skill is not yet indexed by search
	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100

			apiClient := &routingAPI.Configuration.APIClient
			path := routingAPI.Configuration.BasePath + "/api/v2/routing/skillgroups"

			headerParams := buildHeaderParams(routingAPI)
			response, err := apiClient.CallAPI(path, "GET", nil, headerParams, nil, nil, "", nil)

			if err != nil {
				return retry.RetryableError(fmt.Errorf("error encountered while trying to retrieve routing skills group found with name %s, %w", name, err))
			}

			if err == nil && response.Error != nil {
				return retry.RetryableError(fmt.Errorf("error encountered while trying to retrieve routing skills group found with name %s %w", name, err))
			}
			if err == nil && response.StatusCode == http.StatusNotFound {
				return retry.RetryableError(fmt.Errorf("routing skills group not found with name %s", name))
			}

			allSkillGroups := &AllSkillGroups{}

			err = json.Unmarshal(response.RawBody, &allSkillGroups)
			if err != nil {
				return retry.RetryableError(fmt.Errorf("error encountered while trying to retrieve routing skills group found with name %s %w", name, err))
			}

			if allSkillGroups.Entities == nil || len(allSkillGroups.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("no routing skills groups found with name %s", name))
			}

			for _, skillGroup := range allSkillGroups.Entities {
				if skillGroup.Name == name {
					d.SetId(skillGroup.ID)
					return nil
				}
			}
		}

	})
}
