package team

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/chunks"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
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
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Unable to retrieve members for group %s. %s", d.Id(), err), resp)
	}

	memberIds := make([]string, len(*members))
	for i, member := range *members {
		memberIds[i] = *member.Id
	}

	return memberIds, nil
}

func updateTeamMembers(ctx context.Context, d *schema.ResourceData, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	proxy := getTeamProxy(sdkConfig)
	if d.HasChange("member_ids") {
		if membersConfig := d.Get("member_ids"); membersConfig != nil {
			configMemberIds := *lists.SetToStringList(membersConfig.(*schema.Set))
			existingMemberIds, err := getTeamMemberIds(ctx, d, sdkConfig)
			if err != nil {
				return err
			}

			maxMembersPerRequest := 25
			membersToRemoveList := lists.SliceDifference(existingMemberIds, configMemberIds)
			chunkedMemberIdsDelete := chunks.ChunkBy(membersToRemoveList, maxMembersPerRequest)

			chunkProcessor := func(membersToRemove []string) diag.Diagnostics {
				if len(membersToRemove) > 0 {
					if diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
						resp, err := proxy.deleteMembers(ctx, d.Id(), strings.Join(membersToRemove, ","))
						if err != nil {
							return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to remove members from team %s: %s", d.Id(), err), resp)
						}
						return resp, nil
					}); diagErr != nil {
						return diagErr
					}
				}
				return nil
			}

			if err := chunks.ProcessChunks(chunkedMemberIdsDelete, chunkProcessor); err != nil {
				return err
			}

			membersToAdd := lists.SliceDifference(configMemberIds, existingMemberIds)
			if len(membersToAdd) < 1 {
				return nil
			}

			chunkedMemberIds := lists.ChunkStringSlice(membersToAdd, maxMembersPerRequest)
			for _, chunk := range chunkedMemberIds {
				if err := addGroupMembers(ctx, d, chunk, sdkConfig); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func addGroupMembers(ctx context.Context, d *schema.ResourceData, membersToAdd []string, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	proxy := getTeamProxy(sdkConfig)

	teamMembers := &platformclientv2.Teammembers{
		MemberIds: &membersToAdd,
	}

	teamListingResponse, resp, err := proxy.createMembers(ctx, d.Id(), *teamMembers)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to add team members %s: %s", d.Id(), err), resp)
	}
	if len(*teamListingResponse.Failures) > 0 {
		failureReasons := make([]string, len(*teamListingResponse.Failures))
		for i, failure := range *teamListingResponse.Failures {
			failureReasons[i] = fmt.Sprintf("Member %s: %s", *failure.Id, *failure.Reason)
		}
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to add team members for team %s: %v", d.Id(), failureReasons), fmt.Errorf("%v", failureReasons))
	}

	return nil
}

func readTeamMembers(ctx context.Context, teamId string, sdkConfig *platformclientv2.Configuration) (*schema.Set, diag.Diagnostics) {
	proxy := getTeamProxy(sdkConfig)
	members, resp, err := proxy.getMembersById(ctx, teamId)

	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read members for team %s: %s", teamId, err), resp)
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

func GenerateTeamResource(
	teamResource string,
	name string,
	divisionId string,
	description string,
	memberIds ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_team" "%s" {
		name = "%s"
		division_id = %s
		description = "%s"
		%s
	}
	`, teamResource, name, divisionId, description, strings.Join(memberIds, "\n"))
}
