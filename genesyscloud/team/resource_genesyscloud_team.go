package team

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

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
		return nil, diag.Errorf("Failed to get team: %v %v", err, resp)
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
		return diag.Errorf("Failed to create team: %s %v", err, resp)
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

// readTeam is used by the team resource to read an team from genesys cloud
func readTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)
	log.Printf("Reading team %s", d.Id())
	// reading members
	members, err := readMembers(ctx, d, proxy)
	if err != nil {
		return diag.Errorf("failed to read members of the team %s : %s", d.Id(), err)
	}
	if members != nil {
		d.Set("member_ids", members)
	}
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		team, resp, getErr := proxy.getTeamById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read team %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read team %s: %s", d.Id(), getErr))
		}
		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTeam())
		resourcedata.SetNillableValue(d, "name", team.Name)
		resourcedata.SetNillableReferenceWritableDivision(d, "division_id", team.Division)
		resourcedata.SetNillableValue(d, "description", team.Description)
		log.Printf("Read team %s %s", d.Id(), *team.Name)
		return cc.CheckState()
	})
}

// updateTeam is used by the team resource to update an team in Genesys Cloud
func updateTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)
	team := getTeamFromResourceData(d)
	log.Printf("updating team %s", *team.Name)
	teamObj, resp, err := proxy.updateTeam(ctx, d.Id(), &team)
	if err != nil {
		return diag.Errorf("failed to update team %s : %s %v", d.Id(), err, resp)
	}
	members, ok := d.GetOk("member_ids")

	if ok {
		memberList := members.([]interface{})
		if len(memberList) == 0 {
			currentMembers, _ := readMembers(ctx, d, proxy)
			if len(currentMembers) > 0 {
				deleteMembers(ctx, d.Id(), currentMembers, proxy)
			}
		}
		if len(memberList) > 0 {
			currentMembers, _ := readMembers(ctx, d, proxy)
			if len(currentMembers) > 0 {
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
		}
	}
	log.Printf("Updated team %s", *teamObj.Id)
	return readTeam(ctx, d, meta)

}

// deleteTeam is used by the team resource to delete an team from Genesys cloud
func deleteTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)
	resp, err := proxy.deleteTeam(ctx, d.Id())
	if err != nil {
		return diag.Errorf("failed to delete team %s: %s %v", d.Id(), err, resp)
	}
	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getTeamById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("deleted team %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting team %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("team %s still exists", d.Id()))
	})
}

// readMembers is used by the members resource to read a members from genesys cloud
func readMembers(ctx context.Context, d *schema.ResourceData, proxy *teamProxy) ([]interface{}, error) {
	log.Printf("attempting to read members of team %s", d.Id())
	teamMemberListing, resp, err := proxy.getMembersById(ctx, d.Id())
	if err != nil {
		log.Printf("unable to retrieve members of team %s : %s %v", d.Id(), err, resp)
		return nil, err
	}
	log.Printf("success reading members of team %s", d.Id())
	if teamMemberListing != nil {
		return flattenMemberIds(*teamMemberListing), nil
	}
	return nil, nil
}

// deleteMembers is used by the members resource to delete a members from Genesys cloud
func deleteMembers(ctx context.Context, teamId string, memberList []interface{}, proxy *teamProxy) diag.Diagnostics {
	resp, err := proxy.deleteMembers(ctx, teamId, convertMemberListtoString(memberList))
	if err != nil {
		return diag.Errorf("failed to remove members from team %s : %s %v", teamId, err, resp)
	}
	log.Printf("success removing members from team %s", teamId)
	return nil
}

// createMembers is used by the members resource to create Genesys cloud members
func createMembers(ctx context.Context, teamId string, members []interface{}, proxy *teamProxy) diag.Diagnostics {
	log.Printf("adding members to team %s", teamId)

	// API does not allow more than 25 members to be added at once, adding members in chunks of 25
	const chunkSize = 25
	var membersChunk []interface{}
	for _, member := range members {
		membersChunk = append(membersChunk, member)
		if len(membersChunk)%chunkSize == 0 {
			_, resp, err := proxy.createMembers(ctx, teamId, buildTeamMembers(membersChunk))
			if err != nil {
				return diag.Errorf("failed to add members to team %s: %s %v", teamId, err, resp)
			}
			membersChunk = nil
		}
	}

	_, resp, err := proxy.createMembers(ctx, teamId, buildTeamMembers(membersChunk))
	if err != nil {
		return diag.Errorf("failed to add members to team %s: %s %v", teamId, err, resp)
	}

	log.Printf("success adding members to team %s", teamId)
	return nil
}
