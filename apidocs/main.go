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

const (
	swaggerURL            = "https://api.mypurecloud.com/api/v2/docs/swagger"
	proxyFileGlob         = "genesyscloud/*/genesyscloud_*_proxy.go"
	resourceExampleFolder = "examples/resources"
	resourceDocsFolder    = "docs/resources"
	dataSourceExampleFolder = "examples/data-sources"
	dataSourceDocsFolder    = "docs/data-sources"
)

// SwaggerSpec represents the relevant parts of the Swagger/OAS v2 specification
type SwaggerSpec struct {
	Paths map[string]map[string]EndpointSpec `json:"paths"`
}

// EndpointSpec represents an API endpoint specification
type EndpointSpec struct {
	OperationId              string                `json:"operationId"`
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
		apiDocsTag     = "**No APIs**"
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

	// Build operationId map for proxy auditing
	var opMap map[string]APIEndpoint
	if swaggerSpec != nil {
		opMap = buildOperationIdMap(swaggerSpec)
		fmt.Printf("Built operation ID map with %d entries\n", len(opMap))
	}

	// Process resources
	fmt.Println("\nProcessing resources...")
	allResourcePermissions = processDocsFolder(resourceDocsFolder, resourceExampleFolder, apiDocsTag, ignoredExamples, swaggerSpec, opMap, allResourcePermissions, &missingExamples)

	// Process data sources
	fmt.Println("\nProcessing data sources...")
	allResourcePermissions = processDocsFolder(dataSourceDocsFolder, dataSourceExampleFolder, apiDocsTag, ignoredExamples, swaggerSpec, opMap, allResourcePermissions, &missingExamples)

	fmt.Println()
	fmt.Printf("The following resources were explicitly ignored, and so no docs were generated: %v", ignoredExamples)
	fmt.Println()
	fmt.Printf("The following resources/data-sources did not have any examples, and so docs without examples or APIs were generated: %v", missingExamples)

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

// processDocsFolder iterates over doc files in a folder, audits proxy files,
// updates apis.md, and enhances docs with permissions and notes.
func processDocsFolder(docsFolder, examplesFolder, apiDocsTag string, ignoredExamples []string, swaggerSpec *SwaggerSpec, opMap map[string]APIEndpoint, allPerms []ResourcePermissions, missingExamples *[]string) []ResourcePermissions {
	files, err := ioutil.ReadDir(docsFolder)
	if err != nil {
		log.Printf("Warning: Failed to read folder %s: %v", docsFolder, err)
		return allPerms
	}

	for _, file := range files {
		shortResourceName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		resourceName := fmt.Sprintf("genesyscloud_%s", shortResourceName)
		fullFilePath := fmt.Sprintf("%s/%s", docsFolder, file.Name())

		if lists.ItemInSlice(resourceName, ignoredExamples) {
			os.Remove(fullFilePath)
			continue
		}

		examplesDir := fmt.Sprintf("%s/%s", examplesFolder, resourceName)
		if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
			log.Printf("No examples found! %s", resourceName)
			*missingExamples = append(*missingExamples, shortResourceName)
		}

		// Audit proxy and update apis.md with detected endpoints
		apiFileName := fmt.Sprintf("%s/apis.md", examplesDir)
		if opMap != nil {
			updateApisMdFromProxy(resourceName, apiFileName, opMap)
		}

		// Read the (potentially updated) apis.md file
		apiFileBytes, err := ioutil.ReadFile(apiFileName)
		if err != nil {
			fmt.Printf("Missing APIs file: %s\n", apiFileName)
			continue
		}

		// Read optional notes.md file for addendum content
		notesContent := readNotesFile(examplesDir)

		// Build the full content: endpoints + permissions + notes
		enhancedContent := string(apiFileBytes)
		if swaggerSpec != nil {
			endpoints := parseAPIEndpoints(string(apiFileBytes))
			if len(endpoints) > 0 {
				permissionsAndScopes := extractPermissionsAndScopes(endpoints, swaggerSpec)
				if permissionsAndScopes != "" {
					enhancedContent = enhancedContent + "\n" + permissionsAndScopes
				}

				resourcePerms := extractResourcePermissions(resourceName, shortResourceName, endpoints, swaggerSpec)
				if len(resourcePerms.Permissions) > 0 || len(resourcePerms.Scopes) > 0 {
					allPerms = append(allPerms, resourcePerms)
				}
			}
		}

		if notesContent != "" {
			enhancedContent = enhancedContent + "\n" + notesContent
		}

		// Open the doc file and replace the placeholder
		docFile, err := os.OpenFile(fullFilePath, os.O_RDWR, 0666)
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

		newBytes := bytes.Replace(docFileBytes, []byte(apiDocsTag), []byte(enhancedContent), 1)
		docFile.Truncate(0)
		docFile.WriteAt(newBytes, 0)
		fmt.Printf("Updated APIs in doc file: %s\n", file.Name())
	}

	return allPerms
}

// updateApisMdFromProxy scans the proxy file for a resource, detects API endpoints,
// and updates the apis.md file with any missing endpoints.
func updateApisMdFromProxy(resourceName, apisMdFile string, opMap map[string]APIEndpoint) {
	pkgName := strings.TrimPrefix(resourceName, "genesyscloud_")

	proxyFile := filepath.Join("genesyscloud", pkgName, fmt.Sprintf("genesyscloud_%s_proxy.go", pkgName))
	if _, err := os.Stat(proxyFile); os.IsNotExist(err) {
		return
	}

	proxyContent, err := ioutil.ReadFile(proxyFile)
	if err != nil {
		log.Printf("Warning: could not read %s: %v", proxyFile, err)
		return
	}

	detectedEndpoints := detectEndpointsFromProxy(string(proxyContent), opMap)
	if len(detectedEndpoints) == 0 {
		return
	}

	// Read current apis.md (may not exist yet)
	var currentContent string
	if data, err := ioutil.ReadFile(apisMdFile); err == nil {
		currentContent = string(data)
	}

	// Parse what's already documented
	documentedEndpoints := parseAPIEndpoints(currentContent)
	documentedSet := make(map[string]bool)
	for _, ep := range documentedEndpoints {
		key := strings.ToUpper(ep.Method) + " " + ep.Path
		documentedSet[key] = true
	}

	// Find missing endpoints
	var missing []APIEndpoint
	seen := make(map[string]bool)
	for _, ep := range detectedEndpoints {
		key := ep.Method + " " + ep.Path
		if !documentedSet[key] && !seen[key] {
			seen[key] = true
			missing = append(missing, ep)
		}
	}

	if len(missing) == 0 {
		return
	}

	fmt.Printf("  Proxy audit: adding %d missing endpoint(s) to %s apis.md\n", len(missing), resourceName)
	for _, ep := range missing {
		fmt.Printf("    + %s %s\n", ep.Method, ep.Path)
	}

	var buf strings.Builder
	buf.WriteString(strings.TrimSpace(currentContent))
	for _, ep := range missing {
		anchor := buildAnchor(ep.Method, ep.Path)
		buf.WriteString(fmt.Sprintf("\n* [%s %s](https://developer.genesys.cloud/devapps/api-explorer#%s)", ep.Method, ep.Path, anchor))
	}
	buf.WriteString("\n")

	ioutil.WriteFile(apisMdFile, []byte(buf.String()), 0644)
}


// readNotesFile reads the optional notes.md file from a resource's examples directory.
func readNotesFile(examplesDir string) string {
	notesFile := filepath.Join(examplesDir, "notes.md")
	data, err := ioutil.ReadFile(notesFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// buildAnchor generates a URL-safe anchor from an endpoint method and path.
func buildAnchor(method, path string) string {
	anchor := strings.ToLower(method) + "-" + strings.ReplaceAll(strings.ReplaceAll(path, "/", "-"), "{", "-")
	anchor = strings.ReplaceAll(anchor, "}", "-")
	return anchor
}

// writePermissionsJSON writes the permissions data to a JSON file
func writePermissionsJSON(permissions []ResourcePermissions, outputDir, filename, version string) error {
	sort.Slice(permissions, func(i, j int) bool {
		return permissions[i].ResourceType < permissions[j].ResourceType
	})

	data := PermissionsData{
		Version:   version,
		Resources: permissions,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	versionedFilename := fmt.Sprintf("%s-%s.json", filename, version)
	outputPath := filepath.Join(outputDir, versionedFilename)

	if err := ioutil.WriteFile(outputPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// buildOperationIdMap creates a map from operationId -> APIEndpoint using the Swagger spec
func buildOperationIdMap(spec *SwaggerSpec) map[string]APIEndpoint {
	opMap := make(map[string]APIEndpoint)
	for path, methods := range spec.Paths {
		for method, endpointSpec := range methods {
			if endpointSpec.OperationId == "" {
				continue
			}
			opMap[strings.ToLower(endpointSpec.OperationId)] = APIEndpoint{
				Method: strings.ToUpper(method),
				Path:   path,
			}
		}
	}
	return opMap
}

// detectEndpointsFromProxy extracts API endpoints from a proxy file by:
// 1. Matching SDK method calls (e.g., p.scriptsApi.GetScripts(...)) and looking up operationId
// 2. Matching raw HTTP calls with explicit /api/v2/ paths
// 3. Matching BasePath + "/api/v2/..." concatenation patterns
func detectEndpointsFromProxy(content string, opMap map[string]APIEndpoint) []APIEndpoint {
	var endpoints []APIEndpoint
	seen := make(map[string]bool)

	// Pattern 1: SDK method calls like p.someApi.MethodName(...) or a.api.MethodName(...)
	sdkCallRe := regexp.MustCompile(`\.[a-zA-Z]*[Aa]pi\.([A-Z][a-zA-Z0-9]+)\s*\(`)
	matches := sdkCallRe.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			methodName := match[1]
			if ep, ok := opMap[strings.ToLower(methodName)]; ok {
				key := ep.Method + " " + ep.Path
				if !seen[key] {
					seen[key] = true
					endpoints = append(endpoints, ep)
				}
			}
		}
	}

	// Pattern 2: Raw HTTP calls with explicit paths like "/api/v2/scripts/" + scriptId
	rawPathRe := regexp.MustCompile(`"(/api/v2/[^"]+)"`)
	rawMatches := rawPathRe.FindAllStringSubmatch(content, -1)
	for _, match := range rawMatches {
		if len(match) >= 2 {
			rawPath := match[1]
			httpMethod := detectHTTPMethodNearPath(content, match[0])
			if httpMethod != "" {
				normalizedPath := normalizeRawPath(rawPath)
				key := httpMethod + " " + normalizedPath
				if !seen[key] {
					seen[key] = true
					endpoints = append(endpoints, APIEndpoint{Method: httpMethod, Path: normalizedPath})
				}
			}
		}
	}

	// Pattern 3: BasePath + "/api/v2/..." concatenation patterns
	basePathRe := regexp.MustCompile(`BasePath\s*\+\s*"(/api/v2/[^"]+)"`)
	basePathMatches := basePathRe.FindAllStringSubmatch(content, -1)
	for _, match := range basePathMatches {
		if len(match) >= 2 {
			rawPath := match[1]
			httpMethod := detectHTTPMethodNearPath(content, match[0])
			if httpMethod != "" {
				normalizedPath := normalizeRawPath(rawPath)
				key := httpMethod + " " + normalizedPath
				if !seen[key] {
					seen[key] = true
					endpoints = append(endpoints, APIEndpoint{Method: httpMethod, Path: normalizedPath})
				}
			}
		}
	}

	return endpoints
}

// detectHTTPMethodNearPath looks for HTTP method indicators near a raw path usage
func detectHTTPMethodNearPath(content, pathMatch string) string {
	idx := strings.Index(content, pathMatch)
	if idx == -1 {
		return ""
	}

	start := idx - 200
	if start < 0 {
		start = 0
	}
	end := idx + len(pathMatch) + 200
	if end > len(content) {
		end = len(content)
	}
	context := content[start:end]

	if strings.Contains(context, "http.MethodDelete") || strings.Contains(context, "\"DELETE\"") {
		return "DELETE"
	}
	if strings.Contains(context, "http.MethodPost") || strings.Contains(context, "\"POST\"") {
		return "POST"
	}
	if strings.Contains(context, "http.MethodPut") || strings.Contains(context, "\"PUT\"") {
		return "PUT"
	}
	if strings.Contains(context, "http.MethodPatch") || strings.Contains(context, "\"PATCH\"") {
		return "PATCH"
	}
	if strings.Contains(context, "http.MethodGet") || strings.Contains(context, "\"GET\"") {
		return "GET"
	}

	return ""
}

// normalizeRawPath converts raw paths with variable concatenation into Swagger-style parameterized paths
// e.g., "/api/v2/scripts/" becomes "/api/v2/scripts/{scriptsId}" if followed by a variable
func normalizeRawPath(path string) string {
	if strings.HasSuffix(path, "/") {
		parts := strings.Split(strings.TrimRight(path, "/"), "/")
		if len(parts) > 0 {
			lastPart := parts[len(parts)-1]
			path = path + "{" + lastPart + "Id}"
		}
	}
	return path
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
		pathSpec, pathExists := spec.Paths[endpoint.Path]
		if !pathExists {
			continue
		}

		endpointSpec, methodExists := pathSpec[endpoint.Method]
		if !methodExists {
			continue
		}

		for _, perm := range endpointSpec.XIninRequiresPermissions.Permissions {
			if perm != "" {
				permissionsMap[perm] = true
			}
		}

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

	if len(permissionsMap) == 0 && len(scopesMap) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString("## Permissions and Scopes\n\n")

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


// extractResourcePermissions extracts permissions and scopes for a resource and returns structured data
func extractResourcePermissions(resourceType, resourceName string, endpoints []APIEndpoint, spec *SwaggerSpec) ResourcePermissions {
	permissionsMap := make(map[string]bool)
	scopesMap := make(map[string]bool)
	endpointsList := []string{}

	for _, endpoint := range endpoints {
		endpointsList = append(endpointsList, fmt.Sprintf("%s %s", strings.ToUpper(endpoint.Method), endpoint.Path))

		pathSpec, pathExists := spec.Paths[endpoint.Path]
		if !pathExists {
			continue
		}

		endpointSpec, methodExists := pathSpec[endpoint.Method]
		if !methodExists {
			continue
		}

		for _, perm := range endpointSpec.XIninRequiresPermissions.Permissions {
			if perm != "" {
				permissionsMap[perm] = true
			}
		}

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
