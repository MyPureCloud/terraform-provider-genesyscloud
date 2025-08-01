package architect_grammar

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func TestAccResourceArchitectGrammar(t *testing.T) {
	var (
		resourceLabel = "grammar" + uuid.NewString()
		name1         = "Test grammar " + uuid.NewString()
		description1  = "Test description"
		name2         = "Test grammar " + uuid.NewString()
		description2  = "A new description"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create Grammar
				Config: GenerateGrammarResource(
					resourceLabel,
					name1,
					description1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceLabel, "description", description1),
				),
			},
			{
				// Update Grammar
				Config: GenerateGrammarResource(
					resourceLabel,
					name2,
					description2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceLabel, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_architect_grammar."+resourceLabel, "description", description2),
				),
			},
			{
				// Read
				ResourceName:      "genesyscloud_architect_grammar." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyGrammarDestroyed,
	})
}

func testVerifyGrammarDestroyed(state *terraform.State) error {
	architectAPI := platformclientv2.NewArchitectApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_architect_grammar" {
			continue
		}
		grammar, resp, err := architectAPI.GetArchitectGrammar(rs.Primary.ID, false)
		if grammar != nil {
			return fmt.Errorf("Grammar (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Grammar not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All grammars deleted
	return nil
}
