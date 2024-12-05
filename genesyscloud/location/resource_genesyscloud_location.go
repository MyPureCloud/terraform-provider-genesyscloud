package location

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllLocations(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getLocationProxy(clientConfig)

	locations, resp, getErr := proxy.getAllLocation(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of locations error: %s", getErr), resp)
	}

	for _, location := range *locations {
		resources[*location.Id] = &resourceExporter.ResourceMeta{BlockLabel: *location.Name}
	}

	return resources, nil
}

func createLocation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getLocationProxy(sdkConfig)
	name := d.Get("name").(string)
	notes := d.Get("notes").(string)

	create := platformclientv2.Locationcreatedefinition{
		Name:            &name,
		Path:            buildSdkLocationPath(d),
		EmergencyNumber: buildSdkLocationEmergencyNumber(d),
		Address:         buildSdkLocationAddress(d),
	}

	if notes != "" {
		create.Notes = &notes
	}

	log.Printf("Creating location %s", name)
	location, resp, err := proxy.createLocation(ctx, &create)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create location %s error: %s", name, err), resp)
	}

	d.SetId(*location.Id)

	log.Printf("Created location %s %s", name, *location.Id)
	return readLocation(ctx, d, meta)
}

func readLocation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getLocationProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceLocation(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading location %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		location, resp, getErr := proxy.getLocationById(ctx, d.Id(), nil)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read location %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read location %s | error: %s", d.Id(), getErr), resp))
		}

		if location.State != nil && *location.State == "deleted" {
			d.SetId("")
			return nil
		}

		d.Set("name", *location.Name)
		resourcedata.SetNillableValue(d, "notes", location.Notes)
		resourcedata.SetNillableValue(d, "path", location.Path)
		d.Set("emergency_number", flattenLocationEmergencyNumber(location.EmergencyNumber))
		d.Set("address", flattenLocationAddress(location.Address))

		log.Printf("Read location %s %s", d.Id(), *location.Name)
		return cc.CheckState(d)
	})
}

func updateLocation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getLocationProxy(sdkConfig)
	name := d.Get("name").(string)
	notes := d.Get("notes").(string)

	log.Printf("Updating location %s", name)

	if diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current location version
		location, resp, getErr := proxy.getLocationById(ctx, d.Id(), nil)
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read location %s error: %s", name, getErr), resp)
		}

		update := platformclientv2.Locationupdatedefinition{
			Version:         location.Version,
			Name:            &name,
			Path:            buildSdkLocationPath(d),
			EmergencyNumber: buildSdkLocationEmergencyNumber(d),
		}

		if d.HasChange("address") {
			// Even if address is the same, the API does not allow it in the patch request if a number is assigned
			update.Address = buildSdkLocationAddress(d)
		}
		if notes != "" {
			update.Notes = &notes
		} else {
			// nil will result in no change occurring, and an empty string is invalid for this field
			filler := " "
			update.Notes = &filler
			err := d.Set("notes", filler)
			if err != nil {
				return nil, util.BuildDiagnosticError(ResourceType, "error setting the value of 'notes' attribute", err)
			}
		}

		log.Printf("Updating location %s", name)
		_, resp, putErr := proxy.updateLocation(ctx, d.Id(), &update)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update location %s error: %s", name, putErr), resp)
		}
		return resp, nil
	}); diagErr != nil {
		return diagErr
	}

	log.Printf("Updated location %s %s", name, d.Id())
	return readLocation(ctx, d, meta)
}

func deleteLocation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getLocationProxy(sdkConfig)
	name := d.Get("name").(string)

	log.Printf("Deleting location %s", name)

	if diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Directory occasionally returns version errors on deletes if an object was updated at the same time.
		resp, err := proxy.deleteLocation(ctx, d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete location %s error: %s", name, err), resp)
		}
		return nil, nil
	}); diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		location, resp, err := proxy.getLocationById(ctx, d.Id(), nil)
		if err != nil {
			if util.IsStatus404(resp) {
				// Location deleted
				log.Printf("Deleted location %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting location %s | error: %s", d.Id(), err), resp))
		}

		if location.State != nil && *location.State == "deleted" {
			// Location deleted
			log.Printf("Deleted location %s", d.Id())
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Location %s still exists", d.Id()), resp))
	})
}
