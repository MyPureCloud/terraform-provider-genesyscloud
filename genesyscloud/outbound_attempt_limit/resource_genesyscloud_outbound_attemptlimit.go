package outbound_attempt_limit

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
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

var (
	recallSettings = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`nbr_attempts`: {
				Description: `Number of recall attempts. Must be less than max_attempts_per_contact.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeInt,
			},
			`minutes_between_attempts`: {
				Description:  `Number of minutes between attempts. Must be greater than or equal to 5.`,
				Required:     true,
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntAtLeast(5),
			},
		},
	}

	attemptLimitRecallSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`answering_machine`: {
				Optional: true,
				MaxItems: 1,
				Type:     schema.TypeSet,
				Elem:     recallSettings,
			},
			`busy`: {
				Optional: true,
				MaxItems: 1,
				Type:     schema.TypeSet,
				Elem:     recallSettings,
			},
			`fax`: {
				Optional: true,
				MaxItems: 1,
				Type:     schema.TypeSet,
				Elem:     recallSettings,
			},
			`no_answer`: {
				Optional: true,
				MaxItems: 1,
				Type:     schema.TypeSet,
				Elem:     recallSettings,
			},
		},
	}
)

func getAllAttemptLimits(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		attemptLimitConfigs, _, getErr := outboundAPI.GetOutboundAttemptlimits(pageSize, pageNum, true, "", "", "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of attempt limit configs: %v", getErr)
		}

		if attemptLimitConfigs.Entities == nil || len(*attemptLimitConfigs.Entities) == 0 {
			break
		}

		for _, attemptLimitConfig := range *attemptLimitConfigs.Entities {
			resources[*attemptLimitConfig.Id] = &resourceExporter.ResourceMeta{Name: *attemptLimitConfig.Name}
		}
	}

	return resources, nil
}

func OutboundAttemptLimitExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllAttemptLimits),
	}
}

func ResourceOutboundAttemptLimit() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Outbound Attempt Limit`,

		CreateContext: gcloud.CreateWithPooledClient(createOutboundAttemptLimit),
		ReadContext:   gcloud.ReadWithPooledClient(readOutboundAttemptLimit),
		UpdateContext: gcloud.UpdateWithPooledClient(updateOutboundAttemptLimit),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteOutboundAttemptLimit),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name for the attempt limit.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`max_attempts_per_contact`: {
				Description: `The maximum number of times a contact can be called within the resetPeriod. Required if maxAttemptsPerNumber is not defined.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`max_attempts_per_number`: {
				Description: `The maximum number of times a phone number can be called within the resetPeriod. Required if maxAttemptsPerContact is not defined.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`time_zone_id`: {
				Description: `If the resetPeriod is TODAY, this specifies the timezone in which TODAY occurs. Required if the resetPeriod is TODAY.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`reset_period`: {
				Description:  `After how long the number of attempts will be set back to 0.`,
				Optional:     true,
				Type:         schema.TypeString,
				Default:      `NEVER`,
				ValidateFunc: validation.StringInSlice([]string{`NEVER`, `TODAY`}, true),
			},
			`recall_entries`: {
				Description: `Configuration for recall attempts.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeList,
				Elem:        attemptLimitRecallSettingsResource,
			},
		},
	}
}

func createOutboundAttemptLimit(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	maxAttemptsPerContact := d.Get("max_attempts_per_contact").(int)
	maxAttemptsPerNumber := d.Get("max_attempts_per_number").(int)
	timeZoneId := d.Get("time_zone_id").(string)
	resetPeriod := d.Get("reset_period").(string)
	recallEntries := d.Get("recall_entries").([]interface{})

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkAttemptLimits := platformclientv2.Attemptlimits{}

	if name != "" {
		sdkAttemptLimits.Name = &name
	}
	if maxAttemptsPerContact != 0 {
		sdkAttemptLimits.MaxAttemptsPerContact = &maxAttemptsPerContact
	}
	if maxAttemptsPerNumber != 0 {
		sdkAttemptLimits.MaxAttemptsPerNumber = &maxAttemptsPerNumber
	}
	if timeZoneId != "" {
		sdkAttemptLimits.TimeZoneId = &timeZoneId
	}
	if resetPeriod != "" {
		sdkAttemptLimits.ResetPeriod = &resetPeriod
	}
	if recallEntries != nil && len(recallEntries) > 0 {
		sdkAttemptLimits.RecallEntries = buildSdkOutboundAttemptLimitRecallEntryMap(recallEntries)
	}

	log.Printf("Creating Outbound Attempt Limit %s", name)
	outboundAttemptLimit, _, err := outboundApi.PostOutboundAttemptlimits(sdkAttemptLimits)
	if err != nil {
		return diag.Errorf("Failed to create Outbound Attempt Limit %s: %s", name, err)
	}

	d.SetId(*outboundAttemptLimit.Id)

	log.Printf("Created Outbound Attempt Limit %s %s", name, *outboundAttemptLimit.Id)
	return readOutboundAttemptLimit(ctx, d, meta)
}

func updateOutboundAttemptLimit(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	maxAttemptsPerContact := d.Get("max_attempts_per_contact").(int)
	maxAttemptsPerNumber := d.Get("max_attempts_per_number").(int)
	timeZoneId := d.Get("time_zone_id").(string)
	resetPeriod := d.Get("reset_period").(string)
	recallEntries := d.Get("recall_entries").([]interface{})

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkAttemptLimits := platformclientv2.Attemptlimits{}

	if name != "" {
		sdkAttemptLimits.Name = &name
	}
	if maxAttemptsPerContact != 0 {
		sdkAttemptLimits.MaxAttemptsPerContact = &maxAttemptsPerContact
	}
	if maxAttemptsPerNumber != 0 {
		sdkAttemptLimits.MaxAttemptsPerNumber = &maxAttemptsPerNumber
	}
	if timeZoneId != "" {
		sdkAttemptLimits.TimeZoneId = &timeZoneId
	}
	if resetPeriod != "" {
		sdkAttemptLimits.ResetPeriod = &resetPeriod
	}
	if recallEntries != nil && len(recallEntries) > 0 {
		sdkAttemptLimits.RecallEntries = buildSdkOutboundAttemptLimitRecallEntryMap(recallEntries)
	}

	log.Printf("Updating Outbound Attempt Limit %s", name)
	diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound Attempt Limit version
		outboundAttemptLimit, resp, getErr := outboundApi.GetOutboundAttemptlimit(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Outbound Attempt Limit %s: %s", d.Id(), getErr)
		}
		sdkAttemptLimits.Version = outboundAttemptLimit.Version
		outboundAttemptLimit, _, updateErr := outboundApi.PutOutboundAttemptlimit(d.Id(), sdkAttemptLimits)
		if updateErr != nil {
			return resp, diag.Errorf("Failed to update Outbound Attempt Limit %s: %s", name, updateErr)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Outbound Attempt Limit %s", name)
	return readOutboundAttemptLimit(ctx, d, meta)
}

func readOutboundAttemptLimit(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	log.Printf("Reading Outbound Attempt Limit %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkAttemptLimits, resp, getErr := outboundApi.GetOutboundAttemptlimit(d.Id())
		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read Outbound Attempt Limit %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read Outbound Attempt Limit %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundAttemptLimit())

		if sdkAttemptLimits.Name != nil {
			_ = d.Set("name", *sdkAttemptLimits.Name)
		}
		if sdkAttemptLimits.MaxAttemptsPerContact != nil {
			_ = d.Set("max_attempts_per_contact", *sdkAttemptLimits.MaxAttemptsPerContact)
		}
		if sdkAttemptLimits.MaxAttemptsPerNumber != nil {
			_ = d.Set("max_attempts_per_number", *sdkAttemptLimits.MaxAttemptsPerNumber)
		}
		if sdkAttemptLimits.TimeZoneId != nil {
			_ = d.Set("time_zone_id", *sdkAttemptLimits.TimeZoneId)
		}
		if sdkAttemptLimits.ResetPeriod != nil {
			_ = d.Set("reset_period", *sdkAttemptLimits.ResetPeriod)
		}

		if sdkAttemptLimits.RecallEntries != nil && len(*sdkAttemptLimits.RecallEntries) > 0 {
			_ = d.Set("recall_entries", flattenSdkOutboundAttemptLimitRecallEntry(sdkAttemptLimits.RecallEntries))
		}

		log.Printf("Read Outbound Attempt Limit %s %s", d.Id(), *sdkAttemptLimits.Name)
		return cc.CheckState()
	})
}

func deleteOutboundAttemptLimit(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	diagErr := gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Attempt Limit")
		resp, err := outboundApi.DeleteOutboundAttemptlimit(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Outbound Attempt Limit: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := outboundApi.GetOutboundAttemptlimit(d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// Outbound Attempt Limit deleted
				log.Printf("Deleted Outbound Attempt Limit %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting Outbound Attempt Limit %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Outbound Attempt Limit %s still exists", d.Id()))
	})
}

func buildSdkOutboundAttemptLimitRecallEntryMap(recallEntries []interface{}) *map[string]platformclientv2.Recallentry {
	if len(recallEntries) == 0 {
		return nil
	}
	recallEntriesMap := map[string]platformclientv2.Recallentry{}
	if entriesMap, ok := recallEntries[0].(map[string]interface{}); ok {
		types := []string{"busy", "no_answer", "answering_machine", "fax"}
		for _, t := range types {
			entrySet := entriesMap[t].(*schema.Set).List()
			if len(entrySet) == 0 {
				continue
			}
			if entryMap, ok := entrySet[0].(map[string]interface{}); ok && len(entryMap) > 0 {
				recallEntriesMap[gcloud.ToCamelCase(t)] = *buildSdkRecallEntry(entryMap)
			}
		}
	}
	return &recallEntriesMap
}

func buildSdkRecallEntry(entry map[string]interface{}) *platformclientv2.Recallentry {
	sdkRecallEntry := platformclientv2.Recallentry{}
	if nbrAttempts, ok := entry["nbr_attempts"].(int); ok {
		sdkRecallEntry.NbrAttempts = &nbrAttempts
	}
	if minsBetweenAttempts, ok := entry["minutes_between_attempts"].(int); ok {
		sdkRecallEntry.MinutesBetweenAttempts = &minsBetweenAttempts
	}
	return &sdkRecallEntry
}

func flattenSdkOutboundAttemptLimitRecallEntry(sdkRecallEntries *map[string]platformclientv2.Recallentry) []interface{} {
	recallEntries := make(map[string]interface{})
	for key, val := range *sdkRecallEntries {
		recallEntries[gcloud.ToSnakeCase(key)] = flattenSdkRecallEntry(val)
	}
	return []interface{}{recallEntries}
}

func flattenSdkRecallEntry(sdkEntry platformclientv2.Recallentry) *schema.Set {
	var (
		entryMap = make(map[string]interface{})
		entrySet = schema.NewSet(schema.HashResource(recallSettings), []interface{}{})
	)
	entryMap["nbr_attempts"] = *sdkEntry.NbrAttempts
	entryMap["minutes_between_attempts"] = *sdkEntry.MinutesBetweenAttempts
	entrySet.Add(entryMap)
	return entrySet
}

func GenerateAttemptLimitResource(
	resourceId string,
	name string,
	maxAttemptsPerContact string,
	maxAttemptsPerNumber string,
	timeZoneId string,
	resetPeriod string,
	nestedBlocks ...string,
) string {
	if maxAttemptsPerContact != "" {
		maxAttemptsPerContact = fmt.Sprintf(`max_attempts_per_contact = %s`, maxAttemptsPerContact)
	}
	if maxAttemptsPerNumber != "" {
		maxAttemptsPerNumber = fmt.Sprintf(`max_attempts_per_number = %s`, maxAttemptsPerNumber)
	}
	if timeZoneId != "" {
		timeZoneId = fmt.Sprintf(`time_zone_id = "%s"`, timeZoneId)
	}
	if resetPeriod != "" {
		resetPeriod = fmt.Sprintf(`reset_period = "%s"`, resetPeriod)
	}
	return fmt.Sprintf(`
resource "genesyscloud_outbound_attempt_limit" "%s" {
	name = "%s"
	%s
	%s
	%s
	%s
	%s
}
	`, resourceId, name, maxAttemptsPerContact, maxAttemptsPerNumber, timeZoneId, resetPeriod, strings.Join(nestedBlocks, "\n"))
}
