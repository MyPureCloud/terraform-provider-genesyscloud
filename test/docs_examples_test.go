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
		Dependencies []string `hcl:"dependencies"`
	} `hcl:"locals,block"`
}

func TestExampleResources(t *testing.T) {

	resources := []string{
		"genesyscloud_architect_datatable",
		"genesyscloud_architect_datatable_row",
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
				t.Fatal("No tf files found in example directory")
			}

			// Check for "resource.tf" in files and load it up
			if !lists.SubStringInSlice("resource.tf", files) {
				t.Fatal("resource.tf not found in example directory")
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
			resourceExampleContent = append(resourceExampleContent, checkForDependencies(t, exampleDir)...)

			// Add checks for the existence of each attribute defined in the example
			checks := []resource.TestCheckFunc{}
			for _, attr := range resourceAttributes {
				check := resource.TestCheckResourceAttrSet(resourceBlockType+"."+resourceBlockLabel, attr.Name)
				checks = append(checks, check)
			}

			// Run test
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { util.TestAccPreCheck(t) },
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

func checkForDependencies(t *testing.T, dir string) []byte {

	content := []byte{}
	files, err := filepath.Glob(filepath.Join(dir, "*.tf"))
	if err != nil {
		t.Fatal(err)
	}
	// Check for optional "dependencies.tf" in files and load it up
	if lists.SubStringInSlice("dependencies.tf", files) {
		var depConfig DependenciesConfig
		dependencyContentBody, err := os.ReadFile(filepath.Join(dir, "dependencies.tf"))
		if err != nil {
			t.Fatal(err)
		}
		dependencyHCLFile, diagErr := hclsyntax.ParseConfig(dependencyContentBody, "dependencies.tf", hcl.Pos{Line: 1, Column: 1})
		if diagErr != nil && diagErr.HasErrors() {
			t.Fatal(diagErr)
		}
		diagErr = gohcl.DecodeBody(dependencyHCLFile.Body, nil, &depConfig)
		if diagErr != nil && diagErr.HasErrors() {
			t.Fatal(err)
		}
		for _, dependency := range depConfig.Locals.Dependencies {
			dependencyExampleContent, err := os.ReadFile(filepath.Join(dir, dependency))
			if err != nil {
				t.Fatal(err)
			}
			content = append(content, dependencyExampleContent...)

			dependencyBasePath := filepath.Dir(filepath.Join(dir, dependency))
			dependencyContent := checkForDependencies(t, dependencyBasePath)
			content = append(content, dependencyContent...)
		}
	}

	return content

}
