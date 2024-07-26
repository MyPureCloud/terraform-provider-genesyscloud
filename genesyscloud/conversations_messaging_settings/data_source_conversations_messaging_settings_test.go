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
		id     = "conversations_messaging_settings"
		dataId = "conversations_messaging_settings_data"
		name   = "Messaging Settings " + uuid.NewString()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateConversationsMessagingSettingsResource(
					id,
					name,
					generateContentStoryBlock(
						generateMentionInboundOnlySetting("Enabled"),
						generateReplyInboundOnlySetting("Enabled"),
					),
					generateTypingOnSetting(
						"Enabled",
						"Disabled",
					),
				) + generateConversationsMessagingSettingsDataSource(
					dataId,
					name,
					"genesyscloud_conversations_messaging_settings."+id,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_conversations_messaging_settings."+dataId, "id",
						"genesyscloud_conversations_messaging_settings."+id, "id"),
				),
			},
		},
	})
}

func generateConversationsMessagingSettingsDataSource(id, name, dependsOn string) string {
	return fmt.Sprintf(`data "genesyscloud_conversations_messaging_settings" "%s" {
		name = "%s"
		depends_on = [%s]
	}
`, id, name, dependsOn)
}
