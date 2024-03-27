package architect_grammar

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	util "terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The data_source_genesyscloud_architect_grammar.go contains the data source implementation
   for the resource.
*/

// dataSourceArchitectGrammarRead retrieves by name the id in question
func dataSourceArchitectGrammarRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newArchitectGrammarProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		grammarId, retryable, resp, err := proxy.getArchitectGrammarIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error grammar %s: %s %v", name, err, resp))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No grammar found with name %s", name))
		}

		d.SetId(grammarId)
		return nil
	})
}
