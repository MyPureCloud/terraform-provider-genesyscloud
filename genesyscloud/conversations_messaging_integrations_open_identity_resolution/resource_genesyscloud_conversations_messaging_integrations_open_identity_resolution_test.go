package conversations_messaging_integrations_open_identity_resolution

import (
	"fmt"
	"strings"
	"testing"
	"time"

	cmMessagingSetting "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	cmSupportedContent "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	open "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_open"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

const (
	homeDivisionDataSourceLabel = "home"
	homeDivisionIdConfigRef     = "data.genesyscloud_auth_division_home.home.id"
	homeDivisionStateRef        = "data.genesyscloud_auth_division_home.home"

	openIntegrationActivePollInterval = 10 * time.Second
	openIntegrationActivePollTimeout  = 3 * time.Minute
)

func TestAccResourceConversationsMessagingIntegrationsOpenIdentityResolution(t *testing.T) {
	var (
		identityResolutionResourceLabel = "test-identity-resolution"
		openIntegrationResourceLabel    = "test-open-integration"
		openIntegrationResourcePath     = "genesyscloud_conversations_messaging_integrations_open." + openIntegrationResourceLabel
		openIntegrationName             = "Terraform Test Open Integration-" + uuid.NewString()
		homeDivisionConfig              = gcloud.GenerateAuthDivisionHomeDataSource(homeDivisionDataSourceLabel)

		outboundNotificationWebhookUrl                  = "https://mock-server.prv-use1.test-pure.cloud/messaging-service/webhook"
		outboundNotificationWebhookSignatureSecretToken = uuid.NewString()
		nameSupportedContent                            = "TestTerraformSupportedContent-" + uuid.NewString()
		resourceLabelSupportedContent                   = "testSupportedContent"
		inboundType                                     = "*/*"
		nameMessagingSetting                            = "testSettings"
		resourceLabelMessagingSetting                   = "testConversationsMessagingSettings"
	)

	supportedContentResource1 := cmSupportedContent.GenerateSupportedContentResource(
		"genesyscloud_conversations_messaging_supportedcontent",
		resourceLabelSupportedContent,
		nameSupportedContent,
		cmSupportedContent.GenerateInboundTypeBlock(inboundType))

	messagingSettingResource1 := cmMessagingSetting.GenerateConversationsMessagingSettingsResource(
		resourceLabelMessagingSetting,
		nameMessagingSetting,
		cmMessagingSetting.GenerateContentStoryBlock(
			cmMessagingSetting.GenerateMentionInboundOnlySetting("Disabled"),
			cmMessagingSetting.GenerateReplyInboundOnlySetting("Enabled"),
		),
	)

	openIntegrationResource := open.GenerateConversationMessagingOpenResource(
		openIntegrationResourceLabel,
		openIntegrationName,
		"genesyscloud_conversations_messaging_supportedcontent."+resourceLabelSupportedContent+".id",
		"genesyscloud_conversations_messaging_settings."+resourceLabelMessagingSetting+".id",
		outboundNotificationWebhookUrl,
		outboundNotificationWebhookSignatureSecretToken,
		open.GenerateWebhookHeadersProperties("key", "value"),
	)

	openIntegrationDependenciesConfig := supportedContentResource1 + "\n" +
		messagingSettingResource1 +
		openIntegrationResource

	openIntegrationWithIRConfig := func(resolveIdentities, divisionId string) string {
		return homeDivisionConfig +
			openIntegrationDependenciesConfig + "\n" +
			generateConversationsMessagingIntegrationsOpenIdentityResolutionResource(
				identityResolutionResourceLabel,
				openIntegrationResourcePath+".id",
				resolveIdentities,
				divisionId,
			)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		CheckDestroy:      testVerifyConversationsMessagingIntegrationsOpenIdentityResolutionDestroyed,
		Steps: []resource.TestStep{
			{
				// Create Open integration first -> IR API requires an Active parent integration.
				Config: openIntegrationDependenciesConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(openIntegrationResourcePath, "name", openIntegrationName),
					waitForOpenIntegrationActive(openIntegrationResourcePath),
				),
			},
			{
				Config: openIntegrationWithIRConfig("true", homeDivisionIdConfigRef),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"genesyscloud_conversations_messaging_integrations_open_identity_resolution."+identityResolutionResourceLabel, "open_integration_id",
						openIntegrationResourcePath, "id"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_open_identity_resolution."+identityResolutionResourceLabel, "resolve_identities", "true"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_conversations_messaging_integrations_open_identity_resolution."+identityResolutionResourceLabel, "division_id",
						"data.genesyscloud_auth_division_home."+homeDivisionDataSourceLabel, "id"),
					verifyIdentityResolutionConfig(openIntegrationResourcePath, true, homeDivisionStateRef),
				),
			},
			{
				Config: openIntegrationDependenciesConfig + "\n" +
					generateConversationsMessagingIntegrationsOpenIdentityResolutionResource(
						identityResolutionResourceLabel,
						openIntegrationResourcePath+".id",
						"false",
						"",
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"genesyscloud_conversations_messaging_integrations_open_identity_resolution."+identityResolutionResourceLabel,
						"resolve_identities",
						"false",
					),
					verifyIdentityResolutionStateDivisionCleared("genesyscloud_conversations_messaging_integrations_open_identity_resolution."+identityResolutionResourceLabel),
					verifyIdentityResolutionConfig(openIntegrationResourcePath, false, ""),
				),
			},
			{
				Config: openIntegrationDependenciesConfig + "\n" +
					generateConversationsMessagingIntegrationsOpenIdentityResolutionResource(
						identityResolutionResourceLabel,
						openIntegrationResourcePath+".id",
						"false",
						`"*"`,
					),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:            "genesyscloud_conversations_messaging_integrations_open_identity_resolution." + identityResolutionResourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"division_id"},
			},
			{
				Config: openIntegrationDependenciesConfig,
				Check: verifyIdentityResolutionDefault(
					openIntegrationResourcePath,
				),
			},
		},
	})
}

func waitForOpenIntegrationActive(openResourcePath string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		openResource, ok := state.RootModule().Resources[openResourcePath]
		if !ok {
			return fmt.Errorf("failed to find resource %s in state", openResourcePath)
		}

		conversationsApi := platformclientv2.NewConversationsApiWithConfig(sdkConfig)
		openIntegrationId := openResource.Primary.ID
		deadline := time.Now().Add(openIntegrationActivePollTimeout)

		for time.Now().Before(deadline) {
			integration, _, err := conversationsApi.GetConversationsMessagingIntegrationsOpenIntegrationId(openIntegrationId, "")
			if err != nil {
				return fmt.Errorf("failed to get open integration %s: %w", openIntegrationId, err)
			}

			if integration.Status != nil && strings.EqualFold(*integration.Status, "Active") {
				return nil
			}

			status := "unknown"
			if integration.Status != nil {
				status = *integration.Status
			}
			fmt.Printf("waiting for open integration %s to become Active (current status=%s)\n", openIntegrationId, status)
			time.Sleep(openIntegrationActivePollInterval)
		}

		return fmt.Errorf("open integration %s did not become Active within %s", openIntegrationId, openIntegrationActivePollTimeout)
	}
}

func generateConversationsMessagingIntegrationsOpenIdentityResolutionResource(resourceLabel, openIntegrationId, resolveIdentities, divisionId string) string {
	divisionBlock := ""
	if divisionId != "" {
		divisionBlock = fmt.Sprintf("\n    division_id = %s", divisionId)
	}

	return fmt.Sprintf(`resource "genesyscloud_conversations_messaging_integrations_open_identity_resolution" "%s" {
      	open_integration_id = %s
  		resolve_identities = %s%s
	}`, resourceLabel, openIntegrationId, resolveIdentities, divisionBlock)
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

func verifyIdentityResolutionConfig(resourcePath string, resolveIdentities bool, divisionStateRef string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		openResource, ok := state.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("failed to find resource %s in state", resourcePath)
		}

		expectedDivisionId := ""
		if divisionStateRef != "" {
			divisionResource, ok := state.RootModule().Resources[divisionStateRef]
			if !ok {
				return fmt.Errorf("failed to find division %s in state", divisionStateRef)
			}
			expectedDivisionId = divisionResource.Primary.ID
		}

		ConversationsApi := platformclientv2.NewConversationsApiWithConfig(sdkConfig)
		config, _, err := ConversationsApi.GetConversationsMessagingIdentityresolutionIntegrationsOpenIntegrationId(openResource.Primary.ID)
		if err != nil {
			return err
		}

		if config.ResolveIdentities == nil {
			return fmt.Errorf("identity resolution config missing for open integration %s", openResource.Primary.ID)
		}

		if *config.ResolveIdentities != resolveIdentities {
			return fmt.Errorf("expected resolve_identities=%t for open integration %s, got %t", resolveIdentities, openResource.Primary.ID, *config.ResolveIdentities)
		}

		if expectedDivisionId != "" {
			if config.Division == nil || config.Division.Id == nil {
				return fmt.Errorf("expected division_id=%s for open integration %s, got none", expectedDivisionId, openResource.Primary.ID)
			}
			if *config.Division.Id != expectedDivisionId {
				return fmt.Errorf("expected division_id=%s for open integration %s, got %s", expectedDivisionId, openResource.Primary.ID, *config.Division.Id)
			}
		} else if config.Division != nil && config.Division.Id != nil {
			divisionId := *config.Division.Id
			if !isUnassignedDivisionId(divisionId) {
				return fmt.Errorf("expected the unassigned division (* or empty) for open integration %s, got division_id=%s", openResource.Primary.ID, divisionId)
			}
		}

		return nil
	}
}

func verifyIdentityResolutionDefault(resourcePath string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		openResource, ok := state.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("failed to find resource %s in state", resourcePath)
		}

		ConversationsApi := platformclientv2.NewConversationsApiWithConfig(sdkConfig)
		config, _, err := ConversationsApi.GetConversationsMessagingIdentityresolutionIntegrationsOpenIntegrationId(openResource.Primary.ID)
		if err != nil {
			return err
		}

		if !isDefaultIdentityResolutionConfig(config) {
			resolveIdentities := "nil"
			if config.ResolveIdentities != nil {
				resolveIdentities = fmt.Sprintf("%t", *config.ResolveIdentities)
			}
			return fmt.Errorf("expected default identity resolution config for open integration %s, got resolve_identities=%s", openResource.Primary.ID, resolveIdentities)
		}

		return nil
	}
}

// testVerifyConversationsMessagingIntegrationsOpenIdentityResolutionDestroyed verifies destroy reset IR config to the
// platform default. Parent 404 is accepted because the final test destroy may also remove the integration.
func testVerifyConversationsMessagingIntegrationsOpenIdentityResolutionDestroyed(state *terraform.State) error {
	ConversationsApi := platformclientv2.NewConversationsApiWithConfig(sdkConfig)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		openIntegrationId := rs.Primary.Attributes["open_integration_id"]
		if openIntegrationId == "" {
			openIntegrationId = rs.Primary.ID
		}

		config, resp, err := ConversationsApi.GetConversationsMessagingIdentityresolutionIntegrationsOpenIntegrationId(openIntegrationId)
		if util.IsStatus404(resp) {
			continue
		}
		if err != nil {
			return fmt.Errorf("unexpected error verifying identity resolution destroy for open integration %s: %s", openIntegrationId, err)
		}
		if !isDefaultIdentityResolutionConfig(config) {
			return fmt.Errorf("expected default identity resolution config after destroy for open integration %s", openIntegrationId)
		}
	}

	return nil
}
