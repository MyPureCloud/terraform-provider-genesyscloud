package resource_metadata

import (
	"reflect"
	"strings"
	"time"
)

// ResourceAnnotations defines the metadata annotations for a resource
// This struct uses Go struct tags to define metadata that can be extracted
// by the metadata framework
type ResourceAnnotations struct {
	// Resource information - extracted from package and resource type
	ResourceType string `metadata:"resource_type"`
	PackageName  string `metadata:"package_name"`

	// Team ownership information - required annotations
	TeamName     string `metadata:"team_name" required:"true"`
	TeamChatRoom string `metadata:"team_chat_room"`

	// Timestamps - automatically managed
	LastUpdated time.Time `metadata:"last_updated"`
	CreatedAt   time.Time `metadata:"created_at"`

	// Additional metadata
	Description string            `metadata:"description"`
	Tags        map[string]string `metadata:"tags"`
}

// AnnotationExtractor extracts metadata from struct tags
type AnnotationExtractor struct{}

// NewAnnotationExtractor creates a new annotation extractor
func NewAnnotationExtractor() *AnnotationExtractor {
	return &AnnotationExtractor{}
}

// ExtractMetadata extracts metadata from a struct using reflection
func (e *AnnotationExtractor) ExtractMetadata(v interface{}) (*ResourceMetadata, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, &ValidationError{Field: "Type", Message: "Value must be a struct"}
	}

	metadata := &ResourceMetadata{}
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Extract metadata tag
		metadataTag := fieldType.Tag.Get("metadata")
		if metadataTag == "" {
			continue
		}

		// Check if field is required
		required := fieldType.Tag.Get("required") == "true"

		// Extract value based on field type
		var value interface{}
		switch field.Kind() {
		case reflect.String:
			value = field.String()
		case reflect.Map:
			if field.IsNil() {
				value = make(map[string]string)
			} else {
				value = field.Interface()
			}
		case reflect.Struct:
			if field.Type() == reflect.TypeOf(time.Time{}) {
				value = field.Interface()
			}
		}

		// Set the metadata field
		if err := e.setMetadataField(metadata, metadataTag, value, required); err != nil {
			return nil, err
		}
	}

	return metadata, nil
}

// setMetadataField sets a field in the ResourceMetadata struct
func (e *AnnotationExtractor) setMetadataField(metadata *ResourceMetadata, tag string, value interface{}, required bool) error {
	switch tag {
	case "resource_type":
		if str, ok := value.(string); ok {
			metadata.ResourceType = str
		}
	case "package_name":
		if str, ok := value.(string); ok {
			metadata.PackageName = str
		}
	case "team_name":
		if str, ok := value.(string); ok {
			metadata.TeamName = str
		} else if required {
			return &ValidationError{Field: "TeamName", Message: "TeamName is required"}
		}
	case "team_chat_room":
		if str, ok := value.(string); ok {
			metadata.TeamChatRoom = str
		}
	case "last_updated":
		if t, ok := value.(time.Time); ok {
			metadata.LastUpdated = t
		}
	case "created_at":
		if t, ok := value.(time.Time); ok {
			metadata.CreatedAt = t
		}
	case "description":
		if str, ok := value.(string); ok {
			metadata.Description = str
		}
	case "tags":
		if tags, ok := value.(map[string]string); ok {
			metadata.Tags = tags
		}
	}

	return nil
}

// ParseAnnotations parses annotations from a string (useful for build tags)
func ParseAnnotations(annotationString string) map[string]string {
	annotations := make(map[string]string)

	// Remove build tag prefix if present
	if strings.HasPrefix(annotationString, "//go:build ") {
		annotationString = strings.TrimPrefix(annotationString, "//go:build ")
	}
	if strings.HasPrefix(annotationString, "// +build ") {
		annotationString = strings.TrimPrefix(annotationString, "// +build ")
	}

	// Parse key-value pairs
	pairs := strings.Split(annotationString, " ")
	for _, pair := range pairs {
		if strings.Contains(pair, "=") {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				annotations[key] = value
			}
		}
	}

	return annotations
}

// BuildAnnotationString creates a build annotation string from metadata
func BuildAnnotationString(metadata *ResourceMetadata) string {
	var parts []string

	if metadata.TeamName != "" {
		parts = append(parts, "team="+metadata.TeamName)
	}
	if metadata.TeamChatRoom != "" {
		parts = append(parts, "chat="+metadata.TeamChatRoom)
	}
	if !metadata.LastUpdated.IsZero() {
		parts = append(parts, "updated="+metadata.LastUpdated.Format("2006-01-02"))
	}

	if len(parts) == 0 {
		return ""
	}

	return "//go:build " + strings.Join(parts, " ")
}
