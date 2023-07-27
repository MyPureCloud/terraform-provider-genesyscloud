package scripts

import (
	"context"
	"fmt"
	"time"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
DataSource for the Scripts resource
*/

// dataSourceScriptRead provides the main terraform code needed to read a script resource by name
func dataSourceScriptRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	scriptsProxy := getScriptsProxy(sdkConfig)

	name := d.Get("name").(string)

	// Query for scripts by name. Retry in case new script is not yet indexed by search.
	// As script names are non-unique, fail in case of multiple results.
	return gcloud.WithRetries(ctx, 15*time.Second, func() *resource.RetryError {
		scripts, err := scriptsProxy.getPublishedScriptsByName(ctx, name)

		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error requesting script %s: %s", name, err))
		}

		if len(*scripts) == 0 {
			return resource.RetryableError(fmt.Errorf("No scripts found with name %s", name))
		}

		if len(*scripts) > 1 {
			return resource.NonRetryableError(fmt.Errorf("Ambiguous script name: %s", name))
		}

		script := (*scripts)[0]
		d.SetId(*script.Id)
		return nil
	})
}
