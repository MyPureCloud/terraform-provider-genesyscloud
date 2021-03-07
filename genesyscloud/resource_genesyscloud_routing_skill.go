package genesyscloud

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/platformclientv2"
)

func getAllRoutingSkills() (ResourceIDNameMap, diag.Diagnostics) {
	resources := make(map[string]string)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(GetSdkClient())

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
				resources[*skill.Id] = *skill.Name
			}
		}
	}

	return resources, nil
}

func routingSkillExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllRoutingSkills,
		ResourceDef:      resourceRoutingSkill(),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceRoutingSkill() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Skill",

		CreateContext: createRoutingSkill,
		ReadContext:   readRoutingSkill,
		DeleteContext: deleteRoutingSkill,
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

	routingAPI := platformclientv2.NewRoutingApiWithConfig(GetSdkClient())

	log.Printf("Creating skill %s", name)
	skill, _, err := routingAPI.PostRoutingSkills(platformclientv2.Routingskill{
		Name: &name,
	})
	if err != nil {
		return diag.Errorf("Failed to create skill %s: %s", name, err)
	}

	d.SetId(*skill.Id)

	return readRoutingSkill(ctx, d, meta)
}

func readRoutingSkill(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	routingAPI := platformclientv2.NewRoutingApiWithConfig(GetSdkClient())

	skill, resp, getErr := routingAPI.GetRoutingSkill(d.Id())
	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read skill %s: %s", d.Id(), getErr)
	}

	if skill.State != nil && *skill.State == "deleted" {
		d.SetId("")
		return nil
	}

	d.Set("name", *skill.Name)
	return nil
}

func deleteRoutingSkill(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	routingAPI := platformclientv2.NewRoutingApiWithConfig(GetSdkClient())

	log.Printf("Deleting skill %s", name)
	_, err := routingAPI.DeleteRoutingSkill(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete skill %s: %s", name, err)
	}
	return nil
}
