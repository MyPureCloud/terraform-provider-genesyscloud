package outbound_contact_list_contacts_bulk

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const ResourceType = "genesyscloud_outbound_contact_list_contacts_bulk"

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceOutboundContactListContactsBulk())
	regInstance.RegisterExporter(ResourceType, BulkContactsExporter())
}

func BulkContactsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllContactLists),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"contact_list_id": {RefType: "genesyscloud_outbound_contact_list"},
			// "contact_list_template_id": {RefType: "genesyscloud_outbound_contact_list_template"},
		},
		UnResolvableAttributes: map[string]*schema.Schema{
			"filepath": ResourceOutboundContactListContactsBulk().Schema["filepath"],
		},
		CustomFileWriter: resourceExporter.CustomFileWriterSettings{
			RetrieveAndWriteFilesFunc: BulkContactsExporterResolver,
			SubDirectory:              "bulkContacts",
		},
	}
}

func ResourceOutboundContactListContactsBulk() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Outbound Contact List Bulk Contacts Handling`,

		CreateContext: provider.CreateWithPooledClient(createOutboundContactListBulkContacts),
		ReadContext:   provider.ReadWithPooledClient(readOutboundContactListBulkContacts),
		// UpdateContext: provider.UpdateWithPooledClient(createOutboundContactListBulkContacts),
		DeleteContext: provider.DeleteWithPooledClient(deleteOutboundContactListBulkContacts),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		CustomizeDiff: customdiff.All(
			customdiff.ComputedIf("file_content_hash", fileContentHashChanged),
			// importTemplateAttributesSchemaLogic(),
		),
		Schema: map[string]*schema.Schema{
			"filepath": {
				Description:  `The path to the CSV file containing the contacts to be added to the contact list.`,
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validators.ValidatePath,
			},
			"contact_list_id": {
				Description: `The identifier of the contact list. Either this or the "contact_list_template_id" attribute are required to be set.`,
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeString,
				// ExactlyOneOf: []string{"contact_list_id", "contact_list_template_id"},
			},
			// "contact_list_template_id": {
			// 	Description:  `The identifier of the contact list template. Either this or the "contact_list_id" attribute are required to be set.`,
			// 	ForceNew:     true,
			// 	Optional:     true,
			// 	Type:         schema.TypeString,
			// 	ExactlyOneOf: []string{"contact_list_id", "contact_list_template_id"},
			// 	RequiredWith: []string{
			// 		"contact_list_template_id",
			// 		"list_name_prefix",
			// 	},
			// },
			"contact_id_name": {
				Description: `The name of the column in the CSV file that contains the contact's unique contact id.`,
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},
			// "list_name_prefix": {
			// 	Description: `String that will replace %N in the "list_name_format" attribute specified on the import template.`,
			// 	ForceNew:    true,
			// 	Type:        schema.TypeString,
			// 	RequiredWith: []string{
			// 		"contact_list_template_id",
			// 		"list_name_prefix",
			// 	},
			// },
			// "division_id_for_target_contact_lists": {
			// 	Description: `The identifier of the division to be used for the creation of the target contact lists. If not provided, Home division will be used. Only effective in conjunction with "contact_list_template_id" attribute.`,
			// 	ForceNew:    true,
			// 	Optional:    true,
			// 	Type:        schema.TypeString,
			// },
			// Computed attributes
			"file_content_hash": {
				Description: "Hash value of the CSV file content. This is a computed value used to detect changes.",
				Type:        schema.TypeString,
				Computed:    true,
				Required:    false,
				Optional:    false,
				ForceNew:    true,
			},
			"record_count": {
				Description: `The number of contacts in the contact list. This is a read-only identifying attribute.`,
				Optional:    false,
				Required:    false,
				Computed:    true,
				ForceNew:    true,
				Type:        schema.TypeInt,
			},
		},
	}
}
