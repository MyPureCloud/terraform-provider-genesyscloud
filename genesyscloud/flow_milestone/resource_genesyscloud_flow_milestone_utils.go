package flow_milestone

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_flow_outcome_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getFlowMilestoneFromResourceData maps data from schema ResourceData object to a platformclientv2.Flowmilestone
func getFlowMilestoneFromResourceData(d *schema.ResourceData) platformclientv2.Flowmilestone {
	divisionId := d.Get("division_id").(string)
	description := d.Get("description").(string)

	milestone := platformclientv2.Flowmilestone{
		Name: platformclientv2.String(d.Get("name").(string)),
	}
	if divisionId != "" {
		milestone.Division = &platformclientv2.Writabledivision{Id: &divisionId}
	}
	if description != "" {
		milestone.Description = &description
	}

	return milestone
}
