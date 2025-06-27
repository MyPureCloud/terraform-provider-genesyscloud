package examples

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider_registrar"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v161/platformclientv2"
)

// Set to TRUE to display the full output of the content being passed to Terraform with line numbers
// This is useful for debugging the output of the Terraform configuration
var SHOW_EXAMPLE_TERRAFORM_CONFIG_OUTPUT_WITH_LINES = false

type ResultsStatus string

const (
	ResultsStatusSuccess ResultsStatus = "success"
	ResultsStatusSkipped ResultsStatus = "skipped"
	ResultsStatusFailed  ResultsStatus = "failed"
	ResultsStatusErrored ResultsStatus = "errored"
)

func TestAccExampleResources(t *testing.T) {

	var domain string
	var authorizationProducts []string

	fmt.Fprintln(os.Stdout, "Acceptance testing the resources defined in the examples directory...")
	provider.AuthorizeSdk()
	authAPI := platformclientv2.NewAuthorizationApi()
	productEntities, api, err := authAPI.GetAuthorizationProducts()
	if err != nil {
		err = fmt.Errorf("Failed to get authorization products from the API: %s", err)
		t.Fatal(err)
	}
	authorizationProducts = make([]string, *productEntities.Total)
	for _, product := range *productEntities.Entities {
		authorizationProducts = append(authorizationProducts, *product.Id)
	}
	domain = strings.Join(strings.Split(api.Response.Request.URL.Host, ".")[1:], ".")

	providerResources, providerDataSources := provider_registrar.GetProviderResources()
	resources := provider_registrar.GetResourceTypeNames()
	// If you need to just test a specific resource type, you can manually override the resource(s)
	// under test by uncommenting these lines and updating them
	resources = []string{
		"genesyscloud_recording_media_retention_policy",
	}
	sort.Strings(resources)

	providerFactories := provider.GetProviderFactories(providerResources, providerDataSources)

	// Add some extra built in providers to be able to be used
	providerFactories = provider.CombineProviderFactories(providerFactories, ExampleUtilsProviderFactory())

	// External providers
	externalProviders := map[string]resource.ExternalProvider{
		"random": {
			Source:            "hashicorp/random",
			VersionConstraint: "3.7.2",
		},
		"time": {
			Source:            "hashicorp/time",
			VersionConstraint: "0.13.1",
		},
	}

	resourceTypesResults := make(map[string]ResultsStatus, len(resources))

	for _, resourceType := range resources {
		exampleDir := filepath.Join(testrunner.RootDir, "examples", "resources", resourceType)

		t.Run(resourceType, func(t *testing.T) {

			newExample := NewExample()
			processedState := NewProcessedExampleState()
			example, err := newExample.LoadExampleWithDependencies(filepath.Join(exampleDir, "resource.tf"), processedState)
			if err != nil {
				t.Fatal(err)
			}
			resourceExampleContent, err := example.GenerateOutput()
			if err != nil {
				t.Fatal(err)
			}
			checks := example.GenerateChecks()

			// Add arbitrary sleep to allow API to catch up before attempting to delete
			// Also provides a great place to place a breakpoint if needing to pause after Terraform Create and before Delete
			checks = append(checks, func(s *terraform.State) error {
				time.Sleep(2 * time.Second)
				return nil
			})

			resourceTypesResults[resourceType] = ResultsStatusSuccess

			// Run test
			resource.Test(t, resource.TestCase{
				PreCheck: func() {

					util.TestAccPreCheck(t)
					if SHOW_EXAMPLE_TERRAFORM_CONFIG_OUTPUT_WITH_LINES {
						// 12 is the number of lines the provider block (not shown) takes up before outputting the rest of the config
						// Retained for debugging purposes, allows the line numbers in error messages to line up.
						util.PrintStringWithLineNumbers(resourceExampleContent, 12)
					}

				},
				ErrorCheck: func(err error) error {
					resourceTypesResults[resourceType] = ResultsStatusErrored
					return err
				},
				ProviderFactories: providerFactories,
				ExternalProviders: externalProviders,
				Steps: []resource.TestStep{
					{
						SkipFunc: func() (bool, error) {
							shouldSkip := example.ShouldSkipExample(domain, authorizationProducts)
							if shouldSkip {
								resourceTypesResults[resourceType] = ResultsStatusSkipped
							}
							return shouldSkip, nil
						},
						Config: string(resourceExampleContent),
						Check: resource.ComposeTestCheckFunc(
							// arbitrary check with sleep
							checks...,
						),
					},
				},
			})
			if t.Failed() {
				resourceTypesResults[resourceType] = ResultsStatusFailed
			}

			// Pause for five seconds to allow GC API to finish processing delete
			time.Sleep(time.Second * 5)

		})
	}

	// Sort successfulResourceTypes by key
	successfulResourceTypesKeys := make([]string, 0, len(resourceTypesResults))
	for k := range resourceTypesResults {
		successfulResourceTypesKeys = append(successfulResourceTypesKeys, k)
	}
	sort.Strings(successfulResourceTypesKeys)

	io.WriteString(os.Stdout, "The following resources were successfull:\n")
	for _, srtKey := range successfulResourceTypesKeys {
		status := resourceTypesResults[srtKey]
		if status == ResultsStatusSuccess {
			io.WriteString(os.Stdout, fmt.Sprintf("  - %s\n", srtKey))
		}
	}
	io.WriteString(os.Stdout, "The following resources were errored:\n")
	for _, srtKey := range successfulResourceTypesKeys {
		status := resourceTypesResults[srtKey]
		if status == ResultsStatusErrored {
			io.WriteString(os.Stdout, fmt.Sprintf("  - %s\n", srtKey))
		}
	}
	io.WriteString(os.Stdout, "The following resources were failed:\n")
	for _, srtKey := range successfulResourceTypesKeys {
		status := resourceTypesResults[srtKey]
		if status == ResultsStatusFailed {
			io.WriteString(os.Stdout, fmt.Sprintf("  - %s\n", srtKey))
		}
	}
	io.WriteString(os.Stdout, "The following resources were skipped:\n")
	for _, srtKey := range successfulResourceTypesKeys {
		status := resourceTypesResults[srtKey]
		if status == ResultsStatusSkipped {
			io.WriteString(os.Stdout, fmt.Sprintf("  - %s\n", srtKey))
		}
	}
}

func TestUnitExampleResourcesPlanOnly(t *testing.T) {

	fmt.Fprintln(os.Stdout, "Sanity testing the resources defined in the examples directory...")

	providerResources, providerDataSources := provider_registrar.GetProviderResources()
	resources := provider_registrar.GetResourceTypeNames()
	sort.Strings(resources)

	providerFactories := provider.GetProviderFactories(providerResources, providerDataSources)

	// Add some extra built in providers to be able to be used
	providerFactories = provider.CombineProviderFactories(providerFactories, ExampleUtilsProviderFactory())

	// External providers
	externalProviders := map[string]resource.ExternalProvider{
		"random": {
			Source:            "hashicorp/random",
			VersionConstraint: "3.7.2",
		},
		"time": {
			Source:            "hashicorp/time",
			VersionConstraint: "0.13.1",
		},
	}

	// Create a combined example to hold all resources
	combinedExample := NewExample()
	processedState := NewProcessedExampleState()

	for _, resourceType := range resources {
		exampleDir := filepath.Join(testrunner.RootDir, "examples", "resources", resourceType)

		// Warn if the exampleDir doesn't exist
		if _, err := os.Stat(exampleDir); err != nil {
			if os.IsNotExist(err) {
				log.Printf("[WARN] Could not find an example directory for %s at %s", resourceType, exampleDir)
				continue
			}
		}

		// Load this resource's example
		resourceExample, err := NewExample().LoadExampleWithDependencies(filepath.Join(exampleDir, "resource.tf"), processedState)
		if err != nil {
			t.Fatal(err)
		}

		// Manually merge the resource example into the combined example
		combinedExample.Resources = append(combinedExample.Resources, resourceExample.Resources...)
		if resourceExample.Locals != nil {
			if combinedExample.Locals == nil {
				combinedExample.Locals = NewLocals()
			}
			combinedExample.Locals.Merge(resourceExample.Locals)
		}
	}

	t.Run("resources", func(t *testing.T) {

		resourceExampleContent, err := combinedExample.GenerateOutput()
		if err != nil {
			t.Fatal(err)
		}
		// Run test
		resource.Test(t, resource.TestCase{
			PreCheck: func() {
				if SHOW_EXAMPLE_TERRAFORM_CONFIG_OUTPUT_WITH_LINES {
					// 12 is the number of lines the provider block (not shown) takes up before outputting the rest of the config
					// Retained for debugging purposes, allows the line numbers in error messages to line up.
					util.PrintStringWithLineNumbers(resourceExampleContent, 17)
				}
			},
			ProviderFactories: providerFactories,
			ExternalProviders: externalProviders,
			Steps: []resource.TestStep{
				{
					Config:             string(resourceExampleContent),
					PlanOnly:           true,
					ExpectNonEmptyPlan: true,
				},
			},
		})

	})

}

// Utilize this test to explicitly test the simplest functionality available for each resource
func TestAccExampleResourcesAudit(t *testing.T) {

	fmt.Fprintln(os.Stdout, "Acceptance testing the resources defined in the examples directory...")

	providerResources, providerDataSources := provider_registrar.GetProviderResources()
	resources := provider_registrar.GetResourceTypeNames()
	// If you need to just test a specific resource type, you can manually override the resource(s)
	// under test by uncommenting these lines and updating them
	// resources := []string{
	// 	"genesyscloud_knowledge_category",
	// 	"genesyscloud_knowledge_document_variation",
	// 	"genesyscloud_knowledge_document",
	// 	"genesyscloud_knowledge_knowledgebase",
	// 	"genesyscloud_knowledge_label",
	// 	"genesyscloud_quality_forms_survey",
	// 	"genesyscloud_webdeployments_configuration",
	// 	"genesyscloud_webdeployments_deployment",
	// }
	sort.Strings(resources)

	providerFactories := provider.GetProviderFactories(providerResources, providerDataSources)

	// Add some extra built in providers to be able to be used
	providerFactories = provider.CombineProviderFactories(providerFactories, ExampleUtilsProviderFactory())

	// External providers
	externalProviders := map[string]resource.ExternalProvider{
		"random": {
			Source:            "hashicorp/random",
			VersionConstraint: "3.7.2",
		},
		"time": {
			Source:            "hashicorp/time",
			VersionConstraint: "0.13.1",
		},
	}

	provider.AuthorizeSdk()
	orgApi := platformclientv2.NewOrganizationApi()
	organization, _, _ := orgApi.GetOrganizationsMe()
	orgName := *organization.ThirdPartyOrgName

	resourceTypeResults := make(map[string]ResultsStatus, len(resources))

	for _, resourceType := range resources {
		exampleDir := filepath.Join(testrunner.RootDir, "examples", "resources", resourceType)

		t.Run(orgName+"/"+resourceType, func(t *testing.T) {

			newExample := NewExample()
			processedState := NewProcessedExampleState()
			resourceFilePath := filepath.Join(exampleDir, "simplest_resource.tf")
			if _, err := os.Stat(resourceFilePath); os.IsNotExist(err) {
				resourceFilePath = filepath.Join(exampleDir, "resource.tf")
				if _, err := os.Stat(resourceFilePath); os.IsNotExist(err) {
					t.Fatal("No resource.tf file found in the example directory")
				}
			}
			example, err := newExample.LoadExampleWithDependencies(resourceFilePath, processedState)
			if err != nil {
				t.Fatal(err)
			}
			resourceExampleContent, err := example.GenerateOutput()
			if err != nil {
				t.Fatal(err)
			}
			checks := example.GenerateChecks()

			// Add arbitrary sleep to allow API to catch up before attempting to delete
			// Also provides a great place to place a breakpoint if needing to pause after Terraform Create and before Delete
			checks = append(checks, func(s *terraform.State) error {
				time.Sleep(3 * time.Second)
				return nil
			})

			resourceTypeResults[resourceType] = ResultsStatusSuccess

			// Run test
			resource.Test(t, resource.TestCase{
				PreCheck: func() {
					util.TestAccPreCheck(t)
					if SHOW_EXAMPLE_TERRAFORM_CONFIG_OUTPUT_WITH_LINES {
						// 12 is the number of lines the provider block (not shown) takes up before outputting the rest of the config
						// Retained for debugging purposes, allows the line numbers in error messages to line up.
						util.PrintStringWithLineNumbers(resourceExampleContent, 12)
					}
				},
				ErrorCheck: func(err error) error {
					resourceTypeResults[resourceType] = ResultsStatusErrored
					return err
				},
				ProviderFactories: providerFactories,
				ExternalProviders: externalProviders,
				Steps: []resource.TestStep{
					{
						Config: string(resourceExampleContent),
						Check: resource.ComposeTestCheckFunc(
							// arbitrary check with sleep
							checks...,
						),
					},
				},
			})
			if t.Failed() {
				resourceTypeResults[resourceType] = ResultsStatusFailed
			}

			// Pause for five seconds to allow GC API to finish processing delete
			time.Sleep(time.Second * 7)

		})
	}

	// Sort successfulResourceTypes by key
	successfulResourceTypesKeys := make([]string, 0, len(resourceTypeResults))
	for k := range resourceTypeResults {
		successfulResourceTypesKeys = append(successfulResourceTypesKeys, k)
	}
	sort.Strings(successfulResourceTypesKeys)

	io.WriteString(os.Stdout, "The following resources were successfull:\n")
	for _, srtKey := range successfulResourceTypesKeys {
		status := resourceTypeResults[srtKey]
		if status == ResultsStatusSuccess {
			io.WriteString(os.Stdout, fmt.Sprintf("  - %s\n", srtKey))
		}
	}
	io.WriteString(os.Stdout, "The following resources were errored:\n")
	for _, srtKey := range successfulResourceTypesKeys {
		status := resourceTypeResults[srtKey]
		if status == ResultsStatusErrored {
			io.WriteString(os.Stdout, fmt.Sprintf("  - %s\n", srtKey))
		}
	}
	io.WriteString(os.Stdout, "The following resources were failed:\n")
	for _, srtKey := range successfulResourceTypesKeys {
		status := resourceTypeResults[srtKey]
		if status == ResultsStatusFailed {
			io.WriteString(os.Stdout, fmt.Sprintf("  - %s\n", srtKey))
		}
	}
	io.WriteString(os.Stdout, "The following resources were skipped:\n")
	for _, srtKey := range successfulResourceTypesKeys {
		status := resourceTypeResults[srtKey]
		if status == ResultsStatusSkipped {
			io.WriteString(os.Stdout, fmt.Sprintf("  - %s\n", srtKey))
		}
	}
}
