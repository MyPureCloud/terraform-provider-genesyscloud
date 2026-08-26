package integration_config

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

func TestUnitBuildIntegrationConfig(t *testing.T) {
	// Setup test ResourceData
	resourceSchema := ResourceIntegrationConfig().Schema
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
		"integration_id": "test-integration-id",
		"name":           "Test Integration Config",
		"notes":          "Test notes",
		"properties":     `{"key1":"value1","key2":"value2"}`,
		"advanced":       `{"advKey":"advValue"}`,
		"credentials": map[string]interface{}{
			"pureCloudOAuthClient": "credential-id-123",
		},
	})

	version := 5
	config := buildIntegrationConfig(d, &version)

	// Verify version
	if config.Version == nil || *config.Version != 5 {
		t.Errorf("Expected version 5, got %v", config.Version)
	}

	// Verify name
	if config.Name == nil || *config.Name != "Test Integration Config" {
		t.Errorf("Expected name 'Test Integration Config', got %v", config.Name)
	}

	// Verify notes
	if config.Notes == nil || *config.Notes != "Test notes" {
		t.Errorf("Expected notes 'Test notes', got %v", config.Notes)
	}

	// Verify properties is not nil
	if config.Properties == nil {
		t.Fatal("Expected properties to not be nil")
	}

	// Verify advanced is not nil
	if config.Advanced == nil {
		t.Fatal("Expected advanced to not be nil")
	}

	// Verify credentials
	if config.Credentials == nil {
		t.Fatal("Expected credentials to not be nil")
	}
	creds := *config.Credentials
	if cred, ok := creds["pureCloudOAuthClient"]; !ok {
		t.Error("Expected 'pureCloudOAuthClient' key in credentials")
	} else if cred.Id == nil || *cred.Id != "credential-id-123" {
		t.Errorf("Expected credential ID 'credential-id-123', got %v", cred.Id)
	}
}

func TestUnitBuildIntegrationConfigEmptyFields(t *testing.T) {
	// Test that empty/missing fields still produce valid objects (not nil)
	resourceSchema := ResourceIntegrationConfig().Schema
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
		"integration_id": "test-integration-id",
		"name":           "",
		"notes":          "",
	})

	version := 1
	config := buildIntegrationConfig(d, &version)

	// Properties should be empty object, not nil
	if config.Properties == nil {
		t.Fatal("Expected properties to be empty object, got nil")
	}

	// Advanced should be empty object, not nil
	if config.Advanced == nil {
		t.Fatal("Expected advanced to be empty object, got nil")
	}

	// Credentials should be empty map, not nil
	if config.Credentials == nil {
		t.Fatal("Expected credentials to be empty map, got nil")
	}
}

func TestUnitBuildCredentials(t *testing.T) {
	input := map[string]interface{}{
		"pureCloudOAuthClient": "cred-id-1",
		"basicAuth":            "cred-id-2",
	}

	result := buildCredentials(input)

	if len(result) != 2 {
		t.Errorf("Expected 2 credentials, got %d", len(result))
	}

	if cred, ok := result["pureCloudOAuthClient"]; !ok {
		t.Error("Missing 'pureCloudOAuthClient' key")
	} else if *cred.Id != "cred-id-1" {
		t.Errorf("Expected 'cred-id-1', got '%s'", *cred.Id)
	}

	if cred, ok := result["basicAuth"]; !ok {
		t.Error("Missing 'basicAuth' key")
	} else if *cred.Id != "cred-id-2" {
		t.Errorf("Expected 'cred-id-2', got '%s'", *cred.Id)
	}
}

func TestUnitFlattenCredentials(t *testing.T) {
	credId1 := "cred-id-1"
	credId2 := "cred-id-2"

	input := map[string]platformclientv2.Credentialinfo{
		"pureCloudOAuthClient": {Id: &credId1},
		"basicAuth":            {Id: &credId2},
	}

	result := flattenCredentials(input)

	if len(result) != 2 {
		t.Errorf("Expected 2 credentials, got %d", len(result))
	}

	if result["pureCloudOAuthClient"] != "cred-id-1" {
		t.Errorf("Expected 'cred-id-1', got '%s'", result["pureCloudOAuthClient"])
	}

	if result["basicAuth"] != "cred-id-2" {
		t.Errorf("Expected 'cred-id-2', got '%s'", result["basicAuth"])
	}
}

func TestUnitFlattenCredentialsEmpty(t *testing.T) {
	input := map[string]platformclientv2.Credentialinfo{}
	result := flattenCredentials(input)

	if result != nil {
		t.Errorf("Expected nil for empty credentials, got %v", result)
	}
}

func TestUnitFlattenIntegrationConfig(t *testing.T) {
	name := "Test Config"
	notes := "Test Notes"
	credId := "cred-123"

	properties := map[string]interface{}{"key": "value"}
	propsInterface := interface{}(properties)

	advanced := map[string]interface{}{"advKey": "advValue"}
	advInterface := interface{}(advanced)

	config := &platformclientv2.Integrationconfiguration{
		Name:       &name,
		Notes:      &notes,
		Properties: &propsInterface,
		Advanced:   &advInterface,
		Credentials: &map[string]platformclientv2.Credentialinfo{
			"oauth": {Id: &credId},
		},
	}

	// Create test ResourceData to flatten into
	resourceSchema := ResourceIntegrationConfig().Schema
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
		"integration_id": "test-id",
	})

	flattenIntegrationConfig(d, config)

	// Verify name
	if d.Get("name") != "Test Config" {
		t.Errorf("Expected name 'Test Config', got '%s'", d.Get("name"))
	}

	// Verify notes
	if d.Get("notes") != "Test Notes" {
		t.Errorf("Expected notes 'Test Notes', got '%s'", d.Get("notes"))
	}

	// Verify properties is valid JSON
	propsStr := d.Get("properties").(string)
	var propsMap map[string]interface{}
	if err := json.Unmarshal([]byte(propsStr), &propsMap); err != nil {
		t.Fatalf("Expected valid JSON properties, got error: %v", err)
	}
	if propsMap["key"] != "value" {
		t.Errorf("Expected properties key='value', got '%v'", propsMap["key"])
	}

	// Verify advanced is valid JSON
	advStr := d.Get("advanced").(string)
	var advMap map[string]interface{}
	if err := json.Unmarshal([]byte(advStr), &advMap); err != nil {
		t.Fatalf("Expected valid JSON advanced, got error: %v", err)
	}
	if advMap["advKey"] != "advValue" {
		t.Errorf("Expected advanced advKey='advValue', got '%v'", advMap["advKey"])
	}
}

func TestUnitFlattenIntegrationConfigNilFields(t *testing.T) {
	// Test with nil config
	resourceSchema := ResourceIntegrationConfig().Schema
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
		"integration_id": "test-id",
	})

	// Should not panic
	flattenIntegrationConfig(d, nil)
}

func TestUnitFlattenIntegrationConfigNodeDynamoDBEmptyString(t *testing.T) {
	// Test the special "node_dynamodb_empty_string" handling
	name := "Test"
	notes := "node_dynamodb_empty_string"

	config := &platformclientv2.Integrationconfiguration{
		Name:  &name,
		Notes: &notes,
	}

	resourceSchema := ResourceIntegrationConfig().Schema
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
		"integration_id": "test-id",
	})

	flattenIntegrationConfig(d, config)

	if d.Get("notes") != "" {
		t.Errorf("Expected empty notes (node_dynamodb_empty_string conversion), got '%s'", d.Get("notes"))
	}
}
