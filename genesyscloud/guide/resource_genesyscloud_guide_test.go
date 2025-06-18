package guide

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"os"
	"testing"
)

func TestAccResourceGuide(t *testing.T) {
	if v := os.Getenv("GENESYSCLOUD_REGION"); v != "tca" {
		t.Skipf("Skipping test for region %s. genesyscloud_guide is currently only supported in tca", v)
		return
	}
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
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "source", source),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + resourceLabel,
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
		if rs.Type != ResourceType {
			continue
		}
		guide, resp, err := proxy.getGuideById(context.Background(), rs.Primary.ID)
		if guide != nil {
			return fmt.Errorf("%s (%s) still exists", ResourceType, rs.Primary.ID)
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
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		source = "%s"
	}
	`, ResourceType, resourceID, name, source)
}
