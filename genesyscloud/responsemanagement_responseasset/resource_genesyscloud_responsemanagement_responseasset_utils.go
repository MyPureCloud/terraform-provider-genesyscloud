package responsemanagement_responseasset

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
)

func responsemanagementResponseassetResolver(responseAssetId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}, resource resourceExporter.ResourceInfo) error {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRespManagementRespAssetProxy(sdkConfig)

	fullPath := filepath.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}
	ctx := context.Background()

	data, _, err := proxy.getRespManagementRespAssetById(ctx, responseAssetId)
	if err != nil {
		return err
	}

	baseName := strings.TrimSuffix(filepath.Base(*data.Name), filepath.Ext(*data.Name))
	fileName := fmt.Sprintf("%s-%s%s", baseName, responseAssetId, filepath.Ext(*data.Name))
	exportFilename := filepath.Join(subDirectory, fileName)

	if err := files.DownloadExportFile(fullPath, fileName, *data.ContentLocation); err != nil {
		return err
	}
	configMap["filename"] = exportFilename
	resource.State.Attributes["filename"] = exportFilename

	fileContentVal := fmt.Sprintf(`${filesha256("%s")}`, exportFilename)
	configMap["file_content_hash"] = fileContentVal

	hash, er := files.HashFileContent(path.Join(fullPath, fileName))
	if er != nil {
		log.Printf("Error Calculating Hash '%s' ", er)
	} else {
		resource.State.Attributes["file_content_hash"] = hash
	}
	return err
}

func GenerateResponseManagementResponseAssetResource(resourceLabel string, fileName string, divisionId string) string {
	return fmt.Sprintf(`
resource "genesyscloud_responsemanagement_responseasset" "%s" {
    filename    = "%s"
    division_id = %s
	file_content_hash = filesha256("%s")
}
`, resourceLabel, fileName, divisionId, fileName)
}
