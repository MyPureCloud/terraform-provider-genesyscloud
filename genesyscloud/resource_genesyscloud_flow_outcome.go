package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v74/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"log"
)

func resourceFlowOutcome() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud flow outcome`,

		CreateContext: createWithPooledClient(createFlowOutcome),
		ReadContext:   readWithPooledClient(readFlowOutcome),
		UpdateContext: updateWithPooledClient(updateFlowOutcome),
		DeleteContext: deleteWithPooledClient(deleteFlowOutcome),
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
				Description: `TODO: Add appropriate description`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`current_operation_id`: {
				Description: `TODO: Add appropriate description`,
				Optional:    true,
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func createFlowOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	division := d.Get("division_id").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	sdkflowoutcome := platformclientv2.Flowoutcome{}

	if name != "" {
		sdkflowoutcome.Name = &name
	}
	if division != "" {
		sdkflowoutcome.Division = &platformclientv2.Writabledivision{Id: &division}
	}
	if description != "" {
		sdkflowoutcome.Description = &description
	}

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

	sdkConfig := meta.(*providerMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	sdkflowoutcome := platformclientv2.Flowoutcome{}

	if name != "" {
		sdkflowoutcome.Name = &name
	}
	if division != "" {
		sdkflowoutcome.Division = &platformclientv2.Writabledivision{Id: &division}
	}
	if description != "" {
		sdkflowoutcome.Description = &description
	}

	log.Printf("Updating Flow Outcome %s", name)

	_, _, updateErr := architectApi.PutFlowsOutcome(d.Id(), sdkflowoutcome)

	if updateErr != nil {
		return diag.Errorf("Failed to update Flow Outcome %s: %s", name, updateErr)
	}

	log.Printf("Updated Flow Outcome %s", name)
	return readFlowOutcome(ctx, d, meta)
}

func readFlowOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Reading Flow Outcome %s", d.Id())

	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		sdkflowoutcome, resp, getErr := architectApi.GetFlowsOutcome(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read Flow Outcome %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read Flow Outcome %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceFlowOutcome())

		if sdkflowoutcome.Name != nil {
			d.Set("name", *sdkflowoutcome.Name)
		}
		if sdkflowoutcome.Division != nil && sdkflowoutcome.Division.Id != nil {
			d.Set("division_id", *sdkflowoutcome.Division.Id)
		}
		if sdkflowoutcome.Description != nil {
			d.Set("description", *sdkflowoutcome.Description)
		}
		if sdkflowoutcome.CurrentOperation != nil && sdkflowoutcome.CurrentOperation.Id != nil {
			d.Set("current_operation_id", *sdkflowoutcome.CurrentOperation.Id)
		}

		log.Printf("Read Flow Outcome %s %s", d.Id(), *sdkflowoutcome.Name)
		return cc.CheckState()
	})
}

func deleteFlowOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// This resource can not be deleted. TF will no longer manage this resource
	sdkConfig := meta.(*providerMeta).ClientConfig
	apiInstance := platformclientv2.NewObjectsApiWithConfig(sdkConfig)

	divisionId := d.Get("division_id").(string)
	fmt.Printf("Division Id: %s\n", divisionId)
	response, err := apiInstance.DeleteAuthorizationDivision(divisionId, true)
	
	fmt.Printf("Response:\n  Success: %v\n  Status code: %v\n  Correlation ID: %v\n", response.IsSuccess, response.StatusCode, response.CorrelationID)
	if err != nil {
		fmt.Printf("Error calling DeleteAuthorizationDivision: %v\n", err)
	}
	return nil
}
