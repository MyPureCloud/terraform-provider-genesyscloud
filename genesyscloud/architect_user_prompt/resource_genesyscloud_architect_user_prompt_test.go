package architect_user_prompt

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/fileserver"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func TestAccResourceUserPromptBasic(t *testing.T) {
	userPromptResourceLabel1 := "test-user_prompt_1"
	userPromptName1 := "TestUserPrompt_1" + strings.Replace(uuid.NewString(), "-", "", -1)
	userPromptName2 := "TestUserPrompt_2" + strings.Replace(uuid.NewString(), "-", "", -1)
	userPromptDescription1 := "Test description"
	userPromptResourceLang1 := "en-us"
	userPromptResourceLang2 := "ja-jp"
	userPromptResourceTTS1 := "This is a test greeting!"
	userPromptResourceTTS2 := "This is a test greeting too!"
	userPromptResourceTTS3 := "こんにちは!"

	userPromptAsset1 := UserPromptResourceStruct{
		userPromptResourceLang1,
		strconv.Quote(userPromptResourceTTS1),
		util.NullValue,
		util.NullValue,
		util.NullValue,
	}

	userPromptAsset2 := UserPromptResourceStruct{
		userPromptResourceLang1,
		strconv.Quote(userPromptResourceTTS2),
		util.NullValue,
		util.NullValue,
		util.NullValue,
	}

	userPromptAsset3 := UserPromptResourceStruct{
		userPromptResourceLang2,
		strconv.Quote(userPromptResourceTTS3),
		util.NullValue,
		util.NullValue,
		util.NullValue,
	}

	userPromptResources1 := []*UserPromptResourceStruct{&userPromptAsset1}
	userPromptResources2 := []*UserPromptResourceStruct{&userPromptAsset2, &userPromptAsset3}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create Empty user prompt
				Config: GenerateUserPromptResource(&UserPromptStruct{
					userPromptResourceLabel1,
					userPromptName1,
					strconv.Quote(userPromptDescription1),
					nil,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "name", userPromptName1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "description", userPromptDescription1),
				),
			},
			{
				// Update to include TTS message prompt resource
				Config: GenerateUserPromptResource(&UserPromptStruct{
					userPromptResourceLabel1,
					userPromptName2,
					strconv.Quote(userPromptDescription1),
					userPromptResources1,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "name", userPromptName2),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "description", userPromptDescription1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "resources.0.language", userPromptResourceLang1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "resources.0.tts_string", userPromptResourceTTS1),
				),
			},
			{
				// Update existing language TTS
				Config: GenerateUserPromptResource(&UserPromptStruct{
					userPromptResourceLabel1,
					userPromptName2,
					strconv.Quote(userPromptDescription1),
					userPromptResources2,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "name", userPromptName2),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "description", userPromptDescription1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "resources.0.language", userPromptResourceLang1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "resources.0.tts_string", userPromptResourceTTS2),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "resources.1.language", userPromptResourceLang2),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "resources.1.tts_string", userPromptResourceTTS3),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_architect_user_prompt." + userPromptResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyUserPromptsDestroyed,
	})
}

func TestAccResourceUserPromptWavFile(t *testing.T) {
	userPromptResourceLabel1 := "test-user_prompt_wav_file"
	userPromptName1 := "TestUserPromptWav_1" + strings.Replace(uuid.NewString(), "-", "", -1)
	userPromptDescription1 := "Test prompt with wav audio file"
	userPromptResourceLang1 := "en-us"
	userPromptResourceText1 := "This is a test greeting!"
	userPromptResourceFileName1 := testrunner.GetTestDataPath("resource", ResourceType, "test-prompt-01.wav")
	userPromptResourceFileName2 := testrunner.GetTestDataPath("resource", ResourceType, "test-prompt-02.wav")

	userPromptAsset1 := UserPromptResourceStruct{
		Language:        userPromptResourceLang1,
		Tts_string:      util.NullValue,
		Text:            strconv.Quote(userPromptResourceText1),
		Filename:        strconv.Quote(userPromptResourceFileName1),
		FileContentHash: userPromptResourceFileName1,
	}

	userPromptAsset2 := UserPromptResourceStruct{
		Language:        userPromptResourceLang1,
		Tts_string:      util.NullValue,
		Text:            strconv.Quote(userPromptResourceText1),
		Filename:        strconv.Quote(userPromptResourceFileName2),
		FileContentHash: userPromptResourceFileName2,
	}

	userPromptResources1 := []*UserPromptResourceStruct{&userPromptAsset1}
	userPromptResources2 := []*UserPromptResourceStruct{&userPromptAsset2}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create user prompt with an audio file
				Config: GenerateUserPromptResource(&UserPromptStruct{
					userPromptResourceLabel1,
					userPromptName1,
					strconv.Quote(userPromptDescription1),
					userPromptResources1,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "name", userPromptName1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "description", userPromptDescription1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "resources.0.filename", userPromptResourceFileName1),
				),
			},
			{
				// Replace audio file for the prompt
				Config: GenerateUserPromptResource(&UserPromptStruct{
					userPromptResourceLabel1,
					userPromptName1,
					strconv.Quote(userPromptDescription1),
					userPromptResources2,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "name", userPromptName1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "description", userPromptDescription1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "resources.0.filename", userPromptResourceFileName2),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_architect_user_prompt." + userPromptResourceLabel1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"resources"},
			},
		},
		CheckDestroy: testVerifyUserPromptsDestroyed,
	})
}

func TestAccResourceUserPromptWavFileURL(t *testing.T) {
	userPromptResourceLabel1 := "test-user_prompt_wav_file"
	userPromptName1 := "TestUserPromptWav_1" + strings.Replace(uuid.NewString(), "-", "", -1)
	userPromptDescription1 := "Test prompt with wav audio file"
	userPromptResourceLang1 := "en-us"
	userPromptResourceText1 := "This is a test greeting!"
	userPromptResourceFileName1 := "http://localhost:8100/test-prompt-01.wav"
	userPromptResourceFileName2 := "http://localhost:8100/test-prompt-02.wav"

	userPromptAsset1 := UserPromptResourceStruct{
		Language:        userPromptResourceLang1,
		Tts_string:      util.NullValue,
		Text:            strconv.Quote(userPromptResourceText1),
		Filename:        strconv.Quote(userPromptResourceFileName1),
		FileContentHash: util.NullValue,
	}

	userPromptAsset2 := UserPromptResourceStruct{
		Language:        userPromptResourceLang1,
		Tts_string:      util.NullValue,
		Text:            strconv.Quote(userPromptResourceText1),
		Filename:        strconv.Quote(userPromptResourceFileName2),
		FileContentHash: util.NullValue,
	}

	userPromptResources1 := []*UserPromptResourceStruct{&userPromptAsset1}
	userPromptResources2 := []*UserPromptResourceStruct{&userPromptAsset2}

	httpServerExitDone := &sync.WaitGroup{}
	httpServerExitDone.Add(1)
	srv := fileserver.Start(httpServerExitDone, testrunner.GetTestDataPath("resource", ResourceType), 8100)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create user prompt with an audio file
				Config: GenerateUserPromptResource(&UserPromptStruct{
					userPromptResourceLabel1,
					userPromptName1,
					strconv.Quote(userPromptDescription1),
					userPromptResources1,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "name", userPromptName1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "description", userPromptDescription1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "resources.0.filename", userPromptResourceFileName1),
				),
			},
			{
				// Replace audio file for the prompt
				Config: GenerateUserPromptResource(&UserPromptStruct{
					userPromptResourceLabel1,
					userPromptName1,
					strconv.Quote(userPromptDescription1),
					userPromptResources2,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "name", userPromptName1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "description", userPromptDescription1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel1, "resources.0.filename", userPromptResourceFileName2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_architect_user_prompt." + userPromptResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyUserPromptsDestroyed,
	})

	fileserver.ShutDown(srv, httpServerExitDone)
}

func testVerifyUserPromptsDestroyed(state *terraform.State) error {
	architectAPI := platformclientv2.NewArchitectApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_architect_user_prompt" {
			continue
		}

		userPrompt, resp, err := architectAPI.GetArchitectPrompt(rs.Primary.ID, false, false, nil)

		if userPrompt != nil {
			return fmt.Errorf("User Prompt (%s) still exists", rs.Primary.ID)
		}

		if resp != nil && resp.StatusCode == 404 {
			// User prompt not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("Unexpected error: %s", err)
	}
	// Success. All User Prompts destroyed
	return nil
}
