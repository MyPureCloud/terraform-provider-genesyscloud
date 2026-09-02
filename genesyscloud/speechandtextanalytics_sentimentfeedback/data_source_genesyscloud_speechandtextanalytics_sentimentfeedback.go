package speechandtextanalytics_sentimentfeedback

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
The data_source_genesyscloud_speechandtextanalytics_sentimentfeedback.go contains the data source implementation
for the resource.
*/

// dataSourceSentimentFeedbackRead retrieves the sentiment feedback id by phrase and dialect
func dataSourceSentimentFeedbackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSentimentFeedbackProxy(sdkConfig)

	phrase := d.Get("phrase").(string)
	dialect := d.Get("dialect").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		sentimentFeedbackId, resp, retryable, err := proxy.getSentimentFeedbackIdByPhrase(ctx, phrase, dialect)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error searching sentiment feedback %s | error: %s", phrase, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No sentiment feedback found with phrase %s", phrase), resp))
		}

		d.SetId(sentimentFeedbackId)
		return nil
	})
}
