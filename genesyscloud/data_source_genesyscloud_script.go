package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v46/platformclientv2"
)

func dataSourceScript() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Scripts. Select a script by name.",
		ReadContext: readWithPooledClient(dataSourceScriptRead),
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
	sdkConfig := m.(*providerMeta).ClientConfig
	scriptsAPI := platformclientv2.NewScriptsApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query for scripts by name. Retry in case new script is not yet indexed by search.
  // As script names are non-unique, and search is full-text, fail in case of multiple results.
	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		scripts, _, getErr := scriptsAPI.GetScripts(25, 1, "", name, "", "", "", "", "")
		if getErr != nil {
			return resource.NonRetryableError(fmt.Errorf("Error requesting script %s: %s", name, getErr))
		}

		if scripts.Entities == nil || len(*scripts.Entities) == 0 {
			return resource.RetryableError(fmt.Errorf("No scripts found with name %s", name))
		}

		if len(*scripts.Entities) > 1 {
			return resource.NonRetryableError(fmt.Errorf("Ambiguous script name: %s", name))
		}

    script := (*scripts.Entities)[0]
	  d.SetId(*script.Id)
	  return nil
	})
}
