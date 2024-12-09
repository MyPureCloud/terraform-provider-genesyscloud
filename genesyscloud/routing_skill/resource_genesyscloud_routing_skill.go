package routing_skill

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

func GetAllRoutingSkills(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getRoutingSkillProxy(clientConfig)

	skills, resp, getErr := proxy.getAllRoutingSkills(ctx, "")
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get all routing skills | error: %s", getErr), resp)
	}

	for _, skill := range *skills {
		if skill.State != nil && *skill.State != "deleted" {
			resources[*skill.Id] = &resourceExporter.ResourceMeta{BlockLabel: *skill.Name}
		}
	}

	return resources, nil
}

func createRoutingSkill(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSkillProxy(sdkConfig)
	name := d.Get("name").(string)

	log.Printf("Creating skill %s", name)
	skill, resp, err := proxy.createRoutingSkill(ctx, &platformclientv2.Routingskill{
		Name: &name,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create skill %s error: %s", name, err), resp)
	}

	d.SetId(*skill.Id)
	log.Printf("Created skill %s %s", name, *skill.Id)
	return readRoutingSkill(ctx, d, meta)
}

func readRoutingSkill(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSkillProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingSkill(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading skill %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		skill, resp, getErr := proxy.getRoutingSkillById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read skill %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read skill %s | error: %s", d.Id(), getErr), resp))
		}

		if skill.State != nil && *skill.State == "deleted" {
			d.SetId("")
			return nil
		}

		_ = d.Set("name", *skill.Name)
		log.Printf("Read skill %s %s", d.Id(), *skill.Name)
		return cc.CheckState(d)
	})
}

func deleteRoutingSkill(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSkillProxy(sdkConfig)
	name := d.Get("name").(string)

	log.Printf("Deleting Routing skill %s", name)
	resp, err := proxy.deleteRoutingSkill(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete skill %s error: %s", name, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		routingSkill, resp, err := proxy.getRoutingSkillById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted Routing skill %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting Routing skill %s | error: %s", d.Id(), err), resp))
		}

		if routingSkill.State != nil && *routingSkill.State == "deleted" {
			log.Printf("Deleted Routing skill %s", d.Id())
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Routing skill %s still exists", d.Id()), resp))
	})
}

func GenerateRoutingSkillResource(
	resourceLabel string,
	name string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_skill" "%s" {
		name = "%s"
	}
	`, resourceLabel, name)
}
