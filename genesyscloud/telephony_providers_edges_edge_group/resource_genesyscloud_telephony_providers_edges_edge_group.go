package telephony_providers_edges_edge_group

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
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
		edgeGroupResponse, resp, err := edgeGroupProxy.createEdgeGroup(ctx, *edgeGroup)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create edge group %s error: %s", name, err), resp)
		}

		d.SetId(*edgeGroupResponse.Id)
		log.Printf("Created edge group %s", *edgeGroupResponse.Id)

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
				return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("The edge group does not exist %s error: %s", d.Id(), getErr), resp)
			}
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read edge group %s error: %s", d.Id(), getErr), resp)
		}
		edgeGroup.Version = edgeGroupFromApi.Version

		log.Printf("Updating edge group %s", name)
		_, resp, putErr := edgeGroupProxy.updateEdgeGroup(ctx, d.Id(), *edgeGroup)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update edge group %s error: %s", name, putErr), resp)
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
	resp, err := edgeGroupProxy.deleteEdgeGroup(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete edge group %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		edgeGroup, resp, err := edgeGroupProxy.getEdgeGroupById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Edge group deleted
				log.Printf("Deleted Edge group %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting Edge group %s | error: %s", d.Id(), err), resp))
		}

		if edgeGroup.State != nil && *edgeGroup.State == "deleted" {
			// Edge group deleted
			log.Printf("Deleted Edge group %s", d.Id())
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Edge group %s still exists", d.Id()), resp))
	})
}

func readEdgeGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	edgeGroupProxy := getEdgeGroupProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceEdgeGroup(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading edge group %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		edgeGroup, resp, getErr := edgeGroupProxy.getEdgeGroupById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read edge group %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read edge group %s | error: %s", d.Id(), getErr), resp))
		}

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

		return cc.CheckState(d)
	})
}

func getAllEdgeGroups(ctx context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	edgeGroupProxy := getEdgeGroupProxy(sdkConfig)
	edgeGroups, resp, err := edgeGroupProxy.getAllEdgeGroups(ctx, "", false)

	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get edge groups error: %s", err), resp)
	}
	if edgeGroups != nil {
		for _, edgeGroup := range *edgeGroups {
			resources[*edgeGroup.Id] = &resourceExporter.ResourceMeta{BlockLabel: *edgeGroup.Name}
		}
	}
	return resources, nil
}
