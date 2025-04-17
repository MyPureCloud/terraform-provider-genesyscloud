package architect_flow

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
func architectFlowResolver(flowId, exportDirectory, subDirectory string, configMap map[string]any, meta any, resource resourceExporter.ResourceInfo) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("caught in architectFlowResolver: %w", err)
		}
	}()

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newArchitectFlowProxy(sdkConfig)
	ctx := context.Background()
	filename := fmt.Sprintf("%s.yaml", flowId)

	flow, resp, err := proxy.GetFlow(ctx, flowId)
	if err != nil {
		log.Printf("Failed to read flow '%s'. Flow name and type will not be included in the file name. Error: %s. API Response: %s", flowId, err.Error(), resp)
	} else if flow != nil && flow.Name != nil && flow.VarType != nil {
		filename = BuildExportFileName(*flow.Name, *flow.VarType, flowId)
	}

	downloadUrl, err := proxy.generateDownloadUrl(flowId)
	if err != nil {
		return err
	}

	log.Printf("Creating subfolder '%s' inside '%s'", subDirectory, exportDirectory)
	fullPath := filepath.Join(exportDirectory, subDirectory)
	if err = os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}
	log.Printf("Successfully created subfolder '%s' inside '%s'", subDirectory, exportDirectory)

	log.Printf("Downloading export flow '%s' to '%s' from download URL", flowId, filepath.Join(fullPath, filename))
	if resp, err = files.DownloadExportFile(fullPath, filename, downloadUrl); err != nil {
		if resp != nil {
			err = fmt.Errorf("%w. API Response: %s", err, resp.String())
		}
		log.Printf("Failed to download flow file: %s", err.Error())
		return err
	}
	log.Printf("Successfully downloaded export flow '%s' to '%s'", flowId, filepath.Join(fullPath, filename))

	log.Printf("Updating resource config and state file for flow '%s'", flowId)
	updateResourceConfigAndState(configMap, resource, exportDirectory, subDirectory, filename)
	return err
}

func BuildExportFileName(flowName, flowType, flowId string) string {
	return fmt.Sprintf("%s-%s-%s.yaml", sanitizeFlowName(flowName), flowType, flowId)
}

// sanitizeFlowName will replace all forward slashes, backslashes and white spaces with an underscore
func sanitizeFlowName(s string) string {
	// First replace empty strings (multiple spaces) with a single underscore
	noSpaces := strings.ReplaceAll(s, " ", "_")

	// Replace forward slashes with underscore
	noForwardSlash := strings.ReplaceAll(noSpaces, "/", "_")

	// Replace backslashes with underscore
	result := strings.ReplaceAll(noForwardSlash, "\\", "_")

	return result
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
