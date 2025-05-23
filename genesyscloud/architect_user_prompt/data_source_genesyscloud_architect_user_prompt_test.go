package architect_user_prompt

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceArchitectUserPrompt(t *testing.T) {
	userPromptResourceLabel := "test-user_prompt_1"
	userPromptName := "TestUserPrompt_1" + strings.Replace(uuid.NewString(), "-", "", -1)
	userPromptDescription := "Test description"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateUserPromptResource(&UserPromptStruct{
					userPromptResourceLabel,
					userPromptName,
					strconv.Quote(userPromptDescription),
					nil,
				}) + generateUserPromptDataSource(
					userPromptResourceLabel,
					"genesyscloud_architect_user_prompt."+userPromptResourceLabel+".name",
					"genesyscloud_architect_user_prompt."+userPromptResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_architect_user_prompt."+userPromptResourceLabel, "id", "genesyscloud_architect_user_prompt."+userPromptResourceLabel, "id"),
				),
			},
		},
	})
}

func generateUserPromptDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_architect_user_prompt" "%s" {
		name = %s
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
