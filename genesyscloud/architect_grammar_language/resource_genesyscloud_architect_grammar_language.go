package architect_grammar_language

import (
	"context"
	"fmt"
	"log"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

/*
The resource_genesyscloud_architect_grammar_language.go contains all the methods that perform the core logic for a resource.
*/

// getAllAuthArchitectGrammarLanguage retrieves all of the architect grammar languages via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthArchitectGrammarLanguage(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getArchitectGrammarLanguageProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	languages, err := proxy.getAllArchitectGrammarLanguage(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get grammar languages: %v", err)
	}

	for _, language := range *languages {
		languageId := *language.GrammarId + ":" + *language.Language
		resources[languageId] = &resourceExporter.ResourceMeta{Name: *language.Language}
	}

	return resources, nil
}

// createArchitectGrammarLanguage is used by the architect_grammar_language resource to create a Genesys cloud architect grammar language
func createArchitectGrammarLanguage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newArchitectGrammarLanguageProxy(sdkConfig)

	architectGrammarLanguage := getArchitectGrammarLanguageFromResourceData(d)

	log.Printf("Creating Architect Grammar Language %s for grammar %s", *architectGrammarLanguage.Language, *architectGrammarLanguage.GrammarId)
	language, err := proxy.createArchitectGrammarLanguage(ctx, &architectGrammarLanguage)
	if err != nil {
		return diag.Errorf("Failed to create grammar language: %s", err)
	}

	// Language id is always in format <grammar-id>:<language-code>
	d.SetId(*language.GrammarId + ":" + *language.Language)
	log.Printf("Created Architect Grammar Language %s", *language.GrammarId+":"+*language.Language)
	return readArchitectGrammarLanguage(ctx, d, meta)
}

// readArchitectGrammarLanguage is used by the architect_grammar_language resource to read an architect grammar language from genesys cloud.
func readArchitectGrammarLanguage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newArchitectGrammarLanguageProxy(sdkConfig)

	log.Printf("Reading Architect Grammar Language %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		grammarId, languageCode := splitLanguageId(d.Id())
		language, respCode, getErr := proxy.getArchitectGrammarLanguageById(ctx, grammarId, languageCode)

		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read Architect Grammar Language %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Architect Grammar Language %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectGrammarLanguage())

		resourcedata.SetNillableValue(d, "grammar_id", language.GrammarId)
		resourcedata.SetNillableValue(d, "language", language.Language)
		if language.VoiceFileMetadata != nil {
			d.Set("voice_file_data", flattenGrammarLanguageFileMetadata(d, language.VoiceFileMetadata, Voice))
		}
		if language.DtmfFileMetadata != nil {
			d.Set("dtmf_file_data", flattenGrammarLanguageFileMetadata(d, language.DtmfFileMetadata, Dtmf))
		}

		log.Printf("Read Architect Grammar Language %s", d.Id())
		return cc.CheckState()
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
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newArchitectGrammarLanguageProxy(sdkConfig)

	architectGrammarLanguage := getArchitectGrammarLanguageFromResourceData(d)

	log.Printf("Updating Architect Grammar Language %s", d.Id())
	_, err := proxy.updateArchitectGrammarLanguage(ctx, *architectGrammarLanguage.GrammarId, *architectGrammarLanguage.Language, &architectGrammarLanguage)
	if err != nil {
		return diag.Errorf("Failed to update grammar language: %s", err)
	}

	log.Printf("Updated Architect Grammar Language %s", d.Id())
	return readArchitectGrammarLanguage(ctx, d, meta)
}

// deleteArchitectGrammarLanguage is used by the architect_grammar_language resource to delete an architect grammar language from Genesys cloud.
func deleteArchitectGrammarLanguage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newArchitectGrammarLanguageProxy(sdkConfig)

	grammarId, languageCode := splitLanguageId(d.Id())
	_, err := proxy.deleteArchitectGrammarLanguage(ctx, grammarId, languageCode)
	if err != nil {
		return diag.Errorf("Failed to delete grammar language %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getArchitectGrammarLanguageById(ctx, grammarId, d.Id())

		if err != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				log.Printf("Deleted Grammar Language %s", d.Id())
				return nil
			}

			return retry.NonRetryableError(fmt.Errorf("Error deleting grammar language %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Grammar Language %s still exists", d.Id()))
	})
}
