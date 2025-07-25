package responsemanagement_response

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	respmanagementLibrary "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/responsemanagement_library"
	respManagementRespAsset "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/responsemanagement_responseasset"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func TestAccResourceResponseManagementResponseFooterField(t *testing.T) {
	var (
		// Responses initial values
		responseResourceLabel     = "response-resource"
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
		libraryResourceLabel1 = "library-resource1"
		libraryName1          = "Referencelibrary1"
		libraryResourceLabel2 = "library-resource2"
		libraryName2          = "Referencelibrary2"

		// Asset resources variables
		testFilesDir       = "test_responseasset_data"
		assetResourceLabel = "asset-resource-response"
		fileName           = "yeti-img.png"
		fullPath           = filepath.Join(testFilesDir, fileName)
	)

	err := cleanupResponseAssets("yeti")
	if err != nil {
		t.Errorf("failed to cleanup response assets: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with required values
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResourceLabel1,
					libraryName1,
				) + GenerateResponseManagementResponseResource(
					responseResourceLabel,
					name1,
					[]string{"genesyscloud_responsemanagement_library." + libraryResourceLabel1 + ".id"},
					util.NullValue,
					util.NullValue,
					util.NullValue,
					[]string{},
					GenerateTextsBlock(
						textsContent1,
						textsContentTypes[0],
						util.NullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "name", name1),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResourceLabel, "library_ids.0",
						"genesyscloud_responsemanagement_library."+libraryResourceLabel1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.0.content", textsContent1),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.0.content_type", textsContentTypes[0]),
				),
			},
			{
				// Update with new name and texts and add remaining values
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResourceLabel1,
					libraryName1,
				) + respManagementRespAsset.GenerateResponseManagementResponseAssetResource(
					assetResourceLabel,
					fullPath,
					util.NullValue,
				) + GenerateResponseManagementResponseResource(
					responseResourceLabel,
					name2,
					[]string{"genesyscloud_responsemanagement_library." + libraryResourceLabel1 + ".id"},
					strconv.Quote(interactionTypes[0]),
					util.GenerateJsonSchemaDocStr(substitutionsSchema),
					strconv.Quote(responseTypes[3]),
					[]string{"genesyscloud_responsemanagement_responseasset." + assetResourceLabel + ".id"},
					generateFooterBlock(footerType, footerResource),
					GenerateTextsBlock(
						textsContent2,
						textsContentTypes[1],
						util.NullValue,
					),
					generateSubstitutionsBlock(
						substitutionsId,
						substitutionsDescription,
						substitutionsDefaultValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "name", name2),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResourceLabel, "library_ids.0",
						"genesyscloud_responsemanagement_library."+libraryResourceLabel1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.0.content", textsContent2),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.0.content_type", textsContentTypes[1]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "interaction_type", interactionTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions.0.id", substitutionsId),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions.0.description", substitutionsDescription),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions.0.default_value", substitutionsDefaultValue),
					util.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions_schema_id", "type", "object"),
					util.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions_schema_id", "properties."+substitutionsSchema+".type", "string"),
					util.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions_schema_id", "required", substitutionsSchema),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "response_type", responseTypes[3]),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResourceLabel, "asset_ids.0",
						"genesyscloud_responsemanagement_responseasset."+assetResourceLabel, "id"),
				),
			},
			{
				// Add more texts and change libraries
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResourceLabel1,
					libraryName1,
				) + respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResourceLabel2,
					libraryName2,
				) + respManagementRespAsset.GenerateResponseManagementResponseAssetResource(
					assetResourceLabel,
					fullPath,
					util.NullValue,
				) + GenerateResponseManagementResponseResource(
					responseResourceLabel,
					name2,
					[]string{
						"genesyscloud_responsemanagement_library." + libraryResourceLabel1 + ".id",
						"genesyscloud_responsemanagement_library." + libraryResourceLabel2 + ".id",
					},
					strconv.Quote(interactionTypes[0]),
					util.GenerateJsonSchemaDocStr(substitutionsSchema),
					strconv.Quote(responseTypes[3]),
					[]string{"genesyscloud_responsemanagement_responseasset." + assetResourceLabel + ".id"},
					GenerateTextsBlock(
						textsContent1,
						textsContentTypes[0],
						util.NullValue,
					),
					generateFooterBlock(footerType, footerResource),
					GenerateTextsBlock(
						textsContent2,
						textsContentTypes[1],
						util.NullValue,
					),
					generateSubstitutionsBlock(
						substitutionsId,
						substitutionsDescription,
						substitutionsDefaultValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "library_ids.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.0.content", textsContent2),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.0.content_type", textsContentTypes[1]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.1.content", textsContent1),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.1.content_type", textsContentTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "interaction_type", interactionTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions.0.id", substitutionsId),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions.0.description", substitutionsDescription),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions.0.default_value", substitutionsDefaultValue),
					util.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions_schema_id", "type", "object"),
					util.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions_schema_id", "properties."+substitutionsSchema+".type", "string"),
					util.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions_schema_id", "required", substitutionsSchema),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "response_type", responseTypes[3]),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResourceLabel, "asset_ids.0",
						"genesyscloud_responsemanagement_responseasset."+assetResourceLabel, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_responsemanagement_response." + responseResourceLabel,
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
		responseResourceLabel     = "response-resource-message"
		name1                     = "Response-message" + uuid.NewString()
		textsContent1             = "Random text block content string"
		textsContentTypes         = []string{"text/plain", "text/html"}
		interactionTypes          = []string{"chat", "email", "twitter"}
		substitutionsId           = "sub123"
		substitutionsDescription  = "Substitutions description"
		substitutionsDefaultValue = "Substitutions default value"
		substitutionsSchema       = "schema document"
		responseTypes             = []string{`MessagingTemplate`, `CampaignSmsTemplate`, `CampaignEmailTemplate`}
		templateName              = "Sample template name message"
		templateNamespace         = "Template namespace message"

		// Responses Updated values
		name2         = "Response-" + uuid.NewString()
		textsContent2 = "Random text block content string new"

		// Library resources variables
		libraryResourceLabel1 = "library-resource1-message"
		libraryName1          = "ReferencelibraryMessage1"
		libraryResourceLabel2 = "library-resource2-message"
		libraryName2          = "ReferencelibraryMessage2"

		// Asset resources variables
		testFilesDir       = "test_responseasset_data"
		assetResourceLabel = "asset-resource-response-message"
		fileName           = "genesys-img-asset.png"
		fullPath           = filepath.Join(testFilesDir, fileName)
	)

	cleanupResponseAssets("genesys")
	cleanupResponseAssets("yeti")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with required values
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResourceLabel1,
					libraryName1,
				) + GenerateResponseManagementResponseResource(
					responseResourceLabel,
					name1,
					[]string{"genesyscloud_responsemanagement_library." + libraryResourceLabel1 + ".id"},
					util.NullValue,
					util.NullValue,
					util.NullValue,
					[]string{},
					GenerateTextsBlock(
						textsContent1,
						textsContentTypes[0],
						util.NullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "name", name1),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResourceLabel, "library_ids.0",
						"genesyscloud_responsemanagement_library."+libraryResourceLabel1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.0.content", textsContent1),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.0.content_type", textsContentTypes[0]),
				),
			},
			{
				// Update with new name and texts and add remaining values
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResourceLabel1,
					libraryName1,
				) + respManagementRespAsset.GenerateResponseManagementResponseAssetResource(
					assetResourceLabel,
					fullPath,
					util.NullValue,
				) + GenerateResponseManagementResponseResource(
					responseResourceLabel,
					name2,
					[]string{"genesyscloud_responsemanagement_library." + libraryResourceLabel1 + ".id"},
					strconv.Quote(interactionTypes[0]),
					util.GenerateJsonSchemaDocStr(substitutionsSchema),
					strconv.Quote(responseTypes[0]),
					[]string{"genesyscloud_responsemanagement_responseasset." + assetResourceLabel + ".id"},
					GenerateTextsBlock(
						textsContent2,
						textsContentTypes[1],
						util.NullValue,
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
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "name", name2),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResourceLabel, "library_ids.0",
						"genesyscloud_responsemanagement_library."+libraryResourceLabel1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.0.content", textsContent2),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.0.content_type", textsContentTypes[1]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "interaction_type", interactionTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions.0.id", substitutionsId),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions.0.description", substitutionsDescription),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions.0.default_value", substitutionsDefaultValue),
					util.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions_schema_id", "type", "object"),
					util.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions_schema_id", "properties."+substitutionsSchema+".type", "string"),
					util.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions_schema_id", "required", substitutionsSchema),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "response_type", responseTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "messaging_template.0.whats_app.0.name", templateName),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "messaging_template.0.whats_app.0.namespace", templateNamespace),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "messaging_template.0.whats_app.0.language", "en_US"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResourceLabel, "asset_ids.0",
						"genesyscloud_responsemanagement_responseasset."+assetResourceLabel, "id"),
				),
			},
			{
				// Add more texts and change libraries
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResourceLabel1,
					libraryName1,
				) + respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResourceLabel2,
					libraryName2,
				) + respManagementRespAsset.GenerateResponseManagementResponseAssetResource(
					assetResourceLabel,
					fullPath,
					util.NullValue,
				) + GenerateResponseManagementResponseResource(
					responseResourceLabel,
					name2,
					[]string{"genesyscloud_responsemanagement_library." + libraryResourceLabel2 + ".id", "genesyscloud_responsemanagement_library." + libraryResourceLabel1 + ".id"},
					strconv.Quote(interactionTypes[0]),
					util.GenerateJsonSchemaDocStr(substitutionsSchema),
					strconv.Quote(responseTypes[0]),
					[]string{"genesyscloud_responsemanagement_responseasset." + assetResourceLabel + ".id"},
					GenerateTextsBlock(
						textsContent1,
						textsContentTypes[0],
						util.NullValue,
					),
					GenerateTextsBlock(
						textsContent2,
						textsContentTypes[1],
						util.NullValue,
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
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "library_ids.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.0.content", textsContent2),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.0.content_type", textsContentTypes[1]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.1.content", textsContent1),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.1.content_type", textsContentTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "interaction_type", interactionTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions.0.id", substitutionsId),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions.0.description", substitutionsDescription),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions.0.default_value", substitutionsDefaultValue),
					util.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions_schema_id", "type", "object"),
					util.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions_schema_id", "properties."+substitutionsSchema+".type", "string"),
					util.ValidateValueInJsonAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "substitutions_schema_id", "required", substitutionsSchema),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "response_type", responseTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "messaging_template.0.whats_app.0.name", templateName),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "messaging_template.0.whats_app.0.namespace", templateNamespace),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "messaging_template.0.whats_app.0.language", "en_US"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResourceLabel, "asset_ids.0",
						"genesyscloud_responsemanagement_responseasset."+assetResourceLabel, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_responsemanagement_response." + responseResourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"substitutions_schema_id", "messaging_template", "response_type"},
			},
		},
		CheckDestroy: testVerifyResponseManagementResponseDestroyed,
	})
	cleanupResponseAssets(testFilesDir)
}

func TestAccResourceResponseManagementResponseCampaignEmailTemplate(t *testing.T) {
	t.Parallel()
	var (
		// Responses initial values
		responseResourceLabel = "response-resource-campaignemail"
		name1                 = "Response-message" + uuid.NewString()
		textsContentSubject   = "Text as Subject"
		textsContentBody      = "Text as Body! Welocme to Genesys!"
		textsContentTypes     = []string{"text/plain", "text/html"}
		textsType             = []string{"subject", "body"}
		interactionTypes      = []string{"chat", "email", "twitter"}
		responseTypes         = []string{`MessagingTemplate`, `CampaignSmsTemplate`, `CampaignEmailTemplate`}

		// Library resources variables
		libraryResourceLabel1 = "library-resource1-campaignemail"
		libraryName1          = "ReferencelibraryCampaignemail1"
	)

	cleanupResponseAssets("genesys")
	cleanupResponseAssets("yeti")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with required values
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResourceLabel1,
					libraryName1,
				) + GenerateResponseManagementResponseResource(
					responseResourceLabel,
					name1,
					[]string{"genesyscloud_responsemanagement_library." + libraryResourceLabel1 + ".id"},
					strconv.Quote(interactionTypes[1]),
					util.NullValue,
					strconv.Quote(responseTypes[2]),
					[]string{},
					GenerateTextsBlock(
						textsContentSubject,
						textsContentTypes[0],
						strconv.Quote(textsType[0]),
					),
					GenerateTextsBlock(
						textsContentBody,
						textsContentTypes[1],
						strconv.Quote(textsType[1]),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "name", name1),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_responsemanagement_response."+responseResourceLabel, "library_ids.0",
						"genesyscloud_responsemanagement_library."+libraryResourceLabel1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.0.content", textsContentBody),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.0.content_type", textsContentTypes[1]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.0.type", textsType[1]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.1.content", textsContentSubject),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.1.content_type", textsContentTypes[0]),
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_response."+responseResourceLabel, "texts.1.type", textsType[0]),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_responsemanagement_response." + responseResourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"substitutions_schema_id", "messaging_template", "response_type"},
			},
		},
		CheckDestroy: testVerifyResponseManagementResponseDestroyed,
	})

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
		} else if util.IsStatus404(resp) {
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

func cleanupResponseAssets(folderName string) error {
	var (
		name    = "name"
		fields  = []string{name}
		varType = "STARTS_WITH"
	)
	config, err := provider.AuthorizeSdk()
	if err != nil {
		return err
	}
	respManagementApi := platformclientv2.NewResponseManagementApiWithConfig(config)

	var filter = platformclientv2.Responseassetfilter{
		Fields:  &fields,
		Value:   &folderName,
		VarType: &varType,
	}

	var body = platformclientv2.Responseassetsearchrequest{
		Query:  &[]platformclientv2.Responseassetfilter{filter},
		SortBy: &name,
	}

	responseData, _, err := respManagementApi.PostResponsemanagementResponseassetsSearch(body, nil)
	if err != nil {
		log.Printf("Failed to search assets %s", err)
		return err
	}

	if responseData.Results != nil && len(*responseData.Results) > 0 {
		for _, result := range *responseData.Results {
			_, err = respManagementApi.DeleteResponsemanagementResponseasset(*result.Id)
			if err != nil {
				log.Printf("Failed to delete response assets %s: %v", *result.Id, err)
			}
		}
	}
	return nil
}
