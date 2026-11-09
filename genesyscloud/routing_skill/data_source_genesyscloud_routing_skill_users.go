package routing_skill

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRoutingSkillUsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSkillProxy(sdkConfig)

	skillId := d.Get("skill_id").(string)
	skillName := d.Get("skill_name").(string)

	if skillId == "" {
		var diagErr diag.Diagnostics
		skillId, diagErr = getRoutingSkillIdByNameFromCache(ctx, sdkConfig, skillName)
		if diagErr != nil {
			return diagErr
		}
	}

	log.Printf("Reading users assigned to routing skill %s", skillId)
	users, resp, err := proxy.getRoutingSkillUsers(ctx, skillId)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return util.BuildAPIDiagnosticError(UsersResourceType, fmt.Sprintf("routing skill %s not found", skillId), resp)
		}
		return util.BuildAPIDiagnosticError(UsersResourceType, fmt.Sprintf("failed to read users assigned to routing skill %s | error: %s", skillId, err), resp)
	}

	userIds := make([]interface{}, 0, len(*users))
	for _, user := range *users {
		if user.Id != nil {
			userIds = append(userIds, *user.Id)
		}
	}

	d.SetId(skillId)
	_ = d.Set("skill_id", skillId)
	if skillName != "" {
		_ = d.Set("skill_name", skillName)
	}
	_ = d.Set("user_ids", schema.NewSet(schema.HashString, userIds))

	log.Printf("Read %d users assigned to routing skill %s", len(userIds), skillId)
	return nil
}
