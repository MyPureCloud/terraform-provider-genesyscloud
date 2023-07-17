package genesyscloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

func dataSourceOrganizationsMe() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud current organization",
		ReadContext: ReadWithPooledClient(dataSourceOrganizationsMeRead),
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"default_language": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"default_country_code": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"default_site_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"third_party_org_name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"voicemail_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"product_platform": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"support_uri": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
		},
	}
}

func dataSourceOrganizationsMeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	orgAPI := platformclientv2.NewOrganizationApiWithConfig(sdkConfig)

	orgMe, _, getErr := orgAPI.GetOrganizationsMe()
	if getErr != nil {
		return diag.Errorf("Error requesting organization: %s", getErr)
	}

	d.SetId(*orgMe.Id)
	if orgMe.Name != nil {
		d.Set("name", *orgMe.Name)
	}
	if orgMe.DefaultLanguage != nil {
		d.Set("default_language", *orgMe.DefaultLanguage)
	}
	if orgMe.DefaultCountryCode != nil {
		d.Set("default_country_code", *orgMe.DefaultCountryCode)
	}
	if orgMe.Domain != nil {
		d.Set("domain", *orgMe.Domain)
	}
	if orgMe.DefaultSiteId != nil {
		d.Set("default_site_id", *orgMe.DefaultSiteId)
	}
	if orgMe.ThirdPartyOrgName != nil {
		d.Set("third_party_org_name", *orgMe.ThirdPartyOrgName)
	}
	if orgMe.VoicemailEnabled != nil {
		d.Set("voicemail_enabled", *orgMe.VoicemailEnabled)
	}
	if orgMe.ProductPlatform != nil {
		d.Set("product_platform", *orgMe.ProductPlatform)
	}
	if orgMe.SupportURI != nil {
		d.Set("support_uri", *orgMe.SupportURI)
	}

	return nil
}
