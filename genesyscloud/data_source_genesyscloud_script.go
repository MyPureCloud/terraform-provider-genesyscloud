package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v99/platformclientv2"
)

func dataSourceScript() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Scripts. Select a script by name.",
		ReadContext: ReadWithPooledClient(dataSourceScriptRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Script name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceScriptRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	scriptsAPI := platformclientv2.NewScriptsApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query for scripts by name. Retry in case new script is not yet indexed by search.
	// As script names are non-unique, fail in case of multiple results.
	return WithRetries(ctx, 15*time.Second, func() *resource.RetryError {
		const pageSize = 100
		const pageNum = 1
		scripts, _, getErr := scriptsAPI.GetScripts(pageSize, pageNum, "", name, "", "", "", "", "", "")
		if getErr != nil {
			return resource.NonRetryableError(fmt.Errorf("Error requesting script %s: %s", name, getErr))
		}

		matchedScripts := []platformclientv2.Script{}

		if scripts.Entities != nil {
			// Since script name search is full-text, filter out non-exact matches.
			for _, script := range *scripts.Entities {
				if *script.Name == name {
					matchedScripts = append(matchedScripts, script)
				}
			}
		}

		if scripts.Entities == nil || len(matchedScripts) == 0 {
			return resource.RetryableError(fmt.Errorf("No scripts found with name %s", name))
		}

		if len(matchedScripts) > 1 {
			return resource.NonRetryableError(fmt.Errorf("Ambiguous script name: %s", name))
		}

		script := (matchedScripts)[0]
		d.SetId(*script.Id)
		return nil
	})
}
