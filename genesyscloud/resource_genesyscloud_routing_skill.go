package genesyscloud

import (
	"context"
	"log"

	"github.com/MyPureCloud/platform-client-sdk-go/platformclientv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRoutingSkill() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Skill",

		CreateContext: createRoutingSkill,
		ReadContext:   readRoutingSkill,
		DeleteContext: deleteRoutingSkill,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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

	routingAPI := platformclientv2.NewRoutingApi()

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
	routingAPI := platformclientv2.NewRoutingApi()

	skill, _, getErr := routingAPI.GetRoutingSkill(d.Id())
	if getErr != nil {
		return diag.Errorf("Failed to read skill %s: %s", d.Id(), getErr)
	}

    if skill.State != nil && *skill.State == "deleted" {
        return diag.Errorf("Skill %s deleted", d.Id())
    }

	d.Set("name", *skill.Name)
	return nil
}

func deleteRoutingSkill(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	routingAPI := platformclientv2.NewRoutingApi()

	log.Printf("Deleting skill %s", name)
	_, err := routingAPI.DeleteRoutingSkill(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete skill %s: %s", name, err)
	}
	return nil
}
