package speechandtextanalytics_dictionaryfeedback

import (
	"fmt"
	"os"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_speechandtextanalytics_dictionaryfeedback_test.go contains all of the test cases for running the resource
tests for speechandtextanalytics_dictionaryfeedback.
*/

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
		boostValue     = "2.0"
		source         = "Manual"
		soundsLike     = "genesis"
		examplePhrase1 = "welcome to genesys"
		examplePhrase2 = "thanks for calling genesys"
		examplePhrase3 = "Genesys is a platform"
	)

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
