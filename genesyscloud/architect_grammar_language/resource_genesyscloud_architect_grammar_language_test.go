package architect_grammar_language

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	architectGrammar "terraform-provider-genesyscloud/genesyscloud/architect_grammar"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

func TestAccResourceArchitectGrammarLanguage(t *testing.T) {
	var (
		grammarResourceLabel = "grammar" + uuid.NewString()
		grammarResource      = architectGrammar.GenerateGrammarResource(
			grammarResourceLabel,
			"Test grammar"+uuid.NewString(),
			"",
		)
	)

	var (
		languageResourceLabel = "language" + uuid.NewString()

		languageCode = "en-us"
		voiceGrxml1  = generateFilePath("voice-grxml-01.grxml")
		dtmfGrxml1   = generateFilePath("dtmf-grxml-01.grxml")
		voiceGrxml2  = generateFilePath("voice-grxml-02.grxml")
		dtmfGrxml2   = generateFilePath("dtmf-grxml-02.grxml")

		voiceGram1 = generateFilePath("voice-gram-01.gram")
		dtmfGram1  = generateFilePath("dtmf-gram-01.gram")
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create Grammar language
				Config: grammarResource + generateGrammarLanguageResource(
					languageResourceLabel,
					"genesyscloud_architect_grammar."+grammarResourceLabel+".id",
					languageCode,
					generateFileVoiceFileDataBlock(
						voiceGrxml1,
						"Grxml",
					),
					generateFileDtmfFileDataBlock(
						dtmfGrxml1,
						"Grxml",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar_language."+languageResourceLabel, "language", languageCode),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar_language."+languageResourceLabel, "voice_file_data.0.file_name", voiceGrxml1),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar_language."+languageResourceLabel, "voice_file_data.0.file_type", "Grxml"),
					verifyFileUpload("genesyscloud_architect_grammar."+grammarResourceLabel, "en-us", Voice, voiceGrxml1),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar_language."+languageResourceLabel, "dtmf_file_data.0.file_name", dtmfGrxml1),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar_language."+languageResourceLabel, "dtmf_file_data.0.file_type", "Grxml"),
					verifyFileUpload("genesyscloud_architect_grammar."+grammarResourceLabel, "en-us", Dtmf, dtmfGrxml1),
				),
			},
			{
				// Update Grammar language
				Config: grammarResource + generateGrammarLanguageResource(
					languageResourceLabel,
					"genesyscloud_architect_grammar."+grammarResourceLabel+".id",
					languageCode,
					generateFileVoiceFileDataBlock(
						voiceGrxml2,
						"Grxml",
					),
					generateFileDtmfFileDataBlock(
						dtmfGrxml2,
						"Grxml",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar_language."+languageResourceLabel, "language", languageCode),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar_language."+languageResourceLabel, "voice_file_data.0.file_name", voiceGrxml2),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar_language."+languageResourceLabel, "voice_file_data.0.file_type", "Grxml"),
					verifyFileUpload("genesyscloud_architect_grammar."+grammarResourceLabel, "en-us", Voice, voiceGrxml2),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar_language."+languageResourceLabel, "dtmf_file_data.0.file_name", dtmfGrxml2),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar_language."+languageResourceLabel, "dtmf_file_data.0.file_type", "Grxml"),
					verifyFileUpload("genesyscloud_architect_grammar."+grammarResourceLabel, "en-us", Dtmf, dtmfGrxml2),
				),
			},
			{
				// Update Grammar language files to gram files
				Config: grammarResource + generateGrammarLanguageResource(
					languageResourceLabel,
					"genesyscloud_architect_grammar."+grammarResourceLabel+".id",
					languageCode,
					generateFileVoiceFileDataBlock(
						voiceGram1,
						"Gram",
					),
					generateFileDtmfFileDataBlock(
						dtmfGram1,
						"Gram",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar_language."+languageResourceLabel, "language", languageCode),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar_language."+languageResourceLabel, "voice_file_data.0.file_name", voiceGram1),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar_language."+languageResourceLabel, "voice_file_data.0.file_type", "Gram"),
					verifyFileUpload("genesyscloud_architect_grammar."+grammarResourceLabel, "en-us", Voice, voiceGram1),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar_language."+languageResourceLabel, "dtmf_file_data.0.file_name", dtmfGram1),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar_language."+languageResourceLabel, "dtmf_file_data.0.file_type", "Gram"),
					verifyFileUpload("genesyscloud_architect_grammar."+grammarResourceLabel, "en-us", Dtmf, dtmfGram1),
				),
			},
			{
				// Read
				ResourceName:      "genesyscloud_architect_grammar_language." + languageResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"dtmf_file_data",
					"voice_file_data",
				},
			},
		},
		CheckDestroy: testVerifyGrammarLanguageDestroyed,
	})
}

func generateGrammarLanguageResource(
	resourceLabel string,
	grammarId string,
	language string,
	attrs ...string,
) string {
	return fmt.Sprintf(`
		resource "genesyscloud_architect_grammar_language" "%s" {
			grammar_id = %s
			language = "%s"
			%s
		}
	`, resourceLabel, grammarId, language, strings.Join(attrs, "\n"))
}

func generateFileVoiceFileDataBlock(
	fileName string,
	fileType string,
) string {
	fullyQualifiedPath, _ := testrunner.NormalizePath(fileName)
	return fmt.Sprintf(`
		voice_file_data {
			file_name = "%s"
			file_type = "%s"
			file_content_hash = filesha256("%s")
		}
	`, fileName, fileType, fullyQualifiedPath)
}

func generateFileDtmfFileDataBlock(
	fileName string,
	fileType string,
) string {
	fullyQualifiedPath, _ := testrunner.NormalizePath(fileName)
	return fmt.Sprintf(`
		dtmf_file_data {
			file_name = "%s"
			file_type = "%s"
			file_content_hash = filesha256("%s")
		}
	`, fileName, fileType, fullyQualifiedPath)
}

func verifyFileUpload(grammarResourcePath string, language string, fileType FileType, filename string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		grammarResource, ok := state.RootModule().Resources[grammarResourcePath]
		if !ok {
			return fmt.Errorf("failed to find grammar %s in state", grammarResourcePath)
		}
		grammarId := grammarResource.Primary.ID
		architectAPI := platformclientv2.NewArchitectApi()

		grammarLanguage, _, err := architectAPI.GetArchitectGrammarLanguage(grammarId, language)
		if err != nil {
			return fmt.Errorf("failed to find language %s for resource %s", language, grammarResourcePath)
		}

		if fileType == Dtmf {
			if grammarLanguage.DtmfFileUrl == nil {
				return fmt.Errorf("dtmf file url not found for file %s", filename)
			}
			err := validateFileContent(*grammarLanguage.DtmfFileMetadata, *grammarLanguage.DtmfFileUrl)
			if err != nil {
				return err
			}
		}
		if fileType == Voice {
			if grammarLanguage.VoiceFileUrl == nil {
				return fmt.Errorf("voice file url not found for file %s", filename)
			}
			err := validateFileContent(*grammarLanguage.DtmfFileMetadata, *grammarLanguage.DtmfFileUrl)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func validateFileContent(fileData platformclientv2.Grammarlanguagefilemetadata, fileUrl string) error {
	// Download the language file
	downloadedFileContent, err := downloadFile(fileUrl)
	if err != nil {
		return fmt.Errorf("Error downloading %s: %v\n", fileUrl, err)
	}

	// Read the content of the local file
	localFile := *fileData.FileName
	localFileContent, err := os.ReadFile(localFile)
	if err != nil {
		return fmt.Errorf("Error reading %s: %v\n", localFile, err)
	}

	if string(localFileContent) != downloadedFileContent {
		return fmt.Errorf("downloaded file does not match local file")
	}
	return nil
}

func downloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func testVerifyGrammarLanguageDestroyed(state *terraform.State) error {
	architectAPI := platformclientv2.NewArchitectApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_architect_grammar_language" {
			continue
		}
		grammarId, languageCode := splitLanguageId(rs.Primary.ID)
		grammar, resp, err := architectAPI.GetArchitectGrammarLanguage(grammarId, languageCode)
		if grammar != nil {
			return fmt.Errorf("Language (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Language not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All grammar languages deleted
	return nil
}

func generateFilePath(filename string) string {
	testFolder := "../../test/data/resource/architect_grammar_language/"

	return testFolder + filename
}
