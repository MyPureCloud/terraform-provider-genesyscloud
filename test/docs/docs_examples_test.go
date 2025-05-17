package testing

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

	"github.com/mypurecloud/terraform-provider-genesyscloud/examples"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider_registrar"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

// Set to TRUE to display the full output of the content being passed to Terraform with line numbers
// This is useful for debugging the output of the Terraform configuration
var SHOW_EXAMPLE_TERRAFORM_CONFIG_OUTPUT_WITH_LINES = false

func TestExampleResources(t *testing.T) {

	var domain string
	var authorizationProducts []string

	fmt.Fprintln(os.Stdout, "Acceptance testing the resources defined in the examples...")
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
	// resources = []string{
	// 	"genesyscloud_user_roles",
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

	successfulResourceTypes := make(map[string]string, len(resources))

	for _, resourceType := range resources {
		exampleDir := filepath.Join(testrunner.RootDir, "examples", "resources", resourceType)

		t.Run(resourceType, func(t *testing.T) {

			newExample := examples.NewExample()
			processedState := examples.NewProcessedExampleState()
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

			successfulResourceTypes[resourceType] = "success"

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
					successfulResourceTypes[resourceType] = "errored"
					return err
				},
				ProviderFactories: providerFactories,
				ExternalProviders: externalProviders,
				Steps: []resource.TestStep{
					{
						SkipFunc: func() (bool, error) {
							shouldSkip := example.ShouldSkipExample(domain, authorizationProducts)
							if shouldSkip {
								successfulResourceTypes[resourceType] = "skipped"
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
				successfulResourceTypes[resourceType] = "failed"
			}

			// Pause for five seconds to allow GC API to finish processing delete
			time.Sleep(time.Second * 5)

		})
	}

	// Sort successfulResourceTypes by key
	successfulResourceTypesKeys := make([]string, 0, len(successfulResourceTypes))
	for k := range successfulResourceTypes {
		successfulResourceTypesKeys = append(successfulResourceTypesKeys, k)
	}
	sort.Strings(successfulResourceTypesKeys)

	io.WriteString(os.Stdout, "The following resources were successfull:\n")
	for _, srtKey := range successfulResourceTypesKeys {
		status := successfulResourceTypes[srtKey]
		if status == "success" {
			io.WriteString(os.Stdout, fmt.Sprintf("  - %s\n", srtKey))
		}
	}
	io.WriteString(os.Stdout, "The following resources were errored:\n")
	for _, srtKey := range successfulResourceTypesKeys {
		status := successfulResourceTypes[srtKey]
		if status == "errored" {
			io.WriteString(os.Stdout, fmt.Sprintf("  - %s\n", srtKey))
		}
	}
	io.WriteString(os.Stdout, "The following resources were failed:\n")
	for _, srtKey := range successfulResourceTypesKeys {
		status := successfulResourceTypes[srtKey]
		if status == "failed" {
			io.WriteString(os.Stdout, fmt.Sprintf("  - %s\n", srtKey))
		}
	}
	io.WriteString(os.Stdout, "The following resources were skipped:\n")
	for _, srtKey := range successfulResourceTypesKeys {
		status := successfulResourceTypes[srtKey]
		if status == "skipped" {
			io.WriteString(os.Stdout, fmt.Sprintf("  - %s\n", srtKey))
		}
	}
}

func TestExampleResourcesPlanOnly(t *testing.T) {

	fmt.Fprintln(os.Stdout, "Sanity testing the resources defined in the examples...")

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
	combinedExample := examples.NewExample()
	processedState := examples.NewProcessedExampleState()

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
		resourceExample, err := examples.NewExample().LoadExampleWithDependencies(filepath.Join(exampleDir, "resource.tf"), processedState)
		if err != nil {
			t.Fatal(err)
		}

		// Manually merge the resource example into the combined example
		combinedExample.Resources = append(combinedExample.Resources, resourceExample.Resources...)
		if resourceExample.Locals != nil {
			if combinedExample.Locals == nil {
				combinedExample.Locals = examples.NewLocals()
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
					// SkipFunc: func() (bool, error) {
					// 	shouldSkip := example.ShouldSkipExample(domain, authorizationProducts)
					// 	return shouldSkip, nil
					// },
					Config:             string(resourceExampleContent),
					PlanOnly:           true,
					ExpectNonEmptyPlan: true,
				},
			},
		})

	})

}
