// Package wfm holds the helpers that the workforce management resources share.
package wfm

import (
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

// GetAllBusinessUnits retrieves every business unit in the organization.
// The empty strings are for feature (so no special permission checking overrides) and divisionId (so no filtering by division).
func GetAllBusinessUnits(api *platformclientv2.WorkforceManagementApi) (*[]platformclientv2.Businessunitlistitem, *platformclientv2.APIResponse, error) {
	listing, resp, err := api.GetWorkforcemanagementBusinessunits("", "")
	if err != nil {
		return nil, resp, err
	}

	return listing.Entities, resp, nil
}
