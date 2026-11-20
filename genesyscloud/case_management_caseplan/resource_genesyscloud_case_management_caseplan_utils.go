package case_management_caseplan

import (
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/mrmo"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

// versionedAttributes need a draft and a publish to change. name and description are not versioned;
// division_id, customer_intent, reference_prefix and data_schema can only change before the first publish.
var versionedAttributes = []string{
	"default_due_duration_in_seconds",
	"default_ttl_seconds",
	"default_case_owner",
	"intake_settings",
	"stageplan",
}

var frozenAfterPublishAttributes = []string{"division_id", "customer_intent", "reference_prefix", "data_schema"}

// isExporting is true for tfexporter and MRMO exports. MRMO only sets mrmo.IsActive().
func isExporting() bool {
	return tfexporter_state.IsExporterActive() || mrmo.IsActive()
}

// getCaseManagementCaseplanCreateFromResourceData maps ResourceData to Caseplancreate for POST /caseplans.
func getCaseManagementCaseplanCreateFromResourceData(d *schema.ResourceData) platformclientv2.Caseplancreate {
	c := platformclientv2.Caseplancreate{}
	c.Name = platformclientv2.String(d.Get("name").(string))
	if v, ok := d.GetOk("division_id"); ok && v.(string) != "" {
		c.DivisionId = platformclientv2.String(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		c.Description = platformclientv2.String(v.(string))
	}
	if v, ok := d.GetOk("reference_prefix"); ok {
		c.ReferencePrefix = platformclientv2.String(v.(string))
	}
	if v, ok := d.GetOk("default_due_duration_in_seconds"); ok {
		c.DefaultDueDurationInSeconds = platformclientv2.Int(v.(int))
	}
	if v, ok := d.GetOk("default_ttl_seconds"); ok {
		c.DefaultTtlSeconds = platformclientv2.Int(v.(int))
	}
	if v, ok := d.GetOk("default_case_owner"); ok {
		if uid := firstMapString(v.([]interface{}), "id"); uid != "" {
			c.DefaultCaseOwnerId = platformclientv2.String(uid)
		}
	}
	if v, ok := d.GetOk("customer_intent"); ok {
		if cid := firstMapString(v.([]interface{}), "id"); cid != "" {
			c.CustomerIntentId = platformclientv2.String(cid)
		}
	}
	if id := configuredDataSchemaID(d); id != "" {
		c.DataSchemas = &[]platformclientv2.Caseplandataschema{{Id: platformclientv2.String(id)}}
	}
	if intake := expandCaseplanIntakeSettings(d.Get("intake_settings").([]interface{})); len(intake) > 0 {
		c.IntakeSettings = &intake
	}
	return c
}

// buildCaseplanUnversionedPatch diffs the caseplan record fields against live. These are not versioned and need no draft;
// the frozen ones are only sent while they differ, which plan-time validation limits to never-published caseplans.
func buildCaseplanUnversionedPatch(d *schema.ResourceData, live *platformclientv2.Caseplan) (*platformclientv2.Caseplanupdate, bool) {
	patch := &platformclientv2.Caseplanupdate{}
	has := false

	if name := d.Get("name").(string); name != stringValue(live.Name) {
		patch.SetField("Name", platformclientv2.String(name))
		has = true
	}
	if desc := d.Get("description").(string); desc != stringValue(live.Description) {
		patch.SetField("Description", nullableString(desc))
		has = true
	}
	liveDivision := ""
	if live.Division != nil {
		liveDivision = stringValue(live.Division.Id)
	}
	if div := d.Get("division_id").(string); div != "" && div != liveDivision {
		patch.SetField("DivisionId", platformclientv2.String(div))
		has = true
	}
	if prefix := d.Get("reference_prefix").(string); prefix != "" && !strings.EqualFold(prefix, stringValue(live.ReferencePrefix)) {
		patch.SetField("ReferencePrefix", platformclientv2.String(prefix))
		has = true
	}
	liveIntent := ""
	if live.CustomerIntent != nil {
		liveIntent = stringValue(live.CustomerIntent.Id)
	}
	if cid := firstMapString(d.Get("customer_intent").([]interface{}), "id"); cid != "" && cid != liveIntent {
		patch.SetField("CustomerIntentId", platformclientv2.String(cid))
		has = true
	}

	if !has {
		return nil, false
	}
	return patch, true
}

// buildCaseplanVersionedPatch diffs the versioned config fields against the live draft.
func buildCaseplanVersionedPatch(d *schema.ResourceData, draft *platformclientv2.Caseplan) (*platformclientv2.Caseplanupdate, bool) {
	patch := &platformclientv2.Caseplanupdate{}
	has := false

	if due := d.Get("default_due_duration_in_seconds").(int); due != 0 && (draft.DefaultDueDurationInSeconds == nil || due != *draft.DefaultDueDurationInSeconds) {
		patch.SetField("DefaultDueDurationInSeconds", platformclientv2.Int(due))
		has = true
	}
	if ttl := d.Get("default_ttl_seconds").(int); ttl != 0 && (draft.DefaultTtlSeconds == nil || ttl != *draft.DefaultTtlSeconds) {
		patch.SetField("DefaultTtlSeconds", platformclientv2.Int(ttl))
		has = true
	}
	liveOwner := ""
	if draft.DefaultCaseOwner != nil {
		liveOwner = stringValue(draft.DefaultCaseOwner.Id)
	}
	if owner := firstMapString(d.Get("default_case_owner").([]interface{}), "id"); owner != liveOwner {
		patch.SetField("DefaultCaseOwnerId", nullableString(owner))
		has = true
	}

	if !has {
		return nil, false
	}
	return patch, true
}

func expandCaseplanIntakeSettings(raw []interface{}) []platformclientv2.Intakesetting {
	out := make([]platformclientv2.Intakesetting, 0, len(raw))
	for _, item := range raw {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		var row platformclientv2.Intakesetting
		if prop, ok := m["property"].(string); ok && prop != "" {
			row.Property = platformclientv2.String(prop)
		}
		if v, ok := m["required"].(bool); ok {
			row.Required = platformclientv2.Bool(v)
		}
		if v, ok := m["display_order"].(int); ok {
			row.DisplayOrder = platformclientv2.Int(v)
		}
		out = append(out, row)
	}
	return out
}

func flattenCaseplanIntakeSettings(entities *[]platformclientv2.Intakesetting) []interface{} {
	if entities == nil || len(*entities) == 0 {
		return []interface{}{}
	}
	out := make([]interface{}, 0, len(*entities))
	for i := range *entities {
		s := &(*entities)[i]
		m := make(map[string]interface{})
		if s.Property != nil {
			m["property"] = *s.Property
		}
		// Always set required/display_order so state matches config and the consistency checker
		// does not see schema defaults (false/0) when the API omits nil pointers.
		required := false
		if s.Required != nil {
			required = *s.Required
		}
		m["required"] = required
		displayOrder := 0
		if s.DisplayOrder != nil {
			displayOrder = *s.DisplayOrder
		}
		m["display_order"] = displayOrder
		out = append(out, m)
	}
	return out
}

// intakeSettingsEqual compares intake settings in order, treating nil required/display_order as their defaults.
func intakeSettingsEqual(desired []platformclientv2.Intakesetting, live *[]platformclientv2.Intakesetting) bool {
	liveFlat := flattenCaseplanIntakeSettings(live)
	desiredFlat := flattenCaseplanIntakeSettings(&desired)
	if len(liveFlat) != len(desiredFlat) {
		return false
	}
	for i := range liveFlat {
		l, dm := liveFlat[i].(map[string]interface{}), desiredFlat[i].(map[string]interface{})
		if l["property"] != dm["property"] || l["required"] != dm["required"] || l["display_order"] != dm["display_order"] {
			return false
		}
	}
	return true
}

func configuredDataSchemaID(d *schema.ResourceData) string {
	return firstMapString(d.Get("data_schema").([]interface{}), "id")
}

func firstDataSchemaID(schemas *[]platformclientv2.Caseplandataschema) string {
	if schemas == nil {
		return ""
	}
	for _, s := range *schemas {
		if id := stringValue(s.Id); id != "" {
			return id
		}
	}
	return ""
}

func flattenCaseplanDataSchemas(schemas *[]platformclientv2.Caseplandataschema) []interface{} {
	if schemas == nil || len(*schemas) == 0 {
		return nil
	}
	out := make([]interface{}, 0, len(*schemas))
	for i := range *schemas {
		s := &(*schemas)[i]
		m := make(map[string]interface{})
		if s.Id != nil {
			m["id"] = *s.Id
		}
		out = append(out, m)
	}
	return out
}

func publishedVersion(cp *platformclientv2.Caseplan) int {
	if cp.Published == nil {
		return 0
	}
	return *cp.Published
}

// hasDraft is true when the latest version is unpublished, including a caseplan that was never published.
func hasDraft(cp *platformclientv2.Caseplan) bool {
	if cp.Published == nil || cp.Latest == nil {
		return true
	}
	return *cp.Latest != *cp.Published
}

// needsNewDraft is true when the latest version is published, so versioned edits need POST /versions first.
func needsNewDraft(cp *platformclientv2.Caseplan) bool {
	return cp.Published != nil && cp.Latest != nil && *cp.Latest == *cp.Published
}

func versionString(v int) string {
	return fmt.Sprintf("%d", v)
}

func firstMapString(blocks []interface{}, key string) string {
	for _, raw := range blocks {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if v, ok := m[key].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

func flattenUserReference(ref *platformclientv2.Userreference) []interface{} {
	if ref == nil {
		return nil
	}
	m := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(m, "id", ref.Id)
	return []interface{}{m}
}

func flattenCustomerIntentReference(ref *platformclientv2.Customerintentreference) []interface{} {
	if ref == nil {
		return nil
	}
	m := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(m, "id", ref.Id)
	return []interface{}{m}
}

// stepplanConfig is the configured (or live) content of a stepplan.
type stepplanConfig struct {
	id           string
	name         string
	description  string
	activityType string
	worktypeID   string
}

// stageplanConfig is the configured (or live) content of a stageplan.
type stageplanConfig struct {
	id          string
	name        string
	description string
	step        stepplanConfig
}

const stageplanFeatureHint = "Adding, removing or reordering stageplans requires the add/delete stageplans feature (PURE-8006) in the org."

func expandStageplans(raw []interface{}) []stageplanConfig {
	out := make([]stageplanConfig, 0, len(raw))
	for _, item := range raw {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		s := stageplanConfig{
			id:          stringFromMap(m, "id"),
			name:        stringFromMap(m, "name"),
			description: stringFromMap(m, "description"),
		}
		if steps, ok := m["stepplan"].([]interface{}); ok && len(steps) > 0 {
			if sm, ok := steps[0].(map[string]interface{}); ok {
				s.step = stepplanConfig{
					id:           stringFromMap(sm, "id"),
					name:         stringFromMap(sm, "name"),
					description:  stringFromMap(sm, "description"),
					activityType: stringFromMap(sm, "activity_type"),
					worktypeID:   firstMapString(listFromMap(sm, "workitem_settings"), "worktype_id"),
				}
			}
		}
		if s.step.activityType == "" {
			s.step.activityType = activityTypeNone
		}
		out = append(out, s)
	}
	return out
}

func flattenStageplans(stages []stageplanConfig) []interface{} {
	out := make([]interface{}, 0, len(stages))
	for _, s := range stages {
		step := map[string]interface{}{
			"id":                s.step.id,
			"name":              s.step.name,
			"description":       s.step.description,
			"activity_type":     s.step.activityType,
			"workitem_settings": []interface{}{},
		}
		if s.step.worktypeID != "" {
			step["workitem_settings"] = []interface{}{map[string]interface{}{"worktype_id": s.step.worktypeID}}
		}
		out = append(out, map[string]interface{}{
			"id":          s.id,
			"name":        s.name,
			"description": s.description,
			"stepplan":    []interface{}{step},
		})
	}
	return out
}

func stageplanFromAPI(s platformclientv2.Stageplan) stageplanConfig {
	return stageplanConfig{
		id:          stringValue(s.Id),
		name:        stringValue(s.Name),
		description: stringValue(s.Description),
	}
}

func stepplanFromAPI(s platformclientv2.Stepplan) stepplanConfig {
	step := stepplanConfig{
		id:           stringValue(s.Id),
		name:         stringValue(s.Name),
		description:  stringValue(s.Description),
		activityType: stringValue(s.ActivityType),
	}
	if step.activityType == "" {
		step.activityType = activityTypeNone
	}
	if s.WorkitemSettings != nil && s.WorkitemSettings.Worktype != nil {
		step.worktypeID = stringValue(s.WorkitemSettings.Worktype.Id)
	}
	return step
}

// matchStageplans resolves each desired stageplan to an existing stageplan id, or "" when it must be created.
// Desired entries match a prior entry by name first, then by index (a rename). Prior ids that no longer exist are ignored.
// It also returns the existing ids that are not matched and must be deleted, in their current order.
func matchStageplans(prior, actual, desired []stageplanConfig) (desiredIDs []string, toDelete []string) {
	actualIDs := make(map[string]bool, len(actual))
	for _, a := range actual {
		actualIDs[a.id] = true
	}
	desiredIDs = make([]string, len(desired))
	usedPrior := make(map[int]bool)
	usedID := make(map[string]bool)
	claim := func(i, j int) bool {
		id := prior[j].id
		if usedPrior[j] || usedID[id] || !actualIDs[id] {
			return false
		}
		desiredIDs[i] = id
		usedPrior[j] = true
		usedID[id] = true
		return true
	}

	for i, ds := range desired {
		for j, ps := range prior {
			if ps.name == ds.name && claim(i, j) {
				break
			}
		}
	}
	for i := range desired {
		if desiredIDs[i] == "" && i < len(prior) {
			claim(i, i)
		}
	}

	for _, a := range actual {
		if !usedID[a.id] {
			toDelete = append(toDelete, a.id)
		}
	}
	return desiredIDs, toDelete
}

func buildStageplanUpdate(current, desired stageplanConfig) (platformclientv2.Stageplanupdate, bool) {
	update := platformclientv2.Stageplanupdate{}
	changed := false
	if current.name != desired.name {
		update.SetField("Name", platformclientv2.String(desired.name))
		changed = true
	}
	if current.description != desired.description {
		update.SetField("Description", nullableString(desired.description))
		changed = true
	}
	return update, changed
}

// buildStepplanUpdate diffs a stepplan against desired. activityType and workitemSettings are always sent together,
// because the API validates that Workitem has workitem settings and None has none.
func buildStepplanUpdate(current, desired stepplanConfig) (platformclientv2.Stepplanupdate, bool) {
	update := platformclientv2.Stepplanupdate{}
	changed := false
	if current.name != desired.name {
		update.SetField("Name", platformclientv2.String(desired.name))
		changed = true
	}
	if current.description != desired.description {
		update.SetField("Description", nullableString(desired.description))
		changed = true
	}
	if current.activityType != desired.activityType || current.worktypeID != desired.worktypeID {
		update.SetField("ActivityType", platformclientv2.String(desired.activityType))
		if desired.worktypeID == "" {
			update.SetField("WorkitemSettings", nil)
		} else {
			update.SetField("WorkitemSettings", &platformclientv2.Workitemsettings{WorktypeId: platformclientv2.String(desired.worktypeID)})
		}
		changed = true
	}
	return update, changed
}

// longestIncreasingSubsequence marks the indices of one longest strictly increasing subsequence of values.
// Stageplans on it are already in the right relative order; moving every other one after its desired
// predecessor, in desired order, reaches the desired order with the fewest reposition calls.
func longestIncreasingSubsequence(values []int) []bool {
	n := len(values)
	length := make([]int, n)
	prev := make([]int, n)
	best := -1
	for i := range values {
		length[i], prev[i] = 1, -1
		for j := 0; j < i; j++ {
			if values[j] < values[i] && length[j]+1 > length[i] {
				length[i], prev[i] = length[j]+1, j
			}
		}
		if best < 0 || length[i] > length[best] {
			best = i
		}
	}
	keep := make([]bool, n)
	for i := best; i >= 0; i = prev[i] {
		keep[i] = true
	}
	return keep
}

func moveAfter(order []string, id, after string) []string {
	out := slices.DeleteFunc(slices.Clone(order), func(v string) bool { return v == id })
	pos := 0
	if after != "" {
		pos = slices.Index(out, after) + 1
	}
	return slices.Insert(out, pos, id)
}

func nullableString(v string) *string {
	if v == "" {
		return nil
	}
	return platformclientv2.String(v)
}

func stringValue(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func stringFromMap(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

func listFromMap(m map[string]interface{}, key string) []interface{} {
	v, _ := m[key].([]interface{})
	return v
}
