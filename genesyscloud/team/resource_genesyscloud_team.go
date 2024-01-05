package team

import (
	"context"
	"fmt"
	"log"
	"strings"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v116/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

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
	teams, err := proxy.getAllTeam(ctx, "")
	if err != nil {
		return nil, diag.Errorf("Failed to get team: %v", err)
	}

	for _, team := range *teams {
		resources[*team.Id] = &resourceExporter.ResourceMeta{Name: *team.Name}
	}

	return resources, nil
}

// createTeam is used by the team resource to create Genesys cloud team
func createTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)

	team := getTeamFromResourceData(d)

	log.Printf("Creating team %s", *team.Name)
	teamObj, err := proxy.createTeam(ctx, &team)
	if err != nil {
		return diag.Errorf("Failed to create team: %s", err)
	}

	d.SetId(*teamObj.Id)
	log.Printf("Created team %s", *teamObj.Id)
	//adding members to the team
	members, ok := d.GetOk("member_ids")
	if ok {
		memberList := members.([]interface{})
		//creating members along with teams
		if len(memberList) > 0 {
			_, err := proxy.createMembers(ctx, d.Id(), buildTeamMembers(memberList))
			if err != nil {
				return diag.Errorf("Failed to create members: %s", err)
			}
			log.Printf("Created members %s", d.Id())
		}
	}

	return readTeam(ctx, d, meta)
}

// readTeam is used by the team resource to read an team from genesys cloud
func readTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)

	log.Printf("Reading team %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		team, respCode, getErr := proxy.getTeamById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("failed to read team %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read team %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTeam())

		resourcedata.SetNillableValue(d, "name", team.Name)

		resourcedata.SetNillableReferenceWritableDivision(d, "division_id", team.Division)

		resourcedata.SetNillableValue(d, "description", team.Description)

		log.Printf("Read team %s %s", d.Id(), *team.Name)

		// reading members
		members, err := readMembers(ctx, d, proxy)
		if err != nil {
			resourcedata.SetNillableValue(d, "member_ids", flattenMemberIds(members))
		}

		return cc.CheckState()
	})
}

// updateTeam is used by the team resource to update an team in Genesys Cloud
func updateTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)

	team := getTeamFromResourceData(d)

	log.Printf("Updating team %s", *team.Name)
	teamObj, err := proxy.updateTeam(ctx, d.Id(), &team)
	if err != nil {
		return diag.Errorf("Failed to update team: %s", err)
	}

	//check if member list is present
	members, ok := d.GetOk("member_ids")
	if ok {
		memberList := members.([]interface{})
		// check if memberList is Empty
		if len(memberList) == 0 {
			//delete members from the team if memeber list is empty
			currentMembers, err := readMembers(ctx, d, proxy)
			if err != nil {
				deleteMembers(ctx, d.Id(), currentMembers, proxy)
			}
		}

		// get current members and do add/remove based on the difference
		if len(memberList) > 0 {
			currentMembers, err := readMembers(ctx, d, proxy)
			if err != nil {
				removeMembers, addMembers := SliceDifferenceMembers(currentMembers, memberList)
				deleteMembers(ctx, d.Id(), removeMembers, proxy)
				createMembers(ctx, d.Id(), addMembers, proxy)
			}
		}

		log.Printf("Updated team %s", *teamObj.Id)
	}
	return readTeam(ctx, d, meta)

}

// deleteTeam is used by the team resource to delete an team from Genesys cloud
func deleteTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)

	_, err := proxy.deleteTeam(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete team %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getTeamById(ctx, d.Id())

		if err != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				log.Printf("Deleted team %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting team %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("team %s still exists", d.Id()))
	})
}

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

// readMembers is used by the members resource to read a members from genesys cloud
func readMembers(ctx context.Context, d *schema.ResourceData, proxy *teamProxy) ([]interface{}, error) {
	log.Printf("reading members %s", d.Id())
	teamMemberListing, err := proxy.getMembersById(ctx, d.Id())
	if err != nil {
		return nil, err
	}
	if teamMemberListing != nil {
		log.Printf("success reading members %s", d.Id())
		return flattenMemberIds(*teamMemberListing), nil
	}
	return nil, nil
}

// deleteMembers is used by the members resource to delete a members from Genesys cloud
func deleteMembers(ctx context.Context, teamId string, memberList []interface{}, proxy *teamProxy) diag.Diagnostics {

	_, err := proxy.deleteMembers(ctx, teamId, convertMemberListtoString(memberList))
	if err != nil {
		return diag.Errorf("Failed to delete members %s: %s", teamId, err)
	}
	log.Printf("success deleting members %s", teamId)
	return nil
}

// createMembers is used by the members resource to create Genesys cloud members
func createMembers(ctx context.Context, teamId string, members []interface{}, proxy *teamProxy) diag.Diagnostics {

	log.Printf("Creating members for team %s", teamId)
	_, err := proxy.createMembers(ctx, teamId, buildTeamMembers(members))
	if err != nil {
		return diag.Errorf("Failed to create members: %s", err)
	}

	log.Printf("Created members %s", teamId)
	return nil
}

func buildTeamMembers(teamMembers []interface{}) platformclientv2.Teammembers {
	var teamMemberObject platformclientv2.Teammembers
	members := make([]string, len(teamMembers))
	for i, member := range teamMembers {
		members[i] = member.(string)
	}
	teamMemberObject.MemberIds = &members
	return teamMemberObject
}

func convertMemberListtoString(teamMembers []interface{}) string {
	var memberList []string
	for _, v := range teamMembers {
		memberList = append(memberList, v.(string))
	}
	memberString := strings.Join(memberList, ", ")
	return memberString
}

func flattenMemberIds(teamEntityListing []platformclientv2.Userreferencewithname) []interface{} {
	memberList := []interface{}{}

	if len(teamEntityListing) == 0 {
		return nil
	}
	for _, teamEntity := range teamEntityListing {
		memberList = append(memberList, teamEntity.Id)
	}
	return memberList
}

func SliceDifferenceMembers(current, target []interface{}) ([]interface{}, []interface{}) {
	var remove []interface{}
	var add []interface{}

	keysTarget := make(map[interface{}]bool)
	keysCurrent := make(map[interface{}]bool)

	for _, item := range target {
		keysTarget[item] = true
	}

	for _, item := range current {
		keysCurrent[item] = true
	}

	for _, item := range current {
		if _, found := keysTarget[item]; !found {
			remove = append(remove, item)
		}
	}

	for _, item := range target {
		if _, found := keysCurrent[item]; !found {
			add = append(add, item)
		}
	}
	return remove, add
}
