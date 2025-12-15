package resource_metadata

import (
	"testing"
	"time"
)

func TestResourceMetadata_Validation(t *testing.T) {
	tests := []struct {
		name     string
		metadata *ResourceMetadata
		wantErr  bool
		errType  string
	}{
		{
			name: "valid metadata",
			metadata: &ResourceMetadata{
				ResourceType: "genesyscloud_flow",
				TeamName:     "Platform Team",
				TeamChatRoom: "#platform-team",
			},
			wantErr: false,
		},
		{
			name: "missing resource type",
			metadata: &ResourceMetadata{
				TeamName:     "Platform Team",
				TeamChatRoom: "#platform-team",
			},
			wantErr: true,
			errType: "ValidationError",
		},
		{
			name: "missing team name",
			metadata: &ResourceMetadata{
				ResourceType: "genesyscloud_flow",
				TeamChatRoom: "#platform-team",
			},
			wantErr: true,
			errType: "ValidationError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewDefaultRegistry()
			err := registry.RegisterMetadata(tt.metadata)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if tt.errType != "" {
					switch tt.errType {
					case "ValidationError":
						if _, ok := err.(*ValidationError); !ok {
							t.Errorf("Expected ValidationError, got %T", err)
						}
					case "NotFoundError":
						if _, ok := err.(*NotFoundError); !ok {
							t.Errorf("Expected NotFoundError, got %T", err)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestDefaultRegistry_RegisterAndRetrieve(t *testing.T) {
	registry := NewDefaultRegistry()

	metadata := &ResourceMetadata{
		ResourceType: "genesyscloud_flow",
		PackageName:  "architect_flow",
		TeamName:     "Platform Team",
		TeamChatRoom: "#platform-team",
		Description:  "Manages Genesys Cloud flows",
	}

	// Test registration
	err := registry.RegisterMetadata(metadata)
	if err != nil {
		t.Fatalf("Failed to register metadata: %v", err)
	}

	// Test retrieval
	retrieved, err := registry.GetMetadata("genesyscloud_flow")
	if err != nil {
		t.Fatalf("Failed to retrieve metadata: %v", err)
	}

	if retrieved.ResourceType != metadata.ResourceType {
		t.Errorf("Expected ResourceType %s, got %s", metadata.ResourceType, retrieved.ResourceType)
	}

	if retrieved.TeamName != metadata.TeamName {
		t.Errorf("Expected TeamName %s, got %s", metadata.TeamName, retrieved.TeamName)
	}

	// Test that timestamps were set
	if retrieved.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if retrieved.LastUpdated.IsZero() {
		t.Error("Expected LastUpdated to be set")
	}
}

func TestDefaultRegistry_GetAllMetadata(t *testing.T) {
	registry := NewDefaultRegistry()

	metadata1 := &ResourceMetadata{
		ResourceType: "genesyscloud_flow",
		TeamName:     "Platform Team",
	}

	metadata2 := &ResourceMetadata{
		ResourceType: "genesyscloud_queue",
		TeamName:     "Routing Team",
	}

	// Register multiple metadata
	registry.RegisterMetadata(metadata1)
	registry.RegisterMetadata(metadata2)

	// Test retrieval of all metadata
	allMetadata, err := registry.GetAllMetadata()
	if err != nil {
		t.Fatalf("Failed to get all metadata: %v", err)
	}

	if len(allMetadata) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(allMetadata))
	}
}

func TestDefaultRegistry_UpdateLastUpdated(t *testing.T) {
	registry := NewDefaultRegistry()

	metadata := &ResourceMetadata{
		ResourceType: "genesyscloud_flow",
		TeamName:     "Platform Team",
	}

	// Register metadata
	registry.RegisterMetadata(metadata)

	// Get initial timestamp
	initial, _ := registry.GetMetadata("genesyscloud_flow")
	initialTime := initial.LastUpdated

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Update timestamp
	err := registry.UpdateLastUpdated("genesyscloud_flow")
	if err != nil {
		t.Fatalf("Failed to update last updated: %v", err)
	}

	// Get updated metadata
	updated, _ := registry.GetMetadata("genesyscloud_flow")

	if !updated.LastUpdated.After(initialTime) {
		t.Error("Expected LastUpdated to be updated")
	}
}

func TestDefaultRegistry_NotFoundError(t *testing.T) {
	registry := NewDefaultRegistry()

	// Try to get non-existent metadata
	_, err := registry.GetMetadata("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent metadata")
		return
	}

	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}

func TestAnnotationExtractor_ExtractMetadata(t *testing.T) {
	extractor := NewAnnotationExtractor()

	annotations := ResourceAnnotations{
		ResourceType: "genesyscloud_flow",
		PackageName:  "architect_flow",
		TeamName:     "Platform Team",
		TeamChatRoom: "#platform-team",
		Description:  "Manages flows",
	}

	metadata, err := extractor.ExtractMetadata(annotations)
	if err != nil {
		t.Fatalf("Failed to extract metadata: %v", err)
	}

	if metadata.ResourceType != "genesyscloud_flow" {
		t.Errorf("Expected ResourceType %s, got %s", "genesyscloud_flow", metadata.ResourceType)
	}

	if metadata.TeamName != "Platform Team" {
		t.Errorf("Expected TeamName %s, got %s", "Platform Team", metadata.TeamName)
	}
}

func TestParseAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "simple annotations",
			input:    "//go:build team=Platform chat=#platform-team",
			expected: map[string]string{"team": "Platform", "chat": "#platform-team"},
		},
		{
			name:     "legacy build tags",
			input:    "// +build team=Platform chat=#platform-team",
			expected: map[string]string{"team": "Platform", "chat": "#platform-team"},
		},
		{
			name:     "no annotations",
			input:    "//go:build some_other_tag",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseAnnotations(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d annotations, got %d", len(tt.expected), len(result))
			}

			for key, value := range tt.expected {
				if result[key] != value {
					t.Errorf("Expected %s=%s, got %s=%s", key, value, key, result[key])
				}
			}
		})
	}
}

func TestBuildAnnotationString(t *testing.T) {
	metadata := &ResourceMetadata{
		TeamName:     "Platform Team",
		TeamChatRoom: "#platform-team",
		LastUpdated:  time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	result := BuildAnnotationString(metadata)
	expected := "//go:build team=Platform Team chat=#platform-team updated=2024-01-15"

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestResourceDiscovery_ValidateResourceMetadata(t *testing.T) {
	discovery := NewResourceDiscovery("test")

	tests := []struct {
		name     string
		metadata *ResourceMetadata
		wantErr  bool
	}{
		{
			name: "valid metadata",
			metadata: &ResourceMetadata{
				ResourceType: "genesyscloud_flow",
				TeamName:     "Platform Team",
			},
			wantErr: false,
		},
		{
			name: "missing resource type",
			metadata: &ResourceMetadata{
				TeamName: "Platform Team",
			},
			wantErr: true,
		},
		{
			name: "missing team name",
			metadata: &ResourceMetadata{
				ResourceType: "genesyscloud_flow",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := discovery.ValidateResourceMetadata(tt.metadata)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
