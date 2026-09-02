package speechandtextanalytics_sentimentfeedback

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
Test Class for the sentiment feedback Data Source
*/

func TestAccDataSourceSentimentFeedback(t *testing.T) {
	var (
		dataLabel     = "data-sentiment-feedback"
		resourceLabel = "resource-sentiment-feedback"

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
				Config: generateSentimentFeedbackResource(resourceLabel, phrase, dialect, feedbackValue) +
					generateSentimentFeedbackDataSource(dataLabel, phrase, dialect, ResourceType+"."+resourceLabel),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data."+ResourceType+"."+dataLabel, "id",
						ResourceType+"."+resourceLabel, "id",
					),
				),
			},
		},
	})
}

func generateSentimentFeedbackDataSource(dataLabel, phrase, dialect, dependsOn string) string {
	return fmt.Sprintf(`data "%s" "%s" {
	phrase     = "%s"
	dialect    = "%s"
	depends_on = [%s]
}
`, ResourceType, dataLabel, phrase, dialect, dependsOn)
}
