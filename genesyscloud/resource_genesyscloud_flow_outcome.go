package genesyscloud

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func getAllFlowOutcomes(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	archAPI := platformclientv2.NewArchitectApiWithConfig(clientConfig)

	const pageSize = 100
	for pageNum := 1; ; pageNum++ {
		outcomes, _, err := archAPI.GetFlowsOutcomes(pageNum, pageSize, "", "", nil, "", "", "", nil)

		if err != nil {
			return nil, diag.Errorf("Failed to get page of outcomes: %v", err)
		}

		if outcomes.Entities == nil || len(*outcomes.Entities) == 0 {
			break
		}

		for _, outcome := range *outcomes.Entities {
			resources[*outcome.Id] = &resourceExporter.ResourceMeta{Name: *outcome.Name}
		}
	}

	return resources, nil
}

func FlowOutcomeExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllFlowOutcomes),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
		},
	}
}

func ResourceFlowOutcome() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud flow outcome`,

		CreateContext: CreateWithPooledClient(createFlowOutcome),
		ReadContext:   ReadWithPooledClient(readFlowOutcome),
		UpdateContext: UpdateWithPooledClient(updateFlowOutcome),
		DeleteContext: DeleteWithPooledClient(deleteFlowOutcome),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The flow outcome name.`,
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
				Description: `This is a description for the flow outcome.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func createFlowOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	division := d.Get("division_id").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	sdkflowoutcome := platformclientv2.Flowoutcome{}

	if name != "" {
		sdkflowoutcome.Name = &name
	}
	if division != "" {
		sdkflowoutcome.Division = &platformclientv2.Writabledivision{Id: &division}
	}
	sdkflowoutcome.Description = &description

	log.Printf("Creating Flow Outcome %s", name)
	flowOutcome, _, err := architectApi.PostFlowsOutcomes(sdkflowoutcome)
	if err != nil {
		return diag.Errorf("Failed to create Flow Outcome %s: %s", name, err)
	}

	d.SetId(*flowOutcome.Id)

	log.Printf("Created Flow Outcome %s %s", name, *flowOutcome.Id)
	return readFlowOutcome(ctx, d, meta)
}

func updateFlowOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	division := d.Get("division_id").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	sdkflowoutcome := platformclientv2.Flowoutcome{}

	if name != "" {
		sdkflowoutcome.Name = &name
	}
	if division != "" {
		sdkflowoutcome.Division = &platformclientv2.Writabledivision{Id: &division}
	}
	sdkflowoutcome.Description = &description

	log.Printf("Updating Flow Outcome %s", name)

	_, _, updateErr := architectApi.PutFlowsOutcome(d.Id(), sdkflowoutcome)

	if updateErr != nil {
		return diag.Errorf("Failed to update Flow Outcome %s: %s", name, updateErr)
	}

	log.Printf("Updated Flow Outcome %s", name)
	return readFlowOutcome(ctx, d, meta)
}

func readFlowOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Reading Flow Outcome %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkflowoutcome, resp, getErr := architectApi.GetFlowsOutcome(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read Flow Outcome %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Flow Outcome %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceFlowOutcome())

		if sdkflowoutcome.Name != nil {
			d.Set("name", *sdkflowoutcome.Name)
		}
		if sdkflowoutcome.Division != nil && sdkflowoutcome.Division.Id != nil {
			d.Set("division_id", *sdkflowoutcome.Division.Id)
		}
		if sdkflowoutcome.Description != nil {
			d.Set("description", *sdkflowoutcome.Description)
		}

		log.Printf("Read Flow Outcome %s %s", d.Id(), *sdkflowoutcome.Name)
		return cc.CheckState()
	})
}

func deleteFlowOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
