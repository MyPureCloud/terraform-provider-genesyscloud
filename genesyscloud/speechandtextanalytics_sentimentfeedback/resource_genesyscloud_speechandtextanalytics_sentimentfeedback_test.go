package speechandtextanalytics_sentimentfeedback

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
The resource_genesyscloud_speechandtextanalytics_sentimentfeedback_test.go contains all of the test cases for running the resource
tests for speechandtextanalytics_sentimentfeedback.
*/

// cleanupSentimentFeedbackByPhrase removes any existing sentiment feedback entries matching the given phrase and dialect
func cleanupSentimentFeedbackByPhrase(phrase, dialect string) {
	sdkConfig, err := provider.AuthorizeSdk()
	if err != nil {
		log.Printf("failed to authorize SDK for sentiment feedback cleanup: %v", err)
		return
	}
	api := platformclientv2.NewSpeechTextAnalyticsApiWithConfig(sdkConfig)

	feedbacks, _, err := api.GetSpeechandtextanalyticsSentimentfeedback(dialect)
	if err != nil {
		log.Printf("failed to list sentiment feedback for cleanup: %v", err)
		return
	}
	if feedbacks.Entities == nil {
		return
	}
	for _, fb := range *feedbacks.Entities {
		if fb.Id == nil || fb.Phrase == nil || *fb.Phrase != phrase {
			continue
		}
		log.Printf("Cleaning up sentiment feedback %s (phrase=%s, dialect=%s)", *fb.Id, phrase, dialect)
		for attempt := 0; attempt < 5; attempt++ {
			_, err := api.DeleteSpeechandtextanalyticsSentimentfeedbackSentimentFeedbackId(*fb.Id)
			if err == nil {
				break
			}
			log.Printf("attempt %d: failed to delete sentiment feedback %s: %v", attempt+1, *fb.Id, err)
			time.Sleep(3 * time.Second)
		}
	}
}

func TestAccResourceSentimentFeedback(t *testing.T) {
	var (
		resourceLabel = "test-sentiment-feedback"
		phrase        = "the service was excellent"
		dialect       = "en-US"
		feedbackValue = FeedbackValuePositive
	)

	// Clean up any leftover sentiment feedback from previous test runs
	cleanupSentimentFeedbackByPhrase(phrase, dialect)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateSentimentFeedbackResource(resourceLabel, phrase, dialect, feedbackValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phrase", phrase),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "dialect", dialect),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "feedback_value", feedbackValue),
					resource.TestCheckResourceAttrSet(ResourceType+"."+resourceLabel, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySentimentFeedbackDestroyed,
	})
}

// TestAccResourceSentimentFeedbackForceNew covers "modify": every attribute is ForceNew, so changing
// feedback_value replaces the resource. It also exercises the Negative and Neutral values.
func TestAccResourceSentimentFeedbackForceNew(t *testing.T) {
	var (
		resourceLabel = "test-sentiment-feedback-forcenew"
		phrase        = "the wait time was unacceptable"
		dialect       = "en-US"
		resourcePath  = ResourceType + "." + resourceLabel
		firstID       string
		secondID      string
	)

	cleanupSentimentFeedbackByPhrase(phrase, dialect)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateSentimentFeedbackResource(resourceLabel, phrase, dialect, FeedbackValueNegative),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "feedback_value", FeedbackValueNegative),
					testCaptureResourceID(resourcePath, &firstID),
				),
			},
			{
				// Changing feedback_value must replace the resource (new id).
				Config: generateSentimentFeedbackResource(resourceLabel, phrase, dialect, FeedbackValueNeutral),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "feedback_value", FeedbackValueNeutral),
					testCaptureResourceID(resourcePath, &secondID),
					func(_ *terraform.State) error {
						if firstID == "" || firstID == secondID {
							return fmt.Errorf("expected feedback_value change to replace the resource, id stayed %q", firstID)
						}
						return nil
					},
				),
			},
			{
				ResourceName:      resourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySentimentFeedbackDestroyed,
	})
}

func testCaptureResourceID(resourcePath string, dest *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourcePath)
		}
		*dest = rs.Primary.ID
		return nil
	}
}

func generateSentimentFeedbackResource(resourceLabel, phrase, dialect, feedbackValue string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
	phrase         = "%s"
	dialect        = "%s"
	feedback_value = "%s"
}
`, ResourceType, resourceLabel, phrase, dialect, feedbackValue)
}

func testVerifySentimentFeedbackDestroyed(state *terraform.State) error {
	sdkConfig, err := provider.AuthorizeSdk()
	if err != nil {
		return fmt.Errorf("failed to authorize SDK for sentiment feedback destroy check: %v", err)
	}
	api := platformclientv2.NewSpeechTextAnalyticsApiWithConfig(sdkConfig)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		feedbacks, resp, err := api.GetSpeechandtextanalyticsSentimentfeedback("")
		if err != nil {
			if util.IsStatus404(resp) {
				continue
			}
			return fmt.Errorf("unexpected error listing sentiment feedback: %v", err)
		}
		if feedbacks.Entities == nil {
			continue
		}
		for _, fb := range *feedbacks.Entities {
			if fb.Id != nil && *fb.Id == rs.Primary.ID {
				return fmt.Errorf("sentiment feedback %s still exists", rs.Primary.ID)
			}
		}
	}
	// Success. All sentiment feedback destroyed
	return nil
}
