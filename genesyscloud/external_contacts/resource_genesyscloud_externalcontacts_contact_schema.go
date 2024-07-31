package external_contacts

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/validators"
)

/*
resource_genesyscloud_externalcontacts_contacts_schema.go should hold four types of functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the externalcontacts_contacts resource.
3.  The datasource schema definitions for the externalcontacts_contacts datasource.
4.  The resource exporter configuration for the externalcontacts_contacts exporter.
*/
const resourceName = "genesyscloud_externalcontacts_contact"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, DataSourceExternalContactsContact())
	l.RegisterResource(resourceName, ResourceExternalContact())
	l.RegisterExporter(resourceName, ExternalContactExporter())
}

// ResourceExternalContact registers the genesyscloud_externalcontacts_contact resource with Terraform
func ResourceExternalContact() *schema.Resource {
	phoneNumber := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"display": {
				Description: "Display string of the phone number.",
				Type:        schema.TypeString,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return hashFormattedPhoneNumber(old) == hashFormattedPhoneNumber(new)
				},
				Optional: true,
				Computed: true,
			},
			"extension": {
				Description: "Phone extension.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"accepts_sms": {
				Description: "If contact accept SMS.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"e164": {
				Description:      "Phone number in e164 format.",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validators.ValidatePhoneNumber,
			},
			"country_code": {
				Description: "Phone number country code.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}

	address := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"address1": {
				Description: "Contact address 1.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"address2": {
				Description: "Contact address 2.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"city": {
				Description: "Contact address city.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"state": {
				Description: "Contact address state.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"postal_code": {
				Description: "Contact address postal code.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"country_code": {
				Description:      "Contact address country code.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validators.ValidateCountryCode,
			},
		},
	}

	twitterId := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Contact twitter id.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "Contact twitter name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"screen_name": {
				Description: "Contact twitter screen name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"profile_url": {
				Description: "Contact twitter account url.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}

	lineIds := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"user_id": {
				Description: "Contact line id.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	lineId := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"ids": {
				Description: "Contact line id.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        lineIds,
			},
			"display_name": {
				Description: "Contact line display name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	whatsappId := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"phone_number": {
				Description: "Contact whatsapp phone number.",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        phoneNumber,
			},
			"display_name": {
				Description: "Contact whatsapp display name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}

	facebookIds := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"scoped_id": {
				Description: "Contact facebook scoped id.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	facebookId := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"ids": {
				Description: "Contact facebook scoped id.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        facebookIds,
			},
			"display_name": {
				Description: "Contact whatsapp display name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	return &schema.Resource{
		Description: "Genesys Cloud External Contact",

		CreateContext: provider.CreateWithPooledClient(createExternalContact),
		ReadContext:   provider.ReadWithPooledClient(readExternalContact),
		UpdateContext: provider.UpdateWithPooledClient(updateExternalContact),
		DeleteContext: provider.DeleteWithPooledClient(deleteExternalContact),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"first_name": {
				Description: "The first name of the contact.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"middle_name": {
				Description: "The middle name of the contact.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"last_name": {
				Description: "The last name of the contact.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"salutation": {
				Description: "The salutation of the contact.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"title": {
				Description: "The title of the contact.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"work_phone": {
				Description: "Contact work phone settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        phoneNumber,
			},
			"cell_phone": {
				Description: "Contact call phone settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        phoneNumber,
			},
			"home_phone": {
				Description: "Contact home phone settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        phoneNumber,
			},
			"other_phone": {
				Description: "Contact other phone settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        phoneNumber,
			},
			"work_email": {
				Description: "Contact work email.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"personal_email": {
				Description: "Contact personal email.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"other_email": {
				Description: "Contact other email.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"address": {
				Description: "Contact address.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        address,
			},
			"twitter_id": {
				Description: "Contact twitter account informations.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        twitterId,
			},
			"line_id": {
				Description: "Contact line account informations.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        lineId,
			},
			"whatsapp_id": {
				Description: "Contact whatsapp account informations.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        whatsappId,
			},
			"facebook_id": {
				Description: "Contact facebook account informations.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        facebookId,
			},
			"survey_opt_out": {
				Description: "Contact survey opt out preference.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"external_system_url": {
				Description: "Contact external system url.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

// ExternalContactExporter returns the resourceExporter object used to hold the genesyscloud_externalcontacts_contact exporter's config
func ExternalContactExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthExternalContacts),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"external_organization": {}, //Need to add this when we external orgs implemented
		},
	}
}

// DataSourceExternalContactsContact registers the genesyscloud_externalcontacts_contact data source
func DataSourceExternalContactsContact() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud external contacts. Select a contact by any string search.",
		ReadContext: provider.ReadWithPooledClient(dataSourceExternalContactsContactRead),
		Schema: map[string]*schema.Schema{
			"search": {
				Description: "The search string for the contact.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}
