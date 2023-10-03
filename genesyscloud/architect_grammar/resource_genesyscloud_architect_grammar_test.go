package architect_grammar

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v112/platformclientv2"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"testing"
)

type FileType int

const (
	Dtmf FileType = iota
	Voice
)

func TestAccResourceArchitectGrammarGrxml(t *testing.T) {
	var (
		languageCode1 = "en-us"
		voiceGrxml1   = generateFilePath("voice-grxml-01.grxml")
		dtmfGrxml1    = generateFilePath("dtmf-grxml-01.grxml")
		language1     = generateGrammarLanguageBlock(
			languageCode1,
			generateFileVoiceFileDataBlock(
				voiceGrxml1,
				"Grxml",
			),
			generateFileDtmfFileDataBlock(
				dtmfGrxml1,
				"Grxml",
			),
		)
		languageCode2 = "es-ar"
		voiceGrxml2   = generateFilePath("voice-grxml-02.grxml")
		dtmfGrxml2    = generateFilePath("dtmf-grxml-02.grxml")
		language2     = generateGrammarLanguageBlock(
			languageCode2,
			generateFileVoiceFileDataBlock(
				voiceGrxml2,
				"Grxml",
			),
			generateFileDtmfFileDataBlock(
				dtmfGrxml2,
				"Grxml",
			),
		)
	)

	var (
		resourceId   = "grammar" + uuid.NewString()
		name1        = "Test grammar " + uuid.NewString()
		description1 = "Test description"
		name2        = "Test grammar " + uuid.NewString()
		description2 = "A new description"
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
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "languages.0.language", languageCode1),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "languages.0.voice_file_data.0.file_name", voiceGrxml1),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "languages.0.voice_file_data.0.file_type", "Grxml"),
					verifyFileUpload("genesyscloud_architect_grammar."+resourceId, "en-us", Voice, voiceGrxml1),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "languages.0.dtmf_file_data.0.file_name", dtmfGrxml1),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "languages.0.dtmf_file_data.0.file_type", "Grxml"),
					verifyFileUpload("genesyscloud_architect_grammar."+resourceId, "en-us", Dtmf, dtmfGrxml1),
				),
			},
			{
				// Update Grammar
				Config: generateGrammarResource(
					resourceId,
					name2,
					description2,
					language2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "description", description2),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "languages.0.language", languageCode2),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "languages.0.voice_file_data.0.file_name", voiceGrxml2),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "languages.0.voice_file_data.0.file_type", "Grxml"),
					verifyFileUpload("genesyscloud_architect_grammar."+resourceId, "en-us", Voice, voiceGrxml2),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "languages.0.dtmf_file_data.0.file_name", dtmfGrxml2),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceId, "languages.0.dtmf_file_data.0.file_type", "Grxml"),
					verifyFileUpload("genesyscloud_architect_grammar."+resourceId, "en-us", Dtmf, dtmfGrxml2),
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

func generateFileVoiceFileDataBlock(
	fileName string,
	fileType string,
) string {
	return fmt.Sprintf(`
		voice_file_data {
			file_name = "%s"
			file_type = "%s"
		}
	`, fileName, fileType)
}

func generateFileDtmfFileDataBlock(
	fileName string,
	fileType string,
) string {
	return fmt.Sprintf(`
		dtmf_file_data {
			file_name = "%s"
			file_type = "%s"
		}
	`, fileName, fileType)
}

func verifyFileUpload(grammarResourceName string, language string, fileType FileType, filename string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		grammarResource, ok := state.RootModule().Resources[grammarResourceName]
		if !ok {
			return fmt.Errorf("Failed to find grammar %s in state", grammarResourceName)
		}
		grammarId := grammarResource.Primary.ID
		architectAPI := platformclientv2.NewArchitectApi()

		grammarLanguage, _, err := architectAPI.GetArchitectGrammarLanguage(grammarId, language)
		if err != nil {
			return fmt.Errorf("Failed to find langauge %s for resource %s", language, grammarResourceName)
		}

		if fileType == Dtmf {
			if grammarLanguage.DtmfFileUrl == nil {
				return fmt.Errorf("Dtmf file url not found for file %s", filename)
			}
		} else if fileType == Voice {
			if grammarLanguage.VoiceFileUrl == nil {
				return fmt.Errorf("Voice file url not found for file %s", filename)
			}
		} else {
			return fmt.Errorf("Unknown language file type")
		}

		return nil
	}
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

func generateFilePath(filename string) string {
	testFolder := "../../test/data/resource/architect_grammar/"

	return testFolder + filename
}
