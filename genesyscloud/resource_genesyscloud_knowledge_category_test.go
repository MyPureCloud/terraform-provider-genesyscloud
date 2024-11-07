package genesyscloud

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func TestAccResourceKnowledgeCategoryBasic(t *testing.T) {
	var (
		knowledgeBaseResource1     = "test-knowledgebase1"
		knowledgeBaseName1         = "Terraform Knowledge Base" + uuid.NewString()
		knowledgeBaseDescription1  = "test-knowledgebase-description1"
		knowledgeBaseCoreLanguage1 = "en-US"
		knowledgeCategoryResource1 = "test-knowledge-category1"
		categoryName               = "Terraform Knowledge Category" + uuid.NewString()
		categoryDescription        = "test-description1"
		categoryDescription2       = "test-description2"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateKnowledgeKnowledgebaseResource(
					knowledgeBaseResource1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					knowledgeBaseCoreLanguage1,
				) +
					generateKnowledgeCategoryResource(
						knowledgeCategoryResource1,
						knowledgeBaseResource1,
						categoryName,
						categoryDescription,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_knowledge_category."+knowledgeCategoryResource1, "knowledge_base_id", "genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_category."+knowledgeCategoryResource1, "knowledge_category.0.name", categoryName),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_category."+knowledgeCategoryResource1, "knowledge_category.0.description", categoryDescription),
				),
			},
			{
				// Update
				Config: GenerateKnowledgeKnowledgebaseResource(
					knowledgeBaseResource1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					knowledgeBaseCoreLanguage1,
				) +
					generateKnowledgeCategoryResource(
						knowledgeCategoryResource1,
						knowledgeBaseResource1,
						categoryName,
						categoryDescription2,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_knowledge_category."+knowledgeCategoryResource1, "knowledge_base_id", "genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_category."+knowledgeCategoryResource1, "knowledge_category.0.name", categoryName),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_category."+knowledgeCategoryResource1, "knowledge_category.0.description", categoryDescription2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_knowledge_category." + knowledgeCategoryResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyKnowledgeCategoryDestroyed,
	})
}

func TestAccResourceKnowledgeCategoryParentChild(t *testing.T) {
	var (
		knowledgeBaseResource1     = "test-knowledgebase1"
		knowledgeBaseName1         = "Terraform Knowledge Base" + uuid.NewString()
		knowledgeBaseDescription1  = "test-knowledgebase-description1"
		knowledgeBaseCoreLanguage1 = "en-US"

		knowledgeCategoryResourceParent1 = "test-knowledge-category-parent1"
		categoryParentName               = "Terraform Knowledge Category parent" + uuid.NewString()
		categoryParentDescription        = "test-parent-description1"
		categoryParentDescription2       = "test-parent-description2"

		knowledgeCategoryResourceChild1 = "test-knowledge-category-child1"
		categoryChildName               = "Terraform Knowledge Category child" + uuid.NewString()
		categoryChildDescription        = "test-child-description1"
		categoryChildDescription2       = "test-child-description2"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateKnowledgeKnowledgebaseResource(
					knowledgeBaseResource1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					knowledgeBaseCoreLanguage1,
				) + generateKnowledgeCategoryResource(
					knowledgeCategoryResourceParent1,
					knowledgeBaseResource1,
					categoryParentName,
					categoryParentDescription,
				) + generateKnowledgeCategoryChildResource(
					knowledgeCategoryResourceChild1,
					knowledgeBaseResource1,
					categoryChildName,
					categoryChildDescription,
					knowledgeCategoryResourceParent1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_knowledge_category."+knowledgeCategoryResourceParent1, "knowledge_base_id", "genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_category."+knowledgeCategoryResourceParent1, "knowledge_category.0.name", categoryParentName),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_category."+knowledgeCategoryResourceParent1, "knowledge_category.0.description", categoryParentDescription),
				),
			},
			{
				// Update
				Config: GenerateKnowledgeKnowledgebaseResource(
					knowledgeBaseResource1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					knowledgeBaseCoreLanguage1,
				) + generateKnowledgeCategoryResource(
					knowledgeCategoryResourceParent1,
					knowledgeBaseResource1,
					categoryParentName,
					categoryParentDescription2,
				) + generateKnowledgeCategoryChildResource(
					knowledgeCategoryResourceChild1,
					knowledgeBaseResource1,
					categoryChildName,
					categoryChildDescription2,
					knowledgeCategoryResourceParent1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_knowledge_category."+knowledgeCategoryResourceParent1, "knowledge_base_id", "genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_category."+knowledgeCategoryResourceParent1, "knowledge_category.0.name", categoryParentName),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_category."+knowledgeCategoryResourceParent1, "knowledge_category.0.description", categoryParentDescription2),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_category."+knowledgeCategoryResourceChild1, "knowledge_category.0.name", categoryChildName),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_category."+knowledgeCategoryResourceChild1, "knowledge_category.0.description", categoryChildDescription2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_knowledge_category." + knowledgeCategoryResourceParent1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyKnowledgeCategoryDestroyed,
	})
}

func generateKnowledgeCategoryResource(resourceName string, knowledgeBaseResource string, categoryName string, categoryDescription string) string {
	category := fmt.Sprintf(`
        resource "genesyscloud_knowledge_category" "%s" {
            knowledge_base_id = genesyscloud_knowledge_knowledgebase.%s.id
            %s
        }
        `, resourceName,
		knowledgeBaseResource,
		generateKnowledgeCategoryRequestBody(categoryName, categoryDescription, util.NullValue),
	)
	return category
}

func generateKnowledgeCategoryChildResource(resourceName string, knowledgeBaseResource string, categoryName string, categoryDescription string, parentCategoryName string) string {
	category := fmt.Sprintf(`
	resource "genesyscloud_knowledge_category" "%s" {
		knowledge_base_id = genesyscloud_knowledge_knowledgebase.%s.id
		%s
	}
	`, resourceName,
		knowledgeBaseResource,
		generateKnowledgeCategoryRequestBody(categoryName, categoryDescription,
			"genesyscloud_knowledge_category."+parentCategoryName+".id"),
	)
	return category
}

func generateKnowledgeCategoryRequestBody(categoryName string, categoryDescription string, parentCategoryName string) string {

	return fmt.Sprintf(`
        knowledge_category {
            name = "%s"
            description = "%s"
			parent_id = %s
        }
        `, categoryName,
		categoryDescription,
		parentCategoryName,
	)
}

func testVerifyKnowledgeCategoryDestroyed(state *terraform.State) error {
	knowledgeAPI := platformclientv2.NewKnowledgeApi()
	var knowledgeBaseId string
	for _, rs := range state.RootModule().Resources {
		if rs.Type == "genesyscloud_knowledge_knowledgebase" {
			knowledgeBaseId = rs.Primary.ID
			break
		}
	}
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_knowledge_category" {
			continue
		}
		id := strings.Split(rs.Primary.ID, " ")
		knowledgeCategoryId := id[0]
		knowledgeCategory, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseCategory(knowledgeBaseId, knowledgeCategoryId)
		if knowledgeCategory != nil {
			return fmt.Errorf("Knowledge category (%s) still exists", knowledgeCategoryId)
		} else if util.IsStatus404(resp) || util.IsStatus400(resp) {
			// Knowledge base category not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All knowledge categories destroyed
	return nil
}
