package examples

import (
	"fmt"
	"os"
	"path/filepath"
	"terraform-provider-genesyscloud/genesyscloud/provider"

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
	Locals       *Locals
	Resources    []*Resource
	Dependencies []*Example
	ProviderMeta *provider.ProviderMeta
}

type Resource struct {
	FilePath string
	HCL      string
	AST      *hcl.File
}

type Locals struct {
	Dependencies []string               `hcl:"dependencies,optional"`
	WorkingDir   map[string]string      `hcl:"working_dir,optional"`
	Constraints  *Constraints           `hcl:"constraints,block"`
	Other        map[string]interface{} `hcl:",remain"`
}

type Constraints struct {
	SkipIf SkipIfConstraints `hcl:"skip_if,block"`
}

type SkipIfConstraints struct {
	NotInDomains       []string `hcl:"not_in_domains,optional"`
	ProductsMissingAny []string `hcl:"products_missing_any,optional"`
	ProductsMissingAll []string `hcl:"products_missing_all,optional"`
}

// LoadExample loads an example and its dependencies
func LoadExample(resourcePath string, loadedDeps map[string]*Example) (*Example, error) {
	// If this directory has already been loaded, return the existing Example
	if ex, ok := loadedDeps[resourcePath]; ok {
		return ex, nil
	}

	exampleDir := filepath.Dir(resourcePath)
	localsPath := filepath.Join(exampleDir, "locals.tf")

	var locals *Locals
	var err error

	// Check if locals.tf exists
	if _, err := os.Stat(localsPath); err == nil {
		// locals.tf exists, parse it
		locals, err = parseLocals(localsPath)
		if err != nil {
			return nil, fmt.Errorf("error parsing locals.tf: %w", err)
		}
	} else if os.IsNotExist(err) {
		// locals.tf doesn't exist, create an empty Locals struct
		locals = &Locals{
			WorkingDir: make(map[string]string),
			Other:      make(map[string]interface{}),
		}
	} else {
		// Some other error occurred
		return nil, fmt.Errorf("error checking locals.tf: %w", err)
	}

	// Update WorkingDir with absolute paths
	if locals.WorkingDir != nil {
		for key, relativePath := range locals.WorkingDir {
			absPath, err := filepath.Abs(filepath.Join(exampleDir, relativePath))
			if err != nil {
				return nil, fmt.Errorf("error resolving absolute path for %s: %w", key, err)
			}
			locals.WorkingDir[key] = absPath
		}
	}

	example := &Example{
		Locals: locals,
	}

	// Load the main resource.tf
	mainResource, err := loadResource(resourcePath)
	if err != nil {
		return nil, fmt.Errorf("error loading main resource.tf: %w", err)
	}
	example.Resources = append(example.Resources, mainResource)

	// Load dependencies
	if locals.Dependencies != nil {
		for _, depPath := range locals.Dependencies {
			depPath = filepath.Join(exampleDir, depPath)
			depExample, err := LoadExample(depPath, loadedDeps)
			if err != nil {
				return nil, fmt.Errorf("error loading dependency %s: %w", depPath, err)
			}
			example.Dependencies = append(example.Dependencies, depExample)
			// Add resources from dependencies
			example.Resources = append(example.Resources, depExample.Resources...)

			// Merge locals
			example.Locals.merge(depExample.Locals)
		}
	}

	return example, nil
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

	if l.Constraints == nil {
		l.Constraints = &Constraints{}
	}
	l.Constraints.merge(other.Constraints)

	if l.Other == nil {
		l.Other = make(map[string]interface{})
	}
	for k, v := range other.Other {
		l.Other[k] = v
	}
}

func (c *Constraints) merge(other *Constraints) {
	if other == nil {
		return
	}
	c.SkipIf.merge(other.SkipIf)
}

func (s *SkipIfConstraints) merge(other SkipIfConstraints) {
	s.NotInDomains = mergeSlices(s.NotInDomains, other.NotInDomains)
	s.ProductsMissingAny = mergeSlices(s.ProductsMissingAny, other.ProductsMissingAny)
	s.ProductsMissingAll = mergeSlices(s.ProductsMissingAll, other.ProductsMissingAll)
}

func parseLocals(path string) (*Locals, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(content, filepath.Base(path))
	if diags.HasErrors() {
		return nil, fmt.Errorf("error parsing HCL: %s", diags.Error())
	}

	// Define the schema for the locals block
	schema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "locals"},
		},
	}

	contentBody, _, diags := file.Body.PartialContent(schema)
	if diags.HasErrors() {
		return nil, fmt.Errorf("error getting locals block: %s", diags.Error())
	}

	// There should be only one locals block
	if len(contentBody.Blocks) != 1 || contentBody.Blocks[0].Type != "locals" {
		return nil, fmt.Errorf("expected one locals block, found %d", len(contentBody.Blocks))
	}

	var locals Locals
	diags = gohcl.DecodeBody(contentBody.Blocks[0].Body, nil, &locals)
	if diags.HasErrors() {
		return nil, fmt.Errorf("error decoding locals block: %s", diags.Error())
	}

	return &locals, nil
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

		value, diags := hclAttr.Expr.Value(nil)
		if diags.HasErrors() {
			return "", diags
		}

		// Add hclAttr to locals block without trying to evaluate them
		localsBlock.Body().SetAttributeValue(k, value)
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
func (e *Example) GenerateSkipFunc() func() (bool, error) {
	return func() (bool, error) {
		e.ProviderMeta = provider.GetProviderMeta()
		if e.Locals == nil || e.Locals.Constraints == nil {
			return false, nil
		}
		return e.Locals.Constraints.ShouldSkip(e.ProviderMeta), nil
	}
}

// ShouldSkip determines if the test should be skipped based on constraints
func (c *Constraints) ShouldSkip(providerMeta *provider.ProviderMeta) bool {
	if c == nil {
		return false
	}
	return c.SkipIf.shouldSkip(providerMeta)
}

func (s *SkipIfConstraints) shouldSkip(providerMeta *provider.ProviderMeta) bool {
	if len(s.NotInDomains) > 0 && !contains(s.NotInDomains, providerMeta.Domain) {
		return true
	}
	if len(s.ProductsMissingAny) > 0 && !containsAny(providerMeta.AuthorizationProducts, s.ProductsMissingAny) {
		return true
	}
	if len(s.ProductsMissingAll) > 0 && !containsAll(providerMeta.AuthorizationProducts, s.ProductsMissingAll) {
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

func mergeSlices(slices ...[]string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, slice := range slices {
		for _, item := range slice {
			if !seen[item] {
				seen[item] = true
				result = append(result, item)
			}
		}
	}
	return result
}
