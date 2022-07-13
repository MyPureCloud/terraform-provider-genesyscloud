package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v74/platformclientv2"
	"log"
	"time"
)

func resourceFlowsMilestone() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud flows milestone`,

		CreateContext: createWithPooledClient(createFlowsMilestone),
		ReadContext:   readWithPooledClient(readFlowsMilestone),
		UpdateContext: updateWithPooledClient(updateFlowsMilestone),
		DeleteContext: deleteWithPooledClient(deleteFlowsMilestone),
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

func createFlowsMilestone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	division := d.Get("division_id").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
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

	log.Printf("Creating Flows Milestone %s", name)
	flowsMilestone, _, err := architectApi.PostFlowsMilestones(sdkflowmilestone)
	if err != nil {
		return diag.Errorf("Failed to create Flows Milestone %s: %s", name, err)
	}

	d.SetId(*flowsMilestone.Id)

	log.Printf("Created Flows Milestone %s %s", name, *flowsMilestone.Id)
	return readFlowsMilestone(ctx, d, meta)
}

func updateFlowsMilestone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	division := d.Get("division_id").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
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

	log.Printf("Updating Flows Milestone %s", name)
	_, _, err := architectApi.PutFlowsMilestone(d.Id(), sdkflowmilestone)
	if err != nil {
		return diag.Errorf("Failed to update Flows Milestone %s: %s", name, err)
	}

	log.Printf("Updated Flows Milestone %s", name)
	return readFlowsMilestone(ctx, d, meta)
}

func readFlowsMilestone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Reading Flows Milestone %s", d.Id())

	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		sdkflowmilestone, resp, getErr := architectApi.GetFlowsMilestone(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read Flows Milestone %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read Flows Milestone %s: %s", d.Id(), getErr))
		}

		// cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceFlowsMilestone())

		if sdkflowmilestone.Name != nil {
			d.Set("name", *sdkflowmilestone.Name)
		}
		if sdkflowmilestone.Division != nil && sdkflowmilestone.Division.Id != nil {
			d.Set("division_id", *sdkflowmilestone.Division.Id)
		}
		if sdkflowmilestone.Description != nil {
			d.Set("description", *sdkflowmilestone.Description)
		}

		log.Printf("Read Flows Milestone %s %s", d.Id(), *sdkflowmilestone.Name)
		return nil // TODO calling cc.CheckState() can cause some difficult to understand errors in development. When ready for a PR, remove this line and uncomment the consistency_checker initialization and the the below one
		// return cc.CheckState()
	})
}

func deleteFlowsMilestone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	diagErr := retryWhen(isStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Flows Milestone")
		_, resp, err := architectApi.DeleteFlowsMilestone(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Flows Milestone: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := architectApi.GetFlowsMilestone(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// Flows Milestone deleted
				log.Printf("Deleted Flows Milestone %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting Flows Milestone %s: %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("Flows Milestone %s still exists", d.Id()))
	})
}
