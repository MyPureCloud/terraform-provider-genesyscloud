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
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	zclCty "github.com/zclconf/go-cty/cty"
)

func TestExampleResources(t *testing.T) {

	resources := []string{
		// "genesyscloud_architect_datatable",
		// "genesyscloud_architect_datatable_row",
		// "genesyscloud_architect_emergencygroup",
		// "genesyscloud_architect_grammar",
		// "genesyscloud_architect_grammar_language",
		// "genesyscloud_architect_ivr",
		// "genesyscloud_architect_schedulegroups",
		// "genesyscloud_architect_schedules",
		// "genesyscloud_architect_user_prompt",
		// "genesyscloud_auth_division",
		// "genesyscloud_auth_role",
		// // Requires Instagram Product "genesyscloud_conversations_messaging_integrations_instagram",
		// "genesyscloud_conversations_messaging_integrations_open",
		// // Requires Whatsapp Product "genesyscloud_conversations_messaging_integrations_whatsapp",
		// "genesyscloud_conversations_messaging_settings",
		// "genesyscloud_conversations_messaging_settings_default",
		// "genesyscloud_conversations_messaging_supportedcontent",
		// "genesyscloud_conversations_messaging_supportedcontent_default",
		// "genesyscloud_employeeperformance_externalmetrics_definitions",
		// "genesyscloud_externalcontacts_contact",
		// "genesyscloud_externalcontacts_external_source",
		// "genesyscloud_externalcontacts_organization",
		// "genesyscloud_flow",
		// "genesyscloud_flow_loglevel",
		// "genesyscloud_flow_milestone",
		// // No DELETE? "genesyscloud_flow_outcome",
		// "genesyscloud_group",
		// "genesyscloud_group_roles",
		// "genesyscloud_idp_adfs",
		// "genesyscloud_idp_generic",
		// "genesyscloud_idp_gsuite",
		// "genesyscloud_idp_okta",
		// "genesyscloud_idp_onelogin",
		// "genesyscloud_idp_ping",
		// "genesyscloud_idp_salesforce",
		// "genesyscloud_integration_credential",
		// "genesyscloud_integration",
		// "genesyscloud_integration_action",
		// "genesyscloud_integration_custom_auth_action",
		// "genesyscloud_integration_custom_auth_action",
		// "genesyscloud_integration_facebook",

		// // JOURNEY RESOURCE REQUIRE "journeyManagement" product
		// // Available with Cloud CX4 ?
		// // "genesyscloud_journey_action_map",
		// // "genesyscloud_journey_action_template",
		// // "genesyscloud_journey_outcome",
		// // "genesyscloud_journey_outcome_predictor",
		// // "genesyscloud_journey_segment",
		// // "genesyscloud_journey_view_schedule",
		// // "genesyscloud_journey_views",

		// "genesyscloud_knowledge_category",
		// "genesyscloud_knowledge_document",
		// "genesyscloud_knowledge_document_variation",
		// "genesyscloud_knowledge_knowledgebase",
		// "genesyscloud_knowledge_label",

		// "genesyscloud_location",
		// "genesyscloud_oauth_client",
		// "genesyscloud_organization_authentication_settings",
		// "genesyscloud_orgauthorization_pairing",

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

	// Add some extra built in providers to be able to be used
	providerFactories = provider.CombineProviderFactories(providerFactories, UtilsProviderFactory())

	for _, example := range resources {
		exampleDir := filepath.Join("..", "..", "examples", "resources", example)
		// testName := fmt.Sprintf("%s/%s", *providerMeta.Organization.ThirdPartyOrgName, exampleDir)
		t.Run(example, func(t *testing.T) {

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

			// Check for optional "locals.tf" in files and load it up
			resourceExampleContent = append(resourceExampleContent, []byte("\n")...)
			localsAndDependenciesContent, workingDirs, _, remainingLocalAttrs := checkForLocals(t, exampleDir, []string{}, map[string]interface{}{})
			resourceExampleContent = append(resourceExampleContent, localsAndDependenciesContent...)

			// Build locals {} block
			resourceExampleContent = append(resourceExampleContent, []byte("\n")...)
			resourceExampleContent = append(resourceExampleContent, constructLocals(workingDirs, remainingLocalAttrs)...)

			// Creates a list of Test Checks for the existence of each attribute defined in the example resources
			var checks []resource.TestCheckFunc
			for _, block := range resourceHCLFile.Body.(*hclsyntax.Body).Blocks {
				if block.Type == "resource" {

					resourceBlockType := block.Labels[0]
					resourceBlockLabel := block.Labels[1]
					resourceAttributes := block.Body.Attributes

					checks = append(checks, buildAttributeTestChecks(resourceBlockType, resourceBlockLabel, resourceAttributes)...)
				}
			}

			if !planOnly {
				// Add arbitrary sleep to allow API to catch up before attempting to delete
				// Also provides a great place to place a breakpoint if needing to pause after Terraform Create and before Delete
				checks = append(checks, func(s *terraform.State) error {
					time.Sleep(2 * time.Second)
					return nil
				})
			}

			// Run test
			resource.Test(t, resource.TestCase{
				PreCheck: func() {
					util.TestAccPreCheck(t)
					// Uncomment to display the full output of the content being passed to Terraform with line numbers
					// 12 is the number of lines the provider block (not shown) takes up before outputting the rest of the config
					// Retained for debugging purposes, allows the line numbers in error messages to line up.
					// util.PrintBytesWithLineNumbers(resourceExampleContent, 12)
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

// "locals.tf" file example:
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
		// Any extra remaining attributes defined, compliments of hcl.Body.PartialContent()
		// which is called in gohcl.DecodeBody()
		Remain map[string]interface{} `hcl:",remain"`
	} `hcl:"locals,block"`
}

// Recursive function that checks for a "locals.tf" file in the example directory and pulls in any extra dependent files.
// Any files referenced will check for a sibling "locals.tf" to load as well.
func checkForLocals(t *testing.T, examplesDir string, dependencyFilesInput []string, remainingAttrsInput map[string]interface{}) (content []byte, workingDirs map[string]string, dependencyFilesOutput []string, remainingAttrsOutput map[string]interface{}) {
	workingDirs = make(map[string]string)

	files, err := filepath.Glob(filepath.Join(examplesDir, "*.tf"))
	if err != nil {
		t.Fatal(err)
	}
	// Check for optional "locals.tf" in files and load it up
	if lists.SubStringInSlice("locals.tf", files) {
		var depConfig DependenciesConfig

		// Parse the locals.tf file into the depConfig object
		localsTfPath := filepath.Join(examplesDir, "locals.tf")
		dependencyContentBody, err := os.ReadFile(localsTfPath)
		if err != nil {
			t.Fatal(err)
		}
		dependencyHCLFile, diagErr := hclsyntax.ParseConfig(dependencyContentBody, "locals.tf", hcl.Pos{Line: 1, Column: 1})
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
						t.Fatalf("Cannot find the file \"%s\" referenced by the \"%s\" file.", dependencyPath, localsTfPath)
					}
					content = append(content, dependencyExampleContent...)
					content = append(content, []byte("\n")...)

					dependencyBasePath := filepath.Dir(dependencyPath)
					if examplesDir != dependencyBasePath {
						dependencyContent, depWorkingDirs, depPaths, depRemaining := checkForLocals(t, dependencyBasePath, dependencyFilesInput, remainingAttrsInput)
						remainingAttrsInput = updateRemainingAttrs(remainingAttrsInput, depRemaining)
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
			// Set the working directory pointing to the examples dir
			for k, v := range depConfig.Locals.WorkingDir {
				workingDirs[k] = filepath.Join(wd, examplesDir, v)
			}
		}

		if depConfig.Locals.Remain != nil {
			remainingAttrsInput = updateRemainingAttrs(remainingAttrsInput, depConfig.Locals.Remain)
		}
	}

	return content, workingDirs, dependencyFilesInput, remainingAttrsInput

}

// Update the remaining local attributes
func updateRemainingAttrs(allLocalAttributes map[string]interface{}, remainingAttrs map[string]interface{}) map[string]interface{} {
	for k, v := range remainingAttrs {
		allLocalAttributes[k] = v
	}
	return allLocalAttributes
}

// Constructs a single "locals {}" block based off of a map of merged working directories with locals.tf file.
// This is to prevent defining multiple locals {} blocks, which Terraform forbids.
func constructLocals(workingDirs map[string]string, remainingLocalAttrs map[string]interface{}) []byte {

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	// Create the local block
	localsBlock := rootBody.AppendNewBlock("locals", []string{})

	// Create the working_dir objects
	workingDirObj := map[string]zclCty.Value{}
	for k, v := range workingDirs {
		workingDirObj[k] = zclCty.StringVal(v)
	}

	// Add working_dir attribute to locals
	localsBlock.Body().SetAttributeValue("working_dir", zclCty.ObjectVal(workingDirObj))

	// Add remaining attributes
	for k, attr := range remainingLocalAttrs {
		hclAttr := attr.(*hcl.Attribute)

		// Add hclAttr to locals block without trying to evaluate them
		localsBlock.Body().SetAttributeTraversal(k, hclAttr.Expr.Variables()[0])
	}

	return f.Bytes()

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
