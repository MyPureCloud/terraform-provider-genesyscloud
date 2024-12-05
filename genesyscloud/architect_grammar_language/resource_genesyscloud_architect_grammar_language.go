package architect_grammar_language

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_architect_grammar_language.go contains all the methods that perform the core logic for a resource.
*/

// getAllAuthArchitectGrammarLanguage retrieves all of the architect grammar languages via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthArchitectGrammarLanguage(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getArchitectGrammarLanguageProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	languages, resp, err := proxy.getAllArchitectGrammarLanguage(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get grammar languages: %v", err), resp)
	}

	for _, language := range *languages {
		languageId := *language.GrammarId + ":" + *language.Language
		resources[languageId] = &resourceExporter.ResourceMeta{BlockLabel: *language.Language}
	}

	return resources, nil
}

// createArchitectGrammarLanguage is used by the architect_grammar_language resource to create a Genesys cloud architect grammar language
func createArchitectGrammarLanguage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectGrammarLanguageProxy(sdkConfig)

	architectGrammarLanguage := getArchitectGrammarLanguageFromResourceData(d)

	log.Printf("Creating Architect Grammar Language %s for grammar %s", *architectGrammarLanguage.Language, *architectGrammarLanguage.GrammarId)
	language, resp, err := proxy.createArchitectGrammarLanguage(ctx, &architectGrammarLanguage)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create grammar language: %s error %s", d.Id(), err), resp)
	}

	// Language id is always in format <grammar-id>:<language-code>
	languageId := fmt.Sprintf("%s:%s", *language.GrammarId, *language.Language)
	d.SetId(languageId)
	log.Printf("Created Architect Grammar Language %s", languageId)
	return readArchitectGrammarLanguage(ctx, d, meta)
}

// readArchitectGrammarLanguage is used by the architect_grammar_language resource to read an architect grammar language from genesys cloud.
func readArchitectGrammarLanguage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectGrammarLanguageProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectGrammarLanguage(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Architect Grammar Language %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		grammarId, languageCode := splitLanguageId(d.Id())
		language, resp, getErr := proxy.getArchitectGrammarLanguageById(ctx, grammarId, languageCode)

		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Architect Grammar Language %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Architect Grammar Language %s: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "grammar_id", language.GrammarId)
		resourcedata.SetNillableValue(d, "language", language.Language)
		if language.VoiceFileMetadata != nil {
			_ = d.Set("voice_file_data", flattenGrammarLanguageFileMetadata(d, language.VoiceFileMetadata, Voice))
		}
		if language.DtmfFileMetadata != nil {
			_ = d.Set("dtmf_file_data", flattenGrammarLanguageFileMetadata(d, language.DtmfFileMetadata, Dtmf))
		}

		log.Printf("Read Architect Grammar Language %s", d.Id())
		return cc.CheckState(d)
	})
}

func splitLanguageId(languageId string) (string, string) {
	split := strings.SplitN(languageId, ":", 2)
	if len(split) == 2 {
		return split[0], split[1]
	}
	return "", ""
}

// updateArchitectGrammarLanguage is used by the architect_grammar_language resource to update an architect grammar language in Genesys Cloud
func updateArchitectGrammarLanguage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectGrammarLanguageProxy(sdkConfig)

	architectGrammarLanguage := getArchitectGrammarLanguageFromResourceData(d)

	log.Printf("Updating Architect Grammar Language %s", d.Id())
	_, resp, err := proxy.updateArchitectGrammarLanguage(ctx, *architectGrammarLanguage.GrammarId, *architectGrammarLanguage.Language, &architectGrammarLanguage)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update grammar language: %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated Architect Grammar Language %s", d.Id())
	return readArchitectGrammarLanguage(ctx, d, meta)
}

// deleteArchitectGrammarLanguage is used by the architect_grammar_language resource to delete an architect grammar language from Genesys cloud.
func deleteArchitectGrammarLanguage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectGrammarLanguageProxy(sdkConfig)

	grammarId, languageCode := splitLanguageId(d.Id())
	resp, err := proxy.deleteArchitectGrammarLanguage(ctx, grammarId, languageCode)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete grammar language %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getArchitectGrammarLanguageById(ctx, grammarId, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted Grammar Language %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("grammar Language %s still exists", d.Id()), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("grammar Language %s still exists", d.Id()), resp))
	})
}
