package routing_language

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccFrameworkResourceRoutingLanguageBasic(t *testing.T) {
	var (
		langResourceLabel1 = "test-lang1"
		langName1          = "Terraform Framework Lang " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: getFrameworkProviderFactories(),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateFrameworkRoutingLanguageResource(
					langResourceLabel1,
					langName1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_language."+langResourceLabel1, "name", langName1),
					resource.TestCheckResourceAttrSet("genesyscloud_routing_language."+langResourceLabel1, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_language." + langResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: validateFrameworkLanguagesDestroyed,
	})
}

func TestAccFrameworkResourceRoutingLanguageForceNew(t *testing.T) {
	var (
		langResourceLabel = "test-lang-force-new"
		langName1         = "Terraform Framework Lang " + uuid.NewString()
		langName2         = "Terraform Framework Lang Updated " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: getFrameworkProviderFactories(),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateFrameworkRoutingLanguageResource(
					langResourceLabel,
					langName1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_language."+langResourceLabel, "name", langName1),
				),
			},
			{
				// Update (should force new)
				Config: generateFrameworkRoutingLanguageResource(
					langResourceLabel,
					langName2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_language."+langResourceLabel, "name", langName2),
				),
			},
		},
		CheckDestroy: validateFrameworkLanguagesDestroyed,
	})
}

func TestAccFrameworkResourceRoutingLanguageError(t *testing.T) {
	var (
		langResourceLabel = "test-lang-error"
		langName          = "" // Empty name should cause error
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: getFrameworkProviderFactories(),
		Steps: []resource.TestStep{
			{
				// Create with invalid name
				Config: generateFrameworkRoutingLanguageResource(
					langResourceLabel,
					langName,
				),
				ExpectError: regexp.MustCompile(".*"),
			},
		},
	})
}

// validateFrameworkLanguagesDestroyed checks that all routing languages have been destroyed
func validateFrameworkLanguagesDestroyed(state *terraform.State) error {
	routingApi := platformclientv2.NewRoutingApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_routing_language" {
			continue
		}

		lang, resp, err := routingApi.GetRoutingLanguage(rs.Primary.ID)
		if lang != nil {
			if lang.State != nil && *lang.State == "deleted" {
				// Language deleted
				continue
			}
			return fmt.Errorf("Framework routing language (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Language not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error checking Framework routing language: %s", err)
		}
	}
	// Success. All languages destroyed
	return nil
}

// generateFrameworkRoutingLanguageResource generates a routing language resource for Framework testing
func generateFrameworkRoutingLanguageResource(resourceLabel string, name string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_language" "%s" {
		name = "%s"
	}
	`, resourceLabel, name)
}

// getFrameworkProviderFactories returns provider factories for Framework testing
func getFrameworkProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"genesyscloud": func() (tfprotov6.ProviderServer, error) {
			// Create Framework provider with routing_language resource
			frameworkResources := map[string]func() frameworkresource.Resource{
				ResourceType: NewFrameworkRoutingLanguageResource,
			}
			frameworkDataSources := map[string]func() datasource.DataSource{
				ResourceType: NewFrameworkRoutingLanguageDataSource,
			}

			frameworkProvider := provider.NewFrameworkProvider("test", frameworkResources, frameworkDataSources)
			return providerserver.NewProtocol6(frameworkProvider())(), nil
		},
	}
}
