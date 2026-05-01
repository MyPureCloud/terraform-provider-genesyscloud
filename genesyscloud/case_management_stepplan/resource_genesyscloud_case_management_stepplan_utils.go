package case_management_stepplan

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

const stepplanResourceIDSeparator = "|"

func formatStepplanResourceID(caseplanID string, stageNumber int, stepplanID string) string {
	return fmt.Sprintf("%s%s%d%s%s", caseplanID, stepplanResourceIDSeparator, stageNumber, stepplanResourceIDSeparator, stepplanID)
}

func parseStepplanResourceID(id string) (caseplanID string, stageNumber int, stepplanID string, err error) {
	parts := strings.Split(id, stepplanResourceIDSeparator)
	if len(parts) != 3 {
		return "", 0, "", fmt.Errorf("invalid id %q: expected caseplan_id|stage_number|stepplan_id", id)
	}
	caseplanID = parts[0]
	if _, err = fmt.Sscanf(parts[1], "%d", &stageNumber); err != nil {
		return "", 0, "", fmt.Errorf("invalid stage_number in id: %w", err)
	}
	stepplanID = parts[2]
	if caseplanID == "" || stepplanID == "" || stageNumber < 1 {
		return "", 0, "", fmt.Errorf("invalid id %q", id)
	}
	return caseplanID, stageNumber, stepplanID, nil
}

func buildStepplanUpdate(d *schema.ResourceData) platformclientv2.Stepplanupdate {
	u := platformclientv2.Stepplanupdate{}
	if v, ok := d.GetOk("name"); ok {
		u.Name = platformclientv2.String(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		u.Description = platformclientv2.String(v.(string))
	}
	if v, ok := d.GetOk("activity_type"); ok {
		u.ActivityType = platformclientv2.String(v.(string))
	}
	ws := d.Get("workitem_settings").([]interface{})
	if len(ws) > 0 {
		m, ok := ws[0].(map[string]interface{})
		if ok {
			if wtid, ok := m["worktype_id"].(string); ok && wtid != "" {
				u.WorkitemSettings = &platformclientv2.Workitemsettings{
					WorktypeId: platformclientv2.String(wtid),
				}
			}
		}
	}
	return u
}

func flattenCaseplanReference(ref *platformclientv2.Caseplanreference) []interface{} {
	if ref == nil {
		return nil
	}
	m := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(m, "id", ref.Id)
	resourcedata.SetMapValueIfNotNil(m, "name", ref.Name)
	return []interface{}{m}
}

func flattenStageplanReference(ref *platformclientv2.Stageplanreference) []interface{} {
	if ref == nil {
		return nil
	}
	m := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(m, "id", ref.Id)
	resourcedata.SetMapValueIfNotNil(m, "name", ref.Name)
	return []interface{}{m}
}

func flattenWorkitemSettingsResponse(ws *platformclientv2.Workitemsettingsresponse) []interface{} {
	if ws == nil {
		return nil
	}
	m := make(map[string]interface{})
	if ws.Worktype != nil {
		resourcedata.SetMapValueIfNotNil(m, "worktype_id", ws.Worktype.Id)
		resourcedata.SetMapValueIfNotNil(m, "worktype_name", ws.Worktype.Name)
	}
	return []interface{}{m}
}
