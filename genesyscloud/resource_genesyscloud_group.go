package genesyscloud

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/platformclientv2"
	"github.com/nyaruka/phonenumbers"
)

var (
	groupPhoneType       = "PHONE"
	groupAddressResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"number": {
				Description:      "Phone number for this contact type.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validatePhoneNumber,
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

func getAllGroups(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDNameMap, diag.Diagnostics) {
	resources := make(map[string]string)
	groupsAPI := platformclientv2.NewGroupsApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		groups, _, getErr := groupsAPI.GetGroups(100, pageNum, nil, nil, "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of groups: %v", getErr)
		}

		if groups.Entities == nil || len(*groups.Entities) == 0 {
			break
		}

		for _, group := range *groups.Entities {
			resources[*group.Id] = *group.Name
		}
	}

	return resources, nil
}

func groupExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllGroups),
		RefAttrs: map[string]*RefAttrSettings{
			"owner_ids":  {RefType: "genesyscloud_user"},
			"member_ids": {RefType: "genesyscloud_user"},
		},
	}
}

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Directory Group",

		CreateContext: createWithPooledClient(createGroup),
		ReadContext:   readWithPooledClient(readGroup),
		UpdateContext: updateWithPooledClient(updateGroup),
		DeleteContext: deleteWithPooledClient(deleteGroup),
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
				Description:  "Group type (official | social). This cannot be modified.",
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
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        groupAddressResource,
				Set:         groupAddressHash,
			},
			"owner_ids": {
				Description: "IDs of owners of the group.",
				Type:        schema.TypeSet,
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

	sdkConfig := meta.(*providerMeta).ClientConfig
	groupsAPI := platformclientv2.NewGroupsApiWithConfig(sdkConfig)

	log.Printf("Creating group %s", name)
	group, _, err := groupsAPI.PostGroups(platformclientv2.Groupcreate{
		Name:         &name,
		Description:  &description,
		VarType:      &groupType,
		Visibility:   &visibility,
		RulesVisible: &rulesVisible,
		Addresses:    buildSdkGroupAddresses(d),
		OwnerIds:     buildSdkGroupOwners(d),
	})
	if err != nil {
		return diag.Errorf("Failed to create group %s: %s", name, err)
	}

	d.SetId(*group.Id)

	diagErr := updateGroupMembers(d, groupsAPI)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Created group %s %s", name, *group.Id)
	return readGroup(ctx, d, meta)
}

func readGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	groupsAPI := platformclientv2.NewGroupsApiWithConfig(sdkConfig)

	log.Printf("Reading group %s", d.Id())

	group, resp, getErr := groupsAPI.GetGroup(d.Id())
	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read group %s: %s", d.Id(), getErr)
	}

	d.Set("name", *group.Name)
	d.Set("type", *group.VarType)
	d.Set("visibility", *group.Visibility)
	d.Set("rules_visible", *group.RulesVisible)

	if group.Description != nil {
		d.Set("description", *group.Description)
	} else {
		d.Set("description", nil)
	}

	if group.Addresses != nil {
		d.Set("addresses", flattenGroupAddresses(*group.Addresses))
	} else {
		d.Set("addresses", nil)
	}

	if group.Owners != nil {
		d.Set("owner_ids", flattenGroupOwners(*group.Owners))
	} else {
		d.Set("owner_ids", nil)
	}

	members, err := readGroupMembers(d.Id(), groupsAPI)
	if err != nil {
		return err
	}
	d.Set("member_ids", members)

	log.Printf("Read group %s %s", d.Id(), *group.Name)
	return nil
}

func updateGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	visibility := d.Get("visibility").(string)
	rulesVisible := d.Get("rules_visible").(bool)

	sdkConfig := meta.(*providerMeta).ClientConfig
	groupsAPI := platformclientv2.NewGroupsApiWithConfig(sdkConfig)

	diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current group version
		group, resp, getErr := groupsAPI.GetGroup(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read group %s: %s", d.Id(), getErr)
		}

		log.Printf("Updating group %s", name)
		_, resp, putErr := groupsAPI.PutGroup(d.Id(), platformclientv2.Groupupdate{
			Version:      group.Version,
			Name:         &name,
			Description:  &description,
			Visibility:   &visibility,
			RulesVisible: &rulesVisible,
			Addresses:    buildSdkGroupAddresses(d),
			OwnerIds:     buildSdkGroupOwners(d),
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

	sdkConfig := meta.(*providerMeta).ClientConfig
	groupsAPI := platformclientv2.NewGroupsApiWithConfig(sdkConfig)

	retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Directory occasionally returns version errors on deletes if an object was updated at the same time.
		log.Printf("Deleting group %s", name)
		resp, err := groupsAPI.DeleteGroup(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete group %s: %s", name, err)
		}
		log.Printf("Deleted group %s", name)
		return nil, nil
	})
	return nil
}

func groupAddressHash(val interface{}) int {
	// Copy map to avoid modifying state
	phoneMap := make(map[string]interface{})
	for k, v := range val.(map[string]interface{}) {
		phoneMap[k] = v
	}
	if num, ok := phoneMap["number"]; ok {
		// Attempt to format phone numbers before hashing
		number, err := phonenumbers.Parse(num.(string), "US")
		if err == nil {
			phoneMap["number"] = phonenumbers.Format(number, phonenumbers.E164)
		}
	}
	return schema.HashResource(groupAddressResource)(phoneMap)
}

func buildSdkGroupAddresses(d *schema.ResourceData) *[]platformclientv2.Groupcontact {
	addresses := d.Get("addresses").(*schema.Set)
	if addresses != nil {
		addressSlice := addresses.List()
		sdkContacts := make([]platformclientv2.Groupcontact, len(addressSlice))
		for i, configPhone := range addressSlice {
			phoneMap := configPhone.(map[string]interface{})
			phoneType := phoneMap["type"].(string)
			contact := platformclientv2.Groupcontact{
				VarType:   &phoneType,
				MediaType: &groupPhoneType, // Only option is PHONE
			}

			if phoneNum, ok := phoneMap["number"].(string); ok {
				contact.Address = &phoneNum
			}
			if phoneExt, ok := phoneMap["extension"].(string); ok {
				contact.Extension = &phoneExt
			}

			sdkContacts[i] = contact
		}
		return &sdkContacts
	}
	return nil
}

func buildSdkGroupOwners(d *schema.ResourceData) *[]string {
	if permConfig, ok := d.GetOk("owner_ids"); ok {
		return setToStringList(permConfig.(*schema.Set))
	}
	return nil
}

func flattenGroupAddresses(addresses []platformclientv2.Groupcontact) *schema.Set {
	addressSet := schema.NewSet(groupAddressHash, []interface{}{})
	for _, address := range addresses {
		if address.MediaType != nil {
			if *address.MediaType == groupPhoneType {
				phoneNumber := make(map[string]interface{})

				// Strip off any parentheses from phone numbers
				if address.Address != nil {
					phoneNumber["number"] = strings.Trim(*address.Address, "()")
				} else if address.Display != nil {
					// Some numbers are only returned in Display
					phoneNumber["number"] = strings.Trim(*address.Display, "()")
				}

				if address.Extension != nil {
					phoneNumber["extension"] = *address.Extension
				}

				if address.VarType != nil {
					phoneNumber["type"] = *address.VarType
				}
				addressSet.Add(phoneNumber)
			} else {
				log.Printf("Unknown address media type %s", *address.MediaType)
			}
		}
	}
	return addressSet
}

func flattenGroupOwners(owners []platformclientv2.User) *schema.Set {
	interfaceList := make([]interface{}, len(owners))
	for i, v := range owners {
		interfaceList[i] = *v.Id
	}
	return schema.NewSet(schema.HashString, interfaceList)
}

func updateGroupMembers(d *schema.ResourceData, groupsAPI *platformclientv2.GroupsApi) diag.Diagnostics {
	if d.HasChange("member_ids") {
		if membersConfig := d.Get("member_ids"); membersConfig != nil {
			// Get existing members
			members, _, err := groupsAPI.GetGroupIndividuals(d.Id())
			if err != nil {
				return diag.FromErr(err)
			}

			var existingMembers []string
			if members.Entities != nil {
				for _, member := range *members.Entities {
					existingMembers = append(existingMembers, *member.Id)
				}
			}
			configMembers := *setToStringList(membersConfig.(*schema.Set))

			membersToRemove := sliceDifference(existingMembers, configMembers)
			if len(membersToRemove) > 0 {
				_, _, err := groupsAPI.DeleteGroupMembers(d.Id(), strings.Join(membersToRemove, ","))
				if err != nil {
					return diag.Errorf("Failed to remove members from group %s: %s", d.Id(), err)
				}
			}

			membersToAdd := sliceDifference(configMembers, existingMembers)
			if len(membersToAdd) > 0 {
				if diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
					// Need the current group version to add members
					groupInfo, _, getErr := groupsAPI.GetGroup(d.Id())
					if getErr != nil {
						return nil, diag.Errorf("Failed to read group %s: %s", d.Id(), getErr)
					}

					_, resp, postErr := groupsAPI.PostGroupMembers(d.Id(), platformclientv2.Groupmembersupdate{
						MemberIds: &membersToAdd,
						Version:   groupInfo.Version,
					})
					if err != nil {
						return resp, diag.Errorf("Failed to read group %s: %s", d.Id(), postErr)
					}
					return resp, nil
				}); diagErr != nil {
					return diagErr
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
