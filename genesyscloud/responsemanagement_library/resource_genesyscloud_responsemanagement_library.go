package responsemanagement_library

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v123/platformclientv2"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

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

	librarys, err := proxy.getAllResponsemanagementLibrary(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get responsemanagement library: %v", err)
	}

	for _, library := range *librarys {
		resources[*library.Id] = &resourceExporter.ResourceMeta{Name: *library.Name}
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
	library, err := proxy.createResponsemanagementLibrary(ctx, &responsemanagementLibrary)
	if err != nil {
		return diag.Errorf("Failed to create responsemanagement library: %s", err)
	}

	d.SetId(*library.Id)
	log.Printf("Created responsemanagement library %s", *library.Id)
	return readResponsemanagementLibrary(ctx, d, meta)
}

// readResponsemanagementLibrary is used by the responsemanagement_library resource to read an responsemanagement library from genesys cloud
func readResponsemanagementLibrary(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getResponsemanagementLibraryProxy(sdkConfig)

	log.Printf("Reading responsemanagement library %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		library, respCode, getErr := proxy.getResponsemanagementLibraryById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read responsemanagement library %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read responsemanagement library %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceResponsemanagementLibrary())

		resourcedata.SetNillableValue(d, "name", library.Name)

		log.Printf("Read responsemanagement library %s %s", d.Id(), *library.Name)
		return cc.CheckState()
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
	library, err := proxy.updateResponsemanagementLibrary(ctx, d.Id(), &responsemanagementLibrary)
	if err != nil {
		return diag.Errorf("Failed to update responsemanagement library: %s", err)
	}
	log.Printf("Updated responsemanagement library %s", *library.Id)
	return readResponsemanagementLibrary(ctx, d, meta)
}

// deleteResponsemanagementLibrary is used by the responsemanagement_library resource to delete an responsemanagement library from Genesys cloud
func deleteResponsemanagementLibrary(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getResponsemanagementLibraryProxy(sdkConfig)

	_, err := proxy.deleteResponsemanagementLibrary(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete responsemanagement library %s: %s", d.Id(), err)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getResponsemanagementLibraryById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404ByInt(respCode) {
				log.Printf("Deleted responsemanagement library %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting responsemanagement library %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("responsemanagement library %s still exists", d.Id()))
	})
}
