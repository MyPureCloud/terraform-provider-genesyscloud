package genesyscloud

import (
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccDataSourceAuthDivision(t *testing.T) {
	var (
		divResource   = "auth-division"
		divDataSource = "auth-div-data"
		divName       = "Terraform Divisions-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
				Config: GenerateAuthDivisionResource(
					divResource,
					divName,
					util.NullValue,
					util.NullValue,
				) + generateAuthDivisionDataSource(
					divDataSource,
					"genesyscloud_auth_division."+divResource+".name",
					"genesyscloud_auth_division."+divResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_auth_division."+divDataSource, "id", "genesyscloud_auth_division."+divResource, "id"),
					checkDivisionDeleted(divisionID),
				),
			},
		},
		CheckDestroy: testVerifyDivisionsDestroyed,
	})
}

func generateAuthDivisionDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_auth_division" "%s" {
		name = %s
        depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}

func checkDivisionDeleted(id string) resource.TestCheckFunc {
	log.Printf("Fetching division with ID: %s\n", id)
	return func(s *terraform.State) error {
		maxAttempts := 18
		for i := 0; i < maxAttempts; i++ {

			deleted, err := isDivisionDeleted(id)
			if err != nil {
				return err
			}
			if deleted {
				return nil
			}
			time.Sleep(10 * time.Second)
		}
		return fmt.Errorf("division %s was not deleted properly", id)
	}
}

func isDivisionDeleted(id string) (bool, error) {
	mu.Lock()
	defer mu.Unlock()

	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)
	// Attempt to get the division
	_, response, err := authAPI.GetAuthorizationDivision(id, false)

	// Check if the division is not found (deleted)
	if response != nil && response.StatusCode == 404 {
		return true, nil // division is deleted
	}

	// Handle other errors
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		return false, err
	}

	// If division is found, it means the division is not deleted
	return false, nil
}
