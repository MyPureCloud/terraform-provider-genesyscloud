package flow_loglevel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
)

// getFlowLogLevelFromResourceData maps data from schema ResourceData object to a platformclientv2.Flowloglevel
func getFlowLogLevelFromResourceData(d *schema.ResourceData) *platformclientv2.Flowloglevel {
	return &platformclientv2.Flowloglevel{
		Level: platformclientv2.String(d.Get("flow_log_level").(string)),
	}
}
