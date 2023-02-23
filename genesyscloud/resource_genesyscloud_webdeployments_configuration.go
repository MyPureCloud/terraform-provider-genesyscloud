package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v92/platformclientv2"
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
				Default:     true,
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
				Default:     true,
			},
			"allow_agent_control": {
				Description: "Whether agent can take control over customer's screen or not",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"mask_selectors": {
				Description: "List of CSS selectors which should be masked when screen sharing is active",
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
)

func getAllWebDeploymentConfigurations(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	webDeploymentsAPI := platformclientv2.NewWebDeploymentsApiWithConfig(clientConfig)

	configurations, _, getErr := webDeploymentsAPI.GetWebdeploymentsConfigurations(false)
	if getErr != nil {
		return nil, diag.Errorf("Failed to get web deployment configurations: %v", getErr)
	}

	for _, configuration := range *configurations.Entities {
		resources[*configuration.Id] = &ResourceMeta{Name: *configuration.Name}
	}

	return resources, nil
}

func webDeploymentConfigurationExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllWebDeploymentConfigurations),
	}
}

func resourceWebDeploymentConfiguration() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Web Deployment Configuration",

		CreateContext: createWithPooledClient(createWebDeploymentConfiguration),
		ReadContext:   readWithPooledClient(readWebDeploymentConfiguration),
		UpdateContext: updateWithPooledClient(updateWebDeploymentConfiguration),
		DeleteContext: deleteWithPooledClient(deleteWebDeploymentConfiguration),
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
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"default_language": {
				Description: "The default language to use for the configuration.",
				Type:        schema.TypeString,
				Optional:    true,
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
				Optional:    true,
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
	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		configuration, resp, err := api.GetWebdeploymentsConfigurationVersionsDraft(id)
		if err != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Error verifying active status for new web deployment configuration %s: %s", id, err))
			}
			return resource.NonRetryableError(fmt.Errorf("Error verifying active status for new web deployment configuration %s: %s", id, err))
		}

		if *configuration.Status == "Active" {
			return nil
		}

		return resource.RetryableError(fmt.Errorf("Web deployment configuration %s not active yet. Status: %s", id, *configuration.Status))
	})
}

func readWebDeploymentConfigurationFromResourceData(d *schema.ResourceData) (string, *platformclientv2.Webdeploymentconfigurationversion) {
	name := d.Get("name").(string)
	languages := interfaceListToStrings(d.Get("languages").([]interface{}))
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
	log.Println("Processing journey events config: ", cfg)
	enabled, _ := cfg["enabled"].(bool)
	journeySettings := &platformclientv2.Journeyeventssettings{
		Enabled: &enabled,
	}

	excludedQueryParams := interfaceListToStrings(cfg["excluded_query_parameters"].([]interface{}))
	journeySettings.ExcludedQueryParameters = &excludedQueryParams

	if keepUrlFragment, ok := cfg["should_keep_url_fragment"].(bool); ok && keepUrlFragment {
		journeySettings.ShouldKeepUrlFragment = &keepUrlFragment
	}

	searchQueryParameters := interfaceListToStrings(cfg["search_query_parameters"].([]interface{}))
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

	if fileUploads, ok := cfg["file_upload"].([]interface{}); ok && len(fileUploads) > 0 {
		fileUpload := fileUploads[0].(map[string]interface{})
		if modesCfg, ok := fileUpload["mode"].([]interface{}); ok && len(modesCfg) > 0 {
			modes := make([]platformclientv2.Fileuploadmode, len(modesCfg))
			for i, modeCfg := range modesCfg {
				if mode, ok := modeCfg.(map[string]interface{}); ok {
					maxFileSize := mode["max_file_size_kb"].(int)
					fileTypes := interfaceListToStrings(mode["file_types"].([]interface{}))
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
	cobrowseSettings := &platformclientv2.Cobrowsesettings{
		Enabled: &enabled,
	}

	allowAgentControl, _ := cfg["allow_agent_control"].(bool)
	cobrowseSettings = &platformclientv2.Cobrowsesettings{
		AllowAgentControl: &allowAgentControl,
	}

	maskSelectors := interfaceListToStrings(cfg["mask_selectors"].([]interface{}))
	cobrowseSettings = &platformclientv2.Cobrowsesettings{
		MaskSelectors: &maskSelectors,
	}

	return cobrowseSettings
}

func createWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name, inputCfg := readWebDeploymentConfigurationFromResourceData(d)

	log.Printf("Creating web deployment configuration %s", name)

	sdkConfig := meta.(*providerMeta).ClientConfig
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	diagErr := withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		configuration, resp, err := api.PostWebdeploymentsConfigurations(*inputCfg)
		if err != nil {
			if isStatus400(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to create web deployment configuration %s: %s", name, err))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to create web deployment configuration %s: %s", name, err))
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

	diagErr = withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		configuration, resp, err := api.PostWebdeploymentsConfigurationVersionsDraftPublish(d.Id())
		if err != nil {
			if isStatus400(resp) {
				return resource.RetryableError(fmt.Errorf("Error publishing web deployment configuration %s: %s", name, err))
			}
			return resource.NonRetryableError(fmt.Errorf("Error publishing web deployment configuration %s: %s", name, err))
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
	_ = withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		versions, resp, getErr := api.GetWebdeploymentsConfigurationVersions(configurationId)
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to determine latest version %s", getErr))
			}
			log.Printf("Failed to determine latest version. Defaulting to DRAFT. Details: %s", getErr)
			version = "draft"
			return resource.NonRetryableError(fmt.Errorf("Failed to determine latest version %s", getErr))
		}

		maxVersion := 0
		for _, v := range *versions.Entities {
			if *v.Version == "DRAFT" {
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
			version = "draft"
		}

		version = strconv.Itoa(maxVersion)

		return nil
	})

	return version
}

func readWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	version := d.Get("version").(string)
	log.Printf("Reading web deployment configuration %s", d.Id())
	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		if version == "" {
			version = determineLatestVersion(ctx, api, d.Id())
		}
		configuration, resp, getErr := api.GetWebdeploymentsConfigurationVersion(d.Id(), version)
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read web deployment configuration %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read web deployment configuration %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceWebDeploymentConfiguration())
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

		log.Printf("Read web deployment configuration %s %s", d.Id(), *configuration.Name)
		return cc.CheckState()
	})
}

func updateWebDeploymentConfiguration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name, inputCfg := readWebDeploymentConfigurationFromResourceData(d)

	log.Printf("Updating web deployment configuration %s", name)

	sdkConfig := meta.(*providerMeta).ClientConfig
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	diagErr := withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := api.PutWebdeploymentsConfigurationVersionsDraft(d.Id(), *inputCfg)
		if err != nil {
			if isStatus400(resp) {
				return resource.RetryableError(fmt.Errorf("Error updating web deployment configuration %s: %s", name, err))
			}
			return resource.NonRetryableError(fmt.Errorf("Error updating web deployment configuration %s: %s", name, err))
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

	diagErr = withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		configuration, resp, err := api.PostWebdeploymentsConfigurationVersionsDraftPublish(d.Id())
		if err != nil {
			if isStatus400(resp) {
				return resource.RetryableError(fmt.Errorf("Error publishing web deployment configuration %s: %s", name, err))
			}
			return resource.NonRetryableError(fmt.Errorf("Error publishing web deployment configuration %s: %s", name, err))
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

	sdkConfig := meta.(*providerMeta).ClientConfig
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	log.Printf("Deleting web deployment configuration %s", name)
	_, err := api.DeleteWebdeploymentsConfiguration(d.Id())

	if err != nil {
		return diag.Errorf("Failed to delete web deployment configuration %s: %s", name, err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := api.GetWebdeploymentsConfigurationVersionsDraft(d.Id())
		if err != nil {
			if isStatus404(resp) {
				log.Printf("Deleted web deployment configuration %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting web deployment configuration %s: %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("Web deployment configuration %s still exists", d.Id()))
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
		"mask_selectors":      cobrowseSettings.MaskSelectors,
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
