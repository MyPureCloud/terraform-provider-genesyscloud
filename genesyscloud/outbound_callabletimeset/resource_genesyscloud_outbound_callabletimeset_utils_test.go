package outbound_callabletimeset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v204/platformclientv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Unit tests for flattenCallableTimes and keepApiCallableTimeName. When a callable time is created without a `name`
// the API fills one in. `name` is Computed, so Terraform accepts that, but the provider's consistency check compares the
// planned set element (no name) with the element that was read back (with a name) and fails the apply. The API's name
// must therefore only be written to state when the planned or previous element already had one.

func apiCallableTime(timeZoneId, name string, days ...int) platformclientv2.Callabletime {
	slots := make([]platformclientv2.Campaigntimeslot, 0, len(days))
	for _, day := range days {
		slots = append(slots, platformclientv2.Campaigntimeslot{
			StartTime: platformclientv2.String("07:00:00"),
			StopTime:  platformclientv2.String("18:00:00"),
			Day:       platformclientv2.Int(day),
		})
	}
	callableTime := platformclientv2.Callabletime{
		TimeSlots:  &slots,
		TimeZoneId: platformclientv2.String(timeZoneId),
	}
	if name != "" {
		callableTime.Name = platformclientv2.String(name)
	}
	return callableTime
}

// existingCallableTimes builds the set the way the read sees it: the planned values after a create or update, or the
// previous state on a refresh. An empty name stands for an attribute that was not set.
func existingCallableTimes(t *testing.T, elements ...map[string]interface{}) *schema.Set {
	t.Helper()
	list := make([]interface{}, 0, len(elements))
	for _, e := range elements {
		list = append(list, e)
	}
	d := schema.TestResourceDataRaw(t, ResourceOutboundCallabletimeset().Schema, map[string]interface{}{
		"name":           "set",
		"callable_times": list,
	})
	set, ok := d.Get("callable_times").(*schema.Set)
	require.True(t, ok)
	return set
}

func plannedElement(timeZoneId, name string, days ...int) map[string]interface{} {
	slots := make([]interface{}, 0, len(days))
	for _, day := range days {
		slots = append(slots, map[string]interface{}{"start_time": "07:00:00", "stop_time": "18:00:00", "day": day})
	}
	element := map[string]interface{}{"time_zone_id": timeZoneId, "time_slots": slots}
	if name != "" {
		element["name"] = name
	}
	return element
}

// namesOf returns the name of every element in the flattened set.
func namesOf(t *testing.T, set *schema.Set) []string {
	t.Helper()
	require.NotNil(t, set)
	var names []string
	for _, e := range set.List() {
		m, ok := e.(map[string]interface{})
		require.True(t, ok)
		name, _ := m["name"].(string)
		names = append(names, name)
	}
	return names
}

// The reported failure: created without a name, the API returns the time zone as the name.
func TestUnitFlattenCallableTimes_DropsApiNameWhenNotPlanned(t *testing.T) {
	existing := existingCallableTimes(t, plannedElement("Africa/Abidjan", "", 3, 5))
	api := []platformclientv2.Callabletime{apiCallableTime("Africa/Abidjan", "Africa/Abidjan", 3, 5)}

	got := flattenCallableTimes(api, existing)

	assert.Equal(t, []string{""}, namesOf(t, got), "the API's name must not be written when none was planned")
}

func TestUnitFlattenCallableTimes_KeepsPlannedName(t *testing.T) {
	existing := existingCallableTimes(t, plannedElement("Africa/Abidjan", "Morning", 3, 5))
	api := []platformclientv2.Callabletime{apiCallableTime("Africa/Abidjan", "Morning", 3, 5)}

	got := flattenCallableTimes(api, existing)

	assert.Equal(t, []string{"Morning"}, namesOf(t, got))
}

// A name equal to the time zone is a name the user wrote, so it must be kept.
func TestUnitFlattenCallableTimes_KeepsPlannedNameEqualToTimeZone(t *testing.T) {
	existing := existingCallableTimes(t, plannedElement("Africa/Abidjan", "Africa/Abidjan", 3))
	api := []platformclientv2.Callabletime{apiCallableTime("Africa/Abidjan", "Africa/Abidjan", 3)}

	got := flattenCallableTimes(api, existing)

	assert.Equal(t, []string{"Africa/Abidjan"}, namesOf(t, got))
}

// Import and export have no previous value, so the API's name is kept.
func TestUnitFlattenCallableTimes_NoExistingKeepsApiName(t *testing.T) {
	api := []platformclientv2.Callabletime{apiCallableTime("Africa/Abidjan", "Africa/Abidjan", 3, 5)}

	assert.Equal(t, []string{"Africa/Abidjan"}, namesOf(t, flattenCallableTimes(api, nil)))
	assert.Equal(t, []string{"Africa/Abidjan"}, namesOf(t, flattenCallableTimes(api, schema.NewSet(schema.HashResource(timeSlotResource), nil))))
}

// Elements are matched on time zone and time slots. A callable time with no match (changed outside Terraform) keeps
// its name.
func TestUnitFlattenCallableTimes_MatchesOnTimeZoneAndSlots(t *testing.T) {
	existing := existingCallableTimes(t, plannedElement("Africa/Abidjan", "", 3, 5))

	t.Run("different time zone", func(t *testing.T) {
		api := []platformclientv2.Callabletime{apiCallableTime("Europe/Dublin", "Europe/Dublin", 3, 5)}
		assert.Equal(t, []string{"Europe/Dublin"}, namesOf(t, flattenCallableTimes(api, existing)))
	})
	t.Run("different time slots", func(t *testing.T) {
		api := []platformclientv2.Callabletime{apiCallableTime("Africa/Abidjan", "Africa/Abidjan", 1, 2)}
		assert.Equal(t, []string{"Africa/Abidjan"}, namesOf(t, flattenCallableTimes(api, existing)))
	})
}

// Two callable times in one set: one with a name and one without, in either order.
func TestUnitFlattenCallableTimes_MixedNamedAndUnnamed(t *testing.T) {
	existing := existingCallableTimes(t,
		plannedElement("Africa/Abidjan", "Morning", 3),
		plannedElement("Europe/Dublin", "", 5),
	)
	api := []platformclientv2.Callabletime{
		apiCallableTime("Europe/Dublin", "Europe/Dublin", 5),
		apiCallableTime("Africa/Abidjan", "Morning", 3),
	}

	got := flattenCallableTimes(api, existing)

	assert.ElementsMatch(t, []string{"", "Morning"}, namesOf(t, got))
}

// Two callable times with the same time zone and slots that differ only by name are paired one to one.
func TestUnitFlattenCallableTimes_DuplicatesArePairedOneToOne(t *testing.T) {
	existing := existingCallableTimes(t,
		plannedElement("Africa/Abidjan", "Morning", 3),
		plannedElement("Africa/Abidjan", "", 3),
	)
	api := []platformclientv2.Callabletime{
		apiCallableTime("Africa/Abidjan", "Africa/Abidjan", 3),
		apiCallableTime("Africa/Abidjan", "Morning", 3),
	}

	got := flattenCallableTimes(api, existing)

	assert.ElementsMatch(t, []string{"", "Morning"}, namesOf(t, got))
}

func TestUnitFlattenCallableTimes_EdgeCases(t *testing.T) {
	t.Run("no callable times returns nil", func(t *testing.T) {
		assert.Nil(t, flattenCallableTimes(nil, existingCallableTimes(t, plannedElement("Africa/Abidjan", "", 3))))
	})

	t.Run("API callable time without a name", func(t *testing.T) {
		existing := existingCallableTimes(t, plannedElement("Africa/Abidjan", "", 3))
		api := []platformclientv2.Callabletime{apiCallableTime("Africa/Abidjan", "", 3)}
		assert.Equal(t, []string{""}, namesOf(t, flattenCallableTimes(api, existing)))
	})

	t.Run("existing set is not modified", func(t *testing.T) {
		existing := existingCallableTimes(t, plannedElement("Africa/Abidjan", "Morning", 3))
		before := existing.Len()
		flattenCallableTimes([]platformclientv2.Callabletime{apiCallableTime("Africa/Abidjan", "Morning", 3)}, existing)
		assert.Equal(t, before, existing.Len())
	})
}

// End to end through the schema, the way readOutboundCallabletimeset stores it: planned values with no name, API result
// with a name, then d.Set. The stored element must have the same identity as the planned one, which is what the
// provider's consistency check compares.
func TestUnitFlattenCallableTimes_StoredElementMatchesPlannedIdentity(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceOutboundCallabletimeset().Schema, map[string]interface{}{
		"name":           "set",
		"callable_times": []interface{}{plannedElement("Africa/Abidjan", "", 3, 5)},
	})
	planned, ok := d.Get("callable_times").(*schema.Set)
	require.True(t, ok)
	api := []platformclientv2.Callabletime{apiCallableTime("Africa/Abidjan", "Africa/Abidjan", 3, 5)}

	require.NoError(t, d.Set("callable_times", flattenCallableTimes(api, planned)))

	stored, ok := d.Get("callable_times").(*schema.Set)
	require.True(t, ok)
	assert.True(t, planned.Equal(stored), "the stored set must equal the planned set, otherwise the consistency check fails")
}
