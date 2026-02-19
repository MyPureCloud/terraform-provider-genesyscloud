package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/examples"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
)

// SwaggerSpec represents the relevant parts of the Swagger/OAS v2 specification
type SwaggerSpec struct {
	Paths map[string]map[string]EndpointSpec `json:"paths"`
}

// EndpointSpec represents an API endpoint specification
type EndpointSpec struct {
	XIninRequiresPermissions PermissionsSpec       `json:"x-inin-requires-permissions"`
	Security                 []map[string][]string `json:"security"`
}

// PermissionsSpec represents the permissions structure
type PermissionsSpec struct {
	Type        string   `json:"type"`
	Permissions []string `json:"permissions"`
}

// APIEndpoint represents a parsed API endpoint
type APIEndpoint struct {
	Method string
	Path   string
}

// ResourcePermissions represents the permissions and scopes for a resource
type ResourcePermissions struct {
	ResourceType string   `json:"resource_type"`
	ResourceName string   `json:"resource_name"`
	Permissions  []string `json:"permissions"`
	Scopes       []string `json:"scopes"`
	Endpoints    []string `json:"endpoints"`
}

// PermissionsData represents the complete permissions data structure
type PermissionsData struct {
	Version   string                `json:"version"`
	Resources []ResourcePermissions `json:"resources"`
}

// Method to insert the contents of each resource's apis.md file into the markdown documentation
func main() {
	fmt.Println("Updating APIs in docs...")
	const (
		resourceFolder = "docs/resources"
		exampleFolder  = "examples/resources"
		apiDocsTag     = "**No APIs**"
		swaggerURL     = "https://api.mypurecloud.com/api/v2/docs/swagger"
		outputDir      = "public/data"
		outputFilename = "resource_permissions"
	)

	// Get version from command line args or use default
	version := "latest"
	if len(os.Args) > 1 {
		version = os.Args[1]
	}

	missingExamples := []string{}
	ignoredExamples := examples.GetIgnoredResources()

	// Slice to collect all resource permissions
	var allResourcePermissions []ResourcePermissions

	// Fetch Swagger specification once for all resources
	fmt.Println("Fetching Swagger specification...")
	swaggerSpec, err := fetchSwaggerSpec(swaggerURL)
	if err != nil {
		log.Printf("Warning: Failed to fetch Swagger spec: %v. Continuing without permission/scope information.", err)
		swaggerSpec = nil
	} else {
		fmt.Println("Swagger specification loaded successfully")
	}

	files, err := ioutil.ReadDir("docs/resources")
	if err != nil {
		log.Fatalf("Failed to read folder %s", resourceFolder)
	}

	for _, file := range files {

		shortResourceName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		resourceName := fmt.Sprintf("genesyscloud_%s", shortResourceName)
		fullResourceFilePath := fmt.Sprintf("%s/%s", resourceFolder, file.Name())

		// Remove any docs generated for ignored examples
		if lists.ItemInSlice(resourceName, ignoredExamples) {
			os.Remove(fullResourceFilePath)
			continue
		}

		// If no examples are provided, note, and alert at end
		examplesDir := fmt.Sprintf("%s/%s", exampleFolder, resourceName)
		if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
			log.Printf("No examples found! %s", resourceName)
			missingExamples = append(missingExamples, shortResourceName)
		}

		// Open and read the apis.md file for this resource
		apiFileName := fmt.Sprintf("%s/apis.md", examplesDir)
		apisFile, err := os.Open(apiFileName)
		if err != nil {
			fmt.Printf("Missing APIs file: %s\n", apiFileName)
			continue
		}
		defer apisFile.Close()

		apiFileBytes, err := ioutil.ReadAll(apisFile)
		if err != nil {
			fmt.Printf("Couldn't read bytes from %s\n", apiFileName)
			continue
		}

		// Parse API endpoints and enhance with permissions/scopes
		enhancedContent := string(apiFileBytes)
		if swaggerSpec != nil {
			endpoints := parseAPIEndpoints(string(apiFileBytes))
			if len(endpoints) > 0 {
				permissionsAndScopes := extractPermissionsAndScopes(endpoints, swaggerSpec)
				if permissionsAndScopes != "" {
					enhancedContent = insertPermissionsAndScopes(string(apiFileBytes), permissionsAndScopes)
				}

				// Collect permissions data for JSON output
				resourcePerms := extractResourcePermissions(resourceName, shortResourceName, endpoints, swaggerSpec)
				if len(resourcePerms.Permissions) > 0 || len(resourcePerms.Scopes) > 0 {
					allResourcePermissions = append(allResourcePermissions, resourcePerms)
				}
			}
		}

		//open the doc file
		docFile, err := os.OpenFile(fmt.Sprintf("%s/%s", resourceFolder, file.Name()), os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("Couldn't open file: %s\n", file.Name())
			continue
		}
		defer docFile.Close()

		docFileBytes, err := ioutil.ReadAll(docFile)
		if err != nil {
			fmt.Printf("Couldn't read bytes from %s\n", file.Name())
			continue
		}

		// Replace the **No APIs** line with the enhanced content
		newBytes := bytes.Replace(docFileBytes, []byte(apiDocsTag), []byte(enhancedContent), 1)
		docFile.Truncate(0)
		docFile.WriteAt(newBytes, 0)
		fmt.Printf("Updated APIs in doc file: %s\n", file.Name())
	}

	fmt.Println()
	fmt.Printf("The following resources were explicitly ignored, and so no docs were generated: %v", ignoredExamples)
	fmt.Println()
	fmt.Printf("The following resources did not have any examples, and so docs without examples or APIs were generated: %v", missingExamples)

	// Write permissions data to JSON file
	if len(allResourcePermissions) > 0 {
		fmt.Println()
		fmt.Println("Writing permissions data to JSON file...")
		if err := writePermissionsJSON(allResourcePermissions, outputDir, outputFilename, version); err != nil {
			log.Printf("Error writing permissions JSON: %v", err)
		} else {
			outputFile := fmt.Sprintf("%s-%s.json", outputFilename, version)
			fmt.Printf("Permissions data written to: %s/%s\n", outputDir, outputFile)
		}
	}
}

// writePermissionsJSON writes the permissions data to a JSON file
func writePermissionsJSON(permissions []ResourcePermissions, outputDir, filename, version string) error {
	// Sort by resource type
	sort.Slice(permissions, func(i, j int) bool {
		return permissions[i].ResourceType < permissions[j].ResourceType
	})

	// Create the data structure
	data := PermissionsData{
		Version:   version,
		Resources: permissions,
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create versioned filename
	versionedFilename := fmt.Sprintf("%s-%s.json", filename, version)
	outputPath := filepath.Join(outputDir, versionedFilename)

	// Write to file
	if err := ioutil.WriteFile(outputPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// fetchSwaggerSpec downloads and parses the Swagger specification
func fetchSwaggerSpec(url string) (*SwaggerSpec, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch swagger spec: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var spec SwaggerSpec
	if err := json.Unmarshal(body, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse swagger spec: %w", err)
	}

	return &spec, nil
}

// parseAPIEndpoints extracts API endpoints from the apis.md content
func parseAPIEndpoints(content string) []APIEndpoint {
	var endpoints []APIEndpoint

	// Regex to match lines like: - [POST /api/v2/path](url) or * [POST /api/v2/path](url)
	re := regexp.MustCompile(`[-*]\s*\[([A-Z]+)\s+(/api/v2/[^\]]+)\]`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			endpoints = append(endpoints, APIEndpoint{
				Method: strings.ToLower(match[1]),
				Path:   match[2],
			})
		}
	}

	return endpoints
}

// extractPermissionsAndScopes aggregates permissions and scopes from endpoints
func extractPermissionsAndScopes(endpoints []APIEndpoint, spec *SwaggerSpec) string {
	permissionsMap := make(map[string]bool)
	scopesMap := make(map[string]bool)

	for _, endpoint := range endpoints {
		// Normalize path - remove path parameters like {id}
		normalizedPath := endpoint.Path

		// Look up the endpoint in the spec
		pathSpec, pathExists := spec.Paths[normalizedPath]
		if !pathExists {
			continue
		}

		endpointSpec, methodExists := pathSpec[endpoint.Method]
		if !methodExists {
			continue
		}

		// Extract permissions
		for _, perm := range endpointSpec.XIninRequiresPermissions.Permissions {
			if perm != "" {
				permissionsMap[perm] = true
			}
		}

		// Extract scopes
		for _, securityItem := range endpointSpec.Security {
			if oauthScopes, ok := securityItem["PureCloud OAuth"]; ok {
				for _, scope := range oauthScopes {
					if scope != "" {
						scopesMap[scope] = true
					}
				}
			}
		}
	}

	// If no permissions or scopes found, return empty string
	if len(permissionsMap) == 0 && len(scopesMap) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString("## Permissions and Scopes\n\n")

	// Add permissions section
	if len(permissionsMap) > 0 {
		result.WriteString("The following permissions are required to use this resource:\n\n")

		permissions := make([]string, 0, len(permissionsMap))
		for perm := range permissionsMap {
			permissions = append(permissions, perm)
		}
		sort.Strings(permissions)

		for _, perm := range permissions {
			result.WriteString(fmt.Sprintf("* `%s`\n", perm))
		}
		result.WriteString("\n")
	}

	// Add scopes section
	if len(scopesMap) > 0 {
		result.WriteString("The following OAuth scopes are required to use this resource:\n\n")

		scopes := make([]string, 0, len(scopesMap))
		for scope := range scopesMap {
			scopes = append(scopes, scope)
		}
		sort.Strings(scopes)

		for _, scope := range scopes {
			result.WriteString(fmt.Sprintf("* `%s`\n", scope))
		}
	}

	return result.String()
}

// insertPermissionsAndScopes inserts the permissions and scopes section before the first
// markdown heading (##) in the content, or at the end if no headings are present
func insertPermissionsAndScopes(content, permissionsAndScopes string) string {
	lines := strings.Split(content, "\n")

	// Find the first line that starts with ## (markdown heading)
	insertIndex := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "##") {
			insertIndex = i
			break
		}
	}

	// If no heading found, append at the end
	if insertIndex == -1 {
		return content + "\n" + permissionsAndScopes
	}

	// Insert before the first heading
	result := strings.Builder{}

	// Add all lines before the heading
	for i := 0; i < insertIndex; i++ {
		result.WriteString(lines[i])
		result.WriteString("\n")
	}

	// Add the permissions and scopes section
	result.WriteString(permissionsAndScopes)

	// Add the remaining lines (including the heading)
	for i := insertIndex; i < len(lines); i++ {
		result.WriteString(lines[i])
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// extractResourcePermissions extracts permissions and scopes for a resource and returns structured data
func extractResourcePermissions(resourceType, resourceName string, endpoints []APIEndpoint, spec *SwaggerSpec) ResourcePermissions {
	permissionsMap := make(map[string]bool)
	scopesMap := make(map[string]bool)
	endpointsList := []string{}

	for _, endpoint := range endpoints {
		// Add endpoint to list
		endpointsList = append(endpointsList, fmt.Sprintf("%s %s", strings.ToUpper(endpoint.Method), endpoint.Path))

		// Look up the endpoint in the spec
		pathSpec, pathExists := spec.Paths[endpoint.Path]
		if !pathExists {
			continue
		}

		endpointSpec, methodExists := pathSpec[endpoint.Method]
		if !methodExists {
			continue
		}

		// Extract permissions
		for _, perm := range endpointSpec.XIninRequiresPermissions.Permissions {
			if perm != "" {
				permissionsMap[perm] = true
			}
		}

		// Extract scopes
		for _, securityItem := range endpointSpec.Security {
			if oauthScopes, ok := securityItem["PureCloud OAuth"]; ok {
				for _, scope := range oauthScopes {
					if scope != "" {
						scopesMap[scope] = true
					}
				}
			}
		}
	}

	// Convert maps to sorted slices
	permissions := make([]string, 0, len(permissionsMap))
	for perm := range permissionsMap {
		permissions = append(permissions, perm)
	}
	sort.Strings(permissions)

	scopes := make([]string, 0, len(scopesMap))
	for scope := range scopesMap {
		scopes = append(scopes, scope)
	}
	sort.Strings(scopes)

	return ResourcePermissions{
		ResourceType: resourceType,
		ResourceName: resourceName,
		Permissions:  permissions,
		Scopes:       scopes,
		Endpoints:    endpointsList,
	}
}
