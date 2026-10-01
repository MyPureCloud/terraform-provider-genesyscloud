package speechandtextanalytics_category

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

/*
The resource_genesyscloud_speechandtextanalytics_category.go contains all of the methods that perform the core logic for a resource.
*/

// getAllCategories retrieves all of the categories via Terraform in the Genesys Cloud and is used for the exporter
func getAllCategories(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newCategoryProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	categories, resp, err := proxy.getAllCategories(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get categories: %s", err), resp)
	}

	for _, category := range *categories {
		if category.Id != nil && category.Name != nil {
			resources[*category.Id] = &resourceExporter.ResourceMeta{BlockLabel: *category.Name}
		}
	}
	return resources, nil
}

// createCategory is used by the category resource to create a Genesys Cloud category
func createCategory(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCategoryProxy(sdkConfig)

	category, err := buildCategoryFromResourceData(d)
	if err != nil {
		return diag.Errorf("Failed to build category from resource data: %s", err)
	}

	name := d.Get("name").(string)
	log.Printf("Creating category %s", name)

	createdCategory, resp, createErr := proxy.createCategory(ctx, category)
	if createErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create category %s: %s", name, createErr), resp)
	}
	if createdCategory == nil || createdCategory.Id == nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create category %s: API returned no id", name), resp)
	}

	d.SetId(*createdCategory.Id)
	log.Printf("Created category %s %s", name, *createdCategory.Id)
	return readCategory(ctx, d, meta)
}

// readCategory is used by the category resource to read a category from Genesys Cloud
func readCategory(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCategoryProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceCategory(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading category %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		category, resp, getErr := proxy.getCategoryById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read category %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read category %s | error: %s", d.Id(), getErr), resp))
		}

		setCategoryToResourceData(d, category)

		log.Printf("Read category %s", d.Id())
		return cc.CheckState(d)
	})
}

// updateCategory is used by the category resource to update a category in Genesys Cloud
func updateCategory(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCategoryProxy(sdkConfig)

	category, err := buildCategoryFromResourceData(d)
	if err != nil {
		return diag.Errorf("Failed to build category from resource data: %s", err)
	}

	log.Printf("Updating category %s", d.Id())
	_, resp, updateErr := proxy.updateCategory(ctx, d.Id(), category)
	if updateErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update category %s: %s", d.Id(), updateErr), resp)
	}

	log.Printf("Updated category %s", d.Id())
	return readCategory(ctx, d, meta)
}

// deleteCategory is used by the category resource to delete a category from Genesys Cloud
func deleteCategory(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCategoryProxy(sdkConfig)

	log.Printf("Deleting category %s", d.Id())
	resp, deleteErr := proxy.deleteCategory(ctx, d.Id())
	if deleteErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete category %s: %s", d.Id(), deleteErr), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getCategoryById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted category %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting category %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Category %s still exists", d.Id()), resp))
	})
}
