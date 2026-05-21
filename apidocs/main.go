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
	swaggerURL              = "https://api.mypurecloud.com/api/v2/docs/swagger"
	proxyFileGlob           = "genesyscloud/*/genesyscloud_*_proxy.go"
	resourceExampleFolder   = "examples/resources"
	resourceDocsFolder      = "docs/resources"
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
			updateApisMdFromProxy(resourceName, examplesDir, apiFileName, opMap)
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
				// Extract and collect permissions and scopes for writing to API docs
				permissionsAndScopes := extractPermissionsAndScopes(endpoints, swaggerSpec)
				if permissionsAndScopes != "" {
					enhancedContent = enhancedContent + "\n" + permissionsAndScopes
				}

				// Extract and collect resource permissions for writing to JSON file
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

		// Replace the **No APIs** line with the enhanced content
		newBytes := bytes.Replace(docFileBytes, []byte(apiDocsTag), []byte(enhancedContent), 1)
		docFile.Truncate(0)
		docFile.WriteAt(newBytes, 0)
		fmt.Printf("Updated APIs in doc file: %s\n", file.Name())
	}

	return allPerms
}

// updateApisMdFromProxy scans the proxy file for a resource, detects API endpoints,
// and updates the apis.md file with any missing endpoints.
func updateApisMdFromProxy(resourceName, examplesDir, apisMdFile string, opMap map[string]APIEndpoint) {
	pkgName := strings.TrimPrefix(resourceName, "genesyscloud_")

	// Find the source file to scan for API calls
	sourceFile := findSourceFile(pkgName, examplesDir)
	if sourceFile == "" {
		return
	}

	proxyContent, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		log.Printf("Warning: could not read %s: %v", sourceFile, err)
		return
	}

	detectedEndpoints := detectEndpointsFromProxy(string(proxyContent), opMap)

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

	if len(missing) > 0 {
		fmt.Printf("  Proxy audit: adding %d missing endpoint(s) to %s apis.md\n", len(missing), resourceName)
		for _, ep := range missing {
			fmt.Printf("    + %s %s\n", ep.Method, ep.Path)
		}
	}

	// Combine documented and missing, then sort by path first, then method
	allEndpoints := append(documentedEndpoints, missing...)
	if len(allEndpoints) == 0 {
		return
	}

	sort.Slice(allEndpoints, func(i, j int) bool {
		if allEndpoints[i].Path != allEndpoints[j].Path {
			return allEndpoints[i].Path < allEndpoints[j].Path
		}
		return strings.ToUpper(allEndpoints[i].Method) < strings.ToUpper(allEndpoints[j].Method)
	})

	var buf strings.Builder
	for i, ep := range allEndpoints {
		anchor := buildAnchor(ep.Method, ep.Path)
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(fmt.Sprintf("* [%s %s](https://developer.genesys.cloud/devapps/api-explorer#%s)", strings.ToUpper(ep.Method), ep.Path, anchor))
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

// findSourceFile locates the proxy or source file for a given resource package name.
// It tries multiple strategies to handle naming inconsistencies.
func findSourceFile(pkgName, examplesDir string) string {
	// Strategy 0: Check for a .source file that explicitly points to the source
	sourcePointer := filepath.Join(examplesDir, ".source")
	if data, err := ioutil.ReadFile(sourcePointer); err == nil {
		relPath := strings.TrimSpace(string(data))
		resolved := filepath.Join(examplesDir, relPath)
		if _, err := os.Stat(resolved); err == nil {
			return resolved
		}
	}

	// Strategy 1: Exact package match with standard proxy naming
	candidates := []string{
		filepath.Join("genesyscloud", pkgName, fmt.Sprintf("genesyscloud_%s_proxy.go", pkgName)),
		filepath.Join("genesyscloud", pkgName, fmt.Sprintf("resource_genesyscloud_%s_proxy.go", pkgName)),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	// Strategy 2: Find a proxy file inside a matching sub-package directory
	// Handles cases like script->scripts, routing_sms_address->routing_sms_addresses
	matches, _ := filepath.Glob(filepath.Join("genesyscloud", pkgName+"*", "*_proxy.go"))
	if len(matches) > 0 {
		return matches[0]
	}

	// Strategy 3: Find proxy file where the package dir contains the name without underscores
	// Handles cases like external_contacts_organization -> externalcontacts_organization
	compactName := strings.ReplaceAll(pkgName, "_", "")
	allProxies, _ := filepath.Glob(filepath.Join("genesyscloud", "*", "*_proxy.go"))
	for _, p := range allProxies {
		dir := filepath.Base(filepath.Dir(p))
		if strings.ReplaceAll(dir, "_", "") == compactName {
			return p
		}
	}

	// Strategy 4: Legacy data source file in root package
	legacyFile := filepath.Join("genesyscloud", fmt.Sprintf("data_source_genesyscloud_%s.go", pkgName))
	if _, err := os.Stat(legacyFile); err == nil {
		return legacyFile
	}

	return ""
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

	// Build set of known Swagger paths for normalizeRawPath lookups
	swaggerPaths := make(map[string]bool)
	for _, ep := range opMap {
		swaggerPaths[ep.Path] = true
	}

	// Pattern 1: SDK method calls like p.someApi.MethodName(...), a.api.MethodName(...),
	// or standalone apiVar.MethodName(...)
	sdkCallRe := regexp.MustCompile(`(?:^|[^a-zA-Z])[a-zA-Z]*[Aa]pi\.([A-Z][a-zA-Z0-9]+)\s*\(`)
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
				normalizedPath := normalizeRawPath(rawPath, swaggerPaths)
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
				normalizedPath := normalizeRawPath(rawPath, swaggerPaths)
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

// normalizeRawPath converts raw paths with variable concatenation into Swagger-style parameterized paths.
// It looks up the trailing-slash path in the Swagger spec's known paths to find the correct parameter name.
// e.g., "/api/v2/scripts/" becomes "/api/v2/scripts/{scriptId}" by matching against the spec.
func normalizeRawPath(path string, swaggerPaths map[string]bool) string {
	if !strings.HasSuffix(path, "/") {
		return path
	}

	// Look for a Swagger path that starts with this prefix and has exactly one more {param} segment
	for swaggerPath := range swaggerPaths {
		if strings.HasPrefix(swaggerPath, path) {
			remainder := strings.TrimPrefix(swaggerPath, path)
			// Match if remainder is a single path parameter like "{scriptId}"
			if strings.HasPrefix(remainder, "{") && strings.HasSuffix(remainder, "}") && !strings.Contains(remainder, "/") {
				return swaggerPath
			}
		}
	}

	// Fallback: use the last segment name + "Id"
	parts := strings.Split(strings.TrimRight(path, "/"), "/")
	if len(parts) > 0 {
		path = path + "{" + parts[len(parts)-1] + "Id}"
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
