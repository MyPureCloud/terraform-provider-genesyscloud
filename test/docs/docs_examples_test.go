package testing

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider_registrar"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/test/docs/examples"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

// Set to TRUE to display the full output of the content being passed to Terraform with line numbers
// This is useful for debugging the output of the Terraform configuration
var SHOW_EXAMPLE_TERRAFORM_CONFIG_OUTPUT_WITH_LINES = true

func TestExampleResources(t *testing.T) {

	var domain string
	var authorizationProducts []string

	planOnly, err := strconv.ParseBool(os.Getenv("TF_PLAN_ONLY"))
	if err != nil {
		planOnly = false
	}
	if planOnly {
		fmt.Fprintln(os.Stdout, "Sanity testing the resources defined in the examples...")
	} else {
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
	}

	providerResources, providerDataSources := provider_registrar.GetProviderResources()
	resources := provider_registrar.GetResourceTypeNames()
	sort.Strings(resources)
	// resources = []string{
	// 	"genesyscloud_user_roles",
	// }
	// sort.Strings(resources)

	providerFactories := provider.GetProviderFactories(providerResources, providerDataSources)

	// Add some extra built in providers to be able to be used
	providerFactories = provider.CombineProviderFactories(providerFactories, UtilsProviderFactory())

	// Get absolute path of current working directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	successfulResourceTypes := make(map[string]string, len(resources))

	for _, resourceType := range resources {
		exampleDir := filepath.Join(wd, "..", "..", "examples", "resources", resourceType)

		t.Run(resourceType, func(t *testing.T) {

			example, _, err := examples.LoadExampleWithDependencies(filepath.Join(exampleDir, "resource.tf"), nil)
			if err != nil {
				t.Fatal(err)
			}
			resourceExampleContent, err := example.GenerateOutput()
			if err != nil {
				t.Fatal(err)
			}
			checks := example.GenerateChecks()

			if !planOnly {
				// Add arbitrary sleep to allow API to catch up before attempting to delete
				// Also provides a great place to place a breakpoint if needing to pause after Terraform Create and before Delete
				checks = append(checks, func(s *terraform.State) error {
					time.Sleep(2 * time.Second)
					return nil
				})
			}

			successfulResourceTypes[resourceType] = "success"

			// Run test
			resource.Test(t, resource.TestCase{
				PreCheck: func() {
					if !planOnly {
						util.TestAccPreCheck(t)
						if SHOW_EXAMPLE_TERRAFORM_CONFIG_OUTPUT_WITH_LINES {
							// 12 is the number of lines the provider block (not shown) takes up before outputting the rest of the config
							// Retained for debugging purposes, allows the line numbers in error messages to line up.
							util.PrintStringWithLineNumbers(resourceExampleContent, 12)
						}
					}
				},
				ErrorCheck: func(err error) error {
					successfulResourceTypes[resourceType] = "errored"
					return err
				},
				ProviderFactories: providerFactories,
				ExternalProviders: map[string]resource.ExternalProvider{
					"random": {
						Source:            "hashicorp/random",
						VersionConstraint: "3.7.2",
					},
					"time": {
						Source:            "hashicorp/time",
						VersionConstraint: "0.13.1",
					},
				},
				Steps: []resource.TestStep{
					{
						SkipFunc: func() (bool, error) {
							shouldSkip := example.GenerateSkipFunc(domain, authorizationProducts)
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
						PlanOnly:           planOnly,
						ExpectNonEmptyPlan: planOnly,
					},
				},
			})
			if t.Failed() {
				successfulResourceTypes[resourceType] = "failed"
			}

			if !planOnly {
				// Pause for five seconds to allow GC API to finish processing delete
				time.Sleep(time.Second * 5)
			}

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
