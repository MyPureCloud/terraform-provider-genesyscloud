package telephony_providers_edges_edge_group

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v123/platformclientv2"
)

func createEdgeGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	managed := d.Get("managed").(bool)
	hybrid := d.Get("hybrid").(bool)

	edgeGroup := &platformclientv2.Edgegroup{
		Name:            &name,
		Managed:         &managed,
		Hybrid:          &hybrid,
		PhoneTrunkBases: buildSdkTrunkBases(d),
	}

	if description != "" {
		edgeGroup.Description = &description
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	edgeGroupProxy := getEdgeGroupProxy(sdkConfig)

	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Creating edge group %s", name)
		edgeGroup, resp, err := edgeGroupProxy.createEdgeGroup(ctx, *edgeGroup)
		if err != nil {
			return resp, diag.Errorf("Failed to create edge group %s: %s", name, err)
		}

		d.SetId(*edgeGroup.Id)
		log.Printf("Created edge group %s", *edgeGroup.Id)

		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return readEdgeGroup(ctx, d, meta)
}

func updateEdgeGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	managed := d.Get("managed").(bool)
	hybrid := d.Get("hybrid").(bool)
	id := d.Id()

	edgeGroup := &platformclientv2.Edgegroup{
		Id:              &id,
		Name:            &name,
		Managed:         &managed,
		Hybrid:          &hybrid,
		PhoneTrunkBases: buildSdkTrunkBases(d),
	}

	if description != "" {
		edgeGroup.Description = &description
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	edgeGroupProxy := getEdgeGroupProxy(sdkConfig)

	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		edgeGroupFromApi, resp, getErr := edgeGroupProxy.getEdgeGroupById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return resp, diag.Errorf("The edge group does not exist %s: %s", d.Id(), getErr)
			}
			return resp, diag.Errorf("Failed to read edge group %s: %s", d.Id(), getErr)
		}
		edgeGroup.Version = edgeGroupFromApi.Version

		log.Printf("Updating edge group %s", name)
		_, resp, putErr := edgeGroupProxy.updateEdgeGroup(ctx, d.Id(), *edgeGroup)
		if putErr != nil {
			return resp, diag.Errorf("Failed to update edge group %s: %s", name, putErr)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated edge group %s", *edgeGroup.Id)
	return readEdgeGroup(ctx, d, meta)
}

func deleteEdgeGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	edgeGroupProxy := getEdgeGroupProxy(sdkConfig)

	log.Printf("Deleting edge group")
	_, err := edgeGroupProxy.deleteEdgeGroup(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete edge group: %s", err)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		edgeGroup, resp, err := edgeGroupProxy.getEdgeGroupById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Edge group deleted
				log.Printf("Deleted Edge group %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting Edge group %s: %s", d.Id(), err))
		}

		if edgeGroup.State != nil && *edgeGroup.State == "deleted" {
			// Edge group deleted
			log.Printf("Deleted Edge group %s", d.Id())
			return nil
		}

		return retry.RetryableError(fmt.Errorf("Edge group %s still exists", d.Id()))
	})
}

func readEdgeGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	edgeGroupProxy := getEdgeGroupProxy(sdkConfig)

	log.Printf("Reading edge group %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		edgeGroup, resp, getErr := edgeGroupProxy.getEdgeGroupById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read edge group %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read edge group %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceEdgeGroup())
		d.Set("name", *edgeGroup.Name)
		d.Set("state", *edgeGroup.State)
		if edgeGroup.Description != nil {
			d.Set("description", *edgeGroup.Description)
		}
		if edgeGroup.Managed != nil {
			d.Set("managed", *edgeGroup.Managed)
		}
		if edgeGroup.Hybrid != nil {
			d.Set("hybrid", *edgeGroup.Hybrid)
		}
		d.Set("phone_trunk_base_ids", nil)
		if edgeGroup.PhoneTrunkBases != nil {
			d.Set("phone_trunk_base_ids", flattenPhoneTrunkBases(*edgeGroup.PhoneTrunkBases))
		}

		log.Printf("Read edge group %s %s", d.Id(), *edgeGroup.Name)

		return cc.CheckState()
	})
}

func getAllEdgeGroups(ctx context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	edgeGroupProxy := getEdgeGroupProxy(sdkConfig)
	edgeGroups, err := edgeGroupProxy.getAllEdgeGroups(ctx, "", false)

	if err != nil {
		return nil, diag.Errorf("Failed to get Edge groups: %s", err)
	}
	if edgeGroups != nil {
		for _, edgeGroup := range *edgeGroups {
			resources[*edgeGroup.Id] = &resourceExporter.ResourceMeta{Name: *edgeGroup.Name}
		}
	}

	return resources, nil
}
