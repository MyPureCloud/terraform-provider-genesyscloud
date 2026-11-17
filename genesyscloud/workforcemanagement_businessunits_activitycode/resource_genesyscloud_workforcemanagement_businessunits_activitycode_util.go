package workforcemanagement_businessunits_activitycode

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
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
	return *activityCode.Active
}

// activityCodeLabel names an activity code for the exporter. A default activity code that has not been renamed has an empty name, so it is named after its category.
func activityCodeLabel(activityCode platformclientv2.Businessunitactivitycode) string {
	if *activityCode.Name == "" {
		return "default_" + *activityCode.Category
	}
	return *activityCode.Name
}

// activityCodeBlockHash keeps the block labels of exported activity codes unique and stable.
// An inactive code can share its name with an active one, so both would get the same label, which Terraform does not allow.
// The exporter's fallbacks for equal labels either add a suffix to whichever duplicate it reaches second, which can change from one export to the next,
// or flag them as duplicates. Giving the inactive code a hash makes its label differ from the active code's, and stay the same every time.
// Active codes get no hash because their names are unique, so their labels do not change.
// The hash leaves out the id, so the same code gets the same label in other exports and in other organizations.
func activityCodeBlockHash(activityCode platformclientv2.Businessunitactivitycode) (string, error) {
	if isActive(&activityCode) {
		return "", nil
	}
	return util.QuickHashFields(*activityCode.Name, *activityCode.Category, "inactive")
}

// findActivityCodeIdByName returns the id of the active activity code with the name. Default activity codes that have not been renamed all have an empty name,
// so they cannot be told apart by name and are found by category instead.
func findActivityCodeIdByName(activityCodes []platformclientv2.Businessunitactivitycode, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("an activity code with an empty name is a default activity code, so look it up by category")
	}
	for _, activityCode := range activityCodes {
		if *activityCode.Name == name && isActive(&activityCode) {
			return *activityCode.Id, nil
		}
	}
	return "", fmt.Errorf("unable to find an activity code with name %s", name)
}

// findDefaultActivityCodeIdByCategory returns the id of the default activity code of the category, of which there is one in each business unit
func findDefaultActivityCodeIdByCategory(activityCodes []platformclientv2.Businessunitactivitycode, category string) (string, error) {
	for _, activityCode := range activityCodes {
		if *activityCode.DefaultCode && *activityCode.Category == category && isActive(&activityCode) {
			return *activityCode.Id, nil
		}
	}
	return "", fmt.Errorf("unable to find a default activity code with category %s", category)
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
	if metadata.Version != nil {
		request.Metadata = &platformclientv2.Wfmversionedentitymetadata{Version: metadata.Version}
	}
	return request
}

func flattenSecondaryPresenceIds(secondaryPresences *[]platformclientv2.Secondarypresence) *schema.Set {
	ids := make([]interface{}, 0)
	for _, presence := range *secondaryPresences {
		ids = append(ids, *presence.Id)
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
