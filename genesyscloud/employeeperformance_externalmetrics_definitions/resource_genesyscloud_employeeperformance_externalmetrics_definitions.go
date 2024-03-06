package employeeperformance_externalmetrics_definitions

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v123/platformclientv2"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_employeeperformance_externalmetrics_definition.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthEmployeeperformanceExternalmetricsDefinition retrieves all of the employeeperformance externalmetrics definition via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthEmployeeperformanceExternalmetricsDefinitions(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newEmployeeperformanceExternalmetricsDefinitionProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	definitions, err := proxy.getAllEmployeeperformanceExternalmetricsDefinition(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get employeeperformance externalmetrics definition: %v", err)
	}

	for _, definition := range *definitions {
		resources[*definition.Id] = &resourceExporter.ResourceMeta{Name: *definition.Name}
	}

	return resources, nil
}

// createEmployeeperformanceExternalmetricsDefinition is used by the employeeperformance_externalmetrics_definition resource to create Genesys cloud employeeperformance externalmetrics definition
func createEmployeeperformanceExternalmetricsDefinition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getEmployeeperformanceExternalmetricsDefinitionProxy(sdkConfig)

	metricDefinition := platformclientv2.Externalmetricdefinitioncreaterequest{
		Name:                 platformclientv2.String(d.Get("name").(string)),
		Unit:                 platformclientv2.String(d.Get("unit").(string)),
		Enabled:              platformclientv2.Bool(d.Get("enabled").(bool)),
		Precision:            platformclientv2.Int(d.Get("precision").(int)),
		DefaultObjectiveType: platformclientv2.String(d.Get("default_objective_type").(string)),
	}

	unitDefinition := d.Get("unit_definition").(string)
	if unitDefinition != "" {
		metricDefinition.UnitDefinition = &unitDefinition
	}

	log.Printf("Creating employeeperformance externalmetrics definition %s", *metricDefinition.Name)
	definition, err := proxy.createEmployeeperformanceExternalmetricsDefinition(ctx, &metricDefinition)
	if err != nil {
		return diag.Errorf("Failed to create employeeperformance externalmetrics definition %s: %s", *metricDefinition.Name, err)
	}

	d.SetId(*definition.Id)
	log.Printf("Created employeeperformance externalmetrics definition %s: %s", *definition.Name, *definition.Id)
	return readEmployeeperformanceExternalmetricsDefinition(ctx, d, meta)
}

// readEmployeeperformanceExternalmetricsDefinition is used by the employeeperformance_externalmetrics_definition resource to read an employeeperformance externalmetrics definition from genesys cloud
func readEmployeeperformanceExternalmetricsDefinition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getEmployeeperformanceExternalmetricsDefinitionProxy(sdkConfig)

	log.Printf("Reading employeeperformance externalmetrics definition %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		definition, respCode, getErr := proxy.getEmployeeperformanceExternalmetricsDefinitionById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read employeeperformance externalmetrics definition %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read employeeperformance externalmetrics definition %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceEmployeeperformanceExternalmetricsDefinition())

		resourcedata.SetNillableValue(d, "name", definition.Name)
		resourcedata.SetNillableValue(d, "precision", definition.Precision)
		resourcedata.SetNillableValue(d, "default_objective_type", definition.DefaultObjectiveType)
		resourcedata.SetNillableValue(d, "enabled", definition.Enabled)
		resourcedata.SetNillableValue(d, "unit", definition.Unit)
		resourcedata.SetNillableValue(d, "unit_definition", definition.UnitDefinition)

		log.Printf("Read employeeperformance externalmetrics definition %s %s", d.Id(), *definition.Name)
		return cc.CheckState()
	})
}

// updateEmployeeperformanceExternalmetricsDefinition is used by the employeeperformance_externalmetrics_definition resource to update an employeeperformance externalmetrics definition in Genesys Cloud
func updateEmployeeperformanceExternalmetricsDefinition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getEmployeeperformanceExternalmetricsDefinitionProxy(sdkConfig)

	metricDefinition := platformclientv2.Externalmetricdefinitionupdaterequest{
		Name:                 platformclientv2.String(d.Get("name").(string)),
		Enabled:              platformclientv2.Bool(d.Get("enabled").(bool)),
		Precision:            platformclientv2.Int(d.Get("precision").(int)),
		DefaultObjectiveType: platformclientv2.String(d.Get("default_objective_type").(string)),
	}

	log.Printf("Updating employeeperformance externalmetrics definition %s: %s", *metricDefinition.Name, d.Id())
	definition, err := proxy.updateEmployeeperformanceExternalmetricsDefinition(ctx, d.Id(), &metricDefinition)
	if err != nil {
		return diag.Errorf("Failed to update employeeperformance externalmetrics definition: %s", err)
	}

	log.Printf("Updated employeeperformance externalmetrics definition %s", *definition.Id)
	return readEmployeeperformanceExternalmetricsDefinition(ctx, d, meta)
}

// deleteEmployeeperformanceExternalmetricsDefinition is used by the employeeperformance_externalmetrics_definition resource to delete an employeeperformance externalmetrics definition from Genesys cloud
func deleteEmployeeperformanceExternalmetricsDefinition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getEmployeeperformanceExternalmetricsDefinitionProxy(sdkConfig)

	_, err := proxy.deleteEmployeeperformanceExternalmetricsDefinition(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete employeeperformance externalmetrics definition %s: %s", d.Id(), err)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getEmployeeperformanceExternalmetricsDefinitionById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404ByInt(respCode) {
				log.Printf("Deleted employeeperformance externalmetrics definition %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting employeeperformance externalmetrics definition %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("employeeperformance externalmetrics definition %s still exists", d.Id()))
	})
}
