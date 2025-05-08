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
	zclCty "github.com/zclconf/go-cty/cty"
)

type Example struct {
	Locals            *Locals
	Resources         []*Resource
	Dependencies      []*Example
	SkipIfConstraints *SkipIfConstraints
}

type Resource struct {
	FilePath string
	HCL      string
	AST      *hcl.File
}

type LocalsConfig struct {
	Locals *Locals `hcl:"locals,block"`
}

type Locals struct {
	Dependencies      []string               `hcl:"dependencies,optional"`
	WorkingDir        map[string]string      `hcl:"working_dir,optional"`
	SkipIfConstraints map[string][]string    `hcl:"skip_if,optional"`
	EnvironmentVars   map[string]string      `hcl:"environment_vars,optional"`
	Other             map[string]interface{} `hcl:",remain"`
}
type SkipIfConstraints struct {
	NotInDomains       []string `hcl:"not_in_domains,optional"`
	OnlyInDomains      []string `hcl:"only_in_domains,optional"`
	ProductsMissingAny []string `hcl:"products_missing_any,optional"`
	ProductsMissingAll []string `hcl:"products_missing_all,optional"`
}

type ProcessedFiles struct {
	Paths []string
}
type ProcessedExampleState struct {
	DirectoryTracker map[string]*ProcessedFiles
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

// LoadExampleWithDependencies loads an example and its dependencies
func LoadExampleWithDependencies(resourcePath string, processedState *ProcessedExampleState) (*Example, *ProcessedExampleState, error) {

	if processedState == nil {
		processedState = &ProcessedExampleState{
			DirectoryTracker: make(map[string]*ProcessedFiles),
		}
	}

	pathDir := filepath.Dir(resourcePath)

	var processedExample *ProcessedFiles
	if processedState.DirectoryTracker[pathDir] == nil {
		processedExample = &ProcessedFiles{
			Paths: make([]string, 0),
		}
		processedState.DirectoryTracker[pathDir] = processedExample
	} else {
		processedExample = processedState.DirectoryTracker[pathDir]
	}

	var locals *Locals

	// Only process locals.tf if we haven't processed this directory yet
	if !processedExample.isFileProcessed("locals.tf") {

		localsPath := filepath.Join(pathDir, "locals.tf")

		// Check if locals.tf exists
		if _, err := os.Stat(localsPath); err == nil {
			// locals.tf exists, parse it
			locals, err = parseLocals(localsPath)
			if err != nil {
				return nil, processedState, fmt.Errorf("error parsing locals.tf: %w", err)
			}
		} else if os.IsNotExist(err) {
			// locals.tf doesn't exist, create an empty Locals struct
			locals = &Locals{
				WorkingDir: make(map[string]string),
				Other:      make(map[string]interface{}),
			}
		} else {
			// Some other error occurred
			return nil, processedState, fmt.Errorf("error checking locals.tf: %w", err)
		}

		processedExample.markFileAsProcessed("locals.tf")
	} else {
		// If we've already processed this directory, create an empty Locals struct
		locals = &Locals{
			WorkingDir: make(map[string]string),
			Other:      make(map[string]interface{}),
		}
	}

	// Update WorkingDir with absolute paths
	if locals.WorkingDir != nil {
		for key, relativePath := range locals.WorkingDir {
			absPath, err := filepath.Abs(filepath.Join(pathDir, relativePath))
			if err != nil {
				return nil, processedState, fmt.Errorf("error resolving absolute path for %s: %w", key, err)
			}
			locals.WorkingDir[key] = absPath
		}
	}

	example := &Example{
		Locals: locals,
	}

	if !processedExample.isFileProcessed(resourcePath) {

		// Load the resource.tf
		resource, err := loadResource(resourcePath)
		if err != nil {
			return nil, processedState, fmt.Errorf("error loading %s: %w", resourcePath, err)
		}
		example.Resources = append(example.Resources, resource)

		// Mark as processed here so we don't double process a dependency before finishing up
		processedExample.markFileAsProcessed(resourcePath)

		// Load dependencies
		if locals.Dependencies != nil {
			for _, depPath := range locals.Dependencies {
				depPath = filepath.Join(pathDir, depPath)
				depExample, depLoadedDeps, err := LoadExampleWithDependencies(depPath, processedState)
				if err != nil {
					return nil, processedState, fmt.Errorf("error loading dependency %s: %w", depPath, err)
				}
				if depExample == nil {
					continue
				}
				processedState = depLoadedDeps
				example.Dependencies = append(example.Dependencies, depExample)
				// Add resources from dependencies
				example.Resources = append(example.Resources, depExample.Resources...)

				// Merge locals
				example.Locals.merge(depExample.Locals)
			}
		}

		if locals.SkipIfConstraints != nil {
			example.SkipIfConstraints = &SkipIfConstraints{}
			for k, v := range locals.SkipIfConstraints {
				if k == "products_missing_any" {
					example.SkipIfConstraints.ProductsMissingAny = append(example.SkipIfConstraints.ProductsMissingAny, v...)
				} else if k == "products_missing_all" {
					example.SkipIfConstraints.ProductsMissingAll = append(example.SkipIfConstraints.ProductsMissingAll, v...)
				} else if k == "not_in_domains" {
					example.SkipIfConstraints.NotInDomains = append(example.SkipIfConstraints.NotInDomains, v...)
				} else if k == "only_in_domains" {
					example.SkipIfConstraints.OnlyInDomains = append(example.SkipIfConstraints.OnlyInDomains, v...)
				}
			}
		}

		if locals.EnvironmentVars != nil {
			for k, v := range locals.EnvironmentVars {
				err := os.Setenv(k, v)
				if err != nil {
					return nil, processedState, fmt.Errorf("error setting environment variable %s: %w", k, err)
				}
			}
		}

	}

	return example, processedState, nil
}

// merge merges another Locals into this one
func (l *Locals) merge(other *Locals) {
	if other == nil {
		return
	}

	l.Dependencies = append(l.Dependencies, other.Dependencies...)

	if l.WorkingDir == nil {
		l.WorkingDir = make(map[string]string)
	}
	for k, v := range other.WorkingDir {
		l.WorkingDir[k] = v
	}

	if l.SkipIfConstraints == nil {
		l.SkipIfConstraints = make(map[string][]string)
	}
	for k, v := range other.SkipIfConstraints {
		l.SkipIfConstraints[k] = append(l.SkipIfConstraints[k], v...)
	}

	if l.Other == nil {
		l.Other = make(map[string]interface{})
	}
	for k, v := range other.Other {
		l.Other[k] = v
	}
}

func parseLocals(filename string) (*Locals, error) {
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

func loadResource(path string) (*Resource, error) {
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

// GenerateOutput generates the complete Terraform configuration
func (e *Example) GenerateOutput() (string, error) {
	var output string
	var err error
	for _, resource := range e.Resources {
		output += resource.HCL + "\n\n"
	}
	if e.Locals != nil {
		localOutput, err := e.Locals.GenerateOutput()
		if err != nil {
			return "", err
		}
		output += localOutput + "\n\n"
	}
	return output, err
}

// Constructs a single "locals {}" block based off of a map of merged working directories with locals.tf file.
// This is to prevent defining multiple locals {} blocks, which Terraform forbids.
func (l *Locals) GenerateOutput() (string, error) {

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
	for k, attr := range l.Other {
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

// GenerateCheck creates a resource.TestCheckFunc that validates all resources
func (e *Example) GenerateChecks() []resource.TestCheckFunc {
	var checks []resource.TestCheckFunc

	for _, r := range e.Resources {
		resourceChecks := r.GenerateResourceChecks()
		checks = append(checks, resourceChecks...)
	}

	return checks
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

// Creates a list of Test Checks for the existence of each attribute defined in the example
func buildAttributeTestChecks(resourceBlockType, resourceBlockLabel string, resourceAttributes hclsyntax.Attributes) []resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{}
	for _, attr := range resourceAttributes {
		attrName := attr.Name
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

// GenerateSkipFunc creates a function to determine if the test should be skipped
func (e *Example) GenerateSkipFunc(domain string, authorizationProducts []string) func() (bool, error) {
	return func() (bool, error) {
		if e.Locals == nil || e.SkipIfConstraints == nil {
			return false, nil
		}
		shouldSkip := e.SkipIfConstraints.ShouldSkip(domain, authorizationProducts)
		if shouldSkip {
			fmt.Printf("Skipping test due to skip constraints: %v\n", e.SkipIfConstraints)
		}
		return shouldSkip, nil
	}
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

	if len(constraints) == 0 {
		return "No constraints"
	}

	return strings.Join(constraints, "\n")
}

func (s *SkipIfConstraints) ShouldSkip(domain string, authorizationProducts []string) bool {
	if len(s.NotInDomains) > 0 && !contains(s.NotInDomains, domain) {
		return true
	}
	if len(s.OnlyInDomains) > 0 && contains(s.OnlyInDomains, domain) {
		return true
	}
	if len(s.ProductsMissingAny) > 0 && !containsAny(authorizationProducts, s.ProductsMissingAny) {
		return true
	}
	if len(s.ProductsMissingAll) > 0 && !containsAll(authorizationProducts, s.ProductsMissingAll) {
		return true
	}
	return false
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsAny(slice, items []string) bool {
	for _, item := range items {
		if contains(slice, item) {
			return true
		}
	}
	return false
}

func containsAll(slice, items []string) bool {
	for _, item := range items {
		if !contains(slice, item) {
			return false
		}
	}
	return true
}
