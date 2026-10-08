package responsemanagement_response

// @team: Response Management
// @chat: #genesys-cloud-canned-responses
// @pm: Marudhu Pandian
// @jira: RESPONSES
// @description: Response management provides the ability to support 'canned' responses to commonly asked questions for contact center users. The responses are designed to be used with email, chat, SMS, etc.

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesycloud_responsemanagement_response_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the responsemanagement_response resource.
3.  The datasource schema definitions for the responsemanagement_response datasource.
4.  The resource exporter configuration for the responsemanagement_response exporter.
*/
const ResourceType = "genesyscloud_responsemanagement_response"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceResponsemanagementResponse())
	regInstance.RegisterDataSource(ResourceType, DataSourceResponsemanagementResponse())
	regInstance.RegisterExporter(ResourceType, ResponsemanagementResponseExporter())
}

var (
	responsetextResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`content`: {
				Description: `Response text content.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`content_type`: {
				Description:  `Response text content type.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`text/plain`, `text/html`}, false),
			},
			`type`: {
				Description:  `Response text type.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`body`, `subject`}, false),
			},
		},
	}

	footerResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`type`: {
				Description:  `Specifies the type represented by Footer.Valid values: Signature.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`Signature`}, false),
			},
			`applicable_resources`: {
				Description: `Specifies the canned response template where the footer can be used.Valid values: Campaign.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	substitutionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`id`: {
				Description: `Response substitution identifier.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`description`: {
				Description: `Response substitution description.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`default_value`: {
				Description: `Response substitution default value.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
	messagingtemplateResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`whats_app`: {
				Description: `Defines a messaging template for a WhatsApp messaging channel`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeSet,
				Elem:        whatsappDefinitionResource,
				Set: func(_ interface{}) int {
					return 0
				},
			},
		},
	}
	whatsappDefinitionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The messaging template name.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`namespace`: {
				Description: `The messaging template namespace.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`language`: {
				Description: `The messaging template language configured for this template. This is a WhatsApp specific value. For example, 'en_US'`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}

	formMessageResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`title`: {
				Description: `Title of the message.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`subtitle`: {
				Description: `Subtitle of the message.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`image_url`: {
				Description: `URL of the image to display.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	formIntroductionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`title`: {
				Description: `Title of the introduction.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`subtitle`: {
				Description: `Subtitle of the introduction.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`button_text`: {
				Description: `Text for the start button.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`image_url`: {
				Description: `URL of the image to display.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	formListPickerItemResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`title`: {
				Description: `Title of the item.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`image_url`: {
				Description: `URL of the image to display.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	formListPickerSectionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`multiple_selection`: {
				Description: `Whether multiple items can be selected.`,
				Required:    true,
				Type:        schema.TypeBool,
			},
			`items`: {
				Description: `Items in this section. Must contain between 2 and 100 items.`,
				Required:    true,
				Type:        schema.TypeList,
				MinItems:    2,
				MaxItems:    100,
				Elem:        formListPickerItemResource,
			},
			`title`: {
				Description: `Title of the section.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	formListPickerResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`sections`: {
				Description: `Sections in the list picker. At least one section is required.`,
				Required:    true,
				Type:        schema.TypeList,
				MinItems:    1,
				Elem:        formListPickerSectionResource,
			},
		},
	}

	formDatePickerResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`date_display_format`: {
				Description: `Date display format. For example: 'dayMonthYear', 'monthDayYear' or 'yearMonthDay'.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`title`: {
				Description: `Title of the date picker.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`subtitle`: {
				Description: `Subtitle of the date picker.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	formInputResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`is_multiple_line`: {
				Description: `Whether the input supports multiple lines.`,
				Required:    true,
				Type:        schema.TypeBool,
			},
			`is_required`: {
				Description: `Whether the input is required.`,
				Required:    true,
				Type:        schema.TypeBool,
			},
			`title`: {
				Description: `Title of the input field.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`subtitle`: {
				Description: `Subtitle of the input field.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`placeholder_text`: {
				Description: `Placeholder text for the input.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`keyboard_type`: {
				Description:  `Type of keyboard to be shown.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`Default`, `NumberPunctuation`, `Number`, `Phone`, `Email`, `Decimal`, `Websearch`, `URL`}, false),
			},
			`auto_complete_type`: {
				Description: `A string value representing the keyboard and system information about the expected semantic meaning for the content that users enter.`,
				Optional:    true,
				Type:        schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					`Prefix`, `Name`, `GivenName`, `MiddleName`, `FamilyName`, `Suffix`, `Nickname`, `Title`,
					`Organization`, `Location`, `StreetAddress`, `Addressline1`, `Addressline2`, `City`, `State`,
					`Country`, `PostalCode`, `Username`, `OneTimeCode`, `Email`, `Phone`, `PaymentCardNumber`,
					`PaymentCardExpiration`, `PaymentCardExpirationMonth`, `PaymentCardExpirationYear`,
					`PaymentCardSecurityCode`, `PaymentCardType`, `PaymentCardName`, `PaymentCardGivenName`,
					`PaymentCardMiddleName`, `PaymentCardFamilyName`, `Birthdate`, `BirthdateDay`, `BirthdateMonth`,
					`BirthdateYear`, `DateTime`, `FlightNumber`, `Url`,
				}, false),
			},
		},
	}

	formWheelPickerItemResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`title`: {
				Description: `Title of the item.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`value`: {
				Description: `Value of the item.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	formWheelPickerResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`items`: {
				Description: `Items in the wheel picker. Must contain between 2 and 100 items.`,
				Required:    true,
				Type:        schema.TypeList,
				MinItems:    2,
				MaxItems:    100,
				Elem:        formWheelPickerItemResource,
			},
		},
	}

	formPageComponentResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`form_component_type`: {
				Description:  `Type of the component. The matching component block must be set.`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`ListPicker`, `DatePicker`, `Input`, `WheelPicker`}, false),
			},
			`list_picker`: {
				Description: `List picker configuration. Required when form_component_type is 'ListPicker'.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        formListPickerResource,
			},
			`date_picker`: {
				Description: `Date picker configuration. Required when form_component_type is 'DatePicker'.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        formDatePickerResource,
			},
			`input`: {
				Description: `Input field configuration. Required when form_component_type is 'Input'.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        formInputResource,
			},
			`wheel_picker`: {
				Description: `Wheel picker configuration. Required when form_component_type is 'WheelPicker'.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        formWheelPickerResource,
			},
		},
	}

	formPageResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`title`: {
				Description: `Title of the page.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`subtitle`: {
				Description: `Subtitle of the page.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`page_components`: {
				Description: `Components on this page.`,
				Required:    true,
				Type:        schema.TypeList,
				MinItems:    1,
				Elem:        formPageComponentResource,
			},
		},
	}

	formResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`form_description`: {
				Description: `Description of the form.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`received_message`: {
				Description: `Message displayed when the response is received.`,
				Required:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        formMessageResource,
			},
			`reply_message`: {
				Description: `Message displayed as reply.`,
				Required:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        formMessageResource,
			},
			`form_pages`: {
				Description: `Pages of the form. Must contain between 1 and 20 pages.`,
				Required:    true,
				Type:        schema.TypeList,
				MinItems:    1,
				MaxItems:    20,
				Elem:        formPageResource,
			},
			`show_summary`: {
				Description: `Whether to show a summary after form completion.`,
				Required:    true,
				Type:        schema.TypeBool,
			},
			`introduction`: {
				Description: `Introduction section of the form.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        formIntroductionResource,
			},
		},
	}
)

// ResourceResponsemanagementResponse registers the genesyscloud_responsemanagement_response resource with Terraform
func ResourceResponsemanagementResponse() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud responsemanagement response`,

		CreateContext: provider.CreateWithPooledClient(createResponsemanagementResponse),
		ReadContext:   provider.ReadWithPooledClient(readResponsemanagementResponse),
		UpdateContext: provider.UpdateWithPooledClient(updateResponsemanagementResponse),
		DeleteContext: provider.DeleteWithPooledClient(deleteResponsemanagementResponse),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `Name of the responsemanagement response`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`library_ids`: {
				Description: `One or more libraries response is associated with. Changing the library IDs will result in the resource being recreated`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`texts`: {
				Description: `One or more texts associated with the response. Required for the Standard, Footer, MessagingTemplate and CampaignEmailTemplate response types. Not used by the Form response type, which ignores texts and always returns an empty list, so setting this alongside 'form' results in a permanent diff.`,
				Optional:    true,
				Type:        schema.TypeSet,
				Elem:        responsetextResource,
			},
			`interaction_type`: {
				Description:  `The interaction type for this response.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`chat`, `email`, `twitter`}, false),
			},
			`substitutions`: {
				Description: `Details about any text substitutions used in the texts for this response.`,
				Optional:    true,
				Type:        schema.TypeSet,
				Elem:        substitutionResource,
			},
			`substitutions_schema_id`: {
				Description: `Metadata about the text substitutions in json schema format.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`response_type`: {
				Description:  `The response type represented by the response.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`MessagingTemplate`, `CampaignSmsTemplate`, `CampaignEmailTemplate`, `Footer`, `Form`}, false),
			},
			`messaging_template`: {
				Description: `An optional messaging template definition for responseType.MessagingTemplate.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeSet,
				Elem:        messagingtemplateResource,
				Set: func(_ interface{}) int {
					return 0
				},
			},
			`asset_ids`: {
				Description: `Assets used in the response`,
				Optional:    true,
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`footer`: {
				Description: `Footer template identifies the Footer type and its footerUsage`,
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        footerResource,
			},
			`form`: {
				Description: `Form template definition for responseType.Form. Requires response_type to be set to 'Form'.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        formResource,
			},
		},
	}
}

// ResponsemanagementResponseExporter returns the resourceExporter object used to hold the genesyscloud_responsemanagement_response exporter's config
func ResponsemanagementResponseExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthResponsemanagementResponses),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			`library_ids`: {
				RefType: "genesyscloud_responsemanagement_library",
			},
			`library_id`: {
				RefType: "genesyscloud_responsemanagement_library",
			},
			`asset_ids`: {
				RefType: "responsemanagement_responseasset",
			},
		},
		JsonEncodeAttributes: []string{"substitutions_schema_id"},
		// The form block has attributes that are required by the schema but may legitimately come back
		// from the API as an empty string. Without these entries the exporter would strip them and
		// produce a config that no longer validates.
		AllowZeroValues: []string{
			"form.form_description",
			"form.received_message.title",
			"form.reply_message.title",
			"form.introduction.title",
			"form.introduction.subtitle",
			"form.introduction.button_text",
			"form.form_pages.title",
			"form.form_pages.subtitle",
			"form.form_pages.page_components.form_component_type",
			"form.form_pages.page_components.date_picker.date_display_format",
			"form.form_pages.page_components.list_picker.sections.items.title",
			"form.form_pages.page_components.wheel_picker.items.title",
		},
		DataSourceResolver: map[*resourceExporter.DataAttr]*resourceExporter.ResourceAttr{
			{Attr: "library_id"}: {Attr: "library_ids\\.\\d+"},
		},
	}
}

// DataSourceResponsemanagementResponse registers the genesyscloud_responsemanagement_response data source
func DataSourceResponsemanagementResponse() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Responsemanagement Response. Select a Responsemanagement Response by name.`,
		ReadContext: provider.ReadWithPooledClient(dataSourceResponsemanagementResponseRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Responsemanagement Response name.`,
				Type:        schema.TypeString,
				Required:    true,
			},
			"library_id": {
				Description: `ID of the library that contains the response.`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
