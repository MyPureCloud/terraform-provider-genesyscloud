package responsemanagement_responseasset

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_responsemanagement_responseasset.go contains all the methods that perform the core logic for a resource.
*/
func getAllResponseAssets(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getRespManagementRespAssetProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	assets, resp, err := proxy.getAllResponseAssets(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get response management response assets | Error: %s", err), resp)
	}

	for _, asset := range *assets {
		resources[*asset.Id] = &resourceExporter.ResourceMeta{BlockLabel: *asset.Name}
	}

	return resources, nil
}

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
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to upload response asset: %s | error: %s", fileName, err), resp)
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
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceResponseManagementResponseAsset(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Responsemanagement response asset %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkAsset, resp, getErr := proxy.getRespManagementRespAssetById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read response asset %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read response asset %s | error: %s", d.Id(), getErr), resp))
		}

		_ = d.Set("filename", *sdkAsset.Name)

		if sdkAsset.Division != nil && sdkAsset.Division.Id != nil {
			_ = d.Set("division_id", *sdkAsset.Division.Id)
		}

		log.Printf("Read Responsemanagement response asset %s %s", d.Id(), *sdkAsset.Name)

		return cc.CheckState(d)
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
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to update response asset: %s | error: %s", d.Id(), err), resp)
	}

	// Adding a sleep with retry logic to determine when the division ID has actually been updated.
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		log.Printf("Reading response asset %s", d.Id())
		time.Sleep(20 * time.Second)
		getResponseData, resp, err := proxy.getRespManagementRespAssetById(ctx, d.Id())
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to read response asset: %s | error: %s", d.Id(), err), resp)
		}
		if *getResponseData.Division.Id == *putResponseData.Division.Id {
			log.Printf("Updated Responsemanagement response asset %s", d.Id())
			return readRespManagementRespAsset(ctx, d, meta)
		}
	}
	return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Responsemanagement response asset %s did not update properly | error: %s", d.Id(), err), resp)
}

// deleteResponsemanagementResponseasset is used by the responsemanagement_responseasset resource to delete an responsemanagement responseasset from Genesys cloud
func deleteRespManagementRespAsset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRespManagementRespAssetProxy(sdkConfig)

	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Responsemanagement response asset")
		resp, err := proxy.deleteRespManagementRespAsset(ctx, d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to delete response asset: %s | error: %s", d.Id(), err), resp)
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
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting response asset %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("response asset %s still exists", d.Id()), resp))
	})
}
