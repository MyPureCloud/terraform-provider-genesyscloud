package genesyscloud

import (
	"context"
	"fmt"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

const DataSourceOrganizationsMeResourceType = "genesyscloud_organizations_me"

func DataSourceOrganizationsMe() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud current organization",
		ReadContext: provider.ReadWithPooledClient(dataSourceOrganizationsMeRead),
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
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	orgAPI := platformclientv2.NewOrganizationApiWithConfig(sdkConfig)

	orgMe, resp, getErr := orgAPI.GetOrganizationsMe()
	if getErr != nil {
		return util.BuildAPIDiagnosticError(DataSourceOrganizationsMeResourceType, fmt.Sprintf("Error requesting organization: %s", getErr), resp)
	}

	d.SetId(*orgMe.Id)

	resourcedata.SetNillableValue(d, "name", orgMe.Name)
	resourcedata.SetNillableValue(d, "default_language", orgMe.DefaultLanguage)
	resourcedata.SetNillableValue(d, "default_country_code", orgMe.DefaultCountryCode)
	resourcedata.SetNillableValue(d, "domain", orgMe.Domain)
	resourcedata.SetNillableValue(d, "default_site_id", orgMe.DefaultSiteId)
	resourcedata.SetNillableValue(d, "third_party_org_name", orgMe.ThirdPartyOrgName)
	resourcedata.SetNillableValue(d, "voicemail_enabled", orgMe.VoicemailEnabled)
	resourcedata.SetNillableValue(d, "product_platform", orgMe.ProductPlatform)
	resourcedata.SetNillableValue(d, "support_uri", orgMe.SupportURI)

	return nil
}

func GenerateOrganizationMe() string {
	return fmt.Sprintf(`
data "%s" "me" {}
`, DataSourceOrganizationsMeResourceType)
}
