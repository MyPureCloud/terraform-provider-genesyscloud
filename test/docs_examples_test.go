package testing

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/provider_registrar"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type DependenciesConfig struct {
	Locals struct {
		// A list of Terraform configs that should be included to make the example resource pass the test
		Dependencies []string `hcl:"dependencies,optional"`
		// A reference to the existing working directory
		WorkingDir string `hcl:"working_dir,optional"`
	} `hcl:"locals,block"`
}

func TestExampleResources(t *testing.T) {

	resources := []string{
		//"genesyscloud_architect_datatable",
		//"genesyscloud_architect_datatable_row",
		// TODO "genesyscloud_architect_emergencygroup",
		//"genesyscloud_architect_grammar",
		//"genesyscloud_architect_grammar_language",
		// "genesyscloud_architect_ivr",
		// "genesyscloud_architect_schedulegroups",
		// "genesyscloud_architect_schedules",
		//"genesyscloud_flow",
		// "genesyscloud_telephony_providers_edges_did_pool",
	}

	planOnly, err := strconv.ParseBool(os.Getenv("TF_PLAN_ONLY"))
	if err != nil {
		planOnly = false
	}
	if planOnly {
		fmt.Fprintln(os.Stdout, "Sanity testing the resources defined in the examples...")
	} else {
		fmt.Fprintln(os.Stdout, "Acceptance testing the resources defined in the examples...")
	}

	providerResources, providerDataSources := provider_registrar.GetProviderResources()

	for _, example := range resources {
		exampleDir := filepath.Join("..", "examples", "resources", example)
		t.Run(exampleDir, func(t *testing.T) {

			// Get all tf files in the example directory
			files, err := filepath.Glob(filepath.Join(exampleDir, "*.tf"))
			if err != nil {
				t.Fatal(err)
			}
			if len(files) == 0 {
				t.Fatal("No tf files found in example directory " + exampleDir)
			}

			// Check for "resource.tf" in files and load it up
			if !lists.SubStringInSlice("resource.tf", files) {
				t.Fatal("resource.tf not found in example directory " + exampleDir)
			}
			resourceExampleContent, err := os.ReadFile(filepath.Join(exampleDir, "resource.tf"))
			if err != nil {
				t.Fatal(err)
			}

			resourceHCLFile, diagErr := hclsyntax.ParseConfig(resourceExampleContent, "resource.tf", hcl.Pos{Line: 1, Column: 1})
			if diagErr != nil && diagErr.HasErrors() {
				t.Fatal(diagErr)
			}
			resourceBlockType := resourceHCLFile.Body.(*hclsyntax.Body).Blocks[0].Labels[0]
			resourceBlockLabel := resourceHCLFile.Body.(*hclsyntax.Body).Blocks[0].Labels[1]
			resourceAttributes := resourceHCLFile.Body.(*hclsyntax.Body).Blocks[0].Body.Attributes

			// Check for optional "dependencies.tf" in files and load it up
			resourceExampleContent = append(resourceExampleContent, []byte("\n")...)
			resourceExampleContent = append(resourceExampleContent, checkForDependencies(t, exampleDir)...)

			fmt.Fprintln(os.Stdout, string(resourceExampleContent))

			// Add checks for the existence of each attribute defined in the example
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

			// Run test
			resource.Test(t, resource.TestCase{
				PreCheck: func() {
					util.TestAccPreCheck(t)
				},
				ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
				Steps: []resource.TestStep{
					{
						Config: string(resourceExampleContent),
						Check: resource.ComposeTestCheckFunc(
							checks...,
						),
						PlanOnly:           planOnly,
						ExpectNonEmptyPlan: planOnly,
					},
				},
			})

		})
	}
}

func checkForDependencies(t *testing.T, examplesDir string) []byte {

	content := []byte{}
	files, err := filepath.Glob(filepath.Join(examplesDir, "*.tf"))
	if err != nil {
		t.Fatal(err)
	}
	// Check for optional "dependencies.tf" in files and load it up
	if lists.SubStringInSlice("dependencies.tf", files) {
		var depConfig DependenciesConfig
		dependencyContentBody, err := os.ReadFile(filepath.Join(examplesDir, "dependencies.tf"))
		if err != nil {
			t.Fatal(err)
		}
		dependencyHCLFile, diagErr := hclsyntax.ParseConfig(dependencyContentBody, "dependencies.tf", hcl.Pos{Line: 1, Column: 1})
		if diagErr != nil && diagErr.HasErrors() {
			t.Fatal(diagErr)
		}
		diagErr = gohcl.DecodeBody(dependencyHCLFile.Body, nil, &depConfig)
		if diagErr != nil && diagErr.HasErrors() {
			t.Fatal(diagErr)
		}
		// Supports loading an existing example's resource file's content and dependencies
		if len(depConfig.Locals.Dependencies) > 0 {
			for _, dependency := range depConfig.Locals.Dependencies {
				dependencyExampleContent, err := os.ReadFile(filepath.Join(examplesDir, dependency))
				if err != nil {
					t.Fatal(err)
				}
				content = append(content, dependencyExampleContent...)
				content = append(content, []byte("\n")...)

				dependencyBasePath := filepath.Dir(filepath.Join(examplesDir, dependency))
				if examplesDir != dependencyBasePath {
					dependencyContent := checkForDependencies(t, dependencyBasePath)
					content = append(content, dependencyContent...)
					content = append(content, []byte("\n")...)
				}
			}
		}
		// Supports constructing a correct reference to the existing examples directory for referencing a filepath
		// from inside the Terraform config. See genesyscloud_flow for an example.
		if depConfig.Locals.WorkingDir != "" {
			wd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			// Write a locals output with the working directory pointing to the examples dir
			workingDirContent := fmt.Sprintf(`locals {
			  working_dir = "%s"
			}`, filepath.Join(wd, examplesDir, depConfig.Locals.WorkingDir))
			content = append(content, []byte(workingDirContent)...)
			content = append(content, []byte("\n")...)
		}
	}

	return content

}
