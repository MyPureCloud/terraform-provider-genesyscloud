package quality_miner

import (
	"fmt"
	"testing"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

func TestAccResourceQualityMinerTopic(t *testing.T) {
	var (
		minerResourceLabel = "test-miner-topic"
		minerName          = "tf-miner-" + uuid.NewString()[:8]
		minerNameUpdated   = "tf-miner-" + uuid.NewString()[:8]
		resourcePath       = "genesyscloud_quality_miner." + minerResourceLabel
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateQualityMinerResource(
					minerResourceLabel,
					minerName,
					"Topic",
					`language = "en-US"`,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", minerName),
					resource.TestCheckResourceAttr(resourcePath, "miner_type", "Topic"),
					resource.TestCheckResourceAttr(resourcePath, "language", "en-us"),
					resource.TestCheckResourceAttrSet(resourcePath, "id"),
				),
			},
			{
				Config: GenerateQualityMinerResource(
					minerResourceLabel,
					minerNameUpdated,
					"Topic",
					`language = "en-US"`,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", minerNameUpdated),
				),
			},
			{
				ResourceName:      resourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: validateMinersDestroyed,
	})
}

func TestAccResourceQualityMinerIntentNoLanguage(t *testing.T) {
	var (
		minerResourceLabel = "test-miner-intent"
		minerName          = "tf-miner-" + uuid.NewString()[:8]
		resourcePath       = "genesyscloud_quality_miner." + minerResourceLabel
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateQualityMinerResource(
					minerResourceLabel,
					minerName,
					"Intent",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", minerName),
					resource.TestCheckResourceAttr(resourcePath, "miner_type", "Intent"),
					resource.TestCheckResourceAttrSet(resourcePath, "language"),
				),
			},
			{
				ResourceName:      resourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: validateMinersDestroyed,
	})
}

func validateMinersDestroyed(state *terraform.State) error {
	languageUnderstandingApi := platformclientv2.NewLanguageUnderstandingApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_quality_miner" {
			continue
		}

		deadline := time.Now().Add(90 * time.Second)
		for {
			miner, resp, err := languageUnderstandingApi.GetLanguageunderstandingMiner(rs.Primary.ID)
			if util.IsStatus404(resp) {
				break
			}
			if miner != nil && miner.Status != nil && *miner.Status == "Deleted" {
				break
			}
			if err != nil && miner == nil {
				return fmt.Errorf("unexpected error verifying miner %s destruction: %s", rs.Primary.ID, err)
			}
			if time.Now().After(deadline) {
				return fmt.Errorf("miner (%s) still exists after destroy", rs.Primary.ID)
			}
			time.Sleep(5 * time.Second)
		}
	}
	return nil
}
