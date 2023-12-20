package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v116/platformclientv2"
)

var (
	messengerStyle = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"primary_color": {
				Description: "The primary color of messenger in hexadecimal",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	launcherButtonSettings = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"visibility": {
				Description: "The visibility settings for the button.Valid values: On, Off, OnDemand",
				Type:        schema.TypeString,
				Optional:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"On",
					"Off",
					"OnDemand",
				}, false),
			},
		},
	}

	homeScreen = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"enabled": {
				Description: "Whether or not home screen is enabled",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"logo_url": {
				Description: "URL for custom logo to appear in home screen",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}

	fileUploadMode = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"file_types": {
				Description: "A list of supported content types for uploading files.Valid values: image/jpeg, image/gif, image/png",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"max_file_size_kb": {
				Description:  "The maximum file size for file uploads in kilobytes. Default is 10240 (10 MB)",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 10240),
			},
		},
	}

	fileUploadSettings = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"mode": {
				Description: "The list of supported file upload modes",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        fileUploadMode,
			},
		},
	}

	messengerSettings = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"enabled": {
				Description: "Whether or not messenger is enabled",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"styles": {
				Description: "The style settings for messenger",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        messengerStyle,
			},
			"launcher_button": {
				Description: "The settings for the launcher button",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        launcherButtonSettings,
			},
			"home_screen": {
				Description: "The settings for the home screen",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        homeScreen,
			},
			"file_upload": {
				Description: "File upload settings for messenger",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        fileUploadSettings,
			},
		},
	}

	cobrowseSettings = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"enabled": {
				Description: "Whether or not cobrowse is enabled",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"allow_agent_control": {
				Description: "Whether agent can take control over customer's screen or not",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"channels": {
				Description: "List of channels through which cobrowse is available (for now only Webmessaging and Voice)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"Webmessaging", "Voice"}, false),
				},
			},
			"mask_selectors": {
				Description: "List of CSS selectors which should be masked when screen sharing is active",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"readonly_selectors": {
				Description: "List of CSS selectors which should be read-only when screen sharing is active",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	selectorEventTrigger = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"selector": {
				Description: "Element that triggers event",
				Type:        schema.TypeString,
				Required:    true,
			},
			"event_name": {
				Description: "Name of event triggered when element matching selector is interacted with",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}

	formsTrackTrigger = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"selector": {
				Description: "Form element that triggers the form submitted or abandoned event",
				Type:        schema.TypeString,
				Required:    true,
			},
			"form_name": {
				Description: "Prefix for the form submitted or abandoned event name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"capture_data_on_form_abandon": {
				Description: "Whether to capture the form data in the form abandoned event",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"capture_data_on_form_submit": {
				Description: "Whether to capture the form data in the form submitted event",
				Type:        schema.TypeBool,
				Required:    true,
			},
		},
	}

	idleEventTrigger = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"event_name": {
				Description: "Name of event triggered after period of inactivity",
				Type:        schema.TypeString,
				Required:    true,
			},
			"idle_after_seconds": {
				Description:  "Number of seconds of inactivity before an event is triggered",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(30),
			},
		},
	}

	scrollPercentageEventTrigger = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"event_name": {
				Description: "Name of event triggered after scrolling to the specified percentage",
				Type:        schema.TypeString,
				Required:    true,
			},
			"percentage": {
				Description:  "Percentage of a webpage at which an event is triggered",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 100),
			},
		},
	}

	journeyEventsSettings = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"enabled": {
				Description: "Whether or not journey event collection is enabled",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"excluded_query_parameters": {
				Description: "List of parameters to be excluded from the query string",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"should_keep_url_fragment": {
				Description: "Whether or not to keep the URL fragment",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"search_query_parameters": {
				Description: "List of query parameters used for search (e.g. 'q')",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"pageview_config": {
				Description: "Controls how the pageview events are tracked.Valid values: Auto, Once, Off",
				Type:        schema.TypeString,
				Optional:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"Auto",
					"Once",
					"Off",
				}, false),
			},
			"click_event": {
				Description: "Details about a selector event trigger",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        selectorEventTrigger,
			},
			"form_track_event": {
				Description: "Details about a forms tracking event trigger",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        formsTrackTrigger,
			},
			"idle_event": {
				Description: "Details about an idle event trigger",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        idleEventTrigger,
			},
			"in_viewport_event": {
				Description: "Details about a selector event trigger",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        selectorEventTrigger,
			},
			"scroll_depth_event": {
				Description: "Details about a scroll percentage event trigger",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        scrollPercentageEventTrigger,
			},
		},
	}

	supportCenterCustomMessage = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"default_value": {
				Description: "Custom message default value.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"type": {
				Description: "Custom message type.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	supportCenterCategory = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"enabled_categories_id": {
				Description: "Enabled categories id.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"self_uri": {
				Description: "Enabled categories URI",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"image_source": {
				Description: "Enabled categories image source",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	supportCenterStyleSetting = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"hero_style_background_color": {
				Description: "Hero style background color.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"hero_style_text_color": {
				Description: "Hero style background color",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"hero_style_image": {
				Description: "Hero style image",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"global_style_background_color": {
				Description: "Global style background color",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"global_style_primary_color": {
				Description: "Global style primary color",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"global_style_primary_color_dark": {
				Description: "Global style primary color dark",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"global_style_primary_color_light": {
				Description: "Global style primary color light",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"global_style_text_color": {
				Description: "Global style text color",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"global_style_font_family": {
				Description: "Global style font family",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	supportCenterScreens = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description: "The type of the screen.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"module_settings_type": {
				Description: "Screen module type.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"module_settings_enabled": {
				Description: "Whether or not support center screen module is enabled",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"module_settings_compact_category_module_template": {
				Description: "Compact category module template",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"module_settings_detailed_category_module_template": {
				Description: "Customer feedback settings.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
	supportCenterSettings = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"enabled": {
				Description: "Whether or not  knowledge base support center is enabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"knowledge_base_id": {
				Description: "The knowledge base for support center.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"knowledge_base_uri": {
				Description: "The knowledge base uri for support center.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"router_type": {
				Description: "Router type for support center.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"custom_messages": {
				Description: "Customizable display texts for support center tracking event trigger.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        supportCenterCustomMessage,
			},
			"feedback_enabled": {
				Description: "Customer feedback settings.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"enabled_categories": {
				Description: "Enabled article categories for support center.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        supportCenterCategory,
			},
			"style_setting": {
				Description: "Style attributes for support center.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        supportCenterStyleSetting,
			},
			"screens": {
				Description: "Settings concerning support center",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        supportCenterScreens,
			},
		},
	}
)

func getAllWebDeploymentConfigurations(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	webDeploymentsAPI := platformclientv2.NewWebDeploymentsApiWithConfig(clientConfig)

	configurations, _, getErr := webDeploymentsAPI.GetWebdeploymentsConfigurations(false)
	if getErr != nil {
		return nil, diag.Errorf("Failed to get web deployment configurations: %v", getErr)
	}

	for _, configuration := range *configurations.Entities {
		resources[*configuration.Id] = &resourceExporter.ResourceMeta{Name: *configuration.Name}
	}

	return resources, nil
}

func WebDeploymentConfigurationExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc:   GetAllWithPooledClient(getAllWebDeploymentConfigurations),
		ExcludedAttributes: []string{"version"},
	}
}

func ResourceWebDeploymentConfiguration() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Web Deployment Configuration",

		CreateContext: CreateWithPooledClient(createWebDeploymentConfiguration),
		ReadContext:   ReadWithPooledClient(readWebDeploymentConfiguration),
		UpdateContext: UpdateWithPooledClient(updateWebDeploymentConfiguration),
		DeleteContext: DeleteWithPooledClient(deleteWebDeploymentConfiguration),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "Deployment name",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"description": {
				Description: "Deployment description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"languages": {
				Description: "A list of languages supported on the configuration.",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"default_language": {
				Description: "The default language to use for the configuration.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"status": {
				Description: "The current status of the deployment. Valid values: Pending, Active, Inactive, Error, Deleting.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"Pending",
					"Active",
					"Inactive",
					"Error",
					"Deleting",
				}, false),
				DiffSuppressFunc: validateConfigurationStatusChange,
			},
			"version": {
				Description: "The version of the configuration.",
				Type:        schema.TypeString,
				Computed:    true,
				MaxItems:    0,
			},
			"messenger": {
				Description: "Settings concerning messenger",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        messengerSettings,
			},
			"cobrowse": {
				Description: "Settings concerning cobrowse",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        cobrowseSettings,
			},
			"journey_events": {
				Description: "Settings concerning journey events",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        journeyEventsSettings,
			},
			"support_center": {
				Description: "Settings concerning support center",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        supportCenterSettings,
			},
		},
		CustomizeDiff: customizeConfigurationDiff,
	}
}

func customizeConfigurationDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if len(diff.GetChangedKeysPrefix("")) > 0 {
		// When any change is made to the configuration we automatically publish a new version, so mark the version as updated
		// so dependent deployments will update appropriately to reference the newest version
		diff.SetNewComputed("version")
	}
	return nil
}

func waitForConfigurationDraftToBeActive(ctx context.Context, api *platformclientv2.WebDeploymentsApi, id string) diag.Diagnostics {
	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		configuration, resp, err := api.GetWebdeploymentsConfigurationVersionsDraft(id)
		if err != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Error verifying active status for new web deployment configuration %s: %s", id, err))
			}
			return retry.NonRetryableError(fmt.Errorf("Error verifying active status for new web deployment configuration %s: %s", id, err))
		}

		if *configuration.Status == "Active" {
			return nil
		}

		return retry.RetryableError(fmt.Errorf("Web deployment configuration %s not active yet. Status: %s", id, *configuration.Status))
	})
}

func readWebDeploymentConfigurationFromResourceData(d *schema.ResourceData) (string, *platformclientv2.Webdeploymentconfigurationversion) {
	name := d.Get("name").(string)
	languages := lists.InterfaceListToStrings(d.Get("languages").([]interface{}))
	defaultLanguage := d.Get("default_language").(string)

	inputCfg := &platformclientv2.Webdeploymentconfigurationversion{
		Name:            &name,
		Languages:       &languages,
		DefaultLanguage: &defaultLanguage,
	}

	description, ok := d.Get("description").(string)
	if ok {
		inputCfg.Description = &description
	}

	messengerSettings := readMessengerSettings(d)
	if messengerSettings != nil {
		inputCfg.Messenger = messengerSettings
	}

	cobrowseSettings := readCobrowseSettings(d)
	if cobrowseSettings != nil {
		inputCfg.Cobrowse = cobrowseSettings
	}

	journeySettings := readJourneySettings(d)
	if journeySettings != nil {
		inputCfg.JourneyEvents = journeySettings
	}

	supportCenterSettings := readSupportCenterSettings(d)
	if supportCenterSettings != nil {
		inputCfg.SupportCenter = supportCenterSettings
	}

	return name, inputCfg
}

func readJourneySettings(d *schema.ResourceData) *platformclientv2.Journeyeventssettings {
	value, ok := d.GetOk("journey_events")
	if !ok {
		return nil
	}

	cfgs := value.([]interface{})
	if len(cfgs) < 1 {
		return nil
	}

	cfg := cfgs[0].(map[string]interface{})
	enabled, _ := cfg["enabled"].(bool)
	journeySettings := &platformclientv2.Journeyeventssettings{
		Enabled: &enabled,
	}

	excludedQueryParams := lists.InterfaceListToStrings(cfg["excluded_query_parameters"].([]interface{}))
	journeySettings.ExcludedQueryParameters = &excludedQueryParams

	if keepUrlFragment, ok := cfg["should_keep_url_fragment"].(bool); ok && keepUrlFragment {
		journeySettings.ShouldKeepUrlFragment = &keepUrlFragment
	}

	searchQueryParameters := lists.InterfaceListToStrings(cfg["search_query_parameters"].([]interface{}))
	journeySettings.SearchQueryParameters = &searchQueryParameters

	pageviewConfig := cfg["pageview_config"]
	if value, ok := pageviewConfig.(string); ok {
		if value != "" {
			journeySettings.PageviewConfig = &value
		}
	}

	if clickEvents := readSelectorEventTriggers(cfg["click_event"].([]interface{})); clickEvents != nil {
		journeySettings.ClickEvents = clickEvents
	}

	if formsTrackEvents := readFormsTrackTriggers(cfg["form_track_event"].([]interface{})); formsTrackEvents != nil {
		journeySettings.FormsTrackEvents = formsTrackEvents
	}

	if idleEvents := readIdleEventTriggers(cfg["idle_event"].([]interface{})); idleEvents != nil {
		journeySettings.IdleEvents = idleEvents
	}

	if inViewportEvents := readSelectorEventTriggers(cfg["in_viewport_event"].([]interface{})); inViewportEvents != nil {
		journeySettings.InViewportEvents = inViewportEvents
	}

	if scrollDepthEvents := readScrollPercentageEventTriggers(cfg["scroll_depth_event"].([]interface{})); scrollDepthEvents != nil {
		journeySettings.ScrollDepthEvents = scrollDepthEvents
	}

	return journeySettings
}

func readSelectorEventTriggers(triggers []interface{}) *[]platformclientv2.Selectoreventtrigger {
	if triggers == nil || len(triggers) < 1 {
		return nil
	}

	results := make([]platformclientv2.Selectoreventtrigger, len(triggers))
	for i, value := range triggers {
		if trigger, ok := value.(map[string]interface{}); ok {
			selector := trigger["selector"].(string)
			eventName := trigger["event_name"].(string)
			results[i] = platformclientv2.Selectoreventtrigger{
				Selector:  &selector,
				EventName: &eventName,
			}
		}
	}

	return &results
}

func readFormsTrackTriggers(triggers []interface{}) *[]platformclientv2.Formstracktrigger {
	if triggers == nil || len(triggers) < 1 {
		return nil
	}

	results := make([]platformclientv2.Formstracktrigger, len(triggers))
	for i, value := range triggers {
		if trigger, ok := value.(map[string]interface{}); ok {
			selector := trigger["selector"].(string)
			formName := trigger["form_name"].(string)
			captureDataOnAbandon := trigger["capture_data_on_form_abandon"].(bool)
			captureDataOnSubmit := trigger["capture_data_on_form_submit"].(bool)
			results[i] = platformclientv2.Formstracktrigger{
				Selector:                 &selector,
				FormName:                 &formName,
				CaptureDataOnFormAbandon: &captureDataOnAbandon,
				CaptureDataOnFormSubmit:  &captureDataOnSubmit,
			}
		}
	}

	return &results
}

func readIdleEventTriggers(triggers []interface{}) *[]platformclientv2.Idleeventtrigger {
	if triggers == nil || len(triggers) < 1 {
		return nil
	}

	results := make([]platformclientv2.Idleeventtrigger, len(triggers))
	for i, value := range triggers {
		if trigger, ok := value.(map[string]interface{}); ok {
			eventName := trigger["event_name"].(string)
			idleAfterSeconds := trigger["idle_after_seconds"].(int)
			results[i] = platformclientv2.Idleeventtrigger{
				EventName:        &eventName,
				IdleAfterSeconds: &idleAfterSeconds,
			}
		}
	}

	return &results
}

func readScrollPercentageEventTriggers(triggers []interface{}) *[]platformclientv2.Scrollpercentageeventtrigger {
	if triggers == nil || len(triggers) < 1 {
		return nil
	}

	results := make([]platformclientv2.Scrollpercentageeventtrigger, len(triggers))
	for i, value := range triggers {
		if trigger, ok := value.(map[string]interface{}); ok {
			eventName := trigger["event_name"].(string)
			percentage := trigger["percentage"].(int)
			results[i] = platformclientv2.Scrollpercentageeventtrigger{
				EventName:  &eventName,
				Percentage: &percentage,
			}
		}
	}

	return &results
}

func readMessengerSettings(d *schema.ResourceData) *platformclientv2.Messengersettings {
	value, ok := d.GetOk("messenger")
	if !ok {
		return nil
	}

	cfgs := value.([]interface{})
	if len(cfgs) < 1 {
		return nil
	}

	cfg := cfgs[0].(map[string]interface{})
	enabled, _ := cfg["enabled"].(bool)
	messengerSettings := &platformclientv2.Messengersettings{
		Enabled: &enabled,
	}

	if styles, ok := cfg["styles"].([]interface{}); ok && len(styles) > 0 {
		style := styles[0].(map[string]interface{})
		if primaryColor, ok := style["primary_color"].(string); ok {
			messengerSettings.Styles = &platformclientv2.Messengerstyles{
				PrimaryColor: &primaryColor,
			}
		}
	}

	if launchers, ok := cfg["launcher_button"].([]interface{}); ok && len(launchers) > 0 {
		launcher := launchers[0].(map[string]interface{})
		if visibility, ok := launcher["visibility"].(string); ok {
			messengerSettings.LauncherButton = &platformclientv2.Launcherbuttonsettings{
				Visibility: &visibility,
			}
		}
	}

	if screens, ok := cfg["home_screen"].([]interface{}); ok && len(screens) > 0 {
		if screen, ok := screens[0].(map[string]interface{}); ok {
			enabled, enabledOk := screen["enabled"].(bool)
			logoUrl, logoUrlOk := screen["logo_url"].(string)

			if enabledOk && logoUrlOk {
				messengerSettings.HomeScreen = &platformclientv2.Messengerhomescreen{
					Enabled: &enabled,
					LogoUrl: &logoUrl,
				}
			}
		}
	}

	if fileUploads, ok := cfg["file_upload"].([]interface{}); ok && len(fileUploads) > 0 {
		fileUpload := fileUploads[0].(map[string]interface{})
		if modesCfg, ok := fileUpload["mode"].([]interface{}); ok && len(modesCfg) > 0 {
			modes := make([]platformclientv2.Fileuploadmode, len(modesCfg))
			for i, modeCfg := range modesCfg {
				if mode, ok := modeCfg.(map[string]interface{}); ok {
					maxFileSize := mode["max_file_size_kb"].(int)
					fileTypes := lists.InterfaceListToStrings(mode["file_types"].([]interface{}))
					modes[i] = platformclientv2.Fileuploadmode{
						FileTypes:     &fileTypes,
						MaxFileSizeKB: &maxFileSize,
					}
				}
			}

			if len(modes) > 0 {
				messengerSettings.FileUpload = &platformclientv2.Fileuploadsettings{
					Modes: &modes,
				}
			}
		}
	}

	return messengerSettings
}

func readCobrowseSettings(d *schema.ResourceData) *platformclientv2.Cobrowsesettings {
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
	channels := lists.InterfaceListToStrings(cfg["channels"].([]interface{}))
	maskSelectors := lists.InterfaceListToStrings(cfg["mask_selectors"].([]interface{}))
	readonlySelectors := lists.InterfaceListToStrings(cfg["readonly_selectors"].([]interface{}))

	return &platformclientv2.Cobrowsesettings{
		Enabled:           &enabled,
		AllowAgentControl: &allowAgentControl,
		Channels:          &channels,
		MaskSelectors:     &maskSelectors,
		ReadonlySelectors: &readonlySelectors,
	}
}

func readSupportCenterCategory(triggers []interface{}) *[]platformclientv2.Supportcentercategory {
	if triggers == nil || len(triggers) < 1 {
		return nil
	}

	results := make([]platformclientv2.Supportcentercategory, len(triggers))
	for i, value := range triggers {
		if trigger, ok := value.(map[string]interface{}); ok {
			id := trigger["enabled_categories_id"].(string)
			selfuri := trigger["self_uri"].(string)
			imageSource := trigger["image_source"].(string)

			image := &platformclientv2.Supportcenterimage{
				Source: &platformclientv2.Supportcenterimagesource{
					DefaultUrl: &imageSource,
				},
			}

			results[i] = platformclientv2.Supportcentercategory{
				Id:      &id,
				SelfUri: &selfuri,
				Image:   image,
			}
		}
	}

	return &results
}

func readSupportCenterCustomMessage(triggers []interface{}) *[]platformclientv2.Supportcentercustommessage {
	if triggers == nil || len(triggers) < 1 {
		return nil
	}

	results := make([]platformclientv2.Supportcentercustommessage, len(triggers))
	for i, value := range triggers {
		if trigger, ok := value.(map[string]interface{}); ok {
			defaultValue := trigger["default_value"].(string)
			varType := trigger["type"].(string)

			results[i] = platformclientv2.Supportcentercustommessage{
				DefaultValue: &defaultValue,
				VarType:      &varType,
			}
		}
	}

	return &results
}

func readSupportCenterStyleSetting(triggers []interface{}) *[]platformclientv2.Supportcenterstylesetting {
	if triggers == nil || len(triggers) < 1 {
		return nil
	}

	results := make([]platformclientv2.Supportcenterstylesetting, len(triggers))
	for i, value := range triggers {
		if trigger, ok := value.(map[string]interface{}); ok {
			herobackground := trigger["hero_style_background_color"].(string)
			herotextcolor := trigger["hero_style_text_color"].(string)
			heroimage := trigger["hero_style_image"].(string)

			herostyle := platformclientv2.Supportcenterstylesetting{
				HeroStyle: &platformclientv2.Supportcenterherostyle{
					BackgroundColor: &herobackground,
					TextColor:       &herotextcolor,
					Image: &platformclientv2.Supportcenterimage{
						Source: &platformclientv2.Supportcenterimagesource{
							DefaultUrl: &heroimage,
						},
					},
				},
			}

			globalbackground := trigger["global_style_background_color"].(string)
			globalprimarycolor := trigger["global_style_primary_color"].(string)
			globalprimarycolordark := trigger["global_style_primary_color_dark"].(string)
			globalprimarycolorlight := trigger["global_style_primary_color_light"].(string)
			globaltextcolor := trigger["global_style_text_color"].(string)
			globalfontfamily := trigger["global_style_font_family"].(string)

			globalstyle := platformclientv2.Supportcenterstylesetting{
				GlobalStyle: &platformclientv2.Supportcenterglobalstyle{
					BackgroundColor:   &globalbackground,
					PrimaryColor:      &globalprimarycolor,
					PrimaryColorDark:  &globalprimarycolordark,
					PrimaryColorLight: &globalprimarycolorlight,
					TextColor:         &globaltextcolor,
					FontFamily:        &globalfontfamily,
				},
			}

			// Assuming Supportcenterstylesetting has both HeroStyle and GlobalStyle
			results[i] = platformclientv2.Supportcenterstylesetting{
				HeroStyle:   herostyle.HeroStyle,
				GlobalStyle: globalstyle.GlobalStyle,
			}
		}
	}

	return &results
}

func readSupportCenterSettings(d *schema.ResourceData) *platformclientv2.Supportcentersettings {
	value, ok := d.GetOk("support_center")
	if !ok {
		return nil
	}

	cfgs := value.([]interface{})
	if len(cfgs) < 1 {
		return nil
	}

	cfg := cfgs[0].(map[string]interface{})
	enabled, _ := cfg["enabled"].(bool)
	supportCenterSettings := &platformclientv2.Supportcentersettings{
		Enabled: &enabled,
	}

	if id, ok := cfg["knowledge_base_id"].(string); ok {
		supportCenterSettings.KnowledgeBase = &platformclientv2.Addressableentityref{
			Id: &id,
		}
	}

	if selfuri, ok := cfg["knowledge_base_uri"].(string); ok {
		supportCenterSettings.KnowledgeBase = &platformclientv2.Addressableentityref{
			SelfUri: &selfuri,
		}
	}

	if routertype, ok := cfg["router_type"].(string); ok {
		supportCenterSettings.RouterType = &routertype
	}

	if customMessages := readSupportCenterCustomMessage(cfg["custom_messages"].([]interface{})); supportCenterCustomMessage != nil {
		supportCenterSettings.CustomMessages = customMessages
	}

	if supportCenterCategory := readSupportCenterCategory(cfg["enabled_categories"].([]interface{})); supportCenterCategory != nil {
		supportCenterSettings.EnabledCategories = supportCenterCategory
	}

	if feedbackEnabled, ok := cfg["feedback_enabled"].(bool); ok {
		supportCenterSettings.Feedback = &platformclientv2.Supportcenterfeedbacksettings{
			Enabled: &feedbackEnabled,
		}
	}

	return supportCenterSettings
}

func readSupportCenterScreens(triggers []interface{}) *[]platformclientv2.Supportcenterscreen {
	if triggers == nil || len(triggers) < 1 {
		return nil
	}

	results := make([]platformclientv2.Supportcenterscreen, len(triggers))
	for i, value := range triggers {
		if trigger, ok := value.(map[string]interface{}); ok {
			varType := trigger["type"].(string)
			moduleSettingsType := trigger["module_settings_type"].(string)
			moduleSettingsEnabled := trigger["module_settings_enabled"].(bool)
			moduleSettingsCompactCategoryModuleTemplate := trigger["module_settings_compact_category_module_template"].(bool)
			moduleSettingsDetailedCategoryModuleTemplate := trigger["module_settings_detailed_category_module_template"].(bool)

			moduleSettings := []platformclientv2.Supportcentermodulesetting{
				{
					VarType: &moduleSettingsType,
					Enabled: &moduleSettingsEnabled,
					CompactCategoryModuleTemplate: &platformclientv2.Supportcentercompactcategorymoduletemplate{
						Active: &moduleSettingsCompactCategoryModuleTemplate,
					},
					DetailedCategoryModuleTemplate: &platformclientv2.Supportcenterdetailedcategorymoduletemplate{
						Active: &moduleSettingsDetailedCategoryModuleTemplate,
					},
				},
			}

			results[i] = platformclientv2.Supportcenterscreen{
				VarType:        &varType,
				ModuleSettings: &moduleSettings,
			}
		}
	}

	return &results
}

// featureNotImplemented checks the response object to find out if the request failed because a feature is not yet
// implemented in the org that it was ran against. If true, we can pass back the field name and give more context
// in the final error message.
func featureNotImplemented(response *platformclientv2.APIResponse) (bool, string) {
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

func createWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name, inputCfg := readWebDeploymentConfigurationFromResourceData(d)

	log.Printf("Creating web deployment configuration %s", name)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	diagErr := WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		configuration, resp, err := api.PostWebdeploymentsConfigurations(*inputCfg)
		if err != nil {
			var extraErrorInfo string
			featureIsNotImplemented, fieldName := featureNotImplemented(resp)
			if featureIsNotImplemented {
				extraErrorInfo = fmt.Sprintf("Feature '%s' is not yet implemented", fieldName)
			}
			if IsStatus400(resp) {
				return retry.RetryableError(fmt.Errorf("failed to create web deployment configuration %s: %s. %s", name, err, extraErrorInfo))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to create web deployment configuration %s: %s. %s", name, err, extraErrorInfo))
		}
		d.SetId(*configuration.Id)
		d.Set("status", configuration.Status)

		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	activeError := waitForConfigurationDraftToBeActive(ctx, api, d.Id())
	if activeError != nil {
		return diag.Errorf("Web deployment configuration %s did not become active and could not be published", name)
	}

	diagErr = WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		configuration, resp, err := api.PostWebdeploymentsConfigurationVersionsDraftPublish(d.Id())
		if err != nil {
			if IsStatus400(resp) {
				return retry.RetryableError(fmt.Errorf("Error publishing web deployment configuration %s: %s", name, err))
			}
			return retry.NonRetryableError(fmt.Errorf("Error publishing web deployment configuration %s: %s", name, err))
		}
		d.Set("version", configuration.Version)
		d.Set("status", configuration.Status)
		log.Printf("Created web deployment configuration %s %s", name, *configuration.Id)

		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	return readWebDeploymentConfiguration(ctx, d, meta)
}

func determineLatestVersion(ctx context.Context, api *platformclientv2.WebDeploymentsApi, configurationId string) string {
	version := ""
	draft := "DRAFT"
	_ = WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		versions, resp, getErr := api.GetWebdeploymentsConfigurationVersions(configurationId)
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to determine latest version %s", getErr))
			}
			log.Printf("Failed to determine latest version. Defaulting to DRAFT. Details: %s", getErr)
			version = draft
			return retry.NonRetryableError(fmt.Errorf("Failed to determine latest version %s", getErr))
		}

		maxVersion := 0
		for _, v := range *versions.Entities {
			if *v.Version == draft {
				continue
			}
			APIVersion, err := strconv.Atoi(*v.Version)
			if err != nil {
				log.Printf("Failed to convert version %s to an integer", *v.Version)
			} else {
				if APIVersion > maxVersion {
					maxVersion = APIVersion
				}
			}
		}

		if maxVersion == 0 {
			version = draft
		} else {
			version = strconv.Itoa(maxVersion)
		}

		return nil
	})

	return version
}

func readWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	version := d.Get("version").(string)
	log.Printf("Reading web deployment configuration %s", d.Id())
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		if version == "" {
			version = determineLatestVersion(ctx, api, d.Id())
		}
		configuration, resp, getErr := api.GetWebdeploymentsConfigurationVersion(d.Id(), version)
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read web deployment configuration %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read web deployment configuration %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceWebDeploymentConfiguration())
		d.Set("name", *configuration.Name)
		if configuration.Description != nil {
			d.Set("description", *configuration.Description)
		}
		if configuration.Languages != nil {
			d.Set("languages", *configuration.Languages)
		}
		if configuration.DefaultLanguage != nil {
			d.Set("default_language", *configuration.DefaultLanguage)
		}
		if configuration.Status != nil {
			d.Set("status", *configuration.Status)
		}
		if configuration.Version != nil {
			d.Set("version", *configuration.Version)
		}
		if configuration.Messenger != nil {
			d.Set("messenger", flattenMessengerSettings(configuration.Messenger))
		}
		if configuration.Cobrowse != nil {
			d.Set("cobrowse", flattenCobrowseSettings(configuration.Cobrowse))
		}
		if configuration.JourneyEvents != nil {
			d.Set("journey_events", flattenJourneyEvents(configuration.JourneyEvents))
		}
		if configuration.SupportCenter != nil {
			d.Set("support_center", flattenSupportCenterSettings(configuration.SupportCenter))
		}

		log.Printf("Read web deployment configuration %s %s", d.Id(), *configuration.Name)
		return cc.CheckState()
	})
}

func updateWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name, inputCfg := readWebDeploymentConfigurationFromResourceData(d)

	log.Printf("Updating web deployment configuration %s", name)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	diagErr := WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := api.PutWebdeploymentsConfigurationVersionsDraft(d.Id(), *inputCfg)
		if err != nil {
			if IsStatus400(resp) {
				return retry.RetryableError(fmt.Errorf("Error updating web deployment configuration %s: %s", name, err))
			}
			return retry.NonRetryableError(fmt.Errorf("Error updating web deployment configuration %s: %s", name, err))
		}
		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	activeError := waitForConfigurationDraftToBeActive(ctx, api, d.Id())
	if activeError != nil {
		return diag.Errorf("Web deployment configuration %s did not become active and could not be published", name)
	}

	diagErr = WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		configuration, resp, err := api.PostWebdeploymentsConfigurationVersionsDraftPublish(d.Id())
		if err != nil {
			if IsStatus400(resp) {
				return retry.RetryableError(fmt.Errorf("Error publishing web deployment configuration %s: %s", name, err))
			}
			return retry.NonRetryableError(fmt.Errorf("Error publishing web deployment configuration %s: %s", name, err))
		}
		d.Set("version", configuration.Version)
		d.Set("status", configuration.Status)
		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished updating web deployment configuration %s", name)
	return readWebDeploymentConfiguration(ctx, d, meta)
}

func deleteWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	log.Printf("Deleting web deployment configuration %s", name)
	_, err := api.DeleteWebdeploymentsConfiguration(d.Id())

	if err != nil {
		return diag.Errorf("Failed to delete web deployment configuration %s: %s", name, err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := api.GetWebdeploymentsConfigurationVersionsDraft(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				log.Printf("Deleted web deployment configuration %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting web deployment configuration %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Web deployment configuration %s still exists", d.Id()))
	})
}

func validateConfigurationStatusChange(k, old, new string, d *schema.ResourceData) bool {
	// Configs start in a pending status and may not transition to active or error before we retrieve the state, so allow
	// the status to change from pending to something less ephemeral
	return old == "Pending"
}

func flattenMessengerSettings(messengerSettings *platformclientv2.Messengersettings) []interface{} {
	if messengerSettings == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"enabled":         messengerSettings.Enabled,
		"styles":          flattenStyles(messengerSettings.Styles),
		"launcher_button": flattenLauncherButton(messengerSettings.LauncherButton),
		"home_screen":     flattenHomeScreen(messengerSettings.HomeScreen),
		"file_upload":     flattenFileUpload(messengerSettings.FileUpload),
	}}
}

func flattenCobrowseSettings(cobrowseSettings *platformclientv2.Cobrowsesettings) []interface{} {
	if cobrowseSettings == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"enabled":             cobrowseSettings.Enabled,
		"allow_agent_control": cobrowseSettings.AllowAgentControl,
		"channels":            cobrowseSettings.Channels,
		"mask_selectors":      cobrowseSettings.MaskSelectors,
		"readonly_selectors":  cobrowseSettings.ReadonlySelectors,
	}}
}

func flattenStyles(styles *platformclientv2.Messengerstyles) []interface{} {
	if styles == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"primary_color": styles.PrimaryColor,
	}}
}

func flattenLauncherButton(settings *platformclientv2.Launcherbuttonsettings) []interface{} {
	if settings == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"visibility": settings.Visibility,
	}}
}

func flattenHomeScreen(settings *platformclientv2.Messengerhomescreen) []interface{} {
	if settings == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"enabled":  settings.Enabled,
		"logo_url": settings.LogoUrl,
	}}
}

func flattenFileUpload(settings *platformclientv2.Fileuploadsettings) []interface{} {
	if settings == nil || settings.Modes == nil || len(*settings.Modes) < 1 {
		return nil
	}

	modes := make([]map[string]interface{}, len(*settings.Modes))
	for i, mode := range *settings.Modes {
		modes[i] = map[string]interface{}{
			"file_types":       *mode.FileTypes,
			"max_file_size_kb": *mode.MaxFileSizeKB,
		}
	}

	return []interface{}{map[string]interface{}{
		"mode": modes,
	}}
}

func flattenJourneyEvents(journeyEvents *platformclientv2.Journeyeventssettings) []interface{} {
	if journeyEvents == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"enabled":                   journeyEvents.Enabled,
		"excluded_query_parameters": journeyEvents.ExcludedQueryParameters,
		"should_keep_url_fragment":  journeyEvents.ShouldKeepUrlFragment,
		"search_query_parameters":   journeyEvents.SearchQueryParameters,
		"pageview_config":           journeyEvents.PageviewConfig,
		"click_event":               flattenSelectorEventTriggers(journeyEvents.ClickEvents),
		"form_track_event":          flattenFormsTrackTriggers(journeyEvents.FormsTrackEvents),
		"idle_event":                flattenIdleEventTriggers(journeyEvents.IdleEvents),
		"in_viewport_event":         flattenSelectorEventTriggers(journeyEvents.InViewportEvents),
		"scroll_depth_event":        flattenScrollPercentageEventTriggers(journeyEvents.ScrollDepthEvents),
	}}
}

func flattenSelectorEventTriggers(triggers *[]platformclientv2.Selectoreventtrigger) []interface{} {
	if triggers == nil || len(*triggers) < 1 {
		return nil
	}

	result := make([]interface{}, len(*triggers))
	for i, trigger := range *triggers {
		result[i] = map[string]interface{}{
			"selector":   trigger.Selector,
			"event_name": trigger.EventName,
		}
	}
	return result
}

func flattenFormsTrackTriggers(triggers *[]platformclientv2.Formstracktrigger) []interface{} {
	if triggers == nil || len(*triggers) < 1 {
		return nil
	}

	result := make([]interface{}, len(*triggers))
	for i, trigger := range *triggers {
		result[i] = map[string]interface{}{
			"selector":                     trigger.Selector,
			"form_name":                    trigger.FormName,
			"capture_data_on_form_abandon": trigger.CaptureDataOnFormAbandon,
			"capture_data_on_form_submit":  trigger.CaptureDataOnFormSubmit,
		}
	}
	return result
}

func flattenIdleEventTriggers(triggers *[]platformclientv2.Idleeventtrigger) []interface{} {
	if triggers == nil || len(*triggers) < 1 {
		return nil
	}

	result := make([]interface{}, len(*triggers))
	for i, trigger := range *triggers {
		result[i] = map[string]interface{}{
			"event_name":         trigger.EventName,
			"idle_after_seconds": trigger.IdleAfterSeconds,
		}
	}
	return result
}

func flattenScrollPercentageEventTriggers(triggers *[]platformclientv2.Scrollpercentageeventtrigger) []interface{} {
	if triggers == nil || len(*triggers) < 1 {
		return nil
	}

	result := make([]interface{}, len(*triggers))
	for i, trigger := range *triggers {
		result[i] = map[string]interface{}{
			"event_name": trigger.EventName,
			"percentage": trigger.Percentage,
		}
	}
	return result
}

func flattenKnowledgeBase(knowledgebase *platformclientv2.Addressableentityref) []interface{} {
	if knowledgebase == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"knowledgebase_id": knowledgebase.Id,
	}}
}
func flattenSupportCenterCustomMessage(triggers *[]platformclientv2.Supportcentercustommessage) []interface{} {
	if triggers == nil || len(*triggers) < 1 {
		return nil
	}

	result := make([]interface{}, len(*triggers))
	for i, trigger := range *triggers {
		result[i] = map[string]interface{}{
			"default_value": trigger.DefaultValue,
			"type":          trigger.VarType,
		}
	}
	return result
}

func flattenSupportCenterCategory(triggers *[]platformclientv2.Supportcentercategory) []interface{} {
	if triggers == nil || len(*triggers) < 1 {
		return nil
	}

	result := make([]interface{}, len(*triggers))
	for i, trigger := range *triggers {
		result[i] = map[string]interface{}{
			"enabled_categories_id": trigger.Id,
			"self_uri":              trigger.SelfUri,
			"image_source":          trigger.Image,
		}
	}
	return result
}

func flattenSupportCenterSettings(supportCenterSettings *platformclientv2.Supportcentersettings) []interface{} {
	if supportCenterSettings == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"enabled":            supportCenterSettings.Enabled,
		"knowledgebase":      flattenKnowledgeBase(supportCenterSettings.KnowledgeBase),
		"custom_messages":    flattenSupportCenterCustomMessage(supportCenterSettings.CustomMessages),
		"enabled_categories": flattenSupportCenterCategory(supportCenterSettings.EnabledCategories),
	}}
}
