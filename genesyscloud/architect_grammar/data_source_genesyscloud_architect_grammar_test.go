package architect_grammar

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	gcloud "terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
)

func TestAccDataSourceArchitectGrammar(t *testing.T) {
	var (
		grammarResource = "grammar-resource"
		grammarData     = "grammar-data"
		name            = "Grammar" + uuid.NewString()
		description     = "Sample description"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateGrammarResource(
					grammarResource,
					name,
					description,
				) + generateGrammarDataSource(
					grammarData,
					name,
					"genesyscloud_architect_grammar."+grammarResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_architect_grammar."+grammarData, "id", "genesyscloud_architect_grammar."+grammarResource, "id"),
				),
			},
		},
	})
}

func generateGrammarDataSource(
	resourceID string,
	name string,
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_architect_grammar" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
