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

func getAllFlowMilestones(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	archAPI := platformclientv2.NewArchitectApiWithConfig(clientConfig)

	const pageSize = 100
	for pageNum := 1; ; pageNum++ {
		milestones, _, err := archAPI.GetFlowsMilestones(pageNum, pageSize, "", "", nil, "", "", "", nil)

		if err != nil {
			return nil, diag.Errorf("Failed to get page of milestones: %v", err)
		}

		if milestones.Entities == nil || len(*milestones.Entities) == 0 {
			break
		}

		for _, milestone := range *milestones.Entities {
			resources[*milestone.Id] = &resourceExporter.ResourceMeta{Name: *milestone.Name}
		}
	}

	return resources, nil
}

func FlowMilestoneExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllFlowMilestones),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
		},
	}
}

func ResourceFlowMilestone() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud flow milestone`,

		CreateContext: CreateWithPooledClient(createFlowMilestone),
		ReadContext:   ReadWithPooledClient(readFlowMilestone),
		UpdateContext: UpdateWithPooledClient(updateFlowMilestone),
		DeleteContext: DeleteWithPooledClient(deleteFlowMilestone),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The flow milestone name.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`division_id`: {
				Description: `The division to which this entity belongs.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			`description`: {
				Description: `The flow milestone description.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func createFlowMilestone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	division := d.Get("division_id").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	sdkflowmilestone := platformclientv2.Flowmilestone{}

	if name != "" {
		sdkflowmilestone.Name = &name
	}
	if division != "" {
		sdkflowmilestone.Division = &platformclientv2.Writabledivision{Id: &division}
	}
	if description != "" {
		sdkflowmilestone.Description = &description
	}

	log.Printf("Creating Flow Milestone %s", name)
	flowMilestone, _, err := architectApi.PostFlowsMilestones(sdkflowmilestone)
	if err != nil {
		return diag.Errorf("Failed to create Flow Milestone %s: %s", name, err)
	}

	d.SetId(*flowMilestone.Id)

	log.Printf("Created Flow Milestone %s %s", name, *flowMilestone.Id)
	return readFlowMilestone(ctx, d, meta)
}

func updateFlowMilestone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	division := d.Get("division_id").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	sdkflowmilestone := platformclientv2.Flowmilestone{}

	if name != "" {
		sdkflowmilestone.Name = &name
	}
	if division != "" {
		sdkflowmilestone.Division = &platformclientv2.Writabledivision{Id: &division}
	}
	if description != "" {
		sdkflowmilestone.Description = &description
	}

	log.Printf("Updating Flow Milestone %s", name)
	_, _, err := architectApi.PutFlowsMilestone(d.Id(), sdkflowmilestone)
	if err != nil {
		return diag.Errorf("Failed to update Flow Milestone %s: %s", name, err)
	}

	log.Printf("Updated Flow Milestone %s", name)
	return readFlowMilestone(ctx, d, meta)
}

func readFlowMilestone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Reading Flow Milestone %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkflowmilestone, resp, getErr := architectApi.GetFlowsMilestone(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read Flow Milestone %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Flow Milestone %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceFlowMilestone())

		if sdkflowmilestone.Name != nil {
			d.Set("name", *sdkflowmilestone.Name)
		}
		if sdkflowmilestone.Division != nil && sdkflowmilestone.Division.Id != nil {
			d.Set("division_id", *sdkflowmilestone.Division.Id)
		}
		if sdkflowmilestone.Description != nil {
			d.Set("description", *sdkflowmilestone.Description)
		}

		log.Printf("Read Flow Milestone %s %s", d.Id(), *sdkflowmilestone.Name)

		return cc.CheckState()
	})
}

func deleteFlowMilestone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	diagErr := RetryWhen(IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Flow Milestone")
		_, resp, err := architectApi.DeleteFlowsMilestone(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Flow Milestone: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := architectApi.GetFlowsMilestone(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Flow Milestone deleted
				log.Printf("Deleted Flow Milestone %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting Flow Milestone %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Flow Milestone %s still exists", d.Id()))
	})
}
