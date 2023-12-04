package stringmap

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestUnitMergeMaps(t *testing.T) {
	m1 := map[string][]int{
		"key1": {1, 2, 3},
		"key2": {4, 5, 6},
	}

	m2 := map[string][]int{
		"key3": {7, 8, 9},
		"key4": {10, 11, 12},
	}

	// Check if the result contains all the keys from m1 and m2
	expected := map[string][]int{
		"key1": {1, 2, 3},
		"key2": {4, 5, 6},
		"key3": {7, 8, 9},
		"key4": {10, 11, 12},
	}

	result := MergeMaps(m1, m2)

	// Compare the expected result with the actual result
	if diff := cmp.Diff(result, expected); diff != "" {
		t.Errorf("Unexpected result (-want +got):\n%s", diff)
	}
}

func TestUnitMergeMapsStrings(t *testing.T) {
	m1 := map[string][]string{
		"key1": {"1", "2", "3"},
		"key2": {"4", "5", "6"},
	}

	m2 := map[string][]string{
		"key3": {},
		"key4": {},
	}

	expected := map[string][]string{
		"key1": {"1", "2", "3"},
		"key2": {"4", "5", "6"},
		"key3": {},
		"key4": {},
	}

	result := MergeMaps(m1, m2)

	// Compare the expected result with the actual result
	if diff := cmp.Diff(result, expected); diff != "" {
		t.Errorf("Unexpected result (-want +got):\n%s", diff)
	}
}
