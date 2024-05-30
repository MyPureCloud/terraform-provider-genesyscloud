package outbound_contact_list_contact

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
)

func ResourceOutboundContactListContact() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Outbound Contact List Contact`,

		CreateContext: provider.CreateWithPooledClient(nil),
		ReadContext:   provider.ReadWithPooledClient(nil),
		UpdateContext: provider.UpdateWithPooledClient(nil),
		DeleteContext: provider.DeleteWithPooledClient(nil),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The globally unique identifier for the object. If none is provided, a GUID will be generated.",
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"contact_list_id": {
				Description: "",
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},
			"priority": {
				Description: `Contact priority. True means the contact(s) will be dialed next; false means the contact will go to the end of the contact queue. 
Only applicable on the creation of a contact, so updating this field will force the contact to be deleted from the contact list and re-uploaded.`,
				ForceNew: true,
				Optional: true,
				Type:     schema.TypeBool,
			},
			"clear_system_data": {
				Description: `Clear system data. True means the system columns (attempts, callable status, etc) stored on the contact will be cleared if the contact already exists; false means they won't. 
Only applicable on the creation of a contact, so updating this field will force the contact to be deleted from the contact list and re-uploaded.`,
				ForceNew: true,
				Optional: true,
				Type:     schema.TypeBool,
			},
			"do_not_queue": {
				Description: `Do not queue. True means that updated contacts will not have their positions in the queue altered, so contacts that have already been dialed will not be redialed. 
For new contacts, this parameter has no effect; False means that updated contacts will be re-queued, according to the 'priority' parameter. 
Only applicable on the creation of a contact, so updating this field will force the contact to be deleted from the contact list and re-uploaded.`,
				ForceNew: true,
				Optional: true,
				Type:     schema.TypeBool,
			},
			"callable": {
				Description: `Indicates whether or not the contact can be called.`,
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
			},
			"data": {
				Description: `An ordered map of the contact's columns and corresponding values.`,
				Type:        schema.TypeMap,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"phone_number_status": {
				Description: `A map of phone number columns to PhoneNumberStatuses, which indicate if the phone number is callable or not.`,
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Description: `Phone number column identifier.`,
							Type:        schema.TypeString,
							Required:    true,
						},
						"callable": {
							Description: `Indicates whether or not a phone number is callable.`,
							Type:        schema.TypeBool,
							Default:     false,
							Optional:    true,
						},
					},
				},
			},
			"contactable_status": {
				Description: `A map of media types (Voice, SMS and Email) to ContactableStatus, which indicates if the contact can be contacted using the specified media type.`,
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"media_type": {
							Description: `The media type.`,
							Type:        schema.TypeString,
							Required:    true,
						},
						"contactable": {
							Description: `Indicates whether or not the entire contact is contactable for the associated media type.`,
							Type:        schema.TypeBool,
							Required:    true,
						},
						"column_status": {
							Description: `A map of individual contact method columns to whether the individual column is contactable for the associated media type.`,
							Type:        schema.TypeSet,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"column": {
										Description: `Contact method column.`,
										Type:        schema.TypeString,
										Required:    true,
									},
									"contactable": {
										Description: `Indicates whether or not an individual contact method column is contactable.`,
										Type:        schema.TypeBool,
										Required:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
