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
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Example:
//
//	locals {
//	  dependencies = [ ... ]
//	  working_dir = {
//	    auth_role = "./genesyscloud_auth_role/"
//	    user = "./genesyscloud_user/"
//	  }
//	}
type DependenciesConfig struct {
	Locals struct {
		// A list of Terraform configs that should be included to make the example resource pass the test
		Dependencies []string `hcl:"dependencies,optional"`
		// A reference to the existing working directory
		WorkingDir map[string]string `hcl:"working_dir,optional"`
	} `hcl:"locals,block"`
}

func TestExampleResources(t *testing.T) {

	resources := []string{
		// "genesyscloud_architect_datatable",
		// "genesyscloud_architect_datatable_row",
		// // TODO "genesyscloud_architect_emergencygroup",
		// "genesyscloud_architect_grammar",
		// "genesyscloud_architect_grammar_language",
		// "genesyscloud_architect_ivr",
		// "genesyscloud_architect_schedulegroups",
		// "genesyscloud_architect_schedules",
		// "genesyscloud_architect_user_prompt",
		// "genesyscloud_auth_division",
		// "genesyscloud_auth_role",
		// Requires Instagram Product "genesyscloud_conversations_messaging_integrations_instagram",
		"genesyscloud_conversations_messaging_integrations_instagram",
		"genesyscloud_conversations_messaging_integrations_open",
		// Requires Whatsapp Product "genesyscloud_conversations_messaging_integrations_whatsapp",
		"genesyscloud_conversations_messaging_integrations_whatsapp",
		"genesyscloud_conversations_messaging_settings",
		"genesyscloud_conversations_messaging_settings_default",
		"genesyscloud_conversations_messaging_supportedcontent",
		"genesyscloud_conversations_messaging_supportedcontent_default",
		"genesyscloud_employeeperformance_externalmetrics_definitions",
		// "genesyscloud_externalcontacts_contact",
		// "genesyscloud_externalcontacts_external_source",
		// "genesyscloud_externalcontacts_organization",
		// "genesyscloud_flow",
		// "genesyscloud_group",
		// "genesyscloud_location",
		// "genesyscloud_routing_language",
		// "genesyscloud_routing_queue",
		// "genesyscloud_routing_skill",
		// "genesyscloud_routing_sms_address",
		// "genesyscloud_routing_utilization",
		// "genesyscloud_routing_utilization_label",
		// "genesyscloud_routing_wrapupcode",
		// "genesyscloud_script",
		// "genesyscloud_telephony_providers_edges_did_pool",
		// "genesyscloud_user",
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
	providerFactories := provider.GetProviderFactories(providerResources, providerDataSources)

	// providerMeta := provider.GetProviderMeta()

	for _, example := range resources {
		exampleDir := filepath.Join("..", "..", "examples", "resources", example)
		// testName := fmt.Sprintf("%s/%s", *providerMeta.Organization.ThirdPartyOrgName, exampleDir)
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
			dependenciesContent, workingDirs, _ := checkForDependencies(t, exampleDir, []string{})
			resourceExampleContent = append(resourceExampleContent, dependenciesContent...)

			// Build locals {} block
			resourceExampleContent = append(resourceExampleContent, []byte("\n")...)
			resourceExampleContent = append(resourceExampleContent, constructLocals(workingDirs)...)

			// Uncomment to display the full output of the content being passed to Terraform
			// Retained for debugging purposes
			// fmt.Fprintln(os.Stdout, string(resourceExampleContent))

			// Creates a list of Test Checks for the existence of each attribute defined in the example
			checks := buildAttributeTestChecks(resourceBlockType, resourceBlockLabel, resourceAttributes)

			if !planOnly {
				// Add arbitrary sleep to allow API to catch up before attempting to delete
				checks = append(checks, func(s *terraform.State) error {
					time.Sleep(1 * time.Second)
					return nil
				})
			}

			// Run test
			resource.Test(t, resource.TestCase{
				PreCheck: func() {
					util.TestAccPreCheck(t)
				},
				ProviderFactories: providerFactories,
				ExternalProviders: map[string]resource.ExternalProvider{
					"random": {
						Source:            "hashicorp/random",
						VersionConstraint: "3.7.2",
					},
				},
				Steps: []resource.TestStep{
					{
						Config: string(resourceExampleContent),
						Check: resource.ComposeTestCheckFunc(
							// arbitrary check with sleep
							checks...,
						),
						PlanOnly:           planOnly,
						ExpectNonEmptyPlan: planOnly,
					},
				},
			})

			if !planOnly {
				// Pause for five seconds to allow GC API to finish processing delete
				time.Sleep(time.Second * 5)
			}

		})
	}
}

// Recursive function that checks for a "dependencies.tf" file in the example directory and pulls in any extra dependent files.
// Any files referenced will check for a sibling "dependencies.tf" to load as well.
func checkForDependencies(t *testing.T, examplesDir string, dependencyFilesInput []string) (content []byte, workingDirs map[string]string, dependencyFilesOutput []string) {
	workingDirs = make(map[string]string)

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
				dependencyPath := filepath.Join(examplesDir, dependency)
				if !lists.ItemInSlice(dependencyPath, dependencyFilesInput) {
					dependencyFilesInput = append(dependencyFilesInput, dependencyPath)
					dependencyExampleContent, err := os.ReadFile(dependencyPath)
					if err != nil {
						t.Fatal(err)
					}
					content = append(content, dependencyExampleContent...)
					content = append(content, []byte("\n")...)

					dependencyBasePath := filepath.Dir(dependencyPath)
					if examplesDir != dependencyBasePath {
						dependencyContent, depWorkingDirs, depPaths := checkForDependencies(t, dependencyBasePath, dependencyFilesInput)
						content = append(content, dependencyContent...)
						content = append(content, []byte("\n")...)
						for k, v := range depWorkingDirs {
							workingDirs[k] = v
						}
						dependencyFilesInput = depPaths
					}
				}
			}
		}
		// Supports constructing a correct reference to the existing examples directory for referencing a filepath
		// from inside the Terraform config. See genesyscloud_flow for an example.
		if depConfig.Locals.WorkingDir != nil {
			wd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			// Write a locals output with the working directory pointing to the examples dir
			for k, v := range depConfig.Locals.WorkingDir {
				workingDirs[k] = filepath.Join(wd, examplesDir, v)
			}
		}
	}

	return content, workingDirs, dependencyFilesInput

}

// Constructs the "locals {}" block based off of a map of merged working directories. This is to
// prevent duplication of attributes, which Terraform forbids.
func constructLocals(workingDirs map[string]string) []byte {
	locals := []byte{}
	locals = append(locals, []byte("locals {\n")...)
	locals = append(locals, []byte("  working_dir = {\n")...)
	for k, v := range workingDirs {
		locals = append(locals, []byte(fmt.Sprintf("    %s = \"%s\"\n", k, v))...)
	}
	locals = append(locals, []byte("  }\n")...)
	locals = append(locals, []byte("}\n")...)
	return locals
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
