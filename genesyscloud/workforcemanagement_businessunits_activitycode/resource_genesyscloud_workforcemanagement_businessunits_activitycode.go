package workforcemanagement_businessunits_activitycode

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

/*
	NOTE: This resource's Id is in the format <businessUnitId>/<activityCodeId> so we can persist the id of the parent business unit.
	The business_unit_id field can not be used for this because the activity code api does not return the business unit.
*/

/*
The resource_genesyscloud_workforcemanagement_businessunits_activitycode.go contains all of the methods that perform the core logic for a resource.
*/

// getAllBusinessUnitActivityCodes retrieves all of the business unit activity codes, active and inactive, via Terraform in the Genesys Cloud and is used for the exporter.
// Inactive codes are included because old schedules and work plans that have not been updated can still refer to them.
// Default activity codes are included because they can be updated, even though they cannot be created or deleted.
func getAllBusinessUnitActivityCodes(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getBusinessUnitActivityCodeProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	businessUnits, resp, err := proxy.getAllBusinessUnits(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get business units: %v", err), resp)
	}

	for _, businessUnit := range *businessUnits {
		activityCodes, resp, err := proxy.getAllBusinessUnitActivityCodes(ctx, *businessUnit.Id)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get activity codes of business unit %s: %v", *businessUnit.Id, err), resp)
		}

		for _, activityCode := range *activityCodes {
			blockHash, err := activityCodeBlockHash(*businessUnit.Name, activityCode)
			if err != nil {
				return nil, diag.Errorf("failed to generate a block hash for activity code %s: %v", *activityCode.Id, err)
			}
			resources[*businessUnit.Id+"/"+*activityCode.Id] = &resourceExporter.ResourceMeta{BlockLabel: *businessUnit.Name + "_" + activityCodeLabel(activityCode), BlockHash: blockHash}
		}
	}

	return resources, nil
}

// createBusinessUnitActivityCode is used by the workforcemanagement_businessunits_activitycode resource to create a Genesys Cloud business unit activity code
func createBusinessUnitActivityCode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessUnitActivityCodeProxy(sdkConfig)
	businessUnitId := d.Get("business_unit_id").(string)
	activityCode := buildCreateActivityCodeRequest(d)

	log.Printf("Creating business unit %s activity code %s", businessUnitId, *activityCode.Name)
	created, resp, err := proxy.createBusinessUnitActivityCode(ctx, businessUnitId, activityCode)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create business unit %s activity code %s: %s", businessUnitId, *activityCode.Name, err), resp)
	}

	d.SetId(businessUnitId + "/" + *created.Id)
	log.Printf("Created business unit %s activity code %s %s", businessUnitId, *created.Id, *activityCode.Name)
	return readBusinessUnitActivityCode(ctx, d, meta)
}

// readBusinessUnitActivityCode is used by the workforcemanagement_businessunits_activitycode resource to read a business unit activity code from Genesys Cloud
func readBusinessUnitActivityCode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessUnitActivityCodeProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceBusinessUnitActivityCode(), constants.ConsistencyChecks(), ResourceType)
	businessUnitId, activityCodeId := SplitBusinessUnitActivityCodeTerraformId(d.Id())

	log.Printf("Reading business unit %s activity code %s", businessUnitId, activityCodeId)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		activityCode, resp, getErr := proxy.getBusinessUnitActivityCodeById(ctx, businessUnitId, activityCodeId)
		if getErr != nil {
			// A 404 is not retried: the activity code is readable as soon as it is created, so a 404 means it is gone and WithRetriesForRead removes it from state
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read business unit %s activity code %s: %s", businessUnitId, activityCodeId, getErr), resp))
		}

		_ = d.Set("business_unit_id", businessUnitId)
		_ = d.Set("active", *activityCode.Active)
		_ = d.Set("name", *activityCode.Name)
		_ = d.Set("category", *activityCode.Category)
		_ = d.Set("length_in_minutes", *activityCode.LengthInMinutes)
		_ = d.Set("counts_as_paid_time", *activityCode.CountsAsPaidTime)
		_ = d.Set("counts_as_work_time", *activityCode.CountsAsWorkTime)
		resourcedata.SetNillableValue(d, "agent_time_off_selectable", activityCode.AgentTimeOffSelectable)
		_ = d.Set("counts_toward_shrinkage", *activityCode.CountsTowardShrinkage)
		_ = d.Set("planned_shrinkage", *activityCode.PlannedShrinkage)
		_ = d.Set("interruptible", *activityCode.Interruptible)
		_ = d.Set("default_code", *activityCode.DefaultCode)
		_ = d.Set("secondary_presence_ids", flattenSecondaryPresenceIds(activityCode.SecondaryPresences))

		log.Printf("Read business unit %s activity code %s", businessUnitId, activityCodeId)
		return cc.CheckState(d)
	})
}

// updateBusinessUnitActivityCode is used by the workforcemanagement_businessunits_activitycode resource to update a business unit activity code in Genesys Cloud
func updateBusinessUnitActivityCode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessUnitActivityCodeProxy(sdkConfig)
	businessUnitId, activityCodeId := SplitBusinessUnitActivityCodeTerraformId(d.Id())

	log.Printf("Updating business unit %s activity code %s", businessUnitId, activityCodeId)

	// The update request carries the current metadata version for optimistic locking, so a version conflict re-reads the activity code and tries again.
	diagErr := util.WithRetries(ctx, util.UpdateRetryTimeout, func() *retry.RetryError {
		current, resp, err := proxy.getBusinessUnitActivityCodeById(ctx, businessUnitId, activityCodeId)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read business unit %s activity code %s before update: %s", businessUnitId, activityCodeId, err), resp))
		}

		_, resp, err = proxy.updateBusinessUnitActivityCode(ctx, businessUnitId, activityCodeId, buildUpdateActivityCodeRequest(d, current.Metadata))
		if err == nil {
			return nil
		}
		builtDiagErr := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to update business unit %s activity code %s: %s", businessUnitId, activityCodeId, err), resp)
		if util.IsStatus409(resp) {
			return retry.RetryableError(builtDiagErr)
		}
		return retry.NonRetryableError(builtDiagErr)
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated business unit %s activity code %s", businessUnitId, activityCodeId)
	return readBusinessUnitActivityCode(ctx, d, meta)
}

// deleteBusinessUnitActivityCode is used by the workforcemanagement_businessunits_activitycode resource to delete a business unit activity code from Genesys Cloud
func deleteBusinessUnitActivityCode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessUnitActivityCodeProxy(sdkConfig)
	businessUnitId, activityCodeId := SplitBusinessUnitActivityCodeTerraformId(d.Id())

	// Genesys Cloud does not allow a default activity code to be deleted, so destroying one only removes it from Terraform.
	// An activity code that is already inactive is deleted again, because Genesys Cloud allows it and runs the deactivation again.
	if d.Get("default_code").(bool) {
		log.Printf("Business unit %s activity code %s is a default activity code, so it is only removed from Terraform", businessUnitId, activityCodeId)
		return nil
	}

	log.Printf("Deleting business unit %s activity code %s", businessUnitId, activityCodeId)
	resp, err := proxy.deleteBusinessUnitActivityCode(ctx, businessUnitId, activityCodeId)
	if err != nil {
		if util.IsStatus404(resp) {
			log.Printf("Business unit %s activity code %s already deleted", businessUnitId, activityCodeId)
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete business unit %s activity code %s: %s", businessUnitId, activityCodeId, err), resp)
	}

	log.Printf("Deleted business unit %s activity code %s", businessUnitId, activityCodeId)
	return nil
}
