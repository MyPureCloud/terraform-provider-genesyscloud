package architect_flow

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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

// architectFlowResolver downloads and processes an architect flow from Genesys Cloud.
// It creates a subdirectory, downloads the flow file, and updates the resource configuration.
//
// Parameters:
//   - flowId: The ID of the architect flow to resolve
//   - exportDirectory: The base directory where the flow will be exported
//   - subDirectory: The subdirectory name to create within the export directory
//   - configMap: Configuration map containing resource settings
//   - meta: Provider metadata containing client configuration
//   - resource: Resource information for the architect flow
//
// Returns:
//   - error: Returns an error if any operation fails, nil otherwise
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
	fullPath := filepath.Join(exportDirectory, subDirectory)
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

// setFileContentHashToNil resets the file_content_hash in the resource data to nil.
// This is necessary after a flow update fails to ensure Terraform detects changes
// in the file content hash and attempts an update on the next terraform apply,
// even if the file contents haven't changed.
//
// Parameters:
//   - d: The schema.ResourceData containing the resource state
func setFileContentHashToNil(d *schema.ResourceData) {
	_ = d.Set("file_content_hash", nil)
}

// updateResourceConfigAndState updates filepath and file_content_hash in resource and state file to point to exported file
func updateResourceConfigAndState(configMap map[string]any, resource resourceExporter.ResourceInfo, exportDir, subDir, filename string) {
	var (
		exportFilePath                       = filepath.Join(subDir, filename)
		exportFilePathIncludingExportDirName = filepath.Join(exportDir, subDir, filename)
	)

	configMap["filepath"] = exportFilePath
	configMap["file_content_hash"] = fmt.Sprintf(`${filesha256("%s")}`, exportFilePath)

	resource.State.Attributes["filepath"] = exportFilePath
	// Update file_content_hash in exported state file with actual hash
	hash, err := files.HashFileContent(exportFilePathIncludingExportDirName)
	if err != nil {
		log.Printf("Error Calculating Hash '%s' ", err)
	} else {
		resource.State.Attributes["file_content_hash"] = hash
	}
}
