package resource_exporter

import (
	"fmt"
	"os"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitSanitizeResourceBlockLabel(t *testing.T) {
	testCases := []struct {
		name                 string
		input                string
		expectedOriginal     string
		expectedOptimized    string
		expectedBCPOptimized string
	}{
		{
			name:                 "basic string",
			input:                "test label",
			expectedOriginal:     "test_label",
			expectedOptimized:    "test_label",
			expectedBCPOptimized: "test_label",
		},
		{
			name:                 "string with special characters",
			input:                "test@label#123",
			expectedOriginal:     "test_label_123",
			expectedOptimized:    "test_label_123",
			expectedBCPOptimized: "test_label_123",
		},
		{
			name:                 "starts with number",
			input:                "123test",
			expectedOriginal:     "_123test",
			expectedOptimized:    "_123test",
			expectedBCPOptimized: "_123test",
		},
		{
			name:                 "non-latin characters",
			input:                "テスト label",
			expectedOriginal:     "____label",
			expectedOptimized:    "tesuto_label",
			expectedBCPOptimized: "tesuto_label",
		},
		{
			name:                 "empty string",
			input:                "",
			expectedOriginal:     "",
			expectedOptimized:    "",
			expectedBCPOptimized: "",
		},
		{
			name:                 "whitespace only",
			input:                "   ",
			expectedOriginal:     "___",
			expectedOptimized:    "",
			expectedBCPOptimized: "",
		},
		{
			name:                 "mixed case with special chars",
			input:                "Test@Label_123",
			expectedOriginal:     "Test_Label_123",
			expectedOptimized:    "Test_Label_123",
			expectedBCPOptimized: "Test_Label_123",
		},
		{
			name:                 "multiple consecutive special chars",
			input:                "test@@##$$label",
			expectedOriginal:     "test______label",
			expectedOptimized:    "test______label",
			expectedBCPOptimized: "test______label",
		},
		{
			name:                 "dots and dashes",
			input:                "test.label-123",
			expectedOriginal:     "test_label-123",
			expectedOptimized:    "test_label-123",
			expectedBCPOptimized: "test_label-123",
		},
	}

	for _, tc := range testCases {
		t.Run("Original/"+tc.name, func(t *testing.T) {
			sanitizer := &sanitizerOriginal{}
			result := sanitizer.SanitizeResourceBlockLabel(tc.input)
			assert.Equal(t, tc.expectedOriginal, result)
		})

		t.Run("Optimized/"+tc.name, func(t *testing.T) {
			sanitizer := &sanitizerOptimized{}
			result := sanitizer.SanitizeResourceBlockLabel(tc.input)
			assert.Equal(t, tc.expectedOptimized, result)
		})

		t.Run("BCP Optimized/"+tc.name, func(t *testing.T) {
			sanitizer := &sanitizerBCPOptimized{}
			result := sanitizer.SanitizeResourceBlockLabel(tc.input)
			assert.Equal(t, tc.expectedBCPOptimized, result)
		})
	}
}
func TestUnitSanitize(t *testing.T) {
	testCases := []struct {
		name                 string
		input                ResourceIDMetaMap
		validateOriginal     func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerOriginal)
		validateOptimized    func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerOptimized)
		validateBCPOptimized func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerBCPOptimized)
	}{
		{
			name: "unique labels",
			input: ResourceIDMetaMap{
				"id1": &ResourceMeta{BlockLabel: "test1"},
				"id2": &ResourceMeta{BlockLabel: "test2"},
			},
			validateOriginal: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerOriginal) {
				assert.Equal(t, "test1", result["id1"].BlockLabel)
				assert.Equal(t, "test2", result["id2"].BlockLabel)
			},
			validateOptimized: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerOptimized) {
				assert.Equal(t, "test1", result["id1"].BlockLabel)
				assert.Equal(t, "test2", result["id2"].BlockLabel)
			},
			validateBCPOptimized: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerBCPOptimized) {
				assert.True(t, strings.HasPrefix(result["id1"].BlockLabel, "test1_"))
				assert.True(t, strings.HasPrefix(result["id2"].BlockLabel, "test2_"))
				// Hash is always appended for BCP Optimized
				assert.True(t, len(result["id1"].BlockLabel) > len("test1_"))
				assert.True(t, len(result["id2"].BlockLabel) > len("test2_"))
			},
		},
		{
			// Labels that have matching block labels (should never/rarely happen)
			// are handled differently across the sanitizers
			name: "duplicate block labels",
			input: ResourceIDMetaMap{
				"id1": &ResourceMeta{BlockLabel: "testfoo"},
				"id2": &ResourceMeta{BlockLabel: "testfoo"},
				"id3": &ResourceMeta{BlockLabel: "testfoo"},
			},
			// Never appends a hash to all of the labels that have the same original BlockLabel.
			// This can cause issues with mixing up the state file unfortunately, and require
			// extra logic in the buildResourceConfigMap() function to check for this.
			// See DEVTOOLING-1183 for ideas on improving this
			validateOriginal: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerOriginal) {
				assert.Equal(t, result["id1"].BlockLabel, result["id2"].BlockLabel)
				assert.Equal(t, result["id1"].BlockLabel, result["id3"].BlockLabel)
				assert.Equal(t, result["id2"].BlockLabel, result["id3"].BlockLabel)
				assert.True(t, strings.HasPrefix(result["id1"].BlockLabel, "testfoo"))
				assert.True(t, strings.HasPrefix(result["id2"].BlockLabel, "testfoo"))
				assert.True(t, strings.HasPrefix(result["id3"].BlockLabel, "testfoo"))
				assert.True(t, len(result["id1"].BlockLabel) == 7) // Hash NOT appended
				assert.True(t, len(result["id2"].BlockLabel) == 7) // Hash NOT appended
				assert.True(t, len(result["id3"].BlockLabel) == 7) // Hash NOT appended
			},
			// Never appends a hash to all of the labels that have the same original BlockLabel.
			// This can cause issues with mixing up the state file unfortunately, and require
			// extra logic in the buildResourceConfigMap() function to check for this
			// See DEVTOOLING-1183 for ideas on improving this
			validateOptimized: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerOptimized) {
				assert.Equal(t, result["id1"].BlockLabel, result["id2"].BlockLabel)
				assert.Equal(t, result["id1"].BlockLabel, result["id3"].BlockLabel)
				assert.Equal(t, result["id2"].BlockLabel, result["id3"].BlockLabel)
				assert.True(t, strings.HasPrefix(result["id1"].BlockLabel, "testfoo"))
				assert.True(t, strings.HasPrefix(result["id2"].BlockLabel, "testfoo"))
				assert.True(t, strings.HasPrefix(result["id2"].BlockLabel, "testfoo"))
				assert.True(t, len(result["id1"].BlockLabel) == 7) // Hash NOT appended
				assert.True(t, len(result["id2"].BlockLabel) == 7) // Hash NOT appended
				assert.True(t, len(result["id3"].BlockLabel) == 7) // Hash NOT appended
			},
			// Appends a hash to every label processed to ensures consistency so that output
			// is more consistent across runs and between export comparisons across orgs.
			// A _DUPLICATE_INSTANCE_# value is appended to alert on this rare edge case
			// See DEVTOOLING-1183 for ideas on improving this
			validateBCPOptimized: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerBCPOptimized) {
				assert.NotEqual(t, result["id1"].BlockLabel, result["id2"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id3"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id3"].BlockLabel)
				assert.True(t, strings.HasPrefix(result["id1"].BlockLabel, "testfoo_"))
				assert.True(t, strings.HasPrefix(result["id2"].BlockLabel, "testfoo_"))
				assert.True(t, strings.HasPrefix(result["id3"].BlockLabel, "testfoo_"))
				assert.True(t, len(result["id1"].BlockLabel) > 8) // Hash appended
				assert.True(t, len(result["id2"].BlockLabel) > 8) // Hash appended
				assert.True(t, len(result["id3"].BlockLabel) > 8) // Hash appended
				id1Hash := sanitizer.SanitizeResourceHash(result["id1"].OriginalLabel)
				id2Hash := sanitizer.SanitizeResourceHash(result["id2"].OriginalLabel)
				id3Hash := sanitizer.SanitizeResourceHash(result["id3"].OriginalLabel)
				assert.True(t, strings.Contains(result["id1"].BlockLabel, id1Hash))                // Hash included
				assert.True(t, strings.Contains(result["id2"].BlockLabel, id2Hash))                // Hash included
				assert.True(t, strings.Contains(result["id3"].BlockLabel, id3Hash))                // Hash included
				assert.True(t, strings.Contains(result["id1"].BlockLabel, "_DUPLICATE_INSTANCE_")) // Duplicate appended
				assert.True(t, strings.Contains(result["id2"].BlockLabel, "_DUPLICATE_INSTANCE_")) // Duplicate appended
				assert.True(t, strings.Contains(result["id3"].BlockLabel, "_DUPLICATE_INSTANCE_")) // Duplicate appended
			},
		},
		{
			// Labels that have been sanitized and match other labels are handled differently
			// across the sanitizers
			name: "duplicate sanitized labels",
			input: ResourceIDMetaMap{
				"id1": &ResourceMeta{BlockLabel: "test.user@foo.com"},
				"id2": &ResourceMeta{BlockLabel: "test_user@foo.com"},
				"id3": &ResourceMeta{BlockLabel: "test+user@foo.com"},
				"id4": &ResourceMeta{BlockLabel: "test&user@foo.com"},
			},
			// Always appends a hash to all of the labels that match each other
			validateOriginal: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerOriginal) {
				assert.NotEqual(t, result["id1"].BlockLabel, result["id2"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id3"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id4"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id3"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id4"].BlockLabel)
				assert.NotEqual(t, result["id3"].BlockLabel, result["id4"].BlockLabel)
				assert.True(t, strings.HasPrefix(result["id1"].BlockLabel, "test_user_foo_com_"))
				assert.True(t, strings.HasPrefix(result["id2"].BlockLabel, "test_user_foo_com_"))
				assert.True(t, strings.HasPrefix(result["id3"].BlockLabel, "test_user_foo_com_"))
				assert.True(t, strings.HasPrefix(result["id4"].BlockLabel, "test_user_foo_com_"))
				assert.True(t, len(result["id1"].BlockLabel) > len("test_user_foo_com_")) // Hash appended
				assert.True(t, len(result["id2"].BlockLabel) > len("test_user_foo_com_")) // Hash appended
				assert.True(t, len(result["id4"].BlockLabel) > len("test_user_foo_com_")) // Hash appended
				assert.True(t, len(result["id4"].BlockLabel) > len("test_user_foo_com_")) // Hash appended
				id1Hash := sanitizer.SanitizeResourceHash(result["id1"].OriginalLabel)
				id2Hash := sanitizer.SanitizeResourceHash(result["id2"].OriginalLabel)
				id3Hash := sanitizer.SanitizeResourceHash(result["id3"].OriginalLabel)
				id4Hash := sanitizer.SanitizeResourceHash(result["id4"].OriginalLabel)
				assert.True(t, strings.Contains(result["id1"].BlockLabel, id1Hash)) // Hash included
				assert.True(t, strings.Contains(result["id2"].BlockLabel, id2Hash)) // Hash included
				assert.True(t, strings.Contains(result["id3"].BlockLabel, id3Hash)) // Hash included
				assert.True(t, strings.Contains(result["id4"].BlockLabel, id4Hash)) // Hash included
			},
			// Appends a hash to every other label other than the first label that is processed
			// Unfortunately this is not consistent across exports, as the first label is never the same
			// from export to export. This is probably fine for any regular exports. See DEVTOOLING-1182
			validateOptimized: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerOptimized) {
				assert.NotEqual(t, result["id1"].BlockLabel, result["id2"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id3"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id4"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id3"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id4"].BlockLabel)
				assert.NotEqual(t, result["id3"].BlockLabel, result["id4"].BlockLabel)
				assert.True(t, strings.HasPrefix(result["id1"].BlockLabel, "test_user_foo_com"))
				assert.True(t, strings.HasPrefix(result["id2"].BlockLabel, "test_user_foo_com"))
				assert.True(t, strings.HasPrefix(result["id3"].BlockLabel, "test_user_foo_com"))
				assert.True(t, strings.HasPrefix(result["id4"].BlockLabel, "test_user_foo_com"))
				assert.True(t, len(result["id1"].BlockLabel) >= len("test_user_foo_com")) // Hash maybe appended
				assert.True(t, len(result["id2"].BlockLabel) >= len("test_user_foo_com")) // Hash maybe appended
				assert.True(t, len(result["id3"].BlockLabel) >= len("test_user_foo_com")) // Hash maybe appended
				assert.True(t, len(result["id4"].BlockLabel) >= len("test_user_foo_com")) // Hash maybe appended
				// Assert only one has no hash appended
				lengthOfBlockLabelCount := 0
				for _, meta := range result {
					if len(meta.BlockLabel) == len("test_user_foo_com") {
						lengthOfBlockLabelCount++
					} else {
						idHash := sanitizer.SanitizeResourceHash(meta.OriginalLabel)
						assert.True(t, strings.Contains(meta.BlockLabel, idHash))
					}
				}
				assert.Equal(t, 1, lengthOfBlockLabelCount, "Exactly one Block Label should have no hash appended")
			},
			// Appends a hash to every label processed to ensures consistency so that output
			// is more consistent across runs and between export comparisons across orgs. See DEVTOOLING-1182
			validateBCPOptimized: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerBCPOptimized) {
				assert.NotEqual(t, result["id1"].BlockLabel, result["id2"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id3"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id4"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id3"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id4"].BlockLabel)
				assert.NotEqual(t, result["id3"].BlockLabel, result["id4"].BlockLabel)
				assert.True(t, strings.HasPrefix(result["id1"].BlockLabel, "test_user_foo_com_"))
				assert.True(t, strings.HasPrefix(result["id2"].BlockLabel, "test_user_foo_com_"))
				assert.True(t, strings.HasPrefix(result["id3"].BlockLabel, "test_user_foo_com_"))
				assert.True(t, strings.HasPrefix(result["id4"].BlockLabel, "test_user_foo_com_"))
				assert.True(t, len(result["id1"].BlockLabel) > len("test_user_foo_com_")) // Hash appended
				assert.True(t, len(result["id2"].BlockLabel) > len("test_user_foo_com_")) // Hash appended
				assert.True(t, len(result["id3"].BlockLabel) > len("test_user_foo_com_")) // Hash appended
				assert.True(t, len(result["id4"].BlockLabel) > len("test_user_foo_com_")) // Hash appended
				id1Hash := sanitizer.SanitizeResourceHash(result["id1"].OriginalLabel)
				id2Hash := sanitizer.SanitizeResourceHash(result["id2"].OriginalLabel)
				id3Hash := sanitizer.SanitizeResourceHash(result["id3"].OriginalLabel)
				id4Hash := sanitizer.SanitizeResourceHash(result["id4"].OriginalLabel)
				assert.True(t, strings.Contains(result["id1"].BlockLabel, id1Hash))                 // Hash included
				assert.True(t, strings.Contains(result["id2"].BlockLabel, id2Hash))                 // Hash included
				assert.True(t, strings.Contains(result["id3"].BlockLabel, id3Hash))                 // Hash included
				assert.True(t, strings.Contains(result["id4"].BlockLabel, id4Hash))                 // Hash included
				assert.False(t, strings.Contains(result["id1"].BlockLabel, "_DUPLICATE_INSTANCE_")) // Duplicate NOT appended
				assert.False(t, strings.Contains(result["id2"].BlockLabel, "_DUPLICATE_INSTANCE_")) // Duplicate NOT appended
				assert.False(t, strings.Contains(result["id3"].BlockLabel, "_DUPLICATE_INSTANCE_")) // Duplicate NOT appended
				assert.False(t, strings.Contains(result["id4"].BlockLabel, "_DUPLICATE_INSTANCE_")) // Duplicate NOT appended

			},
		},
		{
			name: "duplicates with block hash",
			input: ResourceIDMetaMap{
				"id1": &ResourceMeta{BlockLabel: "test_distinct_user@foo.com", BlockHash: "abc1234"}, // Distinct block label, distinct block hash
				"id2": &ResourceMeta{BlockLabel: "test_distinct_user2@foo.com"},                      // Distinct block label, no block hash
				"id3": &ResourceMeta{BlockLabel: "test_user@foo.com", BlockHash: "abc1234"},          // Same sanitized block label as next, distinct block hash
				"id4": &ResourceMeta{BlockLabel: "test.user@foo.com", BlockHash: "bcd5678"},          // Same block label as next, distinct block hash
				"id5": &ResourceMeta{BlockLabel: "test.user@foo.com", BlockHash: "cde9001"},          // Same block label as prev, distinct block hash
				"id6": &ResourceMeta{BlockLabel: "test+user@foo.com", BlockHash: "def2342"},          // Same block label as next, same block hash
				"id7": &ResourceMeta{BlockLabel: "test+user@foo.com", BlockHash: "def2342"},          // Same block label as prev, same block hash
			},
			// Ignores the BlockHash (for now). This may change in the future.
			// Never append a hash to original labels that distinct or are the same as another.
			// Always appends a hash to all of the labels that were different but match each other directly after being sanitized.
			validateOriginal: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerOriginal) {
				assert.NotEqual(t, result["id1"].BlockLabel, result["id2"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id3"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id4"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id5"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id6"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id7"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id3"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id4"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id5"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id6"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id7"].BlockLabel)
				assert.NotEqual(t, result["id3"].BlockLabel, result["id4"].BlockLabel)
				assert.NotEqual(t, result["id3"].BlockLabel, result["id5"].BlockLabel)
				assert.NotEqual(t, result["id3"].BlockLabel, result["id6"].BlockLabel)
				assert.NotEqual(t, result["id3"].BlockLabel, result["id7"].BlockLabel)
				assert.Equal(t, result["id4"].BlockLabel, result["id5"].BlockLabel)
				assert.NotEqual(t, result["id4"].BlockLabel, result["id6"].BlockLabel)
				assert.NotEqual(t, result["id4"].BlockLabel, result["id7"].BlockLabel)
				assert.NotEqual(t, result["id5"].BlockLabel, result["id6"].BlockLabel)
				assert.NotEqual(t, result["id5"].BlockLabel, result["id7"].BlockLabel)
				assert.Equal(t, result["id6"].BlockLabel, result["id7"].BlockLabel)
				assert.Equal(t, result["id1"].BlockLabel, "test_distinct_user_foo_com")
				assert.Equal(t, result["id2"].BlockLabel, "test_distinct_user2_foo_com")
				assert.True(t, strings.HasPrefix(result["id3"].BlockLabel, "test_user_foo_com_"))
				assert.True(t, strings.HasPrefix(result["id4"].BlockLabel, "test_user_foo_com_"))
				assert.True(t, strings.HasPrefix(result["id5"].BlockLabel, "test_user_foo_com_"))
				assert.True(t, strings.HasPrefix(result["id6"].BlockLabel, "test_user_foo_com_"))
				assert.True(t, strings.HasPrefix(result["id7"].BlockLabel, "test_user_foo_com_"))
				assert.True(t, len(result["id3"].BlockLabel) > len("test_user_foo_com_")) // Hash appended
				assert.True(t, len(result["id4"].BlockLabel) > len("test_user_foo_com_")) // Hash appended
				assert.True(t, len(result["id5"].BlockLabel) > len("test_user_foo_com_")) // Hash appended
				assert.True(t, len(result["id6"].BlockLabel) > len("test_user_foo_com_")) // Hash appended
				assert.True(t, len(result["id7"].BlockLabel) > len("test_user_foo_com_")) // Hash appended
				id3Hash := sanitizer.SanitizeResourceHash(result["id3"].OriginalLabel)
				id4Hash := sanitizer.SanitizeResourceHash(result["id4"].OriginalLabel)
				id5Hash := sanitizer.SanitizeResourceHash(result["id5"].OriginalLabel)
				id6Hash := sanitizer.SanitizeResourceHash(result["id6"].OriginalLabel)
				id7Hash := sanitizer.SanitizeResourceHash(result["id7"].OriginalLabel)
				assert.True(t, strings.HasSuffix(result["id3"].BlockLabel, id3Hash)) // Hash included
				assert.True(t, strings.HasSuffix(result["id4"].BlockLabel, id4Hash)) // Hash included
				assert.True(t, strings.HasSuffix(result["id5"].BlockLabel, id5Hash)) // Hash included
				assert.True(t, strings.HasSuffix(result["id6"].BlockLabel, id6Hash)) // Hash included
				assert.True(t, strings.HasSuffix(result["id7"].BlockLabel, id7Hash)) // Hash included
			},
			// Ignores the BlockHash (for now). This may change in the future.
			// Appends a hash to every other label other than the first label that is processed
			// Unfortunately this is not consistent across exports, as the first label is never the same
			// from export to export. This is probably fine for any regular exports. See DEVTOOLING-1182
			validateOptimized: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerOptimized) {
				assert.NotEqual(t, result["id1"].BlockLabel, result["id2"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id3"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id4"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id5"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id6"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id7"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id3"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id4"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id5"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id6"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id7"].BlockLabel)
				assert.NotEqual(t, result["id3"].BlockLabel, result["id4"].BlockLabel)
				assert.NotEqual(t, result["id3"].BlockLabel, result["id5"].BlockLabel)
				assert.NotEqual(t, result["id3"].BlockLabel, result["id6"].BlockLabel)
				assert.NotEqual(t, result["id3"].BlockLabel, result["id7"].BlockLabel)
				// Intentionally commented out, as this could be equal (same hash) or
				// not equal (one without the hash). Retained for consistency purposes.
				// assert.NotEqual(t, result["id4"].BlockLabel, result["id5"].BlockLabel)
				assert.NotEqual(t, result["id4"].BlockLabel, result["id6"].BlockLabel)
				assert.NotEqual(t, result["id4"].BlockLabel, result["id7"].BlockLabel)
				assert.NotEqual(t, result["id5"].BlockLabel, result["id6"].BlockLabel)
				assert.NotEqual(t, result["id5"].BlockLabel, result["id7"].BlockLabel)
				// Intentionally commented out, as this could be equal (same hash) or
				// not equal (one without the hash). Retained for consistency purposes.
				// assert.NotEqual(t, result["id6"].BlockLabel, result["id7"].BlockLabel)
				assert.Equal(t, result["id1"].BlockLabel, "test_distinct_user_foo_com")
				assert.Equal(t, result["id2"].BlockLabel, "test_distinct_user2_foo_com")
				assert.True(t, strings.HasPrefix(result["id3"].BlockLabel, "test_user_foo_com"))
				assert.True(t, strings.HasPrefix(result["id4"].BlockLabel, "test_user_foo_com"))
				assert.True(t, strings.HasPrefix(result["id5"].BlockLabel, "test_user_foo_com"))
				assert.True(t, strings.HasPrefix(result["id6"].BlockLabel, "test_user_foo_com"))
				assert.True(t, strings.HasPrefix(result["id7"].BlockLabel, "test_user_foo_com"))
				assert.True(t, len(result["id3"].BlockLabel) >= len("test_user_foo_com")) // Hash maybe appended
				assert.True(t, len(result["id4"].BlockLabel) >= len("test_user_foo_com")) // Hash maybe appended
				assert.True(t, len(result["id5"].BlockLabel) >= len("test_user_foo_com")) // Hash maybe appended
				assert.True(t, len(result["id6"].BlockLabel) >= len("test_user_foo_com")) // Hash maybe appended
				assert.True(t, len(result["id7"].BlockLabel) >= len("test_user_foo_com")) // Hash maybe appended
				// Remove distinct results since we've already validated them
				delete(result, "id1")
				delete(result, "id2")
				// Loop over the rest of matching results asserting only one has no hash appended
				lengthOfBlockLabelCount := 0
				for _, meta := range result {
					if len(meta.BlockLabel) == len("test_user_foo_com") {
						lengthOfBlockLabelCount++
					} else {
						hash := sanitizer.SanitizeResourceHash(meta.OriginalLabel)
						assert.True(t, strings.HasSuffix(meta.BlockLabel, hash))
					}
				}
				assert.Equal(t, 1, lengthOfBlockLabelCount, "Exactly one Block Label should have no hash appended")
			},
			// Always includes a hash of the original BlockLabel to every label processed. Also checks for duplicates after
			// appending the hash of the original BlockLabel and if so adds a BlockHash (if available) to append the hash
			// content to the label to create distinct labels to ensures consistency so that output is more consistent across
			// runs and between export comparisons across orgs. See DEVTOOLING-1182 and DEVTOOLING-1183.
			validateBCPOptimized: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerBCPOptimized) {
				assert.NotEqual(t, result["id1"].BlockLabel, result["id2"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id3"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id4"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id5"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id6"].BlockLabel)
				assert.NotEqual(t, result["id1"].BlockLabel, result["id7"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id3"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id4"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id5"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id6"].BlockLabel)
				assert.NotEqual(t, result["id2"].BlockLabel, result["id7"].BlockLabel)
				assert.NotEqual(t, result["id3"].BlockLabel, result["id4"].BlockLabel)
				assert.NotEqual(t, result["id3"].BlockLabel, result["id5"].BlockLabel)
				assert.NotEqual(t, result["id3"].BlockLabel, result["id6"].BlockLabel)
				assert.NotEqual(t, result["id3"].BlockLabel, result["id7"].BlockLabel)
				assert.NotEqual(t, result["id4"].BlockLabel, result["id5"].BlockLabel)
				assert.NotEqual(t, result["id4"].BlockLabel, result["id6"].BlockLabel)
				assert.NotEqual(t, result["id4"].BlockLabel, result["id7"].BlockLabel)
				assert.NotEqual(t, result["id5"].BlockLabel, result["id6"].BlockLabel)
				assert.NotEqual(t, result["id5"].BlockLabel, result["id7"].BlockLabel)
				assert.NotEqual(t, result["id6"].BlockLabel, result["id7"].BlockLabel)
				id1Hash := sanitizer.SanitizeResourceHash(result["id1"].OriginalLabel)
				id2Hash := sanitizer.SanitizeResourceHash(result["id2"].OriginalLabel)
				id3Hash := sanitizer.SanitizeResourceHash(result["id3"].OriginalLabel)
				id4Hash := sanitizer.SanitizeResourceHash(result["id4"].OriginalLabel)
				id5Hash := sanitizer.SanitizeResourceHash(result["id5"].OriginalLabel)
				id6Hash := sanitizer.SanitizeResourceHash(result["id6"].OriginalLabel)
				id7Hash := sanitizer.SanitizeResourceHash(result["id7"].OriginalLabel)

				assert.NotEqual(t, id3Hash, id4Hash)
				assert.NotEqual(t, id3Hash, id5Hash)
				assert.NotEqual(t, id3Hash, id6Hash)
				assert.NotEqual(t, id3Hash, id7Hash)
				assert.Equal(t, id4Hash, id5Hash)
				assert.NotEqual(t, id4Hash, id6Hash)
				assert.NotEqual(t, id4Hash, id7Hash)
				assert.NotEqual(t, id5Hash, id6Hash)
				assert.NotEqual(t, id5Hash, id7Hash)
				assert.Equal(t, id6Hash, id7Hash)

				assert.Equal(t, result["id1"].BlockLabel, fmt.Sprintf("test_distinct_user_foo_com__BLH%s", id1Hash))
				assert.Equal(t, result["id2"].BlockLabel, fmt.Sprintf("test_distinct_user2_foo_com__BLH%s", id2Hash))
				assert.Equal(t, result["id3"].BlockLabel, fmt.Sprintf("test_user_foo_com__BLH%s", id3Hash))
				assert.Equal(t, result["id4"].BlockLabel, fmt.Sprintf("test_user_foo_com__BLH%s_UFH%s", id4Hash, result["id4"].BlockHash))
				assert.Equal(t, result["id5"].BlockLabel, fmt.Sprintf("test_user_foo_com__BLH%s_UFH%s", id5Hash, result["id5"].BlockHash))
				assert.True(t, strings.HasPrefix(result["id6"].BlockLabel, fmt.Sprintf("test_user_foo_com__BLH%s_UFH%s", id6Hash, result["id6"].BlockHash)))
				assert.True(t, strings.HasPrefix(result["id7"].BlockLabel, fmt.Sprintf("test_user_foo_com__BLH%s_UFH%s", id6Hash, result["id7"].BlockHash)))
				assert.False(t, strings.Contains(result["id1"].BlockLabel, "_DUPLICATE_INSTANCE_")) // Duplicate NOT appended
				assert.False(t, strings.Contains(result["id2"].BlockLabel, "_DUPLICATE_INSTANCE_")) // Duplicate NOT appended
				assert.False(t, strings.Contains(result["id3"].BlockLabel, "_DUPLICATE_INSTANCE_")) // Duplicate NOT appended
				assert.False(t, strings.Contains(result["id4"].BlockLabel, "_DUPLICATE_INSTANCE_")) // Duplicate NOT appended
				assert.False(t, strings.Contains(result["id5"].BlockLabel, "_DUPLICATE_INSTANCE_")) // Duplicate NOT appended
				assert.True(t, strings.Contains(result["id6"].BlockLabel, "_DUPLICATE_INSTANCE_"))  // Duplicate IS appended
				assert.True(t, strings.Contains(result["id7"].BlockLabel, "_DUPLICATE_INSTANCE_"))  // Duplicate IS appended

			},
		},
		{
			name: "non-latin characters",
			input: ResourceIDMetaMap{
				"id1": &ResourceMeta{BlockLabel: "テスト1"},
				"id2": &ResourceMeta{BlockLabel: "テスト2"},
			},
			validateOriginal: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerOriginal) {
				assert.Contains(t, result["id1"].BlockLabel, "___1")
				assert.Contains(t, result["id2"].BlockLabel, "___2")
				assert.NotEqual(t, result["id1"].BlockLabel, result["id2"].BlockLabel)
			},
			validateOptimized: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerOptimized) {
				assert.Contains(t, result["id1"].BlockLabel, "tesuto1")
				assert.Contains(t, result["id2"].BlockLabel, "tesuto2")
				assert.NotEqual(t, result["id1"].BlockLabel, result["id2"].BlockLabel)
			},
			validateBCPOptimized: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerBCPOptimized) {
				assert.Contains(t, result["id1"].BlockLabel, "tesuto1")
				assert.Contains(t, result["id2"].BlockLabel, "tesuto2")
				assert.NotEqual(t, result["id1"].BlockLabel, result["id2"].BlockLabel)
			},
		},
		{
			name: "special characters with numbers",
			input: ResourceIDMetaMap{
				"id1": &ResourceMeta{BlockLabel: "123@test"},
				"id2": &ResourceMeta{BlockLabel: "456#test"},
			},
			validateOriginal: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerOriginal) {
				assert.Equal(t, result["id1"].BlockLabel, "_123_test")
				assert.Equal(t, result["id2"].BlockLabel, "_456_test")
			},
			validateOptimized: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerOptimized) {
				assert.Equal(t, result["id1"].BlockLabel, "_123_test")
				assert.Equal(t, result["id2"].BlockLabel, "_456_test")
			},
			validateBCPOptimized: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerBCPOptimized) {
				assert.True(t, strings.HasPrefix(result["id1"].BlockLabel, "_123_test_"))
				assert.True(t, strings.HasPrefix(result["id2"].BlockLabel, "_456_test_"))
			},
		},
		{
			name: "mixed case with special chars",
			input: ResourceIDMetaMap{
				"id1": &ResourceMeta{BlockLabel: "Test@Label_123"},
				"id2": &ResourceMeta{BlockLabel: "Test#Label_456"},
			},
			validateOriginal: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerOriginal) {
				assert.Equal(t, result["id1"].BlockLabel, "Test_Label_123")
				assert.Equal(t, result["id2"].BlockLabel, "Test_Label_456")
				assert.NotEqual(t, result["id1"].BlockLabel, result["id2"].BlockLabel)
			},
			validateOptimized: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerOptimized) {
				assert.Equal(t, result["id1"].BlockLabel, "Test_Label_123")
				assert.Equal(t, result["id2"].BlockLabel, "Test_Label_456")
				assert.NotEqual(t, result["id1"].BlockLabel, result["id2"].BlockLabel)
			},
			validateBCPOptimized: func(t *testing.T, result ResourceIDMetaMap, sanitizer sanitizerBCPOptimized) {
				assert.True(t, strings.HasPrefix(result["id1"].BlockLabel, "Test_Label_123"))
				assert.True(t, strings.HasPrefix(result["id2"].BlockLabel, "Test_Label_456"))
				assert.NotEqual(t, result["id1"].BlockLabel, result["id2"].BlockLabel)
			},
		},
	}

	// Test each sanitizer
	t.Run("Original", func(t *testing.T) {
		sanitizer := &sanitizerOriginal{}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				inputCopy := makeInputCopy(tc.input)
				sanitizer.Sanitize(inputCopy)
				tc.validateOriginal(t, inputCopy, *sanitizer)
			})
		}
	})

	t.Run("Optimized", func(t *testing.T) {
		sanitizer := &sanitizerOptimized{}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				inputCopy := makeInputCopy(tc.input)
				sanitizer.Sanitize(inputCopy)
				tc.validateOptimized(t, inputCopy, *sanitizer)
			})
		}
	})

	t.Run("BCP Optimized", func(t *testing.T) {
		sanitizer := &sanitizerBCPOptimized{}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				inputCopy := makeInputCopy(tc.input)
				sanitizer.Sanitize(inputCopy)
				tc.validateBCPOptimized(t, inputCopy, *sanitizer)
			})
		}
	})
}

// Helper function to create a deep copy of the input map
func makeInputCopy(input ResourceIDMetaMap) ResourceIDMetaMap {
	inputCopy := make(ResourceIDMetaMap)
	for k, v := range input {
		inputCopy[k] = &ResourceMeta{
			BlockLabel:    v.BlockLabel,
			OriginalLabel: v.OriginalLabel,
			BlockHash:     v.BlockHash,
		}
	}
	return inputCopy
}

func TestUnitNewSanitizerProvider(t *testing.T) {
	// Test with default settings (no environment variable)
	provider := NewSanitizerProvider()
	assert.IsType(t, &sanitizerOriginal{}, provider.S)

	// Test by setting optimized variable
	os.Setenv(feature_toggles.ExporterSanitizerOptimizedName(), "true")
	provider = NewSanitizerProvider()
	assert.IsType(t, &sanitizerOptimized{}, provider.S)
	os.Unsetenv(feature_toggles.ExporterSanitizerOptimizedName())

	// Test by setting BCP optimized variable
	os.Setenv(feature_toggles.ExporterSanitizerBCPOptimizedName(), "true")
	provider = NewSanitizerProvider()
	assert.IsType(t, &sanitizerBCPOptimized{}, provider.S)
	os.Unsetenv(feature_toggles.ExporterSanitizerBCPOptimizedName())
}

func TestUnitOriginalLabelPreservation(t *testing.T) {
	testCases := []struct {
		name      string
		sanitizer Sanitizer
		input     ResourceIDMetaMap
	}{
		{
			name:      "Original Sanitizer",
			sanitizer: &sanitizerOriginal{},
			input: ResourceIDMetaMap{
				"id1": &ResourceMeta{BlockLabel: "test@1"},
			},
		},
		{
			name:      "Optimized Sanitizer",
			sanitizer: &sanitizerOptimized{},
			input: ResourceIDMetaMap{
				"id1": &ResourceMeta{BlockLabel: "test@1"},
			},
		},
		{
			name:      "BCP Optimized Sanitizer",
			sanitizer: &sanitizerBCPOptimized{},
			input: ResourceIDMetaMap{
				"id1": &ResourceMeta{BlockLabel: "test@1"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalLabel := tc.input["id1"].BlockLabel
			tc.sanitizer.Sanitize(tc.input)
			assert.Equal(t, originalLabel, tc.input["id1"].OriginalLabel)
			assert.NotEqual(t, originalLabel, tc.input["id1"].BlockLabel)
		})
	}
}
