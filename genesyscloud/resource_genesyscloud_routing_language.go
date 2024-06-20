package genesyscloud

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
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"
)

func getAllRoutingLanguages(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		languages, resp, getErr := routingAPI.GetRoutingLanguages(pageSize, pageNum, "", "", nil)
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_routing_language", fmt.Sprintf("Failed to get page of languages: %v", getErr), resp)
		}

		if languages.Entities == nil || len(*languages.Entities) == 0 {
			break
		}

		for _, language := range *languages.Entities {
			if language.State != nil && *language.State != "deleted" {
				resources[*language.Id] = &resourceExporter.ResourceMeta{Name: *language.Name}
			}
		}
	}

	return resources, nil
}

func RoutingLanguageExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRoutingLanguages),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceRoutingLanguage() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Language",

		CreateContext: provider.CreateWithPooledClient(createRoutingLanguage),
		ReadContext:   provider.ReadWithPooledClient(readRoutingLanguage),
		DeleteContext: provider.DeleteWithPooledClient(deleteRoutingLanguage),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Language name. Changing the language_name attribute will cause the language object to be dropped and recreated with a new ID.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func createRoutingLanguage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Creating language %s", name)
	language, resp, err := routingAPI.PostRoutingLanguages(platformclientv2.Language{
		Name: &name,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_routing_language", fmt.Sprintf("Failed to create language %s error: %s", name, err), resp)
	}

	d.SetId(*language.Id)

	log.Printf("Created language %s %s", name, *language.Id)
	return readRoutingLanguage(ctx, d, meta)
}

func readRoutingLanguage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingLanguage(), constants.DefaultConsistencyChecks, "genesyscloud_routing_language")

	log.Printf("Reading language %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		language, resp, getErr := routingApi.GetRoutingLanguage(d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_language", fmt.Sprintf("Failed to read language %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_language", fmt.Sprintf("Failed to read language %s | error: %s", d.Id(), getErr), resp))
		}

		if language.State != nil && *language.State == "deleted" {
			d.SetId("")
			return nil
		}

		d.Set("name", *language.Name)
		log.Printf("Read language %s %s", d.Id(), *language.Name)
		return cc.CheckState(d)
	})
}

func deleteRoutingLanguage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Deleting language %s", name)
	resp, err := routingApi.DeleteRoutingLanguage(d.Id())

	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_routing_language", fmt.Sprintf("Failed to delete language %s error: %s", name, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		routingLanguage, resp, err := routingApi.GetRoutingLanguage(d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Routing language deleted
				log.Printf("Deleted Routing language %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_language", fmt.Sprintf("Error deleting Routing language %s | error: %s", d.Id(), err), resp))
		}

		if routingLanguage.State != nil && *routingLanguage.State == "deleted" {
			// Routing language deleted
			log.Printf("Deleted Routing language %s", d.Id())
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_language", fmt.Sprintf("Routing language %s still exists", d.Id()), resp))
	})
}

func GenerateRoutingLanguageResource(
	resourceID string,
	name string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_language" "%s" {
		name = "%s"
	}
	`, resourceID, name)
}
