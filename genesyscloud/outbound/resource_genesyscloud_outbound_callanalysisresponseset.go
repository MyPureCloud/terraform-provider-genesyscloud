package outbound

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

var (
	outboundCallAnalysisResponseSetReaction = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`data`: {
				Description: `Parameter for this reaction. For transfer_flow, this would be the outbound flow id.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`name`: {
				Description: `Name of the parameter for this reaction. For transfer_flow, this would be the outbound flow name.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`reaction_type`: {
				Description:  `The reaction to take for a given call analysis result.`,
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{`hangup`, `transfer`, `transfer_flow`, `play_file`}, false),
			},
		},
	}

	outboundCallAnalysisResponseSetResponses = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`callable_lineconnected`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     outboundCallAnalysisResponseSetReaction,
			},
			`callable_person`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     outboundCallAnalysisResponseSetReaction,
			},
			`callable_busy`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     outboundCallAnalysisResponseSetReaction,
			},
			`callable_noanswer`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     outboundCallAnalysisResponseSetReaction,
			},
			`callable_fax`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     outboundCallAnalysisResponseSetReaction,
			},
			`callable_disconnect`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     outboundCallAnalysisResponseSetReaction,
			},
			`callable_machine`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     outboundCallAnalysisResponseSetReaction,
			},
			`callable_sit`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     outboundCallAnalysisResponseSetReaction,
			},
			`uncallable_sit`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     outboundCallAnalysisResponseSetReaction,
			},
			`uncallable_notfound`: {
				Computed: true,
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     outboundCallAnalysisResponseSetReaction,
			},
		},
	}
)

func getAllCallAnalysisResponseSets(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		responseSetConfigs, _, getErr := outboundAPI.GetOutboundCallanalysisresponsesets(pageSize, pageNum, true, "", "", "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of call analysis response set configs: %v", getErr)
		}
		if responseSetConfigs.Entities == nil || len(*responseSetConfigs.Entities) == 0 {
			break
		}
		for _, responseSetConfig := range *responseSetConfigs.Entities {
			resources[*responseSetConfig.Id] = &resourceExporter.ResourceMeta{Name: *responseSetConfig.Name}
		}
	}

	return resources, nil
}

func OutboundCallAnalysisResponseSetExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllCallAnalysisResponseSets),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"responses.callable_person.data":  {RefType: "genesyscloud_flow"},
			"responses.callable_machine.data": {RefType: "genesyscloud_flow"},
		},
	}
}

func ResourceOutboundCallAnalysisResponseSet() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud outbound Call Analysis Response Set`,

		CreateContext: gcloud.CreateWithPooledClient(createOutboundCallAnalysisResponseSet),
		ReadContext:   gcloud.ReadWithPooledClient(readOutboundCallAnalysisResponseSet),
		UpdateContext: gcloud.UpdateWithPooledClient(updateOutboundCallAnalysisResponseSet),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteOutboundCallAnalysisResponseSet),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the Response Set.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`responses`: {
				Description: `List of maps of disposition identifiers to reactions. Required if beep_detection_enabled = true.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeList,
				Elem:        outboundCallAnalysisResponseSetResponses,
			},
			`beep_detection_enabled`: {
				Description: `Whether to enable answering machine beep detection`,
				Optional:    true,
				Default:     false,
				Type:        schema.TypeBool,
			},
		},
	}
}

func createOutboundCallAnalysisResponseSet(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	responses := d.Get("responses").([]interface{})
	beepDetectionEnabled := d.Get("beep_detection_enabled").(bool)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkResponseSet := platformclientv2.Responseset{
		BeepDetectionEnabled: &beepDetectionEnabled,
	}

	if name != "" {
		sdkResponseSet.Name = &name
	}
	if responses != nil && len(responses) > 0 {
		sdkResponseSet.Responses = buildSdkOutboundCallAnalysisResponseSetReaction(responses)
	}

	log.Printf("Creating Outbound Call Analysis Response Set %s", name)
	outboundCallanalysisresponseset, _, err := outboundApi.PostOutboundCallanalysisresponsesets(sdkResponseSet)
	if err != nil {
		return diag.Errorf("Failed to create Outbound Call Analysis Response Set %s: %s", name, err)
	}

	d.SetId(*outboundCallanalysisresponseset.Id)

	log.Printf("Created Outbound Call Analysis Response Set %s %s", name, *outboundCallanalysisresponseset.Id)
	return readOutboundCallAnalysisResponseSet(ctx, d, meta)
}

func updateOutboundCallAnalysisResponseSet(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	responses := d.Get("responses").([]interface{})
	beepDetectionEnabled := d.Get("beep_detection_enabled").(bool)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkResponseSet := platformclientv2.Responseset{
		BeepDetectionEnabled: &beepDetectionEnabled,
	}

	if name != "" {
		sdkResponseSet.Name = &name
	}
	if responses != nil && len(responses) > 0 {
		sdkResponseSet.Responses = buildSdkOutboundCallAnalysisResponseSetReaction(responses)
	}

	log.Printf("Updating Outbound Call Analysis Response Set %s", name)
	diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound Callanalysisresponseset version
		outboundCallanalysisresponseset, resp, getErr := outboundApi.GetOutboundCallanalysisresponseset(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Outbound Call Analysis Response Set %s: %s", d.Id(), getErr)
		}
		sdkResponseSet.Version = outboundCallanalysisresponseset.Version
		outboundCallanalysisresponseset, _, updateErr := outboundApi.PutOutboundCallanalysisresponseset(d.Id(), sdkResponseSet)
		if updateErr != nil {
			return resp, diag.Errorf("Failed to update Outbound Call Analysis Response Set %s: %s", name, updateErr)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Outbound Call Analysis Response Set %s", name)
	return readOutboundCallAnalysisResponseSet(ctx, d, meta)
}

func readOutboundCallAnalysisResponseSet(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	log.Printf("Reading Outbound Call Analysis Response Set %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkResponseSet, resp, getErr := outboundApi.GetOutboundCallanalysisresponseset(d.Id())
		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read Outbound Call Analysis Response Set %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read Outbound Call Analysis Response Set %s: %s", d.Id(), getErr))
		}
		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundCallAnalysisResponseSet())
		if sdkResponseSet.Name != nil {
			_ = d.Set("name", *sdkResponseSet.Name)
		}
		if sdkResponseSet.Responses != nil && len(*sdkResponseSet.Responses) > 0 {
			_ = d.Set("responses", flattenSdkOutboundCallAnalysisResponseSetReaction(sdkResponseSet.Responses))
		}
		if sdkResponseSet.BeepDetectionEnabled != nil {
			_ = d.Set("beep_detection_enabled", *sdkResponseSet.BeepDetectionEnabled)
		}
		log.Printf("Read Outbound Call Analysis Response Set %s %s", d.Id(), *sdkResponseSet.Name)
		return cc.CheckState()
	})
}

func deleteOutboundCallAnalysisResponseSet(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	diagErr := gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Call Analysis Response Set")
		resp, err := outboundApi.DeleteOutboundCallanalysisresponseset(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Outbound Call Analysis Response Set: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := outboundApi.GetOutboundCallanalysisresponseset(d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// Outbound Call Analysis Response Set deleted
				log.Printf("Deleted Outbound Call Analysis Response Set %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting Outbound Call Analysis Response Set %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Outbound Call Analysis Response Set %s still exists", d.Id()))
	})
}

func buildSdkOutboundCallAnalysisResponseSetReaction(responses []interface{}) *map[string]platformclientv2.Reaction {
	if len(responses) == 0 {
		return nil
	}
	sdkResponses := map[string]platformclientv2.Reaction{}
	if responsesMap, ok := responses[0].(map[string]interface{}); ok {
		types := []string{
			"callable_lineconnected",
			"callable_person",
			"callable_busy",
			"callable_noanswer",
			"callable_fax",
			"callable_disconnect",
			"callable_machine",
			"callable_sit",
			"uncallable_sit",
			"uncallable_notfound",
		}
		for _, t := range types {
			reactionSet := responsesMap[t].(*schema.Set).List()
			if len(reactionSet) == 0 {
				continue
			}
			if reactionMap, ok := reactionSet[0].(map[string]interface{}); ok {
				sdkKey := "disposition.classification." + strings.ReplaceAll(t, "_", ".")
				sdkResponses[sdkKey] = *buildSdkReaction(reactionMap)
			}
		}
	}
	return &sdkResponses
}

func buildSdkReaction(reactionMap map[string]interface{}) *platformclientv2.Reaction {
	var sdkReaction platformclientv2.Reaction
	if reactionName, ok := reactionMap["name"].(string); ok {
		sdkReaction.Name = &reactionName
	}
	if reactionData, ok := reactionMap["data"].(string); ok {
		sdkReaction.Data = &reactionData
	}
	if reactionType, ok := reactionMap["reaction_type"].(string); ok {
		sdkReaction.ReactionType = &reactionType
	}
	return &sdkReaction
}

func flattenSdkOutboundCallAnalysisResponseSetReaction(responses *map[string]platformclientv2.Reaction) []interface{} {
	if responses == nil {
		return nil
	}
	responsesMap := make(map[string]interface{})
	for key, val := range *responses {
		schemaKey := strings.Replace(key, "disposition.classification.", "", -1)
		schemaKey = strings.Replace(schemaKey, ".", "_", -1)
		responsesMap[schemaKey] = flattenSdkReaction(val)
	}
	return []interface{}{responsesMap}
}

func flattenSdkReaction(sdkReaction platformclientv2.Reaction) *schema.Set {
	var (
		reactionMap = make(map[string]interface{})
		reactionSet = schema.NewSet(schema.HashResource(outboundCallAnalysisResponseSetReaction), []interface{}{})
	)
	if sdkReaction.Name != nil {
		reactionMap["name"] = *sdkReaction.Name
	}
	if sdkReaction.Data != nil {
		reactionMap["data"] = *sdkReaction.Data
	}
	reactionMap["reaction_type"] = *sdkReaction.ReactionType
	reactionSet.Add(reactionMap)
	return reactionSet
}
