package customer_intent

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v178/platformclientv2"
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
	proxy := newCustomerIntentProxy(clientConfig)
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
	return readCustomerIntent(ctx, d, meta)
}

// readCustomerIntent is used by the customer_intent resource to read an customer intent from genesys cloud
func readCustomerIntent(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCustomerIntentProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceCustomerIntent(), constants.ConsistencyChecks(), resourceName)

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
func getCustomerIntentFromResourceData(d *schema.ResourceData) platformclientv2.Customerintentresponse {
	categoryId := d.Get("category_id").(string)
	return platformclientv2.Customerintentresponse{
		Name:        platformclientv2.String(d.Get("name").(string)),
		Description: platformclientv2.String(d.Get("description").(string)),
		ExpiryTime:  platformclientv2.Int(d.Get("expiry_time").(int)),
		Category: &platformclientv2.Addressableentityref{
			Id: &categoryId,
		},
	}
}
