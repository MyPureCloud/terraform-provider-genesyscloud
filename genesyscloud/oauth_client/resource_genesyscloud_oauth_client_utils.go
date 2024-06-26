package oauth_client

import (
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func buildOAuthRedirectURIs(d *schema.ResourceData) *[]string {
	if config, ok := d.GetOk("registered_redirect_uris"); ok {
		return lists.SetToStringList(config.(*schema.Set))
	}
	return nil
}

func buildOAuthScopes(d *schema.ResourceData) *[]string {
	if config, ok := d.GetOk("scopes"); ok {
		return lists.SetToStringList(config.(*schema.Set))
	}
	return nil
}

func buildOAuthRoles(d *schema.ResourceData) (*[]platformclientv2.Roledivision, diag.Diagnostics) {
	if config, ok := d.GetOk("roles"); ok {
		var sdkRoles []platformclientv2.Roledivision
		roleConfig := config.(*schema.Set).List()
		for _, role := range roleConfig {
			roleMap := role.(map[string]interface{})
			roleId := roleMap["role_id"].(string)

			var divisionId string
			if divConfig, ok := roleMap["division_id"]; ok {
				divisionId = divConfig.(string)
			}

			if divisionId == "" {
				// Set to home division if not set
				var diagErr diag.Diagnostics
				divisionId, diagErr = util.GetHomeDivisionID()
				if diagErr != nil {
					return nil, diagErr
				}
			}

			roleDiv := platformclientv2.Roledivision{
				RoleId:     &roleId,
				DivisionId: &divisionId,
			}
			sdkRoles = append(sdkRoles, roleDiv)
		}
		return &sdkRoles, nil
	}
	return nil, nil
}

func flattenOAuthRoles(sdkRoles []platformclientv2.Roledivision) *schema.Set {
	roleSet := schema.NewSet(schema.HashResource(oauthClientRoleDivResource), []interface{}{})
	for _, roleDiv := range sdkRoles {
		role := make(map[string]interface{})
		if roleDiv.RoleId != nil {
			role["role_id"] = *roleDiv.RoleId
		}
		if roleDiv.DivisionId != nil {
			role["division_id"] = *roleDiv.DivisionId
		}
		roleSet.Add(role)
	}
	return roleSet
}
