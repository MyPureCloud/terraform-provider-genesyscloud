package resource_metadata

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ResourceDiscovery provides functionality to discover and extract metadata from resources
type ResourceDiscovery struct {
	basePath string
}

// NewResourceDiscovery creates a new resource discovery instance
func NewResourceDiscovery(basePath string) *ResourceDiscovery {
	return &ResourceDiscovery{
		basePath: basePath,
	}
}

// DiscoverResources scans the codebase for resource schema files and extracts metadata
func (d *ResourceDiscovery) DiscoverResources() ([]*ResourceMetadata, error) {
	var allMetadata []*ResourceMetadata

	// Walk through the genesyscloud directory
	err := filepath.Walk(d.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Look for schema files
		if strings.Contains(path, "_schema.go") {
			metadata, err := d.extractMetadataFromFile(path)
			if err != nil {
				// Log error but continue scanning
				fmt.Printf("Warning: Failed to extract metadata from %s: %v\n", path, err)
				return nil
			}

			if metadata != nil {
				allMetadata = append(allMetadata, metadata)
			}
		}

		return nil
	})

	return allMetadata, err
}

// extractMetadataFromFile extracts metadata from a single schema file
func (d *ResourceDiscovery) extractMetadataFromFile(filePath string) (*ResourceMetadata, error) {
	// Parse the Go file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Extract package name and resource type
	packageName := node.Name.Name
	resourceType := d.extractResourceType(node)

	if resourceType == "" {
		return nil, fmt.Errorf("could not determine resource type from file")
	}

	// Extract metadata from comments and build tags
	metadata := &ResourceMetadata{
		ResourceType: resourceType,
		PackageName:  packageName,
	}

	// Extract from file comments
	d.extractMetadataFromComments(node.Comments, metadata)

	// Extract from build tags
	d.extractMetadataFromBuildTags(node, metadata)

	// If no team name is found, this might not be a resource we want to track
	if metadata.TeamName == "" {
		return nil, nil
	}

	return metadata, nil
}

// extractResourceType extracts the resource type from the file
func (d *ResourceDiscovery) extractResourceType(node *ast.File) string {
	// Look for ResourceType constant
	for _, decl := range node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					for i, name := range valueSpec.Names {
						if name.Name == "ResourceType" {
							if len(valueSpec.Values) > i {
								if lit, ok := valueSpec.Values[i].(*ast.BasicLit); ok {
									return strings.Trim(lit.Value, `"`)
								}
							}
						}
					}
				}
			}
		}
	}

	// Fallback: try to extract from filename
	fileName := filepath.Base(node.Name.Name)
	if strings.Contains(fileName, "_schema.go") {
		parts := strings.Split(fileName, "_")
		if len(parts) >= 2 {
			return "genesyscloud_" + parts[0]
		}
	}

	return ""
}

// extractMetadataFromComments extracts metadata from file comments
func (d *ResourceDiscovery) extractMetadataFromComments(comments []*ast.CommentGroup, metadata *ResourceMetadata) {
	for _, commentGroup := range comments {
		for _, comment := range commentGroup.List {
			text := comment.Text

			// Look for metadata comments
			if strings.HasPrefix(text, "// @team:") {
				metadata.TeamName = strings.TrimSpace(strings.TrimPrefix(text, "// @team:"))
			} else if strings.HasPrefix(text, "// @chat:") {
				metadata.TeamChatRoom = strings.TrimSpace(strings.TrimPrefix(text, "// @chat:"))
			} else if strings.HasPrefix(text, "// @description:") {
				metadata.Description = strings.TrimSpace(strings.TrimPrefix(text, "// @description:"))
			}
		}
	}
}

// extractMetadataFromBuildTags extracts metadata from build tags
func (d *ResourceDiscovery) extractMetadataFromBuildTags(node *ast.File, metadata *ResourceMetadata) {
	// Look for build tags in the file
	for _, decl := range node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			if genDecl.Tok == token.IMPORT {
				// Check for build tags in import comments
				if genDecl.Doc != nil {
					for _, comment := range genDecl.Doc.List {
						if strings.HasPrefix(comment.Text, "//go:build ") {
							annotations := ParseAnnotations(comment.Text)
							d.applyAnnotations(annotations, metadata)
						}
					}
				}
			}
		}
	}
}

// applyAnnotations applies parsed annotations to metadata
func (d *ResourceDiscovery) applyAnnotations(annotations map[string]string, metadata *ResourceMetadata) {
	if team, exists := annotations["team"]; exists {
		metadata.TeamName = team
	}
	if chat, exists := annotations["chat"]; exists {
		metadata.TeamChatRoom = chat
	}
	if updated, exists := annotations["updated"]; exists {
		// Parse date if needed
		// For now, just store as string in description
		if metadata.Description == "" {
			metadata.Description = "Last updated: " + updated
		}
	}
}

// GenerateMetadataTemplate generates a metadata template for a resource
func (d *ResourceDiscovery) GenerateMetadataTemplate(resourceType, packageName string) string {
	template := fmt.Sprintf(`// Resource Metadata Template for %s
// Add these comments to your schema file:

// @team: [TEAM_NAME]
// @chat: [TEAM_CHAT_ROOM]
// @description: [RESOURCE_DESCRIPTION]

// Or use build tags:
//go:build team=[TEAM_NAME] chat=[TEAM_CHAT_ROOM]

// Example:
// @team: Platform Team
// @chat: #platform-team
// @description: Manages Genesys Cloud flows and their configurations
`, resourceType)

	return template
}

// ValidateResourceMetadata validates that a resource has proper metadata
func (d *ResourceDiscovery) ValidateResourceMetadata(metadata *ResourceMetadata) error {
	if metadata.ResourceType == "" {
		return &ValidationError{Field: "ResourceType", Message: "ResourceType is required"}
	}

	if metadata.TeamName == "" {
		return &ValidationError{Field: "TeamName", Message: "TeamName is required"}
	}

	return nil
}
