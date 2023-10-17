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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func ResourceEmployeeperformanceExternalmetricsDefinition() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud employeeperformance externalmetrics definition`,

		CreateContext: CreateWithPooledClient(createEmployeeperformanceExternalmetricsDefinition),
		ReadContext:   ReadWithPooledClient(readEmployeeperformanceExternalmetricsDefinition),
		UpdateContext: UpdateWithPooledClient(updateEmployeeperformanceExternalmetricsDefinition),
		DeleteContext: DeleteWithPooledClient(deleteEmployeeperformanceExternalmetricsDefinition),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the External Metric Definition`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`precision`: {
				Description:  `The decimal precision of the External Metric Definition. Must be at least 0 and at most 5`,
				Required:     true,
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntBetween(0, 5),
			},
			`default_objective_type`: {
				Description:  `The default objective type of the External Metric Definition`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`HigherIsBetter`, `LowerIsBetter`, `TargetArea`}, false),
			},
			`enabled`: {
				Description: `True if the External Metric Definition is enabled`,
				Required:    true,
				Type:        schema.TypeBool,
			},
			`unit`: {
				Description:  `The unit of the External Metric Definition. Note: Changing the unit property will cause the external metric object to be dropped and recreated with a new ID.`,
				Required:     true,
				ForceNew:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`Seconds`, `Percent`, `Number`, `Currency`}, false),
			},
			`unit_definition`: {
				Description: `The unit definition of the External Metric Definition. Note: Changing the unit definition property will cause the external metric object to be dropped and recreated with a new ID.`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func getAllEmployeeperformanceExternalmetricsDefinition(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	gamificationApi := platformclientv2.NewGamificationApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		sdkexternalmetricdefinitionlisting, _, getErr := gamificationApi.GetEmployeeperformanceExternalmetricsDefinitions(pageSize, pageNum)
		if getErr != nil {
			return nil, diag.Errorf("Error requesting page of Employeeperformance Externalmetrics Definition: %s", getErr)
		}

		if sdkexternalmetricdefinitionlisting.Entities == nil || len(*sdkexternalmetricdefinitionlisting.Entities) == 0 {
			break
		}

		for _, entity := range *sdkexternalmetricdefinitionlisting.Entities {
			resources[*entity.Id] = &resourceExporter.ResourceMeta{Name: *entity.Name}
		}
	}

	return resources, nil
}

func EmployeeperformanceExternalmetricsDefinitionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllEmployeeperformanceExternalmetricsDefinition),
		AllowZeroValues:  []string{"precision"},
	}
}

func createEmployeeperformanceExternalmetricsDefinition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	precision := d.Get("precision").(int)
	defaultObjectiveType := d.Get("default_objective_type").(string)
	enabled := d.Get("enabled").(bool)
	unit := d.Get("unit").(string)
	unitDefinition := d.Get("unit_definition").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	gamificationApi := platformclientv2.NewGamificationApiWithConfig(sdkConfig)

	sdkexternalmetricdefinitioncreaterequest := platformclientv2.Externalmetricdefinitioncreaterequest{
		Precision: &precision,
		Enabled:   &enabled,
	}

	if name != "" {
		sdkexternalmetricdefinitioncreaterequest.Name = &name
	}
	if defaultObjectiveType != "" {
		sdkexternalmetricdefinitioncreaterequest.DefaultObjectiveType = &defaultObjectiveType
	}
	if unit != "" {
		sdkexternalmetricdefinitioncreaterequest.Unit = &unit
	}
	if unitDefinition != "" {
		sdkexternalmetricdefinitioncreaterequest.UnitDefinition = &unitDefinition
	}

	log.Printf("Creating Employeeperformance Externalmetrics Definition %s", name)
	employeeperformanceExternalmetricsDefinition, _, err := gamificationApi.PostEmployeeperformanceExternalmetricsDefinitions(sdkexternalmetricdefinitioncreaterequest)
	if err != nil {
		return diag.Errorf("Failed to create Employeeperformance Externalmetrics Definition %s: %s", name, err)
	}

	d.SetId(*employeeperformanceExternalmetricsDefinition.Id)

	log.Printf("Created Employeeperformance Externalmetrics Definition %s %s", name, *employeeperformanceExternalmetricsDefinition.Id)
	return readEmployeeperformanceExternalmetricsDefinition(ctx, d, meta)
}

func updateEmployeeperformanceExternalmetricsDefinition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	precision := d.Get("precision").(int)
	defaultObjectiveType := d.Get("default_objective_type").(string)
	enabled := d.Get("enabled").(bool)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	gamificationApi := platformclientv2.NewGamificationApiWithConfig(sdkConfig)

	sdkexternalmetricdefinitionupdaterequest := platformclientv2.Externalmetricdefinitionupdaterequest{
		Enabled: &enabled,
	}

	if name != "" {
		sdkexternalmetricdefinitionupdaterequest.Name = &name
	}
	if precision != 0 {
		sdkexternalmetricdefinitionupdaterequest.Precision = &precision
	}
	if defaultObjectiveType != "" {
		sdkexternalmetricdefinitionupdaterequest.DefaultObjectiveType = &defaultObjectiveType
	}

	log.Printf("Updating Employeeperformance Externalmetrics Definition %s", name)
	_, _, err := gamificationApi.PatchEmployeeperformanceExternalmetricsDefinition(d.Id(), sdkexternalmetricdefinitionupdaterequest)
	if err != nil {
		return diag.Errorf("Failed to update Employeeperformance Externalmetrics Definition %s: %s", name, err)
	}

	log.Printf("Updated Employeeperformance Externalmetrics Definition %s", name)
	return readEmployeeperformanceExternalmetricsDefinition(ctx, d, meta)
}

func readEmployeeperformanceExternalmetricsDefinition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	gamificationApi := platformclientv2.NewGamificationApiWithConfig(sdkConfig)

	log.Printf("Reading Employeeperformance Externalmetrics Definition %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkexternalmetricdefinition, resp, getErr := gamificationApi.GetEmployeeperformanceExternalmetricsDefinition(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read Employeeperformance Externalmetrics Definition %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Employeeperformance Externalmetrics Definition %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceEmployeeperformanceExternalmetricsDefinition())

		if sdkexternalmetricdefinition.Name != nil {
			d.Set("name", *sdkexternalmetricdefinition.Name)
		}
		if sdkexternalmetricdefinition.Precision != nil {
			d.Set("precision", *sdkexternalmetricdefinition.Precision)
		}
		if sdkexternalmetricdefinition.DefaultObjectiveType != nil {
			d.Set("default_objective_type", *sdkexternalmetricdefinition.DefaultObjectiveType)
		}
		if sdkexternalmetricdefinition.Enabled != nil {
			d.Set("enabled", *sdkexternalmetricdefinition.Enabled)
		}
		if sdkexternalmetricdefinition.Unit != nil {
			d.Set("unit", *sdkexternalmetricdefinition.Unit)
		}
		if sdkexternalmetricdefinition.UnitDefinition != nil {
			d.Set("unit_definition", *sdkexternalmetricdefinition.UnitDefinition)
		}

		log.Printf("Read Employeeperformance Externalmetrics Definition %s %s", d.Id(), *sdkexternalmetricdefinition.Name)
		return cc.CheckState()
	})
}

func deleteEmployeeperformanceExternalmetricsDefinition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	gamificationApi := platformclientv2.NewGamificationApiWithConfig(sdkConfig)

	diagErr := RetryWhen(IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Employeeperformance Externalmetrics Definition")
		resp, err := gamificationApi.DeleteEmployeeperformanceExternalmetricsDefinition(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Employeeperformance Externalmetrics Definition: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := gamificationApi.GetEmployeeperformanceExternalmetricsDefinition(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Employeeperformance Externalmetrics Definition deleted
				log.Printf("Deleted Employeeperformance Externalmetrics Definition %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting Employeeperformance Externalmetrics Definition %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Employeeperformance Externalmetrics Definition %s still exists", d.Id()))
	})
}
