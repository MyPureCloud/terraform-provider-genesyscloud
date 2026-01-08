package responsemanagement_responseasset

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/aws"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"

	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)

/*
The genesyscloud_responsemanagement_responseasset_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *responsemanagementResponseassetProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllResponseAssetsFunc func(ctx context.Context, p *responsemanagementResponseassetProxy) (*[]platformclientv2.Responseasset, *platformclientv2.APIResponse, error)
type createRespManagementRespAssetFunc func(ctx context.Context, p *responsemanagementResponseassetProxy, respAsset *platformclientv2.Createresponseassetrequest) (*platformclientv2.Createresponseassetresponse, *platformclientv2.APIResponse, error)
type updateRespManagementRespAssetFunc func(ctx context.Context, p *responsemanagementResponseassetProxy, id string, respAsset *platformclientv2.Responseassetrequest) (*platformclientv2.Responseasset, *platformclientv2.APIResponse, error)
type getRespManagementRespAssetByIdFunc func(ctx context.Context, p *responsemanagementResponseassetProxy, id string) (*platformclientv2.Responseasset, *platformclientv2.APIResponse, error)
type getRespManagementRespAssetByNameFunc func(ctx context.Context, p *responsemanagementResponseassetProxy, name string) (string, bool, *platformclientv2.APIResponse, error)
type deleteRespManagementRespAssetFunc func(ctx context.Context, p *responsemanagementResponseassetProxy, id string) (response *platformclientv2.APIResponse, err error)

// responsemanagementResponseassetProxy contains all of the methods that call genesys cloud APIs.
type responsemanagementResponseassetProxy struct {
	clientConfig                         *platformclientv2.Configuration
	responseManagementApi                *platformclientv2.ResponseManagementApi
	getAllResponseAssetsAttr             getAllResponseAssetsFunc
	createRespManagementRespAssetAttr    createRespManagementRespAssetFunc
	updateRespManagementRespAssetAttr    updateRespManagementRespAssetFunc
	getRespManagementRespAssetByIdAttr   getRespManagementRespAssetByIdFunc
	getRespManagementRespAssetByNameAttr getRespManagementRespAssetByNameFunc
	deleteRespManagementRespAssetAttr    deleteRespManagementRespAssetFunc
	assetCache                           rc.CacheInterface[platformclientv2.Responseasset]
}

// newRespManagementRespAssetProxy initializes the responsemanagement responseasset proxy with all of the data needed to communicate with Genesys Cloud
func newRespManagementRespAssetProxy(clientConfig *platformclientv2.Configuration) *responsemanagementResponseassetProxy {
	api := platformclientv2.NewResponseManagementApiWithConfig(clientConfig)
	assetCache := rc.NewResourceCache[platformclientv2.Responseasset]()
	return &responsemanagementResponseassetProxy{
		clientConfig:                         clientConfig,
		responseManagementApi:                api,
		getAllResponseAssetsAttr:             getAllResponseAssetsFn,
		createRespManagementRespAssetAttr:    createRespManagementRespAssetFn,
		updateRespManagementRespAssetAttr:    updateRespManagementRespAssetFn,
		getRespManagementRespAssetByIdAttr:   getRespManagementRespAssetByIdFn,
		getRespManagementRespAssetByNameAttr: getRespManagementRespAssetByNameFn,
		deleteRespManagementRespAssetAttr:    deleteRespManagementRespAssetFn,
		assetCache:                           assetCache,
	}
}

// getRespManagementRespAssetProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getRespManagementRespAssetProxy(clientConfig *platformclientv2.Configuration) *responsemanagementResponseassetProxy {
	if internalProxy == nil {
		internalProxy = newRespManagementRespAssetProxy(clientConfig)
	}
	return internalProxy
}

func (p *responsemanagementResponseassetProxy) getAllResponseAssets(ctx context.Context) (*[]platformclientv2.Responseasset, *platformclientv2.APIResponse, error) {
	return p.getAllResponseAssetsAttr(ctx, p)
}

// createRespManagementRespAsset creates a Genesys Cloud responsemanagement responseasset by Id
func (p *responsemanagementResponseassetProxy) createRespManagementRespAsset(ctx context.Context, respAsset *platformclientv2.Createresponseassetrequest) (*platformclientv2.Createresponseassetresponse, *platformclientv2.APIResponse, error) {
	return p.createRespManagementRespAssetAttr(ctx, p, respAsset)
}

// updateRespManagementRespAsset updates a Genesys Cloud responsemanagement responseasset by Id
func (p *responsemanagementResponseassetProxy) updateRespManagementRespAsset(ctx context.Context, id string, respAsset *platformclientv2.Responseassetrequest) (*platformclientv2.Responseasset, *platformclientv2.APIResponse, error) {
	return p.updateRespManagementRespAssetAttr(ctx, p, id, respAsset)
}

// getRespManagementRespAssetById returns a single Genesys Cloud responsemanagement responseasset by Id
func (p *responsemanagementResponseassetProxy) getRespManagementRespAssetById(ctx context.Context, id string) (*platformclientv2.Responseasset, *platformclientv2.APIResponse, error) {
	return p.getRespManagementRespAssetByIdAttr(ctx, p, id)
}
func (p *responsemanagementResponseassetProxy) getRespManagementRespAssetByName(ctx context.Context, name string) (string, bool, *platformclientv2.APIResponse, error) {
	return p.getRespManagementRespAssetByNameAttr(ctx, p, name)
}

// deleteRespManagementRespAsset deletes a Genesys Cloud responsemanagement responseasset by Id
func (p *responsemanagementResponseassetProxy) deleteRespManagementRespAsset(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteRespManagementRespAssetAttr(ctx, p, id)
}

func (p *responsemanagementResponseassetProxy) uploadRespManagementRespAsset(ctx context.Context, d *schema.ResourceData, name, filePath, divisionId string) (respBody *platformclientv2.Createresponseassetresponse, resp *platformclientv2.APIResponse, err error) {
	s3Path := ""
	localFilePath := filePath
	if aws.IsS3Path(filePath) {
		// In the case of an S3 path, the filename in the request body should be the last part of the path
		// We still want the value in the resource data to be the full S3 Path to avoid a diff
		_ = d.Set("filename", filePath)
		s3Path = filePath

		// get the file name from the end of the s3 path
		localFilePath = strings.Split(filePath, "/")[len(strings.Split(filePath, "/"))-1]

		localFilePath = filepath.Join(os.TempDir(), localFilePath)
		// Download the file if it's not present locally
		if err := downloadFileIfNotPresent(s3Path, localFilePath); err != nil {
			return nil, resp, fmt.Errorf("failed to download file from S3: %w", err)
		}
	}

	if name == "" {
		name = localFilePath
	}

	sdkResponseAsset := platformclientv2.Createresponseassetrequest{
		Name: &name,
	}
	if divisionId != "" {
		sdkResponseAsset.DivisionId = &divisionId
	}

	postResponseData, resp, err := p.createRespManagementRespAsset(ctx, &sdkResponseAsset)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to upload response asset: %s | error: %s", localFilePath, err)
	}

	headers := *postResponseData.Headers
	url := *postResponseData.Url
	reader, _, err := files.DownloadOrOpenFile(ctx, localFilePath, S3Enabled)
	if err != nil {
		return nil, resp, err
	}

	s3Uploader := files.NewS3Uploader(reader, nil, nil, headers, "PUT", url)
	_, err = s3Uploader.Upload()
	return postResponseData, resp, err
}

// downloadFileIfNotPresent will download use the DownloadOrOpenFile function to download the file and give it the path/name of filename
func downloadFileIfNotPresent(s3Path, filename string) error {
	// Check if the file already exists locally
	if _, err := os.Stat(filename); err == nil {
		// File exists - delete it and download it again
		if err := os.Remove(filename); err != nil {
			return fmt.Errorf("failed to remove existing file %s: %w", filename, err)
		}
	}

	ctx := context.Background()
	reader, file, err := files.DownloadOrOpenFile(ctx, s3Path, S3Enabled)
	if err != nil {
		return fmt.Errorf("failed to download file from %s: %w", s3Path, err)
	}

	// If we got a file handle, close it after we're done
	if file != nil {
		defer file.Close()
	}

	// Create the local file
	localFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create local file %s: %w", filename, err)
	}
	defer localFile.Close()

	// Copy the content from the reader to the local file
	_, err = io.Copy(localFile, reader)
	if err != nil {
		return fmt.Errorf("failed to write content to local file %s: %w", filename, err)
	}

	return nil
}

func getAllResponseAssetsFn(ctx context.Context, p *responsemanagementResponseassetProxy) (*[]platformclientv2.Responseasset, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging

	var allResponseAssets []platformclientv2.Responseasset
	var response *platformclientv2.APIResponse
	pageSize := 100

	responseAssets, resp, err := p.responseManagementApi.PostResponsemanagementResponseassetsSearch(platformclientv2.Responseassetsearchrequest{
		PageSize:   &pageSize,
		PageNumber: platformclientv2.Int(1),
	}, []string{})
	response = resp
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get response asset search request: %v", err)
	}

	if responseAssets.Results == nil || len(*responseAssets.Results) == 0 {
		return &allResponseAssets, resp, nil
	}
	allResponseAssets = append(allResponseAssets, *responseAssets.Results...)

	for pageNum := 2; pageNum <= *responseAssets.PageCount; pageNum++ {
		responseAssets, resp, err := p.responseManagementApi.PostResponsemanagementResponseassetsSearch(platformclientv2.Responseassetsearchrequest{
			PageSize:   &pageSize,
			PageNumber: &pageNum,
		}, []string{})
		response = resp
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get response asset search request: %v", err)
		}

		if responseAssets.Results == nil || len(*responseAssets.Results) == 0 {
			break
		}
		allResponseAssets = append(allResponseAssets, *responseAssets.Results...)
	}

	for _, asset := range allResponseAssets {
		rc.SetCache(p.assetCache, *asset.Id, asset)
	}

	return &allResponseAssets, response, nil
}

// createRespManagementRespAssetFn is an implementation of the function to create a Genesys Cloud responsemanagement responseasset
func createRespManagementRespAssetFn(ctx context.Context, p *responsemanagementResponseassetProxy, respAsset *platformclientv2.Createresponseassetrequest) (*platformclientv2.Createresponseassetresponse, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging

	postResponseData, resp, err := p.responseManagementApi.PostResponsemanagementResponseassetsUploads(*respAsset)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to upload response asset: %v", err)
	}
	return postResponseData, resp, nil
}

// updateRespManagementRespAssetFn is an implementation of the function to update a Genesys Cloud responsemanagement responseasset
func updateRespManagementRespAssetFn(ctx context.Context, p *responsemanagementResponseassetProxy, id string, respAsset *platformclientv2.Responseassetrequest) (*platformclientv2.Responseasset, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging

	return p.responseManagementApi.PutResponsemanagementResponseasset(id, *respAsset)
}

// getRespManagementRespAssetByIdFn is an implementation of the function to get a Genesys Cloud responsemanagement responseasset by Id
func getRespManagementRespAssetByIdFn(ctx context.Context, p *responsemanagementResponseassetProxy, id string) (*platformclientv2.Responseasset, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging

	asset := rc.GetCacheItem(p.assetCache, id)
	if asset != nil {
		return asset, nil, nil
	}

	sdkAsset, resp, getErr := p.responseManagementApi.GetResponsemanagementResponseasset(id)
	if getErr != nil {
		return nil, resp, fmt.Errorf("failed to retrieve response asset: %s", getErr)
	}
	return sdkAsset, resp, nil
}

func getRespManagementRespAssetByNameFn(ctx context.Context, p *responsemanagementResponseassetProxy, name string) (string, bool, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging

	var (
		field   = "name"
		fields  = []string{field}
		varType = "TERM"
		filter  = platformclientv2.Responseassetfilter{
			Fields:  &fields,
			Value:   &name,
			VarType: &varType,
		}
		body = platformclientv2.Responseassetsearchrequest{
			Query:  &[]platformclientv2.Responseassetfilter{filter},
			SortBy: &field,
		}
	)

	respAssets, resp, err := p.responseManagementApi.PostResponsemanagementResponseassetsSearch(body, nil)
	if err != nil {
		return "", false, resp, err
	}

	if respAssets == nil || len(*respAssets.Results) == 0 {
		return "", true, resp, fmt.Errorf("No responsemanagement response asset found with name %s", name)
	}

	for _, asset := range *respAssets.Results {
		if *asset.Name == name {
			log.Printf("Retrieved the responsemanagement response asset id %s by name %s", *asset.Id, name)
			return *asset.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("Unable to find responsemanagement response asset with name %s", name)
}

// deleteRespManagementRespAssetFn is an implementation function for deleting a Genesys Cloud responsemanagement responseasset
func deleteRespManagementRespAssetFn(ctx context.Context, p *responsemanagementResponseassetProxy, id string) (response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging

	resp, err := p.responseManagementApi.DeleteResponsemanagementResponseasset(id)
	if err != nil {
		return resp, fmt.Errorf("failed to delete response asset: %s", err)
	}
	rc.DeleteCacheItem(p.assetCache, id)
	return resp, nil
}
