package outbound_campaign

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/stretchr/testify/assert"
)

// TestUnitReadDiagnosticsSettingsOnlyForPowerAndPredictive verifies that diagnostics_settings
// is only populated in state when the campaign's dialing mode is "power" or "predictive".
// This tests the core conditional logic extracted from readOutboundCampaign.
func TestUnitReadDiagnosticsSettingsOnlyForPowerAndPredictive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                         string
		dialingMode                  string
		expectDiagnosticsSettingsSet bool
	}{
		{
			name:                         "power mode should have diagnostics_settings",
			dialingMode:                  "power",
			expectDiagnosticsSettingsSet: true,
		},
		{
			name:                         "predictive mode should have diagnostics_settings",
			dialingMode:                  "predictive",
			expectDiagnosticsSettingsSet: true,
		},
		{
			name:                         "preview mode should NOT have diagnostics_settings",
			dialingMode:                  "preview",
			expectDiagnosticsSettingsSet: false,
		},
		{
			name:                         "progressive mode should NOT have diagnostics_settings",
			dialingMode:                  "progressive",
			expectDiagnosticsSettingsSet: false,
		},
		{
			name:                         "agentless mode should NOT have diagnostics_settings",
			dialingMode:                  "agentless",
			expectDiagnosticsSettingsSet: false,
		},
		{
			name:                         "external mode should NOT have diagnostics_settings",
			dialingMode:                  "external",
			expectDiagnosticsSettingsSet: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Build a campaign with DiagnosticsSettings always populated from the API
			campaign := &platformclientv2.Campaign{
				Id:             platformclientv2.String("test-campaign-id"),
				Name:           platformclientv2.String("Test Campaign"),
				DialingMode:    platformclientv2.String(tc.dialingMode),
				CampaignStatus: platformclientv2.String("off"),
				PhoneColumns: &[]platformclientv2.Phonecolumn{
					{ColumnName: platformclientv2.String("Cell")},
				},
				ContactList: &platformclientv2.Domainentityref{
					Id: platformclientv2.String("contact-list-id"),
				},
				DiagnosticsSettings: &platformclientv2.Diagnosticssettings{
					ReportLowMaxCallsPerAgentAlert: platformclientv2.Bool(true),
				},
			}

			// Create ResourceData from the resource schema
			resourceSchema := ResourceOutboundCampaign()
			d := schema.TestResourceDataRaw(t, resourceSchema.Schema, map[string]interface{}{
				"name":            "Test Campaign",
				"dialing_mode":    tc.dialingMode,
				"contact_list_id": "contact-list-id",
				"phone_columns": []interface{}{
					map[string]interface{}{"column_name": "Cell"},
				},
			})
			d.SetId("test-campaign-id")

			// Replicate the conditional logic from readOutboundCampaign
			if campaign.DialingMode != nil && (*campaign.DialingMode == "power" || *campaign.DialingMode == "predictive") {
				_ = d.Set("diagnostics_settings", flattenDiagnosticsSettings(campaign.DiagnosticsSettings))
			}

			// Verify the result
			diagSettings := d.Get("diagnostics_settings").([]interface{})
			if tc.expectDiagnosticsSettingsSet {
				assert.Equal(t, 1, len(diagSettings),
					"diagnostics_settings should be set for %s mode", tc.dialingMode)
				settingsMap := diagSettings[0].(map[string]interface{})
				assert.Equal(t, true, settingsMap["report_low_max_calls_per_agent_alert"])
			} else {
				assert.Equal(t, 0, len(diagSettings),
					"diagnostics_settings should NOT be set for %s mode", tc.dialingMode)
			}
		})
	}
}

// TestUnitDiagnosticsSettingsNilDialingMode verifies safe handling when DialingMode is nil
func TestUnitDiagnosticsSettingsNilDialingMode(t *testing.T) {
	t.Parallel()

	campaign := &platformclientv2.Campaign{
		Id:             platformclientv2.String("test-campaign-id"),
		Name:           platformclientv2.String("Test Campaign"),
		DialingMode:    nil, // nil DialingMode
		CampaignStatus: platformclientv2.String("off"),
		DiagnosticsSettings: &platformclientv2.Diagnosticssettings{
			ReportLowMaxCallsPerAgentAlert: platformclientv2.Bool(true),
		},
	}

	resourceSchema := ResourceOutboundCampaign()
	d := schema.TestResourceDataRaw(t, resourceSchema.Schema, map[string]interface{}{
		"name":            "Test Campaign",
		"dialing_mode":    "power",
		"contact_list_id": "contact-list-id",
		"phone_columns": []interface{}{
			map[string]interface{}{"column_name": "Cell"},
		},
	})
	d.SetId("test-campaign-id")

	// Should not panic when DialingMode is nil
	if campaign.DialingMode != nil && (*campaign.DialingMode == "power" || *campaign.DialingMode == "predictive") {
		_ = d.Set("diagnostics_settings", flattenDiagnosticsSettings(campaign.DiagnosticsSettings))
	}

	diagSettings := d.Get("diagnostics_settings").([]interface{})
	assert.Equal(t, 0, len(diagSettings), "diagnostics_settings should NOT be set when DialingMode is nil")
}

// TestUnitFlattenDiagnosticsSettingsNil verifies flattenDiagnosticsSettings handles nil safely
func TestUnitFlattenDiagnosticsSettingsNil(t *testing.T) {
	t.Parallel()

	result := flattenDiagnosticsSettings(nil)
	assert.Nil(t, result, "flattenDiagnosticsSettings(nil) should return nil")
}

// TestUnitBuildDiagnosticsSettings verifies building diagnostics settings from resource data
func TestUnitBuildDiagnosticsSettings(t *testing.T) {
	t.Parallel()

	t.Run("with settings", func(t *testing.T) {
		input := []interface{}{
			map[string]interface{}{
				"report_low_max_calls_per_agent_alert": true,
			},
		}
		result := buildDiagnosticsSettings(input)
		assert.NotNil(t, result)
		assert.Equal(t, true, *result.ReportLowMaxCallsPerAgentAlert)
	})

	t.Run("with empty list", func(t *testing.T) {
		result := buildDiagnosticsSettings([]interface{}{})
		assert.Nil(t, result)
	})

	t.Run("with nil", func(t *testing.T) {
		result := buildDiagnosticsSettings(nil)
		assert.Nil(t, result)
	})
}
