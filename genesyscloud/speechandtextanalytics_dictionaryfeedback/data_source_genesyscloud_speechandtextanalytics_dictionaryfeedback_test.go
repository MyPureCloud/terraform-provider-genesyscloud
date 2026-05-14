package speechandtextanalytics_dictionaryfeedback

import (
	"fmt"
	"os"
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
	if v := os.Getenv("GENESYSCLOUD_REGION"); v == "us-east-1" {
		t.Skipf("virtualAgent product not available in %s org", v)
		return
	}
	var (
		speechAndTextAnalyticsDictionaryFeedbackDataLabel     = "data-speechAndTextAnalyticsDictionaryFeedback"
		speechAndTextAnalyticsDictionaryFeedbackResourceLabel = "resource-speechAndTextAnalyticsDictionaryFeedback"

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
		depends_on = [%s]
	}
	`, resourceType, resourceName, term, dependsOn)
}
