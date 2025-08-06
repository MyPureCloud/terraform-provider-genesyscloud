package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

const homeDivisionDataSourceType = "genesyscloud_auth_division_home"

func DataSourceAuthDivisionHome() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Divisions. Get the Home division",
		ReadContext: provider.ReadWithPooledClient(dataSourceAuthDivisionHomeRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Home division name.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Description: "Home division description.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func GenerateAuthDivisionHomeDataSource(resourceLabel string) string {
	return fmt.Sprintf(`
		data "%s" "%s" {}
		`, homeDivisionDataSourceType, resourceLabel)
}

func dataSourceAuthDivisionHomeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	// Query home division
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		division, resp, getErr := authAPI.GetAuthorizationDivisionsHome()
		if getErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(homeDivisionDataSourceType, fmt.Sprintf("Error requesting divisions: %s", getErr), resp))
		}

		d.SetId(*division.Id)
		resourcedata.SetNillableValue(d, "name", division.Name)
		resourcedata.SetNillableValue(d, "description", division.Description)

		return nil
	})
}
