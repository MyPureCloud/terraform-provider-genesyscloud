package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func ResourceEdgeGroup() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Edge Group. NOTE: This resource is being kept here for backwards compatibility with older Genesys Cloud Organization. You may get an error if you try to create an edge group with a Genesys Cloud Organization created in 2022 or later.`,

		CreateContext: CreateWithPooledClient(createEdgeGroup),
		ReadContext:   ReadWithPooledClient(readEdgeGroup),
		UpdateContext: UpdateWithPooledClient(updateEdgeGroup),
		DeleteContext: DeleteWithPooledClient(deleteEdgeGroup),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the entity.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The resource's description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"state": {
				Description: "Indicates if the resource is active, inactive, or deleted.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"managed": {
				Description: "Is this edge group being managed remotely.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"hybrid": {
				Description: "Is this edge group hybrid.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"phone_trunk_base_ids": {
				Description: "A list of trunk base settings IDs of trunkType \"PHONE\" to inherit to edge logical interface for phone communication.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

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

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	diagErr := RetryWhen(IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Creating edge group %s", name)
		edgeGroup, resp, err := edgesAPI.PostTelephonyProvidersEdgesEdgegroups(*edgeGroup)
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

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		edgeGroupFromApi, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesEdgegroup(d.Id(), nil)
		if getErr != nil {
			if IsStatus404(resp) {
				return resp, diag.Errorf("The edge group does not exist %s: %s", d.Id(), getErr)
			}
			return resp, diag.Errorf("Failed to read edge group %s: %s", d.Id(), getErr)
		}
		edgeGroup.Version = edgeGroupFromApi.Version

		log.Printf("Updating edge group %s", name)
		_, resp, putErr := edgesAPI.PutTelephonyProvidersEdgesEdgegroup(d.Id(), *edgeGroup)
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
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Deleting edge group")
	_, err := edgesAPI.DeleteTelephonyProvidersEdgesEdgegroup(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete edge group: %s", err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		edgeGroup, resp, err := edgesAPI.GetTelephonyProvidersEdgesEdgegroup(d.Id(), nil)
		if err != nil {
			if IsStatus404(resp) {
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
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Reading edge group %s", d.Id())
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		edgeGroup, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesEdgegroup(d.Id(), nil)
		if getErr != nil {
			if IsStatus404(resp) {
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

func flattenPhoneTrunkBases(trunkBases []platformclientv2.Trunkbase) *schema.Set {
	interfaceList := make([]interface{}, len(trunkBases))
	for i, v := range trunkBases {
		interfaceList[i] = *v.Id
	}
	return schema.NewSet(schema.HashString, interfaceList)
}

func getAllEdgeGroups(_ context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)

	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		edgeGroups, _, getErr := edgesAPI.GetTelephonyProvidersEdgesEdgegroups(pageSize, pageNum, "", "", false)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of edge groups: %v", getErr)
		}

		if edgeGroups.Entities == nil || len(*edgeGroups.Entities) == 0 {
			break
		}

		for _, edgeGroup := range *edgeGroups.Entities {
			if edgeGroup.State != nil && *edgeGroup.State != "deleted" {
				resources[*edgeGroup.Id] = &resourceExporter.ResourceMeta{Name: *edgeGroup.Name}
			}
		}
	}

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		edgeGroups, _, getErr := edgesAPI.GetTelephonyProvidersEdgesEdgegroups(pageSize, pageNum, "", "", true)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of edge groups: %v", getErr)
		}

		if edgeGroups.Entities == nil || len(*edgeGroups.Entities) == 0 {
			break
		}

		for _, edgeGroup := range *edgeGroups.Entities {
			if edgeGroup.State != nil && *edgeGroup.State != "deleted" {
				resources[*edgeGroup.Id] = &resourceExporter.ResourceMeta{Name: *edgeGroup.Name}
			}
		}
	}

	return resources, nil
}

func EdgeGroupExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllEdgeGroups),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"phone_trunk_base_ids": {RefType: "genesyscloud_telephony_providers_edges_trunkbasesettings"},
		},
	}
}

func buildSdkTrunkBases(d *schema.ResourceData) *[]platformclientv2.Trunkbase {
	returnValue := make([]platformclientv2.Trunkbase, 0)

	if ids, ok := d.GetOk("phone_trunk_base_ids"); ok {
		phoneTrunkBaseIds := lists.SetToStringList(ids.(*schema.Set))
		for _, trunkBaseId := range *phoneTrunkBaseIds {
			id := trunkBaseId
			returnValue = append(returnValue, platformclientv2.Trunkbase{
				Id: &id,
			})
		}
	}

	return &returnValue
}
