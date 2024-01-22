package genesyscloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

type MediaUtilization struct {
	MaximumCapacity         int32    `json:"maximumCapacity"`
	InterruptableMediaTypes []string `json:"interruptableMediaTypes"`
	IncludeNonAcd           bool     `json:"includeNonAcd"`
}

type LabelUtilization struct {
	MaximumCapacity      int32    `json:"maximumCapacity"`
	InterruptingLabelIds []string `json:"interruptingLabelIds"`
}

type OrgUtilizationWithLabels struct {
	Utilization       map[string]MediaUtilization `json:"utilization"`
	LabelUtilizations map[string]LabelUtilization `json:"labelUtilizations"`
}

var (
	// Map of SDK media type name to schema media type name
	utilizationMediaTypes = map[string]string{
		"call":     "call",
		"callback": "callback",
		"chat":     "chat",
		"email":    "email",
		"message":  "message",
	}

	utilizationSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"maximum_capacity": {
				Description:  "Maximum capacity of conversations of this media type. Value must be between 0 and 25.",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 25),
			},
			"interruptible_media_types": {
				Description: fmt.Sprintf("Set of other media types that can interrupt this media type (%s).", strings.Join(getSdkUtilizationTypes(), " | ")),
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"include_non_acd": {
				Description: "Block this media type when on a non-ACD conversation.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}

	utilizationLabelResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"label_id": {
				Description: "Id of the label being configured.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"maximum_capacity": {
				Description:  "Maximum capacity of conversations with this label. Value must be between 0 and 25.",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 25),
			},
			"interrupting_label_ids": {
				Description: "Set of other labels that can interrupt this label.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
)

func getSdkUtilizationTypes() []string {
	types := make([]string, 0, len(utilizationMediaTypes))
	for t := range utilizationMediaTypes {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

func getAllRoutingUtilization(_ context.Context, _ *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Routing utilization config always exists
	resources := make(resourceExporter.ResourceIDMetaMap)
	resources["0"] = &resourceExporter.ResourceMeta{Name: "routing_utilization"}
	return resources, nil
}

func RoutingUtilizationExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllRoutingUtilization),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
		AllowZeroValues:  []string{"maximum_capacity"},
	}
}

func ResourceRoutingUtilization() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Org-wide Routing Utilization Settings.",

		CreateContext: CreateWithPooledClient(createRoutingUtilization),
		ReadContext:   ReadWithPooledClient(readRoutingUtilization),
		UpdateContext: UpdateWithPooledClient(updateRoutingUtilization),
		DeleteContext: DeleteWithPooledClient(deleteRoutingUtilization),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(8 * time.Minute),
			Read:   schema.DefaultTimeout(8 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"call": {
				Description: "Call media settings. If not set, this reverts to the default media type settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        utilizationSettingsResource,
			},
			"callback": {
				Description: "Callback media settings. If not set, this reverts to the default media type settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        utilizationSettingsResource,
			},
			"message": {
				Description: "Message media settings. If not set, this reverts to the default media type settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        utilizationSettingsResource,
			},
			"email": {
				Description: "Email media settings. If not set, this reverts to the default media type settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        utilizationSettingsResource,
			},
			"chat": {
				Description: "Chat media settings. If not set, this reverts to the default media type settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        utilizationSettingsResource,
			},
			"label_utilizations": {
				Description: "Label utilization settings. If not set, default label settings will be applied. This is in PREVIEW and should not be used unless the feature is available to your organization.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem:        utilizationLabelResource,
			},
		},
	}
}

func createRoutingUtilization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating Routing Utilization")
	d.SetId("routing_utilization")
	return updateRoutingUtilization(ctx, d, meta)
}

func readRoutingUtilization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Calling the Utilization API directly while the label feature is not available.
	// Once it is, this code can go back to using platformclientv2's RoutingApi to make the call.
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	apiClient := &routingAPI.Configuration.APIClient

	path := fmt.Sprintf("%s/api/v2/routing/utilization", routingAPI.Configuration.BasePath)
	headerParams := buildHeaderParams(routingAPI)

	log.Printf("Reading Routing Utilization")
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		response, err := apiClient.CallAPI(path, "GET", nil, headerParams, nil, nil, "", nil)
		if err != nil {
			if IsStatus404(response) {
				return retry.RetryableError(fmt.Errorf("Failed to read Routing Utilization: %s", err))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Routing Utilization: %s", err))
		}

		orgUtilization := &OrgUtilizationWithLabels{}
		err = json.Unmarshal(response.RawBody, &orgUtilization)

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingSkill())

		if orgUtilization.Utilization != nil {
			for sdkType, schemaType := range utilizationMediaTypes {
				if mediaSettings, ok := orgUtilization.Utilization[sdkType]; ok {
					d.Set(schemaType, flattenUtilizationSetting(mediaSettings))
				} else {
					d.Set(schemaType, nil)
				}
			}
		}

		if orgUtilization.LabelUtilizations != nil {
			originalLabelUtilizations := d.Get("label_utilizations").([]interface{})

			// Only add to the state the configured labels, in the configured order, but not any extras, to help terraform with matching new and old state.
			flattenedLabelUtilizations := filterAndFlattenLabelUtilizations(orgUtilization.LabelUtilizations, originalLabelUtilizations)
			d.Set("label_utilizations", flattenedLabelUtilizations)
		}

		log.Printf("Read Routing Utilization")
		return cc.CheckState()
	})
}

func updateRoutingUtilization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	var resp *platformclientv2.APIResponse
	var err error

	log.Printf("Updating Routing Utilization")

	labelUtilizations := d.Get("label_utilizations").([]interface{})

	// Retrying on 409s because if a label is created immediately before the utilization update, it can lead to a conflict while the utilization is being updated to handle the new label.
	diagErr := RetryWhen(IsStatus409, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// If the resource has label(s), calls the Utilization API directly.
		// This code can go back to using platformclientv2's RoutingApi to make the call once label utilization is available in platformclientv2's RoutingApi.
		if labelUtilizations != nil && len(labelUtilizations) > 0 {
			apiClient := &routingAPI.Configuration.APIClient

			path := fmt.Sprintf("%s/api/v2/routing/utilization", routingAPI.Configuration.BasePath)
			headerParams := buildHeaderParams(routingAPI)
			requestPayload := make(map[string]interface{})
			requestPayload["utilization"] = buildSdkMediaUtilizations(d)
			requestPayload["labelUtilizations"] = buildLabelUtilizationsRequest(labelUtilizations)
			resp, err = apiClient.CallAPI(path, "PUT", requestPayload, headerParams, nil, nil, "", nil)
		} else {
			_, resp, err = routingAPI.PutRoutingUtilization(platformclientv2.Utilizationrequest{
				Utilization: buildSdkMediaUtilizations(d),
			})
		}

		if err != nil {
			return resp, diag.Errorf("Failed to update Routing Utilization: %s", err)
		}
		return resp, nil
	})

	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Routing Utilization")
	return readRoutingUtilization(ctx, d, meta)
}

func deleteRoutingUtilization(_ context.Context, _ *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	// Resets to default values
	log.Printf("Resetting Routing Utilization")
	_, err := routingAPI.DeleteRoutingUtilization()
	if err != nil {
		return diag.Errorf("Failed to reset Routing Utilization: %s", err)
	}
	log.Printf("Reset Routing Utilization")
	return nil
}

func flattenUtilizationSetting(settings MediaUtilization) []interface{} {
	settingsMap := make(map[string]interface{})

	settingsMap["maximum_capacity"] = settings.MaximumCapacity
	settingsMap["include_non_acd"] = settings.IncludeNonAcd
	if settings.InterruptableMediaTypes != nil {
		settingsMap["interruptible_media_types"] = lists.StringListToSet(settings.InterruptableMediaTypes)
	}

	return []interface{}{settingsMap}
}

func filterAndFlattenLabelUtilizations(labelUtilizations map[string]LabelUtilization, originalLabelUtilizations []interface{}) []interface{} {
	flattenedLabelUtilizations := make([]interface{}, 0)

	for _, originalLabelUtilization := range originalLabelUtilizations {
		originalLabelId := (originalLabelUtilization.(map[string]interface{}))["label_id"].(string)

		for currentLabelId, currentLabelUtilization := range labelUtilizations {
			if currentLabelId == originalLabelId {
				flattenedLabelUtilizations = append(flattenedLabelUtilizations, flattenLabelUtilization(currentLabelId, currentLabelUtilization))
				delete(labelUtilizations, currentLabelId)
				break
			}
		}
	}

	return flattenedLabelUtilizations
}

func flattenLabelUtilization(labelId string, labelUtilization LabelUtilization) map[string]interface{} {
	utilizationMap := make(map[string]interface{})

	utilizationMap["label_id"] = labelId
	utilizationMap["maximum_capacity"] = labelUtilization.MaximumCapacity
	if labelUtilization.InterruptingLabelIds != nil {
		utilizationMap["interrupting_label_ids"] = lists.StringListToSet(labelUtilization.InterruptingLabelIds)
	}

	return utilizationMap
}

func buildSdkMediaUtilizations(d *schema.ResourceData) *map[string]platformclientv2.Mediautilization {
	settings := make(map[string]platformclientv2.Mediautilization)

	for sdkType, schemaType := range utilizationMediaTypes {
		mediaSettings := d.Get(schemaType).([]interface{})
		if mediaSettings != nil && len(mediaSettings) > 0 {
			settings[sdkType] = buildSdkMediaUtilization(mediaSettings)
		}
	}

	return &settings
}

func buildSdkMediaUtilization(settings []interface{}) platformclientv2.Mediautilization {
	settingsMap := settings[0].(map[string]interface{})

	maxCapacity := settingsMap["maximum_capacity"].(int)
	includeNonAcd := settingsMap["include_non_acd"].(bool)

	// Optional
	interruptableMediaTypes := &[]string{}
	if types, ok := settingsMap["interruptible_media_types"]; ok {
		interruptableMediaTypes = lists.SetToStringList(types.(*schema.Set))
	}

	return platformclientv2.Mediautilization{
		MaximumCapacity:         &maxCapacity,
		IncludeNonAcd:           &includeNonAcd,
		InterruptableMediaTypes: interruptableMediaTypes,
	}
}

func buildLabelUtilizationsRequest(labelUtilizations []interface{}) map[string]LabelUtilization {
	request := make(map[string]LabelUtilization)
	for _, labelUtilization := range labelUtilizations {
		labelUtilizationMap := labelUtilization.(map[string]interface{})
		interruptingLabelIds := lists.SetToStringList(labelUtilizationMap["interrupting_label_ids"].(*schema.Set))

		request[labelUtilizationMap["label_id"].(string)] = LabelUtilization{
			MaximumCapacity:      int32(labelUtilizationMap["maximum_capacity"].(int)),
			InterruptingLabelIds: *interruptingLabelIds,
		}
	}
	return request
}
