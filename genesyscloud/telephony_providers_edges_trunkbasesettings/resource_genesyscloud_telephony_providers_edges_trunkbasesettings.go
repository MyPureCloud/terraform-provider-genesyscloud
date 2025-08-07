package telephony_providers_edges_trunkbasesettings

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
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
		_, resp, err := proxy.UpdateTrunkBaseSetting(ctx, d.Id(), trunkBase)
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
	var response *platformclientv2.APIResponse

	log.Printf("Reading trunk base settings %s", d.Id())
	readErr := util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		trunkBaseSettings, resp, getErr := proxy.GetTrunkBaseSettingById(ctx, d.Id())
		response = resp

		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(getErr)
			}
			return retry.NonRetryableError(getErr)
		}

		if trunkBaseSettings == nil {
			return retry.NonRetryableError(fmt.Errorf("succesfully read trunkbase setting '%s', but response body was nil", d.Id()))
		}

		resourcedata.SetNillableValue(d, "name", trunkBaseSettings.Name)
		resourcedata.SetNillableValue(d, "state", trunkBaseSettings.State)
		resourcedata.SetNillableValue(d, "description", trunkBaseSettings.Description)
		resourcedata.SetNillableValue(d, "managed", trunkBaseSettings.Managed)
		resourcedata.SetNillableReference(d, "trunk_meta_base_id", trunkBaseSettings.TrunkMetabase)
		resourcedata.SetNillableReference(d, "inbound_site_id", trunkBaseSettings.InboundSite)
		resourcedata.SetNillableReference(d, "site_id", trunkBaseSettings.Site)
		resourcedata.SetNillableValue(d, "trunk_type", trunkBaseSettings.TrunkType)

		_ = d.Set("properties", nil)
		if trunkBaseSettings.Properties != nil {
			properties, err := util.FlattenTelephonyProperties(trunkBaseSettings.Properties)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			_ = d.Set("properties", properties)
		}

		return cc.CheckState(d)
	})

	if readErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to read trunkbase setting '%s' | error: %v", d.Id(), readErr), response)
	}

	return nil
}

func deleteTrunkBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var (
		err  error
		resp *platformclientv2.APIResponse
	)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTrunkBaseSettingProxy(sdkConfig)

	log.Printf("Deleting trunk base settings for id %s\n", d.Id())
	deleteWithRetriesErr := util.WithRetries(ctx, 50*time.Second, func() *retry.RetryError {
		resp, err = proxy.DeleteTrunkBaseSetting(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// trunk base settings not found, goal achieved!
				return nil
			}
			if util.IsStatus400(resp) {
				return retry.RetryableError(fmt.Errorf("failed to delete trunkbase setting %s due to 400 error: %w", d.Id(), err))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to delete trunkbase setting %s: %w", d.Id(), err))
		}
		return nil
	})
	if deleteWithRetriesErr.HasError() {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete trunk base settings %s error: %v", d.Id(), deleteWithRetriesErr), resp)
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
		// Don't export autogenerated edge trunks
		if *tbs.TrunkType == "EDGE" {
			continue
		}
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

func shouldExportTrunkBaseSettingsAsDataSource(ctx context.Context, sdkConfig *platformclientv2.Configuration, configMap map[string]string) (exportAsData bool, err error) {
	defaultTbsNames := []string{
		"Cloud Proxy Tie TrunkBase for EdgeGroup",
		"Direct Tie TrunkBase for EdgeGroup",
		"Genesys Cloud - CDM SIP Phone Trunk",
		"Genesys Cloud - CDM WebRTC Phone Trunk",
		"Indirect Tie TrunkBase for EdgeGroup",
		"PureCloud Voice - AWS",
		"Tie TrunkBase for EdgeGroup",
	}
	tbsName, ok := configMap["name"]
	if ok {
		if lists.ContainsAnySubStringSlice(tbsName, defaultTbsNames) {
			return true, nil
		}
	}
	return false, nil
}
