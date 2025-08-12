package conversations_messaging_settings

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceConversationsMessagingSettings(t *testing.T) {
	var (
		resourceLabel   = "conversations_messaging_settings"
		dataSourceLabel = "conversations_messaging_settings_data"
		name            = "TestTerraformMessagingSetting-" + uuid.NewString()
	)

	if cleanupErr := CleanupMessagingSettings("TestTerraformMessagingSetting"); cleanupErr != nil {
		t.Logf("Failed to clean up messaging settings with name '%s': %s", name, cleanupErr.Error())
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateConversationsMessagingSettingsResource(
					resourceLabel,
					name,
					GenerateContentStoryBlock(
						GenerateMentionInboundOnlySetting("Enabled"),
						GenerateReplyInboundOnlySetting("Enabled"),
					),
					GenerateTypingOnSetting(
						"Enabled",
						"Disabled",
					),
				) + generateConversationsMessagingSettingsDataSource(
					dataSourceLabel,
					name,
					"genesyscloud_conversations_messaging_settings."+resourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_conversations_messaging_settings."+dataSourceLabel, "id",
						"genesyscloud_conversations_messaging_settings."+resourceLabel, "id"),
				),
			},
		},
	})
}

func generateConversationsMessagingSettingsDataSource(dataSourceLabel, name, dependsOn string) string {
	return fmt.Sprintf(`data "genesyscloud_conversations_messaging_settings" "%s" {
		name = "%s"
		depends_on = [%s]
	}
`, dataSourceLabel, name, dependsOn)
}

func CleanupMessagingSettings(name string) error {
	cmMessagingSettingApi := platformclientv2.NewConversationsApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		cmMessagingSetting, _, getErr := cmMessagingSettingApi.GetConversationsMessagingSettings(pageSize, pageNum)
		if getErr != nil {
			return fmt.Errorf("failed to get page %v of messaging settings: %v", pageNum, getErr)
		}

		if cmMessagingSetting.Entities == nil || len(*cmMessagingSetting.Entities) == 0 {
			break
		}

		for _, setting := range *cmMessagingSetting.Entities {
			if setting.Name != nil && strings.HasPrefix(*setting.Name, name) {
				_, err := cmMessagingSettingApi.DeleteConversationsMessagingSetting(*setting.Id)
				if err != nil {
					return fmt.Errorf("failed to delete messaging settings: %v", err)
				}
				time.Sleep(5 * time.Second)
			}
		}
	}
	return nil
}
