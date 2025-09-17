package resource_exporter

import (
	"testing"

	"github.com/google/uuid"
)

/*
This test file specifically tests the RemoveIfMissing and RemoveIfSelfReferential fields
of the ResourceExporter struct.

RemoveIfMissing: A map of attributes to a list of inner object attributes.
When all specified inner attributes are missing from an object, that object is removed.

RemoveIfSelfReferential: A list of attributes that should be removed from the export config
if the value matches the id of the resource.

These features are used during Terraform resource export to clean up configuration
by removing unnecessary or self-referential attributes.

IMPORTANT: This test directly calls the actual ResourceExporter methods
(RemoveFieldIfMissing and RemoveFieldIfSelfReferential) rather than simulating
the behavior, ensuring we test the real implementation.
*/

// TestUnitRemoveIfMissing tests the RemoveIfMissing functionality.
// This feature removes entire objects from the configuration when all specified
// required attributes are missing from those objects.
// For example, if an outbound_email_address object is missing its route_id,
// the entire outbound_email_address object should be removed from the export.
func TestUnitRemoveIfMissing(t *testing.T) {
	tests := []struct {
		name            string
		removeIfMissing map[string][]string
		attribute       string
		config          map[string]interface{}
		expectedRemove  bool
	}{
		{
			name: "Remove when all specified attributes are missing",
			removeIfMissing: map[string][]string{
				"outbound_email_address": {"route_id"},
			},
			attribute: "outbound_email_address",
			config: map[string]interface{}{
				"domain_id": "test-domain",
				// route_id is missing
			},
			expectedRemove: true,
		},
		{
			name: "Keep when required attributes are present",
			removeIfMissing: map[string][]string{
				"outbound_email_address": {"route_id"},
			},
			attribute: "outbound_email_address",
			config: map[string]interface{}{
				"domain_id": "test-domain",
				"route_id":  "test-route-id",
			},
			expectedRemove: false,
		},
		{
			name: "Remove when required attribute is nil",
			removeIfMissing: map[string][]string{
				"members": {"user_id"},
			},
			attribute: "members",
			config: map[string]interface{}{
				"ring_num": 1,
				"user_id":  nil, // user_id is nil
			},
			expectedRemove: true,
		},
		{
			name: "Keep when all required attributes are present",
			removeIfMissing: map[string][]string{
				"test_object": {"attr1", "attr2"},
			},
			attribute: "test_object",
			config: map[string]interface{}{
				"attr1": "value1",
				"attr2": "value2",
				"attr3": "value3",
			},
			expectedRemove: false,
		},
		{
			name: "Keep when some of multiple required attributes are missing",
			removeIfMissing: map[string][]string{
				"test_object": {"attr1", "attr2"},
			},
			attribute: "test_object",
			config: map[string]interface{}{
				"attr1": "value1",
				// attr2 is missing
				"attr3": "value3",
			},
			expectedRemove: false, // Keep because not ALL required attributes are missing
		},
		{
			name: "Remove when all of multiple required attributes are missing",
			removeIfMissing: map[string][]string{
				"test_object": {"attr1", "attr2"},
			},
			attribute: "test_object",
			config: map[string]interface{}{
				// attr1 is missing
				// attr2 is missing
				"attr3": "value3",
			},
			expectedRemove: true, // Remove because ALL required attributes are missing
		},
		{
			name: "No removal for non-configured attributes",
			removeIfMissing: map[string][]string{
				"other_attribute": {"some_field"},
			},
			attribute: "test_attribute",
			config: map[string]interface{}{
				"some_field": "value",
			},
			expectedRemove: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exporter := &ResourceExporter{
				RemoveIfMissing: test.removeIfMissing,
			}

			// Test the actual method
			actualRemove := exporter.RemoveFieldIfMissing(test.attribute, test.config)

			if actualRemove != test.expectedRemove {
				t.Errorf("Expected RemoveFieldIfMissing to return %v, but got %v", test.expectedRemove, actualRemove)
			}
		})
	}
}

// TestUnitRemoveIfSelfReferential tests the RemoveIfSelfReferential functionality.
// This feature removes attributes from the configuration when their value
// matches the resource's own ID (self-referential).
// For example, if a queue's backup_queue_id points to itself, that attribute
// should be removed from the export to avoid circular references.
func TestUnitRemoveIfSelfReferential(t *testing.T) {
	resourceId := uuid.NewString()
	differentId := uuid.NewString()

	tests := []struct {
		name                    string
		removeIfSelfReferential []string
		attributeKey            string
		attributePath           string
		config                  map[string]interface{}
		expectedRemove          bool
	}{
		{
			name:                    "Remove self-referential attribute",
			removeIfSelfReferential: []string{"backup_queue_id"},
			attributeKey:            "backup_queue_id",
			attributePath:           "backup_queue_id",
			config: map[string]interface{}{
				"id":              resourceId,
				"backup_queue_id": resourceId, // Self-reference
				"name":            "test_queue",
			},
			expectedRemove: true,
		},
		{
			name:                    "Keep non-self-referential attribute",
			removeIfSelfReferential: []string{"backup_queue_id"},
			attributeKey:            "backup_queue_id",
			attributePath:           "backup_queue_id",
			config: map[string]interface{}{
				"id":              resourceId,
				"backup_queue_id": differentId, // Different ID
				"name":            "test_queue",
			},
			expectedRemove: false,
		},
		{
			name:                    "No removal for non-configured attributes",
			removeIfSelfReferential: []string{"other_attribute"},
			attributeKey:            "backup_queue_id",
			attributePath:           "backup_queue_id",
			config: map[string]interface{}{
				"id":              resourceId,
				"backup_queue_id": resourceId, // Self-reference but not configured for removal
				"name":            "test_queue",
			},
			expectedRemove: false,
		},
		{
			name:                    "Remove another self-referential attribute",
			removeIfSelfReferential: []string{"parent_id", "backup_queue_id"},
			attributeKey:            "parent_id",
			attributePath:           "parent_id",
			config: map[string]interface{}{
				"id":        resourceId,
				"parent_id": resourceId, // Self-reference
				"name":      "test_resource",
			},
			expectedRemove: true,
		},
		{
			name:                    "Remove nested self-referential attribute",
			removeIfSelfReferential: []string{"parent_id.backup_queue_id"},
			attributeKey:            "backup_queue_id",
			attributePath:           "parent_id.backup_queue_id",
			config:                  map[string]interface{}{"backup_queue_id": resourceId}, // Nested self-reference [
			expectedRemove:          true,
		},
		{
			name:                    "Keep when attribute value is empty string",
			removeIfSelfReferential: []string{"backup_queue_id"},
			attributeKey:            "backup_queue_id",
			attributePath:           "backup_queue_id",
			config: map[string]interface{}{
				"id":              resourceId,
				"backup_queue_id": "", // Empty string
				"name":            "test_queue",
			},
			expectedRemove: false,
		},
		{
			name:                    "Gracefully handle nil attribute",
			removeIfSelfReferential: []string{"backup_queue_id"},
			attributeKey:            "backup_queue_id",
			attributePath:           "backup_queue_id",
			config: map[string]interface{}{
				"id":              resourceId,
				"backup_queue_id": nil,
				"name":            "test_queue",
			},
			expectedRemove: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exporter := &ResourceExporter{
				RemoveIfSelfReferential: test.removeIfSelfReferential,
			}

			// Test the actual method
			actualRemove := exporter.RemoveFieldIfSelfReferential(resourceId, test.attributePath, test.attributeKey, test.config)

			if actualRemove != test.expectedRemove {
				t.Errorf("Expected RemoveFieldIfSelfReferential to return %v, but got %v", test.expectedRemove, actualRemove)
			}
		})
	}
}
