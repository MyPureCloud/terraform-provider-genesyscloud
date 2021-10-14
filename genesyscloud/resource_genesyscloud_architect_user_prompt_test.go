package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
)

type userPromptStruct struct {
	resourceID  string
	name        string
	description string
	resources   []*userPromptResourceStruct
}

type userPromptResourceStruct struct {
	language   string
	tts_string string
	text       string
	filename   string
}

func TestAccResourceUserPromptBasic(t *testing.T) {
	userPromptResource1 := "test-user_prompt_1"
	userPromptName1 := "TestUserPrompt_1"
	userPromptDescription1 := "Test description"
	userPromptResourceLang1 := "en-us"
	userPromptResourceLang2 := "ja-jp"
	userPromptResourceTTS1 := "This is a test greeting!"
	userPromptResourceTTS2 := "This is a test greeting too!"
	userPromptResourceTTS3 := "こんにちは!"

	userPromptAsset1 := userPromptResourceStruct{
		userPromptResourceLang1,
		strconv.Quote(userPromptResourceTTS1),
		nullValue,
		nullValue,
	}

	userPromptAsset2 := userPromptResourceStruct{
		userPromptResourceLang1,
		strconv.Quote(userPromptResourceTTS2),
		nullValue,
		nullValue,
	}

	userPromptAsset3 := userPromptResourceStruct{
		userPromptResourceLang2,
		strconv.Quote(userPromptResourceTTS3),
		nullValue,
		nullValue,
	}

	userPromptResources1 := []*userPromptResourceStruct{&userPromptAsset1}
	userPromptResources2 := []*userPromptResourceStruct{&userPromptAsset2, &userPromptAsset3}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create Empty user prompt
				Config: generateUserPromptResource(&userPromptStruct{
					userPromptResource1,
					userPromptName1,
					strconv.Quote(userPromptDescription1),
					nil,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "name", userPromptName1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "description", userPromptDescription1),
				),
			},
			{
				// Update to include TTS message prompt resource
				Config: generateUserPromptResource(&userPromptStruct{
					userPromptResource1,
					userPromptName1,
					strconv.Quote(userPromptDescription1),
					userPromptResources1,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "name", userPromptName1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "description", userPromptDescription1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "resources.0.language", userPromptResourceLang1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "resources.0.tts_string", userPromptResourceTTS1),
				),
			},
			{
				// Update existing language TTS
				Config: generateUserPromptResource(&userPromptStruct{
					userPromptResource1,
					userPromptName1,
					strconv.Quote(userPromptDescription1),
					userPromptResources2,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "name", userPromptName1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "description", userPromptDescription1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "resources.0.language", userPromptResourceLang1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "resources.0.tts_string", userPromptResourceTTS2),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "resources.1.language", userPromptResourceLang2),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "resources.1.tts_string", userPromptResourceTTS3),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_architect_user_prompt." + userPromptResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyUserPromptsDestroyed,
	})
}

func TestAccResourceUserPromptWavFile(t *testing.T) {
	userPromptResource1 := "test-user_prompt_wav_file"
	userPromptName1 := "TestUserPromptWav_1"
	userPromptDescription1 := "Test prompt with wav audio file"
	userPromptResourceLang1 := "en-us"
	userPromptResourceText1 := "This is a test greeting!"
	userPromptResourceFileName1 := "test-prompt-01.wav"
	userPromptResourceFileName2 := "test-prompt-02.wav"

	userPromptAsset1 := userPromptResourceStruct{
		userPromptResourceLang1,
		nullValue,
		strconv.Quote(userPromptResourceText1),
		strconv.Quote(userPromptResourceFileName1),
	}

	userPromptAsset2 := userPromptResourceStruct{
		userPromptResourceLang1,
		nullValue,
		strconv.Quote(userPromptResourceText1),
		strconv.Quote(userPromptResourceFileName2),
	}

	userPromptResources1 := []*userPromptResourceStruct{&userPromptAsset1}
	userPromptResources2 := []*userPromptResourceStruct{&userPromptAsset2}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create user prompt with an audio file
				Config: generateUserPromptResource(&userPromptStruct{
					userPromptResource1,
					userPromptName1,
					strconv.Quote(userPromptDescription1),
					userPromptResources1,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "name", userPromptName1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "description", userPromptDescription1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "resources.0.filename", userPromptResourceFileName1),
				),
			},
			{
				// Replace audio file for the prompt
				Config: generateUserPromptResource(&userPromptStruct{
					userPromptResource1,
					userPromptName1,
					strconv.Quote(userPromptDescription1),
					userPromptResources2,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "name", userPromptName1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "description", userPromptDescription1),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResource1, "resources.0.filename", userPromptResourceFileName2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_architect_user_prompt." + userPromptResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyUserPromptsDestroyed,
	})
}

func generateUserPromptResource(userPrompt *userPromptStruct) string {
	resourcesString := ``
	for _, p := range userPrompt.resources {

		resourcesString += fmt.Sprintf(`resources {
            language = "%s"
            tts_string = %s
            text = %s
            filename = %s
        }
        `,
			p.language,
			p.tts_string,
			p.text,
			p.filename,
		)
	}

	return fmt.Sprintf(`resource "genesyscloud_architect_user_prompt" "%s" {
		name = "%s"
		description = %s
        %s
	}
	`, userPrompt.resourceID,
		userPrompt.name,
		userPrompt.description,
		resourcesString,
	)
}

func testVerifyUserPromptsDestroyed(state *terraform.State) error {
	architectAPI := platformclientv2.NewArchitectApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_architect_user_prompt" {
			continue
		}

		userPrompt, resp, err := architectAPI.GetArchitectPrompt(rs.Primary.ID)

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
