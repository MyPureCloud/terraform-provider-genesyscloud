package team

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

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
	memberString := strings.Join(memberList, ",")
	log.Printf("member list is %s", memberString)
	return memberString
}

func flattenMemberIds(teamEntityListing []platformclientv2.Userreferencewithname) []interface{} {
	memberList := []interface{}{}
	if len(teamEntityListing) == 0 {
		return nil
	}
	for _, teamEntity := range teamEntityListing {
		memberList = append(memberList, *teamEntity.Id)
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

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	s := make([]byte, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func generateTeamsWithMemberResource(
	teamResource string,
	name string,
	member_ids []string,
	divisionId string,
) string {
	returnString := fmt.Sprintf(`resource "genesyscloud_team" "%s" {
		name = "%s"
		member_ids = [%s]
		division_id = %s
	}
	`, teamResource, name, strings.Join(member_ids, ", "), divisionId)
	return returnString
}

func GenerateUserWithDivisionId(resourceID string, name string, email string, divisionId string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		name = "%s"
		email = "%s"
		division_id = %s
	}
	`, resourceID, name, email, divisionId)
}

func generateTeamResource(
	teamResource string,
	name string,
	divisionId string,
	description string) string {
	return fmt.Sprintf(`resource "genesyscloud_team" "%s" {
		name = "%s"
		division_id = %s
		description = "%s"
	}
	`, teamResource, name, divisionId, description)
}
