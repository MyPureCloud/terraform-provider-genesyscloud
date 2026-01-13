package speechandtextanalytics_dictionaryfeedback

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_speechandtextanalytics_dictionaryfeedback.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthDictionaryFeedback retrieves all of the dictionary feedback via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthDictionaryFeedbacks(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newDictionaryFeedbackProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	dictionaryFeedbacks, resp, err := proxy.getAllDictionaryFeedback(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get dictionary feedback: %v", err), resp)
	}

	for _, dictionaryFeedback := range *dictionaryFeedbacks {
		resources[*dictionaryFeedback.Id] = &resourceExporter.ResourceMeta{BlockLabel: *dictionaryFeedback.Term}
	}

	return resources, nil
}

// createDictionaryFeedback is used by the speechandtextanalytics_dictionaryfeedback resource to create Genesys cloud dictionary feedback
func createDictionaryFeedback(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getDictionaryFeedbackProxy(sdkConfig)

	dictionaryFeedback := getDictionaryFeedbackFromResourceData(d)

	// validate that term is in the phrases and length of phrases List
	err := validateExamplePhrases(d)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("Creating dictionary feedback %s", *dictionaryFeedback.Term)
	dictionaryFeedbackPtr, resp, err := proxy.createDictionaryFeedback(ctx, &dictionaryFeedback)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create dictionary feedback: %s", err), resp)
	}

	d.SetId(*dictionaryFeedbackPtr.Id)
	log.Printf("Created dictionary feedback %s", *dictionaryFeedbackPtr.Id)
	return readDictionaryFeedback(ctx, d, meta)
}

// readDictionaryFeedback is used by the speechandtextanalytics_dictionaryfeedback resource to read an dictionary feedback from genesys cloud
func readDictionaryFeedback(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getDictionaryFeedbackProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceDictionaryFeedback(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading dictionary feedback %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		dictionaryFeedback, resp, getErr := proxy.getDictionaryFeedbackById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read dictionary feedback %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read dictionary feedback %s: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "term", dictionaryFeedback.Term)
		resourcedata.SetNillableValue(d, "dialect", dictionaryFeedback.Dialect)
		resourcedata.SetNillableValue(d, "boost_value", dictionaryFeedback.BoostValue)
		resourcedata.SetNillableValue(d, "source", dictionaryFeedback.Source)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "example_phrases", dictionaryFeedback.ExamplePhrases, flattenDictionaryFeedbackExamplePhrases)
		resourcedata.SetNillableValue(d, "sounds_like", dictionaryFeedback.SoundsLike)

		log.Printf("Read dictionary feedback %s %s", d.Id(), *dictionaryFeedback.Term)
		return cc.CheckState(d)
	})
}

// updateDictionaryFeedback is used by the speechandtextanalytics_dictionaryfeedback resource to update an dictionary feedback in Genesys Cloud
func updateDictionaryFeedback(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getDictionaryFeedbackProxy(sdkConfig)

	dictionaryFeedback := getDictionaryFeedbackFromResourceData(d)

	// validate that term is in the phrases and length of phrases List
	err := validateExamplePhrases(d)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("Updating dictionary feedback %s", *dictionaryFeedback.Term)
	dictionaryFeedbackPtr, resp, err := proxy.updateDictionaryFeedback(ctx, d.Id(), &dictionaryFeedback)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update dictionary feedback %s: %s", d.Id(), err), resp)
	}

	log.Printf("Updated dictionary feedback %s", *dictionaryFeedbackPtr.Id)
	return readDictionaryFeedback(ctx, d, meta)
}

// deleteDictionaryFeedback is used by the speechandtextanalytics_dictionaryfeedback resource to delete an dictionary feedback from Genesys cloud
func deleteDictionaryFeedback(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getDictionaryFeedbackProxy(sdkConfig)

	resp, err := proxy.deleteDictionaryFeedback(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete dictionary feedback %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getDictionaryFeedbackById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted dictionary feedback %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting dictionary feedback %s: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("dictionary feedback %s still exists", d.Id()), resp))
	})
}
