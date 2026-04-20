package case_management_caseplan

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"
	"github.com/stretchr/testify/assert"
)

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
