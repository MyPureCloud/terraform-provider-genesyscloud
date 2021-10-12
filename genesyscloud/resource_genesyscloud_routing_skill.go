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

func getAllRoutingSkills(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		skills, _, getErr := routingAPI.GetRoutingSkills(100, pageNum, "", nil)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of skills: %v", getErr)
		}

		if skills.Entities == nil || len(*skills.Entities) == 0 {
			break
		}

		for _, skill := range *skills.Entities {
			if *skill.State != "deleted" {
				resources[*skill.Id] = &ResourceMeta{Name: *skill.Name}
			}
		}
	}

	return resources, nil
}

func routingSkillExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllRoutingSkills),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceRoutingSkill() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Skill",

		CreateContext: createWithPooledClient(createRoutingSkill),
		ReadContext:   readWithPooledClient(readRoutingSkill),
		DeleteContext: deleteWithPooledClient(deleteRoutingSkill),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Skill name.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func createRoutingSkill(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Creating skill %s", name)
	skill, _, err := routingAPI.PostRoutingSkills(platformclientv2.Routingskill{
		Name: &name,
	})
	if err != nil {
		return diag.Errorf("Failed to create skill %s: %s", name, err)
	}

	d.SetId(*skill.Id)

	log.Printf("Created skill %s %s", name, *skill.Id)
	return readRoutingSkill(ctx, d, meta)
}

func readRoutingSkill(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Reading skill %s", d.Id())
	return withRetriesForRead(ctx, 30*time.Second, d, func() *resource.RetryError {
		skill, resp, getErr := routingAPI.GetRoutingSkill(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read skill %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read skill %s: %s", d.Id(), getErr))
		}

		if skill.State != nil && *skill.State == "deleted" {
			d.SetId("")
			return nil
		}

		d.Set("name", *skill.Name)
		log.Printf("Read skill %s %s", d.Id(), *skill.Name)
		return nil
	})
}

func deleteRoutingSkill(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Deleting skill %s", name)
	_, err := routingAPI.DeleteRoutingSkill(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete skill %s: %s", name, err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		routingSkill, resp, err := routingAPI.GetRoutingSkill(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// Routing skill deleted
				log.Printf("Deleted Routing skill %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting Routing skill %s: %s", d.Id(), err))
		}

		if *routingSkill.State == "deleted" {
			// Routing skill deleted
			log.Printf("Deleted Routing skill %s", d.Id())
			return nil
		}

		return resource.RetryableError(fmt.Errorf("Routing skill %s still exists", d.Id()))
	})
}
