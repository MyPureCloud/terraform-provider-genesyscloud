package group

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/chunks"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

func updateGroupMembers(ctx context.Context, d *schema.ResourceData, sdkConfig *platformclientv2.Configuration) (diags diag.Diagnostics) {
	gp := getGroupProxy(sdkConfig)

	if !d.HasChange("member_ids") {
		return diags
	}

	membersConfig := d.Get("member_ids")
	if membersConfig == nil {
		return diags
	}

	log.Printf("Updating members for '%s'", d.Id())

	configMemberIds := *lists.SetToStringList(membersConfig.(*schema.Set))

	log.Printf("Reading member IDs for group '%s'", d.Id())

	existingMemberIds, getMemberIdsDiags := getGroupMemberIds(ctx, d, sdkConfig)
	diags = append(diags, getMemberIdsDiags...)
	if diags.HasError() {
		log.Printf("Encountered error while reading member IDs for group '%s': %v", d.Id(), getMemberIdsDiags)
		return diags
	}

	maxMembersPerRequest := 50
	membersToRemoveList := lists.SliceDifference(existingMemberIds, configMemberIds)
	chunkedMemberIdsDelete := chunks.ChunkBy(membersToRemoveList, maxMembersPerRequest)

	chunkProcessor := func(membersToRemove []string) diag.Diagnostics {
		if len(membersToRemove) == 0 {
			log.Printf("No members to remove for group '%s'", d.Id())
			return nil
		}

		log.Printf("Removing %d members for group '%s'", len(membersToRemove), d.Id())
		return util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
			_, resp, err := gp.deleteGroupMembers(ctx, d.Id(), strings.Join(membersToRemove, ","))
			if err != nil {
				log.Printf("Encountered error while removing members from gorup '%s': %s", d.Id(), err.Error())
				return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to remove members from group %s: %s", d.Id(), err), resp)
			}
			log.Printf("Successfully removed %d members from group '%s", len(membersToRemove), d.Id())
			return resp, nil
		})
	}

	log.Printf("Beginning chunking process for group '%s'", d.Id())
	diags = append(diags, chunks.ProcessChunks(chunkedMemberIdsDelete, chunkProcessor)...)
	if diags.HasError() {
		log.Printf("Encountered error while invoking chunk processor for group '%s': %v", d.Id(), diags)
		return diags
	}

	membersToAdd := lists.SliceDifference(configMemberIds, existingMemberIds)
	if len(membersToAdd) < 1 {
		log.Printf("No members to add for group '%s'", d.Id())
		return diags
	}

	chunkedMemberIds := lists.ChunkStringSlice(membersToAdd, maxMembersPerRequest)
	for _, chunk := range chunkedMemberIds {
		diags = append(diags, addGroupMembers(ctx, d, chunk, sdkConfig)...)
	}

	return diags
}

func readGroupMembers(ctx context.Context, groupID string, sdkConfig *platformclientv2.Configuration) (*schema.Set, *platformclientv2.APIResponse, error) {
	gp := getGroupProxy(sdkConfig)

	members, resp, err := gp.getGroupMembers(ctx, groupID)
	if err != nil {
		return nil, resp, err
	}

	interfaceList := make([]interface{}, len(*members))
	for i, v := range *members {
		interfaceList[i] = v
	}

	return schema.NewSet(schema.HashString, interfaceList), resp, nil
}

func getGroupMemberIds(ctx context.Context, d *schema.ResourceData, sdkConfig *platformclientv2.Configuration) ([]string, diag.Diagnostics) {
	gp := getGroupProxy(sdkConfig)
	members, resp, err := gp.getGroupMembers(ctx, d.Id())
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Unable to retrieve members for group %s. %s", d.Id(), err), resp)
	}
	return *members, nil
}

func addGroupMembers(ctx context.Context, d *schema.ResourceData, membersToAdd []string, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	gp := getGroupProxy(sdkConfig)
	return util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Need the current group version to add members
		log.Printf("Reading group '%s' version", d.Id())
		groupInfo, resp, getErr := gp.getGroupById(ctx, d.Id())
		if getErr != nil {
			log.Printf("Encountered error while reading group '%s' version: %s", d.Id(), getErr.Error())
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read group %s: %s", d.Id(), getErr), resp)
		}

		log.Printf("Adding %d members to group '%s'", len(membersToAdd), d.Id())
		groupMemberUpdate := &platformclientv2.Groupmembersupdate{
			MemberIds: &membersToAdd,
			Version:   groupInfo.Version,
		}
		_, resp, postErr := gp.addGroupMembers(ctx, d.Id(), groupMemberUpdate)
		if postErr != nil {
			log.Printf("Encountered error while adding members to group '%s': %s", d.Id(), postErr.Error())
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to add group members %s: %s", d.Id(), postErr), resp)
		}
		return resp, nil
	})
}

// 'number' and 'extension' conflict with eachother. However, one must be set.
// This function validates that the user has satisfied these conditions
func validateAddressesMap(m map[string]interface{}) error {
	number, _ := m["number"].(string)
	extension, _ := m["extension"].(string)

	if (number != "" && extension != "") ||
		(number == "" && extension == "") {
		return fmt.Errorf("either 'number' or 'extension' must be set inside addresses, but both cannot be set")
	}

	return nil
}

func flattenGroupAddresses(d *schema.ResourceData, addresses *[]platformclientv2.Groupcontact) []interface{} {
	addressSlice := make([]interface{}, 0)
	utilE164 := util.NewUtilE164Service()
	for _, address := range *addresses {
		if address.MediaType != nil {
			if *address.MediaType == groupPhoneType {
				phoneNumber := make(map[string]interface{})

				// Strip off any parentheses from phone numbers
				if address.Address != nil {
					phoneNumber["number"] = utilE164.FormatAsCalculatedE164Number(strings.Trim(*address.Address, "()"))
				}

				resourcedata.SetMapValueIfNotNil(phoneNumber, "extension", address.Extension)
				resourcedata.SetMapValueIfNotNil(phoneNumber, "type", address.VarType)

				// Sometimes the number or extension is only returned in Display
				if address.Address == nil &&
					address.Extension == nil &&
					address.Display != nil {
					setExtensionOrNumberBasedOnDisplay(d, phoneNumber, &address, utilE164)
				}

				addressSlice = append(addressSlice, phoneNumber)
			} else {
				log.Printf("Unknown address media type %s", *address.MediaType)
			}
		}
	}
	return addressSlice
}

/**
*  The api can sometimes return only the display which holds the value
*  that was stored in either `address` or `extension`
*  This function establishes which field was set in the schema data (`extension` or `address`)
*  and then sets that field in the map to the value that came back in `display`
 */
func setExtensionOrNumberBasedOnDisplay(d *schema.ResourceData, addressMap map[string]interface{}, address *platformclientv2.Groupcontact, utilE164 *util.UtilE164Service) {
	display := strings.Trim(*address.Display, "()")
	schemaAddresses := d.Get("addresses").([]interface{})
	for _, a := range schemaAddresses {
		currentAddress, ok := a.(map[string]interface{})
		if !ok {
			continue
		}
		addressType, _ := currentAddress["type"].(string)
		if addressType != *address.VarType {
			continue
		}
		if ext, _ := currentAddress["extension"].(string); ext != "" {
			addressMap["extension"] = display
		} else if number, _ := currentAddress["number"].(string); number != "" {
			addressMap["number"] = utilE164.FormatAsCalculatedE164Number(display)
		}
	}
}

func flattenGroupOwners(owners *[]platformclientv2.User) []interface{} {
	interfaceList := make([]interface{}, len(*owners))
	for i, v := range *owners {
		interfaceList[i] = *v.Id
	}
	return interfaceList
}

func buildSdkGroupAddresses(d *schema.ResourceData) (*[]platformclientv2.Groupcontact, error) {
	if addressSlice, ok := d.Get("addresses").([]interface{}); ok && len(addressSlice) > 0 {
		sdkContacts := make([]platformclientv2.Groupcontact, len(addressSlice))
		for i, configPhone := range addressSlice {
			phoneMap := configPhone.(map[string]interface{})
			phoneType := phoneMap["type"].(string)
			contact := platformclientv2.Groupcontact{
				VarType:   &phoneType,
				MediaType: &groupPhoneType, // Only option is PHONE
			}

			if err := validateAddressesMap(phoneMap); err != nil {
				return nil, err
			}

			if phoneNum, ok := phoneMap["number"].(string); ok && phoneNum != "" {
				contact.Address = &phoneNum
			}

			if phoneExt := phoneMap["extension"].(string); ok && phoneExt != "" {
				contact.Extension = &phoneExt
			}

			sdkContacts[i] = contact
		}
		return &sdkContacts, nil
	}
	return nil, nil
}

func GenerateBasicGroupResource(resourceLabel string, name string, nestedBlocks ...string) string {
	return GenerateGroupResource(resourceLabel, name, util.NullValue, util.NullValue, util.NullValue, util.TrueValue, nestedBlocks...)
}

func GenerateGroupResource(
	resourceLabel string,
	name string,
	desc string,
	groupType string,
	visibility string,
	rulesVisible string,
	nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_group" "%s" {
		name = "%s"
		description = %s
		type = %s
		visibility = %s
		rules_visible = %s
        %s
	}
	`, resourceLabel, name, desc, groupType, visibility, rulesVisible, strings.Join(nestedBlocks, "\n"))
}

func generateGroupAddress(number string, phoneType string, extension string) string {
	return fmt.Sprintf(`addresses {
		number = %s
		type = "%s"
		extension = %s
	}
	`, number, phoneType, extension)
}

func GenerateGroupOwners(userIDs ...string) string {
	return fmt.Sprintf(`owner_ids = [%s]
	`, strings.Join(userIDs, ","))
}

func generateGroupMembers(userIDs ...string) string {
	return fmt.Sprintf(`member_ids = [%s]
	`, strings.Join(userIDs, ","))
}
