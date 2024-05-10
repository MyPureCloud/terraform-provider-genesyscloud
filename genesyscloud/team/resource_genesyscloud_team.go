package team

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v129/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_team.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthTeam retrieves all of the team via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthTeams(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getTeamProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)
	teams, resp, err := proxy.getAllTeam(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get team error: %s", err), resp)
	}
	for _, team := range *teams {
		resources[*team.Id] = &resourceExporter.ResourceMeta{Name: *team.Name}
	}
	return resources, nil
}

// createTeam is used by the team resource to create Genesys cloud team
func createTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)
	team := getTeamFromResourceData(d)
	log.Printf("Creating team %s", *team.Name)
	teamObj, resp, err := proxy.createTeam(ctx, &team)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create team %s error: %s", *team.Name, err), resp)
	}
	d.SetId(*teamObj.Id)
	log.Printf("Created team %s", *teamObj.Id)
	//adding members to the team
	members, ok := d.GetOk("member_ids")
	if ok {
		if memberList := members.([]interface{}); len(memberList) > 0 {
			diagErr := createMembers(ctx, *teamObj.Id, memberList, proxy)
			if diagErr != nil {
				return diagErr
			}
		}
	}
	return readTeam(ctx, d, meta)
}

// readTeam is used by the team resource to read a team from genesys cloud
func readTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTeam(), constants.DefaultConsistencyChecks)

	log.Printf("Reading team %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		team, resp, getErr := proxy.getTeamById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("failed to read team %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("failed to read team %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", team.Name)
		resourcedata.SetNillableReferenceWritableDivision(d, "division_id", team.Division)
		resourcedata.SetNillableValue(d, "description", team.Description)

		// reading members
		members, err := readMembers(ctx, d, proxy)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read members of the team %s : %s", d.Id(), err))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read members of the team %s : %s", d.Id(), err))
		}
		_ = d.Set("member_ids", members)

		log.Printf("Read team %s %s", d.Id(), *team.Name)
		return cc.CheckState(d)
	})
}

// updateTeam is used by the team resource to update a team in Genesys Cloud
func updateTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)
	team := getTeamFromResourceData(d)
	log.Printf("Updating team %s", *team.Name)
	teamObj, resp, err := proxy.updateTeam(ctx, d.Id(), &team)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update team %s error: %s", *team.Name, err), resp)
	}

	members := d.Get("member_ids")
	memberList := members.([]interface{})
	currentMembers, _ := readMembers(ctx, d, proxy)
	if len(memberList) == 0 {
		if len(currentMembers) > 0 {
			log.Printf("removing all members from team %s", d.Id())
			deleteMembers(ctx, d.Id(), currentMembers, proxy)
		}
	}

	if len(memberList) > 0 {
		removeMembers, addMembers := SliceDifferenceMembers(currentMembers, memberList)
		if len(removeMembers) > 0 {
			diagErr := deleteMembers(ctx, d.Id(), removeMembers, proxy)
			if diagErr != nil {
				return diagErr
			}
		}
		if len(addMembers) > 0 {
			diagErr := createMembers(ctx, d.Id(), addMembers, proxy)
			if diagErr != nil {
				return diagErr
			}
		}
	}

	log.Printf("Updated team %s", *teamObj.Id)
	return readTeam(ctx, d, meta)
}

// deleteTeam is used by the team resource to delete a team from Genesys cloud
func deleteTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)

	log.Printf("Deleting team %s", d.Id())
	resp, err := proxy.deleteTeam(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete team %s error: %s", d.Id(), err), resp)
	}
	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getTeamById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted team %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error deleting team %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("team %s still exists", d.Id()), resp))
	})
}

// readMembers is used by the members resource to read a members from genesys cloud
func readMembers(ctx context.Context, d *schema.ResourceData, proxy *teamProxy) ([]interface{}, error) {
	log.Printf("Reading members of team %s", d.Id())
	teamMemberListing, resp, err := proxy.getMembersById(ctx, d.Id())
	if err != nil {
		log.Printf("unable to retrieve members of team %s : %s %v", d.Id(), err, resp)
		return nil, err
	}
	log.Printf("Read members of team %s", d.Id())
	if teamMemberListing != nil {
		return flattenMemberIds(*teamMemberListing), nil
	}
	return nil, nil
}

// deleteMembers is used by the members resource to delete members from Genesys cloud
func deleteMembers(ctx context.Context, teamId string, memberList []interface{}, proxy *teamProxy) diag.Diagnostics {
	resp, err := proxy.deleteMembers(ctx, teamId, convertMemberListtoString(memberList))
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update remove members from team %s error: %s", teamId, err), resp)
	}
	log.Printf("success removing members from team %s", teamId)
	return nil
}

// createMembers is used by the members resource to create Genesys cloud members
func createMembers(ctx context.Context, teamId string, members []interface{}, proxy *teamProxy) diag.Diagnostics {
	log.Printf("Adding members to team %s", teamId)

	// API does not allow more than 25 members to be added at once, adding members in chunks of 25
	const chunkSize = 25
	var membersChunk []interface{}
	for _, member := range members {
		membersChunk = append(membersChunk, member)
		if len(membersChunk)%chunkSize == 0 {
			_, resp, err := proxy.createMembers(ctx, teamId, buildTeamMembers(membersChunk))
			if err != nil {
				return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to add members to team %s error: %s", teamId, err), resp)
			}
			membersChunk = nil
		}
	}

	_, resp, err := proxy.createMembers(ctx, teamId, buildTeamMembers(membersChunk))
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to add members to team %s error: %s", teamId, err), resp)
	}

	log.Printf("Added members to team %s", teamId)
	return nil
}
