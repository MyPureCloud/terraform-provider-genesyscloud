package responsemanagement_responseasset

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
	"time"
)

/*
The resource_genesyscloud_responsemanagement_responseasset.go contains all of the methods that perform the core logic for a resource.
*/

// createResponsemanagementResponseasset is used by the responsemanagement_responseasset resource to create Genesys cloud responsemanagement responseasset
func createRespManagementRespAsset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fileName := d.Get("filename").(string)
	divisionId := d.Get("division_id").(string)
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRespManagementRespAssetProxy(sdkConfig)

	sdkResponseAsset := platformclientv2.Createresponseassetrequest{}

	if fileName != "" {
		sdkResponseAsset.Name = &fileName
	}
	if divisionId != "" {
		sdkResponseAsset.DivisionId = &divisionId
	}

	log.Printf("Creating Responsemanagement response asset %s", fileName)
	postResponseData, resp, err := proxy.createRespManagementRespAsset(ctx, &sdkResponseAsset)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create response asset %s", fileName), resp)
	}

	headers := *postResponseData.Headers
	url := *postResponseData.Url
	reader, _, err := files.DownloadOrOpenFile(fileName)
	if err != nil {
		return diag.FromErr(err)
	}

	s3Uploader := files.NewS3Uploader(reader, nil, nil, headers, "PUT", url)
	_, err = s3Uploader.Upload()

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*postResponseData.Id)

	log.Printf("Created Responsemanagement response asset %s %s", fileName, *postResponseData.Id)
	return readRespManagementRespAsset(ctx, d, meta)
}

// readResponsemanagementResponseasset is used by the responsemanagement_responseasset resource to read an responsemanagement responseasset from genesys cloud
func readRespManagementRespAsset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRespManagementRespAssetProxy(sdkConfig)

	log.Printf("Reading Responsemanagement response asset %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkAsset, resp, getErr := proxy.getRespManagementRespAssetById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read response asset %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read response asset %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceResponseManagementResponseAsset())
		d.Set("filename", *sdkAsset.Name)

		if sdkAsset.Division != nil && sdkAsset.Division.Id != nil {
			d.Set("division_id", *sdkAsset.Division.Id)
		}

		log.Printf("Read Responsemanagement response asset %s %s", d.Id(), *sdkAsset.Name)

		return cc.CheckState()
	})
}

// updateResponsemanagementResponseasset is used by the responsemanagement_responseasset resource to update an responsemanagement responseasset in Genesys Cloud
func updateRespManagementRespAsset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRespManagementRespAssetProxy(sdkConfig)
	fileName := d.Get("filename").(string)
	divisionId := d.Get("division_id").(string)

	var bodyRequest platformclientv2.Responseassetrequest
	bodyRequest.Name = &fileName

	if divisionId != "" {
		bodyRequest.DivisionId = &divisionId
	}

	log.Printf("Updating Responsemanagement response asset %s", d.Id())
	putResponseData, resp, err := proxy.updateRespManagementRespAsset(ctx, d.Id(), &bodyRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update Responsemanagement response asset%s", d.Id()), resp)
	}

	// Adding a sleep with retry logic to determine when the division ID has actually been updated.
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		log.Printf("Reading response asset %s", d.Id())
		time.Sleep(20 * time.Second)
		getResponseData, resp, err := proxy.getRespManagementRespAssetById(ctx, d.Id())
		if err != nil {
			return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to read response asset %s", d.Id()), resp)
		}
		if *getResponseData.Division.Id == *putResponseData.Division.Id {
			log.Printf("Updated Responsemanagement response asset %s", d.Id())
			return readRespManagementRespAsset(ctx, d, meta)
		}
	}
	return diag.Errorf("Responsemanagement response asset %s did not update properly", d.Id())
}

// deleteResponsemanagementResponseasset is used by the responsemanagement_responseasset resource to delete an responsemanagement responseasset from Genesys cloud
func deleteRespManagementRespAsset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRespManagementRespAssetProxy(sdkConfig)
	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Responsemanagement response asset")
		resp, err := proxy.deleteRespManagementRespAsset(ctx, d.Id())
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete response asset %s", d.Id()), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	time.Sleep(20 * time.Second)
	return util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getRespManagementRespAssetById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Response asset deleted
				log.Printf("Deleted Responsemanagement response asset %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting response asset %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("Response asset %s still exists", d.Id()))
	})
}

func GenerateResponseManagementResponseAssetResource(resourceId string, fileName string, divisionId string) string {
	return fmt.Sprintf(`
resource "genesyscloud_responsemanagement_responseasset" "%s" {
    filename    = "%s"
    division_id = %s
}
`, resourceId, fileName, divisionId)
}
