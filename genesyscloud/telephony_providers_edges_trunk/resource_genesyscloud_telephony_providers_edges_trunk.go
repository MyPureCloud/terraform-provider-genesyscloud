package telephony_providers_edges_trunk

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

func createTrunk(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	trunkBaseSettingsId := d.Get("trunk_base_settings_id").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	tp := getTrunkProxy(sdkConfig)

	trunkBase, resp, getErr := tp.getTrunkBaseSettings(ctx, trunkBaseSettingsId)
	if getErr != nil {
		if util.IsStatus404(resp) {
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read trunk base settings %s error: %s", d.Id(), getErr), resp)
	}

	// Assign to edge if edge_id is set
	if edgeIdI, ok := d.GetOk("edge_id"); ok {
		edgeId := edgeIdI.(string)
		edge, resp, getErr := tp.getEdge(ctx, edgeId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return nil
			}
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read edge %s error: %s", edgeId, getErr), resp)
		}

		if edge.EdgeGroup == nil {
			edge.EdgeGroup = &platformclientv2.Edgegroup{}
		}
		edge.EdgeGroup.EdgeTrunkBaseAssignment = &platformclientv2.Trunkbaseassignment{
			TrunkBase: trunkBase,
		}

		log.Printf("Assigning trunk base settings to edge %s", edgeId)
		_, resp, err := tp.putEdge(ctx, edgeId, *edge)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to assign trunk base settings to edge %s error: %s", edgeId, err), resp)
		}
	} else if edgeGroupIdI, ok := d.GetOk("edge_group_id"); ok {
		edgeGroupId := edgeGroupIdI.(string)
		edgeGroup, resp, getErr := tp.getEdgeGroup(ctx, edgeGroupId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get edge group %s error: %s", edgeGroupId, getErr), resp)
			}
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read edge group %s error: %s", edgeGroupId, getErr), resp)
		}
		edgeGroup.EdgeTrunkBaseAssignment = &platformclientv2.Trunkbaseassignment{
			TrunkBase: trunkBase,
		}

		log.Printf("Assigning trunk base settings to edge group %s", edgeGroupId)
		_, resp, err := tp.putEdgeGroup(ctx, edgeGroupId, *edgeGroup)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to assign trunk base settings to edge group %s error: %s", edgeGroupId, err), resp)
		}
	} else {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("edge_id or edge_group_id were not set. One must be set in order to assign the trunk base settings"), fmt.Errorf("edge_id or edge_group_id were not set"))
	}

	trunk, resp, err := getTrunkByTrunkBaseId(ctx, trunkBaseSettingsId, meta)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get trunk by trunk base id %s error: %s", trunkBaseSettingsId, err), resp)
	}

	d.SetId(*trunk.Id)

	log.Printf("Created trunk %s", *trunk.Id)

	return readTrunk(ctx, d, meta)
}

func getTrunkByTrunkBaseId(ctx context.Context, trunkBaseId string, meta interface{}) (*platformclientv2.Trunk, *platformclientv2.APIResponse, error) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	tp := getTrunkProxy(sdkConfig)
	var response *platformclientv2.APIResponse
	time.Sleep(2 * time.Second)
	// It should return the trunk as the first object. Paginating to be safe
	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		trunks, resp, getErr := tp.getAllTrunks(ctx, pageNum, pageSize)
		response = resp
		if getErr != nil {
			return nil, resp, fmt.Errorf("Failed to get page of trunks: %v %v", getErr, resp)
		}

		if trunks.Entities == nil || len(*trunks.Entities) == 0 {
			break
		}

		for _, trunk := range *trunks.Entities {
			if *trunk.TrunkBase.Id == trunkBaseId {
				return &trunk, resp, nil
			}
		}
	}

	return nil, response, fmt.Errorf("Could not find trunk for trunk base setting id: %v", trunkBaseId)
}

func updateTrunk(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return createTrunk(ctx, d, meta)
}

func readTrunk(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	tp := getTrunkProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTrunk(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading trunk %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		trunk, resp, getErr := tp.getTrunkById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read trunk %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read trunk %s | error: %s", d.Id(), getErr), resp))
		}

		d.Set("name", *trunk.Name)
		if trunk.TrunkBase != nil {
			d.Set("trunk_base_settings_id", *trunk.TrunkBase.Id)
		}
		if trunk.EdgeGroup != nil {
			d.Set("edge_group_id", *trunk.EdgeGroup.Id)
		}
		if trunk.Edge != nil {
			d.Set("edge_id", *trunk.Edge.Id)
		}

		log.Printf("Read trunk %s %s", d.Id(), *trunk.Name)

		return cc.CheckState(d)
	})
}

func deleteTrunk(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Does not delete the trunk. This resource will just no longer manage trunks.
	return nil
}

func TrunkExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllTrunks),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"trunk_base_settings_id": {RefType: "genesyscloud_telephony_providers_edges_trunkbasesettings"},
			"edge_group_id":          {RefType: "genesyscloud_telephony_providers_edges_edge_group"},
		},
		UnResolvableAttributes: map[string]*schema.Schema{
			"edge_id": ResourceTrunk().Schema["edge_id"],
		},
	}
}

func getAllTrunks(ctx context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	tp := getTrunkProxy(sdkConfig)

	err := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			trunks, resp, getErr := tp.getAllTrunks(ctx, pageNum, pageSize)
			if getErr != nil {
				if util.IsStatus404(resp) {
					return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of trunks: %v", getErr), resp))
				}
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of trunks: %v", getErr), resp))
			}

			if trunks.Entities == nil || len(*trunks.Entities) == 0 {
				break
			}
			for _, trunk := range *trunks.Entities {
				if trunk.State != nil && *trunk.State != "deleted" {
					resources[*trunk.Id] = &resourceExporter.ResourceMeta{BlockLabel: *trunk.Name}
				}
			}
		}
		return nil
	})
	return resources, err
}
