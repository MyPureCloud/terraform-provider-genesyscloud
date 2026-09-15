package case_management_stageplan

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

const stageplanResourceIDSeparator = "|"

// FormatStageplanResourceID builds the case_management_stageplan composite resource ID, which
// is always in the format <caseplan-id>|<stage-number>|<stageplan-id>. Exported so that
// external consumers of this provider's resources don't need to duplicate this logic themselves.
func FormatStageplanResourceID(caseplanID string, stageNumber int, stageplanID string) (id string) {
	return fmt.Sprintf("%s%s%d%s%s", caseplanID, stageplanResourceIDSeparator, stageNumber, stageplanResourceIDSeparator, stageplanID)
}

// ParseStageplanResourceID splits the case_management_stageplan composite resource ID, which
// is always in the format <caseplan-id>|<stage-number>|<stageplan-id>, back into its parts.
// Exported so that external consumers of this provider's resources don't need to duplicate
// this logic themselves.
func ParseStageplanResourceID(id string) (caseplanID string, stageNumber int, stageplanID string, err error) {
	parts := strings.Split(id, stageplanResourceIDSeparator)
	if len(parts) != 3 {
		return "", 0, "", fmt.Errorf("invalid id %q: expected caseplan_id|stage_number|stageplan_id", id)
	}
	caseplanID = parts[0]
	if _, err = fmt.Sscanf(parts[1], "%d", &stageNumber); err != nil {
		return "", 0, "", fmt.Errorf("invalid stage_number in id: %w", err)
	}
	stageplanID = parts[2]
	if caseplanID == "" || stageplanID == "" || stageNumber < 1 {
		return "", 0, "", fmt.Errorf("invalid id %q", id)
	}
	return caseplanID, stageNumber, stageplanID, nil
}

func buildStageplanUpdate(d *schema.ResourceData) platformclientv2.Stageplanupdate {
	u := platformclientv2.Stageplanupdate{}
	if v, ok := d.GetOk("name"); ok {
		u.Name = platformclientv2.String(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		u.Description = platformclientv2.String(v.(string))
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
