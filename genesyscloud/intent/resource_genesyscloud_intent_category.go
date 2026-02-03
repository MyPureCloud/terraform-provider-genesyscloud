package intent_category

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v178/platformclientv2"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"fmt"
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
The resource_genesyscloud_intent_category.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthIntentCategory retrieves all of the intent category via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthIntentCategories(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newIntentCategoryProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	intentsCategorys, resp, err := proxy.getAllIntentCategory(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get intent category: %v", err), resp)
	}

	for _, intentsCategory := range *intentsCategorys {
		resources[*intentsCategory.Id] = &resourceExporter.ResourceMeta{BlockLabel: *intentsCategory.Name}
	}

	return resources, nil
}

// createIntentCategory is used by the intent_category resource to create Genesys cloud intent category
func createIntentCategory(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIntentCategoryProxy(sdkConfig)
	
	intentCategory := getIntentCategoryFromResourceData(d)

	log.Printf("Creating intent category %s", *intentCategory.Name)
	intentsCategory, resp, err := proxy.createIntentCategory(ctx, &intentCategory)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create intent category: %s", err), resp)
	}

	d.SetId(*intentsCategory.Id)
	log.Printf("Created intent category %s", *intentsCategory.Id)
	return readIntentCategory(ctx, d, meta)
}

// readIntentCategory is used by the intent_category resource to read an intent category from genesys cloud
func readIntentCategory(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIntentCategoryProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIntentCategory(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading intent category %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		intentsCategory, resp, getErr := proxy.getIntentCategoryById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read intent category %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read intent category %s: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", intentsCategory.Name)
		resourcedata.SetNillableValue(d, "description", intentsCategory.Description)
		

		log.Printf("Read intent category %s %s", d.Id(), *intentsCategory.Name)
		return cc.CheckState(d)
	})
}

// updateIntentCategory is used by the intent_category resource to update an intent category in Genesys Cloud
func updateIntentCategory(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIntentCategoryProxy(sdkConfig)
	
	intentCategory := getIntentCategoryFromResourceData(d)

	log.Printf("Updating intent category %s", *intentCategory.Name)
	intentsCategory, resp, err := proxy.updateIntentCategory(ctx, d.Id(), &intentCategory)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update intent category %s: %s", d.Id(), err), resp)
	}

	log.Printf("Updated intent category %s", *intentsCategory.Id)
	return readIntentCategory(ctx, d, meta)
}

// deleteIntentCategory is used by the intent_category resource to delete an intent category from Genesys cloud
func deleteIntentCategory(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIntentCategoryProxy(sdkConfig)
	
	resp, err := proxy.deleteIntentCategory(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete intent category %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getIntentCategoryById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted intent category %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting intent category %s: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("intent category %s still exists", d.Id()), resp))
	})
}

// getIntentCategoryFromResourceData maps data from schema ResourceData object to a platformclientv2.Intentscategory
func getIntentCategoryFromResourceData(d *schema.ResourceData) platformclientv2.Intentscategory {
	return platformclientv2.Intentscategory{
                        Name: platformclientv2.String(d.Get("name").(string)),
                Description: platformclientv2.String(d.Get("description").(string)),

	}
}

