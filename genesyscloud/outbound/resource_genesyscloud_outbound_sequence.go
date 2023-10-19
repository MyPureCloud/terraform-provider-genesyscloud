package outbound

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func ResourceOutboundSequence() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud outbound sequence`,

		CreateContext: gcloud.CreateWithPooledClient(createOutboundSequence),
		ReadContext:   gcloud.ReadWithPooledClient(readOutboundSequence),
		UpdateContext: gcloud.UpdateWithPooledClient(updateOutboundSequence),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteOutboundSequence),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `Name of outbound sequence`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`campaign_ids`: {
				Description: `The ordered list of Campaigns that this CampaignSequence will run.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`status`: {
				Description:  `The current status of the CampaignSequence. A CampaignSequence can be turned 'on' or 'off' (default). Changing from "on" to "off" will cause the current sequence to drop and be recreated with a new ID.`,
				Optional:     true,
				Computed:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`on`, `off`}, false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return (old == `complete` && new == `on`)
				},
			},
			`repeat`: {
				Description: `Indicates if a sequence should repeat from the beginning after the last campaign completes. Default is false.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
		},
		CustomizeDiff: customdiff.ForceNewIfChange("status", func(ctx context.Context, old, new, meta any) bool {
			return new.(string) == "off" && (old.(string) == "on" || old.(string) == "complete")
		}),
	}
}

func getAllOutboundSequence(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	outboundApi := platformclientv2.NewOutboundApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		sdkcampaignsequenceentitylisting, _, getErr := outboundApi.GetOutboundSequences(pageSize, pageNum, true, "", "", "", "")
		if getErr != nil {
			return nil, diag.Errorf("Error requesting page of Outbound Sequence: %s", getErr)
		}

		if sdkcampaignsequenceentitylisting.Entities == nil || len(*sdkcampaignsequenceentitylisting.Entities) == 0 {
			break
		}

		for _, entity := range *sdkcampaignsequenceentitylisting.Entities {
			if *entity.Status != "off" && *entity.Status != "on" {
				*entity.Status = "off"
			}
			resources[*entity.Id] = &resourceExporter.ResourceMeta{Name: *entity.Name}
		}
	}

	return resources, nil
}

func OutboundSequenceExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllOutboundSequence),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			`campaign_ids`: {
				RefType: "genesyscloud_outbound_campaign",
			},
		},
	}
}

func createOutboundSequence(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	status := d.Get("status").(string)
	repeat := d.Get("repeat").(bool)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkcampaignsequence := platformclientv2.Campaignsequence{
		Campaigns: gcloud.BuildSdkDomainEntityRefArr(d, "campaign_ids"),
		Repeat:    &repeat,
	}

	if name != "" {
		sdkcampaignsequence.Name = &name
	}

	// All campaigns sequences have to be created in an "off" state to start out with
	defaultStatus := "off"
	sdkcampaignsequence.Status = &defaultStatus

	log.Printf("Creating Outbound Sequence %s", name)
	outboundSequence, _, err := outboundApi.PostOutboundSequences(sdkcampaignsequence)
	if err != nil {
		return diag.Errorf("Failed to create Outbound Sequence %s: %s", name, err)
	}

	d.SetId(*outboundSequence.Id)
	log.Printf("Created Outbound Sequence %s %s", name, *outboundSequence.Id)

	// Campaigns sequences can be enabled after creation
	if status == "on" {
		d.Set("status", status)
		diag := updateOutboundSequence(ctx, d, meta)
		if diag != nil {
			return diag
		}
	}

	return readOutboundSequence(ctx, d, meta)
}

func updateOutboundSequence(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	status := d.Get("status").(string)
	repeat := d.Get("repeat").(bool)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkcampaignsequence := platformclientv2.Campaignsequence{
		Campaigns: gcloud.BuildSdkDomainEntityRefArr(d, "campaign_ids"),
		Repeat:    &repeat,
	}

	if name != "" {
		sdkcampaignsequence.Name = &name
	}
	if status != "" {
		sdkcampaignsequence.Status = &status
	}

	log.Printf("Updating Outbound Sequence %s", name)
	diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound Sequence version
		outboundSequence, resp, getErr := outboundApi.GetOutboundSequence(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Outbound Sequence %s: %s", d.Id(), getErr)
		}
		sdkcampaignsequence.Version = outboundSequence.Version
		outboundSequence, _, updateErr := outboundApi.PutOutboundSequence(d.Id(), sdkcampaignsequence)
		if updateErr != nil {
			return resp, diag.Errorf("Failed to update Outbound Sequence %s: %s", name, updateErr)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Outbound Sequence %s", name)
	return readOutboundSequence(ctx, d, meta)
}

func readOutboundSequence(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	log.Printf("Reading Outbound Sequence %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkcampaignsequence, resp, getErr := outboundApi.GetOutboundSequence(d.Id())
		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read Outbound Sequence %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Outbound Sequence %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundSequence())

		if sdkcampaignsequence.Name != nil {
			d.Set("name", *sdkcampaignsequence.Name)
		}
		if sdkcampaignsequence.Campaigns != nil {
			d.Set("campaign_ids", gcloud.SdkDomainEntityRefArrToList(*sdkcampaignsequence.Campaigns))
		}
		if sdkcampaignsequence.Status != nil {
			d.Set("status", *sdkcampaignsequence.Status)
		}
		if sdkcampaignsequence.Repeat != nil {
			d.Set("repeat", *sdkcampaignsequence.Repeat)
		}

		log.Printf("Read Outbound Sequence %s %s", d.Id(), *sdkcampaignsequence.Name)

		return cc.CheckState()
	})
}

func deleteOutboundSequence(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	diagErr := gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Sequence")
		resp, err := outboundApi.DeleteOutboundSequence(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Outbound Sequence: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := outboundApi.GetOutboundSequence(d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// Outbound Sequence deleted
				log.Printf("Deleted Outbound Sequence %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting Outbound Sequence %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Outbound Sequence %s still exists", d.Id()))
	})
}
