package architect_grammar

import (
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
)

func TestAccDataSourceArchitectGrammar(t *testing.T) {
	var (
		grammarResource = "grammar-resource"
		grammarData     = "grammar-data"
		name            = "GrammarArchitect" + uuid.NewString()
		description     = "Sample description"
	)

	cleanupArchitectGrammar("GrammarArchitect")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
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

func cleanupArchitectGrammar(idPrefix string) {
	architectApi := platformclientv2.NewArchitectApi()

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		architectGrammars, _, getErr := architectApi.GetArchitectGrammars(pageNum, pageSize, "", "", nil, "", "", "", false)
		if getErr != nil {
			return
		}

		if architectGrammars.Entities == nil || len(*architectGrammars.Entities) == 0 {
			break
		}

		for _, grammar := range *architectGrammars.Entities {
			if grammar.Name != nil && strings.HasPrefix(*grammar.Name, idPrefix) {
				_, _, delErr := architectApi.DeleteArchitectGrammar(*grammar.Id)
				if delErr != nil {
					diag.Errorf("failed to delete architect grammar %s", delErr)
					return
				}
				log.Printf("Deleted architect grammar %s (%s)", *grammar.Id, *grammar.Name)
			}
		}
	}
}
