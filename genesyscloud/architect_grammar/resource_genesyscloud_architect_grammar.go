package architect_grammar

import (
	"context"
	"fmt"
	"log"
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
The resource_genesyscloud_architect_grammar.go contains all the methods that perform the core logic for a resource.
*/

// getAllAuthArchitectGrammar retrieves all the architect grammars via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthArchitectGrammar(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getArchitectGrammarProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	grammars, resp, err := proxy.getAllArchitectGrammar(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to retrieve all grammars: %s", err), resp)
	}

	for _, grammar := range *grammars {
		resources[*grammar.Id] = &resourceExporter.ResourceMeta{BlockLabel: *grammar.Name}
	}

	return resources, nil
}

// createArchitectGrammar is used by the architect_grammar_language resource to create a Genesys cloud architect grammar
func createArchitectGrammar(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectGrammarProxy(sdkConfig)

	architectGrammar := platformclientv2.Grammar{
		Name:        platformclientv2.String(d.Get("name").(string)),
		Description: platformclientv2.String(d.Get("description").(string)),
	}

	// Create grammar
	log.Printf("Creating Architect Grammar %s", *architectGrammar.Name)
	grammar, resp, err := proxy.createArchitectGrammar(ctx, &architectGrammar)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create grammar %s error: %s", d.Id(), err), resp)
	}

	d.SetId(*grammar.Id)
	log.Printf("Created Architect Grammar %s", *grammar.Id)
	return readArchitectGrammar(ctx, d, meta)
}

// readArchitectGrammar is used by the architect_grammar_language resource to read an architect grammar from genesys cloud.
func readArchitectGrammar(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectGrammarProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectGrammar(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Architect Grammar %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		grammar, resp, getErr := proxy.getArchitectGrammarById(ctx, d.Id())

		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Architect Grammar %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Architect Grammar %s: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", grammar.Name)
		resourcedata.SetNillableValue(d, "description", grammar.Description)

		log.Printf("Read Architect Grammar %s", d.Id())
		return cc.CheckState(d)
	})
}

// updateArchitectGrammar is used by the architect_grammar_language resource to update an architect grammar in Genesys Cloud
func updateArchitectGrammar(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectGrammarProxy(sdkConfig)

	architectGrammar := platformclientv2.Grammar{
		Name:        platformclientv2.String(d.Get("name").(string)),
		Description: platformclientv2.String(d.Get("description").(string)),
	}

	// Update grammar
	log.Printf("Updating Architect Grammar %s", *architectGrammar.Name)
	grammar, resp, err := proxy.updateArchitectGrammar(ctx, d.Id(), &architectGrammar)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update grammar: %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated Architect Grammar %s", *grammar.Id)
	return readArchitectGrammar(ctx, d, meta)
}

// deleteArchitectGrammar is used by the architect_grammar_language resource to delete an architect grammar from Genesys cloud.
func deleteArchitectGrammar(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectGrammarProxy(sdkConfig)

	resp, err := proxy.deleteArchitectGrammar(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete grammar %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getArchitectGrammarById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted Grammar %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting grammar %s: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("grammar %s still exists", d.Id()), resp))
	})
}

func GenerateGrammarResource(
	resourceLabel string,
	name string,
	description string,
) string {
	return fmt.Sprintf(`
		resource "genesyscloud_architect_grammar" "%s" {
			name = "%s"
			description = "%s"
		}
	`, resourceLabel, name, description)
}
