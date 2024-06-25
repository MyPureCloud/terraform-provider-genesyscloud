package provider

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

type JsonMap map[string]interface{}

// Attempt to get the home division once during a provider run
var divOnce sync.Once
var homeDivID string
var homeDivErr diag.Diagnostics

func getHomeDivisionID() (string, diag.Diagnostics) {
	divOnce.Do(func() {
		authAPI := platformclientv2.NewAuthorizationApi()
		homeDiv, _, err := authAPI.GetAuthorizationDivisionsHome()
		if err != nil {
			homeDivErr = diag.Errorf("Failed to query home division: %s", err)
			return
		}
		homeDivID = *homeDiv.Id
	})

	if homeDivErr != nil {
		return "", homeDivErr
	}
	return homeDivID, nil
}
