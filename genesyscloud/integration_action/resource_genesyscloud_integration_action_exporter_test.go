package integration_action

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
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

func TestUnitIntegrationActionExporterTreatsFunctionZipAsUnresolvable(t *testing.T) {
	exporter := IntegrationActionExporter()
	if exporter.CustomFileWriter.RetrieveAndWriteFilesFunc != nil {
		t.Fatal("function zip custom file writer should not be registered; zip binaries cannot be downloaded")
	}

	attr, ok := exporter.UnResolvableAttributes["function_config.file_path"]
	if !ok {
		t.Fatal("expected function_config.file_path in UnResolvableAttributes")
	}
	if attr == nil || attr.Type != schema.TypeString {
		t.Fatalf("expected string schema for function_config.file_path, got %#v", attr)
	}

	hashResolver, ok := exporter.CustomAttributeResolver["function_config.file_content_hash"]
	if !ok || hashResolver == nil || hashResolver.ResolverFunc == nil {
		t.Fatal("expected custom resolver to strip function_config.file_content_hash")
	}

	configMap := map[string]interface{}{
		"file_path":         "fda_code.zip",
		"file_content_hash": "sha256:stale",
	}
	if err := hashResolver.ResolverFunc(configMap, nil, "label"); err != nil {
		t.Fatalf("hash resolver returned error: %v", err)
	}
	if _, exists := configMap["file_content_hash"]; exists {
		t.Fatal("file_content_hash should be stripped on export")
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
