package speechandtextanalytics_topic

// @team: PureCloud Speech & Text Analytics
// @jira: GIA
// @description: Manage Speech & Text Analytics Topics. These topic IDs can be referenced from Quality Evaluation Forms assistance conditions.

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_speechandtextanalytics_topic"

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceSpeechAndTextAnalyticsTopic())
	regInstance.RegisterDataSource(ResourceType, DataSourceSpeechAndTextAnalyticsTopic())
	regInstance.RegisterExporter(ResourceType, SpeechAndTextAnalyticsTopicExporter())
}

func ResourceSpeechAndTextAnalyticsTopic() *schema.Resource {
	phraseResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"text": {
				Description: "The phrase text.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"strictness": {
				Description:  "The phrase strictness. Valid values: 1, 55, 65, 72, 85, 90.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"1", "55", "65", "72", "85", "90"}, false),
			},
			"sentiment": {
				Description:  "The phrase sentiment. Valid values: Unspecified, Positive, Neutral, Negative.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Unspecified", "Positive", "Neutral", "Negative"}, false),
			},
		},
	}

	return &schema.Resource{
		Description:   "Genesys Cloud Speech & Text Analytics Topic.",
		CreateContext: provider.CreateWithPooledClient(createTopic),
		ReadContext:   provider.ReadWithPooledClient(readTopic),
		UpdateContext: provider.UpdateWithPooledClient(updateTopic),
		DeleteContext: provider.DeleteWithPooledClient(deleteTopic),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The topic name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"dialect": {
				Description: "The topic dialect, e.g. en-US.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The topic description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"strictness": {
				Description:  "The topic strictness. Valid values: 1, 55, 65, 72, 85, 90.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "72",
				ValidateFunc: validation.StringInSlice([]string{"1", "55", "65", "72", "85", "90"}, false),
			},
			"participants": {
				Description:  "Which participants to match. Valid values: External, Internal, All.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "All",
				ValidateFunc: validation.StringInSlice([]string{"External", "Internal", "All"}, false),
			},
			"program_ids": {
				Description: "The IDs of programs associated to the topic.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"tags": {
				Description: "The topic tags.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"phrases": {
				Description: "The topic phrases.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        phraseResource,
			},
			"published": {
				Description: "Whether the topic is published. Assisted QM topic validation may require published topics.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func SpeechAndTextAnalyticsTopicExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllTopics),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func DataSourceSpeechAndTextAnalyticsTopic() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Speech & Text Analytics Topics. Select a topic by name and dialect.",
		ReadContext: provider.ReadWithPooledClient(dataSourceSpeechAndTextAnalyticsTopicRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Topic name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"dialect": {
				Description: "Topic dialect, e.g. en-US.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
