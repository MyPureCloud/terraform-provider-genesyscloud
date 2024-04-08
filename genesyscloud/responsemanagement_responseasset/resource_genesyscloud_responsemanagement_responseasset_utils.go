package responsemanagement_responseasset

import (
	"context"
	"fmt"
	"os"
	"path"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
)

func responsemanagementResponseassetResolver(responseAssetId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}) error {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRespManagementRespAssetProxy(sdkConfig)

	exportFileName := fmt.Sprintf("responseasset-%s.json", responseAssetId)

	fullPath := path.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}

	ctx := context.Background()
	url, _, err := proxy.getRespAssetExportUrl(ctx, responseAssetId)
	if err != nil {
		return err
	}

	if err := files.DownloadExportFile(fullPath, exportFileName, url); err != nil {
		return err
	}

	configMap["filepath"] = path.Join(subDirectory, exportFileName)
	configMap["file_content_hash"] = fmt.Sprintf(`${filesha256("%s")}`, path.Join(subDirectory, exportFileName))

	return err
}
