package case_management_stepplan

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitFormatParseStepplanResourceID(t *testing.T) {
	t.Parallel()
	cp := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	sp := "pppppppp-pppp-pppp-pppp-pppppppppppp"
	id := formatStepplanResourceID(cp, 3, sp)
	assert.Equal(t, cp+"|3|"+sp, id)

	gotCP, gotN, gotSp, err := parseStepplanResourceID(id)
	assert.NoError(t, err)
	assert.Equal(t, cp, gotCP)
	assert.Equal(t, 3, gotN)
	assert.Equal(t, sp, gotSp)

	_, _, _, err = parseStepplanResourceID("x|y")
	assert.Error(t, err)
}

func TestUnitBuildStepplanUpdate(t *testing.T) {
	t.Parallel()
	sch := ResourceCaseManagementStepplan().Schema
	d := schema.TestResourceDataRaw(t, sch, map[string]interface{}{
		"caseplan_id":      "cp-1",
		"stage_number":     1,
		"stepplan_id":      "stp-1",
		"stageplan_id":     "stg-1",
		"name":             "Step n",
		"description":      "d",
		"activity_type":    "Workitem",
		"caseplan":         []interface{}{},
		"stageplan":        []interface{}{},
		"workitem_settings": []interface{}{
			map[string]interface{}{"worktype_id": "wt-1"},
		},
	})
	u := buildStepplanUpdate(d)
	assert.Equal(t, "Step n", *u.Name)
	assert.Equal(t, "d", *u.Description)
	assert.Equal(t, "Workitem", *u.ActivityType)
	assert.NotNil(t, u.WorkitemSettings)
	assert.Equal(t, "wt-1", *u.WorkitemSettings.WorktypeId)
}

func TestUnitFlattenStepplanRefs(t *testing.T) {
	t.Parallel()
	assert.Nil(t, flattenCaseplanReference(nil))
	assert.Nil(t, flattenStageplanReference(nil))
	assert.Nil(t, flattenWorkitemSettingsResponse(nil))

	cid, cn := "c-1", "C"
	sid, sn := "s-1", "S"
	wtid, wtn := "w-1", "WT"

	cpOut := flattenCaseplanReference(&platformclientv2.Caseplanreference{Id: &cid, Name: &cn})
	assert.Equal(t, "c-1", cpOut[0].(map[string]interface{})["id"])

	stOut := flattenStageplanReference(&platformclientv2.Stageplanreference{Id: &sid, Name: &sn})
	assert.Equal(t, "s-1", stOut[0].(map[string]interface{})["id"])

	wsOut := flattenWorkitemSettingsResponse(&platformclientv2.Workitemsettingsresponse{
		Worktype: &platformclientv2.Stepplansworktypereference{Id: &wtid, Name: &wtn},
	})
	m := wsOut[0].(map[string]interface{})
	assert.Equal(t, "w-1", m["worktype_id"])
	assert.Equal(t, "WT", m["worktype_name"])
}
