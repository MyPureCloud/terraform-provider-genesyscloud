package webdeployments_configuration

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/validators"
	wdcUtils "terraform-provider-genesyscloud/genesyscloud/webdeployments_configuration/utils"

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
	customI18nLabel = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"language": {
				Description: "Language of localized labels in homescreen app (eg. en-us, de-de)",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"localized_labels": {
				Description: "Contains localized labels used in homescreen app",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Description:  "Contains localized label key used in messenger homescreen",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"MessengerHomeHeaderTitle", "MessengerHomeHeaderSubTitle"}, false),
						},
						"value": {
							Description: "Contains localized label value used in messenger homescreen",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
		},
	}

	position = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"alignment": {
				Description:  "The alignment for position",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Auto", "Left", "Right"}, false),
			},
			"side_space": {
				Description: "The sidespace value for position",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"bottom_space": {
				Description: "The bottomspace value for position",
				Type:        schema.TypeInt,
				Optional:    true,
			},
		},
	}

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
			"apps": {
				Description: "The apps embedded in the messenger",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"conversations": {
							Description: "Conversation settings that handles chats within the messenger",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Description: "The toggle to enable or disable conversations",
										Type:        schema.TypeBool,
										Optional:    true,
										Computed:    true,
									},
									"show_agent_typing_indicator": {
										Description: "The toggle to enable or disable typing indicator for messenger",
										Type:        schema.TypeBool,
										Optional:    true,
										Computed:    true,
									},
									"show_user_typing_indicator": {
										Description: "The toggle to enable or disable typing indicator for messenger",
										Type:        schema.TypeBool,
										Optional:    true,
										Computed:    true,
									},
									"auto_start_enabled": {
										Description: "The auto start for the messenger conversation",
										Type:        schema.TypeBool,
										Optional:    true,
										Computed:    true,
									},
									"markdown_enabled": {
										Description: "The markdown for the messenger app",
										Type:        schema.TypeBool,
										Optional:    true,
										Computed:    true,
									},
									"conversation_disconnect": {
										Description: "The conversation disconnect for the messenger app",
										Type:        schema.TypeList,
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled": {
													Description: "whether or not conversation disconnect setting is enabled",
													Type:        schema.TypeBool,
													Optional:    true,
												},
												"type": {
													Description:  "Conversation disconnect type",
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice([]string{"Send", "ReadOnly"}, false),
												},
											},
										},
									},
									"conversation_clear_enabled": {
										Description: "The conversation clear settings for the messenger app",
										Type:        schema.TypeBool,
										Optional:    true,
										Computed:    true,
									},
									"humanize": {
										Description: "The humanize conversations settings for the messenger app",
										Type:        schema.TypeList,
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled": {
													Description: "Whether or not humanize conversations setting is enabled",
													Type:        schema.TypeBool,
													Optional:    true,
												},
												"bot": {
													Description: "Bot messenger profile setting",
													Type:        schema.TypeList,
													MaxItems:    1,
													Optional:    true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Description: "The name of the bot",
																Type:        schema.TypeString,
																Optional:    true,
															},
															"avatar_url": {
																Description: "The avatar URL of the bot",
																Type:        schema.TypeString,
																Optional:    true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"knowledge": {
							Description: "The knowledge base config for messenger",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Description: "whether or not knowledge base is enabled",
										Type:        schema.TypeBool,
										Optional:    true,
									},
									"knowledge_base_id": {
										Description: "The knowledge base for messenger",
										Type:        schema.TypeString,
										Optional:    true,
									},
								},
							},
						},
					},
				},
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
			"allow_agent_navigation": {
				Description: "Whether agent can use navigation feature over customer's screen or not",
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
			"pause_criteria": {
				Description: "Pause criteria that will pause cobrowse if some of them are met in the user's URL",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url_fragment": {
							Description: "A string representing a part of the URL that, when matched according to the specified condition, will trigger a pause in the cobrowse session",
							Type:        schema.TypeString,
							Required:    true,
						},
						"condition": {
							Description:  "The condition to be applied to the `url_fragment`. Conditions are 'includes', 'does_not_include', 'starts_with', 'ends_with', 'equals'",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"includes", "does_not_include", "starts_with", "ends_with", "equals"}, false),
						},
					},
				},
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

	customMessage = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"default_value": {
				Description: "Default value for the custom message",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description:  "The custom message type. (Welcome or Fallback)",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Welcome", "Fallback"}, false),
			},
		},
	}

	supportCenterModuleSetting = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "Screen module type",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Search", "Categories", "FAQ", "Contact", "Results", "Article", "TopViewedArticles"}, false),
			},
			"enabled": {
				Description: "Whether or not knowledge portal (previously support center) screen module is enabled",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"compact_category_module_template_active": {
				Description: "Whether the Support Center Compact Category Module Template is active or not",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"detailed_category_module_template": {
				Description: "Detailed category module template settings",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"active": {
							Description: "Whether the Support Center Detailed Category Module Template is active or not",
							Type:        schema.TypeBool,
							Required:    true,
						},
						"sidebar_enabled": {
							Description: "Whether the Support Center Detailed Category Module Sidebar is active or not",
							Type:        schema.TypeBool,
							Required:    true,
						},
					},
				},
			},
		},
	}

	supportCenterScreen = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "The type of the screen",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Home", "Category", "SearchResults", "Article"}, false),
			},
			"module_settings": {
				Description: "Module settings for the screen, valid modules for each screenType: Home: Search, Categories, TopViewedArticles; Category: Search, Categories; SearchResults: Search, Results; Article: Search, Article;",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        supportCenterModuleSetting,
			},
		},
	}

	styleSetting = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"hero_style_setting": {
				Description: "Knowledge portal (previously support center) hero customizations",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"background_color": {
							Description:      "Background color for hero section, in hexadecimal format, eg #ffffff",
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validators.ValidateHexColor,
						},
						"text_color": {
							Description:      "Text color for hero section, in hexadecimal format, eg #ffffff",
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validators.ValidateHexColor,
						},
						"image_uri": {
							Description:  "Background image for hero section",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsURLWithHTTPS,
						},
					},
				},
			},
			"global_style_setting": {
				Description: "Knowledge portal (previously support center) global customizations",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"background_color": {
							Description:      "Global background color, in hexadecimal format, eg #ffffff",
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validators.ValidateHexColor,
						},
						"primary_color": {
							Description:      "Global primary color, in hexadecimal format, eg #ffffff",
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validators.ValidateHexColor,
						},
						"primary_color_dark": {
							Description:      "Global dark primary color, in hexadecimal format, eg #ffffff",
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validators.ValidateHexColor,
						},
						"primary_color_light": {
							Description:      "Global light primary color, in hexadecimal format, eg #ffffff",
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validators.ValidateHexColor,
						},
						"text_color": {
							Description:      "Global text color, in hexadecimal format, eg #ffffff",
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validators.ValidateHexColor,
						},
						"font_family": {
							Description: "Global font family",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
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

	supportCenterSettings = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"enabled": {
				Description: "Whether or not knowledge portal (previously support center) is enabled",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"knowledge_base_id": {
				Description: "The knowledge base for knowledge portal (previously support center)",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"custom_messages": {
				Description: "Customizable display texts for knowledge portal",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        customMessage,
			},
			"router_type": {
				Description:  "Router type for knowledge portal",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Hash", "Browser"}, false),
			},
			"screens": {
				Description: "Available screens for the knowledge portal with its modules",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        supportCenterScreen,
			},
			"enabled_categories": {
				Description: "Featured categories for knowledge portal (previously support center) home screen",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"category_id": {
							Description: "The knowledge base category id",
							Type:        schema.TypeString,
							Required:    true,
						},
						"image_uri": {
							Description:  "Source URL for the featured category",
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsURLWithHTTPS,
						},
					},
				},
			},
			"style_setting": {
				Description: "Style attributes for knowledge portal (previously support center)",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        styleSetting,
			},
			"feedback_enabled": {
				Description: "Whether or not requesting customer feedback on article content and article search results is enabled",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}

	authenticationSettings = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"enabled": {
				Description: "Indicate if these auth is required for this deployment. If, for example, this flag is set to true then webmessaging sessions can not send messages unless the end-user is authenticated.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"integration_id": {
				Description: "The integration identifier which contains the auth settings required on the deployment.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
)

func DataSourceWebDeploymentsConfiguration() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Web Deployments Configurations. Select a configuration by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceConfigurationRead),
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

		CreateContext: provider.CreateWithPooledClient(createWebDeploymentConfiguration),
		ReadContext:   provider.ReadWithPooledClient(readWebDeploymentConfiguration),
		UpdateContext: provider.UpdateWithPooledClient(updateWebDeploymentConfiguration),
		DeleteContext: provider.DeleteWithPooledClient(deleteWebDeploymentConfiguration),
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
			"headless_mode_enabled": {
				Description: "Headless Mode Support which Controls UI components. When enabled, native UI components will be disabled and allows for custom-built UI.",
				Type:        schema.TypeBool,
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
				DiffSuppressFunc: wdcUtils.ValidateConfigurationStatusChange,
			},
			"version": {
				Description: "The version of the configuration.",
				Type:        schema.TypeString,
				Computed:    true,
				MaxItems:    0,
			},
			"custom_i18n_labels": {
				Description: "The localization settings for homescreen app",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        customI18nLabel,
			},
			"position": {
				Description: "Settings concerning position",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        position,
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
				Description: "Settings concerning knowledge portal (previously support center)",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        supportCenterSettings,
			},
			"authentication_settings": {
				Description: "Settings for authenticated webdeployments.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        authenticationSettings,
			},
		},
		CustomizeDiff: wdcUtils.CustomizeConfigurationDiff,
	}
}

func WebDeploymentConfigurationExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc:   provider.GetAllWithPooledClient(getAllWebDeploymentConfigurations),
		ExcludedAttributes: []string{"version"},
		RemoveIfMissing: map[string][]string{
			"authentication_settings": {"integration_id"},
		},
	}
}
