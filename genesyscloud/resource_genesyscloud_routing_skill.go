package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v99/platformclientv2"
)

func getAllRoutingSkills(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		skills, _, getErr := routingAPI.GetRoutingSkills(pageSize, pageNum, "", nil)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of skills: %v", getErr)
		}

		if skills.Entities == nil || len(*skills.Entities) == 0 {
			break
		}

		for _, skill := range *skills.Entities {
			if skill.State != nil && *skill.State != "deleted" {
				resources[*skill.Id] = &ResourceMeta{Name: *skill.Name}
			}
		}
	}

	return resources, nil
}

func routingSkillExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllRoutingSkills),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceRoutingSkill() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Skill",

		CreateContext: CreateWithPooledClient(createRoutingSkill),
		ReadContext:   ReadWithPooledClient(readRoutingSkill),
		DeleteContext: DeleteWithPooledClient(deleteRoutingSkill),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Skill name. Changing the name attribute will cause the skill object object to dropped and recreated with a new ID.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func createRoutingSkill(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
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
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Reading skill %s", d.Id())
	return WithRetriesForRead(ctx, d, func() *resource.RetryError {
		skill, resp, getErr := routingAPI.GetRoutingSkill(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read skill %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read skill %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceRoutingSkill())
		if skill.State != nil && *skill.State == "deleted" {
			d.SetId("")
			return nil
		}

		d.Set("name", *skill.Name)
		log.Printf("Read skill %s %s", d.Id(), *skill.Name)
		return cc.CheckState()
	})
}

func deleteRoutingSkill(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Deleting skill %s", name)
	_, err := routingAPI.DeleteRoutingSkill(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete skill %s: %s", name, err)
	}

	return WithRetries(ctx, 30*time.Second, func() *resource.RetryError {
		routingSkill, resp, err := routingAPI.GetRoutingSkill(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Routing skill deleted
				log.Printf("Deleted Routing skill %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting Routing skill %s: %s", d.Id(), err))
		}

		if routingSkill.State != nil && *routingSkill.State == "deleted" {
			// Routing skill deleted
			log.Printf("Deleted Routing skill %s", d.Id())
			return nil
		}

		return resource.RetryableError(fmt.Errorf("Routing skill %s still exists", d.Id()))
	})
}
