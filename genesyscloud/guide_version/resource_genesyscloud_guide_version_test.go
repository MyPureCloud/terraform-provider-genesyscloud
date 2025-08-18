package guide_version

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/guide"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccResourceGuideVersion(t *testing.T) {
	if os.Getenv("GENESYSCLOUD_REGION") != "tca" {
		t.Skip("Skipping test because GENESYSCLOUD_REGION is not set to tca")
	}

	if !guide.GuideFtIsEnabled() {
		t.Skip("Skipping test as guide feature toggle is not enabled")
		return
	}

	t.Parallel()
	var (
		guideVersionResourceLabel    = "genesyscloud_guide_version"
		guideResourceLabel           = "genesyscloud_guide"
		guideName                    = "Test Guide " + uuid.NewString()
		guideSource                  = "Manual"
		instruction                  = "This is a test instruction for the guide version."
		updatedInstruction           = "This is an updated test instruction for the guide version."
		guideVersionResourceFullPath = ResourceType + "." + guideVersionResourceLabel
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: guide.GenerateGuideResource(guideResourceLabel, guideName, guideSource) +
					GenerateGuideVersionResource(
						guideVersionResourceLabel,
						"${genesyscloud_guide."+guideResourceLabel+".id}",
						instruction,
					),
			},
			{
				// Create guide version with multiple data actions and variables
				Config: guide.GenerateGuideResource(guideResourceLabel, guideName, guideSource) +
					GenerateGuideVersionResource(
						guideVersionResourceLabel,
						"${genesyscloud_guide."+guideResourceLabel+".id}",
						instruction,
						GenerateVariableBlock("testVar1", "String", "Input", "Test variable 1 description"),
						GenerateVariableBlock("testVar2", "Integer", "Output", "Test variable 2 description"),
						GenerateResourcesBlock(
							GenerateDataActionBlock("test_data_action_id_1", "Test Data Action 1", "Test data action 1 description"),
							GenerateDataActionBlock("test_data_action_id_2", "Test Data Action 2", "Test data action 2 description"),
							GenerateDataActionBlock("test_data_action_id_3", "Test Data Action 3", "Test data action 3 description"),
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "instruction", instruction),
					resource.TestCheckResourceAttr("genesyscloud_guide."+guideResourceLabel, "name", guideName),
					resource.TestCheckResourceAttr("genesyscloud_guide."+guideResourceLabel, "source", guideSource),

					// Check variable attributes
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "variables.0.name", "testVar1"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "variables.0.type", "String"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "variables.0.scope", "Input"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "variables.0.description", "Test variable 1 description"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "variables.1.name", "testVar2"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "variables.1.type", "Integer"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "variables.1.scope", "Output"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "variables.1.description", "Test variable 2 description"),

					// Check multiple data action attributes
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.0.data_action_id", "test_data_action_id_1"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.0.label", "Test Data Action 1"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.0.description", "Test data action 1 description"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.1.data_action_id", "test_data_action_id_2"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.1.label", "Test Data Action 2"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.1.description", "Test data action 2 description"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.2.data_action_id", "test_data_action_id_3"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.2.label", "Test Data Action 3"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.2.description", "Test data action 3 description"),
				),
			},
			{
				// Update guide version with different number of data actions
				Config: guide.GenerateGuideResource(guideResourceLabel, guideName, guideSource) +
					GenerateGuideVersionResource(
						guideVersionResourceLabel,
						"${genesyscloud_guide."+guideResourceLabel+".id}",
						updatedInstruction,
						GenerateVariableBlock("testVar1", "String", "Input", "Test variable 1 description"),
						GenerateVariableBlock("testVar2", "Integer", "Output", "Test variable 2 description"),
						GenerateResourcesBlock(
							GenerateDataActionBlock("updated_data_action_id_1", "Updated Data Action 1", "Updated data action 1 description"),
							GenerateDataActionBlock("updated_data_action_id_2", "Updated Data Action 2", "Updated data action 2 description"),
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "instruction", updatedInstruction),

					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.0.data_action_id", "updated_data_action_id_1"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.0.label", "Updated Data Action 1"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.0.description", "Updated data action 1 description"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.1.data_action_id", "updated_data_action_id_2"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.1.label", "Updated Data Action 2"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.1.description", "Updated data action 2 description"),
				),
			},
			{
				// Update guide version with different number of data actions
				Config: guide.GenerateGuideResource(guideResourceLabel, guideName, guideSource) +
					GenerateGuideVersionResource(
						guideVersionResourceLabel,
						"${genesyscloud_guide."+guideResourceLabel+".id}",
						updatedInstruction,
						GenerateVariableBlock("testVar1", "String", "Input", "Test variable 1 description"),
						GenerateVariableBlock("testVar2", "Integer", "Output", "Test variable 2 description"),
						GenerateResourcesBlock(
							GenerateDataActionBlock("updated_data_action_id_1", "Updated Data Action 1", "Updated data action 1 description"),
							GenerateDataActionBlock("updated_data_action_id_2", "Updated Data Action 2", "Updated data action 2 description"),
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "instruction", updatedInstruction),

					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.0.data_action_id", "updated_data_action_id_1"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.0.label", "Updated Data Action 1"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.0.description", "Updated data action 1 description"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.1.data_action_id", "updated_data_action_id_2"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.1.label", "Updated Data Action 2"),
					resource.TestCheckResourceAttr(guideVersionResourceFullPath, "resources.0.data_action.1.description", "Updated data action 2 description"),
				),
			},
			{
				// Import/Read
				ResourceName:            guideVersionResourceFullPath,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"generate_content"},
			},
		},
	})
}

func GenerateGuideVersionResource(resourceLabel string, guideId string, instruction string, nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		guide_id = "%s"
		instruction = "%s"
		%s
	}
	`, ResourceType, resourceLabel, guideId, instruction, strings.Join(nestedBlocks, "\n"))
}

func GenerateVariableBlock(name string, varType string, scope string, description string) string {
	return fmt.Sprintf(`variables {
		name = "%s"
		type = "%s"
		scope = "%s"
		description = "%s"
	}`, name, varType, scope, description)
}

func GenerateResourcesBlock(dataActionBlocks ...string) string {
	return fmt.Sprintf(`resources {
		%s
	}`, strings.Join(dataActionBlocks, "\n"))
}

func GenerateDataActionBlock(dataActionId string, label string, description string) string {
	return fmt.Sprintf(`data_action {
		data_action_id = "%s"
		label = "%s"
		description = "%s"
	}`, dataActionId, label, description)
}
