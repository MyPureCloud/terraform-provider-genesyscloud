package webdeployments_deployment_identity_resolution

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

const (
	homeDivisionDataSourceLabel = "home"
	homeDivisionIdConfigRef     = "data.genesyscloud_auth_division_home.home.id"
	homeDivisionStateRef        = "data.genesyscloud_auth_division_home.home"

	deploymentResourceLabel = "test-deployment"
	deploymentResourcePath  = "genesyscloud_webdeployments_deployment." + deploymentResourceLabel
)

func TestAccResourceWebDeploymentIdentityResolution(t *testing.T) {
	var (
		identityResolutionResourceLabel = "test-identity-resolution"
		identityResolutionResourcePath  = ResourceType + "." + identityResolutionResourceLabel

		// Built once so the base resources stay byte-identical across steps and are
		// not recreated between them.
		baseConfig = webDeploymentBaseConfig(
			"Terraform Test Deployment-"+uuid.NewString(),
			"Terraform Test Config-"+uuid.NewString(),
		)
		homeDivisionConfig = gcloud.GenerateAuthDivisionHomeDataSource(homeDivisionDataSourceLabel)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		CheckDestroy:      testVerifyWebDeploymentIdentityResolutionDestroyed,
		Steps: []resource.TestStep{
			{
				// Create with a real division and web tracking automerge enabled.
				//
				// authenticated_web_messaging stays false throughout: contacts-service
				// AND-s that flag with the deployment's capability, and the capability is
				// false whenever the web deployment has no auth configured. This fixture
				// uses a minimal configuration with no authentication_settings, so asking
				// for true would be silently downgraded to false by the platform. Unit
				// tests cover the flag itself.
				Config: baseConfig + homeDivisionConfig + generateWebDeploymentIdentityResolutionResource(
					identityResolutionResourceLabel,
					deploymentResourcePath+".id",
					"true",
					homeDivisionIdConfigRef,
					"",
					"false",
					"true",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						identityResolutionResourcePath, "deployment_id",
						deploymentResourcePath, "id",
					),
					resource.TestCheckResourceAttr(identityResolutionResourcePath, "resolve_identities", "true"),
					resource.TestCheckResourceAttrPair(
						identityResolutionResourcePath, "division_id",
						homeDivisionStateRef, "id",
					),
					resource.TestCheckResourceAttr(identityResolutionResourcePath, "automerge_config.0.authenticated_web_messaging", "false"),
					resource.TestCheckResourceAttr(identityResolutionResourcePath, "automerge_config.0.web_tracking", "true"),
					verifyIdentityResolutionConfig(true, homeDivisionStateRef, false, true),
				),
			},
			{
				// Clear the division, flip resolve_identities off, and turn automerge off
				// with an explicit all-false block. The block must stay in state even
				// though it now matches the platform default, or the plan cannot converge.
				Config: baseConfig + generateWebDeploymentIdentityResolutionResource(
					identityResolutionResourceLabel,
					deploymentResourcePath+".id",
					"false",
					"",
					"",
					"false",
					"false",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(identityResolutionResourcePath, "resolve_identities", "false"),
					verifyStateAttributeCleared(identityResolutionResourcePath, "division_id"),
					resource.TestCheckResourceAttr(identityResolutionResourcePath, "automerge_config.#", "1"),
					resource.TestCheckResourceAttr(identityResolutionResourcePath, "automerge_config.0.authenticated_web_messaging", "false"),
					resource.TestCheckResourceAttr(identityResolutionResourcePath, "automerge_config.0.web_tracking", "false"),
					verifyIdentityResolutionConfig(false, "", false, false),
				),
			},
			{
				// Explicit "*" is equivalent to an omitted division_id (both mean the
				// unassigned / STAR division), so the plan must stay empty.
				Config: baseConfig + generateWebDeploymentIdentityResolutionResource(
					identityResolutionResourceLabel,
					deploymentResourcePath+".id",
					"false",
					`"*"`,
					"",
					"false",
					"false",
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				// Removing the automerge block leaves no block behind in state -- the case
				// that Computed would have made impossible.
				//
				// The division is set again here so that the final step destroys a resource
				// that still has one, which is what exercises destroy actually clearing it.
				// Without this the division is already unassigned by the time the resource
				// goes away, and the destroy checks only ever see default-from-default.
				Config: baseConfig + homeDivisionConfig + generateWebDeploymentIdentityResolutionResource(
					identityResolutionResourceLabel,
					deploymentResourcePath+".id",
					"false",
					homeDivisionIdConfigRef,
					"",
					"",
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(identityResolutionResourcePath, "automerge_config.#", "0"),
					resource.TestCheckResourceAttrPair(
						identityResolutionResourcePath, "division_id",
						homeDivisionStateRef, "id",
					),
					verifyIdentityResolutionConfig(false, homeDivisionStateRef, false, false),
				),
			},
			{
				ResourceName:      identityResolutionResourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Dropping the identity resolution resource restores the platform default
				// on a deployment that still exists.
				Config: baseConfig,
				Check: resource.ComposeTestCheckFunc(
					verifyIdentityResolutionDefault(),
				),
			},
		},
	})
}

func webDeploymentBaseConfig(deploymentName, configName string) string {
	return fmt.Sprintf(`
resource "genesyscloud_webdeployments_configuration" "minimal" {
  name             = "%s"
  languages        = ["en-us"]
  default_language = "en-us"
}

resource "genesyscloud_webdeployments_deployment" "%s" {
  name              = "%s"
  allow_all_domains = true
  configuration {
    id      = genesyscloud_webdeployments_configuration.minimal.id
    version = genesyscloud_webdeployments_configuration.minimal.version
  }
}
`, configName, deploymentResourceLabel, deploymentName)
}

// generateWebDeploymentIdentityResolutionResource omits division_id / external_source_id
// when empty, and omits the automerge_config block entirely when either flag is empty.
func generateWebDeploymentIdentityResolutionResource(resourceLabel, deploymentId, resolveIdentities, divisionId, externalSourceId, authenticatedWebMessaging, webTracking string) string {
	optionalAttrs := ""
	if divisionId != "" {
		optionalAttrs += fmt.Sprintf("\n  division_id = %s", divisionId)
	}
	if externalSourceId != "" {
		optionalAttrs += fmt.Sprintf("\n  external_source_id = %s", externalSourceId)
	}
	if authenticatedWebMessaging != "" && webTracking != "" {
		optionalAttrs += fmt.Sprintf(`
  automerge_config {
    authenticated_web_messaging = %s
    web_tracking                = %s
  }`, authenticatedWebMessaging, webTracking)
	}

	return fmt.Sprintf(`resource "%s" "%s" {
  deployment_id      = %s
  resolve_identities = %s%s
}`, ResourceType, resourceLabel, deploymentId, resolveIdentities, optionalAttrs)
}

func verifyStateAttributeCleared(resourcePath, attribute string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("failed to find resource %s in state", resourcePath)
		}

		if value, ok := resourceState.Primary.Attributes[attribute]; ok && value != "" {
			return fmt.Errorf("expected %s to be cleared from state for %s, still have %q", attribute, resourcePath, value)
		}

		return nil
	}
}

func verifyIdentityResolutionConfig(resolveIdentities bool, divisionStateRef string, authenticatedWebMessaging, webTracking bool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		config, deploymentId, err := getIdentityResolutionFromApi(state)
		if err != nil {
			return err
		}

		if config.ResolveIdentities == nil {
			return fmt.Errorf("identity resolution config missing resolve_identities for deployment %s", deploymentId)
		}
		if *config.ResolveIdentities != resolveIdentities {
			return fmt.Errorf("expected resolve_identities=%t for deployment %s, got %t", resolveIdentities, deploymentId, *config.ResolveIdentities)
		}

		expectedDivisionId := ""
		if divisionStateRef != "" {
			divisionResource, ok := state.RootModule().Resources[divisionStateRef]
			if !ok {
				return fmt.Errorf("failed to find division %s in state", divisionStateRef)
			}
			expectedDivisionId = divisionResource.Primary.ID
		}

		if expectedDivisionId != "" {
			if config.Division == nil || config.Division.Id == nil {
				return fmt.Errorf("expected division_id=%s for deployment %s, got none", expectedDivisionId, deploymentId)
			}
			if *config.Division.Id != expectedDivisionId {
				return fmt.Errorf("expected division_id=%s for deployment %s, got %s", expectedDivisionId, deploymentId, *config.Division.Id)
			}
		} else if config.Division != nil && config.Division.Id != nil && !isUnassignedDivisionId(*config.Division.Id) {
			return fmt.Errorf("expected the unassigned division (* or empty) for deployment %s, got division_id=%s", deploymentId, *config.Division.Id)
		}

		actualAuthenticatedWebMessaging := false
		actualWebTracking := false
		if config.Automerge != nil {
			if config.Automerge.AuthenticatedWebMessaging != nil {
				actualAuthenticatedWebMessaging = *config.Automerge.AuthenticatedWebMessaging
			}
			if config.Automerge.WebTracking != nil {
				actualWebTracking = *config.Automerge.WebTracking
			}
		}
		if actualAuthenticatedWebMessaging != authenticatedWebMessaging || actualWebTracking != webTracking {
			// The raw payload is included because a missing automerge object and an
			// all-false one are indistinguishable once flattened.
			return fmt.Errorf("expected automerge {authenticated_web_messaging=%t, web_tracking=%t} for deployment %s, got {%t, %t}; raw payload: %s",
				authenticatedWebMessaging, webTracking, deploymentId, actualAuthenticatedWebMessaging, actualWebTracking, config.String())
		}

		return nil
	}
}

func verifyIdentityResolutionDefault() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		config, deploymentId, err := getIdentityResolutionFromApi(state)
		if err != nil {
			return err
		}

		if !isDefaultIdentityResolutionConfig(config) {
			return fmt.Errorf("expected default identity resolution config for deployment %s", deploymentId)
		}

		return nil
	}
}

func getIdentityResolutionFromApi(state *terraform.State) (*platformclientv2.Deploymentidentityresolutionconfig, string, error) {
	deploymentResource, ok := state.RootModule().Resources[deploymentResourcePath]
	if !ok {
		return nil, "", fmt.Errorf("failed to find deployment %s in state", deploymentResourcePath)
	}

	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)
	config, _, err := api.GetWebdeploymentsDeploymentIdentityresolution(deploymentResource.Primary.ID)
	if err != nil {
		return nil, deploymentResource.Primary.ID, err
	}

	return config, deploymentResource.Primary.ID, nil
}

// testVerifyWebDeploymentIdentityResolutionDestroyed verifies destroy reset the identity
// resolution config to the platform default. A 404 is accepted because the final destroy
// also removes the parent deployment.
func testVerifyWebDeploymentIdentityResolutionDestroyed(state *terraform.State) error {
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		deploymentId := rs.Primary.Attributes["deployment_id"]
		if deploymentId == "" {
			deploymentId = rs.Primary.ID
		}

		config, resp, err := api.GetWebdeploymentsDeploymentIdentityresolution(deploymentId)
		if util.IsStatus404(resp) {
			continue
		}
		if err != nil {
			return fmt.Errorf("unexpected error verifying identity resolution destroy for deployment %s: %s", deploymentId, err)
		}
		if !isDefaultIdentityResolutionConfig(config) {
			return fmt.Errorf("expected default identity resolution config after destroy for deployment %s", deploymentId)
		}
	}

	return nil
}
