package external_contacts_organization

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = `genesyscloud_externalcontacts_organization`

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceExternalContactsOrganization())
	regInstance.RegisterDataSource(ResourceType, DataSourceExternalContactsOrganization())
	regInstance.RegisterExporter(ResourceType, ExternalContactsOrganizationExporter())
}

func ResourceExternalContactsOrganization() *schema.Resource {
	contactAddressResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`address1`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`address2`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`city`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`state`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`postal_code`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`country_code`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	phoneNumberResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`display`: {
				Description: `Display string of the phone number.`,
				Type:        schema.TypeString,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return hashFormattedPhoneNumber(old) == hashFormattedPhoneNumber(new)
				},
				Optional: true,
				Computed: true,
			},
			`extension`: {
				Description: `Phone extension.`,
				Type:        schema.TypeInt,
				Optional:    true,
			},
			`accepts_sms`: {
				Description: `If contact accept SMS.`,
				Type:        schema.TypeBool,
				Optional:    true,
			},
			`e164`: {
				Description:      `Phone number in e164 format.`,
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validators.ValidatePhoneNumber,
			},
			`country_code`: {
				Description: `Phone number country code.`,
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}

	tickerResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`symbol`: {
				Description: `The ticker symbol for this organization. Example: ININ, AAPL, MSFT, etc.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`exchange`: {
				Description: `The exchange for this ticker symbol. Examples: NYSE, FTSE, NASDAQ, etc.`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}

	twitterIdResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`twitter_id`: {
				Description: `Contact twitter id.`,
				Type:        schema.TypeString,
				Required:    true,
			},
			`name`: {
				Description: `Contact twitter name.`,
				Type:        schema.TypeString,
				Required:    true,
			},
			`screen_name`: {
				Description: `Contact twitter screen name.`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}

	trustorResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`enabled`: {
				Description: `If disabled no trustee user will have access, even if they were previously added.`,
				Required:    true,
				Type:        schema.TypeBool,
			},
		},
	}

	jsonSchemaDocumentResource := &schema.Resource{
		Schema: map[string]*schema.Schema{

			`description`: {
				Description: `A brief description of the custom fields`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`title`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`required`: {
				Description: `The required fields in the schema`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`properties`: {
				Description:      `The properties for the JSON Schema document.`,
				Optional:         true,
				Type:             schema.TypeString,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
		},
	}

	dataSchemaResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`schema_id`: {
				Description: `The globally unique identifier for the schema. Only required if a schema is used for custom fields during external entity creation or updates.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`name`: {
				Description: `The name of the schema.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`version`: {
				Description: `The schema's version, a positive integer. Required for updates.`,
				Optional:    true,
				Type:        schema.TypeInt,
				Default:     0,
			},
			`enabled`: {
				Description: `The schema's enabled/disabled status. A disabled schema cannot be assigned to any other entities, but the data on those entities from the schema still exists.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`json_schema`: {
				Description: `A JSON schema defining the extension to the built-in entity type.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        jsonSchemaDocumentResource,
			},
		},
	}

	externalDataSourceResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`platform`: {
				Description: `The platform that was the source of the data.  Example: a CRM like SALESFORCE.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`url`: {
				Description: `An URL that links to the source record that contributed data to the associated entity.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud external contacts organization`,

		CreateContext: provider.CreateWithPooledClient(createExternalContactsOrganization),
		ReadContext:   provider.ReadWithPooledClient(readExternalContactsOrganization),
		UpdateContext: provider.UpdateWithPooledClient(updateExternalContactsOrganization),
		DeleteContext: provider.DeleteWithPooledClient(deleteExternalContactsOrganization),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the company.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`company_type`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`industry`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`primary_contact_id`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`address`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        contactAddressResource,
			},
			`phone_number`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        phoneNumberResource,
			},
			`fax_number`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        phoneNumberResource,
			},
			`employee_count`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`revenue`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`tags`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`websites`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`tickers`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        tickerResource,
			},
			`twitter`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        twitterIdResource,
			},
			`external_system_url`: {
				Description: `A string that identifies an external system-of-record resource that may have more detailed information on the organization. It should be a valid URL (including the http/https protocol, port, and path [if any]). The value is automatically trimmed of any leading and trailing whitespace.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`trustor`: {
				Description: ``,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        trustorResource,
			},
			`schema`: {
				Description: `The schema defining custom fields for this contact`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        dataSchemaResource,
			},
			`custom_fields`: {
				Description:      `JSON formatted object for custom field values defined in the schema referenced by the worktype of the workitem.`,
				Optional:         true,
				Computed:         true,
				Type:             schema.TypeString,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
			`external_data_sources`: {
				Description: `Links to the sources of data (e.g. one source might be a CRM) that contributed data to this record.  Read-only, and only populated when requested via expand param.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        externalDataSourceResource,
			},
		},
	}
}

// ExternalContactsOrganizationExporter returns the resourceExporter object used to hold the genesyscloud_external_contacts_organization exporter's config
func ExternalContactsOrganizationExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthExternalContactsOrganizations),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			`external_organizations_schemas`: {}, //Need to add this when we implement organization schemas
		},
	}
}

// DataSourceExternalContactsOrganization registers the genesyscloud_external_contacts_organization data source
func DataSourceExternalContactsOrganization() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud external contacts organization data source. Select an external contacts organization by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceExternalContactsOrganizationRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `external contacts organization name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
