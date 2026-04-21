package case_management_caseplan

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_case_management_caseplan.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthCaseManagementCaseplan retrieves all of the case management caseplan via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthCaseManagementCaseplans(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newCaseManagementCaseplanProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	caseplans, resp, err := proxy.getAllCaseManagementCaseplan(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get case management caseplan: %v", err), resp)
	}

	for _, caseplan := range *caseplans {
		resources[*caseplan.Id] = &resourceExporter.ResourceMeta{BlockLabel: *caseplan.Name}
	}

	return resources, nil
}

// createCaseManagementCaseplan is used by the case_management_caseplan resource to create Genesys cloud case management caseplan
func createCaseManagementCaseplan(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementCaseplanProxy(sdkConfig)

	body := getCaseManagementCaseplanCreateFromResourceData(d)

	log.Printf("Creating case management caseplan")
	created, resp, err := proxy.createCaseManagementCaseplan(ctx, &body)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create case management caseplan: %s", err), resp)
	}
	if created == nil || created.Id == nil {
		return util.BuildAPIDiagnosticError(ResourceType, "Create caseplan returned no id", resp)
	}

	d.SetId(*created.Id)
	log.Printf("Created case management caseplan %s", *created.Id)
	return readCaseManagementCaseplan(ctx, d, meta)
}

// readCaseManagementCaseplan is used by the case_management_caseplan resource to read an case management caseplan from genesys cloud
func readCaseManagementCaseplan(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementCaseplanProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceCaseManagementCaseplan(), constants.ConsistencyChecks(), resourceName)

	log.Printf("Reading case management caseplan %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		caseplan, resp, getErr := proxy.getCaseManagementCaseplanById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read case management caseplan %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read case management caseplan %s: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", caseplan.Name)
		if caseplan.Division != nil {
			resourcedata.SetNillableValue(d, "division_id", caseplan.Division.Id)
		} else {
			_ = d.Set("division_id", nil)
		}
		resourcedata.SetNillableValue(d, "description", caseplan.Description)
		resourcedata.SetNillableValue(d, "reference_prefix", caseplan.ReferencePrefix)
		resourcedata.SetNillableValue(d, "default_due_duration_in_seconds", caseplan.DefaultDueDurationInSeconds)
		resourcedata.SetNillableValue(d, "default_ttl_seconds", caseplan.DefaultTtlSeconds)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "default_case_owner", caseplan.DefaultCaseOwner, flattenUserReference)
		resourcedata.SetNillableValue(d, "latest", caseplan.Latest)
		resourcedata.SetNillableValue(d, "published", caseplan.Published)
		resourcedata.SetNillableTime(d, "date_published", caseplan.DatePublished)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "customer_intent", caseplan.CustomerIntent, flattenCustomerIntentReference)
		resourcedata.SetNillableValue(d, "version_state", caseplan.VersionState)

		ver := caseplanVersionForDataschemaRead(caseplan)
		if ver != "" {
			listing, dsResp, dsErr := proxy.getCaseManagementCaseplanVersionDataschemas(ctx, d.Id(), ver)
			if dsErr != nil {
				if util.IsStatus404(dsResp) {
					_ = d.Set("data_schema", nil)
				} else {
					return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read caseplan %s data schemas: %s", d.Id(), dsErr), dsResp))
				}
			} else {
				var entities *[]platformclientv2.Caseplandataschema
				if listing != nil {
					entities = listing.Entities
				}
				_ = d.Set("data_schema", flattenCaseplanDataSchemas(entities))
			}
		} else {
			_ = d.Set("data_schema", nil)
		}

		log.Printf("Read case management caseplan %s %s", d.Id(), *caseplan.Name)
		return cc.CheckState(d)
	})
}

// updateCaseManagementCaseplan is used by the case_management_caseplan resource to update an case management caseplan in Genesys Cloud
func updateCaseManagementCaseplan(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

// deleteCaseManagementCaseplan is used by the case_management_caseplan resource to delete an case management caseplan from Genesys cloud
func deleteCaseManagementCaseplan(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementCaseplanProxy(sdkConfig)

	resp, err := proxy.deleteCaseManagementCaseplan(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete case management caseplan %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getCaseManagementCaseplanById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted case management caseplan %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting case management caseplan %s: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("case management caseplan %s still exists", d.Id()), resp))
	})
}
