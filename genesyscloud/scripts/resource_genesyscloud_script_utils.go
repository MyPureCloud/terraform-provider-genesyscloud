package scripts

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
)

// ScriptResolver is used to download all Genesys Cloud scripts from Genesys Cloud
func ScriptResolver(scriptId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}, resource resourceExporter.ResourceInfo) error {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	scriptsProxy := getScriptsProxy(sdkConfig)

	exportFileName := fmt.Sprintf("script-%s.json", scriptId)

	fullPath := path.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}
	ctx := context.Background()
	url, _, err := scriptsProxy.getScriptExportUrl(ctx, scriptId)
	if err != nil {
		return err
	}

	if err := files.DownloadExportFile(fullPath, exportFileName, url); err != nil {
		return err
	}

	// Update filepath field in configMap to point to exported script file
	fileNameVal := path.Join(subDirectory, exportFileName)
	fileContentVal := fmt.Sprintf(`${filesha256("%s")}`, path.Join(subDirectory, exportFileName))
	configMap["filepath"] = fileNameVal
	configMap["file_content_hash"] = fileContentVal

	resource.State.Attributes["filepath"] = fileNameVal

	hash, er := files.HashFileContent(path.Join(fullPath, exportFileName))
	if er != nil {
		log.Printf("Error Calculating Hash '%s' ", er)
	} else {
		resource.State.Attributes["file_content_hash"] = hash
	}
	return err
}
