package workforcemanagement_business_unit_activitycode

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
)

// SplitBusinessUnitActivityCodeTerraformId splits the resource id, which is in the form
// <businessUnitId>/<activityCodeId>, into the businessUnitId and activityCodeId
func SplitBusinessUnitActivityCodeTerraformId(id string) (businessUnitId string, activityCodeId string) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		return id, ""
	}
	return parts[0], parts[1]
}

// isActive reports whether the activity code is still active. Deleting an activity code marks it inactive rather than removing it.
func isActive(activityCode *platformclientv2.Businessunitactivitycode) bool {
	return activityCode.Active == nil || *activityCode.Active
}

// isAgentTimeOffSelectableConfigured reports whether agent_time_off_selectable is set in the configuration,
// so an explicit false can be told apart from an unset value
func isAgentTimeOffSelectableConfigured(d *schema.ResourceData) bool {
	return !d.GetRawConfig().GetAttr("agent_time_off_selectable").IsNull()
}

func buildSecondaryPresences(d *schema.ResourceData) *[]platformclientv2.Secondarypresence {
	presenceIds := lists.SetToStringList(d.Get("secondary_presence_ids").(*schema.Set))
	secondaryPresences := make([]platformclientv2.Secondarypresence, 0)
	if presenceIds != nil {
		for _, id := range *presenceIds {
			secondaryPresences = append(secondaryPresences, platformclientv2.Secondarypresence{Id: platformclientv2.String(id)})
		}
	}
	return &secondaryPresences
}

func buildCreateActivityCodeRequest(d *schema.ResourceData) *platformclientv2.Createactivitycoderequest {
	request := &platformclientv2.Createactivitycoderequest{
		Name:                  platformclientv2.String(d.Get("name").(string)),
		Category:              platformclientv2.String(d.Get("category").(string)),
		CountsAsPaidTime:      platformclientv2.Bool(d.Get("counts_as_paid_time").(bool)),
		CountsAsWorkTime:      platformclientv2.Bool(d.Get("counts_as_work_time").(bool)),
		CountsTowardShrinkage: platformclientv2.Bool(d.Get("counts_toward_shrinkage").(bool)),
		PlannedShrinkage:      platformclientv2.Bool(d.Get("planned_shrinkage").(bool)),
		Interruptible:         platformclientv2.Bool(d.Get("interruptible").(bool)),
		SecondaryPresences:    buildSecondaryPresences(d),
	}
	if lengthInMinutes, ok := d.GetOk("length_in_minutes"); ok {
		request.LengthInMinutes = platformclientv2.Int(lengthInMinutes.(int))
	}
	if isAgentTimeOffSelectableConfigured(d) {
		request.AgentTimeOffSelectable = platformclientv2.Bool(d.Get("agent_time_off_selectable").(bool))
	}
	return request
}

func buildUpdateActivityCodeRequest(d *schema.ResourceData, metadata *platformclientv2.Wfmversionedentitymetadata) *platformclientv2.Updateactivitycoderequest {
	request := &platformclientv2.Updateactivitycoderequest{
		Name:                  platformclientv2.String(d.Get("name").(string)),
		Category:              platformclientv2.String(d.Get("category").(string)),
		CountsAsPaidTime:      platformclientv2.Bool(d.Get("counts_as_paid_time").(bool)),
		CountsAsWorkTime:      platformclientv2.Bool(d.Get("counts_as_work_time").(bool)),
		CountsTowardShrinkage: platformclientv2.Bool(d.Get("counts_toward_shrinkage").(bool)),
		PlannedShrinkage:      platformclientv2.Bool(d.Get("planned_shrinkage").(bool)),
		Interruptible:         platformclientv2.Bool(d.Get("interruptible").(bool)),
		SecondaryPresences:    &platformclientv2.Listwrappersecondarypresence{Values: buildSecondaryPresences(d)},
	}
	if lengthInMinutes, ok := d.GetOk("length_in_minutes"); ok {
		request.LengthInMinutes = platformclientv2.Int(lengthInMinutes.(int))
	}
	if isAgentTimeOffSelectableConfigured(d) {
		request.AgentTimeOffSelectable = platformclientv2.Bool(d.Get("agent_time_off_selectable").(bool))
	}
	if metadata != nil && metadata.Version != nil {
		request.Metadata = &platformclientv2.Wfmversionedentitymetadata{Version: metadata.Version}
	}
	return request
}

func flattenSecondaryPresenceIds(secondaryPresences *[]platformclientv2.Secondarypresence) *schema.Set {
	ids := make([]interface{}, 0)
	if secondaryPresences != nil {
		for _, presence := range *secondaryPresences {
			if presence.Id != nil {
				ids = append(ids, *presence.Id)
			}
		}
	}
	return schema.NewSet(schema.HashString, ids)
}

// GenerateBusinessUnitActivityCodeResource returns terraform config for a business unit activity code, used in tests
func GenerateBusinessUnitActivityCodeResource(resourceLabel string, businessUnitId string, name string, category string, attrs ...string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
	business_unit_id = %s
	name             = "%s"
	category         = "%s"
	%s
}
`, ResourceType, resourceLabel, businessUnitId, name, category, strings.Join(attrs, "\n"))
}
