package scripts

import (
	"context"
	"fmt"
	"os"
	"path"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
)

// ScriptResolver is used to download all Genesys Cloud scripts from Genesys Cloud
func ScriptResolver(scriptId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}) error {
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
	configMap["filepath"] = path.Join(subDirectory, exportFileName)
	configMap["file_content_hash"] = fmt.Sprintf(`${filesha256("%s")}`, path.Join(subDirectory, exportFileName))

	return err
}
