package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
	"log"
	"time"
)

func resourceTrunk() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Trunk. Created by assigning a trunk base settings to an edge or edge group",

		CreateContext: createWithPooledClient(createTrunk),
		ReadContext:   readWithPooledClient(readTrunk),
		UpdateContext: updateWithPooledClient(updateTrunk),
		DeleteContext: deleteWithPooledClient(deleteTrunk),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"trunk_base_settings_id": {
				Description: "The trunk base settings reference",
				Type:        schema.TypeString,
				Required:    true,
			},
			"edge_group_id": {
				Description: "The edge group associated with this trunk. Either this or \"edge_id\" must be set",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"edge_id": {
				Description: "The edge associated with this trunk. Either this or \"edge_group_id\" must be set",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "The name of the trunk. This property is read only and populated with the auto generated name.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:	true,
			},
		},
	}
}

func createTrunk(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	trunkBaseSettingsId := d.Get("trunk_base_settings_id").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	trunkBase, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunkbasesetting(trunkBaseSettingsId, true)
	if getErr != nil {
		if isStatus404(resp) {
			return nil
		}
		return diag.Errorf("Failed to read trunk base settings %s: %s", d.Id(), getErr)
	}

	// Assign to edge if edge_id is set
	var trunkBaseId string
	if edgeIdI, ok := d.GetOk("edge_id"); ok {
		edgeId := edgeIdI.(string)
		edge, resp, getErr := edgesAPI.GetTelephonyProvidersEdge(edgeId, nil)
		if getErr != nil {
			if isStatus404(resp) {
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
			if isStatus404(resp) {
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

	trunk, err := getTrunkByTrunkBaseId(trunkBaseId, meta)
	if err != nil {
		return diag.Errorf("Failed to get trunk by trunk base id %s: %s", trunkBaseId, err)
	}

	d.SetId(*trunk.Id)

	log.Printf("Created trunk %s", *trunk.Id)

	time.Sleep(5 * time.Second)
	return readTrunk(ctx, d, meta)
}

func getTrunkByTrunkBaseId(trunkBaseId string, meta interface{}) (*platformclientv2.Trunk, error) {
	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	// It should return the trunk as the first object. Paginating to be safe
	for pageNum := 1; ; pageNum++ {
		trunks, _, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunks(pageNum, 100, "", "", "", trunkBaseId, "")
		if getErr != nil {
			return nil, fmt.Errorf("Failed to get page of edge groups: %v", getErr)
		}

		if trunks.Entities == nil || len(*trunks.Entities) == 0 {
			break
		}

		for _, trunk := range *trunks.Entities {
			if *trunk.Id != trunkBaseId {
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
	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Reading trunk %s", d.Id())
	return withRetriesForRead(ctx, 30*time.Second, d, func() *resource.RetryError {
		trunk, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunk(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read trunk %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read trunk %s: %s", d.Id(), getErr))
		}

		d.Set("name", *trunk.Name)

		log.Printf("Read trunk %s %s", d.Id(), *trunk.Name)

		return nil
	})
}

func deleteTrunk(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Does not delete the trunk. This resource will just no longer manage trunks.
	return nil
}

func trunkExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllTrunks),
		RefAttrs: map[string]*RefAttrSettings{
			"trunk_base_settings_id": {RefType: "genesyscloud_telephony_providers_edges_trunkbasesettings"},
			"edge_group_id":          {RefType: "genesyscloud_telephony_providers_edges_edge_group"},
		},
	}
}

func getAllTrunks(ctx context.Context, sdkConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)

	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		trunks, _, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunks(pageNum, 100, "", "", "", "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of trunks: %v", getErr)
		}

		if trunks.Entities == nil || len(*trunks.Entities) == 0 {
			break
		}

		for _, trunk := range *trunks.Entities {
			if *trunk.State != "deleted" {
				resources[*trunk.Id] = &ResourceMeta{Name: *trunk.Name}
			}
		}
	}

	return resources, nil
}
