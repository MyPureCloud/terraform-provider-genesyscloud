package architect_emergencygroup

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

func getAllEmergencyGroups(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	ap := getArchitectEmergencyGroupProxy(clientConfig)

	emergencyGroupConfigs, _, getErr := ap.getAllArchitectEmergencyGroups(ctx)
	if getErr != nil {
		return nil, diag.FromErr(getErr)
	}

	for _, emergencyGroupConfig := range *emergencyGroupConfigs {
		if emergencyGroupConfig.State != nil && *emergencyGroupConfig.State != "deleted" {
			resources[*emergencyGroupConfig.Id] = &resourceExporter.ResourceMeta{Name: *emergencyGroupConfig.Name}
		}
	}
	return resources, nil
}

func createEmergencyGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	divisionId := d.Get("division_id").(string)
	enabled := d.Get("enabled").(bool)

	sdkConfig := meta.(*genesyscloud.ProviderMeta).ClientConfig
	ap := getArchitectEmergencyGroupProxy(sdkConfig)

	emergencyGroup := platformclientv2.Emergencygroup{
		Name:               &name,
		Enabled:            &enabled,
		EmergencyCallFlows: buildSdkEmergencyGroupCallFlows(d),
	}

	// Optional attributes
	if description != "" {
		emergencyGroup.Description = &description
	}

	if divisionId != "" {
		emergencyGroup.Division = &platformclientv2.Writabledivision{Id: &divisionId}
	}

	log.Printf("Creating emergency group %s", name)
	eGroup, _, err := ap.createArchitectEmergencyGroup(ctx, emergencyGroup)
	if err != nil {
		return diag.Errorf("Failed to create emergency group %s: %s", name, err)
	}

	d.SetId(*eGroup.Id)

	log.Printf("Created emergency group %s %s", name, *eGroup.Id)

	return readEmergencyGroup(ctx, d, meta)
}

func readEmergencyGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*genesyscloud.ProviderMeta).ClientConfig
	ap := getArchitectEmergencyGroupProxy(sdkConfig)

	log.Printf("Reading emergency group %s", d.Id())
	return genesyscloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		emergencyGroup, resp, getErr := ap.getArchitectEmergencyGroup(ctx, d.Id())
		if getErr != nil {
			if genesyscloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read emergency group %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read emergency group %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectEmergencyGroup())

		if emergencyGroup.State != nil && *emergencyGroup.State == "deleted" {
			d.SetId("")
			return nil
		}

		d.Set("name", *emergencyGroup.Name)
		d.Set("division_id", *emergencyGroup.Division.Id)

		resourcedata.SetNillableValue(d, "description", emergencyGroup.Description)
		resourcedata.SetNillableValue(d, "enabled", emergencyGroup.Enabled)

		if emergencyGroup.EmergencyCallFlows != nil && len(*emergencyGroup.EmergencyCallFlows) > 0 {
			d.Set("emergency_call_flows", flattenEmergencyCallFlows(*emergencyGroup.EmergencyCallFlows))
		} else {
			d.Set("emergency_call_flows", nil)
		}

		log.Printf("Read emergency group %s %s", d.Id(), *emergencyGroup.Name)
		return cc.CheckState()
	})
}

func updateEmergencyGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	divisionId := d.Get("division_id").(string)
	enabled := d.Get("enabled").(bool)

	sdkConfig := meta.(*genesyscloud.ProviderMeta).ClientConfig
	ap := getArchitectEmergencyGroupProxy(sdkConfig)

	diagErr := genesyscloud.RetryWhen(genesyscloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current emergency group version
		emergencyGroup, resp, getErr := ap.getArchitectEmergencyGroup(ctx, d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read emergency group %s: %s", d.Id(), getErr)
		}

		log.Printf("Updating emergency group %s", name)
		updatedEmergencyGroup := platformclientv2.Emergencygroup{
			Name:               &name,
			Division:           &platformclientv2.Writabledivision{Id: &divisionId},
			Description:        &description,
			Version:            emergencyGroup.Version,
			State:              emergencyGroup.State,
			Enabled:            &enabled,
			EmergencyCallFlows: buildSdkEmergencyGroupCallFlows(d),
		}

		_, resp, putErr := ap.updateArchitectEmergencyGroup(ctx, d.Id(), updatedEmergencyGroup)

		if putErr != nil {
			return resp, diag.Errorf("Failed to put emergency group %s: %s", d.Id(), putErr)
		}
		return resp, nil
	})

	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished updating emergency group %s", name)
	return readEmergencyGroup(ctx, d, meta)
}

func deleteEmergencyGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*genesyscloud.ProviderMeta).ClientConfig
	ap := getArchitectEmergencyGroupProxy(sdkConfig)

	log.Printf("Deleting emergency group %s", d.Id())
	_, err := ap.deleteArchitectEmergencyGroup(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete emergency group %s: %s", d.Id(), err)
	}
	return genesyscloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		emergencyGroup, resp, err := ap.getArchitectEmergencyGroup(ctx, d.Id())
		if err != nil {
			if genesyscloud.IsStatus404(resp) {
				// group deleted
				log.Printf("Deleted emergency group %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting emergency group %s: %s", d.Id(), err))
		}

		if emergencyGroup.State != nil && *emergencyGroup.State == "deleted" {
			// group deleted
			log.Printf("Deleted emergency group %s", d.Id())
			return nil
		}

		return retry.RetryableError(fmt.Errorf("emergency group %s still exists", d.Id()))
	})
}
