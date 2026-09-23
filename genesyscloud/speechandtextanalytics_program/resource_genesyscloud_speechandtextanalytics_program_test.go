package speechandtextanalytics_program

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccResourceSpeechAndTextAnalyticsProgram(t *testing.T) {
	t.Parallel()

	var (
		resourceLabel = "program_" + uuid.NewString()

		name1        = "tfacc-program-" + uuid.NewString()
		description1 = "Terraform acceptance test program"

		name2        = "tfacc-program-" + uuid.NewString()
		description2 = "Terraform acceptance test program updated"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with topic_ids reference
				Config: generateSpeechAndTextAnalyticsProgramResourceWithTopics(
					resourceLabel,
					name1,
					description1,
					[]string{"tfacc", "stt-program"},
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "description", description1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "tags.#", "2"),
					resource.TestCheckResourceAttrSet(ResourceType+"."+resourceLabel, "topic_ids.0"),
				),
			},
			{
				// Update
				Config: generateSpeechAndTextAnalyticsProgramResource(
					resourceLabel,
					name2,
					description2,
					[]string{"tfacc", "stt-program", "updated"},
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "description", description2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "tags.#", "3"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "topic_ids.#", "0"),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySpeechAndTextAnalyticsProgramDestroyed,
	})
}

func testVerifySpeechAndTextAnalyticsProgramDestroyed(state *terraform.State) error {
	sttAPI := platformclientv2.NewSpeechTextAnalyticsApi()

	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		program, resp, err := sttAPI.GetSpeechandtextanalyticsProgram(rs.Primary.ID)
		if program != nil {
			return fmt.Errorf("speech and text analytics program (%s) still exists", rs.Primary.ID)
		}
		if util.IsStatus404(resp) {
			// program not found as expected
			continue
		}

		return fmt.Errorf("unexpected error checking program destruction: %v", err)
	}

	return nil
}

func generateSpeechAndTextAnalyticsProgramResource(
	resourceLabel, name, description string,
	tags []string,
) string {
	quotedTags := make([]string, 0, len(tags))
	for _, t := range tags {
		quotedTags = append(quotedTags, fmt.Sprintf("%q", t))
	}

	return fmt.Sprintf(`
resource "%s" "%s" {
  name        = %q
  description = %q
  tags        = [%s]
}
`, ResourceType, resourceLabel, name, description, strings.Join(quotedTags, ", "))
}

// generateSpeechAndTextAnalyticsProgramResourceWithTopics generates a program resource with topic dependency.
func generateSpeechAndTextAnalyticsProgramResourceWithTopics(
	resourceLabel, name, description string,
	tags []string,
) string {
	quotedTags := make([]string, 0, len(tags))
	for _, t := range tags {
		quotedTags = append(quotedTags, fmt.Sprintf("%q", t))
	}

	return fmt.Sprintf(`
resource "genesyscloud_speechandtextanalytics_topic" "example_topic" {
  name        = "tfacc-topic-for-program"
  dialect     = "en-US"
  description = "Topic for program acceptance test"
}

resource "%s" "%s" {
  name        = %q
  description = %q
  topic_ids   = [genesyscloud_speechandtextanalytics_topic.example_topic.id]
  tags        = [%s]
}
`, ResourceType, resourceLabel, name, description, strings.Join(quotedTags, ", "))
}
