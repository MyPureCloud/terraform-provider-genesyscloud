package integration_action

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

var functionZipFileNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

// FunctionZipExportResolver updates exported function_config paths for function data actions.
//
// Genesys Cloud does not provide an API (or UI) to download function zip binaries.
// See: https://help.genesys.cloud/articles/add-function-configuration/
// This resolver therefore cannot retrieve the zip. It rewrites file_path to a conventional
// relative location under function_zips/, clears file_content_hash, and writes a README so
// pipelines know they must supply the zip before apply.
func FunctionZipExportResolver(actionId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}, resource resourceExporter.ResourceInfo) error {
	_ = meta

	functionConfigRaw, ok := configMap["function_config"]
	if !ok || functionConfigRaw == nil {
		return nil
	}

	functionConfigList, ok := functionConfigRaw.([]interface{})
	if !ok || len(functionConfigList) == 0 || functionConfigList[0] == nil {
		return nil
	}

	functionMap, ok := functionConfigList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	fullPath := filepath.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}

	if err := writeFunctionZipExportReadme(fullPath); err != nil {
		return err
	}

	zipFileName := buildExportedFunctionZipFileName(functionMap, configMap, actionId)
	exportFilePath := filepath.Join(subDirectory, zipFileName)
	normalizedPath := strings.ReplaceAll(exportFilePath, "\\", "/")

	log.Printf("WARNING: Function data action zip for %s cannot be downloaded from Genesys Cloud. "+
		"Exported file_path set to %s — place the zip at that path (or update file_path) before apply.",
		actionId, normalizedPath)

	functionMap["file_path"] = normalizedPath
	delete(functionMap, "file_content_hash")

	// zip_id is a Genesys Cloud generated, org-specific GUID (Optional+Computed).
	// Exporting it produces a hardcoded value that is invalid in a target org and
	// causes perpetual plan drift, since the platform regenerates it on publish.
	// Drop it so the platform recomputes it on apply.
	delete(functionMap, "zip_id")

	configMap["function_config"] = []interface{}{functionMap}

	// For function data actions, request_url_template is set server-side to the
	// generated function id (an org-specific GUID). Like zip_id, exporting it makes
	// the config non-portable and causes perpetual drift. Remove it so the platform
	// recomputes it on apply.
	removeExportedFunctionRequestUrlTemplate(configMap)

	if resource.State != nil && resource.State.Attributes != nil {
		resource.State.Attributes["function_config.0.file_path"] = normalizedPath
		delete(resource.State.Attributes, "function_config.0.file_content_hash")
		delete(resource.State.Attributes, "function_config.0.zip_id")
		delete(resource.State.Attributes, "config_request.0.request_url_template")
	}

	return nil
}

// removeExportedFunctionRequestUrlTemplate deletes request_url_template from the
// first config_request block in the exported config map, if present.
func removeExportedFunctionRequestUrlTemplate(configMap map[string]interface{}) {
	configRequestRaw, ok := configMap["config_request"]
	if !ok || configRequestRaw == nil {
		return
	}
	configRequestList, ok := configRequestRaw.([]interface{})
	if !ok || len(configRequestList) == 0 || configRequestList[0] == nil {
		return
	}
	requestMap, ok := configRequestList[0].(map[string]interface{})
	if !ok {
		return
	}
	delete(requestMap, "request_url_template")
	configRequestList[0] = requestMap
	configMap["config_request"] = configRequestList
}

func writeFunctionZipExportReadme(directory string) error {
	readmePath := filepath.Join(directory, "README.md")
	if _, err := os.Stat(readmePath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	const contents = `# Function data action zip files

Genesys Cloud does not allow downloading function ZIP files (API or UI).
See: https://help.genesys.cloud/articles/add-function-configuration/

Exported function actions reference zip paths in this directory, but the binaries
are **not** included. Before applying exported Terraform:

1. Copy each function zip into this directory using the exported file name, or
2. Update each resource's function_config.file_path to point at your zip, and
3. Optionally set function_config.file_content_hash = filesha256("<path>").
`
	return os.WriteFile(readmePath, []byte(contents), 0o644)
}

func buildExportedFunctionZipFileName(functionMap, configMap map[string]interface{}, actionId string) string {
	if pathVal, ok := functionMap["file_path"].(string); ok && strings.TrimSpace(pathVal) != "" {
		base := filepath.Base(strings.TrimSpace(pathVal))
		if base != "" && base != "." && base != string(filepath.Separator) {
			if !strings.HasSuffix(strings.ToLower(base), ".zip") {
				base = base + ".zip"
			}
			return functionZipFileNameSanitizer.ReplaceAllString(base, "_")
		}
	}

	baseName := "function"
	if name, ok := configMap["name"].(string); ok && strings.TrimSpace(name) != "" {
		baseName = strings.TrimSpace(name)
	}
	baseName = functionZipFileNameSanitizer.ReplaceAllString(baseName, "_")
	if len(baseName) > 64 {
		baseName = baseName[:64]
	}
	return fmt.Sprintf("%s-%s.zip", baseName, actionId)
}
