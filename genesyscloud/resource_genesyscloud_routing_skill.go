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

func getAllRoutingSkills(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		skills, resp, getErr := routingAPI.GetRoutingSkills(pageSize, pageNum, "", nil)
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_routing_skill", fmt.Sprintf("Failed to get skills error: %s", getErr), resp)
		}

		if skills.Entities == nil || len(*skills.Entities) == 0 {
			break
		}

		for _, skill := range *skills.Entities {
			if skill.State != nil && *skill.State != "deleted" {
				resources[*skill.Id] = &resourceExporter.ResourceMeta{Name: *skill.Name}
			}
		}
	}

	return resources, nil
}

func RoutingSkillExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRoutingSkills),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceRoutingSkill() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Skill",

		CreateContext: provider.CreateWithPooledClient(createRoutingSkill),
		ReadContext:   provider.ReadWithPooledClient(readRoutingSkill),
		DeleteContext: provider.DeleteWithPooledClient(deleteRoutingSkill),
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

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Creating skill %s", name)
	skill, resp, err := routingAPI.PostRoutingSkills(platformclientv2.Routingskill{
		Name: &name,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_routing_skill", fmt.Sprintf("Failed to create skill %s error: %s", name, err), resp)
	}

	d.SetId(*skill.Id)

	log.Printf("Created skill %s %s", name, *skill.Id)
	return readRoutingSkill(ctx, d, meta)
}

func readRoutingSkill(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingSkill(), constants.DefaultConsistencyChecks, "genesyscloud_routing_skill")

	log.Printf("Reading skill %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		skill, resp, getErr := routingAPI.GetRoutingSkill(d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_skill", fmt.Sprintf("Failed to read skill %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_skill", fmt.Sprintf("Failed to read skill %s | error: %s", d.Id(), getErr), resp))
		}

		if skill.State != nil && *skill.State == "deleted" {
			d.SetId("")
			return nil
		}

		d.Set("name", *skill.Name)
		log.Printf("Read skill %s %s", d.Id(), *skill.Name)
		return cc.CheckState(d)
	})
}

func deleteRoutingSkill(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Deleting skill %s", name)
	resp, err := routingAPI.DeleteRoutingSkill(d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_routing_skill", fmt.Sprintf("Failed to delete skill %s error: %s", name, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		routingSkill, resp, err := routingAPI.GetRoutingSkill(d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Routing skill deleted
				log.Printf("Deleted Routing skill %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_skill", fmt.Sprintf("Error deleting Routing skill %s | error: %s", d.Id(), err), resp))
		}

		if routingSkill.State != nil && *routingSkill.State == "deleted" {
			// Routing skill deleted
			log.Printf("Deleted Routing skill %s", d.Id())
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_skill", fmt.Sprintf("Routing skill %s still exists", d.Id()), resp))
	})
}

func GenerateRoutingSkillResource(
	resourceID string,
	name string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_skill" "%s" {
		name = "%s"
	}
	`, resourceID, name)
}
