package integration_action

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

func TestUnitFlattenFunctionConfigRequestPreservesLocalFileFields(t *testing.T) {
	zipName := "uploaded.zip"
	handler := "index.handler"
	functionConfig := platformclientv2.Functionconfig{
		Function: &platformclientv2.Function{
			Handler: &handler,
		},
		Zip: &platformclientv2.Functionzipconfig{
			Name: &zipName,
		},
	}

	flattened := FlattenFunctionConfigRequest(functionConfig, "/local/path/function.zip", "sha256:abc")
	if len(flattened) != 1 {
		t.Fatalf("expected 1 function_config element, got %d", len(flattened))
	}
	m := flattened[0].(map[string]interface{})
	if m["file_path"] != "/local/path/function.zip" {
		t.Fatalf("expected preserved file_path, got %v", m["file_path"])
	}
	if m["file_content_hash"] != "sha256:abc" {
		t.Fatalf("expected preserved file_content_hash, got %v", m["file_content_hash"])
	}
	if m["handler"] != handler {
		t.Fatalf("expected handler %s, got %v", handler, m["handler"])
	}
}

func TestUnitFlattenFunctionConfigRequestFallsBackToZipName(t *testing.T) {
	zipName := "uploaded.zip"
	functionConfig := platformclientv2.Functionconfig{
		Zip: &platformclientv2.Functionzipconfig{
			Name: &zipName,
		},
	}

	flattened := FlattenFunctionConfigRequest(functionConfig, "", "")
	m := flattened[0].(map[string]interface{})
	if m["file_path"] != zipName {
		t.Fatalf("expected zip name fallback, got %v", m["file_path"])
	}
	if _, ok := m["file_content_hash"]; ok {
		t.Fatalf("did not expect file_content_hash when none provided")
	}
}

func TestUnitFunctionZipExportResolver(t *testing.T) {
	exportDir := t.TempDir()
	actionId := "action-123"
	configMap := map[string]interface{}{
		"name": "Get Ticket Status",
		"function_config": []interface{}{
			map[string]interface{}{
				"handler":           "dist/index.handler",
				"file_path":         "xp2025_get_ticket_status.zip",
				"file_content_hash": "sha256:should-be-removed",
				"zip_id":            "11111111-1111-1111-1111-111111111111",
			},
		},
		"config_request": []interface{}{
			map[string]interface{}{
				"request_type":         "POST",
				"request_url_template": "22222222-2222-2222-2222-222222222222",
			},
		},
	}
	resource := resourceExporter.ResourceInfo{
		State: &terraform.InstanceState{
			ID: actionId,
			Attributes: map[string]string{
				"function_config.0.file_path":           "xp2025_get_ticket_status.zip",
				"function_config.0.file_content_hash":   "sha256:should-be-removed",
				"function_config.0.zip_id":              "11111111-1111-1111-1111-111111111111",
				"config_request.0.request_url_template": "22222222-2222-2222-2222-222222222222",
			},
		},
	}

	if err := FunctionZipExportResolver(actionId, exportDir, FunctionZipExportSubDirectory, configMap, nil, resource); err != nil {
		t.Fatalf("FunctionZipExportResolver returned error: %v", err)
	}

	fc := configMap["function_config"].([]interface{})[0].(map[string]interface{})
	expectedPath := filepath.ToSlash(filepath.Join(FunctionZipExportSubDirectory, "xp2025_get_ticket_status.zip"))
	if fc["file_path"] != expectedPath {
		t.Fatalf("expected exported file_path %s, got %v", expectedPath, fc["file_path"])
	}
	if _, ok := fc["file_content_hash"]; ok {
		t.Fatalf("file_content_hash should be cleared on export")
	}
	if _, ok := fc["zip_id"]; ok {
		t.Fatalf("zip_id should be cleared on export")
	}
	cr := configMap["config_request"].([]interface{})[0].(map[string]interface{})
	if _, ok := cr["request_url_template"]; ok {
		t.Fatalf("request_url_template should be cleared on export")
	}
	if cr["request_type"] != "POST" {
		t.Fatalf("request_type should be preserved, got %v", cr["request_type"])
	}
	if resource.State.Attributes["function_config.0.file_path"] != expectedPath {
		t.Fatalf("state file_path not updated")
	}
	if _, ok := resource.State.Attributes["function_config.0.file_content_hash"]; ok {
		t.Fatalf("state file_content_hash should be cleared")
	}
	if _, ok := resource.State.Attributes["function_config.0.zip_id"]; ok {
		t.Fatalf("state zip_id should be cleared")
	}
	if _, ok := resource.State.Attributes["config_request.0.request_url_template"]; ok {
		t.Fatalf("state request_url_template should be cleared")
	}

	readmePath := filepath.Join(exportDir, FunctionZipExportSubDirectory, "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		t.Fatalf("expected README.md to be written: %v", err)
	}
}

func TestUnitFunctionZipExportResolverNoopWithoutFunctionConfig(t *testing.T) {
	configMap := map[string]interface{}{
		"name": "Regular Action",
	}
	resource := resourceExporter.ResourceInfo{
		State: &terraform.InstanceState{Attributes: map[string]string{}},
	}
	if err := FunctionZipExportResolver("id", t.TempDir(), FunctionZipExportSubDirectory, configMap, nil, resource); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestUnitContainsFunctionDataAction(t *testing.T) {
	cases := map[string]bool{
		"Function Data Actions":     true,
		"function-data-action":      true,
		"function_data_action":      true,
		"Genesys Cloud Data Action": false,
		"Get Ticket Status":         false,
	}
	for input, want := range cases {
		if got := containsFunctionDataAction(input); got != want {
			t.Fatalf("containsFunctionDataAction(%q)=%v, want %v", input, got, want)
		}
	}
}
