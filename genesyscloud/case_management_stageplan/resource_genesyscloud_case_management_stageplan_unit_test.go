package case_management_stageplan

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitFormatParseStageplanResourceID(t *testing.T) {
	t.Parallel()
	cp := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	st := "ssssssss-ssss-ssss-ssss-ssssssssssss"
	id := formatStageplanResourceID(cp, 2, st)
	assert.Equal(t, cp+"|2|"+st, id)

	gotCP, gotN, gotSt, err := parseStageplanResourceID(id)
	assert.NoError(t, err)
	assert.Equal(t, cp, gotCP)
	assert.Equal(t, 2, gotN)
	assert.Equal(t, st, gotSt)

	_, _, _, err = parseStageplanResourceID("bad")
	assert.Error(t, err)
}

func TestUnitBuildStageplanUpdate(t *testing.T) {
	t.Parallel()
	sch := ResourceCaseManagementStageplan().Schema
	d := schema.TestResourceDataRaw(t, sch, map[string]interface{}{
		"caseplan_id":  "cp-1",
		"stage_number": 1,
		"name":         "Patched stage",
		"description":  "d",
		"stageplan_id": "st-1",
		"caseplan":     []interface{}{},
	})
	u := buildStageplanUpdate(d)
	assert.Equal(t, "Patched stage", *u.Name)
	assert.Equal(t, "d", *u.Description)
}

func TestUnitFlattenCaseplanReferenceStageplan(t *testing.T) {
	t.Parallel()
	assert.Nil(t, flattenCaseplanReference(nil))
	cid := "c-1"
	cn := "Case"
	out := flattenCaseplanReference(&platformclientv2.Caseplanreference{Id: &cid, Name: &cn})
	assert.Len(t, out, 1)
	m := out[0].(map[string]interface{})
	assert.Equal(t, "c-1", m["id"])
	assert.Equal(t, "Case", m["name"])
}
