package routing_language

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllRoutingLanguages(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getRoutingLanguageProxy(clientConfig)

	languages, resp, getErr := proxy.getAllRoutingLanguages(ctx, "")
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of languages: %v", getErr), resp)
	}

	if languages == nil || len(*languages) == 0 {
		return resources, nil
	}

	for _, language := range *languages {
		if language.State != nil && *language.State != "deleted" {
			resources[*language.Id] = &resourceExporter.ResourceMeta{BlockLabel: *language.Name}
		}
	}
	return resources, nil
}

func createRoutingLanguage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingLanguageProxy(sdkConfig)
	name := d.Get("name").(string)

	log.Printf("Creating language %s", name)

	language, resp, err := proxy.createRoutingLanguage(ctx, &platformclientv2.Language{
		Name: &name,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create language %s error: %s", name, err), resp)
	}

	d.SetId(*language.Id)

	log.Printf("Created language %s %s", name, *language.Id)
	return readRoutingLanguage(ctx, d, meta)
}

func readRoutingLanguage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingLanguageProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingLanguage(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading routing language %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		language, resp, getErr := proxy.getRoutingLanguageById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read language %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read language %s | error: %s", d.Id(), getErr), resp))
		}

		if language.State != nil && *language.State == "deleted" {
			d.SetId("")
			return nil
		}

		_ = d.Set("name", *language.Name)
		log.Printf("Read routing language %s %s", d.Id(), *language.Name)
		return cc.CheckState(d)
	})
}

func deleteRoutingLanguage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingLanguageProxy(sdkConfig)
	name := d.Get("name").(string)

	log.Printf("Deleting language %s", name)
	resp, err := proxy.deleteRoutingLanguage(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete language %s error: %s", name, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		routingLanguage, resp, err := proxy.getRoutingLanguageById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Routing language deleted
				log.Printf("Deleted Routing language %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting Routing language %s | error: %s", d.Id(), err), resp))
		}

		if routingLanguage.State != nil && *routingLanguage.State == "deleted" {
			// Routing language deleted
			log.Printf("Deleted Routing language %s", d.Id())
			return nil
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Routing language %s still exists", d.Id()), resp))
	})
}
