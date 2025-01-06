package routing_queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"
	chunksProcess "terraform-provider-genesyscloud/genesyscloud/util/chunks"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

var ctx = context.Background()

// postRoutingQueueMembers allows up to 100 bulk Member addition/removal per call
func postRoutingQueueMembers(queueID string, membersToUpdate []string, remove bool, proxy *RoutingQueueProxy) diag.Diagnostics {
	// Generic call to prepare chunks for the Update. Takes in three args
	// 1. MemberstoUpdate 2. The Entity prepare func for the update 3. Chunk Size
	if len(membersToUpdate) > 0 {
		chunks := chunksProcess.ChunkItems(membersToUpdate, platformWritableEntityFunc, 100)

		chunkProcessor := func(chunk []platformclientv2.Writableentity) diag.Diagnostics {
			resp, err := proxy.addOrRemoveMembers(ctx, queueID, chunk, remove)
			if err != nil {
				return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update members in queue %s error: %s", queueID, err), resp)
			}
			return nil
		}
		// Generic Function call which takes in the chunks and the processing function
		return chunksProcess.ProcessChunks(chunks, chunkProcessor)
	}
	return nil
}

func getRoutingQueueMembers(queueID string, memberBy string, sdkConfig *platformclientv2.Configuration) ([]platformclientv2.Queuemember, diag.Diagnostics) {
	proxy := GetRoutingQueueProxy(sdkConfig)
	var members []platformclientv2.Queuemember

	// Need to call this method to find the member count for a queue. GetRoutingQueueMembers does not return a `total` property for us to use.
	queue, resp, err := proxy.getRoutingQueueById(ctx, queueID, true)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to find queue %s error: %s", queueID, err), resp)
	}

	if queue.MemberCount == nil {
		log.Printf("no members belong to queue %s", queueID)
		return members, nil
	}

	queueMembers := *queue.MemberCount
	log.Printf("%d members belong to queue %s", queueMembers, queueID)

	for pageNum := 1; ; pageNum++ {
		users, resp, err := sdkGetRoutingQueueMembers(queueID, memberBy, pageNum, 100, sdkConfig)
		if err != nil || resp.StatusCode != http.StatusOK {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to query users for queue %s error: %s", queueID, err), resp)
		}

		if users == nil || users.Entities == nil || len(*users.Entities) == 0 {
			membersFound := len(members)
			log.Printf("%d queue members found for queue %s", membersFound, queueID)

			if membersFound != queueMembers {
				log.Printf("Member count is not equal to queue member found for queue %s, Correlation Id: %s", queueID, resp.CorrelationID)
			}
			return members, nil
		}

		members = append(members, *users.Entities...)
	}
}

func updateQueueMembers(d *schema.ResourceData, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	proxy := GetRoutingQueueProxy(sdkConfig)

	if !d.HasChange("members") {
		return nil
	}

	membersSet, ok := d.Get("members").(*schema.Set)
	if !ok || membersSet.Len() == 0 {
		if err := removeAllExistingUserMembersFromQueue(d.Id(), sdkConfig); err != nil {
			return diag.FromErr(err)
		}
		return nil
	}

	log.Printf("Updating members for Queue %s", d.Get("name"))

	// Get new and Existing users and ring nums
	newUserIds, newUserRingNums := getNewUsersAndRingNums(membersSet)
	oldUserIds, oldUserRingNums, err := getExistingUsersAndRingNums(d.Id(), sdkConfig)
	if err != nil {
		return err
	}

	if diagErr := checkUserMembership(d.Id(), newUserIds, sdkConfig); diagErr != nil {
		return util.BuildDiagnosticError(ResourceType, "failed to update queue member: ", diagErr)
	}

	// Check for members to add or remove
	if diagErr := addOrRemoveMembers(d.Id(), oldUserIds, newUserIds, proxy); err != nil {
		return diagErr
	}

	// Check for ring numbers to update
	if diagErr := updateRingNumbers(d.Id(), newUserRingNums, oldUserRingNums, sdkConfig); diagErr != nil {
		return diagErr
	}

	log.Printf("Members updated for Queue %s", d.Get("name"))
	return nil
}

// removeAllExistingUserMembersFromQueue get all existing user members of a given queue and remove them from the queue
func removeAllExistingUserMembersFromQueue(queueId string, sdkConfig *platformclientv2.Configuration) error {
	proxy := GetRoutingQueueProxy(sdkConfig)

	log.Printf("Reading user members of queue %s", queueId)

	oldSdkUsers, err := getRoutingQueueMembers(queueId, "user", sdkConfig)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	log.Printf("Read user members of queue %s", queueId)

	var oldUserIds []string
	for _, user := range oldSdkUsers {
		oldUserIds = append(oldUserIds, *user.Id)
	}

	if len(oldUserIds) > 0 {
		log.Printf("Removing queue %s user members", queueId)
		if err := postRoutingQueueMembers(queueId, oldUserIds, true, proxy); err != nil {
			return fmt.Errorf("%v", err)
		}
		log.Printf("Removed queue %s user members", queueId)
	}
	return nil
}

func updateRingNumbers(queueID string, newUserRingNums, oldUserRingNums map[string]int, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	for userID, newNum := range newUserRingNums {
		if oldNum, found := oldUserRingNums[userID]; found {
			if newNum != oldNum {
				log.Printf("updating ring_num for user %s because it has updated. New: %v, Old: %v", userID, newNum, oldNum)
				if err := updateQueueUserRingNum(queueID, userID, newNum, sdkConfig); err != nil {
					return err
				}
			}
		} else if newNum != 1 {
			log.Printf("updating user %s ring_num because it is not the default 1", userID)
			if err := updateQueueUserRingNum(queueID, userID, newNum, sdkConfig); err != nil {
				return err
			}
		}
	}
	return nil
}

func updateQueueUserRingNum(queueID string, userID string, ringNum int, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	proxy := GetRoutingQueueProxy(sdkConfig)

	log.Printf("Updating ring number for queue %s user %s", queueID, userID)
	resp, err := proxy.updateRoutingQueueMember(ctx, queueID, userID, platformclientv2.Queuemember{
		Id:         &userID,
		RingNumber: &ringNum,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update ring number for queue %s user %s error: %s", queueID, userID, err), resp)
	}
	return nil
}

func sdkGetRoutingQueueMembers(queueID, memberBy string, pageNumber, pageSize int, sdkConfig *platformclientv2.Configuration) (*platformclientv2.Queuememberentitylisting, *platformclientv2.APIResponse, error) {
	api := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	// SDK does not support nil values for boolean query params yet, so we must manually construct this HTTP request for now
	apiClient := &api.Configuration.APIClient

	// create path and map variables
	path := api.Configuration.BasePath + "/api/v2/routing/queues/{queueId}/members"
	path = strings.Replace(path, "{queueId}", queueID, -1)

	headerParams := make(map[string]string)
	queryParams := make(map[string]string)
	formParams := url.Values{}
	var postBody interface{}
	var postFileName string
	var postFilePath string
	var fileBytes []byte

	// oauth required
	if api.Configuration.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + api.Configuration.AccessToken
	}
	// add default headers if any
	for key := range api.Configuration.DefaultHeader {
		headerParams[key] = api.Configuration.DefaultHeader[key]
	}

	queryParams["pageSize"] = apiClient.ParameterToString(pageSize, "")
	queryParams["pageNumber"] = apiClient.ParameterToString(pageNumber, "")
	if memberBy != "" {
		queryParams["memberBy"] = memberBy
	}

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload *platformclientv2.Queuememberentitylisting
	response, err := apiClient.CallAPI(path, http.MethodGet, postBody, headerParams, queryParams, formParams, postFileName, fileBytes, postFilePath)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		err = json.Unmarshal([]byte(response.RawBody), &successPayload)
	}
	return successPayload, response, err
}

func checkUserMembership(queueId string, newUserIds []string, sdkConfig *platformclientv2.Configuration) error {
	if len(newUserIds) > 0 {
		log.Printf("Sleeping for 10 seconds")
		time.Sleep(10 * time.Second)

		members, diagErr := getRoutingQueueMembers(queueId, "group", sdkConfig)
		if diagErr != nil {
			return fmt.Errorf("%v", diagErr)
		}

		for _, userId := range newUserIds {
			if err := verifyUserIsNotGroupMemberOfQueue(queueId, userId, members); err != nil {
				return err
			}
		}
	}
	return nil
}

// verifyUserIsNotGroupMemberOfQueue Search through queue group members to verify that a given user is not a group member
func verifyUserIsNotGroupMemberOfQueue(queueId, userId string, members []platformclientv2.Queuemember) error {
	log.Printf("verifying that member '%s' is not assigned to the queue '%s' via a group", userId, queueId)

	for _, member := range members {
		if *member.Id == userId {
			return fmt.Errorf("member %s  is already assigned to queue %s via a group, and therefore should not be assigned as a member", userId, queueId)
		}
	}

	log.Printf("User %s not found as group member in queue %s", userId, queueId)
	return nil
}

func getNewUsersAndRingNums(membersSet *schema.Set) ([]string, map[string]int) {
	newUserRingNums := make(map[string]int)
	memberList := membersSet.List()
	newUserIds := make([]string, len(memberList))

	for i, member := range memberList {
		memberMap := member.(map[string]interface{})
		newUserIds[i] = memberMap["user_id"].(string)
		newUserRingNums[newUserIds[i]] = memberMap["ring_num"].(int)
	}
	return newUserIds, newUserRingNums
}

func getExistingUsersAndRingNums(queueID string, sdkConfig *platformclientv2.Configuration) ([]string, map[string]int, diag.Diagnostics) {
	oldSdkUsers, err := getRoutingQueueMembers(queueID, "user", sdkConfig)
	if err != nil {
		return nil, nil, err
	}

	oldUserIds := make([]string, len(oldSdkUsers))
	oldUserRingNums := make(map[string]int)

	for i, user := range oldSdkUsers {
		oldUserIds[i] = *user.Id
		oldUserRingNums[oldUserIds[i]] = *user.RingNumber
	}

	return oldUserIds, oldUserRingNums, nil
}

func addOrRemoveMembers(queueId string, oldUserIds, newUserIds []string, proxy *RoutingQueueProxy) diag.Diagnostics {
	// Remove From Queue
	if len(oldUserIds) > 0 {
		usersToRemove := lists.SliceDifference(oldUserIds, newUserIds)
		if err := postRoutingQueueMembers(queueId, usersToRemove, true, proxy); err != nil {
			return err
		}
	}

	// Add To Queue
	if len(newUserIds) > 0 {
		usersToAdd := lists.SliceDifference(newUserIds, oldUserIds)
		if err := postRoutingQueueMembers(queueId, usersToAdd, false, proxy); err != nil {
			return err
		}
	}
	return nil
}
