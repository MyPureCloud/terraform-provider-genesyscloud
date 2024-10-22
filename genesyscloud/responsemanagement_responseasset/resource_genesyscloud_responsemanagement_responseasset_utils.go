package responsemanagement_responseasset

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"
)

func responsemanagementResponseassetResolver(responseAssetId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}, resource resourceExporter.ResourceInfo) error {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRespManagementRespAssetProxy(sdkConfig)

	fullPath := path.Join(exportDirectory, subDirectory)
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
	exportFilename := path.Join(subDirectory, fileName)

	if err := files.DownloadExportFile(fullPath, fileName, *data.ContentLocation); err != nil {
		return err
	}
	configMap["filename"] = exportFilename
	resource.State.Attributes["filepath"] = exportFilename

	fileContentVal := fmt.Sprintf(`${filesha256("%s")}`, exportFilename)
	configMap["file_content_hash"] = fmt.Sprintf(`${filesha256("%s")}`, exportFilename)
	resource.State.Attributes["file_content_hash"] = fileContentVal

	return nil
}

func GenerateResponseManagementResponseAssetResource(resourceId string, fileName string, divisionId string) string {
	fullyQualifiedPath, _ := testrunner.NormalizePath(fileName)

	return fmt.Sprintf(`
resource "genesyscloud_responsemanagement_responseasset" "%s" {
    filename    = "%s"
    division_id = %s
	file_content_hash = filesha256("%s")
}
`, resourceId, fileName, divisionId, fullyQualifiedPath)
}
