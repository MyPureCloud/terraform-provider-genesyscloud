package speechandtextanalytics_dictionaryfeedback

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
)

/*
The resource_genesyscloud_speechandtextanalytics_dictionaryfeedback_test.go contains all of the test cases for running the resource
tests for speechandtextanalytics_dictionaryfeedback.
*/

// cleanupDictionaryFeedbackByTerm removes any existing dictionary feedback entries matching the given term and dialect
func cleanupDictionaryFeedbackByTerm(term, dialect string) {
	sdkConfig, err := provider.AuthorizeSdk()
	if err != nil {
		log.Printf("failed to authorize SDK for dictionary feedback cleanup: %v", err)
		return
	}
	api := platformclientv2.NewSpeechTextAnalyticsApiWithConfig(sdkConfig)
	feedbacks, _, err := api.GetSpeechandtextanalyticsDictionaryfeedback(dialect, term, "", 100)
	if err != nil {
		log.Printf("failed to list dictionary feedback for cleanup: %v", err)
		return
	}
	if feedbacks.Entities == nil {
		return
	}
	for _, fb := range *feedbacks.Entities {
		if fb.Id != nil {
			log.Printf("Cleaning up dictionary feedback %s (term=%s, dialect=%s)", *fb.Id, term, dialect)
			for attempt := 0; attempt < 5; attempt++ {
				_, err := api.DeleteSpeechandtextanalyticsDictionaryfeedbackDictionaryFeedbackId(*fb.Id)
				if err == nil {
					break
				}
				log.Printf("attempt %d: failed to delete dictionary feedback %s: %v", attempt+1, *fb.Id, err)
				time.Sleep(3 * time.Second)
			}
		}
	}
}

func TestAccResourceDictionaryFeedback(t *testing.T) {
	//t.Parallel()
	if v := os.Getenv("GENESYSCLOUD_REGION"); v == "us-east-1" {
		t.Skipf("virtualAgent product not available in %s org", v)
		return
	}
	var (
		resourceName   = "test-dictionary-feedback"
		term           = "genesys"
		dialect        = "en-AU"
		boostValue     = "2"
		source         = "Manual"
		soundsLike     = "genesis"
		examplePhrase1 = "welcome to genesys"
		examplePhrase2 = "thanks for calling genesys"
		examplePhrase3 = "Genesys is a platform"
	)

	// Clean up any leftover dictionary feedback from previous test runs
	cleanupDictionaryFeedbackByTerm(term, dialect)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateBasicSpeechAndTextAnalyticsDictionaryFeedbackResource(ResourceType, resourceName, term, dialect, examplePhrase1, examplePhrase2, examplePhrase3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceName, "term", term),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceName, "dialect", dialect),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceName, "example_phrases.0.phrase", examplePhrase1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceName, "example_phrases.1.phrase", examplePhrase2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceName, "example_phrases.2.phrase", examplePhrase3),
				),
			},
			{
				// Update
				Config: GenerateFullSpeechAndTextAnalyticsDictionaryFeedbackResource(ResourceType, resourceName, term, dialect, boostValue, source, soundsLike, examplePhrase1, examplePhrase2, examplePhrase3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceName, "term", term),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceName, "dialect", dialect),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceName, "boost_value", boostValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceName, "source", source),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceName, "example_phrases.0.phrase", examplePhrase1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceName, "example_phrases.0.source", source),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceName, "example_phrases.1.phrase", examplePhrase2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceName, "example_phrases.1.source", source),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceName, "example_phrases.2.phrase", examplePhrase3),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceName, "example_phrases.2.source", source),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyDictionaryFeedbackDestroyed,
	})
}

func GenerateFullSpeechAndTextAnalyticsDictionaryFeedbackResource(resourceType, resourceName, term, dialect, boostValue, source, soundsLike, examplePhrase1, examplePhrase2, examplePhrase3 string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
	term = "%s"
	dialect = "%s"
	boost_value = "%s"
	source = "%s"
	sounds_like = ["%s"]
	example_phrases { phrase = "%s" }
	example_phrases { phrase = "%s" }
	example_phrases { phrase = "%s" }
	}
	`, resourceType, resourceName, term, dialect, boostValue, source, soundsLike, examplePhrase1, examplePhrase2, examplePhrase3)
}

func GenerateBasicSpeechAndTextAnalyticsDictionaryFeedbackResource(resourceType, resourceLabel, term, dialect, examplePhrase1, examplePhrase2, examplePhrase3 string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		term = "%s"
		dialect = "%s"
		example_phrases { phrase = "%s" }
		example_phrases { phrase = "%s" }
		example_phrases { phrase = "%s" }
	}
	`, ResourceType, resourceLabel, term, dialect, examplePhrase1, examplePhrase2, examplePhrase3)
}

func testVerifyDictionaryFeedbackDestroyed(state *terraform.State) error {
	return nil
}
