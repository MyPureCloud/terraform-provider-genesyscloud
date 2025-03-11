package telephony_providers_edges_site

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestSiteSchemaUpgradeV1ToV2(t *testing.T) {
	// Create old state with outbound routes
	oldState := map[string]interface{}{
		"id":          "site_1234",
		"name":        "Test Site",
		"description": "Test Description",
		"outbound_routes": []interface{}{
			map[string]interface{}{
				"id":                      "route_5678",
				"name":                    "International",
				"description":             "International calling route",
				"classification_types":    []interface{}{"International"},
				"enabled":                 true,
				"distribution":            "SEQUENTIAL",
				"external_trunk_base_ids": []interface{}{"trunk_base_1", "trunk_base_2"},
			},
			map[string]interface{}{
				"id":                      "route_9012",
				"name":                    "Emergency",
				"description":             "Emergency calling route",
				"classification_types":    []interface{}{"Emergency"},
				"enabled":                 true,
				"distribution":            "SEQUENTIAL",
				"external_trunk_base_ids": []interface{}{"trunk_base_3"},
			},
		},
	}

	// Create expected state without outbound routes
	expectedState := map[string]interface{}{
		"id":          "site_1234",
		"name":        "Test Site",
		"description": "Test Description",
	}

	// Run the upgrade function
	actualState, err := resourceSiteUpgradeV1ToV2(context.Background(), oldState, nil)

	// Assert no error occurred
	assert.NoError(t, err)

	// Assert the outbound_routes field was removed
	assert.NotContains(t, actualState, "outbound_routes")

	// Assert the rest of the state remains unchanged
	assert.Equal(t, expectedState["id"], actualState["id"])
	assert.Equal(t, expectedState["name"], actualState["name"])
	assert.Equal(t, expectedState["description"], actualState["description"])
}

// TestSiteSchemaV1ToV2Output tests that the upgrade outputs the correct migration instructions
func TestSiteSchemaV1ToV2Output(t *testing.T) {
	// Create old state with outbound routes
	oldState := map[string]interface{}{
		"id":   "site_1234",
		"name": "Test Site",
		"outbound_routes": schema.NewSet(schema.HashResource(&schema.Resource{
			Schema: map[string]*schema.Schema{
				"id":                      {Type: schema.TypeString},
				"name":                    {Type: schema.TypeString},
				"description":             {Type: schema.TypeString},
				"classification_types":    {Type: schema.TypeList, Elem: &schema.Schema{Type: schema.TypeString}},
				"enabled":                 {Type: schema.TypeBool},
				"distribution":            {Type: schema.TypeString},
				"external_trunk_base_ids": {Type: schema.TypeList, Elem: &schema.Schema{Type: schema.TypeString}},
			},
		}), []interface{}{
			map[string]interface{}{
				"id":                      "route_5678",
				"name":                    "International",
				"description":             "International calling route",
				"classification_types":    []interface{}{"International"},
				"enabled":                 true,
				"distribution":            "SEQUENTIAL",
				"external_trunk_base_ids": []interface{}{"trunk_base_1", "trunk_base_2"},
			},
		}),
	}

	// Capture output using a custom writer
	output := captureOutput(t, func() {
		_, err := resourceSiteUpgradeV1ToV2(context.Background(), oldState, nil)
		assert.NoError(t, err)
	})

	// Assert the output contains expected migration instructions
	assert.Contains(t, output, "[MIGRATION] Found 1 outbound routes in site site_1234")
	assert.Contains(t, output, `resource "genesyscloud_telephony_providers_edges_site_outbound_route" "example_international"`)
	assert.Contains(t, output, `site_id               = "site_1234"`)
	assert.Contains(t, output, `name                  = "International"`)
	assert.Contains(t, output, "terraform import genesyscloud_telephony_providers_edges_site_outbound_route.example_international site_1234:route_5678")
}

func captureOutput(t *testing.T, f func()) string {
	// Save existing stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the function
	f()

	// Reset stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
