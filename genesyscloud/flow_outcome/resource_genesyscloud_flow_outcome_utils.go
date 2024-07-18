package flow_outcome

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_flow_outcome_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getFlowOutcomeFromResourceData maps data from schema ResourceData object to a platformclientv2.Flowoutcome
func getFlowOutcomeFromResourceData(d *schema.ResourceData) platformclientv2.Flowoutcome {
	divisionId := d.Get("division_id").(string)
	description := d.Get("description").(string)

	outcome := platformclientv2.Flowoutcome{
		Name: platformclientv2.String(d.Get("name").(string)),
	}

	if divisionId != "" {
		outcome.Division = &platformclientv2.Writabledivision{Id: &divisionId}
	}
	if description != "" {
		outcome.Description = &description
	}

	return outcome
}
