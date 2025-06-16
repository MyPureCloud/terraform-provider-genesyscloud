package guide_version

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
)

/*
The resource_genesyscloud_guide_version_test.go contains all of the test cases for running the resource
tests for guide_version.
*/

func TestAccResourceGuideVersion(t *testing.T) {
	t.Parallel()
	var (
		guideResourceLabel = "guide"
		guideName          = "Test Guide " + uuid.NewString()

		guideVersionResourceLabel = "guide-version"
		instruction               = "This is a test instruction for the guide version."
		updatedInstruction        = "This is an updated test instruction for the guide version."
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create a guide and guide version
				Config: generateGuideResource(
					guideResourceLabel,
					guideName,
				) + generateGuideVersionResource(
					guideVersionResourceLabel,
					guideResourceLabel,
					instruction,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_guide_version."+guideVersionResourceLabel, "guide_id", "genesyscloud_guide."+guideResourceLabel, "id"),
					resource.TestCheckResourceAttr("genesyscloud_guide_version."+guideVersionResourceLabel, "instruction", instruction),
				),
			},
			{
				// Update guide version
				Config: generateGuideResource(
					guideResourceLabel,
					guideName,
				) + generateGuideVersionResource(
					guideVersionResourceLabel,
					guideResourceLabel,
					updatedInstruction,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_guide_version."+guideVersionResourceLabel, "guide_id", "genesyscloud_guide."+guideResourceLabel, "id"),
					resource.TestCheckResourceAttr("genesyscloud_guide_version."+guideVersionResourceLabel, "instruction", updatedInstruction),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_guide_version." + guideVersionResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyGuideVersionDestroyed,
	})
}

func generateGuideResource(resourceID string, name string) string {
	return fmt.Sprintf(`resource "genesyscloud_guide" "%s" {
  name = "%s"
}`, resourceID, name)
}

func generateGuideVersionResource(resourceID string, guideResourceID string, instruction string) string {
	return fmt.Sprintf(`resource "genesyscloud_guide_version" "%s" {
  guide_id = genesyscloud_guide.%s.id
  instruction = "%s"
}`, resourceID, guideResourceID, instruction)
}

func testVerifyGuideVersionDestroyed(state *terraform.State) error {
	//guidesAPI := platformclientv2.NewGuidesApi()
	//for _, rs := range state.RootModule().Resources {
	//	if rs.Type != "genesyscloud_guide_version" {
	//		continue
	//	}
	//
	//	guideVersion, resp, err := guidesAPI.GetGuideVersion(rs.Primary.ID)
	//	if guideVersion != nil {
	//		return fmt.Errorf("guide version (%s) still exists", rs.Primary.ID)
	//	} else if util.IsStatus404(resp) {
	//		// Guide version not found as expected
	//		continue
	//	} else {
	//		// Unexpected error
	//		return fmt.Errorf("unexpected error: %s", err)
	//	}
	//}
	//// Success. All guide versions destroyed
	return nil
}
