package outbound_contact_list_contacts_bulk

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const ResourceType = "genesyscloud_outbound_contact_list_contacts_bulk"

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceOutboundContactListContactsBulk())
	regInstance.RegisterExporter(ResourceType, BulkContactsExporter())
}

func BulkContactsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllContacts),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"contact_list_id":          {RefType: "genesyscloud_outbound_contact_list"},
			"contact_list_template_id": {RefType: "genesyscloud_outbound_contact_list_template"},
		},
		UnResolvableAttributes: map[string]*schema.Schema{
			"filepath": ResourceOutboundContactListContactsBulk().Schema["filepath"],
		},
		CustomFlowResolver: map[string]*resourceExporter.CustomFlowResolver{
			"file_content_hash": {ResolverFunc: resourceExporter.FileContentHashResolver},
		},
	}
}

func ResourceOutboundContactListContactsBulk() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Outbound Contact List Bulk Contacts Handling`,

		CreateContext: provider.CreateWithPooledClient(createOutboundContactListContact),
		ReadContext:   provider.ReadWithPooledClient(readOutboundContactListContact),
		UpdateContext: provider.UpdateWithPooledClient(updateOutboundContactListContact),
		DeleteContext: provider.DeleteWithPooledClient(deleteOutboundContactListContact),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"filepath": {
				Description:  `The path to the CSV file containing the contacts to be added to the contact list.`,
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validators.ValidatePath,
			},
			"file_content_hash": {
				Description: "Hash value of the CSV file content. Used to detect changes.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"contact_id_name": {
				Description: `The name of the column in the CSV file that contains the contact's unique contact id.`,
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},
			"contact_list_id": {
				Description:  `The identifier of the contact list. Either this or the "contact_list_template_id" attribute are required to be set.`,
				ForceNew:     true,
				Optional:     true,
				Type:         schema.TypeString,
				ExactlyOneOf: []string{"contact_list_id", "contact_list_template_id"},
			},
			"contact_list_template_id": {
				Description:  `The identifier of the contact list template. Either this or the "contact_list_id" attribute are required to be set.`,
				ForceNew:     true,
				Optional:     true,
				Type:         schema.TypeString,
				ExactlyOneOf: []string{"contact_list_id", "contact_list_template_id"},
				RequiredWith: []string{"list_name_prefix"},
			},
			"list_name_prefix": {
				Description: `String that will replace %N in the "list_name_format" attribute specified on the import template.`,
				ForceNew:    true,
				Type:        schema.TypeString,
				RequiredWith: []string{
					"contact_list_template_id",
				},
			},
			"division_id_for_target_contact_lists": {
				Description: `The identifier of the division to be used for the creation of the target contact lists. If not provided, Home division will be used.`,
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
}
