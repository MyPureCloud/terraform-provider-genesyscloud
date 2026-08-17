package routing_queue_identity_resolution

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

const (
	homeDivisionDataSourceLabel = "home"
	homeDivisionIdConfigRef     = "data.genesyscloud_auth_division_home.home.id"
	homeDivisionStateRef        = "data.genesyscloud_auth_division_home.home"
)

func TestAccResourceRoutingQueueIdentityResolution(t *testing.T) {
	var (
		identityResolutionResourceLabel = "test-identity-resolution"
		queueResourceLabel              = "test-queue"
		queueName                       = "Terraform Test Queue-" + uuid.NewString()
		homeDivisionConfig              = gcloud.GenerateAuthDivisionHomeDataSource(homeDivisionDataSourceLabel)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		CheckDestroy:      testVerifyRoutingQueueIdentityResolutionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: homeDivisionConfig + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel,
					queueName,
				) + generateRoutingQueueIdentityResolutionResource(
					identityResolutionResourceLabel,
					"genesyscloud_routing_queue."+queueResourceLabel+".id",
					"true",
					homeDivisionIdConfigRef,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_identity_resolution."+identityResolutionResourceLabel, "queue_id",
						"genesyscloud_routing_queue."+queueResourceLabel, "id",
					),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_identity_resolution."+identityResolutionResourceLabel, "call_on_behalf_of_queue.0.resolve_identities", "true"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_identity_resolution."+identityResolutionResourceLabel, "call_on_behalf_of_queue.0.division_id",
						"data.genesyscloud_auth_division_home."+homeDivisionDataSourceLabel, "id",
					),
					verifyIdentityResolutionConfig("genesyscloud_routing_queue."+queueResourceLabel, true, homeDivisionStateRef),
				),
			},
			{
				Config: routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel,
					queueName,
				) + generateRoutingQueueIdentityResolutionResource(
					identityResolutionResourceLabel,
					"genesyscloud_routing_queue."+queueResourceLabel+".id",
					"false",
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_identity_resolution."+identityResolutionResourceLabel, "call_on_behalf_of_queue.0.resolve_identities", "false"),
					verifyIdentityResolutionStateDivisionCleared("genesyscloud_routing_queue_identity_resolution."+identityResolutionResourceLabel),
					verifyIdentityResolutionConfig("genesyscloud_routing_queue."+queueResourceLabel, false, ""),
				),
			},
			{
				// Explicit "*" is equivalent to omitted division_id (both mean the unassigned / STAR division).
				Config: routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel,
					queueName,
				) + generateRoutingQueueIdentityResolutionResource(
					identityResolutionResourceLabel,
					"genesyscloud_routing_queue."+queueResourceLabel+".id",
					"false",
					`"*"`,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      "genesyscloud_routing_queue_identity_resolution." + identityResolutionResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel,
					queueName,
				),
				Check: resource.ComposeTestCheckFunc(
					verifyIdentityResolutionDefault("genesyscloud_routing_queue." + queueResourceLabel),
				),
			},
		},
	})
}

func generateRoutingQueueIdentityResolutionResource(resourceLabel, queueId, resolveIdentities, divisionId string) string {
	divisionBlock := ""
	if divisionId != "" {
		divisionBlock = fmt.Sprintf("\n    division_id = %s", divisionId)
	}

	return fmt.Sprintf(`resource "genesyscloud_routing_queue_identity_resolution" "%s" {
  queue_id = %s
  call_on_behalf_of_queue {
    resolve_identities = %s%s
  }
}`, resourceLabel, queueId, resolveIdentities, divisionBlock)
}

func verifyIdentityResolutionStateDivisionCleared(resourcePath string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("failed to find resource %s in state", resourcePath)
		}

		if divisionId, ok := resourceState.Primary.Attributes["call_on_behalf_of_queue.0.division_id"]; ok && divisionId != "" {
			return fmt.Errorf("expected division_id to be cleared from state for %s, still have %q", resourcePath, divisionId)
		}

		return nil
	}
}

func verifyIdentityResolutionConfig(queueResourcePath string, resolveIdentities bool, divisionStateRef string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		queueResource, ok := state.RootModule().Resources[queueResourcePath]
		if !ok {
			return fmt.Errorf("failed to find queue %s in state", queueResourcePath)
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
		config, _, err := routingApi.GetRoutingQueueIdentityresolution(queueResource.Primary.ID)
		if err != nil {
			return err
		}

		if config.CallOnBehalfOfQueue == nil || config.CallOnBehalfOfQueue.ResolveIdentities == nil {
			return fmt.Errorf("identity resolution config missing for queue %s", queueResource.Primary.ID)
		}

		if *config.CallOnBehalfOfQueue.ResolveIdentities != resolveIdentities {
			return fmt.Errorf("expected resolve_identities=%t for queue %s, got %t", resolveIdentities, queueResource.Primary.ID, *config.CallOnBehalfOfQueue.ResolveIdentities)
		}

		if expectedDivisionId != "" {
			if config.CallOnBehalfOfQueue.Division == nil || config.CallOnBehalfOfQueue.Division.Id == nil {
				return fmt.Errorf("expected division_id=%s for queue %s, got none", expectedDivisionId, queueResource.Primary.ID)
			}
			if *config.CallOnBehalfOfQueue.Division.Id != expectedDivisionId {
				return fmt.Errorf("expected division_id=%s for queue %s, got %s", expectedDivisionId, queueResource.Primary.ID, *config.CallOnBehalfOfQueue.Division.Id)
			}
		} else if config.CallOnBehalfOfQueue.Division != nil && config.CallOnBehalfOfQueue.Division.Id != nil {
			divisionId := *config.CallOnBehalfOfQueue.Division.Id
			if !isUnassignedDivisionId(divisionId) {
				return fmt.Errorf("expected the unassigned division (* or empty) for queue %s, got division_id=%s", queueResource.Primary.ID, divisionId)
			}
		}

		return nil
	}
}

func verifyIdentityResolutionDefault(queueResourcePath string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		queueResource, ok := state.RootModule().Resources[queueResourcePath]
		if !ok {
			return fmt.Errorf("failed to find queue %s in state", queueResourcePath)
		}

		routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
		config, _, err := routingApi.GetRoutingQueueIdentityresolution(queueResource.Primary.ID)
		if err != nil {
			return err
		}

		if !isDefaultIdentityResolutionConfig(config) {
			resolveIdentities := "nil"
			if config.CallOnBehalfOfQueue != nil && config.CallOnBehalfOfQueue.ResolveIdentities != nil {
				resolveIdentities = fmt.Sprintf("%t", *config.CallOnBehalfOfQueue.ResolveIdentities)
			}
			return fmt.Errorf("expected default identity resolution config for queue %s, got resolve_identities=%s", queueResource.Primary.ID, resolveIdentities)
		}

		return nil
	}
}

// testVerifyRoutingQueueIdentityResolutionDestroyed verifies destroy reset IR config to the
// platform default. Parent 404 is accepted because the final test destroy may also remove the queue.
func testVerifyRoutingQueueIdentityResolutionDestroyed(state *terraform.State) error {
	routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		queueId := rs.Primary.Attributes["queue_id"]
		if queueId == "" {
			queueId = rs.Primary.ID
		}

		config, resp, err := routingApi.GetRoutingQueueIdentityresolution(queueId)
		if util.IsStatus404(resp) {
			continue
		}
		if err != nil {
			return fmt.Errorf("unexpected error verifying identity resolution destroy for queue %s: %s", queueId, err)
		}
		if !isDefaultIdentityResolutionConfig(config) {
			return fmt.Errorf("expected default identity resolution config after destroy for queue %s", queueId)
		}
	}

	return nil
}
