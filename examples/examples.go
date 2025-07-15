// Package examples provides functionality for loading, processing, and testing Terraform examples
// for the Genesys Cloud Terraform Provider. It handles dependency resolution, test generation,
// and conditional test execution based on environment constraints.
package examples

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/guide"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	zclCty "github.com/zclconf/go-cty/cty"
)

// Example represents a Terraform example configuration with its resources, dependencies,
// and constraints. It manages the loading and processing of example files.
type Example struct {
	// Locals contains local variables defined in the locals.tf file
	Locals *Locals
	// Resources contains the Terraform resources defined in the example
	Resources []*Resource
	// Dependencies contains other examples that this example depends on
	Dependencies []*Example
	// SkipIfConstraints defines conditions under which this example should be skipped during testing
	SkipIfConstraints *SkipIfConstraints
	// Warnings contains any warnings encountered during example processing
	Warnings []string
}

// NewExample creates a new initialized Example with empty slices for Resources and Dependencies.
// This is the starting point for loading and processing example files.
func NewExample() *Example {
	return &Example{
		Resources:    make([]*Resource, 0),
		Dependencies: make([]*Example, 0),
	}
}

// GenerateChecks creates a slice of resource.TestCheckFunc that validates all resources in the example.
// These check functions are used in Terraform acceptance tests to verify that resources were created correctly.
func (e *Example) GenerateChecks() []resource.TestCheckFunc {
	var checks []resource.TestCheckFunc

	for _, r := range e.Resources {
		resourceChecks := r.GenerateResourceChecks()
		checks = append(checks, resourceChecks...)
	}

	return checks
}

// GenerateOutput generates the complete Terraform configuration by combining all resources
// and locals into a single string. This is used to create the configuration that will be
// passed to Terraform during testing.
func (e *Example) GenerateOutput() (string, error) {
	var output string
	var err error
	// Add all resources to the output
	for _, resource := range e.Resources {
		output += resource.HCL + "\n\n"
	}
	// Add locals if they exist
	if e.Locals != nil {
		localOutput, err := e.Locals.generateOutput()
		if err != nil {
			return "", err
		}
		output += localOutput + "\n\n"
	}
	return output, err
}

// LoadExampleWithDependencies loads a Terraform example file and all of its dependencies.
// It handles cyclic dependency detection and ensures each file is processed only once.
// The processedState parameter tracks which files have been processed and the current dependency chain.
func (e *Example) LoadExampleWithDependencies(resourcePath string, processedState *ProcessedExampleState) (*Example, error) {

	// Check for cycles in dependency chain
	absPAth, err := filepath.Abs(resourcePath)
	if err != nil {
		return e, fmt.Errorf("error resolving absolute path for %s: %w", resourcePath, err)
	}

	// Check if this path is already in the dependency chain
	for _, path := range processedState.DependencyChain {
		if path == absPAth {
			return e, fmt.Errorf("cyclic dependency detected: %s", strings.Join(processedState.DependencyChain, " -> "))
		}
	}

	// Add this path to the dependency chain
	processedState.DependencyChain = append(processedState.DependencyChain, absPAth)

	// Remove this path from the dependency chain when the function is completed
	defer func() {
		if len(processedState.DependencyChain) > 0 {
			processedState.DependencyChain = processedState.DependencyChain[:len(processedState.DependencyChain)-1]
		}
	}()

	pathDir := filepath.Dir(resourcePath)
	fileWithoutExt := strings.TrimSuffix(filepath.Base(resourcePath), filepath.Ext(resourcePath))

	var processedExampleFiles *ProcessedFiles
	if processedState.DirectoryTracker[pathDir] == nil {
		processedExampleFiles = NewProcessedFiles()
		processedState.DirectoryTracker[pathDir] = processedExampleFiles
	} else {
		processedExampleFiles = processedState.DirectoryTracker[pathDir]
	}

	// Local locals
	locals, err := e.loadLocals(pathDir)
	if err != nil {
		return e, fmt.Errorf("failed to load locals: %s: %w", resourcePath, err)
	}
	// Update WorkingDir with absolute paths
	if locals.WorkingDir != nil {
		for key, relativePath := range locals.WorkingDir {
			absPath, err := filepath.Abs(filepath.Join(pathDir, relativePath))
			if err != nil {
				return e, fmt.Errorf("error resolving absolute path for %s: %w", key, err)
			}
			locals.WorkingDir[key] = absPath
		}
	}

	processedExampleFiles.markFileAsProcessed("locals.tf")
	e.Locals = locals

	if !processedExampleFiles.isFileProcessed(resourcePath) {

		// Load the resource.tf
		resource, err := e.loadResource(resourcePath)
		if err != nil {
			return e, fmt.Errorf("error loading %s: %w", resourcePath, err)
		}
		e.Resources = append(e.Resources, resource)

		// Mark as processed here so we don't double process a dependency before finishing up
		processedExampleFiles.markFileAsProcessed(resourcePath)

		if err := e.loadDependencies(pathDir, fileWithoutExt, locals, processedState); err != nil {
			return e, err
		}

		e.processSkipConstraints(locals)

		if err := e.setEnvironmentVars(locals); err != nil {
			return e, err
		}
	}

	return e, nil

}

func (e *Example) loadLocals(pathDir string) (*Locals, error) {

	var locals *Locals

	localsPath := filepath.Join(pathDir, "locals.tf")

	// Check if locals.tf exists
	if _, err := os.Stat(localsPath); err == nil {
		// locals.tf exists, parse it
		locals, err = e.parseLocals(localsPath)
		if err != nil {
			return nil, fmt.Errorf("error parsing locals.tf: %w", err)
		}
	} else if os.IsNotExist(err) {
		// locals.tf doesn't exist, create an empty Locals struct
		locals = NewLocals()
	} else {
		// Some other error occurred
		return nil, fmt.Errorf("error checking locals.tf: %w", err)
	}

	return locals, nil
}

func (e *Example) loadDependencies(pathDir string, fileWithoutExt string, locals *Locals, processedState *ProcessedExampleState) error {

	if locals.Dependencies == nil {
		return nil
	}

	for configFile, dependencies := range locals.Dependencies {
		if configFile != fileWithoutExt {
			continue
		}

		for _, depPath := range dependencies {
			depPath = filepath.Join(pathDir, depPath)
			depExample := NewExample()
			depExample, err := depExample.LoadExampleWithDependencies(depPath, processedState)
			if err != nil {
				// Check if this is a cycle error
				if strings.Contains(err.Error(), "cyclic dependency detected") {
					// If it is, log a warning and continue
					e.Warnings = append(e.Warnings, fmt.Sprintf("Warning: %s", err.Error()))
					continue
				}
				return fmt.Errorf("error loading dependency %s: %w", depPath, err)
			}
			if depExample == nil {
				continue
			}

			e.Dependencies = append(e.Dependencies, depExample)
			e.Resources = append(e.Resources, depExample.Resources...)
			e.Warnings = append(e.Warnings, depExample.Warnings...)
			if depExample.Locals != nil {
				if e.Locals == nil {
					e.Locals = NewLocals()
				}
				e.Locals.Merge(depExample.Locals)
			}
		}
	}

	return nil
}

func (e *Example) loadResource(path string) (*Resource, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(content, filepath.Base(path))
	if diags.HasErrors() {
		return nil, fmt.Errorf("error parsing HCL: %s", diags.Error())
	}

	return &Resource{
		FilePath: path,
		HCL:      string(content),
		AST:      file,
	}, nil
}

func (e *Example) parseLocals(filename string) (*Locals, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL file: %s", diags)
	}

	var config struct {
		Locals *Locals `hcl:"locals,block"`
	}

	diags = gohcl.DecodeBody(file.Body, nil, &config)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to decode HCL body: %s", diags)
	}

	if config.Locals == nil {
		return nil, fmt.Errorf("no locals block found")
	}

	return config.Locals, nil
}

func (e *Example) processSkipConstraints(locals *Locals) {
	if locals.SkipIfConstraints == nil {
		return
	}
	e.SkipIfConstraints = &SkipIfConstraints{}
	for k, v := range locals.SkipIfConstraints {
		if k == "products_missing_any" {
			e.SkipIfConstraints.ProductsMissingAny = append(e.SkipIfConstraints.ProductsMissingAny, v...)
		} else if k == "products_missing_all" {
			e.SkipIfConstraints.ProductsMissingAll = append(e.SkipIfConstraints.ProductsMissingAll, v...)
		} else if k == "products_existing_any" {
			e.SkipIfConstraints.ProductsExistingAny = append(e.SkipIfConstraints.ProductsExistingAny, v...)
		} else if k == "products_existing_all" {
			e.SkipIfConstraints.ProductsExistingAll = append(e.SkipIfConstraints.ProductsExistingAll, v...)
		} else if k == "not_in_domains" {
			e.SkipIfConstraints.NotInDomains = append(e.SkipIfConstraints.NotInDomains, v...)
		} else if k == "only_in_domains" {
			e.SkipIfConstraints.OnlyInDomains = append(e.SkipIfConstraints.OnlyInDomains, v...)
		} else if k == "feature_toggles_required" {
			e.SkipIfConstraints.FeatureTogglesRequired = append(e.SkipIfConstraints.FeatureTogglesRequired, v...)
		}
	}
}

func (e *Example) setEnvironmentVars(locals *Locals) error {
	if locals.EnvironmentVars == nil {
		return nil
	}

	for k, v := range locals.EnvironmentVars {
		if err := os.Setenv(k, v); err != nil {
			return fmt.Errorf("error setting environment variable %s: %w", k, err)
		}
	}

	return nil
}

// Resource represents a Terraform resource defined in an example file.
// It contains the file path, the raw HCL content, and the parsed AST for analysis.
type Resource struct {
	// FilePath is the path to the file containing the resource
	FilePath string
	// HCL is the raw HCL content of the resource
	HCL string
	// AST is the parsed abstract syntax tree of the resource
	AST *hcl.File
}

// GenerateResourceChecks creates TestCheckFuncs for a single resource
func (r *Resource) GenerateResourceChecks() []resource.TestCheckFunc {
	var checks []resource.TestCheckFunc

	content, _, diags := r.AST.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "resource", LabelNames: []string{"type", "name"}},
		},
	})
	if diags.HasErrors() {
		return []resource.TestCheckFunc{
			func(*terraform.State) error {
				return fmt.Errorf("error parsing resource block: %s", diags.Error())
			},
		}
	}

	for _, block := range content.Blocks {
		if block.Type == "resource" {
			resourceType := block.Labels[0]
			resourceName := block.Labels[1]

			// Type assert to access the Attributes
			body, ok := block.Body.(*hclsyntax.Body)
			if !ok {
				checks = append(checks, func(*terraform.State) error {
					return fmt.Errorf("unexpected body type for resource %s.%s", resourceType, resourceName)
				})
				continue
			}

			checks = append(checks, buildAttributeTestChecks(resourceType, resourceName, body.Attributes)...)
		}
	}

	return checks
}

type LocalsConfig struct {
	Locals *Locals `hcl:"locals,block"`
}

// Locals represents the contents of a locals.tf file in a Terraform example.
// It defines dependencies, working directories, test constraints, and environment variables.
type Locals struct {
	// Dependencies defines which other example files this example depends on
	Dependencies map[string][]string `hcl:"dependencies,optional"`
	// WorkingDir defines paths to working directories for resources that need file references
	WorkingDir map[string]string `hcl:"working_dir,optional"`
	// SkipIfConstraints defines conditions under which tests should be skipped
	SkipIfConstraints map[string][]string `hcl:"skip_if,optional"`
	// EnvironmentVars defines environment variables to set during testing
	EnvironmentVars map[string]string `hcl:"environment_vars,optional"`
	// ExtraAttributes captures any additional attributes defined in the locals block
	ExtraAttributes map[string]interface{} `hcl:",remain"`
}

func NewLocals() *Locals {
	return &Locals{
		WorkingDir:      make(map[string]string),
		ExtraAttributes: make(map[string]interface{}),
	}
}

// Constructs a single "locals {}" block based off of a map of merged working directories with locals.tf file.
// This is to prevent defining multiple locals {} blocks, which Terraform forbids.
func (l *Locals) generateOutput() (string, error) {

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	// Create the local block
	localsBlock := rootBody.AppendNewBlock("locals", []string{})

	// Create the working_dir objects
	workingDirObj := map[string]zclCty.Value{}
	for k, v := range l.WorkingDir {
		workingDirObj[k] = zclCty.StringVal(v)
	}

	// Add working_dir attribute to locals
	localsBlock.Body().SetAttributeValue("working_dir", zclCty.ObjectVal(workingDirObj))

	// Add remaining attributes
	for k, attr := range l.ExtraAttributes {
		hclAttr := attr.(*hcl.Attribute)

		value := hclAttr.Expr.Variables()
		if len(value) > 0 {
			localsBlock.Body().SetAttributeTraversal(k, value[0])
			continue
		} else {
			value, diags := hclAttr.Expr.Value(nil)
			if diags.HasErrors() {
				return "", diags
			}

			localsBlock.Body().SetAttributeValue(k, value)
		}
	}

	return string(f.Bytes()), nil

}

// Merge merges another Locals into this one
func (l *Locals) Merge(other *Locals) {
	if other == nil {
		return
	}

	if other.Dependencies != nil {
		if l.Dependencies == nil {
			l.Dependencies = make(map[string][]string)
		}
		for k, v := range other.Dependencies {
			if _, exists := l.Dependencies[k]; !exists {
				l.Dependencies[k] = make([]string, 0, len(v))
			}

			// Only add values that don't already exist
			for _, item := range v {
				if !lists.ItemInSlice(item, l.Dependencies[k]) {
					l.Dependencies[k] = append(l.Dependencies[k], item)
				}
			}
		}
	}

	if len(other.WorkingDir) > 0 {
		if l.WorkingDir == nil {
			l.WorkingDir = make(map[string]string)
		}
		for k, v := range other.WorkingDir {
			l.WorkingDir[k] = v
		}
	}

	if other.SkipIfConstraints != nil {
		if l.SkipIfConstraints == nil {
			l.SkipIfConstraints = make(map[string][]string)
		}
		for k, v := range other.SkipIfConstraints {
			if _, exists := l.SkipIfConstraints[k]; !exists {
				l.SkipIfConstraints[k] = make([]string, 0, len(v))
			}

			// Only add values that don't already exist
			for _, item := range v {
				if !lists.ItemInSlice(item, l.SkipIfConstraints[k]) {
					l.SkipIfConstraints[k] = append(l.SkipIfConstraints[k], item)
				}
			}
		}
	}

	if other.EnvironmentVars != nil {
		if l.EnvironmentVars == nil {
			l.EnvironmentVars = make(map[string]string)
		}
		for k, v := range other.EnvironmentVars {
			l.EnvironmentVars[k] = v
		}
	}

	// Any other attributes that could be defined
	if other.ExtraAttributes == nil {
		return
	}
	if l.ExtraAttributes == nil {
		l.ExtraAttributes = make(map[string]interface{})
	}
	for k, v := range other.ExtraAttributes {
		l.ExtraAttributes[k] = v
	}
}

// SkipIfConstraints defines conditions under which tests should be skipped.
// This allows examples to be conditionally tested based on the environment.
type SkipIfConstraints struct {
	// NotInDomains specifies domains where the test should not run
	NotInDomains []string `hcl:"not_in_domains,optional"`
	// OnlyInDomains specifies domains where the test should run
	OnlyInDomains []string `hcl:"only_in_domains,optional"`
	// ProductsMissingAny specifies that the test should run if any of these products are missing
	ProductsMissingAny []string `hcl:"products_missing_any,optional"`
	// ProductsMissingAll specifies that the test should run if all of these products are missing
	ProductsMissingAll []string `hcl:"products_missing_all,optional"`
	// ProductsExistingAny specifies that the test should run if any of these products exist
	ProductsExistingAny []string `hcl:"products_existing_any,optional"`
	// ProductsExistingAll specifies that the test should run if all of these products exist
	ProductsExistingAll []string `hcl:"products_existing_all,optional"`
	// FeatureTogglesRequired specifies feature toggles that must be enabled for the test to run
	FeatureTogglesRequired []string `hcl:"feature_toggles_required,optional"`
}

// ProcessedFiles tracks which files have been processed during example loading.
// This prevents processing the same file multiple times and helps detect cycles.
type ProcessedFiles struct {
	// Paths contains the base filenames of processed files
	Paths []string
}

// NewProcessedFiles creates a new initialized ProcessedFiles
func NewProcessedFiles() *ProcessedFiles {
	return &ProcessedFiles{
		Paths: make([]string, 0),
	}
}

func (p *ProcessedFiles) isFileProcessed(resourcePath string) bool {
	resourcePath = filepath.Base(resourcePath)
	for _, path := range p.Paths {
		if path == resourcePath {
			return true
		}
	}
	return false
}

func (p *ProcessedFiles) markFileAsProcessed(resourcePath string) {
	p.Paths = append(p.Paths, filepath.Base(resourcePath))
}

// ProcessedExampleState tracks the state of example processing across multiple directories.
// It maintains a record of processed files and the current dependency chain to detect cycles.
type ProcessedExampleState struct {
	// DirectoryTracker maps directory paths to their processed files
	DirectoryTracker map[string]*ProcessedFiles
	// DependencyChain tracks the current chain of dependencies being processed
	DependencyChain []string
}

// NewProcessedExampleState creates a new initialized ProcessedExampleState
func NewProcessedExampleState() *ProcessedExampleState {
	return &ProcessedExampleState{
		DirectoryTracker: make(map[string]*ProcessedFiles),
		DependencyChain:  make([]string, 0),
	}
}

// Creates a list of Test Checks for the existence of each attribute defined in the example
func buildAttributeTestChecks(resourceBlockType, resourceBlockLabel string, resourceAttributes hclsyntax.Attributes) []resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{}
	for _, attr := range resourceAttributes {
		attrName := attr.Name
		if attrName == "depends_on" {
			// Don't attempt to check for the `depends_on` attribute as that is not saved in the state
			continue
		}
		expr := attr.Expr
		switch expr.(type) {
		case *hclsyntax.ObjectConsExpr:
			// Handle maps
			check := resource.TestCheckResourceAttrSet(
				resourceBlockType+"."+resourceBlockLabel,
				attrName+".%",
			)
			checks = append(checks, check)
		case *hclsyntax.TupleConsExpr:
			// Handle lists/arrays
			check := resource.TestCheckResourceAttrSet(
				resourceBlockType+"."+resourceBlockLabel,
				attrName+".#",
			)
			checks = append(checks, check)
		default:
			// Handle all other types
			check := resource.TestCheckResourceAttrSet(
				resourceBlockType+"."+resourceBlockLabel,
				attrName,
			)
			checks = append(checks, check)
		}
	}
	return checks
}

// ShouldSkipExample creates a function to determine if the test should be skipped
func (e *Example) ShouldSkipExample(domain string, authorizationProducts []string) bool {

	if e.Locals == nil || e.SkipIfConstraints == nil {
		return false
	}
	shouldSkip := e.SkipIfConstraints.ShouldSkip(domain, authorizationProducts)
	if shouldSkip {
		fmt.Printf("Skipping test due to skip constraints: %v\n", e.SkipIfConstraints)
	}
	return shouldSkip
}

func (s *SkipIfConstraints) String() string {
	var constraints []string

	if len(s.NotInDomains) > 0 {
		constraints = append(constraints,
			fmt.Sprintf("Not in domains: %s", strings.Join(s.NotInDomains, ", ")))
	}

	if len(s.OnlyInDomains) > 0 {
		constraints = append(constraints,
			fmt.Sprintf("Only in domains: %s", strings.Join(s.OnlyInDomains, ", ")))
	}

	if len(s.ProductsMissingAny) > 0 {
		constraints = append(constraints,
			fmt.Sprintf("Missing one of these products: %s", strings.Join(s.ProductsMissingAny, ", ")))
	}

	if len(s.ProductsMissingAll) > 0 {
		constraints = append(constraints,
			fmt.Sprintf("Missing all of these products: %s", strings.Join(s.ProductsMissingAll, ", ")))
	}

	if len(s.ProductsExistingAny) > 0 {
		constraints = append(constraints,
			fmt.Sprintf("Existing one of these products: %s", strings.Join(s.ProductsExistingAny, ", ")))
	}

	if len(s.ProductsExistingAll) > 0 {
		constraints = append(constraints,
			fmt.Sprintf("Existing all of these products: %s", strings.Join(s.ProductsExistingAll, ", ")))
	}

	if len(s.FeatureTogglesRequired) > 0 {
		constraints = append(constraints,
			fmt.Sprintf("Feature toggles required: %s", strings.Join(s.FeatureTogglesRequired, ", ")))
	}

	if len(constraints) == 0 {
		return "No constraints"
	}

	return strings.Join(constraints, "\n")
}

func (s *SkipIfConstraints) ShouldSkip(domain string, authorizationProducts []string) bool {
	if len(s.NotInDomains) > 0 && !lists.ItemInSlice(domain, s.NotInDomains) {
		return true
	}
	if len(s.OnlyInDomains) > 0 && lists.ItemInSlice(domain, s.OnlyInDomains) {
		return true
	}
	if len(s.ProductsMissingAny) > 0 && !containsAny(authorizationProducts, s.ProductsMissingAny) {
		return true
	}
	if len(s.ProductsMissingAll) > 0 && !containsAll(authorizationProducts, s.ProductsMissingAll) {
		return true
	}
	if len(s.ProductsExistingAny) > 0 && containsAny(authorizationProducts, s.ProductsExistingAny) {
		return true
	}
	if len(s.ProductsExistingAll) > 0 && containsAll(authorizationProducts, s.ProductsExistingAll) {
		return true
	}
	if len(s.FeatureTogglesRequired) > 0 {
		for _, toggle := range s.FeatureTogglesRequired {
			if toggle == "guide" && !guide.GuideFtIsEnabled() {
				return true
			}
		}
	}
	return false
}

// Helper functions
func containsAny(slice, items []string) bool {
	for _, item := range items {
		if lists.ItemInSlice(item, slice) {
			return true
		}
	}
	return false
}

func containsAll(slice, items []string) bool {
	for _, item := range items {
		if !lists.ItemInSlice(item, slice) {
			return false
		}
	}
	return true
}
