package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
)

func getAllRoutingLanguages(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		languages, _, getErr := routingAPI.GetRoutingLanguages(100, pageNum, "", "", nil)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of languages: %v", getErr)
		}

		if languages.Entities == nil || len(*languages.Entities) == 0 {
			break
		}

		for _, language := range *languages.Entities {
			if *language.State != "deleted" {
				resources[*language.Id] = &ResourceMeta{Name: *language.Name}
			}
		}
	}

	return resources, nil
}

func routingLanguageExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllRoutingLanguages),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceRoutingLanguage() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Language",

		CreateContext: createWithPooledClient(createRoutingLanguage),
		ReadContext:   readWithPooledClient(readRoutingLanguage),
		DeleteContext: deleteWithPooledClient(deleteRoutingLanguage),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Language name.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func createRoutingLanguage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
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
	sdkConfig := meta.(*providerMeta).ClientConfig
	languagesAPI := platformclientv2.NewLanguagesApiWithConfig(sdkConfig)

	log.Printf("Reading language %s", d.Id())
	language, resp, getErr := languagesAPI.GetRoutingLanguage(d.Id())
	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read language %s: %s", d.Id(), getErr)
	}

	if language.State != nil && *language.State == "deleted" {
		d.SetId("")
		return nil
	}

	d.Set("name", *language.Name)
	log.Printf("Read language %s %s", d.Id(), *language.Name)
	return nil
}

func deleteRoutingLanguage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	languagesAPI := platformclientv2.NewLanguagesApiWithConfig(sdkConfig)

	log.Printf("Deleting language %s", name)
	_, err := languagesAPI.DeleteRoutingLanguage(d.Id())

	if err != nil {
		return diag.Errorf("Failed to delete language %s: %s", name, err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		routingLanguage, resp, err := languagesAPI.GetRoutingLanguage(d.Id())
		if err != nil {
			if resp != nil && resp.StatusCode == 404 {
				// Routing language deleted
				log.Printf("Deleted Routing language %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting Routing language %s: %s", d.Id(), err))
		}

		if *routingLanguage.State == "deleted" {
			// Routing language deleted
			log.Printf("Deleted Routing language %s", d.Id())
			return nil
		}

		return resource.RetryableError(fmt.Errorf("Routing language %s still exists", d.Id()))
	})
}
