package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
	resource_exporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

func ResourceResponsemanagementLibrary() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud responsemanagement library`,

		CreateContext: CreateWithPooledClient(createResponsemanagementLibrary),
		ReadContext:   ReadWithPooledClient(readResponsemanagementLibrary),
		UpdateContext: UpdateWithPooledClient(updateResponsemanagementLibrary),
		DeleteContext: DeleteWithPooledClient(deleteResponsemanagementLibrary),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The library name.`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func getAllResponsemanagementLibrary(_ context.Context, clientConfig *platformclientv2.Configuration) (resource_exporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resource_exporter.ResourceIDMetaMap)
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		sdklibraryentitylisting, _, getErr := responseManagementApi.GetResponsemanagementLibraries(pageNum, pageSize, "", "")
		if getErr != nil {
			return nil, diag.Errorf("Error requesting page of Responsemanagement Library: %s", getErr)
		}

		if sdklibraryentitylisting.Entities == nil || len(*sdklibraryentitylisting.Entities) == 0 {
			break
		}

		for _, entity := range *sdklibraryentitylisting.Entities {
			resources[*entity.Id] = &resource_exporter.ResourceMeta{Name: *entity.Name}
		}
	}

	return resources, nil
}

func ResponsemanagementLibraryExporter() *resource_exporter.ResourceExporter {
	return &resource_exporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllResponsemanagementLibrary),
	}
}

func createResponsemanagementLibrary(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(sdkConfig)

	sdklibrary := platformclientv2.Library{}

	if name != "" {
		sdklibrary.Name = &name
	}

	log.Printf("Creating Responsemanagement Library %s", name)
	responsemanagementLibrary, _, err := responseManagementApi.PostResponsemanagementLibraries(sdklibrary)
	if err != nil {
		return diag.Errorf("Failed to create Responsemanagement Library %s: %s", name, err)
	}

	d.SetId(*responsemanagementLibrary.Id)

	log.Printf("Created Responsemanagement Library %s %s", name, *responsemanagementLibrary.Id)
	return readResponsemanagementLibrary(ctx, d, meta)
}

func updateResponsemanagementLibrary(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(sdkConfig)

	sdklibrary := platformclientv2.Library{}

	if name != "" {
		sdklibrary.Name = &name
	}

	log.Printf("Updating Responsemanagement Library %s", name)
	diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Responsemanagement Library version
		responsemanagementLibrary, resp, getErr := responseManagementApi.GetResponsemanagementLibrary(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Responsemanagement Library %s: %s", d.Id(), getErr)
		}
		sdklibrary.Version = responsemanagementLibrary.Version
		responsemanagementLibrary, _, updateErr := responseManagementApi.PutResponsemanagementLibrary(d.Id(), sdklibrary)
		if updateErr != nil {
			return resp, diag.Errorf("Failed to update Responsemanagement Library %s: %s", name, updateErr)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Responsemanagement Library %s", name)
	return readResponsemanagementLibrary(ctx, d, meta)
}

func readResponsemanagementLibrary(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(sdkConfig)

	log.Printf("Reading Responsemanagement Library %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *resource.RetryError {
		sdklibrary, resp, getErr := responseManagementApi.GetResponsemanagementLibrary(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read Responsemanagement Library %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read Responsemanagement Library %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceResponsemanagementLibrary())

		if sdklibrary.Name != nil {
			d.Set("name", *sdklibrary.Name)
		}

		log.Printf("Read Responsemanagement Library %s %s", d.Id(), *sdklibrary.Name)
		return cc.CheckState()
	})
}

func deleteResponsemanagementLibrary(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(sdkConfig)

	diagErr := RetryWhen(IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Responsemanagement Library")
		resp, err := responseManagementApi.DeleteResponsemanagementLibrary(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Responsemanagement Library: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return WithRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := responseManagementApi.GetResponsemanagementLibrary(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Responsemanagement Library deleted
				log.Printf("Deleted Responsemanagement Library %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting Responsemanagement Library %s: %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("Responsemanagement Library %s still exists", d.Id()))
	})
}
