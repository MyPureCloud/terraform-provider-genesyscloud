package external_contacts_external_source

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const ResourceType = "genesyscloud_externalcontacts_external_source"

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterDataSource(ResourceType, DataSourceExternalContactsExternalSource())
	regInstance.RegisterResource(ResourceType, ResourceExternalContactsExternalSource())
	regInstance.RegisterExporter(ResourceType, ExternalContactsExternalSourceExporter())
}

func ResourceExternalContactsExternalSource() *schema.Resource {
	externalSourceLinkConfiguration := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`uri_template`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud external contacts external source`,

		CreateContext: provider.CreateWithPooledClient(createExternalContactsExternalSource),
		ReadContext:   provider.ReadWithPooledClient(readExternalContactsExternalSource),
		UpdateContext: provider.UpdateWithPooledClient(updateExternalContactsExternalSource),
		DeleteContext: provider.DeleteWithPooledClient(deleteExternalContactsExternalSource),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the external source.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`active`: {
				Description: `Whether the external source is active.`,
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			`link_configuration`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        externalSourceLinkConfiguration,
			},
		},
	}
}

func ExternalContactsExternalSourceExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthExternalContactsExternalSources),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No Reference
	}
}

func DataSourceExternalContactsExternalSource() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud external contacts external source data source. Select an external source by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceExternalContactsExternalSourceRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `external source name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
