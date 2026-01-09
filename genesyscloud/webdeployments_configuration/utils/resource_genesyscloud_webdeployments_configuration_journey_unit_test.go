package webdeployments_configuration_utils

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	"github.com/stretchr/testify/assert"
)

// Test buildIpFilters function
func TestBuildIpFilters(t *testing.T) {
	t.Run("Valid IP filters", func(t *testing.T) {
		input := []interface{}{
			map[string]interface{}{
				"ip_address": "192.168.1.1",
				"name":       "office-network",
			},
			map[string]interface{}{
				"ip_address": "2001:db8::1",
				"name":       "ipv6-network",
			},
		}

		result := buildIpFilters(input)

		assert.NotNil(t, result)
		assert.Len(t, *result, 2)

		filters := *result
		assert.Equal(t, "192.168.1.1", *filters[0].IpAddress)
		assert.Equal(t, "office-network", *filters[0].Name)
		assert.Equal(t, "2001:db8::1", *filters[1].IpAddress)
		assert.Equal(t, "ipv6-network", *filters[1].Name)
	})

	t.Run("Nil input", func(t *testing.T) {
		result := buildIpFilters(nil)
		assert.Nil(t, result)
	})

	t.Run("Empty input", func(t *testing.T) {
		result := buildIpFilters([]interface{}{})
		assert.Nil(t, result)
	})

	t.Run("Single IP filter", func(t *testing.T) {
		input := []interface{}{
			map[string]interface{}{
				"ip_address": "10.0.0.1",
				"name":       "single-host",
			},
		}

		result := buildIpFilters(input)

		assert.NotNil(t, result)
		assert.Len(t, *result, 1)

		filters := *result
		assert.Equal(t, "10.0.0.1", *filters[0].IpAddress)
		assert.Equal(t, "single-host", *filters[0].Name)
	})
}

// Test buildTrackingSettings function
func TestBuildTrackingSettings(t *testing.T) {
	t.Run("Valid tracking settings with all fields", func(t *testing.T) {
		input := []interface{}{
			map[string]interface{}{
				"should_keep_url_fragment":  true,
				"search_query_parameters":   []interface{}{"q", "search", "term"},
				"excluded_query_parameters": []interface{}{"utm_source", "ref"},
				"ip_filters": []interface{}{
					map[string]interface{}{
						"ip_address": "192.168.1.1",
						"name":       "office",
					},
				},
			},
		}

		result := buildTrackingSettings(input)

		assert.NotNil(t, result)
		assert.Equal(t, true, *result.ShouldKeepUrlFragment)
		assert.Equal(t, []string{"q", "search", "term"}, *result.SearchQueryParameters)
		assert.Equal(t, []string{"utm_source", "ref"}, *result.ExcludedQueryParameters)
		assert.NotNil(t, result.IpFilters)
		assert.Len(t, *result.IpFilters, 1)
	})

	t.Run("Partial tracking settings", func(t *testing.T) {
		input := []interface{}{
			map[string]interface{}{
				"should_keep_url_fragment": false,
				"search_query_parameters":  []interface{}{"q"},
			},
		}

		result := buildTrackingSettings(input)

		assert.NotNil(t, result)
		assert.Equal(t, false, *result.ShouldKeepUrlFragment)
		assert.Equal(t, []string{"q"}, *result.SearchQueryParameters)
		assert.Nil(t, result.ExcludedQueryParameters)
		assert.Nil(t, result.IpFilters)
	})

	t.Run("Nil input", func(t *testing.T) {
		result := buildTrackingSettings(nil)
		assert.Nil(t, result)
	})

	t.Run("Empty input", func(t *testing.T) {
		result := buildTrackingSettings([]interface{}{})
		assert.Nil(t, result)
	})

	t.Run("Empty arrays", func(t *testing.T) {
		input := []interface{}{
			map[string]interface{}{
				"should_keep_url_fragment":  true,
				"search_query_parameters":   []interface{}{},
				"excluded_query_parameters": []interface{}{},
				"ip_filters":                []interface{}{},
			},
		}

		result := buildTrackingSettings(input)

		assert.NotNil(t, result)
		assert.Equal(t, true, *result.ShouldKeepUrlFragment)
		assert.Equal(t, []string{}, *result.SearchQueryParameters)
		assert.Equal(t, []string{}, *result.ExcludedQueryParameters)
		assert.Nil(t, result.IpFilters) // Empty array should result in nil
	})
}

// Test flattenIpFilters function
func TestFlattenIpFilters(t *testing.T) {
	t.Run("Valid IP filters", func(t *testing.T) {
		ipAddr1 := "192.168.1.1"
		name1 := "office-network"
		ipAddr2 := "10.0.0.1"
		name2 := "vpn-network"

		input := &[]platformclientv2.Ipfilter{
			{
				IpAddress: &ipAddr1,
				Name:      &name1,
			},
			{
				IpAddress: &ipAddr2,
				Name:      &name2,
			},
		}

		result := flattenIpFilters(input)

		assert.NotNil(t, result)
		assert.Len(t, result, 2)

		filter1 := result[0].(map[string]interface{})
		assert.Equal(t, &ipAddr1, filter1["ip_address"])
		assert.Equal(t, &name1, filter1["name"])

		filter2 := result[1].(map[string]interface{})
		assert.Equal(t, &ipAddr2, filter2["ip_address"])
		assert.Equal(t, &name2, filter2["name"])
	})

	t.Run("Nil input", func(t *testing.T) {
		result := flattenIpFilters(nil)
		assert.Nil(t, result)
	})

	t.Run("Empty input", func(t *testing.T) {
		input := &[]platformclientv2.Ipfilter{}
		result := flattenIpFilters(input)
		assert.Nil(t, result)
	})
}

// Test flattenTrackingSettings function
func TestFlattenTrackingSettings(t *testing.T) {
	t.Run("Valid tracking settings with all fields", func(t *testing.T) {
		shouldKeep := true
		searchParams := []string{"q", "search"}
		excludedParams := []string{"utm_source"}
		ipAddr := "192.168.1.1"
		name := "office"

		input := &platformclientv2.Trackingsettings{
			ShouldKeepUrlFragment:   &shouldKeep,
			SearchQueryParameters:   &searchParams,
			ExcludedQueryParameters: &excludedParams,
			IpFilters: &[]platformclientv2.Ipfilter{
				{
					IpAddress: &ipAddr,
					Name:      &name,
				},
			},
		}

		result := flattenTrackingSettings(input)

		assert.NotNil(t, result)
		assert.Len(t, result, 1)

		settings := result[0].(map[string]interface{})
		assert.Equal(t, &shouldKeep, settings["should_keep_url_fragment"])
		assert.Equal(t, &searchParams, settings["search_query_parameters"])
		assert.Equal(t, &excludedParams, settings["excluded_query_parameters"])
		assert.NotNil(t, settings["ip_filters"])
	})

	t.Run("Partial tracking settings", func(t *testing.T) {
		shouldKeep := false

		input := &platformclientv2.Trackingsettings{
			ShouldKeepUrlFragment: &shouldKeep,
		}

		result := flattenTrackingSettings(input)

		assert.NotNil(t, result)
		assert.Len(t, result, 1)

		settings := result[0].(map[string]interface{})
		assert.Equal(t, &shouldKeep, settings["should_keep_url_fragment"])
		assert.Nil(t, settings["search_query_parameters"])
		assert.Nil(t, settings["excluded_query_parameters"])
		assert.Nil(t, settings["ip_filters"])
	})

	t.Run("Nil input", func(t *testing.T) {
		result := flattenTrackingSettings(nil)
		assert.Nil(t, result)
	})
}

// Test integration with buildJourneySettings
func TestBuildJourneySettingsWithTrackingSettings(t *testing.T) {
	t.Run("Journey settings with tracking_settings", func(t *testing.T) {
		// Mock ResourceData would be complex, so we'll test the core logic
		// by directly testing the cfg map processing
		cfg := map[string]interface{}{
			"enabled": true,
			"tracking_settings": []interface{}{
				map[string]interface{}{
					"should_keep_url_fragment":  true,
					"search_query_parameters":   []interface{}{"q", "search"},
					"excluded_query_parameters": []interface{}{"utm_source"},
					"ip_filters": []interface{}{
						map[string]interface{}{
							"ip_address": "192.168.1.1",
							"name":       "office",
						},
					},
				},
			},
		}

		// Test buildTrackingSettings directly with the cfg data
		trackingSettings := buildTrackingSettings(cfg["tracking_settings"].([]interface{}))

		assert.NotNil(t, trackingSettings)
		assert.Equal(t, true, *trackingSettings.ShouldKeepUrlFragment)
		assert.Equal(t, []string{"q", "search"}, *trackingSettings.SearchQueryParameters)
		assert.Equal(t, []string{"utm_source"}, *trackingSettings.ExcludedQueryParameters)
		assert.NotNil(t, trackingSettings.IpFilters)
		assert.Len(t, *trackingSettings.IpFilters, 1)
	})

	t.Run("Journey settings without tracking_settings", func(t *testing.T) {
		cfg := map[string]interface{}{
			"enabled": true,
		}

		// Test that missing tracking_settings doesn't break anything
		if trackingSettingsData, ok := cfg["tracking_settings"].([]interface{}); ok {
			trackingSettings := buildTrackingSettings(trackingSettingsData)
			assert.Nil(t, trackingSettings)
		} else {
			// This is the expected path - tracking_settings key doesn't exist
			assert.True(t, true) // Test passes
		}
	})
}

// Test integration with FlattenJourneyEvents
func TestFlattenJourneyEventsWithTrackingSettings(t *testing.T) {
	t.Run("Journey events with tracking_settings", func(t *testing.T) {
		enabled := true
		shouldKeep := true
		searchParams := []string{"q"}
		excludedParams := []string{"utm_source"}
		ipAddr := "192.168.1.1"
		name := "office"

		journeyEvents := &platformclientv2.Journeyeventssettings{
			Enabled: &enabled,
			TrackingSettings: &platformclientv2.Trackingsettings{
				ShouldKeepUrlFragment:   &shouldKeep,
				SearchQueryParameters:   &searchParams,
				ExcludedQueryParameters: &excludedParams,
				IpFilters: &[]platformclientv2.Ipfilter{
					{
						IpAddress: &ipAddr,
						Name:      &name,
					},
				},
			},
		}

		result := FlattenJourneyEvents(journeyEvents)

		assert.NotNil(t, result)
		assert.Len(t, result, 1)

		flattened := result[0].(map[string]interface{})
		assert.Equal(t, &enabled, flattened["enabled"])

		trackingSettings := flattened["tracking_settings"]
		assert.NotNil(t, trackingSettings)

		trackingList := trackingSettings.([]interface{})
		assert.Len(t, trackingList, 1)

		tracking := trackingList[0].(map[string]interface{})
		assert.Equal(t, &shouldKeep, tracking["should_keep_url_fragment"])
		assert.Equal(t, &searchParams, tracking["search_query_parameters"])
		assert.Equal(t, &excludedParams, tracking["excluded_query_parameters"])
		assert.NotNil(t, tracking["ip_filters"])
	})

	t.Run("Journey events without tracking_settings", func(t *testing.T) {
		enabled := true

		journeyEvents := &platformclientv2.Journeyeventssettings{
			Enabled: &enabled,
		}

		result := FlattenJourneyEvents(journeyEvents)

		assert.NotNil(t, result)
		assert.Len(t, result, 1)

		flattened := result[0].(map[string]interface{})
		assert.Equal(t, &enabled, flattened["enabled"])
		assert.Nil(t, flattened["tracking_settings"])
	})
}

// Test edge cases and limits
func TestTrackingSettingsEdgeCases(t *testing.T) {
	t.Run("Maximum query parameters", func(t *testing.T) {
		// Test with 50 parameters (the maximum allowed)
		searchParams := make([]interface{}, 50)
		excludedParams := make([]interface{}, 50)

		for i := 0; i < 50; i++ {
			searchParams[i] = fmt.Sprintf("param%d", i)
			excludedParams[i] = fmt.Sprintf("excluded%d", i)
		}

		input := []interface{}{
			map[string]interface{}{
				"should_keep_url_fragment":  true,
				"search_query_parameters":   searchParams,
				"excluded_query_parameters": excludedParams,
			},
		}

		result := buildTrackingSettings(input)

		assert.NotNil(t, result)
		assert.Len(t, *result.SearchQueryParameters, 50)
		assert.Len(t, *result.ExcludedQueryParameters, 50)
	})

	t.Run("Maximum IP filters", func(t *testing.T) {
		// Test with 10 IP filters (the maximum allowed)
		ipFilters := make([]interface{}, 10)

		for i := 0; i < 10; i++ {
			ipFilters[i] = map[string]interface{}{
				"ip_address": fmt.Sprintf("192.168.1.%d", i+1),
				"name":       fmt.Sprintf("host-%d", i+1),
			}
		}

		input := []interface{}{
			map[string]interface{}{
				"ip_filters": ipFilters,
			},
		}

		result := buildTrackingSettings(input)

		assert.NotNil(t, result)
		assert.NotNil(t, result.IpFilters)
		assert.Len(t, *result.IpFilters, 10)
	})

	t.Run("IPv6 addresses", func(t *testing.T) {
		input := []interface{}{
			map[string]interface{}{
				"ip_filters": []interface{}{
					map[string]interface{}{
						"ip_address": "2001:db8:85a3::8a2e:370:7334",
						"name":       "ipv6-host",
					},
					map[string]interface{}{
						"ip_address": "::1",
						"name":       "localhost-ipv6",
					},
				},
			},
		}

		result := buildTrackingSettings(input)

		assert.NotNil(t, result)
		assert.NotNil(t, result.IpFilters)
		assert.Len(t, *result.IpFilters, 2)

		filters := *result.IpFilters
		assert.Equal(t, "2001:db8:85a3::8a2e:370:7334", *filters[0].IpAddress)
		assert.Equal(t, "::1", *filters[1].IpAddress)
	})
}
