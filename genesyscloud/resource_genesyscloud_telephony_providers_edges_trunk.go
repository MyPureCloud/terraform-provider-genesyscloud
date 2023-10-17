package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func ResourceTrunk() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Trunk. Created by assigning a trunk base settings to an edge or edge group",

		CreateContext: CreateWithPooledClient(createTrunk),
		ReadContext:   ReadWithPooledClient(readTrunk),
		UpdateContext: UpdateWithPooledClient(updateTrunk),
		DeleteContext: DeleteWithPooledClient(deleteTrunk),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"trunk_base_settings_id": {
				Description: "The trunk base settings reference",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"edge_group_id": {
				Description: "The edge group associated with this trunk. Either this or \"edge_id\" must be set",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"edge_id": {
				Description: "The edge associated with this trunk. Either this or \"edge_group_id\" must be set",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"name": {
				Description: "The name of the trunk. This property is read only and populated with the auto generated name.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func createTrunk(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	trunkBaseSettingsId := d.Get("trunk_base_settings_id").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	trunkBase, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunkbasesetting(trunkBaseSettingsId, true)
	if getErr != nil {
		if IsStatus404(resp) {
			return nil
		}
		return diag.Errorf("Failed to read trunk base settings %s: %s", d.Id(), getErr)
	}

	// Assign to edge if edge_id is set
	if edgeIdI, ok := d.GetOk("edge_id"); ok {
		edgeId := edgeIdI.(string)
		edge, resp, getErr := edgesAPI.GetTelephonyProvidersEdge(edgeId, nil)
		if getErr != nil {
			if IsStatus404(resp) {
				return nil
			}
			return diag.Errorf("Failed to read edge %s: %s", edgeId, getErr)
		}

		if edge.EdgeGroup == nil {
			edge.EdgeGroup = &platformclientv2.Edgegroup{}
		}
		edge.EdgeGroup.EdgeTrunkBaseAssignment = &platformclientv2.Trunkbaseassignment{
			TrunkBase: trunkBase,
		}

		log.Printf("Assigning trunk base settings to edge %s", edgeId)
		_, _, err := edgesAPI.PutTelephonyProvidersEdge(edgeId, *edge)
		if err != nil {
			return diag.Errorf("Failed to assign trunk base settings to edge %s: %s", edgeId, err)
		}
	} else if edgeGroupIdI, ok := d.GetOk("edge_group_id"); ok {
		edgeGroupId := edgeGroupIdI.(string)
		edgeGroup, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesEdgegroup(edgeGroupId, nil)
		if getErr != nil {
			if IsStatus404(resp) {
				return diag.Errorf("Failed to get edge group %s: %s", edgeGroupId, getErr)
			}
			return diag.Errorf("Failed to read edge group %s: %s", edgeGroupId, getErr)
		}
		edgeGroup.EdgeTrunkBaseAssignment = &platformclientv2.Trunkbaseassignment{
			TrunkBase: trunkBase,
		}

		log.Printf("Assigning trunk base settings to edge group %s", edgeGroupId)
		_, _, err := edgesAPI.PutTelephonyProvidersEdgesEdgegroup(edgeGroupId, *edgeGroup)
		if err != nil {
			return diag.Errorf("Failed to assign trunk base settings to edge group %s: %s", edgeGroupId, err)
		}
	} else {
		return diag.Errorf("edge_id or edge_group_id were not set. One must be set in order to assign the trunk base settings")
	}

	trunk, err := getTrunkByTrunkBaseId(trunkBaseSettingsId, meta)
	if err != nil {
		return diag.Errorf("Failed to get trunk by trunk base id %s: %s", trunkBaseSettingsId, err)
	}

	d.SetId(*trunk.Id)

	log.Printf("Created trunk %s", *trunk.Id)

	return readTrunk(ctx, d, meta)
}

func getTrunkByTrunkBaseId(trunkBaseId string, meta interface{}) (*platformclientv2.Trunk, error) {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	time.Sleep(2 * time.Second)
	// It should return the trunk as the first object. Paginating to be safe
	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		trunks, _, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunks(pageNum, pageSize, "", "", "", "", "")
		if getErr != nil {
			return nil, fmt.Errorf("Failed to get page of trunks: %v", getErr)
		}

		if trunks.Entities == nil || len(*trunks.Entities) == 0 {
			break
		}

		for _, trunk := range *trunks.Entities {
			if *trunk.TrunkBase.Id == trunkBaseId {
				return &trunk, nil
			}
		}
	}

	return nil, fmt.Errorf("Could not find trunk for trunk base setting id: %v", trunkBaseId)
}

func updateTrunk(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return createTrunk(ctx, d, meta)
}

func readTrunk(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Reading trunk %s", d.Id())
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		trunk, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunk(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read trunk %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read trunk %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTrunk())
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

		return cc.CheckState()
	})
}

func deleteTrunk(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Does not delete the trunk. This resource will just no longer manage trunks.
	return nil
}

func TrunkExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllTrunks),
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

	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	err := WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			trunks, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunks(pageNum, pageSize, "", "", "", "", "")
			if getErr != nil {
				if IsStatus404(resp) {
					return retry.RetryableError(fmt.Errorf("Failed to get page of trunks: %v", getErr))
				}
				return retry.NonRetryableError(fmt.Errorf("Failed to get page of trunks: %v", getErr))
			}

			if trunks.Entities == nil || len(*trunks.Entities) == 0 {
				break
			}

			for _, trunk := range *trunks.Entities {
				if trunk.State != nil && *trunk.State != "deleted" {
					resources[*trunk.Id] = &resourceExporter.ResourceMeta{Name: *trunk.Name}
				}
			}
		}

		return nil
	})

	return resources, err
}
