package architect_flow

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFlowRead(ctx context.Context, d *schema.ResourceData, m interface{}) (diagErr diag.Diagnostics) {
	var (
		sdkConfig = m.(*provider.ProviderMeta).ClientConfig
		p         = getArchitectFlowProxy(sdkConfig)
		response  *platformclientv2.APIResponse

		name    = d.Get("name").(string)
		varType = d.Get("type").(string)
	)

	varType = strings.ToLower(varType)

	retryErr := util.WithRetries(ctx, 5*time.Second, func() *retry.RetryError {
		flowId, resp, retryable, err := p.getFlowIdByNameAndType(ctx, name, varType)
		if err != nil {
			response = resp
			if retryable {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		d.SetId(flowId)
		return nil
	})

	if retryErr != nil {
		msg := fmt.Sprintf("error retrieving ID of flow '%s' | error: %v", name, retryErr)
		diagErr = util.BuildAPIDiagnosticError(ResourceType, msg, response)
	}

	if varType == "" {
		diagErr = append(diagErr, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("'type' should be provided to the %s data source", ResourceType),
			Detail:   fmt.Sprintf("Please provide a value for the field 'type' in the %s data source as it will become required in a later version.", ResourceType),
		})
	}

	return diagErr
}
