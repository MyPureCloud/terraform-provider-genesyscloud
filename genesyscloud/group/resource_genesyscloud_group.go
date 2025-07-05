package group

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

func getAllGroups(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getGroupProxy(clientConfig)

	groups, resp, err := proxy.getAllGroups(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to retrieve all groups: %s", err), resp)
	}

	for _, group := range *groups {
		resources[*group.Id] = &resourceExporter.ResourceMeta{BlockLabel: *group.Name}
	}

	return resources, nil
}

func createGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	groupType := d.Get("type").(string)
	visibility := d.Get("visibility").(string)
	rulesVisible := d.Get("rules_visible").(bool)
	rolesEnabled := d.Get("roles_enabled").(bool)
	callsEnabled := d.Get("calls_enabled").(bool)
	includeOwners := d.Get("include_owners").(bool)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	gp := getGroupProxy(sdkConfig)

	addresses, err := buildSdkGroupAddresses(d)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Error Building SDK group addresses"), err)
	}

	groupCreate := &platformclientv2.Groupcreate{
		Name:          &name,
		VarType:       &groupType,
		Visibility:    &visibility,
		RulesVisible:  &rulesVisible,
		Addresses:     addresses,
		RolesEnabled:  &rolesEnabled,
		CallsEnabled:  &callsEnabled,
		OwnerIds:      lists.BuildSdkStringListFromInterfaceArray(d, "owner_ids"),
		IncludeOwners: &includeOwners,
	}

	log.Printf("Creating group %s", name)
	group, resp, err := gp.createGroup(ctx, groupCreate)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create group %s: %s", name, err), resp)
	}

	d.SetId(*group.Id)

	// Description can only be set in a PUT. This is a bug with the API and has been reported
	if description != "" {
		diags = append(diags, updateGroup(ctx, d, meta)...)
		if diags.HasError() {
			return diags
		}
	}

	diags = append(diags, updateGroupMembers(ctx, d, sdkConfig)...)
	if diags.HasError() {
		return diags
	}

	log.Printf("Created group %s %s", name, *group.Id)
	return append(diags, readGroup(ctx, d, meta)...)
}

func readGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceGroup(), constants.ConsistencyChecks(), ResourceType)
	gp := getGroupProxy(sdkConfig)

	retryDiags := util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		log.Printf("Reading group %s", d.Id())
		group, resp, getErr := gp.getGroupById(ctx, d.Id())
		if getErr != nil {
			diagErr := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read group %s | error: %s", d.Id(), getErr), resp)
			if util.IsStatus404(resp) {
				return retry.RetryableError(diagErr)
			}
			return retry.NonRetryableError(diagErr)
		}
		name := *group.Name
		log.Printf("Successfully read group '%s' '%s' from the API", d.Id(), name)

		resourcedata.SetNillableValue(d, "name", group.Name)
		resourcedata.SetNillableValue(d, "type", group.VarType)
		resourcedata.SetNillableValue(d, "visibility", group.Visibility)
		resourcedata.SetNillableValue(d, "rules_visible", group.RulesVisible)
		resourcedata.SetNillableValue(d, "description", group.Description)
		resourcedata.SetNillableValue(d, "roles_enabled", group.RolesEnabled)
		resourcedata.SetNillableValue(d, "calls_enabled", group.CallsEnabled)
		resourcedata.SetNillableValue(d, "include_owners", group.IncludeOwners)

		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "owner_ids", group.Owners, flattenGroupOwners)

		_ = d.Set("addresses", nil)
		if group.Addresses != nil {
			_ = d.Set("addresses", flattenGroupAddresses(d, group.Addresses))
		}

		log.Printf("Reading groups '%s' '%s' members", d.Id(), name)
		members, resp, err := readGroupMembers(ctx, d.Id(), sdkConfig)
		if err != nil {
			log.Printf("Encountered an error while reading group '%s' '%s' members: %s", d.Id(), name, err.Error())
			diagErr := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read members for group %s: %s", d.Id(), err), resp)
			return retry.NonRetryableError(diagErr)
		}

		_ = d.Set("member_ids", members)

		log.Printf("Successfully read group '%s' '%s' into the ResourceData schema", d.Id(), name)
		return cc.CheckState(d)
	})

	if retryDiags != nil {
		diags = append(diags, retryDiags...)
	}

	return diags
}

func updateGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	visibility := d.Get("visibility").(string)
	rulesVisible := d.Get("rules_visible").(bool)
	rolesEnabled := d.Get("roles_enabled").(bool)
	callsEnabled := d.Get("calls_enabled").(bool)
	includeOwners := d.Get("include_owners").(bool)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	gp := getGroupProxy(sdkConfig)

	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current group version
		log.Printf("Reading group '%s' version", d.Id())
		group, resp, getErr := gp.getGroupById(ctx, d.Id())
		if getErr != nil {
			log.Printf("Encountered an error while reading group '%s' version: %s", d.Id(), getErr.Error())
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read group %s: %s", d.Id(), getErr), resp)
		}

		addresses, err := buildSdkGroupAddresses(d)
		if err != nil {
			log.Printf("Encountered error while building group '%s' addresses", err.Error())
			return resp, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Error while trying to buildSdkGroupAddresses for group id: %s", d.Id()), err)
		}

		log.Printf("Updating group %s", name)
		groupUpdate := &platformclientv2.Groupupdate{
			Version:       group.Version,
			Name:          &name,
			Description:   &description,
			Visibility:    &visibility,
			RulesVisible:  &rulesVisible,
			Addresses:     addresses,
			RolesEnabled:  &rolesEnabled,
			CallsEnabled:  &callsEnabled,
			OwnerIds:      lists.BuildSdkStringListFromInterfaceArray(d, "owner_ids"),
			IncludeOwners: &includeOwners,
		}

		// If no owner IDs are provided, assign a list with an empty space, otherwise use the provided owner IDs
		ownerIds := lists.BuildSdkStringListFromInterfaceArray(d, "owner_ids")
		if ownerIds == nil || len(*ownerIds) == 0 {
			emptyList := []string{" "}
			ownerIds = &emptyList
		}
		groupUpdate.OwnerIds = ownerIds

		_, resp, putErr := gp.updateGroup(ctx, d.Id(), groupUpdate)
		if putErr != nil {
			log.Printf("Encountered error while updating group '%s' '%s': %s", d.Id(), name, putErr.Error())
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update group %s: %s", d.Id(), putErr), resp)
		}

		return resp, nil
	})

	diags = append(diags, diagErr...)
	if diags.HasError() {
		return diags
	}

	log.Printf("Updating group '%s' '%s' members", d.Id(), name)
	diags = append(diags, updateGroupMembers(ctx, d, sdkConfig)...)
	if diags.HasError() {
		log.Printf("Error while updating group '%s' members", d.Id())
		return diags
	}

	log.Printf("Successfully updated group '%s' '%s'", d.Id(), name)
	return append(diags, readGroup(ctx, d, meta)...)
}

func deleteGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	gp := getGroupProxy(sdkConfig)

	deleteDiags := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Directory occasionally returns version errors on deletes if an object was updated at the same time.
		log.Printf("Deleting group %s", name)
		resp, err := gp.deleteGroup(ctx, d.Id())
		if err != nil {
			log.Printf("Encountered error while deleting group '%s': %s", d.Id(), err.Error())
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete group %s: %s", name, err), resp)
		}
		return nil, nil
	})

	if deleteDiags != nil {
		log.Printf("Initial delete attempt for group '%s' failed: %v. Re-attempting in case it was a version mismatch error...", name, deleteDiags)
	}

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
		  As a result, the deletion sometimes does not occur, and then we just keep picking it up as it has.
		  After talking with Joe Fruland, the team lead for directory we are putting this extra DELETE in here
		  to keep trying to delete in case of this situation.
		*/
		resp, err = gp.deleteGroup(ctx, d.Id())
		if err != nil {
			log.Printf("Error while trying to delete group '%s' inside of the delete retry. Correlation id of failed call %s", name, resp.CorrelationID)
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Group %s still exists", d.Id()), resp))
	})
}
