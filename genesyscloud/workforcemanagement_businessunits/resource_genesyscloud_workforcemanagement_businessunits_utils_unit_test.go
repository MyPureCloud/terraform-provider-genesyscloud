package workforcemanagement_businessunits

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
)

func TestUnitBuildBuShortTermForecastingSettings(t *testing.T) {
	tests := []struct {
		name        string
		input       []interface{}
		expected    *platformclientv2.Bushorttermforecastingsettings
		shouldPanic bool
	}{
		{
			name: "valid_settings",
			input: []interface{}{
				map[string]interface{}{
					"default_history_weeks": 4,
				},
			},
			expected: &platformclientv2.Bushorttermforecastingsettings{
				DefaultHistoryWeeks: platformclientv2.Int(4),
			},
			shouldPanic: false,
		},
		{
			name: "zero_weeks",
			input: []interface{}{
				map[string]interface{}{
					"default_history_weeks": 0,
				},
			},
			expected: &platformclientv2.Bushorttermforecastingsettings{
				DefaultHistoryWeeks: platformclientv2.Int(0),
			},
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic but function did not panic")
					}
				}()
			}
			result := buildBuShortTermForecastingSettings(tt.input)
			if result == nil {
				t.Errorf("Expected non-nil result but got nil")
				return
			}
			if result.DefaultHistoryWeeks == nil || *result.DefaultHistoryWeeks != *tt.expected.DefaultHistoryWeeks {
				t.Errorf("Expected DefaultHistoryWeeks %d, got %v", *tt.expected.DefaultHistoryWeeks, result.DefaultHistoryWeeks)
			}
		})
	}
}

func TestUnitBuildSchedulerMessageTypeSeverities(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *[]platformclientv2.Schedulermessagetypeseverity
	}{
		{
			name:     "empty_input",
			input:    []interface{}{},
			expected: &[]platformclientv2.Schedulermessagetypeseverity{},
		},
		{
			name: "valid_severities",
			input: []interface{}{
				map[string]interface{}{
					"type":     "AgentSchedule",
					"severity": "Warning",
				},
				map[string]interface{}{
					"type":     "ShortTermForecast",
					"severity": "Error",
				},
			},
			expected: &[]platformclientv2.Schedulermessagetypeseverity{
				{
					VarType:  platformclientv2.String("AgentSchedule"),
					Severity: platformclientv2.String("Warning"),
				},
				{
					VarType:  platformclientv2.String("ShortTermForecast"),
					Severity: platformclientv2.String("Error"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildSchedulerMessageTypeSeverities(tt.input)
			if result == nil && len(*tt.expected) == 0 {
				return
			}
			if len(*result) != len(*tt.expected) {
				t.Errorf("Expected %d severities, got %d", len(*tt.expected), len(*result))
				return
			}
			for i, expected := range *tt.expected {
				if (*result)[i].VarType == nil || *(*result)[i].VarType != *expected.VarType {
					t.Errorf("Expected VarType %s, got %v", *expected.VarType, (*result)[i].VarType)
				}
				if (*result)[i].Severity == nil || *(*result)[i].Severity != *expected.Severity {
					t.Errorf("Expected Severity %s, got %v", *expected.Severity, (*result)[i].Severity)
				}
			}
		})
	}
}

func TestUnitBuildWfmServiceGoalImpact(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *platformclientv2.Wfmservicegoalimpact
	}{
		{
			name: "valid_impact",
			input: []interface{}{
				map[string]interface{}{
					"increase_by_percent": 10.5,
					"decrease_by_percent": 5.0,
				},
			},
			expected: &platformclientv2.Wfmservicegoalimpact{
				IncreaseByPercent: platformclientv2.Float64(10.5),
				DecreaseByPercent: platformclientv2.Float64(5.0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildWfmServiceGoalImpact(tt.input)
			if result == nil {
				t.Errorf("Expected non-nil result but got nil")
				return
			}
			if result.IncreaseByPercent == nil || *result.IncreaseByPercent != *tt.expected.IncreaseByPercent {
				t.Errorf("Expected IncreaseByPercent %f, got %v", *tt.expected.IncreaseByPercent, result.IncreaseByPercent)
			}
			if result.DecreaseByPercent == nil || *result.DecreaseByPercent != *tt.expected.DecreaseByPercent {
				t.Errorf("Expected DecreaseByPercent %f, got %v", *tt.expected.DecreaseByPercent, result.DecreaseByPercent)
			}
		})
	}
}

func TestUnitFlattenBuShortTermForecastingSettings(t *testing.T) {
	tests := []struct {
		name     string
		input    *platformclientv2.Bushorttermforecastingsettings
		expected []interface{}
	}{
		{
			name: "valid_settings",
			input: &platformclientv2.Bushorttermforecastingsettings{
				DefaultHistoryWeeks: platformclientv2.Int(4),
			},
			expected: []interface{}{
				map[string]interface{}{
					"default_history_weeks": 4,
				},
			},
		},
		{
			name: "nil_weeks",
			input: &platformclientv2.Bushorttermforecastingsettings{
				DefaultHistoryWeeks: nil,
			},
			expected: []interface{}{
				map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenBuShortTermForecastingSettings(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d items, got %d", len(tt.expected), len(result))
				return
			}
			if len(result) > 0 {
				resultMap := result[0].(map[string]interface{})
				expectedMap := tt.expected[0].(map[string]interface{})
				if resultMap["default_history_weeks"] != expectedMap["default_history_weeks"] {
					t.Errorf("Expected default_history_weeks %v, got %v", expectedMap["default_history_weeks"], resultMap["default_history_weeks"])
				}
			}
		})
	}
}

func TestUnitFlattenSchedulerMessageTypeSeverities(t *testing.T) {
	tests := []struct {
		name     string
		input    *[]platformclientv2.Schedulermessagetypeseverity
		expected []interface{}
	}{
		{
			name:     "empty_list",
			input:    &[]platformclientv2.Schedulermessagetypeseverity{},
			expected: nil,
		},
		{
			name: "valid_severities",
			input: &[]platformclientv2.Schedulermessagetypeseverity{
				{
					VarType:  platformclientv2.String("AgentSchedule"),
					Severity: platformclientv2.String("Warning"),
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"type":     "AgentSchedule",
					"severity": "Warning",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenSchedulerMessageTypeSeverities(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil but got %v", result)
				}
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d items, got %d", len(tt.expected), len(result))
				return
			}
			resultMap := result[0].(map[string]interface{})
			expectedMap := tt.expected[0].(map[string]interface{})
			if resultMap["type"] != expectedMap["type"] {
				t.Errorf("Expected type %v, got %v", expectedMap["type"], resultMap["type"])
			}
			if resultMap["severity"] != expectedMap["severity"] {
				t.Errorf("Expected severity %v, got %v", expectedMap["severity"], resultMap["severity"])
			}
		})
	}
}

func TestUnitGetCreateWorkforcemanagementBusinessUnitRequestFromResourceData(t *testing.T) {
	resourceSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"division_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"settings": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"start_day_of_week": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"time_zone": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		data     map[string]interface{}
		expected platformclientv2.Createbusinessunitrequest
	}{
		{
			name: "basic_request",
			data: map[string]interface{}{
				"name":        "Test Business Unit",
				"division_id": "division-123",
				"settings": []interface{}{
					map[string]interface{}{
						"start_day_of_week": "Monday",
						"time_zone":         "America/New_York",
					},
				},
			},
			expected: platformclientv2.Createbusinessunitrequest{
				Name:       platformclientv2.String("Test Business Unit"),
				DivisionId: platformclientv2.String("division-123"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.data)
			result := getCreateWorkforcemanagementBusinessUnitRequestFromResourceData(d)
			if result.Name == nil || *result.Name != *tt.expected.Name {
				t.Errorf("Expected Name %s, got %v", *tt.expected.Name, result.Name)
			}
			if result.DivisionId == nil || *result.DivisionId != *tt.expected.DivisionId {
				t.Errorf("Expected DivisionId %s, got %v", *tt.expected.DivisionId, result.DivisionId)
			}
		})
	}
}

func TestUnitGetUpdateWorkforcemanagementBusinessUnitRequestFromResourceData(t *testing.T) {
	resourceSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"division_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"settings": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"start_day_of_week": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"time_zone": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		data     map[string]interface{}
		expected platformclientv2.Updatebusinessunitrequest
	}{
		{
			name: "basic_update_request",
			data: map[string]interface{}{
				"name":        "Updated Business Unit",
				"division_id": "division-456",
				"settings": []interface{}{
					map[string]interface{}{
						"start_day_of_week": "Sunday",
						"time_zone":         "America/Los_Angeles",
					},
				},
			},
			expected: platformclientv2.Updatebusinessunitrequest{
				Name:       platformclientv2.String("Updated Business Unit"),
				DivisionId: platformclientv2.String("division-456"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.data)
			result := getUpdateWorkforcemanagementBusinessUnitRequestFromResourceData(d)
			if result.Name == nil || *result.Name != *tt.expected.Name {
				t.Errorf("Expected Name %s, got %v", *tt.expected.Name, result.Name)
			}
			if result.DivisionId == nil || *result.DivisionId != *tt.expected.DivisionId {
				t.Errorf("Expected DivisionId %s, got %v", *tt.expected.DivisionId, result.DivisionId)
			}
		})
	}
}

func TestUnitFlattenBusinessUnitSettingsResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    *platformclientv2.Businessunitsettingsresponse
		expected int // Expected number of items in the result
	}{
		{
			name:     "nil_input",
			input:    nil,
			expected: 0,
		},
		{
			name: "valid_settings",
			input: &platformclientv2.Businessunitsettingsresponse{
				StartDayOfWeek: platformclientv2.String("Monday"),
				TimeZone:       platformclientv2.String("America/New_York"),
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenBusinessUnitSettingsResponse(tt.input)
			if tt.input == nil {
				if result != nil {
					t.Errorf("Expected nil but got %v", result)
				}
				return
			}
			if len(result) != tt.expected {
				t.Errorf("Expected %d items, got %d", tt.expected, len(result))
				return
			}
			if len(result) > 0 {
				resultMap := result[0].(map[string]interface{})
				if tt.input.StartDayOfWeek != nil {
					if resultMap["start_day_of_week"] != *tt.input.StartDayOfWeek {
						t.Errorf("Expected start_day_of_week %s, got %v", *tt.input.StartDayOfWeek, resultMap["start_day_of_week"])
					}
				}
				if tt.input.TimeZone != nil {
					if resultMap["time_zone"] != *tt.input.TimeZone {
						t.Errorf("Expected time_zone %s, got %v", *tt.input.TimeZone, resultMap["time_zone"])
					}
				}
			}
		})
	}
}

func TestUnitBuildCreateBusinessUnitSettingsRequest(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected bool // Whether we expect a non-nil result
	}{
		{
			name: "valid_settings",
			input: []interface{}{
				map[string]interface{}{
					"start_day_of_week": "Monday",
					"time_zone":         "America/New_York",
				},
			},
			expected: true,
		},
		{
			name: "valid_settings_with_forecasting",
			input: []interface{}{
				map[string]interface{}{
					"start_day_of_week": "Sunday",
					"time_zone":         "America/Los_Angeles",
					"short_term_forecasting": []interface{}{
						map[string]interface{}{
							"default_history_weeks": 6,
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCreateBusinessUnitSettingsRequest(tt.input)
			if tt.expected {
				if result == nil {
					t.Errorf("Expected non-nil result but got nil")
				}
			} else {
				// For empty input, we still get a result but with empty values
				if result == nil {
					t.Errorf("Expected result but got nil")
				}
			}
		})
	}
}

func TestUnitFlattenWfmServiceGoalImpact(t *testing.T) {
	tests := []struct {
		name     string
		input    *platformclientv2.Wfmservicegoalimpact
		expected []interface{}
	}{
		{
			name: "valid_impact",
			input: &platformclientv2.Wfmservicegoalimpact{
				IncreaseByPercent: platformclientv2.Float64(10.5),
				DecreaseByPercent: platformclientv2.Float64(5.0),
			},
			expected: []interface{}{
				map[string]interface{}{
					"increase_by_percent": 10.5,
					"decrease_by_percent": 5.0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenWfmServiceGoalImpact(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d items, got %d", len(tt.expected), len(result))
				return
			}
			if len(result) > 0 {
				resultMap := result[0].(map[string]interface{})
				expectedMap := tt.expected[0].(map[string]interface{})
				if !reflect.DeepEqual(resultMap["increase_by_percent"], expectedMap["increase_by_percent"]) {
					t.Errorf("Expected increase_by_percent %v, got %v", expectedMap["increase_by_percent"], resultMap["increase_by_percent"])
				}
				if !reflect.DeepEqual(resultMap["decrease_by_percent"], expectedMap["decrease_by_percent"]) {
					t.Errorf("Expected decrease_by_percent %v, got %v", expectedMap["decrease_by_percent"], resultMap["decrease_by_percent"])
				}
			}
		})
	}
}
