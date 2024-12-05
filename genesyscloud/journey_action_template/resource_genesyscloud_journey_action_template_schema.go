package journey_action_template

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const ResourceType = "genesyscloud_journey_action_template"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceJourneyActionTemplate())
	regInstance.RegisterDataSource(ResourceType, DataSourceJourneyActionTemplate())
	regInstance.RegisterExporter(ResourceType, JourneyActionTemplateExporter())
}

var (
	journeyActionTemplateSchema = map[string]*schema.Schema{
		"name": {
			Description: "Name of the action template.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"description": {
			Description: "Description of the action template's functionality.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"media_type": {
			Description:  "The media type of the action configured by the action template.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"webchat", "webMessagingOffer", "contentOffer", "architectFlow", "openAction"}, false),
		},
		"state": {
			Description:  "The state of the action template.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"Active", "Inactive", "Deleted"}, false),
		},
		"content_offer": {
			Description: "Properties for configuring a content offer action.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem:        contentOfferResource,
		},
	}

	contentOfferResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"image_url": {
				Description: "URL for image displayed on the content offer.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"display_mode": {
				Description:  "The display mode used by Genesys Widgets when displaying the content offer.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Modal", "Overlay", "Toast"}, false),
			},
			"layout_mode": {
				Description:  "The layout mode used by Genesys Widgets when displaying the content offer.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"TextOnly", "ImageOnly", "LeftText", "RightText", "TopText", "BottomText"}, false),
			},
			"title": {
				Description: "Title in the header of the content offer.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"headline": {
				Description: "Headline displayed above the body text of the content offer.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"body": {
				Description: "Body text of the content offer.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"call_to_action": {
				Description: "Properties customizing the call to action button on the content offer.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        callToActionResource,
			},
			"style": {
				Description: "Properties customizing the styling of the content offer.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        contentOfferStylingConfigurationResource,
			},
		},
	}

	callToActionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"text": {
				Description: "Text displayed on the call to action button.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"url": {
				Description: "URL to open when user clicks on the call to action button.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"target": {
				Description:  "Where should the URL be opened when the user clicks on the call to action button.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Blank", "Self"}, false),
			},
		},
	}

	contentOfferStylingConfigurationResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"position": {
				Description: "Properties for customizing the positioning of the content offer.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        contentPositionPropertiesResource,
			},
			"offer": {
				Description: "Properties for customizing the appearance of the content offer.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        contentOfferStylePropertiesResource,
			},
			"close_button": {
				Description: "Properties for customizing the appearance of the close button.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        closeButtonStylePropertiesResource,
			},
			"cta_button": {
				Description: "Properties for customizing the appearance of the CTA button.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        ctaButtonStylePropertiesResource,
			},
			"title": {
				Description: "Properties for customizing the appearance of the title text.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        textStylePropertiesResource,
			},
			"headline": {
				Description: "Properties for customizing the appearance of the headline text.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        textStylePropertiesResource,
			},
			"body": {
				Description: "Properties for customizing the appearance of the body text.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        textStylePropertiesResource,
			},
		},
	}

	contentPositionPropertiesResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"top": {
				Description: "Top positioning offset.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"bottom": {
				Description: "Bottom positioning offset.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"left": {
				Description: "Left positioning offset.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"right": {
				Description: "Right positioning offset.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	contentOfferStylePropertiesResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"padding": {
				Description: "Padding of the offer.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"color": {
				Description: "Text color of the offer.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"background_color": {
				Description: "Background color of the offer.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	closeButtonStylePropertiesResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"color": {
				Description: "Color of button.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"opacity": {
				Description: "Opacity of button.",
				Type:        schema.TypeFloat,
				Optional:    true,
			},
		},
	}

	ctaButtonStylePropertiesResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"color": {
				Description: "Color of the text.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"font": {
				Description: "Font of the text.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"font_size": {
				Description: "Font size of the text.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"text_align": {
				Description: "Text alignment.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"background_color": {
				Description: "Background color of the CTA button.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	textStylePropertiesResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"color": {
				Description: "Color of the text.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"font": {
				Description: "Font of the text.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"font_size": {
				Description: "Font size of the text.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"text_align": {
				Description:  "Text alignment.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Left", "Right", "Center"}, false),
			},
		},
	}
)

func JourneyActionTemplateExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllJourneyActionTemplates),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No Reference
	}
}

func ResourceJourneyActionTemplate() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Journey Action Template",
		CreateContext: provider.CreateWithPooledClient(createJourneyActionTemplate),
		ReadContext:   provider.ReadWithPooledClient(readJourneyActionTemplate),
		UpdateContext: provider.UpdateWithPooledClient(updateJourneyActionTemplate),
		DeleteContext: provider.DeleteWithPooledClient(deleteJourneyActionTemplate),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema:        journeyActionTemplateSchema,
	}
}

func DataSourceJourneyActionTemplate() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Action Template. Select a journey action template by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceJourneyActionTemplateRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Journey Action Template name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
