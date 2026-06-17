package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/examples"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
)

const (
	swaggerURL              = "https://api.mypurecloud.com/api/v2/docs/swagger"
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
	Version     string                `json:"version"`
	Resources   []ResourcePermissions `json:"resources"`
	DataSources []ResourcePermissions `json:"data_sources"`
}

// docsFolderProcessor represents a structure to pass to the processDocsFolder function
type docsFolderProcessor struct {
	docsFolder      string
	examplesFolder  string
	apiDocsTag      string
	auditOnly       bool
	isDataSource    bool
	ignoredExamples []string
	swaggerSpec     *SwaggerSpec
	opMap           map[string]APIEndpoint
	allPerms        []ResourcePermissions
	missingExamples *[]string
	errors          *[]string
}

// apidocs is a build tool that performs two coupled responsibilities:
//
//  1. Proxy Audit (source mutation): Scans proxy source files for SDK and custom API calls,
//     detects API endpoints via the Swagger spec's operation ID map, and updates
//     examples/resources/*/apis.md and examples/data-sources/*/apis.md with any missing endpoints.
//
//  2. Doc Injection (post-processing): Reads each apis.md, enriches it with permissions and
//     OAuth scopes from the Swagger spec, and injects the result into the generated docs
//     (docs/resources/, docs/data-sources/) by replacing the **No APIs** placeholder that
//     tfplugindocs writes from templates/resources.md.tmpl.
//
// These two steps are kept in a single tool because they share the Swagger spec (expensive to
// fetch), the operation ID map, and must run in sequence (audit before injection). The tool is
// designed to run as part of `go generate` after tfplugindocs, but the proxy audit step can
// also run standalone via `make generate-apidocs` to update apis.md files without regenerating
// full documentation via the -audit-only flag.
func main() {
	auditOnly := flag.Bool("audit-only", false, "Only run proxy audit to update apis.md files; skip doc injection and permissions JSON")
	flag.Parse()

	fmt.Println("Beginning to process API docs from examples directories ...")
	const (
		apiDocsTag     = "**No APIs**"
		outputDir      = "public/data"
		outputFilename = "resource_permissions"
	)

	// Get version from positional args (after flags)
	version := "latest"
	if flag.NArg() > 0 {
		version = flag.Arg(0)
	}

	missingExamples := []string{}
	var errors []string
	ignoredExamples := examples.GetIgnoredResources()

	// Slices to collect permissions separately for resources and data sources
	var resourcePermissions []ResourcePermissions
	var dataSourcePermissions []ResourcePermissions

	// Fetch Swagger specification once for all resources
	fmt.Println("Loading Swagger specification...")
	swaggerSpec, err := fetchSwaggerSpec(swaggerURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to fetch Swagger spec: %v. Continuing without permission/scope information.\n", err)
		swaggerSpec = nil
	}

	// Build operationId map for proxy auditing
	var opMap map[string]APIEndpoint
	if swaggerSpec != nil {
		opMap = buildOperationIdMap(swaggerSpec)
		fmt.Printf("\nBuilt operation ID map from Swagger spec with %d entries\n", len(opMap))
	}

	// Process resources
	fmt.Println("\nProcessing example resources directory...")
	processResourceDocs := docsFolderProcessor{
		docsFolder:      resourceDocsFolder,
		examplesFolder:  resourceExampleFolder,
		apiDocsTag:      apiDocsTag,
		auditOnly:       *auditOnly,
		isDataSource:    false,
		ignoredExamples: ignoredExamples,
		swaggerSpec:     swaggerSpec,
		opMap:           opMap,
		allPerms:        resourcePermissions,
		missingExamples: &missingExamples,
		errors:          &errors,
	}
	resourcePermissions = processDocsFolder(processResourceDocs)
	fmt.Printf("Processed all resources\n")

	// Process data sources
	fmt.Println("\nProcessing example data sources directory...")
	processDataSourceDocs := docsFolderProcessor{
		docsFolder:      dataSourceDocsFolder,
		examplesFolder:  dataSourceExampleFolder,
		apiDocsTag:      apiDocsTag,
		auditOnly:       *auditOnly,
		isDataSource:    true,
		ignoredExamples: ignoredExamples,
		swaggerSpec:     swaggerSpec,
		opMap:           opMap,
		allPerms:        dataSourcePermissions,
		missingExamples: &missingExamples,
		errors:          &errors,
	}
	dataSourcePermissions = processDocsFolder(processDataSourceDocs)
	fmt.Printf("Processed all data sources\n")

	// Write permissions data to JSON file
	if !*auditOnly && (len(resourcePermissions) > 0 || len(dataSourcePermissions) > 0) {
		fmt.Println()
		fmt.Println("Writing permissions data to JSON file...")
		if err := writePermissionsJSON(resourcePermissions, dataSourcePermissions, outputDir, outputFilename, version); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to write permissions JSON: %v\n", err)
		} else {
			outputFile := fmt.Sprintf("%s-%s.json", outputFilename, version)
			fmt.Printf("Permissions data written to: %s/%s\n", outputDir, outputFile)
		}
	}

	fmt.Println("-----")
	if len(ignoredExamples) > 0 {
		fmt.Printf("Ignored resources (no docs generated): %v\n", ignoredExamples)
	}
	if len(missingExamples) > 0 {
		fmt.Printf("Resources/data-sources without examples: %v\n", missingExamples)
	}

	// Surface collected errors
	if len(errors) > 0 {
		fmt.Println()
		fmt.Println("========================================")
		fmt.Printf("ERRORS: %d issue(s) found during processing:\n", len(errors))
		fmt.Println("========================================")
		for i, e := range errors {
			fmt.Printf("\n%d. %s\n", i+1, e)
		}
		fmt.Println()
		os.Exit(1)
	}
}

// processDocsFolder iterates over generated doc files in docs/resources/ or docs/data-sources/.
// It uses the doc folder as the iteration source (rather than examples/) because:
//   - It needs to delete doc files for ignored resources
//   - It ensures only resources with generated docs get API content injected
//
// For each doc file, it derives the resource name, runs the proxy audit against
// the corresponding examples/*/apis.md, then injects enriched content back into the doc file.
func processDocsFolder(docs docsFolderProcessor) []ResourcePermissions {
	// In audit-only mode, only iterate examples directories directly for proxy audit
	if docs.auditOnly {
		return processAuditOnly(docs)
	}

	files, err := os.ReadDir(docs.docsFolder)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to read folder %s: %v\n", docs.docsFolder, err)
		return docs.allPerms
	}

	for _, file := range files {
		shortResourceName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		resourceName := fmt.Sprintf("genesyscloud_%s", shortResourceName)
		fullFilePath := fmt.Sprintf("%s/%s", docs.docsFolder, file.Name())

		if lists.ItemInSlice(resourceName, docs.ignoredExamples) {
			os.Remove(fullFilePath)
			continue
		}

		examplesDir := fmt.Sprintf("%s/%s", docs.examplesFolder, resourceName)
		if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
			*docs.missingExamples = append(*docs.missingExamples, shortResourceName)
		}

		// Audit proxy and update apis.md with detected endpoints
		apiFileName := fmt.Sprintf("%s/apis.md", examplesDir)
		if docs.opMap != nil {
			updateApisMdFromProxy(resourceName, apiFileName, docs.opMap, docs.isDataSource, docs.errors)
		}

		// Read the (potentially updated) apis.md file
		apiFileBytes, err := os.ReadFile(apiFileName)
		if err != nil {
			*docs.errors = append(*docs.errors, fmt.Sprintf("Missing APIs file: %s", apiFileName))
			continue
		}

		// Read optional notes.md file for addendum content
		notesContent := readNotesFile(examplesDir)

		// Build the full document content: endpoints + permissions + notes

		// Start with APIs endpoints listing
		enhancedContent := stripSourcesComment(string(apiFileBytes))

		// Add Permissions/Scopes
		if docs.swaggerSpec != nil {
			endpoints := parseAPIEndpoints(string(apiFileBytes))
			if len(endpoints) > 0 {
				// Extract and collect permissions and scopes for writing to API docs
				permissionsAndScopes := extractPermissionsAndScopes(endpoints, docs.swaggerSpec)
				if permissionsAndScopes != "" {
					enhancedContent = enhancedContent + "\n" + permissionsAndScopes
				}

				// Extract and collect resource permissions for writing to JSON file
				resourcePerms := extractResourcePermissions(resourceName, shortResourceName, endpoints, docs.swaggerSpec)
				if len(resourcePerms.Permissions) > 0 || len(resourcePerms.Scopes) > 0 {
					docs.allPerms = append(docs.allPerms, resourcePerms)
				}
			}
		}

		// Add Notes
		if notesContent != "" {
			enhancedContent = enhancedContent + "\n" + notesContent
		}

		// Write document contents
		changed, err := writeDocFile(fullFilePath, docs.apiDocsTag, enhancedContent)
		if err != nil {
			*docs.errors = append(*docs.errors, fmt.Sprintf("Failed to write doc file %s: %v", fullFilePath, err))
			continue
		}
		if changed {
			fmt.Printf("Updated APIs in doc file: %s\n", file.Name())
		}

	}

	return docs.allPerms
}

// processAuditOnly iterates over examples directories directly (not docs/) to run
// only the proxy audit step. This is needed because in audit-only mode, the docs/
// directory may not exist (tfplugindocs hasn't run yet).
func processAuditOnly(docs docsFolderProcessor) []ResourcePermissions {
	files, err := os.ReadDir(docs.examplesFolder)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to read folder %s: %v\n", docs.examplesFolder, err)
		return docs.allPerms
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		resourceName := file.Name()
		if lists.ItemInSlice(resourceName, docs.ignoredExamples) {
			continue
		}
		apiFileName := filepath.Join(docs.examplesFolder, resourceName, "apis.md")
		if docs.opMap != nil {
			updateApisMdFromProxy(resourceName, apiFileName, docs.opMap, docs.isDataSource, docs.errors)
		}
	}

	return docs.allPerms
}

func writeDocFile(fullFilePath, apiDocsTag, enhancedContent string) (bool, error) {
	// Open the doc file and replace the placeholder
	docFile, err := os.OpenFile(fullFilePath, os.O_RDWR, 0666)
	if err != nil {
		return false, err
	}
	defer docFile.Close()

	docFileBytes, err := io.ReadAll(docFile)
	if err != nil {
		return false, err
	}

	// Replace the **No APIs** line with the enhanced content
	newBytes := bytes.Replace(docFileBytes, []byte(apiDocsTag), []byte(enhancedContent), 1)
	if bytes.Equal(docFileBytes, newBytes) {
		return false, nil
	}

	docFile.Truncate(0)
	_, err = docFile.WriteAt(newBytes, 0)
	return true, err
}

// updateApisMdFromProxy reads source file paths from the <!-- sources --> comment in apis.md,
// scans those files for API endpoints, and updates apis.md with any missing endpoints.
func updateApisMdFromProxy(resourceName, apisMdFile string, opMap map[string]APIEndpoint, isDataSource bool, errors *[]string) {
	// Read current apis.md
	var currentContent string
	if data, err := os.ReadFile(apisMdFile); err == nil {
		currentContent = string(data)
	} else {
		return
	}

	// Parse source files from the <!-- sources --> comment
	sourceFiles := parseSourcesComment(currentContent)
	if len(sourceFiles) == 0 {
		pkgName := strings.TrimPrefix(resourceName, "genesyscloud_")
		assumedPath := filepath.Join("genesyscloud", pkgName, fmt.Sprintf("genesyscloud_%s_proxy.go", pkgName))
		*errors = append(*errors, fmt.Sprintf(
			"%s is missing a <!-- sources --> comment. Add the following to the top of the file:\n"+
				"  <!-- sources\n  %s\n  -->\n"+
				"  If this resource does not use standard API proxy files, use NO_SOURCES instead:\n"+
				"  <!-- sources\n  NO_SOURCES\n  -->",
			apisMdFile, assumedPath))
		return
	}

	// Check for NO_SOURCES sentinel — skip proxy auditing entirely
	if len(sourceFiles) == 1 && sourceFiles[0] == noSourcesSentinel {
		return
	}

	// Validate and read all source files
	var allDetected []APIEndpoint
	for _, sf := range sourceFiles {
		if _, err := os.Stat(sf); err != nil {
			*errors = append(*errors, fmt.Sprintf(
				"Source file defined in %s does not exist: %s", apisMdFile, sf))
			continue
		}
		proxyContent, err := os.ReadFile(sf)
		if err != nil {
			*errors = append(*errors, fmt.Sprintf("Failed to read source file %s: %v", sf, err))
			continue
		}
		allDetected = append(allDetected, detectEndpointsFromProxy(string(proxyContent), opMap)...)
	}

	// For data sources, only consider GET endpoints
	if isDataSource {
		allDetected = filterGETEndpoints(allDetected)
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
	for _, ep := range allDetected {
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
	// For data sources, only keep GET endpoints
	if isDataSource {
		allEndpoints = filterGETEndpoints(allEndpoints)
	}
	if len(allEndpoints) == 0 {
		return
	}

	sort.Slice(allEndpoints, func(i, j int) bool {
		if allEndpoints[i].Path != allEndpoints[j].Path {
			return allEndpoints[i].Path < allEndpoints[j].Path
		}
		return strings.ToUpper(allEndpoints[i].Method) < strings.ToUpper(allEndpoints[j].Method)
	})

	// Rebuild the file preserving the sources comment
	var buf strings.Builder
	buf.WriteString("<!-- sources\n")
	for _, sf := range sourceFiles {
		buf.WriteString(sf + "\n")
	}
	buf.WriteString("-->\n")
	for _, ep := range allEndpoints {
		anchor := buildAnchor(ep.Method, ep.Path)
		buf.WriteString(fmt.Sprintf("* [%s %s](https://developer.genesys.cloud/devapps/api-explorer#%s)\n", strings.ToUpper(ep.Method), ep.Path, anchor))
	}

	if err := os.WriteFile(apisMdFile, []byte(buf.String()), 0644); err != nil {
		*errors = append(*errors, fmt.Sprintf("Failed to write %s: %v", apisMdFile, err))
	}
}

// readNotesFile reads the optional notes.md file from a resource's examples directory.
func readNotesFile(examplesDir string) string {
	notesFile := filepath.Join(examplesDir, "notes.md")
	data, err := os.ReadFile(notesFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// buildAnchor generates a URL-safe anchor from an endpoint method and path.
func buildAnchor(method, path string) string {
	anchor := strings.ToLower(method) + strings.ReplaceAll(strings.ReplaceAll(path, "/", "-"), "{", "-")
	anchor = strings.ReplaceAll(anchor, "}", "-")
	return anchor
}

// writePermissionsJSON writes the permissions data to a JSON file
func writePermissionsJSON(resources, dataSources []ResourcePermissions, outputDir, filename, version string) error {
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].ResourceType < resources[j].ResourceType
	})
	sort.Slice(dataSources, func(i, j int) bool {
		return dataSources[i].ResourceType < dataSources[j].ResourceType
	})

	// Create the data structure
	data := PermissionsData{
		Version:     version,
		Resources:   resources,
		DataSources: dataSources,
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
	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// parseSourcesComment extracts source file paths from the <!-- sources --> comment block
// at the top of an apis.md file. Returns nil if no comment is found.
// If the comment contains "NO_SOURCES" the resource is explicitly opted out of proxy auditing.
func parseSourcesComment(content string) []string {
	re := regexp.MustCompile(`(?s)<!--\s*sources\s*\n(.+?)-->`)
	match := re.FindStringSubmatch(content)
	if len(match) < 2 {
		return nil
	}

	var sources []string
	for _, line := range strings.Split(strings.TrimSpace(match[1]), "\n") {
		line = strings.TrimSpace(line)
		// Strip comment lines (e.g., "# comment" or "// comment" )
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		// Strip inline comments (e.g., "path/to/file.go # TODO: extract to proxy")
		if idx := strings.Index(line, " #"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
		// Add source to list of sources
		if line != "" {
			sources = append(sources, line)
		}
	}
	return sources
}

// stripSourcesComment removes the <!-- sources ... --> comment block from content.
func stripSourcesComment(content string) string {
	re := regexp.MustCompile(`(?s)<!--\s*sources\s*\n.+?-->\n?`)
	return re.ReplaceAllString(content, "")
}

const noSourcesSentinel = "NO_SOURCES"

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
// 2. Matching custom_api_client calls (e.g., customapi.Do[T](ctx, c, customapi.MethodGet, "/api/v2/...", ...))
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

	// Pattern 2: custom_api_client calls like:
	//   customapi.Do[T](ctx, c, customapi.MethodGet, "/api/v2/path", ...)
	//   customapi.DoNoResponse(ctx, c, customapi.MethodDelete, "/api/v2/path/" + id, ...)
	//   customapi.DoRaw(ctx, c, customapi.MethodGet, "/api/v2/path", ...)
	//   customapi.DoWithAcceptHeader(ctx, c, customapi.MethodGet, "/api/v2/path", ...)
	customApiRe := regexp.MustCompile(`customapi\.Do(?:NoResponse|Raw|WithAcceptHeader|\[[^\]]*\])\([^,]+,[^,]+,\s*customapi\.(Method\w+),\s*"(/api/v2/[^"]+)"`)
	customMatches := customApiRe.FindAllStringSubmatch(content, -1)
	for _, match := range customMatches {
		if len(match) >= 3 {
			httpMethod := customApiMethodToHTTP(match[1])
			rawPath := match[2]
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

// customApiMethodToHTTP converts a customapi.Method* constant name to an HTTP method string.
func customApiMethodToHTTP(methodConst string) string {
	switch methodConst {
	case "MethodGet":
		return "GET"
	case "MethodPost":
		return "POST"
	case "MethodPut":
		return "PUT"
	case "MethodPatch":
		return "PATCH"
	case "MethodDelete":
		return "DELETE"
	default:
		return ""
	}
}

// normalizeRawPath converts raw paths with variable concatenation into Swagger-style parameterized paths.
// It looks up the trailing-slash path in the Swagger spec's known paths to find the correct parameter name.
// e.g., "/api/v2/scripts/" becomes "/api/v2/scripts/{scriptId}" by matching against the spec.
// It also corrects casing mismatches by performing case-insensitive lookups.
func normalizeRawPath(path string, swaggerPaths map[string]bool) string {
	// Exact match — return as-is
	if swaggerPaths[path] {
		return path
	}

	// Case-insensitive match for non-trailing-slash paths
	if !strings.HasSuffix(path, "/") {
		lowerPath := strings.ToLower(path)
		for swaggerPath := range swaggerPaths {
			if strings.ToLower(swaggerPath) == lowerPath {
				return swaggerPath
			}
		}
		return path
	}

	// Trailing-slash: look for a Swagger path that starts with this prefix and has exactly one more {param} segment
	lowerPath := strings.ToLower(path)
	for swaggerPath := range swaggerPaths {
		if strings.HasPrefix(strings.ToLower(swaggerPath), lowerPath) {
			remainder := swaggerPath[len(path):]
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

const swaggerCacheFile = "apidocs/swagger_cache.json"

// fetchSwaggerSpec downloads and parses the Swagger specification, using a local cache.
// It sends an If-Modified-Since header based on the cache file's mtime. If the server
// returns 304 Not Modified, the cached version is used.
func fetchSwaggerSpec(url string) (*SwaggerSpec, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if info, err := os.Stat(swaggerCacheFile); err == nil {
		req.Header.Set("If-Modified-Since", info.ModTime().UTC().Format(http.TimeFormat))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Network failure — fall back to cache if available
		spec, cacheErr := loadCachedSpec(swaggerCacheFile)
		if cacheErr != nil {
			return nil, fmt.Errorf("failed to fetch swagger spec and no cache available: %w", err)
		}
		fmt.Println("Network error, using cached Swagger spec")
		return spec, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		fmt.Println("Swagger spec not modified, using cached version")
		return loadCachedSpec(swaggerCacheFile)
	}

	if resp.StatusCode != http.StatusOK {
		spec, cacheErr := loadCachedSpec(swaggerCacheFile)
		if cacheErr != nil {
			return nil, fmt.Errorf("unexpected status code %d and no cache available", resp.StatusCode)
		}
		fmt.Fprintf(os.Stderr, "Warning: Unexpected status code %d fetching Swagger spec, using cached version\n", resp.StatusCode)
		return spec, nil
	}

	fmt.Println("Swagger spec updated, downloading fresh copy...")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Write to cache
	if err := os.MkdirAll(filepath.Dir(swaggerCacheFile), 0755); err == nil {
		_ = os.WriteFile(swaggerCacheFile, body, 0644)
		// Set mtime from Last-Modified header if present
		if lm := resp.Header.Get("Last-Modified"); lm != "" {
			if t, err := time.Parse(http.TimeFormat, lm); err == nil {
				_ = os.Chtimes(swaggerCacheFile, t, t)
			}
		}
	}

	var spec SwaggerSpec
	if err := json.Unmarshal(body, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse swagger spec: %w", err)
	}

	return &spec, nil
}

// loadCachedSpec reads and parses the cached swagger spec file.
func loadCachedSpec(cacheFile string) (*SwaggerSpec, error) {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cached swagger spec: %w", err)
	}
	var spec SwaggerSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse cached swagger spec: %w", err)
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

// filterGETEndpoints returns only endpoints with GET method.
func filterGETEndpoints(endpoints []APIEndpoint) []APIEndpoint {
	var filtered []APIEndpoint
	for _, ep := range endpoints {
		if strings.ToUpper(ep.Method) == "GET" || strings.ToLower(ep.Method) == "get" {
			filtered = append(filtered, ep)
		}
	}
	return filtered
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
