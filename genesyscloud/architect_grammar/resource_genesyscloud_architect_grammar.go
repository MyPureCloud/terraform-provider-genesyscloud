package architect_grammar

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v109/platformclientv2"
	"log"
	genesyscloud2 "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"
)

/*
The resource_genesyscloud_architect_grammar.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthArchitectGrammar retrieves all of the architect grammars via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthArchitectGrammar(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getArchitectGrammarProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	grammars, err := proxy.getAllArchitectGrammar(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get grammars: %v", err)
	}

	for _, grammar := range *grammars {
		log.Printf("Dealing with grammar id : %s", *grammar.Id)
		resources[*grammar.Id] = &resourceExporter.ResourceMeta{Name: *grammar.Id}
	}

	return resources, nil
}

// createArchitectGrammar is used by the architect_grammar resource to create a Genesys cloud architect grammar
func createArchitectGrammar(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*genesyscloud2.ProviderMeta).ClientConfig
	proxy := newArchitectGrammarProxy(sdkConfig)

	architectGrammar := getArchitectGrammarFromResourceData(d)

	// Create grammar
	log.Printf("Creating Architect Grammar %s", *architectGrammar.Name)
	grammar, err := proxy.createArchitectGrammar(ctx, &architectGrammar)
	if err != nil {
		return diag.Errorf("Failed to create grammar: %s", err)
	}

	// Create each language associated with the grammar
	for _, language := range *architectGrammar.Languages {
		_, err := proxy.createArchitectGrammarLanguage(ctx, *grammar.Id, &language)
		if err != nil {
			return diag.Errorf("Failed to create grammar language: %s", err)
		}

		grammar2, _, _ := proxy.getArchitectGrammarById(ctx, *grammar.Id)
		fmt.Println(formatJSON(grammar2))

		// Upload grammar voice file
		if language.VoiceFileMetadata.FileName != nil {
			uploadRequest := platformclientv2.Grammarfileuploadrequest{
				FileType: language.VoiceFileMetadata.FileType,
			}
			err = proxy.uploadGrammarLanguageFile(*grammar.Id, *language.Language, language.VoiceFileMetadata.FileName, &uploadRequest)
			if err != nil {
				return diag.Errorf("Failed to upload language file: %s", err)
			}
		}
	}

	grammar2, _, _ := proxy.getArchitectGrammarById(ctx, *grammar.Id)
	fmt.Println(formatJSON(grammar2))

	d.SetId(*grammar.Id)
	log.Printf("Created Architect Grammar %s", *grammar.Id)
	return readArchitectGrammar(ctx, d, meta)
}

// readArchitectGrammar is used by the architect_grammar resource to read an architect grammar from genesys cloud.
func readArchitectGrammar(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*genesyscloud2.ProviderMeta).ClientConfig
	proxy := newArchitectGrammarProxy(sdkConfig)

	log.Printf("Reading Architect Grammar %s", d.Id())

	return genesyscloud2.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		grammar, respCode, getErr := proxy.getArchitectGrammarById(ctx, d.Id())

		if getErr != nil {
			if genesyscloud2.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read Architect Grammar %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Architect Grammar %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectGrammar())

		resourcedata.SetNillableValue(d, "name", grammar.Name)
		resourcedata.SetNillableValue(d, "description", grammar.Description)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "languages", grammar.Languages, flattenGrammarLanguages)

		log.Printf("Read Architect Grammar %s %s", d.Id(), *grammar.Name)
		return cc.CheckState()
	})
}

// updateArchitectGrammar is used by the architect_grammar resource to update an architect grammar in Genesys Cloud
func updateArchitectGrammar(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*genesyscloud2.ProviderMeta).ClientConfig
	proxy := newArchitectGrammarProxy(sdkConfig)

	architectGrammar := getArchitectGrammarFromResourceData(d)

	grammar, err := proxy.updateArchitectGrammar(ctx, d.Id(), &architectGrammar)
	if err != nil {
		return diag.Errorf("Failed to update grammar: %s", err)
	}

	log.Printf("Updated Architect Grammar %s", *grammar.Id)
	return readArchitectGrammar(ctx, d, meta)
}

// deleteArchitectGrammar is used by the architect_grammar resource to delete an architect grammar from Genesys cloud.
func deleteArchitectGrammar(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*genesyscloud2.ProviderMeta).ClientConfig
	proxy := newArchitectGrammarProxy(sdkConfig)

	_, err := proxy.deleteArchitectGrammar(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete grammar %s: %s", d.Id(), err)
	}

	return genesyscloud2.WithRetries(ctx, 180*time.Second, func() *resource.RetryError {
		_, respCode, err := proxy.getArchitectGrammarById(ctx, d.Id())

		if err != nil {
			if genesyscloud2.IsStatus404ByInt(respCode) {
				log.Printf("Deleted Grammar %s", d.Id())
				return nil
			}

			return retry.NonRetryableError(fmt.Errorf("Error deleting grammar %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Grammar %s still exists", d.Id()))
	})
}
