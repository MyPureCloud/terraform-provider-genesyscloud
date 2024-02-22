package responsemanagement_response

import (
	"fmt"
	"strconv"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	respmanagementLibrary "terraform-provider-genesyscloud/genesyscloud/responsemanagement_library"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
)

func TestAccResourceResponseManagementResponseFooterField(t *testing.T) {
	t.Parallel()
	var (
		// Responses initial values
		responseResource          = "response-resource"
		name1                     = "Response-" + uuid.NewString()
		textsContent1             = "Random text block content string"
		textsContentTypes         = []string{"text/plain", "text/html"}
		interactionTypes          = []string{"chat", "email", "twitter"}
		substitutionsId           = "sub123"
		substitutionsDescription  = "Substitutions description"
		substitutionsDefaultValue = "Substitutions default value"
		substitutionsSchema       = "schema document"
		responseTypes             = []string{`MessagingTemplate`, `CampaignSmsTemplate`, `CampaignEmailTemplate`, `Footer`}
		footerType                = "Signature"
		footerResource            = []string{strconv.Quote("Campaign")}
		// Responses Updated values
		name2         = "Response-" + uuid.NewString()
		textsContent2 = "Random text block content string new"

		// Library resources variables
		libraryResource1 = "library-resource1"
		libraryName1     = "Reference library1"
		libraryResource2 = "library-resource2"
		libraryName2     = "Reference library2"

		// Asset resources variables
		testFilesDir  = "test_responseasset_data"
		assetResource = "asset-resource"
		fileName      = "yeti-img.png"
		fullPath      = fmt.Sprintf("%s/%s", testFilesDir, fileName)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with required values
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResource1,
					libraryName1,
				) + generateResponseManagementResponseResource(
					responseResource,
					name1,
					[]string{"genesyscloud_responsemanagement_library." + libraryResource1 + ".id"},
					gcloud.NullValue,
					gcloud.NullValue,
					gcloud.NullValue,
					[]string{},
					generateTextsBlock(
						textsContent1,
						textsContentTypes[0],
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "name", name1),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResource, "library_ids.0",
						"genesyscloud_responsemanagement_library."+libraryResource1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "texts.0.content", textsContent1),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "texts.0.content_type", textsContentTypes[0]),
				),
			},
			{
				// Update with new name and texts and add remaining values
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResource1,
					libraryName1,
				) + gcloud.GenerateResponseManagementResponseAssetResource(
					assetResource,
					fullPath,
					gcloud.NullValue,
				) + generateResponseManagementResponseResource(
					responseResource,
					name2,
					[]string{"genesyscloud_responsemanagement_library." + libraryResource1 + ".id"},
					strconv.Quote(interactionTypes[0]),
					gcloud.GenerateJsonSchemaDocStr(substitutionsSchema),
					strconv.Quote(responseTypes[3]),
					[]string{"genesyscloud_responsemanagement_responseasset." + assetResource + ".id"},
					generateFooterBlock(footerType, footerResource),
					generateTextsBlock(
						textsContent2,
						textsContentTypes[1],
					),
					generateSubstitutionsBlock(
						substitutionsId,
						substitutionsDescription,
						substitutionsDefaultValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "name", name2),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResource, "library_ids.0",
						"genesyscloud_responsemanagement_library."+libraryResource1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "texts.0.content", textsContent2),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "texts.0.content_type", textsContentTypes[1]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "interaction_type", interactionTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions.0.id", substitutionsId),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions.0.description", substitutionsDescription),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions.0.default_value", substitutionsDefaultValue),
					gcloud.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions_schema_id", "type", "object"),
					gcloud.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions_schema_id", "properties."+substitutionsSchema+".type", "string"),
					gcloud.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions_schema_id", "required", substitutionsSchema),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "response_type", responseTypes[3]),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResource, "asset_ids.0",
						"genesyscloud_responsemanagement_responseasset."+assetResource, "id"),
				),
			},
			{
				// Add more texts and change libraries
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResource1,
					libraryName1,
				) + respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResource2,
					libraryName2,
				) + gcloud.GenerateResponseManagementResponseAssetResource(
					assetResource,
					fullPath,
					gcloud.NullValue,
				) + generateResponseManagementResponseResource(
					responseResource,
					name2,
					[]string{"genesyscloud_responsemanagement_library." + libraryResource2 + ".id", "genesyscloud_responsemanagement_library." + libraryResource1 + ".id"},
					strconv.Quote(interactionTypes[0]),
					gcloud.GenerateJsonSchemaDocStr(substitutionsSchema),
					strconv.Quote(responseTypes[3]),
					[]string{"genesyscloud_responsemanagement_responseasset." + assetResource + ".id"},
					generateTextsBlock(
						textsContent1,
						textsContentTypes[0],
					),
					generateFooterBlock(footerType, footerResource),
					generateTextsBlock(
						textsContent2,
						textsContentTypes[1],
					),
					generateSubstitutionsBlock(
						substitutionsId,
						substitutionsDescription,
						substitutionsDefaultValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "name", name2),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResource, "library_ids.0",
						"genesyscloud_responsemanagement_library."+libraryResource2, "id"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResource, "library_ids.1",
						"genesyscloud_responsemanagement_library."+libraryResource1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "texts.0.content", textsContent2),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "texts.0.content_type", textsContentTypes[1]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "texts.1.content", textsContent1),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "texts.1.content_type", textsContentTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "interaction_type", interactionTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions.0.id", substitutionsId),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions.0.description", substitutionsDescription),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions.0.default_value", substitutionsDefaultValue),
					gcloud.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions_schema_id", "type", "object"),
					gcloud.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions_schema_id", "properties."+substitutionsSchema+".type", "string"),
					gcloud.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions_schema_id", "required", substitutionsSchema),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "response_type", responseTypes[3]),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResource, "asset_ids.0",
						"genesyscloud_responsemanagement_responseasset."+assetResource, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_responsemanagement_response." + responseResource,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"substitutions_schema_id", "messaging_template", "response_type"},
			},
		},
		CheckDestroy: testVerifyResponseManagementResponseDestroyed,
	})
}

func TestAccResourceResponseManagementResponseMessaging(t *testing.T) {
	t.Parallel()
	var (
		// Responses initial values
		responseResource          = "response-resource"
		name1                     = "Response-" + uuid.NewString()
		textsContent1             = "Random text block content string"
		textsContentTypes         = []string{"text/plain", "text/html"}
		interactionTypes          = []string{"chat", "email", "twitter"}
		substitutionsId           = "sub123"
		substitutionsDescription  = "Substitutions description"
		substitutionsDefaultValue = "Substitutions default value"
		substitutionsSchema       = "schema document"
		responseTypes             = []string{`MessagingTemplate`, `CampaignSmsTemplate`, `CampaignEmailTemplate`}
		templateName              = "Sample template name"
		templateNamespace         = "Template namespace"

		// Responses Updated values
		name2         = "Response-" + uuid.NewString()
		textsContent2 = "Random text block content string new"

		// Library resources variables
		libraryResource1 = "library-resource1"
		libraryName1     = "Reference library1"
		libraryResource2 = "library-resource2"
		libraryName2     = "Reference library2"

		// Asset resources variables
		testFilesDir  = "test_responseasset_data"
		assetResource = "asset-resource"
		fileName      = "yeti-img.png"
		fullPath      = fmt.Sprintf("%s/%s", testFilesDir, fileName)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with required values
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResource1,
					libraryName1,
				) + generateResponseManagementResponseResource(
					responseResource,
					name1,
					[]string{"genesyscloud_responsemanagement_library." + libraryResource1 + ".id"},
					gcloud.NullValue,
					gcloud.NullValue,
					gcloud.NullValue,
					[]string{},
					generateTextsBlock(
						textsContent1,
						textsContentTypes[0],
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "name", name1),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResource, "library_ids.0",
						"genesyscloud_responsemanagement_library."+libraryResource1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "texts.0.content", textsContent1),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "texts.0.content_type", textsContentTypes[0]),
				),
			},
			{
				// Update with new name and texts and add remaining values
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResource1,
					libraryName1,
				) + gcloud.GenerateResponseManagementResponseAssetResource(
					assetResource,
					fullPath,
					gcloud.NullValue,
				) + generateResponseManagementResponseResource(
					responseResource,
					name2,
					[]string{"genesyscloud_responsemanagement_library." + libraryResource1 + ".id"},
					strconv.Quote(interactionTypes[0]),
					gcloud.GenerateJsonSchemaDocStr(substitutionsSchema),
					strconv.Quote(responseTypes[0]),
					[]string{"genesyscloud_responsemanagement_responseasset." + assetResource + ".id"},
					generateTextsBlock(
						textsContent2,
						textsContentTypes[1],
					),
					generateSubstitutionsBlock(
						substitutionsId,
						substitutionsDescription,
						substitutionsDefaultValue,
					),
					generateMessagingTemplateBlock(
						generateWhatsappBlock(
							templateName,
							templateNamespace,
							"en_US",
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "name", name2),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResource, "library_ids.0",
						"genesyscloud_responsemanagement_library."+libraryResource1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "texts.0.content", textsContent2),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "texts.0.content_type", textsContentTypes[1]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "interaction_type", interactionTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions.0.id", substitutionsId),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions.0.description", substitutionsDescription),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions.0.default_value", substitutionsDefaultValue),
					gcloud.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions_schema_id", "type", "object"),
					gcloud.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions_schema_id", "properties."+substitutionsSchema+".type", "string"),
					gcloud.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions_schema_id", "required", substitutionsSchema),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "response_type", responseTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "messaging_template.0.whats_app.0.name", templateName),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "messaging_template.0.whats_app.0.namespace", templateNamespace),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "messaging_template.0.whats_app.0.language", "en_US"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResource, "asset_ids.0",
						"genesyscloud_responsemanagement_responseasset."+assetResource, "id"),
				),
			},
			{
				// Add more texts and change libraries
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResource1,
					libraryName1,
				) + respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResource2,
					libraryName2,
				) + gcloud.GenerateResponseManagementResponseAssetResource(
					assetResource,
					fullPath,
					gcloud.NullValue,
				) + generateResponseManagementResponseResource(
					responseResource,
					name2,
					[]string{"genesyscloud_responsemanagement_library." + libraryResource2 + ".id", "genesyscloud_responsemanagement_library." + libraryResource1 + ".id"},
					strconv.Quote(interactionTypes[0]),
					gcloud.GenerateJsonSchemaDocStr(substitutionsSchema),
					strconv.Quote(responseTypes[0]),
					[]string{"genesyscloud_responsemanagement_responseasset." + assetResource + ".id"},
					generateTextsBlock(
						textsContent1,
						textsContentTypes[0],
					),
					generateTextsBlock(
						textsContent2,
						textsContentTypes[1],
					),
					generateSubstitutionsBlock(
						substitutionsId,
						substitutionsDescription,
						substitutionsDefaultValue,
					),
					generateMessagingTemplateBlock(
						generateWhatsappBlock(
							templateName,
							templateNamespace,
							"en_US",
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "name", name2),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResource, "library_ids.0",
						"genesyscloud_responsemanagement_library."+libraryResource2, "id"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResource, "library_ids.1",
						"genesyscloud_responsemanagement_library."+libraryResource1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "texts.0.content", textsContent2),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "texts.0.content_type", textsContentTypes[1]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "texts.1.content", textsContent1),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "texts.1.content_type", textsContentTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "interaction_type", interactionTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions.0.id", substitutionsId),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions.0.description", substitutionsDescription),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions.0.default_value", substitutionsDefaultValue),
					gcloud.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions_schema_id", "type", "object"),
					gcloud.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions_schema_id", "properties."+substitutionsSchema+".type", "string"),
					gcloud.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResource, "substitutions_schema_id", "required", substitutionsSchema),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "response_type", responseTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "messaging_template.0.whats_app.0.name", templateName),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "messaging_template.0.whats_app.0.namespace", templateNamespace),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResource, "messaging_template.0.whats_app.0.language", "en_US"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResource, "asset_ids.0",
						"genesyscloud_responsemanagement_responseasset."+assetResource, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_responsemanagement_response." + responseResource,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"substitutions_schema_id", "messaging_template", "response_type"},
			},
		},
		CheckDestroy: testVerifyResponseManagementResponseDestroyed,
	})
}

func generateResponseManagementResponseResource(
	resourceId string,
	name string,
	libraryIds []string,
	interactionType string,
	schema string,
	responseType string,
	assetIds []string,
	nestedBlocks ...string,
) string {
	return fmt.Sprintf(`
		resource "genesyscloud_responsemanagement_response" "%s" {
			name = "%s"
			library_ids = [%s]
			interaction_type = %s
			substitutions_schema_id = %s
			response_type = %s
			asset_ids = [%s]
			%s
		}
	`, resourceId, name, strings.Join(libraryIds, ", "), interactionType, schema, responseType, strings.Join(assetIds, ", "), strings.Join(nestedBlocks, "\n"))
}

func generateTextsBlock(
	content string,
	contentType string,
) string {
	return fmt.Sprintf(`
		texts {
			content = "%s"
			content_type = "%s"
		}
	`, content, contentType)
}

func generateSubstitutionsBlock(id, description, defaultValue string) string {
	return fmt.Sprintf(`
		substitutions {
			id            = "%s"
			description   = "%s"
			default_value = "%s"
		}
	`, id, description, defaultValue)
}

func generateMessagingTemplateBlock(
	attrs ...string,
) string {
	return fmt.Sprintf(`
		messaging_template {
			%s
		}
	`, strings.Join(attrs, "\n"))
}

func generateWhatsappBlock(
	name string,
	nameSpace string,
	language string,
) string {
	return fmt.Sprintf(`
		whats_app{
			name = "%s"
			namespace = "%s"
			language = "%s"
		}
	`, name, nameSpace, language)
}

func generateFooterBlock(
	footerType string,
	footerResource []string,
) string {
	return fmt.Sprintf(`
		footer {
			type = "%s"
			applicable_resources=[%s]
		}
	`, footerType, strings.Join(footerResource, ", "))
}

func testVerifyResponseManagementResponseDestroyed(state *terraform.State) error {
	managementAPI := platformclientv2.NewResponseManagementApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_responsemanagement_response" {
			continue
		}
		responses, resp, err := managementAPI.GetResponsemanagementResponse(rs.Primary.ID, "")
		if responses != nil {
			return fmt.Errorf("response (%s) still exists", rs.Primary.ID)
		} else if gcloud.IsStatus404(resp) {
			// response not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All responses destroyed
	return nil
}
