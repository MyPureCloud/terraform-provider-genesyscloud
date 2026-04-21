package case_management_caseplan

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_caseplanDataSchemasFromResourceList_coercesFloatVersion(t *testing.T) {
	t.Parallel()
	raw := []interface{}{
		map[string]interface{}{"id": "11111111-1111-1111-1111-111111111111", "version": float64(3)},
	}
	out := caseplanDataSchemasFromResourceList(raw)
	require.Len(t, out, 1)
	assert.Equal(t, 3, *out[0].Version)
}

func TestUnitFlattenExpandCaseplanDataSchemas(t *testing.T) {
	t.Parallel()
	id1 := "11111111-1111-1111-1111-111111111111"
	v1 := 2
	flat := flattenCaseplanDataSchemas(&[]platformclientv2.Caseplandataschema{
		{Id: platformclientv2.String(id1), Version: platformclientv2.Int(v1)},
	})
	assert.Len(t, flat, 1)
	m := flat[0].(map[string]interface{})
	assert.Equal(t, id1, m["id"])
	assert.Equal(t, v1, m["version"])

	assert.Nil(t, flattenCaseplanDataSchemas(nil))
	assert.Nil(t, flattenCaseplanDataSchemas(&[]platformclientv2.Caseplandataschema{}))
}

func TestUnitCaseplanVersionForDataschemaRead(t *testing.T) {
	t.Parallel()
	latest := 5
	pub := 3
	assert.Equal(t, "5", caseplanVersionForDataschemaRead(&platformclientv2.Caseplan{Latest: &latest}))
	assert.Equal(t, "3", caseplanVersionForDataschemaRead(&platformclientv2.Caseplan{Published: &pub}))
	assert.Equal(t, "", caseplanVersionForDataschemaRead(&platformclientv2.Caseplan{}))
	assert.Equal(t, "", caseplanVersionForDataschemaRead(nil))
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
		"customer_intent": []interface{}{
			map[string]interface{}{"id": "intent-1"},
		},
		"default_case_owner": []interface{}{
			map[string]interface{}{"id": "user-1"},
		},
		"data_schema": []interface{}{
			map[string]interface{}{"id": "schema-1", "version": 7},
		},
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
	ds := body.DataSchemas
	assert.NotNil(t, ds)
	assert.Len(t, *ds, 1)
	assert.Equal(t, "schema-1", *(*ds)[0].Id)
	assert.Equal(t, 7, *(*ds)[0].Version)
	isettings := body.IntakeSettings
	assert.NotNil(t, isettings)
	assert.Len(t, *isettings, 1)
	assert.Equal(t, "case_note_text", *(*isettings)[0].Property)
	assert.True(t, *(*isettings)[0].Required)
	assert.Equal(t, 1, *(*isettings)[0].DisplayOrder)
}

func TestUnitFlattenExpandCaseplanIntakeSettings(t *testing.T) {
	t.Parallel()
	prop := "p1"
	req := true
	ord := 2
	flat := flattenCaseplanIntakeSettings(&[]platformclientv2.Intakesetting{
		{Property: &prop, Required: &req, DisplayOrder: &ord},
	})
	assert.Len(t, flat, 1)
	m := flat[0].(map[string]interface{})
	assert.Equal(t, "p1", m["property"])
	assert.Equal(t, true, m["required"])
	assert.Equal(t, 2, m["display_order"])

	assert.Len(t, flattenCaseplanIntakeSettings(nil), 0)
	empty := []platformclientv2.Intakesetting{}
	assert.Len(t, flattenCaseplanIntakeSettings(&empty), 0)

	sch := ResourceCaseManagementCaseplan().Schema
	d := schema.TestResourceDataRaw(t, sch, map[string]interface{}{
		"data_schema": []interface{}{
			map[string]interface{}{"id": "schema-1", "version": 1},
		},
		"intake_settings": []interface{}{
			map[string]interface{}{"property": "a", "required": false, "display_order": 0},
		},
	})
	put := expandCaseplanIntakeSettingsForPut(d)
	assert.Len(t, *put, 1)
	assert.Equal(t, "a", *(*put)[0].Property)
	assert.False(t, *(*put)[0].Required)
	assert.Equal(t, 0, *(*put)[0].DisplayOrder)
}

func TestUnitFlattenUserAndIntentRefs(t *testing.T) {
	t.Parallel()
	assert.Nil(t, flattenUserReference(nil))
	assert.Nil(t, flattenCustomerIntentReference(nil))

	uid := "u-1"
	inid := "i-1"
	ur := flattenUserReference(&platformclientv2.Userreference{Id: &uid})
	assert.Len(t, ur, 1)
	assert.Equal(t, "u-1", ur[0].(map[string]interface{})["id"])

	ir := flattenCustomerIntentReference(&platformclientv2.Customerintentreference{Id: &inid})
	assert.Len(t, ir, 1)
	assert.Equal(t, "i-1", ir[0].(map[string]interface{})["id"])
}
