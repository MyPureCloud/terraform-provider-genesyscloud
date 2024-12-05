package architect_ivr

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

// getAllIvrConfigs retrieves all architect IVRs and is used for the exporter
func getAllIvrConfigs(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	ap := getArchitectIvrProxy(clientConfig)

	allIvrs, resp, err := ap.getAllArchitectIvrs(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get archictect IVRs error: %s", err), resp)
	}

	for _, entity := range *allIvrs {
		resources[*entity.Id] = &resourceExporter.ResourceMeta{BlockLabel: *entity.Name}
	}
	return resources, nil
}

// createIvrConfig is used by the resource to create a Genesys Cloud Architect IVR
func createIvrConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ap := getArchitectIvrProxy(sdkConfig)

	ivrBody := buildArchitectIvrFromResourceData(d)

	// It might need to wait for a dependent did_pool to be created to avoid an eventual consistency issue which
	// would result in the error "Field 'didPoolId' is required and cannot be empty."
	if ivrBody.Dnis != nil {
		time.Sleep(3 * time.Second)
	}

	log.Printf("Creating IVR config %s", *ivrBody.Name)
	ivrConfig, resp, err := ap.createArchitectIvr(ctx, *ivrBody)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create IVR config %s error: %s", *ivrBody.Name, err), resp)
	}

	d.SetId(*ivrConfig.Id)

	log.Printf("Created IVR config %s %s", *ivrBody.Name, *ivrConfig.Id)
	return readIvrConfig(ctx, d, meta)
}

// readIvrConfig is used by the resource to read a Genesys Cloud Architect IVR
func readIvrConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ap := getArchitectIvrProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectIvrConfig(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading IVR config %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		ivrConfig, resp, getErr := ap.getArchitectIvr(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read IVR config %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read IVR config %s | error: %s", d.Id(), getErr), resp))
		}

		if ivrConfig.State != nil && *ivrConfig.State == "deleted" {
			d.SetId("")
			return nil
		}

		_ = d.Set("name", *ivrConfig.Name)
		if ivrConfig.Dnis == nil || *ivrConfig.Dnis == nil {
			_ = d.Set("dnis", nil)
		} else {
			utilE164 := util.NewUtilE164Service()
			dnis := lists.Map(*ivrConfig.Dnis, utilE164.FormatAsCalculatedE164Number)
			_ = d.Set("dnis", lists.StringListToSetOrNil(&dnis))
		}

		resourcedata.SetNillableValue(d, "description", ivrConfig.Description)
		resourcedata.SetNillableReference(d, "open_hours_flow_id", ivrConfig.OpenHoursFlow)
		resourcedata.SetNillableReference(d, "closed_hours_flow_id", ivrConfig.ClosedHoursFlow)
		resourcedata.SetNillableReference(d, "holiday_hours_flow_id", ivrConfig.HolidayHoursFlow)
		resourcedata.SetNillableReference(d, "schedule_group_id", ivrConfig.ScheduleGroup)
		resourcedata.SetNillableReferenceWritableDivision(d, "division_id", ivrConfig.Division)

		log.Printf("Read IVR config %s %s", d.Id(), *ivrConfig.Name)

		return cc.CheckState(d)
	})
}

// updateIvrConfig is used by the resource to update a Genesys Cloud Architect IVR
func updateIvrConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ap := getArchitectIvrProxy(sdkConfig)

	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current version
		ivr, resp, getErr := ap.getArchitectIvr(ctx, d.Id())
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read IVR config %s error: %s", d.Id(), getErr), resp)
		}

		ivrBody := buildArchitectIvrFromResourceData(d)
		ivrBody.Version = ivr.Version

		// It might need to wait for a dependent did_pool to be created to avoid an eventual consistency issue which
		// would result in the error "Field 'didPoolId' is required and cannot be empty."
		if ivrBody.Dnis != nil {
			time.Sleep(3 * time.Second)
		}
		log.Printf("Updating IVR config %s", *ivrBody.Name)
		_, resp, putErr := ap.updateArchitectIvr(ctx, d.Id(), *ivrBody)

		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update IVR config %s error: %s", d.Id(), putErr), resp)
		}

		return resp, nil
	})

	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated IVR config %s", d.Id())
	return readIvrConfig(ctx, d, meta)
}

// deleteIvrConfig is used by the resource to delete a Genesys Cloud Architect IVR
func deleteIvrConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ap := getArchitectIvrProxy(sdkConfig)

	log.Printf("Deleting IVR config %s", name)
	if resp, err := ap.deleteArchitectIvr(ctx, d.Id()); err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete IVR config %s error: %s", name, err), resp)

	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		ivr, resp, err := ap.getArchitectIvr(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// IVR config deleted
				log.Printf("Deleted IVR config %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting IVR config %s | error: %s", d.Id(), err), resp))
		}

		if ivr.State != nil && *ivr.State == "deleted" {
			// IVR config deleted
			log.Printf("Deleted IVR config with a deleted state %s", d.Id())
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("IVR config %s still exists", d.Id()), resp))
	})
}
