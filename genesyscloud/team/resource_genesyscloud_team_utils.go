package team

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/platform-client-sdk-go/v130/platformclientv2"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

// getTeamFromResourceData maps data from schema ResourceData object to a platformclientv2.Team
func getTeamFromResourceData(d *schema.ResourceData) platformclientv2.Team {
	name := d.Get("name").(string)
	division := d.Get("division_id").(string)
	return platformclientv2.Team{
		Name:        &name,
		Division:    &platformclientv2.Writabledivision{Id: &division},
		Description: platformclientv2.String(d.Get("description").(string)),
	}
}

func getTeamMemberIds(ctx context.Context, d *schema.ResourceData, sdkConfig *platformclientv2.Configuration) ([]string, diag.Diagnostics) {
	gp := getTeamProxy(sdkConfig)
	members, resp, err := gp.getMembersById(ctx, d.Id())
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Unable to retrieve members for group %s. %s", d.Id(), err), resp)
	}

	memberIds := make([]string, len(*members))
	for i, member := range *members {
		memberIds[i] = *member.Id
	}

	return memberIds, nil
}

func addGroupMembers(ctx context.Context, d *schema.ResourceData, membersToAdd []string, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	proxy := getTeamProxy(sdkConfig)

	teamMembers := &platformclientv2.Teammembers{
		MemberIds: &membersToAdd,
	}

	_, resp, err := proxy.createMembers(ctx, d.Id(), *teamMembers)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to add team members %s: %s", d.Id(), err), resp)
	}

	return nil
}

func readTeamMembers(ctx context.Context, teamId string, sdkConfig *platformclientv2.Configuration) (*schema.Set, diag.Diagnostics) {
	proxy := getTeamProxy(sdkConfig)
	members, resp, err := proxy.getMembersById(ctx, teamId)

	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to read members for team %s: %s", teamId, err), resp)
	}

	if members == nil || len(*members) == 0 {
		return nil, nil
	}

	interfaceList := make([]interface{}, len(*members))
	for i, member := range *members {
		interfaceList[i] = *member.Id
	}
	return schema.NewSet(schema.HashString, interfaceList), nil
}
