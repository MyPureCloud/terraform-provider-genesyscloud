package speechandtextanalytics_dictionaryfeedback

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_speechandtextanalytics_dictionaryfeedback_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the speechandtextanalytics_dictionaryfeedback resource.
3.  The datasource schema definitions for the speechandtextanalytics_dictionaryfeedback datasource.
4.  The resource exporter configuration for the speechandtextanalytics_dictionaryfeedback exporter.
*/
const ResourceType = "genesyscloud_speechandtextanalytics_dictionaryfeedback"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceDictionaryFeedback())
	regInstance.RegisterDataSource(ResourceType, DataSourceDictionaryFeedback())
	regInstance.RegisterExporter(ResourceType, DictionaryFeedbackExporter())
}

// ResourceDictionaryFeedback registers the genesyscloud_speechandtextanalytics_dictionaryfeedback resource with Terraform
func ResourceDictionaryFeedback() *schema.Resource {
	dictionaryFeedbackExamplePhraseResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`phrase`: {
				Description: `The Example Phrase text. At least 3 words and up to 20 words`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`source`: {
				Description: `The source of the given Example Phrase`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud dictionary feedback`,

		CreateContext: provider.CreateWithPooledClient(createDictionaryFeedback),
		ReadContext:   provider.ReadWithPooledClient(readDictionaryFeedback),
		UpdateContext: provider.UpdateWithPooledClient(updateDictionaryFeedback),
		DeleteContext: provider.DeleteWithPooledClient(deleteDictionaryFeedback),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`term`: {
				Description: `The dictionary term which needs to be added to dictionary feedback system`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`dialect`: {
				Description: `The dialect for the given term, dialect format is {language}-{country} where language follows ISO 639-1 standard and country follows ISO 3166-1 alpha 2 standard`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`boost_value`: {
				Description: `A weighted value assigned to a phrase. The higher the value, the higher the likelihood that the system will choose the word or phrase from the possible alternatives. Boost range is from 1.0 to 10.0. Default is 2.0`,
				Optional:    true,
				Type:        schema.TypeFloat,
			},
			`source`: {
				Description: `The source of the given dictionary feedback`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`example_phrases`: {
				Description: `A list of at least 3 and up to 20 unique phrases that are example usage of the term`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        dictionaryFeedbackExamplePhraseResource,
			},
			`sounds_like`: {
				Description: `A list of up to 10 terms that give examples of how the term sounds`,
				Optional:    true,
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// DictionaryFeedbackExporter returns the resourceExporter object used to hold the genesyscloud_speechandtextanalytics_dictionaryfeedback exporter's config
func DictionaryFeedbackExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthDictionaryFeedbacks),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{
			// TODO: Add any reference attributes here
		},
	}
}

// DataSourceDictionaryFeedback registers the genesyscloud_speechandtextanalytics_dictionaryfeedback data source
func DataSourceDictionaryFeedback() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud dictionary feedback data source. Select an dictionary feedback by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceDictionaryFeedbackRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `dictionary feedback name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
