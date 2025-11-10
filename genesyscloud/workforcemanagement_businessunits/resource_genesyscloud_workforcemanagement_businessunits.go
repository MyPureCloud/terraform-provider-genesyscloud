package workforcemanagement_businessunits

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"log"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"time"

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

// getAllAuthWorkforcemanagementBusinessunits retrieves all the workforcemanagement businessunits via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthWorkforcemanagementBusinessunits(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newWorkforcemanagementBusinessunitsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	businessUnitResponses, resp, err := proxy.getAllWorkforcemanagementBusinessunits(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get workforcemanagement businessunits: %v", err), resp)
	}

	for _, businessUnitResponse := range *businessUnitResponses {
		resources[*businessUnitResponse.Id] = &resourceExporter.ResourceMeta{BlockLabel: *businessUnitResponse.Name}
	}

	return resources, nil
}

// createWorkforcemanagementBusinessUnit is used by the workforcemanagement_businessunits resource to create Genesys cloud workforcemanagement businessunits
func createWorkforcemanagementBusinessUnit(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getWorkforcemanagementBusinessunitsProxy(sdkConfig)

	createBusinessUnitRequest := getCreateWorkforcemanagementBusinessUnitRequestFromResourceData(d)

	log.Printf("Creating workforcemanagement businessunits %s", *createBusinessUnitRequest.Name)
	businessUnitResponse, resp, err := proxy.createWorkforcemanagementBusinessunits(ctx, &createBusinessUnitRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create workforcemanagement businessunits: %s", err), resp)
	}

	d.SetId(*businessUnitResponse.Id)
	log.Printf("Created workforcemanagement businessunits %s", *businessUnitResponse.Id)
	return readWorkforcemanagementBusinessunits(ctx, d, meta)
}

// readWorkforcemanagementBusinessunits is used by the workforcemanagement_businessunits resource to read a workforcemanagement businessunits from genesys cloud
func readWorkforcemanagementBusinessunits(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getWorkforcemanagementBusinessunitsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceWorkforcemanagementBusinessunits(), constants.ConsistencyChecks(), resourceName)

	log.Printf("Reading workforcemanagement businessunits %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		businessUnitResponse, resp, getErr := proxy.getWorkforcemanagementBusinessunitsById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return util.RetryableErrorWithRetryAfter(ctx, util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read workforcemanagement businessunits %s: %s", d.Id(), getErr), resp), resp)
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read workforcemanagement businessunits %s: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", businessUnitResponse.Name)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "settings", businessUnitResponse.Settings, flattenBusinessUnitSettingsResponse)
		resourcedata.SetNillableReferenceWritableDivision(d, "division_id", &platformclientv2.Writabledivision{Id: businessUnitResponse.Division.Id})

		log.Printf("Read workforcemanagement businessunits %s %s", d.Id(), *businessUnitResponse.Name)
		return cc.CheckState(d)
	})
}

// updateWorkforcemanagementBusinessunits is used by the workforcemanagement_businessunits resource to update a workforcemanagement businessunits in Genesys Cloud
func updateWorkforcemanagementBusinessunits(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getWorkforcemanagementBusinessunitsProxy(sdkConfig)

	workforcemanagementBusinessunits := getUpdateWorkforcemanagementBusinessUnitRequestFromResourceData(d)

	log.Printf("Updating workforcemanagement businessunits %s", *workforcemanagementBusinessunits.Name)
	businessUnitResponse, resp, err := proxy.updateWorkforcemanagementBusinessunits(ctx, d.Id(), &workforcemanagementBusinessunits)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update workforcemanagement businessunits %s: %s", d.Id(), err), resp)
	}

	log.Printf("Updated workforcemanagement businessunits %s", *businessUnitResponse.Id)
	return readWorkforcemanagementBusinessunits(ctx, d, meta)
}

// deleteWorkforcemanagementBusinessunits is used by the workforcemanagement_businessunits resource to delete a workforcemanagement businessunits from Genesys cloud
func deleteWorkforcemanagementBusinessunits(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getWorkforcemanagementBusinessunitsProxy(sdkConfig)

	resp, err := proxy.deleteWorkforcemanagementBusinessunits(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete workforcemanagement businessunits %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getWorkforcemanagementBusinessunitsById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted workforcemanagement businessunits %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error deleting workforcemanagement businessunits %s: %s", d.Id(), err), resp))
		}

		return util.RetryableErrorWithRetryAfter(ctx, util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("workforcemanagement businessunits %s still exists", d.Id()), resp), resp)
	})
}
