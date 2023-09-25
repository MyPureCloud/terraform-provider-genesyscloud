package architect_grammar

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v109/platformclientv2"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"testing"
)

func TestAccResourceArchitectGrammarBasic(t *testing.T) {
	var (
		resourceId   = "grammar" + uuid.NewString()
		name1        = "Test grammar " + uuid.NewString()
		description1 = "Test description"
		language1    = generateGrammarLanguageBlock(
			"test language name",
			generateFileVoiceFileMetadataBlock(
				"voiceFile1",
				"256",
				"2023-09-22T15:30:00.123Z",
				"Gram",
			),
			generateFileDtmfFileMetadataBlock(
				"dtmfFile1",
				"214",
				"2023-09-22T15:30:00.123Z",
				"Gram",
			),
		)
		name2        = "Test grammar " + uuid.NewString()
		description2 = "A new test description"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create Grammar
				Config: generateGrammarResource(
					resourceId,
					name1,
					description1,
					language1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "description", description1),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "languages.0.language", "test language name"),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "languages.0.voice_file_metadata.0.file_name", "voiceFile1"),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "languages.0.voice_file_metadata.0.file_size_bytes", "256"),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "languages.0.voice_file_metadata.0.date_uploaded", "2023-09-22T15:30:00.123Z"),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "languages.0.voice_file_metadata.0.file_type", "Gram"),
				),
			},
			{
				// Update Grammar
				Config: generateGrammarResource(
					resourceId,
					name2,
					description2,
					language1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "description", description2),
				),
			},
			{
				// Read
				ResourceName:      "genesyscloud_architect_grammar." + resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyGrammarDestroyed,
	})
}

func generateGrammarResource(
	resourceId string,
	name string,
	description string,
	nestedBlocks ...string,
) string {

	return fmt.Sprintf(`
		resource "genesyscloud_architect_grammar" "%s" {
			name = "%s"
			description = "%s"
			%s
		}
	`, resourceId, name, description, strings.Join(nestedBlocks, "\n"))
}

func generateGrammarLanguageBlock(
	language string,
	attrs ...string,
) string {
	return fmt.Sprintf(`
		languages {
			language = "%s"
			%s
		}
	`, language, strings.Join(attrs, "\n"))
}

func generateFileVoiceFileMetadataBlock(
	fileName string,
	fileSizeBytes string,
	dateUploaded string,
	fileType string,
) string {
	return fmt.Sprintf(`
		voice_file_metadata {
			file_name = "%s"
			file_size_bytes = %s
			date_uploaded = "%s"
			file_type = "%s"
		}
	`, fileName, fileSizeBytes, dateUploaded, fileType)
}

func generateFileDtmfFileMetadataBlock(
	fileName string,
	fileSizeBytes string,
	dateUploaded string,
	fileType string,
) string {
	return fmt.Sprintf(`
		dtmf_file_metadata {
			file_name = "%s"
			file_size_bytes = %s
			date_uploaded = "%s"
			file_type = "%s"
		}
	`, fileName, fileSizeBytes, dateUploaded, fileType)
}

func testVerifyGrammarDestroyed(state *terraform.State) error {
	architectAPI := platformclientv2.NewArchitectApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_architect_grammar" {
			continue
		}
		grammar, resp, err := architectAPI.GetArchitectGrammar(rs.Primary.ID, false)
		if grammar != nil {
			return fmt.Errorf("Grammar (%s) still exists", rs.Primary.ID)
		} else if gcloud.IsStatus404(resp) {
			// Grammar not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All grammars deleted
	return nil
}
