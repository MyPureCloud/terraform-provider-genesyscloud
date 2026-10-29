package case_management_caseplan

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

/*
The resource_genesyscloud_case_management_caseplan.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthCaseManagementCaseplans retrieves all published caseplans for the exporter. Never-published caseplans are skipped.
func getAllAuthCaseManagementCaseplans(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getCaseManagementCaseplanProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	caseplans, resp, err := proxy.getAllCaseManagementCaseplan(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get case management caseplan: %v", err), resp)
	}

	for _, caseplan := range *caseplans {
		if caseplan.Id == nil || *caseplan.Id == "" || caseplan.Published == nil {
			continue
		}
		blockLabel := "caseplan"
		if caseplan.Name != nil && *caseplan.Name != "" {
			blockLabel = *caseplan.Name
		}
		resources[*caseplan.Id] = &resourceExporter.ResourceMeta{BlockLabel: blockLabel}
	}

	return resources, nil
}

// createCaseManagementCaseplan creates the caseplan (draft v1 with 3 default stageplans), reconciles stageplans and publishes.
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
	id := *created.Id
	d.SetId(id)

	if desired := expandStageplans(d.Get("stageplan").([]interface{})); len(desired) > 0 {
		resolved, diags := reconcileStageplans(ctx, proxy, id, nil, desired)
		if diags != nil {
			return diags
		}
		_ = d.Set("stageplan", flattenStageplans(resolved))
	}

	if diags := publishCaseplan(ctx, proxy, d, id); diags != nil {
		return diags
	}
	log.Printf("Created and published case management caseplan %s", id)
	return readCaseManagementCaseplan(ctx, d, meta)
}

// readCaseManagementCaseplan reads the published version. A never-published caseplan falls back to latest,
// and is treated as not found during export.
func readCaseManagementCaseplan(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementCaseplanProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceCaseManagementCaseplan(), constants.ConsistencyChecks(), resourceName)

	log.Printf("Reading case management caseplan %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		latest, resp, getErr := proxy.getCaseManagementCaseplanById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read case management caseplan %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read case management caseplan %s: %s", d.Id(), getErr), resp))
		}

		if latest.Published == nil && isExporting() {
			log.Printf("Skipping never-published case management caseplan %s during export", d.Id())
			d.SetId("")
			return nil
		}

		caseplan, version, resp, err := resolveCaseplanContent(ctx, proxy, latest)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read published version of caseplan %s: %s", d.Id(), err), resp))
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
		_ = d.Set("published_version", publishedVersion(latest))
		_ = d.Set("has_draft", hasDraft(latest))

		schemas, dsResp, dsErr := proxy.getCaseManagementCaseplanVersionDataschemas(ctx, d.Id(), version)
		if dsErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read caseplan %s data schemas: %s", d.Id(), dsErr), dsResp))
		}
		if schemas != nil {
			_ = d.Set("data_schema", flattenCaseplanDataSchemas(schemas.Entities))
		} else {
			_ = d.Set("data_schema", nil)
		}

		intake, inResp, inErr := proxy.getCaseManagementCaseplanVersionIntakesettings(ctx, d.Id(), version)
		if inErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read caseplan %s intake settings: %s", d.Id(), inErr), inResp))
		}
		if intake != nil {
			_ = d.Set("intake_settings", flattenCaseplanIntakeSettings(intake.Entities))
		} else {
			_ = d.Set("intake_settings", []interface{}{})
		}

		// Stageplans are only managed when configured (or seeded by import), so caseplans without stageplan blocks show no diff.
		if len(d.Get("stageplan").([]interface{})) > 0 || isExporting() {
			stages, stResp, stErr := readStageplans(ctx, proxy, d.Id(), version)
			if stErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read caseplan %s stageplans: %s", d.Id(), stErr), stResp))
			}
			_ = d.Set("stageplan", flattenStageplans(stages))
		}

		log.Printf("Read case management caseplan %s %s (version %s)", d.Id(), stringValue(caseplan.Name), version)
		return cc.CheckState(d)
	})
}

// resolveCaseplanContent returns the caseplan content Terraform manages and its version id:
// the published version, or latest when the caseplan has never been published.
func resolveCaseplanContent(ctx context.Context, proxy *caseManagementCaseplanProxy, latest *platformclientv2.Caseplan) (*platformclientv2.Caseplan, string, *platformclientv2.APIResponse, error) {
	if latest.Published == nil {
		return latest, caseplanAPIVersionLatest, nil, nil
	}
	version := versionString(*latest.Published)
	if latest.Latest != nil && *latest.Latest == *latest.Published {
		return latest, version, nil, nil
	}
	published, resp, err := proxy.getCaseManagementCaseplanVersion(ctx, stringValue(latest.Id), version)
	if err != nil {
		return nil, version, resp, err
	}
	return published, version, resp, nil
}

// updateCaseManagementCaseplan patches unversioned fields in place. Versioned changes are written over a draft
// (created, or reused if one exists), diffed against that draft because the API rejects no-op writes, then published.
func updateCaseManagementCaseplan(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementCaseplanProxy(sdkConfig)
	id := d.Id()

	live, resp, err := proxy.getCaseManagementCaseplanById(ctx, id)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read caseplan %s before update: %s", id, err), resp)
	}

	if patch, ok := buildCaseplanUnversionedPatch(d, live); ok {
		if _, resp, err := proxy.patchCaseManagementCaseplan(ctx, id, *patch); err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to patch case management caseplan %s: %s", id, err), resp)
		}
	}

	neverPublished := live.Published == nil
	if !neverPublished && !d.HasChanges(versionedAttributes...) {
		return readCaseManagementCaseplan(ctx, d, meta)
	}

	if needsNewDraft(live) {
		log.Printf("Creating draft version of caseplan %s", id)
		if _, resp, err := proxy.postCaseManagementCaseplanVersions(ctx, id); err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create draft version of caseplan %s: %s", id, err), resp)
		}
	}
	draft, resp, err := proxy.getCaseManagementCaseplanVersion(ctx, id, caseplanAPIVersionLatest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read draft of caseplan %s: %s", id, err), resp)
	}

	if patch, ok := buildCaseplanVersionedPatch(d, draft); ok {
		if _, resp, err := proxy.patchCaseManagementCaseplan(ctx, id, *patch); err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to patch draft of caseplan %s: %s", id, err), resp)
		}
	}
	if neverPublished {
		if diags := syncCaseplanDataSchema(ctx, proxy, d, id); diags != nil {
			return diags
		}
	}
	if diags := syncCaseplanIntakeSettings(ctx, proxy, d, id); diags != nil {
		return diags
	}
	// No stageplan blocks means stageplans are not managed, not "delete all" (the minimum is 1).
	if desired := expandStageplans(d.Get("stageplan").([]interface{})); len(desired) > 0 {
		oldRaw, _ := d.GetChange("stageplan")
		resolved, diags := reconcileStageplans(ctx, proxy, id, expandStageplans(oldRaw.([]interface{})), desired)
		if diags != nil {
			return diags
		}
		_ = d.Set("stageplan", flattenStageplans(resolved))
	}

	if diags := publishCaseplan(ctx, proxy, d, id); diags != nil {
		return diags
	}
	return readCaseManagementCaseplan(ctx, d, meta)
}

// publishCaseplan publishes the draft and records the resulting version, so the read that follows starts from the applied values.
func publishCaseplan(ctx context.Context, proxy *caseManagementCaseplanProxy, d *schema.ResourceData, id string) diag.Diagnostics {
	log.Printf("Publishing caseplan %s", id)
	published, resp, err := proxy.publishCaseManagementCaseplan(ctx, id)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish caseplan %s: %s", id, err), resp)
	}
	if published != nil && published.Published != nil {
		_ = d.Set("published_version", *published.Published)
	}
	_ = d.Set("has_draft", false)
	return nil
}

// syncCaseplanDataSchema binds the configured data schema to the draft (only possible before the first publish).
func syncCaseplanDataSchema(ctx context.Context, proxy *caseManagementCaseplanProxy, d *schema.ResourceData, id string) diag.Diagnostics {
	want := configuredDataSchemaID(d)
	listing, resp, err := proxy.getCaseManagementCaseplanVersionDataschemas(ctx, id, caseplanAPIVersionLatest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read data schemas of caseplan %s: %s", id, err), resp)
	}
	var current string
	if listing != nil {
		current = firstDataSchemaID(listing.Entities)
	}
	if want == "" || want == current {
		return nil
	}
	if current == "" {
		if _, resp, err := proxy.postCaseManagementCaseplanDataschema(ctx, id, want); err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to add data schema %s to caseplan %s: %s", want, id, err), resp)
		}
		return nil
	}
	if _, resp, err := proxy.putCaseManagementCaseplanDataschema(ctx, id, caseplanDataschemaKeyDefault, want); err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to replace data schema of caseplan %s with %s (intake settings and workitem stepplans must not reference the old schema): %s", id, want, err), resp)
	}
	return nil
}

// syncCaseplanIntakeSettings replaces the draft's intake settings when they differ from config.
func syncCaseplanIntakeSettings(ctx context.Context, proxy *caseManagementCaseplanProxy, d *schema.ResourceData, id string) diag.Diagnostics {
	desired := expandCaseplanIntakeSettings(d.Get("intake_settings").([]interface{}))
	listing, resp, err := proxy.getCaseManagementCaseplanVersionIntakesettings(ctx, id, caseplanAPIVersionLatest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read intake settings of caseplan %s: %s", id, err), resp)
	}
	var current *[]platformclientv2.Intakesetting
	if listing != nil {
		current = listing.Entities
	}
	if intakeSettingsEqual(desired, current) {
		return nil
	}
	if _, resp, err := proxy.putCaseManagementCaseplanIntakesettings(ctx, id, platformclientv2.Intakesettingsupdate{IntakeSettings: &desired}); err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update intake settings of caseplan %s: %s", id, err), resp)
	}
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

// importCaseManagementCaseplan marks stageplans as managed with a placeholder; the read that follows import
// (and export) replaces it with every stageplan and stepplan.
func importCaseManagementCaseplan(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	placeholder := []interface{}{map[string]interface{}{"name": "", "stepplan": []interface{}{map[string]interface{}{}}}}
	if err := d.Set("stageplan", placeholder); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

// customizeCaseManagementCaseplanDiff validates stageplans, blocks frozen-field changes after publish, and forces an
// update (which publishes) for a caseplan that has never been published.
func customizeCaseManagementCaseplanDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	if err := validateStageplanConfig(d.Get("stageplan").([]interface{})); err != nil {
		return err
	}
	if d.Id() == "" {
		return nil
	}
	published := d.Get("published_version").(int) > 0
	if published {
		for _, attr := range frozenAfterPublishAttributes {
			if d.HasChange(attr) {
				return fmt.Errorf("%s cannot change after the caseplan has been published; create a new caseplan instead", attr)
			}
		}
	}
	if !published || d.HasChanges(versionedAttributes...) {
		if err := d.SetNewComputed("published_version"); err != nil {
			return err
		}
		return d.SetNewComputed("has_draft")
	}
	return nil
}

func validateStageplanConfig(raw []interface{}) error {
	names := make(map[string]bool)
	for i, s := range expandStageplans(raw) {
		if s.name != "" {
			if names[s.name] {
				return fmt.Errorf("stageplan names must be unique within a caseplan; %q is used more than once", s.name)
			}
			names[s.name] = true
		}
		if s.step.activityType == activityTypeWorkitem && !stepHasWorkitemSettings(raw[i]) {
			return fmt.Errorf("stageplan %d (%q): stepplan activity_type %q requires workitem_settings", i+1, s.name, activityTypeWorkitem)
		}
		if s.step.activityType != activityTypeWorkitem && stepHasWorkitemSettings(raw[i]) {
			return fmt.Errorf("stageplan %d (%q): stepplan workitem_settings is only allowed when activity_type is %q", i+1, s.name, activityTypeWorkitem)
		}
	}
	return nil
}

// stepHasWorkitemSettings checks block presence, so an unknown worktype_id at plan time still counts.
func stepHasWorkitemSettings(rawStage interface{}) bool {
	m, ok := rawStage.(map[string]interface{})
	if !ok {
		return false
	}
	steps := listFromMap(m, "stepplan")
	if len(steps) == 0 {
		return false
	}
	sm, ok := steps[0].(map[string]interface{})
	if !ok {
		return false
	}
	return len(listFromMap(sm, "workitem_settings")) > 0
}

// readStageplans returns the stageplans of a caseplan version in process-flow order, each with its single stepplan.
func readStageplans(ctx context.Context, proxy *caseManagementCaseplanProxy, caseplanID, versionID string) ([]stageplanConfig, *platformclientv2.APIResponse, error) {
	stages, resp, err := proxy.listStageplans(ctx, caseplanID, versionID)
	if err != nil {
		return nil, resp, err
	}
	out := make([]stageplanConfig, 0, len(stages))
	for _, s := range stages {
		stage := stageplanFromAPI(s)
		step, stepResp, err := readSingleStepplan(ctx, proxy, caseplanID, versionID, stage.id)
		if err != nil {
			return nil, stepResp, err
		}
		stage.step = step
		out = append(out, stage)
	}
	return out, resp, nil
}

func readSingleStepplan(ctx context.Context, proxy *caseManagementCaseplanProxy, caseplanID, versionID, stageplanID string) (stepplanConfig, *platformclientv2.APIResponse, error) {
	steps, resp, err := proxy.listStepplans(ctx, caseplanID, versionID, stageplanID)
	if err != nil {
		return stepplanConfig{}, resp, err
	}
	if len(steps) != 1 {
		return stepplanConfig{}, resp, fmt.Errorf("expected exactly 1 stepplan for stageplan %s, found %d", stageplanID, len(steps))
	}
	return stepplanFromAPI(steps[0]), resp, nil
}

// reconcileStageplans makes the latest (draft) version's stageplans and stepplans match desired.
// prior is the previous Terraform state; when empty, the live stageplans are used so they are matched by name, then position.
// The stageplan count stays within 1..maxStageplans after every call. It returns desired with stageplan and stepplan ids resolved.
func reconcileStageplans(ctx context.Context, proxy *caseManagementCaseplanProxy, caseplanID string, prior, desired []stageplanConfig) ([]stageplanConfig, diag.Diagnostics) {
	actual, resp, err := readStageplans(ctx, proxy, caseplanID, caseplanAPIVersionLatest)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to list stageplans for caseplan %s: %s", caseplanID, err), resp)
	}
	if len(prior) == 0 {
		prior = actual
	}
	actualByID := make(map[string]stageplanConfig, len(actual))
	for _, a := range actual {
		actualByID[a.id] = a
	}

	desiredIDs, toDelete := matchStageplans(prior, actual, desired)
	created := make(map[int]bool)
	count := len(actual)

	deleteStage := func(id string) diag.Diagnostics {
		log.Printf("Deleting stageplan %s from caseplan %s", id, caseplanID)
		if resp, err := proxy.deleteStageplan(ctx, caseplanID, id); err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete stageplan %s from caseplan %s: %s. %s", id, caseplanID, err, stageplanFeatureHint), resp)
		}
		count--
		return nil
	}

	var deferred []string
	for _, id := range toDelete {
		if count <= 1 {
			deferred = append(deferred, id)
			continue
		}
		if diags := deleteStage(id); diags != nil {
			return nil, diags
		}
	}

	for i, ds := range desired {
		if desiredIDs[i] != "" {
			continue
		}
		if count >= maxStageplans && len(deferred) > 0 {
			if diags := deleteStage(deferred[0]); diags != nil {
				return nil, diags
			}
			deferred = deferred[1:]
		}
		body := platformclientv2.Stageplancreate{Name: platformclientv2.String(ds.name)}
		if ds.description != "" {
			body.Description = platformclientv2.String(ds.description)
		}
		if i > 0 {
			body.After = platformclientv2.String(desiredIDs[i-1])
		}
		log.Printf("Creating stageplan %q in caseplan %s", ds.name, caseplanID)
		stage, resp, err := proxy.createStageplan(ctx, caseplanID, body)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create stageplan %q in caseplan %s: %s. %s", ds.name, caseplanID, err, stageplanFeatureHint), resp)
		}
		if stage == nil || stage.Id == nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Create stageplan %q in caseplan %s returned no id", ds.name, caseplanID), resp)
		}
		desiredIDs[i] = *stage.Id
		created[i] = true
		count++
	}

	for _, id := range deferred {
		if diags := deleteStage(id); diags != nil {
			return nil, diags
		}
	}

	for i, ds := range desired {
		if created[i] {
			continue
		}
		if update, ok := buildStageplanUpdate(actualByID[desiredIDs[i]], ds); ok {
			if _, resp, err := proxy.patchStageplan(ctx, caseplanID, desiredIDs[i], update); err != nil {
				return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update stageplan %s in caseplan %s: %s", desiredIDs[i], caseplanID, err), resp)
			}
		}
	}

	if diags := repositionStageplans(ctx, proxy, caseplanID, desiredIDs); diags != nil {
		return nil, diags
	}

	resolved := make([]stageplanConfig, len(desired))
	for i, ds := range desired {
		stageID := desiredIDs[i]
		var current stepplanConfig
		if created[i] {
			step, resp, err := readSingleStepplan(ctx, proxy, caseplanID, caseplanAPIVersionLatest, stageID)
			if err != nil {
				return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read stepplan of stageplan %s in caseplan %s: %s", stageID, caseplanID, err), resp)
			}
			current = step
		} else {
			current = actualByID[stageID].step
		}
		resolved[i] = ds
		resolved[i].id = stageID
		resolved[i].step.id = current.id
		if update, ok := buildStepplanUpdate(current, ds.step); ok {
			if _, resp, err := proxy.patchStepplan(ctx, caseplanID, stageID, current.id, update); err != nil {
				return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update stepplan %s of stageplan %s in caseplan %s: %s", current.id, stageID, caseplanID, err), resp)
			}
		}
	}
	return resolved, nil
}

// repositionStageplans moves stageplans until the live order equals desiredIDs. Each move fixes one position, left to right.
func repositionStageplans(ctx context.Context, proxy *caseManagementCaseplanProxy, caseplanID string, desiredIDs []string) diag.Diagnostics {
	stages, resp, err := proxy.listStageplans(ctx, caseplanID, caseplanAPIVersionLatest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to list stageplans for caseplan %s: %s", caseplanID, err), resp)
	}
	order := make([]string, 0, len(stages))
	for _, s := range stages {
		order = append(order, stringValue(s.Id))
	}
	if len(order) != len(desiredIDs) {
		return diag.Errorf("%s: caseplan %s has %d stageplans after reconciling, expected %d", ResourceType, caseplanID, len(order), len(desiredIDs))
	}

	positions := make([]int, len(desiredIDs))
	for i, id := range desiredIDs {
		positions[i] = slices.Index(order, id)
		if positions[i] < 0 {
			return diag.Errorf("%s: stageplan %s not found in caseplan %s after reconciling", ResourceType, id, caseplanID)
		}
	}
	keep := longestIncreasingSubsequence(positions)

	for i, id := range desiredIDs {
		if keep[i] {
			continue
		}
		pos := slices.Index(order, id)
		currentAfter, wantAfter := "", ""
		if pos > 0 {
			currentAfter = order[pos-1]
		}
		if i > 0 {
			wantAfter = desiredIDs[i-1]
		}
		if currentAfter == wantAfter {
			continue
		}
		body := platformclientv2.Stageplanreposition{}
		if wantAfter == "" {
			body.SetField("After", nil)
		} else {
			body.SetField("After", platformclientv2.String(wantAfter))
		}
		log.Printf("Repositioning stageplan %s in caseplan %s after %q", id, caseplanID, wantAfter)
		if resp, err := proxy.repositionStageplan(ctx, caseplanID, id, body); err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to reposition stageplan %s in caseplan %s: %s. %s", id, caseplanID, err, stageplanFeatureHint), resp)
		}
		order = moveAfter(order, id, wantAfter)
	}
	return nil
}
