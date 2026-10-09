package case_management_caseplan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/mrmo"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
)

func TestUnitFlattenCaseplanDataSchemas(t *testing.T) {
	t.Parallel()
	id1 := "11111111-1111-1111-1111-111111111111"
	flat := flattenCaseplanDataSchemas(&[]platformclientv2.Caseplandataschema{{Id: platformclientv2.String(id1)}})
	assert.Len(t, flat, 1)
	assert.Equal(t, id1, flat[0].(map[string]interface{})["id"])

	assert.Nil(t, flattenCaseplanDataSchemas(nil))
	assert.Nil(t, flattenCaseplanDataSchemas(&[]platformclientv2.Caseplandataschema{}))
}

func TestUnitGetCaseManagementCaseplanCreateFromResourceData(t *testing.T) {
	t.Parallel()
	sch := ResourceCaseManagementCaseplan().Schema
	d := schema.TestResourceDataRaw(t, sch, map[string]interface{}{
		"name":                            "cp-name",
		"division_id":                     "div-1",
		"description":                     "desc",
		"reference_prefix":                "AB12",
		"default_due_duration_in_seconds": 100,
		"default_ttl_seconds":             200,
		"customer_intent":                 []interface{}{map[string]interface{}{"id": "intent-1"}},
		"default_case_owner":              []interface{}{map[string]interface{}{"id": "user-1"}},
		"data_schema":                     []interface{}{map[string]interface{}{"id": "schema-1"}},
		"intake_settings": []interface{}{
			map[string]interface{}{"property": "case_note_text", "required": true, "display_order": 1},
		},
	})

	body := getCaseManagementCaseplanCreateFromResourceData(d)
	assert.Equal(t, "cp-name", *body.Name)
	assert.Equal(t, "div-1", *body.DivisionId)
	assert.Equal(t, "desc", *body.Description)
	assert.Equal(t, "AB12", *body.ReferencePrefix)
	assert.Equal(t, 100, *body.DefaultDueDurationInSeconds)
	assert.Equal(t, 200, *body.DefaultTtlSeconds)
	assert.Equal(t, "intent-1", *body.CustomerIntentId)
	assert.Equal(t, "user-1", *body.DefaultCaseOwnerId)
	require.NotNil(t, body.DataSchemas)
	assert.Equal(t, "schema-1", *(*body.DataSchemas)[0].Id)
	require.NotNil(t, body.IntakeSettings)
	assert.Equal(t, "case_note_text", *(*body.IntakeSettings)[0].Property)
	assert.True(t, *(*body.IntakeSettings)[0].Required)
	assert.Equal(t, 1, *(*body.IntakeSettings)[0].DisplayOrder)
}

func TestUnitFlattenExpandCaseplanIntakeSettings(t *testing.T) {
	t.Parallel()
	flat := flattenCaseplanIntakeSettings(&[]platformclientv2.Intakesetting{
		{Property: platformclientv2.String("p1"), Required: platformclientv2.Bool(true), DisplayOrder: platformclientv2.Int(2)},
	})
	assert.Equal(t, []interface{}{map[string]interface{}{"property": "p1", "required": true, "display_order": 2}}, flat)
	assert.Len(t, flattenCaseplanIntakeSettings(nil), 0)

	expanded := expandCaseplanIntakeSettings(flat)
	assert.True(t, intakeSettingsEqual(expanded, &[]platformclientv2.Intakesetting{
		{Property: platformclientv2.String("p1"), Required: platformclientv2.Bool(true), DisplayOrder: platformclientv2.Int(2)},
	}))
	assert.False(t, intakeSettingsEqual(expanded, nil))
	assert.True(t, intakeSettingsEqual(nil, &[]platformclientv2.Intakesetting{}))
	assert.True(t, intakeSettingsEqual(
		[]platformclientv2.Intakesetting{{Property: platformclientv2.String("p"), Required: platformclientv2.Bool(false), DisplayOrder: platformclientv2.Int(0)}},
		&[]platformclientv2.Intakesetting{{Property: platformclientv2.String("p")}},
	))
}

func TestUnitFlattenUserAndIntentRefs(t *testing.T) {
	t.Parallel()
	assert.Nil(t, flattenUserReference(nil))
	assert.Nil(t, flattenCustomerIntentReference(nil))
	assert.Equal(t, "u-1", flattenUserReference(&platformclientv2.Userreference{Id: platformclientv2.String("u-1")})[0].(map[string]interface{})["id"])
	assert.Equal(t, "i-1", flattenCustomerIntentReference(&platformclientv2.Customerintentreference{Id: platformclientv2.String("i-1")})[0].(map[string]interface{})["id"])
}

func TestUnitBuildCaseplanUnversionedPatch(t *testing.T) {
	t.Parallel()
	sch := ResourceCaseManagementCaseplan().Schema
	live := &platformclientv2.Caseplan{
		Name:            platformclientv2.String("cp"),
		Description:     platformclientv2.String("old"),
		ReferencePrefix: platformclientv2.String("AB12"),
		Division:        &platformclientv2.Starrabledivision{Id: platformclientv2.String("div-1")},
		CustomerIntent:  &platformclientv2.Customerintentreference{Id: platformclientv2.String("intent-1")},
	}

	d := schema.TestResourceDataRaw(t, sch, map[string]interface{}{
		"name":             "cp",
		"description":      "old",
		"reference_prefix": "ab12",
		"division_id":      "div-1",
		"customer_intent":  []interface{}{map[string]interface{}{"id": "intent-1"}},
	})
	_, changed := buildCaseplanUnversionedPatch(d, live)
	assert.False(t, changed, "unchanged values (prefix compared case-insensitively) must not be sent")

	d = schema.TestResourceDataRaw(t, sch, map[string]interface{}{"name": "renamed"})
	patch, changed := buildCaseplanUnversionedPatch(d, live)
	require.True(t, changed)
	body, err := json.Marshal(patch)
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"renamed","description":null}`, string(body))
}

func TestUnitBuildCaseplanVersionedPatch(t *testing.T) {
	t.Parallel()
	sch := ResourceCaseManagementCaseplan().Schema
	draft := &platformclientv2.Caseplan{
		DefaultDueDurationInSeconds: platformclientv2.Int(100),
		DefaultTtlSeconds:           platformclientv2.Int(200),
		DefaultCaseOwner:            &platformclientv2.Userreference{Id: platformclientv2.String("user-1")},
	}

	d := schema.TestResourceDataRaw(t, sch, map[string]interface{}{
		"default_due_duration_in_seconds": 100,
		"default_ttl_seconds":             200,
		"default_case_owner":              []interface{}{map[string]interface{}{"id": "user-1"}},
	})
	_, changed := buildCaseplanVersionedPatch(d, draft)
	assert.False(t, changed)

	d = schema.TestResourceDataRaw(t, sch, map[string]interface{}{"default_due_duration_in_seconds": 150})
	patch, changed := buildCaseplanVersionedPatch(d, draft)
	require.True(t, changed)
	body, err := json.Marshal(patch)
	require.NoError(t, err)
	assert.JSONEq(t, `{"defaultDueDurationInSeconds":150,"defaultCaseOwnerId":null}`, string(body))
}

func TestUnitValidateStageplanConfig(t *testing.T) {
	t.Parallel()
	block := func(name, activity string, withSettings bool) interface{} {
		step := map[string]interface{}{"name": "step", "activity_type": activity, "workitem_settings": []interface{}{}}
		if withSettings {
			step["workitem_settings"] = []interface{}{map[string]interface{}{"worktype_id": "wt"}}
		}
		return map[string]interface{}{"name": name, "stepplan": []interface{}{step}}
	}

	assert.NoError(t, validateStageplanConfig([]interface{}{block("A", activityTypeWorkitem, true), block("B", activityTypeNone, false)}))
	assert.ErrorContains(t, validateStageplanConfig([]interface{}{block("A", activityTypeNone, false), block("A", activityTypeNone, false)}), "unique")
	assert.ErrorContains(t, validateStageplanConfig([]interface{}{block("A", activityTypeWorkitem, false)}), "requires workitem_settings")
	assert.ErrorContains(t, validateStageplanConfig([]interface{}{block("A", activityTypeNone, true)}), "only allowed")
}

func TestUnitHasDraftAndNeedsNewDraft(t *testing.T) {
	t.Parallel()
	neverPublished := &platformclientv2.Caseplan{Latest: platformclientv2.Int(1)}
	published := &platformclientv2.Caseplan{Latest: platformclientv2.Int(2), Published: platformclientv2.Int(2)}
	withDraft := &platformclientv2.Caseplan{Latest: platformclientv2.Int(3), Published: platformclientv2.Int(2)}

	assert.True(t, hasDraft(neverPublished))
	assert.False(t, hasDraft(published))
	assert.True(t, hasDraft(withDraft))
	assert.False(t, needsNewDraft(neverPublished))
	assert.True(t, needsNewDraft(published))
	assert.False(t, needsNewDraft(withDraft))
	assert.Equal(t, 0, publishedVersion(neverPublished))
	assert.Equal(t, 2, publishedVersion(withDraft))
}

// The update tests share internalProxy, so they must not run in parallel.

func TestUnitCaseplanSchemaShape(t *testing.T) {
	t.Parallel()
	r := ResourceCaseManagementCaseplan()
	s := r.Schema

	stage := s["stageplan"]
	require.NotNil(t, stage)
	assert.Equal(t, schema.TypeList, stage.Type, "stageplans are ordered")
	assert.Equal(t, 1, stage.MinItems)
	assert.Equal(t, maxStageplans, stage.MaxItems)
	step := stage.Elem.(*schema.Resource).Schema["stepplan"]
	require.NotNil(t, step)
	assert.Equal(t, schema.TypeList, step.Type)
	assert.Equal(t, 1, step.MinItems)
	assert.Equal(t, 1, step.MaxItems)

	assert.True(t, s["data_schema"].Required, "data_schema is required")
	assert.True(t, s["published_version"].Computed)
	assert.False(t, s["published_version"].Optional)
	assert.True(t, s["has_draft"].Computed)
	assert.False(t, s["has_draft"].Optional)
	for _, flag := range []string{"publish", "auto_publish", "revision"} {
		assert.NotContains(t, s, flag, "publishing is implicit; no %s flag", flag)
	}

	require.NotNil(t, r.Importer)
	assert.NotNil(t, r.Importer.StateContext, "custom importer")
	assert.NotNil(t, r.CustomizeDiff)
}

// publishedCaseplanState is the flat state of a published caseplan with one None stageplan.
func publishedCaseplanState(publishedVersion string) map[string]string {
	return map[string]string{
		"id":                                         "cp1",
		"name":                                       "cp",
		"division_id":                                "div-1",
		"reference_prefix":                           "AB12",
		"default_due_duration_in_seconds":            "100",
		"default_ttl_seconds":                        "200",
		"customer_intent.#":                          "1",
		"customer_intent.0.id":                       "intent-1",
		"data_schema.#":                              "1",
		"data_schema.0.id":                           "schema-1",
		"intake_settings.#":                          "0",
		"default_case_owner.#":                       "0",
		"published_version":                          publishedVersion,
		"has_draft":                                  "false",
		"stageplan.#":                                "1",
		"stageplan.0.id":                             "stage-1",
		"stageplan.0.name":                           "A",
		"stageplan.0.description":                    "",
		"stageplan.0.stepplan.#":                     "1",
		"stageplan.0.stepplan.0.id":                  "step-1",
		"stageplan.0.stepplan.0.name":                "Step",
		"stageplan.0.stepplan.0.description":         "",
		"stageplan.0.stepplan.0.activity_type":       activityTypeNone,
		"stageplan.0.stepplan.0.workitem_settings.#": "0",
	}
}

func caseplanConfig(edit func(map[string]interface{})) *terraform.ResourceConfig {
	raw := map[string]interface{}{
		"name":                            "cp",
		"division_id":                     "div-1",
		"reference_prefix":                "AB12",
		"default_due_duration_in_seconds": 100,
		"default_ttl_seconds":             200,
		"customer_intent":                 []interface{}{map[string]interface{}{"id": "intent-1"}},
		"data_schema":                     []interface{}{map[string]interface{}{"id": "schema-1"}},
		"stageplan": []interface{}{map[string]interface{}{
			"name":     "A",
			"stepplan": []interface{}{map[string]interface{}{"name": "Step", "activity_type": activityTypeNone}},
		}},
	}
	if edit != nil {
		edit(raw)
	}
	return terraform.NewResourceConfigRaw(raw)
}

func planCaseplan(t *testing.T, state map[string]string, cfg *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
	t.Helper()
	var s *terraform.InstanceState
	if state != nil {
		s = &terraform.InstanceState{ID: state["id"], Attributes: state}
	}
	return ResourceCaseManagementCaseplan().Diff(context.Background(), s, cfg, nil)
}

func forcesPublish(d *terraform.InstanceDiff) bool {
	if d == nil {
		return false
	}
	a := d.Attributes["published_version"]
	return a != nil && a.NewComputed
}

func TestUnitCaseplanPlan_frozenAttributesBlockedAfterPublish(t *testing.T) {
	t.Parallel()
	edits := map[string]func(map[string]interface{}){
		"division_id":      func(m map[string]interface{}) { m["division_id"] = "div-2" },
		"reference_prefix": func(m map[string]interface{}) { m["reference_prefix"] = "ZZ99" },
		"customer_intent": func(m map[string]interface{}) {
			m["customer_intent"] = []interface{}{map[string]interface{}{"id": "intent-2"}}
		},
		"data_schema": func(m map[string]interface{}) {
			m["data_schema"] = []interface{}{map[string]interface{}{"id": "schema-2"}}
		},
	}
	for attr, edit := range edits {
		t.Run(attr+" blocked when published", func(t *testing.T) {
			_, err := planCaseplan(t, publishedCaseplanState("1"), caseplanConfig(edit))
			require.Error(t, err)
			assert.Contains(t, err.Error(), attr+" cannot change after the caseplan has been published")
		})
		t.Run(attr+" allowed before first publish", func(t *testing.T) {
			diff, err := planCaseplan(t, publishedCaseplanState("0"), caseplanConfig(edit))
			require.NoError(t, err)
			assert.True(t, forcesPublish(diff))
		})
	}
}

func TestUnitCaseplanPlan_publishedVersionReflectsWhatPublishes(t *testing.T) {
	t.Parallel()

	t.Run("no change is an empty plan", func(t *testing.T) {
		diff, err := planCaseplan(t, publishedCaseplanState("1"), caseplanConfig(nil))
		require.NoError(t, err)
		assert.True(t, diff.Empty(), "unexpected diff: %#v", diff)
	})

	t.Run("name and description do not publish", func(t *testing.T) {
		diff, err := planCaseplan(t, publishedCaseplanState("1"), caseplanConfig(func(m map[string]interface{}) {
			m["name"] = "renamed"
			m["description"] = "new"
		}))
		require.NoError(t, err)
		require.NotNil(t, diff.Attributes["name"])
		assert.False(t, forcesPublish(diff))
	})

	versioned := map[string]func(map[string]interface{}){
		"stageplan rename": func(m map[string]interface{}) {
			m["stageplan"].([]interface{})[0].(map[string]interface{})["name"] = "B"
		},
		"stageplan added": func(m map[string]interface{}) {
			m["stageplan"] = append(m["stageplan"].([]interface{}), map[string]interface{}{
				"name":     "C",
				"stepplan": []interface{}{map[string]interface{}{"name": "Step 2", "activity_type": activityTypeNone}},
			})
		},
		"due duration": func(m map[string]interface{}) { m["default_due_duration_in_seconds"] = 150 },
		"intake": func(m map[string]interface{}) {
			m["intake_settings"] = []interface{}{map[string]interface{}{"property": "p"}}
		},
		"default owner": func(m map[string]interface{}) {
			m["default_case_owner"] = []interface{}{map[string]interface{}{"id": "user-1"}}
		},
	}
	for name, edit := range versioned {
		t.Run(name+" publishes", func(t *testing.T) {
			diff, err := planCaseplan(t, publishedCaseplanState("1"), caseplanConfig(edit))
			require.NoError(t, err)
			assert.True(t, forcesPublish(diff))
			assert.True(t, diff.Attributes["has_draft"] != nil && diff.Attributes["has_draft"].NewComputed)
		})
	}

	t.Run("never published forces an update with no config change", func(t *testing.T) {
		diff, err := planCaseplan(t, publishedCaseplanState("0"), caseplanConfig(nil))
		require.NoError(t, err)
		assert.True(t, forcesPublish(diff))
	})
}

func TestUnitCaseplanPlan_stageplanValidation(t *testing.T) {
	t.Parallel()
	_, err := planCaseplan(t, nil, caseplanConfig(func(m map[string]interface{}) {
		m["stageplan"].([]interface{})[0].(map[string]interface{})["stepplan"] = []interface{}{
			map[string]interface{}{"name": "Step", "activity_type": activityTypeWorkitem},
		}
	}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires workitem_settings")

	_, err = planCaseplan(t, nil, caseplanConfig(func(m map[string]interface{}) {
		s := m["stageplan"].([]interface{})[0]
		m["stageplan"] = []interface{}{s, s}
	}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unique")
}

func TestUnitBuildCaseplanUnversionedPatch_omittedDivisionIsNotSent(t *testing.T) {
	t.Parallel()
	live := &platformclientv2.Caseplan{
		Name:     platformclientv2.String("cp"),
		Division: &platformclientv2.Starrabledivision{Id: platformclientv2.String("div-1")},
	}
	d := schema.TestResourceDataRaw(t, ResourceCaseManagementCaseplan().Schema, map[string]interface{}{"name": "cp"})
	patch, changed := buildCaseplanUnversionedPatch(d, live)
	assert.False(t, changed, "omitting division_id must not move the caseplan to division *")
	assert.False(t, patch != nil && patch.SetFieldNames["DivisionId"])
}

func TestUnitCaseplanExporter_refsAndExclusions(t *testing.T) {
	t.Parallel()
	e := CaseManagementCaseplanExporter()

	want := map[string]string{
		"division_id":           "genesyscloud_auth_division",
		"default_case_owner.id": "genesyscloud_user",
		"customer_intent.id":    "genesyscloud_intents_customerintents",
		"data_schema.id":        "genesyscloud_task_management_workitem_schema",
		"stageplan.stepplan.workitem_settings.worktype_id": "genesyscloud_task_management_worktype",
	}
	require.Len(t, e.RefAttrs, len(want))
	for attr, refType := range want {
		require.Contains(t, e.RefAttrs, attr)
		assert.Equal(t, refType, e.RefAttrs[attr].RefType, attr)
	}
	assert.Equal(t, []string{"*"}, e.RefAttrs["division_id"].AltValues)

	for _, attr := range []string{"published_version", "has_draft", "stageplan.id", "stageplan.stepplan.id"} {
		assert.Contains(t, e.ExcludedAttributes, attr)
	}
}

func TestUnitNextAfterCursor(t *testing.T) {
	t.Parallel()
	uri := func(s string) *string { return &s }

	next, done, err := nextAfterCursor(uri("/api/v2/casemanagement/caseplans/cp1/versions/1/stageplans?pageSize=25&after=abc"), "")
	require.NoError(t, err)
	assert.False(t, done)
	assert.Equal(t, "abc", next)

	for name, u := range map[string]*string{"nil": nil, "empty": uri(""), "no cursor": uri("/x?pageSize=25")} {
		_, done, err = nextAfterCursor(u, "")
		require.NoError(t, err, name)
		assert.True(t, done, name)
	}

	_, done, err = nextAfterCursor(uri("/x?after=abc"), "abc")
	require.NoError(t, err)
	assert.True(t, done, "a repeated cursor ends paging instead of looping")
}

// listingServer serves cursor-paged listings and fails the test if a page is requested with anything but the
// cursor from the previous page's nextUri (for example the last entity id).
func listingServer(t *testing.T, pages map[string]map[string]interface{}) *platformclientv2.Configuration {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path + "?after=" + r.URL.Query().Get("after")
		page, ok := pages[key]
		if !ok {
			t.Errorf("unexpected request %s", key)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(page)
	}))
	t.Cleanup(srv.Close)

	cfg := platformclientv2.NewConfiguration()
	cfg.BasePath = srv.URL
	cfg.AccessToken = "token"
	return cfg
}

func TestUnitProxyPagination_usesCursorAndKeepsAPIOrder(t *testing.T) {
	t.Parallel()
	stagePath := "/api/v2/casemanagement/caseplans/cp1/versions/latest/stageplans"
	stepPath := stagePath + "/s3/stepplans"
	caseplansPath := "/api/v2/casemanagement/caseplans"

	cfg := listingServer(t, map[string]map[string]interface{}{
		stagePath + "?after=": {
			"entities": []map[string]string{{"id": "s3", "name": "Zeta"}, {"id": "s1", "name": "Alpha"}},
			"nextUri":  stagePath + "?pageSize=25&after=cursor-2",
		},
		stagePath + "?after=cursor-2": {
			"entities": []map[string]string{{"id": "s2", "name": "Mid"}},
		},
		stepPath + "?after=": {
			"entities": []map[string]string{{"id": "p3", "name": "Step"}},
		},
		caseplansPath + "?after=": {
			"entities": []map[string]string{{"id": "cp-b", "name": "B"}},
			"nextUri":  caseplansPath + "?pageSize=100&after=cp-cursor",
		},
		caseplansPath + "?after=cp-cursor": {
			"entities": []map[string]string{{"id": "cp-a", "name": "A"}},
		},
	})
	p := newCaseManagementCaseplanProxy(cfg)

	stages, _, err := p.listStageplans(context.Background(), "cp1", caseplanAPIVersionLatest)
	require.NoError(t, err)
	ids := []string{}
	for _, s := range stages {
		ids = append(ids, *s.Id)
	}
	assert.Equal(t, []string{"s3", "s1", "s2"}, ids)

	steps, _, err := p.listStepplans(context.Background(), "cp1", caseplanAPIVersionLatest, "s3")
	require.NoError(t, err)
	require.Len(t, steps, 1)
	assert.Equal(t, "p3", *steps[0].Id)

	caseplans, _, err := p.getAllCaseManagementCaseplan(context.Background())
	require.NoError(t, err)
	require.Len(t, *caseplans, 2)
	assert.Equal(t, "cp-b", *(*caseplans)[0].Id)
	assert.Equal(t, "cp-a", *(*caseplans)[1].Id)
}

func stage(name string, step stepplanConfig) stageplanConfig {
	if step.activityType == "" {
		step.activityType = activityTypeNone
	}
	return stageplanConfig{name: name, step: step}
}

func noneStep(name string) stepplanConfig {
	return stepplanConfig{name: name, activityType: activityTypeNone}
}

func workitemStep(name, worktypeID string) stepplanConfig {
	return stepplanConfig{name: name, activityType: activityTypeWorkitem, worktypeID: worktypeID}
}

func TestUnitMatchStageplans(t *testing.T) {
	t.Parallel()
	actual := []stageplanConfig{{id: "1", name: "A"}, {id: "2", name: "B"}, {id: "3", name: "C"}}

	ids, del := matchStageplans(actual, actual, []stageplanConfig{{name: "C"}, {name: "A"}})
	assert.Equal(t, []string{"3", "1"}, ids)
	assert.Equal(t, []string{"2"}, del)

	ids, del = matchStageplans(actual, actual, []stageplanConfig{{name: "X"}, {name: "B"}, {name: "Y"}, {name: "Z"}})
	assert.Equal(t, []string{"1", "2", "3", ""}, ids)
	assert.Empty(t, del)
}

func TestUnitBuildStepplanUpdate(t *testing.T) {
	t.Parallel()

	update, changed := buildStepplanUpdate(noneStep("s"), workitemStep("s", "wt-1"))
	require.True(t, changed)
	body, err := json.Marshal(update)
	require.NoError(t, err)
	assert.JSONEq(t, `{"activityType":"Workitem","workitemSettings":{"worktypeId":"wt-1"}}`, string(body))

	update, changed = buildStepplanUpdate(stepplanConfig{name: "s", description: "d", activityType: activityTypeWorkitem, worktypeID: "wt-1"}, noneStep("s"))
	require.True(t, changed)
	body, err = json.Marshal(update)
	require.NoError(t, err)
	assert.JSONEq(t, `{"description":null,"activityType":"None","workitemSettings":null}`, string(body))

	_, changed = buildStepplanUpdate(workitemStep("s", "wt-1"), workitemStep("s", "wt-1"))
	assert.False(t, changed)
}

func TestUnitStageplanRepositionToFrontSendsNullAfter(t *testing.T) {
	t.Parallel()
	body := platformclientv2.Stageplanreposition{}
	body.SetField("After", nil)
	out, err := json.Marshal(body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"after":null}`, string(out))
}

func TestUnitMoveAfter(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{"c", "a", "b"}, moveAfter([]string{"a", "b", "c"}, "c", ""))
	assert.Equal(t, []string{"b", "a", "c"}, moveAfter([]string{"a", "b", "c"}, "a", "b"))
	assert.Equal(t, []string{"a", "c", "b"}, moveAfter([]string{"a", "b", "c"}, "b", "c"))
}

func TestUnitLongestIncreasingSubsequence(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []bool{true, true, true}, longestIncreasingSubsequence([]int{0, 1, 2}))
	assert.Equal(t, []bool{true, true, false}, longestIncreasingSubsequence([]int{1, 2, 0}))
	assert.Equal(t, []bool{false, true, true}, longestIncreasingSubsequence([]int{2, 0, 1}))
	assert.Equal(t, 1, countTrue(longestIncreasingSubsequence([]int{4, 3, 2, 1, 0})))
	assert.Equal(t, 3, countTrue(longestIncreasingSubsequence([]int{1, 0, 2, 4, 3})))
	assert.Empty(t, longestIncreasingSubsequence(nil))
}

func countTrue(bs []bool) int {
	n := 0
	for _, b := range bs {
		if b {
			n++
		}
	}
	return n
}

func TestUnitFlattenExpandStageplans(t *testing.T) {
	t.Parallel()
	in := []stageplanConfig{
		{id: "s1", name: "A", description: "d", step: stepplanConfig{id: "p1", name: "x", activityType: activityTypeWorkitem, worktypeID: "wt"}},
		{id: "s2", name: "B", step: stepplanConfig{id: "p2", name: "y", activityType: activityTypeNone}},
	}
	assert.Equal(t, in, expandStageplans(flattenStageplans(in)))
}

// ---- In-memory fake of the caseplan API (draft rule, 1..5 stageplans, Workitem invariant) ----

type fakeCaseplanAPI struct {
	t         *testing.T
	id        string
	latest    int
	published int
	record    platformclientv2.Caseplan
	schemaID  string
	intake    []platformclientv2.Intakesetting
	stages    []stageplanConfig
	others    []platformclientv2.Caseplan
	nextID    int
	calls     []string
}

func newFakeCaseplanAPI(t *testing.T, stageNames ...string) *fakeCaseplanAPI {
	f := &fakeCaseplanAPI{t: t, id: "cp1", latest: 1, schemaID: "schema-1"}
	f.record = platformclientv2.Caseplan{
		Id:                          platformclientv2.String(f.id),
		Name:                        platformclientv2.String("cp"),
		ReferencePrefix:             platformclientv2.String("AB12"),
		DefaultDueDurationInSeconds: platformclientv2.Int(100),
		DefaultTtlSeconds:           platformclientv2.Int(200),
	}
	for i, name := range stageNames {
		f.stages = append(f.stages, stageplanConfig{id: f.newID("stage"), name: name,
			step: stepplanConfig{id: f.newID("step"), name: fmt.Sprintf("Step %d", i+1), activityType: activityTypeNone}})
	}
	return f
}

func (f *fakeCaseplanAPI) newID(prefix string) string {
	f.nextID++
	return fmt.Sprintf("%s-%d", prefix, f.nextID)
}

func (f *fakeCaseplanAPI) count(op string) int {
	n := 0
	for _, c := range f.calls {
		if c == op {
			n++
		}
	}
	return n
}

func (f *fakeCaseplanAPI) draftRequired() error {
	if f.published != 0 && f.latest == f.published {
		return errors.New("409 pre.check.draft.caseplan.version.required")
	}
	return nil
}

func (f *fakeCaseplanAPI) checkCount() {
	if len(f.stages) < 1 || len(f.stages) > maxStageplans {
		f.t.Errorf("stageplan count %d outside 1..%d after %v", len(f.stages), maxStageplans, f.calls)
	}
}

func (f *fakeCaseplanAPI) stageIndex(id string) int {
	for i, s := range f.stages {
		if s.id == id {
			return i
		}
	}
	return -1
}

func (f *fakeCaseplanAPI) snapshot() *platformclientv2.Caseplan {
	cp := f.record
	cp.Latest = platformclientv2.Int(f.latest)
	if f.published != 0 {
		cp.Published = platformclientv2.Int(f.published)
	}
	return &cp
}

func ok() *platformclientv2.APIResponse {
	return &platformclientv2.APIResponse{StatusCode: http.StatusOK}
}

func failed(err error) (*platformclientv2.APIResponse, error) {
	return &platformclientv2.APIResponse{StatusCode: http.StatusConflict}, err
}

func (f *fakeCaseplanAPI) assertStages(want []stageplanConfig) {
	f.t.Helper()
	if len(f.stages) != len(want) {
		f.t.Fatalf("want %d stageplans, got %d: %+v", len(want), len(f.stages), f.stages)
	}
	for i := range want {
		got, w := f.stages[i], want[i]
		got.id, got.step.id, w.id, w.step.id = "", "", "", ""
		if got != w {
			f.t.Errorf("stageplan %d: want %+v, got %+v", i, w, got)
		}
	}
}

func (f *fakeCaseplanAPI) proxy() *caseManagementCaseplanProxy {
	return &caseManagementCaseplanProxy{
		createCaseManagementCaseplanAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, body *platformclientv2.Caseplancreate) (*platformclientv2.Caseplancreateresponse, *platformclientv2.APIResponse, error) {
			f.calls = append(f.calls, "create-caseplan")
			f.latest, f.published = 1, 0
			f.record.Name, f.record.Description, f.record.ReferencePrefix = body.Name, body.Description, body.ReferencePrefix
			if body.DataSchemas != nil && len(*body.DataSchemas) > 0 {
				f.schemaID = stringValue((*body.DataSchemas)[0].Id)
			}
			f.stages = nil
			for i := 1; i <= 3; i++ {
				f.stages = append(f.stages, stageplanConfig{id: f.newID("stage"), name: fmt.Sprintf("Stage %d", i),
					step: stepplanConfig{id: f.newID("step"), name: fmt.Sprintf("Step %d", i), activityType: activityTypeNone}})
			}
			return &platformclientv2.Caseplancreateresponse{Id: platformclientv2.String(f.id), Latest: platformclientv2.Int(1)}, ok(), nil
		},
		getAllCaseManagementCaseplanAttr: func(ctx context.Context, p *caseManagementCaseplanProxy) (*[]platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
			all := append([]platformclientv2.Caseplan{*f.snapshot()}, f.others...)
			return &all, ok(), nil
		},
		postCaseManagementCaseplanDataschemaAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId, schemaId string) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error) {
			f.calls = append(f.calls, "post-dataschema")
			f.schemaID = schemaId
			return &platformclientv2.Caseplandataschema{Id: platformclientv2.String(schemaId)}, ok(), nil
		},
		putCaseManagementCaseplanDataschemaAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId, key, schemaId string) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error) {
			f.calls = append(f.calls, "put-dataschema")
			if f.published != 0 {
				resp, err := failed(errors.New("400 data schema cannot change after publish"))
				return nil, resp, err
			}
			f.schemaID = schemaId
			return &platformclientv2.Caseplandataschema{Id: platformclientv2.String(schemaId)}, ok(), nil
		},
		getCaseManagementCaseplanByIdAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, id string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
			return f.snapshot(), ok(), nil
		},
		getCaseManagementCaseplanVersionAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId, versionId string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
			return f.snapshot(), ok(), nil
		},
		getCaseManagementCaseplanVersionDataschemasAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId, versionId string) (*platformclientv2.Caseplandataschemalisting, *platformclientv2.APIResponse, error) {
			return &platformclientv2.Caseplandataschemalisting{Entities: &[]platformclientv2.Caseplandataschema{{Id: platformclientv2.String(f.schemaID)}}}, ok(), nil
		},
		getCaseManagementCaseplanVersionIntakesettingsAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId, versionId string) (*platformclientv2.Intakesettingslisting, *platformclientv2.APIResponse, error) {
			intake := append([]platformclientv2.Intakesetting(nil), f.intake...)
			return &platformclientv2.Intakesettingslisting{Entities: &intake}, ok(), nil
		},
		putCaseManagementCaseplanIntakesettingsAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, body platformclientv2.Intakesettingsupdate) (*platformclientv2.Intakesettingslisting, *platformclientv2.APIResponse, error) {
			f.calls = append(f.calls, "put-intake")
			if err := f.draftRequired(); err != nil {
				resp, err := failed(err)
				return nil, resp, err
			}
			f.intake = *body.IntakeSettings
			return &platformclientv2.Intakesettingslisting{Entities: body.IntakeSettings}, ok(), nil
		},
		patchCaseManagementCaseplanAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, body platformclientv2.Caseplanupdate) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
			versioned := body.SetFieldNames["DefaultDueDurationInSeconds"] || body.SetFieldNames["DefaultTtlSeconds"] || body.SetFieldNames["DefaultCaseOwnerId"]
			if versioned {
				f.calls = append(f.calls, "patch-caseplan-versioned")
				if err := f.draftRequired(); err != nil {
					resp, err := failed(err)
					return nil, resp, err
				}
			} else {
				f.calls = append(f.calls, "patch-caseplan")
			}
			if body.SetFieldNames["Name"] {
				f.record.Name = body.Name
			}
			if body.SetFieldNames["Description"] {
				f.record.Description = body.Description
			}
			if body.SetFieldNames["DefaultDueDurationInSeconds"] {
				f.record.DefaultDueDurationInSeconds = body.DefaultDueDurationInSeconds
			}
			if body.SetFieldNames["DefaultTtlSeconds"] {
				f.record.DefaultTtlSeconds = body.DefaultTtlSeconds
			}
			return f.snapshot(), ok(), nil
		},
		postCaseManagementCaseplanVersionsAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
			f.calls = append(f.calls, "post-version")
			if f.published == 0 || f.latest > f.published {
				resp, err := failed(errors.New("409 precondition.check.failed"))
				return nil, resp, err
			}
			f.latest++
			return f.snapshot(), ok(), nil
		},
		publishCaseManagementCaseplanAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
			f.calls = append(f.calls, "publish")
			if f.published == f.latest {
				resp, err := failed(errors.New("409 nothing to publish"))
				return nil, resp, err
			}
			f.published = f.latest
			return f.snapshot(), ok(), nil
		},
		listStageplansAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId, versionId string) ([]platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
			out := make([]platformclientv2.Stageplan, 0, len(f.stages))
			for _, s := range f.stages {
				out = append(out, platformclientv2.Stageplan{Id: platformclientv2.String(s.id), Name: platformclientv2.String(s.name), Description: nullableString(s.description)})
			}
			return out, ok(), nil
		},
		listStepplansAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId, versionId, stageplanId string) ([]platformclientv2.Stepplan, *platformclientv2.APIResponse, error) {
			i := f.stageIndex(stageplanId)
			if i < 0 {
				resp, err := failed(errors.New("404 stageplan not found"))
				return nil, resp, err
			}
			s := f.stages[i].step
			step := platformclientv2.Stepplan{Id: platformclientv2.String(s.id), Name: platformclientv2.String(s.name),
				Description: nullableString(s.description), ActivityType: platformclientv2.String(s.activityType)}
			if s.worktypeID != "" {
				step.WorkitemSettings = &platformclientv2.Workitemsettingsresponse{Worktype: &platformclientv2.Stepplansworktypereference{Id: platformclientv2.String(s.worktypeID)}}
			}
			return []platformclientv2.Stepplan{step}, ok(), nil
		},
		createStageplanAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId string, body platformclientv2.Stageplancreate) (*platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
			f.calls = append(f.calls, "create")
			if err := f.draftRequired(); err != nil {
				resp, err := failed(err)
				return nil, resp, err
			}
			if len(f.stages) >= maxStageplans {
				resp, err := failed(errors.New("409 pre.check.maximum.limit.reached"))
				return nil, resp, err
			}
			pos := 0
			if body.After != nil {
				if pos = f.stageIndex(*body.After) + 1; pos == 0 {
					resp, err := failed(errors.New("400 invalid.input after"))
					return nil, resp, err
				}
			}
			s := stageplanConfig{id: f.newID("stage"), name: *body.Name, description: stringValue(body.Description),
				step: stepplanConfig{id: f.newID("step"), name: "Step 1", activityType: activityTypeNone}}
			f.stages = append(f.stages[:pos], append([]stageplanConfig{s}, f.stages[pos:]...)...)
			f.checkCount()
			return &platformclientv2.Stageplan{Id: platformclientv2.String(s.id), Name: body.Name}, ok(), nil
		},
		deleteStageplanAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId, stageplanId string) (*platformclientv2.APIResponse, error) {
			f.calls = append(f.calls, "delete")
			if err := f.draftRequired(); err != nil {
				return failed(err)
			}
			i := f.stageIndex(stageplanId)
			if i < 0 {
				return failed(errors.New("404 stageplan not found"))
			}
			if len(f.stages) == 1 {
				return failed(errors.New("409 pre.check.min.stageplans.required"))
			}
			f.stages = append(f.stages[:i], f.stages[i+1:]...)
			f.checkCount()
			return ok(), nil
		},
		repositionStageplanAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId, stageplanId string, body platformclientv2.Stageplanreposition) (*platformclientv2.APIResponse, error) {
			f.calls = append(f.calls, "reposition")
			if err := f.draftRequired(); err != nil {
				return failed(err)
			}
			if !body.SetFieldNames["After"] {
				f.t.Errorf("reposition of %s did not set after explicitly", stageplanId)
			}
			i := f.stageIndex(stageplanId)
			if i < 0 {
				return failed(errors.New("404 stageplan not found"))
			}
			after := stringValue(body.After)
			if after == stageplanId || (after != "" && f.stageIndex(after) < 0) {
				return failed(errors.New("400 invalid.input after"))
			}
			s := f.stages[i]
			f.stages = append(f.stages[:i], f.stages[i+1:]...)
			pos := 0
			if after != "" {
				pos = f.stageIndex(after) + 1
			}
			f.stages = append(f.stages[:pos], append([]stageplanConfig{s}, f.stages[pos:]...)...)
			return ok(), nil
		},
		patchStageplanAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId, stageplanId string, body platformclientv2.Stageplanupdate) (*platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
			f.calls = append(f.calls, "patch-stage")
			if err := f.draftRequired(); err != nil {
				resp, err := failed(err)
				return nil, resp, err
			}
			i := f.stageIndex(stageplanId)
			if i < 0 {
				resp, err := failed(errors.New("404 stageplan not found"))
				return nil, resp, err
			}
			if body.SetFieldNames["Name"] {
				f.stages[i].name = *body.Name
			}
			if body.SetFieldNames["Description"] {
				f.stages[i].description = stringValue(body.Description)
			}
			return &platformclientv2.Stageplan{Id: platformclientv2.String(stageplanId)}, ok(), nil
		},
		patchStepplanAttr: func(ctx context.Context, p *caseManagementCaseplanProxy, caseplanId, stageplanId, stepplanId string, body platformclientv2.Stepplanupdate) (*platformclientv2.Stepplan, *platformclientv2.APIResponse, error) {
			f.calls = append(f.calls, "patch-step")
			if err := f.draftRequired(); err != nil {
				resp, err := failed(err)
				return nil, resp, err
			}
			i := f.stageIndex(stageplanId)
			if i < 0 || f.stages[i].step.id != stepplanId {
				resp, err := failed(errors.New("404 stepplan not found"))
				return nil, resp, err
			}
			step := f.stages[i].step
			if body.SetFieldNames["Name"] {
				step.name = *body.Name
			}
			if body.SetFieldNames["Description"] {
				step.description = stringValue(body.Description)
			}
			if body.SetFieldNames["ActivityType"] {
				step.activityType = *body.ActivityType
			}
			if body.SetFieldNames["WorkitemSettings"] {
				step.worktypeID = ""
				if body.WorkitemSettings != nil {
					step.worktypeID = stringValue(body.WorkitemSettings.WorktypeId)
				}
			}
			if (step.activityType == activityTypeWorkitem) != (step.worktypeID != "") {
				resp, err := failed(errors.New("400 activityType Workitem requires workitemSettings"))
				return nil, resp, err
			}
			f.stages[i].step = step
			return &platformclientv2.Stepplan{Id: platformclientv2.String(stepplanId)}, ok(), nil
		},
	}
}

func liveStages(f *fakeCaseplanAPI) []stageplanConfig {
	return append([]stageplanConfig(nil), f.stages...)
}

func withFakeProxy(t *testing.T, f *fakeCaseplanAPI) *provider.ProviderMeta {
	t.Helper()
	internalProxy = f.proxy()
	t.Cleanup(func() { internalProxy = nil })
	return &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
}

// caseplanStateAttrs is the flat state of a published caseplan whose single stageplan mirrors the fake's.
func caseplanStateAttrs(f *fakeCaseplanAPI) map[string]string {
	return map[string]string{
		"id": f.id, "name": "cp", "reference_prefix": "AB12",
		"default_due_duration_in_seconds": "100", "default_ttl_seconds": "200",
		"published_version": "1", "has_draft": "false",
		"data_schema.#": "1", "data_schema.0.id": "schema-1",
		"stageplan.#": "1", "stageplan.0.id": f.stages[0].id, "stageplan.0.name": f.stages[0].name, "stageplan.0.description": "",
		"stageplan.0.stepplan.#": "1", "stageplan.0.stepplan.0.id": f.stages[0].step.id,
		"stageplan.0.stepplan.0.name": f.stages[0].step.name, "stageplan.0.stepplan.0.activity_type": activityTypeNone,
		"stageplan.0.stepplan.0.description": "", "stageplan.0.stepplan.0.workitem_settings.#": "0",
	}
}

func resourceDataWithDiff(t *testing.T, state map[string]string, diff map[string]*terraform.ResourceAttrDiff) *schema.ResourceData {
	t.Helper()
	d, err := schema.InternalMap(ResourceCaseManagementCaseplan().Schema).Data(
		&terraform.InstanceState{ID: state["id"], Attributes: state}, &terraform.InstanceDiff{Attributes: diff})
	require.NoError(t, err)
	return d
}

// ---- Stageplan reconcile matrix ----

func TestUnitReconcileStageplans(t *testing.T) {
	t.Parallel()
	defaults := []string{"Stage 1", "Stage 2", "Stage 3"}
	run := func(f *fakeCaseplanAPI, prior, desired []stageplanConfig) []stageplanConfig {
		t.Helper()
		resolved, diags := reconcileStageplans(context.Background(), f.proxy(), f.id, prior, desired)
		require.Empty(t, diags)
		f.assertStages(desired)
		return resolved
	}

	t.Run("3 desired map onto the defaults by position", func(t *testing.T) {
		f := newFakeCaseplanAPI(t, defaults...)
		resolved := run(f, nil, []stageplanConfig{stage("Intake", noneStep("Triage")), stage("Work", workitemStep("Do", "wt-1")), stage("Done", noneStep("Close"))})
		assert.Equal(t, 3, f.count("patch-stage"))
		assert.Zero(t, f.count("create")+f.count("delete")+f.count("reposition"))
		for i := range resolved {
			assert.Equal(t, f.stages[i].id, resolved[i].id)
			assert.Equal(t, f.stages[i].step.id, resolved[i].step.id)
		}
	})
	t.Run("1 desired deletes surplus defaults", func(t *testing.T) {
		f := newFakeCaseplanAPI(t, defaults...)
		run(f, nil, []stageplanConfig{stage("Only", noneStep("Step"))})
		assert.Equal(t, 2, f.count("delete"))
	})
	t.Run("5 desired adds 2", func(t *testing.T) {
		f := newFakeCaseplanAPI(t, defaults...)
		run(f, nil, []stageplanConfig{stage("A1", noneStep("s")), stage("B1", noneStep("s")), stage("C1", noneStep("s")), stage("D1", workitemStep("s", "wt-1")), stage("E1", noneStep("s"))})
		assert.Equal(t, 2, f.count("create"))
		assert.Zero(t, f.count("reposition"))
	})
	t.Run("rename in place keeps the id", func(t *testing.T) {
		f := newFakeCaseplanAPI(t, "A", "B", "C")
		prior := liveStages(f)
		run(f, prior, []stageplanConfig{stage("A", noneStep("Step 1")), stage("B renamed", noneStep("Step 2")), stage("C", noneStep("Step 3"))})
		assert.Equal(t, prior[1].id, f.stages[1].id)
		assert.Equal(t, []string{"patch-stage"}, f.calls)
	})
	t.Run("insert in the middle is a single create", func(t *testing.T) {
		f := newFakeCaseplanAPI(t, "A", "B", "C")
		prior := liveStages(f)
		run(f, prior, []stageplanConfig{stage("A", noneStep("Step 1")), stage("New", noneStep("Step 1")), stage("B", noneStep("Step 2")), stage("C", noneStep("Step 3"))})
		assert.Equal(t, []string{"create"}, f.calls)
		assert.Equal(t, prior[1].id, f.stages[2].id)
	})
	t.Run("delete the middle", func(t *testing.T) {
		f := newFakeCaseplanAPI(t, "A", "B", "C")
		run(f, liveStages(f), []stageplanConfig{stage("A", noneStep("Step 1")), stage("C", noneStep("Step 3"))})
		assert.Equal(t, []string{"delete"}, f.calls)
	})
	t.Run("replace all 5", func(t *testing.T) {
		f := newFakeCaseplanAPI(t, "A", "B", "C", "D", "E")
		run(f, liveStages(f), []stageplanConfig{stage("V", noneStep("s")), stage("W", noneStep("s")), stage("X", noneStep("s")), stage("Y", noneStep("s")), stage("Z", noneStep("s"))})
	})
	t.Run("shift at the max deletes before creating", func(t *testing.T) {
		f := newFakeCaseplanAPI(t, "A", "B", "C", "D", "E")
		run(f, liveStages(f), []stageplanConfig{stage("B", noneStep("Step 2")), stage("C", noneStep("Step 3")), stage("D", noneStep("Step 4")), stage("E", noneStep("Step 5")), stage("F", noneStep("Step 1"))})
		assert.Equal(t, 1, f.count("delete"))
		assert.Equal(t, 1, f.count("create"))
	})
	t.Run("single stageplan replaced defers the delete", func(t *testing.T) {
		f := newFakeCaseplanAPI(t, "A")
		run(f, []stageplanConfig{{id: "gone", name: "Old"}}, []stageplanConfig{stage("New", noneStep("Step 1"))})
	})
	t.Run("rename and move becomes delete and create", func(t *testing.T) {
		f := newFakeCaseplanAPI(t, "A", "B")
		prior := liveStages(f)
		run(f, prior, []stageplanConfig{stage("B2", noneStep("Step 1")), stage("A", noneStep("Step 1"))})
		assert.Equal(t, 1, f.count("delete"))
		assert.Equal(t, 1, f.count("create"))
		assert.Equal(t, prior[0].id, f.stages[1].id)
	})
	t.Run("workitem to none clears settings", func(t *testing.T) {
		f := newFakeCaseplanAPI(t, "A")
		f.stages[0].step = workitemStep("Step 1", "wt-1")
		run(f, liveStages(f), []stageplanConfig{stage("A", noneStep("Step 1"))})
	})
	t.Run("no changes makes no writes", func(t *testing.T) {
		f := newFakeCaseplanAPI(t, "A", "B")
		run(f, liveStages(f), []stageplanConfig{stage("A", noneStep("Step 1")), stage("B", noneStep("Step 2"))})
		assert.Empty(t, f.calls)
	})
	t.Run("stageplan deleted outside terraform is recreated", func(t *testing.T) {
		f := newFakeCaseplanAPI(t, "A", "B")
		prior := append(liveStages(f), stageplanConfig{id: "deleted-in-ui", name: "C"})
		run(f, prior, []stageplanConfig{stage("A", noneStep("Step 1")), stage("B", noneStep("Step 2")), stage("C", noneStep("Step 1"))})
		assert.Equal(t, []string{"create"}, f.calls)
	})
	t.Run("create failure mentions the feature toggle", func(t *testing.T) {
		f := newFakeCaseplanAPI(t, "A")
		p := f.proxy()
		p.createStageplanAttr = func(ctx context.Context, pr *caseManagementCaseplanProxy, caseplanId string, body platformclientv2.Stageplancreate) (*platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
			resp, err := failed(assert.AnError)
			return nil, resp, err
		}
		_, diags := reconcileStageplans(context.Background(), p, f.id, liveStages(f), []stageplanConfig{stage("A", noneStep("Step 1")), stage("B", noneStep("Step 1"))})
		require.NotEmpty(t, diags)
		assert.Contains(t, diags[0].Summary+diags[0].Detail, "PURE-8006")
	})

	reorders := []struct {
		name     string
		from, to []string
		moves    int
	}{
		{"first to last", []string{"A", "B", "C"}, []string{"B", "C", "A"}, 1},
		{"last to first", []string{"A", "B", "C"}, []string{"C", "A", "B"}, 1},
		{"swap adjacent", []string{"A", "B", "C", "D"}, []string{"A", "C", "B", "D"}, 1},
		{"swap first and last", []string{"A", "B", "C"}, []string{"C", "B", "A"}, 2},
		{"two disjoint swaps", []string{"A", "B", "C", "D", "E"}, []string{"B", "A", "C", "E", "D"}, 2},
		{"reverse 5", []string{"A", "B", "C", "D", "E"}, []string{"E", "D", "C", "B", "A"}, 4},
	}
	for _, tc := range reorders {
		t.Run("reorder "+tc.name+" uses the minimum repositions", func(t *testing.T) {
			f := newFakeCaseplanAPI(t, tc.from...)
			steps := map[string]string{}
			for _, s := range f.stages {
				steps[s.name] = s.step.name
			}
			desired := make([]stageplanConfig, len(tc.to))
			for i, name := range tc.to {
				desired[i] = stage(name, noneStep(steps[name]))
			}
			run(f, liveStages(f), desired)
			assert.Equal(t, tc.moves, f.count("reposition"))
			assert.Equal(t, tc.moves, len(f.calls), "only repositions: %v", f.calls)
		})
	}
}

// ---- Create / update / read / import flows (the update tests share internalProxy, so no t.Parallel) ----

func TestUnitCreateCaseplan_reconcilesStageplansAndPublishes(t *testing.T) {
	f := newFakeCaseplanAPI(t)
	meta := withFakeProxy(t, f)
	d := schema.TestResourceDataRaw(t, ResourceCaseManagementCaseplan().Schema, map[string]interface{}{
		"name": "cp", "reference_prefix": "AB12",
		"data_schema": []interface{}{map[string]interface{}{"id": "schema-1"}},
		"stageplan": []interface{}{
			map[string]interface{}{"name": "Intake", "stepplan": []interface{}{map[string]interface{}{"name": "Step 1", "activity_type": activityTypeNone}}},
			map[string]interface{}{"name": "Work", "stepplan": []interface{}{map[string]interface{}{
				"name": "Do", "activity_type": activityTypeWorkitem,
				"workitem_settings": []interface{}{map[string]interface{}{"worktype_id": "wt-1"}},
			}}},
		},
	})
	require.Empty(t, createCaseManagementCaseplan(context.Background(), d, meta))
	assert.Equal(t, "cp1", d.Id())
	assert.Equal(t, []string{"create-caseplan", "delete", "patch-stage", "patch-stage", "patch-step", "publish"}, f.calls)
	assert.Zero(t, f.count("post-version"), "a new caseplan is already a draft")
	f.assertStages([]stageplanConfig{stage("Intake", noneStep("Step 1")), stage("Work", workitemStep("Do", "wt-1"))})
	assert.Equal(t, 1, d.Get("published_version"))
	assert.Equal(t, false, d.Get("has_draft"))
	assert.Equal(t, f.stages[1].step.id, d.Get("stageplan.1.stepplan.0.id"))
}

func TestUnitCreateCaseplan_withoutStageplansStillPublishes(t *testing.T) {
	f := newFakeCaseplanAPI(t)
	meta := withFakeProxy(t, f)
	d := schema.TestResourceDataRaw(t, ResourceCaseManagementCaseplan().Schema, map[string]interface{}{
		"name": "cp", "data_schema": []interface{}{map[string]interface{}{"id": "schema-1"}},
	})
	require.Empty(t, createCaseManagementCaseplan(context.Background(), d, meta))
	assert.Equal(t, []string{"create-caseplan", "publish"}, f.calls)
	assert.Len(t, f.stages, 3, "unmanaged stageplans are left alone")
	assert.Equal(t, 0, d.Get("stageplan.#"))
}

func TestUnitUpdateCaseplan_nameOnlyDoesNotCreateDraft(t *testing.T) {
	f := newFakeCaseplanAPI(t, "Stage A")
	f.published = 1
	meta := withFakeProxy(t, f)
	d := resourceDataWithDiff(t, caseplanStateAttrs(f), map[string]*terraform.ResourceAttrDiff{"name": {Old: "cp", New: "renamed"}})
	require.Empty(t, updateCaseManagementCaseplan(context.Background(), d, meta))
	assert.Equal(t, []string{"patch-caseplan"}, f.calls)
	assert.Equal(t, 1, f.latest)
	assert.Equal(t, "renamed", d.Get("name"))
}

func TestUnitUpdateCaseplan_versionedChangeCreatesDraftAndPublishes(t *testing.T) {
	f := newFakeCaseplanAPI(t, "Stage A")
	f.published = 1
	meta := withFakeProxy(t, f)
	stageID := f.stages[0].id
	d := resourceDataWithDiff(t, caseplanStateAttrs(f), map[string]*terraform.ResourceAttrDiff{
		"stageplan.0.stepplan.0.name": {Old: f.stages[0].step.name, New: "Renamed step"},
	})
	require.Empty(t, updateCaseManagementCaseplan(context.Background(), d, meta))
	assert.Equal(t, []string{"post-version", "patch-step", "publish"}, f.calls)
	assert.Equal(t, 2, d.Get("published_version"))
	assert.Equal(t, false, d.Get("has_draft"))
	assert.Equal(t, stageID, d.Get("stageplan.0.id"), "ids are stable across versions")
}

func TestUnitUpdateCaseplan_reusesExistingDraftAndOverwritesIt(t *testing.T) {
	f := newFakeCaseplanAPI(t, "Stage A")
	f.published, f.latest = 1, 2
	f.record.DefaultTtlSeconds = platformclientv2.Int(999)
	meta := withFakeProxy(t, f)
	d := resourceDataWithDiff(t, caseplanStateAttrs(f), map[string]*terraform.ResourceAttrDiff{
		"default_due_duration_in_seconds": {Old: "100", New: "150"},
	})
	require.Empty(t, updateCaseManagementCaseplan(context.Background(), d, meta))
	assert.Equal(t, []string{"patch-caseplan-versioned", "publish"}, f.calls)
	assert.Equal(t, 200, *f.record.DefaultTtlSeconds, "UI draft edits to managed fields are overwritten")
	assert.Equal(t, 2, f.published)
}

func TestUnitUpdateCaseplan_neverPublishedIsPublished(t *testing.T) {
	f := newFakeCaseplanAPI(t, "Stage A")
	meta := withFakeProxy(t, f)
	state := caseplanStateAttrs(f)
	state["published_version"] = "0"
	d := resourceDataWithDiff(t, state, map[string]*terraform.ResourceAttrDiff{"published_version": {Old: "0", NewComputed: true}})
	require.Empty(t, updateCaseManagementCaseplan(context.Background(), d, meta))
	assert.Equal(t, []string{"publish"}, f.calls)
	assert.Equal(t, 1, f.published)
}

func TestUnitUpdateCaseplan_neverPublishedSyncsDataSchema(t *testing.T) {
	f := newFakeCaseplanAPI(t, "Stage A")
	meta := withFakeProxy(t, f)
	state := caseplanStateAttrs(f)
	state["published_version"] = "0"
	d := resourceDataWithDiff(t, state, map[string]*terraform.ResourceAttrDiff{
		"data_schema.0.id": {Old: "schema-1", New: "schema-2"}, "published_version": {Old: "0", NewComputed: true},
	})
	require.Empty(t, updateCaseManagementCaseplan(context.Background(), d, meta))
	assert.Equal(t, []string{"put-dataschema", "publish"}, f.calls)
	assert.Equal(t, "schema-2", f.schemaID)
}

func TestUnitReadCaseplan_readsPublishedVersionWhenDraftExists(t *testing.T) {
	f := newFakeCaseplanAPI(t, "Stage A")
	f.published, f.latest = 1, 2
	p := f.proxy()
	var versionsRead []string
	p.getCaseManagementCaseplanVersionAttr = func(ctx context.Context, pr *caseManagementCaseplanProxy, caseplanId, versionId string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
		versionsRead = append(versionsRead, versionId)
		cp := f.record
		cp.Name = platformclientv2.String("published name")
		return &cp, ok(), nil
	}
	internalProxy = p
	t.Cleanup(func() { internalProxy = nil })
	d := schema.TestResourceDataRaw(t, ResourceCaseManagementCaseplan().Schema, map[string]interface{}{"name": "published name"})
	d.SetId(f.id)
	require.Empty(t, readCaseManagementCaseplan(context.Background(), d, &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}))
	assert.Equal(t, []string{"1"}, versionsRead)
	assert.Equal(t, 1, d.Get("published_version"))
	assert.Equal(t, true, d.Get("has_draft"))
	assert.Empty(t, d.Get("stageplan"), "stageplans are not read when not configured")
}

func TestUnitImportCaseplan_seedsStageplansForRead(t *testing.T) {
	f := newFakeCaseplanAPI(t, "Stage A", "Stage B")
	f.published = 1
	meta := withFakeProxy(t, f)
	d := schema.TestResourceDataRaw(t, ResourceCaseManagementCaseplan().Schema, map[string]interface{}{})
	d.SetId(f.id)
	imported, err := importCaseManagementCaseplan(context.Background(), d, meta)
	require.NoError(t, err)
	require.Len(t, imported, 1)
	require.Empty(t, readCaseManagementCaseplan(context.Background(), imported[0], meta))
	assert.Equal(t, 2, d.Get("stageplan.#"))
	assert.Equal(t, f.stages[1].step.id, d.Get("stageplan.1.stepplan.0.id"))
}

// ---- Exporter ----

func TestUnitCaseplanExporter_skipsNeverPublished(t *testing.T) {
	f := newFakeCaseplanAPI(t)
	f.published = 1
	f.others = []platformclientv2.Caseplan{
		{Id: platformclientv2.String("draft-only"), Name: platformclientv2.String("never published"), Latest: platformclientv2.Int(1)},
		{Name: platformclientv2.String("no id"), Latest: platformclientv2.Int(1), Published: platformclientv2.Int(1)},
		{Id: platformclientv2.String("cp2"), Name: platformclientv2.String("live"), Latest: platformclientv2.Int(3), Published: platformclientv2.Int(2)},
	}
	withFakeProxy(t, f)
	resources, diags := getAllAuthCaseManagementCaseplans(context.Background(), &platformclientv2.Configuration{})
	require.Empty(t, diags)
	assert.Len(t, resources, 2)
	assert.Contains(t, resources, "cp1")
	assert.Contains(t, resources, "cp2")
	assert.Equal(t, "live", resources["cp2"].BlockLabel)
}

func TestUnitCaseplanRead_exportModes(t *testing.T) {
	modes := map[string]func(t *testing.T){
		"tfexporter": func(t *testing.T) {
			tfexporter_state.ActivateExporterState()
			t.Cleanup(tfexporter_state.ResetExporterStateForTests)
		},
		"mrmo": func(t *testing.T) { t.Setenv(mrmo.MRMO_CXASCODE_INTEGRATION_ENABLED, "true") },
	}
	for mode, activate := range modes {
		t.Run(mode+" skips a never-published caseplan", func(t *testing.T) {
			activate(t)
			require.True(t, isExporting())
			f := newFakeCaseplanAPI(t, "A")
			meta := withFakeProxy(t, f)
			d := schema.TestResourceDataRaw(t, ResourceCaseManagementCaseplan().Schema, map[string]interface{}{})
			d.SetId(f.id)
			require.Empty(t, readCaseManagementCaseplan(context.Background(), d, meta))
			assert.Empty(t, d.Id())
		})
		t.Run(mode+" reads stageplans and stepplans from empty state in API order", func(t *testing.T) {
			activate(t)
			f := newFakeCaseplanAPI(t, "Zeta", "Alpha")
			f.published = 1
			f.stages[1].step = workitemStep("Do", "wt-1")
			meta := withFakeProxy(t, f)
			d := schema.TestResourceDataRaw(t, ResourceCaseManagementCaseplan().Schema, map[string]interface{}{})
			d.SetId(f.id)
			require.Empty(t, readCaseManagementCaseplan(context.Background(), d, meta))
			assert.Equal(t, 2, d.Get("stageplan.#"))
			assert.Equal(t, "Zeta", d.Get("stageplan.0.name"))
			assert.Equal(t, "wt-1", d.Get("stageplan.1.stepplan.0.workitem_settings.0.worktype_id"))
		})
	}

	t.Run("outside export a never-published caseplan is kept and read from latest", func(t *testing.T) {
		tfexporter_state.ResetExporterStateForTests()
		require.False(t, isExporting())
		f := newFakeCaseplanAPI(t, "A")
		meta := withFakeProxy(t, f)
		d := schema.TestResourceDataRaw(t, ResourceCaseManagementCaseplan().Schema, map[string]interface{}{"name": "cp"})
		d.SetId(f.id)
		require.Empty(t, readCaseManagementCaseplan(context.Background(), d, meta))
		assert.Equal(t, f.id, d.Id())
		assert.Equal(t, 0, d.Get("published_version"))
		assert.Equal(t, true, d.Get("has_draft"))
	})
}
