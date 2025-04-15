package scripts

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
	"log"
	"os"
	"path"
	"path/filepath"
)

// ScriptResolver is used to download all Genesys Cloud scripts from Genesys Cloud
func ScriptResolver(scriptId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}, resource resourceExporter.ResourceInfo) error {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	scriptsProxy := getScriptsProxy(sdkConfig)

	exportFileName := fmt.Sprintf("script-%s.json", scriptId)

	fullPath := filepath.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}
	ctx := context.Background()
	url, _, err := scriptsProxy.getScriptExportUrl(ctx, scriptId)
	if err != nil {
		return err
	}

	if _, err := files.DownloadExportFile(fullPath, exportFileName, url); err != nil {
		return err
	}

	// Update filepath field in configMap to point to exported script file
	fileNameVal := filepath.Join(subDirectory, exportFileName)
	fileContentVal := fmt.Sprintf(`${filesha256("%s")}`, filepath.Join(subDirectory, exportFileName))
	configMap["filepath"] = fileNameVal
	configMap["file_content_hash"] = fileContentVal

	resource.State.Attributes["filepath"] = fileNameVal

	hash, err := files.HashFileContent(path.Join(fullPath, exportFileName))
	if err != nil {
		log.Printf("Error Calculating Hash '%s' ", err)
	} else {
		resource.State.Attributes["file_content_hash"] = hash
	}
	return err
}

func GenerateScriptResourceBasic(resourceLabel, scriptName, filePath string) string {
	return fmt.Sprintf(`
		resource "%s" "%s" {
			script_name       = "%s"
			filepath          = "%s"
			file_content_hash = filesha256("%s")
		}
	`, ResourceType, resourceLabel, scriptName, filePath, filePath)
}
