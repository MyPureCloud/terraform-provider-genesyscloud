package conversations_messaging_settings

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceConversationsMessagingSettings(t *testing.T) {
	var (
		resourceLabel   = "conversations_messaging_settings"
		dataSourceLabel = "conversations_messaging_settings_data"
		name            = "Messaging Settings " + uuid.NewString()
	)
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
