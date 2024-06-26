package webdeployments_configuration_utils

import (
	"context"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func buildCobrowseSettings(d *schema.ResourceData) *platformclientv2.Cobrowsesettings {
	value, ok := d.GetOk("cobrowse")
	if !ok {
		return nil
	}

	cfgs := value.([]interface{})
	if len(cfgs) < 1 {
		return nil
	}

	cfg := cfgs[0].(map[string]interface{})

	enabled, _ := cfg["enabled"].(bool)
	allowAgentControl, _ := cfg["allow_agent_control"].(bool)
	allowAgentNavigation, _ := cfg["allow_agent_navigation"].(bool)
	channels := lists.InterfaceListToStrings(cfg["channels"].([]interface{}))
	maskSelectors := lists.InterfaceListToStrings(cfg["mask_selectors"].([]interface{}))
	readonlySelectors := lists.InterfaceListToStrings(cfg["readonly_selectors"].([]interface{}))

	var pauseCriteria []platformclientv2.Pausecriteria
	if v, ok := cfg["pause_criteria"]; ok {
		for _, pc := range v.([]interface{}) {
			pcMap := pc.(map[string]interface{})
			urlFragment := pcMap["url_fragment"].(string)
			condition := pcMap["condition"].(string)
			pauseCriteria = append(pauseCriteria, platformclientv2.Pausecriteria{
				UrlFragment: &urlFragment,
				Condition:   &condition,
			})
		}
	}


	return &platformclientv2.Cobrowsesettings{
		Enabled:              &enabled,
		AllowAgentControl:    &allowAgentControl,
		AllowAgentNavigation: &allowAgentNavigation,
		Channels:             &channels,
		MaskSelectors:        &maskSelectors,
		ReadonlySelectors:    &readonlySelectors,
		PauseCriteria:		  &pauseCriteria,
	}
}

func buildLocalizedLabels(labels []interface{}) *[]platformclientv2.Localizedlabels {
	if len(labels) < 1 {
		return nil
	}

	results := make([]platformclientv2.Localizedlabels, len(labels))
	for i, value := range labels {
		if label, ok := value.(map[string]interface{}); ok {
			results[i] = platformclientv2.Localizedlabels{
				Key:   platformclientv2.String(label["key"].(string)),
				Value: platformclientv2.String(label["value"].(string)),
			}
		}
	}

	return &results
}

func buildCustomI18nLabels(d *schema.ResourceData) *[]platformclientv2.Customi18nlabels {
	value, ok := d.GetOk("custom_i18n_labels")
	if !ok {
		return nil
	}

	labels := value.([]interface{})
	if len(labels) < 1 {
		return nil
	}

	results := make([]platformclientv2.Customi18nlabels, len(labels))
	for i, value := range labels {
		if label, ok := value.(map[string]interface{}); ok {
			results[i] = platformclientv2.Customi18nlabels{
				Language:        platformclientv2.String(label["language"].(string)),
				LocalizedLabels: buildLocalizedLabels(label["localized_labels"].([]interface{})),
			}
		}
	}

	return &results
}

func buildPosition(d *schema.ResourceData) *platformclientv2.Positionsettings {
	value, ok := d.GetOk("position")
	if !ok {
		return nil
	}

	position := value.([]interface{})
	if len(position) < 1 || len(position) == 1 && position[0] == nil {
		return nil
	}

	cfg := position[0].(map[string]interface{})
	return &platformclientv2.Positionsettings{
		Alignment:   platformclientv2.String(cfg["alignment"].(string)),
		SideSpace:   platformclientv2.Int(cfg["side_space"].(int)),
		BottomSpace: platformclientv2.Int(cfg["bottom_space"].(int)),
	}
}

func buildAuthenticationSettings(d *schema.ResourceData) *platformclientv2.Authenticationsettings {
	value, ok := d.GetOk("authentication_settings")
	if !ok {
		return nil
	}

	settings := value.([]interface{})
	if len(settings) < 1 {
		return nil
	}

	cfg := settings[0].(map[string]interface{})
	return &platformclientv2.Authenticationsettings{
		Enabled:       platformclientv2.Bool(cfg["enabled"].(bool)),
		IntegrationId: platformclientv2.String(cfg["integration_id"].(string)),
	}
}

func CustomizeConfigurationDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if len(diff.GetChangedKeysPrefix("")) > 0 {
		// When any change is made to the configuration we automatically publish a new version, so mark the version as updated
		// so dependent deployments will update appropriately to reference the newest version
		_ = diff.SetNewComputed("version")
	}
	return nil
}

// FeatureNotImplemented checks the response object to find out if the request failed because a feature is not yet
// implemented in the org that it was ran against. If true, we can pass back the field name and give more context
// in the final error message.
func FeatureNotImplemented(response *platformclientv2.APIResponse) (bool, string) {
	if response.Error == nil || response.Error.Details == nil || len(response.Error.Details) == 0 {
		return false, ""
	}
	for _, err := range response.Error.Details {
		if err.FieldName == nil {
			continue
		}
		if strings.Contains(*err.ErrorCode, "feature is not yet implemented") {
			return true, *err.FieldName
		}
	}
	return false, ""
}

func ValidateConfigurationStatusChange(k, old, new string, d *schema.ResourceData) bool {
	// Configs start in a pending status and may not transition to active or error before we retrieve the state, so allow
	// the status to change from pending to something less ephemeral
	return old == "Pending"
}

func FlattenCobrowseSettings(cobrowseSettings *platformclientv2.Cobrowsesettings) []interface{} {
	if cobrowseSettings == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"enabled":                cobrowseSettings.Enabled,
		"allow_agent_control":    cobrowseSettings.AllowAgentControl,
		"allow_agent_navigation": cobrowseSettings.AllowAgentNavigation,
		"channels":               cobrowseSettings.Channels,
		"mask_selectors":         cobrowseSettings.MaskSelectors,
		"readonly_selectors":     cobrowseSettings.ReadonlySelectors,
	}}
}

func flattenLocalizedLabels(localizedLabels *[]platformclientv2.Localizedlabels) []interface{} {
	if localizedLabels == nil {
		return nil
	}

	results := make([]interface{}, len(*localizedLabels))
	for i, label := range *localizedLabels {
		results[i] = map[string]interface{}{
			"key":   *label.Key,
			"value": *label.Value,
		}
	}

	return results
}

func FlattenCustomI18nLabels(customI18nLabels *[]platformclientv2.Customi18nlabels) []interface{} {
	if customI18nLabels == nil {
		return nil
	}

	results := make([]interface{}, len(*customI18nLabels))
	for i, label := range *customI18nLabels {
		results[i] = map[string]interface{}{
			"language":         label.Language,
			"localized_labels": flattenLocalizedLabels(label.LocalizedLabels),
		}
	}

	return results
}

func FlattenPosition(position *platformclientv2.Positionsettings) []interface{} {
	if position == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"alignment":    position.Alignment,
		"side_space":   position.SideSpace,
		"bottom_space": position.BottomSpace,
	}}
}

func FlattenAuthenticationSettings(authenticationSettings *platformclientv2.Authenticationsettings) []interface{} {
	if authenticationSettings == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"enabled":        authenticationSettings.Enabled,
		"integration_id": authenticationSettings.IntegrationId,
	}}
}

func BuildWebDeploymentConfigurationFromResourceData(d *schema.ResourceData) (string, *platformclientv2.Webdeploymentconfigurationversion) {
	name := d.Get("name").(string)
	languages := lists.InterfaceListToStrings(d.Get("languages").([]interface{}))

	inputCfg := &platformclientv2.Webdeploymentconfigurationversion{
		Name:            &name,
		Description:     resourcedata.GetNillableValue[string](d, "description"),
		Languages:       &languages,
		DefaultLanguage: resourcedata.GetNillableValue[string](d, "default_language"),
	}

	if headlessMode, ok := d.GetOk("headless_mode_enabled"); ok {
		inputCfg.HeadlessMode = &platformclientv2.Webdeploymentheadlessmode{
			Enabled: platformclientv2.Bool(headlessMode.(bool)),
		}
	}

	customI18nLabels := buildCustomI18nLabels(d)
	if customI18nLabels != nil {
		inputCfg.CustomI18nLabels = customI18nLabels
	}

	position := buildPosition(d)
	if position != nil {
		inputCfg.Position = position
	}

	messengerSettings := buildMessengerSettings(d)
	if messengerSettings != nil {
		inputCfg.Messenger = messengerSettings
	}

	cobrowseSettings := buildCobrowseSettings(d)
	if cobrowseSettings != nil {
		inputCfg.Cobrowse = cobrowseSettings
	}

	journeySettings := buildJourneySettings(d)
	if journeySettings != nil {
		inputCfg.JourneyEvents = journeySettings
	}

	supportCenterSettings := buildSupportCenterSettings(d)
	if supportCenterSettings != nil {
		inputCfg.SupportCenter = supportCenterSettings
	}

	authenticationSettings := buildAuthenticationSettings(d)
	if authenticationSettings != nil {
		inputCfg.AuthenticationSettings = authenticationSettings
	}

	return name, inputCfg
}
