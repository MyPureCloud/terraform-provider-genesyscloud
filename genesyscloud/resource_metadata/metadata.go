package resource_metadata

import (
	"time"
)

// ResourceMetadata represents metadata for a Terraform resource
type ResourceMetadata struct {
	// Resource information
	ResourceType string `json:"resource_type" yaml:"resource_type"`
	PackageName  string `json:"package_name" yaml:"package_name"`

	// Team ownership information
	TeamName     string `json:"team_name" yaml:"team_name"`
	TeamChatRoom string `json:"team_chat_room" yaml:"team_chat_room"`

	// Timestamps
	LastUpdated time.Time `json:"last_updated" yaml:"last_updated"`
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`

	// Additional metadata
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Tags        map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// MetadataRegistry provides methods to register and retrieve resource metadata
type MetadataRegistry interface {
	// RegisterMetadata registers metadata for a resource
	RegisterMetadata(metadata *ResourceMetadata) error

	// GetMetadata retrieves metadata for a resource type
	GetMetadata(resourceType string) (*ResourceMetadata, error)

	// GetAllMetadata retrieves all registered metadata
	GetAllMetadata() ([]*ResourceMetadata, error)

	// UpdateLastUpdated updates the last updated timestamp for a resource
	UpdateLastUpdated(resourceType string) error
}

// DefaultRegistry implements MetadataRegistry with in-memory storage
type DefaultRegistry struct {
	metadata map[string]*ResourceMetadata
}

// NewDefaultRegistry creates a new default metadata registry
func NewDefaultRegistry() *DefaultRegistry {
	return &DefaultRegistry{
		metadata: make(map[string]*ResourceMetadata),
	}
}

// RegisterMetadata registers metadata for a resource
func (r *DefaultRegistry) RegisterMetadata(metadata *ResourceMetadata) error {
	if metadata.ResourceType == "" {
		return &ValidationError{Field: "ResourceType", Message: "ResourceType cannot be empty"}
	}
	if metadata.TeamName == "" {
		return &ValidationError{Field: "TeamName", Message: "TeamName cannot be empty"}
	}

	// Set timestamps if not already set
	if metadata.CreatedAt.IsZero() {
		metadata.CreatedAt = time.Now()
	}
	if metadata.LastUpdated.IsZero() {
		metadata.LastUpdated = time.Now()
	}

	r.metadata[metadata.ResourceType] = metadata
	return nil
}

// GetMetadata retrieves metadata for a resource type
func (r *DefaultRegistry) GetMetadata(resourceType string) (*ResourceMetadata, error) {
	metadata, exists := r.metadata[resourceType]
	if !exists {
		return nil, &NotFoundError{ResourceType: resourceType}
	}
	return metadata, nil
}

// GetAllMetadata retrieves all registered metadata
func (r *DefaultRegistry) GetAllMetadata() ([]*ResourceMetadata, error) {
	metadata := make([]*ResourceMetadata, 0, len(r.metadata))
	for _, m := range r.metadata {
		metadata = append(metadata, m)
	}
	return metadata, nil
}

// UpdateLastUpdated updates the last updated timestamp for a resource
func (r *DefaultRegistry) UpdateLastUpdated(resourceType string) error {
	metadata, exists := r.metadata[resourceType]
	if !exists {
		return &NotFoundError{ResourceType: resourceType}
	}

	metadata.LastUpdated = time.Now()
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// NotFoundError represents a not found error
type NotFoundError struct {
	ResourceType string
}

func (e *NotFoundError) Error() string {
	return "metadata not found for resource type: " + e.ResourceType
}
