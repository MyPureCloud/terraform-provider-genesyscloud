package webdeployments_deployment

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceWebDeploymentsDeployment(t *testing.T) {
	t.Parallel()
	var (
		deploymentName        = "Test Deployment " + util.RandString(8)
		deploymentDescription = "Test Deployment description " + util.RandString(32)
		fullResourceName      = "genesyscloud_webdeployments_deployment.basic"
	)

	cleanupWebDeploymentsDeployment(t, "Test Deployment ")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: basicDeploymentResource(deploymentName, deploymentDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", deploymentName),
					resource.TestCheckResourceAttr(fullResourceName, "description", deploymentDescription),
					resource.TestCheckResourceAttr(fullResourceName, "allow_all_domains", "true"),
					resource.TestCheckNoResourceAttr(fullResourceName, "allowed_domains"),
					resource.TestMatchResourceAttr(fullResourceName, "status", regexp.MustCompile("^(Pending|Active)$")),
				),
			},
			{
				ResourceName:            fullResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"status"},
			},
		},
		CheckDestroy: verifyLanguagesDestroyed,
	})
}

func TestAccResourceWebDeploymentsDeployment_AllowedDomains(t *testing.T) {
	t.Parallel()
	var (
		deploymentName   = "Test Deployment " + util.RandString(8)
		fullResourceName = "genesyscloud_webdeployments_deployment.basicWithAllowedDomains"
		firstDomain      = "genesys-" + util.RandString(8) + ".com"
		secondDomain     = "genesys-" + util.RandString(8) + ".com"
	)

	cleanupWebDeploymentsDeployment(t, "Test Deployment ")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: deploymentResourceWithAllowedDomains(t, deploymentName, firstDomain),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(45 * time.Second) // Wait for 30 seconds for status to become active
						return nil
					},
					resource.TestCheckResourceAttr(fullResourceName, "name", deploymentName),
					resource.TestCheckNoResourceAttr(fullResourceName, "description"),
					resource.TestCheckResourceAttr(fullResourceName, "allow_all_domains", "false"),
					resource.TestCheckResourceAttr(fullResourceName, "allowed_domains.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "allowed_domains.0", firstDomain),
					resource.TestMatchResourceAttr(fullResourceName, "status", regexp.MustCompile("^(Pending|Active)$")),
				),
			},
			{
				Config: deploymentResourceWithAllowedDomains(t, deploymentName, firstDomain, secondDomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "allowed_domains.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "allowed_domains.0", firstDomain),
					resource.TestCheckResourceAttr(fullResourceName, "allowed_domains.1", secondDomain),
				),
			},
			{
				ResourceName:            fullResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"status"},
			},
		},
		CheckDestroy: verifyDeploymentDestroyed,
	})
}

func TestAccResourceWebDeploymentsDeployment_Versioning(t *testing.T) {
	t.Parallel()
	var (
		deploymentName             = "Test Deployment " + util.RandString(8)
		fullDeploymentResourceName = "genesyscloud_webdeployments_deployment.versioning"
		fullConfigResourceName     = "genesyscloud_webdeployments_configuration.minimal"
	)

	cleanupWebDeploymentsDeployment(t, "Test Deployment ")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: versioningDeploymentResource(t, deploymentName, "description 1", "en-us", []string{"en-us"}),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper creation
						return nil
					},
					resource.TestCheckResourceAttr(fullDeploymentResourceName, "name", deploymentName),
					resource.TestCheckResourceAttr(fullDeploymentResourceName, "configuration.0.version", "1"),
					resource.TestCheckResourceAttrPair(fullDeploymentResourceName, "configuration.0.id", fullConfigResourceName, "id"),
					resource.TestCheckResourceAttrPair(fullDeploymentResourceName, "configuration.0.version", fullConfigResourceName, "version"),
				),
			},
			{
				Config: versioningDeploymentResource(t, deploymentName, "updated description", "en-us", []string{"en-us", "ja"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullDeploymentResourceName, "name", deploymentName),
					resource.TestCheckResourceAttr(fullDeploymentResourceName, "configuration.0.version", "2"),
					resource.TestCheckResourceAttrPair(fullDeploymentResourceName, "configuration.0.id", fullConfigResourceName, "id"),
					resource.TestCheckResourceAttrPair(fullDeploymentResourceName, "configuration.0.version", fullConfigResourceName, "version"),
				),
			},
			{
				Config: deploymentResourceWithoutConfigVersion(t, deploymentName, "updated description again", "en-us", []string{"en-us", "ja"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullDeploymentResourceName, "name", deploymentName),
					resource.TestCheckResourceAttr(fullDeploymentResourceName, "configuration.0.version", "3"),
					resource.TestCheckResourceAttrPair(fullDeploymentResourceName, "configuration.0.id", fullConfigResourceName, "id"),
					resource.TestCheckResourceAttrPair(fullDeploymentResourceName, "configuration.0.version", fullConfigResourceName, "version"),
				),
			},

			{
				ResourceName:            fullDeploymentResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"status"},
			},
		},
		CheckDestroy: verifyDeploymentDestroyed,
	})
}

func basicDeploymentResource(name, description string) string {
	minimalConfigName := "Minimal Config " + uuid.NewString()
	return fmt.Sprintf(`
	resource "genesyscloud_webdeployments_configuration" "minimal" {
		name             = "%s"
		languages        = ["en-us"]
		default_language = "en-us"
	}

	resource "genesyscloud_webdeployments_deployment" "basic" {
		name = "%s"
		description = "%s"
		allow_all_domains = true
		configuration {
			id = "${genesyscloud_webdeployments_configuration.minimal.id}"
			version = "${genesyscloud_webdeployments_configuration.minimal.version}"
		}
	}
	`, minimalConfigName, name, description)
}

func deploymentResourceWithAllowedDomains(t *testing.T, name string, allowedDomains ...string) string {
	value, err := json.Marshal(allowedDomains)
	if err != nil {
		t.Error(err)
	}
	minimalConfigName := "Minimal Config " + uuid.NewString()

	return fmt.Sprintf(`
	resource "genesyscloud_webdeployments_configuration" "minimal" {
		name             = "%s"
		languages        = ["en-us"]
		default_language = "en-us"
	}

	resource "genesyscloud_webdeployments_deployment" "basicWithAllowedDomains" {
		name = "%s"
		allowed_domains = %s
		configuration {
			id = "${genesyscloud_webdeployments_configuration.minimal.id}"
			version = "${genesyscloud_webdeployments_configuration.minimal.version}"
		}
	}
	`, minimalConfigName, name, value)
}

func versioningDeploymentResource(t *testing.T, name, description, defaultLanguage string, languages []string) string {
	value, err := json.Marshal(languages)
	if err != nil {
		t.Error(err)
	}
	minimalConfigName := "Minimal Config " + uuid.NewString()

	return fmt.Sprintf(`
	resource "genesyscloud_webdeployments_configuration" "minimal" {
		name = "%s"
		languages = %s
		default_language = "%s"
	}

	resource "genesyscloud_webdeployments_deployment" "versioning" {
		name = "%s"
		description = "%s"
		allow_all_domains = true
		configuration {
			id = "${genesyscloud_webdeployments_configuration.minimal.id}"
			version = genesyscloud_webdeployments_configuration.minimal.version
		}
	}
	`, minimalConfigName, value, defaultLanguage, name, description)
}

func deploymentResourceWithoutConfigVersion(t *testing.T, name, description, defaultLanguage string, languages []string) string {
	value, err := json.Marshal(languages)
	if err != nil {
		t.Error(err)
	}
	minimalConfigName := "Minimal Config " + uuid.NewString()

	return fmt.Sprintf(`
	resource "genesyscloud_webdeployments_configuration" "minimal" {
		name = "%s"
		languages = %s
		default_language = "%s"
	}

	resource "genesyscloud_webdeployments_deployment" "versioning" {
		name = "%s"
		description = "%s"
		allow_all_domains = true
		configuration {
			id = "${genesyscloud_webdeployments_configuration.minimal.id}"
		}
	}
	`, minimalConfigName, value, defaultLanguage, name, description)
}

func verifyDeploymentDestroyed(state *terraform.State) error {
	api := platformclientv2.NewWebDeploymentsApi()

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_webdeployments_deployment" {
			continue
		}

		_, response, err := api.GetWebdeploymentsDeployment(rs.Primary.ID, []string{})

		if util.IsStatus404(response) {
			continue
		}

		if err != nil {
			return fmt.Errorf("Unexpected error while checking that deployment has been destroyed: %s", err)
		}

		return fmt.Errorf("Deployment %s still exists when it was expected to have been destroyed", rs.Primary.ID)
	}

	return nil
}

func verifyLanguagesDestroyed(state *terraform.State) error {
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
			return fmt.Errorf("Language (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Language not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All languages destroyed
	return nil
}

func cleanupWebDeploymentsDeployment(t *testing.T, prefix string) {
	config, err := provider.AuthorizeSdk()
	if err != nil {
		t.Logf("Failed to authorize SDK: %s", err)
		return
	}
	deploymentsAPI := platformclientv2.NewWebDeploymentsApiWithConfig(config)

	webDeployments, resp, getErr := deploymentsAPI.GetWebdeploymentsDeployments([]string{})
	if getErr != nil {
		t.Logf("failed to get page of deployments: %v %v", getErr, resp)
		return
	}

	for _, webDeployment := range *webDeployments.Entities {
		if webDeployment.Name != nil && strings.HasPrefix(*webDeployment.Name, prefix) {
			_, err := deploymentsAPI.DeleteWebdeploymentsDeployment(*webDeployment.Id)
			if err != nil {
				t.Logf("failed to delete deployment: %v %v %v", *webDeployment.Id, getErr, resp)
			}
			time.Sleep(5 * time.Second)
		}
	}
}
