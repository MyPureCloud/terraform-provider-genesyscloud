package guide

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
)

func TestAccResourceGuide(t *testing.T) {
	var (
		resourceLabel = "guide"

		name   = "Test Guide Manual" + uuid.NewString()
		source = "Manual"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateGuideResource(
					resourceLabel,
					name,
					source,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_guide."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_guide."+resourceLabel, "source", source),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_guide." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyGuideDestroyed,
	})
}

func testVerifyGuideDestroyed(state *terraform.State) error {
	sdkConfig := provider.GetProviderMeta().ClientConfig
	proxy := getGuideProxy(sdkConfig)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_guide" {
			continue
		}
		guide, resp, err := proxy.getGuideById(context.Background(), rs.Primary.ID)
		if guide != nil {
			return fmt.Errorf("guide (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			continue
		} else {
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	return nil
}

// GenerateGuideResource generates terraform for a guide resource
func GenerateGuideResource(resourceID string, name string, source string) string {
	return fmt.Sprintf(`resource "genesyscloud_guide" "%s" {
		name = "%s"
		source = "%s"
	}
	`, resourceID, name, source)
}
