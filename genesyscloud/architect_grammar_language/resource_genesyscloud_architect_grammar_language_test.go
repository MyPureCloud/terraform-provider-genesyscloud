package architect_grammar_language

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	architectGrammar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_grammar"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/aws/localstack"
	localStackEnv "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/aws/localstack/environment"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v170/platformclientv2"
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
		resource "%s" "%s" {
			grammar_id = %s
			language = "%s"
			%s
		}
	`, ResourceType, resourceLabel, grammarId, language, strings.Join(attrs, "\n"))
}

func generateFileVoiceFileDataBlock(
	fileName string,
	fileType string,
) string {
	return fmt.Sprintf(`
		voice_file_data {
			file_name = "%s"
			file_type = "%s"
			file_content_hash = filesha256("%s")
		}
	`, fileName, fileType, fileName)
}

func generateFileDtmfFileDataBlock(
	fileName string,
	fileType string,
) string {
	return fmt.Sprintf(`
		dtmf_file_data {
			file_name = "%s"
			file_type = "%s"
			file_content_hash = filesha256("%s")
		}
	`, fileName, fileType, fileName)
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
		if rs.Type != ResourceType {
			continue
		}
		grammarId, languageCode := splitGrammarLanguageId(rs.Primary.ID)
		grammar, resp, err := architectAPI.GetArchitectGrammarLanguage(grammarId, languageCode)
		if grammar != nil {
			return fmt.Errorf("language (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Language not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All grammar languages deleted
	return nil
}

func generateFilePath(filename string) string {
	return testrunner.GetTestDataPath("resource", ResourceType, filename)
}

// TestAccResourceArchitectGrammarLanguage_S3 tests the architect grammar language resource using LocalStack for S3 operations.
// This test validates that the terraform-provider-genesyscloud can successfully deploy grammar language files from S3 buckets
// using LocalStack as a local AWS service emulator.
//
// Prerequisites:
//   - LocalStack must be running (either locally or in CI)
//   - Environment variables must be set:
//   - USE_LOCAL_STACK=true
//   - LOCAL_STACK_IMAGE_URI=<localstack-image-uri>
//
// Test Flow:
//
//  1. Creates temporary grammar language files (voice and DTMF)
//
//  2. Sets up LocalStack S3 bucket and uploads the files
//
//  3. Deploys the grammar language using terraform with S3 sources
//
//  4. Updates the file content and re-uploads to S3
//
//  5. Verifies the grammar language is updated correctly
//
//  6. Cleans up S3 bucket and temporary files
//
// This test is designed to run in CI environments where LocalStack is available
// and properly configured with the required environment variables.
func TestAccResourceArchitectGrammarLanguage_S3(t *testing.T) {
	/*
		// To run this test locally, uncomment this block and run `localstack start` from another terminal
		// See more about localstack cli here: https://docs.localstack.cloud/aws/getting-started/installation/
		os.Setenv(localStackEnv.UseLocalStackEnvVar, "true")
		os.Setenv(localStackEnv.LocalStackImageUriEnvVar, "localstack/localstack:latest")
	*/

	imageURI := os.Getenv(localStackEnv.LocalStackImageUriEnvVar)
	if imageURI == "" || !localStackEnv.LocalStackIsActive() {
		t.Skipf("Missing env variables (%s or %s), indicating that localstack is not running", localStackEnv.LocalStackImageUriEnvVar, localStackEnv.UseLocalStackEnvVar)
	}

	ctx := context.Background()
	localStackManager, err := localstack.NewLocalStackManager(ctx)
	if err != nil {
		t.Fatalf("Failed to initialise LocalStackManager: %s", err.Error())
	}

	var (
		grammarResourceLabel = "grammar" + uuid.NewString()
		grammarResource      = architectGrammar.GenerateGrammarResource(
			grammarResourceLabel,
			"Test grammar S3"+uuid.NewString(),
			"",
		)

		languageResourceLabel    = "language"
		languageResourceFullPath = ResourceType + "." + languageResourceLabel
		languageCode             = "en-us"
		bucketName               = "testbucket-" + strings.ToLower(strings.ReplaceAll(uuid.NewString(), "-", ""))
		voiceObjectKey           = "voice-grxml-01.grxml"
		dtmfObjectKey            = "dtmf-grxml-01.grxml"

		voiceS3Path = fmt.Sprintf("s3://%s/%s", bucketName, voiceObjectKey)
		dtmfS3Path  = fmt.Sprintf("s3://%s/%s", bucketName, dtmfObjectKey)
	)

	// Create test voice grammar content
	voiceContent := `<?xml version="1.0" encoding="UTF-8"?>
<grammar xmlns="http://www.w3.org/2001/06/grammar" xml:lang="en-US" root="root">
  <rule id="root">
    <one-of>
      <item>hello</item>
      <item>world</item>
    </one-of>
  </rule>
</grammar>`

	// Create test DTMF grammar content
	dtmfContent := `<?xml version="1.0" encoding="UTF-8"?>
<grammar xmlns="http://www.w3.org/2001/06/grammar" xml:lang="en-US" root="root">
  <rule id="root">
    <one-of>
      <item><one-of><item>1</item><item>2</item></one-of></item>
      <item><one-of><item>3</item><item>4</item></one-of></item>
    </one-of>
  </rule>
</grammar>`

	// Create temporary files
	voiceTempFile, err := os.CreateTemp("", "voice-*.grxml")
	if err != nil {
		t.Fatalf("Failed to create temp voice file: %v", err)
	}

	dtmfTempFile, err := os.CreateTemp("", "dtmf-*.grxml")
	if err != nil {
		t.Fatalf("Failed to create temp dtmf file: %v", err)
	}

	defer func() {
		if err := voiceTempFile.Close(); err != nil {
			t.Logf("[WARN] Failed to close temp voice file: %v", err)
		}
		if err := dtmfTempFile.Close(); err != nil {
			t.Logf("[WARN] Failed to close temp dtmf file: %v", err)
		}
		if err := os.Remove(voiceTempFile.Name()); err != nil {
			t.Logf("[WARN] Failed to remove temp voice file: %v", err)
		}
		if err := os.Remove(dtmfTempFile.Name()); err != nil {
			t.Logf("[WARN] Failed to remove temp dtmf file: %v", err)
		}
	}()

	// Write initial content to temp files
	_, err = voiceTempFile.WriteString(voiceContent)
	if err != nil {
		t.Fatalf("Failed to write voice content: %v", err)
	}
	_, err = dtmfTempFile.WriteString(dtmfContent)
	if err != nil {
		t.Fatalf("Failed to write dtmf content: %v", err)
	}

	// Setup S3 bucket and upload test files
	t.Log("Setting up S3 bucket and uploading test files...")
	err = localStackManager.SetupS3Bucket(bucketName, voiceTempFile.Name(), voiceObjectKey)
	if err != nil {
		t.Fatalf("Failed to setup S3 bucket for voice file: %v", err)
	}
	err = localStackManager.SetupS3Bucket(bucketName, dtmfTempFile.Name(), dtmfObjectKey)
	if err != nil {
		t.Fatalf("Failed to setup S3 bucket for dtmf file: %v", err)
	}

	// Cleanup S3 bucket after test
	defer func() {
		if err := localStackManager.CleanupS3Bucket(bucketName); err != nil {
			t.Logf("[WARN] Failed to cleanup S3 bucket: %v", err)
		}
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create Grammar language with S3 files
				Config: grammarResource + generateGrammarLanguageResource(
					languageResourceLabel,
					"genesyscloud_architect_grammar."+grammarResourceLabel+".id",
					languageCode,
					generateFileVoiceFileDataBlockS3(voiceS3Path, "Grxml"),
					generateFileDtmfFileDataBlockS3(dtmfS3Path, "Grxml"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(languageResourceFullPath, "language", languageCode),
					resource.TestCheckResourceAttr(languageResourceFullPath, "voice_file_data.0.file_name", voiceS3Path),
					resource.TestCheckResourceAttr(languageResourceFullPath, "voice_file_data.0.file_type", "Grxml"),
					resource.TestCheckResourceAttr(languageResourceFullPath, "dtmf_file_data.0.file_name", dtmfS3Path),
					resource.TestCheckResourceAttr(languageResourceFullPath, "dtmf_file_data.0.file_type", "Grxml"),
				),
			},
		},
		CheckDestroy: testVerifyGrammarLanguageDestroyed,
	})
}

// generateFileVoiceFileDataBlockS3 generates a voice file data block for S3 paths (no file_content_hash needed)
func generateFileVoiceFileDataBlockS3(
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

// generateFileDtmfFileDataBlockS3 generates a dtmf file data block for S3 paths (no file_content_hash needed)
func generateFileDtmfFileDataBlockS3(
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
