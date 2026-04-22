package speechandtextanalytics_topic

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccResourceSpeechAndTextAnalyticsTopic(t *testing.T) {
	t.Parallel()

	var (
		resourceLabel = "topic_" + uuid.NewString()

		name1        = "tfacc-topic-" + uuid.NewString()
		description1 = "Terraform acceptance test topic"

		name2        = "tfacc-topic-" + uuid.NewString()
		description2 = "Terraform acceptance test topic updated"

		dialect      = "en-US"
		strictness1  = "72"
		strictness2  = "85"
		participants = "All"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateSpeechAndTextAnalyticsTopicResource(
					resourceLabel,
					name1,
					dialect,
					description1,
					strictness1,
					participants,
					[]string{"tfacc", "stt-topic"},
					[]string{
						generateSpeechAndTextAnalyticsTopicPhrase("hello world", "Neutral"),
					},
					false,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "dialect", dialect),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "description", description1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "strictness", strictness1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "participants", participants),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "tags.#", "2"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phrases.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phrases.0.text", "hello world"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phrases.0.sentiment", "Neutral"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "published", "false"),
				),
			},
			{
				// Update
				Config: generateSpeechAndTextAnalyticsTopicResource(
					resourceLabel,
					name2,
					dialect,
					description2,
					strictness2,
					participants,
					[]string{"tfacc", "stt-topic", "updated"},
					[]string{
						generateSpeechAndTextAnalyticsTopicPhrase("updated phrase", "Positive"),
					},
					false,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "dialect", dialect),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "description", description2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "strictness", strictness2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "participants", participants),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "tags.#", "3"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phrases.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phrases.0.text", "updated phrase"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phrases.0.sentiment", "Positive"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "published", "false"),
				),
			},
			{
				// Import/Read
				ResourceName:            ResourceType + "." + resourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"strictness"},
			},
		},
		CheckDestroy: testVerifySpeechAndTextAnalyticsTopicDestroyed,
	})
}

func testVerifySpeechAndTextAnalyticsTopicDestroyed(state *terraform.State) error {
	sttAPI := platformclientv2.NewSpeechTextAnalyticsApi()

	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		topic, resp, err := sttAPI.GetSpeechandtextanalyticsTopic(rs.Primary.ID)
		if topic != nil {
			return fmt.Errorf("speech and text analytics topic (%s) still exists", rs.Primary.ID)
		}
		if util.IsStatus404(resp) {
			// topic not found as expected
			continue
		}

		return fmt.Errorf("unexpected error checking topic destruction: %v", err)
	}

	return nil
}

func generateSpeechAndTextAnalyticsTopicResource(
	resourceLabel, name, dialect, description, strictness, participants string,
	tags []string,
	phraseBlocks []string,
	published bool,
) string {
	quotedTags := make([]string, 0, len(tags))
	for _, t := range tags {
		quotedTags = append(quotedTags, fmt.Sprintf("%q", t))
	}

	return fmt.Sprintf(`
resource "%s" "%s" {
  name         = %q
  dialect      = %q
  description  = %q
  strictness   = %q
  participants = %q
  tags         = [%s]
%s
  published    = %t
}
`, ResourceType, resourceLabel, name, dialect, description, strictness, participants, strings.Join(quotedTags, ", "), strings.Join(phraseBlocks, "\n"), published)
}

func generateSpeechAndTextAnalyticsTopicPhrase(text, sentiment string) string {
	return fmt.Sprintf(`
  phrases {
    text       = %q
    sentiment  = %q
  }
`, text, sentiment)
}
