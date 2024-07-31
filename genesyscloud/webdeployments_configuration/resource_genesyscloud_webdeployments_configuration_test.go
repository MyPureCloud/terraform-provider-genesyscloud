package webdeployments_configuration

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

type scCustomMessageConfig struct {
	defaultVal string
	varType    string
}

type scModuleSetting struct {
	varType                        string
	enabled                        bool
	compactTemplateActive          bool
	detailedTemplateActive         bool
	detailedTemplateSidebarEnabled bool
}

type scScreenConfig struct {
	varType        string
	moduleSettings []scModuleSetting
}

type scEnabledCategoryConfig struct {
	categoryId string
	imageUri   string
}

type heroStyleSettingConfig struct {
	bgColor   string
	textColor string
	imageUri  string
}

type globalStyleSettingConfig struct {
	bgColor           string
	primaryColor      string
	primaryColorDark  string
	primaryColorLight string
	textColor         string
	fontFamily        string
}

type scStyleSettingConfig struct {
	heroStyleSetting   heroStyleSettingConfig
	globalStyleSetting globalStyleSettingConfig
}

type scConfig struct {
	enabled           bool
	kbId              string
	customMessages    []scCustomMessageConfig
	routerType        string
	screens           []scScreenConfig
	enabledCategories []scEnabledCategoryConfig
	styleSetting      scStyleSettingConfig
	feedbackEnabled   bool
}

func TestAccResourceWebDeploymentsConfiguration(t *testing.T) {
	t.Parallel()
	var (
		resName                  = "webdeploy-config-test"
		fullResName              = resourceName + "." + resName
		configName               = "tf-config-" + uuid.NewString()
		configDescription        = "Test Configuration description"
		updatedConfigDescription = configDescription + " Updated"
		languages1               = []string{"en-us", "ja"}
		defaultLang1             = "en-us"
		languages2               = []string{"es"}
		defaultLang2             = "es"
	)

	cleanupWebDeploymentsConfiguration(t, "Test Configuration ")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateConfigurationResource(
					resName,
					configName,
					configDescription,
					languages1,
					defaultLang1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResName, "name", configName),
					resource.TestCheckResourceAttr(fullResName, "description", configDescription),
					resource.TestMatchResourceAttr(fullResName, "status", regexp.MustCompile("^(Pending|Active)$")),
					resource.TestCheckResourceAttrSet(fullResName, "version"),
					resource.TestCheckResourceAttr(fullResName, "languages.#", strconv.Itoa(len(languages1))),
					resource.TestCheckResourceAttr(fullResName, "languages.0", languages1[0]),
					resource.TestCheckResourceAttr(fullResName, "languages.1", languages1[1]),
					resource.TestCheckResourceAttr(fullResName, "default_language", defaultLang1),
					resource.TestCheckResourceAttr(fullResName, "messenger.#", "0"),
					resource.TestCheckResourceAttr(fullResName, "cobrowse.#", "0"),
					resource.TestCheckResourceAttr(fullResName, "journey_events.#", "0"),
					resource.TestCheckResourceAttr(fullResName, "support_center.#", "0"),
				),
			},
			{
				Config: generateConfigurationResource(
					resName,
					configName,
					updatedConfigDescription,
					languages2,
					defaultLang2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResName, "name", configName),
					resource.TestCheckResourceAttr(fullResName, "description", updatedConfigDescription),
					resource.TestMatchResourceAttr(fullResName, "status", regexp.MustCompile("^(Pending|Active)$")),
					resource.TestCheckResourceAttrSet(fullResName, "version"),
					resource.TestCheckResourceAttr(fullResName, "languages.#", strconv.Itoa(len(languages2))),
					resource.TestCheckResourceAttr(fullResName, "languages.0", languages2[0]),
					resource.TestCheckResourceAttr(fullResName, "default_language", defaultLang2),
					resource.TestCheckResourceAttr(fullResName, "messenger.#", "0"),
					resource.TestCheckResourceAttr(fullResName, "cobrowse.#", "0"),
					resource.TestCheckResourceAttr(fullResName, "journey_events.#", "0"),
					resource.TestCheckResourceAttr(fullResName, "support_center.#", "0"),
				),
			},
			{
				ResourceName:            fullResName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"status"},
			},
		},
		CheckDestroy: verifyConfigurationDestroyed,
	})
}

func TestAccResourceWebDeploymentsConfigurationComplex(t *testing.T) {
	var (
		// Knowledge Base Settings
		kbResName1  = "test-kb-1"
		kbName1     = "tf-kb-" + uuid.NewString()
		kbDesc1     = "kb created for terraform test 1"
		kbCoreLang1 = "en-US"

		// Webdeployment configuration
		configName        = "Test Configuration " + util.RandString(8)
		configDescription = "Test Configuration description " + util.RandString(32)
		fullResourceName  = "genesyscloud_webdeployments_configuration.complex"

		channels       = []string{strconv.Quote("Webmessaging")}
		channelsUpdate = []string{strconv.Quote("Webmessaging"), strconv.Quote("Voice")}
	)

	cleanupWebDeploymentsConfiguration(t, "Test Configuration ")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: gcloud.GenerateKnowledgeKnowledgebaseResource(
					kbResName1,
					kbName1,
					kbDesc1,
					kbCoreLang1,
				) + complexConfigurationResource(
					configName,
					configDescription,
					"genesyscloud_knowledge_knowledgebase."+kbResName1+".id",
					generateWebDeploymentConfigCobrowseSettings(
						util.TrueValue,
						util.TrueValue,
						util.TrueValue,
						channels,
						[]string{strconv.Quote("selector-one")},
						[]string{strconv.Quote("selector-one")},
						generatePauseCriteria("/sensitive", "includes"),
						generatePauseCriteria("/login", "equals"),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", configName),
					resource.TestCheckResourceAttr(fullResourceName, "description", configDescription),
					resource.TestCheckResourceAttr(fullResourceName, "headless_mode_enabled", util.TrueValue),
					resource.TestMatchResourceAttr(fullResourceName, "status", regexp.MustCompile("^(Pending|Active)$")),
					resource.TestCheckResourceAttrSet(fullResourceName, "version"),

					resource.TestCheckResourceAttr(fullResourceName, "languages.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "languages.0", "en-us"),
					resource.TestCheckResourceAttr(fullResourceName, "languages.1", "ja"),
					resource.TestCheckResourceAttr(fullResourceName, "default_language", "en-us"),

					resource.TestCheckResourceAttr(fullResourceName, "custom_i18n_labels.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "custom_i18n_labels.0.language", "en-us"),
					resource.TestCheckResourceAttr(fullResourceName, "custom_i18n_labels.0.localized_labels.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "custom_i18n_labels.0.localized_labels.0.key", "MessengerHomeHeaderTitle"),
					resource.TestCheckResourceAttr(fullResourceName, "custom_i18n_labels.0.localized_labels.0.value", "My Messenger Home Header Title"),
					resource.TestCheckResourceAttr(fullResourceName, "custom_i18n_labels.0.localized_labels.1.key", "MessengerHomeHeaderSubTitle"),
					resource.TestCheckResourceAttr(fullResourceName, "custom_i18n_labels.0.localized_labels.1.value", "My Messenger Home Header SubTitle"),

					resource.TestCheckResourceAttr(fullResourceName, "position.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "position.0.alignment", "Auto"),
					resource.TestCheckResourceAttr(fullResourceName, "position.0.side_space", "10"),
					resource.TestCheckResourceAttr(fullResourceName, "position.0.bottom_space", "20"),

					resource.TestCheckResourceAttr(fullResourceName, "messenger.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.enabled", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.launcher_button.0.visibility", "OnDemand"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.home_screen.0.enabled", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.home_screen.0.logo_url", "https://my-domain/images/my-logo.png"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.styles.0.primary_color", "#B0B0B0"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.0.file_types.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.0.file_types.0", "image/png"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.0.max_file_size_kb", "100"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.1.file_types.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.1.file_types.0", "image/jpeg"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.1.max_file_size_kb", "123"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.0.conversations.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.0.conversations.0.enabled", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.0.conversations.0.show_agent_typing_indicator", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.0.conversations.0.show_user_typing_indicator", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.0.conversations.0.auto_start_enabled", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.0.conversations.0.markdown_enabled", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.0.conversations.0.conversation_clear_enabled", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.0.conversations.0.conversation_disconnect.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.0.conversations.0.conversation_disconnect.0.enabled", "true"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.0.conversations.0.conversation_disconnect.0.type", "Send"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.0.conversations.0.humanize.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.0.conversations.0.humanize.0.enabled", "true"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.0.conversations.0.humanize.0.bot.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.0.conversations.0.humanize.0.bot.0.name", "Marvin"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.0.conversations.0.humanize.0.bot.0.avatar_url", "https://my-domain-example.net/images/marvin.png"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.0.knowledge.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.apps.0.knowledge.0.enabled", "true"),
					resource.TestCheckResourceAttrPair(fullResourceName, "messenger.0.apps.0.knowledge.0.knowledge_base_id", "genesyscloud_knowledge_knowledgebase."+kbResName1, "id"),

					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.enabled", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.allow_agent_control", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.allow_agent_navigation", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.channels.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.channels.0", "Webmessaging"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.mask_selectors.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.mask_selectors.0", "selector-one"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.readonly_selectors.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.readonly_selectors.0", "selector-one"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.pause_criteria.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.pause_criteria.0.url_fragment", "/sensitive"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.pause_criteria.0.condition", "includes"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.pause_criteria.1.url_fragment", "/login"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.pause_criteria.1.condition", "equals"),

					resource.TestCheckResourceAttr(fullResourceName, "journey_events.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.enabled", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.excluded_query_parameters.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.excluded_query_parameters.0", "excluded-one"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.pageview_config", "Auto"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.0.selector", "first-selector"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.0.event_name", "first-click-event-name"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.1.selector", "second-selector"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.1.event_name", "second-click-event-name"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.0.selector", "form-selector-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.0.form_name", "form-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.0.capture_data_on_form_abandon", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.0.capture_data_on_form_submit", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.1.selector", "form-selector-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.1.form_name", "form-3"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.1.capture_data_on_form_abandon", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.1.capture_data_on_form_submit", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.0.event_name", "idle-event-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.0.idle_after_seconds", "88"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.1.event_name", "idle-event-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.1.idle_after_seconds", "30"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.0.selector", "in-viewport-selector-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.0.event_name", "in-viewport-event-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.1.selector", "in-viewport-selector-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.1.event_name", "in-viewport-event-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.0.event_name", "scroll-depth-event-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.0.percentage", "33"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.1.event_name", "scroll-depth-event-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.1.percentage", "66"),
				),
			},
			{
				// Update cobrowse settings
				Config: gcloud.GenerateKnowledgeKnowledgebaseResource(
					kbResName1,
					kbName1,
					kbDesc1,
					kbCoreLang1,
				) + complexConfigurationResource(
					configName,
					configDescription,
					"genesyscloud_knowledge_knowledgebase."+kbResName1+".id",
					generateWebDeploymentConfigCobrowseSettings(
						util.FalseValue,
						util.FalseValue,
						util.FalseValue,
						channelsUpdate,
						[]string{strconv.Quote("selector-one"), strconv.Quote("selector-two")},
						[]string{strconv.Quote("selector-one"), strconv.Quote("selector-two")},
						generatePauseCriteria("/sensitive", "includes"),
						generatePauseCriteria("/login", "equals"),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", configName),
					resource.TestCheckResourceAttr(fullResourceName, "description", configDescription),
					resource.TestMatchResourceAttr(fullResourceName, "status", regexp.MustCompile("^(Pending|Active)$")),
					resource.TestCheckResourceAttrSet(fullResourceName, "version"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.enabled", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.launcher_button.0.visibility", "OnDemand"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.home_screen.0.enabled", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.home_screen.0.logo_url", "https://my-domain/images/my-logo.png"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.styles.0.primary_color", "#B0B0B0"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.0.file_types.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.0.file_types.0", "image/png"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.0.max_file_size_kb", "100"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.1.file_types.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.1.file_types.0", "image/jpeg"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.1.max_file_size_kb", "123"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.enabled", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.allow_agent_control", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.allow_agent_navigation", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.channels.#", "2"),
					util.ValidateStringInArray(fullResourceName, "cobrowse.0.channels", "Webmessaging"),
					util.ValidateStringInArray(fullResourceName, "cobrowse.0.channels", "Voice"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.mask_selectors.#", "2"),
					util.ValidateStringInArray(fullResourceName, "cobrowse.0.mask_selectors", "selector-one"),
					util.ValidateStringInArray(fullResourceName, "cobrowse.0.mask_selectors", "selector-two"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.readonly_selectors.#", "2"),
					util.ValidateStringInArray(fullResourceName, "cobrowse.0.readonly_selectors", "selector-one"),
					util.ValidateStringInArray(fullResourceName, "cobrowse.0.readonly_selectors", "selector-two"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.pause_criteria.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.pause_criteria.0.url_fragment", "/sensitive"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.pause_criteria.0.condition", "includes"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.pause_criteria.1.url_fragment", "/login"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.pause_criteria.1.condition", "equals"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.enabled", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.excluded_query_parameters.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.excluded_query_parameters.0", "excluded-one"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.pageview_config", "Auto"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.0.selector", "first-selector"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.0.event_name", "first-click-event-name"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.1.selector", "second-selector"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.1.event_name", "second-click-event-name"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.0.selector", "form-selector-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.0.form_name", "form-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.0.capture_data_on_form_abandon", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.0.capture_data_on_form_submit", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.1.selector", "form-selector-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.1.form_name", "form-3"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.1.capture_data_on_form_abandon", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.1.capture_data_on_form_submit", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.0.event_name", "idle-event-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.0.idle_after_seconds", "88"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.1.event_name", "idle-event-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.1.idle_after_seconds", "30"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.0.selector", "in-viewport-selector-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.0.event_name", "in-viewport-event-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.1.selector", "in-viewport-selector-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.1.event_name", "in-viewport-event-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.0.event_name", "scroll-depth-event-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.0.percentage", "33"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.1.event_name", "scroll-depth-event-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.1.percentage", "66"),
				),
			},
			{
				ResourceName:            fullResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"status"},
			},
		},
		CheckDestroy: verifyConfigurationDestroyed,
	})
}

func TestAccResourceWebDeploymentsConfigurationSupportCenter(t *testing.T) {
	t.Parallel()

	var (
		// Knowledge Base Settings
		kbResName1  = "test-kb-1"
		kbName1     = "tf-kb-" + uuid.NewString()
		kbDesc1     = "kb created for terraform test 1"
		kbCoreLang1 = "en-US"

		kbResName2  = "test-kb-2"
		kbName2     = "tf-kb-" + uuid.NewString()
		kbDesc2     = "kb created for terraform test 2"
		kbCoreLang2 = "en-US"

		// Support center config
		resName           = "webdeploy-config-test"
		fullResName       = resourceName + "." + resName
		configName        = "tf-config-" + uuid.NewString()
		configDescription = "Test Configuration description. Support Center"
		languages         = []string{"en-us", "ja"}
		defaultLang       = "en-us"
		supportCenter1    = scConfig{
			enabled: true,
			kbId:    "genesyscloud_knowledge_knowledgebase." + kbResName1,

			customMessages: []scCustomMessageConfig{
				{
					defaultVal: "Welcome Message po",
					varType:    "Welcome",
				},
				{
					defaultVal: "Fallback Message po",
					varType:    "Fallback",
				},
			},

			routerType: "Hash",

			screens: []scScreenConfig{
				{
					varType: "Home",
					moduleSettings: []scModuleSetting{
						{
							varType:                        "Search",
							enabled:                        true,
							compactTemplateActive:          true,
							detailedTemplateActive:         true,
							detailedTemplateSidebarEnabled: true,
						},
						{
							varType:                        "Categories",
							enabled:                        true,
							compactTemplateActive:          true,
							detailedTemplateActive:         true,
							detailedTemplateSidebarEnabled: true,
						},
						{
							varType:                        "TopViewedArticles",
							enabled:                        true,
							compactTemplateActive:          true,
							detailedTemplateActive:         true,
							detailedTemplateSidebarEnabled: true,
						},
					},
				},
				{
					varType: "Category",
					moduleSettings: []scModuleSetting{
						{
							varType:                        "Search",
							enabled:                        true,
							compactTemplateActive:          true,
							detailedTemplateActive:         true,
							detailedTemplateSidebarEnabled: true,
						},
						{
							varType:                        "Categories",
							enabled:                        true,
							compactTemplateActive:          true,
							detailedTemplateActive:         true,
							detailedTemplateSidebarEnabled: true,
						},
					},
				},
				{
					varType: "SearchResults",
					moduleSettings: []scModuleSetting{
						{
							varType:                        "Search",
							enabled:                        true,
							compactTemplateActive:          true,
							detailedTemplateActive:         true,
							detailedTemplateSidebarEnabled: true,
						},
						{
							varType:                        "Results",
							enabled:                        true,
							compactTemplateActive:          true,
							detailedTemplateActive:         true,
							detailedTemplateSidebarEnabled: true,
						},
					},
				},
				{
					varType: "Article",
					moduleSettings: []scModuleSetting{
						{
							varType:                        "Search",
							enabled:                        true,
							compactTemplateActive:          true,
							detailedTemplateActive:         true,
							detailedTemplateSidebarEnabled: true,
						},
						{
							varType:                        "Article",
							enabled:                        true,
							compactTemplateActive:          true,
							detailedTemplateActive:         true,
							detailedTemplateSidebarEnabled: true,
						},
					},
				},
			},

			styleSetting: scStyleSettingConfig{
				heroStyleSetting: heroStyleSettingConfig{
					bgColor:   "#000000",
					textColor: "#FFFFFF",
					imageUri:  "https://hero.hero.com/hero.png",
				},

				globalStyleSetting: globalStyleSettingConfig{
					bgColor:           "#000000",
					primaryColor:      "#FFFFFF",
					primaryColorDark:  "#111111",
					primaryColorLight: "#EEEEEE",
					textColor:         "#222222",
					fontFamily:        "Arial",
				},
			},

			feedbackEnabled: true,
		}

		// updated attributes
		supportCenter2 = scConfig{
			enabled: true,
			kbId:    "genesyscloud_knowledge_knowledgebase." + kbResName2,

			customMessages: []scCustomMessageConfig{
				{
					defaultVal: "Welcome Message 2",
					varType:    "Welcome",
				},
				{
					defaultVal: "Fallback Message 2",
					varType:    "Fallback",
				},
			},

			routerType: "Browser",

			screens: []scScreenConfig{
				{
					varType: "Home",
					moduleSettings: []scModuleSetting{
						{
							varType:                        "Search",
							enabled:                        false,
							compactTemplateActive:          false,
							detailedTemplateActive:         false,
							detailedTemplateSidebarEnabled: false,
						},
						{
							varType:                        "Categories",
							enabled:                        false,
							compactTemplateActive:          false,
							detailedTemplateActive:         false,
							detailedTemplateSidebarEnabled: false,
						},
						{
							varType:                        "TopViewedArticles",
							enabled:                        false,
							compactTemplateActive:          false,
							detailedTemplateActive:         false,
							detailedTemplateSidebarEnabled: false,
						},
					},
				},
				{
					varType: "Category",
					moduleSettings: []scModuleSetting{
						{
							varType:                        "Search",
							enabled:                        false,
							compactTemplateActive:          false,
							detailedTemplateActive:         false,
							detailedTemplateSidebarEnabled: false,
						},
						{
							varType:                        "Categories",
							enabled:                        false,
							compactTemplateActive:          false,
							detailedTemplateActive:         false,
							detailedTemplateSidebarEnabled: false,
						},
					},
				},
				{
					varType: "SearchResults",
					moduleSettings: []scModuleSetting{
						{
							varType:                        "Search",
							enabled:                        false,
							compactTemplateActive:          false,
							detailedTemplateActive:         false,
							detailedTemplateSidebarEnabled: false,
						},
						{
							varType:                        "Results",
							enabled:                        false,
							compactTemplateActive:          false,
							detailedTemplateActive:         false,
							detailedTemplateSidebarEnabled: false,
						},
					},
				},
				{
					varType: "Article",
					moduleSettings: []scModuleSetting{
						{
							varType:                        "Search",
							enabled:                        false,
							compactTemplateActive:          false,
							detailedTemplateActive:         false,
							detailedTemplateSidebarEnabled: false,
						},
						{
							varType:                        "Article",
							enabled:                        false,
							compactTemplateActive:          false,
							detailedTemplateActive:         false,
							detailedTemplateSidebarEnabled: false,
						},
					},
				},
			},

			styleSetting: scStyleSettingConfig{
				heroStyleSetting: heroStyleSettingConfig{
					bgColor:   "#000001",
					textColor: "#FFFFFE",
					imageUri:  "https://hero.hero.com/hero2.png",
				},

				globalStyleSetting: globalStyleSettingConfig{
					bgColor:           "#000001",
					primaryColor:      "#FFFFFE",
					primaryColorDark:  "#111112",
					primaryColorLight: "#EEEEEF",
					textColor:         "#222223",
					fontFamily:        "Arial",
				},
			},

			feedbackEnabled: false,
		}
	)

	cleanupWebDeploymentsConfiguration(t, "Test Configuration ")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: gcloud.GenerateKnowledgeKnowledgebaseResource(
					kbResName1,
					kbName1,
					kbDesc1,
					kbCoreLang1,
				) + generateConfigurationResource(
					resName,
					configName,
					configDescription,
					languages,
					defaultLang,
					generateSupportCenterSettings(supportCenter1),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResName, "name", configName),
					resource.TestCheckResourceAttr(fullResName, "description", configDescription),
					resource.TestMatchResourceAttr(fullResName, "status", regexp.MustCompile("^(Pending|Active)$")),
					resource.TestCheckResourceAttrSet(fullResName, "version"),
					resource.TestCheckResourceAttr(fullResName, "languages.#", strconv.Itoa(len(languages))),
					resource.TestCheckResourceAttr(fullResName, "languages.0", languages[0]),
					resource.TestCheckResourceAttr(fullResName, "languages.1", languages[1]),
					resource.TestCheckResourceAttr(fullResName, "default_language", defaultLang),
					resource.TestCheckResourceAttr(fullResName, "messenger.#", "0"),
					resource.TestCheckResourceAttr(fullResName, "cobrowse.#", "0"),
					resource.TestCheckResourceAttr(fullResName, "journey_events.#", "0"),
					resource.TestCheckResourceAttr(fullResName, "support_center.#", "1"),

					resource.TestCheckResourceAttr(fullResName, "support_center.0.enabled", strconv.FormatBool(supportCenter1.enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.knowledge_base_id", supportCenter1.kbId),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.custom_messages.#", strconv.Itoa(len(supportCenter1.customMessages))),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.custom_messages.0.default_value", supportCenter1.customMessages[0].defaultVal),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.custom_messages.0.type", supportCenter1.customMessages[0].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.custom_messages.1.default_value", supportCenter1.customMessages[1].defaultVal),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.custom_messages.1.type", supportCenter1.customMessages[1].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.router_type", supportCenter1.routerType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.#", strconv.Itoa(len(supportCenter1.screens))),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.type", supportCenter1.screens[0].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.#", strconv.Itoa(len(supportCenter1.screens[0].moduleSettings))),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.0.type", supportCenter1.screens[0].moduleSettings[0].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.0.enabled", strconv.FormatBool(supportCenter1.screens[0].moduleSettings[0].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.0.compact_category_module_template_active", strconv.FormatBool(supportCenter1.screens[0].moduleSettings[0].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.0.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter1.screens[0].moduleSettings[0].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.0.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter1.screens[0].moduleSettings[0].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.1.type", supportCenter1.screens[0].moduleSettings[1].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.1.enabled", strconv.FormatBool(supportCenter1.screens[0].moduleSettings[1].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.1.compact_category_module_template_active", strconv.FormatBool(supportCenter1.screens[0].moduleSettings[1].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.1.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter1.screens[0].moduleSettings[1].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.1.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter1.screens[0].moduleSettings[1].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.2.type", supportCenter1.screens[0].moduleSettings[2].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.2.enabled", strconv.FormatBool(supportCenter1.screens[0].moduleSettings[2].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.2.compact_category_module_template_active", strconv.FormatBool(supportCenter1.screens[0].moduleSettings[2].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.2.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter1.screens[0].moduleSettings[2].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.2.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter1.screens[0].moduleSettings[2].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.type", supportCenter1.screens[1].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.#", strconv.Itoa(len(supportCenter1.screens[1].moduleSettings))),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.0.type", supportCenter1.screens[1].moduleSettings[0].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.0.enabled", strconv.FormatBool(supportCenter1.screens[1].moduleSettings[0].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.0.compact_category_module_template_active", strconv.FormatBool(supportCenter1.screens[1].moduleSettings[0].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.0.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter1.screens[1].moduleSettings[0].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.0.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter1.screens[1].moduleSettings[0].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.1.type", supportCenter1.screens[1].moduleSettings[1].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.1.enabled", strconv.FormatBool(supportCenter1.screens[1].moduleSettings[1].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.1.compact_category_module_template_active", strconv.FormatBool(supportCenter1.screens[1].moduleSettings[1].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.1.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter1.screens[1].moduleSettings[1].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.1.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter1.screens[1].moduleSettings[1].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.type", supportCenter1.screens[2].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.#", strconv.Itoa(len(supportCenter1.screens[2].moduleSettings))),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.0.type", supportCenter1.screens[2].moduleSettings[0].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.0.enabled", strconv.FormatBool(supportCenter1.screens[2].moduleSettings[0].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.0.compact_category_module_template_active", strconv.FormatBool(supportCenter1.screens[2].moduleSettings[0].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.0.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter1.screens[2].moduleSettings[0].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.0.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter1.screens[2].moduleSettings[0].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.1.type", supportCenter1.screens[2].moduleSettings[1].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.1.enabled", strconv.FormatBool(supportCenter1.screens[2].moduleSettings[1].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.1.compact_category_module_template_active", strconv.FormatBool(supportCenter1.screens[2].moduleSettings[1].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.1.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter1.screens[2].moduleSettings[1].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.1.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter1.screens[2].moduleSettings[1].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.type", supportCenter1.screens[3].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.#", strconv.Itoa(len(supportCenter1.screens[3].moduleSettings))),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.0.type", supportCenter1.screens[3].moduleSettings[0].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.0.enabled", strconv.FormatBool(supportCenter1.screens[3].moduleSettings[0].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.0.compact_category_module_template_active", strconv.FormatBool(supportCenter1.screens[3].moduleSettings[0].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.0.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter1.screens[3].moduleSettings[0].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.0.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter1.screens[3].moduleSettings[0].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.1.type", supportCenter1.screens[3].moduleSettings[1].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.1.enabled", strconv.FormatBool(supportCenter1.screens[3].moduleSettings[1].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.1.compact_category_module_template_active", strconv.FormatBool(supportCenter1.screens[3].moduleSettings[1].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.1.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter1.screens[3].moduleSettings[1].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.1.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter1.screens[3].moduleSettings[1].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.hero_style_setting.0.background_color", supportCenter1.styleSetting.heroStyleSetting.bgColor),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.hero_style_setting.0.text_color", supportCenter1.styleSetting.heroStyleSetting.textColor),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.hero_style_setting.0.image_uri", supportCenter1.styleSetting.heroStyleSetting.imageUri),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.global_style_setting.0.background_color", supportCenter1.styleSetting.globalStyleSetting.bgColor),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.global_style_setting.0.primary_color", supportCenter1.styleSetting.globalStyleSetting.primaryColor),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.global_style_setting.0.primary_color_dark", supportCenter1.styleSetting.globalStyleSetting.primaryColorDark),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.global_style_setting.0.primary_color_light", supportCenter1.styleSetting.globalStyleSetting.primaryColorLight),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.global_style_setting.0.text_color", supportCenter1.styleSetting.globalStyleSetting.textColor),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.global_style_setting.0.font_family", supportCenter1.styleSetting.globalStyleSetting.fontFamily),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.feedback_enabled", strconv.FormatBool(supportCenter1.feedbackEnabled)),
				),
			},
			{
				Config: gcloud.GenerateKnowledgeKnowledgebaseResource(
					kbResName2,
					kbName2,
					kbDesc2,
					kbCoreLang2,
				) + generateConfigurationResource(
					resName,
					configName,
					configDescription,
					languages,
					defaultLang,
					generateSupportCenterSettings(supportCenter2),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResName, "support_center.0.enabled", strconv.FormatBool(supportCenter2.enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.knowledge_base_id", supportCenter2.kbId),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.custom_messages.#", strconv.Itoa(len(supportCenter2.customMessages))),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.custom_messages.0.default_value", supportCenter2.customMessages[0].defaultVal),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.custom_messages.0.type", supportCenter2.customMessages[0].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.custom_messages.1.default_value", supportCenter2.customMessages[1].defaultVal),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.custom_messages.1.type", supportCenter2.customMessages[1].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.router_type", supportCenter2.routerType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.#", strconv.Itoa(len(supportCenter2.screens))),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.type", supportCenter2.screens[0].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.#", strconv.Itoa(len(supportCenter2.screens[0].moduleSettings))),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.0.type", supportCenter2.screens[0].moduleSettings[0].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.0.enabled", strconv.FormatBool(supportCenter2.screens[0].moduleSettings[0].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.0.compact_category_module_template_active", strconv.FormatBool(supportCenter2.screens[0].moduleSettings[0].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.0.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter2.screens[0].moduleSettings[0].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.0.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter2.screens[0].moduleSettings[0].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.1.type", supportCenter2.screens[0].moduleSettings[1].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.1.enabled", strconv.FormatBool(supportCenter2.screens[0].moduleSettings[1].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.1.compact_category_module_template_active", strconv.FormatBool(supportCenter2.screens[0].moduleSettings[1].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.1.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter2.screens[0].moduleSettings[1].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.1.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter2.screens[0].moduleSettings[1].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.2.type", supportCenter2.screens[0].moduleSettings[2].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.2.enabled", strconv.FormatBool(supportCenter2.screens[0].moduleSettings[2].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.2.compact_category_module_template_active", strconv.FormatBool(supportCenter2.screens[0].moduleSettings[2].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.2.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter2.screens[0].moduleSettings[2].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.0.module_settings.2.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter2.screens[0].moduleSettings[2].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.type", supportCenter2.screens[1].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.#", strconv.Itoa(len(supportCenter2.screens[1].moduleSettings))),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.0.type", supportCenter2.screens[1].moduleSettings[0].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.0.enabled", strconv.FormatBool(supportCenter2.screens[1].moduleSettings[0].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.0.compact_category_module_template_active", strconv.FormatBool(supportCenter2.screens[1].moduleSettings[0].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.0.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter2.screens[1].moduleSettings[0].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.0.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter2.screens[1].moduleSettings[0].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.1.type", supportCenter2.screens[1].moduleSettings[1].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.1.enabled", strconv.FormatBool(supportCenter2.screens[1].moduleSettings[1].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.1.compact_category_module_template_active", strconv.FormatBool(supportCenter2.screens[1].moduleSettings[1].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.1.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter2.screens[1].moduleSettings[1].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.1.module_settings.1.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter2.screens[1].moduleSettings[1].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.type", supportCenter2.screens[2].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.#", strconv.Itoa(len(supportCenter2.screens[2].moduleSettings))),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.0.type", supportCenter2.screens[2].moduleSettings[0].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.0.enabled", strconv.FormatBool(supportCenter2.screens[2].moduleSettings[0].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.0.compact_category_module_template_active", strconv.FormatBool(supportCenter2.screens[2].moduleSettings[0].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.0.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter2.screens[2].moduleSettings[0].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.0.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter2.screens[2].moduleSettings[0].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.1.type", supportCenter2.screens[2].moduleSettings[1].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.1.enabled", strconv.FormatBool(supportCenter2.screens[2].moduleSettings[1].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.1.compact_category_module_template_active", strconv.FormatBool(supportCenter2.screens[2].moduleSettings[1].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.1.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter2.screens[2].moduleSettings[1].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.2.module_settings.1.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter2.screens[2].moduleSettings[1].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.type", supportCenter2.screens[3].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.#", strconv.Itoa(len(supportCenter2.screens[3].moduleSettings))),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.0.type", supportCenter2.screens[3].moduleSettings[0].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.0.enabled", strconv.FormatBool(supportCenter2.screens[3].moduleSettings[0].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.0.compact_category_module_template_active", strconv.FormatBool(supportCenter2.screens[3].moduleSettings[0].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.0.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter2.screens[3].moduleSettings[0].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.0.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter2.screens[3].moduleSettings[0].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.1.type", supportCenter2.screens[3].moduleSettings[1].varType),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.1.enabled", strconv.FormatBool(supportCenter2.screens[3].moduleSettings[1].enabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.1.compact_category_module_template_active", strconv.FormatBool(supportCenter2.screens[3].moduleSettings[1].compactTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.1.detailed_category_module_template.0.active", strconv.FormatBool(supportCenter2.screens[3].moduleSettings[1].detailedTemplateActive)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.screens.3.module_settings.1.detailed_category_module_template.0.sidebar_enabled", strconv.FormatBool(supportCenter2.screens[3].moduleSettings[1].detailedTemplateSidebarEnabled)),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.hero_style_setting.0.background_color", supportCenter2.styleSetting.heroStyleSetting.bgColor),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.hero_style_setting.0.text_color", supportCenter2.styleSetting.heroStyleSetting.textColor),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.hero_style_setting.0.image_uri", supportCenter2.styleSetting.heroStyleSetting.imageUri),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.global_style_setting.0.background_color", supportCenter2.styleSetting.globalStyleSetting.bgColor),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.global_style_setting.0.primary_color", supportCenter2.styleSetting.globalStyleSetting.primaryColor),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.global_style_setting.0.primary_color_dark", supportCenter2.styleSetting.globalStyleSetting.primaryColorDark),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.global_style_setting.0.primary_color_light", supportCenter2.styleSetting.globalStyleSetting.primaryColorLight),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.global_style_setting.0.text_color", supportCenter2.styleSetting.globalStyleSetting.textColor),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.style_setting.0.global_style_setting.0.font_family", supportCenter2.styleSetting.globalStyleSetting.fontFamily),
					resource.TestCheckResourceAttr(fullResName, "support_center.0.feedback_enabled", strconv.FormatBool(supportCenter2.feedbackEnabled)),
				),
			},
			{
				ResourceName:            fullResName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"status"},
			},
		},
		CheckDestroy: verifyConfigurationDestroyed,
	})
}

func generateConfigurationResource(resName, configName, description string, languages []string, defaultLang string, nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_webdeployments_configuration" "%s" {
		name = "%s"
		description = "%s"
		languages        = %s
		default_language = "%s"
		%s
	}
	`,
		resName,
		configName,
		description,
		util.GenerateStringArrayEnquote(languages...),
		defaultLang,
		strings.Join(nestedBlocks, "\n"))
}

func basicConfigurationResource(name, description string) string {
	return fmt.Sprintf(`
	resource "genesyscloud_webdeployments_configuration" "basic" {
		name             = "%s"
		description      = "%s"
		languages        = [ "en-us", "ja" ]
		default_language = "en-us"
	}
	`, name, description)
}

func complexConfigurationResource(name, description, kbId string, nestedBlocks ...string) string {
	return fmt.Sprintf(`
	resource "genesyscloud_webdeployments_configuration" "complex" {
		name = "%s"
		description = "%s"
		languages = [ "en-us", "ja" ]
		default_language = "en-us"
		headless_mode_enabled = true
		custom_i18n_labels {
			language = "en-us"
			localized_labels {
				key = "MessengerHomeHeaderTitle"
				value = "My Messenger Home Header Title"
			}
			localized_labels {
				key = "MessengerHomeHeaderSubTitle"
				value = "My Messenger Home Header SubTitle"
			}
		}
		position {
			alignment = "Auto"
			side_space = 10
			bottom_space = 20
		}
		messenger {
			enabled = true
			launcher_button {
				visibility = "OnDemand"
			}
			home_screen {
				enabled = true
				logo_url = "https://my-domain/images/my-logo.png"
			}
			styles {
				primary_color = "#B0B0B0"
			}
			file_upload {
				mode {
					file_types = [ "image/png" ]
					max_file_size_kb = 100
				}
				mode {
					file_types = [ "image/jpeg" ]
					max_file_size_kb = 123
				}
			}
			apps {
				conversations {
					enabled = true
					show_agent_typing_indicator = true
					show_user_typing_indicator = true
					auto_start_enabled = true
					markdown_enabled = true
					conversation_disconnect {
						enabled = true
						type = "Send"
					}
					conversation_clear_enabled = true
					humanize {
						enabled = true
						bot {
							name = "Marvin"
							avatar_url = "https://my-domain-example.net/images/marvin.png"
						}
					}
				}
				knowledge {
					enabled = true
					knowledge_base_id = %s
				}
			}
		}
		journey_events {
			enabled = true
			excluded_query_parameters = [ "excluded-one" ]

			pageview_config = "Auto"

			click_event {
				selector = "first-selector"
				event_name = "first-click-event-name"
			}
			click_event {
				selector = "second-selector"
				event_name = "second-click-event-name"
			}

			form_track_event {
				selector = "form-selector-1"
				form_name = "form-1"
				capture_data_on_form_abandon = true
				capture_data_on_form_submit = false
			}

			form_track_event {
				selector = "form-selector-2"
				form_name = "form-3"
				capture_data_on_form_abandon = false
				capture_data_on_form_submit = true
			}

			idle_event {
				event_name = "idle-event-1"
				idle_after_seconds = 88
			}

			idle_event {
				event_name = "idle-event-2"
				idle_after_seconds = 30
			}

			in_viewport_event {
				selector = "in-viewport-selector-1"
				event_name = "in-viewport-event-1"
			}

			in_viewport_event {
				selector = "in-viewport-selector-2"
				event_name = "in-viewport-event-2"
			}

			scroll_depth_event {
				event_name = "scroll-depth-event-1"
				percentage = 33
			}

			scroll_depth_event {
				event_name = "scroll-depth-event-2"
				percentage = 66
			}
		}
		%s
	}
	`, name, description, kbId, strings.Join(nestedBlocks, "\n"))
}

func generateWebDeploymentConfigCobrowseSettings(cbEnabled, cbAllowAgentControl string, cbAllowAgentNavigation string, cbChannels []string, cbMaskSelectors []string, cbReadonlySelectors []string, pauseCriteriaBlocks ...string) string {

	return fmt.Sprintf(`
	cobrowse {
		enabled = %s
		allow_agent_control = %s
		allow_agent_navigation = %s
		channels = [ %s ]
		mask_selectors = [ %s ]
		readonly_selectors = [ %s ]
		%s
	}
`, cbEnabled, cbAllowAgentControl, cbAllowAgentNavigation, strings.Join(cbChannels, ", "), strings.Join(cbMaskSelectors, ", "), strings.Join(cbReadonlySelectors, ", "), strings.Join(pauseCriteriaBlocks, "\n"))
}

func generatePauseCriteria(urlFragment, condition string) string {
	return fmt.Sprintf(`pause_criteria {
	url_fragment = "%s"
	condition = "%s"
}`, urlFragment, condition)
}

func generateSupportCenterSettings(supportCenter scConfig) string {
	var customMessages []string
	for _, customMessage := range supportCenter.customMessages {
		customMessages = append(customMessages, fmt.Sprintf(`
		custom_messages {
			default_value = "%s"
			type = "%s"
		}
		`, customMessage.defaultVal, customMessage.varType))
	}

	var screens []string
	for _, screen := range supportCenter.screens {
		var moduleSettings []string
		for _, moduleSetting := range screen.moduleSettings {
			moduleSettings = append(moduleSettings, fmt.Sprintf(`
			module_settings {
				type = "%s"
				enabled = %s
				compact_category_module_template_active = %s
				detailed_category_module_template {
					active = %s
					sidebar_enabled = %s
				}
			}
			`, moduleSetting.varType,
				strconv.FormatBool(moduleSetting.enabled),
				strconv.FormatBool(moduleSetting.compactTemplateActive),
				strconv.FormatBool(moduleSetting.detailedTemplateActive),
				strconv.FormatBool(moduleSetting.detailedTemplateSidebarEnabled)),
			)
		}
		screens = append(screens, fmt.Sprintf(`
		screens {
			type = "%s"
			%s
		}
		`, screen.varType, strings.Join(moduleSettings, "\n")))
	}

	var enabledCategories []string
	for _, enabledCategory := range supportCenter.enabledCategories {
		enabledCategories = append(enabledCategories, fmt.Sprintf(`
		enabled_categories {
			category_id = "%s"
			image_uri = "%s"
		}
		`, enabledCategory.categoryId, enabledCategory.imageUri))
	}

	styleSetting := fmt.Sprintf(`
	style_setting {
		hero_style_setting {
			background_color = "%s"
			text_color = "%s"
			image_uri = "%s"
		}
		global_style_setting {
			background_color = "%s"
			primary_color = "%s"
			primary_color_dark = "%s"
			primary_color_light = "%s"
			text_color = "%s"
			font_family = "%s"
		}
	}
	`, supportCenter.styleSetting.heroStyleSetting.bgColor,
		supportCenter.styleSetting.heroStyleSetting.textColor,
		supportCenter.styleSetting.heroStyleSetting.imageUri,
		supportCenter.styleSetting.globalStyleSetting.bgColor,
		supportCenter.styleSetting.globalStyleSetting.primaryColor,
		supportCenter.styleSetting.globalStyleSetting.primaryColorDark,
		supportCenter.styleSetting.globalStyleSetting.primaryColorLight,
		supportCenter.styleSetting.globalStyleSetting.textColor,
		supportCenter.styleSetting.globalStyleSetting.fontFamily,
	)

	return fmt.Sprintf(`
	support_center {
		enabled = %s
		knowledge_base_id = "%s"
		router_type = "%s"
		%s
		%s
		%s
		%s
		feedback_enabled = %s
	}
	`, strconv.FormatBool(supportCenter.enabled),
		supportCenter.kbId,
		supportCenter.routerType,
		strings.Join(customMessages, "\n"),
		strings.Join(screens, "\n"),
		strings.Join(enabledCategories, "\n"),
		styleSetting,
		strconv.FormatBool(supportCenter.feedbackEnabled),
	)
}

func verifyConfigurationDestroyed(state *terraform.State) error {
	api := platformclientv2.NewWebDeploymentsApi()

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_webdeployments_configuration" {
			continue
		}

		_, response, err := api.GetWebdeploymentsConfigurationVersionsDraft(rs.Primary.ID)

		if util.IsStatus404(response) {
			continue
		}

		if err != nil {
			return fmt.Errorf("Unexpected error while checking that configuration has been destroyed: %s", err)
		}

		return fmt.Errorf("Configuration %s still exists when it was expected to have been destroyed", rs.Primary.ID)
	}

	return nil
}

func cleanupWebDeploymentsConfiguration(t *testing.T, prefix string) {
	config, err := provider.AuthorizeSdk()
	if err != nil {
		t.Logf("Failed to authorize SDK: %s", err)
		return
	}
	deploymentsAPI := platformclientv2.NewWebDeploymentsApiWithConfig(config)

	configurations, resp, getErr := deploymentsAPI.GetWebdeploymentsConfigurations(false)
	if getErr != nil {
		t.Logf("failed to get page of configurations: %v %v", getErr, resp)
		return
	}

	for _, configuration := range *configurations.Entities {
		if configuration.Name != nil && strings.HasPrefix(*configuration.Name, prefix) {
			resp, delErr := deploymentsAPI.DeleteWebdeploymentsConfiguration(*configuration.Id)
			if delErr != nil {
				t.Logf("Failed to delete configuration %s: %s %v", *configuration.Id, delErr, resp)
			}
			time.Sleep(5 * time.Second)
		}
	}
}
