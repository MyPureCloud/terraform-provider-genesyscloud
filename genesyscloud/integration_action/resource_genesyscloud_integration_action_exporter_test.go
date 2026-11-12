package integration_action

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
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

// TestUnitIntegrationActionExporterStripsFunctionZipID verifies the DEVTOOLING-1797 fix:
// zip_id is a Genesys Cloud generated, org-specific GUID and must be dropped on export so
// it does not cause perpetual plan drift in a target org.
func TestUnitIntegrationActionExporterStripsFunctionZipID(t *testing.T) {
	exporter := IntegrationActionExporter()

	zipIDResolver, ok := exporter.CustomAttributeResolver["function_config.zip_id"]
	if !ok || zipIDResolver == nil || zipIDResolver.ResolverFunc == nil {
		t.Fatal("expected custom resolver to strip function_config.zip_id")
	}

	configMap := map[string]interface{}{
		"handler": "dist/index.handler",
		"zip_id":  "11111111-1111-1111-1111-111111111111",
	}
	if err := zipIDResolver.ResolverFunc(configMap, nil, "label"); err != nil {
		t.Fatalf("zip_id resolver returned error: %v", err)
	}
	if _, exists := configMap["zip_id"]; exists {
		t.Fatal("zip_id should be stripped on export")
	}
	if configMap["handler"] != "dist/index.handler" {
		t.Fatalf("handler should be preserved, got %v", configMap["handler"])
	}
}

// requestURLTemplateFromConfigMap is a test helper that reads
// config_request.0.request_url_template out of a top-level action config map, returning the
// value and whether the key is present.
func requestURLTemplateFromConfigMap(t *testing.T, configMap map[string]interface{}) (string, bool) {
	t.Helper()
	configRequestList, ok := configMap["config_request"].([]interface{})
	if !ok || len(configRequestList) == 0 {
		t.Fatalf("expected a config_request block, got %#v", configMap["config_request"])
	}
	configRequest, ok := configRequestList[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected config_request[0] to be a map, got %#v", configRequestList[0])
	}
	val, exists := configRequest["request_url_template"]
	if !exists {
		return "", false
	}
	strVal, _ := val.(string)
	return strVal, true
}

// newActionConfigMap builds a minimal top-level integration action config map for exporter
// resolver tests. functionConfigSet controls whether a non-empty function_config block is
// present (the signal that distinguishes a function data action from an HTTP data action).
func newActionConfigMap(category, urlTemplate string, functionConfigSet bool) map[string]interface{} {
	configMap := map[string]interface{}{
		"category": category,
		"config_request": []interface{}{
			map[string]interface{}{
				"request_type":         "POST",
				"request_url_template": urlTemplate,
			},
		},
	}
	if functionConfigSet {
		configMap["function_config"] = []interface{}{
			map[string]interface{}{
				"handler": "dist/index.handler",
			},
		}
	}
	return configMap
}

// TestUnitIntegrationActionExporterStripsRequestURLTemplate verifies that
// config_request.request_url_template is dropped on export ONLY for function data actions
// (where the platform sets it to the server-generated function id and keeping it causes
// perpetual drift — DEVTOOLING-1797), and is PRESERVED for normal HTTP data actions (where it
// is required config — GitHub issue #2626).
func TestUnitIntegrationActionExporterStripsRequestURLTemplate(t *testing.T) {
	exporter := IntegrationActionExporter()

	resolver, ok := exporter.CustomAttributeResolver["category"]
	if !ok || resolver == nil || resolver.ResolverFunc == nil {
		t.Fatal("expected custom resolver on \"category\" to strip request_url_template for function actions")
	}

	// There must no longer be a blanket resolver on config_request.request_url_template, which
	// stripped the field for every action and broke HTTP data action export (#2626).
	if _, exists := exporter.CustomAttributeResolver["config_request.request_url_template"]; exists {
		t.Fatal("unexpected blanket resolver on config_request.request_url_template; it strips HTTP data actions too (#2626)")
	}

	t.Run("function action by category - stripped", func(t *testing.T) {
		configMap := newActionConfigMap("MyFunction Data Action", "22222222-2222-2222-2222-222222222222", false)
		if err := resolver.ResolverFunc(configMap, nil, "label"); err != nil {
			t.Fatalf("resolver returned error: %v", err)
		}
		if _, exists := requestURLTemplateFromConfigMap(t, configMap); exists {
			t.Fatal("request_url_template should be stripped for a function data action")
		}
	})

	t.Run("function action by function_config presence - stripped", func(t *testing.T) {
		configMap := newActionConfigMap("Some Category", "22222222-2222-2222-2222-222222222222", true)
		if err := resolver.ResolverFunc(configMap, nil, "label"); err != nil {
			t.Fatalf("resolver returned error: %v", err)
		}
		if _, exists := requestURLTemplateFromConfigMap(t, configMap); exists {
			t.Fatal("request_url_template should be stripped when a function_config block is present")
		}
	})

	t.Run("http data action - preserved", func(t *testing.T) {
		const url = "https://api.example.com/api/v2/divisions"
		configMap := newActionConfigMap("MSD Genesys Cloud Public API Integration", url, false)
		if err := resolver.ResolverFunc(configMap, nil, "label"); err != nil {
			t.Fatalf("resolver returned error: %v", err)
		}
		got, exists := requestURLTemplateFromConfigMap(t, configMap)
		if !exists {
			t.Fatal("request_url_template must be preserved for an HTTP data action (#2626)")
		}
		if got != url {
			t.Fatalf("request_url_template changed; want %q, got %q", url, got)
		}
	})
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
