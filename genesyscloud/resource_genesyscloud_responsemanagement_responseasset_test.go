package genesyscloud

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v89/platformclientv2"
)

func TestAccResponseManagementResponseAsset(t *testing.T) {
	t.Parallel()
	var (
		resourceId         = "responseasset"
		testFilesDir       = "test_responseasset_data"
		fileName1          = "yeti-img.png"
		fileName2          = "genesys-img.png"
		fullPath1          = fmt.Sprintf("%s/%s", testFilesDir, fileName1)
		fullPath2          = fmt.Sprintf("%s/%s", testFilesDir, fileName2)
		divisionResourceId = "test_div"
		divisionName       = "test tf divison " + uuid.NewString()
	)

	defer func() {
		err := cleanupResponseAssets(testFilesDir)
		if err != nil {
			log.Printf("error cleaning up response assets: %v. Dangling assets may exist.", err)
		}
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateResponseManagementResponseAssetResource(resourceId, fullPath1, nullValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_responseasset."+resourceId, "filename", fullPath1),
					testDefaultHomeDivision("genesyscloud_responsemanagement_responseasset."+resourceId),
				),
			},
			{
				Config: generateResponseManagementResponseAssetResource(resourceId, fullPath2, "genesyscloud_auth_division."+divisionResourceId+".id") +
					generateAuthDivisionBasic(divisionResourceId, divisionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_responseasset."+resourceId, "filename", fullPath2),
					resource.TestCheckResourceAttrPair("genesyscloud_responsemanagement_responseasset."+resourceId, "division_id",
						"genesyscloud_auth_division."+divisionResourceId, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_responsemanagement_responseasset." + resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyResponseAssetDestroyed,
	})
}

func generateResponseManagementResponseAssetResource(resourceId string, fileName string, divisionId string) string {
	return fmt.Sprintf(`
resource "genesyscloud_responsemanagement_responseasset" "%s" {
    filename    = "%s"
    division_id = %s
}
`, resourceId, fileName, divisionId)
}

func cleanupResponseAssets(folderName string) error {
	var (
		name    = "name"
		fields  = []string{name}
		varType = "STARTS_WITH"
	)
	config := platformclientv2.GetDefaultConfiguration()
	err := config.AuthorizeClientCredentials(os.Getenv("GENESYSCLOUD_OAUTHCLIENT_ID"), os.Getenv("GENESYSCLOUD_OAUTHCLIENT_SECRET"))
	if err != nil {
		return err
	}
	respManagementApi := platformclientv2.NewResponseManagementApiWithConfig(config)

	var filter = platformclientv2.Responseassetfilter{
		Fields:  &fields,
		Value:   &folderName,
		VarType: &varType,
	}

	var body = platformclientv2.Responseassetsearchrequest{
		Query:  &[]platformclientv2.Responseassetfilter{filter},
		SortBy: &name,
	}

	responseData, _, err := respManagementApi.PostResponsemanagementResponseassetsSearch(body, nil)
	if err != nil {
		return err
	}

	if responseData.Results != nil && len(*responseData.Results) > 0 {
		for _, result := range *responseData.Results {
			_, err = respManagementApi.DeleteResponsemanagementResponseasset(*result.Id)
			if err != nil {
				log.Printf("Failed to delete response assets %s: %v", *result.Id, err)
			}
		}
	}
	return nil
}

func testVerifyResponseAssetDestroyed(state *terraform.State) error {
	responseManagementAPI := platformclientv2.NewResponseManagementApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_responsemanagement_responseasset" {
			continue
		}
		responseAsset, resp, err := responseManagementAPI.GetResponsemanagementResponseasset(rs.Primary.ID)
		if responseAsset != nil {
			return fmt.Errorf("response asset (%s) still exists", rs.Primary.ID)
		} else if isStatus404(resp) {
			// response asset not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All response assets destroyed
	return nil
}
