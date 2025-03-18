package routing_queue

import (
	"testing"

	"context"
	"reflect"
)

func TestStateUpgraderRoutingQueueV1ToV2(t *testing.T) {
	tests := []struct {
		name     string
		rawState map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "removes attributes from all media settings",
			rawState: map[string]interface{}{
				"name": "Test Queue",
				"media_settings_callback": []interface{}{
					map[string]interface{}{
						"mode":                     "AgentFirst",
						"enable_auto_dial_and_end": true,
						"auto_dial_delay_seconds":  30,
						"auto_end_delay_seconds":   60,
						"alerting_timeout_sec":     30,
					},
				},
				"media_settings_call": []interface{}{
					map[string]interface{}{
						"mode":                     "AgentFirst",
						"enable_auto_dial_and_end": true,
						"auto_dial_delay_seconds":  30,
						"auto_end_delay_seconds":   60,
						"alerting_timeout_sec":     30,
					},
				},
				"media_settings_email": []interface{}{
					map[string]interface{}{
						"mode":                     "CustomerFirst",
						"enable_auto_dial_and_end": false,
						"auto_dial_delay_seconds":  20,
						"auto_end_delay_seconds":   40,
						"alerting_timeout_sec":     20,
					},
				},
				"media_settings_chat": []interface{}{
					map[string]interface{}{
						"mode":                     "AgentFirst",
						"enable_auto_dial_and_end": true,
						"auto_dial_delay_seconds":  15,
						"auto_end_delay_seconds":   25,
						"alerting_timeout_sec":     15,
					},
				},
				"media_settings_message": []interface{}{
					map[string]interface{}{
						"mode":                     "CustomerFirst",
						"enable_auto_dial_and_end": false,
						"auto_dial_delay_seconds":  10,
						"auto_end_delay_seconds":   20,
						"alerting_timeout_sec":     10,
					},
				},
			},
			expected: map[string]interface{}{
				"name": "Test Queue",
				"media_settings_callback": []interface{}{
					map[string]interface{}{
						"mode":                     "AgentFirst",
						"enable_auto_dial_and_end": true,
						"auto_dial_delay_seconds":  30,
						"auto_end_delay_seconds":   60,
						"alerting_timeout_sec":     30,
					},
				},
				"media_settings_call": []interface{}{
					map[string]interface{}{
						"alerting_timeout_sec": 30,
					},
				},
				"media_settings_email": []interface{}{
					map[string]interface{}{
						"alerting_timeout_sec": 20,
					},
				},
				"media_settings_chat": []interface{}{
					map[string]interface{}{
						"alerting_timeout_sec": 15,
					},
				},
				"media_settings_message": []interface{}{
					map[string]interface{}{
						"alerting_timeout_sec": 10,
					},
				},
			},
		},
		{
			name: "handles empty media settings",
			rawState: map[string]interface{}{
				"name":                 "Test Queue",
				"media_settings_call":  []interface{}{},
				"media_settings_email": []interface{}{},
				"media_settings_chat":  []interface{}{},
				"media_settings_message": []interface{}{
					map[string]interface{}{
						"mode":                     "CustomerFirst",
						"enable_auto_dial_and_end": false,
						"auto_dial_delay_seconds":  10,
						"auto_end_delay_seconds":   20,
						"alerting_timeout_sec":     10,
					},
				},
			},
			expected: map[string]interface{}{
				"name":                 "Test Queue",
				"media_settings_call":  []interface{}{},
				"media_settings_email": []interface{}{},
				"media_settings_chat":  []interface{}{},
				"media_settings_message": []interface{}{
					map[string]interface{}{
						"alerting_timeout_sec": 10,
					},
				},
			},
		},
		{
			name: "handles missing media settings",
			rawState: map[string]interface{}{
				"name": "Test Queue",
			},
			expected: map[string]interface{}{
				"name": "Test Queue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upgraded, err := stateUpgraderRoutingQueueV1ToV2(context.Background(), tt.rawState, nil)
			if err != nil {
				t.Errorf("stateUpgraderRoutingQueueV1ToV2() error = %v", err)
				return
			}

			if !reflect.DeepEqual(upgraded, tt.expected) {
				t.Errorf("stateUpgraderRoutingQueueV1ToV2() = %v, want %v", upgraded, tt.expected)
			}
		})
	}
}
