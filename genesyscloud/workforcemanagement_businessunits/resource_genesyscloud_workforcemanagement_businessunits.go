package workforcemanagement_businessunits

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_workforcemanagement_businessunits.go contains all the methods that perform the core logic for a resource.
*/

// getAllAuthWorkforceManagementBusinessUnits retrieves all the workforce management business units via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthWorkforceManagementBusinessUnits(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newWorkforceManagementBusinessUnitsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	businessUnitResponses, resp, err := proxy.getAllWorkforceManagementBusinessUnits(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceName, fmt.Sprintf("Failed to get workforce management business units: %v", err), resp)
	}

	for _, businessUnitResponse := range *businessUnitResponses {
		resources[*businessUnitResponse.Id] = &resourceExporter.ResourceMeta{BlockLabel: *businessUnitResponse.Name}
	}

	return resources, nil
}

// createWorkforceManagementBusinessUnit is used by the workforcemanagement_businessunits resource to create Genesys cloud workforce management business unit
func createWorkforceManagementBusinessUnit(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getWorkforceManagementBusinessUnitsProxy(sdkConfig)

	createBusinessUnitRequest := getCreateWorkforcemanagementBusinessUnitRequestFromResourceData(d)

	log.Printf("Creating workforce management business unit %s", *createBusinessUnitRequest.Name)
	businessUnitResponse, resp, err := proxy.createWorkforceManagementBusinessUnit(ctx, &createBusinessUnitRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceName, fmt.Sprintf("Failed to create workforce management business unit: %s", err), resp)
	}

	d.SetId(*businessUnitResponse.Id)
	log.Printf("Created workforce management business unit %s", *businessUnitResponse.Id)
	return readWorkforceManagementBusinessUnit(ctx, d, meta)
}

// readWorkforceManagementBusinessUnit is used by the workforcemanagement_businessunits resource to read a workforce management business unit from genesys cloud
func readWorkforceManagementBusinessUnit(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getWorkforceManagementBusinessUnitsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceWorkforcemanagementBusinessunits(), constants.ConsistencyChecks(), ResourceName)

	log.Printf("Reading workforce management business unit %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		businessUnitResponse, resp, getErr := proxy.getWorkforceManagementBusinessUnitById(ctx, d.Id())
		if getErr != nil {
			// 404 indicates the resource does not exist and should not be retried
			if util.IsStatus404(resp) {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceName, fmt.Sprintf("Failed to read workforce management business unit %s: %s", d.Id(), getErr), resp))
			}
			// All other errors are also non-retryable
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceName, fmt.Sprintf("Failed to read workforce management business unit %s: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", businessUnitResponse.Name)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "settings", businessUnitResponse.Settings, flattenBusinessUnitSettingsResponse)
		resourcedata.SetNillableReferenceWritableDivision(d, "division_id", &platformclientv2.Writabledivision{Id: businessUnitResponse.Division.Id})

		log.Printf("Read workforce management business unit %s %s", d.Id(), *businessUnitResponse.Name)
		return cc.CheckState(d)
	})
}

// updateWorkforceManagementBusinessUnit is used by the workforcemanagement_businessunits resource to update a workforce management business unit in Genesys Cloud
func updateWorkforceManagementBusinessUnit(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getWorkforceManagementBusinessUnitsProxy(sdkConfig)

	workforceManagementBusinessUnits := getUpdateWorkforcemanagementBusinessUnitRequestFromResourceData(d)

	log.Printf("Updating workforce management business unit %s", *workforceManagementBusinessUnits.Name)
	businessUnitResponse, resp, err := proxy.updateWorkforceManagementBusinessUnit(ctx, d.Id(), &workforceManagementBusinessUnits)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceName, fmt.Sprintf("Failed to update workforce management business unit %s: %s", d.Id(), err), resp)
	}

	log.Printf("Updated workforce management business unit %s", *businessUnitResponse.Id)
	return readWorkforceManagementBusinessUnit(ctx, d, meta)
}

// deleteWorkforceManagementBusinessUnit is used by the workforcemanagement_businessunits resource to delete a workforce management business unit from Genesys cloud
func deleteWorkforceManagementBusinessUnit(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getWorkforceManagementBusinessUnitsProxy(sdkConfig)

	resp, err := proxy.deleteWorkforceManagementBusinessUnit(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceName, fmt.Sprintf("Failed to delete workforce management business unit %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getWorkforceManagementBusinessUnitById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted workforce management business unit %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceName, fmt.Sprintf("Error deleting workforce management business unit %s: %s", d.Id(), err), resp))
		}

		return util.RetryableErrorWithRetryAfter(ctx, util.BuildWithRetriesApiDiagnosticError(ResourceName, fmt.Sprintf("workforce management business unit %s still exists", d.Id()), resp), resp)
	})
}
