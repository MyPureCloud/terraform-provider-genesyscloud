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
				Config: GenerateBasicSpeechAndTextAnalyticsDictionaryFeedbackResource(resourceName, term, dialect, examplePhrase1, examplePhrase1, examplePhrase3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_speechandtextanalytics_dictionaryfeedback."+resourceName, "term", term),
					resource.TestCheckResourceAttr("genesyscloud_speechandtextanalytics_dictionaryfeedback."+resourceName, "dialect", dialect),
					resource.TestCheckResourceAttr("genesyscloud_speechandtextanalytics_dictionaryfeedback."+resourceName, "example_phrase.0.phrase", examplePhrase1),
					resource.TestCheckResourceAttr("genesyscloud_speechandtextanalytics_dictionaryfeedback."+resourceName, "example_phrase.1.phrase", examplePhrase2),
					resource.TestCheckResourceAttr("genesyscloud_speechandtextanalytics_dictionaryfeedback."+resourceName, "example_phrase.2.phrase", examplePhrase3),
				),
			},
			{
				// Update
				Config: GenerateFullSpeechAndTextAnalyticsDictionaryFeedbackResource(resourceName, term, dialect, boostValue, source, soundsLike, examplePhrase1, examplePhrase2, examplePhrase3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_speechandtextanalytics_dictionaryfeedback."+resourceName, "term", term),
					resource.TestCheckResourceAttr("genesyscloud_speechandtextanalytics_dictionaryfeedback."+resourceName, "dialect", dialect),
					resource.TestCheckResourceAttr("genesyscloud_speechandtextanalytics_dictionaryfeedback."+resourceName, "boost_value", boostValue),
					resource.TestCheckResourceAttr("genesyscloud_speechandtextanalytics_dictionaryfeedback."+resourceName, "source", source),
					resource.TestCheckResourceAttr("genesyscloud_speechandtextanalytics_dictionaryfeedback."+resourceName, "example_phrase.0.phrase", examplePhrase1),
					resource.TestCheckResourceAttr("genesyscloud_speechandtextanalytics_dictionaryfeedback."+resourceName, "example_phrase.0.source", source),
					resource.TestCheckResourceAttr("genesyscloud_speechandtextanalytics_dictionaryfeedback."+resourceName, "example_phrase.1.phrase", examplePhrase2),
					resource.TestCheckResourceAttr("genesyscloud_speechandtextanalytics_dictionaryfeedback."+resourceName, "example_phrase.1.source", source),
					resource.TestCheckResourceAttr("genesyscloud_speechandtextanalytics_dictionaryfeedback."+resourceName, "example_phrase.2.phrase", examplePhrase3),
					resource.TestCheckResourceAttr("genesyscloud_speechandtextanalytics_dictionaryfeedback."+resourceName, "example_phrase.2.source", source),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_speechandtextanalytics_dictionaryfeedback." + resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyDictionaryFeedbackDestroyed,
	})
}

func GenerateFullSpeechAndTextAnalyticsDictionaryFeedbackResource(resourceName, term, dialect, boostValue, source, soundsLike, examplePhrase1, examplePhrase2, examplePhrase3 string) string {
	return fmt.Sprintf(`resource "genesyscloud_speechandtextanalytics_dictionaryfeedback" "%s" {
	term = "%s"
	dialect = "%s"
	boost_value = "%s"
	source = "%s"
	sounds_like ["%s"]
	example_phrases { phrase = "%s" }
	example_phrases { phrase = "%s" }
	example_phrases { phrase = "%s" }
	}
	`, resourceName, term, dialect, boostValue, source, soundsLike, examplePhrase1, examplePhrase2, examplePhrase3)
}

func GenerateBasicSpeechAndTextAnalyticsDictionaryFeedbackResource(resourceLabel, term, dialect, examplePhrase1, examplePhrase2, examplePhrase3 string) string {
	return fmt.Sprintf(`resource "genesyscloud_speechandtextanalytics_dictionaryfeedback" "%s" {
		term = "%s"
		dialect = "%s"
		example_phrases { phrase = "%s" }
		example_phrases { phrase = "%s" }
		example_phrases { phrase = "%s" }
	}
	`, resourceLabel, term, dialect, examplePhrase1, examplePhrase2, examplePhrase3)
}

func testVerifyDictionaryFeedbackDestroyed(state *terraform.State) error {
	return nil
}
