package speechandtextanalytics_sentimentfeedback

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
)

/*
The resource_genesyscloud_speechandtextanalytics_sentimentfeedback.go contains all of the methods that perform the core logic for a resource.

The sentiment feedback API does not support an update operation, so this resource only implements Create, Read and Delete.
Any change to a configured attribute forces the resource to be recreated (all writable attributes are ForceNew).
*/

// getAllSentimentFeedbacks retrieves all of the sentiment feedback via Terraform in the Genesys Cloud and is used for the exporter
func getAllSentimentFeedbacks(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getSentimentFeedbackProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	log.Println("Getting all sentiment feedback")
	sentimentFeedbacks, resp, err := proxy.getAllSentimentFeedback(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get all sentiment feedback: %v", err), resp)
	}

	for _, sentimentFeedback := range *sentimentFeedbacks {
		resources[*sentimentFeedback.Id] = &resourceExporter.ResourceMeta{BlockLabel: sentimentFeedbackExportLabel(sentimentFeedback)}
	}

	return resources, nil
}

// createSentimentFeedback is used by the speechandtextanalytics_sentimentfeedback resource to create a Genesys Cloud sentiment feedback
func createSentimentFeedback(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSentimentFeedbackProxy(sdkConfig)

	sentimentFeedback := getSentimentFeedbackFromResourceData(d)

	log.Printf("Creating sentiment feedback for phrase %s", *sentimentFeedback.Phrase)
	sentimentFeedbackPtr, resp, err := proxy.createSentimentFeedback(ctx, &sentimentFeedback)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create sentiment feedback: %s", err), resp)
	}

	d.SetId(*sentimentFeedbackPtr.Id)
	log.Printf("Created sentiment feedback %s", *sentimentFeedbackPtr.Id)
	return readSentimentFeedback(ctx, d, meta)
}

// readSentimentFeedback is used by the speechandtextanalytics_sentimentfeedback resource to read a sentiment feedback from Genesys Cloud
func readSentimentFeedback(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSentimentFeedbackProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceSentimentFeedback(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading sentiment feedback %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sentimentFeedback, resp, getErr := proxy.getSentimentFeedbackById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read sentiment feedback %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read sentiment feedback %s: %s", d.Id(), getErr), resp))
		}

		setSentimentFeedbackToResourceData(d, sentimentFeedback)

		log.Printf("Read sentiment feedback %s", d.Id())
		return cc.CheckState(d)
	})
}

// deleteSentimentFeedback is used by the speechandtextanalytics_sentimentfeedback resource to delete a sentiment feedback from Genesys Cloud
func deleteSentimentFeedback(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSentimentFeedbackProxy(sdkConfig)

	resp, err := proxy.deleteSentimentFeedback(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete sentiment feedback %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getSentimentFeedbackById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted sentiment feedback %s", d.Id())
				return nil
			}
			// The DELETE call already succeeded; a non-404 error here (e.g. a transient 5xx or
			// timeout on the confirmation list) should be retried rather than aborting the delete.
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error verifying deletion of sentiment feedback %s: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("sentiment feedback %s still exists", d.Id()), resp))
	})
}
