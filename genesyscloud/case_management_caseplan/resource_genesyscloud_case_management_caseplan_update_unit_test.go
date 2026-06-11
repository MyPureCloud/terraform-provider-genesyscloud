package case_management_caseplan

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v191/platformclientv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_caseplanDataSchemaSyncPlanFromState(t *testing.T) {
	t.Parallel()
	t.Run("same id unchanged", func(t *testing.T) {
		raw := []interface{}{map[string]interface{}{"id": "a"}}
		del, puts := caseplanDataSchemaSyncPlanFromState(raw, raw)
		assert.Empty(t, del)
		assert.Empty(t, puts)
	})
	t.Run("replace id single schema", func(t *testing.T) {
		oldRaw := []interface{}{map[string]interface{}{"id": "a"}}
		newRaw := []interface{}{map[string]interface{}{"id": "b"}}
		del, puts := caseplanDataSchemaSyncPlanFromState(oldRaw, newRaw)
		assert.ElementsMatch(t, []string{"a"}, del)
		require.Len(t, puts, 1)
		assert.Equal(t, "b", *puts[0].Id)
	})
	t.Run("remove one of two", func(t *testing.T) {
		oldRaw := []interface{}{
			map[string]interface{}{"id": "a"},
			map[string]interface{}{"id": "b"},
		}
		newRaw := []interface{}{map[string]interface{}{"id": "b"}}
		del, puts := caseplanDataSchemaSyncPlanFromState(oldRaw, newRaw)
		assert.ElementsMatch(t, []string{"a"}, del)
		assert.Empty(t, puts)
	})
}

func TestUnit_execCaseplanDataSchemaSync_deleteDefaultThenPostNewId(t *testing.T) {
	t.Parallel()
	var ops []string
	p := &caseManagementCaseplanProxy{
		postCaseManagementCaseplanDataschemaAttr: func(ctx context.Context, pr *caseManagementCaseplanProxy, caseplanID string, body caseplanDataschemaPostBody) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error) {
			ops = append(ops, "post:"+body.Id)
			return nil, nil, nil
		},
		putCaseManagementCaseplanDataschemaAttr: func(ctx context.Context, pr *caseManagementCaseplanProxy, caseplanID, schemaKey string, body platformclientv2.Caseplandataschema) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error) {
			t.Fatal("put not expected for new schema id")
			return nil, nil, nil
		},
		deleteCaseManagementCaseplanDataschemaAttr: func(ctx context.Context, pr *caseManagementCaseplanProxy, caseplanID, schemaKey string) (*platformclientv2.APIResponse, error) {
			ops = append(ops, "del:"+schemaKey)
			return nil, nil
		},
	}
	oldRaw := []interface{}{map[string]interface{}{"id": "a"}}
	newRaw := []interface{}{map[string]interface{}{"id": "b"}}
	diags := execCaseplanDataSchemaSync(context.Background(), p, "cp", oldRaw, newRaw)
	assert.Empty(t, diags)
	assert.Equal(t, []string{"del:default", "post:b"}, ops)
}

func TestUnit_execCaseplanDataSchemaSync_sameIdNoNetworkCalls(t *testing.T) {
	t.Parallel()
	var ops []string
	p := &caseManagementCaseplanProxy{
		postCaseManagementCaseplanDataschemaAttr: func(ctx context.Context, pr *caseManagementCaseplanProxy, caseplanID string, body caseplanDataschemaPostBody) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error) {
			ops = append(ops, "post")
			return nil, nil, nil
		},
		putCaseManagementCaseplanDataschemaAttr: func(ctx context.Context, pr *caseManagementCaseplanProxy, caseplanID, schemaKey string, body platformclientv2.Caseplandataschema) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error) {
			ops = append(ops, "put")
			return nil, nil, nil
		},
		deleteCaseManagementCaseplanDataschemaAttr: func(ctx context.Context, pr *caseManagementCaseplanProxy, caseplanID, schemaKey string) (*platformclientv2.APIResponse, error) {
			ops = append(ops, "del")
			return nil, nil
		},
	}
	raw := []interface{}{map[string]interface{}{"id": "a"}}
	diags := execCaseplanDataSchemaSync(context.Background(), p, "cp", raw, raw)
	assert.Empty(t, diags)
	assert.Empty(t, ops)
}

func TestUnit_execCaseplanDataSchemaSync_rejectsMultipleNewBlocks(t *testing.T) {
	t.Parallel()
	oldRaw := []interface{}{map[string]interface{}{"id": "a"}}
	newRaw := []interface{}{
		map[string]interface{}{"id": "b"},
		map[string]interface{}{"id": "c"},
	}
	diags := execCaseplanDataSchemaSync(context.Background(), &caseManagementCaseplanProxy{}, "cp", oldRaw, newRaw)
	assert.NotEmpty(t, diags)
}

func TestUnit_execCaseplanDataSchemaSync_deleteThenPutFallbackWhenPutsEmpty(t *testing.T) {
	t.Parallel()
	var ops []string
	p := &caseManagementCaseplanProxy{
		postCaseManagementCaseplanDataschemaAttr: func(ctx context.Context, pr *caseManagementCaseplanProxy, caseplanID string, body caseplanDataschemaPostBody) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error) {
			t.Fatal("post not expected when remaining schema id was already bound")
			return nil, nil, nil
		},
		putCaseManagementCaseplanDataschemaAttr: func(ctx context.Context, pr *caseManagementCaseplanProxy, caseplanID, schemaKey string, body platformclientv2.Caseplandataschema) (*platformclientv2.Caseplandataschema, *platformclientv2.APIResponse, error) {
			ops = append(ops, "put:"+schemaKey+":"+*body.Id)
			return &body, nil, nil
		},
		deleteCaseManagementCaseplanDataschemaAttr: func(ctx context.Context, pr *caseManagementCaseplanProxy, caseplanID, schemaKey string) (*platformclientv2.APIResponse, error) {
			ops = append(ops, "del:"+schemaKey)
			return nil, nil
		},
	}
	oldRaw := []interface{}{
		map[string]interface{}{"id": "a"},
		map[string]interface{}{"id": "b"},
	}
	newRaw := []interface{}{map[string]interface{}{"id": "b"}}
	diags := execCaseplanDataSchemaSync(context.Background(), p, "cp", oldRaw, newRaw)
	assert.Empty(t, diags)
	assert.Equal(t, []string{"del:default", "put:default:b"}, ops)
}

func TestUnit_caseplanApplyPatchIfChanged(t *testing.T) {
	t.Parallel()
	var patchCalls int
	p := &caseManagementCaseplanProxy{
		patchCaseManagementCaseplanAttr: func(ctx context.Context, pr *caseManagementCaseplanProxy, caseplanId string, body platformclientv2.Caseplanupdate) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
			patchCalls++
			assert.Equal(t, "cp1", caseplanId)
			require.NotNil(t, body.Name)
			assert.Equal(t, "new-name", *body.Name)
			return &platformclientv2.Caseplan{}, nil, nil
		},
	}
	sch := ResourceCaseManagementCaseplan().Schema
	state := &terraform.InstanceState{
		Attributes: map[string]string{
			"name":             "old-name",
			"data_schema.#":    "1",
			"data_schema.0.id": "11111111-1111-1111-1111-111111111111",
		},
	}
	diff := &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"name": {Old: "old-name", New: "new-name"},
		},
	}
	d, err := schema.InternalMap(sch).Data(state, diff)
	require.NoError(t, err)
	d.SetId("cp1")

	diags := caseplanApplyPatchIfChanged(context.Background(), p, d, "cp1")
	assert.Empty(t, diags)
	assert.Equal(t, 1, patchCalls)
}

func TestUnit_caseplanApplyPatchIfChanged_noopWhenNoDiff(t *testing.T) {
	t.Parallel()
	var patchCalls int
	p := &caseManagementCaseplanProxy{
		patchCaseManagementCaseplanAttr: func(ctx context.Context, pr *caseManagementCaseplanProxy, caseplanId string, body platformclientv2.Caseplanupdate) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
			patchCalls++
			return nil, nil, nil
		},
	}
	sch := ResourceCaseManagementCaseplan().Schema
	state := &terraform.InstanceState{
		Attributes: map[string]string{
			"name":             "same",
			"data_schema.#":    "1",
			"data_schema.0.id": "11111111-1111-1111-1111-111111111111",
		},
	}
	diff := &terraform.InstanceDiff{Attributes: map[string]*terraform.ResourceAttrDiff{}}
	d, err := schema.InternalMap(sch).Data(state, diff)
	require.NoError(t, err)
	diags := caseplanApplyPatchIfChanged(context.Background(), p, d, "cp1")
	assert.Empty(t, diags)
	assert.Zero(t, patchCalls)
}

func TestUnit_caseplanApplyIntakePutIfChanged(t *testing.T) {
	t.Parallel()
	var putCalls int
	p := &caseManagementCaseplanProxy{
		putCaseManagementCaseplanIntakesettingsAttr: func(ctx context.Context, pr *caseManagementCaseplanProxy, caseplanId string, body platformclientv2.Intakesettingsupdate) (*platformclientv2.Intakesettingslisting, *platformclientv2.APIResponse, error) {
			putCalls++
			require.NotNil(t, body.IntakeSettings)
			require.Len(t, *body.IntakeSettings, 1)
			assert.Equal(t, "p2", *(*body.IntakeSettings)[0].Property)
			return &platformclientv2.Intakesettingslisting{}, nil, nil
		},
	}
	sch := ResourceCaseManagementCaseplan().Schema
	state := &terraform.InstanceState{
		Attributes: map[string]string{
			"name":                            "cp",
			"data_schema.#":                   "1",
			"data_schema.0.id":                "11111111-1111-1111-1111-111111111111",
			"intake_settings.#":               "1",
			"intake_settings.0.property":      "p1",
			"intake_settings.0.required":      "false",
			"intake_settings.0.display_order": "0",
		},
	}
	diff := &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"intake_settings.0.property": {Old: "p1", New: "p2"},
		},
	}
	d, err := schema.InternalMap(sch).Data(state, diff)
	require.NoError(t, err)
	d.SetId("cp1")

	diags := caseplanApplyIntakePutIfChanged(context.Background(), p, d, "cp1")
	assert.Empty(t, diags)
	assert.Equal(t, 1, putCalls)
}

func TestUnit_caseplanApplyIntakePutIfChanged_noop(t *testing.T) {
	t.Parallel()
	var putCalls int
	p := &caseManagementCaseplanProxy{
		putCaseManagementCaseplanIntakesettingsAttr: func(ctx context.Context, pr *caseManagementCaseplanProxy, caseplanId string, body platformclientv2.Intakesettingsupdate) (*platformclientv2.Intakesettingslisting, *platformclientv2.APIResponse, error) {
			putCalls++
			return nil, nil, nil
		},
	}
	sch := ResourceCaseManagementCaseplan().Schema
	state := &terraform.InstanceState{
		Attributes: map[string]string{
			"name":                            "cp",
			"data_schema.#":                   "1",
			"data_schema.0.id":                "11111111-1111-1111-1111-111111111111",
			"intake_settings.#":               "1",
			"intake_settings.0.property":      "p1",
			"intake_settings.0.required":      "false",
			"intake_settings.0.display_order": "0",
		},
	}
	diff := &terraform.InstanceDiff{Attributes: map[string]*terraform.ResourceAttrDiff{}}
	d, err := schema.InternalMap(sch).Data(state, diff)
	require.NoError(t, err)

	diags := caseplanApplyIntakePutIfChanged(context.Background(), p, d, "cp1")
	assert.Empty(t, diags)
	assert.Zero(t, putCalls)
}

func TestUnit_caseplanDiagsIfImmutableFieldsChangeAfterPublish_blocksWhenPublished(t *testing.T) {
	t.Parallel()
	pub := 1
	p := &caseManagementCaseplanProxy{
		getCaseManagementCaseplanByIdAttr: func(ctx context.Context, pr *caseManagementCaseplanProxy, id string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
			return &platformclientv2.Caseplan{Published: &pub}, nil, nil
		},
	}
	sch := ResourceCaseManagementCaseplan().Schema
	state := &terraform.InstanceState{
		Attributes: map[string]string{
			"name":                            "cp",
			"division_id":                     "div-a",
			"reference_prefix":                "AB12",
			"data_schema.#":                   "1",
			"data_schema.0.id":                "11111111-1111-1111-1111-111111111111",
			"customer_intent.#":               "1",
			"customer_intent.0.id":            "22222222-2222-2222-2222-222222222222",
			"intake_settings.#":               "1",
			"intake_settings.0.property":      "case_note_text",
			"intake_settings.0.required":      "false",
			"intake_settings.0.display_order": "0",
		},
	}
	diff := &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"division_id": {Old: "div-a", New: "div-b"},
		},
	}
	d, err := schema.InternalMap(sch).Data(state, diff)
	require.NoError(t, err)
	d.SetId("cp1")

	diags := caseplanDiagsIfImmutableFieldsChangeAfterPublish(context.Background(), p, d, "cp1")
	require.Len(t, diags, 1)
	assert.Contains(t, diags[0].Summary, "division_id")
}

func TestUnit_caseplanDiagsIfImmutableFieldsChangeAfterPublish_allowsWhenUnpublished(t *testing.T) {
	t.Parallel()
	pub := 0
	p := &caseManagementCaseplanProxy{
		getCaseManagementCaseplanByIdAttr: func(ctx context.Context, pr *caseManagementCaseplanProxy, id string) (*platformclientv2.Caseplan, *platformclientv2.APIResponse, error) {
			return &platformclientv2.Caseplan{Published: &pub}, nil, nil
		},
	}
	sch := ResourceCaseManagementCaseplan().Schema
	state := &terraform.InstanceState{
		Attributes: map[string]string{
			"name":             "cp",
			"division_id":      "div-a",
			"data_schema.#":    "1",
			"data_schema.0.id": "11111111-1111-1111-1111-111111111111",
		},
	}
	diff := &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"division_id": {Old: "div-a", New: "div-b"},
		},
	}
	d, err := schema.InternalMap(sch).Data(state, diff)
	require.NoError(t, err)
	d.SetId("cp1")

	diags := caseplanDiagsIfImmutableFieldsChangeAfterPublish(context.Background(), p, d, "cp1")
	assert.Empty(t, diags)
}
