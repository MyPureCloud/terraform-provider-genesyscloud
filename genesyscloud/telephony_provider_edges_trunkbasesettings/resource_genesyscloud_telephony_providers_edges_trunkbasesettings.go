package telephony_provider_edges_trunkbasesettings

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
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

func createTrunkBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	trunkMetaBaseString := d.Get("trunk_meta_base_id").(string)
	trunkMetaBase := util.BuildSdkDomainEntityRef(d, "trunk_meta_base_id")
	inboundSiteString := d.Get("inbound_site_id").(string)
	siteString := d.Get("site_id").(string)
	properties := util.BuildTelephonyProperties(d)
	trunkType := d.Get("trunk_type").(string)
	managed := d.Get("managed").(bool)
	trunkBase := platformclientv2.Trunkbase{
		Name:          &name,
		TrunkMetabase: trunkMetaBase,
		TrunkType:     &trunkType,
		Managed:       &managed,
		Properties:    properties,
	}

	validationInboundSite, errorInboundSite := ValidateInboundSiteSettings(inboundSiteString, trunkMetaBaseString)

	if validationInboundSite && errorInboundSite == nil {
		inboundSite := util.BuildSdkDomainEntityRef(d, "inbound_site_id")
		trunkBase.InboundSite = inboundSite
	}

	if errorInboundSite != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to create trunk base settings %s for inboundSiteId", name), errorInboundSite)
	}

	if siteString != "" {
		trunkBase.Site = util.BuildSdkDomainEntityRef(d, "site_id")
	}

	if description != "" {
		trunkBase.Description = &description
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTrunkBaseSettingProxy(sdkConfig)

	log.Printf("Creating trunk base settings %s", name)
	trunkBaseSettings, resp, err := proxy.CreateTrunkBaseSetting(ctx, trunkBase)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create trunk base settings %s error: %s", name, err), resp)
	}
	if trunkBaseSettings != nil && trunkBaseSettings.Id != nil {
		d.SetId(*trunkBaseSettings.Id)
	} else {
		log.Printf("Error: trunkBaseSettings or its Id is nil\n")
	}

	return readTrunkBaseSettings(ctx, d, meta)
}

func updateTrunkBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	trunkMetaBaseString := d.Get("trunk_meta_base_id").(string)
	trunkMetaBase := util.BuildSdkDomainEntityRef(d, "trunk_meta_base_id")
	inboundSiteString := d.Get("inbound_site_id").(string)
	siteString := d.Get("site_id").(string)

	properties := util.BuildTelephonyProperties(d)
	trunkType := d.Get("trunk_type").(string)
	managed := d.Get("managed").(bool)
	id := d.Id()

	trunkBase := platformclientv2.Trunkbase{
		Id:            &id,
		Name:          &name,
		TrunkMetabase: trunkMetaBase,
		TrunkType:     &trunkType,
		Managed:       &managed,
		Properties:    properties,
	}

	validationInboundSite, errorInboundSite := ValidateInboundSiteSettings(inboundSiteString, trunkMetaBaseString)

	if validationInboundSite && errorInboundSite == nil {
		inboundSite := util.BuildSdkDomainEntityRef(d, "inbound_site_id")
		trunkBase.InboundSite = inboundSite
	}
	if errorInboundSite != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to update trunk base settings %s for inboundSite", name), errorInboundSite)
	}

	if siteString != "" {
		trunkBase.Site = util.BuildSdkDomainEntityRef(d, "site_id")
	}

	if description != "" {
		trunkBase.Description = &description
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTrunkBaseSettingProxy(sdkConfig)

	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get the latest version of the setting
		trunkBaseSettings, resp, getErr := proxy.GetTrunkBaseSettingById(ctx, id)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("The trunk base settings does not exist %s error: %s", d.Id(), getErr), resp)
			}
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read trunk base settings %s error: %s", d.Id(), getErr), resp)
		}
		trunkBase.Version = trunkBaseSettings.Version

		log.Printf("Updating trunk base settings %s", name)
		trunkBaseSettings, resp, err := proxy.UpdateTrunkBaseSetting(ctx, d.Id(), trunkBase)
		if err != nil {

			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update trunk base settings %s error: %s", name, err), resp)
		}
		return resp, nil
	})

	if diagErr != nil {
		return diagErr
	}

	// Get the latest version of the setting
	trunkBaseSettings, resp, getErr := proxy.GetTrunkBaseSettingById(ctx, d.Id())
	if getErr != nil {
		if util.IsStatus404(resp) {
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read trunk base settings %s error: %s", d.Id(), getErr), resp)
	}
	trunkBase.Version = trunkBaseSettings.Version

	log.Printf("Updating trunk base settings %s", name)
	trunkBaseSettings, resp, err := proxy.UpdateTrunkBaseSetting(ctx, d.Id(), trunkBase)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update trunk base settings %s error: %s", d.Id(), err), resp)
	}

	if trunkBaseSettings != nil && trunkBaseSettings.Id != nil {
		log.Printf("Updated trunk base settings %s", *trunkBaseSettings.Id)
	} else {
		log.Printf("Error: trunkBaseSettings or its Id is nil")
	}

	return readTrunkBaseSettings(ctx, d, meta)
}

func readTrunkBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTrunkBaseSettingProxy(sdkConfig)

	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTrunkBaseSettings(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading trunk base settings %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		trunkBaseSettings, resp, getErr := proxy.GetTrunkBaseSettingById(ctx, d.Id())

		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read trunk base settings %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read trunk base settings %s | error: %s", d.Id(), getErr), resp))
		}

		if trunkBaseSettings != nil && trunkBaseSettings.Name != nil {
			d.Set("name", *trunkBaseSettings.Name)
		}

		if trunkBaseSettings != nil && trunkBaseSettings.State != nil {
			d.Set("state", *trunkBaseSettings.State)
		}

		if trunkBaseSettings.Description != nil {
			d.Set("description", *trunkBaseSettings.Description)
		}
		if trunkBaseSettings.Managed != nil {
			d.Set("managed", *trunkBaseSettings.Managed)
		}

		// check if Id is null or not for both metabase and inboundsite
		if trunkBaseSettings != nil && trunkBaseSettings.TrunkMetabase != nil && trunkBaseSettings.TrunkMetabase.Id != nil {
			d.Set("trunk_meta_base_id", *trunkBaseSettings.TrunkMetabase.Id)
		}

		// check if Id is null or not for both metabase and inboundsite
		if trunkBaseSettings != nil && trunkBaseSettings.InboundSite != nil {
			d.Set("inbound_site_id", *trunkBaseSettings.InboundSite.Id)
		}

		if trunkBaseSettings != nil && trunkBaseSettings.Site != nil {
			d.Set("site_id", *trunkBaseSettings.Site.Id)
		}

		if trunkBaseSettings != nil && trunkBaseSettings.TrunkType != nil {
			d.Set("trunk_type", *trunkBaseSettings.TrunkType)
		}
		d.Set("properties", nil)
		if trunkBaseSettings.Properties != nil {
			properties, err := util.FlattenTelephonyProperties(trunkBaseSettings.Properties)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			d.Set("properties", properties)
		}

		return cc.CheckState(d)
	})
}

func deleteTrunkBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTrunkBaseSettingProxy(sdkConfig)
	log.Printf("Deleting trunk base settings for id %s\n", d.Id())
	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {

		resp, err := proxy.DeleteTrunkBaseSetting(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// trunk base settings not found, goal achieved!
				return nil, nil
			}
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete trunk base settings %s error: %s", d.Id(), err), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		trunkBaseSettings, resp, err := proxy.GetTrunkBaseSettingById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// trunk base settings deleted
				log.Printf("Deleted trunk base settings %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting trunk base settings %s | error: %s", d.Id(), err), resp))
		}

		if trunkBaseSettings.State != nil && *trunkBaseSettings.State == "deleted" {
			// trunk base settings deleted
			log.Printf("Deleted trunk base settings %s", d.Id())
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("trunk base settings %s still exists", d.Id()), resp))
	})
}

func getAllTrunkBaseSettings(ctx context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getTrunkBaseSettingProxy(sdkConfig)
	trunkBaseSettings, resp, getErr := proxy.GetAllTrunkBaseSetting(ctx)

	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get all trunk base settings error: %s", getErr), resp)
	}

	for _, tbs := range *trunkBaseSettings {
		resources[*tbs.Id] = &resourceExporter.ResourceMeta{BlockLabel: *tbs.Name}
	}

	return resources, nil
}

func ValidateInboundSiteSettings(inboundSiteString string, trunkBaseMetaId string) (bool, error) {
	externalTrunkName := "external_sip_pcv_byoc"

	if len(inboundSiteString) == 0 && strings.Contains(trunkBaseMetaId, externalTrunkName) {
		return false, errors.New("inboundSite is required for external BYOC trunks")
	}
	if len(inboundSiteString) > 0 && !strings.Contains(trunkBaseMetaId, externalTrunkName) {
		return false, errors.New("inboundSite should be set for external BYOC trunks only")
	}
	if len(inboundSiteString) > 0 && strings.Contains(trunkBaseMetaId, externalTrunkName) {
		return true, nil
	}
	return false, nil
}

func GenerateTrunkBaseSettingsResourceWithCustomAttrs(
	trunkBaseSettingsResourceLabel,
	name,
	description,
	trunkMetaBaseId,
	trunkType string,
	managed bool,
	otherAttrs ...string) string {
	resource := fmt.Sprintf(`resource "genesyscloud_telephony_providers_edges_trunkbasesettings" "%s" {
		name = "%s"
		description = "%s"
		trunk_meta_base_id = "%s"
		trunk_type = "%s"
		managed = %v
		%s
	}
	`, trunkBaseSettingsResourceLabel, name, description, trunkMetaBaseId, trunkType, managed, strings.Join(otherAttrs, "\n"))
	return resource
}
