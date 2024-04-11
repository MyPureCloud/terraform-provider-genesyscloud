package responsemanagement_responseasset

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
)

func responsemanagementResponseassetResolver(responseAssetId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}) error {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRespManagementRespAssetProxy(sdkConfig)

	fileName := fmt.Sprintf("asset-%s.jpeg", responseAssetId)

	fullPath := path.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}

	ctx := context.Background()
	data, _, err := proxy.getRespManagementRespAssetById(ctx, responseAssetId)
	if err != nil {
		return err
	}

	log.Println("Details: ", fullPath, fileName, subDirectory)
	if err := files.DownloadExportFile(fullPath, fileName, *data.ContentLocation); err != nil {
		return err
	}

	configMap["filepath"] = path.Join(subDirectory, fileName)

	return nil
}

func GenerateResponseManagementResponseAssetResource(resourceId string, fileName string, divisionId string, filepath string) string {
	return fmt.Sprintf(`
resource "genesyscloud_responsemanagement_responseasset" "%s" {
    filename    = "%s"
    division_id = %s
	filepath 	= "%s"
}
`, resourceId, fileName, divisionId, filepath)
}
