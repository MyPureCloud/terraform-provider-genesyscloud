package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func getAllRoutingLanguages(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		languages, _, getErr := routingAPI.GetRoutingLanguages(pageSize, pageNum, "", "", nil)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of languages: %v", getErr)
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
		GetResourcesFunc: GetAllWithPooledClient(getAllRoutingLanguages),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceRoutingLanguage() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Language",

		CreateContext: CreateWithPooledClient(createRoutingLanguage),
		ReadContext:   ReadWithPooledClient(readRoutingLanguage),
		DeleteContext: DeleteWithPooledClient(deleteRoutingLanguage),
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

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Creating language %s", name)
	language, _, err := routingAPI.PostRoutingLanguages(platformclientv2.Language{
		Name: &name,
	})
	if err != nil {
		return diag.Errorf("Failed to create language %s: %s", name, err)
	}

	d.SetId(*language.Id)

	log.Printf("Created language %s %s", name, *language.Id)
	return readRoutingLanguage(ctx, d, meta)
}

func readRoutingLanguage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Reading language %s", d.Id())
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		language, resp, getErr := routingApi.GetRoutingLanguage(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read language %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read language %s: %s", d.Id(), getErr))
		}

		if language.State != nil && *language.State == "deleted" {
			d.SetId("")
			return nil
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingLanguage())
		d.Set("name", *language.Name)
		log.Printf("Read language %s %s", d.Id(), *language.Name)
		return cc.CheckState()
	})
}

func deleteRoutingLanguage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Deleting language %s", name)
	_, err := routingApi.DeleteRoutingLanguage(d.Id())

	if err != nil {
		return diag.Errorf("Failed to delete language %s: %s", name, err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		routingLanguage, resp, err := routingApi.GetRoutingLanguage(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Routing language deleted
				log.Printf("Deleted Routing language %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting Routing language %s: %s", d.Id(), err))
		}

		if routingLanguage.State != nil && *routingLanguage.State == "deleted" {
			// Routing language deleted
			log.Printf("Deleted Routing language %s", d.Id())
			return nil
		}

		return retry.RetryableError(fmt.Errorf("Routing language %s still exists", d.Id()))
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
