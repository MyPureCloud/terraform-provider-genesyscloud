package customer_intent

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_customer_intent.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthCustomerIntent retrieves all of the customer intent via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthCustomerIntents(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getCustomerIntentProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	customerIntentResponses, resp, err := proxy.getAllCustomerIntent(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get customer intent: %v", err), resp)
	}

	for _, customerIntentResponse := range *customerIntentResponses {
		resources[*customerIntentResponse.Id] = &resourceExporter.ResourceMeta{BlockLabel: *customerIntentResponse.Name}
	}

	return resources, nil
}

// createCustomerIntent is used by the customer_intent resource to create Genesys cloud customer intent
func createCustomerIntent(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCustomerIntentProxy(sdkConfig)

	customerIntent := getCustomerIntentFromResourceData(d)

	log.Printf("Creating customer intent %s", *customerIntent.Name)
	customerIntentResponse, resp, err := proxy.createCustomerIntent(ctx, &customerIntent)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create customer intent: %s", err), resp)
	}

	d.SetId(*customerIntentResponse.Id)
	log.Printf("Created customer intent %s", *customerIntentResponse.Id)

	// Add source intents if specified
	if sourceIntents := getSourceIntentsFromResourceData(d); len(sourceIntents) > 0 {
		log.Printf("Adding %d source intents to customer intent %s", len(sourceIntents), *customerIntentResponse.Id)
		_, resp, err := proxy.bulkAddSourceIntents(ctx, *customerIntentResponse.Id, sourceIntents)
		if err != nil {
			// Rollback: delete the customer intent since sub-resource creation failed
			log.Printf("Failed to add source intents, rolling back customer intent %s", d.Id())
			if delResp, delErr := proxy.deleteCustomerIntent(ctx, d.Id()); delErr != nil {
				log.Printf("Failed to rollback customer intent %s: %s (resp: %v)", d.Id(), delErr, delResp)
			}
			d.SetId("")
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to add source intents: %s", err), resp)
		}
	}

	return readCustomerIntent(ctx, d, meta)
}

// readCustomerIntent is used by the customer_intent resource to read an customer intent from genesys cloud
func readCustomerIntent(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCustomerIntentProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceCustomerIntent(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading customer intent %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		customerIntentResponse, resp, getErr := proxy.getCustomerIntentById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read customer intent %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read customer intent %s: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", customerIntentResponse.Name)
		resourcedata.SetNillableValue(d, "description", customerIntentResponse.Description)
		resourcedata.SetNillableValue(d, "expiry_time", customerIntentResponse.ExpiryTime)
		if customerIntentResponse.Category != nil {
			resourcedata.SetNillableValue(d, "category_id", customerIntentResponse.Category.Id)
		}

		// Read source intents
		sourceIntents, resp, err := proxy.getSourceIntents(ctx, d.Id())
		if err != nil {
			log.Printf("Failed to read source intents for customer intent %s: %s", d.Id(), err)
		} else if sourceIntents != nil {
			d.Set("source_intents", flattenSourceIntents(*sourceIntents))
		}

		log.Printf("Read customer intent %s %s", d.Id(), *customerIntentResponse.Name)
		return cc.CheckState(d)
	})
}

// updateCustomerIntent is used by the customer_intent resource to update an customer intent in Genesys Cloud
func updateCustomerIntent(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCustomerIntentProxy(sdkConfig)

	customerIntent := getCustomerIntentFromResourceData(d)

	log.Printf("Updating customer intent %s", *customerIntent.Name)
	customerIntentResponse, resp, err := proxy.updateCustomerIntent(ctx, d.Id(), &customerIntent)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update customer intent %s: %s", d.Id(), err), resp)
	}

	// Handle source intent changes
	if d.HasChange("source_intents") {
		oldIntents, newIntents := d.GetChange("source_intents")
		oldSet := oldIntents.(*schema.Set)
		newSet := newIntents.(*schema.Set)

		// Remove old source intents that are no longer in the new set
		toRemove := oldSet.Difference(newSet)
		if toRemove.Len() > 0 {
			sourceIntentsToRemove := buildSourceIntentsFromSet(toRemove)
			log.Printf("Removing %d source intents from customer intent %s", len(sourceIntentsToRemove), d.Id())
			_, resp, err := proxy.bulkRemoveSourceIntents(ctx, d.Id(), sourceIntentsToRemove)
			if err != nil {
				return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to remove source intents: %s", err), resp)
			}
		}

		// Add new source intents that weren't in the old set
		toAdd := newSet.Difference(oldSet)
		if toAdd.Len() > 0 {
			sourceIntentsToAdd := buildSourceIntentsFromSet(toAdd)
			log.Printf("Adding %d source intents to customer intent %s", len(sourceIntentsToAdd), d.Id())
			_, resp, err := proxy.bulkAddSourceIntents(ctx, d.Id(), sourceIntentsToAdd)
			if err != nil {
				return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to add source intents: %s", err), resp)
			}
		}
	}

	log.Printf("Updated customer intent %s", *customerIntentResponse.Id)
	return readCustomerIntent(ctx, d, meta)
}

// deleteCustomerIntent is used by the customer_intent resource to delete an customer intent from Genesys cloud
func deleteCustomerIntent(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCustomerIntentProxy(sdkConfig)

	resp, err := proxy.deleteCustomerIntent(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete customer intent %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getCustomerIntentById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted customer intent %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting customer intent %s: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("customer intent %s still exists", d.Id()), resp))
	})
}

// getCustomerIntentFromResourceData maps data from schema ResourceData object to a platformclientv2.Customerintentresponse
