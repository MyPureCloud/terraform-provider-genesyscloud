package knowledge_knowledgebase

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceKnowledgeKnowledgebase())
	l.RegisterDataSource(ResourceType, dataSourceKnowledgeKnowledgebase())
	l.RegisterExporter(ResourceType, KnowledgeKnowledgebaseExporter())
}

func KnowledgeKnowledgebaseExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllKnowledgeKnowledgebases),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

const ResourceType = "genesyscloud_knowledge_knowledgebase"

func ResourceKnowledgeKnowledgebase() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Knowledge Base",

		CreateContext: provider.CreateWithPooledClient(createKnowledgeKnowledgebase),
		ReadContext:   provider.ReadWithPooledClient(readKnowledgeKnowledgebase),
		UpdateContext: provider.UpdateWithPooledClient(updateKnowledgeKnowledgebase),
		DeleteContext: provider.DeleteWithPooledClient(deleteKnowledgeKnowledgebase),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Knowledge base name",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"description": {
				Description: "Knowledge base description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"core_language": {
				Description:      "Core language for knowledge base in which initial content must be created, language codes [en-US, en-UK, en-AU, de-DE] are supported currently, however the new DX knowledge will support all these language codes",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validators.ValidateLanguageCode,
			},
			"published": {
				Description: "Flag that indicates the knowledge base is published",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func dataSourceKnowledgeKnowledgebase() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Knowledge Base. Select a knowledge base by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceKnowledgeKnowledgebaseRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Knowledge base name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"core_language": {
				Description:      "Core language for knowledge base in which initial content must be created, language codes [en-US, en-UK, en-AU, de-DE] are supported currently, however the new DX knowledge will support all these language codes",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validators.ValidateLanguageCode,
			},
		},
	}
}
