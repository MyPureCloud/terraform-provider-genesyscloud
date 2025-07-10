package guide

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccResourceGuideManual(t *testing.T) {
	if v := os.Getenv("GENESYSCLOUD_REGION"); v != "tca" {
		t.Skipf("Skipping test for region %s. genesyscloud_guide is currently only supported in tca", v)
		return
	}

	if !GuideFtIsEnabled() {
		t.Skip("Skipping test as guide feature toggle is not enabled")
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
					"",
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

func TestAccResourceGuidePrompt(t *testing.T) {
	if v := os.Getenv("GENESYSCLOUD_REGION"); v != "tca" {
		t.Skipf("Skipping test for region %s. genesyscloud_guide is currently only supported in tca", v)
		return
	}
	var (
		resourceLabel = "guide"

		name   = "Test Guide Manual" + uuid.NewString()
		source = "Prompt"
		prompt = "Create a guide that handles customer service interactions"
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
					prompt,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "source", source),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "latest_saved_version", "1.0"),
				),
			},
			{
				// Import/Read
				ResourceName:            ResourceType + "." + resourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"prompt"},
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
