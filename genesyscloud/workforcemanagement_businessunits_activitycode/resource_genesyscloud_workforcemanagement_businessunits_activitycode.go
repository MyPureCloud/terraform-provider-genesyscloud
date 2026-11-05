package workforcemanagement_businessunits_activitycode

import (
	"context"
	"fmt"
	"log"
	"time"

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

// getAllBusinessUnitActivityCodes retrieves all of the business unit activity codes via Terraform in the Genesys Cloud and is used for the exporter.
// Default activity codes are skipped because they are created by Genesys Cloud and cannot be created or deleted.
func getAllBusinessUnitActivityCodes(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getBusinessUnitActivityCodeProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	businessUnits, resp, err := proxy.getAllBusinessUnits(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get business units: %v", err), resp)
	}

	for _, businessUnit := range *businessUnits {
		if businessUnit.Id == nil {
			continue
		}
		activityCodes, resp, err := proxy.getAllBusinessUnitActivityCodes(ctx, *businessUnit.Id)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get activity codes of business unit %s: %v", *businessUnit.Id, err), resp)
		}

		for _, activityCode := range *activityCodes {
			if activityCode.Id == nil || activityCode.Name == nil || !isActive(&activityCode) {
				continue
			}
			if activityCode.DefaultCode != nil && *activityCode.DefaultCode {
				continue
			}
			resources[*businessUnit.Id+"/"+*activityCode.Id] = &resourceExporter.ResourceMeta{BlockLabel: *businessUnit.Name + "_" + *activityCode.Name}
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
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read business unit %s activity code %s: %s", businessUnitId, activityCodeId, getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read business unit %s activity code %s: %s", businessUnitId, activityCodeId, getErr), resp))
		}

		// Deleted activity codes are retained by the api with active=false
		if !isActive(activityCode) {
			d.SetId("")
			return nil
		}

		_ = d.Set("business_unit_id", businessUnitId)
		resourcedata.SetNillableValue(d, "name", activityCode.Name)
		resourcedata.SetNillableValue(d, "category", activityCode.Category)
		resourcedata.SetNillableValue(d, "length_in_minutes", activityCode.LengthInMinutes)
		resourcedata.SetNillableValue(d, "counts_as_paid_time", activityCode.CountsAsPaidTime)
		resourcedata.SetNillableValue(d, "counts_as_work_time", activityCode.CountsAsWorkTime)
		resourcedata.SetNillableValue(d, "agent_time_off_selectable", activityCode.AgentTimeOffSelectable)
		resourcedata.SetNillableValue(d, "counts_toward_shrinkage", activityCode.CountsTowardShrinkage)
		resourcedata.SetNillableValue(d, "planned_shrinkage", activityCode.PlannedShrinkage)
		resourcedata.SetNillableValue(d, "interruptible", activityCode.Interruptible)
		resourcedata.SetNillableValue(d, "default_code", activityCode.DefaultCode)
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

	// The update request carries the current metadata version for optimistic locking, so a version conflict re-reads the activity code and tries again
	diagErr := util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
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

	if d.Get("default_code").(bool) {
		log.Printf("Unable to delete default activity code %s of business unit %s", activityCodeId, businessUnitId)
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

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		activityCode, resp, err := proxy.getBusinessUnitActivityCodeById(ctx, businessUnitId, activityCodeId)
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted business unit %s activity code %s", businessUnitId, activityCodeId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting business unit %s activity code %s: %s", businessUnitId, activityCodeId, err), resp))
		}

		if !isActive(activityCode) {
			log.Printf("Deleted business unit %s activity code %s", businessUnitId, activityCodeId)
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("business unit %s activity code %s still exists", businessUnitId, activityCodeId), resp))
	})
}
