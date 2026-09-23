package conversations_messaging_integrations_instagram_identity_resolution

import (
	"fmt"
	"testing"

	cmMessagingSetting "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	cmSupportedContent "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	instagram "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_instagram"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

const (
	homeDivisionDataSourceLabel = "home"
	homeDivisionIdConfigRef     = "data.genesyscloud_auth_division_home.home.id"
	homeDivisionStateRef        = "data.genesyscloud_auth_division_home.home"
)

func TestAccResourceConversationsMessagingIntegrationsInstagramIdentityResolution(t *testing.T) {
	var (
		identityResolutionResourceLabel   = "test-identity-resolution"
		instagramIntegrationResourceLabel = "test-instagram-integration"
		instagramIntegrationName          = "Terraform Test Instagram Integration-" + uuid.NewString()
		homeDivisionConfig                = gcloud.GenerateAuthDivisionHomeDataSource(homeDivisionDataSourceLabel)

		pageAccessToken1              = uuid.NewString()
		nameSupportedContent          = "TestTerraformSupportedContent-" + uuid.NewString()
		resourceLabelSupportedContent = "testSupportedContent"
		inboundType                   = "*/*"
		nameMessagingSetting          = "testSettings"
		resourceLabelMessagingSetting = "testConversationsMessagingSettings"
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

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		CheckDestroy:      testVerifyConversationsMessagingIntegrationsInstagramIdentityResolutionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: homeDivisionConfig +
					supportedContentResource1 + "\n" +
					messagingSettingResource1 +
					instagram.GenerateInstagramIntegrationResource(
						instagramIntegrationResourceLabel,
						instagramIntegrationName,
						"genesyscloud_conversations_messaging_supportedcontent."+resourceLabelSupportedContent+".id",
						"genesyscloud_conversations_messaging_settings."+resourceLabelMessagingSetting+".id",
						pageAccessToken1,
						"",
						"",
						"",
						"",
					) + "\n" +
					generateConversationsMessagingIntegrationsInstagramIdentityResolutionResource(
						identityResolutionResourceLabel,
						"genesyscloud_conversations_messaging_integrations_instagram."+instagramIntegrationResourceLabel+".id",
						"true",
						homeDivisionIdConfigRef,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"genesyscloud_conversations_messaging_integrations_instagram_identity_resolution."+identityResolutionResourceLabel, "instagram_integration_id",
						"genesyscloud_conversations_messaging_integrations_instagram."+instagramIntegrationResourceLabel, "id"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_messaging_integrations_instagram_identity_resolution."+identityResolutionResourceLabel, "resolve_identities", "true"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_conversations_messaging_integrations_instagram_identity_resolution."+identityResolutionResourceLabel, "division_id",
						"data.genesyscloud_auth_division_home."+homeDivisionDataSourceLabel, "id"),
					verifyIdentityResolutionConfig("genesyscloud_conversations_messaging_integrations_instagram."+instagramIntegrationResourceLabel, true, homeDivisionStateRef),
				),
			},
			{
				Config: supportedContentResource1 + "\n" +
					messagingSettingResource1 +
					instagram.GenerateInstagramIntegrationResource(
						instagramIntegrationResourceLabel,
						instagramIntegrationName,
						"genesyscloud_conversations_messaging_supportedcontent."+resourceLabelSupportedContent+".id",
						"genesyscloud_conversations_messaging_settings."+resourceLabelMessagingSetting+".id",
						pageAccessToken1,
						"",
						"",
						"",
						"",
					) + "\n" +
					generateConversationsMessagingIntegrationsInstagramIdentityResolutionResource(
						identityResolutionResourceLabel,
						"genesyscloud_conversations_messaging_integrations_instagram."+instagramIntegrationResourceLabel+".id",
						"false",
						"",
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"genesyscloud_conversations_messaging_integrations_instagram_identity_resolution."+identityResolutionResourceLabel,
						"resolve_identities",
						"false",
					),
					verifyIdentityResolutionStateDivisionCleared("genesyscloud_conversations_messaging_integrations_instagram_identity_resolution."+identityResolutionResourceLabel),
					verifyIdentityResolutionConfig("genesyscloud_conversations_messaging_integrations_instagram."+instagramIntegrationResourceLabel, false, ""),
				),
			},
			{
				// Explicit "*" is equivalent to omitted division_id (both mean the unassigned / STAR division).
				Config: supportedContentResource1 + "\n" +
					messagingSettingResource1 +
					instagram.GenerateInstagramIntegrationResource(
						instagramIntegrationResourceLabel,
						instagramIntegrationName,
						"genesyscloud_conversations_messaging_supportedcontent."+resourceLabelSupportedContent+".id",
						"genesyscloud_conversations_messaging_settings."+resourceLabelMessagingSetting+".id",
						pageAccessToken1,
						"",
						"",
						"",
						"",
					) + "\n" +
					generateConversationsMessagingIntegrationsInstagramIdentityResolutionResource(
						identityResolutionResourceLabel,
						"genesyscloud_conversations_messaging_integrations_instagram."+instagramIntegrationResourceLabel+".id",
						"false",
						`"*"`,
					),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      "genesyscloud_conversations_messaging_integrations_instagram_identity_resolution." + identityResolutionResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: supportedContentResource1 + "\n" +
					messagingSettingResource1 +
					instagram.GenerateInstagramIntegrationResource(
						instagramIntegrationResourceLabel,
						instagramIntegrationName,
						"genesyscloud_conversations_messaging_supportedcontent."+resourceLabelSupportedContent+".id",
						"genesyscloud_conversations_messaging_settings."+resourceLabelMessagingSetting+".id",
						pageAccessToken1,
						"",
						"",
						"",
						"",
					),
				Check: verifyIdentityResolutionDefault(
					"genesyscloud_conversations_messaging_integrations_instagram." + instagramIntegrationResourceLabel,
				),
			},
		},
	})
}

func generateConversationsMessagingIntegrationsInstagramIdentityResolutionResource(resourceLabel, instagramIntegrationId, resolveIdentities, divisionId string) string {
	divisionBlock := ""
	if divisionId != "" {
		divisionBlock = fmt.Sprintf("\n    division_id = %s", divisionId)
	}

	return fmt.Sprintf(`resource "genesyscloud_conversations_messaging_integrations_instagram_identity_resolution" "%s" {
  instagram_integration_id = %s
  resolve_identities = %s%s
}`, resourceLabel, instagramIntegrationId, resolveIdentities, divisionBlock)
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
		instagramResource, ok := state.RootModule().Resources[resourcePath]
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
		config, _, err := ConversationsApi.GetConversationsMessagingIdentityresolutionIntegrationsInstagramIntegrationId(instagramResource.Primary.ID)
		if err != nil {
			return err
		}

		if config.ResolveIdentities == nil {
			return fmt.Errorf("identity resolution config missing for instagram integration %s", instagramResource.Primary.ID)
		}

		if *config.ResolveIdentities != resolveIdentities {
			return fmt.Errorf("expected resolve_identities=%t for instagram integration %s, got %t", resolveIdentities, instagramResource.Primary.ID, *config.ResolveIdentities)
		}

		if expectedDivisionId != "" {
			if config.Division == nil || config.Division.Id == nil {
				return fmt.Errorf("expected division_id=%s for instagram integration %s, got none", expectedDivisionId, instagramResource.Primary.ID)
			}
			if *config.Division.Id != expectedDivisionId {
				return fmt.Errorf("expected division_id=%s for instagram integration %s, got %s", expectedDivisionId, instagramResource.Primary.ID, *config.Division.Id)
			}
		} else if config.Division != nil && config.Division.Id != nil {
			divisionId := *config.Division.Id
			if !isUnassignedDivisionId(divisionId) {
				return fmt.Errorf("expected the unassigned division (* or empty) for instagram integration %s, got division_id=%s", instagramResource.Primary.ID, divisionId)
			}
		}

		return nil
	}
}

func verifyIdentityResolutionDefault(resourcePath string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		instagramResource, ok := state.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("failed to find resource %s in state", resourcePath)
		}

		ConversationsApi := platformclientv2.NewConversationsApiWithConfig(sdkConfig)
		config, _, err := ConversationsApi.GetConversationsMessagingIdentityresolutionIntegrationsInstagramIntegrationId(instagramResource.Primary.ID)
		if err != nil {
			return err
		}

		if !isDefaultIdentityResolutionConfig(config) {
			resolveIdentities := "nil"
			if config.ResolveIdentities != nil {
				resolveIdentities = fmt.Sprintf("%t", *config.ResolveIdentities)
			}
			return fmt.Errorf("expected default identity resolution config for instagram integration %s, got resolve_identities=%s", instagramResource.Primary.ID, resolveIdentities)
		}

		return nil
	}
}

// testVerifyConversationsMessagingIntegrationsInstagramIdentityResolutionDestroyed verifies destroy reset IR config to the
// platform default. Parent 404 is accepted because the final test destroy may also remove the integration.
func testVerifyConversationsMessagingIntegrationsInstagramIdentityResolutionDestroyed(state *terraform.State) error {
	ConversationsApi := platformclientv2.NewConversationsApiWithConfig(sdkConfig)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		instagramIntegrationId := rs.Primary.Attributes["instagram_integration_id"]
		if instagramIntegrationId == "" {
			instagramIntegrationId = rs.Primary.ID
		}

		config, resp, err := ConversationsApi.GetConversationsMessagingIdentityresolutionIntegrationsInstagramIntegrationId(instagramIntegrationId)
		if util.IsStatus404(resp) {
			continue
		}
		if err != nil {
			return fmt.Errorf("unexpected error verifying identity resolution destroy for instagram integration %s: %s", instagramIntegrationId, err)
		}
		if !isDefaultIdentityResolutionConfig(config) {
			return fmt.Errorf("expected default identity resolution config after destroy for instagram integration %s", instagramIntegrationId)
		}
	}

	return nil
}
