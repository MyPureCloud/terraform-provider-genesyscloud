package group

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/chunks"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllGroups(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	groupProxy := getGroupProxy(clientConfig)

	groups, resp, err := groupProxy.getAllGroups(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to retrieve all groups: %s", err), resp)
	}

	for _, group := range *groups {
		resources[*group.Id] = &resourceExporter.ResourceMeta{BlockLabel: *group.Name}
	}

	return resources, nil
}

func createGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	groupType := d.Get("type").(string)
	visibility := d.Get("visibility").(string)
	rulesVisible := d.Get("rules_visible").(bool)
	rolesEnabled := d.Get("roles_enabled").(bool)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	gp := getGroupProxy(sdkConfig)

	addresses, err := buildSdkGroupAddresses(d)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Error Building SDK group addresses"), err)
	}

	createGroup := &platformclientv2.Groupcreate{
		Name:         &name,
		VarType:      &groupType,
		Visibility:   &visibility,
		RulesVisible: &rulesVisible,
		Addresses:    addresses,
		RolesEnabled: &rolesEnabled,
		OwnerIds:     lists.BuildSdkStringListFromInterfaceArray(d, "owner_ids"),
	}
	log.Printf("Creating group %s", name)
	group, resp, err := gp.createGroup(ctx, createGroup)

	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create group %s: %s", name, err), resp)
	}

	d.SetId(*group.Id)

	// Description can only be set in a PUT. This is a bug with the API and has been reported
	if description != "" {
		diagErr := updateGroup(ctx, d, meta)
		if diagErr != nil {
			return diagErr
		}
	}

	diagErr := updateGroupMembers(ctx, d, sdkConfig)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Created group %s %s", name, *group.Id)
	return readGroup(ctx, d, meta)
}

func readGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceGroup(), constants.ConsistencyChecks(), ResourceType)
	gp := getGroupProxy(sdkConfig)

	log.Printf("Reading group %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {

		group, resp, getErr := gp.getGroupById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read group %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read group %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", group.Name)
		resourcedata.SetNillableValue(d, "type", group.VarType)
		resourcedata.SetNillableValue(d, "visibility", group.Visibility)
		resourcedata.SetNillableValue(d, "rules_visible", group.RulesVisible)
		resourcedata.SetNillableValue(d, "description", group.Description)
		resourcedata.SetNillableValue(d, "roles_enabled", group.RolesEnabled)

		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "owner_ids", group.Owners, flattenGroupOwners)

		if group.Addresses != nil {
			_ = d.Set("addresses", flattenGroupAddresses(d, group.Addresses))
		} else {
			_ = d.Set("addresses", nil)
		}

		members, err := readGroupMembers(ctx, d.Id(), sdkConfig)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("%v", err))
		}
		_ = d.Set("member_ids", members)

		log.Printf("Read group %s %s", d.Id(), *group.Name)
		return cc.CheckState(d)
	})
}

func updateGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	visibility := d.Get("visibility").(string)
	rulesVisible := d.Get("rules_visible").(bool)
	rolesEnabled := d.Get("roles_enabled").(bool)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	gp := getGroupProxy(sdkConfig)

	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current group version
		group, resp, getErr := gp.getGroupById(ctx, d.Id())
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read group %s: %s", d.Id(), getErr), resp)
		}

		addresses, err := buildSdkGroupAddresses(d)
		if err != nil {
			return resp, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Error while trying to buildSdkGroupAddresses for group id: %s", d.Id()), err)
		}

		log.Printf("Updating group %s", name)
		updateGroup := &platformclientv2.Groupupdate{
			Version:      group.Version,
			Name:         &name,
			Description:  &description,
			Visibility:   &visibility,
			RulesVisible: &rulesVisible,
			Addresses:    addresses,
			RolesEnabled: &rolesEnabled,
			OwnerIds:     lists.BuildSdkStringListFromInterfaceArray(d, "owner_ids"),
		}
		_, resp, putErr := gp.updateGroup(ctx, d.Id(), updateGroup)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update group %s: %s", d.Id(), putErr), resp)
		}

		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateGroupMembers(ctx, d, sdkConfig)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated group %s", name)
	return readGroup(ctx, d, meta)
}

func deleteGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	gp := getGroupProxy(sdkConfig)

	util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Directory occasionally returns version errors on deletes if an object was updated at the same time.
		log.Printf("Deleting group %s", name)
		resp, err := gp.deleteGroup(ctx, d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete group %s: %s", name, err), resp)
		}
		return nil, nil
	})

	return util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		group, resp, err := gp.getGroupById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Group %s deleted", name)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting group %s | error: %s", d.Id(), err), resp))
		}

		if group.State != nil && *group.State == "deleted" {
			log.Printf("Group %s deleted", name)
			return nil
		}

		/*
		  This extra delete call is being added here because of  DEVTOOLING-485.  Basically we are in a transition
		  state with the groups API.  We have two services BEVY and Directory that are managing groups.  Bevy is dual
		  writing to directory.  However, Directory always returns a 200 on the delete and then fails asynchronously.
		  As a result the delete sometimes does not occur and then we just keep picking it up as it has.
		  After talking with Joe Fruland, the team lead for directory we are putting this extra DELETE in here
		  to keep trying to delete in case of this situation.
		*/
		resp, err = gp.deleteGroup(ctx, d.Id())
		if err != nil {
			log.Printf("Error while trying to delete group %s inside of the delete retry.  Correlation id of failed call %s",
				name, resp.CorrelationID)
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Group %s still exists", d.Id()), resp))
	})
}

func updateGroupMembers(ctx context.Context, d *schema.ResourceData, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	gp := getGroupProxy(sdkConfig)
	if d.HasChange("member_ids") {
		if membersConfig := d.Get("member_ids"); membersConfig != nil {
			configMemberIds := *lists.SetToStringList(membersConfig.(*schema.Set))
			existingMemberIds, err := getGroupMemberIds(ctx, d, sdkConfig)
			if err != nil {
				return err
			}

			maxMembersPerRequest := 50
			membersToRemoveList := lists.SliceDifference(existingMemberIds, configMemberIds)
			chunkedMemberIdsDelete := chunks.ChunkBy(membersToRemoveList, maxMembersPerRequest)

			chunkProcessor := func(membersToRemove []string) diag.Diagnostics {
				if len(membersToRemove) > 0 {
					if diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
						_, resp, err := gp.deleteGroupMembers(ctx, d.Id(), strings.Join(membersToRemove, ","))
						if err != nil {
							return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to remove members from group %s: %s", d.Id(), err), resp)
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

func readGroupMembers(ctx context.Context, groupID string, sdkConfig *platformclientv2.Configuration) (*schema.Set, diag.Diagnostics) {
	gp := getGroupProxy(sdkConfig)
	members, resp, err := gp.getGroupMembers(ctx, groupID)

	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read members for group %s: %s", groupID, err), resp)
	}

	interfaceList := make([]interface{}, len(*members))
	for i, v := range *members {
		interfaceList[i] = v
	}
	return schema.NewSet(schema.HashString, interfaceList), nil
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
	if diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Need the current group version to add members
		groupInfo, resp, getErr := gp.getGroupById(ctx, d.Id())
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read group %s: %s", d.Id(), getErr), resp)
		}

		groupMemberUpdate := &platformclientv2.Groupmembersupdate{
			MemberIds: &membersToAdd,
			Version:   groupInfo.Version,
		}
		_, resp, postErr := gp.addGroupMembers(ctx, d.Id(), groupMemberUpdate)
		if postErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to add group members %s: %s", d.Id(), postErr), resp)
		}
		return resp, nil
	}); diagErr != nil {
		return diagErr
	}

	return nil
}
