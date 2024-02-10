package employeeperformance_externalmetrics_definition

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
	"log"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

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

	domainOrganizationRoles, err := proxy.getAllEmployeeperformanceExternalmetricsDefinition(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get employeeperformance externalmetrics definition: %v", err)
	}

	for _, domainOrganizationRole := range *domainOrganizationRoles {
		resources[*domainOrganizationRole.Id] = &resourceExporter.ResourceMeta{Name: *domainOrganizationRole.Name}
	}

	return resources, nil
}

// createEmployeeperformanceExternalmetricsDefinition is used by the employeeperformance_externalmetrics_definition resource to create Genesys cloud employeeperformance externalmetrics definition
func createEmployeeperformanceExternalmetricsDefinition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getEmployeeperformanceExternalmetricsDefinitionProxy(sdkConfig)

	employeeperformanceExternalmetricsDefinition := getEmployeeperformanceExternalmetricsDefinitionFromResourceData(d)

	log.Printf("Creating employeeperformance externalmetrics definition %s", *employeeperformanceExternalmetricsDefinition.Name)
	domainOrganizationRole, err := proxy.createEmployeeperformanceExternalmetricsDefinition(ctx, &employeeperformanceExternalmetricsDefinition)
	if err != nil {
		return diag.Errorf("Failed to create employeeperformance externalmetrics definition: %s", err)
	}

	d.SetId(*domainOrganizationRole.Id)
	log.Printf("Created employeeperformance externalmetrics definition %s", *domainOrganizationRole.Id)
	return readEmployeeperformanceExternalmetricsDefinition(ctx, d, meta)
}

// readEmployeeperformanceExternalmetricsDefinition is used by the employeeperformance_externalmetrics_definition resource to read an employeeperformance externalmetrics definition from genesys cloud
func readEmployeeperformanceExternalmetricsDefinition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getEmployeeperformanceExternalmetricsDefinitionProxy(sdkConfig)

	log.Printf("Reading employeeperformance externalmetrics definition %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		domainOrganizationRole, respCode, getErr := proxy.getEmployeeperformanceExternalmetricsDefinitionById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read employeeperformance externalmetrics definition %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read employeeperformance externalmetrics definition %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceEmployeeperformanceExternalmetricsDefinition())

		resourcedata.SetNillableValue(d, "name", domainOrganizationRole.Name)
		resourcedata.SetNillableValue(d, "description", domainOrganizationRole.Description)
		resourcedata.SetNillableValue(d, "default_role_id", domainOrganizationRole.DefaultRoleId)
		resourcedata.SetNillableValue(d, "permissions", domainOrganizationRole.Permissions)
		resourcedata.SetNillableValue(d, "unused_permissions", domainOrganizationRole.UnusedPermissions)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "permission_policies", domainOrganizationRole.PermissionPolicies, flattenDomainPermissionPolicys)
		resourcedata.SetNillableValue(d, "user_count", domainOrganizationRole.UserCount)
		resourcedata.SetNillableValue(d, "role_needs_update", domainOrganizationRole.RoleNeedsUpdate)
		resourcedata.SetNillableValue(d, "base", domainOrganizationRole.Base)
		resourcedata.SetNillableValue(d, "default", domainOrganizationRole.Default)

		log.Printf("Read employeeperformance externalmetrics definition %s %s", d.Id(), *domainOrganizationRole.Name)
		return cc.CheckState()
	})
}

// updateEmployeeperformanceExternalmetricsDefinition is used by the employeeperformance_externalmetrics_definition resource to update an employeeperformance externalmetrics definition in Genesys Cloud
func updateEmployeeperformanceExternalmetricsDefinition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getEmployeeperformanceExternalmetricsDefinitionProxy(sdkConfig)

	employeeperformanceExternalmetricsDefinition := getEmployeeperformanceExternalmetricsDefinitionFromResourceData(d)

	log.Printf("Updating employeeperformance externalmetrics definition %s", *employeeperformanceExternalmetricsDefinition.Name)
	domainOrganizationRole, err := proxy.updateEmployeeperformanceExternalmetricsDefinition(ctx, d.Id(), &employeeperformanceExternalmetricsDefinition)
	if err != nil {
		return diag.Errorf("Failed to update employeeperformance externalmetrics definition: %s", err)
	}

	log.Printf("Updated employeeperformance externalmetrics definition %s", *domainOrganizationRole.Id)
	return readEmployeeperformanceExternalmetricsDefinition(ctx, d, meta)
}

// deleteEmployeeperformanceExternalmetricsDefinition is used by the employeeperformance_externalmetrics_definition resource to delete an employeeperformance externalmetrics definition from Genesys cloud
func deleteEmployeeperformanceExternalmetricsDefinition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getEmployeeperformanceExternalmetricsDefinitionProxy(sdkConfig)

	_, err := proxy.deleteEmployeeperformanceExternalmetricsDefinition(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete employeeperformance externalmetrics definition %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getEmployeeperformanceExternalmetricsDefinitionById(ctx, d.Id())

		if err != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				log.Printf("Deleted employeeperformance externalmetrics definition %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting employeeperformance externalmetrics definition %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("employeeperformance externalmetrics definition %s still exists", d.Id()))
	})
}
