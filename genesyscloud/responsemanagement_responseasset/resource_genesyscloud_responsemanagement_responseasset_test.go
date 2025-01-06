package responsemanagement_responseasset

import (
	"fmt"
	"log"
	"path/filepath"
	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

func TestAccResourceResponseManagementResponseAsset(t *testing.T) {
	var (
		resourceLabel         = "responseasset"
		testFilesDir          = "test_responseasset_data"
		fileName1             = "yeti-img.png"
		fileName2             = "genesys-img.png"
		fullPath1             = filepath.Join(testFilesDir, fileName1)
		fullPath2             = filepath.Join(testFilesDir, fileName2)
		normalizeFileName1, _ = testrunner.NormalizeFileName(fullPath1)
		normalizeFileName2, _ = testrunner.NormalizeFileName(fullPath2)
		divisionResourceLabel = "test_div"
		divisionName          = "test tf divison " + uuid.NewString()
	)

	cleanupResponseAssets("genesys")
	cleanupResponseAssets("yeti")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateResponseManagementResponseAssetResource(resourceLabel, fullPath1, util.NullValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_responseasset."+resourceLabel, "filename", normalizeFileName1),
					provider.TestDefaultHomeDivision("genesyscloud_responsemanagement_responseasset."+resourceLabel),
				),
			},
			{
				Config: GenerateResponseManagementResponseAssetResource(resourceLabel, fullPath2, "genesyscloud_auth_division."+divisionResourceLabel+".id") +
					authDivision.GenerateAuthDivisionBasic(divisionResourceLabel, divisionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_responseasset."+resourceLabel, "filename", normalizeFileName2),
					resource.TestCheckResourceAttrPair("genesyscloud_responsemanagement_responseasset."+resourceLabel, "division_id",
						"genesyscloud_auth_division."+divisionResourceLabel, "id"),
				),
			},
			// Update
			{
				Config: GenerateResponseManagementResponseAssetResource(resourceLabel, fullPath2, "data.genesyscloud_auth_division_home.home.id") +
					fmt.Sprint("\ndata \"genesyscloud_auth_division_home\" \"home\" {}\n"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_responseasset."+resourceLabel, "filename", normalizeFileName2),
					provider.TestDefaultHomeDivision("genesyscloud_responsemanagement_responseasset."+resourceLabel),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_responsemanagement_responseasset." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"file_content_hash",
				},
			},
		},
		CheckDestroy: testVerifyResponseAssetDestroyed,
	})
}

func cleanupResponseAssets(folderName string) error {
	var (
		name    = "name"
		fields  = []string{name}
		varType = "STARTS_WITH"
	)
	config, err := provider.AuthorizeSdk()
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
		} else if util.IsStatus404(resp) {
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
