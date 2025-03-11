package architect_flow

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/files"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func isForceUnlockEnabled(d *schema.ResourceData) bool {
	forceUnlock := d.Get("force_unlock").(bool)
	log.Printf("ForceUnlock: %v, id %v", forceUnlock, d.Id())

	if forceUnlock && d.Id() != "" {
		return true
	}
	return false
}

func GenerateFlowResource(resourceLabel, srcFile, fileContent string, forceUnlock bool, substitutions ...string) string {
	if fileContent != "" {
		updateFile(srcFile, fileContent)
	}

	flowResourceStr := fmt.Sprintf(`resource "genesyscloud_flow" "%s" {
        filepath = %s
		file_content_hash =  filesha256(%s)
		force_unlock = %v
		%s
	}
	`, resourceLabel, strconv.Quote(srcFile), strconv.Quote(srcFile), forceUnlock, strings.Join(substitutions, "\n"))

	return flowResourceStr
}

func architectFlowResolver(flowId, exportDirectory, subDirectory string, configMap map[string]any, meta any, resource resourceExporter.ResourceInfo) error {
	var (
		sdkConfig = meta.(*provider.ProviderMeta).ClientConfig
		proxy     = newArchitectFlowProxy(sdkConfig)
	)

	downloadUrl, err := proxy.generateDownloadUrl(flowId)
	if err != nil {
		return err
	}

	log.Printf("Creating subfolder '%s' inside '%s'", subDirectory, exportDirectory)
	fullPath := path.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}
	log.Printf("Successfully created subfolder '%s' inside '%s'", subDirectory, exportDirectory)

	filename := fmt.Sprintf("flow-%s.yml", flowId)
	log.Printf("Downloading export flow '%s' to '%s' from download URL", flowId, fullPath)
	if resp, err := files.DownloadExportFile(fullPath, filename, downloadUrl); err != nil {
		log.Printf("Failed to download flow file: %s", err.Error())
		if resp != nil {
			log.Printf("API Response: " + resp.String())
		}
		return err
	}
	log.Printf("Successfully downloaded export flow '%s' to '%s'", flowId, fullPath)

	log.Printf("Updating resource config and state file for flow '%s'", flowId)
	updateResourceConfigAndState(configMap, resource, exportDirectory, subDirectory, filename)
	return nil
}

// setFileContentHashToNil This operation is required after a flow update fails because we want Terraform to detect changes
// in the file content hash and re-attempt an update, should the user re-run terraform apply without making changes to the file contents
func setFileContentHashToNil(d *schema.ResourceData) {
	_ = d.Set("file_content_hash", nil)
}

// updateResourceConfigAndState updates filepath and file_content_hash in resource and state file to point to exported file
func updateResourceConfigAndState(configMap map[string]any, resource resourceExporter.ResourceInfo, exportDir, subDir, filename string) {
	configMap["filepath"] = path.Join(subDir, filename)
	configMap["file_content_hash"] = fmt.Sprintf(`${filesha256("%s")}`, path.Join(subDir, filename))

	resource.State.Attributes["filepath"] = path.Join(subDir, filename)
	// Update file_content_hash in exported state file with actual hash
	hash, err := files.HashFileContent(path.Join(exportDir, subDir, filename))
	if err != nil {
		log.Printf("Error Calculating Hash '%s' ", err)
	} else {
		resource.State.Attributes["file_content_hash"] = hash
	}
}
