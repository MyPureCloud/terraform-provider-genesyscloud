package responsemanagement_response

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"
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
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
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

func TestAccResourceResponseManagementResponseFormField(t *testing.T) {
	t.Parallel()
	var (
		responseResourceLabel = "response-resource-form"
		responsePath          = "genesyscloud_responsemanagement_response." + responseResourceLabel
		name1                 = "Response-form-" + uuid.NewString()
		name2                 = "Response-form-" + uuid.NewString()

		// Form initial values
		formDescription1 = "A form for customer feedback"
		receivedTitle1   = "Thank you for your feedback"
		replyTitle1      = "Your response has been received"
		introTitle1      = "Customer Feedback Form"
		introButtonText  = "Start Survey"
		pageTitle1       = "Service Rating"
		pageSubtitle1    = "How would you rate our service?"
		sectionTitle1    = "Select your rating"

		// Form updated values
		formDescription2 = "An updated form for customer feedback"
		receivedTitle2   = "Feedback received"
		pageTitle2       = "Visit Date"
		pageSubtitle2    = "When did you visit us?"

		// Library resource variables
		libraryResourceLabel = "library-resource-form"
		libraryName          = "ReferencelibraryForm1"
	)

	initialForm := generateFormBlock(
		formDescription1,
		util.TrueValue,
		generateFormMessageBlock("received_message", receivedTitle1, "We appreciate your input"),
		generateFormMessageBlock("reply_message", replyTitle1, "We will review your feedback"),
		generateFormIntroductionBlock(introTitle1, "Please help us improve our service", introButtonText),
		generateFormPageBlock(
			pageTitle1,
			pageSubtitle1,
			generateListPickerComponentBlock(sectionTitle1, util.FalseValue, []string{"Excellent", "Good", "Poor"}),
		),
	)

	// Update: change the description and received message, and swap the single list picker page
	// for two pages exercising the remaining component types.
	updatedForm := generateFormBlock(
		formDescription2,
		util.FalseValue,
		generateFormMessageBlock("received_message", receivedTitle2, "We appreciate your input"),
		generateFormMessageBlock("reply_message", replyTitle1, "We will review your feedback"),
		generateFormPageBlock(
			pageTitle2,
			pageSubtitle2,
			generateDatePickerComponentBlock("Select Date", "Choose the date of your visit", "dayMonthYear"),
		),
		generateFormPageBlock(
			"Additional Comments",
			"Please share any additional feedback",
			generateInputComponentBlock("Comments", "Enter your comments here...", util.TrueValue, util.FalseValue, "Default"),
		),
		generateFormPageBlock(
			"Recommendation",
			"Would you recommend us to others?",
			generateWheelPickerComponentBlock(map[string]string{"Definitely": "definitely", "Maybe": "maybe"}),
		),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create a Form response
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResourceLabel,
					libraryName,
				) + GenerateResponseManagementResponseResource(
					responseResourceLabel,
					name1,
					[]string{"genesyscloud_responsemanagement_library." + libraryResourceLabel + ".id"},
					util.NullValue,
					util.NullValue,
					strconv.Quote("Form"),
					[]string{},
					// No texts block: Form responses ignore texts and always return an
					// empty list, so setting it would produce a permanent diff.
					initialForm,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(responsePath, "name", name1),
					resource.TestCheckResourceAttr(responsePath, "response_type", "Form"),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_description", formDescription1),
					resource.TestCheckResourceAttr(responsePath, "form.0.show_summary", "true"),
					resource.TestCheckResourceAttr(responsePath, "form.0.received_message.0.title", receivedTitle1),
					resource.TestCheckResourceAttr(responsePath, "form.0.reply_message.0.title", replyTitle1),
					resource.TestCheckResourceAttr(responsePath, "form.0.introduction.0.title", introTitle1),
					resource.TestCheckResourceAttr(responsePath, "form.0.introduction.0.button_text", introButtonText),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.#", "1"),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.0.title", pageTitle1),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.0.subtitle", pageSubtitle1),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.0.page_components.0.form_component_type", "ListPicker"),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.0.page_components.0.list_picker.0.sections.0.title", sectionTitle1),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.0.page_components.0.list_picker.0.sections.0.multiple_selection", "false"),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.0.page_components.0.list_picker.0.sections.0.items.#", "3"),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.0.page_components.0.list_picker.0.sections.0.items.0.title", "Excellent"),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.0.page_components.0.list_picker.0.sections.0.items.2.title", "Poor"),
				),
			},
			{
				// Update the form: new name, description, show_summary and page set
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResourceLabel,
					libraryName,
				) + GenerateResponseManagementResponseResource(
					responseResourceLabel,
					name2,
					[]string{"genesyscloud_responsemanagement_library." + libraryResourceLabel + ".id"},
					util.NullValue,
					util.NullValue,
					strconv.Quote("Form"),
					[]string{},
					updatedForm,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(responsePath, "name", name2),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_description", formDescription2),
					resource.TestCheckResourceAttr(responsePath, "form.0.show_summary", "false"),
					resource.TestCheckResourceAttr(responsePath, "form.0.received_message.0.title", receivedTitle2),
					// The introduction block was removed on update
					resource.TestCheckResourceAttr(responsePath, "form.0.introduction.#", "0"),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.#", "3"),
					// Page ordering must be preserved
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.0.title", pageTitle2),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.0.page_components.0.form_component_type", "DatePicker"),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.0.page_components.0.date_picker.0.date_display_format", "dayMonthYear"),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.1.page_components.0.form_component_type", "Input"),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.1.page_components.0.input.0.is_multiple_line", "true"),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.1.page_components.0.input.0.is_required", "false"),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.1.page_components.0.input.0.keyboard_type", "Default"),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.2.page_components.0.form_component_type", "WheelPicker"),
					resource.TestCheckResourceAttr(responsePath, "form.0.form_pages.2.page_components.0.wheel_picker.0.items.#", "2"),
				),
			},
			{
				// Import/Read
				ResourceName:            responsePath,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"substitutions_schema_id", "messaging_template", "response_type"},
			},
		},
		CheckDestroy: testVerifyResponseManagementResponseDestroyed,
	})
}

func generateFormBlock(formDescription string, showSummary string, nestedBlocks ...string) string {
	return fmt.Sprintf(`
		form {
			form_description = "%s"
			show_summary     = %s
			%s
		}
	`, formDescription, showSummary, strings.Join(nestedBlocks, "\n"))
}

// generateFormMessageBlock builds a received_message or reply_message block.
func generateFormMessageBlock(blockName string, title string, subtitle string) string {
	return fmt.Sprintf(`
		%s {
			title    = "%s"
			subtitle = "%s"
		}
	`, blockName, title, subtitle)
}

func generateFormIntroductionBlock(title string, subtitle string, buttonText string) string {
	return fmt.Sprintf(`
		introduction {
			title       = "%s"
			subtitle    = "%s"
			button_text = "%s"
		}
	`, title, subtitle, buttonText)
}

func generateFormPageBlock(title string, subtitle string, pageComponents ...string) string {
	return fmt.Sprintf(`
		form_pages {
			title    = "%s"
			subtitle = "%s"
			%s
		}
	`, title, subtitle, strings.Join(pageComponents, "\n"))
}

func generateListPickerComponentBlock(sectionTitle string, multipleSelection string, itemTitles []string) string {
	var items strings.Builder
	for _, itemTitle := range itemTitles {
		items.WriteString(fmt.Sprintf(`
					items {
						title = "%s"
					}
		`, itemTitle))
	}
	return fmt.Sprintf(`
		page_components {
			form_component_type = "ListPicker"
			list_picker {
				sections {
					title              = "%s"
					multiple_selection = %s
					%s
				}
			}
		}
	`, sectionTitle, multipleSelection, items.String())
}

func generateDatePickerComponentBlock(title string, subtitle string, dateDisplayFormat string) string {
	return fmt.Sprintf(`
		page_components {
			form_component_type = "DatePicker"
			date_picker {
				title               = "%s"
				subtitle            = "%s"
				date_display_format = "%s"
			}
		}
	`, title, subtitle, dateDisplayFormat)
}

func generateInputComponentBlock(title string, placeholderText string, isMultipleLine string, isRequired string, keyboardType string) string {
	return fmt.Sprintf(`
		page_components {
			form_component_type = "Input"
			input {
				title            = "%s"
				placeholder_text = "%s"
				is_multiple_line = %s
				is_required      = %s
				keyboard_type    = "%s"
			}
		}
	`, title, placeholderText, isMultipleLine, isRequired, keyboardType)
}

// generateWheelPickerComponentBlock builds a wheel picker from a title -> value mapping.
func generateWheelPickerComponentBlock(items map[string]string) string {
	// Sort the titles so the generated config is deterministic across runs.
	titles := make([]string, 0, len(items))
	for title := range items {
		titles = append(titles, title)
	}
	sort.Strings(titles)

	var itemBlocks strings.Builder
	for _, title := range titles {
		itemBlocks.WriteString(fmt.Sprintf(`
				items {
					title = "%s"
					value = "%s"
				}
		`, title, items[title]))
	}
	return fmt.Sprintf(`
		page_components {
			form_component_type = "WheelPicker"
			wheel_picker {
				%s
			}
		}
	`, itemBlocks.String())
}
