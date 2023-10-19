package architect_ivr

import (
	"context"
	"fmt"
	"log"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

// getAllIvrConfigs retrieves all architect IVRs and is used for the exporter
func getAllIvrConfigs(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	ap := getArchitectIvrProxy(clientConfig)

	allIvrs, err := ap.getAllArchitectIvrs(ctx, "")
	if err != nil {
		return nil, diag.Errorf("failed to get architect ivrs: %v", err)
	}

	for _, entity := range *allIvrs {
		resources[*entity.Id] = &resourceExporter.ResourceMeta{Name: *entity.Name}
	}
	return resources, nil
}

// createIvrConfig is used by the resource to create a Genesys Cloud Architect IVR
func createIvrConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	ap := getArchitectIvrProxy(sdkConfig)

	ivrBody := buildArchitectIvrFromResourceData(d)

	// It might need to wait for a dependent did_pool to be created to avoid an eventual consistency issue which
	// would result in the error "Field 'didPoolId' is required and cannot be empty."
	if ivrBody.Dnis != nil {
		time.Sleep(3 * time.Second)
	}
	log.Printf("Creating IVR config %s", *ivrBody.Name)
	ivrConfig, _, err := ap.createArchitectIvr(ctx, *ivrBody)
	if err != nil {
		return diag.Errorf("Failed to create IVR config %s: %s", *ivrBody.Name, err)
	}
	d.SetId(*ivrConfig.Id)

	log.Printf("Created IVR config %s %s", *ivrBody.Name, *ivrConfig.Id)
	return readIvrConfig(ctx, d, meta)
}

// readIvrConfig is used by the resource to read a Genesys Cloud Architect IVR
func readIvrConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	ap := getArchitectIvrProxy(sdkConfig)

	log.Printf("Reading IVR config %s", d.Id())
	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		ivrConfig, resp, getErr := ap.getArchitectIvr(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read IVR config %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read IVR config %s: %s", d.Id(), getErr))
		}

		if ivrConfig.State != nil && *ivrConfig.State == "deleted" {
			d.SetId("")
			return nil
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectIvrConfig())
		_ = d.Set("name", *ivrConfig.Name)
		_ = d.Set("dnis", lists.StringListToSetOrNil(ivrConfig.Dnis))

		resourcedata.SetNillableValue(d, "description", ivrConfig.Description)
		resourcedata.SetNillableReference(d, "open_hours_flow_id", ivrConfig.OpenHoursFlow)
		resourcedata.SetNillableReference(d, "closed_hours_flow_id", ivrConfig.ClosedHoursFlow)
		resourcedata.SetNillableReference(d, "holiday_hours_flow_id", ivrConfig.HolidayHoursFlow)
		resourcedata.SetNillableReference(d, "schedule_group_id", ivrConfig.ScheduleGroup)
		resourcedata.SetNillableReferenceWritableDivision(d, "division_id", ivrConfig.Division)

		log.Printf("Read IVR config %s %s", d.Id(), *ivrConfig.Name)
		return cc.CheckState()
	})
}

// updateIvrConfig is used by the resource to update a Genesys Cloud Architect IVR
func updateIvrConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	ap := getArchitectIvrProxy(sdkConfig)

	diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current version
		ivr, resp, getErr := ap.getArchitectIvr(ctx, d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read IVR config %s: %s", d.Id(), getErr)
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
			return resp, diag.Errorf("Failed to update IVR config %s: %s", d.Id(), putErr)
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

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	ap := getArchitectIvrProxy(sdkConfig)

	log.Printf("Deleting IVR config %s", name)
	if _, err := ap.deleteArchitectIvr(ctx, d.Id()); err != nil {
		return diag.Errorf("Failed to delete IVR config %s: %s", name, err)
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		ivr, resp, err := ap.getArchitectIvr(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// IVR config deleted
				log.Printf("Deleted IVR config %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting IVR config %s: %s", d.Id(), err))
		}

		if ivr.State != nil && *ivr.State == "deleted" {
			// IVR config deleted
			log.Printf("Deleted IVR config %s", d.Id())
			return nil
		}

		return retry.RetryableError(fmt.Errorf("IVR config %s still exists", d.Id()))
	})
}
