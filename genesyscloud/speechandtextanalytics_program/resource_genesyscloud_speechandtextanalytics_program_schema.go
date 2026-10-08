package speechandtextanalytics_program

// @team: PureCloud Speech & Text Analytics
// @jira: GIA
// @description: Manage Speech & Text Analytics Programs. Programs group topics together and back the Programs area under the Quality section.
// Programs can reference topics (genesyscloud_speechandtextanalytics_topic) via the topic_ids attribute.

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_speechandtextanalytics_program"

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceSpeechAndTextAnalyticsProgram())
	regInstance.RegisterDataSource(ResourceType, DataSourceSpeechAndTextAnalyticsProgram())
	regInstance.RegisterExporter(ResourceType, SpeechAndTextAnalyticsProgramExporter())
}

func ResourceSpeechAndTextAnalyticsProgram() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Speech & Text Analytics Program. Set published = true to publish the program via the publish-job API. Publishing is one-way; setting published = false after a program is published has no effect.",
		CreateContext: provider.CreateWithPooledClient(createProgram),
		ReadContext:   provider.ReadWithPooledClient(readProgram),
		UpdateContext: provider.UpdateWithPooledClient(updateProgram),
		DeleteContext: provider.DeleteWithPooledClient(deleteProgram),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The program name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The program description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"topic_ids": {
				Description: "The IDs of topics associated to the program. Topics are managed by the genesyscloud_speechandtextanalytics_topic resource.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"tags": {
				Description: "The program tags.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"published": {
				Description: "Whether the program is published. Set to true to publish the program via the publish-job API. Publishing is one-way; setting this back to false after publishing has no effect.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				// Publishing is one-way. Once a program is published the API always returns true,
				// so suppress plan drift when config tries to move published from true back to false.
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == "true" && new == "false"
				},
			},
		},
	}
}

func SpeechAndTextAnalyticsProgramExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllPrograms),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"topic_ids": {RefType: "genesyscloud_speechandtextanalytics_topic"},
		},
	}
}

func DataSourceSpeechAndTextAnalyticsProgram() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Speech & Text Analytics Programs. Select a program by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceSpeechAndTextAnalyticsProgramRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Program name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
