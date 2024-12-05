package responsemanagement_library

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_responsemanagement_library.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthResponsemanagementLibrary retrieves all of the responsemanagement library via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthResponsemanagementLibrarys(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newResponsemanagementLibraryProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	librarys, resp, err := proxy.getAllResponsemanagementLibrary(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get responsemanagement library error: %s", err), resp)
	}

	for _, library := range *librarys {
		resources[*library.Id] = &resourceExporter.ResourceMeta{BlockLabel: *library.Name}
	}
	return resources, nil
}

// createResponsemanagementLibrary is used by the responsemanagement_library resource to create Genesys cloud responsemanagement library
func createResponsemanagementLibrary(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getResponsemanagementLibraryProxy(sdkConfig)
	responsemanagementLibrary := platformclientv2.Library{
		Name: platformclientv2.String(d.Get("name").(string)),
	}

	log.Printf("Creating responsemanagement library %s", *responsemanagementLibrary.Name)
	library, resp, err := proxy.createResponsemanagementLibrary(ctx, &responsemanagementLibrary)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create responsemanagement library %s error: %s", *responsemanagementLibrary.Name, err), resp)
	}

	d.SetId(*library.Id)
	log.Printf("Created responsemanagement library %s", *library.Id)
	return readResponsemanagementLibrary(ctx, d, meta)
}

// readResponsemanagementLibrary is used by the responsemanagement_library resource to read an responsemanagement library from genesys cloud
func readResponsemanagementLibrary(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getResponsemanagementLibraryProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceResponsemanagementLibrary(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading responsemanagement library %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		library, resp, getErr := proxy.getResponsemanagementLibraryById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read responsemanagement library %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read responsemanagement library %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", library.Name)
		log.Printf("Read responsemanagement library %s %s", d.Id(), *library.Name)
		return cc.CheckState(d)
	})
}

// updateResponsemanagementLibrary is used by the responsemanagement_library resource to update an responsemanagement library in Genesys Cloud
func updateResponsemanagementLibrary(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getResponsemanagementLibraryProxy(sdkConfig)

	responsemanagementLibrary := platformclientv2.Library{
		Name: platformclientv2.String(d.Get("name").(string)),
	}

	log.Printf("Updating responsemanagement library %s", *responsemanagementLibrary.Name)
	library, resp, err := proxy.updateResponsemanagementLibrary(ctx, d.Id(), &responsemanagementLibrary)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update responsemanagement library %s error: %s", *responsemanagementLibrary.Name, err), resp)
	}
	log.Printf("Updated responsemanagement library %s", *library.Id)
	return readResponsemanagementLibrary(ctx, d, meta)
}

// deleteResponsemanagementLibrary is used by the responsemanagement_library resource to delete an responsemanagement library from Genesys cloud
func deleteResponsemanagementLibrary(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getResponsemanagementLibraryProxy(sdkConfig)

	resp, err := proxy.deleteResponsemanagementLibrary(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete responsemanagement library %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getResponsemanagementLibraryById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted responsemanagement library %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting responsemanagement library %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("responsemanagement library %s still exists", d.Id()), resp))
	})
}
