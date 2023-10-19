package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	files "terraform-provider-genesyscloud/genesyscloud/util/files"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func resourceResponseManagamentResponseAsset() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud responsemanagement response asset`,

		CreateContext: CreateWithPooledClient(createResponsemanagementResponseAsset),
		ReadContext:   ReadWithPooledClient(readResponsemanagementResponseAsset),
		UpdateContext: UpdateWithPooledClient(updateResponsemanagementResponseAsset),
		DeleteContext: DeleteWithPooledClient(deleteResponsemanagementResponseAsset),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`filename`: {
				Description:      "Name of the file to upload. Changing the name attribute will cause the existing response asset to be dropped and recreated with a new ID. It must not start with a dot and not end with a forward slash. Whitespace and the following characters are not allowed: \\{^}%`]\">[~<#|",
				Required:         true,
				ForceNew:         true,
				Type:             schema.TypeString,
				ValidateDiagFunc: validateResponseAssetName,
			},
			`division_id`: {
				Description: `Division to associate to this asset. Can only be used with this division.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func createResponsemanagementResponseAsset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fileName := d.Get("filename").(string)
	divisionId := d.Get("division_id").(string)
	sdkConfig := meta.(*ProviderMeta).ClientConfig
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
	reader, _, err := files.DownloadOrOpenFile(fileName)
	if err != nil {
		return diag.Errorf(err.Error())
	}

	s3Uploader := files.NewS3Uploader(reader, nil, nil, headers, "PUT", url)
	_, err = s3Uploader.Upload()
	if err != nil {
		return diag.Errorf(err.Error())
	}

	d.SetId(*postResponseData.Id)

	log.Printf("Created Responsemanagement response asset %s %s", fileName, *postResponseData.Id)
	return readResponsemanagementResponseAsset(ctx, d, meta)
}

func readResponsemanagementResponseAsset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(sdkConfig)

	log.Printf("Reading Responsemanagement response asset %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkAsset, resp, getErr := responseManagementApi.GetResponsemanagementResponseasset(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read response asset %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read response asset %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceResponsemanagementLibrary())

		_ = d.Set("filename", *sdkAsset.Name)

		if sdkAsset.Division != nil && sdkAsset.Division.Id != nil {
			_ = d.Set("division_id", *sdkAsset.Division.Id)
		}

		log.Printf("Read Responsemanagement response asset %s %s", d.Id(), *sdkAsset.Name)
		return cc.CheckState()
	})
}

func updateResponsemanagementResponseAsset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(sdkConfig)

	fileName := d.Get("filename").(string)
	divisionId := d.Get("division_id").(string)

	var bodyRequest platformclientv2.Responseassetrequest
	bodyRequest.Name = &fileName

	if divisionId != "" {
		bodyRequest.DivisionId = &divisionId
	}

	log.Printf("Updating Responsemanagement response asset %s", d.Id())
	putResponseData, _, err := responseManagementApi.PutResponsemanagementResponseasset(d.Id(), bodyRequest)
	if err != nil {
		diag.Errorf("Failed to delete Responsemanagement response asset %s: %v", d.Id(), err)
	}

	// Adding a sleep with retry logic to determine when the division ID has actually been updated.
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		log.Printf("Reading response asset %s", d.Id())
		time.Sleep(20 * time.Second)
		getResponseData, _, err := responseManagementApi.GetResponsemanagementResponseasset(d.Id())
		if err != nil {
			return diag.Errorf("Failed to read response asset %s: %v", d.Id(), err)
		}
		if *getResponseData.Division.Id == *putResponseData.Division.Id {
			log.Printf("Updated Responsemanagement response asset %s", d.Id())
			return readResponsemanagementResponseAsset(ctx, d, meta)
		}
	}

	return diag.Errorf("Responsemanagement response asset %s did not update properly", d.Id())
}

func deleteResponsemanagementResponseAsset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(sdkConfig)

	diagErr := RetryWhen(IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
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
	time.Sleep(20 * time.Second)
	return WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		_, resp, err := responseManagementApi.GetResponsemanagementResponseasset(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Response asset deleted
				log.Printf("Deleted Responsemanagement response asset %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting response asset %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("Response asset %s still exists", d.Id()))
	})
}
