package speechandtextanalytics_sentimentfeedback

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesyscloud_speechandtextanalytics_sentimentfeedback_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the speechandtextanalytics_sentimentfeedback resource.
3.  The datasource schema definitions for the speechandtextanalytics_sentimentfeedback datasource.
4.  The resource exporter configuration for the speechandtextanalytics_sentimentfeedback exporter.
*/
const ResourceType = "genesyscloud_speechandtextanalytics_sentimentfeedback"

// Valid sentiment feedback values
const (
	FeedbackValuePositive = "Positive"
	FeedbackValueNegative = "Negative"
	FeedbackValueNeutral  = "Neutral"
)

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceSentimentFeedback())
	regInstance.RegisterDataSource(ResourceType, DataSourceSentimentFeedback())
	regInstance.RegisterExporter(ResourceType, SentimentFeedbackExporter())
}

// ResourceSentimentFeedback registers the genesyscloud_speechandtextanalytics_sentimentfeedback resource with Terraform
func ResourceSentimentFeedback() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Speech & Text Analytics Sentiment Feedback. The API does not support updating an existing sentiment feedback, so any change to a configured attribute forces the resource to be recreated.`,

		CreateContext: provider.CreateWithPooledClient(createSentimentFeedback),
		ReadContext:   provider.ReadWithPooledClient(readSentimentFeedback),
		DeleteContext: provider.DeleteWithPooledClient(deleteSentimentFeedback),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"phrase": {
				Description: `The phrase for which sentiment feedback is provided.`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"dialect": {
				Description: `The dialect for the given phrase, dialect format is {language}-{country} where language follows ISO 639-1 standard and country follows ISO 3166-1 alpha 2 standard.`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"feedback_value": {
				Description:  `The sentiment feedback value for the given phrase. Valid values: Positive, Negative, Neutral.`,
				Required:     true,
				ForceNew:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{FeedbackValuePositive, FeedbackValueNegative, FeedbackValueNeutral}, false),
			},
		},
	}
}

// SentimentFeedbackExporter returns the resourceExporter object used to hold the genesyscloud_speechandtextanalytics_sentimentfeedback exporter's config
func SentimentFeedbackExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllSentimentFeedbacks),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
	}
}

// DataSourceSentimentFeedback registers the genesyscloud_speechandtextanalytics_sentimentfeedback data source
func DataSourceSentimentFeedback() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Speech & Text Analytics Sentiment Feedback data source. Select a sentiment feedback by phrase and dialect.`,
		ReadContext: provider.ReadWithPooledClient(dataSourceSentimentFeedbackRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"phrase": {
				Description: `The phrase of the sentiment feedback.`,
				Type:        schema.TypeString,
				Required:    true,
			},
			"dialect": {
				Description: `The dialect of the sentiment feedback. Recommended when the same phrase exists for multiple dialects.`,
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}
