package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v89/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
)

func resourceResponseManagamentResponseAsset() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud responsemanagement response asset`,

		CreateContext: createWithPooledClient(createResponsemanagementResponseAsset),
		ReadContext:   readWithPooledClient(readResponsemanagementResponseAsset),
		DeleteContext: deleteWithPooledClient(deleteResponsemanagementResponseAsset),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`filename`: {
				Description: "Name of the file to upload. Changing the name attribute will cause the existing response asset to be dropped and recreated with a new ID. It must not start with a dot and not end with a forward slash. Whitespace and the following characters are not allowed: \\{^}%`]\">[~<#|",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			`division_id`: {
				Description: `Division to associate to this asset. Can only be used with this division. Changing the division_id attribute will cause the existing response asset to be dropped and recreated with a new ID.`,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func createResponsemanagementResponseAsset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fileName := d.Get("filename").(string)
	divisionId := d.Get("division_id").(string)
	sdkConfig := meta.(*providerMeta).ClientConfig
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(sdkConfig)

	sdkResponseAsset := platformclientv2.Createresponseassetrequest{}

	if fileName != "" {
		sdkResponseAsset.Name = &fileName
	}

	if divisionId != "" {
		sdkResponseAsset.DivisionId = &divisionId
	}

	log.Printf("Creating Responsemanagement response asset %s", fileName)
	postResponseData, _, err := responseManagementApi.PostResponsemanagementResponseassetsUploads(sdkResponseAsset)
	if err != nil {
		return diag.Errorf("Failed to upload response asset %s: %v", fileName, err)
	}

	headers := *postResponseData.Headers
	url := *postResponseData.Url
	_, err = prepareAndUploadFile(fileName, nil, headers, url)
	if err != nil {
		return diag.Errorf(err.Error())
	}

	d.SetId(*postResponseData.Id)

	log.Printf("Created Responsemanagement response asset %s %s", fileName, *postResponseData.Id)
	return readResponsemanagementResponseAsset(ctx, d, meta)
}

func readResponsemanagementResponseAsset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(sdkConfig)

	log.Printf("Reading Responsemanagement response management response asset %s", d.Id())

	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		sdkAsset, resp, getErr := responseManagementApi.GetResponsemanagementResponseasset(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read response asset %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read response asset %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceResponsemanagementLibrary())

		_ = d.Set("filename", *sdkAsset.Name)

		if sdkAsset.Division != nil && sdkAsset.Division.Id != nil {
			_ = d.Set("division_id", *sdkAsset.Division.Id)
		}

		log.Printf("Read Responsemanagement response asset %s %s", d.Id(), *sdkAsset.Name)
		return cc.CheckState()
	})
}

func deleteResponsemanagementResponseAsset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(sdkConfig)

	diagErr := retryWhen(isStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Responsemanagement response asset")
		resp, err := responseManagementApi.DeleteResponsemanagementResponseasset(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete response asset: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return withRetries(ctx, 60*time.Second, func() *resource.RetryError {
		_, resp, err := responseManagementApi.GetResponsemanagementResponseasset(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// Response asset deleted
				log.Printf("Deleted Responsemanagement response asset %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting response asset %s: %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("Response asset %s still exists", d.Id()))
	})
}
