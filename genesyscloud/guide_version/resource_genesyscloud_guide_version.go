package guide_version

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"log"
)

func createGuideVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	skdConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideVersionProxy(skdConfig)
	guideId := d.Get("guide_id").(string)

	log.Printf("Creating Guide Version")

	versionReq := buildGuideVersionFromResourceData(d)

	version, resp, err := proxy.createGuideVersion(ctx, versionReq, guideId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create guide version | error: %s", err), resp)
	}

	d.SetId(*version.Id)
	log.Printf("Created Guide Version")
	return readGuideVersion(ctx, d, meta)
}

func readGuideVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	skdConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideVersionProxy(skdConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceGuideVersion(), constants.ConsistencyChecks(), ResourceType)
	guideId := d.Get("guide_id").(string)

	log.Printf("Reading Guide Version")

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		version, resp, err := proxy.getGuideVersionById(ctx, d.Id(), guideId)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read guide version %s | Error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read guide version %s | Error: %s", d.Id(), err), resp))
		}

		_ = d.Set("instruction", version.Instruction)
		_ = d.Set("resource_data_action", version.Resources)
		_ = d.Set("variables", version.Variables)

		log.Printf("Read Guide Version")
		return cc.CheckState(d)
	})
}

func updateGuideVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	skdConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideVersionProxy(skdConfig)
	guideId := d.Get("guide_id").(string)

	log.Printf("Updating Guide Version %s", d.Id())

	versionReq := buildGuideVersionForUpdate(d)

	_, resp, err := proxy.updateGuideVersion(ctx, d.Id(), guideId, versionReq)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update guide version | error: %s", err), resp)
	}

	log.Printf("Updated Guide Version")
	return readGuideVersion(ctx, d, meta)
}

// Build Function

func buildGuideVersionFromResourceData(d *schema.ResourceData) *CreateGuideVersionRequest {
	log.Printf("Building Guide Version from Resource Data")

	guideVersion := &CreateGuideVersionRequest{
		GuideID:     d.Get("guide_id").(string),
		Instruction: d.Get("instruction").(string),
	}

	if vars := d.Get("variables").([]interface{}); vars != nil {
		guideVersion.Variables = buildGuideVersionVariables(vars)
	}

	if resource := d.Get("resource_data_action").([]interface{}); resource != nil {
		guideVersion.Resources = buildGuideVersionResources(resource)
	}

	log.Printf("Succesfully Built Guide Version from Resource Data")
	return guideVersion
}

func buildGuideVersionResources(resource []interface{}) GuideVersionResources {
	var versionResource = GuideVersionResources{}

	return versionResource
}

func buildGuideVersionVariables(vars []interface{}) []Variable {
	variables := make([]Variable, len(vars))

	for i, v := range vars {
		variables[i] = Variable{
			Name:  v.(map[string]interface{})["name"].(string),
			Type:  v.(map[string]interface{})["value"].(string),
			Scope: v.(map[string]interface{})["scope"].(string),
		}

		if description := v.(map[string]interface{})["description"].(string); description != "" {
			variables[i].Description = description
		}
	}

	return variables
}

func buildGuideVersionForUpdate(d *schema.ResourceData) *UpdateGuideVersion {
	log.Printf("Building Guide Version from Resource Data")

	guideVersion := &UpdateGuideVersion{
		GuideID:     d.Get("guide_id").(string),
		Instruction: d.Get("instruction").(string),
	}

	if vars := d.Get("variables").([]interface{}); vars != nil {
		guideVersion.Variables = buildGuideVersionVariables(vars)
	}

	if resource := d.Get("resource_data_action").([]interface{}); resource != nil {
		guideVersion.Resources = buildGuideVersionResources(resource)
	}

	log.Printf("Succesfully Built Guide Version from Resource Data")
	return guideVersion
}
