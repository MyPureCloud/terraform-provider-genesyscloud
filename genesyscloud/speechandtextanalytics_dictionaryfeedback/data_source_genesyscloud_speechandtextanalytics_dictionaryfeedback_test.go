package speechandtextanalytics_dictionaryfeedback

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the dictionary feedback Data Source
*/

func TestAccDataSourceDictionaryFeedback(t *testing.T) {
	//t.Parallel()
	var (
		speechAndTextAnalyticsDictionaryFeedbackDataLabel     = "data-speechAndTextAnalyticsDictionaryFeedback"
		speechAndTextAnalyticsDictionaryFeedbackResourceLabel = "resource-speechAndTextAnalyticsDictionaryFeedback"

		term           = uniqueDictionaryTerm()
		dialect        = "en-AU"
		boostValue     = "2.0"
		source         = "Manual"
		soundsLike     = "genesis"
		examplePhrase1 = fmt.Sprintf("welcome to %s", term)
		examplePhrase2 = fmt.Sprintf("thanks for calling %s", term)
		examplePhrase3 = fmt.Sprintf("%s is a platform", term)
	)

	// Clean up any leftover dictionary feedback from previous test runs
	cleanupDictionaryFeedbackByTerm(term, dialect)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create feedback setting
				Config: GenerateFullSpeechAndTextAnalyticsDictionaryFeedbackResource(ResourceType, speechAndTextAnalyticsDictionaryFeedbackResourceLabel, term, dialect, boostValue, source, soundsLike, examplePhrase1, examplePhrase2, examplePhrase3) + generateSpeechAndTextAnalyticsDictionaryFeedbackDataSource(ResourceType, speechAndTextAnalyticsDictionaryFeedbackDataLabel, term, ResourceType+"."+speechAndTextAnalyticsDictionaryFeedbackResourceLabel),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data."+ResourceType+"."+speechAndTextAnalyticsDictionaryFeedbackDataLabel, "id",
						""+ResourceType+"."+speechAndTextAnalyticsDictionaryFeedbackResourceLabel, "id",
					),
				),
			},
		},
	})
}

func generateSpeechAndTextAnalyticsDictionaryFeedbackDataSource(resourceType, resourceName, term, dependsOn string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		term = "%s"
		dialect = "en-AU"
		transcription_engine = "%s"
		depends_on = [%s]
	}
	`, resourceType, resourceName, term, TranscriptionEngineGenesys, dependsOn)
}
