package util

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v204/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/homedivision"
)

type JsonMap map[string]interface{}

// HomeDivisionFetcher looks up the home division ID of the org that sdkConfig is authorised against.
type HomeDivisionFetcher = homedivision.Fetcher

func GetHomeDivisionName(key string, divisionName *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		homeDivision, ok := state.RootModule().Resources[key]
		if !ok {
			return fmt.Errorf("Failed to find home division")
		}
		*divisionName = homeDivision.Primary.Attributes["name"]
		return nil
	}
}

// GetHomeDivisionID returns the home division ID for the org the SDK default configuration is
// authorised against. Prefer GetHomeDivisionIDWithConfig wherever a client config is in scope.
func GetHomeDivisionID() (string, diag.Diagnostics) {
	return GetHomeDivisionIDWithConfig(platformclientv2.GetDefaultConfiguration())
}

// GetHomeDivisionIDWithConfig returns the home division ID for the org that sdkConfig is authorised
// against. Results are cached per org credential, so callers working against different orgs in the
// same process each get their own org's home division. A nil sdkConfig is an error rather than a
// fallback to the default configuration, which may belong to a different org.
func GetHomeDivisionIDWithConfig(sdkConfig *platformclientv2.Configuration) (string, diag.Diagnostics) {
	id, err := homedivision.ID(sdkConfig)
	if err != nil {
		return "", diag.Errorf("Failed to query home division: %s", err)
	}
	return id, nil
}

// SetHomeDivisionFetcherForTesting replaces the home division lookup and clears the cache. It returns a
// function that restores the real lookup and clears the cache again. For unit tests only.
func SetHomeDivisionFetcherForTesting(fetcher HomeDivisionFetcher) (restore func()) {
	return homedivision.SetFetcherForTesting(fetcher)
}

func UpdateObjectDivision(d *schema.ResourceData, objType string, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	if d.HasChange("division_id") {
		authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)
		divisionID := d.Get("division_id").(string)
		if divisionID == "" {
			// Default to home division
			homeDivision, diagErr := GetHomeDivisionIDWithConfig(sdkConfig)
			if diagErr != nil {
				return diagErr
			}
			divisionID = homeDivision
		}
		log.Printf("Updating division for %s %s to %s", objType, d.Id(), divisionID)
		_, divErr := authAPI.PostAuthorizationDivisionObject(divisionID, objType, []string{d.Id()})
		if divErr != nil {
			return diag.Errorf("Failed to update division for %s %s: %s", objType, d.Id(), divErr)
		}
	}
	return nil
}
