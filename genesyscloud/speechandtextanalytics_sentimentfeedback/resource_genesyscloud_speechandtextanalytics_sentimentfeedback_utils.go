package speechandtextanalytics_sentimentfeedback

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

/*
The resource_genesyscloud_speechandtextanalytics_sentimentfeedback_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getSentimentFeedbackFromResourceData maps data from schema ResourceData object to a platformclientv2.Sentimentfeedback.
// Read-only fields (Id, DateCreated, CreatedBy) are intentionally not set on the create body.
func getSentimentFeedbackFromResourceData(d *schema.ResourceData) platformclientv2.Sentimentfeedback {
	return platformclientv2.Sentimentfeedback{
		Phrase:        platformclientv2.String(d.Get("phrase").(string)),
		Dialect:       platformclientv2.String(d.Get("dialect").(string)),
		FeedbackValue: platformclientv2.String(d.Get("feedback_value").(string)),
	}
}

// setSentimentFeedbackToResourceData maps a platformclientv2.Sentimentfeedback into the schema ResourceData object.
// This is shared by the resource read and the data source read.
func setSentimentFeedbackToResourceData(d *schema.ResourceData, sentimentFeedback *platformclientv2.Sentimentfeedback) {
	resourcedata.SetNillableValue(d, "phrase", sentimentFeedback.Phrase)
	resourcedata.SetNillableValue(d, "dialect", sentimentFeedback.Dialect)
	resourcedata.SetNillableValue(d, "feedback_value", sentimentFeedback.FeedbackValue)
}

// sentimentFeedbackMatchesFilters returns true when the given sentiment feedback matches the phrase and (optional) dialect
func sentimentFeedbackMatchesFilters(sentimentFeedback platformclientv2.Sentimentfeedback, phrase string, dialect string) bool {
	if sentimentFeedback.Phrase == nil || *sentimentFeedback.Phrase != phrase {
		return false
	}
	if dialect != "" && (sentimentFeedback.Dialect == nil || *sentimentFeedback.Dialect != dialect) {
		return false
	}
	return true
}

// sentimentFeedbackExportLabel builds a stable, human-readable export label for a sentiment feedback
func sentimentFeedbackExportLabel(sentimentFeedback platformclientv2.Sentimentfeedback) string {
	label := ""
	if sentimentFeedback.Phrase != nil {
		label = *sentimentFeedback.Phrase
	}
	if sentimentFeedback.Dialect != nil && *sentimentFeedback.Dialect != "" {
		label = fmt.Sprintf("%s_%s", label, *sentimentFeedback.Dialect)
	}
	if sentimentFeedback.FeedbackValue != nil && *sentimentFeedback.FeedbackValue != "" {
		label = fmt.Sprintf("%s_%s", label, *sentimentFeedback.FeedbackValue)
	}
	return label
}
