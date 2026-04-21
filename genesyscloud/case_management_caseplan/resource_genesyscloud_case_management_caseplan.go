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

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
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
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "customer_intent", caseplan.CustomerIntent, flattenCustomerIntentReference)

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

		if ver != "" {
			intakeListing, inResp, inErr := proxy.getCaseManagementCaseplanVersionIntakesettings(ctx, d.Id(), ver)
			if inErr != nil {
				if util.IsStatus404(inResp) {
					_ = d.Set("intake_settings", []interface{}{})
				} else {
					return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read caseplan %s intake settings: %s", d.Id(), inErr), inResp))
				}
			} else {
				var entities *[]platformclientv2.Intakesetting
				if intakeListing != nil {
					entities = intakeListing.Entities
				}
				_ = d.Set("intake_settings", flattenCaseplanIntakeSettings(entities))
			}
		} else {
			_ = d.Set("intake_settings", []interface{}{})
		}

		log.Printf("Read case management caseplan %s %s", d.Id(), *caseplan.Name)
		return nil
	})
}

func caseplanApplyPatchIfChanged(ctx context.Context, proxy *caseManagementCaseplanProxy, d *schema.ResourceData, id string) diag.Diagnostics {
	if patch, ok := buildCaseplanPatchFromResourceData(d); ok {
		_, resp, err := proxy.patchCaseManagementCaseplan(ctx, id, *patch)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to patch case management caseplan %s: %s", id, err), resp)
		}
	}
	return nil
}

func caseplanDiagsIfImmutableFieldsChangeAfterPublish(ctx context.Context, proxy *caseManagementCaseplanProxy, d *schema.ResourceData, id string) diag.Diagnostics {
	cp, resp, err := proxy.getCaseManagementCaseplanById(ctx, id)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read caseplan %s before update: %s", id, err), resp)
	}
	if cp.Published == nil || *cp.Published == 0 {
		return nil
	}
	var diags diag.Diagnostics
	add := func(attr, detail string) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("%s cannot change after the caseplan has been published", attr),
			Detail:   detail,
		})
	}
	if d.HasChange("division_id") {
		add("division_id", "divisionId is immutable after first publish.")
	}
	if d.HasChange("customer_intent") {
		add("customer_intent", "customerIntentId is immutable after first publish.")
	}
	if d.HasChange("reference_prefix") {
		add("reference_prefix", "referencePrefix is immutable after first publish.")
	}
	if d.HasChange("data_schema") {
		add("data_schema", "dataSchemas are immutable after first publish.")
	}
	if d.HasChange("intake_settings") {
		add("intake_settings", "intakeSettings are immutable after first publish.")
	}
	return diags
}

func caseplanApplyIntakePutIfChanged(ctx context.Context, proxy *caseManagementCaseplanProxy, d *schema.ResourceData, id string) diag.Diagnostics {
	if !d.HasChange("intake_settings") {
		return nil
	}
	body := platformclientv2.Intakesettingsupdate{}
	body.IntakeSettings = expandCaseplanIntakeSettingsForPut(d)
	_, resp, err := proxy.putCaseManagementCaseplanIntakesettings(ctx, id, body)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update caseplan %s intake settings: %s", id, err), resp)
	}
	return nil
}

// execCaseplanDataSchemaSync uses DELETE on .../dataschemas/default when bindings are removed, then for each desired row:
// POST /dataschemas {"id"} for a workitem schema id that was not in the prior config (draft add), or
// PUT .../dataschemas/default with id+version when the id was already present (e.g. version bump).
func execCaseplanDataSchemaSync(ctx context.Context, proxy *caseManagementCaseplanProxy, caseplanID string, oldRaw, newRaw []interface{}) diag.Diagnostics {
	if len(newRaw) > 1 {
		return diag.Errorf("%s: only one data_schema block is supported (API uses .../dataschemas/default); found %d blocks", ResourceType, len(newRaw))
	}
	deleteIDs, puts := caseplanDataSchemaSyncPlanFromState(oldRaw, newRaw)
	if len(deleteIDs) == 0 && len(puts) == 0 {
		return nil
	}
	key := caseplanDataschemaKeyDefault
	oldIDSet := caseplanDataSchemaIDSetFromRaw(oldRaw)
	if len(deleteIDs) > 0 {
		resp, err := proxy.deleteCaseManagementCaseplanDataschema(ctx, caseplanID, key)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete caseplan %s data schema key %q: %s", caseplanID, key, err), resp)
		}
	}
	toPut := puts
	if len(deleteIDs) > 0 && len(toPut) == 0 && len(newRaw) > 0 {
		toPut = caseplanDataSchemasFromResourceList(newRaw)
	}
	for _, row := range toPut {
		if row.Id == nil || *row.Id == "" {
			continue
		}
		sid := *row.Id
		if _, existed := oldIDSet[sid]; existed {
			_, resp, err := proxy.putCaseManagementCaseplanDataschema(ctx, caseplanID, key, row)
			if err != nil {
				return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to put caseplan %s data schema %s (key %q): %s", caseplanID, sid, key, err), resp)
			}
			continue
		}
		_, resp, err := proxy.postCaseManagementCaseplanDataschema(ctx, caseplanID, caseplanDataschemaPostBody{Id: sid})
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to post caseplan %s data schema id %s: %s", caseplanID, sid, err), resp)
		}
	}
	return nil
}

// updateCaseManagementCaseplan is used by the case_management_caseplan resource to update an case management caseplan in Genesys Cloud
func updateCaseManagementCaseplan(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementCaseplanProxy(sdkConfig)
	id := d.Id()

	if diags := caseplanDiagsIfImmutableFieldsChangeAfterPublish(ctx, proxy, d, id); diags != nil {
		return diags
	}

	if diags := caseplanApplyPatchIfChanged(ctx, proxy, d, id); diags != nil {
		return diags
	}

	if d.HasChange("data_schema") {
		oldRaw, newRaw := d.GetChange("data_schema")
		if diags := execCaseplanDataSchemaSync(ctx, proxy, id, oldRaw.([]interface{}), newRaw.([]interface{})); diags != nil {
			return diags
		}
	}

	if diags := caseplanApplyIntakePutIfChanged(ctx, proxy, d, id); diags != nil {
		return diags
	}

	return readCaseManagementCaseplan(ctx, d, meta)
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
