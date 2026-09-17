package routing_email_route_identity_resolution

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	routingEmailDomain "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	routingEmailRoute "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_route"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

const (
	homeDivisionDataSourceLabel = "home"
	homeDivisionIdConfigRef     = "data.genesyscloud_auth_division_home.home.id"
	homeDivisionStateRef        = "data.genesyscloud_auth_division_home.home"
)

func TestAccResourceRoutingEmailRouteIdentityResolution(t *testing.T) {
	var (
		identityResolutionResourceLabel = "test-identity-resolution"
		domainResourceLabel             = "test-domain"
		routeResourceLabel              = "test-route"
		domainId                        = fmt.Sprintf("terraformroutes.%s.com", strings.Replace(uuid.NewString(), "-", "", -1))
		routePattern                    = "tf" + strings.Replace(uuid.NewString(), "-", "", -1)[:8]
		fromName                        = "TF IR Test"
		homeDivisionConfig              = gcloud.GenerateAuthDivisionHomeDataSource(homeDivisionDataSourceLabel)
	)

	domainConfig := routingEmailDomain.GenerateRoutingEmailDomainResource(
		domainResourceLabel,
		domainId,
		util.FalseValue,
		util.NullValue,
	)
	routeConfig := routingEmailRoute.GenerateRoutingEmailRouteResource(
		routeResourceLabel,
		"genesyscloud_routing_email_domain."+domainResourceLabel+".id",
		routePattern,
		fromName,
		fmt.Sprintf(`from_email = "tf@%s"`, domainId),
	)
	domainRef := "genesyscloud_routing_email_domain." + domainResourceLabel + ".id"
	routeRef := "genesyscloud_routing_email_route." + routeResourceLabel + ".id"
	parentConfig := domainConfig + routeConfig

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		CheckDestroy:      testVerifyRoutingEmailRouteIdentityResolutionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: homeDivisionConfig + parentConfig + generateRoutingEmailRouteIdentityResolutionResource(
					identityResolutionResourceLabel,
					domainRef,
					routeRef,
					"true",
					homeDivisionIdConfigRef,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_email_route_identity_resolution."+identityResolutionResourceLabel, "domain_name",
						"genesyscloud_routing_email_domain."+domainResourceLabel, "id",
					),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_email_route_identity_resolution."+identityResolutionResourceLabel, "route_id",
						"genesyscloud_routing_email_route."+routeResourceLabel, "id",
					),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route_identity_resolution."+identityResolutionResourceLabel, "resolve_identities", "true"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_email_route_identity_resolution."+identityResolutionResourceLabel, "division_id",
						"data.genesyscloud_auth_division_home."+homeDivisionDataSourceLabel, "id",
					),
					verifyIdentityResolutionConfig("genesyscloud_routing_email_route."+routeResourceLabel, true, homeDivisionStateRef),
				),
			},
			{
				Config: parentConfig + generateRoutingEmailRouteIdentityResolutionResource(
					identityResolutionResourceLabel,
					domainRef,
					routeRef,
					"false",
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route_identity_resolution."+identityResolutionResourceLabel, "resolve_identities", "false"),
					verifyIdentityResolutionStateDivisionCleared("genesyscloud_routing_email_route_identity_resolution."+identityResolutionResourceLabel),
					verifyIdentityResolutionConfig("genesyscloud_routing_email_route."+routeResourceLabel, false, ""),
				),
			},
			{
				// Explicit "*" is equivalent to omitted division_id (both mean the unassigned / STAR division).
				Config: parentConfig + generateRoutingEmailRouteIdentityResolutionResource(
					identityResolutionResourceLabel,
					domainRef,
					routeRef,
					"false",
					`"*"`,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      "genesyscloud_routing_email_route_identity_resolution." + identityResolutionResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: parentConfig,
				Check: resource.ComposeTestCheckFunc(
					verifyIdentityResolutionDefault("genesyscloud_routing_email_route." + routeResourceLabel),
				),
			},
		},
	})
}

func generateRoutingEmailRouteIdentityResolutionResource(resourceLabel, domainName, routeId, resolveIdentities, divisionId string) string {
	divisionBlock := ""
	if divisionId != "" {
		divisionBlock = fmt.Sprintf("\n    division_id = %s", divisionId)
	}

	return fmt.Sprintf(`resource "genesyscloud_routing_email_route_identity_resolution" "%s" {
		domain_name = %s
		route_id = %s
		resolve_identities = %s%s
	}`, resourceLabel, domainName, routeId, resolveIdentities, divisionBlock)
}

func verifyIdentityResolutionStateDivisionCleared(resourcePath string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("failed to find resource %s in state", resourcePath)
		}

		if divisionId, ok := resourceState.Primary.Attributes["division_id"]; ok && divisionId != "" {
			return fmt.Errorf("expected division_id to be cleared from state for %s, still have %q", resourcePath, divisionId)
		}

		return nil
	}
}

func verifyIdentityResolutionConfig(routeResourcePath string, resolveIdentities bool, divisionStateRef string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		routeResource, ok := state.RootModule().Resources[routeResourcePath]
		if !ok {
			return fmt.Errorf("failed to find route %s in state", routeResourcePath)
		}

		domainName := routeResource.Primary.Attributes["domain_id"]
		routeId := routeResource.Primary.ID
		if domainName == "" || routeId == "" {
			return fmt.Errorf("route %s missing domain_id or id in state", routeResourcePath)
		}

		expectedDivisionId := ""
		if divisionStateRef != "" {
			divisionResource, ok := state.RootModule().Resources[divisionStateRef]
			if !ok {
				return fmt.Errorf("failed to find division %s in state", divisionStateRef)
			}
			expectedDivisionId = divisionResource.Primary.ID
		}

		routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
		config, _, err := routingApi.GetRoutingEmailDomainRouteIdentityresolution(domainName, routeId)
		if err != nil {
			return err
		}

		if config.ResolveIdentities == nil {
			return fmt.Errorf("identity resolution config missing for route %s/%s", domainName, routeId)
		}

		if *config.ResolveIdentities != resolveIdentities {
			return fmt.Errorf("expected resolve_identities=%t for route %s/%s, got %t", resolveIdentities, domainName, routeId, *config.ResolveIdentities)
		}

		if expectedDivisionId != "" {
			if config.Division == nil || config.Division.Id == nil {
				return fmt.Errorf("expected division_id=%s for route %s/%s, got none", expectedDivisionId, domainName, routeId)
			}
			if *config.Division.Id != expectedDivisionId {
				return fmt.Errorf("expected division_id=%s for route %s/%s, got %s", expectedDivisionId, domainName, routeId, *config.Division.Id)
			}
		} else if config.Division != nil && config.Division.Id != nil {
			divisionId := *config.Division.Id
			if !isUnassignedDivisionId(divisionId) {
				return fmt.Errorf("expected the unassigned division (* or empty) for route %s/%s, got division_id=%s", domainName, routeId, divisionId)
			}
		}

		return nil
	}
}

func verifyIdentityResolutionDefault(routeResourcePath string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		routeResource, ok := state.RootModule().Resources[routeResourcePath]
		if !ok {
			return fmt.Errorf("failed to find route %s in state", routeResourcePath)
		}

		domainName := routeResource.Primary.Attributes["domain_id"]
		routeId := routeResource.Primary.ID
		if domainName == "" || routeId == "" {
			return fmt.Errorf("route %s missing domain_id or id in state", routeResourcePath)
		}

		routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
		config, _, err := routingApi.GetRoutingEmailDomainRouteIdentityresolution(domainName, routeId)
		if err != nil {
			return err
		}

		if !isDefaultIdentityResolutionConfig(config) {
			resolveIdentities := "nil"
			if config.ResolveIdentities != nil {
				resolveIdentities = fmt.Sprintf("%t", *config.ResolveIdentities)
			}
			return fmt.Errorf("expected default identity resolution config for route %s/%s, got resolve_identities=%s", domainName, routeId, resolveIdentities)
		}

		return nil
	}
}

// testVerifyRoutingEmailRouteIdentityResolutionDestroyed verifies destroy reset IR config to the
// platform default. Parent 404 is accepted because the final test destroy may also remove the email route.
func testVerifyRoutingEmailRouteIdentityResolutionDestroyed(state *terraform.State) error {
	routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		domainName := rs.Primary.Attributes["domain_name"]
		routeId := rs.Primary.Attributes["route_id"]

		if domainName == "" || routeId == "" {
			parts := strings.Split(rs.Primary.ID, "/")
			if len(parts) == 2 {
				if domainName == "" {
					domainName = parts[0]
				}
				if routeId == "" {
					routeId = parts[1]
				}
			}
		}

		if domainName == "" || routeId == "" {
			return fmt.Errorf("unable to resolve domain_name/route_id for destroyed resource id %s", rs.Primary.ID)
		}

		config, resp, err := routingApi.GetRoutingEmailDomainRouteIdentityresolution(domainName, routeId)
		if util.IsStatus404(resp) {
			continue
		}
		if err != nil {
			return fmt.Errorf("unexpected error verifying identity resolution destroy for route %s/%s: %s", domainName, routeId, err)
		}
		if !isDefaultIdentityResolutionConfig(config) {
			return fmt.Errorf("expected default identity resolution config after destroy for route %s/%s", domainName, routeId)
		}
	}

	return nil
}
