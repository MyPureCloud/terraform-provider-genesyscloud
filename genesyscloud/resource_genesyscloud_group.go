package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/chunks"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

var (
	groupPhoneType       = "PHONE"
	groupAddressResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"number": {
				Description:      "Phone number for this contact type. Must be in an E.164 number format.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: ValidatePhoneNumber,
			},
			"extension": {
				Description: "Phone extension.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"type": {
				Description:  "Contact type of the address. (GROUPRING | GROUPPHONE)",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"GROUPRING", "GROUPPHONE"}, false),
			},
		},
	}
)

func getAllGroups(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	groupsAPI := platformclientv2.NewGroupsApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		groups, _, getErr := groupsAPI.GetGroups(pageSize, pageNum, nil, nil, "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of groups: %v", getErr)
		}

		if groups.Entities == nil || len(*groups.Entities) == 0 {
			break
		}

		for _, group := range *groups.Entities {
			resources[*group.Id] = &resourceExporter.ResourceMeta{Name: *group.Name}
		}
	}

	return resources, nil
}

func GroupExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllGroups),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"owner_ids":  {RefType: "genesyscloud_user"},
			"member_ids": {RefType: "genesyscloud_user"},
		},
		E164Numbers: []string{"addresses.number"},
	}
}

func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Directory Group",

		CreateContext: CreateWithPooledClient(createGroup),
		ReadContext:   ReadWithPooledClient(readGroup),
		UpdateContext: UpdateWithPooledClient(updateGroup),
		DeleteContext: DeleteWithPooledClient(deleteGroup),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Group name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Group description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"type": {
				Description:  "Group type (official | social). This cannot be modified. Changing type attribute will cause the existing genesys_group object to dropped and recreated with a new ID.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "official",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"official", "social"}, false),
			},
			"visibility": {
				Description:  "Who can view this group (public | owners | members).",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "public",
				ValidateFunc: validation.StringInSlice([]string{"public", "owners", "members"}, false),
			},
			"rules_visible": {
				Description: "Are membership rules visible to the person requesting to view the group.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"addresses": {
				Description: "Contact numbers for this group.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        groupAddressResource,
			},
			"owner_ids": {
				Description: "IDs of owners of the group.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"member_ids": {
				Description: "IDs of members assigned to the group. If not set, this resource will not manage group members.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func createGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	groupType := d.Get("type").(string)
	visibility := d.Get("visibility").(string)
	rulesVisible := d.Get("rules_visible").(bool)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	groupsAPI := platformclientv2.NewGroupsApiWithConfig(sdkConfig)

	addresses, err := buildSdkGroupAddresses(d)
	if err != nil {
		return diag.Errorf("%v", err)
	}

	log.Printf("Creating group %s", name)
	group, _, err := groupsAPI.PostGroups(platformclientv2.Groupcreate{
		Name:         &name,
		VarType:      &groupType,
		Visibility:   &visibility,
		RulesVisible: &rulesVisible,
		Addresses:    addresses,
		OwnerIds:     lists.BuildSdkStringListFromInterfaceArray(d, "owner_ids"),
	})
	if err != nil {
		return diag.Errorf("Failed to create group %s: %s", name, err)
	}

	d.SetId(*group.Id)

	// Description can only be set in a PUT. This is a bug with the API and has been reported
	if description != "" {
		diagErr := updateGroup(ctx, d, meta)
		if diagErr != nil {
			return diagErr
		}
	}

	diagErr := updateGroupMembers(d, groupsAPI)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Created group %s %s", name, *group.Id)
	return readGroup(ctx, d, meta)
}

func readGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	groupsAPI := platformclientv2.NewGroupsApiWithConfig(sdkConfig)

	log.Printf("Reading group %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		group, resp, getErr := groupsAPI.GetGroup(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read group %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read group %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceGroup())

		resourcedata.SetNillableValue(d, "name", group.Name)
		resourcedata.SetNillableValue(d, "type", group.VarType)
		resourcedata.SetNillableValue(d, "visibility", group.Visibility)
		resourcedata.SetNillableValue(d, "rules_visible", group.RulesVisible)
		resourcedata.SetNillableValue(d, "description", group.Description)

		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "owner_ids", group.Owners, flattenGroupOwners)

		if group.Addresses != nil {
			d.Set("addresses", flattenGroupAddresses(d, group.Addresses))
		} else {
			d.Set("addresses", nil)
		}

		members, err := readGroupMembers(d.Id(), groupsAPI)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("%v", err))
		}
		d.Set("member_ids", members)

		log.Printf("Read group %s %s", d.Id(), *group.Name)
		return cc.CheckState()
	})
}

func updateGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	visibility := d.Get("visibility").(string)
	rulesVisible := d.Get("rules_visible").(bool)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	groupsAPI := platformclientv2.NewGroupsApiWithConfig(sdkConfig)

	diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current group version
		group, resp, getErr := groupsAPI.GetGroup(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read group %s: %s", d.Id(), getErr)
		}

		addresses, err := buildSdkGroupAddresses(d)
		if err != nil {
			return nil, diag.Errorf("%v", err)
		}

		log.Printf("Updating group %s", name)
		_, resp, putErr := groupsAPI.PutGroup(d.Id(), platformclientv2.Groupupdate{
			Version:      group.Version,
			Name:         &name,
			Description:  &description,
			Visibility:   &visibility,
			RulesVisible: &rulesVisible,
			Addresses:    addresses,
			OwnerIds:     lists.BuildSdkStringListFromInterfaceArray(d, "owner_ids"),
		})
		if putErr != nil {
			return resp, diag.Errorf("Failed to update group %s: %s", d.Id(), putErr)
		}

		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateGroupMembers(d, groupsAPI)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated group %s", name)
	return readGroup(ctx, d, meta)
}

func deleteGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	groupsAPI := platformclientv2.NewGroupsApiWithConfig(sdkConfig)

	RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Directory occasionally returns version errors on deletes if an object was updated at the same time.
		log.Printf("Deleting group %s", name)
		resp, err := groupsAPI.DeleteGroup(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete group %s: %s", name, err)
		}
		return nil, nil
	})

	return WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		group, resp, err := groupsAPI.GetGroup(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				log.Printf("Group %s deleted", name)
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting group %s: %s", d.Id(), err))
		}

		if group.State != nil && *group.State == "deleted" {
			log.Printf("Group %s deleted", name)
			return nil
		}

		return retry.RetryableError(fmt.Errorf("Group %s still exists", d.Id()))
	})
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

// 'number' and 'extension' conflict with eachother. However, one must be set.
// This function validates that the user has satisfied these conditions
func validateAddressesMap(m map[string]interface{}) error {
	number, _ := m["number"].(string)
	extension, _ := m["extension"].(string)

	if (number != "" && extension != "") ||
		(number == "" && extension == "") {
		return fmt.Errorf("Either 'number' or 'extension' must be set inside addresses, but both cannot be set.")
	}

	return nil
}

func flattenGroupAddresses(d *schema.ResourceData, addresses *[]platformclientv2.Groupcontact) []interface{} {
	addressSlice := make([]interface{}, 0)
	for _, address := range *addresses {
		if address.MediaType != nil {
			if *address.MediaType == groupPhoneType {
				phoneNumber := make(map[string]interface{})

				// Strip off any parentheses from phone numbers
				if address.Address != nil {
					phoneNumber["number"] = strings.Trim(*address.Address, "()")
				}

				resourcedata.SetMapValueIfNotNil(phoneNumber, "extension", address.Extension)
				resourcedata.SetMapValueIfNotNil(phoneNumber, "type", address.VarType)

				// Sometimes the number or extension is only returned in Display
				if address.Address == nil &&
					address.Extension == nil &&
					address.Display != nil {
					setExtensionOrNumberBasedOnDisplay(d, phoneNumber, &address)
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
func setExtensionOrNumberBasedOnDisplay(d *schema.ResourceData, addressMap map[string]interface{}, address *platformclientv2.Groupcontact) {
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
			addressMap["number"] = display
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

func updateGroupMembers(d *schema.ResourceData, groupsAPI *platformclientv2.GroupsApi) diag.Diagnostics {
	if d.HasChange("member_ids") {
		if membersConfig := d.Get("member_ids"); membersConfig != nil {
			configMemberIds := *lists.SetToStringList(membersConfig.(*schema.Set))
			existingMemberIds, err := getGroupMemberIds(d, groupsAPI)
			if err != nil {
				return err
			}

			maxMembersPerRequest := 50
			membersToRemoveList := lists.SliceDifference(existingMemberIds, configMemberIds)
			chunkedMemberIdsDelete := chunks.ChunkBy(membersToRemoveList, maxMembersPerRequest)

			chunkProcessor := func(membersToRemove []string) diag.Diagnostics {
				if len(membersToRemove) > 0 {
					if diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
						_, resp, err := groupsAPI.DeleteGroupMembers(d.Id(), strings.Join(membersToRemove, ","))
						if err != nil {
							return resp, diag.Errorf("Failed to remove members from group %s: %s", d.Id(), err)
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
				if err := addGroupMembers(d, chunk, groupsAPI); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func readGroupMembers(groupID string, groupsAPI *platformclientv2.GroupsApi) (*schema.Set, diag.Diagnostics) {
	members, _, err := groupsAPI.GetGroupIndividuals(groupID)
	if err != nil {
		return nil, diag.Errorf("Failed to read members for group %s: %s", groupID, err)
	}

	if members.Entities != nil {
		interfaceList := make([]interface{}, len(*members.Entities))
		for i, v := range *members.Entities {
			interfaceList[i] = *v.Id
		}
		return schema.NewSet(schema.HashString, interfaceList), nil
	}
	return nil, nil
}

func getGroupMemberIds(d *schema.ResourceData, groupsAPI *platformclientv2.GroupsApi) ([]string, diag.Diagnostics) {
	members, _, err := groupsAPI.GetGroupIndividuals(d.Id())
	if err != nil {
		return nil, diag.FromErr(err)
	}

	var existingMembers []string
	if members.Entities != nil {
		for _, member := range *members.Entities {
			existingMembers = append(existingMembers, *member.Id)
		}
	}
	return existingMembers, nil
}

func addGroupMembers(d *schema.ResourceData, membersToAdd []string, groupsAPI *platformclientv2.GroupsApi) diag.Diagnostics {
	if diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Need the current group version to add members
		groupInfo, _, getErr := groupsAPI.GetGroup(d.Id())
		if getErr != nil {
			return nil, diag.Errorf("Failed to read group %s: %s", d.Id(), getErr)
		}

		_, resp, postErr := groupsAPI.PostGroupMembers(d.Id(), platformclientv2.Groupmembersupdate{
			MemberIds: &membersToAdd,
			Version:   groupInfo.Version,
		})
		if postErr != nil {
			return resp, diag.Errorf("Failed to add group members %s: %s", d.Id(), postErr)
		}
		return resp, nil
	}); diagErr != nil {
		return diagErr
	}
	return nil
}

func GenerateBasicGroupResource(resourceID string, name string, nestedBlocks ...string) string {
	return generateGroupResource(resourceID, name, nullValue, nullValue, nullValue, trueValue, nestedBlocks...)
}

func generateGroupResource(
	resourceID string,
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
	`, resourceID, name, desc, groupType, visibility, rulesVisible, strings.Join(nestedBlocks, "\n"))
}

func generateGroupAddress(number string, phoneType string, extension string) string {
	return fmt.Sprintf(`addresses {
				number = %s
				type = "%s"
                extension = %s
			}
			`, number, phoneType, extension)
}

func generateGroupOwners(userIDs ...string) string {
	return fmt.Sprintf(`owner_ids = [%s]
	`, strings.Join(userIDs, ","))
}

func generateGroupMembers(userIDs ...string) string {
	return fmt.Sprintf(`member_ids = [%s]
	`, strings.Join(userIDs, ","))
}
