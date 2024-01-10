package architect_grammar

import (
	"context"
	"fmt"
	"log"
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
The resource_genesyscloud_architect_grammar.go contains all the methods that perform the core logic for a resource.
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
		resources[*grammar.Id] = &resourceExporter.ResourceMeta{Name: *grammar.Name}
	}

	return resources, nil
}

// createArchitectGrammar is used by the architect_grammar_language resource to create a Genesys cloud architect grammar
func createArchitectGrammar(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newArchitectGrammarProxy(sdkConfig)

	architectGrammar := platformclientv2.Grammar{
		Name:        platformclientv2.String(d.Get("name").(string)),
		Description: platformclientv2.String(d.Get("description").(string)),
	}

	// Create grammar
	log.Printf("Creating Architect Grammar %s", *architectGrammar.Name)
	grammar, err := proxy.createArchitectGrammar(ctx, &architectGrammar)
	if err != nil {
		return diag.Errorf("Failed to create grammar: %s", err)
	}

	d.SetId(*grammar.Id)
	log.Printf("Created Architect Grammar %s", *grammar.Id)
	return readArchitectGrammar(ctx, d, meta)
}

// readArchitectGrammar is used by the architect_grammar_language resource to read an architect grammar from genesys cloud.
func readArchitectGrammar(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newArchitectGrammarProxy(sdkConfig)

	log.Printf("Reading Architect Grammar %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		grammar, respCode, getErr := proxy.getArchitectGrammarById(ctx, d.Id())

		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read Architect Grammar %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Architect Grammar %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectGrammar())

		resourcedata.SetNillableValue(d, "name", grammar.Name)
		resourcedata.SetNillableValue(d, "description", grammar.Description)

		log.Printf("Read Architect Grammar %s", d.Id())
		return cc.CheckState()
	})
}

// updateArchitectGrammar is used by the architect_grammar_language resource to update an architect grammar in Genesys Cloud
func updateArchitectGrammar(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newArchitectGrammarProxy(sdkConfig)

	architectGrammar := platformclientv2.Grammar{
		Name:        platformclientv2.String(d.Get("name").(string)),
		Description: platformclientv2.String(d.Get("description").(string)),
	}

	// Update grammar
	log.Printf("Updating Architect Grammar %s", *architectGrammar.Name)
	grammar, err := proxy.updateArchitectGrammar(ctx, d.Id(), &architectGrammar)
	if err != nil {
		return diag.Errorf("Failed to update grammar: %s", err)
	}

	log.Printf("Updated Architect Grammar %s", *grammar.Id)
	return readArchitectGrammar(ctx, d, meta)
}

// deleteArchitectGrammar is used by the architect_grammar_language resource to delete an architect grammar from Genesys cloud.
func deleteArchitectGrammar(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newArchitectGrammarProxy(sdkConfig)

	_, err := proxy.deleteArchitectGrammar(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete grammar %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getArchitectGrammarById(ctx, d.Id())

		if err != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				log.Printf("Deleted Grammar %s", d.Id())
				return nil
			}

			return retry.NonRetryableError(fmt.Errorf("Error deleting grammar %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Grammar %s still exists", d.Id()))
	})
}

func GenerateGrammarResource(
	resourceId string,
	name string,
	description string,
) string {
	return fmt.Sprintf(`
		resource "genesyscloud_architect_grammar" "%s" {
			name = "%s"
			description = "%s"
		}
	`, resourceId, name, description)
}
