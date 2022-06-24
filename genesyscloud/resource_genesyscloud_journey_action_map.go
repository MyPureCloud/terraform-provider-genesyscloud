package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v74/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
)

var (
	journeyActionMapSchema = map[string]*schema.Schema{
		"is_active": {
			Description: "Whether the action map is active.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"display_name": {
			Description: "Display name of the action map.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"trigger_with_segments": {
			Description: "Trigger action map if any segment in the list is assigned to a given customer.",
			Type:        schema.TypeSet,
			Required:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		// TODO
	}
)

func getAllJourneyActionMaps(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	journeyApi := platformclientv2.NewJourneyApiWithConfig(clientConfig)

	pageCount := 1 // Needed because of broken journey common paging
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		const pageSize = 100
		actionMaps, _, getErr := journeyApi.GetJourneyActionmaps(pageNum, pageSize, "", "", "", nil, nil, "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of journey action maps: %v", getErr)
		}

		if actionMaps.Entities == nil || len(*actionMaps.Entities) == 0 {
			break
		}

		for _, actionMap := range *actionMaps.Entities {
			resources[*actionMap.Id] = &ResourceMeta{Name: *actionMap.DisplayName}
		}

		pageCount = *actionMaps.PageCount
	}

	return resources, nil
}

func journeyActionMapExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllJourneyActionMaps),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceJourneyActionMap() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Journey Action Map",
		CreateContext: createWithPooledClient(createJourneyActionMap),
		ReadContext:   readWithPooledClient(readJourneyActionMap),
		UpdateContext: updateWithPooledClient(updateJourneyActionMap),
		DeleteContext: deleteWithPooledClient(deleteJourneyActionMap),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema:        journeyActionMapSchema,
	}
}

func createJourneyActionMap(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	actionMap := buildSdkActionMap(d)

	log.Printf("Creating journey action map %s", *actionMap.DisplayName)
	result, resp, err := journeyApi.PostJourneyActionmaps(*actionMap)
	if err != nil {
		return diag.Errorf("failed to create journey action map %s: %s\n(input: %+v)\n(resp: %s)", *actionMap.DisplayName, err, *actionMap, resp.RawBody)
	}

	d.SetId(*result.Id)

	log.Printf("Created journey action map %s %s", *result.DisplayName, *result.Id)
	return readJourneyActionMap(ctx, d, meta)
}

func readJourneyActionMap(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	log.Printf("Reading journey action map %s", d.Id())
	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		actionMap, resp, getErr := journeyApi.GetJourneyActionmap(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("failed to read journey action map %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("failed to read journey action map %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceJourneyActionMap())
		flattenActionMap(d, actionMap)

		log.Printf("Read journey action map  %s %s", d.Id(), *actionMap.DisplayName)
		return cc.CheckState()
	})
}

func updateJourneyActionMap(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	actionMap := buildSdkPatchActionMap(d)

	log.Printf("Updating journey action map %s", d.Id())
	if _, resp, err := journeyApi.PatchJourneyActionmap(d.Id(), *actionMap); err != nil {
		return diag.Errorf("Error updating journey action map %s: %s\n(input: %+v)\n(resp: %s)", *actionMap.DisplayName, err, *actionMap, resp.RawBody)
	}

	log.Printf("Updated journey action map %s", d.Id())
	return readJourneyActionMap(ctx, d, meta)
}

func deleteJourneyActionMap(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	displayName := d.Get("display_name").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	log.Printf("Deleting journey action map with display name %s", displayName)
	if _, err := journeyApi.DeleteJourneyActionmap(d.Id()); err != nil {
		return diag.Errorf("Failed to delete journey action map with display name %s: %s", displayName, err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := journeyApi.GetJourneyActionmap(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// journey action map deleted
				log.Printf("Deleted journey action map %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error deleting journey action map %s: %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("journey action map %s still exists", d.Id()))
	})
}

func flattenActionMap(d *schema.ResourceData, actionMap *platformclientv2.Actionmap) {
	d.Set("is_active", *actionMap.IsActive)
	d.Set("display_name", *actionMap.DisplayName)
	d.Set("trigger_with_segments", stringListToSet(*actionMap.TriggerWithSegments))
	// TODO
}

func buildSdkActionMap(actionMap *schema.ResourceData) *platformclientv2.Actionmap {
	isActive := getNullableBool(actionMap, "is_active")
	displayName := getNullableValue[string](actionMap, "display_name")
	triggerWithSegments := buildSdkStringList(actionMap, "trigger_with_segments")
	// TODO

	return &platformclientv2.Actionmap{
		IsActive:            isActive,
		DisplayName:         displayName,
		TriggerWithSegments: triggerWithSegments,
		// TODO
	}
}

func buildSdkPatchActionMap(actionMap *schema.ResourceData) *platformclientv2.Patchactionmap {
	isActive := getNullableBool(actionMap, "is_active")
	displayName := getNullableValue[string](actionMap, "display_name")

	return &platformclientv2.Patchactionmap{
		IsActive:    isActive,
		DisplayName: displayName,
		// TODO
	}
}
