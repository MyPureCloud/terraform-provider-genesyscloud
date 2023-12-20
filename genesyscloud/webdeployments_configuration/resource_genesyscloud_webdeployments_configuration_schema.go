package webdeployments_configuration

import (
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const resourceName = "genesyscloud_webdeployments_configuration"

// SetRegistrar registers all the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, DataSourceWebDeploymentsConfiguration())
	l.RegisterResource(resourceName, ResourceWebDeploymentConfiguration())
	l.RegisterExporter(resourceName, WebDeploymentConfigurationExporter())
}

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

func DataSourceWebDeploymentsConfiguration() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Web Deployments Configurations. Select a configuration by name.",
		ReadContext: gcloud.ReadWithPooledClient(dataSourceConfigurationRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the configuration",
				Type:        schema.TypeString,
				Required:    true,
			},
			"version": {
				Description: "The version of the configuration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func ResourceWebDeploymentConfiguration() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Web Deployment Configuration",

		CreateContext: gcloud.CreateWithPooledClient(createWebDeploymentConfiguration),
		ReadContext:   gcloud.ReadWithPooledClient(readWebDeploymentConfiguration),
		UpdateContext: gcloud.UpdateWithPooledClient(updateWebDeploymentConfiguration),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteWebDeploymentConfiguration),
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

func WebDeploymentConfigurationExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc:   gcloud.GetAllWithPooledClient(getAllWebDeploymentConfigurations),
		ExcludedAttributes: []string{"version"},
	}
}
