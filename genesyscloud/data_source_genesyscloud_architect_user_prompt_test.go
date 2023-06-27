package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUserPrompt(t *testing.T) {
	userPromptResource := "test-user_prompt_1"
	userPromptName := "TestUserPrompt_1" + strings.Replace(uuid.NewString(), "-", "", -1)
	userPromptDescription := "Test description"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateUserPromptResource(&UserPromptStruct{
					userPromptResource,
					userPromptName,
					strconv.Quote(userPromptDescription),
					nil,
				}) + generateUserPromptDataSource(
					userPromptResource,
					"genesyscloud_architect_user_prompt."+userPromptResource+".name",
					"genesyscloud_architect_user_prompt."+userPromptResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_architect_user_prompt."+userPromptResource, "id", "genesyscloud_architect_user_prompt."+userPromptResource, "id"),
				),
			},
		},
	})
}

func generateUserPromptDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_architect_user_prompt" "%s" {
		name = %s
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
