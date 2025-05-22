package organization_presence_definition

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

/*
The resource_genesyscloud_organization_presence_definition_test.go contains all of the test cases for running the resource
tests for organization_presence_definition.
*/

func TestAccResourceOrganizationPresenceDefinition(t *testing.T) {
	t.Parallel()
	var (
		codeResourceLabel1 = "organization-presence-definition1"
		languageLabelEnus  = "From Keyboard " + uuid.NewString()
		languageLabelEs    = "del teclado " + uuid.NewString()
		languageLabels1    = map[string]string{"en_US": strconv.Quote(languageLabelEnus)}
		languageLabelsStr1 = util.GenerateMapAttrWithMapProperties("language_labels", languageLabels1)
		languageLabels2    = map[string]string{"en_US": strconv.Quote(languageLabelEnus), "es": strconv.Quote(languageLabelEs)}
		languageLabelsStr2 = util.GenerateMapAttrWithMapProperties("language_labels", languageLabels2)
		systemPresence     = "Away"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create single language label
				Config: GenerateOrganizationPresenceDefinitionResource(
					codeResourceLabel1,
					languageLabelsStr1,
					systemPresence,
					util.NullValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+codeResourceLabel1, "language_labels.%", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+codeResourceLabel1, "language_labels.en_US", languageLabelEnus),
					resource.TestCheckResourceAttr(ResourceType+"."+codeResourceLabel1, "system_presence", systemPresence),
				),
			},
			{
				// Update additional language label
				Config: GenerateOrganizationPresenceDefinitionResource(
					codeResourceLabel1,
					languageLabelsStr2,
					systemPresence,
					util.NullValue,
				),
				Check: resource.ComposeTestCheckFunc(

					resource.TestCheckResourceAttr(ResourceType+"."+codeResourceLabel1, "language_labels.%", "2"),
					resource.TestCheckResourceAttr(ResourceType+"."+codeResourceLabel1, "language_labels.en_US", languageLabelEnus),
					resource.TestCheckResourceAttr(ResourceType+"."+codeResourceLabel1, "language_labels.es", languageLabelEs),
					resource.TestCheckResourceAttr(ResourceType+"."+codeResourceLabel1, "system_presence", systemPresence),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + codeResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyOrganizationPresenceDefinitionDestroyed,
	})
}

func GenerateOrganizationPresenceDefinitionResource(resourceLabel string, languageLabelsStr string, systemPresence string, deactivated interface{}) string {
	var deactivatedStr string
	if deactivated != util.NullValue {
		deactivatedStr = fmt.Sprintf(`deactivated = %v`, deactivated)
	}

	return fmt.Sprintf(`resource "%s" "%s" {
		%s
		system_presence = "%s"
		%s
	}
	`, ResourceType, resourceLabel, languageLabelsStr, systemPresence, deactivatedStr)
}

func testVerifyOrganizationPresenceDefinitionDestroyed(state *terraform.State) error {
	presenceAPI := platformclientv2.NewPresenceApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		organizationPresenceDefinition, _, err := presenceAPI.GetPresenceDefinition(rs.Primary.ID, "")
		if *organizationPresenceDefinition.Deactivated == false {
			return fmt.Errorf("Organization presence definition (%s) still exists", rs.Primary.ID)
		} else if *organizationPresenceDefinition.Deactivated {
			// Organization presence definition not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All Organization presence definitions destroyed
	return nil
}
